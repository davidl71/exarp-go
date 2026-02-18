package database

// SchemaVersion represents the current schema version.
// Must match the highest migration version (migrations/00N_*.sql).
const SchemaVersion = 8

// Table names.
const (
	TableTasks            = "tasks"
	TableTaskTags         = "task_tags"
	TableTaskDependencies = "task_dependencies"
	TableTaskChanges      = "task_changes"
	TableTaskComments     = "task_comments"
	TableTaskActivities   = "task_activities"
	TableSchemaMigrations = "schema_migrations"
)

// Column names for tasks table.
const (
	ColTaskID              = "id"
	ColTaskName            = "name"
	ColTaskContent         = "content"
	ColTaskLongDescription = "long_description"
	ColTaskStatus          = "status"
	ColTaskPriority        = "priority"
	ColTaskCompleted       = "completed"
	ColTaskNumber          = "task_number"
	ColTaskEstimatedHours  = "estimated_hours"
	ColTaskActualHours     = "actual_hours"
	ColTaskCreated         = "created"
	ColTaskLastModified    = "last_modified"
	ColTaskCompletedAt     = "completed_at"
	ColTaskProjectID       = "project_id"
	ColTaskMetadata        = "metadata"
	ColTaskVersion         = "version"
	ColTaskCreatedAt       = "created_at"
	ColTaskUpdatedAt       = "updated_at"
	ColTaskParentID        = "parent_id"
	ColTaskAssignedTo      = "assigned_to"
	ColTaskHost            = "host"
	ColTaskAgent           = "agent"
)

// Status values.
const (
	StatusTodo       = "Todo"
	StatusInProgress = "In Progress"
	StatusReview     = "Review"
	StatusDone       = "Done"
	StatusCancelled  = "Cancelled"
	StatusBlocked    = "Blocked"
)

// Priority values.
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
