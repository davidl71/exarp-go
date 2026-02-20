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
	projectRoot, err := tools.FindProjectRoot()
	projectName := ""

	if err != nil {
		logWarn(context.Background(), "Could not find project root", "error", err, "operation", "runTUI3270Server")
	} else {
		projectName = getProjectName(projectRoot)
		EnsureConfigAndDatabase(projectRoot)

		if database.DB != nil {
			defer func() {
				if err := database.Close(); err != nil {
					logWarn(context.Background(), "Error closing database", "error", err, "operation", "closeDatabase")
				}
			}()
		}
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

// daemonize runs the server in the background and writes PID file.
func daemonize(pidFile string, serverFunc func() error) error {
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
		if err := syscall.Kill(existingPID, 0); err == nil {
			return fmt.Errorf("server already running (PID: %d)", existingPID)
		}
		if rmErr := os.Remove(pidFile); rmErr != nil {
			logWarn(context.Background(), "Failed to remove stale PID file", "error", rmErr, "operation", "daemonize", "pid_file", pidFile)
		}
	}

	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	defer func() {
		if rmErr := os.Remove(pidFile); rmErr != nil {
			logWarn(context.Background(), "Failed to remove PID file", "error", rmErr, "operation", "daemonize", "pid_file", pidFile)
		}
	}()

	logFile := strings.TrimSuffix(pidFile, ".pid") + ".log"
	_ = logFile // Reserved for future file logging support

	fmt.Printf("3270 TUI server running in background (PID: %d)\n", pid)
	fmt.Printf("PID file: %s\n", pidFile)
	fmt.Printf("Log file: %s\n", logFile)
	fmt.Printf("Connect with: x3270 localhost:3270\n")
	fmt.Printf("Stop with: kill %d\n", pid)

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
		tasks:       []*database.Todo2Task{},
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
