package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

// The classify step of the email pipeline: the request the Python service's
// /internal/email/classify receives, and what processEmail does with each
// kind of reply. A stub server stands in for Python.
//
// processEmail is driven only as far as it can go without a database. That
// covers every failure (they all return before the first query) and the two
// "nothing to do" answers; an email that is about a job goes on to look the
// application up, which needs Postgres and is not exercised here.

func classifierStub(t *testing.T, handler http.HandlerFunc) *EmailSyncService {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &EmailSyncService{PythonURL: srv.URL}
}

func testEmail() graphMessage {
	email := graphMessage{
		ID:               "msg-1",
		Subject:          "Your application to Acme",
		ReceivedDateTime: "2026-03-15T10:30:00Z",
		Body:             graphBody{ContentType: "html", Content: "<p>Thanks&nbsp;for applying</p>"},
	}
	email.From.EmailAddress = graphEmailAddress{Name: "Acme Hiring", Address: "no-reply@greenhouse.io"}
	return email
}

func TestClassifyEmailRequest(t *testing.T) {
	var path string
	var body map[string]any
	s := classifierStub(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		w.Write([]byte(`{"isJobEmail": true, "action": "create", "company": "Acme", "jobTitle": "Engineer",
			"status": "applied", "location": "London", "url": null, "appliedAt": "2026-03-15", "notes": null}`))
	})

	analysis, err := s.classifyEmail(context.Background(), testEmail())
	if err != nil {
		t.Fatalf("classifyEmail: %v", err)
	}
	if path != "/internal/email/classify" {
		t.Fatalf("path = %q", path)
	}
	// The whole contract: one key, holding the headers, a blank line and the
	// cleaned body. The Python side hands this string to Claude unchanged.
	want := "Subject: Your application to Acme\nFrom: Acme Hiring <no-reply@greenhouse.io>\nDate: 2026-03-15T10:30:00Z\n\nThanks for applying"
	if len(body) != 1 || body["message"] != want {
		t.Fatalf("request body = %v\nwant message %q", body, want)
	}

	if !analysis.IsJobEmail || analysis.Action != "create" || analysis.Company != "Acme" ||
		analysis.JobTitle != "Engineer" || analysis.Status != "applied" {
		t.Fatalf("analysis = %+v", analysis)
	}
	if analysis.Location == nil || *analysis.Location != "London" || analysis.URL != nil ||
		analysis.AppliedAt == nil || *analysis.AppliedAt != "2026-03-15" || analysis.Notes != nil {
		t.Fatalf("nullable fields = %+v", analysis)
	}
}

func TestProcessEmailWithoutAJob(t *testing.T) {
	cases := []struct {
		name  string
		reply string
	}{
		// What Python sends for a non-job email: every other field defaulted.
		{"not a job email", `{"isJobEmail": false, "action": "", "company": "", "jobTitle": "", "status": "",
			"location": null, "url": null, "appliedAt": null, "notes": null}`},
		// (company, job title) is the only key an application has, so an
		// answer missing either cannot be matched or created.
		{"a job email with no company", `{"isJobEmail": true, "action": "create", "company": "", "jobTitle": "Engineer", "status": "applied"}`},
		{"a job email with no title", `{"isJobEmail": true, "action": "create", "company": "Acme", "jobTitle": "", "status": "applied"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := classifierStub(t, func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(c.reply))
			})
			action, jobAppID, err := s.processEmail(context.Background(), testEmail())
			if err != nil || action != "skipped" || jobAppID != nil {
				t.Fatalf("processEmail = (%q, %v, %v), want (skipped, nil, nil)", action, jobAppID, err)
			}
		})
	}
}

// Every way the classifier can fail must come back from processEmail as an
// error with no action: that is what makes SyncEmails record the email as
// Action "error" with one more attempt used, to be retried next sync.
func TestProcessEmailClassifierFailures(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		body        string
		wantRelayed string // the 502 detail, when there is one
	}{
		{"Claude failed", http.StatusBadGateway, `{"detail": "failed to classify the email"}`, "failed to classify the email"},
		{"Claude's reply was unparseable", http.StatusBadGateway, `{"detail": "failed to parse the email classification"}`, "failed to parse the email classification"},
		{"no API key", http.StatusServiceUnavailable, `{"detail": "CLAUDE_API_KEY is not set"}`, ""},
		{"request rejected", http.StatusUnprocessableEntity, `{"detail": [{"msg": "too long"}]}`, ""},
		{"malformed 200 body", http.StatusOK, `not json`, ""},
		{"200 body of the wrong shape", http.StatusOK, `{"isJobEmail": "yes"}`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := classifierStub(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				w.Write([]byte(c.body))
			})
			action, jobAppID, err := s.processEmail(context.Background(), testEmail())
			if err == nil {
				t.Fatal("expected an error")
			}
			if action != "" || jobAppID != nil {
				t.Fatalf("processEmail = (%q, %v) alongside an error", action, jobAppID)
			}
			// Not a cancellation: SyncEmails skips those without recording
			// anything, which would leave a real failure retried for ever.
			if errors.Is(err, context.Canceled) {
				t.Fatalf("err = %v reads as a cancelled sync", err)
			}
			var pyErr *PythonError
			if errors.As(err, &pyErr) != (c.wantRelayed != "") {
				t.Fatalf("err = %v, want PythonError: %v", err, c.wantRelayed != "")
			}
			if pyErr != nil && pyErr.Detail != c.wantRelayed {
				t.Fatalf("detail = %q, want %q", pyErr.Detail, c.wantRelayed)
			}
		})
	}
}

// The other half of the rule above: a sync abandoned mid-call (the admin
// closed the tab) must be recognisable, so the email's attempts do not advance.
func TestProcessEmailCancelled(t *testing.T) {
	s := classifierStub(t, func(http.ResponseWriter, *http.Request) {})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := s.processEmail(ctx, testEmail()); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want it to wrap context.Canceled", err)
	}
}

func TestClassifierMessageIsCapped(t *testing.T) {
	// An ordinary email is nowhere near the cap and is not touched.
	if got := classifierMessage(testEmail()); !strings.HasSuffix(got, "\n\nThanks for applying") {
		t.Fatalf("message = %q", got)
	}

	// cleanEmailBody bounds the body, so only the headers can push a message
	// over. Multi-byte runes: the cap counts characters, as the Python side
	// does, and the cut must not split one.
	long := testEmail()
	long.Subject = strings.Repeat("é", maxClassifyRunes+50)
	got := classifierMessage(long)
	if n := utf8.RuneCountInString(got); n != maxClassifyRunes {
		t.Fatalf("message is %d runes, want %d", n, maxClassifyRunes)
	}
	if !utf8.ValidString(got) {
		t.Fatal("the cut split a rune")
	}

	// More bytes than the cap but fewer runes: within the limit, so intact.
	wide := testEmail()
	wide.Subject = strings.Repeat("é", maxClassifyRunes/2+10)
	if got := classifierMessage(wide); !strings.HasSuffix(got, "\n\nThanks for applying") {
		t.Fatal("a message under the rune cap was cut")
	}
}
