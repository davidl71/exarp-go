package database

import (
	"context"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestAddComments(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create a task first (required for foreign key)
	task := &models.Todo2Task{
		ID:      "T-COMMENT-1",
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

	err = AddComments(context.Background(), "T-COMMENT-1", comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify comment was added
	retrieved, err := GetComments(context.Background(), "T-COMMENT-1")
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
	if retrieved[0].TaskID != "T-COMMENT-1" {
		t.Errorf("Expected taskID T-COMMENT-1, got %s", retrieved[0].TaskID)
	}
	if retrieved[0].ID == "" {
		t.Error("Expected comment ID to be generated, got empty string")
	}
}

func TestAddCommentsBatch(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:      "T-COMMENT-2",
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

	err = AddComments(context.Background(), "T-COMMENT-2", comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify all comments were added
	retrieved, err := GetComments(context.Background(), "T-COMMENT-2")
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
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:      "T-COMMENT-3",
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
	err = AddComments(context.Background(), "T-COMMENT-3", comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Test GetComments
	retrieved, err := GetComments(context.Background(), "T-COMMENT-3")
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
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create tasks
	task1 := &models.Todo2Task{ID: "T-COMMENT-4", Content: "Task 1", Status: "Todo"}
	task2 := &models.Todo2Task{ID: "T-COMMENT-5", Content: "Task 2", Status: "Todo"}

	err = CreateTask(context.Background(), task1)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	err = CreateTask(context.Background(), task2)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add mixed comments
	err = AddComments(context.Background(), "T-COMMENT-4", []Comment{
		{Type: CommentTypeResearch, Content: "Research 1"},
		{Type: CommentTypeNote, Content: "Note 1"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	err = AddComments(context.Background(), "T-COMMENT-5", []Comment{
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
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:      "T-COMMENT-6",
		Content: "Test task",
		Status:  "Todo",
	}
	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add mixed comments
	err = AddComments(context.Background(), "T-COMMENT-6", []Comment{
		{Type: CommentTypeResearch, Content: "Research"},
		{Type: CommentTypeNote, Content: "Note"},
		{Type: CommentTypeResult, Content: "Result"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Test filtering by type
	notes, err := GetCommentsWithTypeFilter(context.Background(), "T-COMMENT-6", CommentTypeNote)
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
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:      "T-COMMENT-7",
		Content: "Test task",
		Status:  "Todo",
	}
	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add comment
	err = AddComments(context.Background(), "T-COMMENT-7", []Comment{
		{Type: CommentTypeNote, Content: "Comment to delete"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Get comment ID
	comments, err := GetComments(context.Background(), "T-COMMENT-7")
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
	remaining, err := GetComments(context.Background(), "T-COMMENT-7")
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
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:      "T-COMMENT-8",
		Content: "Test task",
		Status:  "Todo",
	}
	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add comments
	err = AddComments(context.Background(), "T-COMMENT-8", []Comment{
		{Type: CommentTypeNote, Content: "Comment 1"},
		{Type: CommentTypeResult, Content: "Comment 2"},
	})
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify comments exist
	comments, err := GetComments(context.Background(), "T-COMMENT-8")
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("Expected 2 comments, got %d", len(comments))
	}

	// Delete task (should cascade delete comments)
	err = DeleteTask(context.Background(), "T-COMMENT-8")
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify comments were cascade deleted
	remaining, err := GetComments(context.Background(), "T-COMMENT-8")
	if err != nil {
		t.Fatalf("GetComments() error = %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("Expected 0 comments after task deletion (cascade), got %d", len(remaining))
	}
}

func TestAddCommentsEmptyList(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Test adding empty comment list (should not error)
	err = AddComments(context.Background(), "T-COMMENT-9", []Comment{})
	if err != nil {
		t.Fatalf("AddComments(context.Background(), ) with empty list should not error, got %v", err)
	}
}

func TestAddCommentsWithProvidedID(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:      "T-COMMENT-10",
		Content: "Test task",
		Status:  "Todo",
	}
	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Add comment with provided ID
	customID := "T-COMMENT-10-C-CUSTOM"
	comments := []Comment{
		{
			ID:      customID,
			Type:    CommentTypeNote,
			Content: "Comment with custom ID",
		},
	}

	err = AddComments(context.Background(), "T-COMMENT-10", comments)
	if err != nil {
		t.Fatalf("AddComments() error = %v", err)
	}

	// Verify comment has custom ID
	retrieved, err := GetComments(context.Background(), "T-COMMENT-10")
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

