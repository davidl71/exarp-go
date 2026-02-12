package utils

import (
	"os"
	"testing"
)

func TestProcessExists(t *testing.T) {
	// Current process must exist
	if !ProcessExists(os.Getpid()) {
		t.Error("ProcessExists(current pid) = false, want true")
	}
	// PID 0 and negative should not exist
	if ProcessExists(0) {
		t.Error("ProcessExists(0) = true, want false")
	}
	if ProcessExists(-1) {
		t.Error("ProcessExists(-1) = true, want false")
	}
	// Very high PID unlikely to exist
	if ProcessExists(99999999) {
		t.Log("ProcessExists(99999999) = true (unusual); on Windows this is a no-op")
	}
}
