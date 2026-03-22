// scorecard_go_format.go — Go scorecard: recommendations, scoring, formatting, and public API.
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/projectroot"
	"github.com/davidl71/exarp-go/internal/security"
	"golang.org/x/sync/singleflight"
)

var scorecardFlight singleflight.Group

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
		recommendations = append(recommendations, "Fix failing tests: make test")
	}

	minCoverage := float64(config.MinCoverage())
	if !fastModeUsed && health.GoTestCoverage < minCoverage {
		recommendations = append(recommendations, fmt.Sprintf("Increase test coverage (currently %.1f%%, target: %.0f%%): make test-coverage", health.GoTestCoverage, minCoverage))
	}

	if !fastModeUsed && health.GoVulnCheckAvailable && !health.GoVulnCheckPasses {
		recommendations = append(recommendations, "Run 'make govulncheck' for security scanning")
	}

	if !fastModeUsed && !health.GoVulnCheckAvailable {
		recommendations = append(recommendations, "Install or expose 'govulncheck' in PATH so the scorecard can verify dependency vulnerabilities")
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

	// Module health (15%)
	maxScore += 15

	if health.GoModExists {
		score += 3.75
	}

	if health.GoSumExists {
		score += 3.75
	}

	if health.GoModTidyPasses {
		score += 3.75
	}

	if health.GoVersionValid {
		score += 3.75
	}

	// Build & Quality (25%)
	maxScore += 25

	if health.GoBuildPasses {
		score += 8.3333333333
	}

	if health.GoVetPasses {
		score += 4.1666666667
	}

	if health.GoFmtCompliant {
		score += 4.1666666667
	}

	if health.GoLintConfigured {
		score += 4.1666666667
	}

	if health.GoLintPasses {
		score += 4.1666666667
	}

	// Testing (25%)
	maxScore += 25

	if health.GoTestPasses {
		score += 12.5
	}

	if health.GoTestCoverage >= float64(config.MinCoverage()) {
		score += 12.5
	} else if health.GoTestCoverage >= 50.0 {
		score += 8.3333333333
	} else if health.GoTestCoverage > 0 {
		score += 4.1666666667
	}

	// Security (20%)
	maxScore += 20

	if health.GoVulnCheckPasses {
		score += 20
	} else {
		// Partial credit if tool not installed
		if !health.GoVulnCheckAvailable {
			score += 5 // Tool not installed, but not a failure
		}
	}

	// Documentation (15%)
	maxScore += 15
	score += (calculateDocsHealthScore(health) / 100.0) * 15

	if maxScore == 0 {
		return 0
	}

	return (score / maxScore) * 100
}

// IsGoProject reports whether the project root (from FindProjectRoot, e.g. PROJECT_ROOT) contains go.mod.
// Uses project root instead of os.Getwd() so MCP invocations with PROJECT_ROOT set are evaluated correctly.
func IsGoProject() bool {
	root, err := FindProjectRoot()
	if err != nil || root == "" {
		return false
	}

	_, err = os.Stat(filepath.Join(root, "go.mod"))
	return err == nil
}

// FormatGoScorecard formats the Go scorecard as text output.
func FormatGoScorecard(scorecard *GoScorecardResult) string {
	var sb strings.Builder

	sb.WriteString("======================================================================\n")
	sb.WriteString("  📊 PROJECT SCORECARD\n")
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

	sb.WriteString(fmt.Sprintf("    govulncheck:          %s\n", checkMarkSkippedOrUnavailable(scorecard.Health.GoVulnCheckPasses, scorecard.FastModeUsed, !scorecard.Health.GoVulnCheckAvailable)))
	sb.WriteString("\n")

	// Security Features
	sb.WriteString("  Security Features:\n")
	sb.WriteString(fmt.Sprintf("    Path boundary enforcement: %s\n", checkMark(scorecard.Health.PathBoundaryEnforcement)))
	sb.WriteString(fmt.Sprintf("    Rate limiting:             %s\n", checkMark(scorecard.Health.RateLimiting)))
	sb.WriteString(fmt.Sprintf("    Access control:            %s\n", checkMark(scorecard.Health.AccessControl)))
	sb.WriteString("\n")

	// Other Languages (polyglot support)
	if len(scorecard.OtherLanguages) > 0 {
		sb.WriteString("  Other Languages:\n")
		for _, lang := range scorecard.OtherLanguages {
			title := strings.ToUpper(lang.Lang)
			if lang.LangRoot != "" {
				title += " (" + lang.LangRoot + ")"
			}
			sb.WriteString(fmt.Sprintf("    %s:\n", title))
			sb.WriteString(fmt.Sprintf("      Files:  %d\n", lang.FileCount))
			sb.WriteString(fmt.Sprintf("      Score: %.1f%%\n", lang.Score))
			sb.WriteString(fmt.Sprintf("      Build: %s\n", checkMark(lang.BuildPasses)))
			sb.WriteString(fmt.Sprintf("      Test:  %s\n", checkMark(lang.TestPasses)))
			sb.WriteString(fmt.Sprintf("      Lint:  %s\n", checkMark(lang.LintPasses)))
			sb.WriteString(fmt.Sprintf("      Fmt:   %s\n", checkMark(lang.FmtPasses)))
			if len(lang.Recommendations) > 0 {
				for _, rec := range lang.Recommendations {
					sb.WriteString(fmt.Sprintf("      • %s\n", rec))
				}
			}
		}
		sb.WriteString("\n")
	}

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

func checkMarkSkippedOrUnavailable(value, skipped, unavailable bool) string {
	if value {
		return "✅"
	}

	if skipped {
		return "— (skipped)"
	}

	if unavailable {
		return "— (unavailable)"
	}

	return "❌"
}

// GenerateGoScorecard generates a Go-specific scorecard
// If opts is nil, uses default options (full checks).
func GenerateGoScorecard(ctx context.Context, projectRoot string, opts *ScorecardOptions) (*GoScorecardResult, error) {
	// Get current working directory if projectRoot is empty
	if projectRoot == "" {
		var err error
		projectRoot, err = GetProjectRootWithFallback()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project root: %w", err)
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

	// Collect multi-language health for non-Go languages (polyglot support)
	otherLangs := collectOtherLanguagesHealth(ctx, projectRoot)

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
		OtherLanguages:      otherLangs,
	}, nil
}

// collectOtherLanguagesHealth collects LangHealth for non-Go languages in the project.
func collectOtherLanguagesHealth(ctx context.Context, projectRoot string) []LangHealth {
	var otherLangs []LangHealth
	allLangs := CollectMultilangHealth(ctx, projectRoot)
	for _, lang := range allLangs {
		if lang.Lang != "go" && lang.Detected {
			otherLangs = append(otherLangs, lang)
		}
	}
	return otherLangs
}
