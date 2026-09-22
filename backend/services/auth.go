package services

import (
	"errors"
	"net/http"
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Auth mints and verifies the HS256 JWTs that back the site's login. Tokens
// are handed to the browser as HTTP-only cookies (see handlers.Store.Login).
//
// There is no token blacklist, but there is coarse server-side revocation:
// every token carries the user's models.User.TokenVersion in a "tv" claim,
// and AuthenticateToken rejects a token whose version has moved on. Bumping
// that column (BumpTokenVersion) therefore kills all of a user's outstanding
// tokens at once — which is what logout and a password change do.
type Auth struct {
	Config *AuthConfig
}

// AuthConfig carries the signing secret and cookie/token lifetimes. Domain is
// the cookie domain; Endpoint is the backend's public base URL.
type AuthConfig struct {
	Secret               []byte
	Domain               string
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
	Endpoint             string
}

// Tokens is a freshly minted access/refresh token pair, both already signed.
type Tokens struct {
	AccessToken  string
	RefreshToken string
}

// InitAuth builds the Auth service from config. An absent or too-short
// secret is fatal: HS256 happily signs and verifies with a zero-length key,
// so starting up with an empty BACKEND_SECRET would let anyone forge admin
// tokens. Fail loudly at boot rather than serving a forgeable auth system.
func InitAuth(config *AuthConfig) *Auth {
	if len(config.Secret) < 32 {
		panic("auth secret must be set and at least 32 bytes (check BACKEND_SECRET)")
	}

	return &Auth{Config: config}
}

// GenerateJWT issues an access/refresh pair for user.
//
// The access token carries the authorisation data the API needs (id, admin)
// so that most requests can be authorised without touching the database. The
// refresh token deliberately carries only the id (plus the version): admin
// status is re-read from the database when refreshing, so revoking someone's
// admin flag takes effect on their next refresh rather than only when they
// log in again.
//
// Both tokens carry "tv", the user's TokenVersion at minting time. Anything
// that validates a token against the database checks it, so the pair dies the
// moment the column is bumped.
func (auth *Auth) GenerateJWT(user *models.User) (*Tokens, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"id":       user.ID,
		"admin":    user.Admin,
		"tv":       tokenVersionOf(user.TokenVersion),
		"exp":      time.Now().Add(auth.Config.AccessTokenLifetime).Unix(),
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"tv":  tokenVersionOf(user.TokenVersion),
		"exp": time.Now().Add(auth.Config.RefreshTokenLifetime).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(auth.Config.Secret)
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := refreshToken.SignedString(auth.Config.Secret)
	if err != nil {
		return nil, err
	}

	return &Tokens{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

// keyFunc supplies the HMAC secret to the JWT parser. The token argument is
// ignored because only one key is ever in use; VerifyJWT separately pins the
// algorithm, which is what actually prevents the "alg: none"/algorithm
// confusion attack.
func (auth *Auth) keyFunc(_ *jwt.Token) (any, error) {
	return auth.Config.Secret, nil
}

// VerifyJWT parses and validates a signed token, returning its claims.
// Expiry is checked by the jwt library. Note the *jwt.MapClaims return: a
// pointer to a map is unusual, but every caller in this codebase expects it,
// so the type is part of the internal contract. Numeric claims come back as
// float64 because that is what encoding/json produces, hence the
// `(*claims)["id"].(float64)` conversions at the call sites.
func (auth *Auth) VerifyJWT(tokenStr string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, auth.keyFunc, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid token claims type")
	}

	return &claims, nil
}

// Typed errors returned by AuthenticateToken so each call site can keep its
// own response for each case — the REST and GraphQL entry points deliberately
// answer differently (401 vs 500 vs 404-plus-cookie-clear).
var (
	// ErrInvalidToken means the token is missing, malformed, expired or not
	// correctly signed.
	ErrInvalidToken = errors.New("invalid token")
	// ErrInvalidClaims means the token verified but its "id" claim is absent
	// or not a number — a server-side inconsistency, not a client error.
	ErrInvalidClaims = errors.New("invalid token claims")
	// ErrUserNotFound means the token named a user that no longer exists (or
	// has been soft deleted).
	ErrUserNotFound = errors.New("user not found")
	// ErrTokenRevoked means the token is otherwise valid but was minted
	// before the user's TokenVersion was bumped, i.e. logout or a password
	// change has invalidated it.
	ErrTokenRevoked = errors.New("token revoked")
)

// tokenVersionOf normalises a stored or claimed token version.
//
// Zero means "unset": either a row that predates the token_version column, or
// a token minted before the "tv" claim existed. Both are normalised to 1, the
// column's default, so introducing versioning logs nobody out. The first
// BumpTokenVersion moves a user to 2 and invalidates every older token.
func tokenVersionOf(v uint) uint {
	if v == 0 {
		return 1
	}
	return v
}

// AuthenticateToken is the one place a token string is turned into a user: it
// verifies the signature and expiry, reads the id claim, loads the row, and
// checks the token version. Every caller that used to hand-roll this sequence
// (handlers.Store.CheckToken, handlers.Store.RefreshToken,
// handlers.Store.tryRefreshAndValidateAdmin and the GraphQL refreshToken
// resolver) now goes through here, so the checks cannot drift apart.
//
// The returned error is always one of the sentinels above, wrapped or not, so
// call sites can map each case onto whatever status code they already used.
func (auth *Auth) AuthenticateToken(db *gorm.DB, tokenStr string) (*models.User, error) {
	user, _, err := auth.AuthenticateTokenWithClaims(db, tokenStr)
	return user, err
}

// AuthenticateTokenWithClaims is AuthenticateToken plus the raw claims, for
// the middlewares that go on to read claims the User struct does not carry
// (notably "admin", which is authoritative from the token rather than the
// row — see handlers.AdminMiddleware).
//
// Note this costs one primary-key SELECT per authenticated request. That is
// the price of revocation: without loading the row there is no version to
// compare against, and logout could not kill a live access token.
func (auth *Auth) AuthenticateTokenWithClaims(db *gorm.DB, tokenStr string) (*models.User, *jwt.MapClaims, error) {
	claims, err := auth.VerifyJWT(tokenStr)
	if err != nil {
		return nil, nil, ErrInvalidToken
	}

	// Claims are decoded by encoding/json, so every number arrives as a
	// float64 regardless of the Go type it was signed from.
	userIDF, ok := (*claims)["id"].(float64)
	if !ok {
		return nil, nil, ErrInvalidClaims
	}

	// Pre-setting the primary key and calling First with no explicit
	// condition makes GORM look the row up by that key. Soft-deleted users
	// are excluded automatically, so a deleted account fails here.
	user := models.User{ID: uint(userIDF)}
	if err := db.First(&user).Error; err != nil {
		return nil, nil, ErrUserNotFound
	}

	// A token with no "tv" claim reads as 0 and is normalised to 1, matching
	// an un-bumped user — see tokenVersionOf.
	claimedVersion, _ := (*claims)["tv"].(float64)
	if tokenVersionOf(uint(claimedVersion)) != tokenVersionOf(user.TokenVersion) {
		return nil, nil, ErrTokenRevoked
	}

	return &user, claims, nil
}

// BumpTokenVersion invalidates every token previously issued to userID by
// incrementing the user's version counter. Callers treat a failure as
// non-fatal (logout still clears the cookies), but it must be logged: a
// silent failure here means a logout that did not actually revoke anything.
//
// COALESCE covers a pre-migration row sitting at 0 or NULL: it is treated as
// 1, so the bump lands on 2 and invalidates tokens claiming either.
func BumpTokenVersion(db *gorm.DB, userID uint) error {
	return db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("token_version", gorm.Expr("COALESCE(token_version, 1) + 1")).
		Error
}

// SetAuthCookies writes the access/refresh cookie pair.
//
// It lives in services because both handlers and graph need it and handlers
// cannot import graph (that would be an import cycle); keeping one copy is
// what stops REST login and GraphQL login drifting apart.
//
// gin's SetCookie takes Secure and HttpOnly as its two trailing booleans, both
// true here: the cookies never travel over plain HTTP and are invisible to
// JavaScript. SameSite=Lax survives a top-level navigation back from an OAuth
// provider while still blocking cross-site POSTs. Each cookie's max-age
// mirrors the lifetime baked into the token itself, so the browser drops it at
// roughly the moment it stops being accepted.
func SetAuthCookies(gc *gin.Context, config *AuthConfig, tokens *Tokens) {
	gc.SetSameSite(http.SameSiteLaxMode)
	gc.SetCookie(
		"access_token",
		tokens.AccessToken,
		int(config.AccessTokenLifetime.Seconds()),
		"/",
		config.Domain,
		true, true,
	)
	gc.SetCookie(
		"refresh_token",
		tokens.RefreshToken,
		int(config.RefreshTokenLifetime.Seconds()),
		"/",
		config.Domain,
		true, true,
	)
}

// ClearAuthCookies expires both auth cookies. The max-age of -1 is what tells
// the browser to delete them; path, domain and the Secure/HttpOnly flags must
// match those used when setting them or the browser keeps the originals.
func ClearAuthCookies(gc *gin.Context, config *AuthConfig) {
	gc.SetSameSite(http.SameSiteLaxMode)
	gc.SetCookie("access_token", "", -1, "/", config.Domain, true, true)
	gc.SetCookie("refresh_token", "", -1, "/", config.Domain, true, true)
}
