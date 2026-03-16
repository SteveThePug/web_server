package graph

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userClaimsKey contextKey = "userClaims"
	ginContextKey contextKey = "ginContext"
)

func UserClaimsFromCtx(ctx context.Context) *jwt.MapClaims {
	claims, ok := ctx.Value(userClaimsKey).(*jwt.MapClaims)
	if !ok {
		return nil
	}
	return claims
}

func UserIDFromCtx(ctx context.Context) (uint, bool) {
	claims := UserClaimsFromCtx(ctx)
	if claims == nil {
		return 0, false
	}
	idF, ok := (*claims)["id"].(float64)
	if !ok {
		return 0, false
	}
	return uint(idF), true
}

func IsAdminFromCtx(ctx context.Context) bool {
	claims := UserClaimsFromCtx(ctx)
	if claims == nil {
		return false
	}
	admin, ok := (*claims)["admin"].(bool)
	return ok && admin
}

func GinContextFromCtx(ctx context.Context) *gin.Context {
	gc, ok := ctx.Value(ginContextKey).(*gin.Context)
	if !ok {
		return nil
	}
	return gc
}
