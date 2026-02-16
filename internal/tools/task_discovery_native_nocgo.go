//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

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

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// handleTaskDiscoveryNative handles task_discovery with native Go (no Apple FM)
// Basic scanning works on all platforms - Apple FM is only for semantic enhancement
func handleTaskDiscoveryNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action, _ := params["action"].(string)
	if action == "" {
		action = "all"
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	discoveries := []map[string]interface{}{}

	// Scan comments
	if action == "comments" || action == "all" {
		filePatterns := []string{"**/*.go", "**/*.py", "**/*.js", "**/*.ts", "**/*.rs", "**/*.java", "**/*.cpp", "**/*.c"}
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
		commentTasks := scanCommentsBasic(projectRoot, filePatterns, includeFIXME)
		discoveries = append(discoveries, commentTasks...)
	}

	// Scan markdown
	if action == "markdown" || action == "all" {
		docPath := ""
		if path, ok := params["doc_path"].(string); ok {
			docPath = path
		}
		markdownTasks := scanMarkdownBasic(projectRoot, docPath)
		discoveries = append(discoveries, markdownTasks...)
	}

	// Find orphans
	if action == "orphans" || action == "all" {
		orphanTasks := findOrphanTasksBasic(ctx, projectRoot)
		discoveries = append(discoveries, orphanTasks...)
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

	// Scan planning documents for task/epic links (regex-based fallback)
	if action == "planning_links" || action == "all" {
		docPath := ""
		if path, ok := params["doc_path"].(string); ok {
			docPath = path
		}
		planningLinks := scanPlanningDocsBasic(projectRoot, docPath)
		discoveries = append(discoveries, planningLinks...)
	}

	// Build summary
	bySource := make(map[string]int)
	byType := make(map[string]int)
	byTag := make(map[string]int)
	withTags := 0

	for _, d := range discoveries {
		if src, ok := d["source"].(string); ok {
			bySource[src]++
		}
		if typ, ok := d["type"].(string); ok {
			byType[typ]++
		}
		// Count tags
		if tags, ok := d["tags"].([]string); ok && len(tags) > 0 {
			withTags++
			for _, tag := range tags {
				byTag[tag]++
			}
		}
	}

	summary := map[string]interface{}{
		"total":     len(discoveries),
		"by_source": bySource,
		"by_type":   byType,
	}

	// Add tag statistics if any tags were found
	if withTags > 0 {
		summary["with_tags"] = withTags
		summary["by_tag"] = byTag
	}

	result := map[string]interface{}{
		"action":      action,
		"discoveries": discoveries,
		"summary":     summary,
		"method":      "native_go",
		"ai_enhanced": false, // No Apple FM on this platform
	}

	// Optionally create tasks if requested
	if createTasks, ok := params["create_tasks"].(bool); ok && createTasks {
		createdTasks := createTasksFromDiscoveries(ctx, projectRoot, discoveries)
		result["tasks_created"] = createdTasks
	}

	// Optionally write result to output_path (parity with CGO build)
	if outputPath, ok := params["output_path"].(string); ok && outputPath != "" {
		fullPath := outputPath
		if !filepath.IsAbs(fullPath) {
			fullPath = filepath.Join(projectRoot, fullPath)
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err == nil {
			raw, _ := json.MarshalIndent(result, "", "  ")
			if err := os.WriteFile(fullPath, raw, 0644); err == nil {
				result["report_path"] = fullPath
			}
		}
	}

	return response.FormatResult(result, "")
}

// scanCommentsBasic scans code files for TODO/FIXME comments (basic version without AI enhancement)
func scanCommentsBasic(projectRoot string, patterns []string, includeFIXME bool) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	// Build regex pattern
	var todoPattern *regexp.Regexp
	if includeFIXME {
		todoPattern = regexp.MustCompile(`(?i)(?:#|//|/\*)\s*(TODO|FIXME|XXX|HACK|NOTE)[\s:]+(.+)`)
	} else {
		todoPattern = regexp.MustCompile(`(?i)(?:#|//|/\*)\s*TODO[\s:]+(.+)`)
	}

	// File extension mapping for pattern matching
	extMap := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".rs": true, ".java": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true,
	}

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and non-code files
		if info.IsDir() {
			// Skip common ignore directories and archive
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
				strings.Contains(path, "__pycache__") || strings.Contains(path, ".venv") ||
				strings.Contains(path, "vendor") || strings.Contains(path, ".idea") ||
				strings.Contains(path, ".vscode") || strings.Contains(path, "dist") ||
				strings.Contains(path, "build") || strings.Contains(path, "target") ||
				strings.Contains(path, "/archive/") {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file extension matches patterns
		ext := filepath.Ext(path)
		matched := false
		for _, pattern := range patterns {
			if strings.Contains(pattern, ext) || pattern == "**/*" || strings.HasSuffix(pattern, ext) {
				matched = true
				break
			}
		}
		// Also check extension map
		if !matched && extMap[ext] {
			matched = true
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

				// Remove comment markers from task text
				taskText = strings.TrimPrefix(taskText, "//")
				taskText = strings.TrimPrefix(taskText, "#")
				taskText = strings.TrimPrefix(taskText, "/*")
				taskText = strings.TrimSuffix(taskText, "*/")
				taskText = strings.TrimSpace(taskText)

				if taskText != "" {
					// Extract hashtag-style tags from the comment
					tags, cleanText := extractTagsFromText(taskText)

					discovery := map[string]interface{}{
						"type":   taskType,
						"text":   taskText, // Keep original text for reference
						"file":   strings.TrimPrefix(path, projectRoot+"/"),
						"line":   lineNum + 1,
						"source": "comment",
					}

					// Add tags if found
					if len(tags) > 0 {
						discovery["tags"] = tags
						discovery["clean_text"] = cleanText // Text without tags
					}

					discoveries = append(discoveries, discovery)
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

// scanMarkdownBasic scans markdown files for task lists (basic version)
func scanMarkdownBasic(projectRoot string, docPath string) []map[string]interface{} {
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
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
				strings.Contains(path, "vendor") || strings.Contains(path, "dist") ||
				strings.Contains(path, "build") || strings.Contains(path, "/archive/") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".md" && filepath.Ext(path) != ".markdown" {
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
					text := strings.TrimSpace(match[2])
					if IsDeprecatedDiscoveryText(text) {
						continue
					}
					discoveries = append(discoveries, map[string]interface{}{
						"type":      "MARKDOWN_TASK",
						"text":      text,
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

// scanPlanningDocsBasic scans markdown files for planning document structure and task/epic links (regex-based)
func scanPlanningDocsBasic(projectRoot string, docPath string) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	searchPath := projectRoot
	if docPath != "" {
		searchPath = filepath.Join(projectRoot, docPath)
	}

	// Regex pattern for task/epic reference extraction
	taskRefPattern := regexp.MustCompile(`(?:Epic|Task)\s+ID[:\s]+` + "`?T-(\\d+)`?")

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
				strings.Contains(path, "vendor") || strings.Contains(path, "dist") ||
				strings.Contains(path, "build") || strings.Contains(path, "/archive/") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".md" && filepath.Ext(path) != ".markdown" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relativePath := strings.TrimPrefix(path, projectRoot+"/")
		contentStr := string(content)

		// Extract task/epic references using regex
		taskRefs := taskRefPattern.FindAllStringSubmatch(contentStr, -1)
		extractedRefs := []string{}
		for _, match := range taskRefs {
			if len(match) > 1 {
				extractedRefs = append(extractedRefs, "T-"+match[1])
			}
		}

		if len(extractedRefs) > 0 {
			discoveries = append(discoveries, map[string]interface{}{
				"type":      "PLANNING_DOC",
				"file":      relativePath,
				"task_refs": extractedRefs,
				"source":    "planning_doc",
			})
		}

		return nil
	})

	if err != nil {
		// Log error but continue
	}

	return discoveries
}

// findOrphanTasksBasic finds orphaned tasks (tasks with invalid structure).
// Uses GetDependencyAnalysisFromTasks (task_analysis) for cycles and missing deps; preserves
// parent_id and incomplete_structure checks. Mirrors findOrphanTasks in task_discovery_native.go.
func findOrphanTasksBasic(ctx context.Context, projectRoot string) []map[string]interface{} {
	orphans := []map[string]interface{}{}

	store := NewDefaultTaskStore(projectRoot)
	list, err := store.ListTasks(ctx, nil)
	if err != nil {
		return orphans
	}
	tasks := tasksFromPtrs(list)

	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	cycles, missing, err := GetDependencyAnalysisFromTasks(tasks)
	if err != nil {
		return orphans
	}

	missingByTask := make(map[string][]string)
	for _, m := range missing {
		tid, _ := m["task_id"].(string)
		dep, _ := m["missing_dep"].(string)
		if tid != "" && dep != "" {
			missingByTask[tid] = append(missingByTask[tid], dep)
		}
	}

	for _, task := range tasks {
		issues := []string{}

		for _, dep := range missingByTask[task.ID] {
			issues = append(issues, fmt.Sprintf("missing_dependency:%s", dep))
		}

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

		parentID := task.ParentID
		if parentID == "" && task.Metadata != nil {
			if pid, ok := task.Metadata["parent_id"].(string); ok && pid != "" {
				parentID = pid
			}
		}
		if parentID != "" && !taskMap[parentID] {
			issues = append(issues, fmt.Sprintf("missing_parent:%s", parentID))
		}

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

	return orphans
}

// createTasksFromDiscoveries is in task_discovery_common.go (shared with CGO build).

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
