package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type SpotifyConfig struct {
	AuthState    string
	RedirectURL  string
	ClientID     string
	ClientSecret string
}

const SPOTIFY_TOKEN_JSON_PATH = "/backend/token/spotify_token.json"

func SaveSpotifyToken(path string, tok *oauth2.Token) error {
	data := struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		Expiry       time.Time `json:"expiry"`
	}{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// 0600 ensures only your user can read/write the file
	return os.WriteFile(path, jsonBytes, 0600)
}

func LoadSpotifyToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var saved struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		TokenType    string    `json:"token_type"`
		Expiry       time.Time `json:"expiry"`
	}

	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}

	tok := &oauth2.Token{
		AccessToken:  saved.AccessToken,
		RefreshToken: saved.RefreshToken,
		TokenType:    saved.TokenType,
		Expiry:       saved.Expiry,
	}

	return tok, nil
}

func InitSpotifyAuth(config *SpotifyConfig) (*spotifyauth.Authenticator, *spotify.Client) {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(config.RedirectURL),
		spotifyauth.WithClientID(config.ClientID),
		spotifyauth.WithClientSecret(config.ClientSecret),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
		),
	)

	// check if token exists locally
	token, err := LoadSpotifyToken(SPOTIFY_TOKEN_JSON_PATH)
	if err != nil || token == nil {
		fmt.Println("No token saved. Authenticate Spotify with:")
		fmt.Println(auth.AuthURL(config.AuthState))
		return auth, nil
	}

	// refresh token and client
	client, err := RefreshClient(auth, token)
	if err != nil {
		fmt.Println("Failed to refresh token. Authenticate Spotify with:")
		fmt.Println(auth.AuthURL(config.AuthState))
		return auth, nil
	}

	return auth, client
}

func RefreshClient(auth *spotifyauth.Authenticator, token *oauth2.Token) (*spotify.Client, error) {
	ctx := context.Background()

	token, err := auth.RefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	client := spotify.New(auth.Client(ctx, token))

	return client, nil
}
