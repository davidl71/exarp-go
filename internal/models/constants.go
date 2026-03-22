package models

import "strings"

// Task status values.
const (
	StatusTodo       = "Todo"
	StatusInProgress = "In Progress"
	StatusReview     = "Review"
	StatusDone       = "Done"
	StatusCancelled  = "Cancelled"
	StatusBlocked    = "Blocked"
)

// AllStatuses returns all valid task statuses.
func AllStatuses() []string {
	return []string{StatusTodo, StatusInProgress, StatusReview, StatusDone, StatusCancelled, StatusBlocked}
}

// IsValidStatus reports whether status is a valid task status.
func IsValidStatus(status string) bool {
	for _, s := range AllStatuses() {
		if s == status || strings.ToLower(s) == strings.ToLower(status) {
			return true
		}
	}
	return false
}

// OpenStatuses returns statuses that represent open/unfinished work: Todo, In Progress, Blocked.
// Review is semi-open but typically considered "not yet done".
func OpenStatuses() []string {
	return []string{StatusTodo, StatusInProgress, StatusBlocked}
}

// IsOpenStatus reports whether status is an open status (not Done, Cancelled).
func IsOpenStatus(status string) bool {
	s := strings.ToLower(status)
	for _, open := range OpenStatuses() {
		if strings.ToLower(open) == s {
			return true
		}
	}
	return false
}

// ClosedStatuses returns terminal/closed statuses: Done, Cancelled.
func ClosedStatuses() []string {
	return []string{StatusDone, StatusCancelled}
}

// IsClosedStatus reports whether status is a closed/terminal status.
func IsClosedStatus(status string) bool {
	s := strings.ToLower(status)
	for _, closed := range ClosedStatuses() {
		if strings.ToLower(closed) == s {
			return true
		}
	}
	return false
}

// Task priority values.
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Comment types.
const (
	CommentTypeResearch = "research_with_links"
	CommentTypeResult   = "result"
	CommentTypeNote     = "note"
	CommentTypeManual   = "manualsetup"
)

// Activity types.
const (
	ActivityTypeCreated       = "todo_created"
	ActivityTypeCommentAdded  = "comment_added"
	ActivityTypeStatusChanged = "status_changed"
	ActivityTypeUpdated       = "todo_updated"
)

// LLM backend values.
const (
	BackendFM     = "fm"
	BackendMLX    = "mlx"
	BackendOllama = "ollama"
	BackendAuto   = "auto"
)
