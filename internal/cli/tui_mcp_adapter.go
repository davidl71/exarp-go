// tui_mcp_adapter.go — Routes TUI data operations through MCP tools instead of direct database access.
// All task CRUD goes through task_workflow; handoffs/analysis already use server.CallTool.
// See docs/TUI_MCP_ADAPTER_DESIGN.md for design rationale.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
)

// mcpListTasksResponse is the JSON envelope from task_workflow sync/list with output_format=json.
type mcpListTasksResponse struct {
	Success bool                `json:"success"`
	Method  string              `json:"method"`
	Tasks   []mcpTaskJSON       `json:"tasks"`
}

// mcpTaskJSON mirrors the JSON fields emitted by task_workflow list.
type mcpTaskJSON struct {
	ID              string                 `json:"id"`
	Content         string                 `json:"content"`
	Status          string                 `json:"status"`
	Priority        string                 `json:"priority,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	LongDescription string                 `json:"long_description,omitempty"`
	Dependencies    []string               `json:"dependencies,omitempty"`
	ParentID        string                 `json:"parent_id,omitempty"`
	LastModified    string                 `json:"last_modified,omitempty"`
	CreatedAt       string                 `json:"created_at,omitempty"`
	CompletedAt     string                 `json:"completed_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

func (j *mcpTaskJSON) toTodo2Task() *database.Todo2Task {
	return &database.Todo2Task{
		ID:              j.ID,
		Content:         j.Content,
		Status:          j.Status,
		Priority:        j.Priority,
		Tags:            j.Tags,
		LongDescription: j.LongDescription,
		Dependencies:    j.Dependencies,
		ParentID:        j.ParentID,
		LastModified:    j.LastModified,
		CreatedAt:       j.CreatedAt,
		CompletedAt:     j.CompletedAt,
		Metadata:        j.Metadata,
	}
}

// callToolText calls a named MCP tool and returns the concatenated text result.
func callToolText(ctx context.Context, server framework.MCPServer, tool string, args map[string]interface{}) (string, error) {
	argsBytes, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("marshal args for %s: %w", tool, err)
	}

	result, err := server.CallTool(ctx, tool, argsBytes)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	for _, c := range result {
		b.WriteString(c.Text)
	}
	return b.String(), nil
}

// listTasksViaMCP loads tasks through task_workflow sync/list.
// When status is empty, loads both Todo and In Progress tasks (matching the old loadTasks behaviour).
func listTasksViaMCP(ctx context.Context, server framework.MCPServer, status string) ([]*database.Todo2Task, error) {
	if status != "" {
		return listTasksByStatusViaMCP(ctx, server, status)
	}

	todo, err := listTasksByStatusViaMCP(ctx, server, "Todo")
	if err != nil {
		return nil, err
	}
	inProgress, err := listTasksByStatusViaMCP(ctx, server, "In Progress")
	if err != nil {
		return nil, err
	}
	return append(todo, inProgress...), nil
}

func listTasksByStatusViaMCP(ctx context.Context, server framework.MCPServer, status string) ([]*database.Todo2Task, error) {
	text, err := callToolText(ctx, server, "task_workflow", map[string]interface{}{
		"action":        "sync",
		"sub_action":    "list",
		"status_filter": status,
		"output_format": "json",
		"compact":       true,
	})
	if err != nil {
		return nil, fmt.Errorf("task_workflow list (status=%s): %w", status, err)
	}

	var resp mcpListTasksResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		return nil, fmt.Errorf("parse task list JSON: %w", err)
	}

	tasks := make([]*database.Todo2Task, len(resp.Tasks))
	for i := range resp.Tasks {
		tasks[i] = resp.Tasks[i].toTodo2Task()
	}
	return tasks, nil
}

// updateTaskStatusViaMCP updates a single task's status through task_workflow update.
func updateTaskStatusViaMCP(ctx context.Context, server framework.MCPServer, taskID, newStatus string) error {
	_, err := callToolText(ctx, server, "task_workflow", map[string]interface{}{
		"action":     "update",
		"task_ids":   taskID,
		"new_status": newStatus,
	})
	return err
}

// bulkUpdateStatusViaMCP updates multiple tasks' status through task_workflow update.
// Returns count of tasks requested (task_workflow update is all-or-nothing per call).
func bulkUpdateStatusViaMCP(ctx context.Context, server framework.MCPServer, taskIDs []string, newStatus string) (int, error) {
	_, err := callToolText(ctx, server, "task_workflow", map[string]interface{}{
		"action":     "update",
		"task_ids":   strings.Join(taskIDs, ","),
		"new_status": newStatus,
	})
	if err != nil {
		return 0, err
	}
	return len(taskIDs), nil
}

// moveTaskToWaveViaMCP updates a task's dependencies through task_workflow update.
func moveTaskToWaveViaMCP(ctx context.Context, server framework.MCPServer, taskID string, newDeps []string) error {
	args := map[string]interface{}{
		"action":       "update",
		"task_ids":     taskID,
		"dependencies": strings.Join(newDeps, ","),
	}
	if len(newDeps) == 0 {
		args["dependencies"] = ""
	}
	_, err := callToolText(ctx, server, "task_workflow", args)
	return err
}
