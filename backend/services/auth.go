package services

import (
	"errors"
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	Config *AuthConfig
}

type AuthConfig struct {
	Secret               []byte
	Domain               string
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
	Endpoint             string
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

func InitAuth(config *AuthConfig) *Auth {
	auth := Auth{Config: config}

	return &auth
}

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

func (auth *Auth) keyFunc(_ *jwt.Token) (any, error) {
	return auth.Config.Secret, nil
}

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
