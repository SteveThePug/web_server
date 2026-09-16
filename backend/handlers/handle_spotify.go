package handlers

import (
	"context"
	"log"
	"net/http"

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
	// CSRF check: the state must be a nonce this process handed out from
	// StartSpotifyAuth and has not seen come back yet. Without it, anyone who
	// could get the site owner's browser to hit this URL with an attacker's
	// authorisation code would bind the site to the attacker's Spotify
	// account. The library also compares the state to the value passed below,
	// but since that value is the one just received, this check is the real
	// one.
	if !store.ConsumeSpotifyState(state) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

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

// StartSpotifyAuth backs the admin-only GET /spotify/auth. It returns the
// Spotify authorisation URL carrying a freshly minted one-shot state nonce,
// which CompleteSpotifyAuth then verifies and consumes.
//
// This exists because the state has to be per-request to be a CSRF defence,
// and the start of the flow therefore has to be a request rather than a line
// printed at start-up.
func (store *Store) StartSpotifyAuth(ctx *gin.Context) {
	state, err := store.NewSpotifyState()
	if err != nil {
		log.Println(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"url": store.SpotifyAuth.AuthURL(state)})
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

// RecentlyPlayed backs GET /spotify/recent, serving the shared one-minute
// cache. Both this endpoint and the GraphQL spotifyRecent resolver go through
// RecentlyPlayedTracks, so either one can fill the cache for the other.
func (store *Store) RecentlyPlayed(ctx *gin.Context) {
	if store.SpotifyClient == nil {
		ctx.JSON(500, gin.H{"error": "Spotify not authenticated"})
		return
	}

	played, err := store.RecentlyPlayedTracks(ctx)
	if err != nil {
		log.Println(err)
		ctx.JSON(500, gin.H{"error": "failed to fetch recently played"})
		return
	}

	ctx.JSON(200, played)
}

// RecentlyPlayedTracks returns the last few played tracks, from the cache when
// it is fresh and from Spotify otherwise — populating the cache in that case.
//
// This is the single fetch-and-cache path for recently-played data; the caller
// must have checked SpotifyClient is non-nil.
func (store *Store) RecentlyPlayedTracks(ctx context.Context) ([]spotify.RecentlyPlayedItem, error) {
	if cached, ok := store.CachedRecentSongs(); ok {
		return cached, nil
	}

	opts := spotify.RecentlyPlayedOptions{Limit: 3}
	played, err := store.SpotifyClient.PlayerRecentlyPlayedOpt(ctx, &opts)
	if err != nil {
		return nil, err
	}

	store.SetRecentSongs(played)

	return played, nil
}
