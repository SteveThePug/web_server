package handlers

import (
	"errors"
	"log"
	"net/http"

	"adam-french.co.uk/backend/models"
	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// This file holds the REST authentication endpoints and the Gin middleware
// that guards the other REST routes. GraphQL has its own parallel
// implementation of login/logout/refresh in graph/auth.resolvers.go; both
// mint the same cookies through services.Auth.
//
// Cookie model: `access_token` is short-lived and carries id + admin, so most
// requests authorise without a database hit; `refresh_token` carries only the
// id and is exchanged for a fresh pair. Both are set HttpOnly and Secure, so
// JavaScript cannot read them and the browser attaches them automatically.

// UserCredentials is the JSON body of POST /auth/login. The binding tags make
// Gin reject a request missing either field.
type UserCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthMiddlewear requires a valid access token on a REST route and stashes
// the claims for AdminMiddleware to read.
//
// (The name is misspelled but it is exported and referenced from main.go, so
// it is left alone rather than renamed.)
//
// It deliberately does NOT fall back to the refresh token: an expired access
// token yields a 401 and the front end is expected to call /auth/refresh and
// retry. A token whose version has been revoked is also a 401.
func (store *Store) AuthMiddlewear(ctx *gin.Context) {
	access_token, err := ctx.Cookie("access_token")
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Validated against the database, not just cryptographically: that is
	// what makes a logged-out or password-changed token stop working here
	// rather than lingering until it expires.
	_, claims, err := store.Auth.AuthenticateTokenWithClaims(store.DB, access_token)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Stored under a plain string key in the *Gin* context. This is a
	// different mechanism from graph/context.go, which puts the same claims
	// into the standard library request context under an unexported typed
	// key; the two do not see each other's values.
	ctx.Set("userClaims", claims)
	ctx.Next()
}

// AdminMiddleware requires the access token to carry admin=true. It must be
// chained after AuthMiddlewear, which is what populates "userClaims"; on its
// own it always rejects.
//
// Admin is read from the token, not the database, so revoking someone's admin
// flag only takes effect once their current access token expires.
func (store *Store) AdminMiddleware(ctx *gin.Context) {
	claims, exists := ctx.Get("userClaims")
	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	mapClaims, ok := claims.(*jwt.MapClaims)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		return
	}

	// A token minted before the admin claim existed simply has no "admin"
	// key, so a failed type assertion is treated as "not an admin".
	admin, ok := (*mapClaims)["admin"].(bool)
	if !ok || !admin {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	ctx.Next()
}

// ValidateAdmin backs GET /auth/validate-admin, which the front end calls to
// decide whether to show admin UI. It answers with a bare status code and no
// body: 200 admin, 403 authenticated but not admin, 401 not authenticated.
//
// Unlike AdminMiddleware it falls back to the refresh token, and will mint a
// new cookie pair as a side effect, so a returning admin is recognised
// without a visible re-login.
func (store *Store) ValidateAdmin(ctx *gin.Context) {
	accessToken, err := ctx.Cookie("access_token")
	if err != nil {
		// No access token — try refreshing
		if !store.tryRefreshAndValidateAdmin(ctx) {
			ctx.Status(http.StatusUnauthorized)
		}
		return
	}

	_, claims, err := store.Auth.AuthenticateTokenWithClaims(store.DB, accessToken)
	if err != nil {
		// Expired, invalid or revoked access token — try refreshing
		if !store.tryRefreshAndValidateAdmin(ctx) {
			ctx.Status(http.StatusUnauthorized)
		}
		return
	}

	admin, ok := (*claims)["admin"].(bool)
	if !ok || !admin {
		ctx.Status(http.StatusForbidden)
		return
	}

	ctx.Status(http.StatusOK)
}

// tryRefreshAndValidateAdmin re-authenticates from the refresh token.
//
// The bool does not mean "is an admin" — it means "this function already
// wrote the response". It returns true both for a successful admin refresh
// (200, new cookies set) and for a valid non-admin (403), and false only when
// it could not decide, leaving the caller to send 401. Getting this backwards
// would send two status codes for one request.
//
// Admin status is re-read from the database here, which is what makes the
// refresh path pick up a revoked admin flag.
func (store *Store) tryRefreshAndValidateAdmin(ctx *gin.Context) bool {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		return false
	}

	// Any failure at all (bad token, unknown user, revoked version) is
	// "cannot decide" here, exactly as before: the caller sends 401.
	user, err := store.Auth.AuthenticateToken(store.DB, refreshToken)
	if err != nil {
		return false
	}

	if !user.Admin {
		ctx.Status(http.StatusForbidden)
		return true
	}

	tokens, err := store.Auth.GenerateJWT(user)
	if err != nil {
		return false
	}

	store.setAuthCookies(ctx, tokens)

	ctx.Status(http.StatusOK)
	return true
}

// setAuthCookies writes the access/refresh cookie pair. The cookie
// attributes live in services.SetAuthCookies so that REST and GraphQL cannot
// disagree about them; this method is just the Store-shaped spelling of it.
func (store *Store) setAuthCookies(ctx *gin.Context, tokens *services.Tokens) {
	services.SetAuthCookies(ctx, store.Auth.Config, tokens)
}

// CheckToken backs GET /auth/check: it returns the current user for a valid
// access token, or 401. If the token is valid but the user no longer exists
// it answers 404 and clears the cookies, so the browser stops presenting a
// token for a deleted account.
func (store *Store) CheckToken(ctx *gin.Context) {
	access_token, err := ctx.Cookie("access_token")
	if err != nil {
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Status codes here are unchanged from the hand-rolled version: a bad
	// token or bad claims are both 401, and only a missing user is 404 with
	// the cookies cleared. A revoked (out-of-date version) token joins the
	// 401 case, since from the client's point of view it is simply no longer
	// a valid token.
	user, err := store.Auth.AuthenticateToken(store.DB, access_token)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			log.Println(err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			store.removeCookies(ctx)
			return
		}
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// RefreshToken backs POST /auth/refresh: it exchanges a valid refresh token
// for a new cookie pair and returns the user.
//
// The old refresh token is not rotated, but it does die when the user logs out
// or changes their password, because those bump the token version this checks.
func (store *Store) RefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// The three outcomes below preserve the original status codes exactly:
	// an unusable token is 401, unreadable claims are 500, and a token for a
	// user that no longer exists is 404 with the cookies cleared. A revoked
	// token is treated as an unusable one (401) — the client's correct
	// response is to log in again.
	user, err := store.Auth.AuthenticateToken(store.DB, refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidClaims):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid token claims"})
		case errors.Is(err, services.ErrUserNotFound):
			log.Println(err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			store.removeCookies(ctx)
		default:
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		}
		return
	}

	tokens, err := store.Auth.GenerateJWT(user)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	store.setAuthCookies(ctx, tokens)

	ctx.JSON(http.StatusAccepted, user)
}

// Login backs POST /auth/login. The user object it returns omits the password
// because models.User tags that field `json:"-"`.
func (store *Store) Login(ctx *gin.Context) {
	// Per-IP limit (5/min) on top of nginx's own login rate limit.
	// ctx.ClientIP is only trustworthy because main.go pins the trusted
	// proxy range, so the X-Forwarded-For header cannot be spoofed by a
	// client talking through nginx.
	if !store.LoginLimiter.Allow(ctx.ClientIP()) {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, please try again later"})
		return
	}

	var input UserCredentials
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Unknown user and wrong password return the identical message, so the
	// response cannot be used to enumerate usernames. (The timing does
	// differ, since only the second path runs bcrypt.)
	user := models.User{}
	if err := store.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(input.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	tokens, err := store.Auth.GenerateJWT(&user)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	store.setAuthCookies(ctx, tokens)

	ctx.JSON(http.StatusAccepted, user)
}

// Logout backs POST /auth/logout. As well as clearing the browser's cookies
// it bumps the user's token version, so the tokens it just discarded — and
// any copy taken beforehand — stop being accepted immediately.
//
// It always answers 200: identifying the caller is best effort (the cookies
// may already be gone or expired), and a client that cannot be identified has
// nothing to revoke anyway.
func (store *Store) Logout(ctx *gin.Context) {
	store.revokeCallerTokens(ctx)
	store.removeCookies(ctx)

	ctx.Status(http.StatusOK)
}

// revokeCallerTokens bumps the token version of whoever made this request,
// identified from the access token or, failing that, the refresh token.
//
// It deliberately verifies the token rather than just reading the id out of
// it: without that, anyone could bump an arbitrary user's version by posting a
// forged cookie, which is a denial-of-service on that account's sessions.
func (store *Store) revokeCallerTokens(ctx *gin.Context) {
	for _, name := range []string{"access_token", "refresh_token"} {
		token, err := ctx.Cookie(name)
		if err != nil {
			continue
		}
		user, err := store.Auth.AuthenticateToken(store.DB, token)
		if err != nil {
			continue
		}
		if err := services.BumpTokenVersion(store.DB, user.ID); err != nil {
			// Logged rather than returned: the cookies are still cleared, but
			// this means the logout did not actually revoke anything.
			log.Printf("logout: failed to revoke tokens for user %d: %v", user.ID, err)
		}
		return
	}
}

// removeCookies expires both auth cookies; see services.ClearAuthCookies for
// why the attributes have to match those used when setting them.
func (store *Store) removeCookies(ctx *gin.Context) {
	services.ClearAuthCookies(ctx, store.Auth.Config)
}
