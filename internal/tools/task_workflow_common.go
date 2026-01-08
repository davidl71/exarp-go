package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTaskWorkflowApprove handles approve action for batch approving tasks
// Uses native Go with Todo2 file access, falls back to MCP if needed
// This is platform-agnostic (doesn't require Apple FM)
func handleTaskWorkflowApprove(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		// Fallback to Todo2 MCP tools if project root not found
		return handleTaskWorkflowApproveMCP(ctx, params)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		// Fallback to Todo2 MCP tools if file access fails
		return handleTaskWorkflowApproveMCP(ctx, params)
	}

	// Extract parameters
	status := "Review"
	if s, ok := params["status"].(string); ok && s != "" {
		status = normalizeStatus(s)
	}

	newStatus := "Todo"
	if s, ok := params["new_status"].(string); ok && s != "" {
		newStatus = normalizeStatus(s)
	}

	clarificationNone := true
	if cn, ok := params["clarification_none"].(bool); ok {
		clarificationNone = cn
	}

	var filterTag string
	if tag, ok := params["filter_tag"].(string); ok {
		filterTag = tag
	}

	var taskIDs []string
	if ids, ok := params["task_ids"].(string); ok && ids != "" {
		// Parse JSON array of task IDs
		if err := json.Unmarshal([]byte(ids), &taskIDs); err != nil {
			// If not JSON, treat as comma-separated
			taskIDs = strings.Split(ids, ",")
			for i := range taskIDs {
				taskIDs[i] = strings.TrimSpace(taskIDs[i])
			}
		}
	} else if idsList, ok := params["task_ids"].([]interface{}); ok {
		// Handle array directly
		taskIDs = make([]string, len(idsList))
		for i, id := range idsList {
			if idStr, ok := id.(string); ok {
				taskIDs[i] = idStr
			}
		}
	}

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	// Filter tasks to approve
	candidates := []Todo2Task{}
	for _, task := range tasks {
		// Filter by status
		if normalizeStatus(task.Status) != status {
			continue
		}

		// Filter by specific task IDs if provided
		if len(taskIDs) > 0 {
			found := false
			for _, id := range taskIDs {
				if task.ID == id {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Filter by tag if provided
		if filterTag != "" {
			hasTag := false
			for _, tag := range task.Tags {
				if tag == filterTag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Filter by clarification requirement if needed
		if clarificationNone {
			// Check if task needs clarification (similar to clarify list logic)
			needsClarification := task.LongDescription == "" || len(task.LongDescription) < 50
			if needsClarification {
				continue
			}
		}

		candidates = append(candidates, task)
	}

	if dryRun {
		// Return list of tasks that would be approved
		taskList := make([]map[string]interface{}, len(candidates))
		for i, task := range candidates {
			taskList[i] = map[string]interface{}{
				"id":      task.ID,
				"content": task.Content,
				"status":  task.Status,
			}
		}

		result := map[string]interface{}{
			"success":        true,
			"method":         "native_go",
			"dry_run":        true,
			"approved_count": len(candidates),
			"task_ids":       extractTaskIDs(candidates),
			"tasks":          taskList,
		}

		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	// Update tasks
	approvedIDs := []string{}
	updatedCount := 0
	for i := range tasks {
		// Check if this task should be approved
		shouldApprove := false
		for _, candidate := range candidates {
			if tasks[i].ID == candidate.ID {
				shouldApprove = true
				break
			}
		}

		if shouldApprove {
			tasks[i].Status = newStatus
			approvedIDs = append(approvedIDs, tasks[i].ID)
			updatedCount++
		}
	}

	// Save updated tasks
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		// If save fails, try MCP fallback
		return handleTaskWorkflowApproveMCP(ctx, params)
	}

	result := map[string]interface{}{
		"success":        true,
		"method":         "native_go",
		"approved_count": updatedCount,
		"task_ids":       approvedIDs,
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleTaskWorkflowApproveMCP fallback to Todo2 MCP tools when file access fails
func handleTaskWorkflowApproveMCP(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// This is a fallback - in a real implementation, we would call Todo2 MCP update_todos
	// For now, return an error indicating MCP fallback is needed
	// The Python bridge will handle it via consolidated_workflow.py
	return nil, fmt.Errorf("approve action: file access failed, falling back to Python bridge")
}

// extractTaskIDs extracts IDs from a slice of tasks
func extractTaskIDs(tasks []Todo2Task) []string {
	ids := make([]string, len(tasks))
	for i, task := range tasks {
		ids[i] = task.ID
	}
	return ids
}
