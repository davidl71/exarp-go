//go:build !unix

package filelock

import "fmt"

func (fl *FileLock) tryLockOS() error {
	return fmt.Errorf("file locking not implemented on this platform")
}

func (fl *FileLock) unlockOS() error {
	return fmt.Errorf("file unlocking not implemented on this platform")
}
