package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"adam-french.co.uk/backend/models"
	"gorm.io/gorm"
)

// Adding a posting to the job tracker without making a second row for one
// that is already there. Three callers share this: the browser capture
// endpoint (handlers/handle_job_capture.go), promoting a discovered lead
// (graph/job_lead.resolvers.go), and the lead fetcher (job_leads.go), which
// uses the URL matching to keep a posting already tracked out of the inbox.
// It lives in services because the fetcher does, and services cannot import
// handlers.

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

// CanonicalJobURL reduces a posting URL to a form that is the same however
// the page was reached, so it can serve as the key that decides whether a
// posting is already tracked. ok is false for anything that is not an
// absolute http(s) URL.
//
// LinkedIn gets its own rule because the same posting appears under several
// paths (search results, collections, the recommended feed), all carrying the
// id in ?currentJobId= rather than in the path.
func CanonicalJobURL(raw string) (canonical string, ok bool) {
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

// FindApplicationByURL returns the newest live application stored under this posting's
// URL, or nil. The raw form is tried as well as the canonical one because rows
// typed into the tracker by hand hold whatever was pasted.
func FindApplicationByURL(db *gorm.DB, canonical, raw string) (*models.JobApplication, error) {
	return firstApplication(db.Where("url IN ?", []string{canonical, strings.TrimSpace(raw)}))
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

// JobCaptureInput is a posting to add to the tracker. Its JSON form is the
// body of POST /jobs/capture: the popup's fields.
type JobCaptureInput struct {
	JobTitle    string `json:"jobTitle"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	Notes       string `json:"notes"`
	Description string `json:"description"`
}

func nonEmpty(p *string) bool { return p != nil && strings.TrimSpace(*p) != "" }

// MergeCapture folds a repeat capture into the application it matched and
// reports whether anything changed.
//
// It only ever adds: the status moves forwards or not at all (the same rule
// the email pipeline follows, so re-capturing a posting marked "saved" cannot
// pull an interviewing application back), and a field already holding a value
// is left alone, because the row may have been corrected by hand since. Title
// and company are never touched — they are the key the email pipeline matches
// on.
func MergeCapture(app *models.JobApplication, in JobCaptureInput, now time.Time) bool {
	changed := false

	if status := NormalizeStatus(in.Status); StatusAdvances(app.Status, status) {
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

// JobInputError is a problem with the input itself (a missing title, a URL
// that is not one), worded for the person who sent it. Anything else
// UpsertJobApplication returns is a database failure and belongs in the log.
type JobInputError struct{ msg string }

func (e *JobInputError) Error() string { return e.msg }

// UpsertJobApplication adds a posting to the tracker, or folds it into the
// application it matches, and reports which: "created", "updated" or
// "unchanged".
//
// The upsert matters beyond tidiness: the email pipeline identifies an
// application by (company, job title), so two rows for one posting means the
// confirmation email advances only one of them.
//
// Match order is the posting URL, then (company, job title) compared the way
// the pipeline compares them. Only live rows match — unlike the pipeline's
// lookup, which also sees deleted rows so that an email cannot resurrect one.
// Capturing or promoting is a deliberate act, so doing it for a posting whose
// row was deleted makes a new row.
//
// The caller bounds Description's length; everything else is trimmed here.
func UpsertJobApplication(db *gorm.DB, input JobCaptureInput, now time.Time) (*models.JobApplication, string, error) {
	input.JobTitle = strings.TrimSpace(input.JobTitle)
	input.Company = strings.TrimSpace(input.Company)
	input.Location = strings.TrimSpace(input.Location)
	input.Notes = strings.TrimSpace(input.Notes)
	input.Description = strings.TrimSpace(input.Description)
	if input.JobTitle == "" || input.Company == "" {
		return nil, "", &JobInputError{"jobTitle and company are required"}
	}

	rawURL := input.URL
	if strings.TrimSpace(rawURL) != "" {
		canonical, ok := CanonicalJobURL(rawURL)
		if !ok {
			return nil, "", &JobInputError{"url must be an absolute http(s) URL"}
		}
		input.URL = canonical
	} else {
		input.URL = ""
	}

	var existing *models.JobApplication
	var err error
	if input.URL != "" {
		existing, err = FindApplicationByURL(db, input.URL, rawURL)
	}
	if err == nil && existing == nil {
		existing, err = firstApplication(db.Where(
			"LOWER(company) = LOWER(?) AND LOWER(job_title) = LOWER(?)", input.Company, input.JobTitle))
	}
	if err != nil {
		return nil, "", fmt.Errorf("looking up the application: %w", err)
	}

	if existing != nil {
		if !MergeCapture(existing, input, now) {
			return existing, "unchanged", nil
		}
		if err := db.Save(existing).Error; err != nil {
			return nil, "", fmt.Errorf("updating the application: %w", err)
		}
		return existing, "updated", nil
	}

	// A new row is the merge of the input into an empty application, which
	// keeps the two paths agreeing on what an empty field and "applied" mean.
	// The empty status counts as unknown, so any status takes over from it.
	app := models.JobApplication{JobTitle: input.JobTitle, Company: input.Company}
	MergeCapture(&app, input, now)
	if err := db.Create(&app).Error; err != nil {
		return nil, "", fmt.Errorf("creating the application: %w", err)
	}
	return &app, "created", nil
}
