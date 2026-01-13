package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleWorkflowModeNative(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "focus action",
			params: map[string]interface{}{
				"action": "focus",
				"mode":   "development",
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
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true")
				}
			},
		},
		{
			name: "suggest action",
			params: map[string]interface{}{
				"action": "suggest",
				"text":   "I need to implement a new feature",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should suggest a workflow mode
			},
		},
		{
			name: "stats action",
			params: map[string]interface{}{
				"action": "stats",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				if success, ok := data["success"].(bool); !ok || !success {
					t.Error("expected success=true")
				}
			},
		},
		{
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantError: false, // Returns error response, not an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleWorkflowModeNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleWorkflowModeNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleWorkflowMode(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "focus action",
			params: map[string]interface{}{
				"action": "focus",
				"mode":   "development",
			},
			wantError: false,
		},
		{
			name: "suggest action",
			params: map[string]interface{}{
				"action": "suggest",
			},
			wantError: false,
		},
		{
			name: "stats action",
			params: map[string]interface{}{
				"action": "stats",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleWorkflowMode(ctx, json.RawMessage(argsJSON))
			if (err != nil) != tt.wantError {
				t.Errorf("handleWorkflowMode() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
