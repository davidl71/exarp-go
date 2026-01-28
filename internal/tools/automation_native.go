package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
)

// handleAutomationNative handles the automation tool with native Go implementation
// Implements all actions: "daily", "nightly", "sprint", and "discover"
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
		"timestamp":       time.Now().Format(time.RFC3339),
		"action":          "daily",
		"tasks_run":       []map[string]interface{}{},
		"tasks_succeeded": []string{},
		"tasks_failed":    []string{},
		"summary":         map[string]interface{}{},
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
		"total_tasks":      totalTasks,
		"succeeded":        succeededCount,
		"failed":           failedCount,
		"success_rate":     0.0,
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

// handleAutomationNightly handles the "nightly" action for automation tool
// Runs maintenance and cleanup tasks that are suitable for overnight execution
func handleAutomationNightly(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	// Results structure
	results := map[string]interface{}{
		"timestamp":       time.Now().Format(time.RFC3339),
		"action":          "nightly",
		"tasks_run":       []map[string]interface{}{},
		"tasks_succeeded": []string{},
		"tasks_failed":    []string{},
		"summary":         map[string]interface{}{},
	}

	// Task 1: memory_maint (consolidate action - cleanup/maintenance)
	task1Result := runDailyTask(ctx, "memory_maint", map[string]interface{}{
		"action": "consolidate",
	})
	tasksRun, _ := results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "memory_maint",
		"task_name": "Memory Consolidation",
		"status":    task1Result["status"],
		"duration":  task1Result["duration"],
		"error":     task1Result["error"],
		"summary":   task1Result["summary"],
	})
	results["tasks_run"] = tasksRun
	if task1Result["status"] == "success" {
		tasksSucceeded, _ := results["tasks_succeeded"].([]string)
		results["tasks_succeeded"] = append(tasksSucceeded, "memory_maint")
	} else {
		tasksFailed, _ := results["tasks_failed"].([]string)
		results["tasks_failed"] = append(tasksFailed, "memory_maint")
	}

	// Task 2: task_analysis (tags action - cleanup/organization)
	task2Result := runDailyTask(ctx, "task_analysis", map[string]interface{}{
		"action": "tags",
	})
	tasksRun, _ = results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "task_analysis",
		"task_name": "Task Tag Analysis",
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

	// Task 3: health (server action - system health check)
	task3Result := runDailyTask(ctx, "health", map[string]interface{}{
		"action": "server",
	})
	tasksRun, _ = results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "health",
		"task_name": "Server Health Check",
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
		"total_tasks":      totalTasks,
		"succeeded":        succeededCount,
		"failed":           failedCount,
		"success_rate":     0.0,
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

// handleAutomationSprint handles the "sprint" action for automation tool
// Runs sprint-specific tasks like alignment analysis and reporting
func handleAutomationSprint(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	// Results structure
	results := map[string]interface{}{
		"timestamp":       time.Now().Format(time.RFC3339),
		"action":          "sprint",
		"tasks_run":       []map[string]interface{}{},
		"tasks_succeeded": []string{},
		"tasks_failed":    []string{},
		"summary":         map[string]interface{}{},
	}

	// Task 1: analyze_alignment (todo2 action - sprint alignment)
	task1Result := runDailyTask(ctx, "analyze_alignment", map[string]interface{}{
		"action": "todo2",
	})
	tasksRun, _ := results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "analyze_alignment",
		"task_name": "Sprint Alignment Analysis",
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

	// Task 2: task_analysis (hierarchy action - sprint planning)
	task2Result := runDailyTask(ctx, "task_analysis", map[string]interface{}{
		"action": "hierarchy",
	})
	tasksRun, _ = results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "task_analysis",
		"task_name": "Task Hierarchy Analysis",
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

	// Task 3: report (overview action - sprint reporting, native Go)
	task3Result := runDailyTask(ctx, "report", map[string]interface{}{
		"action": "overview",
	})
	tasksRun, _ = results["tasks_run"].([]map[string]interface{})
	tasksRun = append(tasksRun, map[string]interface{}{
		"task_id":   "report",
		"task_name": "Sprint Overview Report",
		"status":    task3Result["status"],
		"duration":  task3Result["duration"],
		"error":     task3Result["error"],
		"summary":   task3Result["summary"],
	})
	results["tasks_run"] = tasksRun
	if task3Result["status"] == "success" {
		tasksSucceeded, _ := results["tasks_succeeded"].([]string)
		results["tasks_succeeded"] = append(tasksSucceeded, "report")
	} else {
		tasksFailed, _ := results["tasks_failed"].([]string)
		results["tasks_failed"] = append(tasksFailed, "report")
	}

	// Generate summary
	tasksSucceeded, _ := results["tasks_succeeded"].([]string)
	tasksFailed, _ := results["tasks_failed"].([]string)
	totalTasks := len(tasksRun)
	succeededCount := len(tasksSucceeded)
	failedCount := len(tasksFailed)

	summary := map[string]interface{}{
		"total_tasks":      totalTasks,
		"succeeded":        succeededCount,
		"failed":           failedCount,
		"success_rate":     0.0,
		"duration_seconds": time.Since(startTime).Seconds(),
	}
	if totalTasks > 0 {
		summary["success_rate"] = float64(succeededCount) / float64(totalTasks) * 100.0
	}
	results["summary"] = summary
	results["duration_seconds"] = time.Since(startTime).Seconds()

	// Add recommended backlog order for this sprint (first 15 in dependency order)
	if projectRoot, err := FindProjectRoot(); err == nil {
		if tasks, err := LoadTodo2Tasks(projectRoot); err == nil {
			if orderedIDs, _, details, err := BacklogExecutionOrder(tasks); err == nil && len(orderedIDs) > 0 {
				const sprintOrderLimit = 15
				sprintOrder := make([]map[string]interface{}, 0, sprintOrderLimit)
				detailMap := make(map[string]BacklogTaskDetail)
				for _, d := range details {
					detailMap[d.ID] = d
				}
				for i, id := range orderedIDs {
					if i >= sprintOrderLimit {
						break
					}
					d := detailMap[id]
					sprintOrder = append(sprintOrder, map[string]interface{}{
						"id":       d.ID,
						"content":  d.Content,
						"priority": d.Priority,
						"level":    d.Level,
					})
				}
				results["sprint_backlog_order"] = sprintOrder
			}
		}
	}

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

	// Call appropriate native handler (no Python bridge - per native Go migration plan)
	var err error
	var toolResult []framework.TextContent

	switch toolName {
	case "analyze_alignment":
		toolResult, err = handleAnalyzeAlignmentNative(ctx, params)
	case "task_analysis":
		toolResult, err = handleTaskAnalysisNative(ctx, params)
	case "health":
		toolResult, err = handleHealthNative(ctx, params)
	case "memory_maint":
		argsJSON, marshalErr := json.Marshal(params)
		if marshalErr != nil {
			result["error"] = marshalErr.Error()
			result["duration"] = time.Since(startTime).Seconds()
			return result
		}
		toolResult, err = handleMemoryMaintNative(ctx, argsJSON)
	case "report":
		toolResult, err = handleReportOverview(ctx, params)
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
