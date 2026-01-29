package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/cache"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
	"github.com/davidl71/exarp-go/proto"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
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
	includePlanning, _ := params["include_planning"].(bool)

	// Aggregate project data (proto-based internally)
	overviewProto, err := aggregateProjectDataProto(ctx, projectRoot, includePlanning)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate project data: %w", err)
	}

	// Format output based on requested format (use proto for type-safe formatting)
	var formattedOutput string
	switch outputFormat {
	case "json":
		overviewMap := ProtoToProjectOverviewData(overviewProto)
		contents, err := response.FormatResult(overviewMap, outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to format JSON: %w", err)
		}
		return contents, nil
	case "markdown":
		formattedOutput = formatOverviewMarkdownProto(overviewProto)
	case "html":
		formattedOutput = formatOverviewHTMLProto(overviewProto)
	default:
		formattedOutput = formatOverviewTextProto(overviewProto)
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

// handleReportBriefing handles the briefing action for report tool.
// Uses proto internally (BuildBriefingDataProto) for type-safe briefing data.
func handleReportBriefing(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	var score float64

	if sc, ok := params["score"].(float64); ok {
		score = sc
	} else if sc, ok := params["score"].(int); ok {
		score = float64(sc)
	} else {
		score = 50.0 // Default score
	}

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	engine, err := getWisdomEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize wisdom engine: %w", err)
	}

	// Build briefing from proto (type-safe)
	briefingProto := BuildBriefingDataProto(engine, score)
	briefingMap := BriefingDataToMap(briefingProto)

	return response.FormatResult(briefingMap, "")
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
func aggregateProjectData(ctx context.Context, projectRoot string, includePlanning bool) (map[string]interface{}, error) {
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
				"production_ready": scorecard.Score >= float64(config.MinCoverage()), // Using coverage threshold as production ready indicator
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

	// Optional: planning snippet (critical path + first N backlog order)
	if includePlanning {
		planning := getPlanningSnippet(projectRoot)
		if planning != nil {
			data["planning"] = planning
		}
	}

	data["generated_at"] = time.Now().Format(time.RFC3339)

	return data, nil
}

// aggregateProjectDataProto returns overview data as proto for type-safe report formatting.
func aggregateProjectDataProto(ctx context.Context, projectRoot string, includePlanning bool) (*proto.ProjectOverviewData, error) {
	pb := &proto.ProjectOverviewData{}

	if projectInfo, err := getProjectInfo(projectRoot); err == nil {
		pb.Project = ProjectInfoToProto(projectInfo)
	}

	if IsGoProject() {
		opts := &ScorecardOptions{FastMode: true}
		scorecard, err := GenerateGoScorecard(ctx, projectRoot, opts)
		if err == nil {
			scores := map[string]float64{
				"security":      calculateSecurityScore(scorecard),
				"testing":       calculateTestingScore(scorecard),
				"documentation": calculateDocumentationScore(scorecard),
				"completion":    calculateCompletionScore(scorecard),
			}
			scores["alignment"] = 50.0
			pb.Health = &proto.HealthData{
				OverallScore:    scorecard.Score,
				ProductionReady: scorecard.Score >= float64(config.MinCoverage()),
				Scores:          scores,
			}
		}
	}

	if codebase, err := getCodebaseMetrics(projectRoot); err == nil {
		pb.Codebase = CodebaseMetricsToProto(codebase)
	}

	if tasks, err := getTaskMetrics(projectRoot); err == nil {
		pb.Tasks = TaskMetricsToProto(tasks)
	}

	phasesMap := getProjectPhases()
	for _, phaseRaw := range phasesMap {
		if phase, ok := phaseRaw.(map[string]interface{}); ok {
			pbPhase := &proto.ProjectPhase{}
			if name, ok := phase["name"].(string); ok {
				pbPhase.Name = name
			}
			if status, ok := phase["status"].(string); ok {
				pbPhase.Status = status
			}
			if progress, ok := phase["progress"].(int); ok {
				pbPhase.Progress = int32(progress)
			} else if progress, ok := phase["progress"].(float64); ok {
				pbPhase.Progress = int32(progress)
			}
			pb.Phases = append(pb.Phases, pbPhase)
		}
	}

	if risks, err := getRisksAndBlockers(projectRoot); err == nil {
		for _, r := range risks {
			pb.Risks = append(pb.Risks, &proto.RiskOrBlocker{
				Type:        getStr(r, "type"),
				Description: getStr(r, "description"),
				TaskId:      getStr(r, "task_id"),
				Priority:    getStr(r, "priority"),
			})
		}
	}

	if actions, err := getNextActions(projectRoot); err == nil {
		for _, a := range actions {
			hours := 0.0
			if h, ok := a["estimated_hours"].(float64); ok {
				hours = h
			}
			pb.NextActions = append(pb.NextActions, &proto.NextAction{
				TaskId:         getStr(a, "task_id"),
				Name:           getStr(a, "name"),
				Priority:       getStr(a, "priority"),
				EstimatedHours: hours,
			})
		}
	}

	pb.GeneratedAt = time.Now().Format(time.RFC3339)

	if includePlanning {
		planning := getPlanningSnippet(projectRoot)
		if planning != nil {
			pb.Planning = &proto.PlanningSnippet{}
			if s, ok := planning["critical_path_summary"].(string); ok {
				pb.Planning.CriticalPathSummary = s
			}
			if s, ok := planning["suggested_backlog_summary"].(string); ok {
				pb.Planning.SuggestedBacklogSummary = s
			}
		}
	}

	return pb, nil
}

func getStr(m map[string]interface{}, key string) string {
	if s, ok := m[key].(string); ok {
		return s
	}
	return ""
}

// getPlanningSnippet returns critical path summary and suggested backlog order (first 10).
func getPlanningSnippet(projectRoot string) map[string]interface{} {
	out := make(map[string]interface{})

	// Critical path
	if cp, err := AnalyzeCriticalPath(projectRoot); err == nil {
		if hasCP, _ := cp["has_critical_path"].(bool); hasCP {
			if path, ok := cp["critical_path"].([]string); ok && len(path) > 0 {
				out["critical_path"] = path
				out["critical_path_summary"] = fmt.Sprintf("%d tasks: %s", len(path), strings.Join(path, ", "))
			}
		}
	}

	// First 10 in backlog execution order
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return out
	}
	orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil || len(orderedIDs) == 0 {
		return out
	}
	const n = 10
	if len(orderedIDs) > n {
		orderedIDs = orderedIDs[:n]
	}
	out["suggested_backlog_order"] = orderedIDs
	out["suggested_backlog_summary"] = fmt.Sprintf("First %d: %s", len(orderedIDs), strings.Join(orderedIDs, ", "))

	return out
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

	// Try to get from go.mod (using file cache)
	goModPath := filepath.Join(projectRoot, "go.mod")
	cache := cache.GetGlobalFileCache()
	if data, _, err := cache.ReadFile(goModPath); err == nil {
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
		"tools":        28, // From registry (28 base + 1 conditional Apple FM)
		"prompts":      34, // From templates.go (18 original + 16 migrated from Python)
		"resources":    21, // From resources/handlers.go
	}

	// Count tools, prompts, resources from registry
	// These would need to be exported from registry.go or counted differently
	// For now, use values matching MIGRATION_STATUS_CURRENT.md
	metrics["tools"] = 28
	metrics["prompts"] = 34
	metrics["resources"] = 21

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
			"status":   "complete",
			"progress": 100,
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

// isTaskReady returns true if task is in backlog and all its dependencies are Done.
func isTaskReady(task Todo2Task, taskMap map[string]Todo2Task) bool {
	for _, depID := range task.Dependencies {
		if dep, ok := taskMap[depID]; ok && !IsCompletedStatus(dep.Status) {
			return false
		}
	}
	return true
}

// getNextActions identifies next priority actions (dependency-aware: only suggests ready tasks).
func getNextActions(projectRoot string) ([]map[string]interface{}, error) {
	actions := []map[string]interface{}{}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return actions, nil
	}

	orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil {
		// Fallback: high/critical pending without order
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
		return actions, nil
	}

	taskMap := make(map[string]Todo2Task)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	const maxNext = 10
	for _, taskID := range orderedIDs {
		if len(actions) >= maxNext {
			break
		}
		task, ok := taskMap[taskID]
		if !ok || !IsBacklogStatus(task.Status) {
			continue
		}
		if !isTaskReady(task, taskMap) {
			continue
		}
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

	// Planning (optional)
	if planning, ok := data["planning"].(map[string]interface{}); ok {
		if summary, ok := planning["critical_path_summary"].(string); ok && summary != "" {
			sb.WriteString("Critical Path: " + summary + "\n\n")
		}
		if summary, ok := planning["suggested_backlog_summary"].(string); ok && summary != "" {
			sb.WriteString("Suggested Backlog Order: " + summary + "\n\n")
		}
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

	// Planning (optional)
	if planning, ok := data["planning"].(map[string]interface{}); ok {
		if summary, ok := planning["critical_path_summary"].(string); ok && summary != "" {
			sb.WriteString("## Planning\n\n")
			sb.WriteString("- **Critical Path**: " + summary + "\n\n")
		}
		if summary, ok := planning["suggested_backlog_summary"].(string); ok && summary != "" {
			sb.WriteString("- **Suggested Backlog Order**: " + summary + "\n\n")
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

// formatOverviewTextProto formats overview from proto (type-safe, no map assertions).
func formatOverviewTextProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}
	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  PROJECT OVERVIEW\n")
	sb.WriteString("======================================================================\n\n")

	if pb.Project != nil {
		sb.WriteString("Project Information:\n")
		sb.WriteString(fmt.Sprintf("  Name:        %s\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("  Version:     %s\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("  Type:        %s\n", pb.Project.Type))
		sb.WriteString(fmt.Sprintf("  Status:      %s\n", pb.Project.Status))
		sb.WriteString("\n")
	}

	if pb.Health != nil {
		sb.WriteString("Health Scorecard:\n")
		sb.WriteString(fmt.Sprintf("  Overall Score: %.1f%%\n", pb.Health.OverallScore))
		if pb.Health.ProductionReady {
			sb.WriteString("  Production Ready: YES ✅\n")
		} else {
			sb.WriteString("  Production Ready: NO ❌\n")
		}
		sb.WriteString("\n")
	}

	if pb.Tasks != nil {
		sb.WriteString("Task Status:\n")
		sb.WriteString(fmt.Sprintf("  Total:           %d\n", pb.Tasks.Total))
		sb.WriteString(fmt.Sprintf("  Pending:        %d\n", pb.Tasks.Pending))
		sb.WriteString(fmt.Sprintf("  Completed:      %d\n", pb.Tasks.Completed))
		sb.WriteString(fmt.Sprintf("  Completion:     %.1f%%\n", pb.Tasks.CompletionRate))
		sb.WriteString(fmt.Sprintf("  Remaining Hours: %.1f\n", pb.Tasks.RemainingHours))
		sb.WriteString("\n")
	}

	if len(pb.NextActions) > 0 {
		sb.WriteString("Next Actions:\n")
		for i, action := range pb.NextActions {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("  %d. %s (Priority: %s)\n", i+1, action.Name, action.Priority))
		}
		sb.WriteString("\n")
	}

	if pb.Planning != nil {
		if pb.Planning.CriticalPathSummary != "" {
			sb.WriteString("Critical Path: " + pb.Planning.CriticalPathSummary + "\n\n")
		}
		if pb.Planning.SuggestedBacklogSummary != "" {
			sb.WriteString("Suggested Backlog Order: " + pb.Planning.SuggestedBacklogSummary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewMarkdownProto formats overview as markdown from proto.
func formatOverviewMarkdownProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}
	var sb strings.Builder

	sb.WriteString("# Project Overview\n\n")

	if pb.Project != nil {
		sb.WriteString("## Project Information\n\n")
		sb.WriteString(fmt.Sprintf("- **Name**: %s\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("- **Version**: %s\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("- **Type**: %s\n", pb.Project.Type))
		sb.WriteString(fmt.Sprintf("- **Status**: %s\n\n", pb.Project.Status))
	}

	if pb.Health != nil {
		sb.WriteString("## Health Scorecard\n\n")
		sb.WriteString(fmt.Sprintf("**Overall Score**: %.1f%%\n\n", pb.Health.OverallScore))
	}

	if pb.Tasks != nil {
		sb.WriteString("## Task Status\n\n")
		sb.WriteString(fmt.Sprintf("- **Total**: %d\n", pb.Tasks.Total))
		sb.WriteString(fmt.Sprintf("- **Pending**: %d\n", pb.Tasks.Pending))
		sb.WriteString(fmt.Sprintf("- **Completed**: %d\n", pb.Tasks.Completed))
		sb.WriteString(fmt.Sprintf("- **Completion Rate**: %.1f%%\n\n", pb.Tasks.CompletionRate))
	}

	if pb.Planning != nil {
		if pb.Planning.CriticalPathSummary != "" {
			sb.WriteString("## Planning\n\n")
			sb.WriteString("- **Critical Path**: " + pb.Planning.CriticalPathSummary + "\n\n")
		}
		if pb.Planning.SuggestedBacklogSummary != "" {
			sb.WriteString("- **Suggested Backlog Order**: " + pb.Planning.SuggestedBacklogSummary + "\n\n")
		}
	}

	return sb.String()
}

// formatOverviewHTMLProto formats overview as HTML from proto.
func formatOverviewHTMLProto(pb *proto.ProjectOverviewData) string {
	if pb == nil {
		return ""
	}
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	sb.WriteString("<title>Project Overview</title>\n")
	sb.WriteString("<style>body{font-family:Arial,sans-serif;margin:40px;}</style>\n")
	sb.WriteString("</head>\n<body>\n")
	sb.WriteString("<h1>Project Overview</h1>\n")

	if pb.Project != nil {
		sb.WriteString("<h2>Project Information</h2>\n<ul>\n")
		sb.WriteString(fmt.Sprintf("<li><strong>Name</strong>: %s</li>\n", pb.Project.Name))
		sb.WriteString(fmt.Sprintf("<li><strong>Version</strong>: %s</li>\n", pb.Project.Version))
		sb.WriteString(fmt.Sprintf("<li><strong>Type</strong>: %s</li>\n", pb.Project.Type))
		sb.WriteString("</ul>\n")
	}

	if pb.Health != nil {
		sb.WriteString("<h2>Health Scorecard</h2>\n")
		sb.WriteString(fmt.Sprintf("<p><strong>Overall Score</strong>: %.1f%%</p>\n", pb.Health.OverallScore))
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
