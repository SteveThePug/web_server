package handlers

import (
	"errors"
	"log"
	"net/http"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Upload backs POST /uploads (login required): the one endpoint the frontend
// stores a file through (vue/src/uploads.js). Multipart form:
//
//	file     the file (required)
//	private  "true" to store it as an admin-only chat attachment
//
// It answers {"url": "/uploads/<name>"}, or {"error": "..."} with 400 when
// the file was refused (wrong type, too large) and 500 when storing failed.
// Validation, naming and storage are all services.Uploads.Save; this handler
// only reads the form and decides who may use `private`.
//
// `private` is admin-only because private messages themselves are: an
// attachment URL leaking out is the same disclosure as the message leaking.
func (store *Store) Upload(ctx *gin.Context) {
	private := ctx.PostForm("private") == "true"
	if private && !requestIsAdmin(ctx) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	url, err := store.Uploads.Save(ctx.Request.Context(), file, private)
	var rejected *services.UploadError
	switch {
	case errors.As(err, &rejected):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": rejected.Message})
	case err != nil:
		log.Printf("[upload] storing %s failed: %v", file.Filename, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
	default:
		ctx.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// requestIsAdmin reports whether the caller's access token carries admin=true.
//
// This route sits in the `protected` group, which runs AuthMiddlewear but not
// AdminMiddleware — the endpoint has to stay usable by any signed-in user for
// public attachments — so the admin check is made here from the same
// "userClaims" value AdminMiddleware would have read.
func requestIsAdmin(ctx *gin.Context) bool {
	claims, exists := ctx.Get("userClaims")
	if !exists {
		return false
	}
	mapClaims, ok := claims.(*jwt.MapClaims)
	if !ok {
		return false
	}
	admin, ok := (*mapClaims)["admin"].(bool)

	return ok && admin
}
