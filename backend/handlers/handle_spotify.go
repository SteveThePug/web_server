package handlers

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
)

func (store *Store) ListeningTo(c *gin.Context) {
	ctx := context.Background()

	log.Default().Println("Tets")

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
