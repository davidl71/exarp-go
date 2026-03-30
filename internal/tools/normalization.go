// normalization.go — Task status and priority normalization helpers.
package tools

import (
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

// Package-level lookup tables (not per-call map literals) to avoid GC churn on hot paths.
var (
	statusLowerToCanonical = map[string]string{
		"pending":         "todo",
		"not started":     "todo",
		"new":             "todo",
		"in progress":     "in_progress",
		"in-progress":     "in_progress",
		"in_progress":     "in_progress",
		"working":         "in_progress",
		"active":          "in_progress",
		"inprogress":      "in_progress",
		"review":          "review",
		"needs review":    "review",
		"awaiting review": "review",
		"completed":       "completed",
		"done":            "completed",
		"finished":        "completed",
		"closed":          "completed",
		"blocked":         "blocked",
		"waiting":         "blocked",
		"cancelled":       "cancelled",
		"canceled":        "cancelled",
		"abandoned":       "cancelled",
		"todo":            "todo",
	}

	canonicalLowerToTitle = map[string]string{
		"todo":        models.StatusTodo,
		"in_progress": models.StatusInProgress,
		"review":      models.StatusReview,
		"completed":   models.StatusDone,
		"done":        models.StatusDone,
		"blocked":     models.StatusBlocked,
		"cancelled":   models.StatusCancelled,
	}

	priorityLowerToCanonical = map[string]string{
		"low":      "low",
		"medium":   "medium",
		"high":     "high",
		"critical": "critical",
		"lowest":   "low",
		"highest":  "critical",
		"urgent":   "critical",
		"normal":   "medium",
		"standard": "medium",
	}
)

// NormalizeStatus normalizes task status to canonical lowercase form.
//
// Handles case-insensitive matching and variant status values.
// Maps common variants to canonical forms for consistent processing.
//
// Canonical statuses:
//   - "todo": Pending/not started
//   - "in_progress": Currently being worked on
//   - "review": Awaiting review/approval
//   - "completed": Finished (normalizes "done" to "completed")
//   - "blocked": Cannot proceed
//   - "cancelled": Cancelled/abandoned
//
// Returns canonical lowercase status value (defaults to "todo" if empty/invalid).
//
// Examples:
//   - NormalizeStatus("Todo") -> "todo"
//   - NormalizeStatus("DONE") -> "completed"
//   - NormalizeStatus("in-progress") -> "in_progress"
//   - NormalizeStatus("") -> "todo"
func NormalizeStatus(status string) string {
	s := strings.TrimSpace(status)
	if s == "" {
		return "todo"
	}
	statusLower := strings.ToLower(s)
	if canonical, ok := statusLowerToCanonical[statusLower]; ok {
		return canonical
	}
	return statusLower
}

// NormalizeStatusToTitleCase normalizes task status to Title Case for storage/display.
//
// This ensures consistent capitalization in Todo2 files:
//   - "Todo" (not "todo", "TODO", "pending")
//   - "In Progress" (not "in_progress", "in-progress", "In Progress")
//   - "Review" (not "review", "Review")
//   - "Done" (not "done", "DONE", "completed")
//
// Returns Title Case status value (defaults to "Todo" if empty/invalid).
//
// Examples:
//   - NormalizeStatusToTitleCase("todo") -> "Todo"
//   - NormalizeStatusToTitleCase("DONE") -> "Done"
//   - NormalizeStatusToTitleCase("in_progress") -> "In Progress"
//   - NormalizeStatusToTitleCase("") -> "Todo"
func NormalizeStatusToTitleCase(status string) string {
	if strings.TrimSpace(status) == "" {
		return models.StatusTodo
	}

	normalized := NormalizeStatus(status)
	if titleCase, ok := canonicalLowerToTitle[normalized]; ok {
		return titleCase
	}

	words := strings.Fields(normalized)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// NormalizePriority normalizes priority values to canonical lowercase form.
//
// Handles case-insensitive matching and variant priority values.
//
// Canonical priorities:
//   - "low": Low priority
//   - "medium": Medium priority
//   - "high": High priority
//   - "critical": Critical priority
//
// Returns canonical lowercase priority value (defaults to "medium" if empty/invalid).
//
// Examples:
//   - NormalizePriority("High") -> "high"
//   - NormalizePriority("MEDIUM") -> "medium"
//   - NormalizePriority("") -> "medium"
func NormalizePriority(priority string) string {
	s := strings.TrimSpace(priority)
	if s == "" {
		return "medium"
	}
	priorityLower := strings.ToLower(s)
	if canonical, ok := priorityLowerToCanonical[priorityLower]; ok {
		return canonical
	}
	return priorityLower
}

// IsPendingStatus checks if status represents a pending task.
//
// Returns true if status is "todo" (normalized).
func IsPendingStatusNormalized(status string) bool {
	normalized := NormalizeStatus(status)
	return normalized == "todo"
}

// IsCompletedStatusNormalized checks if status represents a completed task.
//
// Returns true if status is "completed" or "cancelled" (normalized).
func IsCompletedStatusNormalized(status string) bool {
	normalized := NormalizeStatus(status)
	return normalized == "completed" || normalized == "cancelled"
}

// IsActiveStatus checks if status represents an active (non-completed) task.
//
// Returns true if status is "todo", "in_progress", "review", or "blocked" (normalized).
func IsActiveStatusNormalized(status string) bool {
	normalized := NormalizeStatus(status)
	return normalized == "todo" || normalized == "in_progress" || normalized == "review" || normalized == "blocked"
}

// IsReviewStatusNormalized checks if status represents a task awaiting review.
//
// Returns true if status is "review" (normalized).
func IsReviewStatusNormalized(status string) bool {
	normalized := NormalizeStatus(status)
	return normalized == "review"
}
