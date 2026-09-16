package graph

import (
	"adam-french.co.uk/backend/graph/model"
	"adam-french.co.uk/backend/services"
)

// mapGiteaFeed flattens a Gitea feed entry into the GraphQL shape, pulling the
// commit message out of the doubly-encoded Content field as it goes. It
// dereferences feed, so callers must rule out the (nil, nil) that
// FetchLatestFeed returns for an empty feed.
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
