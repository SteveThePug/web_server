package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"adam-french.co.uk/backend/handlers"
	"adam-french.co.uk/backend/services"
)

func main() {
	logsDir := "/backend/logs"
	logFile, err := os.OpenFile(logsDir+"/go.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)

	r := gin.Default()

	err = r.SetTrustedProxies([]string{"172.28.0.0/16"})
	if err != nil {
		panic(err)
	}

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

	// SPOTIFY
	spotifyAuthState := os.Getenv("SPOTIFY_AUTH_STATE")
	spotifyRedirectURL := os.Getenv("SPOTIFY_REDIRECT_URI")
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyConfig := services.SpotifyConfig{AuthState: spotifyAuthState, RedirectURL: spotifyRedirectURL, ClientID: spotifyClientID, ClientSecret: spotifyClientSecret}
	spotifyAuth, spotifyClient := services.InitSpotifyAuth(&spotifyConfig)

	// CLAUDE
	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	claudeConfig := services.ClaudeConfig{APIKey: claudeAPIKey}
	claudeClient := services.InitClaude(&claudeConfig)

	authSecret := os.Getenv("BACKEND_SECRET")
	domainName := os.Getenv("DOMAIN")
	backendEndpoint := os.Getenv("BACKEND_ENDPOINT")
	accessTokenLifetime := 24 * time.Hour
	refreshTokenLifetime := 365 * 24 * time.Hour
	authConfig := services.AuthConfig{Secret: []byte(authSecret), Domain: domainName, RefreshTokenLifetime: refreshTokenLifetime, AccessTokenLifetime: accessTokenLifetime, Endpoint: backendEndpoint}
	auth := services.InitAuth(&authConfig)

	notesDir := "/backend/notes"
	notesConfig := services.NotesConfig{Dir: notesDir}
	notes := services.InitNotes(&notesConfig)

	store := handlers.Store{DB: db, SpotifyAuth: spotifyAuth, SpotifyClient: spotifyClient, ClaudeClient: claudeClient, Auth: auth, Notes: notes}

	protected := r.Group("/", store.AuthMiddlewear)

	// FAVORITES
	r.GET("/favorites", store.GetFavorites)
	protected.POST("/favorites", store.CreateFavorite)

	// ROWING
	r.GET("/rowing", store.GetRowing)
	protected.POST("/rowing", store.CreateRowing)

	// ACTIVITIES
	r.GET("/activity", store.GetActivity)
	protected.POST("/activity", store.CreateActivity)

	// POSTS
	r.GET("/posts", store.GetPosts)
	protected.POST("/posts", store.CreatePost)
	r.GET("/posts/:id", store.GetPost)
	protected.PUT("/posts/:id", store.UpdatePost)
	protected.DELETE("/posts/:id", store.DeletePost)

	// USERS
	r.GET("/user/:id", store.GetUser)
	protected.PUT("/user/:id", store.UpdateUser)
	protected.DELETE("/user/:id", store.DeleteUser)
	r.GET("/user", store.GetUsers)
	r.POST("/user", store.CreateUser)

	// AUTH
	r.POST("/auth/login", store.Login)
	r.POST("/auth/refresh", store.RefreshToken)
	r.GET("/auth/check", store.CheckToken)
	r.POST("/auth/logout", store.Logout)

	// SPOTIFY
	r.GET("/spotify/callback", store.CompleteSpotifyAuth)
	r.GET("/spotify/listening", store.ListeningTo)
	r.GET("/spotify/recent", store.RecentlyPlayed)
	// r.POST("/spotify", store.SendSong)

	// MESSAGES
	r.GET("/ws", store.ConnectWebSocket)

	// NOTES
	r.GET("/notes/*path", store.GetNoteFile)

	// HELLO WORLD
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	port := os.Getenv("BACKEND_PORT")
	r.Run(fmt.Sprintf(":%s", port))
}
