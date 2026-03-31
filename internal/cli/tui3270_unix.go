//go:build !windows

// tui3270_unix.go — detach 3270 TUI server by re-exec in a new session (setsid).
package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func tryDetachTUI3270(pidFileFlag string) (bool, error) {
	if os.Getenv("EXARP_TUI3270_CHILD") == "1" {
		return false, nil
	}

	absPID, err := resolveTUI3270PIDFile(pidFileFlag)
	if err != nil {
		return false, err
	}

	if existingPID, rerr := readPIDFile(absPID); rerr == nil {
		if kerr := syscall.Kill(existingPID, 0); kerr == nil {
			return false, fmt.Errorf("server already running (PID: %d)", existingPID)
		}
		if rmErr := os.Remove(absPID); rmErr != nil {
			logWarn(context.Background(), "Failed to remove stale PID file", "error", rmErr, "operation", "tryDetachTUI3270", "pid_file", absPID)
		}
	}

	logPath := tui3270LogPath(absPID)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return false, fmt.Errorf("open log file: %w", err)
	}

	defer logFile.Close()

	args := stripTUI3270DaemonFlags(os.Args[1:])
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdin = nil
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = append(os.Environ(),
		"EXARP_TUI3270_CHILD=1",
		"EXARP_TUI3270_PIDFILE="+absPID,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return false, fmt.Errorf("start detached process: %w", err)
	}

	childPID := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		logWarn(context.Background(), "Release detached process", "error", err, "operation", "tryDetachTUI3270")
	}

	if err := os.WriteFile(absPID, []byte(strconv.Itoa(childPID)), 0o644); err != nil {
		_ = syscall.Kill(childPID, syscall.SIGTERM)

		return false, fmt.Errorf("write pid file: %w", err)
	}

	fmt.Printf("3270 TUI server started in background (PID: %d)\n", childPID)
	fmt.Printf("PID file: %s\n", absPID)
	fmt.Printf("Log file: %s\n", logPath)
	fmt.Printf("Connect with: x3270 localhost:3270\n")
	fmt.Printf("Stop with: kill %d\n", childPID)

	return true, nil
}

func tui3270LogPath(pidFile string) string {
	if strings.HasSuffix(pidFile, ".pid") {
		return strings.TrimSuffix(pidFile, ".pid") + ".log"
	}

	return pidFile + ".log"
}
