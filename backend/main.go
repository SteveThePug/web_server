package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"adam-french.co.uk/backend/handlers"
	"adam-french.co.uk/backend/services"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func main() {
	db_user := os.Getenv("POSTGRES_USER")
	db_password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	db_host := os.Getenv("POSTGRES_HOST")
	db_port := os.Getenv("POSTGRES_PORT")
	db_config := services.SQLConfig{User: db_user, Password: db_password, DBName: dbname, Host: db_host, Port: db_port}
	db, err := services.InitDatabase(db_config)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	authState := os.Getenv("SPOTIFY_AUTH_STATE")
	redirectURL := os.Getenv("SPOTIFY_REDIRECT_URI")
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURL),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
		),
	)

	fmt.Println("Authenticate spotify with:")
	fmt.Println(auth.AuthURL(authState))

	store := handlers.Store{DB: db, SpotifyAuth: auth, SpotifyToken: nil}

	r.GET("/posts", store.GetPosts)
	r.POST("/posts", store.CreatePost)
	r.PUT("/posts/:id", store.UpdatePost)

	r.GET("/callback", store.CompleteAuth)
	// r.GET("/spotify", store.ListeningTo)
	// r.POST("/spotify", store.SendSong)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	port := os.Getenv("BACKEND_PORT")
	r.Run(fmt.Sprintf(":%s", port))
}
