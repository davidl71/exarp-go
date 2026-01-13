package cli

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type model struct {
	tasks       []*database.Todo2Task
	cursor      int
	selected    map[int]struct{}
	status      string
	server      framework.MCPServer
	loading     bool
	err         error
}

type taskLoadedMsg struct {
	tasks []*database.Todo2Task
	err   error
}

func initialModel(server framework.MCPServer, status string) model {
	return model{
		tasks:    []*database.Todo2Task{},
		cursor:   0,
		selected: make(map[int]struct{}),
		status:   status,
		server:   server,
		loading:  true,
	}
}

func (m model) Init() tea.Cmd {
	return loadTasks(m.status)
}

func loadTasks(status string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		var tasks []*database.Todo2Task
		var err error

		if status != "" {
			tasks, err = database.GetTasksByStatus(ctx, status)
		} else {
			filters := &database.TaskFilters{}
			tasks, err = database.ListTasks(ctx, filters)
		}

		return taskLoadedMsg{tasks: tasks, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case taskLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.tasks = msg.tasks
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case "enter", " ":
			// Toggle selection
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "r":
			// Refresh tasks
			m.loading = true
			return m, loadTasks(m.status)

		case "s":
			// Show task details
			if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
				return m, showTaskDetails(m.tasks[m.cursor])
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.loading {
		return "\n  Loading tasks...\n\n"
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.\n\n", m.err)
	}

	if len(m.tasks) == 0 {
		return fmt.Sprintf("\n  No tasks found (status: %s)\n\n  Press q to quit, r to refresh.\n\n", m.status)
	}

	var b strings.Builder

	// Title
	title := "Task List"
	if m.status != "" {
		title += fmt.Sprintf(" - %s", m.status)
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Task list
	for i, task := range m.tasks {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			if _, ok := m.selected[i]; ok {
				cursor = "✓"
			}
		}

		// Task line
		line := fmt.Sprintf("%s %s", cursor, task.ID)

		// Add status badge
		if task.Status != "" {
			line += " " + statusStyle.Render(task.Status)
		}

		// Add priority badge
		if task.Priority != "" {
			priorityColor := "#3C3C3C"
			switch strings.ToLower(task.Priority) {
			case "high":
				priorityColor = "#FF5F87"
			case "medium":
				priorityColor = "#FFB800"
			case "low":
				priorityColor = "#00D4AA"
			}
			priorityStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFDF5")).
				Background(lipgloss.Color(priorityColor)).
				Padding(0, 1)
			line += " " + priorityStyle.Render(strings.ToUpper(task.Priority))
		}

		// Add task content (truncated)
		content := task.Content
		if content == "" {
			content = task.LongDescription
		}
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		if content != "" {
			line += " " + content
		}

		// Apply selected style if selected
		if m.cursor == i {
			line = selectedStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Help text
	help := helpStyle.Render("\n  ↑/↓: navigate  space/enter: select  s: details  r: refresh  q: quit\n")
	b.WriteString(help)

	return b.String()
}

func showTaskDetails(task *database.Todo2Task) tea.Cmd {
	return func() tea.Msg {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("\nTask: %s\n", task.ID))
		details.WriteString(fmt.Sprintf("Status: %s\n", task.Status))
		details.WriteString(fmt.Sprintf("Priority: %s\n", task.Priority))
		if task.Content != "" {
			details.WriteString(fmt.Sprintf("Content: %s\n", task.Content))
		}
		if task.LongDescription != "" {
			details.WriteString(fmt.Sprintf("Description: %s\n", task.LongDescription))
		}
		if len(task.Tags) > 0 {
			details.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(task.Tags, ", ")))
		}
		if len(task.Dependencies) > 0 {
			details.WriteString(fmt.Sprintf("Dependencies: %s\n", strings.Join(task.Dependencies, ", ")))
		}
		details.WriteString("\nPress any key to continue...\n")
		fmt.Print(details.String())
		return nil
	}
}

// RunTUI starts the TUI interface
func RunTUI(server framework.MCPServer, status string) error {
	// Initialize database if needed
	initializeDatabase()

	p := tea.NewProgram(initialModel(server, status))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
