//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// HistoricalTask and EstimationResult types are defined in estimation_shared.go

// estimateWithAppleFM estimates task duration using the default FM provider for semantic analysis
func estimateWithAppleFM(ctx context.Context, name, details string, tags []string, priority string) (*EstimationResult, error) {
	if !FMAvailable() {
		return nil, ErrFMNotSupported
	}

	// Build analysis prompt
	tagsStr := "none"
	if len(tags) > 0 {
		tagsStr = strings.Join(tags, ", ")
	}

	analysisPrompt := fmt.Sprintf(`Estimate the time required to complete this software development task.

TASK INFORMATION:
- Name: %s
- Details: %s
- Tags: %s
- Priority: %s

ESTIMATION CRITERIA:
1. Technical Complexity: Evaluate the technical difficulty (simple code change, moderate feature, complex architecture)
2. Scope: Assess the amount of work (typo fix, single feature, major refactor, new system)
3. Research Needs: Determine if research is required (none, some documentation, extensive investigation)
4. Testing Requirements: Consider testing needs (no tests, unit tests, integration tests, comprehensive testing)
5. Integration Complexity: Evaluate system integration (standalone, single system, multiple systems, external APIs)

ESTIMATION GUIDELINES:
- Simple bug fixes or typos: 0.5-2 hours
- Small features or improvements: 2-8 hours
- Moderate features with integration: 8-16 hours
- Complex features or refactors: 16-40 hours
- Major architectural changes: 40+ hours (cap at 20 for this tool)

RESPONSE FORMAT (JSON only, no other text):
{
    "estimate_hours": <number between 0.5 and 20>,
    "confidence": <number between 0.0 and 1.0>,
    "complexity": <number between 1 and 10>,
    "reasoning": "<brief 1-2 sentence explanation of your estimate>"
}

EXAMPLE:
{
    "estimate_hours": 3.5,
    "confidence": 0.75,
    "complexity": 6,
    "reasoning": "Moderate complexity feature requiring API integration and testing"
}`, name, details, tagsStr, priority)

	// Prepend system instructions so the model behaves as a task estimator
	systemInstructions := `You are an expert software development task estimator.
Your role is to analyze software development tasks and provide accurate time estimates.
You understand software complexity, development workflows, testing requirements, and integration challenges.
Always provide realistic estimates based on the actual scope of work described.
Respond ONLY with valid JSON in the exact format requested.

`
	fullPrompt := systemInstructions + analysisPrompt

	// Lower temperature (0.2) for more deterministic, consistent estimates
	resp, err := DefaultFM.Generate(ctx, fullPrompt, 200, 0.2)
	if err != nil {
		return nil, fmt.Errorf("FM generate failed: %w", err)
	}

	result, err := parseAppleFMResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse FM response: %w", err)
	}

	return result, nil
}

// parseAppleFMResponse extracts JSON from Apple FM response
func parseAppleFMResponse(text string) (*EstimationResult, error) {
	// Try to extract JSON object from response
	// Look for JSON object pattern
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := text[startIdx : endIdx+1]

	// Parse JSON
	var parsed struct {
		EstimateHours float64 `json:"estimate_hours"`
		Confidence    float64 `json:"confidence"`
		Complexity    float64 `json:"complexity"`
		Reasoning     string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Validate ranges
	if parsed.EstimateHours < 0.5 || parsed.EstimateHours > 20 {
		parsed.EstimateHours = math.Max(0.5, math.Min(20, parsed.EstimateHours))
	}
	if parsed.Confidence < 0 || parsed.Confidence > 1 {
		parsed.Confidence = math.Max(0, math.Min(1, parsed.Confidence))
	}

	// Calculate confidence interval
	stdDev := parsed.EstimateHours * 0.3 // 30% coefficient of variation
	lowerBound := math.Max(0.5, parsed.EstimateHours-1.96*stdDev)
	upperBound := parsed.EstimateHours + 1.96*stdDev

	return &EstimationResult{
		EstimateHours: math.Round(parsed.EstimateHours*10) / 10,
		Confidence:    math.Min(0.95, parsed.Confidence),
		Method:        "apple_foundation_models",
		LowerBound:    math.Round(lowerBound*10) / 10,
		UpperBound:    math.Round(upperBound*10) / 10,
		Metadata: map[string]interface{}{
			"complexity": parsed.Complexity,
			"reasoning":  parsed.Reasoning,
			"model":      "fm_provider",
		},
	}, nil
}

// Shared functions (estimateStatistically, loadHistoricalTasks, estimateFromHistory,
// estimateFromKeywords, getPriorityMultiplier, handleEstimationAnalyze, handleEstimationStats)
// are now in estimation_shared.go to support both CGO and non-CGO builds

// handleEstimationNative handles estimation using native Go with Apple Foundation Models
func handleEstimationNative(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	action := "estimate"
	if actionStr, ok := params["action"].(string); ok && actionStr != "" {
		action = actionStr
	}

	switch action {
	case "estimate":
		return handleEstimationEstimate(ctx, projectRoot, params)
	case "analyze":
		return handleEstimationAnalyze(projectRoot, params)
	case "stats":
		return handleEstimationStats(projectRoot, params)
	default:
		return "", fmt.Errorf("unknown action: %s (supported: estimate, analyze, stats)", action)
	}
}

// handleEstimationEstimate handles the estimate action
func handleEstimationEstimate(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
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

	useAppleFM := true
	if useAFM, ok := params["use_apple_fm"].(bool); ok {
		useAppleFM = useAFM
	}

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

	// Get statistical estimate
	statisticalResult, err := estimateStatistically(projectRoot, name, details, tags, priority, useHistorical)
	if err != nil {
		return "", fmt.Errorf("statistical estimation failed: %w", err)
	}

	// Get Apple FM estimate if enabled
	var appleFMResult *EstimationResult
	if useAppleFM {
		afmResult, err := estimateWithAppleFM(ctx, name, details, tags, priority)
		if err == nil {
			appleFMResult = afmResult
		} else {
			// Log error but continue with statistical only
			// Error is logged to metadata for debugging
			if statisticalResult.Metadata == nil {
				statisticalResult.Metadata = make(map[string]interface{})
			}
			statisticalResult.Metadata["apple_fm_error"] = err.Error()
		}
	}

	// Combine estimates
	var finalResult *EstimationResult
	if appleFMResult != nil && appleFMWeight > 0 {
		// Hybrid: combine statistical and Apple FM
		statisticalWeight := 1.0 - appleFMWeight
		combinedEstimate := statisticalResult.EstimateHours*statisticalWeight + appleFMResult.EstimateHours*appleFMWeight
		combinedConfidence := statisticalResult.Confidence*statisticalWeight + appleFMResult.Confidence*appleFMWeight

		finalResult = &EstimationResult{
			EstimateHours: math.Round(combinedEstimate*10) / 10,
			Confidence:    math.Min(0.95, combinedConfidence),
			Method:        "hybrid_apple_fm",
			LowerBound:    statisticalResult.LowerBound,
			UpperBound:    statisticalResult.UpperBound,
			Metadata: map[string]interface{}{
				"statistical_estimate": statisticalResult.EstimateHours,
				"apple_fm_estimate":    appleFMResult.EstimateHours,
				"statistical_weight":   statisticalWeight,
				"apple_fm_weight":      appleFMWeight,
				"statistical_method":   statisticalResult.Method,
				"apple_fm_metadata":    appleFMResult.Metadata,
			},
		}
	} else {
		// Statistical only
		finalResult = statisticalResult
	}

	// Marshal result
	resultJSON, err := json.MarshalIndent(finalResult, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(resultJSON), nil
}

// handleEstimationAnalyze and handleEstimationStats are in estimation_shared.go
