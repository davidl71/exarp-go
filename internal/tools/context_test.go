package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleContextBudget(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "context_budget with items",
			params: map[string]interface{}{
				"items": []interface{}{
					"Item 1",
					"Item 2",
					"Item 3",
				},
				"budget_tokens": 4000,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}

				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Native budget returns total_tokens, budget_tokens, items, strategy (no success key)
				if _, ok := data["total_tokens"]; !ok {
					t.Error("expected total_tokens in budget result")
				}

				if _, ok := data["items"]; !ok {
					t.Error("expected items in budget result")
				}
			},
		},
		{
			name: "context_budget with items JSON string",
			params: map[string]interface{}{
				"items":         `["Item 1", "Item 2"]`,
				"budget_tokens": 2000,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
			},
		},
		{
			name: "context_budget missing items",
			params: map[string]interface{}{
				"budget_tokens": 4000,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleContextBudget(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleContextBudget() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleContext(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "context_budget action",
			params: map[string]interface{}{
				"action": "budget",
				"items":  []interface{}{"Item 1", "Item 2"},
			},
			wantError: false,
		},
		{
			name: "context_summarize action",
			params: map[string]interface{}{
				"action": "summarize",
				"data":   "Some text to summarize",
			},
			// Summarize is native-only; returns error when Apple FM not available
			wantError: !FMAvailable(),
		},
		{
			name: "context_batch action",
			params: map[string]interface{}{
				"action": "batch",
				"items":  []interface{}{"Item 1", "Item 2"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "context_summarize action" && !FMAvailable() {
				t.Skip("Apple Foundation Models not available on this platform")
			}
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleContext(ctx, argsJSON)
			// Accept "not available" as non-fatal when summarize runs without FM
			if err != nil && tt.name == "context_summarize action" && strings.Contains(err.Error(), "not available") {
				return // skip success check
			}
			if (err != nil) != tt.wantError {
				t.Errorf("handleContext() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}
