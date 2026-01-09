//go:build !(darwin && arm64 && cgo)
// +build !darwin !arm64 !cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
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
	default:
		return "", fmt.Errorf("unknown action: %s (supported: estimate, analyze, stats)", action)
	}
}

// handleEstimationEstimateNoCGO handles estimation without Apple FM (statistical only)
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

	// Get statistical estimate only (no Apple FM on non-Apple platforms)
	statisticalResult, err := estimateStatistically(projectRoot, name, details, tags, priority, useHistorical)
	if err != nil {
		return "", fmt.Errorf("statistical estimation failed: %w", err)
	}

	// Marshal result
	resultJSON, err := json.MarshalIndent(statisticalResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}
