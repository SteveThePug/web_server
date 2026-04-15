package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (store *Store) TriggerEmailSync(ctx *gin.Context) {
	if store.EmailSync == nil || store.EmailSync.HTTPClient == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "email sync not configured or not authenticated"})
		return
	}

	err := store.EmailSync.SyncEmails(ctx.Request.Context())
	if err != nil {
		log.Printf("[EmailSync] Manual sync error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "sync completed"})
}

func (store *Store) CompleteEmailAuth(ctx *gin.Context) {
	if store.EmailSync == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "email sync not configured"})
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	if err := store.EmailSync.CompleteAuth(ctx.Request.Context(), code); err != nil {
		log.Printf("[EmailSync] Auth completion error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "email authentication successful"})
}
