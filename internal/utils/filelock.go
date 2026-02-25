// filelock.go — Cross-platform file locking for concurrent process access.
package utils

import (
	"fmt"
	"path/filepath"
	"time"

	mcpfilelock "github.com/davidl71/mcp-go-core/pkg/mcp/filelock"
)

// FileLock provides cross-process file-based locking (re-exported from core).
type FileLock = mcpfilelock.FileLock

// NewFileLock creates a new file lock.
// path: Path to lock file (created if needed).
// timeout: Maximum time to wait for lock (0 = non-blocking).
var NewFileLock = mcpfilelock.NewFileLock

// DefaultGitLockTimeout is the default wait time for acquiring the Git sync lock.
const DefaultGitLockTimeout = 60 * time.Second

// TaskLock creates a file lock for a specific task.
func TaskLock(projectRoot string, taskID string, timeout time.Duration) (*FileLock, error) {
	lockDir := filepath.Join(projectRoot, ".todo2", "locks")
	lockPath := filepath.Join(lockDir, fmt.Sprintf("task_%s.lock", taskID))
	return NewFileLock(lockPath, timeout)
}

// StateFileLock creates a file lock for the entire state file.
func StateFileLock(projectRoot string, timeout time.Duration) (*FileLock, error) {
	lockPath := filepath.Join(projectRoot, ".todo2", "state.todo2.json.lock")
	return NewFileLock(lockPath, timeout)
}

// GitSyncLockPath returns the path to the repo-level Git sync lock file.
// Any process that performs git add/commit/push must acquire this lock first.
func GitSyncLockPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".todo2", ".git-sync.lock")
}

// WithGitLock runs fn while holding the Git sync lock for projectRoot.
// If timeout is 0, DefaultGitLockTimeout is used. The lock is released when fn returns.
// On non-Unix platforms, file locking is not implemented and WithGitLock returns an error.
func WithGitLock(projectRoot string, timeout time.Duration, fn func() error) error {
	if timeout == 0 {
		timeout = DefaultGitLockTimeout
	}
	lock, err := NewFileLock(GitSyncLockPath(projectRoot), timeout)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := lock.Lock(); err != nil {
		return err
	}
	return fn()
}
