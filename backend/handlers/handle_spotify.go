package handlers

import (
	"context"
	"net/http"

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
		ctx.String(http.StatusInternalServerError, "Couldn't get token: %v", err)
		return
	}

	services.SaveSpotifyToken(services.SPOTIFY_TOKEN_JSON_PATH, token)

	client := spotify.New(store.SpotifyAuth.Client(c, token))

	store.SpotifyClient = client

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"token":   token.AccessToken,
		"type":    token.TokenType,
		"expiry":  token.Expiry,
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
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, playing)
}

func (store *Store) RecentlyPlayed(ctx *gin.Context) {
	opts := spotify.RecentlyPlayedOptions{Limit: 3}

	played, err := store.SpotifyClient.PlayerRecentlyPlayedOpt(ctx, &opts)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, played)
}
