package graph

import (
	"context"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

// AuthContextMiddleware prepares the request context for the GraphQL handler.
//
// It always attaches the *gin.Context, and attaches the JWT claims only when a
// valid access token is present. Crucially it never rejects a request: a
// missing, expired or forged token simply leaves the context anonymous, which
// is what lets public queries work and pushes every authorisation decision
// into the individual resolvers (see IsAdminFromCtx). Both cookie and verify
// errors are swallowed for exactly that reason.
//
// Unlike the REST path there is no refresh-token fallback here, so once the
// access token expires GraphQL treats the caller as anonymous until the front
// end refreshes.
func AuthContextMiddleware(auth *services.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ginContextKey, c)

		accessToken, err := c.Cookie("access_token")
		if err == nil {
			claims, err := auth.VerifyJWT(accessToken)
			if err == nil {
				ctx = context.WithValue(ctx, userClaimsKey, claims)
			}
		}

		// The context is immutable, so the enriched one has to be written
		// back onto the request for the gqlgen handler to see it.
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
