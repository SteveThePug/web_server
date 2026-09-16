package graph

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// This file is the bridge between Gin's request handling and gqlgen's
// resolvers. gqlgen resolvers receive only a context.Context, so
// AuthContextMiddleware (middleware.go) smuggles the verified JWT claims and
// the *gin.Context itself into that context, and these accessors read them
// back out.

// contextKey is a private named type so that these keys can never collide
// with a key set by another package — the reason the standard library warns
// against using bare strings as context keys. Note the string value
// "userClaims" happens to match the key handlers.AuthMiddlewear uses in the
// *Gin* context, but the two are unrelated and do not see each other.
type contextKey string

const (
	userClaimsKey contextKey = "userClaims"
	ginContextKey contextKey = "ginContext"
)

// UserClaimsFromCtx returns the verified JWT claims, or nil when the request
// is anonymous. nil means "not authenticated": a missing or invalid token is
// not an error here, because many queries are public.
func UserClaimsFromCtx(ctx context.Context) *jwt.MapClaims {
	claims, ok := ctx.Value(userClaimsKey).(*jwt.MapClaims)
	if !ok {
		return nil
	}
	return claims
}

// UserIDFromCtx returns the authenticated user's id. The claim arrives as a
// float64 because claims are decoded by encoding/json, hence the conversion.
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

// IsAdminFromCtx reports whether the request carries an admin token. This is
// the authorisation check for nearly every mutation — there is no middleware
// in front of the GraphQL endpoint, so a resolver that forgets to call this
// is public. Admin comes from the token, not the database, so a revoked admin
// flag only bites once the access token expires.
func IsAdminFromCtx(ctx context.Context) bool {
	claims := UserClaimsFromCtx(ctx)
	if claims == nil {
		return false
	}
	admin, ok := (*claims)["admin"].(bool)
	return ok && admin
}

// GinContextFromCtx returns the underlying *gin.Context, which resolvers need
// for things that live outside GraphQL's model: setting auth cookies and
// reading the client IP. Returns nil if the middleware was not installed,
// which the auth resolvers surface as an error rather than panicking.
func GinContextFromCtx(ctx context.Context) *gin.Context {
	gc, ok := ctx.Value(ginContextKey).(*gin.Context)
	if !ok {
		return nil
	}
	return gc
}
