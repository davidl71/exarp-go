package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
)

var (
	// Header/Title styles (top/htop-like)
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#00FF00")).
			Bold(true).
			Padding(0, 1)

	headerLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00FF00")).
				Bold(true)

	headerValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(lipgloss.Color("#00FF00"))

	// Status bar (bottom)
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#000000")).
			Padding(0, 1)

	// Task/item styles
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#008080")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#FFFF00")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))

	oldIDStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	// Priority colors (htop-like)
	highPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	mediumPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00"))

	lowPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

	// Border/separator
	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))
)

type model struct {
	tasks       []*database.Todo2Task
	cursor      int
	selected    map[int]struct{}
	status      string
	server      framework.MCPServer
	loading     bool
	err         error
	autoRefresh bool
	lastUpdate  time.Time
	projectRoot string
	projectName string

	// Terminal dimensions
	width  int
	height int

	// Config editing mode
	mode           string // "tasks" or "config"
	configSections []configSection
	configCursor   int
	configData     *config.FullConfig
	configChanged  bool
}

type configSection struct {
	name        string
	description string
	keys        []string
}

type taskLoadedMsg struct {
	tasks []*database.Todo2Task
	err   error
}

func initialModel(server framework.MCPServer, status string, projectRoot, projectName string) model {
	// Load config
	cfg, _ := config.LoadConfig(projectRoot)

	// Build config sections
	sections := []configSection{
		{name: "Timeouts", description: "Task locks, tool execution, HTTP clients", keys: []string{"task_lock_lease", "tool_default", "http_client"}},
		{name: "Thresholds", description: "Similarity, coverage, confidence, limits", keys: []string{"similarity_threshold", "min_coverage", "min_task_confidence"}},
		{name: "Tasks", description: "Default status/priority, workflow, cleanup", keys: []string{"default_status", "default_priority", "stale_threshold_hours"}},
		{name: "Database", description: "SQLite settings, connection pooling", keys: []string{"sqlite_path", "max_connections", "query_timeout"}},
		{name: "Security", description: "Rate limiting, path validation", keys: []string{"rate_limit.enabled", "max_file_size", "max_path_depth"}},
	}

	return model{
		tasks:          []*database.Todo2Task{},
		cursor:         0,
		selected:       make(map[int]struct{}),
		status:         status,
		server:         server,
		loading:        true,
		width:          80,   // Default width
		height:         24,   // Default height
		autoRefresh:    true, // Enable auto-refresh by default
		lastUpdate:     time.Now(),
		projectRoot:    projectRoot,
		projectName:    projectName,
		mode:           "tasks",
		configSections: sections,
		configCursor:   0,
		configData:     cfg,
		configChanged:  false,
	}
}

func (m model) Init() tea.Cmd {
	// Load tasks, start auto-refresh ticker, and get initial window size
	return tea.Batch(loadTasks(m.status), tick(), tea.WindowSize())
}

// tick returns a command that sends a tick message at configured interval
// Uses config for refresh interval, defaults to 5 seconds if not configured
func tick() tea.Cmd {
	// Use a reasonable default for TUI refresh (5 seconds)
	// This could be made configurable in the future
	refreshInterval := 5 * time.Second
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type tickMsg time.Time

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
	case tea.WindowSizeMsg:
		// Update terminal dimensions
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case taskLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.tasks = msg.tasks
		m.lastUpdate = time.Now()
		// Continue auto-refresh if enabled
		if m.autoRefresh {
			return m, tick()
		}
		return m, nil

	case tickMsg:
		// Auto-refresh tasks periodically (only in tasks mode)
		if m.mode == "tasks" && m.autoRefresh && !m.loading {
			m.loading = true
			return m, loadTasks(m.status)
		}
		// Continue ticking
		if m.mode == "tasks" && m.autoRefresh {
			return m, tick()
		}
		return m, nil

	case configSavedMsg:
		m.configChanged = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == "config" && m.configChanged {
				// Ask for confirmation before quitting with unsaved changes
				// For now, just quit (could add confirmation dialog later)
			}
			return m, tea.Quit

		case "c":
			// Toggle between tasks and config view
			if m.mode == "tasks" {
				m.mode = "config"
				m.configCursor = 0
			} else {
				m.mode = "tasks"
				m.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.mode == "config" {
				if m.configCursor > 0 {
					m.configCursor--
				}
			} else {
				// Task view navigation
				if len(m.tasks) > 0 && m.cursor > 0 {
					m.cursor--
				}
			}
			return m, nil

		case "down", "j":
			if m.mode == "config" {
				if m.configCursor < len(m.configSections)-1 {
					m.configCursor++
				}
			} else {
				// Task view navigation
				if len(m.tasks) > 0 && m.cursor < len(m.tasks)-1 {
					m.cursor++
				}
			}
			return m, nil

		case "enter", " ":
			if m.mode == "config" {
				// Open config section editor (for now, just show section details)
				return m, showConfigSection(m.configSections[m.configCursor], m.configData)
			} else {
				// Toggle selection (only if we have tasks)
				if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
					if _, ok := m.selected[m.cursor]; ok {
						delete(m.selected, m.cursor)
					} else {
						m.selected[m.cursor] = struct{}{}
					}
				}
			}
			return m, nil

		case "r":
			if m.mode == "config" {
				// Reload config
				cfg, err := config.LoadConfig(m.projectRoot)
				if err == nil {
					m.configData = cfg
					m.configChanged = false
				}
				return m, nil
			} else {
				// Refresh tasks
				m.loading = true
				return m, loadTasks(m.status)
			}

		case "a":
			if m.mode == "tasks" {
				// Toggle auto-refresh
				m.autoRefresh = !m.autoRefresh
				if m.autoRefresh {
					return m, tick()
				}
			}
			return m, nil

		case "s":
			if m.mode == "tasks" {
				// Show task details
				if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
					return m, showTaskDetails(m.tasks[m.cursor])
				}
			} else {
				// Save config
				return m, saveConfig(m.projectRoot, m.configData)
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.mode == "config" {
		return m.viewConfig()
	}

	return m.viewTasks()
}

func (m model) viewTasks() string {
	if m.loading {
		return "\n  Loading tasks...\n\n"
	}

	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n\n  Press q to quit.\n\n", m.err)
	}

	if len(m.tasks) == 0 {
		return fmt.Sprintf("\n  No tasks found (status: %s)\n\n  Press q to quit, r to refresh.\n\n", m.status)
	}

	// Calculate available width (account for padding)
	availableWidth := m.width - 4 // Leave some margin
	if availableWidth < 40 {
		availableWidth = 40 // Minimum width
	}

	// Determine layout based on terminal width
	useWideLayout := availableWidth >= 120
	useMediumLayout := availableWidth >= 80

	var b strings.Builder

	// Header bar (top/htop style)
	headerLine := strings.Builder{}

	// Title
	title := "TASKS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}
	if m.status != "" {
		title += fmt.Sprintf(" [%s]", strings.ToUpper(m.status))
	}
	headerLine.WriteString(headerStyle.Render(title))
	headerLine.WriteString(" ")

	// Task count
	taskCount := len(m.tasks)
	headerLine.WriteString(headerLabelStyle.Render("Tasks:"))
	headerLine.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", taskCount)))
	headerLine.WriteString(" ")

	// Selected count
	selectedCount := len(m.selected)
	if selectedCount > 0 {
		headerLine.WriteString(headerLabelStyle.Render("Selected:"))
		headerLine.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", selectedCount)))
		headerLine.WriteString(" ")
	}

	// Auto-refresh status
	if m.autoRefresh {
		lastUpdateStr := time.Since(m.lastUpdate).Round(time.Second).String()
		headerLine.WriteString(headerLabelStyle.Render("Updated:"))
		headerLine.WriteString(headerValueStyle.Render(fmt.Sprintf(" %s ago", lastUpdateStr)))
	} else {
		headerLine.WriteString(headerLabelStyle.Render("Auto-refresh:"))
		headerLine.WriteString(headerValueStyle.Render(" OFF"))
	}

	// Fill remaining space
	headerText := headerLine.String()
	if len(headerText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(headerText))
		headerText += headerValueStyle.Render(padding)
	}

	b.WriteString(headerText)
	b.WriteString("\n")

	// Separator line
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Task list - adjust layout based on terminal width
	if useWideLayout {
		// Wide layout: multi-column or side-by-side
		m.renderWideTaskList(&b, availableWidth)
	} else if useMediumLayout {
		// Medium layout: single column with more details
		m.renderMediumTaskList(&b, availableWidth)
	} else {
		// Narrow layout: single column, minimal info
		m.renderNarrowTaskList(&b, availableWidth)
	}

	// Status bar at bottom (htop style)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Status bar content
	statusBar := strings.Builder{}
	statusBar.WriteString(statusBarStyle.Render("Commands:"))
	statusBar.WriteString(" ")
	statusBar.WriteString(helpStyle.Render("↑↓/jk"))
	statusBar.WriteString(" nav  ")
	statusBar.WriteString(helpStyle.Render("Space"))
	statusBar.WriteString(" select  ")
	statusBar.WriteString(helpStyle.Render("s"))
	statusBar.WriteString(" details  ")
	statusBar.WriteString(helpStyle.Render("r"))
	statusBar.WriteString(" refresh  ")
	statusBar.WriteString(helpStyle.Render("a"))
	statusBar.WriteString(" auto-refresh  ")
	statusBar.WriteString(helpStyle.Render("c"))
	statusBar.WriteString(" config  ")
	statusBar.WriteString(helpStyle.Render("q"))
	statusBar.WriteString(" quit")

	// Fill remaining space
	statusText := statusBar.String()
	if len(statusText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(statusText))
		statusText += statusBarStyle.Render(padding)
	}

	b.WriteString(statusText)

	return b.String()
}

// renderNarrowTaskList renders tasks in a narrow terminal (< 80 chars)
func (m model) renderNarrowTaskList(b *strings.Builder, width int) {
	for i, task := range m.tasks {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			if _, ok := m.selected[i]; ok {
				cursor = "✓"
			}
		}

		// Minimal info: ID, status, truncated content
		line := fmt.Sprintf("%s %s", cursor, task.ID)

		if task.Status != "" {
			line += " " + statusStyle.Render(task.Status)
		}

		// Truncate content to fit
		content := task.Content
		if content == "" {
			content = task.LongDescription
		}
		maxContentWidth := width - len(line) - 10 // Reserve space
		if maxContentWidth > 0 && len(content) > maxContentWidth {
			content = content[:maxContentWidth-3] + "..."
		}
		if content != "" && maxContentWidth > 0 {
			line += " " + content
		}

		if m.cursor == i {
			line = selectedStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// renderMediumTaskList renders tasks in a medium terminal (80-120 chars)
func (m model) renderMediumTaskList(b *strings.Builder, width int) {
	// Column headers (htop style)
	header := fmt.Sprintf("%-4s %-14s %-10s %-8s %s", "PID", "STATUS", "PRIORITY", "PRI", "DESCRIPTION")
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	for i, task := range m.tasks {
		// Cursor indicator
		cursor := "   "
		if m.cursor == i {
			cursor = " → "
			if _, ok := m.selected[i]; ok {
				cursor = " ✓ "
			}
		}

		// Task ID
		taskID := task.ID
		if len(taskID) > 14 {
			taskID = taskID[:11] + "..."
		}

		// Status
		statusStr := task.Status
		if statusStr == "" {
			statusStr = "---"
		}
		if len(statusStr) > 10 {
			statusStr = statusStr[:7] + "..."
		}

		// Priority full name
		priorityFull := strings.ToUpper(task.Priority)
		if priorityFull == "" {
			priorityFull = "---"
		}
		if len(priorityFull) > 8 {
			priorityFull = priorityFull[:5] + "..."
		}

		// Priority short
		priorityShort := "-"
		if task.Priority != "" {
			priorityShort = strings.ToUpper(task.Priority[:1])
		}

		// Content
		content := task.Content
		if content == "" {
			content = task.LongDescription
		}
		maxContentWidth := width - 45 // Reserve space
		if maxContentWidth > 0 {
			if len(content) > maxContentWidth {
				content = content[:maxContentWidth-3] + "..."
			}
		}
		if content == "" {
			content = "(no description)"
		}

		// Build line
		line := fmt.Sprintf("%s%-4s %-14s %-10s %-8s %s", cursor, taskID, statusStr, priorityFull, priorityShort, content)

		// Apply styling
		if m.cursor == i {
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		// Apply priority color to short priority indicator
		if task.Priority != "" {
			switch strings.ToLower(task.Priority) {
			case "high":
				line = strings.Replace(line, priorityShort, highPriorityStyle.Render(priorityShort), 1)
			case "medium":
				line = strings.Replace(line, priorityShort, mediumPriorityStyle.Render(priorityShort), 1)
			case "low":
				line = strings.Replace(line, priorityShort, lowPriorityStyle.Render(priorityShort), 1)
			}
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// renderWideTaskList renders tasks in a wide terminal (>= 120 chars)
func (m model) renderWideTaskList(b *strings.Builder, width int) {
	// Column headers (htop style with more columns)
	header := fmt.Sprintf("%-4s %-16s %-12s %-10s %-4s %-50s", "PID", "STATUS", "PRIORITY", "PRI", "OLD", "DESCRIPTION")
	if width >= 160 {
		header += " TAGS"
	}
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	for i, task := range m.tasks {
		// Cursor indicator
		cursor := "   "
		if m.cursor == i {
			cursor = " → "
			if _, ok := m.selected[i]; ok {
				cursor = " ✓ "
			}
		}

		// Task ID
		taskID := task.ID
		if len(taskID) > 16 {
			taskID = taskID[:13] + "..."
		}

		// Status
		statusStr := task.Status
		if statusStr == "" {
			statusStr = "---"
		}
		if len(statusStr) > 12 {
			statusStr = statusStr[:9] + "..."
		}

		// Priority full
		priorityFull := strings.ToUpper(task.Priority)
		if priorityFull == "" {
			priorityFull = "---"
		}
		if len(priorityFull) > 10 {
			priorityFull = priorityFull[:7] + "..."
		}

		// Priority short
		priorityShort := "-"
		if task.Priority != "" {
			priorityShort = strings.ToUpper(task.Priority[:1])
		}

		// OLD indicator
		oldIndicator := "   "
		if isOldSequentialID(task.ID) {
			oldIndicator = oldIDStyle.Render("OLD")
		}

		// Content
		content := task.Content
		if content == "" {
			content = task.LongDescription
		}
		maxContentWidth := 50
		if width >= 160 {
			maxContentWidth = 60
		}
		if len(content) > maxContentWidth {
			content = content[:maxContentWidth-3] + "..."
		}
		if content == "" {
			content = "(no description)"
		}

		// Build line
		line := fmt.Sprintf("%s%-4s %-16s %-12s %-10s %-4s %-50s",
			cursor, taskID, statusStr, priorityFull, priorityShort, oldIndicator, content)

		// Add tags for very wide terminals
		if width >= 160 && len(task.Tags) > 0 {
			tagsStr := strings.Join(task.Tags, ",")
			if len(tagsStr) > 20 {
				tagsStr = tagsStr[:17] + "..."
			}
			line += " " + helpStyle.Render(tagsStr)
		}

		// Apply styling
		if m.cursor == i {
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		// Apply priority color
		if task.Priority != "" {
			switch strings.ToLower(task.Priority) {
			case "high":
				line = strings.Replace(line, priorityShort, highPriorityStyle.Render(priorityShort), 1)
			case "medium":
				line = strings.Replace(line, priorityShort, mediumPriorityStyle.Render(priorityShort), 1)
			case "low":
				line = strings.Replace(line, priorityShort, lowPriorityStyle.Render(priorityShort), 1)
			}
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

func (m model) viewConfig() string {
	var b strings.Builder
	availableWidth := m.width - 4
	if availableWidth < 40 {
		availableWidth = 40
	}

	// Header bar (top/htop style)
	headerLine := strings.Builder{}
	title := "CONFIG"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}
	headerLine.WriteString(headerStyle.Render(title))
	headerLine.WriteString(" ")

	// Config path
	if m.projectRoot != "" {
		configPath := filepath.Join(m.projectRoot, ".exarp", "config.pb")
		if len(configPath) > 40 {
			configPath = "..." + configPath[len(configPath)-37:]
		}
		headerLine.WriteString(headerLabelStyle.Render("Config:"))
		headerLine.WriteString(headerValueStyle.Render(fmt.Sprintf(" %s", configPath)))
		headerLine.WriteString(" ")
	}

	// Unsaved changes indicator
	if m.configChanged {
		headerLine.WriteString(headerLabelStyle.Render("Status:"))
		headerLine.WriteString(headerValueStyle.Render(" UNSAVED"))
	}

	// Fill remaining space
	headerText := headerLine.String()
	if len(headerText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(headerText))
		headerText += headerValueStyle.Render(padding)
	}

	b.WriteString(headerText)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Column headers
	header := fmt.Sprintf("%-4s %-20s %s", "PID", "SECTION", "DESCRIPTION")
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Config sections list
	for i, section := range m.configSections {
		// Cursor indicator
		cursor := "   "
		if m.configCursor == i {
			cursor = " → "
		}

		// Section number (like PID)
		sectionNum := fmt.Sprintf("%-3d", i+1)

		// Section name
		sectionName := section.name
		if len(sectionName) > 20 {
			sectionName = sectionName[:17] + "..."
		}

		// Description
		description := section.description
		maxDescWidth := availableWidth - 30
		if maxDescWidth > 0 && len(description) > maxDescWidth {
			description = description[:maxDescWidth-3] + "..."
		}

		// Build line
		line := fmt.Sprintf("%s%s %-20s %s", cursor, sectionNum, sectionName, description)

		// Apply styling
		if m.configCursor == i {
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	statusBar := strings.Builder{}
	statusBar.WriteString(statusBarStyle.Render("Commands:"))
	statusBar.WriteString(" ")
	statusBar.WriteString(helpStyle.Render("↑↓"))
	statusBar.WriteString(" nav  ")
	statusBar.WriteString(helpStyle.Render("Enter"))
	statusBar.WriteString(" edit  ")
	statusBar.WriteString(helpStyle.Render("s"))
	statusBar.WriteString(" save  ")
	statusBar.WriteString(helpStyle.Render("r"))
	statusBar.WriteString(" reload  ")
	statusBar.WriteString(helpStyle.Render("c"))
	statusBar.WriteString(" tasks  ")
	statusBar.WriteString(helpStyle.Render("q"))
	statusBar.WriteString(" quit")

	statusText := statusBar.String()
	if len(statusText) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-len(statusText))
		statusText += statusBarStyle.Render(padding)
	}

	b.WriteString(statusText)

	return b.String()
}

// isOldSequentialID checks if a task ID uses the old sequential format (T-1, T-2, etc.)
// vs the new epoch format (T-1768158627000)
// Old format: T- followed by a small number (< 10000, typically 1-999)
// New format: T- followed by epoch milliseconds (13 digits, typically 1.6+ trillion)
func isOldSequentialID(taskID string) bool {
	if !strings.HasPrefix(taskID, "T-") {
		return false
	}

	// Extract the number part
	numStr := strings.TrimPrefix(taskID, "T-")

	// Parse as integer
	var num int64
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return false
	}

	// Old sequential IDs are typically small numbers (< 10000)
	// Epoch milliseconds are 13 digits (1.6+ trillion)
	// Use 1000000 (1 million) as the threshold to be safe
	return num < 1000000
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

func showConfigSection(section configSection, cfg *config.FullConfig) tea.Cmd {
	return func() tea.Msg {
		var details strings.Builder
		details.WriteString(fmt.Sprintf("\nConfig Section: %s\n", section.name))
		details.WriteString(fmt.Sprintf("Description: %s\n", section.description))
		details.WriteString("\nKey values:\n")

		// Show some key values from the section
		// This is a simplified view - full editing would require more complex UI
		switch section.name {
		case "Timeouts":
			details.WriteString(fmt.Sprintf("  Task Lock Lease: %v\n", cfg.Timeouts.TaskLockLease))
			details.WriteString(fmt.Sprintf("  Tool Default: %v\n", cfg.Timeouts.ToolDefault))
			details.WriteString(fmt.Sprintf("  HTTP Client: %v\n", cfg.Timeouts.HTTPClient))
		case "Thresholds":
			details.WriteString(fmt.Sprintf("  Similarity Threshold: %.2f\n", cfg.Thresholds.SimilarityThreshold))
			details.WriteString(fmt.Sprintf("  Min Coverage: %d%%\n", cfg.Thresholds.MinCoverage))
			details.WriteString(fmt.Sprintf("  Min Task Confidence: %.2f\n", cfg.Thresholds.MinTaskConfidence))
		case "Tasks":
			details.WriteString(fmt.Sprintf("  Default Status: %s\n", cfg.Tasks.DefaultStatus))
			details.WriteString(fmt.Sprintf("  Default Priority: %s\n", cfg.Tasks.DefaultPriority))
			details.WriteString(fmt.Sprintf("  Stale Threshold: %d hours\n", cfg.Tasks.StaleThresholdHours))
		case "Database":
			details.WriteString(fmt.Sprintf("  SQLite Path: %s\n", cfg.Database.SQLitePath))
			details.WriteString(fmt.Sprintf("  Max Connections: %d\n", cfg.Database.MaxConnections))
			details.WriteString(fmt.Sprintf("  Query Timeout: %v\n", cfg.Database.QueryTimeout))
		case "Security":
			details.WriteString(fmt.Sprintf("  Rate Limit Enabled: %v\n", cfg.Security.RateLimit.Enabled))
			details.WriteString(fmt.Sprintf("  Max File Size: %d bytes\n", cfg.Thresholds.MaxFileSize))
			details.WriteString(fmt.Sprintf("  Max Path Depth: %d\n", cfg.Thresholds.MaxPathDepth))
		}

		details.WriteString("\nNote: Config is .exarp/config.pb (protobuf). Use 'exarp-go config export yaml' to edit as YAML, then 'convert yaml protobuf' to save.\n")
		details.WriteString("Press any key to continue...\n")
		fmt.Print(details.String())
		return nil
	}
}

func saveConfig(projectRoot string, cfg *config.FullConfig) tea.Cmd {
	return func() tea.Msg {
		if err := config.WriteConfigToProtobufFile(projectRoot, cfg); err != nil {
			fmt.Printf("\n❌ Error writing config file: %v\nPress any key to continue...\n", err)
			return nil
		}

		fmt.Printf("\n✅ Config saved to %s/.exarp/config.pb\nPress any key to continue...\n", projectRoot)
		return configSavedMsg{}
	}
}

type configSavedMsg struct{}

// getProjectName extracts a readable project name from the project root path
func getProjectName(projectRoot string) string {
	if projectRoot == "" {
		return ""
	}

	// Try to get module name from go.mod
	if moduleName := getModuleName(projectRoot); moduleName != "" {
		// Extract just the module name (last part after /)
		parts := strings.Split(moduleName, "/")
		return parts[len(parts)-1]
	}

	// Fallback to directory name
	return filepath.Base(projectRoot)
}

// getModuleName reads the module name from go.mod
func getModuleName(projectRoot string) string {
	goModPath := filepath.Join(projectRoot, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	// Parse "module github.com/user/project" from go.mod
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return moduleName
		}
	}
	return ""
}

// RunTUI starts the TUI interface
func RunTUI(server framework.MCPServer, status string) error {
	// Initialize database if needed (without closing it immediately)
	projectRoot, err := tools.FindProjectRoot()
	projectName := ""
	if err != nil {
		logWarn(nil, "Could not find project root", "error", err, "operation", "RunTUI", "fallback", "JSON")
	} else {
		projectName = getProjectName(projectRoot)
		// Try to use centralized config first
		fullCfg, err := config.LoadConfig(projectRoot)
		if err == nil {
			// Convert centralized config DatabaseConfig to DatabaseConfigFields
			dbCfg := database.DatabaseConfigFields{
				SQLitePath:          fullCfg.Database.SQLitePath,
				JSONFallbackPath:    fullCfg.Database.JSONFallbackPath,
				BackupPath:          fullCfg.Database.BackupPath,
				MaxConnections:      fullCfg.Database.MaxConnections,
				ConnectionTimeout:   int64(fullCfg.Database.ConnectionTimeout.Seconds()),
				QueryTimeout:        int64(fullCfg.Database.QueryTimeout.Seconds()),
				RetryAttempts:       fullCfg.Database.RetryAttempts,
				RetryInitialDelay:   int64(fullCfg.Database.RetryInitialDelay.Seconds()),
				RetryMaxDelay:       int64(fullCfg.Database.RetryMaxDelay.Seconds()),
				RetryMultiplier:     fullCfg.Database.RetryMultiplier,
				AutoVacuum:          fullCfg.Database.AutoVacuum,
				WALMode:             fullCfg.Database.WALMode,
				CheckpointInterval:  fullCfg.Database.CheckpointInterval,
				BackupRetentionDays: fullCfg.Database.BackupRetentionDays,
			}

			if err := database.InitWithCentralizedConfig(projectRoot, dbCfg); err != nil {
				logWarn(nil, "Database initialization with centralized config failed", "error", err, "operation", "RunTUI", "fallback", "legacy config")
				// Fall through to legacy init
			} else {
				// Defer close when TUI exits
				defer func() {
					if err := database.Close(); err != nil {
						logWarn(nil, "Error closing database", "error", err, "operation", "closeDatabase")
					}
				}()
				logInfo(nil, "Database initialized with centralized config", "path", fullCfg.Database.SQLitePath, "operation", "RunTUI")
				goto databaseInitialized
			}
		}

		// Fall back to legacy config
		if err := database.Init(projectRoot); err != nil {
			logWarn(nil, "Database initialization failed", "error", err, "operation", "RunTUI", "fallback", "JSON")
		} else {
			// Defer close when TUI exits
			defer func() {
				if err := database.Close(); err != nil {
					logWarn(nil, "Error closing database", "error", err, "operation", "closeDatabase")
				}
			}()
			logInfo(nil, "Database initialized", "path", projectRoot+"/.todo2/todo2.db", "operation", "RunTUI")
		}
	databaseInitialized:
	}

	p := tea.NewProgram(initialModel(server, status, projectRoot, projectName))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
