package handlers

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"adam-french.co.uk/backend/models"
	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Job capture: the three endpoints behind userscripts/job-capture.user.js,
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

// jobExtractClient bounds the wait on the Python service; see
// rowingReadClient for why two minutes.
var jobExtractClient = &http.Client{Timeout: 2 * time.Minute}

// CaptureTokenMiddleware guards the capture routes with JOB_CAPTURE_TOKEN,
// presented as "Authorization: Bearer <token>".
//
// An unset token turns the routes off (503) rather than leaving them open:
// an empty env var is the stack's failure mode for a forgotten secret (see
// .env.example), and comparing "" to "" would otherwise let everyone in.
func (store *Store) CaptureTokenMiddleware(ctx *gin.Context) {
	if store.JobCaptureToken == "" {
		ctx.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "job capture is not configured"})
		return
	}
	presented, ok := strings.CutPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	if !ok || subtle.ConstantTimeCompare([]byte(presented), []byte(store.JobCaptureToken)) != 1 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	ctx.Next()
}

// trackingParams are query parameters that say how the visitor arrived, not
// which posting this is. Anything not listed is kept: ATS boards identify the
// job in the query string often enough (gh_jid, jk, jobId) that dropping
// unknown parameters would merge different postings.
var trackingParams = map[string]bool{
	"trk": true, "trackingid": true, "refid": true, "ebp": true,
	"gh_src": true, "lever-source": true, "lever-origin": true,
	"source": true, "src": true, "ref": true,
	"fbclid": true, "gclid": true, "mc_cid": true, "mc_eid": true,
}

// linkedInJobPath matches both forms of a LinkedIn posting path: the bare
// /jobs/view/123 and the public slug form /jobs/view/some-title-at-co-123.
var linkedInJobPath = regexp.MustCompile(`^/jobs/view/(?:.*-)?(\d+)/?$`)

// canonicalJobURL reduces a posting URL to a form that is the same however
// the page was reached, so it can serve as the key that decides whether a
// posting is already tracked. ok is false for anything that is not an
// absolute http(s) URL.
//
// LinkedIn gets its own rule because the same posting appears under several
// paths (search results, collections, the recommended feed), all carrying the
// id in ?currentJobId= rather than in the path.
func canonicalJobURL(raw string) (canonical string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", false
	}
	host := strings.ToLower(u.Host)
	query := u.Query()

	if host == "linkedin.com" || strings.HasSuffix(host, ".linkedin.com") {
		id := query.Get("currentJobId")
		if m := linkedInJobPath.FindStringSubmatch(u.Path); m != nil {
			id = m[1]
		}
		if id != "" && strings.Trim(id, "0123456789") == "" {
			return "https://www.linkedin.com/jobs/view/" + id + "/", true
		}
	}

	for key := range query {
		lowered := strings.ToLower(key)
		if trackingParams[lowered] || strings.HasPrefix(lowered, "utm_") {
			query.Del(key)
		}
	}

	out := url.URL{
		Scheme: u.Scheme,
		Host:   host,
		// A trailing slash is not a different posting.
		Path: strings.TrimRight(u.Path, "/"),
		// Encode sorts by key, so parameter order does not matter either.
		RawQuery: query.Encode(),
	}
	return out.String(), true
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

// findByURL returns the newest live application stored under this posting's
// URL, or nil. The raw form is tried as well as the canonical one because rows
// typed into the tracker by hand hold whatever was pasted.
func (store *Store) findByURL(canonical, raw string) (*models.JobApplication, error) {
	return firstApplication(store.DB.Where("url IN ?", []string{canonical, strings.TrimSpace(raw)}))
}

// firstApplication returns the newest row the query matches, or nil for none.
// Limit(1).Find rather than First: a miss is the ordinary answer here (every
// badge lookup on an untracked posting), and First reports one as
// ErrRecordNotFound, which GORM writes to the log as an error each time.
func firstApplication(query *gorm.DB) (*models.JobApplication, error) {
	var apps []models.JobApplication
	if err := query.Order("created_at DESC, id DESC").Limit(1).Find(&apps).Error; err != nil {
		return nil, err
	}
	if len(apps) == 0 {
		return nil, nil
	}
	return &apps[0], nil
}

// JobCaptureStatus backs GET /jobs/capture/status?url=. It is what lets the
// userscript show a posting's tracked status instead of an "add" button.
func (store *Store) JobCaptureStatus(ctx *gin.Context) {
	raw := ctx.Query("url")
	canonical, ok := canonicalJobURL(raw)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "url must be an absolute http(s) URL"})
		return
	}
	app, err := store.findByURL(canonical, raw)
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

// postingReadError is a failure the Python service reported in words that are
// safe to show in the capture popup (its 502 detail); the job-capture
// counterpart of displayReadError.
type postingReadError struct{ detail string }

func (e *postingReadError) Error() string { return e.detail }

// extractJobPosting sends a page's text to the Python service and returns the
// fields Claude picked out. It mirrors readRowingDisplay, including which
// replies are relayed and which are only logged.
func (store *Store) extractJobPosting(ctx context.Context, pageURL, title, text string) (ExtractedJobPosting, error) {
	var posting ExtractedJobPosting

	body, err := json.Marshal(struct {
		URL   string `json:"url"`
		Title string `json:"title"`
		Text  string `json:"text"`
	}{pageURL, title, text})
	if err != nil {
		return posting, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, store.PythonURL+jobExtractPath, bytes.NewReader(body))
	if err != nil {
		return posting, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := jobExtractClient.Do(req)
	if err != nil {
		return posting, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var failure struct {
			Detail string `json:"detail"`
		}
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		if resp.StatusCode == http.StatusBadGateway && json.Unmarshal(raw, &failure) == nil && failure.Detail != "" {
			return posting, &postingReadError{detail: failure.Detail}
		}
		return posting, fmt.Errorf("job extractor returned %d: %s", resp.StatusCode, raw)
	}

	if err := json.NewDecoder(resp.Body).Decode(&posting); err != nil {
		return posting, fmt.Errorf("decoding job extractor response: %w", err)
	}
	return posting, nil
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
		var readErr *postingReadError
		if errors.As(err, &readErr) {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": readErr.detail})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "failed to read the page"})
		return
	}
	ctx.JSON(http.StatusOK, posting)
}

// captureInput is the body of POST /jobs/capture: the popup's fields.
type captureInput struct {
	JobTitle    string `json:"jobTitle"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	Notes       string `json:"notes"`
	Description string `json:"description"`
}

func nonEmpty(p *string) bool { return p != nil && strings.TrimSpace(*p) != "" }

// mergeCapture folds a repeat capture into the application it matched and
// reports whether anything changed.
//
// It only ever adds: the status moves forwards or not at all (the same rule
// the email pipeline follows, so re-capturing a posting marked "saved" cannot
// pull an interviewing application back), and a field already holding a value
// is left alone, because the row may have been corrected by hand since. Title
// and company are never touched — they are the key the email pipeline matches
// on.
func mergeCapture(app *models.JobApplication, in captureInput, now time.Time) bool {
	changed := false

	if status := services.NormalizeStatus(in.Status); services.StatusAdvances(app.Status, status) {
		app.Status = status
		changed = true
	}
	// Saved-then-applied is the common path here, and the moment of the
	// second capture is the application date.
	if app.AppliedAt == nil && app.Status == "applied" {
		app.AppliedAt = &now
		changed = true
	}

	fill := func(field **string, value string) {
		if value != "" && !nonEmpty(*field) {
			*field = &value
			changed = true
		}
	}
	fill(&app.Location, in.Location)
	fill(&app.URL, in.URL)
	fill(&app.Description, in.Description)

	// Notes accumulate, as they do when the pipeline updates a row, but a
	// repeat capture of the same page must not append the same line twice.
	if in.Notes != "" {
		switch {
		case !nonEmpty(app.Notes):
			app.Notes = &in.Notes
			changed = true
		case !strings.Contains(*app.Notes, in.Notes):
			combined := *app.Notes + "\n" + in.Notes
			app.Notes = &combined
			changed = true
		}
	}
	return changed
}

// JobCapture backs POST /jobs/capture, an upsert.
//
// A posting already in the tracker is updated in place rather than added
// again. That matters beyond tidiness: the email pipeline identifies an
// application by (company, job title), so two rows for one posting means the
// confirmation email advances only one of them.
//
// Match order is the posting URL, then (company, job title) compared the way
// the pipeline compares them. Only live rows match — unlike the pipeline's
// lookup, which also sees deleted rows so that an email cannot resurrect one.
// A capture is a deliberate act, so capturing a posting whose row was deleted
// makes a new row.
func (store *Store) JobCapture(ctx *gin.Context) {
	var input captureInput
	if !bindCapture(ctx, &input) {
		return
	}
	input.JobTitle = strings.TrimSpace(input.JobTitle)
	input.Company = strings.TrimSpace(input.Company)
	input.Location = strings.TrimSpace(input.Location)
	input.Notes = strings.TrimSpace(input.Notes)
	input.Description = truncateRunes(strings.TrimSpace(input.Description), maxPageRunes)
	if input.JobTitle == "" || input.Company == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "jobTitle and company are required"})
		return
	}

	rawURL := input.URL
	if strings.TrimSpace(rawURL) != "" {
		canonical, ok := canonicalJobURL(rawURL)
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "url must be an absolute http(s) URL"})
			return
		}
		input.URL = canonical
	} else {
		input.URL = ""
	}

	var existing *models.JobApplication
	var err error
	if input.URL != "" {
		existing, err = store.findByURL(input.URL, rawURL)
	}
	if err == nil && existing == nil {
		existing, err = firstApplication(store.DB.Where(
			"LOWER(company) = LOWER(?) AND LOWER(job_title) = LOWER(?)", input.Company, input.JobTitle))
	}
	if err != nil {
		log.Println("Job capture lookup failed:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}

	now := time.Now()
	if existing != nil {
		action := "unchanged"
		if mergeCapture(existing, input, now) {
			if err := store.DB.Save(existing).Error; err != nil {
				log.Println("Job capture update failed:", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
				return
			}
			action = "updated"
		}
		ctx.JSON(http.StatusOK, gin.H{"action": action, "application": capturedJobOf(existing)})
		return
	}

	// A new row is the merge of the capture into an empty application, which
	// keeps the two paths agreeing on what an empty field and "applied" mean.
	// The empty status counts as unknown, so any status takes over from it.
	app := models.JobApplication{JobTitle: input.JobTitle, Company: input.Company}
	mergeCapture(&app, input, now)
	if err := store.DB.Create(&app).Error; err != nil {
		log.Println("Job capture insert failed:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"action": "created", "application": capturedJobOf(&app)})
}
