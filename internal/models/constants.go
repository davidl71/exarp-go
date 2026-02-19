package models

// Task status values.
const (
	StatusTodo       = "Todo"
	StatusInProgress = "In Progress"
	StatusReview     = "Review"
	StatusDone       = "Done"
	StatusCancelled  = "Cancelled"
	StatusBlocked    = "Blocked"
)

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
