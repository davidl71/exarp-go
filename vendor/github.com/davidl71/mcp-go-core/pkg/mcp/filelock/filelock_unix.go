//go:build unix

package filelock

import (
	"fmt"
	"syscall"
)

func (fl *FileLock) tryLockOS() error {
	err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		if err == syscall.EWOULDBLOCK {
			return fmt.Errorf("lock is held by another process")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	return nil
}

func (fl *FileLock) unlockOS() error {
	return syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)
}
