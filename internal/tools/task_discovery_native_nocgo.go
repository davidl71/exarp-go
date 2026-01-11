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
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
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
		orphanTasks := findOrphanTasksBasic(projectRoot)
		discoveries = append(discoveries, orphanTasks...)
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
		"ai_enhanced": false, // No Apple FM on this platform
	}

	// Optionally create tasks if requested
	if createTasks, ok := params["create_tasks"].(bool); ok && createTasks {
		createdTasks := createTasksFromDiscoveries(projectRoot, discoveries)
		result["tasks_created"] = createdTasks
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	return []framework.TextContent{
		{Type: "text", Text: string(output)},
	}, nil
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
			// Skip common ignore directories
			if strings.Contains(path, ".git") || strings.Contains(path, "node_modules") ||
				strings.Contains(path, "__pycache__") || strings.Contains(path, ".venv") ||
				strings.Contains(path, "vendor") || strings.Contains(path, ".idea") ||
				strings.Contains(path, ".vscode") || strings.Contains(path, "dist") ||
				strings.Contains(path, "build") || strings.Contains(path, "target") {
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
					discoveries = append(discoveries, map[string]interface{}{
						"type":   taskType,
						"text":   taskText,
						"file":   strings.TrimPrefix(path, projectRoot+"/"),
						"line":   lineNum + 1,
						"source": "comment",
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
				strings.Contains(path, "build") {
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

// findOrphanTasksBasic finds orphaned tasks (tasks with invalid structure)
func findOrphanTasksBasic(projectRoot string) []map[string]interface{} {
	orphans := []map[string]interface{}{}

	tasks, err := LoadTodo2Tasks(projectRoot)
	if err != nil {
		return orphans
	}

	// Build task map
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

// createTasksFromDiscoveries creates Todo2 tasks from discovered items
func createTasksFromDiscoveries(projectRoot string, discoveries []map[string]interface{}) []map[string]interface{} {
	createdTasks := []map[string]interface{}{}

	// Load existing tasks to avoid duplicates
	existingTasks, err := LoadTodo2Tasks(projectRoot)
	if err == nil {
		existingContent := make(map[string]bool)
		for _, task := range existingTasks {
			existingContent[strings.ToLower(strings.TrimSpace(task.Content))] = true
		}

		// Create tasks from discoveries
		for _, discovery := range discoveries {
			text, ok := discovery["text"].(string)
			if !ok || text == "" {
				continue
			}

			// Check for duplicates
			textLower := strings.ToLower(strings.TrimSpace(text))
			if existingContent[textLower] {
				continue
			}

			// Create new task
			taskID := generateTaskID(existingTasks)
			newTask := Todo2Task{
				ID:       taskID,
				Content:  text,
				Status:   "Todo",
				Priority: "medium",
				Tags:     []string{"discovered", discovery["source"].(string)},
				Metadata: map[string]interface{}{
					"discovered_from": discovery["file"],
					"discovered_line": discovery["line"],
					"discovery_type":  discovery["type"],
				},
			}

			// Add to existing tasks
			existingTasks = append(existingTasks, newTask)
			existingContent[textLower] = true

			createdTasks = append(createdTasks, map[string]interface{}{
				"id":      taskID,
				"content": text,
				"source":  discovery["source"],
			})
		}

		// Save tasks
		if len(createdTasks) > 0 {
			_ = SaveTodo2Tasks(projectRoot, existingTasks)
		}
	}

	return createdTasks
}

// generateTaskID generates a task ID using epoch milliseconds (T-{epoch_milliseconds})
// This is O(1) and doesn't require loading all tasks, matching task_workflow implementation
// Format: T-{epoch_milliseconds} (e.g., T-1768158627000)
// Note: tasks parameter is kept for backward compatibility but is no longer used
func generateTaskID(tasks []Todo2Task) string {
	// Using epoch milliseconds for O(1) ID generation
	epochMillis := time.Now().UnixMilli()
	return fmt.Sprintf("T-%d", epochMillis)
}
