package handlers

import (
	"adam-french.co.uk/backend/models"
	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

func (store *Store) ConnectWebSocket(ctx *gin.Context) {
	// Work out admin status before upgrading: the browser sends the auth
	// cookies with the upgrade request, but they are not available afterwards.
	isAdmin := store.isAdminRequest(ctx)

	conn, err := services.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		// Upgrader already wrote the HTTP error response, so just return
		return
	}

	services.HandleWebSocket(conn, isAdmin)
}

// isAdminRequest reports whether the request carries valid admin credentials.
// It prefers the access token and falls back to the refresh token plus a DB
// lookup, so an admin whose access token has just expired is still recognised.
func (store *Store) isAdminRequest(ctx *gin.Context) bool {
	if accessToken, err := ctx.Cookie("access_token"); err == nil {
		if claims, err := store.Auth.VerifyJWT(accessToken); err == nil {
			admin, ok := (*claims)["admin"].(bool)
			return ok && admin
		}
	}

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
	return user.Admin
}
