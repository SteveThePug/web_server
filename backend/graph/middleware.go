package graph

import (
	"context"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
// It takes the database because validating a token means loading the user to
// compare token versions; see services.Auth.AuthenticateTokenWithClaims.
//
// Unlike the REST path there is no refresh-token fallback here, so once the
// access token expires GraphQL treats the caller as anonymous until the front
// end refreshes.
func AuthContextMiddleware(auth *services.Auth, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ginContextKey, c)

		accessToken, err := c.Cookie("access_token")
		if err == nil {
			// The database is needed to check the token version, so a
			// revoked token (logout, password change) leaves the request
			// anonymous instead of continuing to authorise mutations.
			_, claims, err := auth.AuthenticateTokenWithClaims(db, accessToken)
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
