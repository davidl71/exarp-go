package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// todo2TaskJSON is used when parsing JSON so metadata can be sanitized.
// Metadata is raw to handle invalid JSON (string, malformed) without failing unmarshal.
// Supports both "created"/"updated" and "created_at"/"last_modified" for compatibility.
// Supports "name"/"title" (aliases for content) and "description" (alias for long_description) for Todo2 extension compatibility.
type todo2TaskJSON struct {
	ID              string          `json:"id"`
	Name            string          `json:"name,omitempty"`
	Title           string          `json:"title,omitempty"`
	Content         string          `json:"content"`
	Description     string          `json:"description,omitempty"`
	LongDescription string          `json:"long_description,omitempty"`
	Status          string          `json:"status"`
	Priority        string          `json:"priority,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	Dependencies    []string        `json:"dependencies,omitempty"`
	Completed       bool            `json:"completed,omitempty"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
	Created         string          `json:"created,omitempty"`
	CreatedAt       string          `json:"created_at,omitempty"`
	Updated         string          `json:"updated,omitempty"`
	LastModified    string          `json:"last_modified,omitempty"`
	CompletedAt     string          `json:"completed_at,omitempty"`
}

func convertTodo2TaskJSONToTask(raw todo2TaskJSON) models.Todo2Task {
	createdAt := raw.CreatedAt
	if createdAt == "" {
		createdAt = raw.Created
	}
	lastMod := raw.LastModified
	if lastMod == "" {
		lastMod = raw.Updated
	}
	content := raw.Content
	if content == "" && raw.Name != "" {
		content = raw.Name
	}
	if content == "" && raw.Title != "" {
		content = raw.Title
	}
	longDesc := raw.LongDescription
	if longDesc == "" && raw.Description != "" {
		longDesc = raw.Description
	}
	t := models.Todo2Task{
		ID:              raw.ID,
		Content:         content,
		LongDescription: longDesc,
		Status:          raw.Status,
		Priority:        raw.Priority,
		Tags:            raw.Tags,
		Dependencies:    raw.Dependencies,
		Completed:       raw.Completed,
		CreatedAt:       createdAt,
		LastModified:    lastMod,
		CompletedAt:     raw.CompletedAt,
	}
	if len(raw.Metadata) > 0 {
		t.Metadata = database.DeserializeTaskMetadata(string(raw.Metadata), nil, "")
	}
	t.NormalizeEpochDates()
	return t
}

// todo2TaskWrite is used when writing state JSON. It includes "name" (same as content),
// "description" (same as long_description), and "created"/"updated" (same as created_at/last_modified)
// so the Todo2 extension and overview never show Invalid Date or 1970.
type todo2TaskWrite struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"` // Always write for Todo2 extension display; fallback from content
	Content         string                 `json:"content"`
	Description     string                 `json:"description,omitempty"`
	LongDescription string                 `json:"long_description,omitempty"`
	Status          string                 `json:"status"`
	Priority        string                 `json:"priority,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	Completed       bool                   `json:"completed,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Created         string                 `json:"created,omitempty"`
	CreatedAt       string                 `json:"created_at,omitempty"`
	Updated         string                 `json:"updated,omitempty"`
	LastModified    string                 `json:"last_modified,omitempty"`
	CompletedAt     string                 `json:"completed_at,omitempty"`
}

// MarshalTasksToStateJSON marshals tasks to state.todo2.json format with "name", "description",
// and "created"/"updated" for Todo2 extension compatibility (no Invalid Date or 1970).
// Never writes epoch dates; empty dates get a fallback (last_modified or now) so the extension always has a parseable date.
func MarshalTasksToStateJSON(tasks []models.Todo2Task) ([]byte, error) {
	nowFallback := time.Now().UTC().Format(time.RFC3339)
	writes := make([]todo2TaskWrite, len(tasks))
	for i := range tasks {
		t := &tasks[i]
		createdAt := t.CreatedAt
		if models.IsEpochDate(createdAt) {
			createdAt = ""
		}
		lastMod := t.LastModified
		if models.IsEpochDate(lastMod) {
			lastMod = ""
		}
		completedAt := t.CompletedAt
		if models.IsEpochDate(completedAt) {
			completedAt = ""
		}
		// Fallback so extension never sees missing/invalid date (avoids Invalid Date / 1970)
		if createdAt == "" {
			createdAt = lastMod
		}
		if createdAt == "" {
			createdAt = nowFallback
		}
		if lastMod == "" {
			lastMod = createdAt
		}
		if lastMod == "" {
			lastMod = nowFallback
		}
		writes[i] = todo2TaskWrite{
			ID:              t.ID,
			Name:            t.Content,
			Content:         t.Content,
			Description:     t.LongDescription,
			LongDescription: t.LongDescription,
			Status:          t.Status,
			Priority:        t.Priority,
			Tags:            t.Tags,
			Dependencies:    t.Dependencies,
			Completed:       t.Completed,
			Metadata:        database.SanitizeMetadataForWrite(t.Metadata),
			Created:         createdAt,
			CreatedAt:       createdAt,
			Updated:         lastMod,
			LastModified:    lastMod,
			CompletedAt:     completedAt,
		}
	}
	state := struct {
		Todos []todo2TaskWrite `json:"todos"`
	}{Todos: writes}
	return json.MarshalIndent(state, "", "  ")
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
