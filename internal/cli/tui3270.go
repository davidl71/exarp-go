package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/racingmars/go3270"
)

// tui3270State holds the state for a 3270 TUI session.
type tui3270State struct {
	server                framework.MCPServer
	projectRoot           string
	projectName           string
	status                string
	tasks                 []*database.Todo2Task
	cursor                int
	listOffset            int    // For scrolling in list view
	mode                  string // "tasks", "taskdetail", "config", "editor"
	selectedTask          *database.Todo2Task
	devInfo               go3270.DevInfo
	command               string   // Command line input
	filter                string   // Current filter/search term
	scorecardRecs         []string // Last scorecard recommendations (for Run #)
	scorecardFullModeNext bool     // When true, next scorecard load uses full checks (e.g. after Run #)
}

// loadTasksForStatus loads tasks by status. When status is empty, returns open tasks only (Todo + In Progress).
func loadTasksForStatus(ctx context.Context, status string) ([]*database.Todo2Task, error) {
	if status != "" {
		return database.GetTasksByStatus(ctx, status)
	}

	todo, err := database.GetTasksByStatus(ctx, "Todo")
	if err != nil {
		return nil, err
	}

	inProgress, err := database.GetTasksByStatus(ctx, "In Progress")
	if err != nil {
		return nil, err
	}

	return append(todo, inProgress...), nil
}

// RunTUI3270 starts a 3270 TUI server.
func RunTUI3270(server framework.MCPServer, status string, port int, daemon bool, pidFile string) error {
	// Daemonize if requested
	if daemon {
		return daemonize(pidFile, func() error {
			return runTUI3270Server(server, status, port)
		})
	}

	// Run in foreground
	return runTUI3270Server(server, status, port)
}

// runTUI3270Server runs the actual server (foreground or background).
func runTUI3270Server(server framework.MCPServer, status string, port int) error {
	// Logger initialized automatically via adapter
	// Initialize config and database
	projectRoot, err := tools.FindProjectRoot()
	projectName := ""

	if err != nil {
		logWarn(nil, "Could not find project root", "error", err, "operation", "runTUI3270Server")
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

	// Listen for tn3270 connections
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}
	defer listener.Close()

	logInfo(nil, "3270 TUI server listening", "port", port, "operation", "runTUI3270Server")
	logInfo(nil, "Connect with x3270", "host", "localhost", "port", port, "operation", "runTUI3270Server")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start accept loop in goroutine
	errChan := make(chan error, 1)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Check if listener was closed
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}

				logError(nil, "Error accepting connection", "error", err, "operation", "acceptConnection")

				continue
			}

			// Handle each connection in a goroutine
			go handle3270Connection(conn, server, status, projectRoot, projectName)
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		logInfo(nil, "Received signal, shutting down gracefully", "signal", sig.String(), "operation", "shutdown")
		listener.Close()

		return nil
	case err := <-errChan:
		return err
	}
}

// daemonize runs the server in the background and writes PID file.
func daemonize(pidFile string, serverFunc func() error) error {
	// Determine PID file location
	if pidFile == "" {
		projectRoot, _ := tools.FindProjectRoot()
		if projectRoot != "" {
			pidFile = filepath.Join(projectRoot, ".exarp-go-tui3270.pid")
		} else {
			pidFile = ".exarp-go-tui3270.pid"
		}
	}

	// Check if already running
	if existingPID, err := readPIDFile(pidFile); err == nil {
		// Check if process is still running
		if err := syscall.Kill(existingPID, 0); err == nil {
			return fmt.Errorf("server already running (PID: %d)", existingPID)
		}
		// PID file exists but process is dead, remove it
		os.Remove(pidFile)
	}

	// Write PID file with current process PID
	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Set up cleanup on exit
	defer func() {
		os.Remove(pidFile)
	}()

	// Redirect logs to file when daemonized
	// Note: Logging is handled by slog which outputs to stderr
	// File logging can be added in Phase 2 (JSON output with file support)
	logFile := strings.TrimSuffix(pidFile, ".pid") + ".log"
	_ = logFile // Reserved for future file logging support

	fmt.Printf("3270 TUI server running in background (PID: %d)\n", pid)
	fmt.Printf("PID file: %s\n", pidFile)
	fmt.Printf("Log file: %s\n", logFile)
	fmt.Printf("Connect with: x3270 localhost:3270\n")
	fmt.Printf("Stop with: kill %d\n", pid)

	// Run server (this will block)
	return serverFunc()
}

// readPIDFile reads PID from file.
func readPIDFile(pidFile string) (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

// handle3270Connection handles a single 3270 connection.
func handle3270Connection(conn net.Conn, server framework.MCPServer, status, projectRoot, projectName string) {
	defer conn.Close()

	// Negotiate telnet options
	devInfo, err := go3270.NegotiateTelnet(conn)
	if err != nil {
		logError(nil, "Telnet negotiation failed", "error", err, "operation", "negotiateTelnet")
		return
	}

	// Create initial state
	state := &tui3270State{
		server:      server,
		projectRoot: projectRoot,
		projectName: projectName,
		status:      status,
		tasks:       []*database.Todo2Task{},
		cursor:      0,
		listOffset:  0,
		mode:        "tasks",
		devInfo:     devInfo,
		command:     "",
		filter:      "",
	}

	// Load initial tasks (open only when status is empty)
	ctx := context.Background()

	state.tasks, err = loadTasksForStatus(ctx, state.status)
	if err != nil {
		logError(nil, "Error loading tasks", "error", err, "operation", "loadTasks")
	}

	// Run transactions (main application loop)
	// Start with main menu transaction
	if err := go3270.RunTransactions(conn, devInfo, state.mainMenuTransaction, state); err != nil {
		logError(nil, "Transaction error", "error", err, "operation", "runTransactions")
	}
}

// mainMenuTransaction shows the main menu.
func (state *tui3270State) mainMenuTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	title := "EXARP-GO TASK MANAGEMENT"
	if state.projectName != "" {
		title = fmt.Sprintf("%s - %s", state.projectName, title)
	}

	screen := go3270.Screen{
		{Row: 2, Col: 2, Content: title, Intense: true},
		{Row: 4, Col: 2, Content: "Select an option (1-6) or type a command below:"},
		{Row: 6, Col: 4, Content: "1. Task List"},
		{Row: 7, Col: 4, Content: "2. Configuration"},
		{Row: 8, Col: 4, Content: "3. Scorecard"},
		{Row: 9, Col: 4, Content: "4. Session handoffs"},
		{Row: 10, Col: 4, Content: "5. Exit"},
		{Row: 11, Col: 4, Content: "6. Run in child agent (task/plan/wave/handoff)"},
		{Row: 12, Col: 2, Content: "Option: "},
		{Row: 12, Col: 10, Content: "", Write: true, Name: "option"},
		{Row: 22, Col: 2, Content: "Command ===>", Intense: true},
		{Row: 22, Col: 15, Write: true, Name: "command", Content: ""},
		{Row: 23, Col: 2, Content: "PF1=Help  PF3=Exit  Enter=Select option or run command"},
	}

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	// Check AID (Action ID) for PF keys
	if response.AID == go3270.AIDPF3 {
		return nil, nil, nil // Exit
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	// Check command line first
	cmd := strings.TrimSpace(response.Values["command"])
	if cmd != "" {
		return state.handleCommand(cmd, state.mainMenuTransaction)
	}

	// Check field value for option (allow "1"-"6", or value after "Option:")
	optionRaw := strings.TrimSpace(response.Values["option"])

	option := extractMenuOption(optionRaw)
	switch option {
	case "1":
		return state.taskListTransaction, state, nil
	case "2":
		return state.configTransaction, state, nil
	case "3":
		return state.scorecardTransaction, state, nil
	case "4":
		return state.handoffTransaction, state, nil
	case "5", "":
		if response.AID == go3270.AIDEnter && option == "" {
			return nil, nil, nil // Exit
		}

		return state.mainMenuTransaction, state, nil
	case "6":
		return state.childAgentMenuTransaction, state, nil
	default:
		// Invalid option, stay on main menu
		return state.mainMenuTransaction, state, nil
	}
}

// extractMenuOption returns "1".."6" from user input, or "" if none.
func extractMenuOption(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// Single digit
	if len(s) == 1 && s >= "1" && s <= "6" {
		return s
	}
	// Last character (e.g. "Option: 1" -> "1")
	if len(s) > 0 {
		c := s[len(s)-1:]
		if c >= "1" && c <= "6" {
			return c
		}
	}
	// First digit in string
	for _, r := range s {
		if r >= '1' && r <= '6' {
			return string(r)
		}
	}

	return ""
}

// childAgentMenuTransaction shows Run in child agent submenu (option 6).
func (state *tui3270State) childAgentMenuTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	screen := go3270.Screen{
		{Row: 2, Col: 2, Content: "RUN IN CHILD AGENT", Intense: true},
		{Row: 4, Col: 2, Content: "Launches Cursor CLI 'agent' in project root with a prompt."},
		{Row: 5, Col: 2, Content: "Select (1-5):"},
		{Row: 7, Col: 4, Content: "1. Task (current or first)"},
		{Row: 8, Col: 4, Content: "2. Plan"},
		{Row: 9, Col: 4, Content: "3. Wave (first wave)"},
		{Row: 10, Col: 4, Content: "4. Handoff (first handoff)"},
		{Row: 11, Col: 4, Content: "5. Back to main menu"},
		{Row: 13, Col: 2, Content: "Option: "},
		{Row: 13, Col: 10, Content: "", Write: true, Name: "option_val"},
		{Row: 22, Col: 2, Content: "PF3=Back"},
	}

	opts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}
	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}
	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	opt := strings.TrimSpace(response.Values["option_val"])
	if opt == "" {
		return state.childAgentMenuTransaction, state, nil
	}

	var prompt string
	var kind ChildAgentKind
	switch opt {
	case "5":
		return state.mainMenuTransaction, state, nil
	case "1":
		// Task: use current cursor task or first task
		ctx := context.Background()
		tasks, err := loadTasksForStatus(ctx, state.status)
		if err != nil || len(tasks) == 0 {
			msg := "No tasks"
			if err != nil {
				msg = err.Error()
			}
			return state.showChildAgentResultTransaction(msg, state.mainMenuTransaction), state, nil
		}
		idx := state.cursor
		if idx >= len(tasks) {
			idx = 0
		}
		task := tasks[idx]
		prompt = PromptForTask(task.ID, task.Content)
		kind = ChildAgentTask
	case "2":
		prompt = PromptForPlan(state.projectRoot)
		kind = ChildAgentPlan
	case "3":
		ctx := context.Background()
		tasks, err := loadTasksForStatus(ctx, state.status)
		if err != nil || len(tasks) == 0 {
			msg := "No tasks for wave"
			if err != nil {
				msg = err.Error()
			}
			return state.showChildAgentResultTransaction(msg, state.mainMenuTransaction), state, nil
		}
		taskList := make([]tools.Todo2Task, 0, len(tasks))
		for _, t := range tasks {
			if t != nil {
				taskList = append(taskList, *t)
			}
		}
		_, waves, _, err := tools.BacklogExecutionOrder(taskList, nil)
		if err != nil || len(waves) == 0 {
			msg := "No waves"
			if err != nil {
				msg = err.Error()
			}
			return state.showChildAgentResultTransaction(msg, state.mainMenuTransaction), state, nil
		}
		levels := make([]int, 0, len(waves))
		for k := range waves {
			levels = append(levels, k)
		}
		sort.Ints(levels)
		ids := waves[levels[0]]
		prompt = PromptForWave(levels[0], ids)
		kind = ChildAgentWave
	case "4":
		ctx := context.Background()
		args := map[string]interface{}{"action": "handoff", "sub_action": "list", "limit": float64(5)}
		argsBytes, _ := json.Marshal(args)
		result, err := state.server.CallTool(ctx, "session", argsBytes)
		if err != nil {
			return state.showChildAgentResultTransaction(err.Error(), state.mainMenuTransaction), state, nil
		}
		var text strings.Builder
		for _, c := range result {
			text.WriteString(c.Text)
		}
		var payload struct {
			Handoffs []map[string]interface{} `json:"handoffs"`
		}
		if json.Unmarshal([]byte(text.String()), &payload) != nil || len(payload.Handoffs) == 0 {
			return state.showChildAgentResultTransaction("No handoffs", state.mainMenuTransaction), state, nil
		}
		h := payload.Handoffs[0]
		sum, _ := h["summary"].(string)
		var steps []interface{}
		if s, ok := h["next_steps"].([]interface{}); ok {
			steps = s
		}
		prompt = PromptForHandoff(sum, steps)
		kind = ChildAgentHandoff
	default:
		return state.childAgentMenuTransaction, state, nil
	}

	r := RunChildAgent(state.projectRoot, prompt)
	r.Kind = kind
	msg := r.Message
	if !r.Launched {
		msg = "Error: " + msg
	}
	return state.showChildAgentResultTransaction(msg, state.mainMenuTransaction), state, nil
}

// showChildAgentResultTransaction shows a one-screen result then returns to nextTx.
func (state *tui3270State) showChildAgentResultTransaction(message string, nextTx go3270.Tx) go3270.Tx {
	return func(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
		screen := go3270.Screen{
			{Row: 2, Col: 2, Content: "CHILD AGENT", Intense: true},
			{Row: 4, Col: 2, Content: message},
			{Row: 22, Col: 2, Content: "PF3=Back to menu"},
		}
		if len(message) > 76 {
			// Wrap
			screen = append(screen, go3270.Field{Row: 5, Col: 2, Content: message[76:]})
		}
		opts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}
		response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
		if err != nil {
			return nil, nil, err
		}
		if response.AID == go3270.AIDPF1 {
			return state.helpTransaction, state, nil
		}
		return nextTx, state, nil
	}
}

// helpTransaction shows the help screen (PF1).
func (state *tui3270State) helpTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	lines := []string{
		"EXARP-GO 3270 - HELP",
		"",
		"Main menu options:",
		"  1 = Task List  2 = Config  3 = Scorecard  4 = Handoffs  5 = Exit  6 = Run in child agent",
		"",
		"Commands (type in Command ===> field, then Enter):",
		"  TASKS or T     - Go to Task List",
		"  CONFIG         - Go to Configuration",
		"  SCORECARD or SC - Go to Scorecard",
		"  HANDOFFS or HO - Go to Session handoffs",
		"  MENU or M      - Go to Main Menu",
		"  RUN TASK       - Run current task in child agent",
		"  RUN PLAN       - Run plan in child agent",
		"  RUN WAVE       - Run first wave in child agent",
		"  RUN HANDOFF    - Run first handoff in child agent",
		"  FIND <text>    - Filter tasks by text",
		"  RESET          - Clear filter",
		"  VIEW [id]      - View task details",
		"  EDIT [id]      - Edit task",
		"  TOP / BOTTOM   - Scroll to top/bottom of list",
		"",
		"PF keys:",
		"  PF1 = Help   PF3 = Back/Exit   PF7/PF8 = Scroll (task list)",
		"",
		"Press PF3 to return.",
	}
	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "HELP", Intense: true},
		{Row: 22, Col: 2, Content: "PF3=Back to previous screen"},
	}

	for i, line := range lines {
		if i >= 20 {
			break
		}

		if len(line) > 78 {
			line = line[:75] + "..."
		}

		screen = append(screen, go3270.Field{Row: 2 + i, Col: 2, Content: line})
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	return state.helpTransaction, state, nil
}

// 3270 task list column layout (fixed-width, aligned like mainframe job list).
const (
	t3270ColS        = 2
	t3270ColID       = 5
	t3270ColStatus   = 24
	t3270ColPriority = 37
	t3270ColContent  = 48
	t3270WidS        = 2
	t3270WidID       = 18
	t3270WidStatus   = 12
	t3270WidPriority = 10
	t3270WidContent  = 32
)

func t3270Pad(s string, width int) string {
	if len(s) >= width {
		if width <= 3 {
			return s[:width]
		}

		return s[:width-3] + "..."
	}

	return s + strings.Repeat(" ", width-len(s))
}

// taskListTransaction shows the task list in 3270-style tabular layout (aligned columns, header row, COMMAND/SCROLL line).
func (state *tui3270State) taskListTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	var err error

	state.tasks, err = loadTasksForStatus(ctx, state.status)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	title := "TASK LIST *"
	if state.status != "" {
		title = fmt.Sprintf("TASK LIST (%s) *", state.status)
	}

	if state.filter != "" {
		title = fmt.Sprintf("TASK LIST [%s] *", state.filter)
	}

	maxVisible := 18

	totalLines := len(state.tasks)
	if totalLines == 0 {
		totalLines = 1
	}

	currentLine := state.cursor + 1
	if currentLine < 1 {
		currentLine = 1
	}

	if currentLine > totalLines {
		currentLine = totalLines
	}

	scrollIndicator := "CS"
	if state.listOffset > 0 {
		scrollIndicator = "MORE"
	} else if state.listOffset+maxVisible >= len(state.tasks) && len(state.tasks) > maxVisible {
		scrollIndicator = "BOTTOM"
	} else if len(state.tasks) <= maxVisible {
		scrollIndicator = "CS"
	} else {
		scrollIndicator = "TOP"
	}

	// Row 1: title left, LINE x OF y, COMMAND ===> with input, SCROLL ===> xx (mainframe-style)
	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: title, Intense: true},
		{Row: 1, Col: 45, Content: fmt.Sprintf("LINE %d OF %d", currentLine, totalLines), Intense: true},
		{Row: 1, Col: 55, Content: "COMMAND ===>", Intense: true},
		{Row: 1, Col: 68, Write: true, Name: "command", Content: state.command},
		{Row: 1, Col: 72, Content: "SCROLL ===> " + scrollIndicator, Intense: true},
	}

	// Row 2: blank
	screen = append(screen, go3270.Field{Row: 2, Col: 2, Content: ""})

	// Row 3: column headers (intense, like yellow headers)
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColS, Content: t3270Pad("S", t3270WidS), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColID, Content: t3270Pad("ID", t3270WidID), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColStatus, Content: t3270Pad("STATUS", t3270WidStatus), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColPriority, Content: t3270Pad("PRIORITY", t3270WidPriority), Intense: true})
	screen = append(screen, go3270.Field{Row: 3, Col: t3270ColContent, Content: t3270Pad("CONTENT", t3270WidContent), Intense: true})

	startIdx := state.listOffset

	endIdx := startIdx + maxVisible
	if endIdx > len(state.tasks) {
		endIdx = len(state.tasks)
	}

	if endIdx < startIdx {
		endIdx = startIdx
	}

	for row, i := 4, startIdx; i < endIdx; row, i = row+1, i+1 {
		task := state.tasks[i]

		s := "  "
		if i == state.cursor {
			s = "> "
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColS, Content: t3270Pad(s, t3270WidS)})
		screen = append(screen, go3270.Field{Row: row, Col: t3270ColID, Content: t3270Pad(task.ID, t3270WidID)})

		st := task.Status
		if st == "" {
			st = "-"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColStatus, Content: t3270Pad(st, t3270WidStatus)})

		pri := strings.ToUpper(task.Priority)
		if pri == "" {
			pri = "-"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColPriority, Content: t3270Pad(pri, t3270WidPriority)})

		content := task.Content
		if content == "" {
			content = task.LongDescription
		}

		if content == "" {
			content = "(no description)"
		}

		screen = append(screen, go3270.Field{Row: row, Col: t3270ColContent, Content: t3270Pad(content, t3270WidContent)})
	}

	// Row 22: status left, help text
	statusLine := fmt.Sprintf("Tasks: %d", len(state.tasks))
	if state.cursor < len(state.tasks) {
		statusLine += fmt.Sprintf("  Cursor: %d/%d", state.cursor+1, len(state.tasks))
	}

	if state.filter != "" {
		statusLine += fmt.Sprintf("  Filter: %s", state.filter)
	}

	screen = append(screen, go3270.Field{Row: 22, Col: 2, Content: statusLine})

	// Row 23: PF key line
	screen = append(screen, go3270.Field{
		Row:     23,
		Col:     2,
		Content: "PF1=Help  PF3=Back  PF7=Up  PF8=Down  Enter=Select  PF12=Cancel  PF2=Edit",
	})

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	// Prepare initial values for command line
	initialValues := map[string]string{
		"command": state.command,
	}

	response, err := go3270.ShowScreenOpts(screen, initialValues, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	// Save command line input
	state.command = strings.TrimSpace(response.Values["command"])

	// Handle command line first
	if state.command != "" {
		return state.handleCommand(state.command, state.taskListTransaction)
	}

	// Handle attention keys
	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		state.command = ""
		state.filter = ""
		state.listOffset = 0

		return state.mainMenuTransaction, state, nil
	}

	if response.AID == go3270.AIDPF7 {
		// Scroll up
		if state.cursor > 0 {
			state.cursor--
			// Adjust offset if cursor moved above visible area
			if state.cursor < state.listOffset {
				state.listOffset = state.cursor
			}
		}

		return state.taskListTransaction, state, nil
	}

	if response.AID == go3270.AIDPF8 {
		// Scroll down
		if state.cursor < len(state.tasks)-1 {
			state.cursor++
			// Adjust offset if cursor moved below visible area
			if state.cursor >= state.listOffset+18 {
				state.listOffset = state.cursor - 17
			}
		}

		return state.taskListTransaction, state, nil
	}

	if response.AID == go3270.AIDEnter || response.AID == go3270.AIDPF1 {
		if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			return state.taskDetailTransaction, state, nil
		}
	}

	if response.AID == go3270.AIDPF2 {
		// Edit selected task
		if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			return state.taskEditorTransaction, state, nil
		}
	}

	if response.AID == go3270.AIDPF12 {
		state.command = ""
		state.filter = ""
		state.listOffset = 0

		return state.mainMenuTransaction, state, nil
	}

	// Default: stay on task list
	return state.taskListTransaction, state, nil
}

// taskDetailTransaction shows task details.
func (state *tui3270State) taskDetailTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	if state.selectedTask == nil {
		return state.taskListTransaction, state, nil
	}

	task := state.selectedTask

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "TASK DETAILS", Intense: true},
		{Row: 3, Col: 2, Content: "Task ID:", Intense: true},
		{Row: 3, Col: 12, Content: task.ID},
		{Row: 4, Col: 2, Content: "Status:", Intense: true},
		{Row: 4, Col: 12, Content: task.Status},
		{Row: 5, Col: 2, Content: "Priority:", Intense: true},
		{Row: 5, Col: 12, Content: task.Priority},
	}

	row := 7
	if task.Content != "" {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Content:",
			Intense: true,
		})
		row++
		// Split content into multiple lines if needed
		content := task.Content
		for len(content) > 0 {
			lineLen := 76
			if len(content) < lineLen {
				lineLen = len(content)
			}

			screen = append(screen, go3270.Field{
				Row:     row,
				Col:     2,
				Content: content[:lineLen],
			})
			row++

			if len(content) > lineLen {
				content = content[lineLen:]
			} else {
				break
			}
		}
	}

	if task.LongDescription != "" {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Description:",
			Intense: true,
		})
		row++
		// Split description into multiple lines
		desc := task.LongDescription
		for len(desc) > 0 && row < 20 {
			lineLen := 76
			if len(desc) < lineLen {
				lineLen = len(desc)
			}

			screen = append(screen, go3270.Field{
				Row:     row,
				Col:     2,
				Content: desc[:lineLen],
			})
			row++

			if len(desc) > lineLen {
				desc = desc[lineLen:]
			} else {
				break
			}
		}
	}

	if len(task.Tags) > 0 {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Tags:",
			Intense: true,
		})
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     8,
			Content: strings.Join(task.Tags, ", "),
		})
		row++
	}

	if len(task.Dependencies) > 0 {
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: "Dependencies:",
			Intense: true,
		})
		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     15,
			Content: strings.Join(task.Dependencies, ", "),
		})
		row++
	}

	// Add help line
	screen = append(screen, go3270.Field{
		Row:     23,
		Col:     2,
		Content: "PF3=Back  PF12=Cancel",
	})

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 || response.AID == go3270.AIDPF12 {
		return state.taskListTransaction, state, nil
	}

	if response.AID == go3270.AIDPF2 {
		// Edit task
		return state.taskEditorTransaction, state, nil
	}

	// Default: stay on detail screen
	return state.taskDetailTransaction, state, nil
}

// configTransaction shows configuration editor.
func (state *tui3270State) configTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	cfg, err := config.LoadConfig(state.projectRoot)
	if err != nil {
		// Show error screen instead of failing the connection; user can PF3 back
		msg := err.Error()
		if len(msg) > 72 {
			msg = msg[:69] + "..."
		}

		errScreen := go3270.Screen{
			{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true},
			{Row: 3, Col: 2, Content: "Config could not be loaded (protobuf required)."},
			{Row: 5, Col: 2, Content: "Run: exarp-go config init"},
			{Row: 6, Col: 2, Content: "  or: exarp-go config convert yaml protobuf"},
			{Row: 8, Col: 2, Content: msg},
			{Row: 22, Col: 2, Content: "PF3=Back to menu"},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		response, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts)
		if showErr != nil {
			return nil, nil, showErr
		}

		if response.AID == go3270.AIDPF1 {
			return state.helpTransaction, state, nil
		}

		return state.mainMenuTransaction, state, nil
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "CONFIGURATION", Intense: true},
		{Row: 3, Col: 2, Content: "Configuration sections:"},
		{Row: 5, Col: 4, Content: "1. Timeouts"},
		{Row: 6, Col: 4, Content: "2. Thresholds"},
		{Row: 7, Col: 4, Content: "3. Tasks"},
		{Row: 8, Col: 4, Content: "4. Database"},
		{Row: 9, Col: 4, Content: "5. Security"},
		{Row: 11, Col: 2, Content: "Section:", Write: true, Name: "section"},
		{Row: 22, Col: 2, Content: "PF3=Back  Enter=Select"},
	}

	// Show some config values (simplified - just show that config is loaded)
	if cfg != nil {
		screen = append(screen, go3270.Field{
			Row:     13,
			Col:     2,
			Content: "Configuration loaded successfully",
		})
	}

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	// For now, just go back to main menu
	// In a full implementation, you'd handle section selection
	return state.mainMenuTransaction, state, nil
}

// recommendationToCommand3270 maps scorecard recommendation text to (name, args); prefers Makefile targets; ok false if not runnable.
func recommendationToCommand3270(rec string) (name string, args []string, ok bool) {
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

// runRecommendation3270 runs a recommendation command in projectRoot and returns output and error.
func runRecommendation3270(projectRoot, rec string) (output string, err error) {
	name, args, ok := recommendationToCommand3270(rec)
	if !ok {
		return "", fmt.Errorf("no runnable command for this recommendation")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = projectRoot

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return string(out), nil
}

// scorecardTransaction shows project scorecard (Go scorecard when Go project + project overview).
// When scorecardFullModeNext is set (e.g. after running a recommendation), uses full checks so coverage is shown.
func (state *tui3270State) scorecardTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()

	var combined strings.Builder

	var recommendations []string

	// Use full mode when returning from "Run #" so updated coverage/lint is shown
	useFullMode := state.scorecardFullModeNext
	state.scorecardFullModeNext = false

	if tools.IsGoProject() {
		scorecardOpts := &tools.ScorecardOptions{FastMode: !useFullMode}

		scorecard, err := tools.GenerateGoScorecard(ctx, state.projectRoot, scorecardOpts)
		if err != nil {
			errScreen := go3270.Screen{
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true},
				{Row: 4, Col: 2, Content: "Error: " + err.Error()},
				{Row: 22, Col: 2, Content: "PF3=Back"},
			}
			screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

			if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
				return nil, nil, showErr
			}

			return state.mainMenuTransaction, state, nil
		}

		recommendations = scorecard.Recommendations

		combined.WriteString("=== Go Scorecard ===\n\n")
		combined.WriteString(tools.FormatGoScorecard(scorecard))
	}

	overviewText, err := tools.GetOverviewText(ctx, state.projectRoot)
	if err != nil {
		if combined.Len() > 0 {
			combined.WriteString("\n\n=== Project Overview ===\n\n(overview failed: ")
			combined.WriteString(err.Error())
			combined.WriteString(")")
		} else {
			errScreen := go3270.Screen{
				{Row: 2, Col: 2, Content: "SCORECARD", Intense: true},
				{Row: 4, Col: 2, Content: "Error: " + err.Error()},
				{Row: 22, Col: 2, Content: "PF3=Back"},
			}
			screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

			if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
				return nil, nil, showErr
			}

			return state.mainMenuTransaction, state, nil
		}
	} else {
		if combined.Len() > 0 {
			combined.WriteString("\n\n")
		}

		combined.WriteString("=== Project Overview ===\n\n")
		combined.WriteString(overviewText)
	}

	state.scorecardRecs = recommendations
	text := combined.String()
	lines := strings.Split(text, "\n")
	maxRows := 16

	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "PROJECT SCORECARD", Intense: true},
		{Row: 22, Col: 2, Content: "PF3=Back to menu"},
	}
	if len(state.scorecardRecs) > 0 {
		screen = append(screen,
			go3270.Field{Row: 20, Col: 2, Content: "Run # (1-" + strconv.Itoa(len(state.scorecardRecs)) + "):", Intense: true},
			go3270.Field{Row: 20, Col: 24, Write: true, Name: "run_rec", Content: ""},
		)
		screen[1] = go3270.Field{Row: 22, Col: 2, Content: "PF3=Back  Enter=Run selected #"}
	}

	for i, line := range lines {
		if len(line) > 78 {
			line = line[:75] + "..."
		}

		screen = append(screen, go3270.Field{Row: 2 + i, Col: 2, Content: line})
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}
	// Check if user entered a recommendation number to run
	runRec := strings.TrimSpace(response.Values["run_rec"])
	if runRec != "" && len(state.scorecardRecs) > 0 {
		var n int
		if _, parseErr := fmt.Sscanf(runRec, "%d", &n); parseErr == nil && n >= 1 && n <= len(state.scorecardRecs) {
			rec := state.scorecardRecs[n-1]
			out, runErr := runRecommendation3270(state.projectRoot, rec)
			resultLines := strings.Split(strings.TrimSpace(out), "\n")

			if runErr != nil {
				resultLines = append([]string{"Error: " + runErr.Error()}, resultLines...)
			}

			if len(resultLines) > 18 {
				resultLines = resultLines[:18]
			}

			resultScreen := go3270.Screen{
				{Row: 1, Col: 2, Content: "IMPLEMENT RESULT (#" + runRec + ")", Intense: true},
				{Row: 2, Col: 2, Content: rec},
				{Row: 22, Col: 2, Content: "PF3=Back to scorecard"},
			}

			for i, line := range resultLines {
				if len(line) > 78 {
					line = line[:75] + "..."
				}

				resultScreen = append(resultScreen, go3270.Field{Row: 4 + i, Col: 2, Content: line})
			}

			resp, showErr := go3270.ShowScreenOpts(resultScreen, nil, conn, screenOpts)
			if showErr != nil {
				return nil, nil, showErr
			}

			if resp.AID == go3270.AIDPF1 {
				return state.helpTransaction, state, nil
			}
			// PF3 or any key -> back to scorecard; next load uses full checks so coverage updates
			state.scorecardFullModeNext = true

			return state.scorecardTransaction, state, nil
		}
	}

	return state.scorecardTransaction, state, nil
}

// handoffTransaction shows session handoff notes (from session tool, handoff list).
func (state *tui3270State) handoffTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()
	args := map[string]interface{}{"action": "handoff", "sub_action": "list"}

	argsBytes, err := json.Marshal(args)
	if err != nil {
		errScreen := go3270.Screen{
			{Row: 2, Col: 2, Content: "SESSION HANDOFFS", Intense: true},
			{Row: 4, Col: 2, Content: "Error: " + err.Error()},
			{Row: 22, Col: 2, Content: "PF3=Back to menu"},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
			return nil, nil, showErr
		}

		return state.mainMenuTransaction, state, nil
	}

	result, err := state.server.CallTool(ctx, "session", argsBytes)
	if err != nil {
		errScreen := go3270.Screen{
			{Row: 2, Col: 2, Content: "SESSION HANDOFFS", Intense: true},
			{Row: 4, Col: 2, Content: "Error: " + err.Error()},
			{Row: 22, Col: 2, Content: "PF3=Back to menu"},
		}
		screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

		if _, showErr := go3270.ShowScreenOpts(errScreen, nil, conn, screenOpts); showErr != nil {
			return nil, nil, showErr
		}

		return state.mainMenuTransaction, state, nil
	}

	var text strings.Builder
	for _, c := range result {
		text.WriteString(c.Text)
		text.WriteString("\n")
	}

	content := strings.TrimSpace(text.String())
	lines := strings.Split(content, "\n")
	maxRows := 18

	if len(lines) > maxRows {
		lines = lines[:maxRows]
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "SESSION HANDOFFS", Intense: true},
		{Row: 22, Col: 2, Content: "PF3=Back to menu"},
	}

	for i, line := range lines {
		if len(line) > 78 {
			line = line[:75] + "..."
		}

		screen = append(screen, go3270.Field{Row: 2 + i, Col: 2, Content: line})
	}

	if len(lines) == 0 {
		screen = append(screen, go3270.Field{Row: 2, Col: 2, Content: "No handoff notes."})
	}

	screenOpts := go3270.ScreenOpts{Codepage: devInfo.Codepage()}

	response, err := go3270.ShowScreenOpts(screen, nil, conn, screenOpts)
	if err != nil {
		return nil, nil, err
	}

	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	return state.handoffTransaction, state, nil
}

// taskEditorTransaction shows ISPF-like task editor.
func (state *tui3270State) taskEditorTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	if state.selectedTask == nil {
		return state.taskListTransaction, state, nil
	}

	task := state.selectedTask

	// Split description into lines for editing (max 10 lines)
	descLines := splitIntoLines(task.LongDescription, 10, 70)

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: "EDIT TASK", Intense: true},
		{Row: 3, Col: 2, Content: "Task ID:", Intense: true},
		{Row: 3, Col: 12, Content: task.ID}, // Read-only
		{Row: 4, Col: 2, Content: "Status:", Intense: true},
		{Row: 4, Col: 12, Write: true, Name: "status", Content: task.Status},
		{Row: 5, Col: 2, Content: "Priority:", Intense: true},
		{Row: 5, Col: 12, Write: true, Name: "priority", Content: task.Priority},
		{Row: 7, Col: 2, Content: "Description:", Intense: true},
	}

	// Add description lines as editable fields
	row := 8
	for i, line := range descLines {
		if row > 17 {
			break
		}

		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Write:   true,
			Name:    fmt.Sprintf("desc_line%d", i+1),
			Content: line,
		})
		row++
	}

	// Add command line
	screen = append(screen, go3270.Field{
		Row:     22,
		Col:     2,
		Content: "Command ===>",
		Intense: true,
	})
	screen = append(screen, go3270.Field{
		Row:     22,
		Col:     15,
		Write:   true,
		Name:    "command",
		Content: state.command,
	})

	// Add help line
	screen = append(screen, go3270.Field{
		Row:     23,
		Col:     2,
		Content: "PF3=Exit  PF15=Save  Enter=Execute Command  CANCEL=Abandon",
	})

	opts := go3270.ScreenOpts{
		Codepage: devInfo.Codepage(),
	}

	// Prepare initial values
	initialValues := map[string]string{
		"status":   task.Status,
		"priority": task.Priority,
		"command":  state.command,
	}
	for i, line := range descLines {
		initialValues[fmt.Sprintf("desc_line%d", i+1)] = line
	}

	response, err := go3270.ShowScreenOpts(screen, initialValues, conn, opts)
	if err != nil {
		return nil, nil, err
	}

	// Save command line
	state.command = strings.TrimSpace(response.Values["command"])

	// Handle commands
	cmd := strings.ToUpper(state.command)
	if cmd == "CANCEL" || cmd == "C" {
		state.command = ""
		return state.taskListTransaction, state, nil
	}

	if cmd == "SAVE" || cmd == "S" || response.AID == go3270.AIDPF15 {
		// Save changes
		ctx := context.Background()
		task.Status = strings.TrimSpace(response.Values["status"])
		task.Priority = strings.TrimSpace(response.Values["priority"])

		// Reconstruct description from lines
		var descParts []string

		for i := 1; i <= 10; i++ {
			lineKey := fmt.Sprintf("desc_line%d", i)
			if line, ok := response.Values[lineKey]; ok && strings.TrimSpace(line) != "" {
				descParts = append(descParts, strings.TrimSpace(line))
			}
		}

		task.LongDescription = strings.Join(descParts, "\n")

		// Update task in database
		if err := database.UpdateTask(ctx, task); err != nil {
			logError(nil, "Error updating task", "error", err, "operation", "updateTask", "task_id", task.ID)
			// Show error message (could enhance with error field)
		} else {
			state.command = ""
			return state.taskListTransaction, state, nil
		}
	}

	// Handle other commands
	if state.command != "" {
		return state.handleCommand(state.command, state.taskEditorTransaction)
	}

	// Handle PF keys
	if response.AID == go3270.AIDPF1 {
		return state.helpTransaction, state, nil
	}

	if response.AID == go3270.AIDPF3 {
		state.command = ""
		return state.taskListTransaction, state, nil
	}

	// Default: stay in editor
	return state.taskEditorTransaction, state, nil
}

// handleCommand processes command line input (ISPF-style).
func (state *tui3270State) handleCommand(cmd string, currentTx go3270.Tx) (go3270.Tx, any, error) {
	cmd = strings.TrimSpace(cmd)
	cmdUpper := strings.ToUpper(cmd)

	// Parse command
	parts := strings.Fields(cmdUpper)
	if len(parts) == 0 {
		return currentTx, state, nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "1":
		state.command = ""
		return state.taskListTransaction, state, nil
	case "2":
		state.command = ""
		return state.configTransaction, state, nil
	case "3":
		state.command = ""
		return state.scorecardTransaction, state, nil
	case "4":
		state.command = ""
		return state.handoffTransaction, state, nil
	case "5":
		state.command = ""
		return nil, nil, nil // Exit
	case "SCORECARD", "SC":
		state.command = ""
		return state.scorecardTransaction, state, nil
	case "HANDOFFS", "HO":
		state.command = ""
		return state.handoffTransaction, state, nil
	case "MENU", "M", "MAIN":
		state.command = ""
		return state.mainMenuTransaction, state, nil
	case "TASKS", "T":
		state.command = ""
		return state.taskListTransaction, state, nil
	case "CONFIG":
		state.command = ""
		return state.configTransaction, state, nil
	case "HELP", "H":
		state.command = ""
		return state.helpTransaction, state, nil
	case "FIND", "F":
		// Filter/search tasks
		if len(args) > 0 {
			state.filter = strings.Join(args, " ")
			state.cursor = 0
			state.listOffset = 0
			ctx := context.Background()

			var err error

			state.tasks, err = loadTasksForStatus(ctx, state.status)
			if err == nil {
				// Simple text search filter
				filtered := []*database.Todo2Task{}
				searchTerm := strings.ToLower(state.filter)

				for _, task := range state.tasks {
					content := strings.ToLower(task.Content + " " + task.LongDescription)
					if strings.Contains(content, searchTerm) {
						filtered = append(filtered, task)
					}
				}

				state.tasks = filtered
			}
		} else {
			state.filter = ""
			ctx := context.Background()

			var err error

			state.tasks, err = loadTasksForStatus(ctx, state.status)
			if err != nil {
				logError(nil, "Error reloading tasks", "error", err, "operation", "reloadTasks")
			}
		}

		state.command = ""

		return state.taskListTransaction, state, nil

	case "RESET", "RES":
		// Reset filter
		state.filter = ""
		state.command = ""
		ctx := context.Background()

		var err error

		state.tasks, err = loadTasksForStatus(ctx, state.status)
		if err != nil {
			logError(nil, "Error reloading tasks", "error", err, "operation", "reloadTasks")
		}

		return state.taskListTransaction, state, nil

	case "EDIT", "E":
		// Edit task by ID or line number
		if len(args) > 0 {
			taskID := args[0]
			// Check if it's a line number
			if strings.HasPrefix(taskID, "T-") {
				// Find task by ID
				ctx := context.Background()

				task, err := database.GetTask(ctx, taskID)
				if err == nil {
					state.selectedTask = task
					state.command = ""

					return state.taskEditorTransaction, state, nil
				}
			} else {
				// Try as line number
				var lineNum int
				if _, err := fmt.Sscanf(taskID, "%d", &lineNum); err == nil {
					if lineNum > 0 && lineNum <= len(state.tasks) {
						state.selectedTask = state.tasks[lineNum-1]
						state.cursor = lineNum - 1
						state.command = ""

						return state.taskEditorTransaction, state, nil
					}
				}
			}
		} else if state.cursor < len(state.tasks) {
			// Edit current task
			state.selectedTask = state.tasks[state.cursor]
			state.command = ""

			return state.taskEditorTransaction, state, nil
		}

		state.command = ""

		return currentTx, state, nil

	case "VIEW", "V":
		// View task details
		if len(args) > 0 {
			taskID := args[0]
			if strings.HasPrefix(taskID, "T-") {
				ctx := context.Background()

				task, err := database.GetTask(ctx, taskID)
				if err == nil {
					state.selectedTask = task
					state.command = ""

					return state.taskDetailTransaction, state, nil
				}
			}
		} else if state.cursor < len(state.tasks) {
			state.selectedTask = state.tasks[state.cursor]
			state.command = ""

			return state.taskDetailTransaction, state, nil
		}

		state.command = ""

		return currentTx, state, nil

	case "TOP":
		// Go to top of list
		state.cursor = 0
		state.listOffset = 0
		state.command = ""

		return state.taskListTransaction, state, nil

	case "BOTTOM", "BOT":
		// Go to bottom of list
		if len(state.tasks) > 0 {
			state.cursor = len(state.tasks) - 1
			if state.cursor >= 18 {
				state.listOffset = state.cursor - 17
			}
		}

		state.command = ""

		return state.taskListTransaction, state, nil

	case "RUN":
		// RUN TASK | PLAN | WAVE | HANDOFF - execute in child agent
		state.command = ""
		sub := ""
		if len(args) > 0 {
			sub = args[0]
		}
		switch strings.ToUpper(sub) {
		case "TASK":
			if state.cursor < len(state.tasks) {
				task := state.tasks[state.cursor]
				prompt := PromptForTask(task.ID, task.Content)
				r := RunChildAgent(state.projectRoot, prompt)
				return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
			}
			return state.showChildAgentResultTransaction("No task selected", state.taskListTransaction), state, nil
		case "PLAN":
			prompt := PromptForPlan(state.projectRoot)
			r := RunChildAgent(state.projectRoot, prompt)
			return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
		case "WAVE":
			if len(state.tasks) == 0 {
				return state.showChildAgentResultTransaction("No tasks", state.taskListTransaction), state, nil
			}
			taskList := make([]tools.Todo2Task, 0, len(state.tasks))
			for _, t := range state.tasks {
				if t != nil {
					taskList = append(taskList, *t)
				}
			}
			_, waves, _, err := tools.BacklogExecutionOrder(taskList, nil)
			if err != nil || len(waves) == 0 {
				return state.showChildAgentResultTransaction("No waves", state.taskListTransaction), state, nil
			}
			levels := make([]int, 0, len(waves))
			for k := range waves {
				levels = append(levels, k)
			}
			sort.Ints(levels)
			prompt := PromptForWave(levels[0], waves[levels[0]])
			r := RunChildAgent(state.projectRoot, prompt)
			return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
		case "HANDOFF":
			ctx := context.Background()
			sessionArgs := map[string]interface{}{"action": "handoff", "sub_action": "list", "limit": float64(5)}
			argsBytes, _ := json.Marshal(sessionArgs)
			result, err := state.server.CallTool(ctx, "session", argsBytes)
			if err != nil {
				return state.showChildAgentResultTransaction(err.Error(), state.taskListTransaction), state, nil
			}
			var text strings.Builder
			for _, c := range result {
				text.WriteString(c.Text)
			}
			var payload struct {
				Handoffs []map[string]interface{} `json:"handoffs"`
			}
			if json.Unmarshal([]byte(text.String()), &payload) != nil || len(payload.Handoffs) == 0 {
				return state.showChildAgentResultTransaction("No handoffs", state.taskListTransaction), state, nil
			}
			h := payload.Handoffs[0]
			sum, _ := h["summary"].(string)
			var steps []interface{}
			if s, ok := h["next_steps"].([]interface{}); ok {
				steps = s
			}
			prompt := PromptForHandoff(sum, steps)
			r := RunChildAgent(state.projectRoot, prompt)
			return state.showChildAgentResultTransaction(r.Message, state.taskListTransaction), state, nil
		default:
			return state.showChildAgentResultTransaction("RUN TASK|PLAN|WAVE|HANDOFF", state.taskListTransaction), state, nil
		}

	case "AGENT":
		// Alias for RUN (same args)
		state.command = ""
		parts := append([]string{"RUN"}, args...)
		return state.handleCommand(strings.Join(parts, " "), currentTx)

	default:
		// Unknown command - stay on current screen
		state.command = ""
		return currentTx, state, nil
	}
}

// splitIntoLines splits text into lines with max length and max lines.
func splitIntoLines(text string, maxLines, maxLen int) []string {
	if text == "" {
		return make([]string, maxLines)
	}

	lines := []string{}
	current := text

	for len(lines) < maxLines && len(current) > 0 {
		if len(current) <= maxLen {
			lines = append(lines, current)
			break
		}

		// Try to break at word boundary
		line := current[:maxLen]
		lastSpace := strings.LastIndex(line, " ")

		if lastSpace > maxLen/2 {
			line = current[:lastSpace]
			current = strings.TrimSpace(current[lastSpace:])
		} else {
			current = current[maxLen:]
		}

		lines = append(lines, line)
	}

	// Pad to maxLines
	for len(lines) < maxLines {
		lines = append(lines, "")
	}

	return lines
}
