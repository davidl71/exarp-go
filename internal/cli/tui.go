package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

// Column widths for task list tabulation (header and rows must match)
const (
	colCursor   = 3  // " → " or "   "
	colIDMedium = 18 // task ID (e.g. T-1769992681538)
	colIDWide   = 20
	colStatus   = 12
	colPriority = 10
	colPRI      = 4
	colOLD      = 4
	colDescMed  = 45 // reserved for description in medium; rest is content
	colDescWide = 50
)

// truncatePad truncates s to max width with "..." or right-pads with spaces.
func truncatePad(s string, width int) string {
	if width <= 0 {
		return s
	}
	if len(s) > width {
		if width <= 3 {
			return s[:width]
		}
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}

// highlightRow pads the line to full width and applies selectedStyle so the current row is clearly highlighted.
func highlightRow(line string, width int, selected bool) string {
	if !selected {
		return line
	}
	visible := lipgloss.Width(line)
	if visible < width {
		line += strings.Repeat(" ", width-visible)
	}
	return selectedStyle.Render(line)
}

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
	mode           string // "tasks", "config", "scorecard", or "handoffs"
	configSections []configSection
	configCursor   int
	configData     *config.FullConfig
	configChanged  bool

	// Scorecard view (project health)
	scorecardText      string
	scorecardRecs      []string // recommendations from scorecard
	scorecardRecCursor int      // selected recommendation index
	scorecardLoading   bool
	scorecardErr       error
	scorecardRunOutput string // last "run recommendation" output or error

	// Handoffs view (session handoff notes)
	handoffText    string
	handoffLoading bool
	handoffErr     error

	// Help overlay
	showHelp bool
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

type scorecardLoadedMsg struct {
	text            string
	recommendations []string
	err             error
}

type handoffLoadedMsg struct {
	text string
	err  error
}

type runRecommendationResultMsg struct {
	output string
	err    error
}

func initialModel(server framework.MCPServer, status string, projectRoot, projectName string) model {
	// Ensure config file exists so TUI config view can load and save (initialize if necessary)
	pbPath := filepath.Join(projectRoot, ".exarp", "config.pb")
	if _, err := os.Stat(pbPath); os.IsNotExist(err) {
		_ = config.WriteConfigToProtobufFile(projectRoot, config.GetDefaults())
	}
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
		tasks:            []*database.Todo2Task{},
		cursor:           0,
		selected:         make(map[int]struct{}),
		status:           status,
		server:           server,
		loading:          true,
		width:            80,   // Default width
		height:           24,   // Default height
		autoRefresh:      true, // Enable auto-refresh by default
		lastUpdate:       time.Now(),
		projectRoot:      projectRoot,
		projectName:      projectName,
		mode:             "tasks",
		configSections:   sections,
		configCursor:     0,
		configData:       cfg,
		configChanged:    false,
		scorecardLoading: false,
		showHelp:         false,
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
			// Default: open tasks only (Todo + In Progress), matching CLI task list
			todo, errTodo := database.GetTasksByStatus(ctx, "Todo")
			inProgress, errIP := database.GetTasksByStatus(ctx, "In Progress")
			if errTodo != nil {
				err = errTodo
			} else if errIP != nil {
				err = errIP
			} else {
				tasks = append(tasks, todo...)
				tasks = append(tasks, inProgress...)
			}
		}

		return taskLoadedMsg{tasks: tasks, err: err}
	}
}

func loadScorecard(projectRoot string) tea.Cmd {
	return func() tea.Msg {
		if projectRoot == "" {
			return scorecardLoadedMsg{err: fmt.Errorf("no project root")}
		}
		ctx := context.Background()
		var combined strings.Builder
		var recommendations []string

		// Go scorecard (when Go project)
		if tools.IsGoProject() {
			opts := &tools.ScorecardOptions{FastMode: true}
			scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
			if err != nil {
				return scorecardLoadedMsg{err: err}
			}
			combined.WriteString("=== Go Scorecard ===\n\n")
			combined.WriteString(tools.FormatGoScorecard(scorecard))
			recommendations = scorecard.Recommendations
		}

		// Project overview (always; includes health when Go, project/tasks/codebase for all)
		overviewText, err := tools.GetOverviewText(ctx, projectRoot)
		if err != nil {
			// If we already have Go scorecard, append error note; else fail
			if combined.Len() > 0 {
				combined.WriteString("\n\n=== Project Overview ===\n\n(overview failed: ")
				combined.WriteString(err.Error())
				combined.WriteString(")")
			} else {
				return scorecardLoadedMsg{err: err}
			}
		} else {
			if combined.Len() > 0 {
				combined.WriteString("\n\n")
			}
			combined.WriteString("=== Project Overview ===\n\n")
			combined.WriteString(overviewText)
		}

		return scorecardLoadedMsg{text: combined.String(), recommendations: recommendations}
	}
}

// loadHandoffs fetches session handoff list via the session tool and returns a handoffLoadedMsg.
func loadHandoffs(server framework.MCPServer) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return handoffLoadedMsg{err: fmt.Errorf("no server")}
		}
		ctx := context.Background()
		args := map[string]interface{}{"action": "handoff", "sub_action": "list"}
		argsBytes, err := json.Marshal(args)
		if err != nil {
			return handoffLoadedMsg{err: err}
		}
		result, err := server.CallTool(ctx, "session", argsBytes)
		if err != nil {
			return handoffLoadedMsg{err: err}
		}
		var b strings.Builder
		for _, c := range result {
			b.WriteString(c.Text)
			b.WriteString("\n")
		}
		return handoffLoadedMsg{text: strings.TrimSpace(b.String()), err: nil}
	}
}

// recommendationToCommand maps scorecard recommendation text to a runnable command (exec name + args).
// Prefers Makefile targets (make <target>) when available; ok is false for recommendations with no single command.
func recommendationToCommand(rec string) (name string, args []string, ok bool) {
	rec = strings.TrimSpace(rec)
	switch {
	case strings.Contains(rec, "go mod tidy"):
		return "make", []string{"go-mod-tidy"}, true
	case strings.Contains(rec, "go fmt"):
		return "make", []string{"go-fmt"}, true
	case strings.Contains(rec, "go vet"):
		return "make", []string{"go-vet"}, true
	case strings.Contains(rec, "Fix Go build"):
		return "make", []string{"build"}, true
	case strings.Contains(rec, "golangci-lint issues"):
		return "make", []string{"golangci-lint-fix"}, true
	case strings.Contains(rec, "failing Go tests"), strings.Contains(rec, "go test"):
		return "make", []string{"test"}, true
	case strings.Contains(rec, "govulncheck"):
		return "make", []string{"govulncheck"}, true
	case strings.Contains(rec, "test coverage"):
		return "make", []string{"test-coverage"}, true
	default:
		return "", nil, false
	}
}

func runRecommendationCmd(projectRoot, rec string) tea.Cmd {
	return func() tea.Msg {
		name, args, ok := recommendationToCommand(rec)
		if !ok {
			return runRecommendationResultMsg{output: "", err: fmt.Errorf("no runnable command for this recommendation")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = projectRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			return runRecommendationResultMsg{output: string(out), err: err}
		}
		return runRecommendationResultMsg{output: string(out), err: nil}
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

	case scorecardLoadedMsg:
		m.scorecardLoading = false
		m.scorecardErr = msg.err
		m.scorecardRunOutput = ""
		if msg.err == nil {
			m.scorecardText = msg.text
			m.scorecardRecs = msg.recommendations
			if len(m.scorecardRecs) > 0 {
				m.scorecardRecCursor = 0
			}
		}
		return m, nil

	case handoffLoadedMsg:
		m.handoffLoading = false
		m.handoffErr = msg.err
		if msg.err == nil {
			m.handoffText = msg.text
		}
		return m, nil

	case runRecommendationResultMsg:
		if msg.err != nil {
			m.scorecardRunOutput = "Error: " + msg.err.Error()
		} else {
			m.scorecardRunOutput = strings.TrimSpace(msg.output)
			if m.scorecardRunOutput == "" {
				m.scorecardRunOutput = "(command completed)"
			}
		}
		// Refresh scorecard so updated state (e.g. after make test, make lint) is shown
		m.scorecardLoading = true
		return m, loadScorecard(m.projectRoot)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.mode == "config" && m.configChanged {
				// Ask for confirmation before quitting with unsaved changes
				// For now, just quit (could add confirmation dialog later)
			}
			return m, tea.Quit

		case "?", "h":
			m.showHelp = !m.showHelp
			return m, nil

		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			return m, nil
		}

		// When help is open, ignore all other keys (handled above)
		if m.showHelp {
			return m, nil
		}

		switch msg.String() {
		case "p":
			// Toggle between tasks and scorecard (project health)
			if m.mode == "scorecard" {
				m.mode = "tasks"
				m.cursor = 0
			} else if m.mode == "tasks" {
				m.mode = "scorecard"
				m.scorecardLoading = true
				m.scorecardErr = nil
				m.scorecardText = ""
				return m, loadScorecard(m.projectRoot)
			} else if m.mode == "handoffs" {
				m.mode = "tasks"
				m.cursor = 0
			}
			return m, nil

		case "H":
			// Toggle handoffs view (session handoff notes)
			if m.mode == "handoffs" {
				m.mode = "tasks"
				m.cursor = 0
				return m, nil
			}
			m.mode = "handoffs"
			m.handoffLoading = true
			m.handoffErr = nil
			m.handoffText = ""
			return m, loadHandoffs(m.server)

		case "c":
			// Toggle between tasks and config view
			if m.mode == "tasks" {
				m.mode = "config"
				m.configCursor = 0
			} else if m.mode == "config" {
				m.mode = "tasks"
				m.cursor = 0
			} else if m.mode == "handoffs" {
				m.mode = "tasks"
				m.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.mode == "scorecard" {
				if len(m.scorecardRecs) > 0 && m.scorecardRecCursor > 0 {
					m.scorecardRecCursor--
				}
				return m, nil
			}
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
			if m.mode == "scorecard" {
				if len(m.scorecardRecs) > 0 && m.scorecardRecCursor < len(m.scorecardRecs)-1 {
					m.scorecardRecCursor++
				}
				return m, nil
			}
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

		case "enter", " ", "e":
			if m.mode == "scorecard" {
				if len(m.scorecardRecs) > 0 && m.scorecardRecCursor < len(m.scorecardRecs) {
					rec := m.scorecardRecs[m.scorecardRecCursor]
					if _, _, ok := recommendationToCommand(rec); ok {
						return m, runRecommendationCmd(m.projectRoot, rec)
					}
				}
				return m, nil
			}
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
			} else if m.mode == "scorecard" {
				// Refresh scorecard
				m.scorecardLoading = true
				return m, loadScorecard(m.projectRoot)
			} else if m.mode == "handoffs" {
				// Refresh handoffs
				m.handoffLoading = true
				return m, loadHandoffs(m.server)
			} else {
				// Refresh tasks
				m.loading = true
				return m, loadTasks(m.status)
			}

		case "a":
			if m.mode == "scorecard" {
				return m, nil
			}
			if m.mode == "tasks" {
				// Toggle auto-refresh
				m.autoRefresh = !m.autoRefresh
				if m.autoRefresh {
					return m, tick()
				}
			}
			return m, nil

		case "s":
			if m.mode == "scorecard" {
				return m, nil
			}
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
	if m.showHelp {
		return m.viewHelp()
	}
	if m.mode == "config" {
		return m.viewConfig()
	}
	if m.mode == "scorecard" {
		return m.viewScorecard()
	}
	if m.mode == "handoffs" {
		return m.viewHandoffs()
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
	// Use full width minus small margin for better window utilization
	availableWidth := m.width - 2
	if availableWidth < 40 {
		availableWidth = 40
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
	statusBar.WriteString(helpStyle.Render("p"))
	statusBar.WriteString(" scorecard  ")
	statusBar.WriteString(helpStyle.Render("?/h"))
	statusBar.WriteString(" help  ")
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
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// renderMediumTaskList renders tasks in a medium terminal (80-120 chars)
func (m model) renderMediumTaskList(b *strings.Builder, width int) {
	// Column headers aligned with column width constants
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %s",
		colCursor, "", colIDMedium, "ID", colStatus, "STATUS", colPriority, "PRIORITY", colPRI, "PRI", "DESCRIPTION")
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	minDescWidth := width - (colCursor + 1 + colIDMedium + 1 + colStatus + 1 + colPriority + 1 + colPRI + 1)
	if minDescWidth < 10 {
		minDescWidth = 10
	}

	for i, task := range m.tasks {
		// Cursor indicator (fixed width)
		cursor := "   "
		if m.cursor == i {
			cursor = " → "
			if _, ok := m.selected[i]; ok {
				cursor = " ✓ "
			}
		}

		taskID := truncatePad(task.ID, colIDMedium)
		statusStr := task.Status
		if statusStr == "" {
			statusStr = "---"
		}
		statusStr = truncatePad(statusStr, colStatus)
		priorityFull := strings.ToUpper(task.Priority)
		if priorityFull == "" {
			priorityFull = "---"
		}
		priorityFull = truncatePad(priorityFull, colPriority)
		priorityShort := "-"
		if task.Priority != "" {
			priorityShort = strings.ToUpper(task.Priority[:1])
		}
		priorityShort = truncatePad(priorityShort, colPRI)

		content := task.Content
		if content == "" {
			content = task.LongDescription
		}
		if len(content) > minDescWidth {
			content = content[:minDescWidth-3] + "..."
		}
		if content == "" {
			content = "(no description)"
		}

		line := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %s",
			colCursor, cursor, colIDMedium, taskID, colStatus, statusStr, colPriority, priorityFull, colPRI, priorityShort, content)

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

		if m.cursor == i {
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// renderWideTaskList renders tasks in a wide terminal (>= 120 chars)
func (m model) renderWideTaskList(b *strings.Builder, width int) {
	// Column headers aligned with column width constants
	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		colCursor, "", colIDWide, "ID", colStatus, "STATUS", colPriority, "PRIORITY", colPRI, "PRI", colOLD, "OLD", colDescWide, "DESCRIPTION")
	if width >= 160 {
		header += " TAGS"
	}
	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	maxDescWidth := colDescWide
	if width >= 160 {
		maxDescWidth = 60
	}
	tagsWidth := 0
	if width >= 160 {
		tagsWidth = width - (colCursor + 1 + colIDWide + 1 + colStatus + 1 + colPriority + 1 + colPRI + 1 + colOLD + 1 + maxDescWidth + 1)
		if tagsWidth < 5 {
			tagsWidth = 0
		}
	}

	for i, task := range m.tasks {
		cursor := "   "
		if m.cursor == i {
			cursor = " → "
			if _, ok := m.selected[i]; ok {
				cursor = " ✓ "
			}
		}

		taskID := truncatePad(task.ID, colIDWide)
		statusStr := task.Status
		if statusStr == "" {
			statusStr = "---"
		}
		statusStr = truncatePad(statusStr, colStatus)
		priorityFull := strings.ToUpper(task.Priority)
		if priorityFull == "" {
			priorityFull = "---"
		}
		priorityFull = truncatePad(priorityFull, colPriority)
		priorityShort := "-"
		if task.Priority != "" {
			priorityShort = strings.ToUpper(task.Priority[:1])
		}
		priorityShort = truncatePad(priorityShort, colPRI)

		oldStr := "   "
		if isOldSequentialID(task.ID) {
			oldStr = "OLD"
		}
		oldIndicator := truncatePad(oldStr, colOLD)

		content := task.Content
		if content == "" {
			content = task.LongDescription
		}
		if len(content) > maxDescWidth {
			content = content[:maxDescWidth-3] + "..."
		}
		if content == "" {
			content = "(no description)"
		}
		content = truncatePad(content, maxDescWidth)

		line := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %s",
			colCursor, cursor, colIDWide, taskID, colStatus, statusStr, colPriority, priorityFull, colPRI, priorityShort, colOLD, oldIndicator, content)

		if tagsWidth > 0 && len(task.Tags) > 0 {
			tagsStr := strings.Join(task.Tags, ",")
			if len(tagsStr) > tagsWidth {
				tagsStr = tagsStr[:tagsWidth-3] + "..."
			}
			line += " " + helpStyle.Render(tagsStr)
		}

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
		if isOldSequentialID(task.ID) {
			line = strings.Replace(line, "OLD", oldIDStyle.Render("OLD"), 1)
		}

		if m.cursor == i {
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

func (m model) viewConfig() string {
	var b strings.Builder
	availableWidth := m.width - 2
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

		// Apply styling and full-row highlight for current line
		if m.configCursor == i {
			line = highlightRow(line, availableWidth, true)
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

func (m model) viewHandoffs() string {
	availableWidth := m.width - 2
	if availableWidth < 40 {
		availableWidth = 40
	}
	maxTextLines := m.height - 8
	if maxTextLines < 4 {
		maxTextLines = 4
	}

	var b strings.Builder
	title := "SESSION HANDOFFS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}
	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("H=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=refresh"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if m.handoffLoading {
		b.WriteString("\n  Loading handoffs...\n\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  q quit"))
		return b.String()
	}
	if m.handoffErr != nil {
		b.WriteString(fmt.Sprintf("\n  Error: %v\n\n", m.handoffErr))
		b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))
		return b.String()
	}
	if m.handoffText == "" {
		b.WriteString("\n  No handoff notes.\n\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))
		return b.String()
	}

	lines := strings.Split(m.handoffText, "\n")
	textLinesShown := 0
	for _, line := range lines {
		if textLinesShown >= maxTextLines {
			b.WriteString(helpStyle.Render("  ... (run 'exarp-go session handoffs' for full output)"))
			b.WriteString("\n")
			break
		}
		if len(line) > availableWidth {
			line = line[:availableWidth-3] + "..."
		}
		b.WriteString(normalStyle.Render(line))
		b.WriteString("\n")
		textLinesShown++
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))
	return b.String()
}

func (m model) viewScorecard() string {
	// Use full width minus small margin for better window utilization
	availableWidth := m.width - 2
	if availableWidth < 40 {
		availableWidth = 40
	}
	// Reserve vertical space for header, recommendations, last run, status bar
	maxTextLines := m.height - 14
	if maxTextLines < 4 {
		maxTextLines = 4
	}

	var b strings.Builder

	// Header
	title := "SCORECARD"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}
	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("p=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=refresh"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if m.scorecardLoading {
		b.WriteString("\n  Loading scorecard...\n\n")
		b.WriteString(statusBarStyle.Render("Commands: p back  q quit"))
		return b.String()
	}
	if m.scorecardErr != nil {
		b.WriteString(fmt.Sprintf("\n  Error: %v\n\n", m.scorecardErr))
		b.WriteString(statusBarStyle.Render("Commands: p back  r refresh  q quit"))
		return b.String()
	}

	// Show scorecard text; cap lines so recommendations and status stay visible
	lines := strings.Split(m.scorecardText, "\n")
	textLinesShown := 0
	for _, line := range lines {
		if textLinesShown >= maxTextLines {
			b.WriteString(helpStyle.Render("  ... (run report/scorecard for full output)"))
			b.WriteString("\n")
			break
		}
		if len(line) > availableWidth {
			for len(line) > 0 && textLinesShown < maxTextLines {
				end := availableWidth
				if end > len(line) {
					end = len(line)
				}
				b.WriteString(normalStyle.Render(line[:end]))
				b.WriteString("\n")
				line = line[end:]
				textLinesShown++
			}
			if textLinesShown >= maxTextLines {
				b.WriteString(helpStyle.Render("  ... (run report/scorecard for full output)"))
				b.WriteString("\n")
				break
			}
		} else {
			b.WriteString(normalStyle.Render(line))
			b.WriteString("\n")
			textLinesShown++
		}
	}

	// Interactive recommendations: select and run
	if len(m.scorecardRecs) > 0 {
		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(normalStyle.Render("Run recommendation (↑↓ select, e or Enter run):"))
		b.WriteString("\n")
		for i, rec := range m.scorecardRecs {
			prefix := "   "
			if i == m.scorecardRecCursor {
				_, _, ok := recommendationToCommand(rec)
				if ok {
					prefix = " ▶ " // runnable
				} else {
					prefix = " → " // selected but not runnable
				}
			}
			line := rec
			if len(line) > availableWidth-len(prefix)-2 {
				line = line[:availableWidth-len(prefix)-5] + "..."
			}
			fullLine := prefix + line
			if i == m.scorecardRecCursor {
				b.WriteString(highlightRow(fullLine, availableWidth, true))
			} else {
				b.WriteString(normalStyle.Render(fullLine))
			}
			b.WriteString("\n")
		}
	}
	if m.scorecardRunOutput != "" {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Last run:"))
		b.WriteString("\n")
		for _, line := range strings.Split(m.scorecardRunOutput, "\n") {
			if len(line) > availableWidth {
				line = line[:availableWidth-3] + "..."
			}
			b.WriteString(normalStyle.Render("  " + line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Commands: p back  r refresh  e implement selected  q quit"))
	return b.String()
}

func (m model) viewHelp() string {
	availableWidth := m.width - 2
	if availableWidth < 40 {
		availableWidth = 40
	}
	var b strings.Builder
	b.WriteString(headerStyle.Render("HELP - Key bindings"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render("Navigation:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("↑↓ / j / k"))
	b.WriteString("     Move cursor (tasks, config)\n  ")
	b.WriteString(helpStyle.Render("Enter / Space"))
	b.WriteString("  Toggle selection (tasks) or open (config)\n\n")
	b.WriteString(normalStyle.Render("Views:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("c"))
	b.WriteString("  Switch to Config\n  ")
	b.WriteString(helpStyle.Render("p"))
	b.WriteString("  Switch to Scorecard (project health)\n  ")
	b.WriteString(helpStyle.Render("H"))
	b.WriteString("  Switch to Session handoffs\n\n")
	b.WriteString(normalStyle.Render("Actions:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("r"))
	b.WriteString("  Refresh (tasks, config, or scorecard)\n  ")
	b.WriteString(helpStyle.Render("e"))
	b.WriteString("  Implement selected recommendation (scorecard)\n  ")
	b.WriteString(helpStyle.Render("a"))
	b.WriteString("  Toggle auto-refresh (tasks only)\n  ")
	b.WriteString(helpStyle.Render("s"))
	b.WriteString("  Show task details (tasks) or Save config (config)\n\n")
	b.WriteString(normalStyle.Render("Other:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("q"))
	b.WriteString("  Quit\n  ")
	b.WriteString(helpStyle.Render("? / h"))
	b.WriteString("  This help\n\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Press ? or h or Esc to close help"))
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
		EnsureConfigAndDatabase(projectRoot)
		if database.DB != nil {
			defer func() {
				if err := database.Close(); err != nil {
					logWarn(nil, "Error closing database", "error", err, "operation", "closeDatabase")
				}
			}()
		}
	}

	p := tea.NewProgram(initialModel(server, status, projectRoot, projectName))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
