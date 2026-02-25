//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

// task_discovery_native_scanners.go — Task discovery: Apple FM enhancement, planning doc scanner, and git JSON scanner.
// See also: task_discovery_native.go
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
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   enhanceTaskWithAppleFM
//   enhancePlanningDocWithAppleFM — enhancePlanningDocWithAppleFM uses the default FM to extract task/epic references and structure from planning documents.
//   scanPlanningDocs — scanPlanningDocs scans markdown files for planning document structure and task/epic links.
//   scanGitJSON — scanGitJSON scans git repository for JSON files containing tasks
// ────────────────────────────────────────────────────────────────────────────

// ─── enhanceTaskWithAppleFM ─────────────────────────────────────────────────
func enhanceTaskWithAppleFM(ctx context.Context, taskText string) map[string]interface{} {
	if !FMAvailable() {
		return nil
	}

	fmCtx, cancel := context.WithTimeout(ctx, fmDiscoveryTimeout)
	defer cancel()

	prompt := fmt.Sprintf(`Extract structured information from this task comment:

"%s"

Return JSON with: {"description": "cleaned task description", "priority": "low|medium|high", "category": "bug|feature|refactor|docs"}`,
		taskText)

	result, err := DefaultFMProvider().Generate(fmCtx, prompt, 200, 0.2)
	if err != nil {
		return nil
	}

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

// ─── enhancePlanningDocWithAppleFM ──────────────────────────────────────────
// enhancePlanningDocWithAppleFM uses the default FM to extract task/epic references and structure from planning documents.
// A per-call timeout prevents a hung FM from blocking the entire discovery scan.
func enhancePlanningDocWithAppleFM(ctx context.Context, content string, filePath string) map[string]interface{} {
	if !FMAvailable() {
		return nil
	}

	fmCtx, cancel := context.WithTimeout(ctx, fmDiscoveryTimeout)
	defer cancel()

	contentLimit := len(content)
	if contentLimit > 5000 {
		contentLimit = 5000
	}

	contentPreview := content[:contentLimit]

	prompt := fmt.Sprintf(`Analyze this planning document and extract structured information:

File: %s

Content:
%s

Extract:
1. Task IDs referenced (format: T-123 or T-1234567890)
2. Epic IDs referenced (tasks tagged with #epic)
3. Planning document type (epic_planning, feature_planning, migration_planning, architecture_planning, etc.)
4. Related planning documents mentioned (file paths)
5. Task/epic relationships described

Return JSON only (no other text):
{
  "task_refs": ["T-123", "T-456"],
  "epic_refs": ["T-789"],
  "doc_type": "epic_planning",
  "related_docs": ["docs/planning/related-plan.md"],
  "relationships": [{"from": "T-123", "to": "T-456", "type": "depends_on"}]
}`,
		filePath, contentPreview)

	result, err := DefaultFMProvider().Generate(fmCtx, prompt, 1000, 0.2)
	if err != nil {
		return nil
	}

	// Parse JSON from result
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

// ─── scanPlanningDocs ───────────────────────────────────────────────────────
// scanPlanningDocs scans markdown files for planning document structure and task/epic links.
func scanPlanningDocs(ctx context.Context, projectRoot string, docPath string, useAppleFM bool) []map[string]interface{} {
	discoveries := []map[string]interface{}{}

	searchPath := projectRoot
	if docPath != "" {
		searchPath = filepath.Join(projectRoot, docPath)
	}

	// Basic regex patterns (fallback if Apple FM unavailable or for validation)
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

		// Use default FM for semantic extraction if available
		if useAppleFM {
			enhanced := enhancePlanningDocWithAppleFM(ctx, contentStr, relativePath)
			if enhanced != nil {
				discoveries = append(discoveries, map[string]interface{}{
					"type":          "PLANNING_DOC",
					"file":          relativePath,
					"task_refs":     enhanced["task_refs"],
					"epic_refs":     enhanced["epic_refs"],
					"doc_type":      enhanced["doc_type"],
					"related_docs":  enhanced["related_docs"],
					"relationships": enhanced["relationships"],
					"source":        "planning_doc",
					"ai_enhanced":   true,
				})

				return nil
			}
		}

		// Fallback to regex-based extraction
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

// ─── scanGitJSON ────────────────────────────────────────────────────────────
// scanGitJSON scans git repository for JSON files containing tasks
// Finds JSON files committed in git and extracts tasks from them.
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
