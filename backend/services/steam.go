package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// This file wraps the two public Steam Web API endpoints used by the home
// page. Results are cached by the caller (handlers.Store, 5 minute TTL) to
// stay well inside Steam's rate limits.

// SteamRecentGame is one game from GetRecentlyPlayedGames. Playtimes are in
// minutes.
type SteamRecentGame struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	Playtime2Weeks  int    `json:"playtime_2weeks"`
	PlaytimeForever int    `json:"playtime_forever"`
}

// SteamRecentGamesResponse mirrors Steam's envelope, which wraps everything
// in a "response" object.
type SteamRecentGamesResponse struct {
	Response struct {
		TotalCount int               `json:"total_count"`
		Games      []SteamRecentGame `json:"games"`
	} `json:"response"`
}

// SteamPlayerSummary is the subset of GetPlayerSummaries we care about.
// PersonaState is Steam's online status enum: 0 = offline, anything higher
// is some flavour of online (online, busy, away, snooze, looking to trade or
// play), which is why the caller treats "> 0" as online.
type SteamPlayerSummary struct {
	PersonaState int `json:"personastate"`
}

// SteamPlayerSummariesResponse mirrors Steam's response envelope.
type SteamPlayerSummariesResponse struct {
	Response struct {
		Players []SteamPlayerSummary `json:"players"`
	} `json:"response"`
}

// FetchRecentlyPlayedGames returns the three most recently played games.
//
// Steam answers with HTTP 200 and an empty envelope for a private profile or
// a bad key, so an empty slice and no error is the normal "nothing to show"
// result rather than a failure.
func FetchRecentlyPlayedGames(apiKey, steamID string) ([]SteamRecentGame, error) {
	url := fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetRecentlyPlayedGames/v1/?key=%s&steamid=%s&count=3", apiKey, steamID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SteamRecentGamesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Response.Games, nil
}

// FetchPlayerSummary returns the profile's online status, or (nil, nil) when
// Steam returns no players (private or unknown profile).
func FetchPlayerSummary(apiKey, steamID string) (*SteamPlayerSummary, error) {
	url := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s", apiKey, steamID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SteamPlayerSummariesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Response.Players) == 0 {
		return nil, nil
	}

	return &result.Response.Players[0], nil
}
