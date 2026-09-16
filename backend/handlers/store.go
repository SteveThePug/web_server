package handlers

import (
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
// Concurrency note: the cache fields below are read and written by concurrent
// requests without a mutex. The races are benign in practice (worst case two
// requests refresh the same third-party API at once, or one reads a value
// being replaced), but this is a deliberate simplification, not an oversight
// — do not add a field here that would be unsafe to tear.
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
}

// GiteaFeedFresh reports whether the cached Gitea feed entry is still within
// its one-minute TTL. A nil cache is never fresh, so the first request always
// hits Gitea.
func (s *Store) GiteaFeedFresh() bool {
	if s.GiteaFeed == nil {
		return false
	}
	return time.Since(s.GiteaFeedFetchedAt) < time.Minute
}

// SteamFresh reports whether the cached Steam data is still within its
// five-minute TTL. Note it tests only SteamRecentGames: if Steam legitimately
// returns no recent games, the cache never registers as fresh and every
// request re-queries the API.
func (s *Store) SteamFresh() bool {
	if s.SteamRecentGames == nil {
		return false
	}
	return time.Since(s.SteamFetchedAt) < 5*time.Minute
}
