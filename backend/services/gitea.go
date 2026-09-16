package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// This file talks to the self-hosted Gitea instance over the private Docker
// network to surface "what did Adam push most recently" on the home page.
// The response is cached by the caller (handlers.Store, 1 minute TTL).

// GiteaFeedCommit is one commit inside a push event's Content payload.
// The capitalised JSON keys are Gitea's, not a typo.
type GiteaFeedCommit struct {
	Message string `json:"Message"`
}

// GiteaFeedContent is the shape of the JSON *string* stored in
// GiteaFeedResponse.Content for push events; see ParseCommitMessage.
type GiteaFeedContent struct {
	Commits []GiteaFeedCommit `json:"Commits"`
}

// GiteaFeedUser is the acting user on a feed entry.
type GiteaFeedUser struct {
	AvatarURL string `json:"avatar_url"`
}

// GiteaFeedRepo is the repository a feed entry belongs to.
type GiteaFeedRepo struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

// GiteaFeedResponse is a single Gitea activity feed entry. Only the fields
// the front end renders are declared; the rest of the payload is discarded.
type GiteaFeedResponse struct {
	ActUser GiteaFeedUser `json:"act_user"`
	Repo    GiteaFeedRepo `json:"repo"`
	OpType  string        `json:"op_type"`
	Content string        `json:"content"`
	Created time.Time     `json:"created"`
}

// FetchLatestFeed returns the single most recent public activity entry for
// user "adamf", or (nil, nil) when the feed is empty — callers must handle
// that nil-without-error case.
//
// The username is hard-coded and the call is plain HTTP with no token: this
// only ever reaches the Gitea container over the internal Docker network,
// where the public feed needs no authentication.
func FetchLatestFeed(host, port string) (*GiteaFeedResponse, error) {
	url := fmt.Sprintf("http://%s:%s/api/v1/users/adamf/activities/feeds?limit=1", host, port)

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

	var items []GiteaFeedResponse
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	return &items[0], nil
}

// ParseCommitMessage pulls the first commit message out of a feed entry's
// Content field, which Gitea delivers as JSON encoded inside a JSON string
// (so it needs a second Unmarshal). Non-push events have Content in a
// different shape entirely, so any parse failure is intentionally swallowed
// and reported as "no commit message".
func ParseCommitMessage(content string) string {
	var c GiteaFeedContent
	if err := json.Unmarshal([]byte(content), &c); err != nil {
		return ""
	}
	if len(c.Commits) == 0 {
		return ""
	}
	return c.Commits[0].Message
}
