package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
)

// handleAutomationNative handles the automation tool with native Go implementation
// Implements "daily", "discover", "nightly", and "sprint" actions
func handleAutomationNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "daily"
	}

	switch action {
	case "daily":
		return handleAutomationDaily(ctx, params)
	case "discover":
		return handleAutomationDiscover(ctx, params)
	case "nightly":
		return handleAutomationNightly(ctx, params)
	case "sprint":
		return handleAutomationSprint(ctx, params)
	default:
		return nil, fmt.Errorf("unknown automation action: %s (use 'daily', 'nightly', 'sprint', or 'discover')", action)
	}
}

// handleAutomationDaily handles the "daily" action for automation tool
func handleAutomationDaily(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	// Results structure
	results := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"action":        "daily",
		"tasks_run":     []map[string]interface{}{},
		"tasks_succeeded": []string{},
		"tasks_failed":   []string{},
		"summary":       map[string]interface{}{},
	}

	// Task 1: analyze_alignment (todo2 action)
	task1Result := runDailyTask(ctx, "analyze_alignment", map[string]interface{}{
		"action": "todo2",
	})
	tasksRun, _ := results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "analyze_alignment",
		"task_name": "Todo2 Alignment Analysis",
		"status":    task1Result["status"],
		"duration":  task1Result["duration"],
		"error":     task1Result["error"],
		"summary":   task1Result["summary"],
	})
	results["tasks_run"] = tasksRun
	if task1Result["status"] == "success" {
		tasksSucceeded, _ := results["tasks_succeeded"].([]string)
		results["tasks_succeeded"] = append(tasksSucceeded, "analyze_alignment")
	} else {
		tasksFailed, _ := results["tasks_failed"].([]string)
		results["tasks_failed"] = append(tasksFailed, "analyze_alignment")
	}

	// Task 2: task_analysis (duplicates action)
	task2Result := runDailyTask(ctx, "task_analysis", map[string]interface{}{
		"action": "duplicates",
	})
	tasksRun, _ = results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "task_analysis",
		"task_name": "Duplicate Task Detection",
		"status":    task2Result["status"],
		"duration":  task2Result["duration"],
		"error":     task2Result["error"],
		"summary":   task2Result["summary"],
	})
	results["tasks_run"] = tasksRun
	if task2Result["status"] == "success" {
		tasksSucceeded, _ := results["tasks_succeeded"].([]string)
		results["tasks_succeeded"] = append(tasksSucceeded, "task_analysis")
	} else {
		tasksFailed, _ := results["tasks_failed"].([]string)
		results["tasks_failed"] = append(tasksFailed, "task_analysis")
	}

	// Task 3: health (docs action) - now uses native Go
	task3Result := runDailyTask(ctx, "health", map[string]interface{}{
		"action": "docs",
	})
	tasksRun, _ = results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "health",
		"task_name": "Documentation Health Check",
		"status":    task3Result["status"],
		"duration":  task3Result["duration"],
		"error":     task3Result["error"],
		"summary":   task3Result["summary"],
	})
	results["tasks_run"] = tasksRun
	if task3Result["status"] == "success" {
		tasksSucceeded, _ := results["tasks_succeeded"].([]string)
		results["tasks_succeeded"] = append(tasksSucceeded, "health")
	} else {
		tasksFailed, _ := results["tasks_failed"].([]string)
		results["tasks_failed"] = append(tasksFailed, "health")
	}

	// Generate summary
	tasksSucceeded, _ := results["tasks_succeeded"].([]string)
	tasksFailed, _ := results["tasks_failed"].([]string)
	totalTasks := len(tasksRun)
	succeededCount := len(tasksSucceeded)
	failedCount := len(tasksFailed)
	
	summary := map[string]interface{}{
		"total_tasks":   totalTasks,
		"succeeded":     succeededCount,
		"failed":        failedCount,
		"success_rate":  0.0,
		"duration_seconds": time.Since(startTime).Seconds(),
	}
	if totalTasks > 0 {
		summary["success_rate"] = float64(succeededCount) / float64(totalTasks) * 100.0
	}
	results["summary"] = summary
	results["duration_seconds"] = time.Since(startTime).Seconds()

	// Build response
	responseData := map[string]interface{}{
		"status":  "success",
		"results": results,
	}

	resultJSON, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// handleAutomationDiscover handles the "discover" action for automation tool
func handleAutomationDiscover(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Use task_discovery tool (native Go)
	// Set action to "all" to find tasks from all sources
	taskDiscoveryParams := map[string]interface{}{
		"action": "all",
	}
	
	// Get optional parameters
	if minValueScore, ok := params["min_value_score"].(float64); ok {
		// Note: task_discovery doesn't have min_value_score, but we can filter results
		// For now, just pass it through (might be used for filtering)
		taskDiscoveryParams["min_value_score"] = minValueScore
	}
	if outputPath, ok := params["output_path"].(string); ok && outputPath != "" {
		taskDiscoveryParams["output_path"] = outputPath
	}

	// Call task_discovery native handler
	result, err := handleTaskDiscoveryNative(ctx, taskDiscoveryParams)
	if err != nil {
		return nil, fmt.Errorf("task_discovery failed: %w", err)
	}

	// Return result as-is (already formatted as TextContent)
	return result, nil
}

// runDailyTask runs a native Go tool task and returns result
func runDailyTask(ctx context.Context, toolName string, params map[string]interface{}) map[string]interface{} {
	startTime := time.Now()

	result := map[string]interface{}{
		"status":   "error",
		"duration": 0.0,
		"error":    "",
		"summary":  map[string]interface{}{},
	}

	// Call appropriate native handler
	var err error
	var toolResult []framework.TextContent

	switch toolName {
	case "analyze_alignment":
		toolResult, err = handleAnalyzeAlignmentNative(ctx, params)
	case "task_analysis":
		toolResult, err = handleTaskAnalysisNative(ctx, params)
	case "health":
		toolResult, err = handleHealthNative(ctx, params)
	default:
		result["error"] = fmt.Sprintf("unknown tool: %s", toolName)
		result["duration"] = time.Since(startTime).Seconds()
		return result
	}

	duration := time.Since(startTime).Seconds()
	result["duration"] = duration

	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		return result
	}

	// Parse result JSON
	if len(toolResult) > 0 && toolResult[0].Text != "" {
		var resultData map[string]interface{}
		if err := json.Unmarshal([]byte(toolResult[0].Text), &resultData); err == nil {
			result["status"] = "success"
			result["summary"] = resultData
		} else {
			// If not JSON, treat as text summary
			result["status"] = "success"
			result["summary"] = map[string]interface{}{
				"output": toolResult[0].Text,
			}
		}
	} else {
		result["status"] = "success"
		result["summary"] = map[string]interface{}{}
	}

	return result
}

// runDailyTaskPython runs a Python bridge tool task and returns result
func runDailyTaskPython(ctx context.Context, toolName string, params map[string]interface{}) map[string]interface{} {
	startTime := time.Now()

	result := map[string]interface{}{
		"status":   "error",
		"duration": 0.0,
		"error":    "",
		"summary":  map[string]interface{}{},
	}

	// Call Python bridge
	bridgeResult, err := bridge.ExecutePythonTool(ctx, toolName, params)
	duration := time.Since(startTime).Seconds()
	result["duration"] = duration

	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		return result
	}

	// Parse result JSON
	var resultData map[string]interface{}
	if err := json.Unmarshal([]byte(bridgeResult), &resultData); err == nil {
		result["status"] = "success"
		result["summary"] = resultData
	} else {
		// If not JSON, treat as text summary
		result["status"] = "success"
		result["summary"] = map[string]interface{}{
			"output": bridgeResult,
		}
	}

	return result
}

// handleAutomationNightly handles the "nightly" action for automation tool
// Processes background-capable tasks and assigns them to agents
func handleAutomationNightly(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	maxTasksPerHost := 5
	if max, ok := params["max_tasks_per_host"].(float64); ok {
		maxTasksPerHost = int(max)
	}

	maxParallelTasks := 10
	if max, ok := params["max_parallel_tasks"].(float64); ok {
		maxParallelTasks = int(max)
	}

	priorityFilter := ""
	if prio, ok := params["priority_filter"].(string); ok {
		priorityFilter = prio
	}

	tagFilter := []string{}
	if tagsRaw, ok := params["tag_filter"].([]interface{}); ok {
		for _, tag := range tagsRaw {
			if tagStr, ok := tag.(string); ok {
				tagFilter = append(tagFilter, tagStr)
			}
		}
	}

	dryRun := false
	if dry, ok := params["dry_run"].(bool); ok {
		dryRun = dry
	}

	// Load tasks
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// Filter background-capable tasks (Todo or In Progress status, with background tag)
	backgroundTasks := make([]Todo2Task, 0)
	for _, task := range tasks {
		if task.Status != "Todo" && task.Status != "In Progress" {
			continue
		}

		// Check if task has background tag
		hasBackgroundTag := false
		for _, tag := range task.Tags {
			if tag == "background" || tag == "background-agent" || tag == "background-capable" {
				hasBackgroundTag = true
				break
			}
		}

		if !hasBackgroundTag {
			continue
		}

		// Apply priority filter
		if priorityFilter != "" && task.Priority != priorityFilter {
			continue
		}

		// Apply tag filter
		if len(tagFilter) > 0 {
			hasTag := false
			for _, filterTag := range tagFilter {
				for _, taskTag := range task.Tags {
					if taskTag == filterTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		backgroundTasks = append(backgroundTasks, task)
	}

	// Limit to max parallel tasks
	if len(backgroundTasks) > maxParallelTasks {
		backgroundTasks = backgroundTasks[:maxParallelTasks]
	}

	// Get agent ID
	agentID, err := database.GetAgentID()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent ID: %w", err)
	}

	// Process tasks (claim and assign)
	processed := 0
	assigned := 0
	for i, task := range backgroundTasks {
		if i >= maxTasksPerHost {
			break
		}

		if !dryRun {
			// Try to claim task
			result, err := database.ClaimTaskForAgent(ctx, task.ID, agentID, 30*time.Minute)
			if err != nil || !result.Success {
				continue
			}

			// Update task status to "In Progress"
			task.Status = "In Progress"
			if err := database.UpdateTask(ctx, &task); err != nil {
				continue
			}

			assigned++
		}
		processed++
	}

	results := map[string]interface{}{
		"status":               "success",
		"action":               "nightly",
		"timestamp":            time.Now().Format(time.RFC3339),
		"background_tasks_found": len(backgroundTasks),
		"tasks_processed":      processed,
		"tasks_assigned":       assigned,
		"max_tasks_per_host":   maxTasksPerHost,
		"max_parallel_tasks":   maxParallelTasks,
		"duration_seconds":     time.Since(startTime).Seconds(),
	}

	if dryRun {
		results["dry_run"] = true
		results["message"] = "Dry run mode - no tasks were actually assigned"
	}

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// handleAutomationSprint handles the "sprint" action for automation tool
// Orchestrates sprint workflow with subtask extraction, auto-approval, analysis/testing, and background task processing
func handleAutomationSprint(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()
	_, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	maxIterations := 10
	if max, ok := params["max_iterations"].(float64); ok {
		maxIterations = int(max)
	}

	autoApprove := true
	if approve, ok := params["auto_approve"].(bool); ok {
		autoApprove = approve
	}

	extractSubtasks := true
	if extract, ok := params["extract_subtasks"].(bool); ok {
		extractSubtasks = extract
	}

	runAnalysisTools := true
	if run, ok := params["run_analysis_tools"].(bool); ok {
		runAnalysisTools = run
	}

	runTestingTools := true
	if run, ok := params["run_testing_tools"].(bool); ok {
		runTestingTools = run
	}

	priorityFilter := ""
	if prio, ok := params["priority_filter"].(string); ok {
		priorityFilter = prio
	}

	tagFilter := []string{}
	if tagsRaw, ok := params["tag_filter"].([]interface{}); ok {
		for _, tag := range tagsRaw {
			if tagStr, ok := tag.(string); ok {
				tagFilter = append(tagFilter, tagStr)
			}
		}
	}

	dryRun := false
	if dry, ok := params["dry_run"].(bool); ok {
		dryRun = dry
	}

	results := map[string]interface{}{
		"status":              "success",
		"action":              "sprint",
		"timestamp":           time.Now().Format(time.RFC3339),
		"iterations":          0,
		"tasks_processed":     0,
		"tasks_completed":     0,
		"subtasks_extracted": 0,
		"tasks_auto_approved": 0,
		"analysis_results":    map[string]interface{}{},
		"testing_results":     map[string]interface{}{},
	}

	// Step 1: Extract subtasks (if enabled)
	if extractSubtasks {
		// Note: Subtask extraction would require task_analysis tool with hierarchy action
		// For now, we'll skip this step and note it in results
		results["subtasks_extracted"] = 0
		results["subtask_extraction"] = "not yet implemented in native Go"
	}

	// Step 2: Auto-approve tasks (if enabled)
	if autoApprove && !dryRun {
		// Use task_workflow approve action
		approveParams := map[string]interface{}{
			"action": "approve",
		}
		approveResult, err := handleTaskWorkflowNative(ctx, approveParams)
		if err == nil && len(approveResult) > 0 {
			var approveData map[string]interface{}
			if err := json.Unmarshal([]byte(approveResult[0].Text), &approveData); err == nil {
				if tasksApproved, ok := approveData["tasks_approved"].(float64); ok {
					results["tasks_auto_approved"] = int(tasksApproved)
				}
			}
		}
	}

	// Step 3: Run analysis tools (if enabled)
	if runAnalysisTools {
		analysisResults := map[string]interface{}{}

		// Run alignment analysis
		alignParams := map[string]interface{}{
			"action": "todo2",
		}
		alignResult := runDailyTask(ctx, "analyze_alignment", alignParams)
		analysisResults["alignment"] = alignResult["summary"]

		// Run duplicate detection
		dupParams := map[string]interface{}{
			"action": "duplicates",
		}
		dupResult := runDailyTask(ctx, "task_analysis", dupParams)
		analysisResults["duplicates"] = dupResult["summary"]

		// Run docs health check
		healthParams := map[string]interface{}{
			"action": "docs",
		}
		healthResult := runDailyTask(ctx, "health", healthParams)
		analysisResults["docs_health"] = healthResult["summary"]

		results["analysis_results"] = analysisResults
	}

	// Step 4: Run testing tools (if enabled)
	if runTestingTools {
		testingResults := map[string]interface{}{}

		// Note: Testing tools would require testing tool implementation
		// For now, we'll note this in results
		testingResults["status"] = "not yet implemented in native Go"
		results["testing_results"] = testingResults
	}

	// Step 5: Process background tasks (similar to nightly)
	nightlyParams := map[string]interface{}{
		"action":             "nightly",
		"max_tasks_per_host": 5,
		"max_parallel_tasks": 10,
		"priority_filter":    priorityFilter,
		"tag_filter":         tagFilter,
		"dry_run":            dryRun,
	}
	nightlyResult, err := handleAutomationNightly(ctx, nightlyParams)
	if err == nil && len(nightlyResult) > 0 {
		var nightlyData map[string]interface{}
		if err := json.Unmarshal([]byte(nightlyResult[0].Text), &nightlyData); err == nil {
			if tasksProcessed, ok := nightlyData["tasks_processed"].(float64); ok {
				results["tasks_processed"] = int(tasksProcessed)
			}
		}
	}

	results["iterations"] = 1 // Simplified: single iteration
	results["duration_seconds"] = time.Since(startTime).Seconds()

	if dryRun {
		results["dry_run"] = true
		results["message"] = "Dry run mode - no tasks were actually processed"
	}

	// Suppress unused variable warning
	_ = maxIterations

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}
