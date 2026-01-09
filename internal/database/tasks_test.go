package database

import (
	"testing"

	"github.com/davidl71/exarp-go/internal/models"
)

func TestCreateTask(t *testing.T) {
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
	err = CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Verify
	retrieved, err := GetTask("T-1")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
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
	err = CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Test
	retrieved, err := GetTask("T-2")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrieved.ID != "T-2" {
		t.Errorf("Expected ID T-2, got %s", retrieved.ID)
	}
	if retrieved.Status != "In Progress" {
		t.Errorf("Expected Status In Progress, got %s", retrieved.Status)
	}
}

func TestUpdateTask(t *testing.T) {
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
	err = CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Update task
	task.Content = "Updated content"
	task.Status = "Done"
	task.Priority = "high"
	task.Tags = []string{"updated"}
	err = UpdateTask(task)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	// Verify
	retrieved, err := GetTask("T-3")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
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
	err = CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Delete task
	err = DeleteTask("T-4")
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify task is deleted
	_, err = GetTask("T-4")
	if err == nil {
		t.Error("Expected error when getting deleted task, got nil")
	}
}

func TestListTasks(t *testing.T) {
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
		err = CreateTask(task)
		if err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
	}

	// Test ListTasks without filters
	allTasks, err := ListTasks(nil)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(allTasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(allTasks))
	}

	// Test filter by status
	statusFilter := "Done"
	filters := &TaskFilters{Status: &statusFilter}
	doneTasks, err := ListTasks(filters)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(doneTasks) != 1 {
		t.Errorf("Expected 1 Done task, got %d", len(doneTasks))
	}
	if doneTasks[0].ID != "T-7" {
		t.Errorf("Expected task T-7, got %s", doneTasks[0].ID)
	}
}

func TestGetTasksByStatus(t *testing.T) {
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
		err = CreateTask(task)
		if err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
	}

	// Test
	todoTasks, err := GetTasksByStatus("Todo")
	if err != nil {
		t.Fatalf("GetTasksByStatus() error = %v", err)
	}

	if len(todoTasks) != 2 {
		t.Errorf("Expected 2 Todo tasks, got %d", len(todoTasks))
	}
}

func TestGetDependencies(t *testing.T) {
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

	err = CreateTask(task1)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	err = CreateTask(task2)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
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
		err = CreateTask(task)
		if err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
	}

	// Test
	backendTasks, err := GetTasksByTag("backend")
	if err != nil {
		t.Fatalf("GetTasksByTag() error = %v", err)
	}

	if len(backendTasks) != 2 {
		t.Errorf("Expected 2 tasks with 'backend' tag, got %d", len(backendTasks))
	}
}

func TestTaskWithMetadata(t *testing.T) {
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

	err = CreateTask(task)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Retrieve and verify metadata
	retrieved, err := GetTask("T-16")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if retrieved.Metadata == nil {
		t.Fatal("Expected metadata, got nil")
	}
	if retrieved.Metadata["custom"] != "value" {
		t.Errorf("Expected metadata['custom'] = 'value', got %v", retrieved.Metadata["custom"])
	}
}

