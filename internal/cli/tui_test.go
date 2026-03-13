package cli

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
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
			Status:          models.StatusTodo,
			Priority:        models.PriorityMedium,
			LongDescription: "Test task description",
			Tags:            []string{"test"},
			Dependencies:    []string{},
			Completed:       false,
		}
	}

	return tasks
}

// keyPress creates a tea.KeyPressMsg for a special key
func keyPress(code rune) tea.KeyMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}

// runeKey creates a tea.KeyPressMsg for a character key
func runeKey(r rune) tea.KeyMsg {
	return tea.KeyPressMsg(tea.Key{Code: r, Text: string(r)})
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
			key:       keyPress(tea.KeyDown),
			startPos:  0,
			taskCount: 3,
			wantPos:   1,
		},
		{
			name:      "j key from start",
			key:       runeKey('j'),
			startPos:  0,
			taskCount: 3,
			wantPos:   1,
		},
		{
			name:      "up arrow from middle",
			key:       keyPress(tea.KeyUp),
			startPos:  1,
			taskCount: 3,
			wantPos:   0,
		},
		{
			name:      "k key from middle",
			key:       runeKey('k'),
			startPos:  1,
			taskCount: 3,
			wantPos:   0,
		},
		{
			name:      "down arrow at end",
			key:       keyPress(tea.KeyDown),
			startPos:  2,
			taskCount: 3,
			wantPos:   2,
		},
		{
			name:      "up arrow at start",
			key:       keyPress(tea.KeyUp),
			startPos:  0,
			taskCount: 3,
			wantPos:   0,
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

	updated, _ := m.Update(runeKey('c'))

	updatedModel := updated.(model)
	if updatedModel.mode != "config" {
		t.Errorf("Expected mode 'config', got %s", updatedModel.mode)
	}

	if updatedModel.configCursor != 0 {
		t.Errorf("Expected config cursor at 0, got %d", updatedModel.configCursor)
	}

	updated2, _ := updatedModel.Update(runeKey('c'))

	updatedModel2 := updated2.(model)
	if updatedModel2.mode != "tasks" {
		t.Errorf("Expected mode 'tasks', got %s", updatedModel2.mode)
	}

	if updatedModel2.cursor != 0 {
		t.Errorf("Expected task cursor at 0, got %d", updatedModel2.cursor)
	}
}

// TestTUISelection tests task selection with enter key.
func TestTUISelection(t *testing.T) {
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 0, 0)
	m.tasks = createTestTasks(3)
	m.loading = false
	m.cursor = 1

	// Select task at cursor
	updated, _ := m.Update(keyPress(tea.KeyEnter))

	updatedModel := updated.(model)
	if _, ok := updatedModel.selected[1]; !ok {
		t.Error("Expected task 1 to be selected")
	}

	// Deselect task
	updated2, _ := updatedModel.Update(keyPress(tea.KeyEnter))

	updatedModel2 := updated2.(model)
	if _, ok := updatedModel2.selected[1]; ok {
		t.Error("Expected task 1 to be deselected")
	}
}
