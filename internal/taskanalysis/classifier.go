// Package taskanalysis provides heuristic task complexity classification for the
// Model-Assisted Workflow. Tag hints for Todo2: #feature #task-analysis
package taskanalysis

import (
	"strings"

	"github.com/davidl71/exarp-go/internal/models"
)

// TaskComplexity indicates whether a task is simple (auto-executable), medium
// (needs breakdown), or complex (needs human review).
type TaskComplexity string

const (
	ComplexitySimple  TaskComplexity = "simple"  // Can auto-execute
	ComplexityMedium  TaskComplexity = "medium"  // Needs breakdown
	ComplexityComplex TaskComplexity = "complex" // Needs human review
)

// AnalysisResult holds the result of AnalyzeTask.
type AnalysisResult struct {
	Complexity     TaskComplexity
	CanAutoExecute bool
	NeedsBreakdown bool
	Reason         string
}

// AnalyzeTask determines task complexity and recommendations using heuristic rules.
// Decision criteria (from MODEL_ASSISTED_WORKFLOW.md):
// - Simple: Well-defined, routine, low-risk, <1h estimated
// - Medium: Needs breakdown, multiple steps, some uncertainty
// - Complex: High-stakes, experimental, requires human judgment.
func AnalyzeTask(task *models.Todo2Task) AnalysisResult {
	if task == nil {
		return AnalysisResult{
			Complexity:     ComplexityMedium,
			CanAutoExecute: false,
			NeedsBreakdown: true,
			Reason:         "nil task, default to medium",
		}
	}

	text := strings.ToLower(task.Content + " " + task.LongDescription)
	depCount := len(task.Dependencies)

	tagLower := make([]string, len(task.Tags))
	for i, t := range task.Tags {
		tagLower[i] = strings.ToLower(t)
	}

	textWithTags := text + " " + strings.Join(tagLower, " ")

	// Complex indicators: high-stakes keywords
	if hasComplexKeywords(textWithTags) {
		return AnalysisResult{
			Complexity:     ComplexityComplex,
			CanAutoExecute: false,
			NeedsBreakdown: true,
			Reason:         "high-stakes or experimental keywords detected",
		}
	}

	// Many dependencies -> medium or complex
	if depCount >= 5 {
		return AnalysisResult{
			Complexity:     ComplexityComplex,
			CanAutoExecute: false,
			NeedsBreakdown: true,
			Reason:         "many dependencies (>=5)",
		}
	}

	if depCount >= 3 {
		return AnalysisResult{
			Complexity:     ComplexityMedium,
			CanAutoExecute: false,
			NeedsBreakdown: true,
			Reason:         "moderate dependencies (>=3)",
		}
	}

	// Long description -> likely needs breakdown
	descLen := len(task.LongDescription) + len(task.Content)
	if descLen > 1500 {
		return AnalysisResult{
			Complexity:     ComplexityMedium,
			CanAutoExecute: false,
			NeedsBreakdown: true,
			Reason:         "long description suggests multiple steps",
		}
	}

	// Simple indicators: routine, well-defined
	if hasSimpleKeywords(textWithTags) && depCount == 0 && descLen < 500 {
		return AnalysisResult{
			Complexity:     ComplexitySimple,
			CanAutoExecute: true,
			NeedsBreakdown: false,
			Reason:         "routine task, low complexity",
		}
	}

	// Default: medium (conservative for empty/minimal)
	return AnalysisResult{
		Complexity:     ComplexityMedium,
		CanAutoExecute: false,
		NeedsBreakdown: true,
		Reason:         "default heuristic",
	}
}

func hasComplexKeywords(s string) bool {
	complex := []string{
		"security", "migration", "migrate", "experimental", "refactor", "architecture",
		"breaking", "critical", "high-stakes", "redesign", "rewrite",
	}
	for _, k := range complex {
		if strings.Contains(s, k) {
			return true
		}
	}

	return false
}

func hasSimpleKeywords(s string) bool {
	simple := []string{
		"add comment", "fix typo", "update readme", "format", "lint",
		"rename", "add test", "document", "docs", "changelog",
	}
	for _, k := range simple {
		if strings.Contains(s, k) {
			return true
		}
	}

	return false
}
