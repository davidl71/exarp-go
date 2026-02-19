package database

import "github.com/davidl71/exarp-go/internal/models"

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

// Re-export constants from models so existing callers of database.StatusTodo etc. keep working.
const (
	StatusTodo       = models.StatusTodo
	StatusInProgress = models.StatusInProgress
	StatusReview     = models.StatusReview
	StatusDone       = models.StatusDone
	StatusCancelled  = models.StatusCancelled
	StatusBlocked    = models.StatusBlocked
)

const (
	PriorityLow      = models.PriorityLow
	PriorityMedium   = models.PriorityMedium
	PriorityHigh     = models.PriorityHigh
	PriorityCritical = models.PriorityCritical
)

const (
	CommentTypeResearch = models.CommentTypeResearch
	CommentTypeResult   = models.CommentTypeResult
	CommentTypeNote     = models.CommentTypeNote
	CommentTypeManual   = models.CommentTypeManual
)

const (
	ActivityTypeCreated       = models.ActivityTypeCreated
	ActivityTypeCommentAdded  = models.ActivityTypeCommentAdded
	ActivityTypeStatusChanged = models.ActivityTypeStatusChanged
	ActivityTypeUpdated       = models.ActivityTypeUpdated
)
