package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Todo2Task represents a Todo2 task
type Todo2Task struct {
	ID             string                 `json:"id"`
	Content        string                 `json:"content"`
	LongDescription string                `json:"long_description,omitempty"`
	Status         string                 `json:"status"`
	Priority       string                 `json:"priority,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	Dependencies   []string               `json:"dependencies,omitempty"`
	Completed      bool                   `json:"completed,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Todo2State represents the Todo2 state file structure
type Todo2State struct {
	Todos []Todo2Task `json:"todos"`
}

// LoadTodo2Tasks loads tasks from .todo2/state.todo2.json
func LoadTodo2Tasks(projectRoot string) ([]Todo2Task, error) {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")
	
	data, err := os.ReadFile(todo2Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Todo2Task{}, nil
		}
		return nil, fmt.Errorf("failed to read Todo2 file: %w", err)
	}

	var state Todo2State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Todo2 JSON: %w", err)
	}

	return state.Todos, nil
}

// SaveTodo2Tasks saves tasks to .todo2/state.todo2.json
func SaveTodo2Tasks(projectRoot string, tasks []Todo2Task) error {
	todo2Path := filepath.Join(projectRoot, ".todo2", "state.todo2.json")
	
	// Ensure directory exists
	dir := filepath.Dir(todo2Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create .todo2 directory: %w", err)
	}

	state := Todo2State{Todos: tasks}
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
func FindProjectRoot() (string, error) {
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

