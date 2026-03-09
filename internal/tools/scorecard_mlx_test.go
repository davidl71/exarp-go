package tools

import "testing"

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
