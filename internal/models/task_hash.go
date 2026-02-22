// task_hash.go — Content hash and normalization helpers for task deduplication and sync conflict detection.
// See docs/TASK_CONTENT_HASH_DESIGN.md for the design document.
package models

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"unicode"
)

// NormalizeForComparison normalizes task content + description for comparison.
// Trims whitespace, lowercases, and collapses multiple spaces into one.
func NormalizeForComparison(content, description string) string {
	combined := content + " " + description
	combined = strings.ToLower(combined)
	combined = strings.TrimSpace(combined)
	var b strings.Builder
	prevSpace := false
	for _, r := range combined {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return b.String()
}

// ContentHashFromString computes SHA-256 hex digest of a normalized string.
func ContentHashFromString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

// ContentHash computes the content hash for a task from its content and long_description.
func ContentHash(task *Todo2Task) string {
	normalized := NormalizeForComparison(task.Content, task.LongDescription)
	return ContentHashFromString(normalized)
}

// SetContentHash computes and stores the content hash in task metadata.
func SetContentHash(task *Todo2Task) {
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}
	task.Metadata["content_hash"] = ContentHash(task)
}

// EnsureContentHash sets the content hash only if not already present.
func EnsureContentHash(task *Todo2Task) {
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}
	if _, ok := task.Metadata["content_hash"]; !ok {
		task.Metadata["content_hash"] = ContentHash(task)
	}
}

// GetContentHash returns the stored content hash, or empty string if not set.
func GetContentHash(task *Todo2Task) string {
	if task.Metadata == nil {
		return ""
	}
	if h, ok := task.Metadata["content_hash"].(string); ok {
		return h
	}
	return ""
}
