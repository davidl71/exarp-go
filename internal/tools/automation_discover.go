// automation_discover.go — Automation discover workflow: task discovery and dead agent cleanup.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"sync"
	"time"
)

// handleAutomationDiscover handles the "discover" action for automation tool.
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

	if useLLM, ok := params["use_llm"].(bool); ok {
		taskDiscoveryParams["use_llm"] = useLLM
	}

	// Call task_discovery native handler
	result, err := handleTaskDiscoveryNative(ctx, taskDiscoveryParams)
	if err != nil {
		return nil, fmt.Errorf("task_discovery failed: %w", err)
	}

	// Return result as-is (already formatted as TextContent)
	return result, nil
}

// runDeadAgentCleanup runs dead agent lock cleanup (T-76). Returns result in runDailyTask shape.
func runDeadAgentCleanup(ctx context.Context) map[string]interface{} {
	startTime := time.Now()
	result := map[string]interface{}{
		"status":   "success",
		"duration": 0.0,
		"error":    "",
		"summary":  map[string]interface{}{"cleaned": 0, "task_ids": []string{}},
	}

	if _, err := database.GetDB(); err != nil {
		result["summary"] = map[string]interface{}{"skipped": true, "reason": "database not available"}
		result["duration"] = time.Since(startTime).Seconds()

		return result
	}

	staleThreshold := config.GetGlobalConfig().Timeouts.StaleLockThreshold
	if staleThreshold <= 0 {
		staleThreshold = 5 * time.Minute
	}

	cleaned, taskIDs, err := database.CleanupDeadAgentLocks(ctx, staleThreshold)
	result["duration"] = time.Since(startTime).Seconds()

	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		result["summary"] = map[string]interface{}{"error": err.Error()}

		return result
	}

	result["summary"] = map[string]interface{}{
		"cleaned":  cleaned,
		"task_ids": taskIDs,
	}

	return result
}

// parallelTask describes a task for parallel execution (T-228).
type parallelTask struct {
	toolName string
	params   map[string]interface{}
	taskName string
}

// runParallelTasks runs multiple tasks concurrently with max parallel limit (T-228).
func runParallelTasks(ctx context.Context, tasks []parallelTask, maxParallel int) []map[string]interface{} {
	if maxParallel <= 0 {
		maxParallel = 3
	}

	results := make([]map[string]interface{}, len(tasks))
	sem := make(chan struct{}, maxParallel)

	var wg sync.WaitGroup
	for i, t := range tasks {
		wg.Add(1)

		go func(idx int, task parallelTask) {
			defer wg.Done()

			sem <- struct{}{}

			defer func() { <-sem }()

			results[idx] = runDailyTask(ctx, task.toolName, task.params)
		}(i, t)
	}

	wg.Wait()

	return results
}

// runDailyTask runs a native Go tool task and returns result.
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
	case "tool_count_health":
		toolResult, err = handleHealthNative(ctx, params)
	case "dependency_security":
		toolResult, err = handleSecurityScan(ctx, params)
	case "handoff_check":
		toolResult, err = handleSessionNative(ctx, params)
	case "task_progress_inference":
		toolResult, err = handleInferTaskProgressNative(ctx, params)
	case "stale_task_cleanup":
		toolResult, err = handleTaskWorkflowNative(ctx, params)
	case "dead_agent_cleanup":
		return runDeadAgentCleanup(ctx)
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
