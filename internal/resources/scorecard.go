package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/tools"
)

// handleScorecard handles the stdio://scorecard resource.
// Native Go only: uses GenerateGoScorecard for Go projects; returns a clear JSON error for non-Go projects
// (aligned with report tool: scorecard is Go-only; no Python bridge fallback).
func handleScorecard(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	if !tools.IsGoProject() {
		errResult := map[string]interface{}{
			"success":   false,
			"error":     "stdio://scorecard is supported only for Go projects; use report(action=\"scorecard\") for Go projects",
			"uri":       uri,
			"timestamp": time.Now().Format(time.RFC3339),
		}
		jsonData, _ := json.Marshal(errResult)

		return jsonData, "application/json", nil
	}

	opts := &tools.ScorecardOptions{FastMode: true}

	scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate Go scorecard: %w", err)
	}

	result := convertScorecardToResourceFormat(scorecard)

	jsonData, err := json.Marshal(result)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal scorecard: %w", err)
	}

	return jsonData, "application/json", nil
}

// convertScorecardToResourceFormat converts GoScorecardResult to resource JSON format
// Matches Python generate_project_scorecard() output format
// Reuses helper functions from internal/tools/scorecard_mlx.go.
func convertScorecardToResourceFormat(scorecard *tools.GoScorecardResult) map[string]interface{} {
	// Use the existing goScorecardToMap function which already calculates component scores
	scorecardMap := tools.GoScorecardToMap(scorecard)

	// Extract blockers (using helper from scorecard_mlx.go)
	blockers := tools.ExtractBlockers(scorecard)

	// Determine production readiness (score >= 70 and no critical blockers)
	productionReady := scorecard.Score >= 70.0 && len(blockers) == 0

	// Build result matching Python format
	result := map[string]interface{}{
		"overall_score":    scorecard.Score,
		"production_ready": productionReady,
		"scores":           scorecardMap["scores"],
		"blockers":         blockers,
		"recommendations":  scorecard.Recommendations,
		"metrics":          scorecardMap["metrics"],
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	return result
}
