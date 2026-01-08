//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

	// Find orphans (requires task_analysis, so we'll use Python bridge for now)
	if action == "orphans" || action == "all" {
		// Orphans detection requires hierarchy analysis - fall back to Python for now
		// Could be implemented later
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
		"action":     action,
		"discoveries": discoveries,
		"summary":    summary,
		"method":     "native_go",
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
								"type":     taskType,
								"text":     taskText,
								"file":     strings.TrimPrefix(path, projectRoot+"/"),
								"line":     lineNum + 1,
								"source":   "comment",
								"priority": priority,
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

