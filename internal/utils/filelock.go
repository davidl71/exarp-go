package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/sys/unix"
)

// FileLock provides cross-process file-based locking using OS-level locks
// Similar to Python's fcntl/msvcrt but native Go implementation
// Currently supports Unix-like systems (macOS, Linux, etc.)
type FileLock struct {
	file    *os.File
	path    string
	timeout time.Duration
	isUnix  bool
}

// NewFileLock creates a new file lock
// path: Path to lock file (will be created if needed)
// timeout: Maximum time to wait for lock (0 = non-blocking)
func NewFileLock(path string, timeout time.Duration) (*FileLock, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	// Open lock file (create if doesn't exist)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	return &FileLock{
		file:    file,
		path:    path,
		timeout: timeout,
		isUnix:  runtime.GOOS != "windows",
	}, nil
}

// TryLock attempts to acquire the lock (non-blocking)
// Returns error if lock is held or failed
func (fl *FileLock) TryLock() error {
	if fl.file == nil {
		return fmt.Errorf("lock file not open")
	}

	// Unix-like systems: use fcntl
	if fl.isUnix {
		err := unix.FcntlFlock(fl.file.Fd(), unix.F_SETLK, &unix.Flock_t{
			Type:   unix.F_WRLCK,
			Whence: 0,
			Start:  0,
			Len:    0, // Lock entire file
		})

		if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
			return fmt.Errorf("lock is held by another process")
		}
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		return nil
	}

	// Windows: Not yet implemented (would use msvcrt.locking)
	// For now, return error on Windows
	return fmt.Errorf("file locking not implemented for Windows (use database locking instead)")
}

// Lock acquires the lock, blocking until available or timeout
// Returns error if timeout exceeded or lock failed
func (fl *FileLock) Lock() error {
	if fl.timeout == 0 {
		// Non-blocking
		return fl.TryLock()
	}

	start := time.Now()
	pollInterval := 50 * time.Millisecond

	for {
		// Try to acquire lock
		err := fl.TryLock()
		if err == nil {
			return nil // Lock acquired
		}

		// Check timeout
		if time.Since(start) >= fl.timeout {
			return fmt.Errorf("lock timeout after %v", fl.timeout)
		}

		// Wait before retry
		time.Sleep(pollInterval)
	}
}

// Unlock releases the lock
func (fl *FileLock) Unlock() error {
	if fl.file == nil {
		return nil // Already closed
	}

	if fl.isUnix {
		err := unix.FcntlFlock(fl.file.Fd(), unix.F_SETLK, &unix.Flock_t{
			Type:   unix.F_UNLCK,
			Whence: 0,
			Start:  0,
			Len:    0,
		})
		return err
	}

	// Windows: Not yet implemented
	return fmt.Errorf("file unlocking not implemented for Windows")
}

// Close closes the lock file and releases the lock
func (fl *FileLock) Close() error {
	if fl.file == nil {
		return nil
	}

	_ = fl.Unlock() // Release lock
	err := fl.file.Close()
	fl.file = nil
	return err
}

// TaskLock creates a file lock for a specific task
// Returns FileLock instance for managing task-level locks
func TaskLock(projectRoot string, taskID string, timeout time.Duration) (*FileLock, error) {
	lockDir := filepath.Join(projectRoot, ".todo2", "locks")
	lockPath := filepath.Join(lockDir, fmt.Sprintf("task_%s.lock", taskID))
	return NewFileLock(lockPath, timeout)
}

// StateFileLock creates a file lock for the entire state file
// Returns FileLock instance for managing state-level locks
func StateFileLock(projectRoot string, timeout time.Duration) (*FileLock, error) {
	lockPath := filepath.Join(projectRoot, ".todo2", "state.todo2.json.lock")
	return NewFileLock(lockPath, timeout)
}

// DefaultGitLockTimeout is the default wait time for acquiring the Git sync lock (T-78).
const DefaultGitLockTimeout = 60 * time.Second

// GitSyncLockPath returns the path to the repo-level Git sync lock file.
// Any process that performs git add/commit/push must acquire this lock first.
func GitSyncLockPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".todo2", ".git-sync.lock")
}

// WithGitLock runs fn while holding the Git sync lock for projectRoot.
// If timeout is 0, DefaultGitLockTimeout is used. The lock is released when fn returns.
// On Windows, file locking is not implemented and WithGitLock returns an error.
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
