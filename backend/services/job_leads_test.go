package services

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
)

// The parts of the lead pipeline that need no database: the pre-filter, how a
// fetched posting becomes a row, source validation, the note a promoted lead
// leaves, and the scoring call. Storing, re-filtering and the scoring loop's
// bookkeeping all run queries and are exercised in the dev stack instead.

func TestLeadFilter(t *testing.T) {
	filter := newLeadFilter(&models.JobSearchProfile{
		IncludeKeywords: "Engineer\ndeveloper, SRE",
		ExcludeKeywords: "intern\nsenior staff\nsales",
		Locations:       "London\nremote",
	})
	cases := []struct {
		title, location string
		want            bool
	}{
		{"Backend Engineer", "London, UK", true},
		{"Software Engineers (Platform)", "Remote - EMEA", true},
		{"SRE", "Cardiff, London or Remote (UK)", true},
		{"Engineering Manager", "London", false}, // "engineer" is not a whole word here
		{"Account Executive", "London", false},
		{"Engineer Intern", "London", false},
		{"Internal Tools Engineer", "London", true}, // "intern" must not catch "Internal"
		{"Senior Staff Engineer", "London", false},
		{"Sales Engineer", "London", false},
		{"Backend Engineer", "New York, NY", false},
		{"Backend Engineer", "", true}, // unknown location passes
	}
	for _, c := range cases {
		if got := filter.matches(c.title, c.location); got != c.want {
			t.Errorf("matches(%q, %q) = %v, want %v", c.title, c.location, got, c.want)
		}
	}

	// Nothing configured: everything passes.
	if !newLeadFilter(&models.JobSearchProfile{}).matches("Chief Vibes Officer", "Mars") {
		t.Error("an empty filter should match everything")
	}
}

func TestContainsWord(t *testing.T) {
	cases := []struct {
		text, word string
		want       bool
	}{
		{"go developer", "go", true},
		{"golang developer", "go", false},
		{"c++ engineer", "c++", true},
		{"developers wanted", "developer", true},
		{"developerss", "developer", false},
		{"site reliability engineer (sre)", "sre", true},
		{"ingénieur logiciel", "nieur", false},
		{"", "go", false},
	}
	for _, c := range cases {
		if got := containsWord(c.text, c.word); got != c.want {
			t.Errorf("containsWord(%q, %q) = %v, want %v", c.text, c.word, got, c.want)
		}
	}
}

func TestNewLead(t *testing.T) {
	now := time.Date(2026, 10, 12, 9, 0, 0, 0, time.UTC)
	src := models.JobSource{ID: 7, Kind: SourceGreenhouse, Label: "Acme", Token: "acme"}
	filter := newLeadFilter(&models.JobSearchProfile{IncludeKeywords: "engineer"})
	tracked := map[string]uint{"https://boards.greenhouse.io/acme/jobs/2": 42}

	t.Run("a fresh posting is a new, matched lead", func(t *testing.T) {
		lead, ok := newLead(src, fetchedJob{
			ExternalID: "acme/1", Title: " Backend Engineer ", Company: "Acme", Location: "London",
			URL: "https://boards.greenhouse.io/acme/jobs/1/?gh_src=abc", Description: "Text",
		}, filter, tracked, now)
		if !ok {
			t.Fatal("not ok")
		}
		if lead.State != LeadNew || !lead.Matched || lead.JobAppID != nil || lead.SourceID != 7 ||
			lead.SourceKind != SourceGreenhouse || lead.JobTitle != "Backend Engineer" || !lead.LastSeenAt.Equal(now) {
			t.Errorf("lead = %+v", lead)
		}
		// Stored canonical, so it is the same key the tracker would hold.
		if lead.URL != "https://boards.greenhouse.io/acme/jobs/1" {
			t.Errorf("url = %q", lead.URL)
		}
	})

	t.Run("a posting already in the tracker is filed as promoted", func(t *testing.T) {
		lead, ok := newLead(src, fetchedJob{
			ExternalID: "acme/2", Title: "Backend Engineer", Company: "Acme",
			URL: "https://boards.greenhouse.io/acme/jobs/2?utm_source=x",
		}, filter, tracked, now)
		if !ok || lead.State != LeadPromoted || lead.JobAppID == nil || *lead.JobAppID != 42 {
			t.Errorf("lead = %+v, ok = %v", lead, ok)
		}
	})

	t.Run("a posting that fails the filter is kept, unmatched", func(t *testing.T) {
		lead, ok := newLead(src, fetchedJob{ExternalID: "acme/3", Title: "Account Executive", Company: "Acme"}, filter, tracked, now)
		if !ok || lead.Matched || lead.State != LeadNew {
			t.Errorf("lead = %+v, ok = %v", lead, ok)
		}
	})

	t.Run("a posting with no id, title or company is dropped", func(t *testing.T) {
		for _, job := range []fetchedJob{
			{Title: "Engineer", Company: "Acme"},
			{ExternalID: "acme/4", Company: "Acme"},
			{ExternalID: "acme/5", Title: "Engineer", Company: "  "},
		} {
			if _, ok := newLead(src, job, filter, tracked, now); ok {
				t.Errorf("%+v should be dropped", job)
			}
		}
	})

	t.Run("a long description is cut on a rune boundary", func(t *testing.T) {
		lead, _ := newLead(src, fetchedJob{
			ExternalID: "acme/6", Title: "Engineer", Company: "Acme", Description: strings.Repeat("é", maxLeadRunes+10),
		}, filter, tracked, now)
		if got := len([]rune(lead.Description)); got != maxLeadRunes {
			t.Errorf("description is %d runes", got)
		}
	})
}

func TestValidateJobSource(t *testing.T) {
	ok := []struct {
		in   models.JobSource
		want models.JobSource
	}{
		// Fields the kind does not use are cleared.
		{models.JobSource{Kind: " Greenhouse ", Label: " Acme ", Token: " acme ", Query: "stale"},
			models.JobSource{Kind: "greenhouse", Label: "Acme", Token: "acme"}},
		// A search's label defaults to its query.
		{models.JobSource{Kind: "reed", Query: "golang", Location: "London", Token: "stale"},
			models.JobSource{Kind: "reed", Label: "golang", Query: "golang", Location: "London"}},
	}
	for _, c := range ok {
		src := c.in
		if err := ValidateJobSource(&src); err != nil || src != c.want {
			t.Errorf("ValidateJobSource(%+v) = %+v, %v; want %+v", c.in, src, err, c.want)
		}
	}

	bad := []models.JobSource{
		{Kind: "monster", Label: "x", Query: "go"},
		{Kind: "lever", Label: "Acme"},
		{Kind: "lever", Label: "Acme", Token: "https://jobs.lever.co/acme"},
		{Kind: "ashby", Token: "acme"},
		{Kind: "adzuna", Label: "x"},
	}
	for _, src := range bad {
		if err := ValidateJobSource(&src); err == nil {
			t.Errorf("ValidateJobSource(%+v) should fail", src)
		}
	}
}

func TestLeadNotes(t *testing.T) {
	score := 82
	reasons := " Strong Go match; hybrid London. "
	if got := leadNotes(&models.JobLead{}); got != "" {
		t.Errorf("unscored lead notes = %q", got)
	}
	if got := leadNotes(&models.JobLead{Score: &score}); got != "Lead score 82/100" {
		t.Errorf("notes = %q", got)
	}
	if got := leadNotes(&models.JobLead{Score: &score, ScoreReasons: &reasons}); got != "Lead score 82/100: Strong Go match; hybrid London." {
		t.Errorf("notes = %q", got)
	}
}

func TestScoreLeadRequestAndErrors(t *testing.T) {
	var path string
	var body map[string]any
	reply := func(w http.ResponseWriter) { w.Write([]byte(`{"score": 82, "reasons": "Strong match"}`)) }
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		reply(w)
	}))
	t.Cleanup(srv.Close)
	s := &JobLeadService{PythonURL: srv.URL}
	lead := &models.JobLead{JobTitle: "Engineer", Company: "Acme", Location: "London", Description: "Text"}

	got, err := s.scoreLead(context.Background(), "Five years of Go", lead)
	if err != nil || got.Score != 82 || got.Reasons != "Strong match" {
		t.Fatalf("scoreLead = %+v, %v", got, err)
	}
	if path != "/internal/jobs/score" {
		t.Errorf("path = %q", path)
	}
	want := map[string]any{"profile": "Five years of Go", "jobTitle": "Engineer", "company": "Acme", "location": "London", "description": "Text"}
	for key, value := range want {
		if body[key] != value {
			t.Errorf("body[%q] = %v, want %v", key, body[key], value)
		}
	}
	if len(body) != len(want) {
		t.Errorf("body = %v", body)
	}

	// The two kinds of failure scoreLeads tells apart: Claude answering badly
	// for this posting (a PythonError, counted against the lead) and scoring
	// being unavailable altogether (anything else, which stops the run).
	var readErr *PythonError
	reply = func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"detail": "failed to parse the score"}`))
	}
	if _, err := s.scoreLead(context.Background(), "p", lead); !errors.As(err, &readErr) {
		t.Errorf("a 502 should be a PythonError, got %v", err)
	}
	reply = func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"detail": "CLAUDE_API_KEY is not set"}`))
	}
	if _, err := s.scoreLead(context.Background(), "p", lead); err == nil || errors.As(err, &readErr) {
		t.Errorf("a 503 should be a plain error, got %v", err)
	}
}

func TestLeadSyncStatusStartsIdle(t *testing.T) {
	s := InitJobLeads(&JobLeadConfig{}, nil, "http://python:8000")
	if status := s.Status(); status.Running || status.LastRunAt != nil || status.LastError != "" {
		t.Errorf("status = %+v", status)
	}
}
