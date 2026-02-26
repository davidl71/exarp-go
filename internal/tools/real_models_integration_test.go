// Integration tests that call real LLM backends (FM, Ollama, MLX) when available.
// Skip with: go test -short ./internal/tools/...
// Run with: go test -run RealModels ./internal/tools/... (omit -short)
package tools

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
)

// isConnectionRefused reports whether err indicates a connection refused (e.g. Ollama not running).
func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Err != nil {
		var sysErr *syscall.Errno
		if errors.As(opErr.Err, &sysErr) && *sysErr == syscall.ECONNREFUSED {
			return true
		}
	}
	return strings.Contains(err.Error(), "connection refused")
}

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
		if isConnectionRefused(err) {
			t.Skip("Ollama not running:", err)
		}
		t.Fatalf("Generate with real model: %v", err)
	}

	if text == "" {
		t.Error("expected non-empty response from real model")
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
		if isConnectionRefused(err) {
			t.Skip("Ollama not running:", err)
		}
		t.Fatalf("AnalyzeTask with real model: %v", err)
	}

	if result == nil || len(result.Subtasks) == 0 {
		t.Fatal("expected at least one subtask from task breakdown")
	}
}

// TestRealModels_TaskExecutionFlow runs the full task execution flow (T-215) with a real model.
// Uses a temp DB and fixture task; Apply=false so no file writes occur.
func TestRealModels_TaskExecutionFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test with real models in short mode")
	}

	gen := DefaultFMProvider()
	if gen == nil || !gen.Supported() {
		t.Skip("no real model backend available (FM/Ollama/MLX)")
	}

	ctx := context.Background()

	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		t.Fatalf("create .todo2 dir: %v", err)
	}

	if err := database.Init(projectRoot); err != nil {
		t.Fatalf("database.Init: %v", err)
	}

	defer func() { _ = database.Close() }()

	task := &database.Todo2Task{
		Content:         "Add a comment to main.go",
		LongDescription: "Add a simple comment at the top of the file.",
		Status:          "Todo",
		Priority:        "medium",
		Tags:            []string{"test"},
	}
	if err := database.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:      task.ID,
		ProjectRoot: projectRoot,
		Apply:       false,
	})
	if err != nil {
		if isConnectionRefused(err) {
			t.Skip("Ollama not running:", err)
		}
		t.Fatalf("RunTaskExecutionFlow with real model: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil RunTaskExecutionFlowResult")
	}

	if result.Explanation == "" && result.ParseError == "" {
		t.Error("expected either Explanation or ParseError to be non-empty from execution flow")
	}
}

