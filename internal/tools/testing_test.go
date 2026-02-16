package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleTestingRun(t *testing.T) {
	// Use temp dir (no go.mod) so native handler returns "only supported for Go projects"
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
			name: "run action with test_path (non-Go project)",
			params: map[string]interface{}{
				"action":    "run",
				"test_path": "./internal/tools",
			},
			wantError: true,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result may be test output or error message
			},
		},
		{
			name: "run action with verbose (non-Go project)",
			params: map[string]interface{}{
				"action":  "run",
				"verbose": true,
			},
			wantError: true,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be test output
			},
		},
		{
			name: "run action with coverage (non-Go project)",
			params: map[string]interface{}{
				"action":   "run",
				"coverage": true,
			},
			wantError: true,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be coverage output
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleTestingRun(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTestingRun() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleTestingCoverage(t *testing.T) {
	// Use temp dir (no go.mod) so native handler returns "only supported for Go projects"
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
			name: "coverage action (non-Go project)",
			params: map[string]interface{}{
				"action": "coverage",
			},
			wantError: true,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
			},
		},
		{
			name: "coverage with format (non-Go project)",
			params: map[string]interface{}{
				"action": "coverage",
				"format": "html",
			},
			wantError: true,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be coverage report
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleTestingCoverage(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTestingCoverage() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleTestingValidate(t *testing.T) {
	// Use temp dir (no go.mod) so native handler returns "only supported for Go projects"
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
			name: "validate action (non-Go project)",
			params: map[string]interface{}{
				"action": "validate",
			},
			wantError: true,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handleTestingValidate(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTestingValidate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleTesting(t *testing.T) {
	// Use temp dir (no go.mod) so native handler returns "only supported for Go projects"
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)

	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "run action (non-Go project)",
			params: map[string]interface{}{
				"action": "run",
			},
			wantError: true,
		},
		{
			name: "coverage action (non-Go project)",
			params: map[string]interface{}{
				"action": "coverage",
			},
			wantError: true,
		},
		{
			name: "validate action (non-Go project)",
			params: map[string]interface{}{
				"action": "validate",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)

			result, err := handleTesting(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleTesting() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && (result == nil || len(result) == 0) {
				t.Error("expected non-empty result")
			}
		})
	}
}
