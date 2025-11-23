package services

import (
	"fmt"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type SpotifyConfig struct {
	AuthState    string
	RedirectURL  string
	ClientID     string
	ClientSecret string
}

func InitSpotifyAuth(config SpotifyConfig) *spotifyauth.Authenticator {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(config.RedirectURL),
		spotifyauth.WithClientID(config.ClientID),
		spotifyauth.WithClientSecret(config.ClientSecret),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
		),
	)

	fmt.Println("Authenticate spotify with:")
	fmt.Println(auth.AuthURL(config.AuthState))

	return auth
}
