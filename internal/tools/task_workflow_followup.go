// task_workflow_followup.go — AI-suggested follow-up task creation.
// Automatically suggests and optionally creates follow-up tasks when a task is completed.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

// SuggestFollowUps uses LLM to suggest follow-up tasks for a completed task.
// Uses FM chain (Apple FM -> fallback) for suggestions.
func SuggestFollowUps(ctx context.Context, task *models.Todo2Task) ([]map[string]string, error) {
	if task == nil {
		return nil, fmt.Errorf("task is nil")
	}

	// Build context about the completed task
	taskInfo := fmt.Sprintf("Task: %s\n", task.Content)
	if task.LongDescription != "" {
		taskInfo += fmt.Sprintf("Description: %s\n", task.LongDescription)
	}
	if len(task.Tags) > 0 {
		taskInfo += fmt.Sprintf("Tags: %s\n", strings.Join(task.Tags, ", "))
	}
	if task.Priority != "" {
		taskInfo += fmt.Sprintf("Priority: %s\n", task.Priority)
	}

	prompt := fmt.Sprintf(`Based on the completed task below, suggest 1-3 potential follow-up tasks that would naturally come next.

%s

For each follow-up, provide:
1. A concise task name (short, action-oriented)
2. A brief description (1-2 sentences explaining why this follow-up is needed)

Respond in JSON format:
[
  {"name": "follow-up task name", "description": "brief description"},
  {"name": "another follow-up", "description": "description"}
]

Only suggest realistic follow-ups that would help complete the broader goal.`, taskInfo)

	// Try FM provider (Apple FM or fallback chain)
	var response string
	var err error

	if FMAvailable() {
		response, err = DefaultFMProvider().Generate(ctx, prompt, 500, 0.3)
	}

	if err != nil || response == "" {
		return nil, fmt.Errorf("no LLM available for follow-up suggestions")
	}

	// Parse JSON response
	var suggestions []map[string]string

	// Try to extract JSON from response
	jsonStr := extractJSONFromFollowup(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no valid JSON found in LLM response")
	}

	if err := json.Unmarshal([]byte(jsonStr), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse suggestions: %w", err)
	}

	return suggestions, nil
}

// extractJSONFromFollowup extracts JSON array from LLM response text.
func extractJSONFromFollowup(text string) string {
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start == -1 || end == -1 || end <= start {
		return ""
	}
	return text[start : end+1]
}

// CreateFollowUpTasks creates follow-up tasks from suggestions.
// Returns IDs of created tasks.
func CreateFollowUpTasks(ctx context.Context, suggestions []map[string]string, parentID string) ([]string, error) {
	if len(suggestions) == 0 {
		return nil, nil
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	var createdIDs []string

	for _, suggestion := range suggestions {
		name, hasName := suggestion["name"]
		if !hasName || name == "" {
			continue
		}

		description, _ := suggestion["description"]

		newTask := &models.Todo2Task{
			Content:         name,
			LongDescription: description,
			Status:          models.StatusTodo,
			Priority:        "medium",
		}

		if parentID != "" {
			newTask.ParentID = parentID
		}

		if err := store.CreateTask(ctx, newTask); err != nil {
			continue
		}

		createdIDs = append(createdIDs, newTask.ID)
	}

	return createdIDs, nil
}
