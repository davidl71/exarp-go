package tools

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/models"
)

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
func LoadJSONStateFromContent(data []byte) ([]models.Todo2Task, []database.Comment, error) {
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Load tasks
	var tasks []models.Todo2Task
	if todos, ok := state["todos"].([]interface{}); ok {
		for _, todo := range todos {
			todoBytes, err := json.Marshal(todo)
			if err != nil {
				continue
			}

			var task models.Todo2Task
			if err := json.Unmarshal(todoBytes, &task); err != nil {
				continue
			}

			tasks = append(tasks, task)
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
