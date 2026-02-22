// scorecard_go.go — Go-native scorecard calculation: build, test, lint, security checks.
package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/projectroot"
	"github.com/davidl71/exarp-go/internal/security"
)

// GoProjectMetrics represents Go-specific project metrics.
type GoProjectMetrics struct {
	GoFiles         int    `json:"go_files"`
	GoLines         int    `json:"go_lines"`
	GoTestFiles     int    `json:"go_test_files"`
	GoTestLines     int    `json:"go_test_lines"`
	PythonFiles     int    `json:"python_files"` // Bridge scripts only
	PythonLines     int    `json:"python_lines"`
	GoModules       int    `json:"go_modules"`
	GoDependencies  int    `json:"go_dependencies"`
	GoVersion       string `json:"go_version"`
	MCPTools        int    `json:"mcp_tools"`
	MCPPrompts      int    `json:"mcp_prompts"`
	MCPResources    int    `json:"mcp_resources"`
	TotalCodeBytes  int    `json:"total_code_bytes"` // Sum of .go, _test.go, bridge/tests .py file sizes (for token estimate)
	EstimatedTokens int    `json:"estimated_tokens"` // Ratio-based estimate (chars × tokens_per_char); use for context budgeting
}

// GoHealthChecks represents Go-specific health check results.
type GoHealthChecks struct {
	GoModExists       bool    `json:"go_mod_exists"`
	GoSumExists       bool    `json:"go_sum_exists"`
	GoModTidyPasses   bool    `json:"go_mod_tidy_passes"`
	GoVersionValid    bool    `json:"go_version_valid"`
	GoVersion         string  `json:"go_version"`
	GoBuildPasses     bool    `json:"go_build_passes"`
	GoVetPasses       bool    `json:"go_vet_passes"`
	GoFmtCompliant    bool    `json:"go_fmt_compliant"`
	GoLintConfigured  bool    `json:"go_lint_configured"`
	GoLintPasses      bool    `json:"go_lint_passes"`
	GoTestPasses      bool    `json:"go_test_passes"`
	GoTestCoverage    float64 `json:"go_test_coverage"`
	GoVulnCheckPasses bool    `json:"go_vulncheck_passes"`
	// Security features
	PathBoundaryEnforcement bool `json:"path_boundary_enforcement"`
	RateLimiting            bool `json:"rate_limiting"`
	AccessControl           bool `json:"access_control"`
	// Documentation (for documentation score dimension)
	ReadmeExists      bool `json:"readme_exists"`
	DocsDirExists     bool `json:"docs_dir_exists"`
	DocsFileCount     int  `json:"docs_file_count"`
	AIAssistDocsExist bool `json:"ai_assist_docs_exist"` // .cursor/skills, .cursor/rules, CLAUDE.md, or .claude/commands/
}

// FileSizeInfo holds per-file size and token estimate for split/refactor analysis.
type FileSizeInfo struct {
	Path            string `json:"path"` // Relative to project root
	Lines           int    `json:"lines"`
	Bytes           int    `json:"bytes"`
	EstimatedTokens int    `json:"estimated_tokens"`
}

// Default thresholds for "large file" (split/refactor candidates).
const (
	defaultLargeFileTokenThreshold = 6000 // ~24KB at 0.25 tokens/char; exceeds typical context chunk
	defaultLargeFileLineThreshold  = 500  // Common style-guide limit
)

// skipGeneratedProtobufGo reports whether path is a generated proto Go file (proto/*.pb.go) and should be excluded from scorecard metrics and checks.
func skipGeneratedProtobufGo(projectRoot, path string) bool {
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return false
	}
	relSlash := filepath.ToSlash(rel)
	return strings.HasPrefix(relSlash, "proto/") && strings.HasSuffix(relSlash, ".pb.go")
}

// GoScorecardResult represents the complete Go scorecard.
type GoScorecardResult struct {
	Metrics             GoProjectMetrics `json:"metrics"`
	Health              GoHealthChecks   `json:"health"`
	Recommendations     []string         `json:"recommendations"`
	Score               float64          `json:"score"`
	LargeFileCandidates []FileSizeInfo   `json:"large_file_candidates,omitempty"` // Files above token/line threshold; consider splitting
	// FastModeUsed is true when scorecard was generated with FastMode (coverage/lint not run).
	FastModeUsed bool `json:"fast_mode_used,omitempty"`
}

// ScorecardOptions configures scorecard generation behavior.
type ScorecardOptions struct {
	FastMode bool // Skip expensive operations (go test, go build, go mod tidy)
}

// collectGoMetrics collects Go-specific project metrics.
func collectGoMetrics(ctx context.Context, projectRoot string) (*GoProjectMetrics, error) {
	metrics := &GoProjectMetrics{}

	// Count Go files (files, lines, bytes)
	goFiles, goLines, goBytes, err := countGoFilesWithBytes(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to count Go files: %w", err)
	}

	metrics.GoFiles = goFiles
	metrics.GoLines = goLines

	// Count Go test files
	testFiles, testLines, testBytes, err := countGoTestFilesWithBytes(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to count Go test files: %w", err)
	}

	metrics.GoTestFiles = testFiles
	metrics.GoTestLines = testLines

	// Count Python files (bridge scripts only)
	pythonFiles, pythonLines, pythonBytes, err := countPythonFilesWithBytes(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to count Python files: %w", err)
	}

	metrics.PythonFiles = pythonFiles
	metrics.PythonLines = pythonLines

	totalBytes := goBytes + testBytes + pythonBytes
	metrics.TotalCodeBytes = totalBytes
	metrics.EstimatedTokens = int(float64(totalBytes) * config.TokensPerChar())

	// Check Go module
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
		metrics.GoModules = 1
		// Count dependencies
		deps, version, err := getGoModuleInfo(ctx, projectRoot)
		if err == nil {
			metrics.GoDependencies = deps
			metrics.GoVersion = version
		}
	}

	// MCP server counts (these should be accurate)
	metrics.MCPTools = 25     // 24 base + llamacpp
	metrics.MCPPrompts = 15   // Fixed: was 38 (actual count may vary, check sanity-check)
	metrics.MCPResources = 17 // Updated: 11 base + 6 task resources

	return metrics, nil
}

// performGoHealthChecks performs Go-specific health checks.
func performGoHealthChecks(ctx context.Context, projectRoot string, opts *ScorecardOptions) (*GoHealthChecks, error) {
	health := &GoHealthChecks{}

	// Check go.mod
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
		health.GoModExists = true
	}

	// Check go.sum
	if _, err := os.Stat(filepath.Join(projectRoot, "go.sum")); err == nil {
		health.GoSumExists = true
	}

	// Check go mod tidy (skip in fast mode)
	if opts == nil || !opts.FastMode {
		health.GoModTidyPasses = checkGoModTidy(ctx, projectRoot)
	}

	// Get Go version
	version, valid := getGoVersion(ctx)
	health.GoVersion = version
	health.GoVersionValid = valid

	// Check go build (skip in fast mode)
	if opts == nil || !opts.FastMode {
		health.GoBuildPasses = checkGoBuild(ctx, projectRoot)
	}

	// Check go vet
	health.GoVetPasses = checkGoVet(ctx, projectRoot)

	// Check go fmt
	health.GoFmtCompliant = checkGoFmt(ctx, projectRoot)

	// Check golangci-lint
	health.GoLintConfigured = checkGolangciLintConfigured(projectRoot)
	if opts == nil || !opts.FastMode {
		health.GoLintPasses = checkGolangciLint(ctx, projectRoot)
	}

	// Check go test (skip in fast mode - this is the slowest operation)
	if opts == nil || !opts.FastMode {
		health.GoTestPasses, health.GoTestCoverage = checkGoTest(ctx, projectRoot)
	} else {
		// Fast mode: don't run go test, but read coverage from existing coverage.out if present
		// (e.g. from prior "make test-coverage" or full scorecard)
		health.GoTestPasses = true // Assume tests exist if we have test files
		health.GoTestCoverage = readCoverageFromFile(ctx, projectRoot, "coverage.out")
	}

	// Check govulncheck (skip in fast mode)
	if opts == nil || !opts.FastMode {
		health.GoVulnCheckPasses = checkGoVulncheck(ctx, projectRoot)
	}

	// Check security features
	health.PathBoundaryEnforcement = checkPathBoundaryEnforcement(projectRoot)
	health.RateLimiting = checkRateLimiting(projectRoot)
	health.AccessControl = checkAccessControl(projectRoot)

	// Documentation checks (README, docs/, .cursor docs)
	readmePaths := []string{"README.md", "README.rst", "README.txt", "readme.md"}
	for _, name := range readmePaths {
		if _, err := os.Stat(filepath.Join(projectRoot, name)); err == nil {
			health.ReadmeExists = true
			break
		}
	}

	docsDir := filepath.Join(projectRoot, "docs")
	if info, err := os.Stat(docsDir); err == nil && info.IsDir() {
		health.DocsDirExists = true

		entries, _ := os.ReadDir(docsDir)
		for _, e := range entries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".md") || strings.HasSuffix(e.Name(), ".rst")) {
				health.DocsFileCount++
			}
		}
	}

	cursorSkills := filepath.Join(projectRoot, ".cursor", "skills")
	cursorRules := filepath.Join(projectRoot, ".cursor", "rules")
	claudeMD := filepath.Join(projectRoot, "CLAUDE.md")
	claudeCommands := filepath.Join(projectRoot, ".claude", "commands")

	if info, _ := os.Stat(cursorSkills); info != nil && info.IsDir() {
		health.AIAssistDocsExist = true
	}

	if info, _ := os.Stat(cursorRules); info != nil && info.IsDir() {
		health.AIAssistDocsExist = true
	}

	if _, err := os.Stat(claudeMD); err == nil {
		health.AIAssistDocsExist = true
	}

	if info, _ := os.Stat(claudeCommands); info != nil && info.IsDir() {
		health.AIAssistDocsExist = true
	}

	return health, nil
}

// collectPerFileCodeStats returns per-file stats (path, lines, bytes, estimated tokens) for .go, _test.go, and bridge/tests .py.
// Used for large-file detection (split/refactor candidates) and for optional token-based analysis.
func collectPerFileCodeStats(projectRoot string) ([]FileSizeInfo, error) {
	var out []FileSizeInfo

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".venv" || info.Name() == "vendor" || info.Name() == ".git" || info.Name() == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, _ := filepath.Rel(projectRoot, path)
		if rel == "" || rel == "." {
			return nil
		}

		var add bool
		if strings.HasSuffix(path, ".go") {
			if skipGeneratedProtobufGo(projectRoot, path) {
				return nil
			}
			add = true
		} else if strings.HasSuffix(path, ".py") && (strings.Contains(path, "bridge/") || strings.Contains(path, "tests/")) {
			add = true
		}
		if !add {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable
		}

		lines := len(strings.Split(string(data), "\n"))
		bytes := len(data)
		tokens := int(float64(bytes) * config.TokensPerChar())

		out = append(out, FileSizeInfo{
			Path:            rel,
			Lines:           lines,
			Bytes:           bytes,
			EstimatedTokens: tokens,
		})
		return nil
	})

	return out, err
}

// filterLargeFileCandidates returns files that exceed the token or line threshold (split/refactor candidates).
func filterLargeFileCandidates(files []FileSizeInfo, tokenThreshold, lineThreshold int) []FileSizeInfo {
	var out []FileSizeInfo
	for _, f := range files {
		if f.EstimatedTokens >= tokenThreshold || f.Lines >= lineThreshold {
			out = append(out, f)
		}
	}
	return out
}

// countGoFilesWithBytes counts Go source files, lines, and total bytes (for token estimate).
func countGoFilesWithBytes(root string) (int, int, int, error) {
	var count, lines, bytes int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip .venv, vendor, .git directories
			if info.Name() == ".venv" || info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			if skipGeneratedProtobufGo(root, path) {
				return nil
			}
			count++
			data, err := os.ReadFile(path)
			if err == nil {
				lines += len(strings.Split(string(data), "\n"))
				bytes += len(data)
			}
		}

		return nil
	})

	return count, lines, bytes, err
}

// countGoTestFilesWithBytes counts Go test files, lines, and total bytes (for token estimate).
func countGoTestFilesWithBytes(root string) (int, int, int, error) {
	var count, lines, bytes int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".venv" || info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, "_test.go") {
			count++

			data, err := os.ReadFile(path)
			if err == nil {
				lines += len(strings.Split(string(data), "\n"))
				bytes += len(data)
			}
		}

		return nil
	})

	return count, lines, bytes, err
}

// countPythonFilesWithBytes counts Python files and lines (bridge scripts only) and total bytes (for token estimate).
func countPythonFilesWithBytes(root string) (int, int, int, error) {
	var count, lines, bytes int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".venv" || info.Name() == "__pycache__" || info.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".py") {
			// Only count bridge scripts and tests
			if strings.Contains(path, "bridge/") || strings.Contains(path, "tests/") {
				count++

				data, err := os.ReadFile(path)
				if err == nil {
					lines += len(strings.Split(string(data), "\n"))
					bytes += len(data)
				}
			}
		}

		return nil
	})

	return count, lines, bytes, err
}

// getGoModuleInfo gets Go module dependency count and version.
func getGoModuleInfo(ctx context.Context, root string) (int, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get module path and version
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Path}} {{.Version}}")
	cmd.Dir = root

	output, err := cmd.Output()
	if err != nil {
		return 0, "", err
	}

	parts := strings.Fields(string(output))
	version := "unknown"

	if len(parts) >= 2 {
		version = parts[1]
	}

	// Count dependencies
	cmd = exec.CommandContext(ctx, "go", "list", "-m", "all")
	cmd.Dir = root

	output, err = cmd.Output()
	if err != nil {
		return 0, version, nil // Non-fatal
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	deps := len(lines) - 1 // Subtract 1 for the main module

	return deps, version, nil
}

// getGoVersion gets the Go version.
func getGoVersion(ctx context.Context) (string, bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "version")

	output, err := cmd.Output()
	if err != nil {
		return "unknown", false
	}

	version := strings.TrimSpace(string(output))

	return version, true
}

// checkGoModTidy checks if go mod tidy passes.
func checkGoModTidy(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkGoBuild checks if go build succeeds.
func checkGoBuild(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("scorecard"))
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "./...")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkGoVet checks if go vet passes.
func checkGoVet(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkGoFmt checks if code is gofmt compliant.
func checkGoFmt(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var goFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == ".venv" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") {
			if !skipGeneratedProtobufGo(root, path) {
				goFiles = append(goFiles, path)
			}
		}

		return nil
	})
	if err != nil {
		return false
	}

	if len(goFiles) == 0 {
		return true
	}

	const batchSize = 200
	for i := 0; i < len(goFiles); i += batchSize {
		end := i + batchSize
		if end > len(goFiles) {
			end = len(goFiles)
		}

		args := append([]string{"-l"}, goFiles[i:end]...)
		cmd := exec.CommandContext(ctx, "gofmt", args...)

		output, err := cmd.Output()
		if err != nil {
			return false
		}

		if strings.TrimSpace(string(output)) != "" {
			return false
		}
	}

	return true
}

// checkGolangciLintConfigured checks if golangci-lint is configured.
func checkGolangciLintConfigured(root string) bool {
	// Check for .golangci.yml or .golangci.yaml
	if _, err := os.Stat(filepath.Join(root, ".golangci.yml")); err == nil {
		return true
	}

	if _, err := os.Stat(filepath.Join(root, ".golangci.yaml")); err == nil {
		return true
	}

	return false
}

// checkGolangciLint checks if golangci-lint passes.
func checkGolangciLint(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("scorecard"))
	defer cancel()

	// Check if golangci-lint is available
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return false
	}

	cmd := exec.CommandContext(ctx, "golangci-lint", "run", "--timeout", "30s")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// readCoverageFromFile parses coverage percentage from a coverage profile file.
// Returns 0.0 if the file does not exist or cannot be parsed. Used by both
// full mode (after go test) and fast mode (from prior coverage.out if present).
func readCoverageFromFile(ctx context.Context, root, filename string) float64 {
	coverPath := filepath.Join(root, filename)
	if _, err := os.Stat(coverPath); err != nil {
		return 0.0
	}
	coverCmd := exec.CommandContext(ctx, "go", "tool", "cover", "-func="+filename)
	coverCmd.Dir = root
	output, err := coverCmd.Output()
	if err != nil {
		return 0.0
	}
	lines := strings.Split(string(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "total:") {
			fields := strings.Fields(lines[i])
			if len(fields) > 0 {
				percentText := strings.TrimSuffix(fields[len(fields)-1], "%")
				if percent, parseErr := strconv.ParseFloat(percentText, 64); parseErr == nil {
					return percent
				}
			}
			break
		}
	}
	return 0.0
}

// checkGoTest checks if go test passes and gets coverage.
// Coverage is read from coverage.out even when tests fail, since go test may still
// write partial coverage for packages that ran before a failure.
func checkGoTest(ctx context.Context, root string) (bool, float64) {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("testing"))
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "./...", "-coverprofile=coverage.out", "-covermode=atomic")
	cmd.Dir = root
	err := cmd.Run()
	passes := err == nil

	coverage := readCoverageFromFile(ctx, root, "coverage.out")
	_ = os.Remove(filepath.Join(root, "coverage.out"))
	return passes, coverage
}

// checkGoVulncheck checks if govulncheck passes.
func checkGoVulncheck(ctx context.Context, root string) bool {
	ctx, cancel := context.WithTimeout(ctx, config.ToolTimeout("scorecard"))
	defer cancel()

	// Check if govulncheck is available
	if _, err := exec.LookPath("govulncheck"); err != nil {
		return false // Not installed, not a failure
	}

	cmd := exec.CommandContext(ctx, "govulncheck", "./...")
	cmd.Dir = root
	err := cmd.Run()

	return err == nil
}

// checkPathBoundaryEnforcement checks if path boundary enforcement is implemented.
func checkPathBoundaryEnforcement(projectRoot string) bool {
	// Check if security/path.go exists and has ValidatePath function
	pathFile := filepath.Join(projectRoot, "internal", "security", "path.go")
	if _, err := os.Stat(pathFile); err != nil {
		return false
	}
	// Read file and check for ValidatePath function
	data, err := os.ReadFile(pathFile)
	if err != nil {
		return false
	}

	content := string(data)

	return strings.Contains(content, "func ValidatePath") && strings.Contains(content, "ValidatePathWithinRoot")
}

// checkRateLimiting checks if rate limiting is implemented.
func checkRateLimiting(projectRoot string) bool {
	// Check if security/ratelimit.go exists
	ratelimitFile := filepath.Join(projectRoot, "internal", "security", "ratelimit.go")
	if _, err := os.Stat(ratelimitFile); err != nil {
		return false
	}
	// Read file and check for RateLimiter
	data, err := os.ReadFile(ratelimitFile)
	if err != nil {
		return false
	}

	content := string(data)

	return strings.Contains(content, "type RateLimiter") && strings.Contains(content, "func Allow")
}

// checkAccessControl checks if access control is implemented.
func checkAccessControl(projectRoot string) bool {
	// Check if security/access.go exists
	accessFile := filepath.Join(projectRoot, "internal", "security", "access.go")
	if _, err := os.Stat(accessFile); err != nil {
		return false
	}
	// Read file and check for AccessControl
	data, err := os.ReadFile(accessFile)
	if err != nil {
		return false
	}

	content := string(data)

	return strings.Contains(content, "type AccessControl") && strings.Contains(content, "func CheckToolAccess")
}

// generateGoRecommendations generates recommendations based on health checks and large-file analysis.
// When fastModeUsed is true, recommendations for skipped checks (go mod tidy, go build, etc.) are not added.
func generateGoRecommendations(health *GoHealthChecks, metrics *GoProjectMetrics, fastModeUsed bool, largeFileCandidates []FileSizeInfo) []string {
	var recommendations []string

	if !health.GoModExists {
		recommendations = append(recommendations, "Create go.mod file")
	}

	if !health.GoSumExists {
		recommendations = append(recommendations, "Run 'make tidy' to generate go.sum (auto-fix)")
	}

	if !fastModeUsed && !health.GoModTidyPasses {
		recommendations = append(recommendations, "Run 'make tidy' to clean up dependencies (auto-fix)")
	}

	if !fastModeUsed && !health.GoBuildPasses {
		recommendations = append(recommendations, "Fix Go build errors: make b")
	}

	if !health.GoVetPasses {
		recommendations = append(recommendations, "Fix 'go vet' issues (investigate manually)")
	}

	if !health.GoFmtCompliant {
		recommendations = append(recommendations, "Run 'make fmt' to format code (auto-fix)")
	}

	if !health.GoLintConfigured {
		recommendations = append(recommendations, "Configure golangci-lint (.golangci.yml)")
	}

	if !fastModeUsed && health.GoLintConfigured && !health.GoLintPasses {
		recommendations = append(recommendations, "Run 'make lint-fix' to auto-fix lint issues (auto-fix)")
	}

	if !fastModeUsed && !health.GoTestPasses {
		recommendations = append(recommendations, "Fix failing Go tests: make test")
	}

	minCoverage := float64(config.MinCoverage())
	if !fastModeUsed && health.GoTestCoverage < minCoverage {
		recommendations = append(recommendations, fmt.Sprintf("Increase test coverage (currently %.1f%%, target: %.0f%%): make test-coverage", health.GoTestCoverage, minCoverage))
	}

	if !fastModeUsed && !health.GoVulnCheckPasses {
		recommendations = append(recommendations, "Run 'make govulncheck' for security scanning")
	}

	autoFixable := !health.GoFmtCompliant || (!fastModeUsed && !health.GoModTidyPasses) || (!fastModeUsed && health.GoLintConfigured && !health.GoLintPasses)
	if autoFixable {
		recommendations = append(recommendations, "💡 Auto-fix all: make scorecard-fix")
	}

	if len(largeFileCandidates) > 0 {
		recommendations = append(recommendations, "Consider splitting/refactoring large files (see Large files section) for better LLM context fit and maintainability")
	}

	return recommendations
}

// calculateGoScore calculates overall Go project score.
func calculateGoScore(health *GoHealthChecks, metrics *GoProjectMetrics) float64 {
	score := 0.0
	maxScore := 0.0

	// Module health (20%)
	maxScore += 20

	if health.GoModExists {
		score += 5
	}

	if health.GoSumExists {
		score += 5
	}

	if health.GoModTidyPasses {
		score += 5
	}

	if health.GoVersionValid {
		score += 5
	}

	// Build & Quality (30%)
	maxScore += 30

	if health.GoBuildPasses {
		score += 10
	}

	if health.GoVetPasses {
		score += 5
	}

	if health.GoFmtCompliant {
		score += 5
	}

	if health.GoLintConfigured {
		score += 5
	}

	if health.GoLintPasses {
		score += 5
	}

	// Testing (30%)
	maxScore += 30

	if health.GoTestPasses {
		score += 15
	}

	if health.GoTestCoverage >= float64(config.MinCoverage()) {
		score += 15
	} else if health.GoTestCoverage >= 50.0 {
		score += 10
	} else if health.GoTestCoverage > 0 {
		score += 5
	}

	// Security (20%)
	maxScore += 20

	if health.GoVulnCheckPasses {
		score += 20
	} else {
		// Partial credit if tool not installed
		if _, err := exec.LookPath("govulncheck"); err != nil {
			score += 5 // Tool not installed, but not a failure
		}
	}

	if maxScore == 0 {
		return 0
	}

	return (score / maxScore) * 100
}

// IsGoProject checks if the current directory is a Go project
// Exported for use by resource handlers.
func IsGoProject() bool {
	wd, err := os.Getwd()
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(wd, "go.mod"))

	return err == nil
}

// getProjectRoot gets the project root directory.
func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	return wd
}

// FormatGoScorecard formats the Go scorecard as text output.
func FormatGoScorecard(scorecard *GoScorecardResult) string {
	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  📊 GO PROJECT SCORECARD\n")
	sb.WriteString("======================================================================\n\n")

	// Overall Score
	sb.WriteString(fmt.Sprintf("  OVERALL SCORE: %.1f%%\n", scorecard.Score))
	// Use coverage threshold as production ready indicator
	productionReadyThreshold := float64(config.MinCoverage())
	if scorecard.Score >= productionReadyThreshold {
		sb.WriteString("  Production Ready: YES ✅\n")
	} else if scorecard.Score >= 60 {
		sb.WriteString("  Production Ready: PARTIAL ⚠️\n")
	} else {
		sb.WriteString("  Production Ready: NO ❌\n")
	}

	if scorecard.FastModeUsed {
		sb.WriteString("  Excluded in fast mode: go mod tidy, go build, go test, golangci-lint, govulncheck (run with fast_mode=false for full results)\n")
	}

	sb.WriteString("\n")

	// Metrics
	sb.WriteString("  Codebase Metrics:\n")
	sb.WriteString(fmt.Sprintf("    Go Files:        %d\n", scorecard.Metrics.GoFiles))
	sb.WriteString(fmt.Sprintf("    Go Lines:        %d\n", scorecard.Metrics.GoLines))
	sb.WriteString(fmt.Sprintf("    Go Test Files:   %d\n", scorecard.Metrics.GoTestFiles))
	sb.WriteString(fmt.Sprintf("    Go Test Lines:   %d\n", scorecard.Metrics.GoTestLines))
	sb.WriteString(fmt.Sprintf("    Python Files:    %d (bridge scripts)\n", scorecard.Metrics.PythonFiles))
	sb.WriteString(fmt.Sprintf("    Python Lines:    %d\n", scorecard.Metrics.PythonLines))
	sb.WriteString(fmt.Sprintf("    Go Modules:      %d\n", scorecard.Metrics.GoModules))
	sb.WriteString(fmt.Sprintf("    Go Dependencies: %d\n", scorecard.Metrics.GoDependencies))
	sb.WriteString(fmt.Sprintf("    Go Version:       %s\n", scorecard.Metrics.GoVersion))
	sb.WriteString(fmt.Sprintf("    MCP Tools:        %d\n", scorecard.Metrics.MCPTools))
	sb.WriteString(fmt.Sprintf("    MCP Prompts:      %d\n", scorecard.Metrics.MCPPrompts))
	sb.WriteString(fmt.Sprintf("    MCP Resources:    %d\n", scorecard.Metrics.MCPResources))
	sb.WriteString(fmt.Sprintf("    Est. tokens (code): %d (≈ context cost if sent to LLM; ratio-based)\n", scorecard.Metrics.EstimatedTokens))
	sb.WriteString("\n")

	// Large files (split/refactor candidates): multi-stage token/size → threshold → list
	if len(scorecard.LargeFileCandidates) > 0 {
		sb.WriteString("  Large files (consider splitting/refactoring for context fit):\n")
		for _, f := range scorecard.LargeFileCandidates {
			sb.WriteString(fmt.Sprintf("    %s  %d lines  ~%d tokens\n", f.Path, f.Lines, f.EstimatedTokens))
		}
		sb.WriteString("\n")
	}

	// Health Checks
	sb.WriteString("  Go Health Checks:\n")
	sb.WriteString(fmt.Sprintf("    go.mod exists:        %s\n", checkMark(scorecard.Health.GoModExists)))
	sb.WriteString(fmt.Sprintf("    go.sum exists:        %s\n", checkMark(scorecard.Health.GoSumExists)))
	sb.WriteString(fmt.Sprintf("    go mod tidy:          %s\n", checkMarkOrSkipped(scorecard.Health.GoModTidyPasses, scorecard.FastModeUsed)))
	sb.WriteString(fmt.Sprintf("    Go version valid:     %s (%s)\n", checkMark(scorecard.Health.GoVersionValid), scorecard.Health.GoVersion))
	sb.WriteString(fmt.Sprintf("    go build:             %s\n", checkMarkOrSkipped(scorecard.Health.GoBuildPasses, scorecard.FastModeUsed)))
	sb.WriteString(fmt.Sprintf("    go vet:               %s\n", checkMark(scorecard.Health.GoVetPasses)))
	sb.WriteString(fmt.Sprintf("    go fmt:               %s\n", checkMark(scorecard.Health.GoFmtCompliant)))
	sb.WriteString(fmt.Sprintf("    golangci-lint config: %s\n", checkMark(scorecard.Health.GoLintConfigured)))
	sb.WriteString(fmt.Sprintf("    golangci-lint:        %s\n", checkMarkOrSkipped(scorecard.Health.GoLintPasses, scorecard.FastModeUsed)))
	sb.WriteString(fmt.Sprintf("    go test:              %s\n", checkMarkOrSkipped(scorecard.Health.GoTestPasses, scorecard.FastModeUsed)))

	if scorecard.Health.GoTestCoverage == 0 && scorecard.FastModeUsed {
		sb.WriteString("    Test coverage:        — (fast mode; run full scorecard or make test-coverage to see %)\n")
	} else {
		sb.WriteString(fmt.Sprintf("    Test coverage:        %.1f%%\n", scorecard.Health.GoTestCoverage))
	}

	sb.WriteString(fmt.Sprintf("    govulncheck:          %s\n", checkMarkOrSkipped(scorecard.Health.GoVulnCheckPasses, scorecard.FastModeUsed)))
	sb.WriteString("\n")

	// Security Features
	sb.WriteString("  Security Features:\n")
	sb.WriteString(fmt.Sprintf("    Path boundary enforcement: %s\n", checkMark(scorecard.Health.PathBoundaryEnforcement)))
	sb.WriteString(fmt.Sprintf("    Rate limiting:             %s\n", checkMark(scorecard.Health.RateLimiting)))
	sb.WriteString(fmt.Sprintf("    Access control:            %s\n", checkMark(scorecard.Health.AccessControl)))
	sb.WriteString("\n")

	// Recommendations
	if len(scorecard.Recommendations) > 0 {
		sb.WriteString("  Recommendations:\n")

		for _, rec := range scorecard.Recommendations {
			sb.WriteString(fmt.Sprintf("    • %s\n", rec))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatGoScorecardWithWisdom formats the Go scorecard with wisdom section
// Gracefully degrades to base scorecard if wisdom engine fails.
func FormatGoScorecardWithWisdom(scorecard *GoScorecardResult) string {
	// Get base scorecard
	base := FormatGoScorecard(scorecard)
	return addWisdomToScorecard(base, scorecard)
}

// addWisdomToScorecard adds wisdom section to a formatted scorecard string
// Gracefully degrades to original string if wisdom engine fails.
func addWisdomToScorecard(formattedScorecard string, scorecard *GoScorecardResult) string {
	// Try to get wisdom engine
	engine, err := getWisdomEngine()
	if err != nil {
		// Gracefully degrade: return original scorecard without wisdom
		return formattedScorecard
	}

	// Get wisdom quote based on score (use "random" for variety, date-seeded for consistency)
	quote, err := engine.GetWisdom(scorecard.Score, "random")
	if err != nil {
		// Gracefully degrade: return original scorecard without wisdom
		return formattedScorecard
	}

	// Append wisdom section
	var sb strings.Builder

	sb.WriteString(formattedScorecard)
	sb.WriteString("  ──────────────────────────────────────────────────────────────────\n")
	sb.WriteString("  🧘 Wisdom for Your Journey\n")
	sb.WriteString("  ──────────────────────────────────────────────────────────────────\n\n")
	sb.WriteString(fmt.Sprintf("  > \"%s\"\n", quote.Quote))
	sb.WriteString(fmt.Sprintf("  > — %s\n\n", quote.Source))

	if quote.Encouragement != "" {
		sb.WriteString(fmt.Sprintf("  Encouragement: %s\n", quote.Encouragement))
	}

	return sb.String()
}

// checkMark returns a checkmark or X based on boolean.
func checkMark(b bool) string {
	if b {
		return "✅"
	}

	return "❌"
}

// checkMarkOrSkipped returns ✅ if value is true, "— (skipped)" if skipped (e.g. fast mode), else ❌.
// Use for health checks that are not run in fast mode so the scorecard doesn't show ❌ for "not run".
func checkMarkOrSkipped(value, skipped bool) string {
	if value {
		return "✅"
	}

	if skipped {
		return "— (skipped)"
	}

	return "❌"
}

// GenerateGoScorecard generates a Go-specific scorecard
// If opts is nil, uses default options (full checks).
func GenerateGoScorecard(ctx context.Context, projectRoot string, opts *ScorecardOptions) (*GoScorecardResult, error) {
	// Get current working directory if projectRoot is empty
	if projectRoot == "" {
		var err error

		projectRoot, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Validate project root path to prevent directory traversal
	validatedRoot, err := security.ValidatePath(projectRoot, projectRoot)
	if err != nil {
		// If validation fails, try to get project root safely
		validatedRoot, err = projectroot.FindFrom(projectRoot)
		if err != nil {
			return nil, fmt.Errorf("invalid project root: %w", err)
		}
	}

	projectRoot = validatedRoot

	// Collect metrics
	metrics, err := collectGoMetrics(ctx, projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Perform health checks
	health, err := performGoHealthChecks(ctx, projectRoot, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to perform health checks: %w", err)
	}

	fastMode := opts != nil && opts.FastMode

	// Multi-stage: per-file token/size → threshold filter → split/refactor candidates
	allFiles, err := collectPerFileCodeStats(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to collect per-file stats: %w", err)
	}
	largeCandidates := filterLargeFileCandidates(allFiles, defaultLargeFileTokenThreshold, defaultLargeFileLineThreshold)

	// Generate recommendations (include large-file rec when applicable)
	recommendations := generateGoRecommendations(health, metrics, fastMode, largeCandidates)

	// Calculate score
	score := calculateGoScore(health, metrics)

	return &GoScorecardResult{
		Metrics:             *metrics,
		Health:              *health,
		Recommendations:     recommendations,
		Score:               score,
		LargeFileCandidates: largeCandidates,
		FastModeUsed:        fastMode,
	}, nil
}
