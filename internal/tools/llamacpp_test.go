// llamacpp_test.go — Tests for llamacpp tool, provider, GPU detection, and Ollama model discovery.
package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLlamaCppProviderNilWhenNotBuilt(t *testing.T) {
	p := DefaultLlamaCppProvider()
	if p != nil && p.Supported() {
		t.Skip("llamacpp provider is available — skipping nil test")
	}
	t.Log("DefaultLlamaCppProvider() returned nil or unsupported (expected without llamacpp build tag)")
}

func TestLlamaCppIfAvailable(t *testing.T) {
	g := llamacppIfAvailable()
	if g != nil {
		t.Log("llamacpp is available in the FM chain")
	} else {
		t.Log("llamacpp is not available (expected when not built with llamacpp tag)")
	}
}

func TestDetectGPU(t *testing.T) {
	info := DetectGPU()
	t.Logf("GPU: available=%v, backend=%s, device=%s, memory_mb=%d, layers=%d",
		info.Available, info.Backend, info.DeviceName, info.MemoryMB, info.Layers)
	if info.Backend == "" {
		t.Error("DetectGPU().Backend should not be empty")
	}
}

func TestRecommendGPULayers(t *testing.T) {
	tests := []struct {
		name       string
		modelMB    int64
		gpuMB      int64
		wantLayers int
	}{
		{"all layers fit", 4000, 8000, -1},
		{"partial fit 40", 8000, 8000, 40},
		{"partial fit 24", 8000, 5000, 24},
		{"partial fit 16", 8000, 3000, 16},
		{"no GPU memory", 4000, 0, 0},
		{"tiny GPU", 8000, 1000, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RecommendGPULayers(tt.modelMB, tt.gpuMB)
			if got != tt.wantLayers {
				t.Errorf("RecommendGPULayers(%d, %d) = %d, want %d", tt.modelMB, tt.gpuMB, got, tt.wantLayers)
			}
		})
	}
}

func TestDiscoverOllamaModels(t *testing.T) {
	models, err := DiscoverOllamaModels()
	if err != nil {
		t.Logf("No Ollama models found (expected in CI): %v", err)
		return
	}
	t.Logf("Found %d Ollama GGUF models", len(models))
	for _, m := range models {
		t.Logf("  %s -> %s (%d bytes, family=%s)", m.Name, m.GGUFPath, m.Size, m.Family)
	}
}

func TestResolveModelAlias(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"llama3", "llama3.2"},
		{"llama", "llama3.2"},
		{"phi", "phi3"},
		{"gemma", "gemma2"},
		{"unknown-model", "unknown-model"},
		{"llama3:1b", "llama3.2:1b"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ResolveModelAlias(tt.input)
			if got != tt.want {
				t.Errorf("ResolveModelAlias(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestLlamaCppToolStatus(t *testing.T) {
	result, err := handleLlamaCppStatus()
	if err != nil {
		t.Fatalf("handleLlamaCppStatus() error = %v", err)
	}
	if len(result) == 0 {
		t.Error("handleLlamaCppStatus() returned empty result")
	}
	t.Logf("Status: %s", result[0].Text)
}

func TestLlamaCppToolModels(t *testing.T) {
	result, err := handleLlamaCppModels()
	if err != nil {
		t.Fatalf("handleLlamaCppModels() error = %v", err)
	}
	if len(result) == 0 {
		t.Error("handleLlamaCppModels() returned empty result")
	}
	text := result[0].Text
	if len(text) > 200 {
		text = text[:200]
	}
	t.Logf("Models: %s", text)
}

func TestLlamaCppToolDispatch_UnknownAction(t *testing.T) {
	args, _ := json.Marshal(map[string]interface{}{"action": "notanaction"})
	_, err := handleLlamaCppTool(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for unknown action, got nil")
	}
	if !strings.Contains(err.Error(), "unknown llamacpp action") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLlamaCppToolDispatch_DefaultAction(t *testing.T) {
	// Empty action should default to "status" (same as TestLlamaCppToolStatus).
	args, _ := json.Marshal(map[string]interface{}{})
	result, err := handleLlamaCppTool(context.Background(), args)
	if err != nil {
		t.Fatalf("handleLlamaCppTool default action error = %v", err)
	}
	if len(result) == 0 {
		t.Error("handleLlamaCppTool default action returned empty result")
	}
}

func TestLlamaCppToolLoad_MissingModel(t *testing.T) {
	// Neither model_path nor model provided → error.
	_, err := handleLlamaCppLoad(map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error when model_path and model are both absent")
	}
	if !strings.Contains(err.Error(), "model_path or model name is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLlamaCppToolLoad_WithModelPath(t *testing.T) {
	// Providing a model_path: nocgo stub returns an error from LlamaCppLoad,
	// which should surface as a load-failed error (not a missing-model error).
	_, err := handleLlamaCppLoad(map[string]interface{}{"model_path": "/nonexistent/model.gguf"})
	if err == nil {
		t.Skip("llamacpp CGO build available — load may succeed or fail differently; skipping nocgo assertion")
	}
	// In the nocgo build the stub always returns an error.
	if !strings.Contains(err.Error(), "llamacpp load failed") && !strings.Contains(err.Error(), "not built") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLlamaCppToolUnload(t *testing.T) {
	// Unload is always a no-op in the nocgo build; must return success.
	result, err := handleLlamaCppUnload(map[string]interface{}{})
	if err != nil {
		t.Fatalf("handleLlamaCppUnload() error = %v", err)
	}
	if len(result) == 0 || result[0].Text != "Model unloaded" {
		t.Errorf("unexpected unload result: %+v", result)
	}
}

func TestLlamaCppToolGenerate_NoProvider(t *testing.T) {
	if p := DefaultLlamaCppProvider(); p != nil && p.Supported() {
		t.Skip("llamacpp provider available — skipping unsupported path test")
	}
	_, err := handleLlamaCppGenerate(context.Background(), map[string]interface{}{"prompt": "hello"})
	if err == nil {
		t.Fatal("expected error when provider is unavailable")
	}
	if !strings.Contains(err.Error(), "llamacpp provider not available") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestModelManagerStats_NoCGO(t *testing.T) {
	stats := ModelManagerStats()
	if stats == nil {
		t.Fatal("ModelManagerStats() returned nil")
	}
	// In a non-CGO build the nocgo stub returns {"available": false}.
	// In a CGO build the real manager returns a richer map.
	// We only assert the key exists in the nocgo case.
	if p := DefaultLlamaCppProvider(); p != nil && p.Supported() {
		t.Skip("llamacpp CGO build — skipping nocgo stub assertion")
	}
	avail, ok := stats["available"]
	if !ok {
		t.Error("ModelManagerStats() missing 'available' key")
	}
	if avail != false {
		t.Errorf("expected available=false in nocgo build, got %v", avail)
	}
}

func TestLlamaCppNoCGO_Stubs(t *testing.T) {
	if p := DefaultLlamaCppProvider(); p != nil && p.Supported() {
		t.Skip("llamacpp CGO build — skipping nocgo stub tests")
	}

	// LlamaCppLoad must return a non-nil error.
	if err := LlamaCppLoad("/any/path.gguf"); err == nil {
		t.Error("LlamaCppLoad() should return error in nocgo build")
	}

	// LlamaCppUnload must not panic.
	LlamaCppUnload()

	// LlamaCppModelPath must return empty string.
	if got := LlamaCppModelPath(); got != "" {
		t.Errorf("LlamaCppModelPath() = %q, want \"\"", got)
	}
}
