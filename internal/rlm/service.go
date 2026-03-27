package rlm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	AnthropicClaudeOpus   = anthropic.ModelClaudeOpus4_6
	AnthropicClaudeSonnet = anthropic.ModelClaudeSonnet4_6
	MaxTokens             = 20000
)

// AnthropicLLM implements LLMClient using the Anthropic SDK.
type AnthropicLLM struct {
	client *anthropic.Client
	model  string
}

// AnthropicConfig holds configuration for an Anthropic-backed LLM client.
type AnthropicConfig struct {
	APIKey string
	Model  string // e.g. "claude-sonnet-4-5-20241022"
}

// NewAnthropicLLM creates an LLM client backed by the Anthropic API.
func NewAnthropicLLM(cfg AnthropicConfig) *AnthropicLLM {
	client := anthropic.NewClient(
		option.WithAPIKey(cfg.APIKey),
	)
	return &AnthropicLLM{
		client: &client,
		model:  cfg.Model,
	}
}

// Complete sends messages to the Anthropic API and returns the assistant's text response.
func (a *AnthropicLLM) Complete(ctx context.Context, messages []Message) (string, error) {
	// Separate system message from conversation messages.
	var systemPrompt string
	var chatMessages []anthropic.MessageParam
	for _, m := range messages {
		switch m.Role {
		case "system":
			systemPrompt = m.Content
		case "user":
			chatMessages = append(chatMessages, anthropic.NewUserMessage(
				anthropic.NewTextBlock(m.Content),
			))
		case "assistant":
			chatMessages = append(chatMessages, anthropic.NewAssistantMessage(
				anthropic.NewTextBlock(m.Content),
			))
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: MaxTokens,
		Messages:  chatMessages,
	}
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	resp, err := a.client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("anthropic completion: %w", err)
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("anthropic completion: no text content in response")
}
