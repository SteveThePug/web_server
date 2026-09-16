package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// SpotifyConfig holds the OAuth app credentials.
//
// There is no state value here any more: the `state` parameter is a one-shot
// CSRF nonce minted per authorisation request by handlers.Store, so a fixed
// configured value would defeat the point.
type SpotifyConfig struct {
	RedirectURL  string
	ClientID     string
	ClientSecret string
}

// SPOTIFY_TOKEN_JSON_PATH is inside a Docker volume, so the refresh token
// survives container restarts and the OAuth dance only has to be done once.
const SPOTIFY_TOKEN_JSON_PATH = "/backend/token/spotify_token.json"

// SaveSpotifyToken persists an OAuth token to disk.
//
// It copies the fields into a local anonymous struct rather than marshalling
// oauth2.Token directly, so the on-disk format stays stable regardless of what
// the oauth2 package adds to its own struct or its JSON tags.
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

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// 0600 ensures only your user can read/write the file
	return os.WriteFile(path, jsonBytes, 0600)
}

// LoadSpotifyToken reads a token previously written by SaveSpotifyToken.
// A missing file returns an error, which callers treat as "not yet
// authenticated" rather than a fault.
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

// InitSpotifyAuth builds the OAuth authenticator and, if a saved token
// exists, an authenticated client.
//
// Start-up never fails on Spotify problems: when there is no token or the
// refresh is rejected it logs how to authenticate and returns a nil client.
// Every Spotify handler and resolver therefore has to nil-check
// Store.SpotifyClient. The authorisation URL can no longer be printed here
// because its state nonce is per-request (see handlers.Store.StartSpotifyAuth);
// following it sends the browser to /spotify/callback, which fills the client
// in at runtime.
func InitSpotifyAuth(config *SpotifyConfig) (*spotifyauth.Authenticator, *spotify.Client) {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(config.RedirectURL),
		spotifyauth.WithClientID(config.ClientID),
		spotifyauth.WithClientSecret(config.ClientSecret),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserReadRecentlyPlayed,
		),
	)

	// check if token exists locally
	token, err := LoadSpotifyToken(SPOTIFY_TOKEN_JSON_PATH)
	if err != nil || token == nil {
		fmt.Println("No Spotify token saved. Authenticate by calling GET /api/spotify/auth as an admin and opening the URL it returns.")
		return auth, nil
	}

	// refresh token and client
	client, err := RefreshClient(auth, token)
	if err != nil {
		fmt.Println("Failed to refresh Spotify token. Re-authenticate by calling GET /api/spotify/auth as an admin and opening the URL it returns.")
		return auth, nil
	}

	return auth, client
}

// RefreshClient exchanges a stored token for a fresh one and returns a client
// using it.
//
// Spotify may issue a new refresh token during this exchange, so the result is
// written straight back to disk; the write error is ignored because a failed
// save only costs a re-authentication later, and failing the whole start-up
// over it would be worse. The returned client also holds the token source
// internally and keeps refreshing on its own — but those later refreshes are
// NOT persisted, so after a long uptime the file on disk can be stale.
func RefreshClient(auth *spotifyauth.Authenticator, token *oauth2.Token) (*spotify.Client, error) {
	// context.Background rather than a request context: the resulting client
	// is long-lived and must outlive whichever request triggered this.
	ctx := context.Background()

	token, err := auth.RefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}
	SaveSpotifyToken(SPOTIFY_TOKEN_JSON_PATH, token)

	client := spotify.New(auth.Client(ctx, token))

	return client, nil
}
