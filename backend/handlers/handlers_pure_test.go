package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zmb3/spotify/v2"
)

// What is deliberately NOT covered here, and why:
//
//   - Every handler method (Login, UploadMessageFile, the radio file movers,
//     the Spotify and email endpoints). They all need a live *gorm.DB, a real
//     multipart request, the container's /backend/... directories, or a paid
//     Claude call. Only postgres is in go.mod, so a DB would mean a server in
//     CI.
//   - The radio/chat upload filename sanitisation itself. It is written inline
//     inside the handler bodies rather than in a helper, so testing it would
//     mean either driving a full multipart upload against the real upload
//     directories or refactoring production code. Neither was in scope, so the
//     part that *is* extractable — the extension allow-lists those handlers
//     gate on — is covered as a table contract below instead.
//   - Store.isAdminRequest and the refresh fallback in ValidateAdmin: both
//     reach the database on the refresh path.
//
// What is covered: the Spotify OAuth CSRF nonce store, the TTL rules of the
// three shared caches, the admin claim check on the upload trust boundary, and
// the upload allow-list tables.

// --- Spotify OAuth CSRF nonces -------------------------------------------

func TestSpotifyStateSingleUse(t *testing.T) {
	s := &Store{}

	state, err := s.NewSpotifyState()
	if err != nil {
		t.Fatalf("NewSpotifyState: %v", err)
	}
	if !s.ConsumeSpotifyState(state) {
		t.Fatal("a freshly minted state must be accepted")
	}
	// The whole point of the nonce: a replayed callback must not authenticate
	// a second time.
	if s.ConsumeSpotifyState(state) {
		t.Fatal("a consumed state must not be accepted again")
	}
}

func TestConsumeSpotifyStateRejections(t *testing.T) {
	cases := []struct {
		name string
		// setup returns the state string to present to ConsumeSpotifyState.
		setup func(s *Store) string
		want  bool
	}{
		{
			name:  "empty state is rejected before the map is even consulted",
			setup: func(*Store) string { return "" },
			want:  false,
		},
		{
			name:  "a state this process never minted is rejected",
			setup: func(*Store) string { return "deadbeef" },
			want:  false,
		},
		{
			name: "a state minted within the TTL is accepted",
			setup: func(s *Store) string {
				st, _ := s.NewSpotifyState()
				return st
			},
			want: true,
		},
		{
			name: "a state older than the TTL is rejected",
			setup: func(s *Store) string {
				st, _ := s.NewSpotifyState()
				// Backdating beats sleeping for ten minutes; the map is in the
				// same package so the creation stamp is writable here.
				s.spotifyStates[st] = time.Now().Add(-spotifyStateTTL - time.Second)
				return st
			},
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &Store{}
			state := c.setup(s)
			if got := s.ConsumeSpotifyState(state); got != c.want {
				t.Fatalf("ConsumeSpotifyState = %v, want %v", got, c.want)
			}
		})
	}
}

// An expired nonce must also be removed, not merely refused: otherwise the map
// grows without bound for any flow that is started and abandoned.
func TestNewSpotifyStatePrunesExpired(t *testing.T) {
	s := &Store{}
	stale, _ := s.NewSpotifyState()
	s.spotifyStates[stale] = time.Now().Add(-spotifyStateTTL - time.Second)

	if _, err := s.NewSpotifyState(); err != nil {
		t.Fatalf("NewSpotifyState: %v", err)
	}
	if _, ok := s.spotifyStates[stale]; ok {
		t.Fatal("minting a new state should have pruned the expired one")
	}
}

// --- Cache TTL rules ------------------------------------------------------

// Each cache treats "empty" as stale, which is deliberate (an empty listening
// history or games list is re-fetched rather than cached). These tables pin
// that down so it cannot be "fixed" into caching emptiness by accident.
func TestCachedRecentSongsFreshness(t *testing.T) {
	song := []spotify.RecentlyPlayedItem{{}}

	cases := []struct {
		name  string
		setup func(s *Store)
		wantOK bool
	}{
		{"never fetched", func(*Store) {}, false},
		{"fetched but empty counts as stale", func(s *Store) {
			s.SetRecentSongs([]spotify.RecentlyPlayedItem{})
		}, false},
		{"just fetched", func(s *Store) { s.SetRecentSongs(song) }, true},
		{"inside the one-minute TTL", func(s *Store) {
			s.SetRecentSongs(song)
			s.RecentSongsFetchedAt = time.Now().Add(-59 * time.Second)
		}, true},
		{"past the one-minute TTL", func(s *Store) {
			s.SetRecentSongs(song)
			s.RecentSongsFetchedAt = time.Now().Add(-time.Minute - time.Second)
		}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &Store{}
			c.setup(s)
			if _, ok := s.CachedRecentSongs(); ok != c.wantOK {
				t.Fatalf("fresh = %v, want %v", ok, c.wantOK)
			}
		})
	}
}

func TestCachedGiteaFeedFreshness(t *testing.T) {
	feed := &services.GiteaFeedResponse{}

	cases := []struct {
		name   string
		setup  func(s *Store)
		wantOK bool
	}{
		{"never fetched", func(*Store) {}, false},
		{"just fetched", func(s *Store) { s.SetGiteaFeed(feed) }, true},
		{"past the one-minute TTL", func(s *Store) {
			s.SetGiteaFeed(feed)
			s.GiteaFeedFetchedAt = time.Now().Add(-time.Minute - time.Second)
		}, false},
		// Unlike the songs cache, a nil feed is the only "empty" here, and a
		// nil stored feed must never read as fresh.
		{"explicitly cached nil is never fresh", func(s *Store) { s.SetGiteaFeed(nil) }, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &Store{}
			c.setup(s)
			if _, ok := s.CachedGiteaFeed(); ok != c.wantOK {
				t.Fatalf("fresh = %v, want %v", ok, c.wantOK)
			}
		})
	}
}

func TestCachedSteamFreshness(t *testing.T) {
	games := []services.SteamRecentGame{{}}

	cases := []struct {
		name       string
		setup      func(s *Store)
		wantOK     bool
		wantOnline bool
	}{
		{"never fetched", func(*Store) {}, false, false},
		{"just fetched, online flag preserved", func(s *Store) { s.SetSteam(games, true) }, true, true},
		{"just fetched, offline", func(s *Store) { s.SetSteam(games, false) }, true, false},
		{"nil games list counts as stale even when online", func(s *Store) { s.SetSteam(nil, true) }, false, false},
		{"past the five-minute TTL", func(s *Store) {
			s.SetSteam(games, true)
			s.SteamFetchedAt = time.Now().Add(-5*time.Minute - time.Second)
		}, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &Store{}
			c.setup(s)
			_, online, ok := s.CachedSteam()
			if ok != c.wantOK {
				t.Fatalf("fresh = %v, want %v", ok, c.wantOK)
			}
			if online != c.wantOnline {
				t.Fatalf("online = %v, want %v", online, c.wantOnline)
			}
		})
	}
}

// --- Upload trust boundary ------------------------------------------------

// requestIsAdmin gates private (admin-only) attachment uploads, so every way
// the claim can be absent or the wrong shape must read as "not admin". No DB
// or network is involved: it only inspects the Gin context.
func TestRequestIsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	claims := func(m jwt.MapClaims) any { return &m }

	cases := []struct {
		name string
		// set is nil when nothing should be put in the context at all.
		set  any
		want bool
	}{
		{"no claims in context at all", nil, false},
		{"claims of the wrong type are not trusted", "not-claims", false},
		// A by-value jwt.MapClaims is not what AuthMiddlewear stores; accepting
		// it would mean the type assertion had been loosened.
		{"claims stored by value rather than by pointer", jwt.MapClaims{"admin": true}, false},
		{"pre-admin-claim token has no admin key", claims(jwt.MapClaims{"id": 1.0}), false},
		{"admin false", claims(jwt.MapClaims{"admin": false}), false},
		{"admin as a string is not a truthy admin", claims(jwt.MapClaims{"admin": "true"}), false},
		{"admin true", claims(jwt.MapClaims{"admin": true}), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			if c.set != nil {
				ctx.Set("userClaims", c.set)
			}
			if got := requestIsAdmin(ctx); got != c.want {
				t.Fatalf("requestIsAdmin = %v, want %v", got, c.want)
			}
		})
	}
}

// The allow-list tables are the upload trust boundary's data. Testing them as
// data catches the realistic failure: someone adding an executable or markup
// extension, or adding an extension to the MIME map without allowing it (which
// would be dead configuration).
func TestUploadAllowLists(t *testing.T) {
	t.Run("dangerous extensions stay off the chat allow-list", func(t *testing.T) {
		// Served by nginx from a public directory with no auth, so anything
		// the browser will execute or render as markup must not be storable.
		for _, ext := range []string{".html", ".htm", ".svg", ".js", ".php", ".exe", ".sh", ".xml", ""} {
			if allowedExtensions[ext] {
				t.Errorf("%q must not be an allowed chat upload extension", ext)
			}
		}
	})

	t.Run("allow-list keys are lower-cased with a leading dot", func(t *testing.T) {
		// The handlers look up strings.ToLower(filepath.Ext(...)), so a key in
		// any other shape is unreachable.
		for _, m := range []map[string]bool{allowedExtensions, allowedAudioExtensions} {
			for ext := range m {
				if ext == "" || ext[0] != '.' || ext != strings.ToLower(ext) {
					t.Errorf("malformed allow-list key %q", ext)
				}
			}
		}
	})

	t.Run("every MIME-checked extension is also allowed", func(t *testing.T) {
		for ext := range extensionToMIMEPrefix {
			if !allowedExtensions[ext] {
				t.Errorf("%q has a MIME prefix but is not an allowed extension, so the check never runs", ext)
			}
		}
	})

	t.Run("audio extensions are content-unverified by design", func(t *testing.T) {
		// http.DetectContentType is unreliable for these, so a missing entry
		// means "extension allowed, content not verified". Adding one would
		// start rejecting legitimate uploads.
		for _, ext := range []string{".mp3", ".ogg"} {
			if _, ok := extensionToMIMEPrefix[ext]; ok {
				t.Errorf("%q should not be MIME-checked", ext)
			}
		}
	})

	t.Run("radio allow-list is audio only", func(t *testing.T) {
		for _, ext := range []string{".mp4", ".webm", ".pdf", ".txt", ".png", ".html"} {
			if allowedAudioExtensions[ext] {
				t.Errorf("%q must not be an allowed radio upload extension", ext)
			}
		}
	})
}
