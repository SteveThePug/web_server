package graph

import (
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
