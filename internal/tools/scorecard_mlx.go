package tools

import (
	"fmt"
	"strings"
)

// GoScorecardToMap converts GoScorecardResult to map for MLX processing
// Exported for use by resource handlers.
func GoScorecardToMap(scorecard *GoScorecardResult) map[string]interface{} {
	result := make(map[string]interface{})

	result["overall_score"] = scorecard.Score
	result["scores"] = map[string]interface{}{
		"testing":       calculateTestingScore(scorecard),
		"security":      calculateSecurityScore(scorecard),
		"documentation": calculateDocumentationScore(scorecard),
		"completion":    calculateCompletionScore(scorecard),
		"ci_cd":         calculateCICDScore(scorecard),
	}

	result["blockers"] = ExtractBlockers(scorecard)
	result["recommendations"] = scorecard.Recommendations
	result["metrics"] = map[string]interface{}{
		"go_files":      scorecard.Metrics.GoFiles,
		"go_lines":      scorecard.Metrics.GoLines,
		"go_test_files": scorecard.Metrics.GoTestFiles,
		"go_test_lines": scorecard.Metrics.GoTestLines,
		"test_coverage": scorecard.Health.GoTestCoverage,
		"mcp_tools":     scorecard.Metrics.MCPTools,
		"mcp_prompts":   scorecard.Metrics.MCPPrompts,
		"mcp_resources": scorecard.Metrics.MCPResources,
	}

	return result
}

// FormatGoScorecardWithMLX formats scorecard with MLX insights.
func FormatGoScorecardWithMLX(scorecard *GoScorecardResult, insights map[string]interface{}) string {
	var sb strings.Builder

	// Standard scorecard format
	sb.WriteString(FormatGoScorecard(scorecard))

	// Add AI insights section (from insight provider: MLX or FM)
	if insightsText, ok := insights["insights"].(string); ok && insightsText != "" {
		sb.WriteString("\n\n")
		sb.WriteString("═══════════════════════════════════════════════════════════\n")
		sb.WriteString("🤖 AI-Generated Insights\n")
		sb.WriteString("═══════════════════════════════════════════════════════════\n\n")

		if model, ok := insights["model"].(string); ok {
			sb.WriteString(fmt.Sprintf("Model: %s\n\n", model))
		}

		if generatedBy, ok := insights["generated_by"].(string); ok {
			sb.WriteString(fmt.Sprintf("Source: %s\n\n", generatedBy))
		}

		sb.WriteString(insightsText)
		sb.WriteString("\n")
	}

	return sb.String()
}

// Helper functions to calculate component scores.
func calculateTestingScore(scorecard *GoScorecardResult) float64 {
	score := 0.0
	checks := 0

	if scorecard.Health.GoTestPasses {
		score += 30
		checks++
	}

	if scorecard.Health.GoTestCoverage >= 80 {
		score += 40
		checks++
	} else if scorecard.Health.GoTestCoverage >= 50 {
		score += 20
		checks++
	}

	if scorecard.Metrics.GoTestFiles > 0 {
		score += 30
		checks++
	}

	if checks > 0 {
		return score / float64(checks) * 100
	}

	return 0
}

func calculateSecurityScore(scorecard *GoScorecardResult) float64 {
	score := 0.0
	checks := 0

	if scorecard.Health.GoVulnCheckPasses {
		score += 40
		checks++
	}

	if scorecard.Health.PathBoundaryEnforcement {
		score += 30
		checks++
	}

	if scorecard.Health.RateLimiting {
		score += 15
		checks++
	}

	if scorecard.Health.AccessControl {
		score += 15
		checks++
	}

	if checks > 0 {
		return score / float64(checks) * 100
	}

	return 50 // Default if no checks
}

func calculateDocumentationScore(scorecard *GoScorecardResult) float64 {
	score := 0.0
	checks := 0

	if scorecard.Health.ReadmeExists {
		score += 40
		checks++
	}

	if scorecard.Health.DocsDirExists {
		score += 30
		checks++
	}

	if scorecard.Health.DocsFileCount > 0 {
		score += 15
		checks++
	}

	if scorecard.Health.CursorDocsExist {
		score += 15
		checks++
	}

	if checks > 0 {
		return score
	}

	return 0
}

func calculateCompletionScore(scorecard *GoScorecardResult) float64 {
	// Completion scoring would need task metrics
	// For now, return a default
	return 50.0
}

func calculateCICDScore(scorecard *GoScorecardResult) float64 {
	score := 0.0
	checks := 0

	if scorecard.Health.GoBuildPasses {
		score += 25
		checks++
	}

	if scorecard.Health.GoTestPasses {
		score += 25
		checks++
	}

	if scorecard.Health.GoLintPasses {
		score += 25
		checks++
	}

	if scorecard.Health.GoVetPasses {
		score += 25
		checks++
	}

	if checks > 0 {
		return score / float64(checks) * 100
	}

	return 0
}

// ExtractBlockers extracts blockers from scorecard
// Exported for use by resource handlers.
func ExtractBlockers(scorecard *GoScorecardResult) []string {
	blockers := []string{}

	if !scorecard.Health.GoModExists {
		blockers = append(blockers, "Missing go.mod file")
	}

	if !scorecard.Health.GoBuildPasses {
		blockers = append(blockers, "Go build fails")
	}

	if !scorecard.Health.GoTestPasses {
		blockers = append(blockers, "Go tests fail")
	}

	if scorecard.Health.GoTestCoverage < 50 && scorecard.Metrics.GoTestFiles > 0 {
		blockers = append(blockers, fmt.Sprintf("Low test coverage: %.1f%%", scorecard.Health.GoTestCoverage))
	}

	if !scorecard.Health.GoVulnCheckPasses {
		blockers = append(blockers, "Security vulnerabilities detected")
	}

	return blockers
}
