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

func (store *Store) CompleteSpotifyAuth(ctx *gin.Context) {
	state := ctx.Query("state")
	c := context.Background()
	// code := c.Query("code")

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

func (s *Store) RecentSongsFresh() bool {
	if s.RecentSongs == nil {
		return false
	}

	if len(*s.RecentSongs) == 0 {
		return false
	}

	return time.Since(s.RecentSongsFetchedAt) < time.Minute
}
