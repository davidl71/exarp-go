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
	"github.com/davidl71/exarp-go/internal/models"
	mcpframework "github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
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
	case "fix_empty_descriptions":
		return handleTaskWorkflowFixEmptyDescriptions(ctx, params)
	case "clarity":
		return handleTaskWorkflowClarity(ctx, params)
	case "cleanup":
		return handleTaskWorkflowCleanup(ctx, params)
	case "create":
		return handleTaskWorkflowCreate(ctx, params)
	case "sanity_check":
		return handleTaskWorkflowSanityCheck(ctx, params)
	case "fix_invalid_ids":
		return handleTaskWorkflowFixInvalidIDs(ctx, params)
	case "link_planning":
		return handleTaskWorkflowLinkPlanning(ctx, params)
	case "delete":
		return handleTaskWorkflowDelete(ctx, params)
	case "update":
		return handleTaskWorkflowUpdate(ctx, params)
	case "request_approval":
		return handleTaskWorkflowRequestApproval(ctx, params)
	case "sync_approvals":
		return handleTaskWorkflowSyncApprovals(ctx, params)
	case "apply_approval_result":
		return handleTaskWorkflowApplyApprovalResult(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// handleTaskWorkflowSyncApprovals returns approval requests for all tasks in Review (T-111).
// The client can send each to gotoHuman via request-human-review-with-form.
func handleTaskWorkflowSyncApprovals(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	formID, _ := params["form_id"].(string)
	var tasks []*models.Todo2Task
	if db, err := database.GetDB(); err == nil && db != nil {
		reviewTasks, err := database.GetTasksByStatus(ctx, "Review")
		if err == nil {
			tasks = reviewTasks
		}
	}
	if tasks == nil {
		projectRoot, err := FindProjectRoot()
		if err != nil {
			return nil, fmt.Errorf("sync_approvals: %w", err)
		}
		all, err := LoadTodo2Tasks(projectRoot)
		if err != nil {
			return nil, fmt.Errorf("sync_approvals: failed to load tasks: %w", err)
		}
		for i := range all {
			if all[i].Status == "Review" {
				tasks = append(tasks, &all[i])
			}
		}
	}
	approvalRequests := make([]ApprovalRequest, 0, len(tasks))
	for _, task := range tasks {
		approvalRequests = append(approvalRequests, BuildApprovalRequestFromTask(task, formID))
	}
	result := map[string]interface{}{
		"review_count":      len(approvalRequests),
		"approval_requests": approvalRequests,
		"instructions":      "Call @gotoHuman request-human-review-with-form for each approval_request (form_id, field_data). Set GOTOHUMAN_API_KEY if needed. See docs/GOTOHUMAN_API_REFERENCE.md.",
	}
	return response.FormatResult(result, "")
}

// handleTaskWorkflowApplyApprovalResult updates a task when human approves or rejects in gotoHuman (T-112).
// Params: task_id (required), result (required: "approved" or "rejected"), feedback (optional, for rejection).
func handleTaskWorkflowApplyApprovalResult(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("apply_approval_result requires task_id")
	}
	resultVal, _ := params["result"].(string)
	resultVal = strings.TrimSpace(strings.ToLower(resultVal))
	if resultVal != "approved" && resultVal != "rejected" {
		return nil, fmt.Errorf("apply_approval_result requires result=approved or result=rejected")
	}
	feedback, _ := params["feedback"].(string)
	newStatus := "Done"
	if resultVal == "rejected" {
		newStatus = "In Progress"
	}
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("apply_approval_result: %w", err)
	}
	// Update via same path as update action
	updateParams := map[string]interface{}{
		"task_ids":   taskID,
		"new_status": newStatus,
	}
	out, err := handleTaskWorkflowUpdate(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("apply_approval_result: %w", err)
	}
	// Optionally add feedback to task (e.g. as comment or in long_description)
	if feedback != "" && resultVal == "rejected" {
		tasks, _ := LoadTodo2Tasks(projectRoot)
		for i := range tasks {
			if tasks[i].ID == taskID {
				if tasks[i].LongDescription == "" {
					tasks[i].LongDescription = "Rejection feedback: " + feedback
				} else {
					tasks[i].LongDescription += "\n\nRejection feedback: " + feedback
				}
				_ = SaveTodo2Tasks(projectRoot, tasks)
				break
			}
		}
	}
	return out, nil
}

// handleTaskWorkflowRequestApproval builds a gotoHuman approval request payload for a Todo2 task.
// The client (e.g. Cursor) should call @gotoHuman request-human-review-with-form with the returned payload.
func handleTaskWorkflowRequestApproval(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("request_approval requires task_id")
	}
	formID, _ := params["form_id"].(string)

	var task *models.Todo2Task
	if db, err := database.GetDB(); err == nil && db != nil {
		t, err := database.GetTask(ctx, taskID)
		if err == nil && t != nil {
			task = t
		}
	}
	if task == nil {
		projectRoot, err := FindProjectRoot()
		if err != nil {
			return nil, fmt.Errorf("request_approval: %w", err)
		}
		tasks, err := LoadTodo2Tasks(projectRoot)
		if err != nil {
			return nil, fmt.Errorf("request_approval: failed to load tasks: %w", err)
		}
		for i := range tasks {
			if tasks[i].ID == taskID {
				task = &tasks[i]
				break
			}
		}
	}
	if task == nil {
		return nil, fmt.Errorf("request_approval: task %s not found", taskID)
	}

	req := BuildApprovalRequestFromTask(task, formID)
	payload, _ := json.Marshal(req)
	instructions := "Call @gotoHuman request-human-review-with-form with formId and fieldData from approval_request. Set GOTOHUMAN_API_KEY if needed. See docs/GOTOHUMAN_API_REFERENCE.md."
	result := map[string]interface{}{
		"task_id":          taskID,
		"approval_request": req,
		"instructions":     instructions,
	}
	return response.FormatResult(result, string(payload))
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
		return response.FormatResult(result, "")
	}
	if err := SyncTodo2Tasks(projectRoot); err != nil {
		result := map[string]interface{}{
			"success":       true,
			"method":        "native_go",
			"tasks_updated": updated,
			"sync_error":    err.Error(),
		}
		return response.FormatResult(result, "")
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
	return response.FormatResult(result, "")
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

	return response.FormatResult(result, "")
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

		// Sync DB to JSON (shared workflow)
		if projectRoot, syncErr := FindProjectRoot(); syncErr == nil {
			_ = SyncTodo2Tasks(projectRoot)
		}

		result := map[string]interface{}{
			"success": true,
			"method":  "database",
			"task_id": taskID,
			"message": "Clarification resolved",
		}

		return response.FormatResult(result, "")
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

	return response.FormatResult(result, "")
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

		// Sync DB to JSON (shared workflow)
		if resolved > 0 {
			if projectRoot, syncErr := FindProjectRoot(); syncErr == nil {
				_ = SyncTodo2Tasks(projectRoot)
			}
		}

		result := map[string]interface{}{
			"success":  true,
			"method":   "database",
			"resolved": resolved,
			"total":    len(decisions),
			"message":  fmt.Sprintf("Resolved %d clarifications", resolved),
		}

		return response.FormatResult(result, "")
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

	return response.FormatResult(result, "")
}

// handleTaskWorkflowDelete deletes one or more tasks by ID. Accepts task_id (single) or task_ids (comma-separated or array). Syncs DB to JSON once after all deletes.
func handleTaskWorkflowDelete(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Optional MCP Elicitation: confirm delete when confirm_via_elicitation is true
	if confirm, _ := params["confirm_via_elicitation"].(bool); confirm {
		if eliciter := mcpframework.EliciterFromContext(ctx); eliciter != nil {
			schema := map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"proceed": map[string]interface{}{"type": "boolean", "description": "Proceed with delete?"},
				},
			}
			action, content, err := eliciter.ElicitForm(ctx, "Proceed with deleting the specified task(s)?", schema)
			if err != nil || action != "accept" {
				result := map[string]interface{}{"cancelled": true, "message": "Delete cancelled by user or elicitation unavailable"}
				return response.FormatResult(result, "")
			}
			if content != nil {
				if proceed, ok := content["proceed"].(bool); ok && !proceed {
					result := map[string]interface{}{"cancelled": true, "message": "Delete cancelled by user"}
					return response.FormatResult(result, "")
				}
			}
		}
	}

	var ids []string
	if taskIDsRaw, ok := params["task_ids"]; ok && taskIDsRaw != nil {
		switch v := taskIDsRaw.(type) {
		case string:
			if v != "" {
				ids = strings.Split(strings.TrimSpace(v), ",")
				for i := range ids {
					ids[i] = strings.TrimSpace(ids[i])
				}
			}
		case []interface{}:
			for _, x := range v {
				if s, ok := x.(string); ok && s != "" {
					ids = append(ids, strings.TrimSpace(s))
				}
			}
		}
	}
	if len(ids) == 0 {
		if taskID, _ := params["task_id"].(string); taskID != "" {
			ids = []string{strings.TrimSpace(taskID)}
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("task_id or task_ids is required for delete action")
	}
	var deleted, failed []string
	for _, id := range ids {
		if id == "" {
			continue
		}
		if err := database.DeleteTask(ctx, id); err != nil {
			failed = append(failed, id+": "+err.Error())
			continue
		}
		deleted = append(deleted, id)
	}
	projectRoot, err := FindProjectRoot()
	if err != nil {
		result := map[string]interface{}{"success": len(failed) == 0, "method": "database", "deleted": deleted, "failed": failed, "sync_skipped": true}
		return response.FormatResult(result, "")
	}
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		result := map[string]interface{}{"success": len(failed) == 0, "method": "database", "deleted": deleted, "failed": failed, "sync_skipped": true}
		return response.FormatResult(result, "")
	}
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		result := map[string]interface{}{"success": len(failed) == 0, "method": "database", "deleted": deleted, "failed": failed, "sync_error": err.Error()}
		return response.FormatResult(result, "")
	}
	result := map[string]interface{}{"success": len(failed) == 0, "method": "database", "deleted": deleted, "failed": failed, "synced": true}
	return response.FormatResult(result, "")
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
