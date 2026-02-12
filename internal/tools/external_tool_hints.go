package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// handleAddExternalToolHintsNative handles the add_external_tool_hints tool with native Go implementation
func handleAddExternalToolHintsNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get project root
	projectRoot, err := FindProjectRoot()
	if err != nil {
		// Fallback to PROJECT_ROOT env var or current directory
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	// Get optional parameters
	dryRun := false
	if dryRunRaw, ok := params["dry_run"].(bool); ok {
		dryRun = dryRunRaw
	}

	minFileSize := 50
	if minFileSizeRaw, ok := params["min_file_size"]; ok {
		switch v := minFileSizeRaw.(type) {
		case int:
			minFileSize = v
		case float64:
			minFileSize = int(v)
		}
	}

	outputPath := ""
	if outputPathRaw, ok := params["output_path"].(string); ok && outputPathRaw != "" {
		outputPath = outputPathRaw
	} else {
		outputPath = filepath.Join(projectRoot, "docs", "EXTERNAL_TOOL_HINTS_REPORT.md")
	}

	// Perform external tool hints analysis
	results := processExternalToolHints(projectRoot, minFileSize, dryRun)

	// Save report if output path specified
	if outputPath != "" {
		report := generateExternalToolHintsReport(results, projectRoot)
		reportPath := outputPath
		if !filepath.IsAbs(reportPath) {
			reportPath = filepath.Join(projectRoot, reportPath)
		}

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err == nil {
			os.WriteFile(reportPath, []byte(report), 0644)
		}
		results.ReportPath = reportPath
	}

	// Build response
	responseData := map[string]interface{}{
		"files_scanned":        results.FilesScanned,
		"files_modified":       results.FilesModified,
		"files_skipped":        results.FilesSkipped,
		"hints_added_count":    len(results.HintsAdded),
		"hints_skipped_count":  len(results.HintsSkipped),
		"report_path":          results.ReportPath,
		"dry_run":              dryRun,
		"status":               "success",
		"hints_added":          results.HintsAdded,
		"hints_skipped_sample": results.HintsSkipped,
	}

	return response.FormatResult(responseData, "")
}

// ExternalToolHintsResults represents the results of external tool hints processing
type ExternalToolHintsResults struct {
	FilesScanned  int
	FilesModified int
	FilesSkipped  int
	HintsAdded    []map[string]interface{}
	HintsSkipped  []map[string]interface{}
	ReportPath    string
}

// processExternalToolHints processes markdown files to add external tool hints
func processExternalToolHints(projectRoot string, minFileSize int, dryRun bool) ExternalToolHintsResults {
	results := ExternalToolHintsResults{
		FilesScanned:  0,
		FilesModified: 0,
		FilesSkipped:  0,
		HintsAdded:    []map[string]interface{}{},
		HintsSkipped:  []map[string]interface{}{},
	}

	// Common library patterns to detect
	libraryPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(react|vue|angular|svelte)`),
		regexp.MustCompile(`(?i)(express|fastapi|flask|django|rails)`),
		regexp.MustCompile(`(?i)(postgresql|mysql|mongodb|redis)`),
		regexp.MustCompile(`(?i)(aws|azure|gcp|cloud)`),
		regexp.MustCompile(`(?i)(docker|kubernetes|terraform)`),
		regexp.MustCompile(`(?i)(pytest|jest|mocha|junit)`),
	}

	// Scan docs directory
	docsPath := filepath.Join(projectRoot, "docs")
	if _, err := os.Stat(docsPath); err != nil {
		// No docs directory, return empty results
		return results
	}

	// Walk docs directory
	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		results.FilesScanned++

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			results.FilesSkipped++
			results.HintsSkipped = append(results.HintsSkipped, map[string]interface{}{
				"file":   path,
				"reason": fmt.Sprintf("Error reading file: %v", err),
			})
			return nil
		}

		// Check file size
		lines := strings.Split(string(content), "\n")
		if len(lines) < minFileSize {
			results.FilesSkipped++
			results.HintsSkipped = append(results.HintsSkipped, map[string]interface{}{
				"file":   path,
				"reason": "File too short",
			})
			return nil
		}

		// Check if hint already exists
		contentStr := string(content)
		if hasExistingHint(contentStr) {
			results.FilesSkipped++
			results.HintsSkipped = append(results.HintsSkipped, map[string]interface{}{
				"file":   path,
				"reason": "Hint already exists",
			})
			return nil
		}

		// Detect libraries
		libraries := detectLibraries(contentStr, libraryPatterns)
		if len(libraries) == 0 {
			results.FilesSkipped++
			results.HintsSkipped = append(results.HintsSkipped, map[string]interface{}{
				"file":   path,
				"reason": "No external libraries detected",
			})
			return nil
		}

		// Generate hint
		hintText := generateHint(libraries)

		// Insert hint if not dry run
		if !dryRun {
			newContent := insertHint(contentStr, hintText)
			if err := os.WriteFile(path, []byte(newContent), 0644); err == nil {
				results.FilesModified++
				results.HintsAdded = append(results.HintsAdded, map[string]interface{}{
					"file":      path,
					"libraries": libraries,
					"hint_type": "section",
				})
			} else {
				results.FilesSkipped++
				results.HintsSkipped = append(results.HintsSkipped, map[string]interface{}{
					"file":   path,
					"reason": fmt.Sprintf("Error writing file: %v", err),
				})
			}
		} else {
			// Dry run - just report
			results.HintsAdded = append(results.HintsAdded, map[string]interface{}{
				"file":      path,
				"libraries": libraries,
				"hint_type": "section",
				"would_add": true,
			})
		}

		return nil
	})

	return results
}

// hasExistingHint checks if file already has a Context7 hint
func hasExistingHint(content string) bool {
	hintPatterns := []string{
		"context7",
		"resolve-library-id",
		"query-docs",
		"external.*tool",
	}

	contentLower := strings.ToLower(content)
	for _, pattern := range hintPatterns {
		matched, _ := regexp.MatchString(pattern, contentLower)
		if matched {
			return true
		}
	}

	return false
}

// detectLibraries detects external libraries in content
func detectLibraries(content string, patterns []*regexp.Regexp) []string {
	libraries := make(map[string]bool)

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 0 {
				lib := strings.ToLower(match[0])
				libraries[lib] = true
			}
		}
	}

	result := make([]string, 0, len(libraries))
	for lib := range libraries {
		result = append(result, lib)
	}

	return result
}

// generateHint generates a Context7 hint for detected libraries
func generateHint(libraries []string) string {
	if len(libraries) == 0 {
		return ""
	}

	// Generate hint section
	hint := "\n## External Tool Hints\n\n"
	hint += "For documentation on external libraries used in this document, use Context7:\n\n"

	for _, lib := range libraries {
		hint += fmt.Sprintf("- **%s**: Use `resolve-library-id` then `query-docs` for %s documentation\n", lib, lib)
	}

	return hint
}

// insertHint inserts a hint into markdown content
func insertHint(content string, hint string) string {
	// Try to insert after first heading or at the end
	lines := strings.Split(content, "\n")

	// Find first heading
	insertIndex := len(lines)
	for i, line := range lines {
		if strings.HasPrefix(line, "#") && i > 0 {
			insertIndex = i + 1
			break
		}
	}

	// Insert hint
	newLines := make([]string, 0, len(lines)+10)
	newLines = append(newLines, lines[:insertIndex]...)
	newLines = append(newLines, "")
	newLines = append(newLines, hint)
	newLines = append(newLines, "")
	newLines = append(newLines, lines[insertIndex:]...)

	return strings.Join(newLines, "\n")
}

// generateExternalToolHintsReport generates a markdown report
func generateExternalToolHintsReport(results ExternalToolHintsResults, projectRoot string) string {
	report := fmt.Sprintf(`# External Tool Hints Report

**Generated:** 2026-01-09

## Summary

- **Files Scanned:** %d
- **Files Modified:** %d
- **Files Skipped:** %d
- **Hints Added:** %d
- **Hints Skipped:** %d

## Hints Added

`,
		results.FilesScanned,
		results.FilesModified,
		results.FilesSkipped,
		len(results.HintsAdded),
		len(results.HintsSkipped),
	)

	for _, hint := range results.HintsAdded {
		if file, ok := hint["file"].(string); ok {
			report += fmt.Sprintf("- ✅ %s\n", file)
		}
	}

	if len(results.HintsSkipped) > 0 {
		report += "\n## Hints Skipped\n\n"
		for _, hint := range results.HintsSkipped {
			if file, ok := hint["file"].(string); ok {
				if reason, ok := hint["reason"].(string); ok {
					report += fmt.Sprintf("- ⏭️ %s: %s\n", file, reason)
				}
			}
		}
	}

	return report
}
