package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
)

// PromptEntry represents a single prompt log entry
type PromptEntry struct {
	Timestamp    string `json:"timestamp"`
	Prompt       string `json:"prompt"`
	TaskID       string `json:"task_id,omitempty"`
	Mode         string `json:"mode"`
	Outcome      string `json:"outcome"`
	Iteration    int    `json:"iteration"`
	PromptLength int    `json:"prompt_length"`
}

// PromptLog represents the log file structure
type PromptLog struct {
	Created     string        `json:"created"`
	LastUpdated string        `json:"last_updated"`
	Entries     []PromptEntry `json:"entries"`
}

// handlePromptTrackingNative handles the prompt_tracking tool with native Go implementation
func handlePromptTrackingNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "analyze"
	}

	switch action {
	case "log":
		return handlePromptTrackingLog(ctx, params)
	case "analyze":
		return handlePromptTrackingAnalyze(ctx, params)
	default:
		return nil, fmt.Errorf("unknown action: %s (use 'log' or 'analyze')", action)
	}
}

// handlePromptTrackingLog handles the log action
func handlePromptTrackingLog(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return nil, fmt.Errorf("prompt parameter is required for log action")
	}

	var taskID string
	if tid, ok := params["task_id"].(string); ok {
		taskID = tid
	}

	var mode string
	if m, ok := params["mode"].(string); ok {
		mode = m
	} else {
		mode = "unknown"
	}

	var outcome string
	if o, ok := params["outcome"].(string); ok {
		outcome = o
	} else {
		outcome = "pending"
	}

	iteration := 1
	if iter, ok := params["iteration"].(float64); ok {
		iteration = int(iter)
	}

	// Truncate long prompts
	promptText := prompt
	if len(promptText) > 500 {
		promptText = promptText[:500]
	}

	// Find project root
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	// Ensure log directory exists
	logDir := filepath.Join(projectRoot, ".cursor", "prompt_history")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Get current session log file (daily log)
	now := time.Now()
	logFile := filepath.Join(logDir, fmt.Sprintf("session_%s.json", now.Format("20060102")))

	// Load existing log or create new
	var log PromptLog
	if data, err := os.ReadFile(logFile); err == nil {
		if err := json.Unmarshal(data, &log); err != nil {
			// Invalid JSON, create new log
			log = PromptLog{
				Created:     now.Format(time.RFC3339),
				LastUpdated: now.Format(time.RFC3339),
				Entries:     []PromptEntry{},
			}
		}
	} else {
		// File doesn't exist, create new log
		log = PromptLog{
			Created:     now.Format(time.RFC3339),
			LastUpdated: now.Format(time.RFC3339),
			Entries:     []PromptEntry{},
		}
	}

	// Create entry
	entry := PromptEntry{
		Timestamp:    now.Format(time.RFC3339),
		Prompt:       promptText,
		TaskID:       taskID,
		Mode:         mode,
		Outcome:      outcome,
		Iteration:    iteration,
		PromptLength: len(prompt),
	}

	// Append entry
	log.Entries = append(log.Entries, entry)
	log.LastUpdated = now.Format(time.RFC3339)

	// Save log
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal log: %w", err)
	}

	if err := os.WriteFile(logFile, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to save log: %w", err)
	}

	result := map[string]interface{}{
		"success":   true,
		"method":    "native_go",
		"data":      entry,
		"timestamp": time.Now().Unix(),
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// handlePromptTrackingAnalyze handles the analyze action
func handlePromptTrackingAnalyze(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	days := 7
	if d, ok := params["days"].(float64); ok {
		days = int(d)
	}

	// Find project root
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	logDir := filepath.Join(projectRoot, ".cursor", "prompt_history")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		// No logs yet
		result := map[string]interface{}{
			"success":         true,
			"method":          "native_go",
			"period_days":     days,
			"total_prompts":   0,
			"by_mode":         map[string]int{},
			"by_outcome":      map[string]int{},
			"avg_iterations":  0.0,
			"patterns":        []string{},
			"recommendations": []string{"No prompt history found. Use log action to track prompts."},
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(output)},
		}, nil
	}

	// Load all log files
	cutoffDate := time.Now().AddDate(0, 0, -days)
	allEntries := []PromptEntry{}

	entries, err := os.ReadDir(logDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) != ".json" {
				continue
			}

			logFile := filepath.Join(logDir, entry.Name())
			data, err := os.ReadFile(logFile)
			if err != nil {
				continue
			}

			var log PromptLog
			if err := json.Unmarshal(data, &log); err != nil {
				continue
			}

			// Filter entries by date
			for _, e := range log.Entries {
				if entryTime, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
					if entryTime.After(cutoffDate) {
						allEntries = append(allEntries, e)
					}
				}
			}
		}
	}

	// Analyze entries
	analysis := map[string]interface{}{
		"period_days":    days,
		"total_prompts":  len(allEntries),
		"by_mode":        map[string]int{},
		"by_outcome":     map[string]int{},
		"avg_iterations": 0.0,
		"patterns":       []string{},
		"recommendations": []string{},
	}

	byMode := make(map[string]int)
	byOutcome := make(map[string]int)
	taskIterations := make(map[string]int)

	for _, entry := range allEntries {
		byMode[entry.Mode]++
		byOutcome[entry.Outcome]++

		if entry.TaskID != "" {
			if entry.Iteration > taskIterations[entry.TaskID] {
				taskIterations[entry.TaskID] = entry.Iteration
			}
		}
	}

	analysis["by_mode"] = byMode
	analysis["by_outcome"] = byOutcome

	// Calculate average iterations
	if len(taskIterations) > 0 {
		sum := 0
		for _, iter := range taskIterations {
			sum += iter
		}
		avg := float64(sum) / float64(len(taskIterations))
		analysis["avg_iterations"] = fmt.Sprintf("%.2f", avg)
	}

	// Generate patterns
	patterns := []string{}
	if avgIter, ok := analysis["avg_iterations"].(string); ok {
		var avg float64
		fmt.Sscanf(avgIter, "%f", &avg)
		if avg > 3 {
			patterns = append(patterns, "High iteration count - consider more detailed initial prompts")
		}
	}

	agentCount := byMode["AGENT"]
	askCount := byMode["ASK"]
	if agentCount > askCount*2 {
		patterns = append(patterns, "Heavy AGENT usage - consider ASK for simpler queries")
	}

	failed := byOutcome["failed"]
	if len(allEntries) > 0 && failed > len(allEntries)/5 {
		patterns = append(patterns, "High failure rate - review prompt quality")
	}

	analysis["patterns"] = patterns

	// Generate recommendations
	recommendations := []string{}
	if avgIter, ok := analysis["avg_iterations"].(string); ok {
		var avg float64
		fmt.Sscanf(avgIter, "%f", &avg)
		if avg > 2 {
			recommendations = append(recommendations, "Break down complex tasks into smaller, more specific prompts")
		}
	}
	if len(byMode) == 0 {
		recommendations = append(recommendations, "Track workflow mode (AGENT/ASK) to optimize tool selection")
	}

	analysis["recommendations"] = recommendations

	result := map[string]interface{}{
		"success": true,
		"method":  "native_go",
		"data":    analysis,
		"timestamp": time.Now().Unix(),
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}
