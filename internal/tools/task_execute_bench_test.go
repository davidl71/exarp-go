package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
)

func BenchmarkRunTaskExecutionFlow(b *testing.B) {
	ctx := context.Background()

	projectRoot := b.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, ".todo2"), 0755); err != nil {
		b.Fatalf("create .todo2 dir: %v", err)
	}

	if err := database.Init(projectRoot); err != nil {
		b.Fatalf("database.Init: %v", err)
	}
	defer func() { _ = database.Close() }()

	task := &database.Todo2Task{
		Content:         "Benchmark task",
		LongDescription: "Task used for benchmarking RunTaskExecutionFlow.",
		Status:          "Todo",
		Priority:        "medium",
		Tags:            []string{"bench"},
	}
	if err := database.CreateTask(ctx, task); err != nil {
		b.Fatalf("CreateTask: %v", err)
	}

	mockJSON := `{"changes":[],"explanation":"Benchmark execution result.","confidence":0.9}`
	router := &mockExecutionRouter{response: mockJSON}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
			TaskID:      task.ID,
			ProjectRoot: projectRoot,
			Apply:       false,
			ModelRouter: router,
		})
		if err != nil {
			b.Fatalf("RunTaskExecutionFlow: %v", err)
		}
	}
}
