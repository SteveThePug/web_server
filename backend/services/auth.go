package services

import (
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	config *AuthConfig
}

type AuthConfig struct {
	Secret []byte
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

func InitAuth(config *AuthConfig) *Auth {
	auth := Auth{config: config}

	return &auth
}

func (auth *Auth) GenerateJWT(user *models.User) (*Tokens, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"id":       user.ID,
		"admin":    user.Admin,
		"exp":      time.Now().AddDate(0, 0, 1).Unix(),
	})

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().AddDate(1, 0, 0).Unix(),
	})

	accessTokenString, err := accessToken.SignedString(auth.config.Secret)
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := refreshToken.SignedString(auth.config.Secret)
	if err != nil {
		return nil, err
	}

	return &Tokens{AccessToken: accessTokenString, RefreshToken: refreshTokenString}, nil
}

func (auth *Auth) VerifyJWT(tokens Tokens) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokens.AccessToken, func(token *jwt.Token) (any, error) {
		return auth.config.Secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, err
}
