package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestCreateTask(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	task := &models.Todo2Task{
		ID:              "T-1",
		Content:         "Test task",
		LongDescription: "This is a test task",
		Status:          "Todo",
		Priority:        "medium",
		Tags:            []string{"test", "database"},
		Dependencies:    []string{},
		Completed:       false,
	}

	// Test
	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
	}

	// Verify
	retrieved, err := GetTask(context.Background(), "T-1")
	if err != nil {
		t.Fatalf("GetTask(context.Background(), ) error = %v", err)
	}

	if retrieved.ID != task.ID {
		t.Errorf("Expected ID %s, got %s", task.ID, retrieved.ID)
	}

	if retrieved.Content != task.Content {
		t.Errorf("Expected Content %s, got %s", task.Content, retrieved.Content)
	}

	if retrieved.Status != task.Status {
		t.Errorf("Expected Status %s, got %s", task.Status, retrieved.Status)
	}

	if len(retrieved.Tags) != len(task.Tags) {
		t.Errorf("Expected %d tags, got %d", len(task.Tags), len(retrieved.Tags))
	}
}

func TestGetTask(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create task first
	task := &models.Todo2Task{
		ID:       "T-2",
		Content:  "Test task 2",
		Status:   "In Progress",
		Priority: "high",
		Tags:     []string{"test"},
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
	}

	// Test
	retrieved, err := GetTask(context.Background(), "T-2")
	if err != nil {
		t.Fatalf("GetTask(context.Background(), ) error = %v", err)
	}

	if retrieved.ID != "T-2" {
		t.Errorf("Expected ID T-2, got %s", retrieved.ID)
	}

	if retrieved.Status != "In Progress" {
		t.Errorf("Expected Status In Progress, got %s", retrieved.Status)
	}
}

func TestUpdateTask(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:       "T-3",
		Content:  "Original content",
		Status:   "Todo",
		Priority: "low",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
	}

	// Update task
	task.Content = "Updated content"
	task.Status = "Done"
	task.Priority = "high"
	task.Tags = []string{"updated"}

	err = UpdateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("UpdateTask(context.Background(), ) error = %v", err)
	}

	// Verify
	retrieved, err := GetTask(context.Background(), "T-3")
	if err != nil {
		t.Fatalf("GetTask(context.Background(), ) error = %v", err)
	}

	if retrieved.Content != "Updated content" {
		t.Errorf("Expected Content 'Updated content', got '%s'", retrieved.Content)
	}

	if retrieved.Status != "Done" {
		t.Errorf("Expected Status 'Done', got '%s'", retrieved.Status)
	}

	if len(retrieved.Tags) != 1 || retrieved.Tags[0] != "updated" {
		t.Errorf("Expected tag 'updated', got %v", retrieved.Tags)
	}
}

func TestDeleteTask(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create task
	task := &models.Todo2Task{
		ID:      "T-4",
		Content: "Task to delete",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
	}

	// Delete task
	err = DeleteTask(context.Background(), "T-4")
	if err != nil {
		t.Fatalf("DeleteTask(context.Background(), ) error = %v", err)
	}

	// Verify task is deleted
	_, err = GetTask(context.Background(), "T-4")
	if err == nil {
		t.Error("Expected error when getting deleted task, got nil")
	}
}

func TestListTasks(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create multiple tasks
	tasks := []*models.Todo2Task{
		{ID: "T-5", Content: "Task 1", Status: "Todo", Priority: "high"},
		{ID: "T-6", Content: "Task 2", Status: "In Progress", Priority: "medium"},
		{ID: "T-7", Content: "Task 3", Status: "Done", Priority: "low"},
	}

	for _, task := range tasks {
		err = CreateTask(context.Background(), task)
		if err != nil {
			t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
		}
	}

	// Test ListTasks without filters
	allTasks, err := ListTasks(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTasks(context.Background(), ) error = %v", err)
	}

	if len(allTasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(allTasks))
	}

	// Test filter by status
	statusFilter := "Done"
	filters := &TaskFilters{Status: &statusFilter}

	doneTasks, err := ListTasks(context.Background(), filters)
	if err != nil {
		t.Fatalf("ListTasks(context.Background(), ) error = %v", err)
	}

	if len(doneTasks) != 1 {
		t.Errorf("Expected 1 Done task, got %d", len(doneTasks))
	}

	if doneTasks[0].ID != "T-7" {
		t.Errorf("Expected task T-7, got %s", doneTasks[0].ID)
	}
}

func TestGetTasksByStatus(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create tasks with different statuses
	tasks := []*models.Todo2Task{
		{ID: "T-8", Content: "Todo task", Status: "Todo"},
		{ID: "T-9", Content: "Another todo", Status: "Todo"},
		{ID: "T-10", Content: "In progress task", Status: "In Progress"},
	}

	for _, task := range tasks {
		err = CreateTask(context.Background(), task)
		if err != nil {
			t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
		}
	}

	// Test
	todoTasks, err := GetTasksByStatus(context.Background(), "Todo")
	if err != nil {
		t.Fatalf("GetTasksByStatus(context.Background(), ) error = %v", err)
	}

	if len(todoTasks) != 2 {
		t.Errorf("Expected 2 Todo tasks, got %d", len(todoTasks))
	}
}

func TestGetDependencies(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create tasks
	task1 := &models.Todo2Task{ID: "T-11", Content: "Parent", Status: "Todo"}
	task2 := &models.Todo2Task{ID: "T-12", Content: "Child", Status: "Todo", Dependencies: []string{"T-11"}}

	err = CreateTask(context.Background(), task1)
	if err != nil {
		t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
	}

	err = CreateTask(context.Background(), task2)
	if err != nil {
		t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
	}

	// Test
	deps, err := GetDependencies("T-12")
	if err != nil {
		t.Fatalf("GetDependencies() error = %v", err)
	}

	if len(deps) != 1 {
		t.Fatalf("Expected 1 dependency, got %d", len(deps))
	}

	if deps[0] != "T-11" {
		t.Errorf("Expected dependency T-11, got %s", deps[0])
	}

	// Test dependents
	dependents, err := GetDependents("T-11")
	if err != nil {
		t.Fatalf("GetDependents() error = %v", err)
	}

	if len(dependents) != 1 {
		t.Fatalf("Expected 1 dependent, got %d", len(dependents))
	}

	if dependents[0] != "T-12" {
		t.Errorf("Expected dependent T-12, got %s", dependents[0])
	}
}

func TestGetTasksByTag(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create tasks with tags
	tasks := []*models.Todo2Task{
		{ID: "T-13", Content: "Task 1", Status: "Todo", Tags: []string{"backend", "database"}},
		{ID: "T-14", Content: "Task 2", Status: "Todo", Tags: []string{"frontend"}},
		{ID: "T-15", Content: "Task 3", Status: "Todo", Tags: []string{"backend", "api"}},
	}

	for _, task := range tasks {
		err = CreateTask(context.Background(), task)
		if err != nil {
			t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
		}
	}

	// Test
	backendTasks, err := GetTasksByTag(context.Background(), "backend")
	if err != nil {
		t.Fatalf("GetTasksByTag(context.Background(), ) error = %v", err)
	}

	if len(backendTasks) != 2 {
		t.Errorf("Expected 2 tasks with 'backend' tag, got %d", len(backendTasks))
	}
}

func TestTaskWithMetadata(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Create task with metadata
	task := &models.Todo2Task{
		ID:       "T-16",
		Content:  "Task with metadata",
		Status:   "Todo",
		Metadata: map[string]interface{}{"custom": "value", "number": 42},
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask(context.Background(), ) error = %v", err)
	}

	// Retrieve and verify metadata
	retrieved, err := GetTask(context.Background(), "T-16")
	if err != nil {
		t.Fatalf("GetTask(context.Background(), ) error = %v", err)
	}

	if retrieved.Metadata == nil {
		t.Fatal("Expected metadata, got nil")
	}

	if retrieved.Metadata["custom"] != "value" {
		t.Errorf("Expected metadata['custom'] = 'value', got %v", retrieved.Metadata["custom"])
	}
}

func TestSerializeTaskMetadata(t *testing.T) {
	task := &Todo2Task{
		ID:       "T-serial",
		Content:  "x",
		Metadata: map[string]interface{}{"k": "v"},
	}

	metadataJSON, metadataProtobuf, metadataFormat, err := SerializeTaskMetadata(task)
	if err != nil {
		t.Fatalf("SerializeTaskMetadata() error = %v", err)
	}

	if metadataFormat != "protobuf" && metadataFormat != "json" {
		t.Errorf("metadataFormat want protobuf or json, got %q", metadataFormat)
	}

	if metadataJSON == "" {
		t.Error("metadataJSON should be non-empty when task has metadata")
	}

	if metadataFormat == "protobuf" && len(metadataProtobuf) == 0 {
		t.Error("metadataProtobuf should be non-empty when format is protobuf")
	}
	// Round-trip via DeserializeTaskMetadata
	got := DeserializeTaskMetadata(metadataJSON, metadataProtobuf, metadataFormat)
	if got == nil || got["k"] != "v" {
		t.Errorf("DeserializeTaskMetadata round-trip: got %v, want map with k=v", got)
	}
}

func TestDeserializeTaskMetadata(t *testing.T) {
	// Empty inputs
	if got := DeserializeTaskMetadata("", nil, ""); got != nil {
		t.Errorf("DeserializeTaskMetadata(empty) = %v, want nil", got)
	}
	// JSON fallback
	got := DeserializeTaskMetadata(`{"a":1}`, nil, "json")
	if got == nil || got["a"] != float64(1) {
		t.Errorf("DeserializeTaskMetadata(json) = %v, want map a=1", got)
	}
}

func TestIsValidTaskID(t *testing.T) {
	tests := []struct {
		id   string
		want bool
	}{
		{"", false},
		{"T-NaN", false},
		{"T-undefined", false},
		{"T-123", true},
		{"T-1768158627000", true},
		{"T-0", true},
		{"T-1a", false},
		{"X-123", false},
		{"T-", false},
	}
	for _, tt := range tests {
		if got := IsValidTaskID(tt.id); got != tt.want {
			t.Errorf("IsValidTaskID(%q) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

func TestCreateTaskReplacesInvalidID(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	task := &models.Todo2Task{
		ID:       "T-NaN",
		Content:  "Task with invalid ID",
		Status:   "Todo",
		Priority: "medium",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	// CreateTask should have replaced T-NaN with a generated ID
	if task.ID == "T-NaN" {
		t.Error("CreateTask should have replaced T-NaN with generated ID")
	}

	if !IsValidTaskID(task.ID) {
		t.Errorf("task.ID after CreateTask = %q, should be valid", task.ID)
	}
	// Task should be retrievable by the new ID
	_, err = GetTask(context.Background(), task.ID)
	if err != nil {
		t.Errorf("GetTask(%q) error = %v", task.ID, err)
	}
}

// TestForeignKeysEnabled verifies that foreign key constraints are enabled
// and properly reject invalid dependency references.
func TestForeignKeysEnabled(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	// Setup
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	// Verify foreign keys are enabled
	db, err := GetDB()
	if err != nil {
		t.Fatalf("GetDB() error = %v", err)
	}

	var fkEnabled int

	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("Failed to check foreign_keys PRAGMA: %v", err)
	}

	if fkEnabled != 1 {
		t.Errorf("Expected foreign_keys = 1, got %d", fkEnabled)
	}

	// Test: Create task with invalid dependency should fail
	invalidTask := &models.Todo2Task{
		ID:           "T-17",
		Content:      "Task with invalid dependency",
		Status:       "Todo",
		Dependencies: []string{"T-NONEXISTENT"},
	}

	err = CreateTask(context.Background(), invalidTask)
	if err == nil {
		t.Fatal("Expected error when creating task with invalid dependency, got nil")
	}

	if !strings.Contains(err.Error(), "foreign key") && !strings.Contains(err.Error(), "FOREIGN KEY constraint") {
		t.Errorf("Expected foreign key constraint error, got: %v", err)
	}

	// Verify task was not created (transaction should have rolled back)
	_, err = GetTask(context.Background(), "T-17")
	if err == nil {
		t.Error("Expected task T-17 not to exist after failed creation, but it exists")
	}

	// Test: Create task with valid dependency should succeed
	parentTask := &models.Todo2Task{
		ID:      "T-18",
		Content: "Parent task",
		Status:  "Todo",
	}

	err = CreateTask(context.Background(), parentTask)
	if err != nil {
		t.Fatalf("CreateTask() for parent task error = %v", err)
	}

	validTask := &models.Todo2Task{
		ID:           "T-19",
		Content:      "Task with valid dependency",
		Status:       "Todo",
		Dependencies: []string{"T-18"},
	}

	err = CreateTask(context.Background(), validTask)
	if err != nil {
		t.Fatalf("CreateTask() for valid task error = %v", err)
	}

	// Verify valid task was created
	retrieved, err := GetTask(context.Background(), "T-19")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if len(retrieved.Dependencies) != 1 || retrieved.Dependencies[0] != "T-18" {
		t.Errorf("Expected dependency T-18, got %v", retrieved.Dependencies)
	}

	// Test: Update task with invalid dependency should fail
	taskToUpdate := &models.Todo2Task{
		ID:           "T-19",
		Content:      "Updated task",
		Status:       "Todo",
		Dependencies: []string{"T-NONEXISTENT"},
	}

	err = UpdateTask(context.Background(), taskToUpdate)
	if err == nil {
		t.Fatal("Expected error when updating task with invalid dependency, got nil")
	}

	if !strings.Contains(err.Error(), "foreign key") && !strings.Contains(err.Error(), "FOREIGN KEY constraint") {
		t.Errorf("Expected foreign key constraint error, got: %v", err)
	}

	// Verify original dependency is still intact (transaction should have rolled back)
	retrieved, err = GetTask(context.Background(), "T-19")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if len(retrieved.Dependencies) != 1 || retrieved.Dependencies[0] != "T-18" {
		t.Errorf("Expected dependency T-18 to remain after failed update, got %v", retrieved.Dependencies)
	}
}

func TestIsVersionMismatchError(t *testing.T) {
	if IsVersionMismatchError(nil) {
		t.Error("IsVersionMismatchError(nil) should be false")
	}

	if IsVersionMismatchError(errors.New("other error")) {
		t.Error("IsVersionMismatchError(other) should be false")
	}

	if !IsVersionMismatchError(ErrVersionMismatch) {
		t.Error("IsVersionMismatchError(ErrVersionMismatch) should be true")
	}

	if !IsVersionMismatchError(fmt.Errorf("wrap: %w", ErrVersionMismatch)) {
		t.Error("IsVersionMismatchError(wrapped ErrVersionMismatch) should be true")
	}
}

func TestCheckUpdateConflict(t *testing.T) {
	testDBMu.Lock()
	defer testDBMu.Unlock()
	tmpDir := t.TempDir()

	err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defer Close()

	taskID := "T-3000001"
	task := &models.Todo2Task{
		ID:       taskID,
		Content:  "Task for conflict detection",
		Status:   "Todo",
		Priority: "medium",
	}

	err = CreateTask(context.Background(), task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// After CreateTask, DB version is 1
	expectedVer := int64(1)

	// No conflict when version matches
	hasConflict, currentVer, err := CheckUpdateConflict(context.Background(), taskID, expectedVer)
	if err != nil {
		t.Fatalf("CheckUpdateConflict() error = %v", err)
	}

	if hasConflict {
		t.Error("CheckUpdateConflict(matching version) should report no conflict")
	}

	if currentVer != expectedVer {
		t.Errorf("currentVersion = %d, want %d", currentVer, expectedVer)
	}

	// Conflict when version differs
	hasConflict, currentVer, err = CheckUpdateConflict(context.Background(), taskID, expectedVer-1)
	if err != nil {
		t.Fatalf("CheckUpdateConflict() error = %v", err)
	}

	if !hasConflict {
		t.Error("CheckUpdateConflict(stale version) should report conflict")
	}

	if currentVer != expectedVer {
		t.Errorf("currentVersion = %d, want %d", currentVer, expectedVer)
	}

	// Not found
	_, _, err = CheckUpdateConflict(context.Background(), "T-NOTFOUND", 0)
	if err == nil {
		t.Error("CheckUpdateConflict(nonexistent task) should error")
	}
}
