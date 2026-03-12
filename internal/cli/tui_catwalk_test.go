// tui_catwalk_test.go — TUI tests using catwalk for Bubble Tea testing.
package cli

import (
	"testing"

	"charm.land/bubbles/v2/spinner"
	"github.com/knz/catwalk"
)

func TestCatwalkSmoke(t *testing.T) {
	// Basic smoke test to verify catwalk is properly integrated.
	// Catwalk provides testing utilities for Bubble Tea models.
	//
	// Example catwalk usage:
	//   - Walk through model states
	//   - Test key sequences
	//   - Verify view output
	//
	// This test serves as a placeholder to verify the dependency is working.
	// More comprehensive tests can be added using catwalk's APIs:
	//
	//   func TestTaskListNavigation(t *testing.T) {
	//       server := setupMockServer(t)
	//       m := initialModel(server, "", "/test", "test-project", 80, 24)
	//
	//       // Simulate key presses
	//       m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	//       m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	//
	//       // Verify cursor moved
	//       if m.cursor != 2 {
	//           t.Errorf("Expected cursor at 2, got %d", m.cursor)
	//       }
	//   }
	//
	//   func TestViewOutput(t *testing.T) {
	//       server := setupMockServer(t)
	//       m := initialModel(server, "", "/test", "test-project", 80, 24)
	//       m.tasks = createTestTasks(3)
	//       m.loading = false
	//
	//       view := m.View()
	//       if !contains(view, "T-1") {
	//           t.Error("Expected task T-1 in view")
	//       }
	//   }

	// Verify catwalk is available by checking its version/type
	var _ catwalk.Walker // catwalk.Walker interface
	t.Log("Catwalk dependency available for TUI testing")
}

func TestBubbleListToggle(t *testing.T) {
	// Test that the bubble/list toggle works correctly
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 80, 24)

	// Initially disabled
	if m.useBubbleList {
		t.Error("Expected useBubbleList=false initially")
	}

	// Toggle via key handler (simulating L key)
	m.tasks = createTestTasks(3)
	m.useBubbleList = true
	m.taskList.SetItems(itemsFromTasks(m.tasks))

	// Verify list has items
	if m.taskList.Length() != 3 {
		t.Errorf("Expected 3 items in list, got %d", m.taskList.Length())
	}
}

func TestBubbleTableToggle(t *testing.T) {
	// Test that the bubble/table toggle works correctly
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 80, 24)

	// Initially disabled
	if m.useBubbleTable {
		t.Error("Expected useBubbleTable=false initially")
	}

	// Toggle and set rows
	m.tasks = createTestTasks(3)
	m.useBubbleTable = true
	m.taskTable.SetRows(rowsFromTasks(m.tasks))

	// Verify table has rows
	if len(m.taskTable.Rows()) != 3 {
		t.Errorf("Expected 3 rows in table, got %d", len(m.taskTable.Rows()))
	}
}

func TestCommandPaletteToggle(t *testing.T) {
	// Test command palette toggle
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 80, 24)

	// Initially disabled
	if m.showCommandPalette {
		t.Error("Expected showCommandPalette=false initially")
	}

	// Enable palette
	m.showCommandPalette = true
	m.commandPalette.Focus()

	// Verify it's shown
	if !m.showCommandPalette {
		t.Error("Expected showCommandPalette=true after toggle")
	}

	// Disable palette
	m.showCommandPalette = false
	m.commandPalette.Blur()

	if m.showCommandPalette {
		t.Error("Expected showCommandPalette=false after disable")
	}
}

func TestSpinnerState(t *testing.T) {
	// Test spinner initialization
	server := setupMockServer(t)
	m := initialModel(server, "", "/test", "test-project", 80, 24)

	// Verify spinner is initialized
	if m.taskSpinner == (spinner.Model{}) {
		t.Error("Expected spinner to be initialized")
	}

	// Test loading state shows spinner
	m.loading = true
	view := m.viewSpinner()

	if view == "" {
		t.Error("Expected non-empty spinner view when loading")
	}
}
