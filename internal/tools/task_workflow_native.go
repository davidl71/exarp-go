//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	fm "github.com/blacktop/go-foundationmodels"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
)

// handleTaskWorkflowNative handles task_workflow with native Go and Apple FM
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
	case "sync", "clarity", "cleanup":
		// These actions don't use native Go yet - fall through to Python bridge
		return nil, fmt.Errorf("action %s not yet implemented in native Go", action)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// handleTaskWorkflowClarify handles clarify action with Apple FM for question generation
func handleTaskWorkflowClarify(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Check platform support
	support := platform.CheckAppleFoundationModelsSupport()
	if !support.Supported {
		return nil, fmt.Errorf("Apple Foundation Models not supported: %s", support.Reason)
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

		// Check for unclear descriptions
		if task.LongDescription == "" || len(task.LongDescription) < 50 {
			needsClarification = true
			clarificationText = "Task description is too brief or missing"
		}

		// Use Apple FM to detect if task needs clarification
		if needsClarification {
			question := generateClarificationQuestion(task, clarificationText)
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
		"success":                true,
		"method":                 "apple_foundation_models",
		"tasks_awaiting_clarification": len(needingClarification),
		"tasks":                   needingClarification,
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
		"method":  "native_go",
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
		"success":      true,
		"method":       "native_go",
		"resolved":     resolved,
		"total":        len(decisions),
		"message":      fmt.Sprintf("Resolved %d clarifications", resolved),
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// generateClarificationQuestion uses Apple FM to generate clarification questions
func generateClarificationQuestion(task Todo2Task, reason string) string {
	support := platform.CheckAppleFoundationModelsSupport()
	if !support.Supported {
		return fmt.Sprintf("Task needs clarification: %s", reason)
	}

	sess := fm.NewSession()
	defer sess.Release()

	prompt := fmt.Sprintf(`Generate a specific clarification question for this task:

Task ID: %s
Content: %s
Description: %s
Reason: %s

Generate a single, specific question that would help clarify what needs to be done. Return only the question, no explanation.`,
		task.ID, task.Content, task.LongDescription, reason)

	result := sess.RespondWithOptions(prompt, 100, 0.3)
	return strings.TrimSpace(result)
}

