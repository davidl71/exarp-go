// tui_command_palette.go — Command palette UI for TUI (suggestions, input, execution).
package cli

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// commandPaletteSuggestions returns available commands for the command palette.
var commandPaletteSuggestions = []string{
	"refresh - Reload tasks",
	"sort:status - Sort by status",
	"sort:priority - Sort by priority",
	"sort:updated - Sort by updated",
	"sort:hierarchy - Sort by hierarchy",
	"filter:todo - Filter by Todo",
	"filter:inprogress - Filter by In Progress",
	"filter:done - Filter by Done",
	"filter:review - Filter by Review",
	"list - Toggle list view",
	"table - Toggle table view",
	"help - Show help",
	"quit - Quit application",
}

// newCommandPalette creates a new text input for the command palette.
func newCommandPalette() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Type a command... (Esc to close)"
	ti.Prompt = ">"
	ti.SetWidth(40)
	ti.ShowSuggestions = true
	ti.SetSuggestions(commandPaletteSuggestions)
	return ti
}

// viewCommandPalette renders the command palette overlay.
func (m model) viewCommandPalette() string {
	m.commandPalette.SetWidth(m.effectiveWidth() - 4)
	return "\n  " + m.commandPalette.View() + "\n\n  ↑/↓: navigate • Enter: execute • Esc: close\n"
}

// executeCommand parses and executes a command from the command palette.
func (m *model) executeCommand(cmd string) tea.Cmd {
	base := strings.TrimSpace(strings.Split(cmd, " - ")[0])

	switch base {
	case "refresh":
		m.loading = true
		return loadTasks(m.server, m.status)
	case "sort:status":
		m.sortOrder = SortByStatus
		m.sortAsc = true
	case "sort:priority":
		m.sortOrder = SortByPriority
		m.sortAsc = true
	case "sort:updated":
		m.sortOrder = SortByUpdated
		m.sortAsc = false
	case "sort:hierarchy":
		m.sortOrder = SortByHierarchy
		m.sortAsc = true
	case "filter:todo":
		m.status = "Todo"
	case "filter:inprogress":
		m.status = "In Progress"
	case "filter:done":
		m.status = "Done"
	case "filter:review":
		m.status = "Review"
	case "list":
		m.useBubbleList = !m.useBubbleList
		if m.useBubbleList {
			m.syncTaskComponents()
		}
	case "table":
		m.useBubbleTable = !m.useBubbleTable
		if m.useBubbleTable {
			m.syncTaskComponents()
		}
	case "help":
		m.showHelp = !m.showHelp
	case "quit":
		return tea.Quit
	}

	return nil
}
