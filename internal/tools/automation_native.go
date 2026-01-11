package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
)

// handleAutomationNative handles the automation tool with native Go implementation
// Currently implements "daily" and "discover" actions
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
	default:
		// For nightly and sprint actions, return error to trigger Python bridge fallback
		return nil, fmt.Errorf("automation action '%s' not yet implemented in native Go, using Python bridge", action)
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

	// Task 3: health (docs action) - uses Python bridge (not yet native)
	task3Result := runDailyTaskPython(ctx, "health", map[string]interface{}{
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
