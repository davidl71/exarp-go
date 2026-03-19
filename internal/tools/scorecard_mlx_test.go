package tools

import (
	"math"
	"testing"
)

func TestExtractBlockers_DoesNotFlagUnavailableGovulncheckAsVulnerability(t *testing.T) {
	scorecard := &GoScorecardResult{
		Metrics: GoProjectMetrics{GoTestFiles: 1},
		Health: GoHealthChecks{
			GoModExists:             true,
			GoBuildPasses:           true,
			GoTestPasses:            true,
			GoTestCoverage:          75,
			GoVulnCheckAvailable:    false,
			GoVulnCheckPasses:       false,
			PathBoundaryEnforcement: true,
			RateLimiting:            true,
			AccessControl:           true,
		},
	}

	blockers := ExtractBlockers(scorecard)
	for _, blocker := range blockers {
		if blocker == "Security vulnerabilities detected" {
			t.Fatalf("unexpected security blocker when govulncheck is unavailable: %v", blockers)
		}
	}
}

func TestExtractBlockers_FlagsFailedGovulncheckWhenAvailable(t *testing.T) {
	scorecard := &GoScorecardResult{
		Health: GoHealthChecks{
			GoModExists:          true,
			GoBuildPasses:        true,
			GoTestPasses:         true,
			GoVulnCheckAvailable: true,
			GoVulnCheckPasses:    false,
		},
	}

	blockers := ExtractBlockers(scorecard)
	found := false
	for _, blocker := range blockers {
		if blocker == "Security vulnerabilities detected" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected security blocker when govulncheck is available and failing: %v", blockers)
	}
}

func TestExtractBlockers_FlagsDocumentationDrift(t *testing.T) {
	scorecard := &GoScorecardResult{
		Health: GoHealthChecks{
			GoModExists:               true,
			GoBuildPasses:             true,
			GoTestPasses:              true,
			DocsStalePathMatches:      2,
			DocsMissingReferenceCount: 1,
		},
	}

	blockers := ExtractBlockers(scorecard)
	foundStale := false
	foundMissing := false
	for _, blocker := range blockers {
		if blocker == "Stale documentation paths: 2" {
			foundStale = true
		}
		if blocker == "Broken documentation references: 1" {
			foundMissing = true
		}
	}
	if !foundStale || !foundMissing {
		t.Fatalf("expected documentation blockers, got %v", blockers)
	}
}

func TestCalculateDocumentationScore_UsesDocsHealthSignals(t *testing.T) {
	scorecard := &GoScorecardResult{
		Health: GoHealthChecks{
			ReadmeExists:              true,
			DocsDirExists:             true,
			DocsLiveCount:             3,
			DocsStalePathMatches:      1,
			DocsMissingReferenceCount: 0,
			AIAssistDocsExist:         true,
		},
	}

	score := calculateDocumentationScore(scorecard)
	if score != 90 {
		t.Fatalf("documentation score = %v, want 90", score)
	}
}

func TestCalculateGoScore_IncludesDocumentationDimension(t *testing.T) {
	metrics := &GoProjectMetrics{}
	healthy := &GoHealthChecks{
		GoModExists:               true,
		GoSumExists:               true,
		GoModTidyPasses:           true,
		GoVersionValid:            true,
		GoBuildPasses:             true,
		GoVetPasses:               true,
		GoFmtCompliant:            true,
		GoLintConfigured:          true,
		GoLintPasses:              true,
		GoTestPasses:              true,
		GoTestCoverage:            90,
		GoVulnCheckAvailable:      true,
		GoVulnCheckPasses:         true,
		ReadmeExists:              true,
		DocsDirExists:             true,
		DocsLiveCount:             2,
		DocsStalePathMatches:      0,
		DocsMissingReferenceCount: 0,
	}
	drifted := *healthy
	drifted.DocsStalePathMatches = 2
	drifted.DocsMissingReferenceCount = 1

	healthyScore := calculateGoScore(healthy, metrics)
	driftedScore := calculateGoScore(&drifted, metrics)

	if healthyScore <= driftedScore {
		t.Fatalf("expected docs health to affect overall score: healthy=%v drifted=%v", healthyScore, driftedScore)
	}
	if math.Abs(healthyScore-100) > 1e-6 {
		t.Fatalf("expected fully healthy scorecard to score 100, got %v", healthyScore)
	}
}
