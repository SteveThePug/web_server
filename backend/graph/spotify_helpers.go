package graph

import (
	"adam-french.co.uk/backend/graph/model"
	"github.com/zmb3/spotify/v2"
)

func mapSpotifyImages(images []spotify.Image) []*model.SpotifyImage {
	result := make([]*model.SpotifyImage, len(images))
	for i, img := range images {
		result[i] = &model.SpotifyImage{URL: img.URL}
	}
	return result
}

func mapSpotifyTrack(track *spotify.FullTrack) *model.SpotifyTrack {
	artists := make([]*model.SpotifyArtist, len(track.Artists))
	for i, a := range track.Artists {
		artists[i] = &model.SpotifyArtist{Name: a.Name}
	}
	return &model.SpotifyTrack{
		Name:    track.Name,
		Artists: artists,
		Album: &model.SpotifyAlbum{
			Name:   track.Album.Name,
			Images: mapSpotifyImages(track.Album.Images),
		},
	}
}

func mapRecentItems(items []spotify.RecentlyPlayedItem) []*model.SpotifyRecentItem {
	result := make([]*model.SpotifyRecentItem, len(items))
	for i, item := range items {
		artists := make([]*model.SpotifyArtist, len(item.Track.Artists))
		for j, a := range item.Track.Artists {
			artists[j] = &model.SpotifyArtist{Name: a.Name}
		}
		result[i] = &model.SpotifyRecentItem{
			PlayedAt: item.PlayedAt,
			Track: &model.SpotifyTrack{
				Name:    item.Track.Name,
				Artists: artists,
				Album: &model.SpotifyAlbum{
					Name:   item.Track.Album.Name,
					Images: mapSpotifyImages(item.Track.Album.Images),
				},
			},
		}
	}
	return result
}
