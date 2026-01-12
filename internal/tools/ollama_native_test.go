package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleOllamaDocs(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid file path",
			params: map[string]interface{}{
				"action":     "docs",
				"file_path":  testFile,
				"model":      "llama3",
				"style":      "google",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "missing file_path",
			params: map[string]interface{}{
				"action": "docs",
				"model":  "llama3",
			},
			wantError: true,
		},
		{
			name: "non-existent file",
			params: map[string]interface{}{
				"action":    "docs",
				"file_path": "/nonexistent/file.go",
				"model":     "llama3",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			host := "http://localhost:11434"
			if envHost := os.Getenv("OLLAMA_HOST"); envHost != "" {
				host = envHost
			}

			result, err := handleOllamaDocs(ctx, tt.params, host)
			if (err != nil) != tt.wantError {
				t.Errorf("handleOllamaDocs() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && result != nil && len(result) > 0 {
				// Validate JSON format
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON response: %v", err)
				}
			}
		})
	}
}

func TestHandleOllamaQuality(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func add(a, b int) int {
	return a + b
}
`
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid file path",
			params: map[string]interface{}{
				"action":    "quality",
				"file_path": testFile,
				"model":     "llama3",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "missing file_path",
			params: map[string]interface{}{
				"action": "quality",
				"model":  "llama3",
			},
			wantError: true,
		},
		{
			name: "with include_suggestions",
			params: map[string]interface{}{
				"action":             "quality",
				"file_path":          testFile,
				"model":              "llama3",
				"include_suggestions": true,
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			host := "http://localhost:11434"
			if envHost := os.Getenv("OLLAMA_HOST"); envHost != "" {
				host = envHost
			}

			result, err := handleOllamaQuality(ctx, tt.params, host)
			if (err != nil) != tt.wantError {
				t.Errorf("handleOllamaQuality() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && result != nil && len(result) > 0 {
				// Validate JSON format
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON response: %v", err)
				}
			}
		})
	}
}

func TestHandleOllamaSummary(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "valid data string",
			params: map[string]interface{}{
				"action": "summary",
				"data":   "This is a test document that needs to be summarized.",
				"model":  "llama3",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "missing data",
			params: map[string]interface{}{
				"action": "summary",
				"model":  "llama3",
			},
			wantError: true,
		},
		{
			name: "with level parameter",
			params: map[string]interface{}{
				"action": "summary",
				"data":   "Long document text here...",
				"model":  "llama3",
				"level":  "detailed",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "with data as map",
			params: map[string]interface{}{
				"action": "summary",
				"data": map[string]interface{}{
					"content": "Test content",
				},
				"model": "llama3",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			host := "http://localhost:11434"
			if envHost := os.Getenv("OLLAMA_HOST"); envHost != "" {
				host = envHost
			}

			result, err := handleOllamaSummary(ctx, tt.params, host)
			if (err != nil) != tt.wantError {
				t.Errorf("handleOllamaSummary() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && result != nil && len(result) > 0 {
				// Validate JSON format
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON response: %v", err)
				}
			}
		})
	}
}

func TestHandleOllamaNative(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name: "status action",
			params: map[string]interface{}{
				"action": "status",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "models action",
			params: map[string]interface{}{
				"action": "models",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "docs action",
			params: map[string]interface{}{
				"action":    "docs",
				"file_path": "/tmp/test.go",
				"code":      "func test() {}",
			},
			wantError: true, // Missing file_path or code
		},
		{
			name: "quality action",
			params: map[string]interface{}{
				"action":    "quality",
				"file_path": "/tmp/test.go",
				"code":      "func test() {}",
			},
			wantError: true, // Missing file_path or code
		},
		{
			name: "summary action",
			params: map[string]interface{}{
				"action": "summary",
				"text":   "Test text",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
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
			result, err := handleOllamaNative(ctx, tt.params)
			if (err != nil) != tt.wantError {
				t.Errorf("handleOllamaNative() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && result != nil && len(result) > 0 {
				// Validate JSON format
				var data map[string]interface{}
				if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
					t.Errorf("invalid JSON response: %v", err)
				}
			}
		})
	}
}
