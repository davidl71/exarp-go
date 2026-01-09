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
	"github.com/davidl71/exarp-go/internal/security"
)

// handleCheckAttributionNative handles the check_attribution tool with native Go implementation
func handleCheckAttributionNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
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

	// Get optional parameters
	createTasks := true
	if createTasksRaw, ok := params["create_tasks"].(bool); ok {
		createTasks = createTasksRaw
	}

	outputPath := ""
	if outputPathRaw, ok := params["output_path"].(string); ok && outputPathRaw != "" {
		outputPath = outputPathRaw
	} else {
		outputPath = filepath.Join(projectRoot, "docs", "ATTRIBUTION_COMPLIANCE_REPORT.md")
	}

	// Perform attribution check
	results := performAttributionCheck(projectRoot)

	// Create tasks if requested (for now, skip - complex logic)
	tasksCreated := 0
	if createTasks {
		// TODO: Implement task creation logic
		tasksCreated = 0
	}

	// Save report if output path specified
	if outputPath != "" {
		report := generateAttributionReport(results, projectRoot)
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
		"attribution_score":   results.AttributionScore,
		"compliant_files":     results.CompliantFiles,
		"missing_attribution": results.MissingAttribution,
		"warnings":            results.Warnings,
		"issues":              results.Issues,
		"report_path":         results.ReportPath,
		"tasks_created":       tasksCreated,
		"status":              "success",
	}

	resultJSON, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// AttributionResults represents the results of attribution checking
type AttributionResults struct {
	AttributionScore   float64
	CompliantFiles     []string
	MissingAttribution []string
	Warnings           []map[string]interface{}
	Issues             []map[string]interface{}
	ReportPath         string
}

// performAttributionCheck performs attribution compliance checking
func performAttributionCheck(projectRoot string) AttributionResults {
	results := AttributionResults{
		AttributionScore:   100.0,
		CompliantFiles:     []string{},
		MissingAttribution: []string{},
		Warnings:           []map[string]interface{}{},
		Issues:             []map[string]interface{}{},
	}

	// Check for ATTRIBUTIONS.md
	attributionsPath := filepath.Join(projectRoot, "ATTRIBUTIONS.md")
	if _, err := os.Stat(attributionsPath); err == nil {
		results.CompliantFiles = append(results.CompliantFiles, "ATTRIBUTIONS.md exists")
	} else {
		results.MissingAttribution = append(results.MissingAttribution, "ATTRIBUTIONS.md")
		results.AttributionScore -= 20.0
		results.Issues = append(results.Issues, map[string]interface{}{
			"type":    "missing_file",
			"file":    "ATTRIBUTIONS.md",
			"message": "ATTRIBUTIONS.md file not found",
		})
	}

	// Check dependency files for license info
	checkDependencyFiles(projectRoot, &results)

	// Check for common attribution patterns in code files
	checkCodeAttribution(projectRoot, &results)

	// Ensure score doesn't go below 0
	if results.AttributionScore < 0 {
		results.AttributionScore = 0
	}

	return results
}

// checkDependencyFiles checks dependency files for license information
func checkDependencyFiles(projectRoot string, results *AttributionResults) {
	// Check go.mod
	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		content, err := os.ReadFile(goModPath)
		if err == nil {
			// Check for license comment or note
			contentStr := string(content)
			if strings.Contains(strings.ToLower(contentStr), "license") {
				results.CompliantFiles = append(results.CompliantFiles, "go.mod: License reference found")
			} else {
				results.Warnings = append(results.Warnings, map[string]interface{}{
					"type":    "license_field",
					"file":    "go.mod",
					"message": "go.mod may be missing license comment",
				})
			}
		}
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(projectRoot, "pyproject.toml")
	if _, err := os.Stat(pyprojectPath); err == nil {
		content, err := os.ReadFile(pyprojectPath)
		if err == nil {
			contentStr := strings.ToLower(string(content))
			if strings.Contains(contentStr, "license") {
				results.CompliantFiles = append(results.CompliantFiles, "pyproject.toml: License field present")
			} else {
				results.Warnings = append(results.Warnings, map[string]interface{}{
					"type":    "license_field",
					"file":    "pyproject.toml",
					"message": "pyproject.toml may be missing license field",
				})
			}
		}
	}
}

// checkCodeAttribution checks code files for attribution patterns
func checkCodeAttribution(projectRoot string, results *AttributionResults) {
	// Common attribution patterns
	attributionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(inspired by|based on|from|attribution|credit|license)`),
		regexp.MustCompile(`(?i)(github\.com|gitlab\.com|sourceforge\.net)`),
		regexp.MustCompile(`(?i)(mit|apache|gpl|bsd|license)`),
	}

	// Check specific files that might need attribution
	filesToCheck := []string{
		"project_management_automation/tools/wisdom",
		"bridge",
		"internal/tools",
	}

	for _, dir := range filesToCheck {
		dirPath := filepath.Join(projectRoot, dir)
		if _, err := os.Stat(dirPath); err == nil {
			checkDirectoryForAttribution(dirPath, attributionPatterns, results)
		}
	}
}

// checkDirectoryForAttribution checks a directory for attribution patterns
func checkDirectoryForAttribution(dirPath string, patterns []*regexp.Regexp, results *AttributionResults) {
	// Limit depth and file types
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check source files
		ext := filepath.Ext(path)
		if ext != ".go" && ext != ".py" && ext != ".js" && ext != ".ts" {
			return nil
		}

		// Limit depth
		relPath, _ := filepath.Rel(dirPath, path)
		if strings.Count(relPath, string(filepath.Separator)) > 3 {
			return nil
		}

		// Read file and check for attribution patterns
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		hasAttribution := false

		// Check first 50 lines for attribution
		lines := strings.Split(contentStr, "\n")
		headerLines := lines
		if len(headerLines) > 50 {
			headerLines = headerLines[:50]
		}
		header := strings.Join(headerLines, "\n")

		for _, pattern := range patterns {
			if pattern.MatchString(header) {
				hasAttribution = true
				break
			}
		}

		if hasAttribution {
			results.CompliantFiles = append(results.CompliantFiles, fmt.Sprintf("%s: Attribution pattern found", path))
		} else {
			// Don't penalize too much for individual files
			results.Warnings = append(results.Warnings, map[string]interface{}{
				"type":    "attribution_check",
				"file":    path,
				"message": "File may need attribution verification",
			})
		}

		return nil
	})
}

// generateAttributionReport generates a markdown report
func generateAttributionReport(results AttributionResults, projectRoot string) string {
	report := fmt.Sprintf(`# Attribution Compliance Report

**Generated:** 2026-01-09

## Summary

- **Attribution Score:** %.1f%%
- **Compliant Files:** %d
- **Missing Attribution:** %d
- **Warnings:** %d
- **Issues:** %d

## Compliant Files

`,
		results.AttributionScore,
		len(results.CompliantFiles),
		len(results.MissingAttribution),
		len(results.Warnings),
		len(results.Issues),
	)

	for _, file := range results.CompliantFiles {
		report += fmt.Sprintf("- ✅ %s\n", file)
	}

	if len(results.MissingAttribution) > 0 {
		report += "\n## Missing Attribution\n\n"
		for _, file := range results.MissingAttribution {
			report += fmt.Sprintf("- ❌ %s\n", file)
		}
	}

	if len(results.Warnings) > 0 {
		report += "\n## Warnings\n\n"
		for _, warning := range results.Warnings {
			report += fmt.Sprintf("- ⚠️ %s: %s\n", warning["file"], warning["message"])
		}
	}

	return report
}
