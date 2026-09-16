package services

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeConfig holds the Anthropic API credentials.
type ClaudeConfig struct {
	APIKey string
}

// InitClaude builds the shared Anthropic client. It is used in two places:
// reading a rowing machine display from a photo (handlers.CreateRowing) and
// classifying job-application emails (EmailSyncService.processEmail).
//
// No validation happens here: with an empty or wrong API key construction
// still succeeds and the failure only surfaces on the first request.
func InitClaude(config *ClaudeConfig) *anthropic.Client {
	// anthropic.NewClient returns a value, not a pointer, so take its address
	// to share one client across every caller.
	client := anthropic.NewClient(option.WithAPIKey(config.APIKey))
	return &client
}
