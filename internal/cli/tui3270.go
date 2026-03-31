// tui3270.go — IBM 3270 mainframe TUI: state struct, server setup, connection handling.
// Transactions live in tui3270_transactions.go, menus in tui3270_menu.go,
// command/help logic in tui3270_helpers.go.
package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/tools"
	"github.com/racingmars/go3270"
)

// tui3270Session captures a return-point in the session stack (PF11 swap).
type tui3270Session struct {
	name string    // Human-readable label (e.g. "Tasks", "Scorecard")
	tx   go3270.Tx // Transaction to return to
}

// tui3270State holds the state for a 3270 TUI session.
type tui3270State struct {
	server                framework.MCPServer
	projectRoot           string
	projectName           string
	status                string
	tasks                 []*models.Todo2Task
	cursor                int
	listOffset            int    // For scrolling in list view
	mode                  string // "tasks", "taskdetail", "config", "editor"
	selectedTask          *models.Todo2Task
	devInfo               go3270.DevInfo
	command               string           // Command line input
	filter                string           // Current filter/search term
	scorecardRecs         []string         // Last scorecard recommendations (for Run #)
	scorecardFullModeNext bool             // When true, next scorecard load uses full checks (e.g. after Run #)
	sessionStack          []tui3270Session // Stack of saved sessions for PF11 swap
}

// pushSession saves the current transaction as a named session on the stack.
func (state *tui3270State) pushSession(name string, tx go3270.Tx) {
	const maxStack = 8
	state.sessionStack = append(state.sessionStack, tui3270Session{name: name, tx: tx})
	if len(state.sessionStack) > maxStack {
		state.sessionStack = state.sessionStack[len(state.sessionStack)-maxStack:]
	}
}

// popSession removes and returns the most recent session, or nil if empty.
func (state *tui3270State) popSession() *tui3270Session {
	if len(state.sessionStack) == 0 {
		return nil
	}
	s := state.sessionStack[len(state.sessionStack)-1]
	state.sessionStack = state.sessionStack[:len(state.sessionStack)-1]
	return &s
}

// RunTUI3270 starts a 3270 TUI server in the foreground. Unix daemon detach is handled in
// cli dispatch via tryDetachTUI3270 before setupServer.
func RunTUI3270(server framework.MCPServer, status string, port int) error {
	// Suppress debug logs when running TUI (interactive UI shouldn't show logs)
	CLIOutputOpts.Quiet = true

	return runTUI3270Server(server, status, port)
}

// runTUI3270Server runs the TCP listener and accept loop.
func runTUI3270Server(server framework.MCPServer, status string, port int) error {
	if pf := os.Getenv("EXARP_TUI3270_PIDFILE"); pf != "" {
		defer func() {
			if rmErr := os.Remove(pf); rmErr != nil {
				logWarn(context.Background(), "Failed to remove PID file", "error", rmErr, "operation", "runTUI3270Server", "pid_file", pf)
			}
		}()
	}

	projectRoot, err := tools.FindProjectRoot()
	projectName := ""

	if err != nil {
		logWarn(context.Background(), "Could not find project root", "error", err, "operation", "runTUI3270Server")
	} else {
		projectName = getProjectName(projectRoot)
		EnsureConfigAndDatabase(projectRoot)

		defer func() {
			if err := CloseDatabaseIfOpen(); err != nil {
				logWarn(context.Background(), "Error closing database", "error", err, "operation", "closeDatabase")
			}
		}()
	}

	// Listen for tn3270 connections
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	defer func() {
		if closeErr := listener.Close(); closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
			logWarn(context.Background(), "Error closing listener", "error", closeErr, "operation", "runTUI3270Server")
		}
	}()

	logInfo(context.Background(), "3270 TUI server listening", "port", port, "operation", "runTUI3270Server")
	logInfo(context.Background(), "Connect with x3270", "host", "localhost", "port", port, "operation", "runTUI3270Server")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start accept loop in goroutine
	errChan := make(chan error, 1)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}

				logError(context.Background(), "Error accepting connection", "error", err, "operation", "acceptConnection")

				continue
			}

			go handle3270Connection(conn, server, status, projectRoot, projectName)
		}
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		logInfo(context.Background(), "Received signal, shutting down gracefully", "signal", sig.String(), "operation", "shutdown")

		if closeErr := listener.Close(); closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
			logWarn(context.Background(), "Error closing listener", "error", closeErr, "operation", "shutdown")
		}

		return nil
	case err := <-errChan:
		return err
	}
}

// resolveTUI3270PIDFile returns an absolute path for the PID file.
func resolveTUI3270PIDFile(pidFile string) (string, error) {
	var base string

	if pidFile != "" {
		base = pidFile
	} else {
		projectRoot, err := tools.FindProjectRoot()
		if err == nil && projectRoot != "" {
			base = filepath.Join(projectRoot, ".exarp-go-tui3270.pid")
		} else {
			wd, werr := os.Getwd()
			if werr != nil {
				return "", werr
			}
			base = filepath.Join(wd, ".exarp-go-tui3270.pid")
		}
	}

	out, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}

	return out, nil
}

// stripTUI3270DaemonFlags removes daemon/pid-file flags before re-exec so the child does not
// attempt a second detach.
func stripTUI3270DaemonFlags(args []string) []string {
	out := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		a := args[i]

		switch {
		case a == "--daemon" || a == "-d":
			continue
		case a == "--pid-file" || a == "--pidfile":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
			continue
		case strings.HasPrefix(a, "--pid-file=") || strings.HasPrefix(a, "--pidfile="):
			continue
		default:
			out = append(out, a)
		}
	}

	return out
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
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logWarn(context.Background(), "Error closing connection", "error", closeErr, "operation", "handle3270Connection")
		}
	}()

	devInfo, err := go3270.NegotiateTelnet(conn)
	if err != nil {
		logError(context.Background(), "Telnet negotiation failed", "error", err, "operation", "negotiateTelnet")
		return
	}

	state := &tui3270State{
		server:      server,
		projectRoot: projectRoot,
		projectName: projectName,
		status:      status,
		tasks:       []*models.Todo2Task{},
		cursor:      0,
		listOffset:  0,
		mode:        "tasks",
		devInfo:     devInfo,
		command:     "",
		filter:      "",
	}

	ctx := context.Background()

	state.tasks, err = state.loadTasksForStatus(ctx, state.status)
	if err != nil {
		logError(context.Background(), "Error loading tasks", "error", err, "operation", "loadTasks")
	}

	if err := go3270.RunTransactions(conn, devInfo, state.mainMenuTransaction, state); err != nil {
		logError(context.Background(), "Transaction error", "error", err, "operation", "runTransactions")
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

	for len(lines) < maxLines {
		lines = append(lines, "")
	}

	return lines
}
