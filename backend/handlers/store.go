package handlers

import (
	"time"

	"adam-french.co.uk/backend/services"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"gorm.io/gorm"
)

type Store struct {
	DB            *gorm.DB
	SpotifyAuth   *spotifyauth.Authenticator
	SpotifyClient *spotify.Client
	ClaudeClient  *anthropic.Client
	Auth          *services.Auth
	Notes         *services.Notes

	RecentSongs          *[]spotify.RecentlyPlayedItem
	RecentSongsFetchedAt time.Time
}
