package graph

import (
	"net/http"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

// setAuthCookies writes the access/refresh cookie pair on a GraphQL response.
//
// This mirrors handlers.Store.setAuthCookies — the two cannot be shared
// because importing graph from handlers would be a cycle, so any change to
// cookie attributes has to be made in both places or login via REST and login
// via GraphQL will disagree.
//
// gin's SetCookie takes Secure and HttpOnly as its two trailing booleans, both
// true here: the cookies never travel over plain HTTP and are invisible to
// JavaScript. SameSite=Lax survives a top-level navigation back from an OAuth
// provider while still blocking cross-site POSTs. Each max-age mirrors the
// lifetime baked into the token itself.
//
// It lives in its own file rather than in auth.resolvers.go because gqlgen may
// relocate hand-written code in the files it regenerates.
func setAuthCookies(gc *gin.Context, config *services.AuthConfig, tokens *services.Tokens) {
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
