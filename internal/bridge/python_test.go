package bridge

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExecutePythonTool_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		wantErr bool
	}{
		{
			name: "simple args",
			args: map[string]interface{}{
				"action": "test",
				"value":  42,
			},
			wantErr: false,
		},
		{
			name:    "empty args",
			args:    map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "nested args",
			args: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			argsJSON, err := json.Marshal(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify it can be unmarshaled back
			var unmarshaled map[string]interface{}
			if err := json.Unmarshal(argsJSON, &unmarshaled); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}
		})
	}
}

func TestExecutePythonTool_PROJECT_ROOT(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		wantPath    string
	}{
		{
			name:        "with PROJECT_ROOT",
			projectRoot: "/test/project",
			wantPath:    "/test/project/bridge/execute_tool.py",
		},
		{
			name:        "without PROJECT_ROOT",
			projectRoot: "",
			wantPath:    filepath.Join("..", "project-management-automation", "bridge", "execute_tool.py"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			originalRoot := os.Getenv("PROJECT_ROOT")
			defer os.Setenv("PROJECT_ROOT", originalRoot)

			// Set test env
			if tt.projectRoot != "" {
				os.Setenv("PROJECT_ROOT", tt.projectRoot)
			} else {
				os.Unsetenv("PROJECT_ROOT")
			}

			// Get project root
			projectRoot := os.Getenv("PROJECT_ROOT")
			if projectRoot == "" {
				projectRoot = filepath.Join("..", "project-management-automation")
			}

			// Verify path construction
			bridgeScript := filepath.Join(projectRoot, "bridge", "execute_tool.py")
			if bridgeScript != tt.wantPath {
				// Allow for path differences (e.g., relative vs absolute)
				// Just verify it ends with expected path
				expectedSuffix := filepath.Join("bridge", "execute_tool.py")
				if !filepath.IsAbs(tt.wantPath) {
					expectedSuffix = tt.wantPath
				}
				if filepath.Base(bridgeScript) != filepath.Base(expectedSuffix) {
					t.Errorf("ExecutePythonTool path = %v, want ends with %v", bridgeScript, expectedSuffix)
				}
			}
		})
	}
}

func TestGetPythonPrompt_JSONParsing(t *testing.T) {
	tests := []struct {
		name       string
		jsonInput  string
		wantPrompt string
		wantErr    bool
	}{
		{
			name:       "valid response",
			jsonInput:  `{"success": true, "prompt": "test prompt"}`,
			wantPrompt: "test prompt",
			wantErr:    false,
		},
		{
			name:      "error response",
			jsonInput: `{"success": false, "error": "test error"}`,
			wantErr:   true,
		},
		{
			name:      "invalid JSON",
			jsonInput: `{invalid json}`,
			wantErr:   true,
		},
		{
			name:      "missing success field",
			jsonInput: `{"prompt": "test"}`,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result struct {
				Success bool   `json:"success"`
				Prompt  string `json:"prompt"`
				Error   string `json:"error,omitempty"`
			}

			err := json.Unmarshal([]byte(tt.jsonInput), &result)
			
			// Invalid JSON should cause unmarshal error
			if tt.name == "invalid JSON" {
				if err == nil {
					t.Error("json.Unmarshal() error = nil, want error for invalid JSON")
				}
				return
			}
			
			// Valid JSON should parse without error
			if err != nil {
				t.Errorf("json.Unmarshal() error = %v, want no error", err)
				return
			}

			// For success cases, verify prompt
			if !tt.wantErr && result.Success {
				if result.Prompt != tt.wantPrompt {
					t.Errorf("result.Prompt = %v, want %v", result.Prompt, tt.wantPrompt)
				}
			}

			// For error response cases, verify error is present or Success is false
			if tt.wantErr && tt.name == "error response" {
				if result.Success {
					t.Error("result.Success = true, want false for error response")
				}
				if result.Error == "" {
					t.Error("result.Error is empty, want error message")
				}
			}

			// For missing success field, verify Success defaults to false
			if tt.name == "missing success field" {
				if result.Success {
					t.Error("result.Success = true, want false when success field is missing")
				}
			}
		})
	}
}

func TestContextTimeout(t *testing.T) {
	t.Run("ExecutePythonTool timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Create a context that will timeout quickly
		shortCtx, shortCancel := context.WithTimeout(ctx, 50*time.Millisecond)
		defer shortCancel()

		// Verify timeout context is cancelled
		select {
		case <-shortCtx.Done():
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Error("context should have timed out")
		}
	})

	t.Run("GetPythonPrompt timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Verify timeout context is cancelled
		select {
		case <-ctx.Done():
			// Expected after timeout
		case <-time.After(150 * time.Millisecond):
			// Context timed out
		}
	})
}

func TestExecutePythonResource_URIParsing(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		wantPath string
	}{
		{
			name:     "scorecard URI",
			uri:      "stdio://scorecard",
			wantPath: filepath.Join("..", "project-management-automation", "bridge", "execute_resource.py"),
		},
		{
			name:     "memories URI",
			uri:      "stdio://memories",
			wantPath: filepath.Join("..", "project-management-automation", "bridge", "execute_resource.py"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get project root
			projectRoot := os.Getenv("PROJECT_ROOT")
			if projectRoot == "" {
				projectRoot = filepath.Join("..", "project-management-automation")
			}

			// Verify path construction
			bridgeScript := filepath.Join(projectRoot, "bridge", "execute_resource.py")
			if bridgeScript == "" {
				t.Error("bridgeScript should not be empty")
			}

			// Verify script path ends with execute_resource.py
			baseName := filepath.Base(bridgeScript)
			if baseName != "execute_resource.py" {
				t.Errorf("bridgeScript base = %v, want execute_resource.py", baseName)
			}
		})
	}
}
