package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
)

// These REST endpoints predate the equivalent GraphQL queries in
// graph/spotify.resolvers.go and are kept for the OAuth callback (which must
// be REST, because Spotify redirects a browser to it) and for direct use.

// CompleteSpotifyAuth handles Spotify's OAuth redirect to
// GET /spotify/callback. It swaps the code for a token, saves it so the
// authorisation survives restarts, and installs the authenticated client on
// the Store — which is how Store.SpotifyClient becomes non-nil without a
// restart.
func (store *Store) CompleteSpotifyAuth(ctx *gin.Context) {
	state := ctx.Query("state")
	// context.Background, not the request context: the client built below
	// outlives this request and would otherwise be cancelled when it ends.
	c := context.Background()

	// The library pulls the `code` out of the request itself and checks the
	// state parameter matches the one passed here.
	token, err := store.SpotifyAuth.Token(c, state, ctx.Request)
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}

	if err := services.SaveSpotifyToken(services.SPOTIFY_TOKEN_JSON_PATH, token); err != nil {
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	client := spotify.New(store.SpotifyAuth.Client(c, token))

	store.SpotifyClient = client

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
	})
}

// ListeningTo backs GET /spotify/listening with the currently playing track.
// A nil client means the OAuth flow has never been completed, reported here
// as a 500.
func (store *Store) ListeningTo(ctx *gin.Context) {
	c := ctx.Request.Context()

	if store.SpotifyClient == nil {
		ctx.JSON(500, gin.H{"error": "Spotify not authenticated"})
		return
	}

	playing, err := store.SpotifyClient.PlayerCurrentlyPlaying(c)
	if err != nil {
		log.Println(err)
		ctx.JSON(500, gin.H{"error": "failed to fetch currently playing"})
		return
	}

	ctx.JSON(200, playing)
}

// RecentlyPlayed backs GET /spotify/recent.
//
// It reads the cache but never writes it: only the GraphQL spotifyRecent
// resolver populates Store.RecentSongs. So this endpoint serves cached data
// only when a GraphQL query happened to fill the cache in the last minute,
// and otherwise calls Spotify every time.
func (store *Store) RecentlyPlayed(ctx *gin.Context) {
	if store.SpotifyClient == nil {
		ctx.JSON(500, gin.H{"error": "Spotify not authenticated"})
		return
	}

	opts := spotify.RecentlyPlayedOptions{Limit: 3}

	if store.RecentSongsFresh() {
		ctx.JSON(200, *store.RecentSongs)
		return
	}

	played, err := store.SpotifyClient.PlayerRecentlyPlayedOpt(ctx, &opts)
	if err != nil {
		log.Println(err)
		ctx.JSON(500, gin.H{"error": "failed to fetch recently played"})
		return
	}

	ctx.JSON(200, played)
}

// RecentSongsFresh reports whether the cached recently-played list is within
// its one-minute TTL. An empty list counts as stale, so an empty listening
// history is re-fetched every time.
func (s *Store) RecentSongsFresh() bool {
	if s.RecentSongs == nil {
		return false
	}

	if len(*s.RecentSongs) == 0 {
		return false
	}

	return time.Since(s.RecentSongsFetchedAt) < time.Minute
}
