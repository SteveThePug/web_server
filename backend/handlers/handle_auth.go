package handlers

import (
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (store *Store) AuthMiddlewear(ctx *gin.Context) {
	access_token, err := ctx.Cookie("access_token")
	if err != nil {
		ctx.AbortWithStatusJSON(401, err.Error())
		return
	}

	claims, err := store.Auth.VerifyJWT(access_token)
	if err != nil {
		ctx.AbortWithStatusJSON(401, err.Error())
		return
	}

	// store claims in Gin context
	ctx.Set("userClaims", claims)
	ctx.Next()
}

func (store *Store) CheckToken(ctx *gin.Context) {
	access_token, err := ctx.Cookie("access_token")
	if err != nil {
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	claims, err := store.Auth.VerifyJWT(access_token)
	if err != nil {
		ctx.JSON(401, err.Error())
		return
	}

	userID, ok := (*claims)["id"].(uint)
	if !ok {
		ctx.JSON(401, gin.H{"error": "claims does not contain id"})
		return
	}

	user := models.User{Model: gorm.Model{ID: userID}}
	tx := store.DB.First(&user)
	if tx.Error != nil {
		ctx.JSON(http.StatusNotFound, tx.Error.Error())
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (store *Store) RefreshToken(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		ctx.JSON(http.StatusUnauthorized,  err.Error())
		return
	}

	claims, err := store.Auth.VerifyJWT(refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized,  err.Error())
	}

	userID, ok := (*claims)["id"].(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid token claims"})
		return
	}

	user := models.User{}
	tx := store.DB.First(&user, userID)
	if tx.Error != nil {
		ctx.JSON(http.StatusNotFound, tx.Error.Error())
		return
	}

	tokens, err := store.Auth.GenerateJWT(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.SetCookie(
		"access_token",
		tokens.AccessToken,
		int(store.Auth.Config.AccessTokenLifetime.Seconds()),
		store.Auth.Config.Endpoint,
		store.Auth.Config.Domain,
		true, true,
	)
	ctx.SetCookie(
		"refresh_token",
		tokens.RefreshToken,
		int(store.Auth.Config.RefreshTokenLifetime.Seconds()),
		store.Auth.Config.Endpoint,
		store.Auth.Config.Domain,
		true, true,
	)

	ctx.JSON(http.StatusAccepted, user)
}

func (store *Store) Login(ctx *gin.Context) {
	var input UserCredentials
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest,  err.Error())
		return
	}

	user := models.User{}
	if err := store.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(input.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized,  err.Error())
		return
	}

	tokens, err := store.Auth.GenerateJWT(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError,  err.Error())
		return
	}

	ctx.SetCookie(
		"access_token",
		tokens.AccessToken,
		int(store.Auth.Config.AccessTokenLifetime.Seconds()),
		store.Auth.Config.Endpoint,
		store.Auth.Config.Domain,
		true, true,
	)
	ctx.SetCookie(
		"refresh_token",
		tokens.RefreshToken,
		int(store.Auth.Config.RefreshTokenLifetime.Seconds()),
		store.Auth.Config.Endpoint,
		store.Auth.Config.Domain,
		true, true,
	)

	ctx.JSON(http.StatusAccepted, user)
}
