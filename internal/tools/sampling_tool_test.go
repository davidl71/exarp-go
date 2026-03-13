// Package tools provides MCP tools for exarp-go.
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

type mockSampler struct {
	result framework.CreateMessageResult
	err    error
}

func (m *mockSampler) CreateMessage(ctx context.Context, params framework.CreateMessageParams) (framework.CreateMessageResult, error) {
	return m.result, m.err
}

func TestHandleAskClient_NoPrompt(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{}

	_, err := handleAskClient(ctx, params)
	if !errors.Is(err, errPromptRequired) {
		t.Errorf("expected errPromptRequired, got %v", err)
	}
}

func TestHandleAskClient_NoSampler(t *testing.T) {
	ctx := context.Background()
	params := map[string]interface{}{
		"prompt": "test prompt",
	}

	_, err := handleAskClient(ctx, params)
	if !errors.Is(err, errSamplingNotSupported) {
		t.Errorf("expected errSamplingNotSupported, got %v", err)
	}
}

func TestHandleAskClient_Success(t *testing.T) {
	ctx := framework.ContextWithSampler(context.Background(), &mockSampler{
		result: framework.CreateMessageResult{
			Content:    "Hello from client LLM",
			Model:      "claude-3",
			StopReason: "end_turn",
		},
	})
	params := map[string]interface{}{
		"prompt": "Say hello",
	}

	result, err := handleAskClient(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	if result[0].Text != "Hello from client LLM" {
		t.Errorf("expected 'Hello from client LLM', got '%s'", result[0].Text)
	}
}

func TestHandleAskClient_WithOptions(t *testing.T) {
	ctx := framework.ContextWithSampler(context.Background(), &mockSampler{
		result: framework.CreateMessageResult{
			Content: "Custom response",
		},
	})
	params := map[string]interface{}{
		"prompt":        "Test prompt",
		"temperature":   0.9,
		"max_tokens":    100,
		"system_prompt": "You are helpful",
	}

	result, err := handleAskClient(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	if result[0].Text != "Custom response" {
		t.Errorf("expected 'Custom response', got '%s'", result[0].Text)
	}
}

func TestHandleAskClient_SamplerError(t *testing.T) {
	ctx := framework.ContextWithSampler(context.Background(), &mockSampler{
		err: errors.New("sampler unavailable"),
	})
	params := map[string]interface{}{
		"prompt": "test",
	}

	_, err := handleAskClient(ctx, params)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHandleAskClientWrapper(t *testing.T) {
	ctx := framework.ContextWithSampler(context.Background(), &mockSampler{
		result: framework.CreateMessageResult{
			Content: "Wrapped response",
		},
	})

	args := json.RawMessage(`{"prompt": "test prompt", "temperature": 0.5, "max_tokens": 256}`)

	result, err := handleAskClientWrapper(ctx, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	if result[0].Text != "Wrapped response" {
		t.Errorf("expected 'Wrapped response', got '%s'", result[0].Text)
	}
}

func TestHandleAskClientWrapper_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	args := json.RawMessage(`{invalid`)

	_, err := handleAskClientWrapper(ctx, args)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
