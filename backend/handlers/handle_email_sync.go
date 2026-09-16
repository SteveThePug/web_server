package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// REST endpoints for the job-application email pipeline: a manual
// admin-triggered sync, and the OAuth redirect target (which has to be REST
// because Microsoft redirects a browser to it).

// TriggerEmailSync backs the admin-only POST /email/sync and runs one sync
// synchronously, so the response reflects the result. If the scheduler is
// mid-sync the service refuses via TryLock and this returns 500.
func (store *Store) TriggerEmailSync(ctx *gin.Context) {
	// A nil HTTPClient is the service's readiness flag — see
	// services.EmailSyncService.
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

// CompleteEmailAuth handles Microsoft's OAuth redirect to
// GET /email/callback and stores the resulting token.
//
// The route is unauthenticated (Microsoft's redirect carries no site cookie),
// which is safe only because the code is worthless without the app's client
// secret. Note that completing auth here does not start the scheduler, which
// only checks readiness at start-up — a restart is needed for that.
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
