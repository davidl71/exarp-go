// alignment_analysis.go — MCP "analyze_alignment" tool: todo2 and PRD alignment checks.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/models"
)

// prdPersona holds persona metadata for PRD alignment (matches Python prd_generator.PERSONAS).
type prdPersona struct {
	Name     string
	Advisor  string
	Keywords []string
}

// prdPersonas defines personas used for task-to-persona alignment (subset of Python PERSONAS).
var prdPersonas = map[string]prdPersona{
	"developer":         {Name: "Developer", Advisor: "tao_of_programming", Keywords: []string{"code", "implement", "fix", "build", "debug", "api", "endpoint", "mcp", "tool"}},
	"project_manager":   {Name: "Project Manager", Advisor: "art_of_war", Keywords: []string{"sprint", "planning", "status", "delivery", "blockers", "progress", "schedule"}},
	"code_reviewer":     {Name: "Code Reviewer", Advisor: "stoic", Keywords: []string{"pr", "review", "approve", "merge", "quality", "standards"}},
	"architect":         {Name: "Architect", Advisor: "enochian", Keywords: []string{"design", "architecture", "coupling", "patterns", "structure", "scalability"}},
	"security_engineer": {Name: "Security Engineer", Advisor: "bofh", Keywords: []string{"vulnerability", "security", "cve", "audit", "scan", "risk"}},
	"qa_engineer":       {Name: "QA Engineer", Advisor: "stoic", Keywords: []string{"test", "coverage", "quality", "defect", "validation", "qa"}},
	"executive":         {Name: "Executive/Stakeholder", Advisor: "pistis_sophia", Keywords: []string{"status", "summary", "overview", "stakeholder", "report", "dashboard"}},
	"tech_writer":       {Name: "Technical Writer", Advisor: "confucius", Keywords: []string{"doc", "documentation", "docs", "readme", "guide", "tutorial"}},
}

// handleAnalyzeAlignmentNative handles the analyze_alignment tool with native Go implementation
// Implements both "todo2" and "prd" actions natively.
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
		return handleAlignmentPRD(ctx, params)
	default:
		return nil, fmt.Errorf("unknown alignment action: %s. Use 'todo2' or 'prd'", action)
	}
}

// handleAlignmentTodo2 handles the "todo2" action for analyze_alignment.
func handleAlignmentTodo2(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get project root
	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		return nil, fmt.Errorf("project root: %w", err)
	}

	// Load tasks via TaskStore
	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load Todo2 tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

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

	return framework.FormatResult(responseData, "")
}

// handleAlignmentPRD handles the "prd" action: task-to-persona alignment using PRD.md and persona keywords.
func handleAlignmentPRD(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := GetProjectRootWithFallback()
	if err != nil {
		return nil, fmt.Errorf("project root: %w", err)
	}

	store := NewDefaultTaskStore(projectRoot)

	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load Todo2 tasks: %w", err)
	}

	tasks := tasksFromPtrs(list)

	prdPath := filepath.Join(projectRoot, "docs", "PRD.md")
	prdExists := false

	if _, err := os.Stat(prdPath); err == nil {
		prdExists = true
	}

	prdUserStories := 0
	if prdExists {
		prdUserStories = parsePRDUserStoriesCount(prdPath)
	}

	personaCounts := make(map[string]int)
	for id := range prdPersonas {
		personaCounts[id] = 0
	}

	var aligned []map[string]interface{}

	var unaligned []map[string]interface{}

	for _, task := range tasks {
		bestPersona, bestScore, bestAdvisor, reason := prdScoreTask(task)
		if bestScore >= 2 {
			personaCounts[bestPersona]++
			aligned = append(aligned, map[string]interface{}{
				"id": task.ID, "name": task.Content,
				"persona": bestPersona, "advisor": bestAdvisor, "alignment_score": bestScore,
			})
		} else {
			unaligned = append(unaligned, map[string]interface{}{
				"id": task.ID, "name": task.Content, "reason": reason,
			})
		}
	}

	overallScore := 0.0
	if len(tasks) > 0 {
		overallScore = float64(len(aligned)) / float64(len(tasks)) * 100.0
	}

	recommendations := prdRecommendations(personaCounts, unaligned)

	alignmentByPersona := make(map[string]interface{})

	for id, count := range personaCounts {
		if count > 0 {
			p := prdPersonas[id]
			alignmentByPersona[id] = map[string]interface{}{
				"name": p.Name, "count": count, "advisor": p.Advisor,
			}
		}
	}

	unalignedTop10 := unaligned
	if len(unalignedTop10) > 10 {
		unalignedTop10 = unalignedTop10[:10]
	}

	data := map[string]interface{}{
		"timestamp":               time.Now().Format(time.RFC3339),
		"prd_exists":              prdExists,
		"goals_exists":            fileExists(filepath.Join(projectRoot, "PROJECT_GOALS.md")),
		"tasks_analyzed":          len(tasks),
		"prd_user_stories":        prdUserStories,
		"persona_coverage":        personaCounts,
		"aligned_count":           len(aligned),
		"unaligned_count":         len(unaligned),
		"unaligned_tasks":         unalignedTop10,
		"overall_alignment_score": roundFloat(overallScore, 1),
		"recommendations":         recommendations,
		"alignment_by_persona":    alignmentByPersona,
	}

	outputPath, _ := params["output_path"].(string)
	if outputPath != "" {
		fullPath := outputPath
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(projectRoot, fullPath)
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err == nil {
			raw, _ := json.MarshalIndent(data, "", "  ")
			_ = os.WriteFile(fullPath, raw, 0644)
			data["report_path"] = fullPath
		}
	}

	envelope := map[string]interface{}{
		"success":   true,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	return framework.FormatResult(envelope, "")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func roundFloat(v float64, decimals int) float64 {
	mult := 1.0
	for i := 0; i < decimals; i++ {
		mult *= 10
	}

	return float64(int(v*mult+0.5)) / mult
}

// parsePRDUserStoriesCount returns the number of user stories in PRD.md (## N. User Stories / ### US-N: title).
func parsePRDUserStoriesCount(prdPath string) int {
	content, err := os.ReadFile(prdPath)
	if err != nil {
		return 0
	}
	// Match ### US-<num>: <title>
	re := regexp.MustCompile(`### US-\d+:\s*.+`)
	matches := re.FindAll(content, -1)

	return len(matches)
}

// prdScoreTask scores a task against personas; returns best persona id, score, advisor, and reason if unaligned.
func prdScoreTask(task Todo2Task) (personaID string, score int, advisor, reason string) {
	content := strings.ToLower(task.Content + " " + task.LongDescription)

	tags := make([]string, 0, len(task.Tags))
	for _, t := range task.Tags {
		tags = append(tags, strings.ToLower(t))
	}

	bestScore := 0

	for id, p := range prdPersonas {
		s := 0

		for _, kw := range p.Keywords {
			if strings.Contains(content, kw) {
				s += 2
			}

			for _, tag := range tags {
				if strings.Contains(tag, kw) {
					s += 3
					break
				}
			}
		}

		if s > bestScore {
			bestScore = s
			personaID = id
			advisor = p.Advisor
		}
	}

	if bestScore >= 2 {
		return personaID, bestScore, advisor, ""
	}

	return "", 0, "", "No strong persona match found"
}

func prdRecommendations(personaCounts map[string]int, unaligned []map[string]interface{}) []string {
	var recs []string

	for id, count := range personaCounts {
		if count == 0 {
			if p, ok := prdPersonas[id]; ok {
				recs = append(recs, fmt.Sprintf("No tasks aligned with %s persona - consider their needs", p.Name))
			}
		}
	}

	total := 0
	for _, c := range personaCounts {
		total += c
	}

	total += len(unaligned)
	if total > 0 {
		ratio := float64(len(unaligned)) / float64(total)
		if ratio > 0.2 {
			recs = append(recs, fmt.Sprintf("%d tasks (%.0f%%) lack persona alignment - add relevant tags", len(unaligned), ratio*100))
		}
	}

	maxCount := 0
	minCount := -1

	for _, c := range personaCounts {
		if c > maxCount {
			maxCount = c
		}

		if minCount < 0 || c < minCount {
			minCount = c
		}
	}

	if maxCount > 0 && minCount == 0 {
		recs = append(recs, "Task distribution is unbalanced across personas")
	}

	return recs
}

// AlignmentAnalysis represents the results of alignment analysis.
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

// analyzeTaskAlignment analyzes task alignment with project goals.
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

// extractGoalTerms extracts key terms from project goals.
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

// checkTaskAlignment checks if a task aligns with project goals.
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

// isInfrastructureTask checks if a task is infrastructure-related.
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

// isStaleTask checks if a task is stale (old and not updated).
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

// generateAlignmentReport generates a markdown report.
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
		time.Now().Format(time.RFC3339),
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

// createAlignmentFollowupTasks creates Todo2 tasks for alignment issues.
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
			ID:       taskID,
			Content:  "Review misaligned tasks",
			Status:   models.StatusTodo,
			Priority: "medium",
			Tags:     []string{"alignment", "review"},
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
			ID:       taskID,
			Content:  "Review stale tasks",
			Status:   models.StatusTodo,
			Priority: "low",
			Tags:     []string{"alignment", "cleanup"},
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
