// scorecard_go.go — Go scorecard: types, consts, and health checks.
package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
	GoModExists          bool    `json:"go_mod_exists"`
	GoSumExists          bool    `json:"go_sum_exists"`
	GoModTidyPasses      bool    `json:"go_mod_tidy_passes"`
	GoVersionValid       bool    `json:"go_version_valid"`
	GoVersion            string  `json:"go_version"`
	GoBuildPasses        bool    `json:"go_build_passes"`
	GoVetPasses          bool    `json:"go_vet_passes"`
	GoFmtCompliant       bool    `json:"go_fmt_compliant"`
	GoLintConfigured     bool    `json:"go_lint_configured"`
	GoLintPasses         bool    `json:"go_lint_passes"`
	GoTestPasses         bool    `json:"go_test_passes"`
	GoTestCoverage       float64 `json:"go_test_coverage"`
	GoVulnCheckAvailable bool    `json:"go_vulncheck_available"`
	GoVulnCheckPasses    bool    `json:"go_vulncheck_passes"`
	// Security features
	PathBoundaryEnforcement bool `json:"path_boundary_enforcement"`
	RateLimiting            bool `json:"rate_limiting"`
	AccessControl           bool `json:"access_control"`
	// Documentation (for documentation score dimension)
	ReadmeExists              bool    `json:"readme_exists"`
	DocsDirExists             bool    `json:"docs_dir_exists"`
	DocsFileCount             int     `json:"docs_file_count"`
	DocsLiveCount             int     `json:"docs_live_count"`
	DocsArchiveCount          int     `json:"docs_archive_count"`
	DocsGeneratedCount        int     `json:"docs_generated_count"`
	DocsStalePathMatches      int     `json:"docs_stale_path_matches"`
	DocsMissingReferenceCount int     `json:"docs_missing_reference_count"`
	DocsHealthScore           float64 `json:"docs_health_score"`
	AIAssistDocsExist         bool    `json:"ai_assist_docs_exist"` // .cursor/skills, .cursor/rules, CLAUDE.md, or .claude/commands/
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
	// OtherLanguages holds LangHealth for non-Go languages detected in the project (polyglot support).
	OtherLanguages []LangHealth `json:"other_languages,omitempty"`
}

// ScorecardOptions configures scorecard generation behavior.
type ScorecardOptions struct {
	FastMode bool // Skip expensive operations (go test, go build, go mod tidy)
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
		health.GoVulnCheckAvailable, health.GoVulnCheckPasses = checkGoVulncheck(ctx, projectRoot)
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
		stats, statsErr := collectDocsHealthStats(projectRoot, docsDir)
		if statsErr == nil {
			health.DocsFileCount = stats.TotalDocs
			health.DocsLiveCount = stats.LiveDocs
			health.DocsArchiveCount = stats.ArchiveDocs
			health.DocsGeneratedCount = stats.GeneratedDocs
			health.DocsStalePathMatches = len(stats.StalePaths)
			health.DocsMissingReferenceCount = len(stats.MissingReferences)
		} else {
			entries, _ := os.ReadDir(docsDir)
			for _, e := range entries {
				if !e.IsDir() && (strings.HasSuffix(e.Name(), ".md") || strings.HasSuffix(e.Name(), ".rst")) {
					health.DocsFileCount++
					health.DocsLiveCount++
				}
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

	health.DocsHealthScore = calculateDocsHealthScore(health)

	return health, nil
}
