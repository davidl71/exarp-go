// Integration tests that call real LLM backends (FM, Ollama, MLX) when available.
// Skip with: go test -short ./internal/tools/...
// Run with: go test -run RealModels ./internal/tools/... (omit -short)
package tools

import (
	"context"
	"testing"
)

const realModelsTimeoutSeconds = 60

func TestRealModels_TextGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test with real models in short mode")
	}
	ctx := context.Background()
	gen := DefaultFMProvider()
	if gen == nil || !gen.Supported() {
		t.Skip("no real model backend available (FM/Ollama/MLX)")
	}
	text, err := gen.Generate(ctx, "Reply with exactly the word OK and nothing else.", 10, 0)
	if err != nil {
		t.Fatalf("Generate with real model: %v", err)
	}
	if text == "" {
		t.Error("expected non-empty response from real model")
	}
}

func TestRealModels_AnalyzePrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test with real models in short mode")
	}
	ctx := context.Background()
	gen := DefaultFMProvider()
	if gen == nil || !gen.Supported() {
		t.Skip("no real model backend available")
	}
	result, err := AnalyzePrompt(ctx, "List three colors.", "", "test", gen)
	if err != nil {
		t.Fatalf("AnalyzePrompt with real model: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil PromptAnalysis")
	}
	// At least one dimension should be set (model may not return full JSON)
	hasScore := result.Clarity > 0 || result.Specificity > 0 || result.Completeness > 0 ||
		result.Structure > 0 || result.Actionability > 0
	if !hasScore && result.Summary == "" {
		t.Error("expected at least one score or summary from prompt analysis")
	}
}

func TestRealModels_AnalyzeTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test with real models in short mode")
	}
	ctx := context.Background()
	gen := DefaultFMProvider()
	if gen == nil || !gen.Supported() {
		t.Skip("no real model backend available")
	}
	result, err := AnalyzeTask(ctx, "Add a unit test for function Foo.", "", "", gen)
	if err != nil {
		t.Fatalf("AnalyzeTask with real model: %v", err)
	}
	if result == nil || len(result.Subtasks) == 0 {
		t.Fatal("expected at least one subtask from task breakdown")
	}
}
