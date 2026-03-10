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

type SetAdminInput struct {
	Admin *bool `json:"admin" binding:"required"`
}

func (store *Store) CreateUser(ctx *gin.Context) {
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
	if !(*claims)["admin"].(bool) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var input UserCredentials
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	user := models.User{Username: input.Username, Password: hashedPassword}
	tx := store.DB.Create(&user)
	if tx.Error != nil {
		ctx.JSON(http.StatusInternalServerError, tx.Error.Error())
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (store *Store) GetUser(ctx *gin.Context) {
	userID := ctx.Param("id")
	var user models.User
	if err := store.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (store *Store) GetUsers(ctx *gin.Context) {
	var users []models.User
	if err := store.DB.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, users)
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

	userIDF, ok := (*claims)["id"].(float64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}
	userID := uint(userIDF)

	var user models.User
	if err := store.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, err.Error())
		return
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "will be implemented"})
}

func (store *Store) SetUserAdmin(ctx *gin.Context) {
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
	if !(*claims)["admin"].(bool) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	callerIDF, ok := (*claims)["id"].(float64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}
	callerID := uint(callerIDF)

	targetID := ctx.Param("id")

	var input SetAdminInput
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := store.DB.First(&user, targetID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.ID == callerID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own admin status"})
		return
	}

	user.Admin = *input.Admin
	if err := store.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
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

	userIDF, ok := (*claims)["id"].(float64)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in claims"})
		return
	}
	userID := uint(userIDF)

	var user models.User
	if err := store.DB.First(&user, userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, err.Error())
		return
	}

	tx := store.DB.Delete(&user)
	if tx.Error != nil {
		ctx.JSON(http.StatusInternalServerError, tx.Error.Error())
		return
	}

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		"access_token",
		"",
		-1,
		store.Auth.Config.Endpoint,
		store.Auth.Config.Domain,
		true, true,
	)
	ctx.SetCookie(
		"refresh_token",
		"",
		-1,
		store.Auth.Config.Endpoint,
		store.Auth.Config.Domain,
		true, true,
	)

	ctx.JSON(http.StatusOK, user)
}
