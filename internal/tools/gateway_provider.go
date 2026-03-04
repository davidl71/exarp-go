// Package tools: OpenAI-compatible gateway as a TextGenerator.
// Optional backend for any OpenAI-compatible router (inference-gateway, pLLM, kcolemangt/llm-router, radlab llm-router, RouteLLM).
// Base URL from env OPENAI_GATEWAY_BASE_URL; optional OPENAI_GATEWAY_MODEL and OPENAI_GATEWAY_API_KEY.
// See docs/research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md §6.5.

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	gatewayEnvBaseURL     = "OPENAI_GATEWAY_BASE_URL"
	gatewayEnvModel       = "OPENAI_GATEWAY_MODEL"
	gatewayEnvAPIKey      = "OPENAI_GATEWAY_API_KEY"
	gatewayDefaultModel   = "gpt-3.5-turbo"
	gatewayDefaultTimeout = 120 * time.Second
)

// gatewayTextGenerator implements TextGenerator using an OpenAI-compatible /v1/chat/completions endpoint.
type gatewayTextGenerator struct{}

// DefaultGateway is the shared gateway provider for text_generate (provider=gateway).
var DefaultGateway TextGenerator = &gatewayTextGenerator{}

// DefaultGatewayProvider returns the default gateway provider (implements TextGenerator).
func DefaultGatewayProvider() TextGenerator {
	return DefaultGateway
}

// Supported returns true when OPENAI_GATEWAY_BASE_URL is set (non-empty).
func (*gatewayTextGenerator) Supported() bool {
	return strings.TrimSpace(os.Getenv(gatewayEnvBaseURL)) != ""
}

// Generate sends the prompt to the gateway /v1/chat/completions and returns the first choice content.
func (g *gatewayTextGenerator) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	return g.generateWithModel(ctx, prompt, maxTokens, temperature, "")
}

// GenerateWithModel sends the prompt with an explicit model override (for RouteLLM router-mf-*, kcolemangt prefix routing, etc.).
// If modelOverride is empty, env OPENAI_GATEWAY_MODEL or default is used.
func (g *gatewayTextGenerator) GenerateWithModel(ctx context.Context, prompt string, maxTokens int, temperature float32, modelOverride string) (string, error) {
	return g.generateWithModel(ctx, prompt, maxTokens, temperature, modelOverride)
}

func (g *gatewayTextGenerator) generateWithModel(ctx context.Context, prompt string, maxTokens int, temperature float32, modelOverride string) (string, error) {
	baseURL := strings.TrimSpace(os.Getenv(gatewayEnvBaseURL))
	if baseURL == "" {
		return "", fmt.Errorf("OPENAI_GATEWAY_BASE_URL is not set")
	}
	model := strings.TrimSpace(modelOverride)
	if model == "" {
		model = strings.TrimSpace(os.Getenv(gatewayEnvModel))
	}
	if model == "" {
		model = gatewayDefaultModel
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  maxTokens,
		"temperature": float64(temperature),
		"stream":      false,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gateway marshal request: %w", err)
	}

	url := strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("gateway create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := strings.TrimSpace(os.Getenv(gatewayEnvAPIKey)); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	client := &http.Client{Timeout: gatewayDefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gateway API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gateway API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("gateway decode response: %w", err)
	}
	if len(out.Choices) == 0 || out.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("gateway empty response")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
