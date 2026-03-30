// ollama_integration_test.go — Integration tests for Ollama tool using a mock HTTP server (CI-friendly)
// or optional live tests when OLLAMA_HOST is set. Tag hints for Todo2: #testing #ollama

package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// mockOllamaServer returns an httptest.Server that mimics Ollama API (GET /api/tags, POST /api/generate).
// Used so integration tests run in CI without a real Ollama instance.
func mockOllamaServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// GET /api/tags — list models (same shape as Ollama)
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"models": []map[string]interface{}{
				{"name": "test-model", "modified_at": "2025-01-01T00:00:00Z", "size": 1000000, "digest": "abc123"},
			},
		})
	})

	// POST /api/generate — generate (non-streaming response)
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Model  string `json:"model"`
			Prompt string `json:"prompt"`
			Stream bool   `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"model":          req.Model,
			"response":       "mock response for: " + req.Prompt,
			"done":           true,
			"created_at":     "2025-01-01T00:00:00Z",
			"total_duration": 1000000,
		})
	})

	return httptest.NewServer(mux)
}

func TestOllamaIntegration_Mock_Status(t *testing.T) {
	srv := mockOllamaServer(t)
	defer srv.Close()

	ctx := context.Background()
	params := map[string]interface{}{
		"action": "status",
		"host":   srv.URL,
	}

	result, err := handleOllamaNative(ctx, params)
	if err != nil {
		t.Fatalf("handleOllamaNative(status) error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if data["status"] != "running" && data["success"] != true {
		t.Errorf("expected status running or success true, got %v", data)
	}
	if data["host"] != srv.URL {
		t.Errorf("host = %v, want %s", data["host"], srv.URL)
	}
}

func TestOllamaIntegration_Mock_Models(t *testing.T) {
	srv := mockOllamaServer(t)
	defer srv.Close()

	ctx := context.Background()
	params := map[string]interface{}{
		"action": "models",
		"host":   srv.URL,
	}

	result, err := handleOllamaNative(ctx, params)
	if err != nil {
		t.Fatalf("handleOllamaNative(models) error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := data["models"]; !ok {
		t.Errorf("expected 'models' key, got %v", data)
	}
	if data["success"] != true {
		t.Errorf("success = %v, want true", data["success"])
	}
}

func TestOllamaIntegration_Mock_Generate(t *testing.T) {
	srv := mockOllamaServer(t)
	defer srv.Close()

	ctx := context.Background()
	params := map[string]interface{}{
		"action": "generate",
		"host":   srv.URL,
		"prompt": "Say hello",
		"model":  "test-model",
		"stream": false,
	}

	result, err := handleOllamaNative(ctx, params)
	if err != nil {
		t.Fatalf("handleOllamaNative(generate) error = %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if data["success"] != true {
		t.Errorf("success = %v, want true", data["success"])
	}
	if resp, ok := data["response"].(string); !ok || !strings.Contains(resp, "mock response") {
		t.Errorf("expected response to contain mock text, got %v", data["response"])
	}
}

// TestOllamaIntegration_Live runs against a real Ollama server when OLLAMA_HOST is set.
// Skipped in CI when no Ollama is running; use for local/manual verification.
func TestOllamaIntegration_Live(t *testing.T) {
	if os.Getenv("OLLAMA_HOST") == "" {
		t.Skip("OLLAMA_HOST not set; skipping live integration test")
	}
	skipIfOllamaNotReachable(t)

	ctx := context.Background()
	host := "http://localhost:11434"
	if h := os.Getenv("OLLAMA_HOST"); h != "" {
		host = h
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
	}

	// status
	params := map[string]interface{}{"action": "status", "host": host}
	result, err := handleOllamaNative(ctx, params)
	if err != nil {
		t.Fatalf("live status: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("live status: empty result")
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("live status JSON: %v", err)
	}
	if data["status"] != "running" {
		t.Errorf("live status: got %v", data["status"])
	}

	// models
	params = map[string]interface{}{"action": "models", "host": host}
	result, err = handleOllamaNative(ctx, params)
	if err != nil {
		t.Fatalf("live models: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("live models: empty result")
	}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("live models JSON: %v", err)
	}
	if _, ok := data["models"]; !ok {
		t.Errorf("live models: no 'models' key")
	}

	// generate (use test model; skip if not available)
	params = map[string]interface{}{
		"action": "generate", "host": host,
		"prompt": "Reply with one word: ok", "model": getOllamaTestModel(), "stream": false,
	}
	result, err = handleOllamaNative(ctx, params)
	if err != nil {
		skipIfOllamaModelUnavailable(t, err)
		t.Fatalf("live generate: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("live generate: empty result")
	}
	if err := json.Unmarshal([]byte(result[0].Text), &data); err != nil {
		t.Fatalf("live generate JSON: %v", err)
	}
	if data["success"] != true {
		t.Errorf("live generate success = %v", data["success"])
	}
}
