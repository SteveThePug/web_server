package graph

import (
	"context"
	"time"

	"adam-french.co.uk/backend/graph/model"
	"adam-french.co.uk/backend/services"
)

func mapGiteaFeed(feed *services.GiteaFeedResponse) *model.GiteaFeedItem {
	return &model.GiteaFeedItem{
		AvatarURL:     feed.ActUser.AvatarURL,
		RepoURL:       feed.Repo.HTMLURL,
		RepoName:      feed.Repo.FullName,
		OpType:        feed.OpType,
		CommitMessage: services.ParseCommitMessage(feed.Content),
		CreatedAt:     feed.Created,
	}
}

// GiteaFeed is the resolver for the giteaFeed field.
func (r *queryResolver) GiteaFeed(ctx context.Context) (*model.GiteaFeedItem, error) {
	if r.Store.GiteaFeedFresh() {
		return mapGiteaFeed(r.Store.GiteaFeed), nil
	}

	feed, err := services.FetchLatestFeed(r.Store.GiteaHost, r.Store.GiteaPort)
	if err != nil {
		return nil, err
	}
	if feed == nil {
		return nil, nil
	}

	r.Store.GiteaFeed = feed
	r.Store.GiteaFeedFetchedAt = time.Now()

	return mapGiteaFeed(feed), nil
}
