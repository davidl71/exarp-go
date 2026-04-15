//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// handleEstimationNative handles estimation using native Go (non-Apple platforms)
// Uses statistical estimation only (no Apple Foundation Models)
func handleEstimationNative(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	action := "estimate"
	if actionStr, ok := params["action"].(string); ok && actionStr != "" {
		action = actionStr
	}

	switch action {
	case "estimate":
		return handleEstimationEstimateNoCGO(ctx, projectRoot, params)
	case "analyze":
		return handleEstimationAnalyze(projectRoot, params)
	case "stats":
		return handleEstimationStats(projectRoot, params)
	case "estimate_batch":
		return handleEstimationBatch(projectRoot, params)
	default:
		return "", fmt.Errorf("unknown action: %s (supported: estimate, analyze, stats, estimate_batch)", action)
	}
}

// handleEstimationEstimateNoCGO handles estimation without Apple FM (statistical + optional Ollama)
func handleEstimationEstimateNoCGO(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	name, _ := params["name"].(string)
	details, _ := params["details"].(string)
	tagsRaw, _ := params["tags"]
	tagListRaw, _ := params["tag_list"]
	priority, _ := params["priority"].(string)
	if priority == "" {
		priority = "medium"
	}

	useHistorical := true
	if useHist, ok := params["use_historical"].(bool); ok {
		useHistorical = useHist
	}

	backend, _ := params["local_ai_backend"].(string)
	backend = strings.TrimSpace(strings.ToLower(backend))

	appleFMWeight := 0.3
	if weight, ok := params["apple_fm_weight"].(float64); ok {
		appleFMWeight = weight
	}

	// Parse tags
	var tags []string
	if tagList, ok := tagListRaw.([]interface{}); ok {
		for _, tag := range tagList {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	} else if tagsStr, ok := tagsRaw.(string); ok {
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}
	}

	statisticalResult, err := estimateStatistically(projectRoot, name, details, tags, priority, useHistorical)
	if err != nil {
		return "", fmt.Errorf("statistical estimation failed: %w", err)
	}

	var finalResult *EstimationResult
	if backend == "ollama" {
		ollamaResult, err := EstimateWithOllama(ctx, name, details, tags, priority, "")
		if err == nil && appleFMWeight > 0 {
			statisticalWeight := 1.0 - appleFMWeight
			combinedEstimate := statisticalResult.EstimateHours*statisticalWeight + ollamaResult.EstimateHours*appleFMWeight
			combinedConfidence := statisticalResult.Confidence*statisticalWeight + ollamaResult.Confidence*appleFMWeight
			finalResult = &EstimationResult{
				EstimateHours: math.Round(combinedEstimate*10) / 10,
				Confidence:    math.Min(0.95, combinedConfidence),
				Method:        "hybrid_ollama",
				LowerBound:    statisticalResult.LowerBound,
				UpperBound:    statisticalResult.UpperBound,
				Metadata: map[string]interface{}{
					"statistical_estimate": statisticalResult.EstimateHours,
					"llm_estimate":         ollamaResult.EstimateHours,
					"llm_backend":          "ollama",
					"statistical_weight":   statisticalWeight,
					"llm_weight":           appleFMWeight,
					"llm_metadata":         ollamaResult.Metadata,
				},
			}
		} else {
			if statisticalResult.Metadata == nil {
				statisticalResult.Metadata = make(map[string]interface{})
			}
			statisticalResult.Metadata["ollama_error"] = err.Error()
			finalResult = statisticalResult
		}
	} else {
		finalResult = statisticalResult
	}

	resultJSON, err := json.MarshalIndent(finalResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(resultJSON), nil
}
