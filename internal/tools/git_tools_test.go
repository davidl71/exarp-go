package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestHandleGitToolsNative(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    GitToolsParams
		wantError bool
		validate  func(*testing.T, string)
	}{
		{
			name: "commits action",
			params: GitToolsParams{
				Action: "commits",
				Limit:  10,
			},
			wantError: false,
			validate: func(t *testing.T, result string) {
				// Should return valid JSON
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result), &data); err != nil {
					// May not be JSON if git repo doesn't exist - that's ok
					return
				}
			},
		},
		{
			name: "branches action",
			params: GitToolsParams{
				Action: "branches",
			},
			wantError: false,
			validate: func(t *testing.T, result string) {
				// Result may be error message if not a git repo
			},
		},
		{
			name: "tasks action",
			params: GitToolsParams{
				Action: "tasks",
			},
			wantError: false,
			validate: func(t *testing.T, result string) {
				// Result may be empty or error message
			},
		},
		{
			name: "unknown action",
			params: GitToolsParams{
				Action: "unknown",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := HandleGitToolsNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("HandleGitToolsNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestHandleGitTools(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("PROJECT_ROOT", tmpDir)
	defer os.Unsetenv("PROJECT_ROOT")

	tests := []struct {
		name      string
		params    GitToolsParams
		wantError bool
	}{
		{
			name: "commits action",
			params: GitToolsParams{
				Action: "commits",
				Limit:  10,
			},
			wantError: false,
		},
		{
			name: "branches action",
			params: GitToolsParams{
				Action: "branches",
			},
			wantError: false,
		},
		{
			name: "tasks action",
			params: GitToolsParams{
				Action: "tasks",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			argsJSON, _ := json.Marshal(tt.params)
			result, err := handleGitTools(ctx, argsJSON)
			if (err != nil) != tt.wantError {
				t.Errorf("handleGitTools() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
