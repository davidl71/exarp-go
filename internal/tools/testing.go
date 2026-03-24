// testing.go — MCP "testing" tool: validate, run, and coverage for project tests.
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/proto"
	"github.com/spf13/cast"
)

// handleTestingRun handles the run action for testing tool.
func handleTestingRun(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	testPath := "./..."
	if path := cast.ToString(params["test_path"]); path != "" {
		testPath = path

		// Validate path against MCP Roots if available
		if _, err := ValidatePathAgainstMCPRoots(ctx, testPath); err != nil {
			return nil, fmt.Errorf("path validation failed: %w", err)
		}
	}

	verbose := true
	if _, has := params["verbose"]; has {
		verbose = cast.ToBool(params["verbose"])
	}

	coverage := false
	if _, has := params["coverage"]; has {
		coverage = cast.ToBool(params["coverage"])
	}

	// Determine framework: use param if specified, otherwise auto-detect
	testFramework := cast.ToString(params["framework"])
	if testFramework == "" || testFramework == "auto" {
		testFramework = detectTestFramework(projectRoot)
	}

	var result string
	switch testFramework {
	case "go":
		result, err = runGoTests(ctx, projectRoot, testPath, verbose, coverage)
	case "python":
		result, err = runPyTests(ctx, projectRoot, testPath, verbose)
	case "rust":
		result, err = runCargoTests(ctx, projectRoot, testPath, verbose)
	case "node":
		result, err = runNpmTests(ctx, projectRoot, verbose)
	case "":
		return nil, fmt.Errorf("testing run: cannot detect project framework (no go.mod, pyproject.toml, Cargo.toml, or package.json found)")
	default:
		return nil, fmt.Errorf("testing run: unsupported framework %q (supported: go, python, rust, node)", testFramework)
	}

	if err != nil {
		return nil, fmt.Errorf("testing run: %w", err)
	}

	resp := &proto.TestingResponse{Success: true, Action: "run", ResultJson: result}

	return framework.FormatResult(TestingResponseToMap(resp), "")
}

// handleTestingCoverage handles the coverage action for testing tool.
func handleTestingCoverage(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	coverageFile := cast.ToString(params["coverage_file"])

	// Use config default, allow override from params
	minCoverage := config.MinCoverage()
	if _, has := params["min_coverage"]; has {
		minCoverage = int(cast.ToFloat64(params["min_coverage"]))
	}

	format := "html"
	if f := cast.ToString(params["format"]); f != "" {
		format = f
	}

	// Determine framework: use param if specified, otherwise auto-detect
	testFramework := cast.ToString(params["framework"])
	if testFramework == "" || testFramework == "auto" {
		testFramework = detectTestFramework(projectRoot)
	}

	switch testFramework {
	case "go":
		result, err := analyzeGoCoverage(ctx, projectRoot, coverageFile, minCoverage, format)
		if err != nil {
			return nil, fmt.Errorf("testing coverage: %w", err)
		}
		resp := &proto.TestingResponse{Success: true, Action: "coverage", ResultJson: result}
		return framework.FormatResult(TestingResponseToMap(resp), "")
	case "python":
		result, err := analyzePyCoverage(ctx, projectRoot, minCoverage, format)
		if err != nil {
			return nil, fmt.Errorf("testing coverage: %w", err)
		}
		resp := &proto.TestingResponse{Success: true, Action: "coverage", ResultJson: result}
		return framework.FormatResult(TestingResponseToMap(resp), "")
	case "rust":
		result, err := analyzeRustCoverage(ctx, projectRoot, minCoverage, format)
		if err != nil {
			return nil, fmt.Errorf("testing coverage: %w", err)
		}
		resp := &proto.TestingResponse{Success: true, Action: "coverage", ResultJson: result}
		return framework.FormatResult(TestingResponseToMap(resp), "")
	case "node":
		result, err := analyzeNodeCoverage(ctx, projectRoot, minCoverage, format)
		if err != nil {
			return nil, fmt.Errorf("testing coverage: %w", err)
		}
		resp := &proto.TestingResponse{Success: true, Action: "coverage", ResultJson: result}
		return framework.FormatResult(TestingResponseToMap(resp), "")
	case "":
		return nil, fmt.Errorf("testing coverage: cannot detect project framework")
	default:
		return nil, fmt.Errorf("testing coverage: unsupported framework %q (supported: go, python, rust, node)", testFramework)
	}
}

// handleTestingValidate handles the validate action for testing tool.
func handleTestingValidate(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find project root: %w", err)
	}

	testPath := "./..."
	if path := cast.ToString(params["test_path"]); path != "" {
		testPath = path
	}

	// Determine framework: use param if specified, otherwise auto-detect
	testFramework := cast.ToString(params["framework"])
	if testFramework == "" || testFramework == "auto" {
		testFramework = detectTestFramework(projectRoot)
	}

	var result string
	switch testFramework {
	case "go":
		result, err = validateGoTests(projectRoot, testPath)
	case "python":
		result, err = validatePyTests(projectRoot, testPath)
	case "rust":
		result, err = validateCargoTests(projectRoot, testPath)
	case "":
		return nil, fmt.Errorf("testing validate: cannot detect project framework")
	default:
		return nil, fmt.Errorf("testing validate: unsupported framework %q", testFramework)
	}

	if err != nil {
		return nil, fmt.Errorf("testing validate: %w", err)
	}

	resp := &proto.TestingResponse{Success: true, Action: "validate", ResultJson: result}

	return framework.FormatResult(TestingResponseToMap(resp), "")
}

// runGoTests runs Go tests and returns formatted results.
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
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
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

// analyzeGoCoverage analyzes Go test coverage.
func analyzeGoCoverage(ctx context.Context, projectRoot, coverageFile string, minCoverage int, format string) (string, error) {
	// Generate coverage profile
	coverProfile := "coverage.out"
	if coverageFile != "" {
		coverProfile = coverageFile
	}

	cmd := exec.CommandContext(ctx, "go", "test", "./...", "-coverprofile", coverProfile, "-covermode=atomic", "-coverpkg=./internal/...")
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

// walkTestFiles walks searchPath and returns all files where match(path, base) is true.
func walkTestFiles(searchPath string, match func(path, base string) bool) ([]string, error) {
	var files []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && match(path, filepath.Base(path)) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// validateGoTests validates Go test structure.
func validateGoTests(projectRoot, testPath string) (string, error) {
	testFiles, err := walkTestFiles(
		filepath.Join(projectRoot, testPath),
		func(path, _ string) bool { return strings.HasSuffix(path, "_test.go") },
	)
	if err != nil {
		return "", fmt.Errorf("failed to walk test path: %w", err)
	}

	var issues []string
	if len(testFiles) == 0 {
		issues = append(issues, "No test files found")
	}

	// Check for test functions in test files
	for _, testFile := range testFiles {
		data, err := os.ReadFile(testFile)
		if err != nil {
			continue
		}
		if !strings.Contains(string(data), "func Test") {
			issues = append(issues, fmt.Sprintf("No test functions in %s", testFile))
		}
	}

	result := map[string]interface{}{
		"valid":      len(issues) == 0,
		"test_files": len(testFiles),
		"issues":     issues,
		"test_path":  testPath,
	}
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// validatePyTests validates Python test structure (pytest).
func validatePyTests(projectRoot, testPath string) (string, error) {
	searchPath := projectRoot
	if testPath != "" && testPath != "./..." {
		searchPath = filepath.Join(projectRoot, testPath)
	}

	testFiles, err := walkTestFiles(searchPath, func(_, base string) bool {
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk test path: %w", err)
	}

	var issues []string
	if len(testFiles) == 0 {
		issues = append(issues, "No pytest test files found (looked for test_*.py and *_test.py)")
	}

	result := map[string]interface{}{
		"valid":      len(issues) == 0,
		"test_files": len(testFiles),
		"issues":     issues,
		"test_path":  testPath,
		"framework":  "python",
	}
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// validateCargoTests validates Rust test structure (cargo test).
func validateCargoTests(projectRoot, testPath string) (string, error) {
	searchPath := projectRoot
	if testPath != "" && testPath != "./..." {
		searchPath = filepath.Join(projectRoot, testPath)
	}

	testFiles, err := walkTestFiles(searchPath, func(path, _ string) bool {
		return strings.HasSuffix(path, ".rs")
	})
	if err != nil {
		return "", fmt.Errorf("failed to walk test path: %w", err)
	}

	var issues []string
	if len(testFiles) == 0 {
		issues = append(issues, "No Rust source files found")
	}

	result := map[string]interface{}{
		"valid":      len(issues) == 0,
		"test_files": len(testFiles),
		"issues":     issues,
		"test_path":  testPath,
		"framework":  "rust",
	}
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// analyzePyCoverage runs pytest with coverage.py and returns a coverage result JSON.
// Requires pytest-cov: pip install pytest-cov. Falls back to plain coverage report if available.
func analyzePyCoverage(ctx context.Context, projectRoot string, minCoverage int, format string) (string, error) {
	args := []string{"-m", "pytest", "--cov=.", "--cov-report=term-missing"}
	if format == "html" {
		args = append(args, "--cov-report=html:coverage_html")
	}

	cmd := exec.CommandContext(ctx, "python", args...)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	coveragePercent := parsePyCoveragePercent(outputStr)

	result := map[string]interface{}{
		"framework":       "python",
		"coverage_percent": coveragePercent,
		"min_coverage":    minCoverage,
		"meets_threshold": coveragePercent >= float64(minCoverage),
		"format":          format,
		"output":          outputStr,
	}
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			result["returncode"] = exitErr.ExitCode()
		}
	}
	if format == "html" {
		result["html_dir"] = filepath.Join(projectRoot, "coverage_html")
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// parsePyCoveragePercent extracts the total coverage percentage from pytest-cov / coverage.py output.
// Looks for "TOTAL ... X%" at the end of the summary table.
func parsePyCoveragePercent(output string) float64 {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "TOTAL" {
			// Last field is "XX%" when pytest-cov summary is present
			last := fields[len(fields)-1]
			if strings.HasSuffix(last, "%") {
				if pct, err := parseFloat(strings.TrimSuffix(last, "%")); err == nil {
					return pct
				}
			}
		}
	}
	return 0
}

// analyzeRustCoverage runs cargo tarpaulin and returns a coverage result JSON.
// Requires: cargo install cargo-tarpaulin
func analyzeRustCoverage(ctx context.Context, projectRoot string, minCoverage int, format string) (string, error) {
	args := []string{"tarpaulin", "--out", "Stdout"}
	if format == "html" {
		args = append(args, "--out", "Html")
	}

	cmd := exec.CommandContext(ctx, "cargo", args...)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	coveragePercent := parseRustCoveragePercent(outputStr)

	result := map[string]interface{}{
		"framework":        "rust",
		"coverage_percent": coveragePercent,
		"min_coverage":     minCoverage,
		"meets_threshold":  coveragePercent >= float64(minCoverage),
		"format":           format,
		"output":           outputStr,
	}
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			result["returncode"] = exitErr.ExitCode()
		}
	}
	if format == "html" {
		result["html_file"] = filepath.Join(projectRoot, "tarpaulin-report.html")
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// parseRustCoveragePercent extracts the total coverage from cargo tarpaulin output.
// Tarpaulin prints a line like: "XX.XX% coverage, ..."
func parseRustCoveragePercent(output string) float64 {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "% coverage"); idx > 0 {
			start := strings.LastIndex(line[:idx], " ")
			if start < 0 {
				start = 0
			} else {
				start++
			}
			if pct, err := parseFloat(line[start:idx]); err == nil {
				return pct
			}
		}
	}
	return 0
}

// analyzeNodeCoverage runs jest with --coverage and returns a coverage result JSON.
// Uses npx jest so it works whether jest is installed locally or globally.
func analyzeNodeCoverage(ctx context.Context, projectRoot string, minCoverage int, format string) (string, error) {
	args := []string{"jest", "--coverage", "--coverageReporters=text"}
	if format == "html" {
		args = append(args, "--coverageReporters=html")
	}

	cmd := exec.CommandContext(ctx, "npx", args...)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	coveragePercent := parseNodeCoveragePercent(outputStr)

	result := map[string]interface{}{
		"framework":        "node",
		"coverage_percent": coveragePercent,
		"min_coverage":     minCoverage,
		"meets_threshold":  coveragePercent >= float64(minCoverage),
		"format":           format,
		"output":           outputStr,
	}
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			result["returncode"] = exitErr.ExitCode()
		}
	}
	if format == "html" {
		result["html_dir"] = filepath.Join(projectRoot, "coverage")
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// parseNodeCoveragePercent extracts the "All files" statements coverage from jest text output.
// Jest prints a table: "All files | XX | ..." — we take the Statements column (index 1).
func parseNodeCoveragePercent(output string) float64 {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "All files") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				if pct, err := parseFloat(strings.TrimSpace(parts[1])); err == nil {
					return pct
				}
			}
		}
	}
	return 0
}

// parseFloat is a simple float parser helper.
func parseFloat(s string) (float64, error) {
	var f float64

	_, err := fmt.Sscanf(s, "%f", &f)

	return f, err
}

// detectTestFramework returns the detected test framework based on project markers.
// Returns "go", "python", "rust", "node", or "" if none detected.
func detectTestFramework(projectRoot string) string {
	if DetectGoProject(projectRoot) {
		return "go"
	}
	if DetectPythonProject(projectRoot) {
		return "python"
	}
	if DetectRustProject(projectRoot) {
		return "rust"
	}
	if DetectTypeScriptProject(projectRoot) {
		return "node"
	}
	return ""
}

// runPyTests runs Python tests using pytest.
func runPyTests(ctx context.Context, projectRoot, testPath string, verbose bool) (string, error) {
	args := []string{"-m", "pytest"}
	if verbose {
		args = append(args, "-v")
	}
	if testPath != "" && testPath != "./..." {
		args = append(args, testPath)
	}

	cmd := exec.CommandContext(ctx, "python", args...)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"framework":  "python",
		"test_path":  testPath,
		"output":     string(output),
		"returncode": 0,
	}

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			result["returncode"] = exitErr.ExitCode()
		}
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// runCargoTests runs Rust tests using cargo test.
func runCargoTests(ctx context.Context, projectRoot, testPath string, verbose bool) (string, error) {
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if testPath != "" && testPath != "./..." {
		args = append(args, "--", testPath)
	}

	cmd := exec.CommandContext(ctx, "cargo", args...)
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"framework":  "rust",
		"test_path":  testPath,
		"output":     string(output),
		"returncode": 0,
	}

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			result["returncode"] = exitErr.ExitCode()
		}
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// runNpmTests runs Node.js tests using npm test.
func runNpmTests(ctx context.Context, projectRoot string, verbose bool) (string, error) {
	cmd := exec.CommandContext(ctx, "npm", "test", "--if-present")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()

	result := map[string]interface{}{
		"framework":  "node",
		"output":     string(output),
		"returncode": 0,
	}

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			result["returncode"] = exitErr.ExitCode()
		}
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
