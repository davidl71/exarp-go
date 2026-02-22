// llamacpp_test.go — Tests for llamacpp tool, provider, GPU detection, and Ollama model discovery.
package tools

import "testing"

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
