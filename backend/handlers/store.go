package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"adam-french.co.uk/backend/services"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"gorm.io/gorm"
)

// Package handlers holds the REST endpoints and the shared Store that every
// request handler — and, through graph.Resolver, every GraphQL resolver —
// depends on.
//
// Store is the backend's dependency container: it is built once in main and
// passed around by pointer, so the caches on it are shared process-wide.
//
// Concurrency note: the third-party caches below are read and written by
// concurrent requests, so every access goes through cacheMu and the accessor
// methods at the bottom of this file. Do not touch the cache fields directly
// — an unsynchronised read of one of them is a data race, not merely a stale
// value.
type Store struct {
	DB            *gorm.DB
	SpotifyAuth   *spotifyauth.Authenticator
	SpotifyClient *spotify.Client
	ClaudeClient  *anthropic.Client
	Auth          *services.Auth
	Notes         *services.Notes
	LoginLimiter  *services.RateLimiter
	EmailSync     *services.EmailSyncService

	// Cached "recently played" tracks, refreshed at most once a minute by the
	// GraphQL spotifyRecent resolver. A pointer to a slice so that nil can
	// mean "never fetched" distinctly from "fetched, empty".
	// cacheMu guards every cache field below (and only those). One mutex for
	// all of them keeps the locking obvious; the critical sections are just
	// field copies, never network calls, so there is no contention worth
	// splitting it up for.
	cacheMu sync.RWMutex

	RecentSongs          *[]spotify.RecentlyPlayedItem
	RecentSongsFetchedAt time.Time

	// Cached latest Gitea activity entry, 1 minute TTL.
	GiteaHost          string
	GiteaPort          string
	GiteaFeed          *services.GiteaFeedResponse
	GiteaFeedFetchedAt time.Time

	// Cached Steam status and recent games, 5 minute TTL.
	SteamAPIKey      string
	SteamID          string
	SteamRecentGames []services.SteamRecentGame
	SteamOnline      bool
	SteamFetchedAt   time.Time

	// spotifyStates holds the one-shot CSRF nonces handed out by
	// GET /spotify/auth, mapped to their creation time. In-memory is enough:
	// there is one backend instance, and losing them on restart only means
	// restarting an OAuth flow that takes seconds.
	spotifyStatesMu sync.Mutex
	spotifyStates   map[string]time.Time

	// spotifyNilLoggedAt rate-limits the "Spotify not authenticated" log line:
	// every page visit hits spotifyRecent, so an unauthenticated backend would
	// otherwise write a flood of identical lines. nil means "recently logged".
	spotifyNilLoggedMu sync.Mutex
	spotifyNilLoggedAt time.Time
}

// spotifyStateTTL is how long an unused Spotify OAuth nonce stays valid. Long
// enough to log in at Spotify and approve the scopes, short enough that a
// leaked authorisation URL is not reusable later.
const spotifyStateTTL = 10 * time.Minute

// NewSpotifyState mints a one-shot CSRF nonce for the Spotify authorisation
// URL and remembers it. Expired entries are pruned on the way through, which
// is the only cleanup there is — the map is otherwise unbounded, and only an
// admin can add to it.
func (s *Store) NewSpotifyState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	s.spotifyStatesMu.Lock()
	defer s.spotifyStatesMu.Unlock()
	if s.spotifyStates == nil {
		s.spotifyStates = make(map[string]time.Time)
	}
	for k, created := range s.spotifyStates {
		if time.Since(created) > spotifyStateTTL {
			delete(s.spotifyStates, k)
		}
	}
	s.spotifyStates[state] = time.Now()

	return state, nil
}

// ConsumeSpotifyState checks a nonce returned by Spotify's redirect and
// removes it, so a replayed callback fails. An unknown or expired nonce
// returns false, which the callback must treat as a rejected request.
func (s *Store) ConsumeSpotifyState(state string) bool {
	if state == "" {
		return false
	}

	s.spotifyStatesMu.Lock()
	defer s.spotifyStatesMu.Unlock()
	created, ok := s.spotifyStates[state]
	if !ok {
		return false
	}
	delete(s.spotifyStates, state)

	return time.Since(created) <= spotifyStateTTL
}

// The cache accessors below are the only supported way to touch the cache
// fields. Each "Cached..." method returns the value plus a bool that is true
// only when the entry is present and inside its TTL; each "Set..." method
// stores a value and stamps it.

// LogSpotifyUnauthenticated logs that the Spotify OAuth flow has never been
// completed (or a stored token stopped working), with a where hint. It is
// rate-limited to one line per hour — see spotifyNilLoggedAt — because
// spotifyRecent is hit by every home-page visit.
func (s *Store) LogSpotifyUnauthenticated(where string) {
	s.spotifyNilLoggedMu.Lock()
	defer s.spotifyNilLoggedMu.Unlock()
	if time.Since(s.spotifyNilLoggedAt) < time.Hour {
		return
	}
	s.spotifyNilLoggedAt = time.Now()
	log.Printf("[Spotify] not authenticated (seen in %s) — an admin can reconnect via GET /api/spotify/auth", where)
}

// CachedRecentSongs returns the cached Spotify recently-played list if it is
// within its one-minute TTL. An empty list counts as stale, so an empty
// listening history is re-fetched every time (unchanged behaviour).
func (s *Store) CachedRecentSongs() ([]spotify.RecentlyPlayedItem, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if s.RecentSongs == nil || len(*s.RecentSongs) == 0 {
		return nil, false
	}
	if time.Since(s.RecentSongsFetchedAt) >= time.Minute {
		return nil, false
	}
	return *s.RecentSongs, true
}

// SetRecentSongs replaces the cached recently-played list.
func (s *Store) SetRecentSongs(songs []spotify.RecentlyPlayedItem) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.RecentSongs = &songs
	s.RecentSongsFetchedAt = time.Now()
}

// CachedGiteaFeed returns the cached Gitea activity entry if it is within its
// one-minute TTL. A nil cache is never fresh, so the first request always hits
// Gitea.
func (s *Store) CachedGiteaFeed() (*services.GiteaFeedResponse, bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if s.GiteaFeed == nil || time.Since(s.GiteaFeedFetchedAt) >= time.Minute {
		return nil, false
	}
	return s.GiteaFeed, true
}

// SetGiteaFeed replaces the cached Gitea feed entry.
func (s *Store) SetGiteaFeed(feed *services.GiteaFeedResponse) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.GiteaFeed = feed
	s.GiteaFeedFetchedAt = time.Now()
}

// CachedSteam returns the cached Steam recent games and online flag if they
// are within their five-minute TTL. As before, freshness is judged only on the
// games list: if Steam legitimately returns no recent games the cache never
// registers as fresh and every request re-queries the API.
func (s *Store) CachedSteam() (games []services.SteamRecentGame, online bool, ok bool) {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if s.SteamRecentGames == nil || time.Since(s.SteamFetchedAt) >= 5*time.Minute {
		return nil, false, false
	}
	return s.SteamRecentGames, s.SteamOnline, true
}

// SetSteam replaces the cached Steam data.
func (s *Store) SetSteam(games []services.SteamRecentGame, online bool) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.SteamRecentGames = games
	s.SteamOnline = online
	s.SteamFetchedAt = time.Now()
}
