// report_data.go — report data helpers.
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
	"github.com/davidl71/exarp-go/internal/models"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

// aggregateProjectDataProto returns overview data as proto for type-safe report formatting.
func aggregateProjectDataProto(ctx context.Context, projectRoot string, includePlanning bool) (*proto.ProjectOverviewData, error) {
	pb := &proto.ProjectOverviewData{}
	var healthWarning string

	if projectInfo, err := getProjectInfo(projectRoot); err == nil {
		pb.Project = ProjectInfoToProto(projectInfo)
	}

	if isGoProjectRoot(projectRoot) {
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
		} else {
			healthWarning = fmt.Sprintf("scorecard generation failed: %v", err)
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

	if healthWarning != "" {
		if pb.Project == nil {
			pb.Project = &proto.ProjectInfo{}
		}
		if pb.Project.Description != "" {
			pb.Project.Description += " | "
		}
		pb.Project.Description += "Health warning: " + healthWarning
	}

	return pb, nil
}

func isGoProjectRoot(projectRoot string) bool {
	if projectRoot == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(projectRoot, "go.mod"))
	return err == nil
}

func getStr(m map[string]interface{}, key string) string {
	return cast.ToString(m[key])
}

// planFilenameFromTitle returns a safe filename stem from a plan title (e.g. "github.com/davidl71/exarp-go" -> "exarp-go").
func planFilenameFromTitle(title string) string {
	if title == "" {
		return "plan"
	}

	stem := title
	if strings.Contains(title, "/") {
		stem = filepath.Base(title)
	}

	stem = strings.ReplaceAll(stem, " ", "-")
	stem = strings.ReplaceAll(stem, ":", "-")

	return strings.TrimSpace(stem)
}

// ensurePlanMdSuffix ensures path ends with .plan.md. If path ends with .md but not .plan.md, replaces with .plan.md; otherwise appends .plan.md.
func ensurePlanMdSuffix(path string) string {
	if path == "" {
		return path
	}

	path = filepath.Clean(path)
	base := filepath.Base(path)

	if strings.HasSuffix(strings.ToLower(base), ".plan.md") {
		return path
	}

	if strings.HasSuffix(strings.ToLower(base), ".md") {
		return filepath.Join(filepath.Dir(path), base[:len(base)-3]+".plan.md")
	}

	return filepath.Join(filepath.Dir(path), base+".plan.md")
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
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return out
	}

	tasks := tasksFromPtrs(list)

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

// getProjectInfo extracts project metadata.
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

// getCodebaseMetrics collects codebase statistics.
func getCodebaseMetrics(projectRoot string) (map[string]interface{}, error) {
	var goFiles, pythonFiles, cppFiles, rustFiles, totalFiles int

	// Count files (simplified - could use more sophisticated counting)
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			// Skip hidden and vendor directories
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" || info.Name() == "node_modules" || info.Name() == "build" || info.Name() == "target" || info.Name() == "__pycache__" {
				return filepath.SkipDir
			}

			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go":
			goFiles++
		case ".py":
			pythonFiles++
		case ".cpp", ".cc", ".cxx", ".c", ".h", ".hpp", ".hxx":
			cppFiles++
		case ".rs":
			rustFiles++
		}

		totalFiles++

		return nil
	})

	// Keep these metadata counts aligned with the current native registries.
	// Resources remain a fixed count here because tools cannot import internal/resources
	// without a package cycle.
	const expectedResourceCount = 24

	metrics := map[string]interface{}{
		"go_files":     goFiles,
		"go_lines":     0, // Could count lines if needed
		"python_files": pythonFiles,
		"python_lines": 0, // Could count lines if needed
		"cpp_files":    cppFiles,
		"rust_files":   rustFiles,
		"total_files":  totalFiles,
		"total_lines":  0, // Could count lines if needed
		"tools":        ExpectedToolCountBase,
		"prompts":      len(prompts.ListAllPromptNames()),
		"resources":    expectedResourceCount,
	}

	return metrics, err
}

// getTaskMetrics collects task statistics.
func getTaskMetrics(projectRoot string) (map[string]interface{}, error) {
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	tasks := tasksFromPtrs(list)

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

// getProjectPhases returns current phase status.
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

// getRisksAndBlockers identifies project risks.
func getRisksAndBlockers(projectRoot string) ([]map[string]interface{}, error) {
	risks := []map[string]interface{}{}

	// Check for incomplete tasks with high priority
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err == nil {
		tasks := tasksFromPtrs(list)
		for _, task := range tasks {
			if IsPendingStatus(task.Status) && task.Priority == models.PriorityCritical {
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

	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(context.Background(), nil)
	if err != nil {
		return actions, nil
	}

	tasks := tasksFromPtrs(list)

	orderedIDs, _, _, err := BacklogExecutionOrder(tasks, nil)
	if err != nil {
		// Fallback: high/critical pending without order
		for _, task := range tasks {
			if IsPendingStatus(task.Status) && (task.Priority == models.PriorityHigh || task.Priority == models.PriorityCritical) {
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

// generatePRD generates a Product Requirements Document.
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

		store := NewDefaultTaskStore(projectRoot)

		list, err := store.ListTasks(ctx, nil)
		if err == nil {
			tasks := tasksFromPtrs(list)
			// Group tasks by priority
			criticalTasks := []Todo2Task{}
			highTasks := []Todo2Task{}
			mediumTasks := []Todo2Task{}

			for _, task := range tasks {
				if IsPendingStatus(task.Status) {
					switch task.Priority {
					case models.PriorityCritical:
						criticalTasks = append(criticalTasks, task)
					case models.PriorityHigh:
						highTasks = append(highTasks, task)
					case models.PriorityMedium:
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

// formatOverviewText formats overview as plain text.
