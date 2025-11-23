package handlers

import (
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Store struct {
	DB           *gorm.DB
	SpotifyAuth  *spotifyauth.Authenticator
	SpotifyToken *oauth2.Token
}
