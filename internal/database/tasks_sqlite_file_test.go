package database

import (
	"context"
	"path/filepath"
	"testing"
)

func TestListTasksFromSQLiteFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dir := t.TempDir()
	if err := Init(dir); err != nil {
		t.Fatal(err)
	}
	if err := CreateTask(ctx, &Todo2Task{ID: "T-9000700000000000001", Content: "hello", Status: "Todo"}); err != nil {
		t.Fatal(err)
	}
	if err := Close(); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(dir, ".todo2", "todo2.db")
	tasks, err := ListTasksFromSQLiteFile(ctx, dbPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 || tasks[0].ID != "T-9000700000000000001" {
		t.Fatalf("got %#v", tasks)
	}
	if tasks[0].Content != "hello" && tasks[0].Name != "hello" {
		t.Fatalf("content/name = %q / %q", tasks[0].Content, tasks[0].Name)
	}
}
