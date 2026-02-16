package database

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

// initCommentsTestDB sets up a temp DB with migrations from the repo (same pattern as initLockTestDB).
func initCommentsTestDB(t *testing.T) string {
	t.Helper()
	testDBMu.Lock()
	t.Cleanup(func() { testDBMu.Unlock() })
	_, self, _, _ := runtime.Caller(0)
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(self)))
	migrationsDir := filepath.Join(repoRoot, "migrations")
	tmpDir := t.TempDir()
	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	cfg.Driver = DriverSQLite
	cfg.DSN = filepath.Join(tmpDir, ".todo2", "todo2.db")
	cfg.MigrationsDir = migrationsDir
	cfg.AutoMigrate = true
	if err = InitWithConfig(cfg); err != nil {
		t.Fatalf("InitWithConfig() error = %v", err)
	}
	t.Cleanup(func() { _ = Close() })
	return tmpDir
}

func TestAddComments(t *testing.T) {
	initCommentsTestDB(t)
	var err error
	// Create a task first (required for foreign key; use valid ID T-<digits>)
	taskID := "T-1000001"
	task := &models.Todo2Task{
		ID:      taskID,
		Content: "Test task for comments",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Test adding single comment
	comments := []Comment{
		{
			Type:    CommentTypeNote,
			Content: "This is a test note",
		},
	}

	err = AddComments(context.Background(), taskID, comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify comment was added
	retrieved, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(retrieved) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(retrieved))
	}

	if retrieved[0].Type != CommentTypeNote {
		t.Errorf("Expected type %s, got %s", CommentTypeNote, retrieved[0].Type)
	}

	if retrieved[0].Content != "This is a test note" {
		t.Errorf("Expected content 'This is a test note', got '%s'", retrieved[0].Content)
	}

	if retrieved[0].TaskID != taskID {
		t.Errorf("Expected taskID %s, got %s", taskID, retrieved[0].TaskID)
	}

	if retrieved[0].ID == "" {
		t.Error("Expected comment ID to be generated, got empty string")
	}
}

func TestAddCommentsBatch(t *testing.T) {
	initCommentsTestDB(t)
	var err error
	// Create task (valid ID: T-<digits>)
	taskID := "T-1000002"
	task := &models.Todo2Task{
		ID:      taskID,
		Content: "Test task",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Test adding multiple comments in one transaction
	comments := []Comment{
		{
			Type:    CommentTypeResearch,
			Content: "Research comment 1",
		},
		{
			Type:    CommentTypeResult,
			Content: "Result comment",
		},
		{
			Type:    CommentTypeNote,
			Content: "Note comment",
		},
	}

	err = AddComments(context.Background(), taskID, comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify all comments were added
	retrieved, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(retrieved) != 3 {
		t.Fatalf("Expected 3 comments, got %d", len(retrieved))
	}

	// Verify comment types
	types := map[string]bool{
		retrieved[0].Type: true,
		retrieved[1].Type: true,
		retrieved[2].Type: true,
	}
	if !types[CommentTypeResearch] || !types[CommentTypeResult] || !types[CommentTypeNote] {
		t.Error("Expected all comment types to be present")
	}
}

func TestGetComments(t *testing.T) {
	initCommentsTestDB(t)
	var err error
	// Create task (valid ID)
	taskID := "T-1000003"
	task := &models.Todo2Task{
		ID:      taskID,
		Content: "Test task",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add comments
	comments := []Comment{
		{
			Type:    CommentTypeNote,
			Content: "First comment",
		},
		{
			Type:    CommentTypeResult,
			Content: "Second comment",
		},
	}

	err = AddComments(context.Background(), taskID, comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Test GetComments
	retrieved, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(retrieved) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(retrieved))
	}

	// Test empty task
	empty, err := GetComments(context.Background(), "T-NONEXISTENT")
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(empty) != 0 {
		t.Errorf("Expected 0 comments for non-existent task, got %d", len(empty))
	}
}

func TestGetCommentsByType(t *testing.T) {
	initCommentsTestDB(t)
	var err error
	// Create tasks (valid IDs)
	taskID1, taskID2 := "T-1000004", "T-1000005"
	task1 := &models.Todo2Task{ID: taskID1, Content: "Task 1", Status: "Todo"}
	task2 := &models.Todo2Task{ID: taskID2, Content: "Task 2", Status: "Todo"}

	err = CreateTask(context.Background(), task1)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	err = CreateTask(context.Background(), task2)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add mixed comments
	err = AddComments(context.Background(), taskID1, []Comment{
		{Type: CommentTypeResearch, Content: "Research 1"},
		{Type: CommentTypeNote, Content: "Note 1"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	err = AddComments(context.Background(), taskID2, []Comment{
		{Type: CommentTypeResearch, Content: "Research 2"},
		{Type: CommentTypeResult, Content: "Result 1"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Test GetCommentsByType
	researchComments, err := GetCommentsByType(context.Background(), CommentTypeResearch)
	if err != nil {
		t.Fatalf("GetCommentsByType() error = %v", err)
	}

	if len(researchComments) != 2 {
		t.Fatalf("Expected 2 research comments, got %d", len(researchComments))
	}

	for _, comment := range researchComments {
		if comment.Type != CommentTypeResearch {
			t.Errorf("Expected type %s, got %s", CommentTypeResearch, comment.Type)
		}
	}
}

func TestGetCommentsWithTypeFilter(t *testing.T) {
	initCommentsTestDB(t)
	var err error
	// Create task (valid ID)
	taskID := "T-1000006"
	task := &models.Todo2Task{
		ID:      taskID,
		Content: "Test task",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add mixed comments
	err = AddComments(context.Background(), taskID, []Comment{
		{Type: CommentTypeResearch, Content: "Research"},
		{Type: CommentTypeNote, Content: "Note"},
		{Type: CommentTypeResult, Content: "Result"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Test filtering by type
	notes, err := GetCommentsWithTypeFilter(context.Background(), taskID, CommentTypeNote)
	if err != nil {
		t.Fatalf("GetCommentsWithTypeFilter() error = %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	if notes[0].Type != CommentTypeNote {
		t.Errorf("Expected type %s, got %s", CommentTypeNote, notes[0].Type)
	}
}

func TestDeleteComment(t *testing.T) {
	initCommentsTestDB(t)
	var err error
	// Create task (valid ID)
	taskID := "T-1000007"
	task := &models.Todo2Task{
		ID:      taskID,
		Content: "Test task",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add comment
	err = AddComments(context.Background(), taskID, []Comment{
		{Type: CommentTypeNote, Content: "Comment to delete"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Get comment ID
	comments, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(comments))
	}

	commentID := comments[0].ID

	// Delete comment
	err = DeleteComment(context.Background(), commentID)
	if err != nil {
		t.Fatalf("DeleteComment() error = %v", err)
	}

	// Verify comment is deleted
	remaining, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(remaining) != 0 {
		t.Errorf("Expected 0 comments after deletion, got %d", len(remaining))
	}

	// Test deleting non-existent comment
	err = DeleteComment(context.Background(), "NONEXISTENT")
	if err == nil {
		t.Error("Expected error when deleting non-existent comment, got nil")
	}
}

func TestCommentCascadeDelete(t *testing.T) {
	initCommentsTestDB(t)
	var err error
	// Create task (valid ID)
	taskID := "T-1000008"
	task := &models.Todo2Task{
		ID:      taskID,
		Content: "Test task",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add comments
	err = AddComments(context.Background(), taskID, []Comment{
		{Type: CommentTypeNote, Content: "Comment 1"},
		{Type: CommentTypeResult, Content: "Comment 2"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify comments exist
	comments, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(comments) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(comments))
	}

	// Delete task (should cascade delete comments)
	err = DeleteTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify comments were cascade deleted
	remaining, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(remaining) != 0 {
		t.Errorf("Expected 0 comments after task deletion (cascade), got %d", len(remaining))
	}
}

func TestAddCommentsEmptyList(t *testing.T) {
	initCommentsTestDB(t)
	// Test adding empty comment list (should not error; task need not exist for empty list)
	err := AddComments(context.Background(), "T-1000009", []Comment{})
	if err != nil {
		t.Fatalf("AddComments(context.Background(), ) with empty list should not error, got %v", err)
	}
}

func TestAddCommentsWithProvidedID(t *testing.T) {
	initCommentsTestDB(t)
	// Create task (valid ID)
	taskID := "T-1000010"
	task := &models.Todo2Task{
		ID:      taskID,
		Content: "Test task",
		Status:  "Todo",
	}

	err := CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add comment with provided ID
	customID := taskID + "-C-CUSTOM"
	comments := []Comment{
		{
			ID:      customID,
			Type:    CommentTypeNote,
			Content: "Comment with custom ID",
		},
	}

	err = AddComments(context.Background(), taskID, comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify comment has custom ID
	retrieved, err := GetComments(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}

	if len(retrieved) != 1 {
		t.Fatalf("Expected 1 comment, got %d", len(retrieved))
	}

	if retrieved[0].ID != customID {
		t.Errorf("Expected ID %s, got %s", customID, retrieved[0].ID)
	}
}
