package services

import (
	"errors"
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/golang-jwt/jwt/v5"
)

// Auth mints and verifies the HS256 JWTs that back the site's login. Tokens
// are handed to the browser as HTTP-only cookies (see handlers.Store.Login);
// there is no server-side session or token blacklist, so a token stays valid
// until it expires even after logout.
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

// InitAuth builds the Auth service from config.
func InitAuth(config *AuthConfig) *Auth {
	auth := Auth{Config: config}

	return &auth
}

// GenerateJWT issues an access/refresh pair for user.
//
// The access token carries the authorisation data the API needs (id, admin)
// so that most requests can be authorised without touching the database. The
// refresh token deliberately carries only the id: admin status is re-read
// from the database when refreshing, so revoking someone's admin flag takes
// effect on their next refresh rather than only when they log in again.
func (auth *Auth) GenerateJWT(user *models.User) (*Tokens, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"id":       user.ID,
		"admin":    user.Admin,
		"exp":      time.Now().Add(auth.Config.AccessTokenLifetime).Unix(),
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
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
