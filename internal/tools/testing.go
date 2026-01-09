package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/framework"
)

// handleTestingRun handles the run action for testing tool
func handleTestingRun(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	testPath := "./..."
	if path, ok := params["test_path"].(string); ok && path != "" {
		testPath = path
	}

	verbose := true
	if v, ok := params["verbose"].(bool); ok {
		verbose = v
	}

	coverage := false
	if c, ok := params["coverage"].(bool); ok {
		coverage = c
	}

	// For Go projects, use native Go test execution
	if IsGoProject() {
		result, err := runGoTests(ctx, projectRoot, testPath, verbose, coverage)
		if err == nil {
			return []framework.TextContent{
				{Type: "text", Text: result},
			}, nil
		}
		// If Go test fails, fall through to Python bridge
	}

	// For non-Go projects or if Go test fails, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "testing", params)
	if err != nil {
		return nil, fmt.Errorf("testing run failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTestingCoverage handles the coverage action for testing tool
func handleTestingCoverage(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	coverageFile := ""
	if file, ok := params["coverage_file"].(string); ok && file != "" {
		coverageFile = file
	}

	minCoverage := 80
	if min, ok := params["min_coverage"].(float64); ok {
		minCoverage = int(min)
	}

	format := "html"
	if f, ok := params["format"].(string); ok && f != "" {
		format = f
	}

	// For Go projects, use native Go coverage analysis
	if IsGoProject() {
		result, err := analyzeGoCoverage(ctx, projectRoot, coverageFile, minCoverage, format)
		if err == nil {
			return []framework.TextContent{
				{Type: "text", Text: result},
			}, nil
		}
		// If Go coverage fails, fall through to Python bridge
	}

	// For non-Go projects or if Go coverage fails, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "testing", params)
	if err != nil {
		return nil, fmt.Errorf("testing coverage failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// handleTestingValidate handles the validate action for testing tool
func handleTestingValidate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	testPath := "./..."
	if path, ok := params["test_path"].(string); ok && path != "" {
		testPath = path
	}

	testFramework := "auto"
	if f, ok := params["framework"].(string); ok && f != "" {
		testFramework = f
	}

	// For Go projects, use native Go validation
	if IsGoProject() || testFramework == "go" || testFramework == "auto" {
		result, err := validateGoTests(projectRoot, testPath)
		if err == nil {
			return []framework.TextContent{
				{Type: "text", Text: result},
			}, nil
		}
		// If Go validation fails, fall through to Python bridge
	}

	// For non-Go projects or if Go validation fails, use Python bridge
	result, err := bridge.ExecutePythonTool(ctx, "testing", params)
	if err != nil {
		return nil, fmt.Errorf("testing validate failed: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: result},
	}, nil
}

// runGoTests runs Go tests and returns formatted results
func runGoTests(ctx context.Context, projectRoot, testPath string, verbose, coverage bool) (string, error) {
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if coverage {
		args = append(args, "-cover")
	}
	args = append(args, testPath)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"framework":    "go",
		"test_path":    testPath,
		"output":       string(output),
		"returncode":   0,
		"tests_run":    0,
		"tests_passed": 0,
		"tests_failed": 0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result["returncode"] = exitErr.ExitCode()
		}
		// Parse output for test results
		outputStr := string(output)
		if strings.Contains(outputStr, "PASS") {
			result["tests_passed"] = strings.Count(outputStr, "PASS")
		}
		if strings.Contains(outputStr, "FAIL") {
			result["tests_failed"] = strings.Count(outputStr, "FAIL")
		}
		result["tests_run"] = result["tests_passed"].(int) + result["tests_failed"].(int)
	} else {
		// Parse successful output
		outputStr := string(output)
		if strings.Contains(outputStr, "PASS") {
			result["tests_passed"] = strings.Count(outputStr, "PASS")
		}
		result["tests_run"] = result["tests_passed"].(int)
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// analyzeGoCoverage analyzes Go test coverage
func analyzeGoCoverage(ctx context.Context, projectRoot, coverageFile string, minCoverage int, format string) (string, error) {
	// Generate coverage profile
	coverProfile := "coverage.out"
	if coverageFile != "" {
		coverProfile = coverageFile
	}

	cmd := exec.CommandContext(ctx, "go", "test", "./...", "-coverprofile", coverProfile)
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to generate coverage: %w", err)
	}

	// Get coverage percentage
	cmd = exec.CommandContext(ctx, "go", "tool", "cover", "-func", coverProfile)
	cmd.Dir = projectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to analyze coverage: %w", err)
	}

	// Parse coverage percentage from output
	coveragePercent := 0.0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			// Extract percentage from line like "total:                          (statements)    85.2%"
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasSuffix(part, "%") {
					if percentStr := strings.TrimSuffix(part, "%"); percentStr != "" {
						if percent, err := parseFloat(percentStr); err == nil {
							coveragePercent = percent
							break
						}
					}
				}
			}
		}
	}

	result := map[string]interface{}{
		"coverage_percent": coveragePercent,
		"min_coverage":     minCoverage,
		"meets_threshold":  coveragePercent >= float64(minCoverage),
		"format":           format,
	}

	// Generate HTML report if requested
	if format == "html" {
		htmlFile := "coverage.html"
		cmd = exec.CommandContext(ctx, "go", "tool", "cover", "-html", coverProfile, "-o", htmlFile)
		cmd.Dir = projectRoot
		if err := cmd.Run(); err == nil {
			result["html_file"] = htmlFile
		}
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// validateGoTests validates Go test structure
func validateGoTests(projectRoot, testPath string) (string, error) {
	issues := []string{}

	// Check for test files
	testFiles := []string{}
	err := filepath.Walk(filepath.Join(projectRoot, testPath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to walk test path: %w", err)
	}

	if len(testFiles) == 0 {
		issues = append(issues, "No test files found")
	}

	// Check for test functions in test files
	for _, testFile := range testFiles {
		data, err := os.ReadFile(testFile)
		if err != nil {
			continue
		}
		content := string(data)
		if !strings.Contains(content, "func Test") {
			issues = append(issues, fmt.Sprintf("No test functions in %s", testFile))
		}
	}

	result := map[string]interface{}{
		"valid":        len(issues) == 0,
		"test_files":   len(testFiles),
		"issues":       issues,
		"test_path":    testPath,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// parseFloat is a simple float parser helper
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

