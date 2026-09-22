package handlers

import (
	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

// ConnectWebSocket upgrades GET /ws into a chat connection. The route is
// public: anyone may chat, and admin status only adds the ability to see and
// send private messages and to delete.
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
// Both branches go through AuthenticateTokenWithClaims rather than bare
// VerifyJWT: that is what enforces the "tv" token-version claim, so a token
// revoked by logout or a password change is rejected here too. A bare
// VerifyJWT would leave /ws honouring stale tokens for their full lifetime
// while every other path rejects them.
func (store *Store) isAdminRequest(ctx *gin.Context) bool {
	if accessToken, err := ctx.Cookie("access_token"); err == nil {
		if _, claims, err := store.Auth.AuthenticateTokenWithClaims(store.DB, accessToken); err == nil {
			admin, ok := (*claims)["admin"].(bool)
			return ok && admin
		}
	}

	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		return false
	}
	user, err := store.Auth.AuthenticateToken(store.DB, refreshToken)
	if err != nil {
		return false
	}
	return user.Admin
}
