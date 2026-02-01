package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// handleTaskWorkflowApprove handles approve action for batch approving tasks
// Uses database for efficient updates, falls back to file-based approach if needed
// This is platform-agnostic (doesn't require Apple FM)
func handleTaskWorkflowApprove(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
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

	// Try database first for efficient filtering and updates
	if db, err := database.GetDB(); err == nil && db != nil {
		// Build filters
		filters := &database.TaskFilters{Status: &status}
		if filterTag != "" {
			filters.Tag = &filterTag
		}

		// Get tasks matching filters
		allTasks, err := database.ListTasks(context.Background(), filters)
		if err != nil {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}

		// Filter candidates
		candidates := []*models.Todo2Task{}
		for _, task := range allTasks {
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

			// Filter by clarification requirement if needed
			if clarificationNone {
				minDescLen := config.TaskMinDescriptionLength()
				needsClarification := task.LongDescription == "" || len(task.LongDescription) < minDescLen
				if needsClarification {
					continue
				}
			}

			candidates = append(candidates, task)
		}

		if dryRun {
			taskList := make([]map[string]interface{}, len(candidates))
			for i, task := range candidates {
				taskList[i] = map[string]interface{}{
					"id":      task.ID,
					"content": task.Content,
					"status":  task.Status,
				}
			}

			taskIDList := make([]string, len(candidates))
			for i, task := range candidates {
				taskIDList[i] = task.ID
			}

			result := map[string]interface{}{
				"success":        true,
				"method":         "database",
				"dry_run":        true,
				"approved_count": len(candidates),
				"task_ids":       taskIDList,
				"tasks":          taskList,
			}

			return response.FormatResult(result, "")
		}

		// Update tasks in database
		approvedIDs := []string{}
		updatedCount := 0
		for _, task := range candidates {
			task.Status = newStatus
			if err := database.UpdateTask(context.Background(), task); err == nil {
				approvedIDs = append(approvedIDs, task.ID)
				updatedCount++
			}
		}

		// Sync DB to JSON so CLI and MCP leave Todo2 in sync (shared workflow)
		if updatedCount > 0 {
			if projectRoot, syncErr := FindProjectRoot(); syncErr == nil {
				_ = SyncTodo2Tasks(projectRoot)
			}
		}

		result := map[string]interface{}{
			"success":        true,
			"method":         "database",
			"approved_count": updatedCount,
			"task_ids":       approvedIDs,
		}

		return response.FormatResult(result, "")
	}

	// Fallback to file-based approach
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return handleTaskWorkflowApproveMCP(ctx, params)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return handleTaskWorkflowApproveMCP(ctx, params)
	}

	// Filter tasks to approve
	candidates := []*models.Todo2Task{}
	for _, task := range tasks {
		if normalizeStatus(task.Status) != status {
			continue
		}
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
		if clarificationNone {
			minDescLen := config.TaskMinDescriptionLength()
			needsClarification := task.LongDescription == "" || len(task.LongDescription) < minDescLen
			if needsClarification {
				continue
			}
		}
		candidates = append(candidates, &task)
	}

	if dryRun {
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
			"method":         "file",
			"dry_run":        true,
			"approved_count": len(candidates),
			"task_ids":       extractTaskIDs(candidates),
			"tasks":          taskList,
		}
		return response.FormatResult(result, "")
	}

	// Update tasks (file-based)
	updatedCount := 0
	for _, task := range candidates {
		task.Status = newStatus
		// Find index in tasks and update
		for i := range tasks {
			if tasks[i].ID == task.ID {
				tasks[i].Status = newStatus
				if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
					return nil, fmt.Errorf("failed to save tasks: %w", err)
				}
				updatedCount++
				break
			}
		}
	}

	result := map[string]interface{}{
		"success":        true,
		"method":         "file",
		"approved_count": updatedCount,
		"task_ids":       extractTaskIDs(candidates),
	}
	return response.FormatResult(result, "")
}

// handleTaskWorkflowUpdate updates task(s) by ID with optional new_status and/or priority.
// Uses database first; syncs DB to JSON after update.
// Params: task_ids (required, array or JSON string), new_status (optional), priority (optional).
func handleTaskWorkflowUpdate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	var taskIDs []string
	if ids, ok := params["task_ids"].(string); ok && ids != "" {
		if err := json.Unmarshal([]byte(ids), &taskIDs); err != nil {
			taskIDs = strings.Split(ids, ",")
			for i := range taskIDs {
				taskIDs[i] = strings.TrimSpace(taskIDs[i])
			}
		}
	} else if idsList, ok := params["task_ids"].([]interface{}); ok {
		taskIDs = make([]string, 0, len(idsList))
		for _, id := range idsList {
			if idStr, ok := id.(string); ok {
				taskIDs = append(taskIDs, idStr)
			}
		}
	}
	if len(taskIDs) == 0 {
		return nil, fmt.Errorf("update action requires task_ids")
	}

	newStatus, _ := params["new_status"].(string)
	if newStatus != "" {
		newStatus = normalizeStatus(newStatus)
	}
	priority, _ := params["priority"].(string)
	if priority != "" {
		priority = normalizePriority(priority)
	}
	if newStatus == "" && priority == "" {
		return nil, fmt.Errorf("update action requires at least one of new_status or priority")
	}

	// Try database first
	if db, err := database.GetDB(); err == nil && db != nil {
		updatedIDs := []string{}
		updatedCount := 0
		for _, id := range taskIDs {
			task, err := database.GetTask(ctx, id)
			if err != nil {
				continue
			}
			if newStatus != "" {
				task.Status = newStatus
			}
			if priority != "" {
				task.Priority = priority
			}
			if err := database.UpdateTask(ctx, task); err != nil {
				continue
			}
			updatedIDs = append(updatedIDs, id)
			updatedCount++
		}

		// Sync DB to JSON
		projectRoot, syncErr := FindProjectRoot()
		if syncErr == nil {
			_ = SyncTodo2Tasks(projectRoot)
		}

		result := map[string]interface{}{
			"success":       true,
			"method":        "database",
			"updated_count": updatedCount,
			"task_ids":      updatedIDs,
		}
		return response.FormatResult(result, "")
	}

	// Fallback: file-based load, update, save
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	taskMap := make(map[string]*models.Todo2Task)
	for i := range tasks {
		taskMap[tasks[i].ID] = &tasks[i]
	}
	updatedIDs := []string{}
	for _, id := range taskIDs {
		t, ok := taskMap[id]
		if !ok {
			continue
		}
		if newStatus != "" {
			t.Status = newStatus
		}
		if priority != "" {
			t.Priority = priority
		}
		updatedIDs = append(updatedIDs, id)
	}
	if len(updatedIDs) == 0 {
		result := map[string]interface{}{
			"success":       true,
			"method":        "file",
			"updated_count": 0,
			"task_ids":      []string{},
		}
		return response.FormatResult(result, "")
	}
	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		return nil, fmt.Errorf("failed to save tasks: %w", err)
	}
	result := map[string]interface{}{
		"success":       true,
		"method":        "file",
		"updated_count": len(updatedIDs),
		"task_ids":      updatedIDs,
	}
	return response.FormatResult(result, "")
}

// extractTaskIDs returns task IDs from a slice of *Todo2Task (for approve file fallback).
func extractTaskIDs(candidates []*models.Todo2Task) []string {
	ids := make([]string, len(candidates))
	for i, t := range candidates {
		ids[i] = t.ID
	}
	return ids
}

// handleTaskWorkflowList is moved below to avoid duplicating the file-based approve loop.
func handleTaskWorkflowListPlaceholder(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	tasks, err := LoadTodo2Tasks(ctx.Value("projectRoot").(string))
	if err != nil {
		return nil, err
	}
	_ = params
	_ = tasks
	return nil, nil
}

var _ = handleTaskWorkflowListPlaceholder // avoid unused; real list is handleTaskWorkflowList

// handleTaskWorkflowApproveMCP returns an error when project root or task load fails (no bridge).
func handleTaskWorkflowApproveMCP(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	return nil, fmt.Errorf("approve action: project root or task load failed; cannot approve tasks")
}

// handleTaskWorkflowList handles list sub-action for displaying tasks
func handleTaskWorkflowList(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// Load tasks
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Apply filters
	var status, priority, filterTag, taskID string
	var limit int

	if s, ok := params["status"].(string); ok {
		status = s
	}
	if p, ok := params["priority"].(string); ok {
		priority = p
	}
	if tag, ok := params["filter_tag"].(string); ok {
		filterTag = tag
	}
	if tid, ok := params["task_id"].(string); ok {
		taskID = tid
	}
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	} else if l, ok := params["limit"].(int); ok {
		limit = l
	}

	// Filter tasks
	filtered := []Todo2Task{}
	for _, task := range tasks {
		if taskID != "" && task.ID != taskID {
			continue
		}
		if status != "" && task.Status != status {
			continue
		}
		if priority != "" && task.Priority != priority {
			continue
		}
		if filterTag != "" {
			found := false
			for _, tag := range task.Tags {
				if tag == filterTag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		filtered = append(filtered, task)
	}

	// Optional: sort by execution order (dependency order)
	if order, _ := params["order"].(string); order == "execution" || order == "dependency" {
		orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
		if err == nil {
			filteredMap := make(map[string]Todo2Task)
			for _, t := range filtered {
				filteredMap[t.ID] = t
			}
			orderedSet := make(map[string]bool)
			for _, id := range orderedIDs {
				orderedSet[id] = true
			}
			orderedFiltered := make([]Todo2Task, 0, len(filtered))
			for _, id := range orderedIDs {
				if t, ok := filteredMap[id]; ok {
					orderedFiltered = append(orderedFiltered, t)
				}
			}
			for _, t := range filtered {
				if !orderedSet[t.ID] {
					orderedFiltered = append(orderedFiltered, t)
				}
			}
			filtered = orderedFiltered
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Format output
	outputFormat, _ := params["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "text"
	}

	if outputFormat == "json" {
		return response.FormatResult(map[string]interface{}{"tasks": filtered}, "")
	}
	// Text format
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tasks (%d total, %d shown)\n", len(tasks), len(filtered)))
	sb.WriteString(strings.Repeat("=", 80) + "\n")
	sb.WriteString(fmt.Sprintf("%-8s | %-15s | %-10s | %s\n", "ID", "Status", "Priority", "Content"))
	sb.WriteString(strings.Repeat("-", 80) + "\n")
	for _, task := range filtered {
		content := task.Content
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-8s | %-15s | %-10s | %s\n", task.ID, task.Status, task.Priority, content))
	}
	return []framework.TextContent{
		{Type: "text", Text: sb.String()},
	}, nil
}

// handleTaskWorkflowSync handles sync action for synchronizing tasks between SQLite and JSON.
// The "external" param (sync with external sources, e.g. infer_task_progress) is a future nice-to-have; if passed, it is ignored and SQLite↔JSON sync is performed.
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
	subAction, _ := params["sub_action"].(string)
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
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks after sync: %w", err)
	}

	// Validate task consistency
	issues := []string{}
	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	// Check for missing dependencies
	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if !taskMap[dep] {
				issues = append(issues, fmt.Sprintf("Task %s depends on %s which doesn't exist", task.ID, dep))
			}
		}

		// Validate planning document links
		if linkMeta := GetPlanningLinkMetadata(&task); linkMeta != nil {
			if linkMeta.PlanningDoc != "" {
				if err := ValidatePlanningLink(projectRoot, linkMeta.PlanningDoc); err != nil {
					issues = append(issues, fmt.Sprintf("Task %s has invalid planning doc link: %v", task.ID, err))
				}
			}
			if linkMeta.EpicID != "" {
				if err := ValidateTaskReference(linkMeta.EpicID, tasks); err != nil {
					issues = append(issues, fmt.Sprintf("Task %s has invalid epic ID: %v", task.ID, err))
				}
			}
		}
	}

	syncResults := map[string]interface{}{
		"validated_tasks": len(tasks),
		"issues_found":    len(issues),
		"issues":          issues,
		"synced":          !dryRun,
	}
	if external, _ := params["external"].(bool); external {
		syncResults["external_sync_note"] = "External sync is a future nice-to-have; performed SQLite↔JSON sync only."
	}

	result := map[string]interface{}{
		"success":      true,
		"method":       "native_go",
		"dry_run":      dryRun,
		"sync_results": syncResults,
	}

	outputPath, _ := params["output_path"].(string)
	return response.FormatResult(result, outputPath)
}

// Valid task statuses for sanity check.
var validTaskStatuses = map[string]bool{
	"Todo": true, "In Progress": true, "Done": true, "Review": true,
	"todo": true, "in progress": true, "done": true, "review": true,
}

// handleTaskWorkflowSanityCheck runs generic Todo2 task sanity checks (epoch dates, empty content, valid status, duplicate IDs, missing deps).
// Use action=sanity_check; optional output_path to write report.
func handleTaskWorkflowSanityCheck(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	issues := []string{}
	taskMap := make(map[string]bool)

	for _, task := range tasks {
		// Invalid task ID (e.g. T-NaN from JS "T-" + NaN)
		if !isValidTaskID(task.ID) {
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
		if strings.TrimSpace(strings.ToLower(task.Status)) == "done" && models.IsEpochDate(task.CompletedAt) {
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
		key := normalizeTaskContent(task.Content, task.LongDescription)
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

	outputPath, _ := params["output_path"].(string)
	return response.FormatResult(result, outputPath)
}

// handleTaskWorkflowClarity handles clarity action for improving task clarity
func handleTaskWorkflowClarity(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

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

	outputPath, _ := params["output_path"].(string)

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

// isValidTaskID returns true unless the task ID looks like a malformed T- ID (e.g. T-NaN).
// IDs with other prefixes (AFM-, AUTO-, etc.) are allowed. Only T-<non-integer> is invalid.
func isValidTaskID(taskID string) bool {
	if !strings.HasPrefix(taskID, "T-") {
		return true
	}
	numStr := strings.TrimPrefix(taskID, "T-")
	var num int64
	if _, err := fmt.Sscanf(numStr, "%d", &num); err != nil {
		return false
	}
	return num > 0
}

// isOldSequentialID checks if a task ID uses the old sequential format (T-1, T-2, etc.)
// vs the new epoch format (T-1768158627000)
// Old format: T- followed by a small number (< 10000, typically 1-999)
// New format: T- followed by epoch milliseconds (13 digits, typically 1.6+ trillion)
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

// handleTaskWorkflowCleanup handles cleanup action for removing stale tasks and legacy tasks
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
		pendingStatus := "Todo"
		filters := &database.TaskFilters{Status: &pendingStatus}
		tasks, err := database.ListTasks(context.Background(), filters)
		if err != nil {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}

		// Also get "In Progress" and "Review" tasks
		inProgressStatus := "In Progress"
		filters.Status = &inProgressStatus
		inProgressTasks, _ := database.ListTasks(context.Background(), filters)
		tasks = append(tasks, inProgressTasks...)

		reviewStatus := "Review"
		filters.Status = &reviewStatus
		reviewTasks, _ := database.ListTasks(context.Background(), filters)
		tasks = append(tasks, reviewTasks...)

		// Also get "Done" tasks if including legacy (legacy tasks might be marked Done)
		if includeLegacy {
			doneStatus := "Done"
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
		if len(removedIDs) > 0 {
			if projectRoot, syncErr := FindProjectRoot(); syncErr == nil {
				_ = SyncTodo2Tasks(projectRoot)
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

		outputPath, _ := params["output_path"].(string)
		return response.FormatResult(result, outputPath)
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

	// Save updated tasks
	if err := SaveTodo2Tasks(projectRoot, remainingTasks); err != nil {
		return nil, fmt.Errorf("failed to save tasks: %w", err)
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

	outputPath, _ := params["output_path"].(string)
	return response.FormatResult(result, outputPath)
}

// handleTaskWorkflowFixEmptyDescriptions sets long_description from content for tasks with empty long_description, then syncs to JSON.
func handleTaskWorkflowFixEmptyDescriptions(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	if db, err := database.GetDB(); err == nil && db != nil {
		tasks, err := database.ListTasks(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}

		var toUpdate []*models.Todo2Task
		for _, task := range tasks {
			if strings.TrimSpace(task.LongDescription) == "" && task.Content != "" {
				t := *task
				t.LongDescription = task.Content
				toUpdate = append(toUpdate, &t)
			}
		}

		if dryRun {
			ids := make([]string, len(toUpdate))
			for i, t := range toUpdate {
				ids[i] = t.ID
			}
			result := map[string]interface{}{
				"success":       true,
				"method":        "database",
				"dry_run":       true,
				"tasks_updated": len(toUpdate),
				"task_ids":      ids,
			}
			return response.FormatResult(result, "")
		}

		updatedIDs := []string{}
		for _, task := range toUpdate {
			if err := database.UpdateTask(ctx, task); err == nil {
				updatedIDs = append(updatedIDs, task.ID)
			}
		}

		if len(updatedIDs) > 0 {
			if projectRoot, syncErr := FindProjectRoot(); syncErr == nil {
				_ = SyncTodo2Tasks(projectRoot)
			}
		}

		result := map[string]interface{}{
			"success":       true,
			"method":        "database",
			"tasks_updated": len(updatedIDs),
			"task_ids":      updatedIDs,
		}
		outputPath, _ := params["output_path"].(string)
		return response.FormatResult(result, outputPath)
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	if dryRun {
		var ids []string
		for _, t := range tasks {
			if strings.TrimSpace(t.LongDescription) == "" && t.Content != "" {
				ids = append(ids, t.ID)
			}
		}
		result := map[string]interface{}{
			"success":       true,
			"method":        "file",
			"dry_run":       true,
			"tasks_updated": len(ids),
			"task_ids":      ids,
		}
		return response.FormatResult(result, "")
	}

	var updated int
	for i := range tasks {
		if strings.TrimSpace(tasks[i].LongDescription) == "" && tasks[i].Content != "" {
			tasks[i].LongDescription = tasks[i].Content
			updated++
		}
	}

	if updated > 0 {
		if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
			return nil, fmt.Errorf("failed to save tasks: %w", err)
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"method":        "file",
		"tasks_updated": updated,
	}
	outputPath, _ := params["output_path"].(string)
	return response.FormatResult(result, outputPath)
}

// Helper functions for clarity action

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

// allowedStatusForLinkPlanning restricts link_planning to Todo and In Progress only.
var allowedStatusForLinkPlanning = map[string]bool{
	"Todo":        true,
	"In Progress": true,
}

// handleTaskWorkflowLinkPlanning sets planning_doc and/or epic_id on existing tasks.
// Only tasks with status Todo or In Progress are updated.
func handleTaskWorkflowLinkPlanning(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	planningDoc, _ := params["planning_doc"].(string)
	epicID, _ := params["epic_id"].(string)
	if planningDoc == "" && epicID == "" {
		return nil, fmt.Errorf("at least one of planning_doc or epic_id is required for link_planning")
	}

	if planningDoc != "" {
		if err := ValidatePlanningLink(projectRoot, planningDoc); err != nil {
			return nil, fmt.Errorf("invalid planning_doc: %w", err)
		}
	}

	// Resolve task IDs: task_id (single) or task_ids (comma or JSON array)
	var uniqueIDs []string
	seen := make(map[string]bool)
	if tid, ok := params["task_id"].(string); ok && tid != "" {
		tid = strings.TrimSpace(tid)
		if !seen[tid] {
			seen[tid] = true
			uniqueIDs = append(uniqueIDs, tid)
		}
	}
	if ids, ok := params["task_ids"].(string); ok && ids != "" {
		var parsed []string
		if json.Unmarshal([]byte(ids), &parsed) == nil {
			for _, id := range parsed {
				id = strings.TrimSpace(id)
				if id != "" && !seen[id] {
					seen[id] = true
					uniqueIDs = append(uniqueIDs, id)
				}
			}
		} else {
			for _, id := range strings.Split(ids, ",") {
				id = strings.TrimSpace(id)
				if id != "" && !seen[id] {
					seen[id] = true
					uniqueIDs = append(uniqueIDs, id)
				}
			}
		}
	}
	if len(uniqueIDs) == 0 {
		return nil, fmt.Errorf("task_id or task_ids is required for link_planning")
	}

	// Load all tasks for lookup and (if no DB) for file-based save
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	taskMap := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	if epicID != "" {
		if err := ValidateTaskReference(epicID, tasks); err != nil {
			return nil, fmt.Errorf("invalid epic_id: %w", err)
		}
	}

	updatedIDs := make([]string, 0)
	var skippedStatus []string
	db, dbErr := database.GetDB()
	useDB := dbErr == nil && db != nil

	for _, id := range uniqueIDs {
		task, ok := taskMap[id]
		if !ok {
			skippedStatus = append(skippedStatus, id+": not found")
			continue
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
		}
		SetPlanningLinkMetadata(&task, linkMeta)

		if useDB {
			if err := database.UpdateTask(ctx, &task); err != nil {
				skippedStatus = append(skippedStatus, id+": update failed: "+err.Error())
				continue
			}
		} else {
			for i := range tasks {
				if tasks[i].ID == id {
					tasks[i].Metadata = task.Metadata
					break
				}
			}
		}
		updatedIDs = append(updatedIDs, id)
	}

	if useDB && len(updatedIDs) > 0 {
		if projectRoot, syncErr := FindProjectRoot(); syncErr == nil {
			_ = SyncTodo2Tasks(projectRoot)
		}
	}
	if !useDB && len(updatedIDs) > 0 {
		if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
			return nil, fmt.Errorf("failed to save tasks: %w", err)
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"method":        "native_go",
		"action":        "link_planning",
		"updated_count": len(updatedIDs),
		"updated_ids":   updatedIDs,
		"skipped":       skippedStatus,
	}
	return response.FormatResult(result, "")
}

// handleTaskWorkflowCreate handles create action for creating new tasks
// Uses database for efficient creation, falls back to file-based approach if needed
// This is platform-agnostic (doesn't require Apple FM)
func handleTaskWorkflowCreate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// Extract required parameters
	name, _ := params["name"].(string)
	if name == "" {
		return nil, fmt.Errorf("name is required for task creation")
	}

	longDescription, _ := params["long_description"].(string)
	if longDescription == "" {
		longDescription = name // Use name as fallback
	}

	// Extract optional parameters - use config defaults
	status := config.DefaultTaskStatus()
	if s, ok := params["status"].(string); ok && s != "" {
		status = normalizeStatus(s)
	}

	priority := config.DefaultTaskPriority()
	if p, ok := params["priority"].(string); ok && p != "" {
		priority = normalizePriority(p)
	}

	// Use config default tags, allow override from params
	tags := config.DefaultTaskTags()
	if len(tags) == 0 {
		tags = []string{} // Ensure it's a slice, not nil
	}
	if t, ok := params["tags"].([]interface{}); ok {
		for _, tag := range t {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	} else if tStr, ok := params["tags"].(string); ok && tStr != "" {
		// Support comma-separated tags string
		tagList := strings.Split(tStr, ",")
		for _, tag := range tagList {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	dependencies := []string{}
	if d, ok := params["dependencies"].([]interface{}); ok {
		for _, dep := range d {
			if depStr, ok := dep.(string); ok {
				dependencies = append(dependencies, depStr)
			}
		}
	} else if dStr, ok := params["dependencies"].(string); ok && dStr != "" {
		// Support comma-separated dependencies string
		depList := strings.Split(dStr, ",")
		for _, dep := range depList {
			dep = strings.TrimSpace(dep)
			if dep != "" {
				dependencies = append(dependencies, dep)
			}
		}
	}

	// Load existing tasks for dependency validation only (ID generation is now O(1))
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Generate next task ID using epoch milliseconds (O(1) - no need to load all tasks)
	nextID := generateEpochTaskID()

	// Validate dependencies exist
	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true
	}
	for _, dep := range dependencies {
		if !taskMap[dep] {
			return nil, fmt.Errorf("dependency %s does not exist", dep)
		}
	}

	// Extract planning document link (optional)
	var planningDoc string
	if pd, ok := params["planning_doc"].(string); ok && pd != "" {
		planningDoc = pd
		// Validate planning document link
		if err := ValidatePlanningLink(projectRoot, planningDoc); err != nil {
			return nil, fmt.Errorf("invalid planning document link: %w", err)
		}
	}

	var epicID string
	if eid, ok := params["epic_id"].(string); ok && eid != "" {
		epicID = eid
		// Validate epic ID exists
		if err := ValidateTaskReference(epicID, tasks); err != nil {
			return nil, fmt.Errorf("invalid epic ID: %w", err)
		}
	}

	// Create task
	task := &models.Todo2Task{
		ID:              nextID,
		Content:         name,
		LongDescription: longDescription,
		Status:          status,
		Priority:        priority,
		Tags:            tags,
		Dependencies:    dependencies,
		Completed:       false,
		Metadata:        make(map[string]interface{}),
	}

	// Store planning document link in metadata if provided
	if planningDoc != "" || epicID != "" {
		linkMeta := &PlanningLinkMetadata{
			PlanningDoc: planningDoc,
			EpicID:      epicID,
		}
		SetPlanningLinkMetadata(task, linkMeta)
	}

	// Try database first
	if err := database.CreateTask(ctx, task); err != nil {
		// Database failed, try file-based fallback
		tasks = append(tasks, *task)
		if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
			return nil, fmt.Errorf("failed to create task: database error: %v, file error: %w", err, err)
		}
	} else {
		// Sync DB to JSON (shared workflow)
		_ = SyncTodo2Tasks(projectRoot)
	}

	// Return created task information
	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"task": map[string]interface{}{
			"id":               task.ID,
			"name":             task.Content,
			"long_description": task.LongDescription,
			"status":           task.Status,
			"priority":         task.Priority,
			"tags":             task.Tags,
			"dependencies":     task.Dependencies,
		},
	}

	// Auto-estimate task if enabled (default: true)
	// This happens after task creation succeeds, so failures don't affect task creation
	autoEstimate := true
	if ae, ok := params["auto_estimate"].(bool); ok {
		autoEstimate = ae
	}

	if autoEstimate {
		// Estimate task duration and add as comment (graceful failure - don't fail task creation)
		if err := addEstimateComment(ctx, projectRoot, task, name, longDescription, tags, priority); err != nil {
			// Log error but don't fail task creation
			// Error is logged to result metadata for debugging
			if metadata, ok := result["metadata"].(map[string]interface{}); ok {
				metadata["estimation_error"] = err.Error()
			} else {
				result["metadata"] = map[string]interface{}{
					"estimation_error": err.Error(),
				}
			}
		} else {
			// Add estimation success indicator
			if metadata, ok := result["metadata"].(map[string]interface{}); ok {
				metadata["estimation_added"] = true
			} else {
				result["metadata"] = map[string]interface{}{
					"estimation_added": true,
				}
			}
		}
	}

	outputPath, _ := params["output_path"].(string)
	return response.FormatResult(result, outputPath)
}

// handleTaskWorkflowFixInvalidIDs finds tasks with invalid IDs (e.g. T-NaN from JS), assigns new epoch IDs,
// updates dependencies, removes old DB rows, and saves. Use after sanity_check reports invalid_task_id.
func handleTaskWorkflowFixInvalidIDs(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	type fix struct{ oldID, newID string }
	var fixes []fix
	for i := range tasks {
		if !isValidTaskID(tasks[i].ID) {
			oldID := tasks[i].ID
			newID := generateEpochTaskID()
			tasks[i].ID = newID
			fixes = append(fixes, fix{oldID, newID})
		}
	}

	if len(fixes) == 0 {
		result := map[string]interface{}{
			"success": true, "method": "native_go",
			"fixed_count": 0,
			"message":     "no invalid task IDs found",
		}
		return response.FormatResult(result, "")
	}

	// Update dependencies: any task referencing an old ID should reference the new ID
	oldToNew := make(map[string]string)
	for _, f := range fixes {
		oldToNew[f.oldID] = f.newID
	}
	for i := range tasks {
		for j, dep := range tasks[i].Dependencies {
			if newID, ok := oldToNew[dep]; ok {
				tasks[i].Dependencies[j] = newID
			}
		}
	}

	// Remove old rows from DB so SaveTodo2Tasks will create new rows with new IDs
	if db, err := database.GetDB(); err == nil && db != nil {
		for _, f := range fixes {
			_ = database.DeleteTask(ctx, f.oldID)
		}
	}

	if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
		return nil, fmt.Errorf("failed to save after fixing IDs: %w", err)
	}

	// Regenerate overview so .cursor/rules/todo2-overview.mdc reflects new IDs
	if overviewErr := WriteTodo2Overview(projectRoot); overviewErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write todo2-overview.mdc: %v\n", overviewErr)
	}

	mapping := make(map[string]string)
	for _, f := range fixes {
		mapping[f.oldID] = f.newID
	}
	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"fixed_count": len(fixes),
		"id_mapping":  mapping,
	}
	return response.FormatResult(result, "")
}

// generateEpochTaskID generates a task ID using epoch milliseconds (T-{epoch_milliseconds})
// This is O(1) and doesn't require loading all tasks, solving the performance bottleneck
// Format: T-{epoch_milliseconds} (e.g., T-1768158627000)
func generateEpochTaskID() string {
	epochMillis := time.Now().UnixMilli()
	return fmt.Sprintf("T-%d", epochMillis)
}

// normalizePriority normalizes priority values to canonical lowercase form.
// This is a wrapper around NormalizePriority for backward compatibility.
func normalizePriority(priority string) string {
	return NormalizePriority(priority)
}

// addEstimateComment estimates task duration and adds it as a comment
// This is called after task creation succeeds, and failures are handled gracefully
func addEstimateComment(ctx context.Context, projectRoot string, task *models.Todo2Task, name, details string, tags []string, priority string) error {
	// Call estimation tool
	estimationParams := map[string]interface{}{
		"action":         "estimate",
		"name":           name,
		"details":        details,
		"tag_list":       tags,
		"priority":       priority,
		"use_historical": true,
		"detailed":       false,
	}

	// Try native estimation first (platform-agnostic, will use statistical if Apple FM unavailable)
	estimationResult, err := handleEstimationNative(ctx, projectRoot, estimationParams)
	if err != nil {
		return fmt.Errorf("estimation failed: %w", err)
	}

	// Parse estimation result
	var estimate EstimationResult
	if err := json.Unmarshal([]byte(estimationResult), &estimate); err != nil {
		return fmt.Errorf("failed to parse estimation result: %w", err)
	}

	// Format estimate as markdown comment
	commentContent := formatEstimateComment(estimate)

	// Create comment
	comment := database.Comment{
		TaskID:  task.ID,
		Type:    "note",
		Content: commentContent,
	}

	// Add comment via database
	if err := database.AddComments(ctx, task.ID, []database.Comment{comment}); err != nil {
		// Database failed, try file-based fallback (future enhancement)
		// For now, just return error (task creation already succeeded)
		return fmt.Errorf("failed to add estimate comment: %w", err)
	}

	return nil
}

// formatEstimateComment formats estimation result as a markdown comment
func formatEstimateComment(estimate EstimationResult) string {
	var builder strings.Builder
	builder.WriteString("## Task Duration Estimate\n\n")
	builder.WriteString(fmt.Sprintf("**Estimated Duration:** %.1f hours\n\n", estimate.EstimateHours))
	builder.WriteString(fmt.Sprintf("**Confidence:** %.0f%%\n\n", estimate.Confidence*100))
	builder.WriteString(fmt.Sprintf("**Method:** %s\n", estimate.Method))

	if estimate.LowerBound > 0 && estimate.UpperBound > 0 {
		builder.WriteString(fmt.Sprintf("\n**Range:** %.1f - %.1f hours\n", estimate.LowerBound, estimate.UpperBound))
	}

	if len(estimate.Metadata) > 0 {
		if statisticalEst, ok := estimate.Metadata["statistical_estimate"].(float64); ok {
			builder.WriteString(fmt.Sprintf("\n**Statistical Estimate:** %.1f hours\n", statisticalEst))
		}
		if appleFMEst, ok := estimate.Metadata["apple_fm_estimate"].(float64); ok {
			builder.WriteString(fmt.Sprintf("**AI Estimate:** %.1f hours\n", appleFMEst))
		}
	}

	return builder.String()
}
