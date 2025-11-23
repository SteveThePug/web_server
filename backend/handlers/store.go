package handlers

import (
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"gorm.io/gorm"
)

type Store struct {
	DB            *gorm.DB
	SpotifyAuth   *spotifyauth.Authenticator
	SpotifyClient *spotify.Client
}
