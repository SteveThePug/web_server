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

	GiteaHost          string
	GiteaPort          string
	GiteaFeed          *services.GiteaFeedResponse
	GiteaFeedFetchedAt time.Time
}

func (s *Store) GiteaFeedFresh() bool {
	if s.GiteaFeed == nil {
		return false
	}
	return time.Since(s.GiteaFeedFetchedAt) < time.Minute
}
