package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
)

// handleReportOverview handles the overview action for report tool
func handleReportOverview(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	outputFormat := "text"
	if format, ok := params["output_format"].(string); ok && format != "" {
		outputFormat = format
	}

	outputPath, _ := params["output_path"].(string)

	// Aggregate project data
	overviewData, err := aggregateProjectData(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate project data: %w", err)
	}

	// Format output based on requested format
	var formattedOutput string
	switch outputFormat {
	case "json":
		jsonBytes, err := json.MarshalIndent(overviewData, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		formattedOutput = string(jsonBytes)
	case "markdown":
		formattedOutput = formatOverviewMarkdown(overviewData)
	case "html":
		formattedOutput = formatOverviewHTML(overviewData)
	default:
		formattedOutput = formatOverviewText(overviewData)
	}

	// Save to file if requested
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(formattedOutput), 0644); err != nil {
			return nil, fmt.Errorf("failed to write output file: %w", err)
		}
		formattedOutput += fmt.Sprintf("\n\n[Report saved to: %s]", outputPath)
	}

	return []framework.TextContent{
		{Type: "text", Text: formattedOutput},
	}, nil
}

// handleReportBriefing handles the briefing action for report tool
// Uses devwisdom-go wisdom engine directly (no MCP client needed)
func handleReportBriefing(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	var score float64

	if sc, ok := params["score"].(float64); ok {
		score = sc
	} else if sc, ok := params["score"].(int); ok {
		score = float64(sc)
	} else {
		score = 50.0 // Default score
	}

	// Validate and clamp score
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	// Get wisdom engine
	engine, err := getWisdomEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wisdom engine: %w", err)
	}

	// Get multiple quotes from different sources
	sources := engine.ListSources()
	briefing := map[string]interface{}{
		"date":    time.Now().Format("2006-01-02"),
		"score":   score,
		"quotes":  []interface{}{},
		"sources": sources,
	}

	// Get quotes from a few sources (up to 3)
	selectedSources := []string{"pistis_sophia", "stoic", "tao"}
	if len(sources) > 0 {
		selectedSources = sources
		if len(selectedSources) > 3 {
			selectedSources = selectedSources[:3]
		}
	}

	quotes := []interface{}{}
	for _, src := range selectedSources {
		quote, err := engine.GetWisdom(score, src)
		if err == nil {
			quotes = append(quotes, quote)
		}
	}

	briefing["quotes"] = quotes

	// Convert to JSON
	resultJSON, err := json.MarshalIndent(briefing, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal briefing: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// handleReportPRD handles the prd action for report tool
func handleReportPRD(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	projectName, _ := params["project_name"].(string)
	if projectName == "" {
		// Try to get from go.mod or default
		projectName = filepath.Base(projectRoot)
	}

	includeArchitecture := true
	if arch, ok := params["include_architecture"].(bool); ok {
		includeArchitecture = arch
	}

	includeMetrics := true
	if metrics, ok := params["include_metrics"].(bool); ok {
		includeMetrics = metrics
	}

	includeTasks := true
	if tasks, ok := params["include_tasks"].(bool); ok {
		includeTasks = tasks
	}

	outputPath, _ := params["output_path"].(string)

	// Generate PRD
	prd, err := generatePRD(ctx, projectRoot, projectName, includeArchitecture, includeMetrics, includeTasks)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PRD: %w", err)
	}

	// Save to file if requested
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(prd), 0644); err != nil {
			return nil, fmt.Errorf("failed to write PRD file: %w", err)
		}
		prd += fmt.Sprintf("\n\n[PRD saved to: %s]", outputPath)
	}

	return []framework.TextContent{
		{Type: "text", Text: prd},
	}, nil
}

// aggregateProjectData aggregates all project data for overview
func aggregateProjectData(ctx context.Context, projectRoot string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// Project info
	projectInfo, err := getProjectInfo(projectRoot)
	if err == nil {
		data["project"] = projectInfo
	}

	// Health metrics (try Go scorecard first, fallback to Python)
	if IsGoProject() {
		opts := &ScorecardOptions{FastMode: true}
		scorecard, err := GenerateGoScorecard(ctx, projectRoot, opts)
		if err == nil {
			// Calculate component scores using helper functions
			scores := map[string]float64{
				"security":      calculateSecurityScore(scorecard),
				"testing":       calculateTestingScore(scorecard),
				"documentation": calculateDocumentationScore(scorecard),
				"completion":    calculateCompletionScore(scorecard),
			}
			// Alignment score would need to be calculated separately
			scores["alignment"] = 50.0 // Default, could calculate from tasks

			data["health"] = map[string]interface{}{
				"overall_score":    scorecard.Score,
				"production_ready": scorecard.Score >= 80,
				"scores":           scores,
			}
		}
	}

	// Codebase metrics
	codebaseMetrics, err := getCodebaseMetrics(projectRoot)
	if err == nil {
		data["codebase"] = codebaseMetrics
	}

	// Task metrics
	taskMetrics, err := getTaskMetrics(projectRoot)
	if err == nil {
		data["tasks"] = taskMetrics
	}

	// Project phases
	data["phases"] = getProjectPhases()

	// Risks and blockers
	risks, err := getRisksAndBlockers(projectRoot)
	if err == nil {
		data["risks"] = risks
	}

	// Next actions
	nextActions, err := getNextActions(projectRoot)
	if err == nil {
		data["next_actions"] = nextActions
	}

	data["generated_at"] = time.Now().Format(time.RFC3339)

	return data, nil
}

// getProjectInfo extracts project metadata
func getProjectInfo(projectRoot string) (map[string]interface{}, error) {
	info := map[string]interface{}{
		"name":        filepath.Base(projectRoot),
		"version":     "0.1.0",
		"description": "MCP Server",
		"type":        "MCP Server",
		"status":      "Active Development",
	}

	// Try to get from go.mod
	goModPath := filepath.Join(projectRoot, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		content := string(data)
		if strings.Contains(content, "module ") {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "module ") {
					moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
					info["name"] = moduleName
					break
				}
			}
		}
	}

	return info, nil
}

// getCodebaseMetrics collects codebase statistics
func getCodebaseMetrics(projectRoot string) (map[string]interface{}, error) {
	var goFiles, pythonFiles, totalFiles int

	// Count files (simplified - could use more sophisticated counting)
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			// Skip hidden and vendor directories
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			goFiles++
		case ".py":
			pythonFiles++
		}
		totalFiles++
		return nil
	})

	metrics := map[string]interface{}{
		"go_files":     goFiles,
		"go_lines":     0, // Could count lines if needed
		"python_files": pythonFiles,
		"python_lines": 0, // Could count lines if needed
		"total_files":  totalFiles,
		"total_lines":  0,  // Could count lines if needed
		"tools":        30, // From registry
		"prompts":      19, // From templates.go
		"resources":    11, // From resources registry
	}

	// Count tools, prompts, resources from registry
	// These would need to be exported from registry.go or counted differently
	// For now, use approximate values
	metrics["tools"] = 30
	metrics["prompts"] = 19
	metrics["resources"] = 11

	return metrics, err
}

// getTaskMetrics collects task statistics
func getTaskMetrics(projectRoot string) (map[string]interface{}, error) {
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, err
	}

	pending := 0
	completed := 0
	totalHours := 0.0

	for _, task := range tasks {
		if IsPendingStatus(task.Status) {
			pending++
			// Check for estimated hours in metadata
			if task.Metadata != nil {
				if hours, ok := task.Metadata["estimatedHours"].(float64); ok {
					totalHours += hours
				}
			}
		} else if IsCompletedStatus(task.Status) {
			completed++
		}
	}

	completionRate := 0.0
	if len(tasks) > 0 {
		completionRate = float64(completed) / float64(len(tasks)) * 100
	}

	return map[string]interface{}{
		"total":           len(tasks),
		"pending":         pending,
		"completed":       completed,
		"completion_rate": completionRate,
		"remaining_hours": totalHours,
	}, nil
}

// getProjectPhases returns current phase status
func getProjectPhases() map[string]interface{} {
	return map[string]interface{}{
		"phase_1": map[string]interface{}{
			"name":     "Foundation Tools",
			"status":   "complete",
			"progress": 100,
		},
		"phase_2": map[string]interface{}{
			"name":     "Medium Complexity Tools",
			"status":   "in_progress",
			"progress": 44,
		},
		"phase_3": map[string]interface{}{
			"name":     "Complex Tools",
			"status":   "complete",
			"progress": 100,
		},
		"phase_4": map[string]interface{}{
			"name":     "Resources",
			"status":   "complete",
			"progress": 100,
		},
		"phase_5": map[string]interface{}{
			"name":     "Prompts",
			"status":   "complete",
			"progress": 100,
		},
	}
}

// getRisksAndBlockers identifies project risks
func getRisksAndBlockers(projectRoot string) ([]map[string]interface{}, error) {
	risks := []map[string]interface{}{}

	// Check for incomplete tasks with high priority
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err == nil {
		for _, task := range tasks {
			if IsPendingStatus(task.Status) && task.Priority == "critical" {
				risks = append(risks, map[string]interface{}{
					"type":        "blocker",
					"description": task.Content,
					"task_id":     task.ID,
					"priority":    task.Priority,
				})
			}
		}
	}

	return risks, nil
}

// getNextActions identifies next priority actions
func getNextActions(projectRoot string) ([]map[string]interface{}, error) {
	actions := []map[string]interface{}{}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err == nil {
		// Get high priority pending tasks
		for _, task := range tasks {
			if IsPendingStatus(task.Status) && (task.Priority == "high" || task.Priority == "critical") {
				estimatedHours := 0.0
				if task.Metadata != nil {
					if hours, ok := task.Metadata["estimatedHours"].(float64); ok {
						estimatedHours = hours
					}
				}
				actions = append(actions, map[string]interface{}{
					"task_id":         task.ID,
					"name":            task.Content,
					"priority":        task.Priority,
					"estimated_hours": estimatedHours,
				})
				if len(actions) >= 10 {
					break
				}
			}
		}
	}

	return actions, nil
}

// generatePRD generates a Product Requirements Document
func generatePRD(ctx context.Context, projectRoot, projectName string, includeArchitecture, includeMetrics, includeTasks bool) (string, error) {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Product Requirements Document: %s\n\n", projectName))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString("---\n\n")

	// Executive Summary
	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(fmt.Sprintf("%s is an MCP (Model Context Protocol) server providing project management automation tools.\n\n", projectName))

	// Requirements from Tasks
	if includeTasks {
		sb.WriteString("## Requirements\n\n")
		tasks, err := LoadTodo2Tasks(projectRoot)
		if err == nil {
			// Group tasks by priority
			criticalTasks := []Todo2Task{}
			highTasks := []Todo2Task{}
			mediumTasks := []Todo2Task{}

			for _, task := range tasks {
				if IsPendingStatus(task.Status) {
					switch task.Priority {
					case "critical":
						criticalTasks = append(criticalTasks, task)
					case "high":
						highTasks = append(highTasks, task)
					case "medium":
						mediumTasks = append(mediumTasks, task)
					}
				}
			}

			if len(criticalTasks) > 0 {
				sb.WriteString("### Critical Priority\n\n")
				for _, task := range criticalTasks {
					sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", task.Content, task.ID, task.LongDescription))
				}
				sb.WriteString("\n")
			}

			if len(highTasks) > 0 {
				sb.WriteString("### High Priority\n\n")
				for _, task := range highTasks {
					sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", task.Content, task.ID, task.LongDescription))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Architecture
	if includeArchitecture {
		sb.WriteString("## Architecture\n\n")
		sb.WriteString("### System Components\n\n")
		sb.WriteString("- **MCP Server**: Go-based server implementing MCP protocol\n")
		sb.WriteString("- **Tool Handlers**: Native Go implementations for project management tools\n")
		sb.WriteString("- **Python Bridge**: Subprocess bridge for remaining Python tools\n")
		sb.WriteString("- **Database**: SQLite for Todo2 task storage\n")
		sb.WriteString("- **Resources**: Memory and scorecard resources\n\n")
	}

	// Metrics
	if includeMetrics {
		sb.WriteString("## Metrics\n\n")
		metrics, err := getCodebaseMetrics(projectRoot)
		if err == nil {
			sb.WriteString(fmt.Sprintf("- **Total Files**: %d\n", metrics["total_files"]))
			sb.WriteString(fmt.Sprintf("- **Go Files**: %d\n", metrics["go_files"]))
			sb.WriteString(fmt.Sprintf("- **Tools**: %d\n", metrics["tools"]))
			sb.WriteString(fmt.Sprintf("- **Prompts**: %d\n", metrics["prompts"]))
			sb.WriteString(fmt.Sprintf("- **Resources**: %d\n", metrics["resources"]))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// formatOverviewText formats overview as plain text
func formatOverviewText(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  PROJECT OVERVIEW\n")
	sb.WriteString("======================================================================\n\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("Project Information:\n")
		sb.WriteString(fmt.Sprintf("  Name:        %s\n", project["name"]))
		sb.WriteString(fmt.Sprintf("  Version:     %s\n", project["version"]))
		sb.WriteString(fmt.Sprintf("  Type:        %s\n", project["type"]))
		sb.WriteString(fmt.Sprintf("  Status:      %s\n", project["status"]))
		sb.WriteString("\n")
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("Health Scorecard:\n")
		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Overall Score: %.1f%%\n", score))
		}
		if ready, ok := health["production_ready"].(bool); ok {
			if ready {
				sb.WriteString("  Production Ready: YES ✅\n")
			} else {
				sb.WriteString("  Production Ready: NO ❌\n")
			}
		}
		sb.WriteString("\n")
	}

	// Tasks
	if tasks, ok := data["tasks"].(map[string]interface{}); ok {
		sb.WriteString("Task Status:\n")
		sb.WriteString(fmt.Sprintf("  Total:           %d\n", tasks["total"]))
		sb.WriteString(fmt.Sprintf("  Pending:        %d\n", tasks["pending"]))
		sb.WriteString(fmt.Sprintf("  Completed:      %d\n", tasks["completed"]))
		if rate, ok := tasks["completion_rate"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Completion:     %.1f%%\n", rate))
		}
		if hours, ok := tasks["remaining_hours"].(float64); ok {
			sb.WriteString(fmt.Sprintf("  Remaining Hours: %.1f\n", hours))
		}
		sb.WriteString("\n")
	}

	// Next Actions
	if actions, ok := data["next_actions"].([]map[string]interface{}); ok && len(actions) > 0 {
		sb.WriteString("Next Actions:\n")
		for i, action := range actions {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("  %d. %s (Priority: %s)\n", i+1, action["name"], action["priority"]))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// formatOverviewMarkdown formats overview as markdown
func formatOverviewMarkdown(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("# Project Overview\n\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("## Project Information\n\n")
		sb.WriteString(fmt.Sprintf("- **Name**: %s\n", project["name"]))
		sb.WriteString(fmt.Sprintf("- **Version**: %s\n", project["version"]))
		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", project["type"]))
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", project["status"]))
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("## Health Scorecard\n\n")
		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("**Overall Score**: %.1f%%\n\n", score))
		}
	}

	// Tasks
	if tasks, ok := data["tasks"].(map[string]interface{}); ok {
		sb.WriteString("## Task Status\n\n")
		sb.WriteString(fmt.Sprintf("- **Total**: %d\n", tasks["total"]))
		sb.WriteString(fmt.Sprintf("- **Pending**: %d\n", tasks["pending"]))
		sb.WriteString(fmt.Sprintf("- **Completed**: %d\n", tasks["completed"]))
		if rate, ok := tasks["completion_rate"].(float64); ok {
			sb.WriteString(fmt.Sprintf("- **Completion Rate**: %.1f%%\n\n", rate))
		}
	}

	return sb.String()
}

// formatOverviewHTML formats overview as HTML
func formatOverviewHTML(data map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>Project Overview</title>\n")
	sb.WriteString("<style>body{font-family:Arial,sans-serif;margin:40px;}</style>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<h1>Project Overview</h1>\n")

	// Project Info
	if project, ok := data["project"].(map[string]interface{}); ok {
		sb.WriteString("<h2>Project Information</h2>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Name</strong>: %s</li>\n", project["name"]))
		sb.WriteString(fmt.Sprintf("<li><strong>Version</strong>: %s</li>\n", project["version"]))
		sb.WriteString(fmt.Sprintf("<li><strong>Type</strong>: %s</li>\n", project["type"]))
		sb.WriteString("</ul>\n")
	}

	// Health Score
	if health, ok := data["health"].(map[string]interface{}); ok {
		sb.WriteString("<h2>Health Scorecard</h2>\n")
		if score, ok := health["overall_score"].(float64); ok {
			sb.WriteString(fmt.Sprintf("<p><strong>Overall Score</strong>: %.1f%%</p>\n", score))
		}
	}

	sb.WriteString("</body>\n</html>\n")

	return sb.String()
}

// getFloatParam safely extracts float64 from params
func getFloatParam(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := params[key].(float64); ok {
		return val
	}
	return defaultValue
}
