package tasksync

import (
	"context"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/models"
)

// ExternalSourceType represents the type of external task source
type ExternalSourceType string

const (
	SourceAgenticTools  ExternalSourceType = "agentic-tools"
	SourceGoogleTasks   ExternalSourceType = "google-tasks"
	SourceAnyDo         ExternalSourceType = "anydo"
	SourceMicrosoftTodo ExternalSourceType = "microsoft-todo"
	SourceGitHubIssues  ExternalSourceType = "github-issues"
)

// ExternalSource defines the interface for external task synchronization
type ExternalSource interface {
	// Type returns the source type
	Type() ExternalSourceType

	// Name returns a human-readable name for the source
	Name() string

	// SyncTasks synchronizes tasks between Todo2 and the external source
	// Returns sync results including matches, conflicts, new tasks, and updates
	SyncTasks(ctx context.Context, todo2Tasks []*models.Todo2Task, options SyncOptions) (*SyncResult, error)

	// GetTasks retrieves tasks from the external source
	GetTasks(ctx context.Context, options GetTasksOptions) ([]*ExternalTask, error)

	// CreateTask creates a task in the external source
	CreateTask(ctx context.Context, task *models.Todo2Task) (*ExternalTask, error)

	// UpdateTask updates a task in the external source
	UpdateTask(ctx context.Context, externalID string, task *models.Todo2Task) (*ExternalTask, error)

	// DeleteTask deletes a task from the external source
	DeleteTask(ctx context.Context, externalID string) error

	// TestConnection tests the connection to the external source
	TestConnection(ctx context.Context) error
}

// ExternalTask represents a task from an external source
type ExternalTask struct {
	// ExternalID is the unique identifier in the external system
	ExternalID string

	// Title is the task title/name
	Title string

	// Description is the task description
	Description string

	// Status is the task status (will be mapped to Todo2 status)
	Status string

	// Priority is the task priority
	Priority string

	// DueDate is the task due date
	DueDate *time.Time

	// Completed is whether the task is completed
	Completed bool

	// Tags are task tags/labels
	Tags []string

	// Metadata contains additional source-specific metadata
	Metadata map[string]interface{}

	// Source is the external source type
	Source ExternalSourceType

	// LastModified is when the task was last modified
	LastModified time.Time
}

// SyncOptions contains options for task synchronization
type SyncOptions struct {
	// DryRun performs a dry run without making changes
	DryRun bool

	// Direction specifies sync direction (bidirectional, todo2_to_external, external_to_todo2)
	Direction SyncDirection

	// ConflictResolution specifies how to handle conflicts
	ConflictResolution ConflictResolution

	// StatusMapping is a custom status mapping (optional, uses default if nil)
	StatusMapping map[string]string
}

// SyncDirection specifies the direction of synchronization
type SyncDirection string

const (
	DirectionBidirectional   SyncDirection = "bidirectional"
	DirectionTodo2ToExternal SyncDirection = "todo2_to_external"
	DirectionExternalToTodo2 SyncDirection = "external_to_todo2"
)

// ConflictResolution specifies how to handle conflicts
type ConflictResolution string

const (
	ConflictResolutionTodo2Wins    ConflictResolution = "todo2_wins"
	ConflictResolutionExternalWins ConflictResolution = "external_wins"
	ConflictResolutionNewerWins    ConflictResolution = "newer_wins"
	ConflictResolutionManual       ConflictResolution = "manual"
)

// GetTasksOptions contains options for retrieving tasks
type GetTasksOptions struct {
	// Status filters tasks by status
	Status string

	// IncludeCompleted includes completed tasks
	IncludeCompleted bool

	// Limit limits the number of tasks returned
	Limit int

	// Since filters tasks modified since this time
	Since *time.Time
}

// SyncResult contains the results of a synchronization operation
type SyncResult struct {
	// Matches are tasks that exist in both systems and match
	Matches []TaskMatch

	// Conflicts are tasks that exist in both systems but differ
	Conflicts []TaskConflict

	// NewTodo2Tasks are tasks created in Todo2 from external source
	NewTodo2Tasks []*models.Todo2Task

	// NewExternalTasks are tasks created in external source from Todo2
	NewExternalTasks []*ExternalTask

	// UpdatedTasks are tasks that were updated
	UpdatedTasks []TaskUpdate

	// Errors are any errors encountered during sync
	Errors []SyncError

	// Stats contains synchronization statistics
	Stats SyncStats
}

// TaskMatch represents a task that exists in both systems and matches
type TaskMatch struct {
	Todo2Task    *models.Todo2Task
	ExternalTask *ExternalTask
	MatchScore   float64 // 0.0 to 1.0, how well they match
}

// TaskConflict represents a task that exists in both systems but differs
type TaskConflict struct {
	Todo2Task    *models.Todo2Task
	ExternalTask *ExternalTask
	Differences  []string // List of fields that differ
}

// TaskUpdate represents a task that was updated during sync
type TaskUpdate struct {
	TaskID  string
	Source  ExternalSourceType
	Updated bool
	Fields  []string // Fields that were updated
}

// SyncError represents an error during synchronization
type SyncError struct {
	TaskID    string
	Source    ExternalSourceType
	Operation string
	Error     error
}

// SyncStats contains synchronization statistics
type SyncStats struct {
	TotalTodo2Tasks    int
	TotalExternalTasks int
	MatchesFound       int
	ConflictsDetected  int
	NewTodo2Tasks      int
	NewExternalTasks   int
	UpdatedTasks       int
	Errors             int
	Duration           time.Duration
}

// StatusMapper handles status mapping between Todo2 and external sources
type StatusMapper interface {
	// Todo2ToExternal maps Todo2 status to external source status
	Todo2ToExternal(todo2Status string) string

	// ExternalToTodo2 maps external source status to Todo2 status
	ExternalToTodo2(externalStatus string) string
}

// DefaultStatusMapper provides default status mapping
type DefaultStatusMapper struct {
	todo2ToExternal map[string]string
	externalToTodo2 map[string]string
}

// NewDefaultStatusMapper creates a default status mapper
func NewDefaultStatusMapper() *DefaultStatusMapper {
	return &DefaultStatusMapper{
		todo2ToExternal: map[string]string{
			"Todo":        "pending",
			"In Progress": "in-progress",
			"Done":        "completed",
			"Review":      "in-progress",
			"Cancelled":   "cancelled",
		},
		externalToTodo2: map[string]string{
			"pending":     "Todo",
			"in-progress": "In Progress",
			"completed":   "Done",
			"done":        "Done",
			"cancelled":   "Cancelled",
		},
	}
}

// Todo2ToExternal maps Todo2 status to external source status
func (m *DefaultStatusMapper) Todo2ToExternal(todo2Status string) string {
	if mapped, ok := m.todo2ToExternal[todo2Status]; ok {
		return mapped
	}
	return "pending" // Default fallback
}

// ExternalToTodo2 maps external source status to Todo2 status
func (m *DefaultStatusMapper) ExternalToTodo2(externalStatus string) string {
	if mapped, ok := m.externalToTodo2[externalStatus]; ok {
		return mapped
	}
	return "Todo" // Default fallback
}

// ExternalSourceRegistry manages external task sources
type ExternalSourceRegistry struct {
	sources map[ExternalSourceType]ExternalSource
}

// NewExternalSourceRegistry creates a new registry
func NewExternalSourceRegistry() *ExternalSourceRegistry {
	return &ExternalSourceRegistry{
		sources: make(map[ExternalSourceType]ExternalSource),
	}
}

// Register registers an external source
func (r *ExternalSourceRegistry) Register(source ExternalSource) {
	r.sources[source.Type()] = source
}

// Get retrieves a source by type
func (r *ExternalSourceRegistry) Get(sourceType ExternalSourceType) (ExternalSource, error) {
	source, ok := r.sources[sourceType]
	if !ok {
		return nil, fmt.Errorf("external source %s not registered", sourceType)
	}
	return source, nil
}

// List returns all registered source types
func (r *ExternalSourceRegistry) List() []ExternalSourceType {
	types := make([]ExternalSourceType, 0, len(r.sources))
	for t := range r.sources {
		types = append(types, t)
	}
	return types
}

// Global registry
var globalRegistry *ExternalSourceRegistry

func init() {
	globalRegistry = NewExternalSourceRegistry()
}

// RegisterExternalSource registers a source in the global registry
func RegisterExternalSource(source ExternalSource) {
	globalRegistry.Register(source)
}

// GetExternalSource retrieves a source from the global registry
func GetExternalSource(sourceType ExternalSourceType) (ExternalSource, error) {
	return globalRegistry.Get(sourceType)
}

// ListExternalSources returns all registered source types
func ListExternalSources() []ExternalSourceType {
	return globalRegistry.List()
}
