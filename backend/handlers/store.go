package handlers

import (
	"github.com/zmb3/spotify/v2"
	"gorm.io/gorm"
)

type Store struct {
	DB            *gorm.DB
	SpotifyClient *spotify.Client
}
