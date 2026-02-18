package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleSetupPatternHooks(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "dry run with default patterns",
			params: map[string]interface{}{
				"action":  "patterns",
				"dry_run": true,
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

				if dryRun, ok := data["dry_run"].(bool); !ok || !dryRun {
					t.Error("expected dry_run=true")
				}

				if configured, ok := data["patterns_configured"].([]interface{}); !ok || len(configured) == 0 {
					t.Error("expected patterns_configured in result")
				}
			},
		},
		{
			name: "install with default patterns",
			params: map[string]interface{}{
				"action":  "patterns",
				"install": true,
				"dry_run": false,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				if status, ok := data["status"].(string); !ok || status != "success" {
					t.Errorf("expected status=success, got %v", data["status"])
				}
				// Check if config file was created
				if configFile, ok := data["config_file"].(string); ok {
					if _, err := os.Stat(configFile); err != nil {
						t.Errorf("config file not created: %v", err)
					}
				}
			},
		},
		{
			name: "with custom patterns",
			params: map[string]interface{}{
				"action":  "patterns",
				"install": true,
				"dry_run": false,
				"patterns": map[string]interface{}{
					"file_changes": map[string]interface{}{
						"*.go": map[string]interface{}{
							"tools": []string{"lint", "test"},
						},
					},
				},
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}
				// Should have configured custom patterns
			},
		},
		{
			name: "uninstall",
			params: map[string]interface{}{
				"action":  "patterns",
				"install": false,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON: %v", err)
					return
				}

				if status, ok := data["status"].(string); !ok || status != "uninstalled" {
					t.Errorf("expected status=uninstalled, got %v", data["status"])
				}
			},
		},
	}

	// Set PROJECT_ROOT for tests
	oldRoot := os.Getenv("PROJECT_ROOT")

	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Setenv("PROJECT_ROOT", oldRoot)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleSetupPatternHooks(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSetupPatternHooks() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestGetDefaultPatterns(t *testing.T) {
	patterns := getDefaultPatterns()
	if len(patterns) == 0 {
		t.Error("expected non-empty default patterns")
	}
	// Check for expected categories
	expectedCategories := []string{"file_changes", "git_events", "task_status"}
	for _, category := range expectedCategories {
		if _, ok := patterns[category]; !ok {
			t.Errorf("expected category %s in default patterns", category)
		}
	}
}

func TestHandleSetupHooksNative(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "git action",
			params: map[string]interface{}{
				"action": "git",
			},
			wantError: false,
		},
		{
			name: "patterns action",
			params: map[string]interface{}{
				"action":  "patterns",
				"dry_run": true,
			},
			wantError: false,
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

			result, err := handleSetupHooksNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleSetupHooksNative() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}
