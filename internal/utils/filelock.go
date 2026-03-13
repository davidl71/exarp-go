// Package utils provides cross-platform file locking for concurrent process access.
package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileLock provides cross-process file-based locking using OS-level locks.
type FileLock struct {
	file    *os.File
	path    string
	timeout time.Duration
}

// NewFileLock creates a new file lock.
func NewFileLock(path string, timeout time.Duration) (*FileLock, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	return &FileLock{file: file, path: path, timeout: timeout}, nil
}

// TryLock attempts to acquire the lock non-blocking.
func (fl *FileLock) TryLock() error {
	if fl.file == nil {
		return fmt.Errorf("lock file not open")
	}
	return fl.tryLockOS()
}

// Lock acquires the lock, blocking until available or timeout.
func (fl *FileLock) Lock() error {
	if fl.timeout == 0 {
		return fl.TryLock()
	}

	start := time.Now()
	pollInterval := 50 * time.Millisecond
	for {
		if err := fl.TryLock(); err == nil {
			return nil
		}
		if time.Since(start) >= fl.timeout {
			return fmt.Errorf("lock timeout after %v", fl.timeout)
		}
		time.Sleep(pollInterval)
	}
}

// Unlock releases the lock.
func (fl *FileLock) Unlock() error {
	if fl.file == nil {
		return nil
	}
	return fl.unlockOS()
}

// Close closes the lock file and releases the lock.
func (fl *FileLock) Close() error {
	if fl.file == nil {
		return nil
	}
	_ = fl.Unlock()
	err := fl.file.Close()
	fl.file = nil
	return err
}

// DefaultGitLockTimeout is the default wait time for acquiring the Git sync lock.
const DefaultGitLockTimeout = 60 * time.Second

// GitSyncLockPath returns the path to the repo-level Git sync lock file.
func GitSyncLockPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".todo2", ".git-sync.lock")
}

// WithGitLock runs fn while holding the Git sync lock for projectRoot.
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
