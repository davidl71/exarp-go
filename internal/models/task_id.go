package models

import (
	"fmt"
	"strings"
	"time"
)

// IsValidTaskID returns true if id is a valid task ID (T-<digits>).
// Rejects empty, "T-NaN", "T-undefined", and any non-numeric suffix.
func IsValidTaskID(id string) bool {
	if id == "" || id == "T-NaN" || id == "T-undefined" {
		return false
	}

	if !strings.HasPrefix(id, "T-") || len(id) <= 2 {
		return false
	}

	for _, c := range id[2:] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}

// GenerateTaskID returns a new task ID in the form T-<epoch_nanoseconds>.
func GenerateTaskID() string {
	return fmt.Sprintf("T-%d", time.Now().UnixNano())
}
