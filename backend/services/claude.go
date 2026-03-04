package services

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeConfig struct {
	APIKey string
}

func InitClaude(config *ClaudeConfig) *anthropic.Client {
	client := anthropic.NewClient(option.WithAPIKey(config.APIKey))
	return &client
}
