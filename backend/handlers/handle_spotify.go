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

	// If Spotify says "nothing is currently playing"
	if playing == nil || !playing.Playing || playing.Item == nil {
		ctx.JSON(200, gin.H{
			"playing": false,
			"message": "User is not currently listening to anything",
		})
		return
	}

	// Extract fields safely
	item := playing.Item
	artistName := ""
	if len(item.Artists) > 0 {
		artistName = item.Artists[0].Name
	}

	imgURL := ""
	if len(item.Album.Images) > 0 {
		imgURL = item.Album.Images[0].URL
	}

	ctx.JSON(200, gin.H{
		"playing":     true,
		"song_name":   item.Name,
		"song_url":    item.PreviewURL,
		"artist_name": artistName,
		"album_image": imgURL,
	})
}
