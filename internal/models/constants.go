package models

import "strings"

// TaskStatus is the internal typed representation of a Todo2 task status.
// It is intentionally independent of protobuf (no proto import) so it can be used
// throughout internal packages including database and tools.
type TaskStatus int

const (
	TaskStatusUnspecified TaskStatus = iota
	TaskStatusTodo
	TaskStatusInProgress
	TaskStatusReview
	TaskStatusDone
	TaskStatusCancelled
	TaskStatusBlocked
)

func (s TaskStatus) TitleString() string {
	switch s {
	case TaskStatusTodo:
		return StatusTodo
	case TaskStatusInProgress:
		return StatusInProgress
	case TaskStatusReview:
		return StatusReview
	case TaskStatusDone:
		return StatusDone
	case TaskStatusCancelled:
		return StatusCancelled
	case TaskStatusBlocked:
		return StatusBlocked
	default:
		return ""
	}
}

// CanonicalString returns the canonical lowercase form used by normalization helpers.
func (s TaskStatus) CanonicalString() string {
	switch s {
	case TaskStatusTodo:
		return "todo"
	case TaskStatusInProgress:
		return "in_progress"
	case TaskStatusReview:
		return "review"
	case TaskStatusDone:
		return "completed"
	case TaskStatusCancelled:
		return "cancelled"
	case TaskStatusBlocked:
		return "blocked"
	default:
		return ""
	}
}

// ParseTaskStatus parses a status string into a TaskStatus.
// Accepts both display strings ("In Progress") and canonical/variant strings ("in_progress", "done", etc.).
func ParseTaskStatus(status string) TaskStatus {
	s := strings.TrimSpace(strings.ToLower(status))
	if s == "" {
		return TaskStatusUnspecified
	}

	switch s {
	case "todo", "pending", "not started", "new":
		return TaskStatusTodo
	case "in progress", "in_progress", "in-progress", "working", "active", "inprogress":
		return TaskStatusInProgress
	case "review", "needs review", "awaiting review":
		return TaskStatusReview
	case "done", "completed", "finished", "closed":
		return TaskStatusDone
	case "blocked", "waiting":
		return TaskStatusBlocked
	case "cancelled", "canceled", "abandoned":
		return TaskStatusCancelled
	default:
		return TaskStatusUnspecified
	}
}

// TaskPriority is the internal typed representation of a Todo2 task priority.
type TaskPriority int

const (
	TaskPriorityUnspecified TaskPriority = iota
	TaskPriorityLow
	TaskPriorityMedium
	TaskPriorityHigh
	TaskPriorityCritical
)

func (p TaskPriority) CanonicalString() string {
	switch p {
	case TaskPriorityLow:
		return PriorityLow
	case TaskPriorityMedium:
		return PriorityMedium
	case TaskPriorityHigh:
		return PriorityHigh
	case TaskPriorityCritical:
		return PriorityCritical
	default:
		return ""
	}
}

// ParseTaskPriority parses a priority string into a TaskPriority.
func ParseTaskPriority(priority string) TaskPriority {
	s := strings.TrimSpace(strings.ToLower(priority))
	if s == "" {
		return TaskPriorityUnspecified
	}

	switch s {
	case "low", "lowest":
		return TaskPriorityLow
	case "medium", "normal", "standard":
		return TaskPriorityMedium
	case "high":
		return TaskPriorityHigh
	case "critical", "urgent", "highest":
		return TaskPriorityCritical
	default:
		return TaskPriorityUnspecified
	}
}

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

// AllPriorities returns all valid task priorities.
func AllPriorities() []string {
	return []string{PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical}
}

// IsValidPriority reports whether priority is a valid task priority.
func IsValidPriority(priority string) bool {
	for _, p := range AllPriorities() {
		if strings.EqualFold(p, priority) {
			return true
		}
	}
	return false
}

// IsHighPriority reports whether priority is high or critical.
func IsHighPriority(priority string) bool {
	s := strings.ToLower(priority)
	return s == PriorityHigh || s == PriorityCritical
}

// IsCritical reports whether priority is critical.
func IsCritical(priority string) bool {
	return strings.EqualFold(priority, PriorityCritical)
}

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
