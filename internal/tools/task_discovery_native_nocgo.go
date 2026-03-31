//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/spf13/cast"
)

// handleTaskDiscoveryNative handles task_discovery with native Go (no Apple FM)
// Basic scanning works on all platforms - Apple FM is only for semantic enhancement
func handleTaskDiscoveryNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action := cast.ToString(params["action"])
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
		filePatterns := []string{
			"**/*.go", "**/*.py", "**/*.js", "**/*.ts", "**/*.tsx", "**/*.jsx",
			"**/*.rs", "**/*.java", "**/*.cpp", "**/*.c", "**/*.h", "**/*.hpp",
			"**/*.toml",
		}
		ignorePaths := discoveryIgnorePathsForProject(projectRoot, params)
		if patterns := cast.ToString(params["file_patterns"]); patterns != "" {
			var parsed []string
			if err := json.Unmarshal([]byte(patterns), &parsed); err == nil {
				filePatterns = parsed
			}
		}
		includeFIXME := true
		if _, has := params["include_fixme"]; has {
			includeFIXME = cast.ToBool(params["include_fixme"])
		}
		commentTasks := scanCommentsBasic(projectRoot, filePatterns, ignorePaths, includeFIXME)
		discoveries = append(discoveries, commentTasks...)
	}

	// Scan markdown
	if action == "markdown" || action == "all" {
		docPath := cast.ToString(params["doc_path"])
		ignorePaths := discoveryIgnorePathsForProject(projectRoot, params)
		markdownTasks := scanMarkdownBasic(projectRoot, docPath, ignorePaths)
		discoveries = append(discoveries, markdownTasks...)
	}

	// Find orphans
	if action == "orphans" || action == "all" {
		orphanTasks := findOrphanTasksBasic(ctx, projectRoot)
		discoveries = append(discoveries, orphanTasks...)
	}

	// Scan git repository for JSON files
	if action == "git_json" || action == "all" {
		jsonPattern := cast.ToString(params["json_pattern"])
		gitJSONTasks := scanGitJSON(ctx, projectRoot, jsonPattern)
		discoveries = append(discoveries, gitJSONTasks...)
	}

	// Scan planning documents for task/epic links (regex-based fallback)
	if action == "planning_links" || action == "all" {
		docPath := cast.ToString(params["doc_path"])
		ignorePaths := discoveryIgnorePathsForProject(projectRoot, params)
		planningLinks := scanPlanningDocsBasic(projectRoot, docPath, ignorePaths)
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
	if createTasks := cast.ToBool(params["create_tasks"]); createTasks {
		createdTasks := createTasksFromDiscoveries(ctx, projectRoot, discoveries)
		result["tasks_created"] = createdTasks
	}

	// Optionally write result to output_path (parity with CGO build; default out/task_discovery_report.json when not set)
	outputPath := DefaultReportOutputPath(projectRoot, "task_discovery_report.json", params)
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

	return framework.FormatResult(result, "")
}

// scanCommentsBasic scans code files for TODO/FIXME comments (basic version without AI enhancement)
func scanCommentsBasic(projectRoot string, patterns []string, ignorePaths []string, includeFIXME bool) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	// Build regex pattern
	var todoPattern *regexp.Regexp
	if includeFIXME {
		todoPattern = regexp.MustCompile(`(?i)(?:#|//|/\*)\s*(TODO|FIXME|XXX|HACK|NOTE)(?:\([^)]+\))?[\s:]+(.+)`)
	} else {
		todoPattern = regexp.MustCompile(`(?i)(?:#|//|/\*)\s*TODO(?:\([^)]+\))?[\s:]+(.+)`)
	}

	// File extension mapping for pattern matching
	extMap := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
		".rs": true, ".java": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".toml": true,
	}

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and non-code files
		if info.IsDir() {
			if shouldSkipDiscoveryDir(projectRoot, path, ignorePaths) {
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
func scanMarkdownBasic(projectRoot string, docPath string, ignorePaths []string) []map[string]interface{} {
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
			if shouldSkipDiscoveryDir(projectRoot, path, ignorePaths) {
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
func scanPlanningDocsBasic(projectRoot string, docPath string, ignorePaths []string) []map[string]interface{} {
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
			if shouldSkipDiscoveryDir(projectRoot, path, ignorePaths) {
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

	cycles, missing, err := GetDependencyAnalysisFromTasksWithStore(ctx, store, tasks)
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
			if _, perr := store.GetTask(ctx, parentID); perr != nil {
				issues = append(issues, fmt.Sprintf("missing_parent:%s", parentID))
			}
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
