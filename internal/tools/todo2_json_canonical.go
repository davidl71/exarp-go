// todo2_json_canonical.go — Canonical (non-compat) JSON helpers for Todo2 state.
//
// These helpers intentionally do NOT support legacy alias fields like "title"/"description"
// or "created"/"updated". They exist only for:
// - reading old `state.todo2.json` during one-off migrations (cmd/migrate)
// - decoding point-in-time snapshots stored as JSON
//
// SQLite (`.todo2/todo2.db`) is the canonical store; JSON is not used for ongoing sync.
package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

// MarshalTasksToStateJSON marshals tasks to the canonical state.todo2.json shape:
// `{ "todos": [ ... ] }`.
func MarshalTasksToStateJSON(tasks []models.Todo2Task) ([]byte, error) {
	state := struct {
		Todos []models.Todo2Task `json:"todos"`
	}{Todos: tasks}

	return json.MarshalIndent(state, "", "  ")
}

// ParseTasksFromJSON parses the canonical state.todo2.json shape:
// `{ "todos": [ ... ] }`.
func ParseTasksFromJSON(data []byte) ([]models.Todo2Task, error) {
	var state struct {
		Todos []models.Todo2Task `json:"todos"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse Todo2 JSON: %w", err)
	}
	return state.Todos, nil
}

// LoadJSONStateFromFile loads tasks and comments from a JSON file.
// Comments are best-effort (empty slice if missing/unparseable).
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

// LoadJSONStateFromContent loads tasks and comments from JSON byte content.
// It only supports canonical field names; unknown fields are ignored by json.Unmarshal.
func LoadJSONStateFromContent(data []byte) ([]models.Todo2Task, []database.Comment, error) {
	var raw struct {
		Todos    []models.Todo2Task     `json:"todos"`
		Comments []map[string]any      `json:"comments"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	comments := make([]database.Comment, 0, len(raw.Comments))
	for _, c := range raw.Comments {
		taskID, _ := c["todo_id"].(string)
		if taskID == "" {
			taskID, _ = c["task_id"].(string)
		}
		if taskID == "" {
			continue
		}

		typ, _ := c["type"].(string)
		content, _ := c["content"].(string)
		if typ == "" || content == "" {
			continue
		}

		comments = append(comments, database.Comment{
			TaskID:  taskID,
			Type:    typ,
			Content: content,
		})
	}

	return raw.Todos, comments, nil
}

