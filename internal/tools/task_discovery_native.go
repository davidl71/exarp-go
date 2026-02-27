//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

// task_discovery_native.go — Task discovery: main handler, comment scanner, markdown scanner, and orphan finder.
// See also: task_discovery_native_scanners.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/spf13/cast"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   fmDiscoveryTimeout
//   handleTaskDiscoveryNative — handleTaskDiscoveryNative handles task_discovery with native Go and Apple FM.
//   scanComments — scanComments scans code files for TODO/FIXME comments.
//   scanMarkdown — scanMarkdown scans markdown files for task lists.
//   findOrphanTasks — findOrphanTasks finds orphaned tasks (tasks with invalid structure).
// ────────────────────────────────────────────────────────────────────────────

// ─── fmDiscoveryTimeout ─────────────────────────────────────────────────────
const fmDiscoveryTimeout = 10 * time.Second

// ─── handleTaskDiscoveryNative ──────────────────────────────────────────────
// handleTaskDiscoveryNative handles task_discovery with native Go and Apple FM.
func handleTaskDiscoveryNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	action := cast.ToString(params["action"])
	if action == "" {
		action = "all"
	}

	// Use default FM for semantic extraction when available (can be explicitly disabled via params["use_llm"]).
	useAppleFM := FMAvailable()
	if _, has := params["use_llm"]; has {
		useAppleFM = useAppleFM && cast.ToBool(params["use_llm"])
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	discoveries := []map[string]interface{}{}

	// Scan comments
	if action == "comments" || action == "all" {
		filePatterns := []string{"**/*.go", "**/*.py", "**/*.js", "**/*.ts"}

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

		commentTasks := scanComments(ctx, projectRoot, filePatterns, includeFIXME, useAppleFM)
		discoveries = append(discoveries, commentTasks...)
	}

	// Scan markdown
	if action == "markdown" || action == "all" {
		docPath := cast.ToString(params["doc_path"])

		markdownTasks := scanMarkdown(projectRoot, docPath)
		discoveries = append(discoveries, markdownTasks...)
	}

	// Find orphans
	if action == "orphans" || action == "all" {
		orphanTasks := findOrphanTasks(ctx, projectRoot)
		discoveries = append(discoveries, orphanTasks...)
	}

	// Scan git repository for JSON files
	if action == "git_json" || action == "all" {
		jsonPattern := cast.ToString(params["json_pattern"])

		gitJSONTasks := scanGitJSON(projectRoot, jsonPattern)
		discoveries = append(discoveries, gitJSONTasks...)
	}

	// Scan planning documents for task/epic links
	if action == "planning_links" || action == "all" {
		docPath := cast.ToString(params["doc_path"])

		planningLinks := scanPlanningDocs(ctx, projectRoot, docPath, useAppleFM)
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
	}

	if useAppleFM {
		result["ai_enhanced"] = true
	}

	// Optionally create tasks if requested (parity with nocgo build)
	if createTasks := cast.ToBool(params["create_tasks"]); createTasks {
		createdTasks := createTasksFromDiscoveries(ctx, projectRoot, discoveries)
		result["tasks_created"] = createdTasks
	}

	// Optionally write result to output_path (default docs/task_discovery_report.json when not set)
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

// ─── scanComments ───────────────────────────────────────────────────────────
// scanComments scans code files for TODO/FIXME comments.
func scanComments(ctx context.Context, projectRoot string, patterns []string, includeFIXME bool, useAppleFM bool) []map[string]interface{} {
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
			// Skip common ignore directories: vendor, build outputs, archive
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
				strings.Contains(path, "__pycache__") || strings.Contains(path, ".venv") ||
				strings.Contains(path, "vendor") || strings.Contains(path, ".idea") ||
				strings.Contains(path, ".vscode") || strings.Contains(path, "dist") ||
				strings.Contains(path, "build") || strings.Contains(path, "target") ||
				strings.Contains(path, "/archive/") || filepath.Base(path) == "bin" {
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

				// Extract hashtag-style tags from the comment
				tags, cleanText := extractTagsFromText(taskText)

				// Use default FM for semantic extraction if available
				if useAppleFM && taskText != "" {
					enhanced := enhanceTaskWithAppleFM(ctx, taskText)
					if enhanced != nil {
						if desc, ok := enhanced["description"].(string); ok && desc != "" {
							taskText = desc
						}

						if priority, ok := enhanced["priority"].(string); ok {
							discovery := map[string]interface{}{
								"type":        taskType,
								"text":        taskText,
								"file":        strings.TrimPrefix(path, projectRoot+"/"),
								"line":        lineNum + 1,
								"source":      "comment",
								"priority":    priority,
								"ai_enhanced": true,
							}
							// Add tags if found
							if len(tags) > 0 {
								discovery["tags"] = tags
								discovery["clean_text"] = cleanText
							}

							discoveries = append(discoveries, discovery)

							continue
						}
					}
				}

				discovery := map[string]interface{}{
					"type":   taskType,
					"text":   taskText,
					"file":   strings.TrimPrefix(path, projectRoot+"/"),
					"line":   lineNum + 1,
					"source": "comment",
				}
				// Add tags if found
				if len(tags) > 0 {
					discovery["tags"] = tags
					discovery["clean_text"] = cleanText
				}

				discoveries = append(discoveries, discovery)
			}
		}

		return nil
	})
	if err != nil {
		// Log error but continue
	}

	return discoveries
}

// ─── scanMarkdown ───────────────────────────────────────────────────────────
// scanMarkdown scans markdown files for task lists.
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
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
				strings.Contains(path, "/archive/") {
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

// ─── findOrphanTasks ────────────────────────────────────────────────────────
// findOrphanTasks finds orphaned tasks (tasks with invalid structure).
// Uses GetDependencyAnalysisFromTasks (task_analysis) for cycles and missing deps; preserves
// parent_id and incomplete_structure checks.
func findOrphanTasks(ctx context.Context, projectRoot string) []map[string]interface{} {
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

	// Build missing_deps by task for efficient lookup
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

// enhanceTaskWithAppleFM uses the default FM to extract structured task information.
// A per-call timeout prevents a hung FM from blocking the entire discovery scan.
