package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
)

func TestNewDefaultTaskStore_FileFallback(t *testing.T) {
	// Use a temp dir with no DB so we exercise the file fallback.
	projectRoot := t.TempDir()

	todo2Dir := filepath.Join(projectRoot, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatalf("mkdir .todo2: %v", err)
	}

	store := NewDefaultTaskStore(projectRoot)
	ctx := context.Background()

	// Ensure it implements TaskStore
	var _ = store

	task := &database.Todo2Task{
		ID:       "T-998001",
		Content:  "File fallback task",
		Status:   "Todo",
		Priority: "medium",
		Tags:     []string{"file"},
	}
	if err := store.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := store.GetTask(ctx, "T-998001")
	if err != nil || got == nil {
		t.Fatalf("GetTask: err=%v got=%v", err, got)
	}

	if got.Content != "File fallback task" {
		t.Errorf("GetTask content = %q, want File fallback task", got.Content)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListTasks: err=%v len=%d", err, len(list))
	}

	got.Status = "Done"
	if err := store.UpdateTask(ctx, got); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got2, _ := store.GetTask(ctx, "T-998001")
	if got2.Status != "Done" {
		t.Errorf("after UpdateTask Status = %q, want Done", got2.Status)
	}

	if err := store.DeleteTask(ctx, "T-998001"); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	got3, _ := store.GetTask(ctx, "T-998001")
	if got3 != nil {
		t.Errorf("GetTask after delete = %v, want nil", got3)
	}
}
