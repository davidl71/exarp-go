package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/queue"
	"github.com/davidl71/exarp-go/internal/tools"
	"golang.org/x/term"
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
	case models.PriorityCritical:
		return 0
	case models.PriorityHigh:
		return 1
	case models.PriorityMedium:
		return 2
	case models.PriorityLow:
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
