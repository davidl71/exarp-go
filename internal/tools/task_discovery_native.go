//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	fm "github.com/blacktop/go-foundationmodels"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
)

// handleTaskDiscoveryNative handles task_discovery with native Go and Apple FM
func handleTaskDiscoveryNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "all"
	}

	// Check platform support for Apple FM (used for semantic extraction)
	support := platform.CheckAppleFoundationModelsSupport()
	useAppleFM := support.Supported

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	discoveries := []map[string]interface{}{}

	// Scan comments
	if action == "comments" || action == "all" {
		filePatterns := []string{"**/*.go", "**/*.py", "**/*.js", "**/*.ts"}
		if patterns, ok := params["file_patterns"].(string); ok && patterns != "" {
			var parsed []string
			if err := json.Unmarshal([]byte(patterns), &parsed); err == nil {
				filePatterns = parsed
			}
		}
		includeFIXME := true
		if inc, ok := params["include_fixme"].(bool); ok {
			includeFIXME = inc
		}
		commentTasks := scanComments(projectRoot, filePatterns, includeFIXME, useAppleFM)
		discoveries = append(discoveries, commentTasks...)
	}

	// Scan markdown
	if action == "markdown" || action == "all" {
		docPath := ""
		if path, ok := params["doc_path"].(string); ok {
			docPath = path
		}
		markdownTasks := scanMarkdown(projectRoot, docPath)
		discoveries = append(discoveries, markdownTasks...)
	}

	// Find orphans
	if action == "orphans" || action == "all" {
		orphanTasks := findOrphanTasks(projectRoot)
		for _, orphan := range orphanTasks {
			discoveries = append(discoveries, orphan)
		}
	}

	// Scan git repository for JSON files
	if action == "git_json" || action == "all" {
		jsonPattern := ""
		if pattern, ok := params["json_pattern"].(string); ok && pattern != "" {
			jsonPattern = pattern
		}
		gitJSONTasks := scanGitJSON(projectRoot, jsonPattern)
		discoveries = append(discoveries, gitJSONTasks...)
	}

	// Build summary
	bySource := make(map[string]int)
	byType := make(map[string]int)

	for _, d := range discoveries {
		if src, ok := d["source"].(string); ok {
			bySource[src]++
		}
		if typ, ok := d["type"].(string); ok {
			byType[typ]++
		}
	}

	summary := map[string]interface{}{
		"total":     len(discoveries),
		"by_source": bySource,
		"by_type":   byType,
	}

	result := map[string]interface{}{
		"action":      action,
		"discoveries": discoveries,
		"summary":     summary,
		"method":      "native_go",
	}

	if useAppleFM {
		result["ai_enhanced"] = true
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
}

// scanComments scans code files for TODO/FIXME comments
func scanComments(projectRoot string, patterns []string, includeFIXME bool, useAppleFM bool) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	// Build regex pattern
	var todoPattern *regexp.Regexp
	if includeFIXME {
		todoPattern = regexp.MustCompile(`(?i)(?:#|//)\s*(TODO|FIXME)[\s:]+(.+)`)
	} else {
		todoPattern = regexp.MustCompile(`(?i)(?:#|//)\s*TODO[\s:]+(.+)`)
	}

	// Simple file walking (could be enhanced with glob support)
	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and non-code files
		if info.IsDir() {
			// Skip common ignore directories
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
				strings.Contains(path, "__pycache__") || strings.Contains(path, ".venv") {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches patterns (simplified - full glob support would be better)
		ext := filepath.Ext(path)
		matched := false
		for _, pattern := range patterns {
			if strings.Contains(pattern, ext) || pattern == "**/*" {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Find TODO/FIXME comments
		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			matches := todoPattern.FindStringSubmatch(line)
			if len(matches) > 0 {
				taskType := "TODO"
				taskText := ""
				if includeFIXME && len(matches) > 2 {
					taskType = strings.ToUpper(matches[1])
					taskText = strings.TrimSpace(matches[2])
				} else if len(matches) > 1 {
					taskText = strings.TrimSpace(matches[1])
				}

				// Use Apple FM for semantic extraction if available
				if useAppleFM && taskText != "" {
					enhanced := enhanceTaskWithAppleFM(taskText)
					if enhanced != nil {
						taskText = enhanced["description"].(string)
						if priority, ok := enhanced["priority"].(string); ok {
							discoveries = append(discoveries, map[string]interface{}{
								"type":        taskType,
								"text":        taskText,
								"file":        strings.TrimPrefix(path, projectRoot+"/"),
								"line":        lineNum + 1,
								"source":      "comment",
								"priority":    priority,
								"ai_enhanced": true,
							})
							continue
						}
					}
				}

				discoveries = append(discoveries, map[string]interface{}{
					"type":   taskType,
					"text":   taskText,
					"file":   strings.TrimPrefix(path, projectRoot+"/"),
					"line":   lineNum + 1,
					"source": "comment",
				})
			}
		}

		return nil
	})

	if err != nil {
		// Log error but continue
	}

	return discoveries
}

// scanMarkdown scans markdown files for task lists
func scanMarkdown(projectRoot string, docPath string) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	searchPath := projectRoot
	if docPath != "" {
		searchPath = filepath.Join(projectRoot, docPath)
	}

	taskPattern := regexp.MustCompile(`(?m)^[\s]*[-*]\s*\[([ xX])\]\s*(.+)`)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		matches := taskPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				isDone := strings.ToLower(match[1]) == "x"
				if !isDone {
					discoveries = append(discoveries, map[string]interface{}{
						"type":      "MARKDOWN_TASK",
						"text":      strings.TrimSpace(match[2]),
						"file":      strings.TrimPrefix(path, projectRoot+"/"),
						"completed": isDone,
						"source":    "markdown",
					})
				}
			}
		}

		return nil
	})

	if err != nil {
		// Log error but continue
	}

	return discoveries
}

// findOrphanTasks finds orphaned tasks (tasks with invalid structure)
func findOrphanTasks(projectRoot string) []map[string]interface{} {
	orphans := []map[string]interface{}{}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return orphans
	}

	// Build task map and dependency graph
	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	// Build dependency graph using gonum
	tg, err := BuildTaskGraph(tasks)
	if err != nil {
		// If graph building fails, continue with empty graph
		tg = NewTaskGraph()
	}

	// Find missing dependencies
	missing := findMissingDependencies(tasks, tg)

	// Find circular dependencies using gonum
	cycles := DetectCycles(tg)

	// Identify orphan tasks
	for _, task := range tasks {
		issues := []string{}

		// Check for missing dependencies
		for _, dep := range task.Dependencies {
			if !taskMap[dep] {
				issues = append(issues, fmt.Sprintf("missing_dependency:%s", dep))
			}
		}

		// Check if task is part of a cycle
		for _, cycle := range cycles {
			for _, cycleTaskID := range cycle {
				if cycleTaskID == task.ID {
					issues = append(issues, "circular_dependency")
					break
				}
			}
			if len(issues) > 0 {
				break
			}
		}

		// Check for invalid parent references (if tasks have parent field in metadata)
		if task.Metadata != nil {
			if parentID, ok := task.Metadata["parent_id"].(string); ok && parentID != "" {
				if !taskMap[parentID] {
					issues = append(issues, fmt.Sprintf("missing_parent:%s", parentID))
				}
			}
		}

		// Check for tasks that should have structure but don't
		if len(task.Dependencies) > 0 && len(task.Tags) == 0 && task.Priority == "" {
			issues = append(issues, "incomplete_structure")
		}

		if len(issues) > 0 {
			orphans = append(orphans, map[string]interface{}{
				"type":    "ORPHAN",
				"text":    task.Content,
				"task_id": task.ID,
				"status":  task.Status,
				"issues":  issues,
				"source":  "orphan_detection",
			})
		}
	}

	// Use missing dependencies info (unused variable fix)
	_ = missing

	return orphans
}

// enhanceTaskWithAppleFM uses Apple FM to extract structured task information
func enhanceTaskWithAppleFM(taskText string) map[string]interface{} {
	support := platform.CheckAppleFoundationModelsSupport()
	if !support.Supported {
		return nil
	}

	sess := fm.NewSession()
	defer sess.Release()

	prompt := fmt.Sprintf(`Extract structured information from this task comment:

"%s"

Return JSON with: {"description": "cleaned task description", "priority": "low|medium|high", "category": "bug|feature|refactor|docs"}`,
		taskText)

	result := sess.RespondWithOptions(prompt, 200, 0.2)

	// Try to parse JSON from result
	jsonStart := strings.Index(result, "{")
	jsonEnd := strings.LastIndex(result, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		var enhanced map[string]interface{}
		if err := json.Unmarshal([]byte(result[jsonStart:jsonEnd+1]), &enhanced); err == nil {
			return enhanced
		}
	}

	return nil
}

// scanGitJSON scans git repository for JSON files containing tasks
// Finds JSON files committed in git and extracts tasks from them
func scanGitJSON(projectRoot string, jsonPattern string) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	// Default pattern: look for .todo2/state.todo2.json files
	if jsonPattern == "" {
		jsonPattern = "**/.todo2/state.todo2.json"
	}

	// Use git to find JSON files
	// First, try git ls-files to find tracked JSON files
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "ls-files", "*.json", "**/*.json")
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		// Git not available or not a git repo - return empty
		return discoveries
	}

	// Parse git output to get list of JSON files
	jsonFiles := []string{}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	defaultPattern := "**/.todo2/state.todo2.json"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Use default pattern if not specified
		patternToUse := jsonPattern
		if patternToUse == "" {
			patternToUse = defaultPattern
		}

		// Filter by pattern
		matched := false
		if patternToUse == defaultPattern {
			// Default: only match state.todo2.json files
			matched = strings.Contains(line, "state.todo2.json")
		} else {
			// Custom pattern: try exact match first
			matched, _ = filepath.Match(patternToUse, line)
			if !matched {
				// Try with ** prefix for glob matching
				if strings.HasPrefix(patternToUse, "**/") {
					pattern := strings.TrimPrefix(patternToUse, "**/")
					matched, _ = filepath.Match(pattern, filepath.Base(line))
				}
				// Also try simple contains match for flexibility
				if !matched && strings.Contains(line, strings.TrimPrefix(patternToUse, "**/")) {
					matched = true
				}
			}
		}

		if matched {
			jsonFiles = append(jsonFiles, line)
		}
	}

	// For each JSON file, extract tasks
	for _, jsonFile := range jsonFiles {
		fullPath := filepath.Join(projectRoot, jsonFile)

		// Try to read file from git history (all commits)
		// Use git log to find all versions of this file
		cmd = exec.CommandContext(ctx, "git", "log", "--all", "--pretty=format:%H", "--", jsonFile)
		cmd.Dir = projectRoot
		commitOutput, err := cmd.Output()
		if err != nil {
			// If git log fails, try reading current file
			tasks, _, err := LoadJSONStateFromFile(fullPath)
			if err == nil {
				for _, task := range tasks {
					discoveries = append(discoveries, map[string]interface{}{
						"type":      "JSON_TASK",
						"text":      task.Content,
						"task_id":   task.ID,
						"status":    task.Status,
						"priority":  task.Priority,
						"file":      jsonFile,
						"source":    "git_json",
						"completed": task.Completed,
					})
				}
			}
			continue
		}

		// Process each commit that modified this file
		commits := strings.Split(strings.TrimSpace(string(commitOutput)), "\n")
		processedTasks := make(map[string]bool) // Track unique task IDs to avoid duplicates

		for _, commit := range commits {
			commit = strings.TrimSpace(commit)
			if commit == "" {
				continue
			}

			// Get file content from this commit
			cmd = exec.CommandContext(ctx, "git", "show", commit+":"+jsonFile)
			cmd.Dir = projectRoot
			fileContent, err := cmd.Output()
			if err != nil {
				continue
			}

			// Parse JSON and extract tasks
			tasks, _, err := LoadJSONStateFromContent(fileContent)
			if err != nil {
				continue
			}

			// Add tasks to discoveries (avoid duplicates)
			for _, task := range tasks {
				// Use task ID + commit as unique key to track tasks across commits
				uniqueKey := fmt.Sprintf("%s:%s", task.ID, commit)
				if processedTasks[uniqueKey] {
					continue
				}
				processedTasks[uniqueKey] = true

				discoveries = append(discoveries, map[string]interface{}{
					"type":      "JSON_TASK",
					"text":      task.Content,
					"task_id":   task.ID,
					"status":    task.Status,
					"priority":  task.Priority,
					"file":      jsonFile,
					"commit":    commit[:8], // Short commit hash
					"source":    "git_json",
					"completed": task.Completed,
				})
			}
		}
	}

	return discoveries
}

// LoadJSONStateFromFile and LoadJSONStateFromContent are now in todo2_json.go
