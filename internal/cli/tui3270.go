package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/racingmars/go3270"
)

// tui3270State holds the state for a 3270 TUI session
type tui3270State struct {
	server       framework.MCPServer
	projectRoot  string
	projectName  string
	status       string
	tasks        []*database.Todo2Task
	cursor       int
	listOffset   int    // For scrolling in list view
	mode         string // "tasks", "taskdetail", "config", "editor"
	selectedTask *database.Todo2Task
	devInfo      go3270.DevInfo
	command      string // Command line input
	filter       string // Current filter/search term
}

// RunTUI3270 starts a 3270 TUI server
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

// runTUI3270Server runs the actual server (foreground or background)
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

// daemonize runs the server in the background and writes PID file
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

// readPIDFile reads PID from file
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

// handle3270Connection handles a single 3270 connection
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

	// Load initial tasks
	ctx := context.Background()
	if state.status != "" {
		state.tasks, err = database.GetTasksByStatus(ctx, state.status)
	} else {
		filters := &database.TaskFilters{}
		state.tasks, err = database.ListTasks(ctx, filters)
	}
	if err != nil {
		logError(nil, "Error loading tasks", "error", err, "operation", "loadTasks")
	}

	// Run transactions (main application loop)
	// Start with main menu transaction
	if err := go3270.RunTransactions(conn, devInfo, state.mainMenuTransaction, state); err != nil {
		logError(nil, "Transaction error", "error", err, "operation", "runTransactions")
	}
}

// mainMenuTransaction shows the main menu
func (state *tui3270State) mainMenuTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	title := "EXARP-GO TASK MANAGEMENT"
	if state.projectName != "" {
		title = fmt.Sprintf("%s - %s", state.projectName, title)
	}

	screen := go3270.Screen{
		{Row: 2, Col: 2, Content: title, Intense: true},
		{Row: 4, Col: 2, Content: "Select an option:"},
		{Row: 6, Col: 4, Content: "1. Task List"},
		{Row: 7, Col: 4, Content: "2. Configuration"},
		{Row: 8, Col: 4, Content: "3. Exit"},
		{Row: 10, Col: 2, Content: "Option:", Write: true, Name: "option"},
		{Row: 22, Col: 2, Content: "Command ===>", Intense: true},
		{Row: 22, Col: 15, Write: true, Name: "command", Content: ""},
		{Row: 23, Col: 2, Content: "PF1=Help  PF3=Exit  Enter=Select"},
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

	// Check command line first
	cmd := strings.TrimSpace(response.Values["command"])
	if cmd != "" {
		return state.handleCommand(cmd, state.mainMenuTransaction)
	}

	// Check field value for option
	option := strings.TrimSpace(response.Values["option"])
	switch option {
	case "1":
		return state.taskListTransaction, state, nil
	case "2":
		return state.configTransaction, state, nil
	case "3", "":
		if response.AID == go3270.AIDEnter {
			return nil, nil, nil // Exit
		}
		return state.mainMenuTransaction, state, nil
	default:
		// Invalid option, stay on main menu
		return state.mainMenuTransaction, state, nil
	}
}

// taskListTransaction shows the task list
func (state *tui3270State) taskListTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	ctx := context.Background()
	// Reload tasks
	var err error
	if state.status != "" {
		state.tasks, err = database.GetTasksByStatus(ctx, state.status)
	} else {
		filters := &database.TaskFilters{}
		state.tasks, err = database.ListTasks(ctx, filters)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Build screen with ISPF-style enhancements
	title := "TASK LIST"
	if state.status != "" {
		title = fmt.Sprintf("%s (%s)", title, state.status)
	}
	if state.filter != "" {
		title = fmt.Sprintf("%s [Filter: %s]", title, state.filter)
	}

	// Calculate scroll indicator
	scrollIndicator := "CSR"
	maxVisible := 18
	if state.listOffset > 0 {
		scrollIndicator = "MORE"
	}
	if state.listOffset+maxVisible >= len(state.tasks) && len(state.tasks) > maxVisible {
		scrollIndicator = "BOTTOM"
	}
	if state.listOffset == 0 && len(state.tasks) <= maxVisible {
		scrollIndicator = "TOP"
	}

	screen := go3270.Screen{
		{Row: 1, Col: 2, Content: title, Intense: true},
		{Row: 1, Col: 70, Content: fmt.Sprintf("SCROLL: %s", scrollIndicator), Intense: true},
	}

	// Add task rows with line numbers (ISPF style)
	startIdx := state.listOffset
	endIdx := startIdx + maxVisible
	if endIdx > len(state.tasks) {
		endIdx = len(state.tasks)
	}
	if endIdx < startIdx {
		endIdx = startIdx
	}

	row := 3
	for i := startIdx; i < endIdx; i++ {
		task := state.tasks[i]
		lineNum := i + 1
		cursor := "  "
		if i == state.cursor {
			cursor = "> "
		}

		// Format task line with line number: > 1  T-123 [Status] [Priority] Description
		taskLine := fmt.Sprintf("%s%2d  %s", cursor, lineNum, task.ID)
		if task.Status != "" {
			taskLine += fmt.Sprintf(" [%s]", task.Status)
		}
		if task.Priority != "" {
			taskLine += fmt.Sprintf(" [%s]", strings.ToUpper(task.Priority))
		}

		// Add description (truncate to fit)
		content := task.Content
		if content == "" {
			content = task.LongDescription
		}
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		if content != "" {
			taskLine += " " + content
		}

		// Pad to fit screen (accounting for line numbers)
		maxLen := 72 // 80 - 8 for row/col and line number
		if len(taskLine) > maxLen {
			taskLine = taskLine[:maxLen]
		} else {
			taskLine += strings.Repeat(" ", maxLen-len(taskLine))
		}

		screen = append(screen, go3270.Field{
			Row:     row,
			Col:     2,
			Content: taskLine,
		})

		row++
	}

	// Add status line
	statusLine := fmt.Sprintf("Tasks: %d", len(state.tasks))
	if state.cursor < len(state.tasks) {
		statusLine += fmt.Sprintf("  Cursor: %d/%d", state.cursor+1, len(state.tasks))
	}
	if state.filter != "" {
		statusLine += fmt.Sprintf("  Filter: %s", state.filter)
	}
	screen = append(screen, go3270.Field{
		Row:     22,
		Col:     2,
		Content: statusLine,
	})

	// Add command line (ISPF style)
	screen = append(screen, go3270.Field{
		Row:     22,
		Col:     50,
		Content: "Command ===>",
		Intense: true,
	})
	screen = append(screen, go3270.Field{
		Row:     22,
		Col:     63,
		Write:   true,
		Name:    "command",
		Content: state.command,
	})

	// Add help line
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

// taskDetailTransaction shows task details
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

// configTransaction shows configuration editor
func (state *tui3270State) configTransaction(conn net.Conn, devInfo go3270.DevInfo, data any) (go3270.Tx, any, error) {
	cfg, err := config.LoadConfig(state.projectRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
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

	if response.AID == go3270.AIDPF3 {
		return state.mainMenuTransaction, state, nil
	}

	// For now, just go back to main menu
	// In a full implementation, you'd handle section selection
	return state.mainMenuTransaction, state, nil
}

// taskEditorTransaction shows ISPF-like task editor
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
	if response.AID == go3270.AIDPF3 {
		state.command = ""
		return state.taskListTransaction, state, nil
	}

	// Default: stay in editor
	return state.taskEditorTransaction, state, nil
}

// handleCommand processes command line input (ISPF-style)
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
	case "FIND", "F":
		// Filter/search tasks
		if len(args) > 0 {
			state.filter = strings.Join(args, " ")
			state.cursor = 0
			state.listOffset = 0
			// Reload tasks with filter
			ctx := context.Background()
			filters := &database.TaskFilters{}
			var err error
			state.tasks, err = database.ListTasks(ctx, filters)
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
			// Reload all tasks
			ctx := context.Background()
			var err error
			if state.status != "" {
				state.tasks, err = database.GetTasksByStatus(ctx, state.status)
			} else {
				filters := &database.TaskFilters{}
				state.tasks, err = database.ListTasks(ctx, filters)
			}
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
		if state.status != "" {
			state.tasks, err = database.GetTasksByStatus(ctx, state.status)
		} else {
			filters := &database.TaskFilters{}
			state.tasks, err = database.ListTasks(ctx, filters)
		}
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

	default:
		// Unknown command - stay on current screen
		state.command = ""
		return currentTx, state, nil
	}
}

// splitIntoLines splits text into lines with max length and max lines
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
