package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"adam-french.co.uk/backend/models"
	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

// The capture handlers themselves need a database, and the only driver in
// go.mod is postgres, so what is covered here is everything around the
// queries: the token gate, the URL key that decides whether a posting is
// already tracked, the merge rule for a repeat capture, and the call to the
// Python service.

func TestCaptureTokenMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serve := func(store *Store, header string) int {
		r := gin.New()
		r.GET("/x", store.CaptureTokenMiddleware, func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}
	newToken := func() *services.CaptureToken {
		tok, err := services.InitCaptureToken(filepath.Join(t.TempDir(), "token.json"))
		if err != nil {
			t.Fatalf("InitCaptureToken: %v", err)
		}
		return tok
	}

	// The case the middleware exists to get right: with no token generated,
	// a request presenting nothing (or an empty bearer) must not get in.
	t.Run("no token generated switches the routes off", func(t *testing.T) {
		store := &Store{CaptureToken: newToken()}
		for _, header := range []string{"", "Bearer ", "Bearer anything"} {
			if got := serve(store, header); got != http.StatusServiceUnavailable {
				t.Fatalf("header %q: status = %d, want 503", header, got)
			}
		}
	})
	t.Run("a store built without the token service is off too", func(t *testing.T) {
		if got := serve(&Store{}, "Bearer x"); got != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", got)
		}
	})

	store := &Store{CaptureToken: newToken()}
	token, _, err := store.CaptureToken.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"the generated token is let through", "Bearer " + token, http.StatusOK},
		{"a wrong token is refused", "Bearer nope", http.StatusUnauthorized},
		{"no header is refused", "", http.StatusUnauthorized},
		{"the bare token without the scheme is refused", token, http.StatusUnauthorized},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := serve(store, c.header); got != c.want {
				t.Fatalf("status = %d, want %d", got, c.want)
			}
		})
	}
}

// The admin endpoints, driven without the admin middleware (covered
// elsewhere): generate shows the token once, status never does, and a second
// generate or a revoke stops the first token working.
func TestCaptureTokenAdminEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tok, err := services.InitCaptureToken(filepath.Join(t.TempDir(), "token.json"))
	if err != nil {
		t.Fatalf("InitCaptureToken: %v", err)
	}
	store := &Store{CaptureToken: tok}
	r := gin.New()
	r.GET("/t", store.GetCaptureToken)
	r.POST("/t", store.GenerateCaptureToken)
	r.DELETE("/t", store.RevokeCaptureToken)

	call := func(method string) (map[string]any, *httptest.ResponseRecorder) {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(method, "/t", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, body = %s", method, rec.Code, rec.Body)
		}
		var out map[string]any
		json.Unmarshal(rec.Body.Bytes(), &out)
		return out, rec
	}

	if out, _ := call(http.MethodGet); out["set"] != false {
		t.Fatalf("fresh status = %v", out)
	}

	out, rec := call(http.MethodPost)
	first, _ := out["token"].(string)
	if !tok.Verify(first) {
		t.Fatalf("generated token %q does not verify", first)
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("the response carrying the token must not be cacheable")
	}

	out, _ = call(http.MethodGet)
	if out["set"] != true || out["createdAt"] == nil {
		t.Fatalf("status after generate = %v", out)
	}
	if _, leaked := out["token"]; leaked {
		t.Fatal("status must never return the token")
	}

	call(http.MethodPost)
	if tok.Verify(first) {
		t.Fatal("regenerating must invalidate the previous token")
	}

	if out, _ := call(http.MethodDelete); out["set"] != false {
		t.Fatalf("status after revoke = %v", out)
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
