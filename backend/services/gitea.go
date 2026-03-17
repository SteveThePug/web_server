package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GiteaFeedCommit struct {
	Message string `json:"Message"`
}

type GiteaFeedContent struct {
	Commits []GiteaFeedCommit `json:"Commits"`
}

type GiteaFeedUser struct {
	AvatarURL string `json:"avatar_url"`
}

type GiteaFeedRepo struct {
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
}

type GiteaFeedResponse struct {
	ActUser GiteaFeedUser `json:"act_user"`
	Repo    GiteaFeedRepo `json:"repo"`
	OpType  string        `json:"op_type"`
	Content string        `json:"content"`
	Created time.Time     `json:"created"`
}

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
