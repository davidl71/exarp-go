package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// SaveTodo2Tasks saves tasks to database (preferred) or .todo2/state.todo2.json (fallback)
func SaveTodo2Tasks(projectRoot string, tasks []Todo2Task) error {
	// Try database first
	if err := saveTodo2TasksToDB(tasks); err == nil {
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

	state := models.Todo2State{Todos: tasks}
	data, err := json.MarshalIndent(state, "", "  ")
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
