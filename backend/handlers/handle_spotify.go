package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zmb3/spotify/v2"
)

func (store *Store) CompleteAuth(c *gin.Context) {
	state := c.Query("state")
	// code := c.Query("code")

	token, err := store.SpotifyAuth.Token(c.Request.Context(), state, c.Request)
	if err != nil {
		c.String(http.StatusInternalServerError, "Couldn't get token: %v", err)
		return
	}

	store.SpotifyToken = token

	client := spotify.New(store.SpotifyAuth.Client(c.Request.Context(), token))

	store.SpotifyClient = client

	c.JSON(http.StatusOK, gin.H{
		"message": "Authentication successful",
		"token":   token.AccessToken,
		"type":    token.TokenType,
		"expiry":  token.Expiry,
	})
}

func (store *Store) ListeningTo(c *gin.Context) {
	ctx := c.Request.Context()

	if store.SpotifyClient == nil {
		c.JSON(500, gin.H{"error": "Spotify not authenticated"})
		return
	}

	playing, err := store.SpotifyClient.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// If Spotify says "nothing is currently playing"
	if playing == nil || !playing.Playing || playing.Item == nil {
		c.JSON(200, gin.H{
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

	c.JSON(200, gin.H{
		"playing":     true,
		"song_name":   item.Name,
		"artist_name": artistName,
		"album_image": imgURL,
	})
}
