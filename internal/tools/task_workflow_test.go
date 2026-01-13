package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleTaskWorkflowNative(t *testing.T) {
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
			name: "sync action",
			params: map[string]interface{}{
				"action": "sync",
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
			},
		},
		{
			name: "approve action",
			params: map[string]interface{}{
				"action": "approve",
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
			name: "create action",
			params: map[string]interface{}{
				"action":         "create",
				"name":           "Test Task",
				"long_description": "Test description",
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
			name: "unknown action",
			params: map[string]interface{}{
				"action": "unknown",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleTaskWorkflowNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskWorkflowNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleTaskWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "sync action",
			params: map[string]interface{}{
				"action": "sync",
			},
			wantError: false,
		},
		{
			name: "approve action",
			params: map[string]interface{}{
				"action": "approve",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleTaskWorkflow(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTaskWorkflow() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}
