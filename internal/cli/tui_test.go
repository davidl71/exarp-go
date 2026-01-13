package cli

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/tests/fixtures"
)

// setupMockServer creates a mock MCP server for testing
func setupMockServer(t *testing.T) framework.MCPServer {
	t.Helper()
	return fixtures.NewMockServer("test-server")
}

// createTestTasks creates sample tasks for testing
func createTestTasks(count int) []*database.Todo2Task {
	tasks := make([]*database.Todo2Task, count)
	for i := 0; i < count; i++ {
		tasks[i] = &database.Todo2Task{
			ID:              "T-" + string(rune('1'+i)),
			Content:         "Test Task " + string(rune('1'+i)),
			Status:          "Todo",
			Priority:        "medium",
			LongDescription: "Test task description",
			Tags:            []string{"test"},
			Dependencies:    []string{},
			Completed:       false,
		}
	}
	return tasks
}

// TestTUIInitialState tests the initial state of the TUI model
func TestTUIInitialState(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")

	// Verify initial state
	if model.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", model.cursor)
	}
	if model.mode != "tasks" {
		t.Errorf("Expected mode 'tasks', got %s", model.mode)
	}
	if !model.loading {
		t.Error("Expected loading=true initially")
	}
	if !model.autoRefresh {
		t.Error("Expected autoRefresh=true by default")
	}
}

// TestTUINavigation tests keyboard navigation in task view
func TestTUINavigation(t *testing.T) {
	tests := []struct {
		name      string
		key       tea.KeyMsg
		startPos  int
		taskCount int
		wantPos   int
	}{
		{
			name:      "down arrow from start",
			key:       tea.KeyMsg{Type: tea.KeyDown},
			startPos:  0,
			taskCount: 3,
			wantPos:   1,
		},
		{
			name:      "j key from start",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			startPos:  0,
			taskCount: 3,
			wantPos:   1,
		},
		{
			name:      "up arrow from middle",
			key:       tea.KeyMsg{Type: tea.KeyUp},
			startPos:  1,
			taskCount: 3,
			wantPos:   0,
		},
		{
			name:      "k key from middle",
			key:       tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			startPos:  1,
			taskCount: 3,
			wantPos:   0,
		},
		{
			name:      "down arrow at end",
			key:       tea.KeyMsg{Type: tea.KeyDown},
			startPos:  2,
			taskCount: 3,
			wantPos:   2, // Should not go beyond end
		},
		{
			name:      "up arrow at start",
			key:       tea.KeyMsg{Type: tea.KeyUp},
			startPos:  0,
			taskCount: 3,
			wantPos:   0, // Should not go below 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t)
			model := initialModel(server, "", "/test", "test-project")
			model.tasks = createTestTasks(tt.taskCount)
			model.loading = false
			model.cursor = tt.startPos

			updated, _ := model.Update(tt.key)
			updatedModel := updated.(model)

			if updatedModel.cursor != tt.wantPos {
				t.Errorf("Expected cursor %d, got %d", tt.wantPos, updatedModel.cursor)
			}
		})
	}
}

// TestTUIModeSwitching tests switching between tasks and config modes
func TestTUIModeSwitching(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")
	model.tasks = createTestTasks(3)
	model.loading = false

	// Switch to config mode
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updatedModel := updated.(model)
	if updatedModel.mode != "config" {
		t.Errorf("Expected mode 'config', got %s", updatedModel.mode)
	}
	if updatedModel.configCursor != 0 {
		t.Errorf("Expected config cursor at 0, got %d", updatedModel.configCursor)
	}

	// Switch back to tasks mode
	updated2, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updatedModel2 := updated2.(model)
	if updatedModel2.mode != "tasks" {
		t.Errorf("Expected mode 'tasks', got %s", updatedModel2.mode)
	}
	if updatedModel2.cursor != 0 {
		t.Errorf("Expected task cursor at 0, got %d", updatedModel2.cursor)
	}
}

// TestTUISelection tests task selection with enter/space
func TestTUISelection(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")
	model.tasks = createTestTasks(3)
	model.loading = false
	model.cursor = 1

	// Select task at cursor
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel := updated.(model)
	if _, ok := updatedModel.selected[1]; !ok {
		t.Error("Expected task 1 to be selected")
	}

	// Deselect task
	updated2, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel2 := updated2.(model)
	if _, ok := updatedModel2.selected[1]; ok {
		t.Error("Expected task 1 to be deselected")
	}

	// Test space key also toggles selection
	updated3, _ := updatedModel2.Update(tea.KeyMsg{Type: tea.KeySpace})
	updatedModel3 := updated3.(model)
	if _, ok := updatedModel3.selected[1]; !ok {
		t.Error("Expected task 1 to be selected with space key")
	}
}

// TestTUIWindowResize tests handling of window size changes
func TestTUIWindowResize(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")

	// Simulate window resize
	updated, _ := model.Update(tea.WindowSizeMsg{
		Width:  120,
		Height: 40,
	})
	updatedModel := updated.(model)

	if updatedModel.width != 120 {
		t.Errorf("Expected width 120, got %d", updatedModel.width)
	}
	if updatedModel.height != 40 {
		t.Errorf("Expected height 40, got %d", updatedModel.height)
	}
}

// TestTUIRefresh tests manual refresh with 'r' key
func TestTUIRefresh(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")
	model.tasks = createTestTasks(2)
	model.loading = false

	// Trigger refresh
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updatedModel := updated.(model)

	if !updatedModel.loading {
		t.Error("Expected loading=true after refresh")
	}
	if cmd == nil {
		t.Error("Expected command to be returned for refresh")
	}
}

// TestTUIAutoRefreshToggle tests toggling auto-refresh with 'a' key
func TestTUIAutoRefreshToggle(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")
	model.tasks = createTestTasks(2)
	model.loading = false

	// Toggle auto-refresh off
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updatedModel := updated.(model)
	if updatedModel.autoRefresh {
		t.Error("Expected autoRefresh=false after toggle")
	}
	if cmd != nil {
		t.Error("Expected no command when disabling auto-refresh")
	}

	// Toggle auto-refresh on
	updated2, cmd2 := updatedModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	updatedModel2 := updated2.(model)
	if !updatedModel2.autoRefresh {
		t.Error("Expected autoRefresh=true after toggle")
	}
	if cmd2 == nil {
		t.Error("Expected command when enabling auto-refresh")
	}
}

// TestTUIQuit tests quitting with 'q' or Ctrl+C
func TestTUIQuit(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupMockServer(t)
			model := initialModel(server, "", "/test", "test-project")
			model.tasks = createTestTasks(2)
			model.loading = false

			updated, cmd := model.Update(tt.key)
			_ = updated // Updated model not needed for quit test

			if cmd == nil {
				t.Error("Expected quit command")
			} else {
				// Verify it's a quit command
				if _, ok := cmd().(tea.QuitMsg); !ok {
					t.Error("Expected tea.Quit command")
				}
			}
		})
	}
}

// TestTUITaskLoading tests task loading message handling
func TestTUITaskLoading(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")

	// Simulate tasks loaded successfully
	tasks := createTestTasks(3)
	updated, _ := model.Update(taskLoadedMsg{tasks: tasks, err: nil})
	updatedModel := updated.(model)

	if updatedModel.loading {
		t.Error("Expected loading=false after tasks loaded")
	}
	if len(updatedModel.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(updatedModel.tasks))
	}
}

// TestTUITaskLoadingError tests error handling during task loading
func TestTUITaskLoadingError(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")

	// Simulate task loading error
	err := context.DeadlineExceeded
	updated, _ := model.Update(taskLoadedMsg{tasks: nil, err: err})
	updatedModel := updated.(model)

	if updatedModel.loading {
		t.Error("Expected loading=false after error")
	}
	if updatedModel.err == nil {
		t.Error("Expected error to be set")
	}
}

// TestTUIConfigNavigation tests navigation in config mode
func TestTUIConfigNavigation(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")
	model.mode = "config"

	// Navigate down in config
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel := updated.(model)
	if updatedModel.configCursor != 1 {
		t.Errorf("Expected config cursor 1, got %d", updatedModel.configCursor)
	}

	// Navigate up in config
	updated2, _ := updatedModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	updatedModel2 := updated2.(model)
	if updatedModel2.configCursor != 0 {
		t.Errorf("Expected config cursor 0, got %d", updatedModel2.configCursor)
	}
}

// TestTUIEmptyState tests rendering when no tasks are available
func TestTUIEmptyState(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "Todo", "/test", "test-project")
	model.tasks = []*database.Todo2Task{}
	model.loading = false

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
	// Should contain "No tasks found"
	if len(view) < 10 {
		t.Error("Expected meaningful view content")
	}
}

// TestTUIViewRendering tests that view renders correctly with tasks
func TestTUIViewRendering(t *testing.T) {
	server := setupMockServer(t)
	model := initialModel(server, "", "/test", "test-project")
	model.tasks = createTestTasks(5)
	model.loading = false
	model.width = 80
	model.height = 24

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
	// View should contain task information
	if len(view) < 50 {
		t.Error("Expected substantial view content with tasks")
	}
}
