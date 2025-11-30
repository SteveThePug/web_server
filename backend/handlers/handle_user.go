package handlers

import (
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (store *Store) CreateUser(ctx *gin.Context) {
	var input UserCredentials
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user := models.User{Username: input.Username, Password: hashedPassword}
	store.DB.Create(&user)

	// Generate JWT token
	tokens, err := store.Auth.GenerateJWT(&user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (store *Store) GetUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	var user models.User
	if err := store.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

func (store *Store) GetUsers(ctx *gin.Context) {
	var users []models.User
	if err := store.DB.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": users})
}

func (store *Store) UpdateUser(ctx *gin.Context) {
	claimsVal, ok := ctx.Get("userClaims")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user claims could not be found"})
		return
	}

	claims, ok := claimsVal.(*jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims"})
		return
	}

	userID, ok := (*claims)["id"].(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}

	var user models.User
	if err := store.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "will be implemented"})
}

func (store *Store) DeleteUser(ctx *gin.Context) {
	claimsVal, ok := ctx.Get("userClaims")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user claims could not be found"})
		return
	}

	claims, ok := claimsVal.(*jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid claims"})
		return
	}

	userID, ok := (*claims)["id"].(uint)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}

	var user models.User
	if err := store.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	store.DB.Delete(&user)
	ctx.JSON(http.StatusOK, gin.H{"data": user})

}
