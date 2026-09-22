package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Covered here: the pure logic in ratelimit.go, notes.go and gitea.go —
// sliding-window arithmetic, the path-traversal guard, and the two-stage JSON
// decoding of a Gitea feed entry (from a hand-written fixture).
//
// Deliberately not covered:
//   - FetchLatestFeed: it is an HTTP call to the Gitea container. Its only
//     logic beyond the request is the empty-feed nil,nil case, and testing it
//     would mean either a network or an injected client that does not exist.
//   - claude.go: InitClaude is a single SDK constructor with no validation and
//     no prompt building; there is nothing to assert that is not the SDK's.
//   - The 1-minute Gitea cache TTL: it lives in handlers.Store, not here.
//   - RateLimiter's window expiry uses the real clock (Allow calls time.Now
//     directly and there is no injectable clock), so the expiry case below
//     uses a deliberately tiny window rather than a long sleep.

func TestRateLimiterAllow(t *testing.T) {
	t.Run("calls up to max are allowed", func(t *testing.T) {
		rl := NewRateLimiter(3, time.Minute)
		for i := 1; i <= 3; i++ {
			if !rl.Allow("ip") {
				t.Fatalf("call %d denied, want allowed", i)
			}
		}
	})

	t.Run("the call after max is denied", func(t *testing.T) {
		rl := NewRateLimiter(2, time.Minute)
		rl.Allow("ip")
		rl.Allow("ip")
		if rl.Allow("ip") {
			t.Fatal("third call allowed, want denied")
		}
	})

	t.Run("a denied attempt is not recorded, so it cannot extend the lockout", func(t *testing.T) {
		rl := NewRateLimiter(1, time.Minute)
		rl.Allow("ip")
		for i := 0; i < 5; i++ {
			rl.Allow("ip")
		}
		// One timestamp only: the rejected hammering must not have been stored.
		if got := len(rl.attempts["ip"]); got != 1 {
			t.Fatalf("stored %d attempts, want 1", got)
		}
	})

	t.Run("keys are limited independently", func(t *testing.T) {
		rl := NewRateLimiter(1, time.Minute)
		rl.Allow("a")
		if !rl.Allow("b") {
			t.Fatal("second key denied, want allowed")
		}
	})

	t.Run("attempts older than the window stop counting", func(t *testing.T) {
		rl := NewRateLimiter(1, 5*time.Millisecond)
		rl.Allow("ip")
		time.Sleep(10 * time.Millisecond)
		if !rl.Allow("ip") {
			t.Fatal("denied after the window elapsed, want allowed")
		}
	})

	t.Run("max of zero denies everything", func(t *testing.T) {
		rl := NewRateLimiter(0, time.Minute)
		if rl.Allow("ip") {
			t.Fatal("call allowed with max 0, want denied")
		}
	})
}

func TestNotesParsePath(t *testing.T) {
	base := t.TempDir()
	notes := InitNotes(&NotesConfig{Dir: base})

	cases := []struct {
		name    string
		in      string
		want    string // relative to base; ignored when wantErr
		wantErr bool
	}{
		{"empty path is the wiki index", "", "Index.md", false},
		{"bare slash is the wiki index", "/", "Index.md", false},
		{"a plain file resolves inside the dir", "/Notes.md", "Notes.md", false},
		{"a nested file resolves inside the dir", "/private/Secret.md", "private/Secret.md", false},
		{"embedded .. that stays inside is fine", "/a/../b.md", "b.md", false},
		{"traversal above the base is rejected", "/../escape.md", "", true},
		{"deep traversal is rejected", "/a/../../escape.md", "", true},
		{"the base dir itself is rejected, only children are servable", "/.", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := notes.ParsePath(c.in)
			if c.wantErr {
				if err == nil {
					t.Fatalf("ParsePath(%q) = %q, want error", c.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePath(%q) errored: %v", c.in, err)
			}
			if want := filepath.Join(base, c.want); got != want {
				t.Fatalf("ParsePath(%q) = %q, want %q", c.in, got, want)
			}
		})
	}
}

// The sibling-directory case is the reason ParsePath appends a separator
// before the prefix check, so it gets its own test with real neighbouring dirs.
func TestNotesParsePathRejectsSiblingPrefix(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "notes")
	if err := os.Mkdir(base, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(parent, "notes-private"), 0o755); err != nil {
		t.Fatal(err)
	}

	notes := InitNotes(&NotesConfig{Dir: base})
	if got, err := notes.ParsePath("/../notes-private/Secret.md"); err == nil {
		t.Fatalf("ParsePath resolved into a sibling dir: %q", got)
	}
}

func TestParseCommitMessage(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			"first commit message of a push event",
			`{"Commits":[{"Message":"fix nginx template\n"},{"Message":"second"}]}`,
			"fix nginx template\n",
		},
		{"push event with no commits", `{"Commits":[]}`, ""},
		{"non-push event whose Content is a different shape", `"some string"`, ""},
		{"empty Content, as sent for events with no payload", "", ""},
		{"malformed JSON is swallowed rather than surfaced", `{"Commits":`, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ParseCommitMessage(c.in); got != c.want {
				t.Fatalf("ParseCommitMessage(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// Guards the struct tags against a rename: Gitea's feed uses capitalised keys
// inside Content but snake_case at the top level, which is easy to "tidy" wrongly.
func TestGiteaFeedResponseDecoding(t *testing.T) {
	const fixture = `[{
		"act_user": {"avatar_url": "http://gitea:3000/avatars/1", "login": "adamf"},
		"repo": {"full_name": "adamf/web_server", "html_url": "http://gitea:3000/adamf/web_server"},
		"op_type": "commit_repo",
		"content": "{\"Commits\":[{\"Message\":\"add tests\"}]}",
		"created": "2026-09-22T10:30:00Z",
		"unknown_field": 1
	}]`

	var items []GiteaFeedResponse
	if err := json.Unmarshal([]byte(fixture), &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d entries, want 1", len(items))
	}

	got := items[0]
	if got.ActUser.AvatarURL != "http://gitea:3000/avatars/1" {
		t.Errorf("AvatarURL = %q", got.ActUser.AvatarURL)
	}
	if got.Repo.FullName != "adamf/web_server" {
		t.Errorf("FullName = %q", got.Repo.FullName)
	}
	if got.Repo.HTMLURL != "http://gitea:3000/adamf/web_server" {
		t.Errorf("HTMLURL = %q", got.Repo.HTMLURL)
	}
	if got.OpType != "commit_repo" {
		t.Errorf("OpType = %q", got.OpType)
	}
	if !got.Created.Equal(time.Date(2026, 9, 22, 10, 30, 0, 0, time.UTC)) {
		t.Errorf("Created = %v", got.Created)
	}
	// Content is the nested JSON string ParseCommitMessage then decodes.
	if msg := ParseCommitMessage(got.Content); msg != "add tests" {
		t.Errorf("ParseCommitMessage(Content) = %q, want %q", msg, "add tests")
	}
}
