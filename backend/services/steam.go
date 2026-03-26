package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SteamRecentGame struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	Playtime2Weeks  int    `json:"playtime_2weeks"`
	PlaytimeForever int    `json:"playtime_forever"`
}

type SteamRecentGamesResponse struct {
	Response struct {
		TotalCount int               `json:"total_count"`
		Games      []SteamRecentGame `json:"games"`
	} `json:"response"`
}

type SteamPlayerSummary struct {
	PersonaState int `json:"personastate"`
}

type SteamPlayerSummariesResponse struct {
	Response struct {
		Players []SteamPlayerSummary `json:"players"`
	} `json:"response"`
}

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
