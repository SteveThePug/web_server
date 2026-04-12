package handlers

import (
	"log"
	"net/http"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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

	// store claims in Gin context
	ctx.Set("userClaims", claims)
	ctx.Next()
}

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

	admin, ok := (*mapClaims)["admin"].(bool)
	if !ok || !admin {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	ctx.Next()
}

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

	ctx.Status(http.StatusOK)
	return true
}

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

	ctx.JSON(http.StatusAccepted, user)
}

func (store *Store) Login(ctx *gin.Context) {
	var input UserCredentials
	if err := ctx.ShouldBindBodyWithJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

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

	ctx.JSON(http.StatusAccepted, user)
}

func (store *Store) Logout(ctx *gin.Context) {
	store.removeCookies(ctx)

	ctx.Status(http.StatusOK)
}

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
