// Package tools: LocalAI as a TextGenerator (OpenAI-compatible API).
// Optional backend for self-hosted LocalAI alongside Ollama/FM/MLX.
// Base URL from env LOCALAI_BASE_URL; optional LOCALAI_MODEL. See docs/GO_AI_ECOSYSTEM.md §5.2.

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

const localAIEnvBaseURL = "LOCALAI_BASE_URL"
const localAIEnvModel = "LOCALAI_MODEL"
const localAIDefaultModel = "gpt-3.5-turbo"
const localAIDefaultTimeout = 120 * time.Second

// localaiTextGenerator implements TextGenerator using LocalAI's OpenAI-compatible /v1/chat/completions.
type localaiTextGenerator struct{}

// DefaultLocalAI is the shared LocalAI provider for text_generate (provider=localai).
var DefaultLocalAI TextGenerator = &localaiTextGenerator{}

// DefaultLocalAIProvider returns the default LocalAI provider (implements TextGenerator).
func DefaultLocalAIProvider() TextGenerator {
	return DefaultLocalAI
}

// Supported returns true when LOCALAI_BASE_URL is set (non-empty).
func (*localaiTextGenerator) Supported() bool {
	return strings.TrimSpace(os.Getenv(localAIEnvBaseURL)) != ""
}

// Generate sends the prompt to LocalAI /v1/chat/completions and returns the first choice content.
func (*localaiTextGenerator) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	baseURL := strings.TrimSpace(os.Getenv(localAIEnvBaseURL))
	if baseURL == "" {
		return "", fmt.Errorf("LOCALAI_BASE_URL is not set")
	}
	model := strings.TrimSpace(os.Getenv(localAIEnvModel))
	if model == "" {
		model = localAIDefaultModel
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
		return "", fmt.Errorf("localai marshal request: %w", err)
	}

	url := strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("localai create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: localAIDefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("localai API request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("localai API status %d: %s", resp.StatusCode, string(body))
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("localai decode response: %w", err)
	}
	if len(out.Choices) == 0 || out.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("localai empty response")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}
