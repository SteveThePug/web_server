package graph

import (
	"context"
	"time"

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

// SpotifyListening is the resolver for the spotifyListening field.
func (r *queryResolver) SpotifyListening(ctx context.Context) (*model.SpotifyPlaying, error) {
	if r.Store.SpotifyClient == nil {
		return nil, nil
	}

	playing, err := r.Store.SpotifyClient.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return nil, err
	}

	result := &model.SpotifyPlaying{Playing: playing.Playing}
	if playing.Item != nil {
		result.Track = mapSpotifyTrack(playing.Item)
	}

	return result, nil
}

// SpotifyRecent is the resolver for the spotifyRecent field.
func (r *queryResolver) SpotifyRecent(ctx context.Context) ([]*model.SpotifyRecentItem, error) {
	if r.Store.SpotifyClient == nil {
		return []*model.SpotifyRecentItem{}, nil
	}

	if r.Store.RecentSongsFresh() {
		return mapRecentItems(*r.Store.RecentSongs), nil
	}

	opts := spotify.RecentlyPlayedOptions{Limit: 3}
	played, err := r.Store.SpotifyClient.PlayerRecentlyPlayedOpt(ctx, &opts)
	if err != nil {
		return nil, err
	}

	r.Store.RecentSongs = &played
	r.Store.RecentSongsFetchedAt = time.Now()

	return mapRecentItems(played), nil
}
