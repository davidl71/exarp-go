// Package utils provides process and file utilities.
// process.go: verify agent process exists (for stale lock detection and T-319).

package utils

import (
	"runtime"

	"golang.org/x/sys/unix"
)

// ProcessExists returns true if the process with the given PID exists and is running.
// Used to verify that an agent holding a task lock is still alive (process monitoring).
// On Unix (darwin, linux, etc.) uses kill(pid, 0); on Windows returns true (no-op)
// to avoid false positives until Windows support is added.
func ProcessExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	if runtime.GOOS == "windows" {
		// Windows: no simple signal-0 check; assume process exists to avoid premature cleanup.
		return true
	}
	err := unix.Kill(pid, 0)
	return err == nil
}
