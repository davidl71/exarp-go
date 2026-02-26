package tools

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// skipIfOllamaNotReachable skips the test if Ollama server is not reachable at localhost:11434.
// Use for tests that require Ollama to be running.
func skipIfOllamaNotReachable(t *testing.T) {
	t.Helper()
	host := "localhost:11434"
	if h := os.Getenv("OLLAMA_HOST"); h != "" {
		// OLLAMA_HOST may be "http://localhost:11434"; extract host:port
		if len(h) >= 8 && h[:7] == "http://" {
			host = h[7:]
		} else if len(h) >= 9 && h[:8] == "https://" {
			host = h[8:]
		} else {
			host = h
		}
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Skipf("Ollama not reachable at %s: %v", host, err)
	}
	_ = conn.Close()
}

// getOllamaTestModel returns the Ollama model for tests (env OLLAMA_TEST_MODEL or default qwen2.5:1.5b).
// Use a light model for fast tests; override to your preferred model if needed.
func getOllamaTestModel() string {
	if s := os.Getenv("OLLAMA_TEST_MODEL"); s != "" {
		return s
	}
	return "qwen2.5:1.5b"
}

// getOllamaTestCodeModel returns the Ollama code model for tests (env OLLAMA_TEST_CODE_MODEL or default qwen2.5:1.5b).
// Same light default as getOllamaTestModel for speed; override for code-specific models (e.g. qwen3-coder:latest).
func getOllamaTestCodeModel() string {
	if s := os.Getenv("OLLAMA_TEST_CODE_MODEL"); s != "" {
		return s
	}
	return "qwen2.5:1.5b"
}

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
				"action":    "docs",
				"file_path": testFile,
				"model":     getOllamaTestModel(),
				"style":     "google",
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "missing file_path",
			params: map[string]interface{}{
				"action": "docs",
				"model":  getOllamaTestModel(),
			},
			wantError: true,
		},
		{
			name: "non-existent file",
			params: map[string]interface{}{
				"action":    "docs",
				"file_path": "/nonexistent/file.go",
				"model":     getOllamaTestModel(),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantError {
				skipIfOllamaNotReachable(t)
			}
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
				"model":     getOllamaTestModel(),
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "missing file_path",
			params: map[string]interface{}{
				"action": "quality",
				"model":  getOllamaTestModel(),
			},
			wantError: true,
		},
		{
			name: "with include_suggestions",
			params: map[string]interface{}{
				"action":              "quality",
				"file_path":           testFile,
				"model":               getOllamaTestModel(),
				"include_suggestions": true,
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantError {
				skipIfOllamaNotReachable(t)
			}
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
				"model":  getOllamaTestModel(),
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
		{
			name: "missing data",
			params: map[string]interface{}{
				"action": "summary",
				"model":  getOllamaTestModel(),
			},
			wantError: true,
		},
		{
			name: "with level parameter",
			params: map[string]interface{}{
				"action": "summary",
				"data":   "Long document text here...",
				"model":  getOllamaTestModel(),
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
				"model": getOllamaTestModel(),
			},
			wantError: false, // Will fail if Ollama not running, but that's expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantError {
				skipIfOllamaNotReachable(t)
			}
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
				"data":   "Test text",
				"model":  getOllamaTestCodeModel(),
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
			if !tt.wantError {
				skipIfOllamaNotReachable(t)
			}
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
