package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"adam-french.co.uk/backend/models"
	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

// Job capture: the three endpoints behind vue/public/job-capture.user.js,
// which adds the posting open in the browser to the job tracker.
//
// These are REST, and token-authenticated, because the caller is a userscript
// running on someone else's origin (linkedin.com, an ATS). The auth cookies
// are SameSite=Lax and never accompany such a request, so the GraphQL
// mutations, which rely on them, are out of reach from there.
//
// The token is deliberately narrow: it opens these routes and nothing else,
// so a token leaked from the browser can add rows to the tracker and spend
// Claude calls, but cannot read the tracker, edit other data or log in.
//
// As with rowing, the Claude call lives in the Python service
// (python/app/jobs.py) and everything touching the database stays here.

const (
	// jobExtractPath is internal in the same way as rowingReadPath: nginx
	// refuses /py/internal/ and the route refuses anything nginx forwarded.
	jobExtractPath = "/internal/jobs/extract"

	// maxCaptureBody bounds a capture request. A long posting is tens of
	// kilobytes of text; the userscript trims the page to maxPageRunes before
	// sending, so anything near this limit is not a job page.
	maxCaptureBody = 512 << 10

	// maxPageRunes caps both the text sent to Claude and the description
	// stored. It mirrors MAX_TEXT_CHARS in python/app/jobs.py, which rejects
	// anything longer, so the cut has to happen on this side.
	maxPageRunes = 40_000
)

// CaptureTokenMiddleware guards the capture routes with the token generated
// in the admin panel (services.CaptureToken), presented as
// "Authorization: Bearer <token>".
//
// Until a token has been generated the routes are off (503) rather than
// open; "no token" must never be satisfied by a request that sends none.
func (store *Store) CaptureTokenMiddleware(ctx *gin.Context) {
	if store.CaptureToken == nil {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "job capture is not configured"})
		return
	}
	if set, _ := store.CaptureToken.Status(); !set {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "no capture token has been generated"})
		return
	}
	presented, ok := strings.CutPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if !ok || !store.CaptureToken.Verify(presented) {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx.Next()
}

// The three handlers below manage that token from the admin panel. They sit
// in the cookie-authenticated admin group, not behind the token: the token
// must not be able to replace itself.

func captureTokenStatus(token *services.CaptureToken) gin.H {
	set, createdAt := token.Status()
	if !set {
		return gin.H{"set": false}
	}
	return gin.H{"set": true, "createdAt": createdAt}
}

// GetCaptureToken backs GET /admin/capture-token: whether a token exists and
// when it was made. The token itself is not recoverable — only its hash is
// stored.
func (store *Store) GetCaptureToken(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, captureTokenStatus(store.CaptureToken))
}

// GenerateCaptureToken backs POST /admin/capture-token. It replaces any
// existing token, so the userscript must be reconfigured with the new one,
// and its response is the only time the token is ever sent anywhere.
func (store *Store) GenerateCaptureToken(ctx *gin.Context) {
	token, createdAt, err := store.CaptureToken.Generate()
	if err != nil {
		log.Println("Generating capture token failed:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store the token"})
		return
	}
	// The response carries a credential; keep it out of any cache.
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, gin.H{"token": token, "createdAt": createdAt})
}

// RevokeCaptureToken backs DELETE /admin/capture-token, switching the capture
// routes off until a new token is generated.
func (store *Store) RevokeCaptureToken(ctx *gin.Context) {
	if err := store.CaptureToken.Revoke(); err != nil {
		log.Println("Revoking capture token failed:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke the token"})
		return
	}
	ctx.JSON(http.StatusOK, captureTokenStatus(store.CaptureToken))
}

// capturedJob is the slice of an application the userscript needs back: enough
// to show "already tracked as …" and no more. The token is not meant to be
// able to read the tracker, so notes and the description stay out of it.
type capturedJob struct {
	ID       uint   `json:"id"`
	JobTitle string `json:"jobTitle"`
	Company  string `json:"company"`
	Status   string `json:"status"`
}

func capturedJobOf(app *models.JobApplication) *capturedJob {
	return &capturedJob{ID: app.ID, JobTitle: app.JobTitle, Company: app.Company, Status: app.Status}
}

// JobCaptureStatus backs GET /jobs/capture/status?url=. It is what lets the
// userscript show a posting's tracked status instead of an "add" button.
func (store *Store) JobCaptureStatus(ctx *gin.Context) {
	raw := ctx.Query("url")
	canonical, ok := services.CanonicalJobURL(raw)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "url must be an absolute http(s) URL"})
		return
	}
	app, err := services.FindApplicationByURL(store.DB, canonical, raw)
	if err != nil {
		log.Println("Job capture status lookup failed:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}
	if app == nil {
		ctx.JSON(http.StatusOK, gin.H{"found": false})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"found": true, "application": capturedJobOf(app)})
}

// ExtractedJobPosting is the JSON contract for what Claude reads off a job
// page: the body of the Python service's reply (JobPosting in
// python/app/jobs.py). Empty strings and nulls mean "not found on the page".
type ExtractedJobPosting struct {
	JobTitle string  `json:"jobTitle"`
	Company  string  `json:"company"`
	Location *string `json:"location"`
	Notes    *string `json:"notes"`
}

// extractJobPosting sends a page's text to the Python service and returns the
// fields Claude picked out. Like readRowingDisplay it is a thin wrapper over
// services.CallPython, which decides which replies are relayed (a
// *services.PythonError) and which are only logged.
func (store *Store) extractJobPosting(ctx context.Context, pageURL, title, text string) (ExtractedJobPosting, error) {
	return services.CallPython[ExtractedJobPosting](ctx, store.PythonURL, jobExtractPath, struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Text  string `json:"text"`
	}{pageURL, title, text})
}

// truncateRunes cuts s to at most max runes, on a rune boundary so the result
// is still valid UTF-8 for Postgres and for JSON.
func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

// bindCapture decodes a size-limited JSON body, answering 400 itself when it
// cannot.
func bindCapture(ctx *gin.Context, into any) bool {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxCaptureBody)
	if err := ctx.ShouldBindJSON(into); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return false
	}
	return true
}

// JobCaptureExtract backs POST /jobs/capture/extract: page text in, the
// popup's pre-filled fields out. Nothing is stored; saving is a separate
// request made once the fields have been checked by eye.
func (store *Store) JobCaptureExtract(ctx *gin.Context) {
	var input struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Text  string `json:"text"`
	}
	if !bindCapture(ctx, &input) {
		return
	}
	text := strings.TrimSpace(input.Text)
	if text == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "text is required"})
		return
	}

	posting, err := store.extractJobPosting(ctx.Request.Context(),
		truncateRunes(input.URL, 2000), truncateRunes(input.Title, 500), truncateRunes(text, maxPageRunes))
	if err != nil {
		log.Println("Job extract failed:", err)
		var readErr *services.PythonError
		if errors.As(err, &readErr) {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": readErr.Detail})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "failed to read the page"})
		return
	}
	ctx.JSON(http.StatusOK, posting)
}

// JobCapture backs POST /jobs/capture, an upsert: a posting already in the
// tracker is updated in place rather than added again. The matching and
// merging rules are services.UpsertJobApplication's; this handler owns the
// request body and the status codes.
func (store *Store) JobCapture(ctx *gin.Context) {
	var input services.JobCaptureInput
	if !bindCapture(ctx, &input) {
		return
	}
	input.Description = truncateRunes(strings.TrimSpace(input.Description), maxPageRunes)

	app, action, err := services.UpsertJobApplication(store.DB, input, time.Now())
	var invalid *services.JobInputError
	switch {
	case errors.As(err, &invalid):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": invalid.Error()})
	case err != nil:
		log.Println("Job capture failed:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
	case action == "created":
		ctx.JSON(http.StatusCreated, gin.H{"action": action, "application": capturedJobOf(app)})
	default:
		ctx.JSON(http.StatusOK, gin.H{"action": action, "application": capturedJobOf(app)})
	}
}
