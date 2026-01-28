package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// todo2TaskJSON is used when parsing JSON so metadata can be sanitized.
// Metadata is raw to handle invalid JSON (string, malformed) without failing unmarshal.
type todo2TaskJSON struct {
	ID              string          `json:"id"`
	Content         string          `json:"content"`
	LongDescription string          `json:"long_description,omitempty"`
	Status          string          `json:"status"`
	Priority        string          `json:"priority,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	Dependencies    []string        `json:"dependencies,omitempty"`
	Completed       bool            `json:"completed,omitempty"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
}

func convertTodo2TaskJSONToTask(raw todo2TaskJSON) models.Todo2Task {
	t := models.Todo2Task{
		ID:              raw.ID,
		Content:         raw.Content,
		LongDescription: raw.LongDescription,
		Status:          raw.Status,
		Priority:        raw.Priority,
		Tags:            raw.Tags,
		Dependencies:    raw.Dependencies,
		Completed:       raw.Completed,
	}
	if len(raw.Metadata) > 0 {
		t.Metadata = database.SanitizeTaskMetadata(string(raw.Metadata))
	}
	return t
}

// ParseTasksFromJSON parses "todos" from state JSON and returns tasks with sanitized metadata.
// Invalid metadata is coerced to {"raw": "..."} so callers never see parse errors.
func ParseTasksFromJSON(data []byte) ([]models.Todo2Task, error) {
	var state struct {
		Todos []todo2TaskJSON `json:"todos"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Todo2 JSON: %w", err)
	}
	out := make([]models.Todo2Task, 0, len(state.Todos))
	for _, raw := range state.Todos {
		out = append(out, convertTodo2TaskJSONToTask(raw))
	}
	return out, nil
}

// LoadJSONStateFromFile loads tasks and comments from a JSON file
// Supports both comment formats:
// 1. Top-level "comments" array with "todoId" or "todo_id" field
// 2. Comments nested inside each task object
func LoadJSONStateFromFile(jsonPath string) ([]models.Todo2Task, []database.Comment, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Todo2Task{}, []database.Comment{}, nil
		}
		return nil, nil, fmt.Errorf("failed to read JSON file: %w", err)
	}
	return LoadJSONStateFromContent(data)
}

// LoadJSONStateFromContent loads tasks and comments from JSON byte content
// Supports both comment formats:
// 1. Top-level "comments" array with "todoId" or "todo_id" field
// 2. Comments nested inside each task object
// Task metadata is sanitized on load; invalid JSON is coerced to {"raw": "..."}.
func LoadJSONStateFromContent(data []byte) ([]models.Todo2Task, []database.Comment, error) {
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Load tasks with metadata sanitization (invalid JSON -> {"raw": "..."})
	var tasks []models.Todo2Task
	if todos, ok := state["todos"].([]interface{}); ok {
		for _, todo := range todos {
			todoBytes, err := json.Marshal(todo)
			if err != nil {
				continue
			}
			var raw todo2TaskJSON
			if err := json.Unmarshal(todoBytes, &raw); err != nil {
				continue
			}
			tasks = append(tasks, convertTodo2TaskJSONToTask(raw))
		}
	}

	// Load comments - check both formats
	var comments []database.Comment

	// Format 1: Top-level comments array
	if commentsArray, ok := state["comments"].([]interface{}); ok {
		for _, comment := range commentsArray {
			commentBytes, err := json.Marshal(comment)
			if err != nil {
				continue
			}

			var commentMap map[string]interface{}
			if err := json.Unmarshal(commentBytes, &commentMap); err != nil {
				continue
			}

			todoID, _ := commentMap["todoId"].(string)
			if todoID == "" {
				todoID, _ = commentMap["todo_id"].(string) // Try alternate field name
			}

			if todoID == "" {
				continue
			}

			commentType, _ := commentMap["type"].(string)
			content, _ := commentMap["content"].(string)

			comments = append(comments, database.Comment{
				TaskID:  todoID,
				Type:    commentType,
				Content: content,
			})
		}
	}

	// Format 2: Comments nested in tasks
	// (Note: tasks already loaded above, but we need to check for nested comments separately)
	if todos, ok := state["todos"].([]interface{}); ok {
		for _, todo := range todos {
			todoMap, ok := todo.(map[string]interface{})
			if !ok {
				continue
			}

			taskID, _ := todoMap["id"].(string)
			if taskID == "" {
				continue
			}

			if commentsArray, ok := todoMap["comments"].([]interface{}); ok {
				for _, comment := range commentsArray {
					commentMap, ok := comment.(map[string]interface{})
					if !ok {
						continue
					}

					commentType, _ := commentMap["type"].(string)
					content, _ := commentMap["content"].(string)

					comments = append(comments, database.Comment{
						TaskID:  taskID,
						Type:    commentType,
						Content: content,
					})
				}
			}
		}
	}

	return tasks, comments, nil
}
