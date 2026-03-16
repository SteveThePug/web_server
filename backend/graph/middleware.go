package graph

import (
	"context"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

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

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
