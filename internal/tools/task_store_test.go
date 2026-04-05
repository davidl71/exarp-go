package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
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

	if gotHash := models.GetContentHash(got); gotHash == "" {
		t.Error("GetTask content_hash = empty, want populated hash")
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
	if gotHash := models.GetContentHash(got2); gotHash == "" {
		t.Error("after UpdateTask content_hash = empty, want populated hash")
	}

	if err := store.DeleteTask(ctx, "T-998001"); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	got3, _ := store.GetTask(ctx, "T-998001")
	if got3 != nil {
		t.Errorf("GetTask after delete = %v, want nil", got3)
	}
}

func TestNewDefaultTaskStore_DBCreateAndUpdateSetContentHash(t *testing.T) {
	cleanup := initSessionTestDB(t)
	defer cleanup()

	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		t.Fatalf("GetProjectRootWithFallback: %v", err)
	}

	store := NewDefaultTaskStore(projectRoot)
	ctx := context.Background()

	task := &database.Todo2Task{
		ID:              "T-998002",
		Content:         "DB-backed task",
		LongDescription: "Initial description",
		Status:          "Todo",
		Priority:        "medium",
	}
	if err := store.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := store.GetTask(ctx, task.ID)
	if err != nil || got == nil {
		t.Fatalf("GetTask after CreateTask: err=%v got=%v", err, got)
	}

	initialHash := models.GetContentHash(got)
	if initialHash == "" {
		t.Fatal("CreateTask content_hash = empty, want populated hash")
	}

	got.LongDescription = "Updated description"
	if err := store.UpdateTask(ctx, got); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got2, err := store.GetTask(ctx, task.ID)
	if err != nil || got2 == nil {
		t.Fatalf("GetTask after UpdateTask: err=%v got=%v", err, got2)
	}

	updatedHash := models.GetContentHash(got2)
	if updatedHash == "" {
		t.Fatal("UpdateTask content_hash = empty, want populated hash")
	}
	if updatedHash == initialHash {
		t.Errorf("content_hash did not change after content update: got %q", updatedHash)
	}
}

func TestNewDefaultTaskStoreListIncludesLegacyNullProjectRows(t *testing.T) {
	cleanup := initSessionTestDB(t)
	defer cleanup()

	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		t.Fatalf("GetProjectRootWithFallback: %v", err)
	}

	db, err := database.GetDBx()
	if err != nil {
		t.Fatalf("GetDBx: %v", err)
	}

	_, err = db.ExecContext(context.Background(), `
		INSERT INTO tasks (
			id, name, content, long_description,
			status, status_enum,
			priority, priority_enum, priority_rank,
			completed,
			created, last_modified, completed_at,
			created_ts, last_modified_ts, completed_at_ts,
			metadata, metadata_protobuf, metadata_format,
			parent_id, project_id, assigned_to, host, agent,
			assignee, assigned_at, lock_until,
			version, created_at, updated_at
		) VALUES (
		  ?, ?, ?, ?,
		  ?, ?,
		  ?, ?, 0,
		  ?,
		  ?, ?, '',
		  strftime('%s', ?), strftime('%s', ?), 0,
		  ?, ?, ?,
		  ?, ?, ?, ?, ?,
		  '', 0, 0,
		  1, strftime('%s', 'now'), strftime('%s', 'now')
		)
	`, "T-legacy-null-project", "", "Legacy null project", "", database.StatusTodo, 1, "medium", 2, 0, "2026-03-24T00:00:00Z", "2026-03-24T00:00:00Z", "2026-03-24T00:00:00Z", "2026-03-24T00:00:00Z", "", []byte(nil), "json", "", nil, "", "", "")
	if err != nil {
		t.Fatalf("insert legacy task: %v", err)
	}

	store := NewDefaultTaskStore(projectRoot)
	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	for _, task := range list {
		if task.ID == "T-legacy-null-project" {
			return
		}
	}

	t.Fatal("expected legacy NULL-project task to be visible through the default task store list")
}
