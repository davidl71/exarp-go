package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/tests/fixtures"
)

// setupMockServer creates a mock MCP server for testing.
func setupMockServer(t *testing.T) framework.MCPServer {
	t.Helper()
	return fixtures.NewMockServer("test-server")
}

// createTestTasks creates sample tasks for testing.
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

// TestTUIInitialState tests the initial state of the TUI model.
func TestTUIInitialState(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)

	// Verify initial state
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	if m.mode != "tasks" {
		t.Errorf("Expected mode 'tasks', got %s", m.mode)
	}

	if !m.loading {
		t.Error("Expected loading=true initially")
	}

	if !m.autoRefresh {
		t.Error("Expected autoRefresh=true by default")
	}
}

// TestTUINavigation tests keyboard navigation in task view.
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
			m := initialModel(server, "", "/test", "test-project", 0, 0)
			m.tasks = createTestTasks(tt.taskCount)
			m.loading = false
			m.cursor = tt.startPos

			updated, _ := m.Update(tt.key)
			updatedModel := updated.(model)

			if updatedModel.cursor != tt.wantPos {
				t.Errorf("Expected cursor %d, got %d", tt.wantPos, updatedModel.cursor)
			}
		})
	}
}

// TestTUIModeSwitching tests switching between tasks and config modes.
func TestTUIModeSwitching(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.tasks = createTestTasks(3)
	m.loading = false

	// Switch to config mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

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

// TestTUISelection tests task selection with enter/space.
func TestTUISelection(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.tasks = createTestTasks(3)
	m.loading = false
	m.cursor = 1

	// Select task at cursor
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

// TestTUIWindowResize tests handling of window size changes.
func TestTUIWindowResize(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)

	// Simulate window resize
	updated, _ := m.Update(tea.WindowSizeMsg{
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

// TestTUIRefresh tests manual refresh with 'r' key.
func TestTUIRefresh(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.tasks = createTestTasks(2)
	m.loading = false

	// Trigger refresh
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updatedModel := updated.(model)

	if !updatedModel.loading {
		t.Error("Expected loading=true after refresh")
	}

	if cmd == nil {
		t.Error("Expected command to be returned for refresh")
	}
}

// TestTUIAutoRefreshToggle tests toggling auto-refresh with 'a' key.
func TestTUIAutoRefreshToggle(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.tasks = createTestTasks(2)
	m.loading = false

	// Toggle auto-refresh off
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

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

// TestTUIQuit tests quitting with 'q' or Ctrl+C.
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
			m := initialModel(server, "", "/test", "test-project", 0, 0)
			m.tasks = createTestTasks(2)
			m.loading = false

			updated, cmd := m.Update(tt.key)
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

// TestTUITaskLoading tests task loading message handling.
func TestTUITaskLoading(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)

	// Simulate tasks loaded successfully
	tasks := createTestTasks(3)
	updated, _ := m.Update(taskLoadedMsg{tasks: tasks, err: nil})
	updatedModel := updated.(model)

	if updatedModel.loading {
		t.Error("Expected loading=false after tasks loaded")
	}

	if len(updatedModel.tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(updatedModel.tasks))
	}
}

// TestTUITaskLoadingError tests error handling during task loading.
func TestTUITaskLoadingError(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)

	// Simulate task loading error
	err := context.DeadlineExceeded
	updated, _ := m.Update(taskLoadedMsg{tasks: nil, err: err})
	updatedModel := updated.(model)

	if updatedModel.loading {
		t.Error("Expected loading=false after error")
	}

	if updatedModel.err == nil {
		t.Error("Expected error to be set")
	}
}

// TestTUIConfigNavigation tests navigation in config mode.
func TestTUIConfigNavigation(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.mode = "config"

	// Navigate down in config
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})

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

// TestTUIEmptyState tests rendering when no tasks are available.
func TestTUIEmptyState(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "Todo", "/test", "test-project", 0, 0)
	m.tasks = []*database.Todo2Task{}
	m.loading = false

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
	// Should contain "No tasks found"
	if len(view) < 10 {
		t.Error("Expected meaningful view content")
	}
}

// TestTUIViewRendering tests that view renders correctly with tasks.
func TestTUIViewRendering(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.tasks = createTestTasks(5)
	m.loading = false
	m.width = 80
	m.height = 24

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
	// View should contain task information
	if len(view) < 50 {
		t.Error("Expected substantial view content with tasks")
	}
}

// TestTUITaskDetailRecordsOutput simulates pressing 's' to open task detail,
// captures the view output, and records it to testdata for inspection.
func TestTUITaskDetailRecordsOutput(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	tasks := []*database.Todo2Task{
		{
			ID:              "T-1771245906548",
			Content:         "Fix tools health and hooks tests",
			Status:          "Todo",
			Priority:        "medium",
			LongDescription: "Objective: Fix internal/tools TestHandleHealthDocs, TestHandleHealthCICD. Acceptance: Health docs/DOD/CICD tests pass or are skipped.",
			Tags:            []string{"testing", "health"},
			Dependencies:    []string{},
		},
	}
	m.tasks = tasks
	m.loading = false
	m.cursor = 0
	m.width = 80
	m.height = 24

	// Simulate pressing 's' to show task details
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	updatedModel, ok := updated.(model)

	if !ok {
		t.Fatalf("Update returned %T, want model", updated)
	}

	if updatedModel.mode != "taskDetail" {
		t.Errorf("mode = %q, want taskDetail", updatedModel.mode)
	}

	if updatedModel.taskDetailTask == nil {
		t.Fatal("taskDetailTask is nil after pressing s")
	}

	if updatedModel.taskDetailTask.ID != "T-1771245906548" {
		t.Errorf("taskDetailTask.ID = %q, want T-1771245906548", updatedModel.taskDetailTask.ID)
	}

	// Capture view output
	output := updatedModel.View()
	if output == "" {
		t.Fatal("View() returned empty string")
	}

	// Record to testdata
	testdataDir := filepath.Join("testdata")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Logf("Could not create testdata dir: %v", err)
	} else {
		outPath := filepath.Join(testdataDir, "tui_task_detail_output.txt")
		if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
			t.Logf("Could not write output file: %v", err)
		} else {
			t.Logf("Recorded view output to %s", outPath)
		}
	}

	// Log first 500 runes for visibility in test run
	t.Logf("View output (first 500 chars):\n%s", truncateForLog(output, 500))

	// Sanity checks on content
	if !strings.Contains(output, "TASK DETAIL") {
		t.Error("output should contain 'TASK DETAIL'")
	}

	if !strings.Contains(output, "T-1771245906548") {
		t.Error("output should contain task ID T-1771245906548")
	}

	if !strings.Contains(output, "Fix tools health and hooks tests") {
		t.Error("output should contain task content")
	}

	if !strings.Contains(output, "Esc/Enter/Space") {
		t.Error("output should contain close hint")
	}
}

func truncateForLog(s string, max int) string {
	if len(s) <= max {
		return s
	}

	return s[:max] + "..."
}
