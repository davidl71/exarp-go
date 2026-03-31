// task_workflow_maintenance_helpers.go — Task workflow maintenance: fix descriptions, clarity/stale formatters, link planning.
// See also: task_workflow_maintenance.go
package tools

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/spf13/cast"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleTaskWorkflowFixEmptyDescriptions — handleTaskWorkflowFixEmptyDescriptions sets long_description from content for tasks with empty long_description, then syncs to JSON.
//   buildClarityRecommendations — Helper functions for clarity action
//   formatClarityAnalysisText
//   formatStaleTasks — Helper functions for cleanup action
//   formatStaleTasksFromPtrs
//   allowedStatusForLinkPlanning — allowedStatusForLinkPlanning restricts link_planning to Todo and In Progress only.
//   handleTaskWorkflowLinkPlanning — handleTaskWorkflowLinkPlanning sets planning_doc and/or epic_id on existing tasks.
// ────────────────────────────────────────────────────────────────────────────

// ─── handleTaskWorkflowFixEmptyDescriptions ─────────────────────────────────
// handleTaskWorkflowFixEmptyDescriptions sets long_description from content for tasks with empty long_description, then syncs to JSON.
func handleTaskWorkflowFixEmptyDescriptions(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	dryRun := false
	if _, has := params["dry_run"]; has {
		dryRun = cast.ToBool(params["dry_run"])
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	if dryRun {
		var ids []string

		for _, t := range list {
			if t != nil && strings.TrimSpace(t.LongDescription) == "" && t.Content != "" {
				ids = append(ids, t.ID)
			}
		}

		result := map[string]interface{}{
			"success":       true,
			"method":        "store",
			"dry_run":       true,
			"tasks_updated": len(ids),
			"task_ids":      ids,
		}

		return framework.FormatResult(result, "")
	}

	var updated int

	for _, task := range list {
		if task == nil {
			continue
		}

		if strings.TrimSpace(task.LongDescription) == "" && task.Content != "" {
			task.LongDescription = task.Content
			if err := store.UpdateTask(ctx, task); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to update task %s: %v\n", task.ID, err)
				continue
			}

			updated++
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"method":        "store",
		"tasks_updated": updated,
	}
	outputPath := ParamOutputPath(params)

	return framework.FormatResult(result, outputPath)
}

// ─── handleTaskWorkflowFixEmptyNames ────────────────────────────────────────
// handleTaskWorkflowFixEmptyNames backfills name from content (fallback: long_description)
// for tasks with empty name. This improves list UIs and enables name-first summary queries.
func handleTaskWorkflowFixEmptyNames(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	dryRun := false
	if _, has := params["dry_run"]; has {
		dryRun = cast.ToBool(params["dry_run"])
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	if dryRun {
		ids := make([]string, 0)
		for _, t := range list {
			if t == nil {
				continue
			}
			if strings.TrimSpace(t.Name) == "" {
				before := t.Name
				t.EnsureName()
				if strings.TrimSpace(t.Name) != "" && strings.TrimSpace(before) == "" {
					ids = append(ids, t.ID)
				}
			}
		}

		result := map[string]interface{}{
			"success":       true,
			"method":        "store",
			"dry_run":       true,
			"tasks_updated": len(ids),
			"task_ids":      ids,
		}

		return framework.FormatResult(result, "")
	}

	updated := 0
	for _, task := range list {
		if task == nil {
			continue
		}

		if strings.TrimSpace(task.Name) != "" {
			continue
		}

		task.EnsureName()
		if strings.TrimSpace(task.Name) == "" {
			continue
		}

		if err := store.UpdateTask(ctx, task); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update task %s: %v\n", task.ID, err)
			continue
		}
		updated++
	}

	result := map[string]interface{}{
		"success":       true,
		"method":        "store",
		"tasks_updated": updated,
	}
	outputPath := ParamOutputPath(params)

	return framework.FormatResult(result, outputPath)
}

// ─── buildClarityRecommendations ────────────────────────────────────────────
// Helper functions for clarity action

type storeDriftReport struct {
	JSONOnlyIDs []string `json:"json_only_ids,omitempty"`
	DBOnlyIDs   []string `json:"db_only_ids,omitempty"`
}

func detectTaskStoreDrift(projectRoot string) (*storeDriftReport, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root is required")
	}

	jsonTasks, err := loadTodo2TasksFromJSON(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("load json for store drift: %w", err)
	}

	dbTasks, err := loadTodo2TasksFromDB(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("load database for store drift: %w", err)
	}

	dbSet := make(map[string]struct{}, len(dbTasks))
	for _, task := range dbTasks {
		dbSet[task.ID] = struct{}{}
	}
	jsonSet := make(map[string]struct{}, len(jsonTasks))
	for _, task := range jsonTasks {
		jsonSet[task.ID] = struct{}{}
	}

	var jsonOnly, dbOnly []string
	for _, task := range jsonTasks {
		if _, ok := dbSet[task.ID]; !ok {
			jsonOnly = append(jsonOnly, task.ID)
		}
	}
	for _, task := range dbTasks {
		if _, ok := jsonSet[task.ID]; !ok {
			dbOnly = append(dbOnly, task.ID)
		}
	}
	sort.Strings(jsonOnly)
	sort.Strings(dbOnly)

	return &storeDriftReport{
		JSONOnlyIDs: jsonOnly,
		DBOnlyIDs:   dbOnly,
	}, nil
}

func buildClarityRecommendations(issues []map[string]interface{}) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	if len(issues) > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":    "clarity_improvement",
			"count":   len(issues),
			"message": fmt.Sprintf("%d tasks need clarity improvements", len(issues)),
		})
	}

	return recommendations
}

// ─── formatClarityAnalysisText ──────────────────────────────────────────────
func formatClarityAnalysisText(result map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("Task Clarity Analysis\n")
	sb.WriteString("=====================\n\n")

	if total, ok := result["total_tasks"].(int); ok {
		sb.WriteString(fmt.Sprintf("Total Tasks: %d\n", total))
	}

	if analyzed, ok := result["tasks_analyzed"].(int); ok {
		sb.WriteString(fmt.Sprintf("Tasks Needing Improvement: %d\n\n", analyzed))
	}

	if issues, ok := result["clarity_issues"].([]map[string]interface{}); ok && len(issues) > 0 {
		sb.WriteString("Clarity Issues:\n\n")

		for i, issue := range issues {
			if taskID, ok := issue["task_id"].(string); ok {
				sb.WriteString(fmt.Sprintf("%d. Task %s\n", i+1, taskID))

				if content, ok := issue["content"].(string); ok {
					sb.WriteString(fmt.Sprintf("   Content: %s\n", content))
				}

				if issuesList, ok := issue["issues"].([]string); ok {
					sb.WriteString(fmt.Sprintf("   Issues: %s\n", strings.Join(issuesList, ", ")))
				}

				if suggestions, ok := issue["suggestions"].([]string); ok {
					sb.WriteString("   Suggestions:\n")

					for _, sug := range suggestions {
						sb.WriteString(fmt.Sprintf("     - %s\n", sug))
					}
				}

				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// ─── formatStaleTasks ───────────────────────────────────────────────────────
// Helper functions for cleanup action

func formatStaleTasks(tasks []Todo2Task) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		result[i] = map[string]interface{}{
			"id":      task.ID,
			"content": task.Content,
			"status":  task.Status,
		}
	}

	return result
}

// ─── formatStaleTasksFromPtrs ───────────────────────────────────────────────
func formatStaleTasksFromPtrs(tasks []*models.Todo2Task) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		result[i] = map[string]interface{}{
			"id":      task.ID,
			"content": task.Content,
			"status":  task.Status,
		}
	}

	return result
}

// ─── allowedStatusForLinkPlanning ───────────────────────────────────────────
// allowedStatusForLinkPlanning restricts link_planning to Todo and In Progress only.
var allowedStatusForLinkPlanning = map[string]bool{
	models.StatusTodo:       true,
	models.StatusInProgress: true,
}

// ─── handleTaskWorkflowLinkPlanning ─────────────────────────────────────────
// handleTaskWorkflowLinkPlanning sets planning_doc and/or epic_id on existing tasks.
// Only tasks with status Todo or In Progress are updated.
func handleTaskWorkflowLinkPlanning(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	planningDoc := cast.ToString(params["planning_doc"])
	epicID := cast.ToString(params["epic_id"])

	if planningDoc == "" && epicID == "" {
		return nil, fmt.Errorf("at least one of planning_doc or epic_id is required for link_planning")
	}

	if planningDoc != "" {
		if err := ValidatePlanningLink(projectRoot, planningDoc); err != nil {
			return nil, fmt.Errorf("invalid planning_doc: %w", err)
		}
	}

	uniqueIDs := ParseTaskIDsFromParams(params)
	if len(uniqueIDs) == 0 {
		return nil, fmt.Errorf("task_id or task_ids is required for link_planning")
	}

	// Load all tasks for lookup; use TaskStore for unified DB/file path
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)
	taskMap := make(map[string]Todo2Task)

	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	if epicID != "" {
		if err := ValidateTaskReferenceInStore(ctx, store, epicID, tasks); err != nil {
			return nil, fmt.Errorf("invalid epic_id: %w", err)
		}
	}

	updatedIDs := make([]string, 0)

	var skippedStatus []string

	for _, id := range uniqueIDs {
		task, ok := taskMap[id]
		if !ok {
			ptr, gerr := store.GetTask(ctx, id)
			if gerr != nil {
				skippedStatus = append(skippedStatus, id+": not found")
				continue
			}
			task = *ptr
		}

		norm := normalizeStatus(strings.TrimSpace(task.Status))
		if !allowedStatusForLinkPlanning[norm] {
			skippedStatus = append(skippedStatus, id+": status is "+norm+", only Todo or In Progress allowed")
			continue
		}

		linkMeta := GetPlanningLinkMetadata(&task)
		if linkMeta == nil {
			linkMeta = &PlanningLinkMetadata{}
		}

		if planningDoc != "" {
			linkMeta.PlanningDoc = planningDoc
		}

		if epicID != "" {
			linkMeta.EpicID = epicID
			task.ParentID = epicID
		}

		SetPlanningLinkMetadata(&task, linkMeta)

		taskPtr := &task
		if err := store.UpdateTask(ctx, taskPtr); err != nil {
			skippedStatus = append(skippedStatus, id+": update failed: "+err.Error())
			continue
		}

		updatedIDs = append(updatedIDs, id)
	}

	result := map[string]interface{}{
		"success":       true,
		"method":        "store",
		"action":        "link_planning",
		"updated_count": len(updatedIDs),
		"updated_ids":   updatedIDs,
		"skipped":       skippedStatus,
	}

	return framework.FormatResult(result, "")
}

// handleTaskWorkflowCreate handles create action for creating new tasks.
// Supports single task (name param) or batch create (tasks JSON array param).
// Uses database for efficient creation, falls back to file-based approach if needed.
