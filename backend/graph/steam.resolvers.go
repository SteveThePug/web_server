package graph

import (
	"context"
	"fmt"
	"time"

	"adam-french.co.uk/backend/graph/model"
	"adam-french.co.uk/backend/services"
)

func mapSteamGames(games []services.SteamRecentGame) []*model.SteamGame {
	result := make([]*model.SteamGame, len(games))
	for i, g := range games {
		result[i] = &model.SteamGame{
			AppID:           g.AppID,
			Name:            g.Name,
			Playtime2Weeks:  g.Playtime2Weeks,
			PlaytimeForever: g.PlaytimeForever,
			HeaderImageURL:  fmt.Sprintf("https://cdn.akamai.steamstatic.com/steam/apps/%d/header.jpg", g.AppID),
		}
	}
	return result
}

// SteamStatus is the resolver for the steamStatus field.
func (r *queryResolver) SteamStatus(ctx context.Context) (*model.SteamStatus, error) {
	if r.Store.SteamAPIKey == "" {
		return nil, nil
	}

	if r.Store.SteamFresh() {
		return &model.SteamStatus{
			Online:      r.Store.SteamOnline,
			RecentGames: mapSteamGames(r.Store.SteamRecentGames),
		}, nil
	}

	games, err := services.FetchRecentlyPlayedGames(r.Store.SteamAPIKey, r.Store.SteamID)
	if err != nil {
		return nil, err
	}

	summary, err := services.FetchPlayerSummary(r.Store.SteamAPIKey, r.Store.SteamID)
	if err != nil {
		return nil, err
	}

	online := false
	if summary != nil {
		online = summary.PersonaState > 0
	}

	r.Store.SteamRecentGames = games
	r.Store.SteamOnline = online
	r.Store.SteamFetchedAt = time.Now()

	return &model.SteamStatus{
		Online:      online,
		RecentGames: mapSteamGames(games),
	}, nil
}
