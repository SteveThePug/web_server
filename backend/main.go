package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"

	"adam-french.co.uk/backend/graph"
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

	if os.Getenv("DEV_MODE") != "true" {
		gin.SetMode(gin.ReleaseMode)
	}

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
	if os.Getenv("SEED_DB") == "true" {
		services.SeedDatabase(db)
	}
	domainName := os.Getenv("DOMAIN")
	services.InitWebSocket(db, domainName)

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
	backendEndpoint := os.Getenv("BACKEND_ENDPOINT")
	accessTokenLifetime := 7 * 24 * time.Hour
	refreshTokenLifetime := 365 * 24 * time.Hour
	authConfig := services.AuthConfig{Secret: []byte(authSecret), Domain: domainName, RefreshTokenLifetime: refreshTokenLifetime, AccessTokenLifetime: accessTokenLifetime, Endpoint: backendEndpoint}
	auth := services.InitAuth(&authConfig)

	notesDir := "/backend/notes"
	notesConfig := services.NotesConfig{Dir: notesDir}
	notes := services.InitNotes(&notesConfig)

	giteaHost := os.Getenv("GITEA_HOST")
	giteaPort := os.Getenv("GITEA_PORT")

	steamAPIKey := os.Getenv("STEAM_API_KEY")
	steamID := os.Getenv("STEAM_ID")

	store := handlers.Store{DB: db, SpotifyAuth: spotifyAuth, SpotifyClient: spotifyClient, ClaudeClient: claudeClient, Auth: auth, Notes: notes, GiteaHost: giteaHost, GiteaPort: giteaPort, SteamAPIKey: steamAPIKey, SteamID: steamID}

	protected := r.Group("/", store.AuthMiddlewear)
	admin := r.Group("/", store.AuthMiddlewear, store.AdminMiddleware)

	// ROWING
	r.GET("/rowing", store.GetRowing)
	admin.POST("/rowing", store.CreateRowing)

	// AUTH
	r.POST("/auth/login", store.Login)
	r.POST("/auth/refresh", store.RefreshToken)
	r.GET("/auth/check", store.CheckToken)
	r.POST("/auth/logout", store.Logout)
	r.GET("/auth/validate-admin", store.ValidateAdmin)

	// SPOTIFY
	r.GET("/spotify/callback", store.CompleteSpotifyAuth)
	r.GET("/spotify/listening", store.ListeningTo)
	r.GET("/spotify/recent", store.RecentlyPlayed)
	// r.POST("/spotify", store.SendSong)

	// RADIO
	admin.POST("/radio/upload", store.UploadRadioSong)
	admin.GET("/radio/songs", store.ListRadioSongs)
	admin.DELETE("/radio/songs/:filename", store.DeleteRadioSong)
	admin.PATCH("/radio/songs/:filename/disable", store.DisableRadioSong)
	admin.PATCH("/radio/songs/:filename/enable", store.EnableRadioSong)

	// MESSAGES
	r.GET("/ws", store.ConnectWebSocket)
	protected.POST("/messages/upload", store.UploadMessageFile)

	// NOTES
	r.GET("/notes/*path", store.GetNoteFile)

	// GRAPHQL
	gqlSrv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{Store: &store},
	}))
	gqlSrv.AddTransport(transport.Websocket{KeepAlivePingInterval: 10 * time.Second})
	gqlSrv.AddTransport(transport.Options{})
	gqlSrv.AddTransport(transport.GET{})
	gqlSrv.AddTransport(transport.POST{})
	gqlSrv.AddTransport(transport.MultipartForm{})
	gqlSrv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	gqlSrv.Use(extension.FixedComplexityLimit(200))
	if os.Getenv("GQL_INTROSPECTION") == "true" {
		gqlSrv.Use(extension.Introspection{})
	}
	r.POST("/graphql", graph.AuthContextMiddleware(auth), func(c *gin.Context) {
		gqlSrv.ServeHTTP(c.Writer, c.Request)
	})
	if os.Getenv("GQL_PLAYGROUND") == "true" {
		r.GET("/graphql", func(c *gin.Context) {
			playground.Handler("GraphQL Playground", "/graphql").ServeHTTP(c.Writer, c.Request)
		})
	}

	// HELLO WORLD
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	port := os.Getenv("BACKEND_PORT")
	r.Run(fmt.Sprintf(":%s", port))
}
