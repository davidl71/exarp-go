package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// loadTodo2TasksFromJSON loads tasks from JSON file (fallback method)
func loadTodo2TasksFromJSON(projectRoot string) ([]Todo2Task, error) {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")

	data, err := os.ReadFile(todo2Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo2Task{}, nil
		}
		return nil, fmt.Errorf("failed to read Todo2 file: %w", err)
	}

	var state models.Todo2State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Todo2 JSON: %w", err)
	}

	return state.Todos, nil
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

// IsPendingStatus checks if a status is pending
func IsPendingStatus(status string) bool {
	status = normalizeStatus(status)
	return status == "Todo" || status == "In Progress" || status == "Review"
}

// IsCompletedStatus checks if a status is completed
func IsCompletedStatus(status string) bool {
	status = normalizeStatus(status)
	return status == "Done" || status == "Cancelled"
}

// normalizeStatus normalizes status to Title Case
func normalizeStatus(status string) string {
	if status == "" {
		return "Todo"
	}
	// Simple normalization - could be enhanced
	switch status {
	case "todo", "TODO", "pending":
		return "Todo"
	case "in_progress", "in-progress", "working":
		return "In Progress"
	case "review", "Review":
		return "Review"
	case "done", "DONE", "completed":
		return "Done"
	case "blocked", "Blocked":
		return "Blocked"
	case "cancelled", "canceled", "Cancelled":
		return "Cancelled"
	default:
		return status
	}
}
