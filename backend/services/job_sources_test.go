package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"adam-french.co.uk/backend/models"
)

// The source fetchers, each against a stub server replying with a cut-down
// copy of a real response (the board kinds) or of the documented shape (the
// search kinds, which need API keys). What is under test is the request each
// one makes and how its JSON becomes a fetchedJob.

// sourceStub points every source's base URL at one stub server for the test.
func sourceStub(t *testing.T, config JobLeadConfig, handler http.HandlerFunc) *JobLeadService {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	for _, base := range []*string{&greenhouseBase, &leverBase, &ashbyBase, &adzunaBase, &reedBase} {
		old := *base
		*base = srv.URL
		t.Cleanup(func() { *base = old })
	}
	return &JobLeadService{Config: &config, HTTPClient: srv.Client()}
}

func TestFetchGreenhouse(t *testing.T) {
	var path string
	s := sourceStub(t, JobLeadConfig{}, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.RequestURI()
		w.Write([]byte(`{"jobs": [{
			"absolute_url": "https://job-boards.greenhouse.io/monzo/jobs/8143930",
			"location": {"name": "Cardiff, London or Remote (UK)"},
			"id": 8143930,
			"updated_at": "2026-09-25T05:17:09-04:00",
			"title": "Backend Engineer",
			"first_published": "2026-08-24T09:04:42-04:00",
			"content": "&lt;div class=&quot;intro&quot;&gt;&lt;p&gt;&lt;strong&gt;About us&lt;/strong&gt;&lt;/p&gt;&lt;ul&gt;&lt;li&gt;Go &amp;amp; Kotlin&lt;/li&gt;&lt;li&gt;On-call&lt;/li&gt;&lt;/ul&gt;&lt;/div&gt;"
		}], "meta": {"total": 1}}`))
	})

	jobs, err := s.fetchGreenhouse(context.Background(), models.JobSource{Kind: SourceGreenhouse, Label: "Monzo", Token: "Monzo"})
	if err != nil {
		t.Fatalf("fetchGreenhouse: %v", err)
	}
	if path != "/v1/boards/Monzo/jobs?content=true" {
		t.Errorf("request = %q", path)
	}
	if len(jobs) != 1 {
		t.Fatalf("got %d jobs", len(jobs))
	}
	j := jobs[0]
	// The id is namespaced by the board, lowercased so "Monzo" and "monzo"
	// are one board.
	if j.ExternalID != "monzo/8143930" || j.Title != "Backend Engineer" || j.Company != "Monzo" ||
		j.Location != "Cardiff, London or Remote (UK)" || j.URL != "https://job-boards.greenhouse.io/monzo/jobs/8143930" {
		t.Errorf("job = %+v", j)
	}
	// Doubly escaped HTML comes out as plain lines.
	if j.Description != "About us\n\nGo & Kotlin\n\nOn-call" {
		t.Errorf("description = %q", j.Description)
	}
	if j.PostedAt == nil || j.PostedAt.UTC().Format("2006-01-02") != "2026-08-24" {
		t.Errorf("postedAt = %v", j.PostedAt)
	}
}

func TestFetchLever(t *testing.T) {
	s := sourceStub(t, JobLeadConfig{}, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/v0/postings/acme?mode=json" {
			t.Errorf("request = %q", r.URL.RequestURI())
		}
		w.Write([]byte(`[{
			"additionalPlain": "We are an equal opportunity employer.",
			"categories": {"commitment": "Full-time", "location": "London, UK", "team": "Platform"},
			"createdAt": 1786469891368,
			"descriptionPlain": "A World-Changing Company\n",
			"id": "6ed76ce8-4156",
			"lists": [{"text": "What you'll do", "content": "<li>Build things</li><li>Run them</li>"}],
			"text": "Platform Engineer",
			"workplaceType": "hybrid",
			"hostedUrl": "https://jobs.lever.co/acme/6ed76ce8-4156"
		}]`))
	})

	jobs, err := s.fetchLever(context.Background(), models.JobSource{Kind: SourceLever, Label: "Acme", Token: "acme"})
	if err != nil || len(jobs) != 1 {
		t.Fatalf("fetchLever: %v, %d jobs", err, len(jobs))
	}
	j := jobs[0]
	if j.ExternalID != "acme/6ed76ce8-4156" || j.Title != "Platform Engineer" || j.Location != "London, UK (hybrid)" {
		t.Errorf("job = %+v", j)
	}
	want := "A World-Changing Company\n\nWhat you'll do\nBuild things\n\nRun them\n\nWe are an equal opportunity employer."
	if j.Description != want {
		t.Errorf("description = %q\nwant %q", j.Description, want)
	}
	if j.PostedAt == nil || j.PostedAt.Year() != 2026 {
		t.Errorf("postedAt = %v", j.PostedAt)
	}
}

func TestFetchAshbySkipsUnlisted(t *testing.T) {
	s := sourceStub(t, JobLeadConfig{}, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/posting-api/job-board/ramp" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Write([]byte(`{"apiVersion": "1", "jobs": [
			{"id": "34413f8d", "title": " Security Engineer, Cloud", "location": "London", "workplaceType": "Remote",
			 "publishedAt": "2026-04-07T17:12:35.753+00:00", "isListed": true,
			 "jobUrl": "https://jobs.ashbyhq.com/ramp/34413f8d", "descriptionPlain": "ABOUT RAMP\n\nRamp is building."},
			{"id": "hidden", "title": "Secret", "isListed": false, "jobUrl": "https://jobs.ashbyhq.com/ramp/hidden"}
		]}`))
	})

	jobs, err := s.fetchAshby(context.Background(), models.JobSource{Kind: SourceAshby, Label: "Ramp", Token: "ramp"})
	if err != nil || len(jobs) != 1 {
		t.Fatalf("fetchAshby: %v, %d jobs", err, len(jobs))
	}
	j := jobs[0]
	if j.ExternalID != "ramp/34413f8d" || j.Title != "Security Engineer, Cloud" || j.Location != "London (remote)" ||
		j.Description != "ABOUT RAMP\n\nRamp is building." || j.PostedAt == nil {
		t.Errorf("job = %+v", j)
	}
}

func TestFetchAdzuna(t *testing.T) {
	var requests []string
	s := sourceStub(t, JobLeadConfig{AdzunaAppID: "id1", AdzunaAppKey: "key1"}, func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		q := r.URL.Query()
		if q.Get("app_id") != "id1" || q.Get("app_key") != "key1" || q.Get("what") != "golang" || q.Get("where") != "London" {
			t.Errorf("query = %v", q)
		}
		w.Write([]byte(`{"results": [{
			"id": "5012345678", "title": "<strong>Golang</strong> Developer",
			"description": "We need a <strong>Golang</strong> developer…",
			"redirect_url": "https://www.adzuna.co.uk/jobs/land/ad/5012345678?se=abc",
			"created": "2026-10-01T09:00:00Z",
			"company": {"display_name": "Acme Ltd"}, "location": {"display_name": "London, UK"}
		}]}`))
	})

	jobs, err := s.fetchAdzuna(context.Background(), models.JobSource{Kind: SourceAdzuna, Query: "golang", Location: "London"})
	if err != nil || len(jobs) != 1 {
		t.Fatalf("fetchAdzuna: %v, %d jobs", err, len(jobs))
	}
	// A short page is the last page: no second request.
	if len(requests) != 1 || requests[0] != "/v1/api/jobs/gb/search/1" {
		t.Errorf("requests = %v", requests)
	}
	j := jobs[0]
	if j.ExternalID != "5012345678" || j.Title != "Golang Developer" || j.Company != "Acme Ltd" ||
		j.Description != "We need a Golang developer…" || j.detailID != "" {
		t.Errorf("job = %+v", j)
	}
}

func TestFetchReed(t *testing.T) {
	s := sourceStub(t, JobLeadConfig{ReedAPIKey: "reedkey"}, func(w http.ResponseWriter, r *http.Request) {
		// The key is the basic-auth username, with an empty password.
		if user, pass, ok := r.BasicAuth(); !ok || user != "reedkey" || pass != "" {
			t.Errorf("basic auth = %q %q %v", user, pass, ok)
		}
		switch r.URL.Path {
		case "/api/1.0/search":
			if q := r.URL.Query(); q.Get("keywords") != "golang" || q.Get("locationName") != "London" {
				t.Errorf("query = %v", q)
			}
			w.Write([]byte(`{"results": [{"jobId": 51234567, "employerName": "Acme", "jobTitle": "Go Engineer",
				"locationName": "London", "date": "03/10/2026", "jobDescription": "Snippet…",
				"jobUrl": "https://www.reed.co.uk/jobs/go-engineer/51234567"}], "totalResults": 1}`))
		case "/api/1.0/jobs/51234567":
			w.Write([]byte(`{"jobId": 51234567, "jobDescription": "<p>The full advert.</p><p>Second paragraph.</p>"}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})

	jobs, err := s.fetchReed(context.Background(), models.JobSource{Kind: SourceReed, Query: "golang", Location: "London"})
	if err != nil || len(jobs) != 1 {
		t.Fatalf("fetchReed: %v, %d jobs", err, len(jobs))
	}
	j := jobs[0]
	if j.ExternalID != "51234567" || j.Company != "Acme" || j.Description != "Snippet…" || j.detailID != "51234567" {
		t.Errorf("job = %+v", j)
	}
	if j.PostedAt == nil || j.PostedAt.Format("2006-01-02") != "2026-10-03" {
		t.Errorf("postedAt = %v", j.PostedAt)
	}

	full, err := s.fetchReedDescription(context.Background(), j.detailID)
	if err != nil || full != "The full advert.\n\nSecond paragraph." {
		t.Errorf("fetchReedDescription = %q, %v", full, err)
	}
}

func TestSearchSourcesNeedTheirKeys(t *testing.T) {
	s := sourceStub(t, JobLeadConfig{}, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no request should be made without a key, got %s", r.URL.Path)
	})
	if _, err := s.fetchSource(context.Background(), models.JobSource{Kind: SourceAdzuna, Query: "go"}); err == nil || !strings.Contains(err.Error(), "ADZUNA_APP_ID") {
		t.Errorf("adzuna err = %v", err)
	}
	if _, err := s.fetchSource(context.Background(), models.JobSource{Kind: SourceReed, Query: "go"}); err == nil || !strings.Contains(err.Error(), "REED_API_KEY") {
		t.Errorf("reed err = %v", err)
	}
	if _, err := s.fetchSource(context.Background(), models.JobSource{Kind: "monster"}); err == nil {
		t.Error("an unknown kind should be an error")
	}
}

// The error ends up in JobSource.LastError, on the page. It must say what to
// fix and must never quote the request URL, which for Adzuna holds the key.
func TestSourceErrorsAreSafeToShow(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{http.StatusNotFound, "check the board token"},
		{http.StatusUnauthorized, "check the API key"},
		{http.StatusInternalServerError, "source returned 500"},
	}
	for _, c := range cases {
		s := sourceStub(t, JobLeadConfig{AdzunaAppID: "id1", AdzunaAppKey: "sekrit"}, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.status)
		})
		_, err := s.fetchSource(context.Background(), models.JobSource{Kind: SourceAdzuna, Query: "go"})
		if err == nil || !strings.Contains(err.Error(), c.want) || strings.Contains(err.Error(), "sekrit") {
			t.Errorf("status %d: err = %v", c.status, err)
		}
	}

	// A connection failure is the case where net/http would quote the URL.
	s := sourceStub(t, JobLeadConfig{AdzunaAppID: "id1", AdzunaAppKey: "sekrit"}, func(http.ResponseWriter, *http.Request) {})
	adzunaBase = "http://127.0.0.1:1"
	_, err := s.fetchSource(context.Background(), models.JobSource{Kind: SourceAdzuna, Query: "go"})
	if err == nil || strings.Contains(err.Error(), "sekrit") {
		t.Errorf("connection failure: err = %v", err)
	}

	s = sourceStub(t, JobLeadConfig{}, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("<html>")) })
	if _, err := s.fetchSource(context.Background(), models.JobSource{Kind: SourceLever, Token: "acme"}); err == nil || !strings.Contains(err.Error(), "unexpected response") {
		t.Errorf("non-JSON reply: err = %v", err)
	}
}

func TestHTMLToText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"<p>One</p><p>Two</p>", "One\n\nTwo"},
		{"<ul><li>a</li><li>b &amp; c</li></ul>", "a\n\nb & c"},
		{"plain   text here", "plain text here"},
		{"<span>inline</span> <b>tags</b> stay on a line", "inline tags stay on a line"},
		{"line<br>break<br/>twice", "line\nbreak\ntwice"},
		{"", ""},
	}
	for _, c := range cases {
		if got := htmlToText(c.in); got != c.want {
			t.Errorf("htmlToText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestWithWorkplace(t *testing.T) {
	cases := []struct{ location, workplace, want string }{
		{"London", "Hybrid", "London (hybrid)"},
		{"Remote - UK", "remote", "Remote - UK"},
		{"", "Remote", "remote"},
		{"London", "", "London"},
		{"London", "unspecified", "London"},
	}
	for _, c := range cases {
		if got := withWorkplace(c.location, c.workplace); got != c.want {
			t.Errorf("withWorkplace(%q, %q) = %q, want %q", c.location, c.workplace, got, c.want)
		}
	}
}
