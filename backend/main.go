package main

import (
	"context"
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

// Command backend is the Go API for adam-french.co.uk.
//
// main is the single place everything is wired together: it reads every
// setting from the environment, builds the services, packs them into one
// handlers.Store, and registers the routes. There are three route groups —
// open, `protected` (valid access token) and `admin` (token with admin=true)
// — plus the GraphQL endpoint, which does its own per-resolver authorisation
// instead of using middleware. See backend/README.md for the wider picture.
func main() {
	// Logs go to both stdout (so `docker compose logs` works) and a file in a
	// mounted volume (so they survive the container).
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

	// Only the Docker network nginx sits on may set X-Forwarded-For. Without
	// this pin any client could spoof its IP and defeat the per-IP login rate
	// limiter, which keys on ctx.ClientIP(). The CIDR must match the network
	// defined in docker-compose.yml.
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
	// Dev only; the seeder creates an admin with a known password.
	if os.Getenv("SEED_DB") == "true" {
		services.SeedDatabase(db)
	}
	// DOMAIN does double duty: the WebSocket origin allow-list and the auth
	// cookie domain.
	domainName := os.Getenv("DOMAIN")
	// The chat hub is a package-level singleton rather than a value on the
	// Store, so it is initialised rather than constructed.
	services.InitWebSocket(db, domainName)

	// SPOTIFY
	// No SPOTIFY_AUTH_STATE: the OAuth `state` is now a one-shot nonce minted
	// per authorisation request, so there is nothing to configure.
	spotifyRedirectURL := os.Getenv("SPOTIFY_REDIRECT_URI")
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyConfig := services.SpotifyConfig{RedirectURL: spotifyRedirectURL, ClientID: spotifyClientID, ClientSecret: spotifyClientSecret}
	spotifyAuth, spotifyClient := services.InitSpotifyAuth(&spotifyConfig)

	// CLAUDE
	claudeAPIKey := os.Getenv("CLAUDE_API_KEY")
	claudeConfig := services.ClaudeConfig{APIKey: claudeAPIKey}
	claudeClient := services.InitClaude(&claudeConfig)

	authSecret := os.Getenv("BACKEND_SECRET")
	backendEndpoint := os.Getenv("BACKEND_ENDPOINT")
	// Long-lived by design: this is a personal site with a single admin, and
	// there is no server-side revocation, so a leaked token stays valid for
	// its whole lifetime.
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

	// EMAIL SYNC
	// An unparseable EMAIL_SYNC_INTERVAL is silently ignored and the 30
	// minute default stands — a typo here fails quietly.
	emailSyncInterval := 30 * time.Minute
	if interval := os.Getenv("EMAIL_SYNC_INTERVAL"); interval != "" {
		if parsed, err := time.ParseDuration(interval); err == nil {
			emailSyncInterval = parsed
		}
	}
	emailSyncConfig := services.EmailSyncConfig{
		Backend:      os.Getenv("EMAIL_BACKEND"),
		ClientID:     os.Getenv("MSGRAPH_CLIENT_ID"),
		ClientSecret: os.Getenv("MSGRAPH_CLIENT_SECRET"),
		TenantID:     os.Getenv("MSGRAPH_TENANT_ID"),
		RedirectURI:  os.Getenv("MSGRAPH_REDIRECT_URI"),
		IMAP: &services.IMAPConfig{
			Host:     os.Getenv("IMAP_HOST"),
			Port:     os.Getenv("IMAP_PORT"),
			Email:    os.Getenv("IMAP_EMAIL"),
			Password: os.Getenv("IMAP_PASSWORD"),
		},
		SyncInterval: emailSyncInterval,
		Enabled:      os.Getenv("EMAIL_SYNC_ENABLED") == "true",
	}
	emailSync := services.InitEmailSync(&emailSyncConfig, db, claudeClient)

	// 5 login attempts per IP per minute, applied by both the REST handler
	// and the GraphQL login mutation. Nginx rate-limits the login route too.
	loginLimiter := services.NewRateLimiter(5, time.Minute)

	store := handlers.Store{DB: db, SpotifyAuth: spotifyAuth, SpotifyClient: spotifyClient, ClaudeClient: claudeClient, Auth: auth, Notes: notes, LoginLimiter: loginLimiter, EmailSync: emailSync, GiteaHost: giteaHost, GiteaPort: giteaPort, SteamAPIKey: steamAPIKey, SteamID: steamID}

	// Route groups. AdminMiddleware reads the claims AuthMiddlewear stored in
	// the Gin context, so the order in the admin group matters.
	//
	// Paths below have no /api prefix: nginx strips it before proxying.
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

	// EMAIL SYNC
	r.GET("/email/callback", store.CompleteEmailAuth)
	admin.POST("/email/sync", store.TriggerEmailSync)

	// SPOTIFY
	// The callback must be open — Spotify redirects a logged-out browser to
	// it — and is protected by the state nonce instead. Starting the flow is
	// admin-only, since that is what mints the nonce.
	r.GET("/spotify/callback", store.CompleteSpotifyAuth)
	admin.GET("/spotify/auth", store.StartSpotifyAuth)
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
	// /ws is deliberately outside the protected group: chat is public, and
	// the handler checks the cookies itself to decide admin privileges.
	r.GET("/ws", store.ConnectWebSocket)
	protected.POST("/messages/upload", store.UploadMessageFile)

	// NOTES
	admin.GET("/notes/*path", store.GetNoteFile)

	// GRAPHQL
	// GraphQL is assembled by hand rather than with handler.NewDefaultServer
	// so the transports and extensions below are explicit.
	//
	// Note the whole API is one POST route with no middleware guard:
	// authorisation happens per resolver via IsAdminFromCtx / UserIDFromCtx,
	// so a new resolver is PUBLIC until it checks for itself.
	gqlSrv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{Store: &store},
	}))
	gqlSrv.AddTransport(transport.Websocket{KeepAlivePingInterval: 10 * time.Second})
	gqlSrv.AddTransport(transport.Options{})
	gqlSrv.AddTransport(transport.POST{})
	gqlSrv.AddTransport(transport.MultipartForm{})
	// Caches parsed query documents so repeated queries skip parsing.
	gqlSrv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	// Rejects queries above a computed complexity, the standard defence
	// against a deeply nested query being used to exhaust the server.
	gqlSrv.Use(extension.FixedComplexityLimit(200))
	// Introspection and the playground are both off unless DEV_MODE *and*
	// their own flag are set, so production does not publish its schema.
	devMode := os.Getenv("DEV_MODE") == "true"
	if devMode && os.Getenv("GQL_INTROSPECTION") == "true" {
		gqlSrv.Use(extension.Introspection{})
	}
	// AuthContextMiddleware copies the Gin context and any verified claims
	// into the request context, which is what resolvers read. The closure is
	// needed because gqlgen speaks net/http, not gin.
	r.POST("/graphql", graph.AuthContextMiddleware(auth, db), func(c *gin.Context) {
		gqlSrv.ServeHTTP(c.Writer, c.Request)
	})
	if devMode && os.Getenv("GQL_PLAYGROUND") == "true" {
		r.GET("/graphql", func(c *gin.Context) {
			playground.Handler("GraphQL Playground", "/graphql").ServeHTTP(c.Writer, c.Request)
		})
	}

	// HELLO WORLD
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	// The scheduler runs until ctx is cancelled, which in practice only
	// happens if r.Run returns — there is no signal handling or graceful
	// shutdown, so the container is simply killed.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go store.EmailSync.StartScheduler(ctx)

	port := os.Getenv("BACKEND_PORT")
	// Blocks until the server stops. Any return is a failure — a port clash,
	// most often — so it must exit non-zero, or Docker sees a clean shutdown
	// and reports the container as having succeeded.
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("http server stopped: %v", err)
	}
}
