package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/davidl71/exarp-go/internal/framework"
)

func TestHandleTestingRun(t *testing.T) {
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
			name: "run action with test_path",
			params: map[string]interface{}{
				"action":    "run",
				"test_path": "./internal/tools",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
				// Result may be test output or error message
			},
		},
		{
			name: "run action with verbose",
			params: map[string]interface{}{
				"action":  "run",
				"verbose": true,
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				// Result may be test output
			},
		},
		{
			name: "run action with coverage",
			params: map[string]interface{}{
				"action":   "run",
				"coverage": true,
			},
			wantError: false,
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
			name: "coverage action",
			params: map[string]interface{}{
				"action": "coverage",
			},
			wantError: false,
			validate: func(t *testing.T, result []framework.TextContent) {
				if len(result) == 0 {
					t.Error("expected non-empty result")
					return
				}
			},
		},
		{
			name: "coverage with format",
			params: map[string]interface{}{
				"action": "coverage",
				"format": "html",
			},
			wantError: false,
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
			name: "validate action",
			params: map[string]interface{}{
				"action": "validate",
			},
			wantError: false,
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
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "run action",
			params: map[string]interface{}{
				"action": "run",
			},
			wantError: false,
		},
		{
			name: "coverage action",
			params: map[string]interface{}{
				"action": "coverage",
			},
			wantError: false,
		},
		{
			name: "validate action",
			params: map[string]interface{}{
				"action": "validate",
			},
			wantError: false,
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
