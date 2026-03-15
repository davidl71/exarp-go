// tui.go — Bubbletea TUI: core model struct, Init, RunTUI, project name detection.
// Sorting/hierarchy/search logic lives in tui_sorting.go.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/queue"
	"github.com/davidl71/exarp-go/internal/tools"
	"golang.org/x/term"
)

// availableSpinners contains terminal-compatible spinner styles users can cycle through.
// Only ASCII/Unicode box-drawing spinners are included for tmux compatibility.
var availableSpinners = []struct {
	name    string
	spinner spinner.Spinner
}{
	{"Dot", spinner.Dot},
	{"Line", spinner.Line},
	{"MiniDot", spinner.MiniDot},
	{"Jump", spinner.Jump},
	{"Pulse", spinner.Pulse},
	{"Points", spinner.Points},
	{"Meter", spinner.Meter},
	{"Ellipsis", spinner.Ellipsis},
}

type model struct {
	CoreState
	ConfigViewState
	ScorecardViewState
	HandoffViewState
	WavesViewState
	TaskAnalysisViewState
	JobsViewState
	TaskDetailViewState
	TaskListViewState
}

type CoreState struct {
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

	// Terminal dimensions (detected at startup and on SIGWINCH via tea.WindowSizeMsg)
	width  int
	height int

	// Current active view
	mode string
}

type ConfigViewState struct {
	// Config editing mode
	configSections    []configSection
	configCursor      int
	configData        *config.FullConfig
	configChanged     bool
	configSectionText string // overlay when Enter on a section
	configSaveMessage string // after save: success or error line
	configSaveSuccess bool
}

type ScorecardViewState struct {
	// Scorecard view (project health)
	scorecardText      string
	scorecardRecs      []string // recommendations from scorecard
	scorecardRecCursor int      // selected recommendation index
	scorecardLoading   bool
	scorecardErr       error
	scorecardRunOutput string // last "run recommendation" output or error
}

type HandoffViewState struct {
	// Handoffs view (session handoff notes)
	handoffText        string
	handoffLoading     bool
	handoffErr         error
	handoffEntries     []map[string]interface{} // parsed list for list/detail view
	handoffCursor      int
	handoffSelected    map[int]struct{}
	handoffDetailIndex int    // -1 = list, >= 0 = showing detail for that index
	handoffActionMsg   string // result of close/approve action
}

type WavesViewState struct {
	// Waves view (dependency-order waves from BacklogExecutionOrder)
	waves           map[int][]string // level -> task IDs (computed when entering waves view)
	waveDetailLevel int              // -1 = wave list (collapsed), >= 0 = viewing tasks for that wave level
	waveCursor      int              // cursor in wave list (when collapsed)
	waveTaskCursor  int              // cursor within expanded wave's task list (when waveDetailLevel >= 0)
	waveMoveTaskID  string           // when set, showing "move task to wave" prompt; Esc clears
	waveMoveMsg     string           // result message after move (success or error)
	waveUpdateMsg   string           // result message after update waves from plan (success or error)
	queueEnabled    bool             // true when REDIS_ADDR is set (queue available)
	queueEnqueueMsg string           // result message after enqueue wave (success or error)
}

type TaskAnalysisViewState struct {
	// Task analysis view (run task_analysis tool and show result)
	taskAnalysisText           string // result text
	taskAnalysisLoading        bool
	taskAnalysisErr            error
	taskAnalysisAction         string // e.g. "parallelization", "dependencies"
	taskAnalysisReturnMode     string // "tasks" or "waves" when leaving view
	taskAnalysisApproveLoading bool
	taskAnalysisApproveMsg     string // "Saved to ..." or error after y=write waves
}

type JobsViewState struct {
	// Background jobs view (child agent launches)
	jobs            []BackgroundJob
	jobsCursor      int
	jobsDetailIndex int // -1 = list, >= 0 = showing detail for that job
}

type TaskDetailViewState struct {
	// Task detail overlay (pressing 's' on a task)
	taskDetailTask      *database.Todo2Task
	taskDetailScrollTop int // scroll offset for long detail content
}

type TaskListViewState struct {
	// Viewport: scroll offset for task list so only visible rows are rendered
	viewportOffset int

	// Sort: order (id|status|priority|updated|hierarchy), ascending
	sortOrder          string
	sortAsc            bool
	hierarchyOrder     []int               // display order when sortOrder==hierarchy
	hierarchyDepth     map[int]int         // task index -> indent level (0=root), used for hierarchy order
	hierarchyDepthByID map[string]int      // task ID -> indent level, for indent regardless of sort
	collapsedTaskIDs   map[string]struct{} // task IDs that are collapsed (their descendants hidden in tree)

	// Search/filter: / to open, n/N next/prev match
	searchQuery     string
	searchMode      bool
	filteredIndices []int

	// Help overlay
	showHelp bool

	// Compact vs spacious list density (toggle with 'v', inspired by omm)
	spaciousMode bool

	// Inline task creation (inspired by omm): + to start, type name, Enter to create, Esc cancel
	createMode  bool
	createInput string

	// Child agent: last result message for status display (cleared on next key or refresh)
	childAgentMsg string

	// Bulk status update: when bulkStatusPrompt is true, showing status selection menu
	bulkStatusPrompt bool
	bulkStatusMsg    string // result message after bulk update

	// Bubble list for task rendering (replaces custom narrow/medium/wide rendering)
	taskList      list.Model
	useBubbleList bool // toggle for bubble/list vs custom rendering

	// Bubble table for sprintboard-style view
	taskTable      table.Model
	useBubbleTable bool // toggle for bubble/table vs custom rendering

	// Bubble spinner for async operations
	taskSpinner       spinner.Model
	spinnerMessage    string // message to display with spinner
	spinnerStyleIndex int    // index into available spinner styles

	// Command palette
	commandPalette     textinput.Model
	showCommandPalette bool

	// Contextual help bubble: shows brief hints based on current context
	contextualHelp     string // Current contextual help message
	contextualHelpTime int64  // Timestamp when help was shown (for auto-hide)
}

// configuredSpinner returns a spinner.Model initialized with the style from cfg.
// Falls back to spinner.Line when cfg is nil or the style is unrecognized.
func configuredSpinner(cfg *config.FullConfig) spinner.Model {
	style := "line"
	if cfg != nil && cfg.TUI.SpinnerStyle != "" {
		style = cfg.TUI.SpinnerStyle
	}

	return spinner.New(spinner.WithSpinner(SpinnerStyleFromString(style)))
}

// initialModel creates the TUI model. initialWidth and initialHeight are optional;
// when > 0 they set the terminal size at startup (e.g. from term.GetSize).
// Resize (SIGWINCH on Unix) is handled by Bubble Tea and delivered as tea.WindowSizeMsg.
func initialModel(server framework.MCPServer, status string, projectRoot, projectName string, initialWidth, initialHeight int) model {
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

	width, height := minTermWidth, minTermHeight
	if initialWidth > 0 && initialHeight > 0 {
		width, height = initialWidth, initialHeight
	}

	return model{
		CoreState: CoreState{
			tasks:       []*database.Todo2Task{},
			cursor:      0,
			selected:    make(map[int]struct{}),
			status:      status,
			server:      server,
			loading:     true,
			width:       width,
			height:      height,
			autoRefresh: true,
			lastUpdate:  time.Now(),
			projectRoot: projectRoot,
			projectName: projectName,
			mode:        ModeTasks,
		},
		ConfigViewState: ConfigViewState{
			configSections:    sections,
			configCursor:      0,
			configData:        cfg,
			configChanged:     false,
			configSectionText: "",
		},
		ScorecardViewState: ScorecardViewState{
			scorecardLoading: false,
		},
		HandoffViewState: HandoffViewState{
			handoffEntries:     nil,
			handoffCursor:      0,
			handoffSelected:    make(map[int]struct{}),
			handoffDetailIndex: -1,
		},
		WavesViewState: WavesViewState{
			waveDetailLevel: -1,
			waveCursor:      0,
			queueEnabled:    queue.ConfigFromEnv().Enabled(),
		},
		TaskAnalysisViewState: TaskAnalysisViewState{},
		JobsViewState: JobsViewState{
			jobsCursor:      0,
			jobsDetailIndex: -1,
		},
		TaskDetailViewState: TaskDetailViewState{},
		TaskListViewState: TaskListViewState{
			sortOrder:         SortByHierarchy,
			sortAsc:           true,
			collapsedTaskIDs:  make(map[string]struct{}),
			taskList:          list.New(nil, list.NewDefaultDelegate(), 0, 0),
			useBubbleList:     false, // set to true to enable bubble/list rendering
			taskTable:         table.New(table.WithColumns(defaultTableColumns), table.WithRows(nil)),
			useBubbleTable:    false,                                                          // set to true to enable bubble/table rendering
			taskSpinner:       configuredSpinner(cfg),
			spinnerMessage:    "",
			spinnerStyleIndex: 1, // Default to Line spinner (may be overridden by config)
			commandPalette:    newCommandPalette(),
		},
	}
}

// handoffSelectedIDs returns handoff id strings for currently selected handoff indices.
func (m model) handoffSelectedIDs() []string {
	var ids []string

	for i := range m.handoffSelected {
		if i < 0 || i >= len(m.handoffEntries) {
			continue
		}

		if id, ok := m.handoffEntries[i]["id"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}

	return ids
}

func (m model) Init() tea.Cmd {
	// Load tasks, start auto-refresh ticker, get initial window size, and start spinner
	m.taskSpinner, _ = m.taskSpinner.Update(spinner.TickMsg{})
	return tea.Batch(loadTasks(m.server, m.status), tick(), func() tea.Msg { return tea.RequestWindowSize() }, func() tea.Msg { return m.taskSpinner.Tick() })
}

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

// getModuleName reads the module name from go.mod.
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

// RunTUI starts the TUI interface. Terminal size is detected at startup when
// stdout is a TTY; SIGWINCH (window resize) is handled by Bubble Tea and
// updates the layout via tea.WindowSizeMsg.
func RunTUI(server framework.MCPServer, status string) error {
	// Suppress debug logs when running TUI (interactive UI shouldn't show logs)
	CLIOutputOpts.Quiet = true

	// Initialize database if needed (without closing it immediately)
	projectRoot, err := tools.FindProjectRoot()
	projectName := ""

	if err != nil {
		logWarn(context.TODO(), "Could not find project root", "error", err, "operation", "RunTUI", "fallback", "JSON")
	} else {
		projectName = getProjectName(projectRoot)
		EnsureConfigAndDatabase(projectRoot)

		if database.DB != nil {
			defer func() {
				if err := database.Close(); err != nil {
					logWarn(context.TODO(), "Error closing database", "error", err, "operation", "closeDatabase")
				}
			}()
		}
	}

	// Detect terminal size at startup so the first frame uses correct dimensions.
	// Resize (SIGWINCH) is handled by Bubble Tea and sends WindowSizeMsg on change.
	initialWidth, initialHeight := 0, 0

	if term.IsTerminal(int(os.Stdout.Fd())) {
		if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
			initialWidth, initialHeight = w, h
		}
	}

	p := tea.NewProgram(initialModel(server, status, projectRoot, projectName, initialWidth, initialHeight))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
