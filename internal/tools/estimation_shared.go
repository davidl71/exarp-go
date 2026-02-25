// estimation_shared.go — Estimation: consts, GetPreferredBackend, HistoricalTask type, estimateStatistically.
package tools

import (
	"math"
	"strings"
	"github.com/spf13/cast"
)

// MetadataKeyPreferredBackend is the task metadata key for local AI backend preference (fm, mlx, ollama).
const MetadataKeyPreferredBackend = "preferred_backend"

// MetadataKeyRecommendedTools is the task metadata key for recommended MCP tools (e.g. tractatus_thinking, context7).
const MetadataKeyRecommendedTools = "recommended_tools"

// GetRecommendedTools returns the recommended tool IDs from task metadata, or nil if unset.
func GetRecommendedTools(metadata map[string]interface{}) []string {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata[MetadataKeyRecommendedTools]
	if !ok || raw == nil {
		return nil
	}
	slice := cast.ToStringSlice(raw)
	if len(slice) == 0 {
		return nil
	}
	out := make([]string, 0, len(slice))
	for _, s := range slice {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// GetPreferredBackend returns the preferred local AI backend from task metadata, or "" if unset.
// Valid values: "fm", "mlx", "ollama".
func GetPreferredBackend(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	s := strings.TrimSpace(strings.ToLower(cast.ToString(metadata[MetadataKeyPreferredBackend])))
	switch s {
	case "fm", "mlx", "ollama":
		return s
	default:
		return ""
	}
}

// HistoricalTask represents a completed task for historical analysis.
type HistoricalTask struct {
	Name           string
	Details        string
	Tags           []string
	Priority       string
	EstimatedHours float64
	ActualHours    float64
	Created        string
	CompletedAt    string
}

// EstimationResult represents the result of task duration estimation.
type EstimationResult struct {
	EstimateHours float64                `json:"estimate_hours"`
	Confidence    float64                `json:"confidence"`
	Method        string                 `json:"method"`
	LowerBound    float64                `json:"lower_bound"`
	UpperBound    float64                `json:"upper_bound"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// estimateStatistically estimates task duration using statistical methods.
func estimateStatistically(projectRoot, name, details string, tags []string, priority string, useHistorical bool) (*EstimationResult, error) {
	text := strings.ToLower(name + " " + details)

	var historicalEstimate *float64

	var historicalConfidence float64

	// Strategy 1: Historical data matching
	if useHistorical {
		historical, err := loadHistoricalTasks(projectRoot)
		if err == nil && len(historical) > 0 {
			est, conf := estimateFromHistory(text, tags, priority, historical)
			if est != nil {
				historicalEstimate = est
				historicalConfidence = conf
			}
		}
	}

	// Strategy 2: Keyword-based heuristic
	heuristicEstimate := estimateFromKeywords(text)
	heuristicConfidence := 0.3

	// Strategy 3: Priority-based adjustment
	priorityMultiplier := getPriorityMultiplier(priority)

	// Combine estimates
	var baseEstimate float64

	var confidence float64

	var method string

	if historicalEstimate != nil && historicalConfidence > 0.2 {
		// Use historical estimate (already includes actual work time)
		baseEstimate = *historicalEstimate
		confidence = historicalConfidence
		method = "historical_match"
	} else {
		// Use keyword heuristic
		baseEstimate = heuristicEstimate
		confidence = heuristicConfidence
		method = "keyword_heuristic"
	}

	// Apply priority multiplier once (not twice!)
	finalEstimate := baseEstimate * priorityMultiplier

	// Cap estimates at reasonable maximum (20 hours for this tool)
	// Most tasks should be much shorter
	finalEstimate = math.Min(20.0, finalEstimate)

	// Round to 1 decimal place
	finalEstimate = math.Round(finalEstimate*10) / 10

	// Calculate confidence interval
	stdDev := finalEstimate * 0.3
	lowerBound := math.Max(0.5, finalEstimate-1.96*stdDev)
	upperBound := finalEstimate + 1.96*stdDev

	return &EstimationResult{
		EstimateHours: finalEstimate,
		Confidence:    math.Min(0.95, confidence),
		Method:        method,
		LowerBound:    math.Round(lowerBound*10) / 10,
		UpperBound:    math.Round(upperBound*10) / 10,
		Metadata: map[string]interface{}{
			"historical_match":      historicalEstimate != nil,
			"historical_confidence": historicalConfidence,
			"heuristic_estimate":    heuristicEstimate,
			"priority_multiplier":   priorityMultiplier,
		},
	}, nil
}

