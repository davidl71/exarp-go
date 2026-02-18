package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/queue"
	"github.com/davidl71/exarp-go/internal/tools"
	humanize "github.com/dustin/go-humanize"
	"golang.org/x/term"
)

var (
	// Header/Title styles (top/htop-like).
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

		// Status bar (bottom).
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#000000")).
			Padding(0, 1)

		// Task/item styles.
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

		// Priority colors (htop-like).
	highPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Bold(true)

	mediumPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFF00"))

	lowPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00"))

		// Border/separator.
	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#808080"))
)

// Column widths for task list tabulation (header and rows must match).
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

	// Minimum terminal dimensions for layout. Some terminals (e.g. iTerm2) may
	// report 0 or stale size; clamping avoids broken task detail ("s") and layout.
	minTermWidth  = 80
	minTermHeight = 24
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

// indentForTask returns the indent string for hierarchical display (by parent/dependency depth).
// Uses hierarchyDepthByID so indent stays correct after any sort order.
func (m model) indentForTask(realIdx int) string {
	if m.hierarchyDepthByID == nil || realIdx < 0 || realIdx >= len(m.tasks) {
		return ""
	}

	task := m.tasks[realIdx]
	if task == nil {
		return ""
	}

	d := m.hierarchyDepthByID[task.ID]
	if d <= 0 {
		return ""
	}

	return strings.Repeat("  ", d)
}

// treeMarkerForTask returns a prefix for tasks that have children: "▶ " when collapsed, "▼ " when expanded, "" otherwise.
func (m model) treeMarkerForTask(realIdx int) string {
	if realIdx < 0 || realIdx >= len(m.tasks) {
		return ""
	}

	task := m.tasks[realIdx]
	if task == nil || !m.taskHasChildren(task.ID) {
		return ""
	}

	if _, ok := m.collapsedTaskIDs[task.ID]; ok {
		return "▶ "
	}

	return "▼ "
}

// wordWrap wraps s to at most width runes per line, breaking at spaces when possible.
func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}

	var out strings.Builder

	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			out.WriteString("\n")
			continue
		}

		for len(line) > width {
			cut := width
			if idx := strings.LastIndex(line[:min(len(line), width)], " "); idx > 0 {
				cut = idx
			}

			out.WriteString(line[:cut])
			out.WriteString("\n")

			line = strings.TrimLeft(line[cut:], " ")
		}

		out.WriteString(line)
		out.WriteString("\n")
	}

	return strings.TrimSuffix(out.String(), "\n")
}

// sortedWaveLevels returns sorted wave level keys from waves map.
func sortedWaveLevels(waves map[int][]string) []int {
	levels := make([]int, 0, len(waves))
	for k := range waves {
		levels = append(levels, k)
	}

	sort.Ints(levels)

	return levels
}

// effectiveWidth returns width for layout, never below minTermWidth (avoids iTerm2/terminal size quirks).
func (m model) effectiveWidth() int {
	if m.width >= minTermWidth {
		return m.width
	}

	return minTermWidth
}

// effectiveHeight returns height for layout, never below minTermHeight (avoids iTerm2/terminal size quirks).
func (m model) effectiveHeight() int {
	if m.height >= minTermHeight {
		return m.height
	}

	return minTermHeight
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

	// Terminal dimensions (detected at startup and on SIGWINCH via tea.WindowSizeMsg)
	width  int
	height int

	// Config editing mode
	mode              string // "tasks", "config", "scorecard", "handoffs", "waves", "jobs", or "configSection"
	configSections    []configSection
	configCursor      int
	configData        *config.FullConfig
	configChanged     bool
	configSectionText string // overlay when Enter on a section
	configSaveMessage string // after save: success or error line
	configSaveSuccess bool

	// Scorecard view (project health)
	scorecardText      string
	scorecardRecs      []string // recommendations from scorecard
	scorecardRecCursor int      // selected recommendation index
	scorecardLoading   bool
	scorecardErr       error
	scorecardRunOutput string // last "run recommendation" output or error

	// Handoffs view (session handoff notes)
	handoffText        string
	handoffLoading     bool
	handoffErr         error
	handoffEntries     []map[string]interface{} // parsed list for list/detail view
	handoffCursor      int
	handoffSelected    map[int]struct{}
	handoffDetailIndex int    // -1 = list, >= 0 = showing detail for that index
	handoffActionMsg   string // result of close/approve action

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

	// Task analysis view (run task_analysis tool and show result)
	taskAnalysisText           string // result text
	taskAnalysisLoading        bool
	taskAnalysisErr            error
	taskAnalysisAction         string // e.g. "parallelization", "dependencies"
	taskAnalysisReturnMode     string // "tasks" or "waves" when leaving view
	taskAnalysisApproveLoading bool
	taskAnalysisApproveMsg     string // "Saved to ..." or error after y=write waves

	// Background jobs view (child agent launches)
	jobs            []BackgroundJob
	jobsCursor      int
	jobsDetailIndex int // -1 = list, >= 0 = showing detail for that job

	// Task detail overlay (pressing 's' on a task)
	taskDetailTask *database.Todo2Task

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

	// Child agent: last result message for status display (cleared on next key or refresh)
	childAgentMsg string
}

// BackgroundJob represents a launched child agent (or other background job).
type BackgroundJob struct {
	Kind      ChildAgentKind
	Prompt    string
	StartedAt time.Time
	Pid       int
	Output    string // captured stdout+stderr (non-interactive only)
}

// jobCompletedMsg is sent when a background agent process exits (with captured output).
type jobCompletedMsg struct {
	Pid      int
	Output   string
	ExitCode int
	Err      error
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
	text    string
	entries []map[string]interface{}
	err     error
}

type handoffActionDoneMsg struct {
	action  string // "close" or "approve"
	updated int
	err     error
}

type runRecommendationResultMsg struct {
	output string
	err    error
}

type configSectionDetailMsg struct {
	text string
}

type configSaveResultMsg struct {
	message string
	success bool
}

// wavesRefreshDoneMsg is sent after running exarp tools to refresh waves (task_workflow sync, task_analysis, etc.).
type wavesRefreshDoneMsg struct {
	err error
}

// taskAnalysisLoadedMsg is sent after running the task_analysis tool.
type taskAnalysisLoadedMsg struct {
	text   string
	action string
	err    error
}

// taskAnalysisApproveDoneMsg is sent after writing the waves plan (report parallel_execution_plan).
type taskAnalysisApproveDoneMsg struct {
	message string
	err     error
}

// moveTaskToWaveDoneMsg is sent after updating a task's dependencies to move it to another wave.
type moveTaskToWaveDoneMsg struct {
	taskID string
	err    error
}

// updateWavesFromPlanDoneMsg is sent after report(update_waves_from_plan) completes.
type updateWavesFromPlanDoneMsg struct {
	message string
	err     error
}

// enqueueWaveDoneMsg is sent after enqueueing a wave to Redis+Asynq.
type enqueueWaveDoneMsg struct {
	waveLevel int
	enqueued  int
	err       error
}

// childAgentResultMsg is sent after starting a child agent (for status display).
type childAgentResultMsg struct {
	Result ChildAgentRunResult
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
		tasks:              []*database.Todo2Task{},
		cursor:             0,
		selected:           make(map[int]struct{}),
		status:             status,
		server:             server,
		loading:            true,
		width:              width,
		height:             height,
		autoRefresh:        true, // Enable auto-refresh by default
		lastUpdate:         time.Now(),
		projectRoot:        projectRoot,
		projectName:        projectName,
		mode:               "tasks",
		configSections:     sections,
		configCursor:       0,
		configData:         cfg,
		configChanged:      false,
		configSectionText:  "",
		configSaveMessage:  "",
		configSaveSuccess:  false,
		scorecardLoading:   false,
		taskDetailTask:     nil,
		sortOrder:          "hierarchy",
		sortAsc:            true,
		searchQuery:        "",
		searchMode:         false,
		filteredIndices:    nil,
		showHelp:           false,
		childAgentMsg:      "",
		handoffEntries:     nil,
		handoffCursor:      0,
		handoffSelected:    make(map[int]struct{}),
		handoffDetailIndex: -1,
		handoffActionMsg:   "",
		waveDetailLevel:    -1,
		waveCursor:         0,
		queueEnabled:       queue.ConfigFromEnv().Enabled(),
		jobs:               nil,
		jobsCursor:         0,
		jobsDetailIndex:    -1,
		collapsedTaskIDs:   make(map[string]struct{}),
	}
}

// sortTasksBy sorts tasks in place by order (id|status|priority|updated) and direction.
func sortTasksBy(tasks []*database.Todo2Task, order string, asc bool) {
	if len(tasks) == 0 {
		return
	}

	less := func(i, j int) bool {
		a, b := tasks[i], tasks[j]

		var cmp int

		switch order {
		case "status":
			cmp = strings.Compare(strings.ToLower(a.Status), strings.ToLower(b.Status))
		case "priority":
			cmp = strings.Compare(strings.ToLower(a.Priority), strings.ToLower(b.Priority))
		case "updated":
			cmp = strings.Compare(a.LastModified, b.LastModified)
		default:
			cmp = strings.Compare(a.ID, b.ID)
		}

		if cmp == 0 {
			cmp = strings.Compare(a.ID, b.ID)
		}

		if asc {
			return cmp < 0
		}

		return cmp > 0
	}
	sort.Slice(tasks, less)
}

// priorityOrderForSort returns sort key for priority (lower = higher priority).
func priorityOrderForSort(p string) int {
	switch strings.ToLower(p) {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

// computeHierarchyOrder builds hierarchyOrder and hierarchyDepth from task graph (dependencies + parent_id).
func (m *model) computeHierarchyOrder() {
	if len(m.tasks) == 0 {
		m.hierarchyOrder = nil
		m.hierarchyDepth = nil
		m.hierarchyDepthByID = nil

		return
	}

	taskList := make([]tools.Todo2Task, 0, len(m.tasks))

	for _, p := range m.tasks {
		if p != nil {
			taskList = append(taskList, *p)
		}
	}

	tg, err := tools.BuildTaskGraph(taskList)
	if err != nil {
		m.hierarchyOrder = nil
		m.hierarchyDepth = nil
		m.hierarchyDepthByID = nil

		return
	}

	levels := tools.GetTaskLevels(tg)
	m.hierarchyDepthByID = make(map[string]int)

	for id, level := range levels {
		m.hierarchyDepthByID[id] = level
	}

	type item struct {
		idx      int
		level    int
		priority string
		id       string
	}

	var items []item

	for i, p := range m.tasks {
		if p == nil {
			continue
		}

		level := levels[p.ID]
		items = append(items, item{i, level, p.Priority, p.ID})
	}

	asc := m.sortAsc

	sort.Slice(items, func(i, j int) bool {
		if items[i].level != items[j].level {
			if asc {
				return items[i].level < items[j].level
			}

			return items[i].level > items[j].level
		}

		pi := priorityOrderForSort(items[i].priority)
		pj := priorityOrderForSort(items[j].priority)

		if pi != pj {
			if asc {
				return pi < pj
			}

			return pi > pj
		}

		if asc {
			return items[i].id < items[j].id
		}

		return items[i].id > items[j].id
	})

	m.hierarchyOrder = make([]int, len(items))
	m.hierarchyDepth = make(map[int]int)

	for i, it := range items {
		m.hierarchyOrder[i] = it.idx
		m.hierarchyDepth[it.idx] = it.level
	}
}

// taskParentMap returns task ID -> parent task ID for all tasks that have a parent.
func (m model) taskParentMap() map[string]string {
	out := make(map[string]string, len(m.tasks))

	for _, t := range m.tasks {
		if t.ParentID != "" {
			out[t.ID] = t.ParentID
		}
	}

	return out
}

// taskAncestorIDs returns for each task ID the set of ancestor IDs (parent + dependencies).
// Used to hide both parent_id descendants and dependency dependents when a node is collapsed.
func (m model) taskAncestorIDs() map[string][]string {
	out := make(map[string][]string, len(m.tasks))
	taskIDs := make(map[string]struct{}, len(m.tasks))

	for _, t := range m.tasks {
		if t == nil {
			continue
		}

		taskIDs[t.ID] = struct{}{}
	}

	for _, t := range m.tasks {
		if t == nil {
			continue
		}

		var ancestors []string

		if t.ParentID != "" {
			if _, ok := taskIDs[t.ParentID]; ok {
				ancestors = append(ancestors, t.ParentID)
			}
		}

		for _, depID := range t.Dependencies {
			if _, ok := taskIDs[depID]; ok {
				ancestors = append(ancestors, depID)
			}
		}

		if len(ancestors) > 0 {
			out[t.ID] = ancestors
		}
	}

	return out
}

// isDescendantOfCollapsed returns true if taskID has an ancestor in m.collapsedTaskIDs,
// following both parent_id and dependency links (so collapsing hides parent children and dependents).
func (m model) isDescendantOfCollapsed(taskID string, ancestorIDs map[string][]string) bool {
	seen := make(map[string]struct{})

	var queue []string

	queue = append(queue, taskID)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}

		for _, anc := range ancestorIDs[id] {
			if _, collapsed := m.collapsedTaskIDs[anc]; collapsed {
				return true
			}

			queue = append(queue, anc)
		}
	}

	return false
}

// taskHasChildren returns true if at least one task has ParentID == taskID or taskID in Dependencies.
func (m model) taskHasChildren(taskID string) bool {
	for _, t := range m.tasks {
		if t == nil {
			continue
		}

		if t.ParentID == taskID {
			return true
		}

		for _, depID := range t.Dependencies {
			if depID == taskID {
				return true
			}
		}
	}

	return false
}

// visibleIndices returns the indices of tasks to display (filtered by search, collapse, or all).
func (m model) visibleIndices() []int {
	if len(m.tasks) == 0 {
		return nil
	}

	var base []int
	if m.sortOrder == "hierarchy" && len(m.hierarchyOrder) > 0 {
		base = m.hierarchyOrder
	} else {
		base = make([]int, len(m.tasks))
		for i := range m.tasks {
			base[i] = i
		}
	}

	if m.filteredIndices != nil {
		filterSet := make(map[int]struct{})
		for _, i := range m.filteredIndices {
			filterSet[i] = struct{}{}
		}

		var out []int

		for _, i := range base {
			if _, ok := filterSet[i]; ok {
				out = append(out, i)
			}
		}

		base = out
	}
	// Hide descendants of collapsed tasks (parent_id children and dependency dependents)
	if len(m.collapsedTaskIDs) == 0 {
		return base
	}

	ancestorIDs := m.taskAncestorIDs()

	var final []int

	for _, i := range base {
		if m.tasks[i] == nil {
			continue
		}

		if !m.isDescendantOfCollapsed(m.tasks[i].ID, ancestorIDs) {
			final = append(final, i)
		}
	}

	return final
}

// realIndexAt returns the index into m.tasks for the current cursor position.
func (m model) realIndexAt(cursorPos int) int {
	vis := m.visibleIndices()
	if vis == nil || cursorPos < 0 || cursorPos >= len(vis) {
		return 0
	}

	return vis[cursorPos]
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

// computeFilteredIndices returns indices of m.tasks that match m.searchQuery (case-insensitive).
func (m model) computeFilteredIndices() []int {
	q := strings.TrimSpace(strings.ToLower(m.searchQuery))
	if q == "" {
		return nil
	}

	var out []int

	for i, t := range m.tasks {
		if taskMatchesSearch(t, q) {
			out = append(out, i)
		}
	}

	return out
}

func taskMatchesSearch(t *database.Todo2Task, q string) bool {
	if strings.Contains(strings.ToLower(t.ID), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.Content), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.LongDescription), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.Status), q) {
		return true
	}

	if strings.Contains(strings.ToLower(t.Priority), q) {
		return true
	}

	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}

	return false
}

func (m model) Init() tea.Cmd {
	// Load tasks, start auto-refresh ticker, and get initial window size
	return tea.Batch(loadTasks(m.status), tick(), tea.WindowSize())
}

// tick returns a command that sends a tick message at configured interval
// Uses config for refresh interval, defaults to 5 seconds if not configured.
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

// loadScorecard loads the project scorecard. When fullMode is true, runs full checks (including
// go test coverage) so the scorecard reflects state after implementing fixes.
func loadScorecard(projectRoot string, fullMode bool) tea.Cmd {
	return func() tea.Msg {
		if projectRoot == "" {
			return scorecardLoadedMsg{err: fmt.Errorf("no project root")}
		}

		ctx := context.Background()

		var combined strings.Builder

		var recommendations []string

		// Go scorecard (when Go project)
		if tools.IsGoProject() {
			opts := &tools.ScorecardOptions{FastMode: !fullMode}

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
		args := map[string]interface{}{
			"action":         "handoff",
			"sub_action":     "list",
			"limit":          float64(20),
			"include_closed": false,
		}

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

		text := strings.TrimSpace(b.String())

		var entries []map[string]interface{}

		var payload struct {
			Handoffs []map[string]interface{} `json:"handoffs"`
		}

		if err := json.Unmarshal([]byte(text), &payload); err == nil && len(payload.Handoffs) > 0 {
			entries = payload.Handoffs
		}

		return handoffLoadedMsg{text: text, entries: entries, err: nil}
	}
}

// runWavesRefreshTools runs exarp tools that refresh wave-related state (task_workflow sync, task_analysis parallelization),
// then returns wavesRefreshDoneMsg. After handling that message, the caller should run loadTasks to recompute waves.
func runWavesRefreshTools(server framework.MCPServer) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return wavesRefreshDoneMsg{err: fmt.Errorf("no server")}
		}

		ctx := context.Background()

		// 1. Sync task store (SQLite ↔ JSON if applicable)
		syncArgs, _ := json.Marshal(map[string]interface{}{"action": "sync", "sub_action": "list"})
		if _, err := server.CallTool(ctx, "task_workflow", syncArgs); err != nil {
			return wavesRefreshDoneMsg{err: fmt.Errorf("task_workflow sync: %w", err)}
		}

		// 2. Task analysis parallelization (refreshes dependency/wave view data)
		taArgs, _ := json.Marshal(map[string]interface{}{"action": "parallelization", "output_format": "text"})
		if _, err := server.CallTool(ctx, "task_analysis", taArgs); err != nil {
			return wavesRefreshDoneMsg{err: fmt.Errorf("task_analysis parallelization: %w", err)}
		}

		return wavesRefreshDoneMsg{err: nil}
	}
}

// runTaskAnalysis runs the task_analysis tool with the given action (e.g. "parallelization", "dependencies", "execution_plan")
// and returns taskAnalysisLoadedMsg. Used from tasks and waves views.
func runTaskAnalysis(server framework.MCPServer, action string) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return taskAnalysisLoadedMsg{action: action, err: fmt.Errorf("no server")}
		}

		ctx := context.Background()
		args, _ := json.Marshal(map[string]interface{}{"action": action, "output_format": "text"})

		result, err := server.CallTool(ctx, "task_analysis", args)
		if err != nil {
			return taskAnalysisLoadedMsg{action: action, err: err}
		}

		var b strings.Builder
		for _, c := range result {
			b.WriteString(c.Text)
			b.WriteString("\n")
		}

		return taskAnalysisLoadedMsg{text: strings.TrimSpace(b.String()), action: action, err: nil}
	}
}

// runReportUpdateWavesFromPlan runs report(action=update_waves_from_plan) to sync Todo2 task dependencies
// from docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md. Returns updateWavesFromPlanDoneMsg.
func runReportUpdateWavesFromPlan(server framework.MCPServer, projectRoot string) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return updateWavesFromPlanDoneMsg{message: "", err: fmt.Errorf("no server")}
		}

		ctx := context.Background()
		args, _ := json.Marshal(map[string]interface{}{
			"action": "update_waves_from_plan",
		})

		result, err := server.CallTool(ctx, "report", args)
		if err != nil {
			return updateWavesFromPlanDoneMsg{message: "", err: err}
		}

		var msg string
		for _, c := range result {
			msg += c.Text
		}

		return updateWavesFromPlanDoneMsg{message: strings.TrimSpace(msg), err: nil}
	}
}

// runEnqueueWave enqueues all tasks from the given wave level to Redis+Asynq.
func runEnqueueWave(projectRoot string, waveLevel int) tea.Cmd {
	return func() tea.Msg {
		cfg := queue.ConfigFromEnv()
		if !cfg.Enabled() {
			return enqueueWaveDoneMsg{waveLevel: waveLevel, err: fmt.Errorf("REDIS_ADDR not set")}
		}
		producer, err := queue.NewProducer(cfg)
		if err != nil {
			return enqueueWaveDoneMsg{waveLevel: waveLevel, err: err}
		}
		defer producer.Close()

		enqueued, err := producer.EnqueueWave(context.Background(), projectRoot, waveLevel)
		return enqueueWaveDoneMsg{waveLevel: waveLevel, enqueued: enqueued, err: err}
	}
}

// runReportParallelExecutionPlan runs task_analysis(action=execution_plan, output_format=subagents_plan)
// to write waves to .cursor/plans/parallel-execution-subagents.plan.md using exarp's wave detection.
// Returns taskAnalysisApproveDoneMsg.
func runReportParallelExecutionPlan(server framework.MCPServer, projectRoot string) tea.Cmd {
	return func() tea.Msg {
		if server == nil {
			return taskAnalysisApproveDoneMsg{err: fmt.Errorf("no server")}
		}

		ctx := context.Background()
		outputPath := filepath.Join(projectRoot, ".cursor", "plans", "parallel-execution-subagents.plan.md")
		args, _ := json.Marshal(map[string]interface{}{
			"action":        "execution_plan",
			"output_format": "subagents_plan",
			"output_path":   outputPath,
		})

		result, err := server.CallTool(ctx, "task_analysis", args)
		if err != nil {
			return taskAnalysisApproveDoneMsg{err: err}
		}

		var msg string
		for _, c := range result {
			msg += c.Text
		}

		msg = strings.TrimSpace(msg)
		if msg == "" {
			msg = "Waves plan written to .cursor/plans/parallel-execution-subagents.plan.md"
		}

		return taskAnalysisApproveDoneMsg{message: msg, err: nil}
	}
}

// moveTaskToWaveCmd updates a task's dependencies so it lands in the given wave (0 = no deps, K = depend on one task from wave K-1), then sends moveTaskToWaveDoneMsg.
func moveTaskToWaveCmd(task *database.Todo2Task, newDeps []string) tea.Cmd {
	return func() tea.Msg {
		if task == nil {
			return moveTaskToWaveDoneMsg{taskID: "", err: fmt.Errorf("no task")}
		}

		ctx := context.Background()
		// Copy so we don't mutate the shared task
		updated := *task
		updated.Dependencies = make([]string, len(newDeps))
		copy(updated.Dependencies, newDeps)

		if err := database.UpdateTask(ctx, &updated); err != nil {
			return moveTaskToWaveDoneMsg{taskID: task.ID, err: err}
		}

		return moveTaskToWaveDoneMsg{taskID: task.ID, err: nil}
	}
}

// runHandoffAction runs session tool handoff close or approve for the given handoff IDs, then sends handoffActionDoneMsg.
func runHandoffAction(server framework.MCPServer, projectRoot string, handoffIDs []string, action string) tea.Cmd {
	return func() tea.Msg {
		if server == nil || len(handoffIDs) == 0 {
			return handoffActionDoneMsg{action: action, err: fmt.Errorf("no server or no handoff IDs")}
		}

		ctx := context.Background()
		args := map[string]interface{}{
			"action":      "handoff",
			"sub_action":  action,
			"handoff_ids": handoffIDs,
		}

		argsBytes, err := json.Marshal(args)
		if err != nil {
			return handoffActionDoneMsg{action: action, err: err}
		}

		_, err = server.CallTool(ctx, "session", argsBytes)
		if err != nil {
			return handoffActionDoneMsg{action: action, err: err}
		}

		return handoffActionDoneMsg{action: action, updated: len(handoffIDs)}
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
		// Update terminal dimensions (SIGWINCH on Unix triggers this via Bubble Tea).
		// Clamp to minimums so iTerm2 (or other terminals that report 0/stale size)
		// don't break window size or task detail ("s") layout.
		m.width = msg.Width
		m.height = msg.Height

		if m.width < minTermWidth {
			m.width = minTermWidth
		}

		if m.height < minTermHeight {
			m.height = minTermHeight
		}

		return m, nil

	case taskLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.tasks = msg.tasks
		m.computeHierarchyOrder() // always compute so hierarchy depth/order available

		if m.sortOrder == "hierarchy" && len(m.hierarchyOrder) > 0 {
			// use hierarchy order (parent then children, with depth)
		} else {
			sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
		}

		if m.searchQuery != "" {
			m.filteredIndices = m.computeFilteredIndices()
		} else {
			m.filteredIndices = nil
		}

		vis := m.visibleIndices()
		if len(vis) > 0 && m.cursor >= len(vis) {
			m.cursor = len(vis) - 1
		}

		m.lastUpdate = time.Now()
		// If in waves view, recompute waves (prefer docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md)
		if m.mode == "waves" && len(m.tasks) > 0 {
			taskList := make([]tools.Todo2Task, 0, len(m.tasks))

			for _, t := range m.tasks {
				if t != nil {
					taskList = append(taskList, *t)
				}
			}

			waves, err := tools.ComputeWavesForTUI(m.projectRoot, taskList)
			if err == nil {
				m.waves = waves
				if m.waveDetailLevel >= 0 {
					if ids := m.waves[m.waveDetailLevel]; len(ids) > 0 {
						if m.waveTaskCursor >= len(ids) {
							m.waveTaskCursor = len(ids) - 1
						}
					} else {
						m.waveTaskCursor = 0
					}
				}
			} else {
				m.waves = nil
			}
		}
		// Continue auto-refresh if enabled
		if m.autoRefresh {
			return m, tick()
		}

		return m, nil

	case wavesRefreshDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		// Reload tasks so waves recompute from updated backlog
		m.loading = true

		return m, loadTasks(m.status)

	case taskAnalysisLoadedMsg:
		m.taskAnalysisLoading = false
		m.taskAnalysisErr = msg.err
		m.taskAnalysisAction = msg.action

		if msg.err == nil {
			m.taskAnalysisText = msg.text
		} else {
			m.taskAnalysisText = ""
		}

		return m, nil

	case taskAnalysisApproveDoneMsg:
		m.taskAnalysisApproveLoading = false
		if msg.err != nil {
			m.taskAnalysisApproveMsg = "Error: " + msg.err.Error()
		} else {
			m.taskAnalysisApproveMsg = msg.message
		}

		return m, nil

	case moveTaskToWaveDoneMsg:
		m.waveMoveTaskID = ""
		if msg.err != nil {
			m.waveMoveMsg = "Error: " + msg.err.Error()
		} else {
			m.waveMoveMsg = "Moved " + msg.taskID + " to wave"
		}

		return m, loadTasks(m.status)

	case updateWavesFromPlanDoneMsg:
		m.loading = false
		m.waveUpdateMsg = ""

		if msg.err != nil {
			m.waveUpdateMsg = "Error: " + msg.err.Error()
		} else if msg.message != "" {
			m.waveUpdateMsg = msg.message
		}

		return m, loadTasks(m.status)

	case enqueueWaveDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.queueEnqueueMsg = "Enqueue error: " + msg.err.Error()
		} else {
			m.queueEnqueueMsg = fmt.Sprintf("Enqueued %d tasks from wave %d to Redis", msg.enqueued, msg.waveLevel)
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

	case configSectionDetailMsg:
		m.mode = "configSection"
		m.configSectionText = msg.text

		return m, nil

	case configSaveResultMsg:
		m.configSaveMessage = msg.message
		m.configSaveSuccess = msg.success

		if msg.success {
			m.configChanged = false
		}

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
			m.handoffEntries = msg.entries

			if len(m.handoffEntries) > 0 {
				if m.handoffCursor >= len(m.handoffEntries) {
					m.handoffCursor = len(m.handoffEntries) - 1
				}
			} else {
				m.handoffCursor = 0
			}
			// Keep selection only for indices that still exist
			newSel := make(map[int]struct{})

			for i := range m.handoffSelected {
				if i >= 0 && i < len(m.handoffEntries) {
					newSel[i] = struct{}{}
				}
			}

			m.handoffSelected = newSel
		}

		return m, nil

	case handoffActionDoneMsg:
		if msg.err != nil {
			m.handoffActionMsg = fmt.Sprintf("%s failed: %v", msg.action, msg.err)
		} else {
			verb := msg.action + "d"
			switch msg.action {
			case "delete":
				verb = "deleted"
			case "close":
				verb = "closed"
			case "approve":
				verb = "approved"
			}

			m.handoffActionMsg = fmt.Sprintf("%d handoff(s) %s.", msg.updated, verb)
			m.handoffSelected = make(map[int]struct{})
			m.handoffDetailIndex = -1 // return to list after close/approve/delete
		}

		return m, loadHandoffs(m.server)

	case runRecommendationResultMsg:
		if msg.err != nil {
			m.scorecardRunOutput = "Error: " + msg.err.Error()
		} else {
			m.scorecardRunOutput = strings.TrimSpace(msg.output)
			if m.scorecardRunOutput == "" {
				m.scorecardRunOutput = "(command completed)"
			}
		}
		// Refresh scorecard with full checks so updated state (e.g. coverage after make test) is shown
		m.scorecardLoading = true

		return m, loadScorecard(m.projectRoot, true)

	case childAgentResultMsg:
		if msg.Result.Launched {
			m.childAgentMsg = msg.Result.Message
			m.jobs = append(m.jobs, BackgroundJob{
				Kind:      msg.Result.Kind,
				Prompt:    msg.Result.Prompt,
				StartedAt: time.Now(),
				Pid:       msg.Result.Pid,
			})
		} else {
			m.childAgentMsg = "Child agent: " + msg.Result.Message
		}

		return m, nil

	case jobCompletedMsg:
		for i := range m.jobs {
			if m.jobs[i].Pid == msg.Pid {
				m.jobs[i].Output = msg.Output
				if msg.Err != nil {
					m.jobs[i].Output += "\n(exit: " + msg.Err.Error() + ")"
				}

				break
			}
		}

		return m, nil

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

			if m.searchMode {
				m.searchMode = false
				m.searchQuery = ""
				m.filteredIndices = nil

				return m, nil
			}

			if m.mode == "taskAnalysis" {
				m.mode = m.taskAnalysisReturnMode
				if m.taskAnalysisReturnMode == "" {
					m.mode = "tasks"
				}

				return m, nil
			}

			if m.mode == "waves" {
				if m.waveDetailLevel >= 0 {
					m.waveDetailLevel = -1
				} else {
					m.mode = "tasks"
					m.cursor = 0
				}

				return m, nil
			}

			if m.mode == "jobs" {
				if m.jobsDetailIndex >= 0 {
					m.jobsDetailIndex = -1
				} else {
					m.mode = "tasks"
					m.cursor = 0
				}

				return m, nil
			}

			return m, nil
		}

		// When help is open, ignore all other keys (handled above)
		if m.showHelp {
			return m, nil
		}

		// Search mode: accept input, Enter to apply, Esc already handled above
		if m.searchMode && m.mode == "tasks" {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				m.filteredIndices = m.computeFilteredIndices()

				if len(m.visibleIndices()) > 0 && m.cursor >= len(m.visibleIndices()) {
					m.cursor = len(m.visibleIndices()) - 1
				}

				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}

				return m, nil
			default:
				if len(msg.String()) == 1 && msg.Type == tea.KeyRunes {
					m.searchQuery += msg.String()
				}

				return m, nil
			}
		}

		// When task detail is open, Esc / Enter / Space close it
		if m.mode == "taskDetail" {
			switch msg.String() {
			case "esc", "enter", " ":
				m.mode = "tasks"
				m.taskDetailTask = nil

				return m, nil
			}
		}

		// When config section detail is open, Esc / Enter / Space close it
		if m.mode == "configSection" {
			switch msg.String() {
			case "esc", "enter", " ":
				m.mode = "config"
				m.configSectionText = ""

				return m, nil
			}
		}

		// When handoff detail is open, Esc / Enter / Space close it
		if m.mode == "handoffs" && m.handoffDetailIndex >= 0 {
			switch msg.String() {
			case "esc", "enter", " ":
				m.handoffDetailIndex = -1
				return m, nil
			}
		}

		// When wave detail is open (viewing tasks for a wave), Esc / Enter / Space close it (or cancel move)
		if m.mode == "waves" && m.waveDetailLevel >= 0 {
			switch msg.String() {
			case "esc", "enter", " ":
				if m.waveMoveTaskID != "" {
					m.waveMoveTaskID = ""
					m.waveMoveMsg = ""

					return m, nil
				}

				m.waveUpdateMsg = ""
				m.waveDetailLevel = -1

				return m, nil
			}
		}

		// When job detail is open, Esc / Enter / Space close it
		if m.mode == "jobs" && m.jobsDetailIndex >= 0 {
			switch msg.String() {
			case "esc", "enter", " ":
				m.jobsDetailIndex = -1
				return m, nil
			}
		}

		switch msg.String() {
		case "p":
			// Back from scorecard or task analysis to previous view
			switch m.mode {
			case "scorecard":
				m.mode = "tasks"
				m.cursor = 0
			case "taskAnalysis":
				m.mode = m.taskAnalysisReturnMode
				if m.taskAnalysisReturnMode == "" {
					m.mode = "tasks"
				}
			case "tasks":
				m.mode = "scorecard"
				m.scorecardLoading = true
				m.scorecardErr = nil
				m.scorecardText = ""

				return m, loadScorecard(m.projectRoot, false)
			case "handoffs":
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
			m.handoffEntries = nil
			m.handoffCursor = 0
			m.handoffSelected = make(map[int]struct{})
			m.handoffDetailIndex = -1
			m.handoffActionMsg = ""

			return m, loadHandoffs(m.server)

		case "b":
			// Toggle background jobs view
			if m.mode == "jobs" {
				m.mode = "tasks"
				m.cursor = 0

				return m, nil
			}

			m.mode = "jobs"
			m.jobsCursor = 0
			m.jobsDetailIndex = -1

			return m, nil

		case "w":
			// Toggle waves view (dependency-order waves from backlog)
			if m.mode == "waves" {
				m.mode = "tasks"
				m.cursor = 0
				m.waveDetailLevel = -1

				return m, nil
			}

			if m.mode == "tasks" && len(m.tasks) > 0 {
				m.mode = "waves"
				m.waveDetailLevel = -1
				m.waveCursor = 0
				// Compute waves (prefer docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md)
				taskList := make([]tools.Todo2Task, 0, len(m.tasks))

				for _, t := range m.tasks {
					if t != nil {
						taskList = append(taskList, *t)
					}
				}

				waves, err := tools.ComputeWavesForTUI(m.projectRoot, taskList)
				if err != nil {
					m.waves = nil
				} else {
					m.waves = waves
				}
			}

			return m, nil

		case "c":
			// Toggle between tasks and config view
			switch m.mode {
			case "tasks":
				m.mode = "config"
				m.configCursor = 0
			case "config":
				m.mode = "tasks"
				m.cursor = 0
				m.configSaveMessage = ""
			case "handoffs":
				m.mode = "tasks"
				m.cursor = 0
			case "waves", "jobs":
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
			} else if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor > 0 {
					m.cursor--
				}
			} else if m.mode == "handoffs" && m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 && m.handoffCursor > 0 {
				m.handoffCursor--
			} else if m.mode == "waves" && m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
				ids := m.waves[m.waveDetailLevel]
				if len(ids) > 0 && m.waveTaskCursor > 0 {
					m.waveTaskCursor--
				}
			} else if m.mode == "waves" && m.waveDetailLevel < 0 && len(m.waves) > 0 && m.waveCursor > 0 {
				m.waveCursor--
			} else if m.mode == "jobs" && m.jobsDetailIndex < 0 && len(m.jobs) > 0 && m.jobsCursor > 0 {
				m.jobsCursor--
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
			} else if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis)-1 {
					m.cursor++
				}
			} else if m.mode == "handoffs" && m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 && m.handoffCursor < len(m.handoffEntries)-1 {
				m.handoffCursor++
			} else if m.mode == "waves" && m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
				ids := m.waves[m.waveDetailLevel]
				if len(ids) > 0 && m.waveTaskCursor < len(ids)-1 {
					m.waveTaskCursor++
				}
			} else if m.mode == "waves" && m.waveDetailLevel < 0 && len(m.waves) > 0 {
				levels := sortedWaveLevels(m.waves)
				if m.waveCursor < len(levels)-1 {
					m.waveCursor++
				}
			} else if m.mode == "jobs" && m.jobsDetailIndex < 0 && len(m.jobs) > 0 && m.jobsCursor < len(m.jobs)-1 {
				m.jobsCursor++
			}

			return m, nil

		case "enter", " ", "e", "i":
			// In handoffs: "i" = start interactive agent with handoff (do not close)
			if m.mode == "handoffs" && msg.String() == "i" && len(m.handoffEntries) > 0 {
				var h map[string]interface{}
				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					h = m.handoffEntries[m.handoffDetailIndex]
				} else if m.handoffCursor < len(m.handoffEntries) {
					h = m.handoffEntries[m.handoffCursor]
				}

				if h != nil {
					m.childAgentMsg = ""
					sum, _ := h["summary"].(string)

					var steps []interface{}
					if s, ok := h["next_steps"].([]interface{}); ok {
						steps = s
					}

					prompt := PromptForHandoff(sum, steps)

					return m, runChildAgentCmdInteractive(m.projectRoot, prompt, ChildAgentHandoff)
				}

				return m, nil
			}

			// In handoffs: "e" = execute current handoff in agent and close it
			if m.mode == "handoffs" && msg.String() == "e" && len(m.handoffEntries) > 0 {
				var h map[string]interface{}

				var id string

				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					h = m.handoffEntries[m.handoffDetailIndex]
					id, _ = h["id"].(string)
				} else if m.handoffCursor < len(m.handoffEntries) {
					h = m.handoffEntries[m.handoffCursor]
					id, _ = h["id"].(string)
				}

				if id != "" && h != nil {
					m.childAgentMsg = ""
					sum, _ := h["summary"].(string)

					var steps []interface{}
					if s, ok := h["next_steps"].([]interface{}); ok {
						steps = s
					}

					prompt := PromptForHandoff(sum, steps)

					return m, tea.Batch(
						runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff),
						runHandoffAction(m.server, m.projectRoot, []string{id}, "close"),
					)
				}

				return m, nil
			}

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
			} else if m.mode == "tasks" {
				// Toggle selection (cursor is index into visible list)
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					realIdx := m.realIndexAt(m.cursor)
					if _, ok := m.selected[realIdx]; ok {
						delete(m.selected, realIdx)
					} else {
						m.selected[realIdx] = struct{}{}
					}
				}
			} else if m.mode == "handoffs" && m.handoffDetailIndex < 0 && len(m.handoffEntries) > 0 {
				if msg.String() == " " {
					// Space: toggle selection
					if _, ok := m.handoffSelected[m.handoffCursor]; ok {
						delete(m.handoffSelected, m.handoffCursor)
					} else {
						m.handoffSelected[m.handoffCursor] = struct{}{}
					}
				} else {
					// Enter or e: open detail
					m.handoffDetailIndex = m.handoffCursor
				}
			} else if m.mode == "waves" && m.waveDetailLevel < 0 && len(m.waves) > 0 {
				levels := sortedWaveLevels(m.waves)
				if m.waveCursor < len(levels) {
					m.waveDetailLevel = levels[m.waveCursor]
					m.waveTaskCursor = 0
				}
			} else if m.mode == "jobs" && m.jobsDetailIndex < 0 && len(m.jobs) > 0 {
				m.jobsDetailIndex = m.jobsCursor
			}

			return m, nil

		case "r":
			switch m.mode {
			case "config":
				// Reload config
				m.configSaveMessage = ""

				cfg, err := config.LoadConfig(m.projectRoot)
				if err == nil {
					m.configData = cfg
					m.configChanged = false
				}

				return m, nil
			case "scorecard":
				// Refresh scorecard (fast mode for manual refresh; use Run then Enter for full refresh)
				m.scorecardLoading = true
				return m, loadScorecard(m.projectRoot, false)
			case "handoffs":
				// Refresh handoffs
				m.handoffLoading = true
				return m, loadHandoffs(m.server)
			case "waves":
				// Refresh tasks (waves recompute on taskLoadedMsg)
				m.loading = true
				return m, loadTasks(m.status)
			case "taskAnalysis":
				if !m.taskAnalysisLoading {
					m.taskAnalysisLoading = true

					action := m.taskAnalysisAction
					if action == "" {
						action = "parallelization"
					}

					return m, runTaskAnalysis(m.server, action)
				}

				return m, nil
			default:
				// Refresh tasks
				m.loading = true
				return m, loadTasks(m.status)
			}

		case "x":
			// Close/dismiss handoffs: from detail view (current item) or list view (selected or current)
			if m.mode == "handoffs" && len(m.handoffEntries) > 0 {
				var ids []string

				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					if id, _ := m.handoffEntries[m.handoffDetailIndex]["id"].(string); id != "" {
						ids = []string{id}
					}
				}

				if len(ids) == 0 {
					ids = m.handoffSelectedIDs()
					if len(ids) == 0 && m.handoffCursor < len(m.handoffEntries) {
						if id, _ := m.handoffEntries[m.handoffCursor]["id"].(string); id != "" {
							ids = []string{id}
						}
					}
				}

				if len(ids) > 0 {
					return m, runHandoffAction(m.server, m.projectRoot, ids, "close")
				}
			}

			return m, nil

		case "a":
			if m.mode == "scorecard" {
				return m, nil
			}

			if m.mode == "handoffs" && len(m.handoffEntries) > 0 {
				// Approve: from detail view (current item) or list view (selected or current)
				var ids []string

				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					if id, _ := m.handoffEntries[m.handoffDetailIndex]["id"].(string); id != "" {
						ids = []string{id}
					}
				}

				if len(ids) == 0 {
					ids = m.handoffSelectedIDs()
					if len(ids) == 0 && m.handoffCursor < len(m.handoffEntries) {
						if id, _ := m.handoffEntries[m.handoffCursor]["id"].(string); id != "" {
							ids = []string{id}
						}
					}
				}

				if len(ids) > 0 {
					return m, runHandoffAction(m.server, m.projectRoot, ids, "approve")
				}
			}

			if m.mode == "tasks" {
				// Toggle auto-refresh
				m.autoRefresh = !m.autoRefresh
				if m.autoRefresh {
					return m, tick()
				}
			}

			return m, nil

		case "d":
			// Delete handoffs: from detail view (current item) or list view (selected or current)
			if m.mode == "handoffs" && len(m.handoffEntries) > 0 {
				var ids []string

				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					if id, _ := m.handoffEntries[m.handoffDetailIndex]["id"].(string); id != "" {
						ids = []string{id}
					}
				}

				if len(ids) == 0 {
					ids = m.handoffSelectedIDs()
					if len(ids) == 0 && m.handoffCursor < len(m.handoffEntries) {
						if id, _ := m.handoffEntries[m.handoffCursor]["id"].(string); id != "" {
							ids = []string{id}
						}
					}
				}

				if len(ids) > 0 {
					return m, runHandoffAction(m.server, m.projectRoot, ids, "delete")
				}
			}

			return m, nil

		case "o":
			// Cycle sort order (id → status → priority → updated → hierarchy → id)
			if m.mode == "tasks" && len(m.tasks) > 0 {
				switch m.sortOrder {
				case "id":
					m.sortOrder = "status"
				case "status":
					m.sortOrder = "priority"
				case "priority":
					m.sortOrder = "updated"
				case "updated":
					m.sortOrder = "hierarchy"
				default:
					m.sortOrder = "id"
				}

				if m.sortOrder == "hierarchy" {
					m.computeHierarchyOrder()
				} else {
					sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
				}

				if m.cursor >= len(m.visibleIndices()) {
					m.cursor = len(m.visibleIndices()) - 1
				}
			}

			return m, nil

		case "O":
			// Toggle sort direction (asc ↔ desc)
			if m.mode == "tasks" && len(m.tasks) > 0 {
				m.sortAsc = !m.sortAsc
				if m.sortOrder == "hierarchy" {
					m.computeHierarchyOrder()
				} else {
					sortTasksBy(m.tasks, m.sortOrder, m.sortAsc)
				}
			}

			return m, nil

		case "/":
			// Start search/filter (vim-style)
			if m.mode == "tasks" {
				m.searchMode = true
				// Keep previous searchQuery so user can extend or backspace
			}

			return m, nil

		case "n":
			// Next search match (vim-style)
			if m.mode == "tasks" && m.searchQuery != "" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis)-1 {
					m.cursor++
				}
			}

			return m, nil

		case "N":
			// Previous search match (vim-style)
			if m.mode == "tasks" && m.searchQuery != "" {
				if m.cursor > 0 {
					m.cursor--
				}
			}

			return m, nil

		case "tab", "\t":
			// In tasks mode: collapse/expand tree node under cursor (if it has children)
			if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					realIdx := m.realIndexAt(m.cursor)

					task := m.tasks[realIdx]
					if task != nil && m.taskHasChildren(task.ID) {
						if _, ok := m.collapsedTaskIDs[task.ID]; ok {
							delete(m.collapsedTaskIDs, task.ID)
						} else {
							m.collapsedTaskIDs[task.ID] = struct{}{}
						}
					}
				}
			}

			return m, nil

		case "u":
			// In config view: update (save current config to .exarp/config.pb protobuf)
			if m.mode == "config" {
				return m, saveConfig(m.projectRoot, m.configData)
			}

			return m, nil

		case "s":
			if m.mode == "scorecard" {
				return m, nil
			}

			switch m.mode {
			case "tasks":
				// Show task details in-TUI (word-wrapped)
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					m.mode = "taskDetail"
					m.taskDetailTask = m.tasks[m.realIndexAt(m.cursor)]

					return m, nil
				}
			case "taskDetail":
				// Close task detail on 's' too (so same key can close)
				m.mode = "tasks"
				m.taskDetailTask = nil

				return m, nil
			default:
				// Save config (writes to .exarp/config.pb protobuf)
				return m, saveConfig(m.projectRoot, m.configData)
			}

		case "R":
			// In waves view: run exarp tools (task_workflow sync, task_analysis parallelization) then refresh waves
			if m.mode == "waves" {
				m.loading = true
				m.err = nil

				return m, runWavesRefreshTools(m.server)
			}

			return m, nil

		case "U":
			// In waves view: update Todo2 task dependencies from docs/PARALLEL_EXECUTION_PLAN_RESEARCH.md
			if m.mode == "waves" {
				m.loading = true
				m.waveUpdateMsg = ""

				return m, runReportUpdateWavesFromPlan(m.server, m.projectRoot)
			}

			return m, nil

		case "Q":
			if m.mode == "waves" && m.queueEnabled && len(m.waves) > 0 {
				levels := sortedWaveLevels(m.waves)
				waveIdx := 0
				if m.waveDetailLevel >= 0 {
					for i, l := range levels {
						if l == m.waveDetailLevel {
							waveIdx = i
							break
						}
					}
				} else if m.waveCursor < len(levels) {
					waveIdx = m.waveCursor
				}
				if waveIdx < len(levels) {
					level := levels[waveIdx]
					m.loading = true
					m.queueEnqueueMsg = ""
					return m, runEnqueueWave(m.projectRoot, level)
				}
			}

			return m, nil

		case "A":
			// In tasks or waves: run task_analysis and show result in dedicated view
			if m.mode == "tasks" || m.mode == "waves" {
				m.taskAnalysisReturnMode = m.mode
				m.mode = "taskAnalysis"
				m.taskAnalysisLoading = true
				m.taskAnalysisErr = nil
				m.taskAnalysisText = ""
				m.taskAnalysisAction = "parallelization"
				m.taskAnalysisApproveMsg = ""
				m.taskAnalysisApproveLoading = false

				return m, runTaskAnalysis(m.server, "parallelization")
			}

			if m.mode == "taskAnalysis" && !m.taskAnalysisLoading {
				// Rerun same action
				m.taskAnalysisLoading = true
				m.taskAnalysisApproveMsg = ""

				return m, runTaskAnalysis(m.server, m.taskAnalysisAction)
			}

			return m, nil

		case "y":
			// In task analysis view: approve = write waves plan to .cursor/plans/parallel-execution-subagents.plan.md
			if m.mode == "taskAnalysis" && !m.taskAnalysisLoading && !m.taskAnalysisApproveLoading {
				m.taskAnalysisApproveLoading = true
				m.taskAnalysisApproveMsg = ""

				return m, runReportParallelExecutionPlan(m.server, m.projectRoot)
			}

			return m, nil

		case "m":
			// In waves expanded view: start "move task to wave" (then press 0-9 to pick target wave)
			if m.mode == "waves" && m.waveDetailLevel >= 0 && m.waveMoveTaskID == "" {
				ids := m.waves[m.waveDetailLevel]
				if len(ids) > 0 && m.waveTaskCursor < len(ids) {
					m.waveMoveTaskID = ids[m.waveTaskCursor]
					m.waveMoveMsg = ""
				}
			}

			return m, nil

		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.mode == "waves" && m.waveMoveTaskID != "" {
				targetLevel := int(msg.String()[0] - '0')
				levels := sortedWaveLevels(m.waves)

				if targetLevel < 0 || targetLevel >= len(levels) {
					m.waveMoveMsg = "Invalid wave number"
					return m, nil
				}

				level := levels[targetLevel]

				var newDeps []string

				if level == 0 {
					newDeps = nil
				} else {
					prevLevel := levels[targetLevel-1]

					prevIDs := m.waves[prevLevel]
					if len(prevIDs) == 0 {
						m.waveMoveMsg = "No tasks in previous wave"
						return m, nil
					}

					newDeps = []string{prevIDs[0]}
				}

				taskByID := make(map[string]*database.Todo2Task)

				for _, t := range m.tasks {
					if t != nil {
						taskByID[t.ID] = t
					}
				}

				task := taskByID[m.waveMoveTaskID]
				if task == nil {
					m.waveMoveMsg = "Task not found"
					return m, nil
				}

				return m, moveTaskToWaveCmd(task, newDeps)
			}

			return m, nil

		case "E":
			// Execute current context (task, handoff, wave) in child agent
			m.childAgentMsg = ""
			if m.mode == "tasks" {
				vis := m.visibleIndices()
				if len(vis) > 0 && m.cursor < len(vis) {
					task := m.tasks[m.realIndexAt(m.cursor)]
					if task != nil {
						prompt := PromptForTask(task.ID, task.Content)
						return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentTask)
					}
				}
			} else if m.mode == "taskDetail" && m.taskDetailTask != nil {
				prompt := PromptForTask(m.taskDetailTask.ID, m.taskDetailTask.Content)
				return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentTask)
			} else if m.mode == "handoffs" {
				if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
					h := m.handoffEntries[m.handoffDetailIndex]
					sum, _ := h["summary"].(string)

					var steps []interface{}
					if s, ok := h["next_steps"].([]interface{}); ok {
						steps = s
					}

					prompt := PromptForHandoff(sum, steps)

					return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff)
				} else if len(m.handoffEntries) > 0 && m.handoffCursor < len(m.handoffEntries) {
					h := m.handoffEntries[m.handoffCursor]
					sum, _ := h["summary"].(string)

					var steps []interface{}
					if s, ok := h["next_steps"].([]interface{}); ok {
						steps = s
					}

					prompt := PromptForHandoff(sum, steps)

					return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentHandoff)
				}
			} else if m.mode == "waves" && len(m.waves) > 0 {
				levels := sortedWaveLevels(m.waves)
				waveIdx := 0

				if m.waveDetailLevel >= 0 {
					for i, l := range levels {
						if l == m.waveDetailLevel {
							waveIdx = i
							break
						}
					}
				} else if m.waveCursor < len(levels) {
					waveIdx = m.waveCursor
				}

				if waveIdx < len(levels) {
					level := levels[waveIdx]
					ids := m.waves[level]
					prompt := PromptForWave(level, ids)

					return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentWave)
				}
			}

			return m, nil

		case "L":
			// Launch plan in child agent (tasks or taskDetail)
			m.childAgentMsg = ""
			if m.mode == "tasks" || m.mode == "taskDetail" {
				prompt := PromptForPlan(m.projectRoot)
				return m, runChildAgentCmd(m.projectRoot, prompt, ChildAgentPlan)
			}

			return m, nil

		default:
			// Clear child-agent status on any other key
			if m.childAgentMsg != "" {
				m.childAgentMsg = ""
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.showHelp {
		return m.viewHelp()
	}

	if m.mode == "taskDetail" && m.taskDetailTask != nil {
		return m.viewTaskDetail()
	}

	if m.mode == "configSection" {
		return m.viewConfigSection()
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

	if m.mode == "waves" {
		return m.viewWaves()
	}

	if m.mode == "taskAnalysis" {
		return m.viewTaskAnalysis()
	}

	if m.mode == "jobs" {
		return m.viewJobs()
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

	// Calculate available width (account for padding). Use effective dimensions so
	// iTerm2 and other terminals that report 0 or stale size still get correct layout.
	availableWidth := m.effectiveWidth() - 2
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

	// Task count (visible = filtered or all)
	visCount := len(m.visibleIndices())
	totalCount := len(m.tasks)
	taskCountStr := " " + humanize.Comma(int64(visCount))

	if m.searchQuery != "" && totalCount != visCount {
		taskCountStr = fmt.Sprintf(" %s/%s", humanize.Comma(int64(visCount)), humanize.Comma(int64(totalCount)))
	}

	headerLine.WriteString(headerLabelStyle.Render("Tasks:"))
	headerLine.WriteString(headerValueStyle.Render(taskCountStr))
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
		headerLine.WriteString(headerLabelStyle.Render("Updated:"))
		headerLine.WriteString(headerValueStyle.Render(" " + humanize.Time(m.lastUpdate)))
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

	// Search mode prompt
	if m.searchMode {
		searchPrompt := "/" + m.searchQuery + "_"
		b.WriteString(helpStyle.Render("Search: " + searchPrompt + " (Enter=apply Esc=cancel)"))
		b.WriteString("\n")
	}

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

	// Child agent result (one-line feedback)
	if m.childAgentMsg != "" {
		msgLine := m.childAgentMsg
		if len(msgLine) > availableWidth-2 {
			msgLine = msgLine[:availableWidth-5] + "..."
		}

		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  " + msgLine))
		b.WriteString("\n")
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
	statusBar.WriteString(helpStyle.Render("/"))
	statusBar.WriteString(" search  ")
	statusBar.WriteString(helpStyle.Render("n/N"))
	statusBar.WriteString(" next/prev  ")
	statusBar.WriteString(helpStyle.Render("o/O"))
	statusBar.WriteString(" sort  ")
	statusBar.WriteString(helpStyle.Render("Space"))
	statusBar.WriteString(" select  ")
	statusBar.WriteString(helpStyle.Render("Tab"))
	statusBar.WriteString(" collapse  ")
	statusBar.WriteString(helpStyle.Render("s"))
	statusBar.WriteString(" details  ")
	statusBar.WriteString(helpStyle.Render("r"))
	statusBar.WriteString(" refresh  ")
	statusBar.WriteString(helpStyle.Render("a"))
	statusBar.WriteString(" auto  ")
	statusBar.WriteString(helpStyle.Render("c"))
	statusBar.WriteString(" config  ")
	statusBar.WriteString(helpStyle.Render("p"))
	statusBar.WriteString(" scorecard  ")
	statusBar.WriteString(helpStyle.Render("w"))
	statusBar.WriteString(" w waves  ")
	statusBar.WriteString(helpStyle.Render("A"))
	statusBar.WriteString(" analysis  ")
	statusBar.WriteString(helpStyle.Render("b"))
	statusBar.WriteString(" jobs  ")
	statusBar.WriteString(helpStyle.Render("E"))
	statusBar.WriteString(" child agent  ")
	statusBar.WriteString(helpStyle.Render("L"))
	statusBar.WriteString(" plan  ")
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

// renderNarrowTaskList renders tasks in a narrow terminal (< 80 chars).
func (m model) renderNarrowTaskList(b *strings.Builder, width int) {
	vis := m.visibleIndices()
	for idx, realIdx := range vis {
		task := m.tasks[realIdx]
		indent := m.indentForTask(realIdx)
		marker := m.treeMarkerForTask(realIdx)

		cursor := " "
		if m.cursor == idx {
			cursor = ">"
			if _, ok := m.selected[realIdx]; ok {
				cursor = "✓"
			}
		}

		// Minimal info: indent + tree marker + cursor + ID, status, truncated content
		line := fmt.Sprintf("%s%s%s %s", indent, marker, cursor, task.ID)

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

		if m.cursor == idx {
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// renderMediumTaskList renders tasks in a medium terminal (80-120 chars).
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

	vis := m.visibleIndices()
	for idx, realIdx := range vis {
		task := m.tasks[realIdx]
		indent := m.indentForTask(realIdx)
		marker := m.treeMarkerForTask(realIdx)

		cursor := "   "
		if m.cursor == idx {
			cursor = " → "
			if _, ok := m.selected[realIdx]; ok {
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

		line = indent + marker + line
		if m.cursor == idx {
			line = highlightRow(line, width, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}
}

// Wide-layout constants: compact column widths to maximize space for Description.
// Single space between columns; fixed total = 52 so description gets (width - 53) or more.
const (
	wideColCursor          = 3  // " → " or " ✓ "
	wideColID              = 18 // T-xxxxxxxxxxxxx
	wideColStatus          = 11 // "In Progress"
	wideColPriority        = 8  // "PRIORITY" / "medium"
	wideColPRI             = 3  // H/M/L
	wideColOLD             = 3  // "OLD" or "   "
	wideColSpaces          = 6  // one space between each of 7 columns
	wideFixedColsTotal     = wideColCursor + wideColID + wideColStatus + wideColPriority + wideColPRI + wideColOLD + wideColSpaces
	wideMinDescWidth       = 50
	wideTagsColMin         = 24
	wideShowTagsThreshold  = 160
	wideFixedPlusDescSpace = wideFixedColsTotal + 1
)

// renderWideTaskList renders tasks in a wide terminal (>= 120 chars). Uses full width:
// description column grows with terminal width; TAGS column appears when width >= 160.
func (m model) renderWideTaskList(b *strings.Builder, width int) {
	// Compute description and optional tags column widths so each row fills width.
	maxDescWidth := width - wideFixedPlusDescSpace
	if maxDescWidth < wideMinDescWidth {
		maxDescWidth = wideMinDescWidth
	}

	tagsWidth := 0
	if width >= wideShowTagsThreshold {
		// Reserve space for TAGS; description gets the rest.
		tagsWidth = wideTagsColMin

		maxDescWidth = width - wideFixedPlusDescSpace - 1 - tagsWidth
		if maxDescWidth < wideMinDescWidth {
			maxDescWidth = wideMinDescWidth

			tagsWidth = width - wideFixedPlusDescSpace - maxDescWidth - 1
			if tagsWidth < 5 {
				tagsWidth = 0
			}
		}
	}

	descHeader := truncatePad("DESCRIPTION", maxDescWidth)

	header := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*s %-*s %s",
		wideColCursor, "", wideColID, "ID", wideColStatus, "STATUS", wideColPriority, "PRIORITY", wideColPRI, "PRI", wideColOLD, "OLD", descHeader)
	if tagsWidth > 0 {
		header += " " + truncatePad("TAGS", tagsWidth)
	}

	b.WriteString(helpStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", width)))
	b.WriteString("\n")

	vis := m.visibleIndices()
	for idx, realIdx := range vis {
		task := m.tasks[realIdx]
		indent := m.indentForTask(realIdx)
		marker := m.treeMarkerForTask(realIdx)

		cursor := "   "
		if m.cursor == idx {
			cursor = " → "
			if _, ok := m.selected[realIdx]; ok {
				cursor = " ✓ "
			}
		}

		taskID := truncatePad(task.ID, wideColID)

		statusStr := task.Status
		if statusStr == "" {
			statusStr = "---"
		}

		statusStr = truncatePad(statusStr, wideColStatus)

		priorityFull := strings.ToUpper(task.Priority)
		if priorityFull == "" {
			priorityFull = "---"
		}

		priorityFull = truncatePad(priorityFull, wideColPriority)

		priorityShort := "-"
		if task.Priority != "" {
			priorityShort = strings.ToUpper(task.Priority[:1])
		}

		priorityShort = truncatePad(priorityShort, wideColPRI)

		oldStr := "   "
		if isOldSequentialID(task.ID) {
			oldStr = "OLD"
		}

		oldIndicator := truncatePad(oldStr, wideColOLD)

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
			wideColCursor, cursor, wideColID, taskID, wideColStatus, statusStr, wideColPriority, priorityFull, wideColPRI, priorityShort, wideColOLD, oldIndicator, content)

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

		line = indent + marker + line
		if m.cursor == idx {
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

	availableWidth := m.effectiveWidth() - 2
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

	if m.configSaveMessage != "" {
		msgLine := m.configSaveMessage
		if len(msgLine) > availableWidth {
			msgLine = msgLine[:availableWidth-3] + "..."
		}

		if m.configSaveSuccess {
			b.WriteString(headerValueStyle.Render("  ✅ " + msgLine))
		} else {
			b.WriteString(highPriorityStyle.Render("  ❌ " + msgLine))
		}

		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
	}

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
	statusBar.WriteString(" section  ")
	statusBar.WriteString(helpStyle.Render("u"))
	statusBar.WriteString(" update (protobuf)  ")
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
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxContentLines := m.effectiveHeight() - 10
	if maxContentLines < 6 {
		maxContentLines = 6
	}

	wrapWidth := availableWidth - 2
	if wrapWidth < 38 {
		wrapWidth = 38
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
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("Loading handoffs..."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  q quit"))

		return b.String()
	}

	if m.handoffErr != nil {
		b.WriteString("\n  ")
		b.WriteString(normalStyle.Render(fmt.Sprintf("Error: %v", m.handoffErr)))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))

		return b.String()
	}

	if !m.handoffLoading && m.handoffErr == nil && len(m.handoffEntries) == 0 {
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("No handoff notes."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))

		return b.String()
	}

	// Detail view: show full handoff when one is selected
	if m.handoffDetailIndex >= 0 && m.handoffDetailIndex < len(m.handoffEntries) {
		return m.viewHandoffDetail(m.handoffEntries[m.handoffDetailIndex], availableWidth, wrapWidth)
	}

	// List view: show handoffs as list with cursor and selection
	if len(m.handoffEntries) > 0 {
		// Header with count and selection
		b.WriteString(headerLabelStyle.Render("Handoffs:"))
		b.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", len(m.handoffEntries))))

		if len(m.handoffSelected) > 0 {
			b.WriteString(" ")
			b.WriteString(headerLabelStyle.Render("Selected:"))
			b.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", len(m.handoffSelected))))
		}

		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		// Column header
		colCursor := 3
		colNum := 5
		colHost := 18
		colTime := 22

		colSummary := availableWidth - colCursor - colNum - colHost - colTime - 4
		if colSummary < 15 {
			colSummary = 15
		}

		b.WriteString(helpStyle.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s %s", colCursor, "", colNum, "#", colHost, "HOST", colTime, "TIME", "SUMMARY")))
		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")

		for i, h := range m.handoffEntries {
			cursor := "   "
			if m.handoffCursor == i {
				cursor = " → "
				if _, ok := m.handoffSelected[i]; ok {
					cursor = " ✓ "
				}
			} else if _, ok := m.handoffSelected[i]; ok {
				cursor = " ✓ "
			}

			host, _ := h["host"].(string)
			if host == "" {
				host = "—"
			}

			if len(host) > colHost {
				host = host[:colHost-3] + "..."
			}

			ts, _ := h["timestamp"].(string)
			if ts != "" && len(ts) > colTime {
				ts = ts[:colTime-3] + "..."
			}

			if ts == "" {
				ts = "—"
			}

			sum, _ := h["summary"].(string)
			if len(sum) > colSummary {
				sum = sum[:colSummary-3] + "..."
			}

			if sum == "" {
				sum = "(no summary)"
			}

			line := fmt.Sprintf("%-*s %-*d %-*s %-*s %s", colCursor, cursor, colNum, i+1, colHost, host, colTime, ts, sum)
			if m.handoffCursor == i {
				line = highlightRow(line, availableWidth, true)
			} else {
				line = normalStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}

		if m.handoffActionMsg != "" {
			b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
			b.WriteString("\n  ")
			b.WriteString(helpStyle.Render(m.handoffActionMsg))
			b.WriteString("\n")
		}

		if m.childAgentMsg != "" {
			b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
			b.WriteString("\n  ")
			b.WriteString(helpStyle.Render(m.childAgentMsg))
			b.WriteString("\n")
		}

		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Enter detail  Space select  i interactive agent  e run agent & close  x close  a approve  d delete  H back  r refresh  q quit"))

		return b.String()
	}

	// Fallback: try to parse as JSON and render (no list state)
	var payload struct {
		Handoffs []map[string]interface{} `json:"handoffs"`
		Count    int                      `json:"count"`
		Total    int                      `json:"total"`
	}

	if err := json.Unmarshal([]byte(m.handoffText), &payload); err == nil && len(payload.Handoffs) > 0 {
		linesUsed := 0
		for i, h := range payload.Handoffs {
			if linesUsed >= maxContentLines {
				b.WriteString("  ")
				b.WriteString(helpStyle.Render("... (run 'exarp-go session handoffs' for full list)"))
				b.WriteString("\n")

				break
			}
			// Header: Handoff N · host · timestamp, then separator line
			host, _ := h["host"].(string)
			if host == "" {
				host = "unknown"
			}

			ts, _ := h["timestamp"].(string)
			if ts != "" && len(ts) > 25 {
				ts = ts[:25]
			}

			b.WriteString("  ")
			b.WriteString(headerLabelStyle.Render(fmt.Sprintf("Handoff %d", i+1)))
			b.WriteString(headerValueStyle.Render(" · " + host))

			if ts != "" {
				b.WriteString(helpStyle.Render(" · " + ts))
			}

			b.WriteString("\n  ")
			b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
			b.WriteString("\n")

			linesUsed += 2

			// Summary
			if sum, ok := h["summary"].(string); ok && sum != "" {
				b.WriteString("  ")
				b.WriteString(headerValueStyle.Render("Summary:"))
				b.WriteString("\n  ")

				for _, wl := range strings.Split(wordWrap(sum, wrapWidth), "\n") {
					if linesUsed >= maxContentLines {
						b.WriteString(helpStyle.Render("  ..."))
						b.WriteString("\n")

						linesUsed++

						break
					}

					b.WriteString(normalStyle.Render("  "+wl) + "\n")

					linesUsed++
				}
			}

			// Blockers
			if blockers, ok := h["blockers"].([]interface{}); ok && len(blockers) > 0 {
				b.WriteString("  ")
				b.WriteString(headerValueStyle.Render("Blockers:"))
				b.WriteString("\n")

				linesUsed++
				for _, bi := range blockers {
					if linesUsed >= maxContentLines {
						break
					}

					bl := ""

					switch v := bi.(type) {
					case string:
						bl = v
					default:
						bl = fmt.Sprintf("%v", v)
					}

					for _, wl := range strings.Split(wordWrap(bl, wrapWidth), "\n") {
						b.WriteString("  ")
						b.WriteString(normalStyle.Render("  • "+wl) + "\n")

						linesUsed++
					}
				}
			}

			// Next steps
			if steps, ok := h["next_steps"].([]interface{}); ok && len(steps) > 0 {
				b.WriteString("  ")
				b.WriteString(headerValueStyle.Render("Next steps:"))
				b.WriteString("\n")

				linesUsed++
				for _, si := range steps {
					if linesUsed >= maxContentLines {
						break
					}

					st := ""

					switch v := si.(type) {
					case string:
						st = v
					default:
						st = fmt.Sprintf("%v", v)
					}

					for _, wl := range strings.Split(wordWrap(st, wrapWidth), "\n") {
						b.WriteString("  ")
						b.WriteString(normalStyle.Render("  • "+wl) + "\n")

						linesUsed++
					}
				}
			}

			b.WriteString("\n")

			linesUsed++
		}

		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: e run agent & close  E run agent  H back  r refresh  q quit"))

		return b.String()
	}

	// Fallback: plain text (e.g. non-JSON or parse failure) with word wrap
	lines := strings.Split(m.handoffText, "\n")
	textLinesShown := 0

	for _, line := range lines {
		if textLinesShown >= maxContentLines {
			b.WriteString("  ")
			b.WriteString(helpStyle.Render("... (run 'exarp-go session handoffs' for full output)"))
			b.WriteString("\n")

			break
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			b.WriteString("\n")

			textLinesShown++

			continue
		}

		for _, wl := range strings.Split(wordWrap(trimmed, wrapWidth), "\n") {
			if textLinesShown >= maxContentLines {
				break
			}

			b.WriteString("  ")
			b.WriteString(normalStyle.Render(wl))
			b.WriteString("\n")

			textLinesShown++
		}
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Commands: H back  r refresh  q quit"))

	return b.String()
}

func (m model) viewWaves() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	wrapWidth := availableWidth - 2
	if wrapWidth < 30 {
		wrapWidth = 30
	}

	var b strings.Builder

	title := "WAVES (dependency order)"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("H/w=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Enter=expand"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=refresh"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("R=refresh tools"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("U=update from plan"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("A=analysis"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if len(m.waves) == 0 {
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("No backlog waves (Todo/In Progress with dependencies). Refresh tasks (r) or run tools then refresh (R)."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: H or w back  r refresh  R refresh tools  q quit"))

		return b.String()
	}

	levels := sortedWaveLevels(m.waves)
	taskByID := make(map[string]*database.Todo2Task)

	for _, t := range m.tasks {
		if t != nil {
			taskByID[t.ID] = t
		}
	}

	if m.waveDetailLevel >= 0 {
		// Expanded: show tasks for the selected wave only
		ids := m.waves[m.waveDetailLevel]
		levels := sortedWaveLevels(m.waves)

		maxWaveIdx := len(levels) - 1
		if maxWaveIdx < 0 {
			maxWaveIdx = 0
		}

		b.WriteString("  ")
		b.WriteString(headerLabelStyle.Render(fmt.Sprintf("Wave %d", m.waveDetailLevel)))
		b.WriteString(headerValueStyle.Render(fmt.Sprintf(" (%d tasks) — Esc/Enter to collapse", len(ids))))
		b.WriteString("\n  ")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")

		for i, id := range ids {
			content := id
			if t := taskByID[id]; t != nil && t.Content != "" {
				content = truncatePad(t.Content, availableWidth-6)
			}

			line := "  • " + helpStyle.Render(id) + "  " + normalStyle.Render(content)
			if i == m.waveTaskCursor {
				line = highlightRow(line, availableWidth, true)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}

		if m.waveMoveTaskID != "" {
			b.WriteString("\n  ")
			b.WriteString(helpStyle.Render(fmt.Sprintf("Move %s to wave (0-%d): press 0-%d  Esc cancel", m.waveMoveTaskID, maxWaveIdx, maxWaveIdx)))
			b.WriteString("\n")
		}

		if m.waveMoveMsg != "" {
			b.WriteString("  ")
			b.WriteString(statusBarStyle.Render(m.waveMoveMsg))
			b.WriteString("\n")
		}

		if m.waveUpdateMsg != "" {
			b.WriteString("  ")
			b.WriteString(statusBarStyle.Render(m.waveUpdateMsg))
			b.WriteString("\n")
		}
		if m.queueEnqueueMsg != "" {
			b.WriteString("  ")
			b.WriteString(statusBarStyle.Render(m.queueEnqueueMsg))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")

		statusLine := "↑↓ move  m move task  U update from plan  Esc/Enter collapse  E run wave  R refresh  H/w back  q quit"
		if m.queueEnabled {
			statusLine += "  Q enqueue"
		}
		if m.waveMoveTaskID != "" {
			statusLine = "0-9 pick wave  Esc cancel  " + statusLine
		}

		b.WriteString(statusBarStyle.Render(statusLine))

		return b.String()
	}

	// Collapsed: wave list only
	b.WriteString(helpStyle.Render("  #   WAVE        TASKS"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, level := range levels {
		ids := m.waves[level]

		cursor := "   "
		if m.waveCursor == i {
			cursor = " → "
		}

		line := fmt.Sprintf("%s %-3d  Wave %-3d   %d tasks", cursor, i+1, level, len(ids))
		if m.waveCursor == i {
			line = highlightRow(line, availableWidth, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	if m.waveUpdateMsg != "" {
		b.WriteString("  ")
		b.WriteString(statusBarStyle.Render(m.waveUpdateMsg))
		b.WriteString("\n")
	}
	if m.queueEnqueueMsg != "" {
		b.WriteString("  ")
		b.WriteString(statusBarStyle.Render(m.queueEnqueueMsg))
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	statusLine := "Enter expand  E run wave  R refresh tools  U update from plan  A analysis  H/w back  r refresh  q quit"
	if m.queueEnabled {
		statusLine = "[Queue] " + statusLine + "  Q enqueue wave"
	}
	b.WriteString(statusBarStyle.Render(statusLine))

	return b.String()
}

func (m model) viewTaskAnalysis() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxTextLines := m.effectiveHeight() - 10
	if maxTextLines < 4 {
		maxTextLines = 4
	}

	var b strings.Builder

	title := "TASK ANALYSIS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("action=" + m.taskAnalysisAction))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("p=back"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("r=rerun"))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("y=write waves"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if m.taskAnalysisLoading {
		b.WriteString("\n  Running task_analysis (" + m.taskAnalysisAction + ")...\n\n")
		b.WriteString(statusBarStyle.Render("Commands: p back  q quit"))

		return b.String()
	}

	if m.taskAnalysisApproveLoading {
		b.WriteString("\n  Writing waves plan to .cursor/plans/parallel-execution-subagents.plan.md ...\n\n")
		b.WriteString(statusBarStyle.Render("Commands: p back  q quit"))

		return b.String()
	}

	if m.taskAnalysisErr != nil {
		b.WriteString(fmt.Sprintf("\n  Error: %v\n\n", m.taskAnalysisErr))
		b.WriteString(statusBarStyle.Render("Commands: p back  r rerun  y write waves  q quit"))

		return b.String()
	}

	lines := strings.Split(m.taskAnalysisText, "\n")
	shown := 0

	for _, line := range lines {
		if shown >= maxTextLines {
			b.WriteString(helpStyle.Render("  ... (run exarp-go -tool task_analysis for full output)"))
			b.WriteString("\n")

			break
		}

		if len(line) > availableWidth {
			for len(line) > 0 && shown < maxTextLines {
				end := availableWidth
				if end > len(line) {
					end = len(line)
				}

				b.WriteString(normalStyle.Render(line[:end]))
				b.WriteString("\n")

				line = line[end:]
				shown++
			}
		} else {
			b.WriteString(normalStyle.Render(line))
			b.WriteString("\n")

			shown++
		}
	}

	if m.taskAnalysisApproveMsg != "" {
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("  " + m.taskAnalysisApproveMsg))
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Commands: p back  r rerun  y write waves  q quit"))

	return b.String()
}

func (m model) viewJobs() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	wrapWidth := availableWidth - 2
	if wrapWidth < 30 {
		wrapWidth = 30
	}

	var b strings.Builder

	title := "BACKGROUND JOBS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc/b=back"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if len(m.jobs) == 0 {
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render("No background jobs. Launch a child agent (E from task/handoff/wave) to add jobs."))
		b.WriteString("\n\n")
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n")
		b.WriteString(statusBarStyle.Render("Commands: Esc or b back  q quit"))

		return b.String()
	}

	// Job detail view
	if m.jobsDetailIndex >= 0 && m.jobsDetailIndex < len(m.jobs) {
		return m.viewJobDetail(m.jobs[m.jobsDetailIndex], availableWidth, wrapWidth)
	}

	// Job list view
	b.WriteString(helpStyle.Render("  #   KIND      PID    STARTED              PROMPT"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, job := range m.jobs {
		cursor := "   "
		if m.jobsCursor == i {
			cursor = " → "
		}

		pidStr := "—"
		if job.Pid > 0 {
			pidStr = fmt.Sprintf("%d", job.Pid)
		}

		started := job.StartedAt.Format("2006-01-02 15:04")

		promptPreview := job.Prompt
		if len(promptPreview) > availableWidth-45 {
			promptPreview = promptPreview[:availableWidth-48] + "..."
		}

		line := fmt.Sprintf("%s %-3d  %-6s  %-6s  %s  %s", cursor, i+1, string(job.Kind), pidStr, started, promptPreview)
		if m.jobsCursor == i {
			line = highlightRow(line, availableWidth, true)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Enter view detail  Esc/b back  q quit"))

	return b.String()
}

func (m model) viewJobDetail(job BackgroundJob, availableWidth, wrapWidth int) string {
	var b strings.Builder

	title := "JOB DETAIL"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc/Enter back"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Kind:"))
	b.WriteString(headerValueStyle.Render(" " + string(job.Kind)))
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("PID:"))
	b.WriteString(headerValueStyle.Render(fmt.Sprintf(" %d", job.Pid)))
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Started:"))
	b.WriteString(headerValueStyle.Render(" " + job.StartedAt.Format("2006-01-02 15:04:05")))
	b.WriteString("\n  ")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Prompt:"))
	b.WriteString("\n  ")

	for _, line := range strings.Split(wordWrap(job.Prompt, wrapWidth), "\n") {
		b.WriteString(normalStyle.Render("  "+line) + "\n")
	}

	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Output:"))
	b.WriteString("\n")

	if job.Output != "" {
		for _, line := range strings.Split(wordWrap(job.Output, wrapWidth), "\n") {
			b.WriteString("  ")
			b.WriteString(normalStyle.Render(line))
			b.WriteString("\n")
		}
	} else {
		b.WriteString("  ")
		b.WriteString(helpStyle.Render("(running or interactive job — no capture)"))
		b.WriteString("\n")
	}

	b.WriteString("  ")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Esc/Enter back  q quit"))

	return b.String()
}

// viewHandoffDetail renders a single handoff in detail (full summary, blockers, next steps).
func (m model) viewHandoffDetail(h map[string]interface{}, availableWidth, wrapWidth int) string {
	var b strings.Builder

	title := "SESSION HANDOFFS"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc back"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	host, _ := h["host"].(string)
	if host == "" {
		host = "unknown"
	}

	ts, _ := h["timestamp"].(string)
	id, _ := h["id"].(string)

	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Host:"))
	b.WriteString(headerValueStyle.Render(" " + host))
	b.WriteString("  ")
	b.WriteString(headerLabelStyle.Render("Time:"))
	b.WriteString(headerValueStyle.Render(" " + ts))

	if id != "" {
		b.WriteString("  ")
		b.WriteString(helpStyle.Render("ID: " + id))
	}

	b.WriteString("\n  ")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	if sum, ok := h["summary"].(string); ok && sum != "" {
		b.WriteString("  ")
		b.WriteString(headerValueStyle.Render("Summary:"))
		b.WriteString("\n  ")

		for _, wl := range strings.Split(wordWrap(sum, wrapWidth), "\n") {
			b.WriteString(normalStyle.Render("  "+wl) + "\n")
		}

		b.WriteString("\n")
	}

	if blockers, ok := h["blockers"].([]interface{}); ok && len(blockers) > 0 {
		b.WriteString("  ")
		b.WriteString(headerValueStyle.Render("Blockers:"))
		b.WriteString("\n")

		for _, bi := range blockers {
			bl := ""

			switch v := bi.(type) {
			case string:
				bl = v
			default:
				bl = fmt.Sprintf("%v", v)
			}

			for _, wl := range strings.Split(wordWrap(bl, wrapWidth), "\n") {
				b.WriteString("  ")
				b.WriteString(normalStyle.Render("  • "+wl) + "\n")
			}
		}

		b.WriteString("\n")
	}

	if steps, ok := h["next_steps"].([]interface{}); ok && len(steps) > 0 {
		b.WriteString("  ")
		b.WriteString(headerValueStyle.Render("Next steps:"))
		b.WriteString("\n")

		for _, si := range steps {
			st := ""

			switch v := si.(type) {
			case string:
				st = v
			default:
				st = fmt.Sprintf("%v", v)
			}

			for _, wl := range strings.Split(wordWrap(st, wrapWidth), "\n") {
				b.WriteString("  ")
				b.WriteString(normalStyle.Render("  • "+wl) + "\n")
			}
		}

		b.WriteString("\n")
	}

	if m.childAgentMsg != "" {
		b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
		b.WriteString("\n  ")
		b.WriteString(helpStyle.Render(m.childAgentMsg))
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("i interactive  e run & close  x close  a approve  d delete  Esc back  q quit"))

	return b.String()
}

func (m model) viewTaskDetail() string {
	task := m.taskDetailTask
	if task == nil {
		return ""
	}
	// Use effective dimensions so task detail ("s") behaves consistently in iTerm2 and other terminals.
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxContentLines := m.effectiveHeight() - 10
	if maxContentLines < 8 {
		maxContentLines = 8
	}

	var body strings.Builder
	if task.ID != "" {
		body.WriteString("Task: ")
		body.WriteString(task.ID)
		body.WriteString("\n\n")
	}

	if task.Status != "" {
		body.WriteString("Status: ")
		body.WriteString(task.Status)
		body.WriteString("\n\n")
	}

	if task.Priority != "" {
		body.WriteString("Priority: ")
		body.WriteString(task.Priority)
		body.WriteString("\n\n")
	}

	if task.Content != "" {
		body.WriteString("Content:\n")
		body.WriteString(wordWrap(task.Content, availableWidth-2))
		body.WriteString("\n\n")
	}

	if task.LongDescription != "" {
		body.WriteString("Description:\n")
		body.WriteString(wordWrap(task.LongDescription, availableWidth-2))
		body.WriteString("\n\n")
	}

	if len(task.Tags) > 0 {
		body.WriteString("Tags: ")
		body.WriteString(strings.Join(task.Tags, ", "))
		body.WriteString("\n\n")
	}

	if len(task.Dependencies) > 0 {
		body.WriteString("Dependencies: ")
		body.WriteString(strings.Join(task.Dependencies, ", "))
		body.WriteString("\n")
	}

	allLines := strings.Split(strings.TrimSuffix(body.String(), "\n"), "\n")

	var b strings.Builder

	// Header bar (full-width, same as task list)
	title := "TASK DETAIL"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	headerLine := headerStyle.Render(title) + " " + headerLabelStyle.Render("Esc/Enter/Space=close")
	if lipgloss.Width(headerLine) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-lipgloss.Width(headerLine))
		headerLine += headerValueStyle.Render(padding)
	}

	b.WriteString(headerLine)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, line := range allLines {
		if i >= maxContentLines {
			b.WriteString(helpStyle.Render("  ... (truncated)\n"))
			break
		}
		// Constrain line to available width so it never overflows (matches menu behavior)
		if lipgloss.Width(line) > availableWidth {
			runes := []rune(line)
			maxRunes := availableWidth - 3

			if maxRunes > 0 && len(runes) > maxRunes {
				line = string(runes[:maxRunes]) + "..."
			}
		}

		b.WriteString(normalStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	// Status bar (full-width, same as task list)
	statusLine := statusBarStyle.Render("Press Esc, Enter, or Space to close  s=close")
	if lipgloss.Width(statusLine) < availableWidth {
		padding := strings.Repeat(" ", availableWidth-lipgloss.Width(statusLine))
		statusLine += statusBarStyle.Render(padding)
	}

	b.WriteString(statusLine)

	return b.String()
}

func (m model) viewConfigSection() string {
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxContentLines := m.effectiveHeight() - 8
	if maxContentLines < 6 {
		maxContentLines = 6
	}

	wrapped := wordWrap(m.configSectionText, availableWidth)
	allLines := strings.Split(wrapped, "\n")

	var b strings.Builder

	title := "CONFIG SECTION"
	if m.projectName != "" {
		title = fmt.Sprintf("%s - %s", strings.ToUpper(m.projectName), title)
	}

	b.WriteString(headerStyle.Render(title))
	b.WriteString(" ")
	b.WriteString(headerLabelStyle.Render("Esc/Enter/Space=close"))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")

	for i, line := range allLines {
		if i >= maxContentLines {
			b.WriteString(helpStyle.Render("  ... (truncated)\n"))
			break
		}

		b.WriteString(normalStyle.Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", availableWidth)))
	b.WriteString("\n")
	b.WriteString(statusBarStyle.Render("Press Esc, Enter, or Space to close"))

	return b.String()
}

func (m model) viewScorecard() string {
	// Use effective dimensions for consistent layout in iTerm2 and other terminals.
	availableWidth := m.effectiveWidth() - 2
	if availableWidth < 40 {
		availableWidth = 40
	}

	maxTextLines := m.effectiveHeight() - 14
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
	availableWidth := m.effectiveWidth() - 2
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
	b.WriteString(normalStyle.Render("Search & sort (tasks):"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("/"))
	b.WriteString("  Search/filter (type query, Enter apply, Esc cancel)\n  ")
	b.WriteString(helpStyle.Render("n / N"))
	b.WriteString("  Next/previous match\n  ")
	b.WriteString(helpStyle.Render("o"))
	b.WriteString("  Cycle sort order (id → status → priority → updated → hierarchy)\n  ")
	b.WriteString(helpStyle.Render("O"))
	b.WriteString("  Toggle sort direction (asc/desc)\n\n")
	b.WriteString(normalStyle.Render("Views:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("c"))
	b.WriteString("  Switch to Config\n  ")
	b.WriteString(helpStyle.Render("p"))
	b.WriteString("  Switch to Scorecard (project health)\n  ")
	b.WriteString(helpStyle.Render("H"))
	b.WriteString("  Switch to Session handoffs (i interactive agent  e run & close  x close  a approve  d delete)\n  ")
	b.WriteString(helpStyle.Render("w"))
	b.WriteString("  Switch to Waves (Enter expand wave)\n  ")
	b.WriteString(helpStyle.Render("A"))
	b.WriteString("  Task analysis (tasks/waves: run parallelization, show result)\n  ")
	b.WriteString(helpStyle.Render("b"))
	b.WriteString("  Switch to Background jobs (child agent launches)\n\n")
	b.WriteString(normalStyle.Render("Child agent (run Cursor agent in project root):"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("E"))
	b.WriteString("  Execute current context in child agent (task, handoff, or wave)\n  ")
	b.WriteString(helpStyle.Render("i"))
	b.WriteString("  In handoffs: start interactive agent (don't close handoff)\n  ")
	b.WriteString(helpStyle.Render("L"))
	b.WriteString("  Launch plan in child agent (from tasks view)\n\n")
	b.WriteString(normalStyle.Render("Actions:"))
	b.WriteString("\n  ")
	b.WriteString(helpStyle.Render("r"))
	b.WriteString("  Refresh (tasks, config, or scorecard)\n  ")
	b.WriteString(helpStyle.Render("R"))
	b.WriteString("  In waves: run exarp tools (sync, task_analysis) then refresh\n  ")
	b.WriteString(helpStyle.Render("y"))
	b.WriteString("  In task analysis: write waves plan (parallel-execution-subagents.plan.md)\n  ")
	b.WriteString(helpStyle.Render("e"))
	b.WriteString("  Implement selected recommendation (scorecard)\n  ")
	b.WriteString(helpStyle.Render("a"))
	b.WriteString("  Toggle auto-refresh (tasks only)\n  ")
	b.WriteString(helpStyle.Render("s"))
	b.WriteString("  Show task details (tasks); Save/Update config (config): u or s → .exarp/config.pb\n\n")
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
// New format: T- followed by epoch milliseconds (13 digits, typically 1.6+ trillion).
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

		details.WriteString("\nUpdate from TUI: press ")
		details.WriteString("u")
		details.WriteString(" or ")
		details.WriteString("s")
		details.WriteString(" to save to .exarp/config.pb (protobuf).")
		details.WriteString("\nOr use CLI: 'exarp-go config export yaml' to edit as YAML, then 'convert yaml protobuf' to save.")

		return configSectionDetailMsg{text: details.String()}
	}
}

func saveConfig(projectRoot string, cfg *config.FullConfig) tea.Cmd {
	return func() tea.Msg {
		if err := config.WriteConfigToProtobufFile(projectRoot, cfg); err != nil {
			return configSaveResultMsg{message: "Error writing config: " + err.Error(), success: false}
		}

		return configSaveResultMsg{message: "Config saved to " + projectRoot + "/.exarp/config.pb", success: true}
	}
}

// runChildAgentCmd runs the non-interactive child agent, captures output, and returns childAgentResultMsg plus jobCompletedMsg when done.
func runChildAgentCmd(projectRoot, prompt string, kind ChildAgentKind) tea.Cmd {
	return func() tea.Msg {
		r, done := RunChildAgentWithOutputCapture(projectRoot, prompt)
		r.Kind = kind

		return tea.Batch(
			func() tea.Msg { return childAgentResultMsg{Result: r} },
			func() tea.Msg {
				out := <-done
				return jobCompletedMsg(out)
			},
		)
	}
}

// runChildAgentCmdInteractive runs the child agent in interactive mode (Cursor agent with initial prompt, no -p).
func runChildAgentCmdInteractive(projectRoot, prompt string, kind ChildAgentKind) tea.Cmd {
	return func() tea.Msg {
		r := RunChildAgentInteractive(projectRoot, prompt)
		r.Kind = kind

		return childAgentResultMsg{Result: r}
	}
}

type configSavedMsg struct{}

// getProjectName extracts a readable project name from the project root path.
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
