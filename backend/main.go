package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"adam-french.co.uk/backend/handlers"
	"adam-french.co.uk/backend/services"
)

func main() {
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbConfig := services.SQLConfig{User: dbUser, Password: dbPassword, DBName: dbName, Host: dbHost, Port: dbPort}
	db, err := services.InitDatabase(&dbConfig)
	if err != nil {
		log.Fatal(err)
	}

	spotifyAuthState := os.Getenv("SPOTIFY_AUTH_STATE")
	spotifyRedirectURL := os.Getenv("SPOTIFY_REDIRECT_URI")
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyConfig := services.SpotifyConfig{AuthState: spotifyAuthState, RedirectURL: spotifyRedirectURL, ClientID: spotifyClientID, ClientSecret: spotifyClientSecret}
	spotifyAuth, client := services.InitSpotifyAuth(&spotifyConfig)

	authSecret := os.Getenv("BACKEND_SECRET")
	domainName := os.Getenv("DOMAIN")
	backendEndpoint := os.Getenv("BACKEND_ENDPOINT")
	accessTokenLifetime := 24 * time.Hour
	refreshTokenLifetime := 365 * 24 * time.Hour
	authConfig := services.AuthConfig{Secret: []byte(authSecret), Domain: domainName, RefreshTokenLifetime: refreshTokenLifetime, AccessTokenLifetime: accessTokenLifetime, Endpoint: backendEndpoint}
	auth := services.InitAuth(&authConfig)

	store := handlers.Store{DB: db, SpotifyAuth: spotifyAuth, SpotifyClient: client, Auth: auth}

	r := gin.Default()
	protected := r.Group("/", store.AuthMiddlewear)

	r.GET("/posts", store.GetPosts)
	protected.POST("/posts", store.CreatePost)
	r.GET("/posts/:id", store.GetPost)
	protected.PUT("/posts/:id", store.UpdatePost)
	protected.DELETE("/posts/:id", store.DeletePost)

	r.GET("/user/:id", store.GetUser)
	protected.PUT("/user/:id", store.UpdateUser)
	protected.DELETE("/user/:id", store.DeleteUser)
	r.GET("/user", store.GetUsers)
	r.POST("/user", store.CreateUser)

	r.POST("/auth/login", store.Login)
	r.POST("/auth/refresh", store.RefreshToken)

	r.GET("/spotify/callback", store.CompleteSpotifyAuth)
	r.GET("/spotify/listening", store.ListeningTo)
	r.GET("/spotify/recent", store.RecentlyPlayed)
	// r.POST("/spotify", store.SendSong)

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	port := os.Getenv("BACKEND_PORT")
	r.Run(fmt.Sprintf(":%s", port))
}
