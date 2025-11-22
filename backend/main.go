package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"adam-french.co.uk/backend/handlers"
	"adam-french.co.uk/backend/services"
)

func main() {
	db, err := services.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}

	spotify_client, err := services.InitSpotify()
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	store := handlers.Store{DB: db, SpotifyClient: spotify_client}

	r.GET("/posts", store.GetPosts)
	r.POST("/posts", store.CreatePost)
	r.PUT("/posts/:id", store.UpdatePost)

	r.GET("/spotify", store.ListeningTo)
	// r.POST("/spotify", store.SendSong)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	port := os.Getenv("BACKEND_PORT")
	r.Run(fmt.Sprintf(":%s", port))
}
