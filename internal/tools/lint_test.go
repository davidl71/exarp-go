package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleLint(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		validate  func(*testing.T, []framework.TextContent)
	}{
		{
			name: "lint with golangci-lint",
			params: map[string]interface{}{
				"action": "run",
				"linter": "golangci-lint",
				"path":   ".",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result may be lint output or error message (if linter not available)
			},
		},
		{
			name: "lint with auto detection",
			params: map[string]interface{}{
				"action": "run",
				"linter": "auto",
				"path":   ".",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be lint output or error message
			},
		},
		{
			name: "lint with fix",
			params: map[string]interface{}{
				"action": "run",
				"linter": "golangci-lint",
				"path":   ".",
				"fix":    true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be lint output or error message
			},
		},
		{
			name: "analyze action",
			params: map[string]interface{}{
				"action": "analyze",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be analysis output
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleLint(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleLint() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestRunLinter(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("PROJECT_ROOT", tmpDir)

	tests := []struct {
		name      string
		linter    string
		path      string
		fix       bool
		wantError bool
	}{
		{
			name:      "golangci-lint",
			linter:    "golangci-lint",
			path:      ".",
			fix:       false,
			wantError: false, // May fail if linter not available, but function should handle it
		},
		{
			name:      "unsupported linter",
			linter:    "unsupported-linter",
			path:      ".",
			fix:       false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := runLinter(ctx, tt.linter, tt.path, tt.fix)
			if (err != nil) != tt.wantError {
				t.Errorf("runLinter() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}
