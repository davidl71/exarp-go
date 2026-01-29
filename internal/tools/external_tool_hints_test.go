package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleAddExternalToolHintsNative(t *testing.T) {
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
			name: "add external tool hints with dry_run",
			params: map[string]interface{}{
				"dry_run":       true,
				"min_file_size": 50,
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
				if status, ok := data["status"].(string); !ok || status != "success" {
					t.Errorf("expected status='success', got %v", data["status"])
				}
			},
		},
		{
			name: "add external tool hints with output_path",
			params: map[string]interface{}{
				"dry_run":     true,
				"output_path": "/tmp/test_hints.md",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handleAddExternalToolHintsNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAddExternalToolHintsNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleAddExternalToolHints(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "add external tool hints",
			params: map[string]interface{}{
				"dry_run": true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleAddExternalToolHints(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleAddExternalToolHints() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
