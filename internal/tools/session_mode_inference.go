package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
)

// SessionMode represents the inferred session mode
type SessionMode string

const (
	SessionModeAGENT   SessionMode = "agent"
	SessionModeASK     SessionMode = "ask"
	SessionModeMANUAL  SessionMode = "manual"
	SessionModeUNKNOWN SessionMode = "unknown"
)

// ModeInferenceResult represents the result of session mode inference
type ModeInferenceResult struct {
	Mode       SessionMode            `json:"mode"`
	Confidence float64                `json:"confidence"`
	Reasoning  []string               `json:"reasoning"`
	Metrics    map[string]interface{} `json:"metrics"`
	Timestamp  string                 `json:"timestamp"`
}

var lastInferenceCache *ModeInferenceResult
var lastInferenceTime time.Time

// HandleInferSessionModeNative handles the infer_session_mode tool with native Go implementation
// Exported for use by resources package
func HandleInferSessionModeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	forceRecompute, _ := params["force_recompute"].(bool)

	// Check cache (within 2 minutes) unless forcing recompute
	if !forceRecompute && lastInferenceCache != nil {
		if time.Since(lastInferenceTime) < 2*time.Minute {
			result, err := json.MarshalIndent(lastInferenceCache, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal cached result: %w", err)
			}
			return []framework.TextContent{
				{Type: "text", Text: string(result)},
			}, nil
		}
	}

	// Find project root
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" || projectRoot == "unknown" {
		var err error
		projectRoot, err = FindProjectRoot()
		if err != nil {
			// If project root not found, return unknown mode
			result := ModeInferenceResult{
				Mode:       SessionModeUNKNOWN,
				Confidence: 0.0,
				Reasoning:  []string{"Project root not found"},
				Metrics:    map[string]interface{}{},
				Timestamp:  time.Now().Format(time.RFC3339),
			}
			return marshalInferenceResult(result)
		}
	}

	// Load Todo2 tasks
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		result := ModeInferenceResult{
			Mode:       SessionModeUNKNOWN,
			Confidence: 0.0,
			Reasoning:  []string{fmt.Sprintf("Failed to load Todo2 tasks: %v", err)},
			Metrics:    map[string]interface{}{},
			Timestamp:  time.Now().Format(time.RFC3339),
		}
		return marshalInferenceResult(result)
	}

	// Infer mode from task patterns
	result := inferModeFromTasks(tasks, projectRoot)

	// Cache result
	lastInferenceCache = &result
	lastInferenceTime = time.Now()

	return marshalInferenceResult(result)
}

// handleInferSessionModeNative is an alias for HandleInferSessionModeNative
// Kept for backward compatibility with existing tool handler
func handleInferSessionModeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	return HandleInferSessionModeNative(ctx, params)
}

// marshalInferenceResult marshals the inference result to JSON
func marshalInferenceResult(result ModeInferenceResult) ([]framework.TextContent, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inference result: %w", err)
	}
	return []framework.TextContent{
		{Type: "text", Text: string(data)},
	}, nil
}

// inferModeFromTasks infers session mode from Todo2 task patterns
func inferModeFromTasks(tasks []Todo2Task, projectRoot string) ModeInferenceResult {
	if len(tasks) == 0 {
		return ModeInferenceResult{
			Mode:       SessionModeASK,
			Confidence: 0.3,
			Reasoning:  []string{"No tasks found - defaulting to ASK mode"},
			Metrics: map[string]interface{}{
				"total_tasks": 0,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
	}

	// Analyze task patterns
	metrics := analyzeTaskPatterns(tasks)
	reasoning := []string{}

	// Count tasks by status
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0
	taskWithDeps := 0
	multiFileTags := 0

	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pendingCount++
		}
		if task.Status == "In Progress" {
			inProgressCount++
		}
		if IsCompletedStatus(task.Status) {
			completedCount++
		}
		if len(task.Dependencies) > 0 {
			taskWithDeps++
		}
		// Check for multi-file indicators in tags or content
		content := strings.ToLower(task.Content + " " + task.LongDescription)
		if strings.Contains(content, "multi") || strings.Contains(content, "file") ||
			strings.Contains(content, "refactor") || strings.Contains(content, "implement") {
			multiFileTags++
		}
	}

	totalTasks := len(tasks)
	pendingRatio := float64(pendingCount) / float64(totalTasks)
	inProgressRatio := float64(inProgressCount) / float64(totalTasks)
	completionRatio := float64(completedCount) / float64(totalTasks)
	depsRatio := float64(taskWithDeps) / float64(totalTasks)

	metrics["total_tasks"] = totalTasks
	metrics["pending_tasks"] = pendingCount
	metrics["in_progress_tasks"] = inProgressCount
	metrics["completed_tasks"] = completedCount
	metrics["tasks_with_dependencies"] = taskWithDeps
	metrics["pending_ratio"] = pendingRatio
	metrics["in_progress_ratio"] = inProgressRatio
	metrics["completion_ratio"] = completionRatio
	metrics["dependencies_ratio"] = depsRatio

	// Infer mode based on heuristics
	var mode SessionMode
	var confidence float64

	// AGENT mode indicators:
	// - High number of in-progress tasks (>30% of total)
	// - Many tasks with dependencies (suggests multi-step work)
	// - Multiple pending tasks (suggests autonomous execution)
	agentScore := 0.0
	if inProgressRatio > 0.3 {
		agentScore += 0.4
		reasoning = append(reasoning, fmt.Sprintf("High in-progress ratio (%.1f%%) suggests active work", inProgressRatio*100))
	}
	if depsRatio > 0.4 {
		agentScore += 0.3
		reasoning = append(reasoning, fmt.Sprintf("Many tasks with dependencies (%.1f%%) suggest multi-step execution", depsRatio*100))
	}
	if pendingRatio > 0.5 && totalTasks > 5 {
		agentScore += 0.2
		reasoning = append(reasoning, "Many pending tasks suggest autonomous task management")
	}
	if float64(multiFileTags) > float64(totalTasks)*0.3 {
		agentScore += 0.1
		reasoning = append(reasoning, "Task content suggests multi-file work")
	}

	// ASK mode indicators:
	// - Low task count (< 5)
	// - Mostly completed tasks (suggests quick questions/completions)
	// - Low dependency ratio
	askScore := 0.0
	if totalTasks < 5 {
		askScore += 0.4
		reasoning = append(reasoning, "Low task count suggests focused questions")
	}
	if completionRatio > 0.7 && totalTasks > 0 {
		askScore += 0.3
		reasoning = append(reasoning, fmt.Sprintf("High completion ratio (%.1f%%) suggests quick task completions", completionRatio*100))
	}
	if depsRatio < 0.2 {
		askScore += 0.2
		reasoning = append(reasoning, "Low dependency ratio suggests simple, focused tasks")
	}

	// MANUAL mode indicators:
	// - Very low task activity (mostly completed, few pending)
	manualScore := 0.0
	if completionRatio > 0.8 && pendingRatio < 0.1 && totalTasks > 3 {
		manualScore += 0.5
		reasoning = append(reasoning, "Very low activity suggests manual work")
	}

	// Determine mode
	if agentScore > askScore && agentScore > manualScore {
		mode = SessionModeAGENT
		confidence = min(agentScore, 0.95)
		if confidence < 0.5 {
			confidence = 0.5 // Minimum confidence for mode selection
		}
	} else if askScore > manualScore {
		mode = SessionModeASK
		confidence = min(askScore, 0.95)
		if confidence < 0.5 {
			confidence = 0.5
		}
	} else if manualScore > 0.3 {
		mode = SessionModeMANUAL
		confidence = min(manualScore, 0.95)
		if confidence < 0.5 {
			confidence = 0.5
		}
	} else {
		// Default to ASK if unclear
		mode = SessionModeASK
		confidence = 0.5
		reasoning = append(reasoning, "Unclear pattern - defaulting to ASK mode")
	}

	if len(reasoning) == 0 {
		reasoning = append(reasoning, fmt.Sprintf("Inferred from %d tasks", totalTasks))
	}

	return ModeInferenceResult{
		Mode:       mode,
		Confidence: confidence,
		Reasoning:  reasoning,
		Metrics:    metrics,
		Timestamp:  time.Now().Format(time.RFC3339),
	}
}

// analyzeTaskPatterns analyzes task patterns and returns metrics
func analyzeTaskPatterns(tasks []Todo2Task) map[string]interface{} {
	metrics := make(map[string]interface{})

	statusCounts := make(map[string]int)
	priorityCounts := make(map[string]int)
	tagCounts := make(map[string]int)

	for _, task := range tasks {
		statusCounts[task.Status]++
		if task.Priority != "" {
			priorityCounts[task.Priority]++
		}
		for _, tag := range task.Tags {
			tagCounts[tag]++
		}
	}

	metrics["status_distribution"] = statusCounts
	metrics["priority_distribution"] = priorityCounts
	metrics["tag_counts"] = tagCounts

	return metrics
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
