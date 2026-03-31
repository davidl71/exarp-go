// todo2_utils.go — Shared Todo2 utilities: load, save, format, and sync helpers.
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// Todo2Task is an alias for models.Todo2Task (for backward compatibility).
type Todo2Task = models.Todo2Task

// Todo2State is an alias for models.Todo2State (for backward compatibility).
type Todo2State = models.Todo2State

// LoadTodo2Tasks loads tasks from database (preferred) or .todo2/state.todo2.json (fallback).
func LoadTodo2Tasks(projectRoot string) ([]Todo2Task, error) {
	// Try database first (scoped to project when projectRoot is set)
	if tasks, err := loadTodo2TasksFromDB(projectRoot); err == nil {
		return tasks, nil
	}

	// Database not available or query failed, fallback to JSON
	return loadTodo2TasksFromJSON(projectRoot)
}

// loadTodo2TasksFromJSON loads tasks from JSON file (fallback method).
// This is canonical-only: no alias fields (title/description, created/updated) are supported.
func loadTodo2TasksFromJSON(projectRoot string) ([]Todo2Task, error) {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")
	data, err := os.ReadFile(todo2Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo2Task{}, nil
		}
		return nil, fmt.Errorf("failed to read Todo2 file: %w", err)
	}

	tasks, err := ParseTasksFromJSON(data)
	if err != nil {
		return nil, err
	}

	projectID := filepath.Base(projectRoot)
	if projectID != "" && projectID != "." {
		for i := range tasks {
			if tasks[i].ProjectID == "" {
				tasks[i].ProjectID = projectID
			}
		}
	}

	for i := range tasks {
		models.EnsureContentHash(&tasks[i])
	}

	return tasks, nil
}

// SaveTodo2Tasks saves tasks to database (preferred) or .todo2/state.todo2.json (fallback).
// When database save succeeds, also writes the same list to JSON so both stores stay in sync
// (avoids merge/sync reintroducing removed tasks from stale JSON).
func SaveTodo2Tasks(projectRoot string, tasks []Todo2Task) error {
	// Try database first
	if err := saveTodo2TasksToDB(projectRoot, tasks); err == nil {
		// Keep JSON in sync so a later sync does not reintroduce removed tasks (e.g. after merge)
		if jsonErr := saveTodo2TasksToJSON(projectRoot, tasks); jsonErr != nil {
			return fmt.Errorf("database saved but JSON write failed: %w", jsonErr)
		}
		return nil
	}

	// Database not available or save failed, fallback to JSON
	return saveTodo2TasksToJSON(projectRoot, tasks)
}

// saveTodo2TasksToJSON saves tasks to JSON file (fallback method).
// Writes canonical state.todo2.json with no legacy alias fields.
func saveTodo2TasksToJSON(projectRoot string, tasks []Todo2Task) error {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")
	if err := os.MkdirAll(filepath.Dir(todo2Path), 0755); err != nil {
		return fmt.Errorf("failed to create .todo2 directory: %w", err)
	}

	for i := range tasks {
		tasks[i].NormalizeEpochDates()
	}

	data, err := MarshalTasksToStateJSON(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal Todo2 state: %w", err)
	}

	if err := os.WriteFile(todo2Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write Todo2 file: %w", err)
	}
	return nil
}

// FindProjectRoot finds the exarp project root. Delegates to config.FindProjectRoot() (single implementation).
// Use this for tools, resources, and handlers that need the project root.
func FindProjectRoot() (string, error) {
	return config.FindProjectRoot()
}

// GetProjectRootWithFallback returns project root: FindProjectRoot, else PROJECT_ROOT env.
// Use when a best-effort root is needed, but still require an explicit or discoverable project root.
func GetProjectRootWithFallback() (string, error) {
	root, err := config.FindProjectRoot()
	if err == nil && root != "" {
		return root, nil
	}
	if env := os.Getenv("PROJECT_ROOT"); env != "" && !strings.Contains(env, "{{PROJECT_ROOT}}") {
		return filepath.Clean(env), nil
	}
	return "", fmt.Errorf("project root not found; set PROJECT_ROOT or run from a project with .todo2/.exarp markers")
}

// SyncTodo2Tasks synchronizes tasks between database and JSON file
// It loads from both sources, merges them (database takes precedence for conflicts),
// and saves to both to ensure consistency.
func SyncTodo2Tasks(projectRoot string) error {
	// Load from both sources (DB scoped to current project)
	dbTasksLoaded, dbErr := loadTodo2TasksFromDB(projectRoot)
	jsonTasksLoaded, _ := loadTodo2TasksFromJSON(projectRoot)

	// Build merged task map (database takes precedence)
	taskMap := make(map[string]Todo2Task)
	for _, task := range jsonTasksLoaded {
		taskMap[task.ID] = task
	}
	for _, task := range dbTasksLoaded {
		taskMap[task.ID] = task
	}

	mergedTasks := make([]Todo2Task, 0, len(taskMap))
	for _, task := range taskMap {
		mergedTasks = append(mergedTasks, task)
	}

	// Save to both sources
	if dbErr == nil {
		if err := saveTodo2TasksToDB(projectRoot, mergedTasks); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Database save had errors: %v\n", err)
		}
	}
	if err := saveTodo2TasksToJSON(projectRoot, mergedTasks); err != nil {
		return err
	}

	return nil
}

// GetTaskByID returns a task by ID via TaskStore (database or JSON fallback).
// Caller must not mutate the task if storage is shared; for updates use database.UpdateTask.
func GetTaskByID(ctx context.Context, projectRoot string, id string) (*Todo2Task, error) {
	if id == "" {
		return nil, fmt.Errorf("task id is required")
	}

	store := NewDefaultTaskStore(projectRoot)

	return store.GetTask(ctx, id)
}

// normalizeStatus normalizes status to Title Case.
// This is a wrapper around NormalizeStatusToTitleCase for backward compatibility.
func normalizeStatus(status string) string {
	return NormalizeStatusToTitleCase(status)
}

// IsPendingStatus checks if a status is pending (only "Todo", not "In Progress" or "Review").
// Note: This matches Python implementation where only "todo" is considered pending.
// For active tasks (todo, in_progress, review, blocked), use IsActiveStatusNormalized.
func IsPendingStatus(status string) bool {
	normalized := NormalizeStatus(status)
	return normalized == "todo"
}

// IsCompletedStatus checks if a status is completed.
func IsCompletedStatus(status string) bool {
	normalized := NormalizeStatus(status)
	return normalized == "completed" || normalized == "cancelled"
}

// cleanupAutoTasksFromDB removes all AUTO-* tasks from the database
// AUTO-* tasks are automated/system tasks that should only exist in JSON.
func cleanupAutoTasksFromDB() error {
	if db, err := database.GetDB(); err != nil || db == nil {
		return fmt.Errorf("database not available")
	}

	ctx := context.Background()

	// Get all AUTO-* tasks from database
	allTasks, err := database.ListTasks(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	// Delete each AUTO-* task
	deletedCount := 0

	for _, task := range allTasks {
		if strings.HasPrefix(task.ID, "AUTO-") {
			if err := database.DeleteTask(ctx, task.ID); err != nil {
				// Log but continue - don't fail on individual deletions
				fmt.Fprintf(os.Stderr, "Warning: Failed to delete AUTO task %s: %v\n", task.ID, err)
			} else {
				deletedCount++
			}
		}
	}

	if deletedCount > 0 {
		fmt.Fprintf(os.Stderr, "Cleaned up %d AUTO-* tasks from database\n", deletedCount)
	}

	return nil
}


// formatTaskDate returns a display string for a task date; never returns 1970.
// Empty or epoch dates return "—".
func formatTaskDate(s string) string {
	if s == "" || models.IsEpochDate(s) {
		return "—"
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return "—"
	}

	return t.Format("01/02/2006, 03:04 PM")
}

// commentCounts holds per-type comment counts for a task.
type commentCounts struct {
	Research, Result, Note, Manual int
}

// getCommentCounts returns comment counts from the database (DB only; JSON fallback has no comments).
func getCommentCounts(ctx context.Context, taskID string) commentCounts {
	comments, err := database.GetComments(ctx, taskID)
	if err != nil {
		return commentCounts{}
	}

	var c commentCounts

	for _, cmt := range comments {
		switch cmt.Type {
		case database.CommentTypeResearch:
			c.Research++
		case database.CommentTypeResult:
			c.Result++
		case database.CommentTypeNote:
			c.Note++
		case database.CommentTypeManual:
			c.Manual++
		}
	}

	return c
}

// getKeyInsight returns a truncated key insight from the most recent result or note comment.
func getKeyInsight(ctx context.Context, taskID string, maxLen int) string {
	comments, err := database.GetComments(ctx, taskID)
	if err != nil || len(comments) == 0 {
		return ""
	}

	for i := len(comments) - 1; i >= 0; i-- {
		if comments[i].Type == database.CommentTypeResult || comments[i].Type == database.CommentTypeNote {
			s := strings.TrimSpace(comments[i].Content)
			if len(s) > maxLen {
				s = s[:maxLen-3] + "..."
			}

			return s
		}
	}

	return ""
}

// GetSuggestedNextTasks returns dependency-ordered tasks ready to start (deps done), up to limit.
// Used by todo2-overview and stdio://suggested-tasks resource. Uses TaskStore.
func GetSuggestedNextTasks(projectRoot string, limit int) []BacklogTaskDetail {
	ctx := context.Background()
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil || limit <= 0 {
		return nil
	}

	tasks := tasksFromPtrs(list)

	orderedIDs, _, details, orderErr := BacklogExecutionOrder(tasks, nil)
	if orderErr != nil || len(orderedIDs) == 0 {
		return nil
	}

	ready := tasksReadyToStart(tasks)

	detailMap := make(map[string]BacklogTaskDetail)
	for _, d := range details {
		detailMap[d.ID] = d
	}

	out := make([]BacklogTaskDetail, 0, limit)

	for _, id := range orderedIDs {
		if ready[id] {
			if d, ok := detailMap[id]; ok {
				out = append(out, d)
				if len(out) >= limit {
					break
				}
			}
		}
	}

	return out
}

// tasksReadyToStart returns task IDs whose dependencies are all Done.
func tasksReadyToStart(tasks []Todo2Task) map[string]bool {
	done := make(map[string]bool)

	for _, t := range tasks {
		// Prefer typed status when available; fall back to string for legacy callers.
		if t.StatusEnum == models.TaskStatusDone || strings.EqualFold(t.Status, "done") {
			done[t.ID] = true
		}
	}

	ready := make(map[string]bool)

	for _, t := range tasks {
		if !IsBacklogStatus(t.Status) {
			continue
		}

		allDone := true

		for _, dep := range t.Dependencies {
			if !done[dep] {
				allDone = false
				break
			}
		}

		if allDone {
			ready[t.ID] = true
		}
	}

	return ready
}

// WriteTodo2Overview writes .cursor/rules/todo2-overview.mdc from current tasks.
// Uses TaskStore. Uses real dates or "—" for unknown; never displays 1970.
func WriteTodo2Overview(projectRoot string) error {
	ctx := context.Background()
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return fmt.Errorf("load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Sort by last_modified desc (newest first), then take last 20 for "newest first" display
	sort.Slice(tasks, func(i, j int) bool {
		a, b := tasks[i].LastModified, tasks[j].LastModified
		if a == "" {
			a = tasks[i].CreatedAt
		}

		if b == "" {
			b = tasks[j].CreatedAt
		}

		return a > b
	})

	displayCount := 20
	if len(tasks) < displayCount {
		displayCount = len(tasks)
	}

	displayTasks := tasks
	if len(tasks) > displayCount {
		displayTasks = tasks[:displayCount]
	}

	suggestedNext := GetSuggestedNextTasks(projectRoot, 5)

	now := time.Now().Format("01/02/2006, 03:04 PM")

	var b strings.Builder

	b.WriteString("---\n")
	b.WriteString("description: Todo2 task overview for Cursor AI awareness - provides real-time context of current project tasks, priorities, and progress\n")
	b.WriteString("alwaysApply: true\n")
	b.WriteString("---\n\n")
	b.WriteString("# Todo2 Project Context\n\n")
	b.WriteString("*Last updated: " + now + "*\n")
	b.WriteString("*Generated automatically from .todo2/state.todo2.json*\n\n")

	if len(suggestedNext) > 0 {
		b.WriteString("## Suggested Next Tasks (dependency-ready)\n\n")

		for _, d := range suggestedNext {
			b.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", d.ID, d.Priority, d.Content))
		}

		b.WriteString("\n")
	}

	b.WriteString("## Current Task Overview (Last 20 Tasks - Newest First)\n\n")

	for _, t := range displayTasks {
		name := t.Content
		if name == "" {
			name = "undefined"
		}

		b.WriteString("### " + t.ID + ": " + name + "\n")
		b.WriteString("- **Status:** " + t.Status + " ")

		switch t.StatusEnum {
		case models.TaskStatusDone:
			b.WriteString("✅")
		case models.TaskStatusInProgress:
			b.WriteString("⚡")
		case models.TaskStatusReview:
			b.WriteString("👀")
		default:
			b.WriteString("📋")
		}

		b.WriteString(" | **Priority:** " + t.Priority + " ")
		// Priority emoji
		switch strings.ToLower(t.Priority) {
		case "high", "critical":
			b.WriteString("🟠")
		case "medium":
			b.WriteString("🟡")
		default:
			b.WriteString("🟢")
		}

		b.WriteString(" | **Created:** " + formatTaskDate(t.CreatedAt) + " | **Updated:** " + formatTaskDate(t.LastModified) + "\n")
		b.WriteString("- **Tags:** " + strings.Join(t.Tags, ", ") + "\n")
		b.WriteString("- **Dependencies:** " + strings.Join(t.Dependencies, ", ") + "\n")
		cc := getCommentCounts(ctx, t.ID)
		b.WriteString(fmt.Sprintf("- **Comments**: %d research_with_links, %d result, %d notes, %d manualsetup\n",
			cc.Research, cc.Result, cc.Note, cc.Manual))
		b.WriteString("- **Status:** " + t.Status + "\n")

		insight := getKeyInsight(ctx, t.ID, 80)
		if insight == "" {
			insight = "*[No key insight available]*"
		} else {
			insight = "*" + insight + "*"
		}

		b.WriteString("- **Key Insight:** " + insight + "\n\n")
	}

	// Task statistics
	var todo, inProgress, done int

	for _, t := range tasks {
		switch t.StatusEnum {
		case models.TaskStatusTodo:
			todo++
		case models.TaskStatusInProgress:
			inProgress++
		case models.TaskStatusDone:
			done++
		}
	}

	high := 0

	for _, t := range tasks {
		if models.IsHighPriority(t.Priority) {
			high++
		}
	}

	b.WriteString("## Task Statistics\n")
	b.WriteString(fmt.Sprintf("- **Total Tasks:** %d\n", len(tasks)))
	b.WriteString(fmt.Sprintf("- **In Progress:** %d tasks\n", inProgress))
	b.WriteString(fmt.Sprintf("- **Todo:** %d tasks \n", todo))
	b.WriteString(fmt.Sprintf("- **Done:** %d tasks\n", done))

	critical := 0

	for _, t := range tasks {
		if models.IsCritical(t.Priority) {
			critical++
		}
	}

	b.WriteString(fmt.Sprintf("- **High Priority:** %d tasks\n", high))
	b.WriteString(fmt.Sprintf("- **Critical Priority:** %d tasks\n", critical))
	b.WriteString(fmt.Sprintf("- **Tasks with Dependencies:** %d tasks\n\n", countWithDeps(tasks)))

	b.WriteString("## Recent Activity\n")

	for i, t := range displayTasks {
		if i >= 10 {
			break
		}

		b.WriteString(fmt.Sprintf("- %s: %s (%s)\n", t.ID, t.Status, formatTaskDate(t.LastModified)))
	}

	b.WriteString("\n## Key Project Context\n")
	b.WriteString("This is an implementation of an automated cursor rules system that will maintain real-time awareness of Todo2 task status for enhanced AI assistance. The system monitors task changes and automatically updates this overview file to provide contextual information to Cursor chat.\n\n")
	b.WriteString("*This file is automatically maintained by Todo2. Last generation: " + time.Now().Format("01/02/2006, 15:04:05") + "*\n")

	outPath := filepath.Join(projectRoot, ".cursor", "rules", "todo2-overview.mdc")
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("create .cursor/rules: %w", err)
	}

	if err := os.WriteFile(outPath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("write overview: %w", err)
	}

	return nil
}

func countWithDeps(tasks []Todo2Task) int {
	n := 0

	for _, t := range tasks {
		if len(t.Dependencies) > 0 {
			n++
		}
	}

	return n
}
