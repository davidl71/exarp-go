package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncTodo2TasksRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	todo2Dir := filepath.Join(dir, ".todo2")
	if err := os.MkdirAll(todo2Dir, 0755); err != nil {
		t.Fatal(err)
	}
	statePath := filepath.Join(todo2Dir, "state.todo2.json")
	if err := os.WriteFile(statePath, []byte("{not valid json"), 0644); err != nil {
		t.Fatal(err)
	}

	err := SyncTodo2Tasks(dir)
	if err == nil {
		t.Fatal("expected error for corrupt state.todo2.json")
	}
	if !strings.Contains(err.Error(), "sync: load json") {
		t.Fatalf("expected sync load json error, got %v", err)
	}
}
