package services

import (
	"testing"
	"time"

	"adam-french.co.uk/backend/models"
)

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
			got, ok := CanonicalJobURL(c.in)
			if ok != (c.want != "") || got != c.want {
				t.Fatalf("CanonicalJobURL(%q) = %q, %v; want %q", c.in, got, ok, c.want)
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
		in := JobCaptureInput{Status: "applied", Location: "London", URL: "https://x.com/1", Notes: "Hybrid", Description: "Full text"}
		if !MergeCapture(&app, in, now) {
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
		MergeCapture(&app, JobCaptureInput{Status: "saved"}, now)
		if app.Status != "saved" || app.AppliedAt != nil {
			t.Fatalf("status = %q, appliedAt = %v", app.Status, app.AppliedAt)
		}
	})

	t.Run("saved then applied advances and dates the application", func(t *testing.T) {
		app := models.JobApplication{Status: "saved"}
		if !MergeCapture(&app, JobCaptureInput{Status: "applied"}, now) {
			t.Fatal("expected a change")
		}
		if app.Status != "applied" || app.AppliedAt == nil || !app.AppliedAt.Equal(now) {
			t.Fatalf("status = %q, appliedAt = %v", app.Status, app.AppliedAt)
		}
	})

	t.Run("a repeat capture cannot move an application backwards", func(t *testing.T) {
		app := models.JobApplication{Status: "interviewing", AppliedAt: &earlier}
		if MergeCapture(&app, JobCaptureInput{Status: "saved"}, now) {
			t.Fatal("expected no change")
		}
		if app.Status != "interviewing" || !app.AppliedAt.Equal(earlier) {
			t.Fatalf("status = %q, appliedAt = %v", app.Status, app.AppliedAt)
		}
	})

	t.Run("fields already set are kept, empty ones are filled", func(t *testing.T) {
		app := models.JobApplication{Status: "applied", AppliedAt: &earlier, Location: strPtr("Remote"), URL: strPtr("")}
		in := JobCaptureInput{Status: "applied", Location: "London", URL: "https://x.com/1", Description: "Full text"}
		if !MergeCapture(&app, in, now) {
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
		in := JobCaptureInput{Status: "applied", Notes: "Hybrid, £70k"}
		if !MergeCapture(&app, in, now) || *app.Notes != "Recruiter: Sam\nHybrid, £70k" {
			t.Fatalf("notes = %q", *app.Notes)
		}
		if MergeCapture(&app, in, now) {
			t.Fatalf("the same note was appended twice: %q", *app.Notes)
		}
	})
}
