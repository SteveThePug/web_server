package handlers

import (
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
// retry.
func (store *Store) AuthMiddlewear(ctx *gin.Context) {
	access_token, err := ctx.Cookie("access_token")
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claims, err := store.Auth.VerifyJWT(access_token)
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

	claims, err := store.Auth.VerifyJWT(accessToken)
	if err != nil {
		// Expired/invalid access token — try refreshing
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

	claims, err := store.Auth.VerifyJWT(refreshToken)
	if err != nil {
		return false
	}

	userIDF, ok := (*claims)["id"].(float64)
	if !ok {
		return false
	}

	// Pre-setting ID and calling First with no explicit condition makes GORM
	// use the primary key already in the struct as the lookup. Soft-deleted
	// users are excluded automatically, so a deleted admin fails here.
	user := models.User{ID: uint(userIDF)}
	if err := store.DB.First(&user).Error; err != nil {
		return false
	}

	if !user.Admin {
		ctx.Status(http.StatusForbidden)
		return true
	}

	tokens, err := store.Auth.GenerateJWT(&user)
	if err != nil {
		return false
	}

	store.setAuthCookies(ctx, tokens)

	ctx.Status(http.StatusOK)
	return true
}

// setAuthCookies writes the access/refresh cookie pair.
//
// The two trailing booleans of gin's SetCookie are Secure and HttpOnly, both
// true here: the cookies are never sent over plain HTTP and are invisible to
// JavaScript. SameSite=Lax lets them survive a top-level navigation back from
// an OAuth provider while still blocking cross-site POSTs. Each cookie's
// max-age mirrors the lifetime baked into the token itself, so the browser
// drops it at roughly the moment it stops being accepted.
func (store *Store) setAuthCookies(ctx *gin.Context, tokens *services.Tokens) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"access_token",
		tokens.AccessToken,
		int(store.Auth.Config.AccessTokenLifetime.Seconds()),
		"/",
		store.Auth.Config.Domain,
		true, true,
	)
	ctx.SetCookie(
		"refresh_token",
		tokens.RefreshToken,
		int(store.Auth.Config.RefreshTokenLifetime.Seconds()),
		"/",
		store.Auth.Config.Domain,
		true, true,
	)
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

	claims, err := store.Auth.VerifyJWT(access_token)
	if err != nil {
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// JWT claims come back through encoding/json, so every number is a
	// float64 regardless of the Go type it was signed from.
	userIDF, ok := (*claims)["id"].(float64)
	if !ok {
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	userID := uint(userIDF)

	user := models.User{ID: userID}
	tx := store.DB.First(&user)
	if tx.Error != nil {
		log.Println(tx.Error)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		store.removeCookies(ctx)
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// RefreshToken backs POST /auth/refresh: it exchanges a valid refresh token
// for a new cookie pair and returns the user.
//
// Tokens are not rotated in any revocable sense — the old refresh token stays
// valid until it expires, because nothing is tracked server-side.
func (store *Store) RefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	claims, err := store.Auth.VerifyJWT(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDF, ok := (*claims)["id"].(float64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid token claims"})
		return
	}
	userID := uint(userIDF)

	user := models.User{ID: userID}
	tx := store.DB.First(&user)
	if tx.Error != nil {
		log.Println(tx.Error)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		store.removeCookies(ctx)
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

// Logout backs POST /auth/logout. It only clears the browser's cookies —
// the tokens themselves stay cryptographically valid until they expire, so a
// copy taken beforehand would still work.
func (store *Store) Logout(ctx *gin.Context) {
	store.removeCookies(ctx)

	ctx.Status(http.StatusOK)
}

// removeCookies expires both auth cookies. The max-age of -1 is what tells
// the browser to delete them; path, domain and the Secure/HttpOnly flags must
// match those used when setting them or the browser keeps the originals.
func (store *Store) removeCookies(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		store.Auth.Config.Domain,
		true, true,
	)
	ctx.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		store.Auth.Config.Domain,
		true, true,
	)
}
