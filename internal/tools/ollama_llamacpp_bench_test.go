// ollama_llamacpp_bench_test.go — benchmarks Ollama (HTTP) vs llamacpp (local) generate performance.
// Run: make bench (or go test -bench=BenchmarkOllama -bench=BenchmarkLlamaCpp ./internal/tools/).
// Ollama benchmark requires ollama serve; llamacpp benchmark requires build with -tags llamacpp,cgo and a loaded model.
package tools

import (
	"context"
	"testing"
	"time"
)

const (
	benchPrompt    = "Say hello in one word."
	benchMaxTokens = 10
	benchTemp      = 0
)

// BenchmarkOllamaGenerate measures end-to-end latency for one Ollama HTTP generate.
// Requires Ollama server (e.g. ollama serve) and a model (default from config or llama3.2).
// Reports ns/op; divide by 1e6 for ms per call.
func BenchmarkOllamaGenerate(b *testing.B) {
	ctx := context.Background()
	gen := DefaultOllamaTextGenerator()
	probe, err := gen.Generate(ctx, benchPrompt, benchMaxTokens, benchTemp)
	if err != nil {
		b.Skipf("Ollama not available (start with: ollama serve): %v", err)
	}
	_ = probe

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(ctx, benchPrompt, benchMaxTokens, benchTemp)
	}
}

// BenchmarkLlamaCppGenerate measures end-to-end latency for one llamacpp local generate.
// Requires build with -tags llamacpp,cgo and a loaded model (e.g. via LLAMACPP_MODEL_PATH or load action).
// Skipped when llamacpp is not built or no model is loaded.
func BenchmarkLlamaCppGenerate(b *testing.B) {
	p := DefaultLlamaCppProvider()
	if p == nil || !p.Supported() {
		b.Skip("llamacpp not available (build with -tags llamacpp,cgo and load a model)")
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Generate(ctx, benchPrompt, benchMaxTokens, benchTemp)
	}
}

// TestOllamaVsLlamaCppPerformanceReport runs one generate per backend and logs timings.
// Use: go test -run TestOllamaVsLlamaCppPerformanceReport -v ./internal/tools/
// Useful for a quick comparison without full benchmark runs.
func TestOllamaVsLlamaCppPerformanceReport(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ollama
	gen := DefaultOllamaTextGenerator()
	start := time.Now()
	ollamaOut, ollamaErr := gen.Generate(ctx, benchPrompt, benchMaxTokens, benchTemp)
	ollamaElapsed := time.Since(start)
	if ollamaErr != nil {
		t.Logf("Ollama: skipped (%v)", ollamaErr)
	} else {
		tokensApprox := len(ollamaOut) / 4
		if tokensApprox < 1 && len(ollamaOut) > 0 {
			tokensApprox = 1
		}
		t.Logf("Ollama: %v total, ~%d tokens, ~%.2f tokens/s",
			ollamaElapsed.Round(time.Millisecond), tokensApprox,
			float64(tokensApprox)/ollamaElapsed.Seconds())
	}

	// LlamaCpp
	p := DefaultLlamaCppProvider()
	if p == nil || !p.Supported() {
		t.Log("LlamaCpp: skipped (not built or no model loaded)")
		return
	}
	start = time.Now()
	llamaOut, llamaErr := p.Generate(ctx, benchPrompt, benchMaxTokens, benchTemp)
	llamaElapsed := time.Since(start)
	if llamaErr != nil {
		t.Logf("LlamaCpp: error %v", llamaErr)
		return
	}
	tokensApprox := len(llamaOut) / 4
	if tokensApprox < 1 && len(llamaOut) > 0 {
		tokensApprox = 1
	}
	t.Logf("LlamaCpp: %v total, ~%d tokens, ~%.2f tokens/s",
		llamaElapsed.Round(time.Millisecond), tokensApprox,
		float64(tokensApprox)/llamaElapsed.Seconds())
}
