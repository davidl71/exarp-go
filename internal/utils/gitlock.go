package utils

import (
	"path/filepath"
	"time"
)

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
