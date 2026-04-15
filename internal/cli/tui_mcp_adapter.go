// tui_mcp_adapter.go — Routes TUI data operations through MCP tools instead of direct database access.
// All task CRUD goes through task_workflow; handoffs/analysis already use server.CallTool.
// See docs/TUI_MCP_ADAPTER_DESIGN.md for design rationale.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/tools"
)

// mcpListTasksResponse is the JSON envelope from task_workflow sync/list with output_format=json.
type mcpListTasksResponse struct {
	Success bool          `json:"success"`
	Method  string        `json:"method"`
	Tasks   []mcpTaskJSON `json:"tasks"`
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

// createTaskViaMCP creates a new task through task_workflow create action.
func createTaskViaMCP(ctx context.Context, server framework.MCPServer, name string) (string, error) {
	text, err := callToolText(ctx, server, "task_workflow", map[string]interface{}{
		"action":        "create",
		"name":          name,
		"priority":      "medium",
		"output_format": "json",
	})
	if err != nil {
		return "", err
	}
	var resp struct {
		TaskID string `json:"task_id"`
		ID     string `json:"id"`
	}
	if json.Unmarshal([]byte(text), &resp) == nil && (resp.TaskID != "" || resp.ID != "") {
		if resp.TaskID != "" {
			return resp.TaskID, nil
		}
		return resp.ID, nil
	}
	return "", nil
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

// getTaskViaMCP fetches a single task by ID through task_workflow sync/list with task_ids filter.
func getTaskViaMCP(ctx context.Context, server framework.MCPServer, taskID string) (*database.Todo2Task, error) {
	text, err := callToolText(ctx, server, "task_workflow", map[string]interface{}{
		"action":        "sync",
		"sub_action":    "list",
		"task_ids":      taskID,
		"output_format": "json",
		"compact":       true,
	})
	if err != nil {
		return nil, fmt.Errorf("task_workflow list for %s: %w", taskID, err)
	}

	var resp mcpListTasksResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		return nil, fmt.Errorf("parse task list JSON: %w", err)
	}

	if len(resp.Tasks) > 0 {
		return resp.Tasks[0].toTodo2Task(), nil
	}

	return nil, fmt.Errorf("task %s not found", taskID)
}

// HandoffEntry is a typed representation of a session handoff note.
// Replaces map[string]interface{} access across both TUIs.
type HandoffEntry struct {
	ID        string   `json:"id,omitempty"`
	Summary   string   `json:"summary,omitempty"`
	Host      string   `json:"host,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
	Status    string   `json:"status,omitempty"`
	NextSteps []string `json:"next_steps,omitempty"`
	Blockers  []string `json:"blockers,omitempty"`
}

// handoffListPayload is the JSON envelope from session handoff list.
type handoffListPayload struct {
	Handoffs []HandoffEntry `json:"handoffs"`
	Count    int            `json:"count"`
	Total    int            `json:"total"`
}

// fetchHandoffs loads handoff entries via the session MCP tool.
func fetchHandoffs(ctx context.Context, server framework.MCPServer, limit int) ([]HandoffEntry, error) {
	args := map[string]interface{}{"action": "handoff", "sub_action": "list"}
	if limit > 0 {
		args["limit"] = float64(limit)
	}

	text, err := callToolText(ctx, server, "session", args)
	if err != nil {
		return nil, fmt.Errorf("session handoff list: %w", err)
	}

	var payload handoffListPayload
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil, fmt.Errorf("parse handoff JSON: %w", err)
	}

	return payload.Handoffs, nil
}

// firstWaveTaskIDs converts task pointers to values and returns the first wave's task IDs.
func firstWaveTaskIDs(projectRoot string, tasks []*database.Todo2Task) (waveLevel int, ids []string, err error) {
	taskList := make([]database.Todo2Task, 0, len(tasks))
	for _, t := range tasks {
		if t != nil {
			taskList = append(taskList, *t)
		}
	}

	waves, err := tools.ComputeWavesForTUI(projectRoot, taskList)
	if err != nil {
		return 0, nil, err
	}
	if len(waves) == 0 {
		return 0, nil, fmt.Errorf("no waves")
	}

	levels := make([]int, 0, len(waves))
	for k := range waves {
		levels = append(levels, k)
	}
	sort.Ints(levels)

	return levels[0], waves[levels[0]], nil
}

// updateTaskFieldsViaMCP updates a task's status, priority, and description through task_workflow.
func updateTaskFieldsViaMCP(ctx context.Context, server framework.MCPServer, taskID, status, priority, description string) error {
	args := map[string]interface{}{
		"action":   "update",
		"task_ids": taskID,
	}

	if status != "" {
		args["new_status"] = status
	}

	if priority != "" {
		args["new_priority"] = priority
	}

	if description != "" {
		args["long_description"] = description
	}

	_, err := callToolText(ctx, server, "task_workflow", args)
	return err
}
