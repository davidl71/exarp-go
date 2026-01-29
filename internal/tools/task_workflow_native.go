package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTaskWorkflowNative handles task_workflow with native Go and FM chain (Apple → Ollama → stub)
func handleTaskWorkflowNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "sync"
	}

	switch action {
	case "clarify":
		return handleTaskWorkflowClarify(ctx, params)
	case "approve":
		return handleTaskWorkflowApprove(ctx, params)
	case "sync":
		return handleTaskWorkflowSync(ctx, params)
	case "fix_dates":
		return handleTaskWorkflowFixDates(ctx, params)
	case "clarity":
		return handleTaskWorkflowClarity(ctx, params)
	case "cleanup":
		return handleTaskWorkflowCleanup(ctx, params)
	case "create":
		return handleTaskWorkflowCreate(ctx, params)
	case "sanity_check":
		return handleTaskWorkflowSanityCheck(ctx, params)
	case "link_planning":
		return handleTaskWorkflowLinkPlanning(ctx, params)
	case "delete":
		return handleTaskWorkflowDelete(ctx, params)
	case "update":
		return handleTaskWorkflowUpdate(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// handleTaskWorkflowFixDates backfills created/last_modified from DB created_at/updated_at for tasks with empty or 1970 dates, then syncs to JSON.
func handleTaskWorkflowFixDates(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	updated, err := database.FixTaskDates(ctx)
	if err != nil {
		return nil, fmt.Errorf("fix_dates: %w", err)
	}
	projectRoot, err := FindProjectRoot()
	if err != nil {
		// Return success with count even if sync fails (DB is fixed)
		result := map[string]interface{}{
			"success":       true,
			"method":        "native_go",
			"tasks_updated": updated,
			"sync_skipped":  true,
			"sync_error":    err.Error(),
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
	}
	if err := SyncTodo2Tasks(projectRoot); err != nil {
		result := map[string]interface{}{
			"success":       true,
			"method":        "native_go",
			"tasks_updated": updated,
			"sync_error":    err.Error(),
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
	}
	// Regenerate overview so dates never show 1970
	if overviewErr := WriteTodo2Overview(projectRoot); overviewErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write todo2-overview.mdc: %v\n", overviewErr)
	}
	result := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"tasks_updated": updated,
		"synced":        true,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
}

// handleTaskWorkflowClarify handles clarify action with default FM for question generation
func handleTaskWorkflowClarify(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	if !FMAvailable() {
		return nil, ErrFMNotSupported
	}

	subAction, _ := params["sub_action"].(string)
	if subAction == "" {
		subAction = "list"
	}

	switch subAction {
	case "list":
		return listTasksAwaitingClarification(ctx, params)
	case "resolve":
		return resolveTaskClarification(ctx, params)
	case "batch":
		return resolveBatchClarifications(ctx, params)
	default:
		return nil, fmt.Errorf("unknown sub_action: %s (use 'list', 'resolve', or 'batch')", subAction)
	}
}

// listTasksAwaitingClarification lists tasks that need clarification
func listTasksAwaitingClarification(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Find tasks that need clarification (have clarification comments or are unclear)
	needingClarification := []map[string]interface{}{}

	for _, task := range tasks {
		if !IsPendingStatus(task.Status) {
			continue
		}

		// Check if task needs clarification
		needsClarification := false
		clarificationText := ""

		// Check for unclear descriptions (use config for min length)
		minDescLen := config.TaskMinDescriptionLength()
		if task.LongDescription == "" || len(task.LongDescription) < minDescLen {
			needsClarification = true
			clarificationText = "Task description is too brief or missing"
		}

		// Use default FM to generate clarification question
		if needsClarification {
			question := generateClarificationQuestion(ctx, task, clarificationText)
			needingClarification = append(needingClarification, map[string]interface{}{
				"task_id":            task.ID,
				"content":            task.Content,
				"status":             task.Status,
				"clarification_text": question,
				"reason":             clarificationText,
			})
		}
	}

	result := map[string]interface{}{
		"success":                      true,
		"method":                       "fm_chain",
		"tasks_awaiting_clarification": len(needingClarification),
		"tasks":                        needingClarification,
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// resolveTaskClarification resolves a single task clarification
func resolveTaskClarification(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required for resolve action")
	}

	// Try database first for efficient single-task update
	if db, err := database.GetDB(); err == nil && db != nil {
		task, err := database.GetTask(ctx, taskID)
		if err != nil {
			return nil, fmt.Errorf("task %s not found: %w", taskID, err)
		}

		clarificationText, _ := params["clarification_text"].(string)
		decision, _ := params["decision"].(string)

		// Update task with clarification response
		if clarificationText != "" {
			// Add clarification to long_description
			if task.LongDescription == "" {
				task.LongDescription = fmt.Sprintf("Clarification: %s", clarificationText)
			} else {
				task.LongDescription += fmt.Sprintf("\n\nClarification: %s", clarificationText)
			}
		}

		if decision != "" {
			// Update task based on decision
			if task.Metadata == nil {
				task.Metadata = make(map[string]interface{})
			}
			task.Metadata["clarification_decision"] = decision
		}

		moveToTodo := true
		if move, ok := params["move_to_todo"].(bool); ok {
			moveToTodo = move
		}

		if moveToTodo {
			task.Status = "Todo"
		}

		// Update task in database
		if err := database.UpdateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("failed to update task: %w", err)
		}

		result := map[string]interface{}{
			"success": true,
			"method":  "database",
			"task_id": taskID,
			"message": "Clarification resolved",
		}

		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	// Fallback to file-based approach
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Find and update task
	found := false
	for i := range tasks {
		if tasks[i].ID == taskID {
			clarificationText, _ := params["clarification_text"].(string)
			decision, _ := params["decision"].(string)

			// Update task with clarification response
			if clarificationText != "" {
				// Add clarification to long_description or metadata
				if tasks[i].LongDescription == "" {
					tasks[i].LongDescription = fmt.Sprintf("Clarification: %s", clarificationText)
				} else {
					tasks[i].LongDescription += fmt.Sprintf("\n\nClarification: %s", clarificationText)
				}
			}

			if decision != "" {
				// Update task based on decision
				if tasks[i].Metadata == nil {
					tasks[i].Metadata = make(map[string]interface{})
				}
				tasks[i].Metadata["clarification_decision"] = decision
			}

			moveToTodo := true
			if move, ok := params["move_to_todo"].(bool); ok {
				moveToTodo = move
			}

			if moveToTodo {
				tasks[i].Status = "Todo"
			}

			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	// Save updated tasks
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		return nil, fmt.Errorf("failed to save tasks: %w", err)
	}

	result := map[string]interface{}{
		"success": true,
		"method":  "file",
		"task_id": taskID,
		"message": "Clarification resolved",
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// resolveBatchClarifications resolves multiple clarifications
func resolveBatchClarifications(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	decisionsJSON, _ := params["decisions_json"].(string)
	if decisionsJSON == "" {
		return nil, fmt.Errorf("decisions_json is required for batch action")
	}

	var decisions []map[string]interface{}
	if err := json.Unmarshal([]byte(decisionsJSON), &decisions); err != nil {
		return nil, fmt.Errorf("failed to parse decisions_json: %w", err)
	}

	// Try database first for efficient batch updates
	if db, err := database.GetDB(); err == nil && db != nil {
		resolved := 0
		for _, decision := range decisions {
			taskID, _ := decision["task_id"].(string)
			if taskID == "" {
				continue
			}

			task, err := database.GetTask(ctx, taskID)
			if err != nil {
				// Task not found, skip
				continue
			}

			clarificationText, _ := decision["clarification_text"].(string)
			decisionText, _ := decision["decision"].(string)

			if clarificationText != "" {
				if task.LongDescription == "" {
					task.LongDescription = fmt.Sprintf("Clarification: %s", clarificationText)
				} else {
					task.LongDescription += fmt.Sprintf("\n\nClarification: %s", clarificationText)
				}
			}

			if decisionText != "" {
				if task.Metadata == nil {
					task.Metadata = make(map[string]interface{})
				}
				task.Metadata["clarification_decision"] = decisionText
			}

			moveToTodo := true
			if move, ok := decision["move_to_todo"].(bool); ok {
				moveToTodo = move
			}

			if moveToTodo {
				task.Status = "Todo"
			}

			// Update task in database
			if err := database.UpdateTask(ctx, task); err == nil {
				resolved++
			}
		}

		result := map[string]interface{}{
			"success":  true,
			"method":   "database",
			"resolved": resolved,
			"total":    len(decisions),
			"message":  fmt.Sprintf("Resolved %d clarifications", resolved),
		}

		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	// Fallback to file-based approach
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	resolved := 0
	for _, decision := range decisions {
		taskID, _ := decision["task_id"].(string)
		if taskID == "" {
			continue
		}

		for i := range tasks {
			if tasks[i].ID == taskID {
				clarificationText, _ := decision["clarification_text"].(string)
				decisionText, _ := decision["decision"].(string)

				if clarificationText != "" {
					if tasks[i].LongDescription == "" {
						tasks[i].LongDescription = fmt.Sprintf("Clarification: %s", clarificationText)
					} else {
						tasks[i].LongDescription += fmt.Sprintf("\n\nClarification: %s", clarificationText)
					}
				}

				if decisionText != "" {
					if tasks[i].Metadata == nil {
						tasks[i].Metadata = make(map[string]interface{})
					}
					tasks[i].Metadata["clarification_decision"] = decisionText
				}

				moveToTodo := true
				if move, ok := decision["move_to_todo"].(bool); ok {
					moveToTodo = move
				}

				if moveToTodo {
					tasks[i].Status = "Todo"
				}

				resolved++
				break
			}
		}
	}

	// Save updated tasks
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		return nil, fmt.Errorf("failed to save tasks: %w", err)
	}

	result := map[string]interface{}{
		"success":  true,
		"method":   "file",
		"resolved": resolved,
		"total":    len(decisions),
		"message":  fmt.Sprintf("Resolved %d clarifications", resolved),
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handleTaskWorkflowDelete deletes a task by ID (e.g. wrong project). Syncs DB to JSON after delete.
func handleTaskWorkflowDelete(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required for delete action")
	}
	if err := database.DeleteTask(ctx, taskID); err != nil {
		return nil, fmt.Errorf("delete task: %w", err)
	}
	projectRoot, err := FindProjectRoot()
	if err != nil {
		result := map[string]interface{}{"success": true, "method": "database", "task_id": taskID, "deleted": true, "sync_skipped": true}
		out, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
	}
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		result := map[string]interface{}{"success": true, "method": "database", "task_id": taskID, "deleted": true, "sync_skipped": true}
		out, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
	}
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		result := map[string]interface{}{"success": true, "method": "database", "task_id": taskID, "deleted": true, "sync_error": err.Error()}
		out, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
	}
	result := map[string]interface{}{"success": true, "method": "database", "task_id": taskID, "deleted": true, "synced": true}
	out, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
}

// generateClarificationQuestion uses the default FM to generate clarification questions
func generateClarificationQuestion(ctx context.Context, task Todo2Task, reason string) string {
	if !FMAvailable() {
		return fmt.Sprintf("Task needs clarification: %s", reason)
	}

	prompt := fmt.Sprintf(`Generate a specific clarification question for this task:

Task ID: %s
Content: %s
Description: %s
Reason: %s

Generate a single, specific question that would help clarify what needs to be done. Return only the question, no explanation.`,
		task.ID, task.Content, task.LongDescription, reason)

	result, err := DefaultFMProvider().Generate(ctx, prompt, 100, 0.3)
	if err != nil {
		return fmt.Sprintf("Task needs clarification: %s", reason)
	}
	return strings.TrimSpace(result)
}
