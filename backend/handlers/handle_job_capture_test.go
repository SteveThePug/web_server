package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
)

// The capture handlers themselves need a database, and the only driver in
// go.mod is postgres, so what is covered here is everything around the
// queries: the token gate, the URL key that decides whether a posting is
// already tracked, the merge rule for a repeat capture, and the call to the
// Python service.

func TestCaptureTokenMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name       string
		configured string
		header     string
		want       int
	}{
		{"the right token is let through", "s3cret", "Bearer s3cret", http.StatusOK},
		{"a wrong token is refused", "s3cret", "Bearer nope", http.StatusUnauthorized},
		{"no header is refused", "s3cret", "", http.StatusUnauthorized},
		{"the bare token without the scheme is refused", "s3cret", "s3cret", http.StatusUnauthorized},
		// The case the middleware exists to get right: an unset token must
		// not match the empty credential of an anonymous request.
		{"an unset token switches the routes off", "", "", http.StatusServiceUnavailable},
		{"an unset token is not matched by an empty bearer", "", "Bearer ", http.StatusServiceUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := &Store{JobCaptureToken: c.configured}
			r := gin.New()
			r.GET("/x", store.CaptureTokenMiddleware, func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			if c.header != "" {
				req.Header.Set("Authorization", c.header)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d", rec.Code, c.want)
			}
		})
	}
}

func TestCanonicalJobURL(t *testing.T) {
	const linkedIn = "https://www.linkedin.com/jobs/view/4012345678/"
	cases := []struct {
		name string
		in   string
		want string // "" means rejected
	}{
		// One LinkedIn posting, reached four ways, must be one key.
		{"linkedin direct", "https://www.linkedin.com/jobs/view/4012345678/?trk=public_jobs", linkedIn},
		{"linkedin slug form", "https://uk.linkedin.com/jobs/view/backend-engineer-at-acme-4012345678?refId=abc", linkedIn},
		{"linkedin search pane", "https://www.linkedin.com/jobs/search/?currentJobId=4012345678&keywords=go", linkedIn},
		{"linkedin collections pane", "https://www.linkedin.com/jobs/collections/recommended/?currentJobId=4012345678", linkedIn},
		{"linkedin page with no job id is left as a normal URL", "https://www.linkedin.com/jobs/search/?keywords=go", "https://www.linkedin.com/jobs/search?keywords=go"},
		{"a non-numeric linkedin id is not trusted", "https://www.linkedin.com/jobs/search/?currentJobId=4/../x", "https://www.linkedin.com/jobs/search?currentJobId=4%2F..%2Fx"},
		{"a lookalike host does not get the linkedin rule", "https://notlinkedin.com/jobs/view/42", "https://notlinkedin.com/jobs/view/42"},

		{"tracking parameters and the fragment are dropped", "https://jobs.lever.co/acme/abc-123?lever-source=LinkedIn&utm_medium=x#apply", "https://jobs.lever.co/acme/abc-123"},
		// gh_jid IS the posting on an embedded Greenhouse board.
		{"a parameter that identifies the job is kept", "https://acme.com/careers?gh_jid=77&gh_src=abc", "https://acme.com/careers?gh_jid=77"},
		{"parameter order does not matter", "https://x.com/job?b=2&a=1", "https://x.com/job?a=1&b=2"},
		{"host case and a trailing slash do not matter", "https://Boards.Greenhouse.io/acme/jobs/55/", "https://boards.greenhouse.io/acme/jobs/55"},

		{"a relative path is rejected", "/jobs/view/1", ""},
		{"a non-http scheme is rejected", "javascript:alert(1)", ""},
		{"empty is rejected", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := canonicalJobURL(c.in)
			if ok != (c.want != "") || got != c.want {
				t.Fatalf("canonicalJobURL(%q) = %q, %v; want %q", c.in, got, ok, c.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

func TestMergeCapture(t *testing.T) {
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	earlier := now.Add(-48 * time.Hour)

	t.Run("a new row takes every field and an applied date", func(t *testing.T) {
		app := models.JobApplication{JobTitle: "Engineer", Company: "Acme"}
		in := captureInput{Status: "applied", Location: "London", URL: "https://x.com/1", Notes: "Hybrid", Description: "Full text"}
		if !mergeCapture(&app, in, now) {
			t.Fatal("expected a change")
		}
		if app.Status != "applied" || app.AppliedAt == nil || !app.AppliedAt.Equal(now) {
			t.Fatalf("status = %q, appliedAt = %v", app.Status, app.AppliedAt)
		}
		if *app.Location != "London" || *app.URL != "https://x.com/1" || *app.Notes != "Hybrid" || *app.Description != "Full text" {
			t.Fatalf("fields not filled: %+v", app)
		}
	})

	t.Run("a saved posting has no applied date", func(t *testing.T) {
		app := models.JobApplication{}
		mergeCapture(&app, captureInput{Status: "saved"}, now)
		if app.Status != "saved" || app.AppliedAt != nil {
			t.Fatalf("status = %q, appliedAt = %v", app.Status, app.AppliedAt)
		}
	})

	t.Run("saved then applied advances and dates the application", func(t *testing.T) {
		app := models.JobApplication{Status: "saved"}
		if !mergeCapture(&app, captureInput{Status: "applied"}, now) {
			t.Fatal("expected a change")
		}
		if app.Status != "applied" || app.AppliedAt == nil || !app.AppliedAt.Equal(now) {
			t.Fatalf("status = %q, appliedAt = %v", app.Status, app.AppliedAt)
		}
	})

	t.Run("a repeat capture cannot move an application backwards", func(t *testing.T) {
		app := models.JobApplication{Status: "interviewing", AppliedAt: &earlier}
		if mergeCapture(&app, captureInput{Status: "saved"}, now) {
			t.Fatal("expected no change")
		}
		if app.Status != "interviewing" || !app.AppliedAt.Equal(earlier) {
			t.Fatalf("status = %q, appliedAt = %v", app.Status, app.AppliedAt)
		}
	})

	t.Run("fields already set are kept, empty ones are filled", func(t *testing.T) {
		app := models.JobApplication{Status: "applied", AppliedAt: &earlier, Location: strPtr("Remote"), URL: strPtr("")}
		in := captureInput{Status: "applied", Location: "London", URL: "https://x.com/1", Description: "Full text"}
		if !mergeCapture(&app, in, now) {
			t.Fatal("expected a change")
		}
		if *app.Location != "Remote" {
			t.Fatalf("location overwritten: %q", *app.Location)
		}
		if *app.URL != "https://x.com/1" || *app.Description != "Full text" {
			t.Fatalf("empty fields not filled: url=%q", *app.URL)
		}
		if !app.AppliedAt.Equal(earlier) {
			t.Fatal("an existing applied date must not be replaced")
		}
	})

	t.Run("notes append once", func(t *testing.T) {
		app := models.JobApplication{Status: "applied", AppliedAt: &earlier, Notes: strPtr("Recruiter: Sam")}
		in := captureInput{Status: "applied", Notes: "Hybrid, £70k"}
		if !mergeCapture(&app, in, now) || *app.Notes != "Recruiter: Sam\nHybrid, £70k" {
			t.Fatalf("notes = %q", *app.Notes)
		}
		if mergeCapture(&app, in, now) {
			t.Fatalf("the same note was appended twice: %q", *app.Notes)
		}
	})
}

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("££££", 2); got != "££" {
		t.Fatalf("got %q; the cut must fall on a rune boundary", got)
	}
	if got := truncateRunes("short", 10); got != "short" {
		t.Fatalf("got %q", got)
	}
}

func jobExtractor(t *testing.T, handler http.HandlerFunc) *Store {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &Store{PythonURL: srv.URL}
}

func TestExtractJobPostingRequestAndReply(t *testing.T) {
	var got map[string]string
	var path string
	var forwarded bool
	store := jobExtractor(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		// As with the rowing reader, either header makes the Python route
		// answer 404.
		forwarded = r.Header.Get("X-Real-IP") != "" || r.Header.Get("X-Forwarded-For") != ""
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		w.Write([]byte(`{"jobTitle": "Engineer", "company": "Acme", "location": "London", "notes": null}`))
	})

	posting, err := store.extractJobPosting(context.Background(), "https://x.com/1", "Engineer | Acme", "page text")
	if err != nil {
		t.Fatalf("extractJobPosting: %v", err)
	}
	if posting.JobTitle != "Engineer" || posting.Company != "Acme" || posting.Location == nil || *posting.Location != "London" || posting.Notes != nil {
		t.Fatalf("got %+v", posting)
	}
	if path != jobExtractPath {
		t.Fatalf("path = %q, want %q", path, jobExtractPath)
	}
	if forwarded {
		t.Fatal("request carried a proxy header the Python route rejects")
	}
	if got["url"] != "https://x.com/1" || got["title"] != "Engineer | Acme" || got["text"] != "page text" {
		t.Fatalf("request body = %+v", got)
	}
}

func TestExtractJobPostingErrors(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		body       string
		wantDetail string // non-empty: a postingReadError with this message
	}{
		{"extractor failure is relayed", http.StatusBadGateway, `{"detail": "failed to parse the posting"}`, "failed to parse the posting"},
		{"validation error is not relayed", http.StatusUnprocessableEntity, `{"detail": [{"msg": "bad"}]}`, ""},
		{"missing key is not relayed", http.StatusServiceUnavailable, `{"detail": "CLAUDE_API_KEY is not set"}`, ""},
		{"malformed success body", http.StatusOK, `not json`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := jobExtractor(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				w.Write([]byte(c.body))
			})
			_, err := store.extractJobPosting(context.Background(), "https://x.com/1", "t", "text")
			if err == nil {
				t.Fatal("expected an error")
			}
			var readErr *postingReadError
			if relayed := errors.As(err, &readErr); relayed != (c.wantDetail != "") {
				t.Fatalf("relayed = %v, err = %v", relayed, err)
			}
			if readErr != nil && readErr.detail != c.wantDetail {
				t.Fatalf("detail = %q, want %q", readErr.detail, c.wantDetail)
			}
		})
	}
}

// The extract handler runs without a database, so it can be driven whole: the
// page text must be cut to the limit the Python side enforces before it is
// forwarded, or a long page would come back as an opaque 422.
func TestJobCaptureExtractTruncatesAndRelays(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var forwardedLen int
	store := jobExtractor(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		forwardedLen = len([]rune(body["text"]))
		w.Write([]byte(`{"jobTitle": "Engineer", "company": "Acme", "location": null, "notes": null}`))
	})
	r := gin.New()
	r.POST("/extract", store.JobCaptureExtract)

	post := func(body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/extract", strings.NewReader(body)))
		return rec
	}

	long, _ := json.Marshal(map[string]string{"url": "https://x.com/1", "title": "t", "text": strings.Repeat("é", maxPageRunes+500)})
	rec := post(string(long))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if forwardedLen != maxPageRunes {
		t.Fatalf("forwarded %d runes, want %d", forwardedLen, maxPageRunes)
	}
	var out ExtractedJobPosting
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.Company != "Acme" {
		t.Fatalf("body = %s (%v)", rec.Body, err)
	}

	if rec := post(`{"url": "https://x.com/1", "text": "   "}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("blank text: status = %d", rec.Code)
	}
	if rec := post(`not json`); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad body: status = %d", rec.Code)
	}
}
