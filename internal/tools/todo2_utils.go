package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// Todo2Task is an alias for models.Todo2Task (for backward compatibility)
type Todo2Task = models.Todo2Task

// Todo2State is an alias for models.Todo2State (for backward compatibility)
type Todo2State = models.Todo2State

// LoadTodo2Tasks loads tasks from database (preferred) or .todo2/state.todo2.json (fallback)
func LoadTodo2Tasks(projectRoot string) ([]Todo2Task, error) {
	// Try database first
	if tasks, err := loadTodo2TasksFromDB(); err == nil {
		return tasks, nil
	}

	// Database not available or query failed, fallback to JSON
	return loadTodo2TasksFromJSON(projectRoot)
}

// loadTodo2TasksFromJSON loads tasks from JSON file (fallback method).
// Metadata is sanitized on load; invalid JSON is coerced to {"raw": "..."}.
func loadTodo2TasksFromJSON(projectRoot string) ([]Todo2Task, error) {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")

	// Use file cache for frequently accessed todo2.json file
	fileCache := cache.GetGlobalFileCache()
	data, _, err := fileCache.ReadFile(todo2Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo2Task{}, nil
		}
		return nil, fmt.Errorf("failed to read Todo2 file: %w", err)
	}

	return ParseTasksFromJSON(data)
}

// SaveTodo2Tasks saves tasks to database (preferred) or .todo2/state.todo2.json (fallback).
// When database save succeeds, also writes the same list to JSON so both stores stay in sync
// (avoids merge/sync reintroducing removed tasks from stale JSON).
func SaveTodo2Tasks(projectRoot string, tasks []Todo2Task) error {
	// Try database first
	if err := saveTodo2TasksToDB(tasks); err == nil {
		// Keep JSON in sync so a later sync does not reintroduce removed tasks (e.g. after merge)
		if jsonErr := saveTodo2TasksToJSON(projectRoot, tasks); jsonErr != nil {
			return fmt.Errorf("database saved but JSON write failed: %w", jsonErr)
		}
		return nil
	}

	// Database not available or save failed, fallback to JSON
	return saveTodo2TasksToJSON(projectRoot, tasks)
}

// saveTodo2TasksToJSON saves tasks to JSON file (fallback method)
func saveTodo2TasksToJSON(projectRoot string, tasks []Todo2Task) error {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")

	// Ensure directory exists
	dir := filepath.Dir(todo2Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create .todo2 directory: %w", err)
	}

	// Normalize epoch dates so we never persist 1970-01-01
	for i := range tasks {
		tasks[i].NormalizeEpochDates()
	}
	// Use MarshalTasksToStateJSON so written JSON includes "name" and "description" for Todo2 extension/overview
	data, err := MarshalTasksToStateJSON(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal Todo2 state: %w", err)
	}

	if err := os.WriteFile(todo2Path, data, 0644); err != nil {
		return fmt.Errorf("failed to write Todo2 file: %w", err)
	}

	return nil
}

// FindProjectRoot finds the project root by looking for .todo2 directory
// It first checks the PROJECT_ROOT environment variable (set by Cursor IDE from {{PROJECT_ROOT}}),
// then searches up from the current working directory for a .todo2 directory.
func FindProjectRoot() (string, error) {
	// Check PROJECT_ROOT environment variable first (highest priority)
	// This is set by Cursor IDE when using {{PROJECT_ROOT}} in mcp.json
	if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
		// Skip if placeholder wasn't substituted (contains {{PROJECT_ROOT}})
		if !strings.Contains(envRoot, "{{PROJECT_ROOT}}") {
			// Validate that the path exists and contains .todo2
			absPath, err := filepath.Abs(envRoot)
			if err == nil {
				todo2Path := filepath.Join(absPath, ".todo2")
				if _, err := os.Stat(todo2Path); err == nil {
					return absPath, nil
				}
				// If PROJECT_ROOT is set but no .todo2, still use it (might be valid project)
				// This allows working with projects that don't have .todo2 yet
				return absPath, nil
			}
		}
	}

	// Fallback: search up from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		todo2Path := filepath.Join(dir, ".todo2")
		if _, err := os.Stat(todo2Path); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found (no .todo2 directory)")
}

// SyncTodo2Tasks synchronizes tasks between database and JSON file
// It loads from both sources, merges them (database takes precedence for conflicts),
// and saves to both to ensure consistency
func SyncTodo2Tasks(projectRoot string) error {
	// Load from both sources
	dbTasksLoaded, dbErr := loadTodo2TasksFromDB()
	jsonTasksLoaded, _ := loadTodo2TasksFromJSON(projectRoot)

	// Build merged task map (database takes precedence)
	taskMap := make(map[string]Todo2Task)

	// First, add JSON tasks
	for _, task := range jsonTasksLoaded {
		taskMap[task.ID] = task
	}

	// Then, override with database tasks (database takes precedence)
	for _, task := range dbTasksLoaded {
		taskMap[task.ID] = task
	}

	// Convert map back to slice
	mergedTasks := make([]Todo2Task, 0, len(taskMap))
	for _, task := range taskMap {
		mergedTasks = append(mergedTasks, task)
	}

	// Filter out AUTO-* tasks for database (keep them in JSON only)
	// AUTO-* tasks are automated/system tasks that don't need to be in database
	dbTasksToSave := make([]Todo2Task, 0, len(mergedTasks))
	jsonTasksForSave := make([]Todo2Task, 0, len(mergedTasks))
	for _, task := range mergedTasks {
		if strings.HasPrefix(task.ID, "AUTO-") {
			// Skip AUTO-* tasks for database, but keep in JSON
			jsonTasksForSave = append(jsonTasksForSave, task)
		} else {
			// Regular tasks go to both
			dbTasksToSave = append(dbTasksToSave, task)
			jsonTasksForSave = append(jsonTasksForSave, task)
		}
	}

	// Save to both sources
	var dbSaveErr, jsonSaveErr error

	// Try to save to database first (without AUTO-* tasks)
	if dbErr == nil {
		// Database is available, save to it (excluding AUTO-* tasks)
		// Also clean up any existing AUTO-* tasks from database
		if err := cleanupAutoTasksFromDB(); err != nil {
			// Log but don't fail - cleanup is best effort
			fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup AUTO tasks from database: %v\n", err)
		}
		dbSaveErr = saveTodo2TasksToDB(dbTasksToSave)
		if dbSaveErr != nil {
			// Database save had errors - log but continue
			// The error message includes details about which tasks failed
			// We still want to save to JSON as backup
			// Log the error for debugging
			fmt.Fprintf(os.Stderr, "WARNING: Database save had errors: %v\n", dbSaveErr)
		}
	} else {
		// Database not available, skip
		dbSaveErr = fmt.Errorf("database not available: %w", dbErr)
	}

	// Always save to JSON (as fallback, including AUTO-* tasks)
	jsonSaveErr = saveTodo2TasksToJSON(projectRoot, jsonTasksForSave)

	// Return error if both failed
	if dbSaveErr != nil && jsonSaveErr != nil {
		return fmt.Errorf("failed to save to both sources: database=%v, json=%v", dbSaveErr, jsonSaveErr)
	}

	// If database save had errors, return the error (don't silently ignore)
	// This ensures we know about sync issues even if JSON save succeeded
	if dbSaveErr != nil {
		// Database save failed - return error so caller knows
		// JSON save succeeded, so we have a backup, but we should report the issue
		return fmt.Errorf("database save failed (JSON saved as backup): %w", dbSaveErr)
	}

	return nil
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
// AUTO-* tasks are automated/system tasks that should only exist in JSON
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
// Used by todo2-overview and stdio://suggested-tasks resource.
func GetSuggestedNextTasks(projectRoot string, limit int) []BacklogTaskDetail {
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil || limit <= 0 {
		return nil
	}
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
		if strings.EqualFold(t.Status, "done") {
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
// Uses real dates or "—" for unknown; never displays 1970.
func WriteTodo2Overview(projectRoot string) error {
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return fmt.Errorf("load tasks: %w", err)
	}
	ctx := context.Background()

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
		switch strings.ToLower(t.Status) {
		case "done":
			b.WriteString("✅")
		case "in progress":
			b.WriteString("⚡")
		case "review":
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
		switch strings.ToLower(t.Status) {
		case "todo":
			todo++
		case "in progress":
			inProgress++
		case "done":
			done++
		}
	}
	high := 0
	for _, t := range tasks {
		if strings.ToLower(t.Priority) == "high" || strings.ToLower(t.Priority) == "critical" {
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
		if strings.ToLower(t.Priority) == "critical" {
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
