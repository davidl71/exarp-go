// task_workflow_maintenance.go — Task workflow maintenance: sync, sanity-check, clarity, and cleanup handlers.
// See also: task_workflow_maintenance_helpers.go
package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/spf13/cast"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleTaskWorkflowSync
//   validTaskStatuses — Valid task statuses for sanity check.
//   handleTaskWorkflowSanityCheck — handleTaskWorkflowSanityCheck runs generic Todo2 task sanity checks (epoch dates, empty content, valid status, duplicate IDs, missing deps).
//   handleTaskWorkflowClarity — handleTaskWorkflowClarity handles clarity action for improving task clarity.
//   isOldSequentialID — isOldSequentialID checks if a task ID uses the old sequential format (T-1, T-2, etc.)
//   handleTaskWorkflowCleanup — handleTaskWorkflowCleanup handles cleanup action for removing stale tasks and legacy tasks.
// ────────────────────────────────────────────────────────────────────────────

// ─── handleTaskWorkflowSync ─────────────────────────────────────────────────
func handleTaskWorkflowSync(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	// Check if this is a list sub-action (for listing tasks)
	// If sub_action is "list", we just load and return tasks (no sync needed)
	subAction := cast.ToString(params["sub_action"])
	if subAction == "list" {
		// For list, just load tasks and format them (no sync)
		return handleTaskWorkflowList(ctx, params)
	}

	// Perform bidirectional sync between SQLite and JSON
	if !dryRun {
		if err := SyncTodo2Tasks(projectRoot); err != nil {
			return nil, fmt.Errorf("failed to sync tasks: %w", err)
		}
		// Regenerate overview so dates never show 1970
		if overviewErr := WriteTodo2Overview(projectRoot); overviewErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write todo2-overview.mdc: %v\n", overviewErr)
		}
	}

	// Load tasks after sync to validate
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks after sync: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Validate task consistency
	issues := []string{}

	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	// Tasks that had invalid epic ID cleared (so we persist the fix)
	var tasksWithOrphanedEpicFixed []*models.Todo2Task

	// Check for missing dependencies and invalid planning links
	for i := range tasks {
		task := &tasks[i]
		for _, dep := range task.Dependencies {
			if !taskMap[dep] {
				issues = append(issues, fmt.Sprintf("Task %s depends on %s which doesn't exist", task.ID, dep))
			}
		}

		// Validate planning document links; auto-clear orphaned epic_id
		if linkMeta := GetPlanningLinkMetadata(task); linkMeta != nil {
			if linkMeta.PlanningDoc != "" {
				if err := ValidatePlanningLink(projectRoot, linkMeta.PlanningDoc); err != nil {
					issues = append(issues, fmt.Sprintf("Task %s has invalid planning doc link: %v", task.ID, err))
				}
			}

			if linkMeta.EpicID != "" {
				if err := ValidateTaskReference(linkMeta.EpicID, tasks); err != nil {
					issues = append(issues, fmt.Sprintf("Task %s has invalid epic ID: %v", task.ID, err))
				}
				// Clear orphaned epic_id so future syncs are clean
				linkMeta.EpicID = ""
				SetPlanningLinkMetadata(task, linkMeta)
				tasksWithOrphanedEpicFixed = append(tasksWithOrphanedEpicFixed, task)
			}
		}
	}

	// Persist cleared epic_id for tasks that referenced a missing epic
	if !dryRun && len(tasksWithOrphanedEpicFixed) > 0 {
		for _, t := range tasksWithOrphanedEpicFixed {
			if err := database.UpdateTask(ctx, t); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to clear orphaned epic_id on task %s: %v\n", t.ID, err)
			}
		}

		if len(tasksWithOrphanedEpicFixed) > 0 {
			if syncErr := SyncTodo2Tasks(projectRoot); syncErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: sync after clearing orphaned epic_id failed: %v\n", syncErr)
			}
		}
	}

	syncResults := &proto.SyncResults{
		ValidatedTasks: int32(len(tasks)),
		IssuesFound:    int32(len(issues)),
		Issues:         issues,
		Synced:         !dryRun,
	}
	resp := &proto.TaskWorkflowResponse{
		Success:     true,
		Method:      "native_go",
		DryRun:      dryRun,
		SyncResults: syncResults,
	}

	result := TaskWorkflowResponseToMap(resp)
	if cast.ToBool(params["external"]) {
		if result["sync_results"] != nil {
			if m, ok := result["sync_results"].(map[string]interface{}); ok {
				m["external_sync_note"] = "External sync is a future nice-to-have; performed SQLite↔JSON sync only."
			}
		}
	}

	outputPath := cast.ToString(params["output_path"])

	return response.FormatResult(result, outputPath)
}

// ─── validTaskStatuses ──────────────────────────────────────────────────────
// Valid task statuses for sanity check.
var validTaskStatuses = map[string]bool{
	models.StatusTodo: true, models.StatusInProgress: true, models.StatusDone: true, models.StatusReview: true,
	strings.ToLower(models.StatusTodo): true, strings.ToLower(models.StatusInProgress): true, strings.ToLower(models.StatusDone): true, strings.ToLower(models.StatusReview): true,
}

// ─── handleTaskWorkflowSanityCheck ──────────────────────────────────────────
// handleTaskWorkflowSanityCheck runs generic Todo2 task sanity checks (epoch dates, empty content, valid status, duplicate IDs, missing deps).
// Use action=sanity_check; optional output_path to write report.
func handleTaskWorkflowSanityCheck(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	issues := []string{}
	taskMap := make(map[string]bool)

	for _, task := range tasks {
		// Invalid task ID (e.g. T-NaN from JS "T-" + NaN)
		if !models.IsValidTaskID(task.ID) {
			issues = append(issues, fmt.Sprintf("Task has invalid ID format: %q (expected T-<integer>)", task.ID))
		}

		// Duplicate ID
		if taskMap[task.ID] {
			issues = append(issues, fmt.Sprintf("Duplicate task ID: %s", task.ID))
		}

		taskMap[task.ID] = true

		// Epoch or invalid dates (1970, 0) for created/last_modified; completed_at only for Done tasks
		if models.IsEpochDate(task.CreatedAt) {
			issues = append(issues, fmt.Sprintf("Task %s has epoch/invalid created_at", task.ID))
		}

		if models.IsEpochDate(task.LastModified) {
			issues = append(issues, fmt.Sprintf("Task %s has epoch/invalid last_modified", task.ID))
		}

		if strings.EqualFold(task.Status, models.StatusDone) && models.IsEpochDate(task.CompletedAt) {
			issues = append(issues, fmt.Sprintf("Task %s (Done) has epoch/invalid completed_at", task.ID))
		}

		// Empty content/name (task title)
		if strings.TrimSpace(task.Content) == "" {
			issues = append(issues, fmt.Sprintf("Task %s has empty content/name", task.ID))
		}

		// Valid status
		norm := strings.TrimSpace(strings.ToLower(task.Status))
		if task.Status == "" || !validTaskStatuses[norm] {
			issues = append(issues, fmt.Sprintf("Task %s has invalid or empty status: %q", task.ID, task.Status))
		}
	}

	// Missing dependencies
	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if !taskMap[dep] {
				issues = append(issues, fmt.Sprintf("Task %s depends on %s which doesn't exist", task.ID, dep))
			}
		}
	}

	// Duplicate content (same name/description, different IDs)
	contentToIDs := make(map[string][]string)

	for _, task := range tasks {
		key := models.NormalizeForComparison(task.Content, task.LongDescription)
		if key == "" {
			continue
		}

		contentToIDs[key] = append(contentToIDs[key], task.ID)
	}

	for contentKey, ids := range contentToIDs {
		if len(ids) > 1 {
			// Truncate content for display
			preview := contentKey
			if len(preview) > 60 {
				preview = preview[:57] + "..."
			}

			issues = append(issues, fmt.Sprintf("Duplicate content (%q): tasks %v", preview, ids))
		}
	}

	passed := len(issues) == 0
	result := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"passed":        passed,
		"total_tasks":   len(tasks),
		"issues_found":  len(issues),
		"issues":        issues,
		"sanity_checks": []string{"invalid_task_id", "epoch_dates", "empty_content", "valid_status", "duplicate_ids", "duplicate_content", "missing_dependencies"},
	}

	outputPath := cast.ToString(params["output_path"])

	return response.FormatResult(result, outputPath)
}

// ─── handleTaskWorkflowClarity ──────────────────────────────────────────────
// handleTaskWorkflowClarity handles clarity action for improving task clarity.
func handleTaskWorkflowClarity(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	autoApply := false
	if apply, ok := params["auto_apply"].(bool); ok {
		autoApply = apply
	}

	outputFormat := "text"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	// Analyze tasks for clarity issues
	clarityIssues := []map[string]interface{}{}
	improvements := []map[string]interface{}{}

	for _, task := range tasks {
		if !IsPendingStatus(task.Status) {
			continue
		}

		issues := []string{}
		suggestions := []string{}

		// Check description length
		if task.LongDescription == "" {
			issues = append(issues, "missing_description")
			suggestions = append(suggestions, "Add a detailed description explaining what needs to be done")
		} else if len(task.LongDescription) < config.TaskMinDescriptionLength() {
			issues = append(issues, "brief_description")
			suggestions = append(suggestions, "Expand description with more details about requirements and acceptance criteria")
		}

		// Check for acceptance criteria keywords
		descLower := strings.ToLower(task.LongDescription)
		if !strings.Contains(descLower, "acceptance") && !strings.Contains(descLower, "criteria") {
			issues = append(issues, "missing_acceptance_criteria")
			suggestions = append(suggestions, "Add acceptance criteria to clarify when the task is complete")
		}

		// Check for scope boundaries
		if !strings.Contains(descLower, "scope") && !strings.Contains(descLower, "boundary") {
			issues = append(issues, "missing_scope")
			suggestions = append(suggestions, "Define scope boundaries to prevent scope creep")
		}

		if len(issues) > 0 {
			clarityIssues = append(clarityIssues, map[string]interface{}{
				"task_id":     task.ID,
				"content":     task.Content,
				"issues":      issues,
				"suggestions": suggestions,
			})

			if autoApply {
				// Auto-apply improvements (simplified - would need more sophisticated logic)
				improvements = append(improvements, map[string]interface{}{
					"task_id": task.ID,
					"applied": false, // Would apply in real implementation
				})
			}
		}
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "native_go",
		"total_tasks":     len(tasks),
		"tasks_analyzed":  len(clarityIssues),
		"clarity_issues":  clarityIssues,
		"auto_apply":      autoApply,
		"improvements":    improvements,
		"recommendations": buildClarityRecommendations(clarityIssues),
	}

	outputPath := cast.ToString(params["output_path"])

	// Handle text vs JSON formatting
	if outputFormat == "json" {
		// Use response.FormatResult for JSON format
		return response.FormatResult(result, outputPath)
	}

	// Text format - use custom formatter
	output := formatClarityAnalysisText(result)

	// Write to file if outputPath is provided
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err == nil {
			// File written successfully - note: can't add output_path to text output
			// but we can log it or include in a comment
		}
	}

	return []framework.TextContent{
		{Type: "text", Text: output},
	}, nil
}

// ─── isOldSequentialID ──────────────────────────────────────────────────────
// isOldSequentialID checks if a task ID uses the old sequential format (T-1, T-2, etc.)
// vs the new epoch format (T-1768158627000)
// Old format: T- followed by a small number (< 10000, typically 1-999)
// New format: T- followed by epoch milliseconds (13 digits, typically 1.6+ trillion).
func isOldSequentialID(taskID string) bool {
	if !strings.HasPrefix(taskID, "T-") {
		return false
	}

	// Extract the number part
	numStr := strings.TrimPrefix(taskID, "T-")

	// Parse as integer
	var num int64
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return false
	}

	// Old sequential IDs are typically small numbers (< 10000)
	// Epoch milliseconds are 13 digits (1.6+ trillion)
	// Use 1000000 (1 million) as the threshold to be safe
	return num < 1000000
}

// ─── handleTaskWorkflowCleanup ──────────────────────────────────────────────
// handleTaskWorkflowCleanup handles cleanup action for removing stale tasks and legacy tasks.
func handleTaskWorkflowCleanup(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	staleThresholdHours := 2.0
	if threshold, ok := params["stale_threshold_hours"].(float64); ok {
		staleThresholdHours = threshold
	}

	includeLegacy := false
	if legacy, ok := params["include_legacy"].(bool); ok {
		includeLegacy = legacy
	}

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	// Try database first for efficient filtering and deletion
	if db, err := database.GetDB(); err == nil && db != nil {
		// Get all pending tasks
		pendingStatus := models.StatusTodo
		filters := &database.TaskFilters{Status: &pendingStatus}

		tasks, err := database.ListTasks(context.Background(), filters)
		if err != nil {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}

		// Also get "In Progress" and "Review" tasks
		inProgressStatus := models.StatusInProgress
		filters.Status = &inProgressStatus
		inProgressTasks, _ := database.ListTasks(context.Background(), filters)
		tasks = append(tasks, inProgressTasks...)

		reviewStatus := models.StatusReview
		filters.Status = &reviewStatus
		reviewTasks, _ := database.ListTasks(context.Background(), filters)
		tasks = append(tasks, reviewTasks...)

		// Also get "Done" tasks if including legacy (legacy tasks might be marked Done)
		if includeLegacy {
			doneStatus := models.StatusDone
			filters.Status = &doneStatus
			doneTasks, _ := database.ListTasks(context.Background(), filters)
			tasks = append(tasks, doneTasks...)
		}

		// Identify stale and legacy tasks
		staleTasks := []*models.Todo2Task{}
		legacyTasks := []*models.Todo2Task{}

		for _, task := range tasks {
			// Check for legacy task ID (old sequential format)
			if includeLegacy && isOldSequentialID(task.ID) {
				legacyTasks = append(legacyTasks, task)
				continue
			}

			if !IsPendingStatus(task.Status) {
				continue
			}

			// Check for stale tag
			isStale := false

			for _, tag := range task.Tags {
				if strings.ToLower(tag) == "stale" {
					isStale = true
					break
				}
			}

			// Check metadata for last update time (if available)
			if !isStale && task.Metadata != nil {
				if lastUpdate, ok := task.Metadata["last_updated"].(string); ok {
					if strings.Contains(strings.ToLower(lastUpdate), "stale") {
						isStale = true
					}
				}
			}

			if isStale {
				staleTasks = append(staleTasks, task)
			}
		}

		// Combine stale and legacy tasks
		tasksToRemove := append(staleTasks, legacyTasks...)

		if dryRun {
			result := map[string]interface{}{
				"success":         true,
				"method":          "database",
				"dry_run":         true,
				"stale_count":     len(staleTasks),
				"stale_tasks":     formatStaleTasksFromPtrs(staleTasks),
				"legacy_count":    len(legacyTasks),
				"legacy_tasks":    formatStaleTasksFromPtrs(legacyTasks),
				"total_to_remove": len(tasksToRemove),
				"threshold_hours": staleThresholdHours,
				"include_legacy":  includeLegacy,
			}

			return response.FormatResult(result, "")
		}

		// Delete stale and legacy tasks from database
		removedIDs := []string{}

		for _, task := range tasksToRemove {
			if err := database.DeleteTask(context.Background(), task.ID); err == nil {
				removedIDs = append(removedIDs, task.ID)
			}
		}

		// Sync DB to JSON (shared workflow)
		var syncErr error

		if len(removedIDs) > 0 {
			if projectRoot, findErr := FindProjectRoot(); findErr == nil {
				syncErr = SyncTodo2Tasks(projectRoot)
				if syncErr != nil {
					fmt.Fprintf(os.Stderr, "Warning: sync DB to JSON after cleanup failed: %v\n", syncErr)
				}
			}
		}

		// Get remaining count
		remainingCount := len(tasks) - len(removedIDs)

		result := map[string]interface{}{
			"success":         true,
			"method":          "database",
			"removed_count":   len(removedIDs),
			"stale_removed":   len(staleTasks),
			"legacy_removed":  len(legacyTasks),
			"remaining_count": remainingCount,
			"removed_tasks":   removedIDs,
			"threshold_hours": staleThresholdHours,
			"include_legacy":  includeLegacy,
		}
		if syncErr != nil {
			result["sync_error"] = syncErr.Error()
		}

		outputPath := cast.ToString(params["output_path"])

		return response.FormatResult(result, outputPath)
	}

	// Fallback to TaskStore
	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task store: %w", err)
	}

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	// Identify stale and legacy tasks
	staleTasks := []Todo2Task{}
	legacyTasks := []Todo2Task{}

	for _, task := range tasks {
		// Check for legacy task ID (old sequential format)
		if includeLegacy && isOldSequentialID(task.ID) {
			legacyTasks = append(legacyTasks, task)
			continue
		}

		if IsPendingStatus(task.Status) {
			// Check for stale tag
			isStale := false

			for _, tag := range task.Tags {
				if strings.ToLower(tag) == "stale" {
					isStale = true
					break
				}
			}

			// Check metadata for last update time (if available)
			if !isStale && task.Metadata != nil {
				if lastUpdate, ok := task.Metadata["last_updated"].(string); ok {
					if strings.Contains(strings.ToLower(lastUpdate), "stale") {
						isStale = true
					}
				}
			}

			if isStale {
				staleTasks = append(staleTasks, task)
			}
		}
	}

	// Combine stale and legacy tasks
	tasksToRemove := append(staleTasks, legacyTasks...)

	if dryRun {
		result := map[string]interface{}{
			"success":         true,
			"method":          "file",
			"dry_run":         true,
			"stale_count":     len(staleTasks),
			"stale_tasks":     formatStaleTasks(staleTasks),
			"legacy_count":    len(legacyTasks),
			"legacy_tasks":    formatStaleTasks(legacyTasks),
			"total_to_remove": len(tasksToRemove),
			"threshold_hours": staleThresholdHours,
			"include_legacy":  includeLegacy,
		}

		return response.FormatResult(result, "")
	}

	// Remove stale and legacy tasks
	remainingTasks := []Todo2Task{}

	removeMap := make(map[string]bool)
	for _, taskToRemove := range tasksToRemove {
		removeMap[taskToRemove.ID] = true
	}

	removedIDs := []string{}

	for _, task := range tasks {
		if removeMap[task.ID] {
			removedIDs = append(removedIDs, task.ID)
		} else {
			remainingTasks = append(remainingTasks, task)
		}
	}

	// Delete removed tasks via store
	for _, id := range removedIDs {
		if err := store.DeleteTask(ctx, id); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete task %s: %v\n", id, err)
		}
	}

	result := map[string]interface{}{
		"success":         true,
		"method":          "file",
		"removed_count":   len(removedIDs),
		"stale_removed":   len(staleTasks),
		"legacy_removed":  len(legacyTasks),
		"remaining_count": len(remainingTasks),
		"removed_tasks":   removedIDs,
		"threshold_hours": staleThresholdHours,
		"include_legacy":  includeLegacy,
	}

	outputPath := cast.ToString(params["output_path"])

	return response.FormatResult(result, outputPath)
}

