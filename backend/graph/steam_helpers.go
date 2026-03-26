package graph

import (
	"fmt"

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
