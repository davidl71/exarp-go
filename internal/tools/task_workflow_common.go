package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
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
				needsClarification := task.LongDescription == "" || len(task.LongDescription) < 50
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

			output, _ := json.MarshalIndent(result, "", "  ")
			return []framework.TextContent{
				{Type: "text", Text: string(output)},
			}, nil
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

		result := map[string]interface{}{
			"success":        true,
			"method":         "database",
			"approved_count": updatedCount,
			"task_ids":       approvedIDs,
		}

		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
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
			needsClarification := task.LongDescription == "" || len(task.LongDescription) < 50
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

		result := map[string]interface{}{
			"success":        true,
			"method":         "file",
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
		return handleTaskWorkflowApproveMCP(ctx, params)
	}

	result := map[string]interface{}{
		"success":        true,
		"method":         "file",
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

// handleTaskWorkflowSync handles sync action for synchronizing tasks across systems
func handleTaskWorkflowSync(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	dryRun := false
	if dr, ok := params["dry_run"].(bool); ok {
		dryRun = dr
	}

	// For now, sync is a simplified implementation that validates task consistency
	// Full sync with external systems would require additional integration
	syncResults := map[string]interface{}{
		"validated_tasks": len(tasks),
		"issues_found":    0,
		"issues":          []string{},
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
	}

	syncResults["issues_found"] = len(issues)
	syncResults["issues"] = issues

	if !dryRun && len(issues) > 0 {
		// Could auto-fix some issues here
		// For now, just report them
	}

	result := map[string]interface{}{
		"success":     true,
		"method":      "native_go",
		"dry_run":     dryRun,
		"sync_results": syncResults,
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		output, _ := json.MarshalIndent(result, "", "  ")
		if err := os.WriteFile(outputPath, output, 0644); err == nil {
			result["output_path"] = outputPath
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
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
		} else if len(task.LongDescription) < 50 {
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

	var output string
	if outputFormat == "json" {
		outputBytes, _ := json.MarshalIndent(result, "", "  ")
		output = string(outputBytes)
	} else {
		output = formatClarityAnalysisText(result)
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err == nil {
			result["output_path"] = outputPath
		}
	}

	return []framework.TextContent{
		{Type: "text", Text: output},
	}, nil
}

// handleTaskWorkflowCleanup handles cleanup action for removing stale tasks
func handleTaskWorkflowCleanup(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	staleThresholdHours := 2.0
	if threshold, ok := params["stale_threshold_hours"].(float64); ok {
		staleThresholdHours = threshold
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

		// Identify stale tasks
		staleTasks := []*models.Todo2Task{}
		for _, task := range tasks {
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

		if dryRun {
			result := map[string]interface{}{
				"success":        true,
				"method":         "database",
				"dry_run":        true,
				"stale_count":    len(staleTasks),
				"stale_tasks":    formatStaleTasksFromPtrs(staleTasks),
				"threshold_hours": staleThresholdHours,
			}

			output, _ := json.MarshalIndent(result, "", "  ")
			return []framework.TextContent{
				{Type: "text", Text: string(output)},
			}, nil
		}

		// Delete stale tasks from database
		removedIDs := []string{}
		for _, task := range staleTasks {
			if err := database.DeleteTask(context.Background(), task.ID); err == nil {
				removedIDs = append(removedIDs, task.ID)
			}
		}

		// Get remaining count
		remainingCount := len(tasks) - len(removedIDs)

		result := map[string]interface{}{
			"success":         true,
			"method":          "database",
			"removed_count":   len(removedIDs),
			"remaining_count": remainingCount,
			"removed_tasks":   removedIDs,
			"threshold_hours": staleThresholdHours,
		}

		outputPath, _ := params["output_path"].(string)
		if outputPath != "" {
			output, _ := json.MarshalIndent(result, "", "  ")
			if err := os.WriteFile(outputPath, output, 0644); err == nil {
				result["output_path"] = outputPath
			}
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

	// Identify stale tasks
	staleTasks := []Todo2Task{}
	for _, task := range tasks {
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

	if dryRun {
		result := map[string]interface{}{
			"success":        true,
			"method":         "file",
			"dry_run":        true,
			"stale_count":    len(staleTasks),
			"stale_tasks":    formatStaleTasks(staleTasks),
			"threshold_hours": staleThresholdHours,
		}

		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	// Remove stale tasks
	remainingTasks := []Todo2Task{}
	staleMap := make(map[string]bool)
	for _, stale := range staleTasks {
		staleMap[stale.ID] = true
	}

	for _, task := range tasks {
		if !staleMap[task.ID] {
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
		"removed_count":   len(staleTasks),
		"remaining_count": len(remainingTasks),
		"removed_tasks":   extractTaskIDs(staleTasks),
		"threshold_hours": staleThresholdHours,
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		output, _ := json.MarshalIndent(result, "", "  ")
		if err := os.WriteFile(outputPath, output, 0644); err == nil {
			result["output_path"] = outputPath
		}
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
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

	// Extract optional parameters
	status := "Todo"
	if s, ok := params["status"].(string); ok && s != "" {
		status = normalizeStatus(s)
	}

	priority := "medium"
	if p, ok := params["priority"].(string); ok && p != "" {
		priority = normalizePriority(p)
	}

	tags := []string{}
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

	// Load existing tasks to generate next ID and validate dependencies
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Generate next task ID (T-{next_number})
	nextID := generateNextTaskID(tasks)

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

	// Try database first
	if err := database.CreateTask(ctx, task); err != nil {
		// Database failed, try file-based fallback
		tasks = append(tasks, *task)
		if err := SaveTodo2Tasks(projectRoot, tasks); err != nil {
			return nil, fmt.Errorf("failed to create task: database error: %v, file error: %w", err, err)
		}
	}

	// Return created task information
	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"task": map[string]interface{}{
			"id":              task.ID,
			"name":            task.Content,
			"long_description": task.LongDescription,
			"status":          task.Status,
			"priority":        task.Priority,
			"tags":            task.Tags,
			"dependencies":   task.Dependencies,
		},
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		output, _ := json.MarshalIndent(result, "", "  ")
		if err := os.WriteFile(outputPath, output, 0644); err == nil {
			result["output_path"] = outputPath
		}
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

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// generateNextTaskID generates the next sequential task ID (T-{number})
// Finds the highest existing task ID number and increments it
func generateNextTaskID(tasks []Todo2Task) string {
	maxNum := 0

	// Parse existing task IDs to find the highest number
	for _, task := range tasks {
		var num int
		if _, err := fmt.Sscanf(task.ID, "T-%d", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	// Return next ID
	return fmt.Sprintf("T-%d", maxNum+1)
}

// normalizePriority normalizes priority to valid values
func normalizePriority(priority string) string {
	priority = strings.ToLower(strings.TrimSpace(priority))
	switch priority {
	case "low", "medium", "high", "critical":
		return priority
	default:
		return "medium" // Default
	}
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

	if estimate.Metadata != nil && len(estimate.Metadata) > 0 {
		if statisticalEst, ok := estimate.Metadata["statistical_estimate"].(float64); ok {
			builder.WriteString(fmt.Sprintf("\n**Statistical Estimate:** %.1f hours\n", statisticalEst))
		}
		if appleFMEst, ok := estimate.Metadata["apple_fm_estimate"].(float64); ok {
			builder.WriteString(fmt.Sprintf("**AI Estimate:** %.1f hours\n", appleFMEst))
		}
	}

	return builder.String()
}
