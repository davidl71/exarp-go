package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/security"
)

// handleAnalyzeAlignmentNative handles the analyze_alignment tool with native Go implementation
// Implements "todo2" action fully (including followup task creation), falls back to Python bridge for "prd" action
func handleAnalyzeAlignmentNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get action (default: "todo2")
	action := "todo2"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	switch action {
	case "todo2":
		return handleAlignmentTodo2(ctx, params)
	case "prd":
		// PRD alignment is complex, use Python bridge
		return nil, fmt.Errorf("PRD alignment not yet implemented in native Go, using Python bridge")
	default:
		return nil, fmt.Errorf("unknown alignment action: %s. Use 'todo2' or 'prd'", action)
	}
}

// handleAlignmentTodo2 handles the "todo2" action for analyze_alignment
func handleAlignmentTodo2(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get project root
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		// Fallback to PROJECT_ROOT env var or current directory
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	// Load tasks
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load Todo2 tasks: %w", err)
	}

	// Load project goals if available
	goalsPath := filepath.Join(projectRoot, "PROJECT_GOALS.md")
	goalsContent := ""
	if content, err := os.ReadFile(goalsPath); err == nil {
		goalsContent = string(content)
	}

	// Analyze alignment
	analysis := analyzeTaskAlignment(tasks, goalsContent)

	// Get optional parameters
	createFollowupTasks := true
	if createRaw, ok := params["create_followup_tasks"].(bool); ok {
		createFollowupTasks = createRaw
	}

	outputPath := ""
	if outputPathRaw, ok := params["output_path"].(string); ok && outputPathRaw != "" {
		outputPath = outputPathRaw
	} else {
		outputPath = filepath.Join(projectRoot, "docs", "TODO2_ALIGNMENT_REPORT.md")
	}

	// Create followup tasks if requested
	tasksCreated := 0
	if createFollowupTasks {
		tasksCreated = createAlignmentFollowupTasks(ctx, analysis)
	}

	// Save report if output path specified
	if outputPath != "" {
		report := generateAlignmentReport(analysis, projectRoot)
		reportPath := outputPath
		if !filepath.IsAbs(reportPath) {
			reportPath = filepath.Join(projectRoot, reportPath)
		}

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err == nil {
			os.WriteFile(reportPath, []byte(report), 0644)
		}
		analysis.ReportPath = reportPath
	}

	// Build response
	responseData := map[string]interface{}{
		"total_tasks_analyzed":    analysis.TotalTasks,
		"misaligned_count":        len(analysis.MisalignedTasks),
		"infrastructure_count":    len(analysis.InfrastructureTasks),
		"stale_count":             len(analysis.StaleTasks),
		"average_alignment_score": analysis.AlignmentScore,
		"by_priority":             analysis.ByPriority,
		"by_status":               analysis.ByStatus,
		"report_path":             analysis.ReportPath,
		"tasks_created":           tasksCreated,
		"status":                  "success",
	}

	resultJSON, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// AlignmentAnalysis represents the results of alignment analysis
type AlignmentAnalysis struct {
	TotalTasks          int
	MisalignedTasks     []Todo2Task
	InfrastructureTasks []Todo2Task
	StaleTasks          []Todo2Task
	AlignmentScore      float64
	ByPriority          map[string]int
	ByStatus            map[string]int
	ReportPath          string
}

// analyzeTaskAlignment analyzes task alignment with project goals
func analyzeTaskAlignment(tasks []Todo2Task, goalsContent string) AlignmentAnalysis {
	analysis := AlignmentAnalysis{
		TotalTasks:          len(tasks),
		MisalignedTasks:     []Todo2Task{},
		InfrastructureTasks: []Todo2Task{},
		StaleTasks:          []Todo2Task{},
		ByPriority:          make(map[string]int),
		ByStatus:            make(map[string]int),
	}

	// Extract key terms from goals
	goalTerms := extractGoalTerms(goalsContent)

	// Analyze each task
	alignedCount := 0
	for _, task := range tasks {
		// Count by priority and status
		priority := task.Priority
		if priority == "" {
			priority = "medium"
		}
		analysis.ByPriority[priority]++

		status := normalizeStatus(task.Status)
		analysis.ByStatus[status]++

		// Check alignment
		isAligned := checkTaskAlignment(task, goalTerms)
		if isAligned {
			alignedCount++
		} else {
			// Check if it's infrastructure or stale
			if isInfrastructureTask(task) {
				analysis.InfrastructureTasks = append(analysis.InfrastructureTasks, task)
			} else if isStaleTask(task) {
				analysis.StaleTasks = append(analysis.StaleTasks, task)
			} else {
				analysis.MisalignedTasks = append(analysis.MisalignedTasks, task)
			}
		}
	}

	// Calculate alignment score
	if len(tasks) > 0 {
		analysis.AlignmentScore = float64(alignedCount) / float64(len(tasks)) * 100.0
	}

	return analysis
}

// extractGoalTerms extracts key terms from project goals
func extractGoalTerms(goalsContent string) []string {
	if goalsContent == "" {
		return []string{}
	}

	terms := []string{}
	content := strings.ToLower(goalsContent)

	// Look for common goal keywords
	keywords := []string{
		"migrate", "migration", "go", "native", "performance",
		"framework", "tool", "server", "mcp", "automation",
		"test", "testing", "documentation", "security",
	}

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			terms = append(terms, keyword)
		}
	}

	return terms
}

// checkTaskAlignment checks if a task aligns with project goals
func checkTaskAlignment(task Todo2Task, goalTerms []string) bool {
	if len(goalTerms) == 0 {
		// No goals available, consider all tasks aligned
		return true
	}

	// Combine task content and description
	taskText := strings.ToLower(task.Content + " " + task.LongDescription)

	// Check if task mentions any goal terms
	for _, term := range goalTerms {
		if strings.Contains(taskText, term) {
			return true
		}
	}

	// Check tags for alignment
	for _, tag := range task.Tags {
		tagLower := strings.ToLower(tag)
		for _, term := range goalTerms {
			if strings.Contains(tagLower, term) {
				return true
			}
		}
	}

	return false
}

// isInfrastructureTask checks if a task is infrastructure-related
func isInfrastructureTask(task Todo2Task) bool {
	infraKeywords := []string{"infrastructure", "ci/cd", "cicd", "deployment", "docker", "kubernetes", "setup", "config"}
	taskText := strings.ToLower(task.Content + " " + task.LongDescription)

	for _, keyword := range infraKeywords {
		if strings.Contains(taskText, keyword) {
			return true
		}
	}

	return false
}

// isStaleTask checks if a task is stale (old and not updated)
func isStaleTask(task Todo2Task) bool {
	// For now, consider tasks with "stale" tag or very old tasks as stale
	// This is a simplified check - full implementation would check dates
	for _, tag := range task.Tags {
		if strings.ToLower(tag) == "stale" {
			return true
		}
	}

	return false
}

// generateAlignmentReport generates a markdown report
func generateAlignmentReport(analysis AlignmentAnalysis, projectRoot string) string {
	report := fmt.Sprintf(`# Todo2 Alignment Analysis Report

**Generated:** %s

## Summary

- **Total Tasks Analyzed:** %d
- **Alignment Score:** %.1f%%
- **Misaligned Tasks:** %d
- **Infrastructure Tasks:** %d
- **Stale Tasks:** %d

## Alignment by Priority

`,
		"2026-01-09", // TODO: Use actual timestamp
		analysis.TotalTasks,
		analysis.AlignmentScore,
		len(analysis.MisalignedTasks),
		len(analysis.InfrastructureTasks),
		len(analysis.StaleTasks),
	)

	for priority, count := range analysis.ByPriority {
		report += fmt.Sprintf("- **%s:** %d\n", priority, count)
	}

	report += "\n## Alignment by Status\n\n"
	for status, count := range analysis.ByStatus {
		report += fmt.Sprintf("- **%s:** %d\n", status, count)
	}

	if len(analysis.MisalignedTasks) > 0 {
		report += "\n## Misaligned Tasks\n\n"
		for _, task := range analysis.MisalignedTasks {
			report += fmt.Sprintf("- %s (%s)\n", task.Content, task.ID)
		}
	}

	return report
}

// createAlignmentFollowupTasks creates Todo2 tasks for alignment issues
func createAlignmentFollowupTasks(ctx context.Context, analysis AlignmentAnalysis) int {
	tasksCreated := 0

	// Create task for misaligned tasks
	if len(analysis.MisalignedTasks) > 0 {
		var description strings.Builder
		description.WriteString(fmt.Sprintf("Review and align %d misaligned tasks with project goals:\n\n", len(analysis.MisalignedTasks)))
		for _, task := range analysis.MisalignedTasks {
			description.WriteString(fmt.Sprintf("- %s (%s): %s\n", task.ID, task.Status, task.Content))
		}

		taskID := generateEpochTaskID()
		task := &models.Todo2Task{
			ID:      taskID,
			Content: "Review misaligned tasks",
			Status:  "Todo",
			Priority: "medium",
			Tags:    []string{"alignment", "review"},
			Metadata: map[string]interface{}{
				"misaligned_count": len(analysis.MisalignedTasks),
				"alignment_score":  analysis.AlignmentScore,
			},
		}

		if err := database.CreateTask(ctx, task); err == nil {
			comment := database.Comment{
				TaskID:  taskID,
				Type:    "description",
				Content: description.String(),
			}
			_ = database.AddComments(ctx, taskID, []database.Comment{comment})
			tasksCreated++
		}
	}

	// Create task for stale tasks
	if len(analysis.StaleTasks) > 0 {
		var description strings.Builder
		description.WriteString(fmt.Sprintf("Review %d stale tasks that may need updating or removal:\n\n", len(analysis.StaleTasks)))
		for _, task := range analysis.StaleTasks {
			description.WriteString(fmt.Sprintf("- %s (%s): %s\n", task.ID, task.Status, task.Content))
		}

		taskID := generateEpochTaskID()
		task := &models.Todo2Task{
			ID:      taskID,
			Content: "Review stale tasks",
			Status:  "Todo",
			Priority: "low",
			Tags:    []string{"alignment", "cleanup"},
			Metadata: map[string]interface{}{
				"stale_count": len(analysis.StaleTasks),
			},
		}

		if err := database.CreateTask(ctx, task); err == nil {
			comment := database.Comment{
				TaskID:  taskID,
				Type:    "description",
				Content: description.String(),
			}
			_ = database.AddComments(ctx, taskID, []database.Comment{comment})
			tasksCreated++
		}
	}

	return tasksCreated
}
