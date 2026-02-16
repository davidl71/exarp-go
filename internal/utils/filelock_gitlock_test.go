package utils

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestGitSyncLockPath(t *testing.T) {
	tests := []struct {
		root string
		want string
	}{
		{"/proj", filepath.Join("/proj", ".todo2", ".git-sync.lock")},
		{"", filepath.Join(".todo2", ".git-sync.lock")},
	}
	for _, tt := range tests {
		got := GitSyncLockPath(tt.root)
		if got != tt.want {
			t.Errorf("GitSyncLockPath(%q) = %q, want %q", tt.root, got, tt.want)
		}
	}
}

func TestWithGitLock_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file locking not implemented on Windows")
	}

	dir := t.TempDir()
	ran := false

	err := WithGitLock(dir, 2*time.Second, func() error {
		ran = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithGitLock: %v", err)
	}

	if !ran {
		t.Error("fn did not run")
	}
}

func TestWithGitLock_PropagatesFnError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file locking not implemented on Windows")
	}

	dir := t.TempDir()
	wantErr := errors.New("fn failed")
	err := WithGitLock(dir, 2*time.Second, func() error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("WithGitLock returned %v, want %v", err, wantErr)
	}
	// Lock should be released; second call should succeed
	ran := false

	err = WithGitLock(dir, 2*time.Second, func() error {
		ran = true
		return nil
	})
	if err != nil {
		t.Errorf("second WithGitLock: %v", err)
	}

	if !ran {
		t.Error("second fn did not run")
	}
}

// TestWithGitLock_Contention verifies that a second caller cannot acquire the lock
// while the first holds it (and times out). Skipped on Darwin and Windows because
// lock semantics may allow same-process re-entrancy or are unimplemented.
func TestWithGitLock_Contention(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file locking not implemented on Windows")
	}

	if runtime.GOOS == "darwin" {
		t.Skip("same-process fcntl lock semantics on Darwin may allow concurrent acquisition; test on Linux for contention")
	}

	dir := t.TempDir()
	firstHolding := make(chan struct{})

	go func() {
		_ = WithGitLock(dir, 5*time.Second, func() error {
			close(firstHolding)
			time.Sleep(300 * time.Millisecond)

			return nil
		})
	}()
	<-firstHolding

	err := WithGitLock(dir, 50*time.Millisecond, func() error {
		return nil
	})
	if err == nil {
		t.Error("expected timeout error from second caller")
	}
}

func TestWithGitLock_DefaultTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file locking not implemented on Windows")
	}

	dir := t.TempDir()
	ran := false

	err := WithGitLock(dir, 0, func() error {
		ran = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithGitLock(0): %v", err)
	}

	if !ran {
		t.Error("fn did not run")
	}
}
