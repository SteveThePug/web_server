package services

import (
	"strings"
	"testing"
	"time"

	"adam-french.co.uk/backend/models"
)

// These tests cover only the pure helpers of the email sync: body cleaning,
// truncation, fence stripping, timestamp parsing, the retry predicate and the
// filter. Everything else in email_sync.go — token acquisition, the Graph
// paging loop, processEmail and the create/update paths — needs either a live
// *gorm.DB, a Microsoft Graph endpoint or a Claude call, none of which exist
// in a test run, so they are deliberately not exercised here.
//
// matchesFilter is a method but reads no field of the receiver (it matches
// against the package-level subjectKeywords/senderDomains), which is why a
// zero-value &EmailSyncService{} is a legitimate receiver below.

func TestCleanEmailBody(t *testing.T) {
	cases := []struct {
		name        string
		content     string
		contentType string
		want        string
	}{
		{
			"html tags are stripped and entities decoded",
			"<p>Hi&nbsp;Adam &amp; team</p><b>thanks</b>",
			"html",
			"Hi Adam & team thanks",
		},
		{
			"contentType match is case-insensitive",
			"<p>hello</p>",
			"HTML",
			"hello",
		},
		{
			// Entities are only decoded on the HTML path, so plain text keeps
			// them verbatim rather than being silently rewritten.
			"text bodies are not entity-decoded",
			"a &amp; b",
			"text",
			"a &amp; b",
		},
		{
			"runs of spaces and tabs collapse to one space",
			"one   \t  two",
			"text",
			"one two",
		},
		{
			"three or more newlines collapse to a blank line",
			"a\n\n\n\n\nb",
			"text",
			"a\n\nb",
		},
		{
			"signature marker cuts the tail",
			"body text\n-- \nAdam French\n07000",
			"text",
			"body text",
		},
		{
			"Sent from my iPhone is treated as a signature",
			"see attached\nSent from my iPhone",
			"text",
			"see attached",
		},
		{
			// idx > 0, not >= 0: a body that is nothing but a signature would
			// otherwise be blanked entirely, losing the only content there is.
			"a body starting with a signature marker is left intact",
			"Sent from my iPhone\nreal content",
			"text",
			"Sent from my iPhone\nreal content",
		},
		{
			"quoted lines are dropped",
			"my reply\n> their mail\n  > indented quote\nmore reply",
			"text",
			"my reply\nmore reply",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := cleanEmailBody(c.content, c.contentType); got != c.want {
				t.Fatalf("cleanEmailBody(%q, %q) = %q, want %q", c.content, c.contentType, got, c.want)
			}
		})
	}
}

// The 4000-byte cap is what bounds the prompt cost, so it gets its own test:
// the cut is by bytes and is applied before the final TrimSpace.
func TestCleanEmailBodyCap(t *testing.T) {
	got := cleanEmailBody(strings.Repeat("a", 5000), "text")
	if len(got) != 4000 {
		t.Fatalf("len = %d, want 4000", len(got))
	}

	// A body exactly at the cap must pass through untouched.
	exact := strings.Repeat("b", 4000)
	if got := cleanEmailBody(exact, "text"); got != exact {
		t.Fatalf("body at the cap was modified: len = %d", len(got))
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"shorter than the cap is unchanged", "abc", 10, "abc"},
		{"exactly at the cap is unchanged", "abcde", 5, "abcde"},
		{"longer is cut to the cap", "abcdef", 3, "abc"},
		{"zero cap yields the empty string", "abc", 0, ""},
		{"empty input", "", 5, ""},
		// Bytes, not runes: the cut can land mid-rune. Asserted so the
		// behaviour is a documented choice rather than a lurking surprise.
		{"multi-byte runes are cut by byte", "é", 1, "\xc3"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := truncate(c.s, c.maxLen); got != c.want {
				t.Fatalf("truncate(%q, %d) = %q, want %q", c.s, c.maxLen, got, c.want)
			}
		})
	}
}

func TestStripMarkdownFence(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"json-tagged fence", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"bare fence", "```\n{\"a\":1}\n```", `{"a":1}`},
		{"unfenced input is returned as-is", `{"a":1}`, `{"a":1}`},
		{"surrounding whitespace is trimmed", "\n\n  {\"a\":1}  \n", `{"a":1}`},
		{"leading whitespace before the fence", "  ```json\n{\"a\":1}\n```  ", `{"a":1}`},
		{"opening fence only", "```json\n{\"a\":1}", `{"a":1}`},
		{"empty input", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := StripMarkdownFence(c.raw); got != c.want {
				t.Fatalf("StripMarkdownFence(%q) = %q, want %q", c.raw, got, c.want)
			}
		})
	}
}

func TestParseReceivedAt(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want time.Time
	}{
		{"valid RFC3339", "2024-03-15T10:30:00Z", time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)},
		{"empty means no timestamp", "", time.Time{}},
		{"malformed falls back to zero rather than failing the sync", "15/03/2024", time.Time{}},
		{"a date without a time is not RFC3339", "2024-03-15", time.Time{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := parseReceivedAt(c.raw); !got.Equal(c.want) {
				t.Fatalf("parseReceivedAt(%q) = %v, want %v", c.raw, got, c.want)
			}
		})
	}
}

func TestRetryableFailure(t *testing.T) {
	cases := []struct {
		name     string
		action   string
		attempts int
		want     bool
	}{
		{"a failure with attempts left is retried", "error", maxEmailAttempts - 1, true},
		{"a failure at the cap is abandoned", "error", maxEmailAttempts, false},
		{"a failure past the cap is abandoned", "error", maxEmailAttempts + 1, false},
		{"a success is terminal even with attempts left", "created", 0, false},
		{"a skip is terminal too", "skipped", 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			row := &models.ProcessedEmail{Action: c.action, Attempts: c.attempts}
			if got := retryableFailure(row); got != c.want {
				t.Fatalf("retryableFailure(%s, %d) = %v, want %v", c.action, c.attempts, got, c.want)
			}
		})
	}
}

func TestMatchesFilter(t *testing.T) {
	cases := []struct {
		name    string
		subject string
		sender  string
		want    bool
	}{
		{"subject keyword anywhere in the line", "Your application to Acme", "hr@acme.com", true},
		{"subject keyword is case-insensitive", "INTERVIEW CONFIRMED", "hr@acme.com", true},
		{"sender on a tracked ATS domain", "hello", "no-reply@greenhouse.io", true},
		{"sender on a subdomain of a tracked domain", "hello", "no-reply@mail.greenhouse.io", true},
		{"sender domain match is case-insensitive", "hello", "No-Reply@Lever.co", true},
		// The rule is an @ or . boundary, so a lookalike domain that merely
		// ends in the same letters must not match.
		{"lookalike domain does not match", "hello", "spam@notgreenhouse.io", false},
		{"neither subject nor sender matches", "lunch tomorrow?", "friend@example.com", false},
	}

	s := &EmailSyncService{}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			email := graphMessage{Subject: c.subject}
			email.From.EmailAddress.Address = c.sender
			if got := s.matchesFilter(email); got != c.want {
				t.Fatalf("matchesFilter(%q, %q) = %v, want %v", c.subject, c.sender, got, c.want)
			}
		})
	}
}
