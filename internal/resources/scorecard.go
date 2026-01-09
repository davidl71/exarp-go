package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/tools"
)

// handleScorecard handles the scorecard resource
// Uses native Go implementation for Go projects, falls back to Python bridge for non-Go projects
func handleScorecard(ctx context.Context, uri string) ([]byte, string, error) {
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		return nil, "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Check if Go project
	if tools.IsGoProject() {
		// Use native Go scorecard implementation
		opts := &tools.ScorecardOptions{FastMode: true}
		scorecard, err := tools.GenerateGoScorecard(ctx, projectRoot, opts)
		if err == nil {
			// Convert to resource JSON format matching Python implementation
			result := convertScorecardToResourceFormat(scorecard)
			jsonData, err := json.Marshal(result)
			if err != nil {
				return nil, "", fmt.Errorf("failed to marshal scorecard: %w", err)
			}
			return jsonData, "application/json", nil
		}
		// If native fails, fall through to Python bridge
	}

	// Fallback to Python bridge for non-Go projects or if native fails
	return bridge.ExecutePythonResource(ctx, uri)
}

// convertScorecardToResourceFormat converts GoScorecardResult to resource JSON format
// Matches Python generate_project_scorecard() output format
// Reuses helper functions from internal/tools/scorecard_mlx.go
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
