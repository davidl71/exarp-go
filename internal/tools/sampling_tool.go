// Package tools provides MCP tools for exarp-go.
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

var (
	errPromptRequired       = errors.New("prompt is required")
	errSamplingNotSupported = errors.New("sampling not supported by client")
)

const (
	samplingDefaultTemperature = 0.7
	samplingDefaultMaxTokens   = 512
)

func handleAskClient(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	prompt, ok := params["prompt"].(string)
	if !ok || prompt == "" {
		return nil, errPromptRequired
	}

	sampler := framework.SamplerFromContext(ctx)
	if sampler == nil {
		return nil, errSamplingNotSupported
	}

	temperature := samplingDefaultTemperature
	if t, ok := params["temperature"].(float64); ok {
		temperature = t
	}

	maxTokens := samplingDefaultMaxTokens
	if m, ok := params["max_tokens"].(int); ok {
		maxTokens = m
	}

	systemPrompt := ""
	if s, ok := params["system_prompt"].(string); ok {
		systemPrompt = s
	}

	result, err := sampler.CreateMessage(ctx, framework.CreateMessageParams{
		Messages: []framework.SamplingMessage{
			{Role: "user", Content: prompt},
		},
		SystemPrompt: systemPrompt,
		Temperature:  temperature,
		MaxTokens:    maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("sampling failed: %w", err)
	}

	return []framework.TextContent{
		{Text: result.Content, Type: "text"},
	}, nil
}

func handleAskClientWrapper(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	framework.ApplyDefaults(params, map[string]interface{}{
		"temperature":   samplingDefaultTemperature,
		"max_tokens":    samplingDefaultMaxTokens,
		"system_prompt": "",
	})

	return handleAskClient(ctx, params)
}
