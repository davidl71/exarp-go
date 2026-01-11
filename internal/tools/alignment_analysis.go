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

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
)

// handleAnalyzeAlignmentNative handles the analyze_alignment tool with native Go implementation
// Currently implements basic "todo2" action, falls back to Python bridge for "prd" and complex analysis
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
	if createFollowupTasks && len(analysis.MisalignedTasks) > 0 {
		// For now, we'll skip creating tasks in native Go (complex logic)
		// This can be added later or left to Python bridge
		tasksCreated = 0
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

// handleAlignmentPRD handles the "prd" action for analyze_alignment
// Analyzes task alignment against PRD personas and user stories
func handleAlignmentPRD(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
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

	// Check if PRD.md exists
	prdPath := filepath.Join(projectRoot, "docs", "PRD.md")
	goalsPath := filepath.Join(projectRoot, "PROJECT_GOALS.md")

	results := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"prd_exists":       false,
		"goals_exists":     false,
		"tasks_analyzed":   0,
		"persona_coverage": map[string]int{},
		"unaligned_tasks":  []interface{}{},
		"alignment_by_persona": map[string]interface{}{},
		"recommendations": []string{},
	}

	// Check file existence
	if _, err := os.Stat(prdPath); err == nil {
		results["prd_exists"] = true
	}
	if _, err := os.Stat(goalsPath); err == nil {
		results["goals_exists"] = true
	}

	if !results["prd_exists"].(bool) {
		results["recommendations"] = []string{"Run generate_prd to create PRD.md for persona-based alignment"}
		resultJSON, _ := json.MarshalIndent(results, "", "  ")
		return []framework.TextContent{
			{Type: "text", Text: string(resultJSON)},
		}, nil
	}

	// Load tasks
	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load Todo2 tasks: %w", err)
	}

	results["tasks_analyzed"] = len(tasks)

	// Parse PRD for user stories
	prdContent, _ := os.ReadFile(prdPath)
	userStories := parsePRDUserStories(string(prdContent))
	results["prd_user_stories"] = len(userStories)

	// Get default personas (simplified - hard-coded for now)
	personas := getDefaultPersonas()

	// Analyze each task
	personaCounts := make(map[string]int)
	for pid := range personas {
		personaCounts[pid] = 0
	}

	alignedTasks := []interface{}{}
	unalignedTasks := []interface{}{}

	for _, task := range tasks {
		alignment := analyzeTaskPRDAlignment(task, personas)

		if alignment["persona"] != nil && alignment["persona"].(string) != "" {
			personaID := alignment["persona"].(string)
			personaCounts[personaID]++
			alignedTasks = append(alignedTasks, map[string]interface{}{
				"id":             task.ID,
				"name":           task.Content,
				"persona":        personaID,
				"advisor":        alignment["advisor"],
				"alignment_score": alignment["score"],
			})
		} else {
			unalignedTasks = append(unalignedTasks, map[string]interface{}{
				"id":     task.ID,
				"name":   task.Content,
				"reason": alignment["reason"],
			})
		}
	}

	results["persona_coverage"] = personaCounts
	results["aligned_count"] = len(alignedTasks)
	results["unaligned_count"] = len(unalignedTasks)

	// Top 10 unaligned tasks
	if len(unalignedTasks) > 10 {
		results["unaligned_tasks"] = unalignedTasks[:10]
	} else {
		results["unaligned_tasks"] = unalignedTasks
	}

	// Calculate alignment score
	if len(tasks) > 0 {
		results["overall_alignment_score"] = roundFloat(float64(len(alignedTasks))/float64(len(tasks))*100, 1)
	} else {
		results["overall_alignment_score"] = 0.0
	}

	// Generate recommendations
	results["recommendations"] = generatePRDRecommendations(personaCounts, unalignedTasks, personas)

	// Add persona details
	alignmentByPersona := make(map[string]interface{})
	for pid, count := range personaCounts {
		if count > 0 {
			persona := personas[pid]
			alignmentByPersona[pid] = map[string]interface{}{
				"name":    persona["name"],
				"count":   count,
				"advisor": persona["advisor"],
			}
		}
	}
	results["alignment_by_persona"] = alignmentByPersona

	// Get output path
	outputPath := ""
	if outputPathRaw, ok := params["output_path"].(string); ok && outputPathRaw != "" {
		outputPath = outputPathRaw
	}

	// Save report if requested
	if outputPath != "" {
		if !filepath.IsAbs(outputPath) {
			outputPath = filepath.Join(projectRoot, outputPath)
		}

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err == nil {
			reportData, _ := json.MarshalIndent(results, "", "  ")
			os.WriteFile(outputPath, reportData, 0644)
			results["report_path"] = outputPath
		}
	}

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// parsePRDUserStories parses user stories from PRD.md content
func parsePRDUserStories(prdContent string) []map[string]string {
	userStories := []map[string]string{}

	// Find User Stories section
	storySectionRegex := regexp.MustCompile(`(?i)##\s+\d+\.\s*User\s+Stories\s*\n(.+?)(?=\n##\s+\d+\.|\Z)`)
	matches := storySectionRegex.FindStringSubmatch(prdContent)

	if len(matches) < 2 {
		return userStories
	}

	storiesText := matches[1]

	// Parse individual stories (US-1: Title format)
	storyRegex := regexp.MustCompile(`(?i)###?\s*US-(\d+):\s*(.+?)\n`)
	storyMatches := storyRegex.FindAllStringSubmatch(storiesText, -1)

	for _, match := range storyMatches {
		if len(match) >= 3 {
			userStories = append(userStories, map[string]string{
				"id":    fmt.Sprintf("US-%s", match[1]),
				"title": strings.TrimSpace(match[2]),
			})
		}
	}

	return userStories
}

// getDefaultPersonas returns default persona definitions
func getDefaultPersonas() map[string]map[string]interface{} {
	return map[string]map[string]interface{}{
		"developer": {
			"name":    "Developer",
			"keywords": []string{"implement", "code", "refactor", "migrate", "test", "bug", "fix", "feature"},
			"advisor": "sage",
		},
		"architect": {
			"name":    "Architect",
			"keywords": []string{"architecture", "design", "system", "framework", "structure", "pattern", "scalability"},
			"advisor": "sage",
		},
		"qa": {
			"name":    "QA Engineer",
			"keywords": []string{"test", "testing", "quality", "coverage", "validation", "verify", "bug"},
			"advisor": "sage",
		},
		"devops": {
			"name":    "DevOps Engineer",
			"keywords": []string{"deploy", "ci/cd", "cicd", "infrastructure", "docker", "kubernetes", "automation"},
			"advisor": "sage",
		},
		"product": {
			"name":    "Product Manager",
			"keywords": []string{"feature", "requirement", "user", "story", "roadmap", "priority", "release"},
			"advisor": "sage",
		},
		"security": {
			"name":    "Security Engineer",
			"keywords": []string{"security", "vulnerability", "scan", "audit", "compliance", "encryption", "auth"},
			"advisor": "sage",
		},
	}
}

// analyzeTaskPRDAlignment analyzes a single task's alignment with personas
func analyzeTaskPRDAlignment(task Todo2Task, personas map[string]map[string]interface{}) map[string]interface{} {
	content := strings.ToLower(task.Content + " " + task.LongDescription)
	tags := make([]string, len(task.Tags))
	for i, tag := range task.Tags {
		tags[i] = strings.ToLower(tag)
	}

	// Score each persona
	bestPersona := ""
	bestScore := 0
	bestAdvisor := "sage"

	for personaID, personaData := range personas {
		score := 0

		// Get keywords
		keywordsRaw, ok := personaData["keywords"].([]string)
		if !ok {
			continue
		}

		// Keyword matching
		keywordMatches := 0
		for _, kw := range keywordsRaw {
			if strings.Contains(content, strings.ToLower(kw)) {
				keywordMatches++
			}
		}
		score += keywordMatches * 2

		// Tag matching
		tagMatches := 0
		for _, kw := range keywordsRaw {
			for _, tag := range tags {
				if strings.Contains(tag, strings.ToLower(kw)) {
					tagMatches++
					break
				}
			}
		}
		score += tagMatches * 3

		if score > bestScore {
			bestScore = score
			bestPersona = personaID
			if advisor, ok := personaData["advisor"].(string); ok {
				bestAdvisor = advisor
			}
		}
	}

	if bestScore >= 2 { // Threshold for alignment
		return map[string]interface{}{
			"persona": bestPersona,
			"advisor": bestAdvisor,
			"score":   bestScore,
			"reason":  nil,
		}
	}

	return map[string]interface{}{
		"persona": nil,
		"advisor": nil,
		"score":   0,
		"reason":  "No strong persona match found",
	}
}

// generatePRDRecommendations generates alignment recommendations
func generatePRDRecommendations(personaCounts map[string]int, unalignedTasks []interface{}, personas map[string]map[string]interface{}) []string {
	recommendations := []string{}

	// Check for underrepresented personas
	for personaID, count := range personaCounts {
		if count == 0 {
			if persona, ok := personas[personaID]; ok {
				if name, ok := persona["name"].(string); ok {
					recommendations = append(recommendations, fmt.Sprintf("No tasks aligned with %s persona - consider their needs", name))
				}
			}
		}
	}

	// Check unaligned ratio
	total := 0
	for _, count := range personaCounts {
		total += count
	}
	total += len(unalignedTasks)

	if total > 0 {
		unalignedRatio := float64(len(unalignedTasks)) / float64(total)
		if unalignedRatio > 0.2 {
			recommendations = append(recommendations, fmt.Sprintf("%d tasks (%.0f%%) lack persona alignment - add relevant tags", len(unalignedTasks), unalignedRatio*100))
		}
	}

	// Check for imbalance
	if len(personaCounts) > 0 {
		maxCount := 0
		minCount := 0
		first := true
		for _, count := range personaCounts {
			if first || count > maxCount {
				maxCount = count
			}
			if first || count < minCount {
				minCount = count
			}
			first = false
		}

		if maxCount > 0 && minCount == 0 {
			recommendations = append(recommendations, "Task distribution is unbalanced across personas")
		}
	}

	return recommendations
}
