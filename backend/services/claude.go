package services

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeConfig holds the Anthropic API credentials.
type ClaudeConfig struct {
	APIKey string
}

// InitClaude builds the Anthropic client used to classify job-application
// emails (EmailSyncService.processEmail). The other Claude caller, the rowing
// photo reader, lives in the Python service and builds its own client from
// the same CLAUDE_API_KEY.
//
// No validation happens here: with an empty or wrong API key construction
// still succeeds and the failure only surfaces on the first request.
func InitClaude(config *ClaudeConfig) *anthropic.Client {
	// anthropic.NewClient returns a value, not a pointer, so take its address
	// so it can be handed around without copying.
	client := anthropic.NewClient(option.WithAPIKey(config.APIKey))
	return &client
}
