// estimation_shared_v2.go — Shared estimation handling logic for both CGO and no-CGO builds.
// Uses runtime FMAvailable() for feature detection.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// EstimateWithFM estimates task duration using Apple FM if available.
func EstimateWithFM(ctx context.Context, name, details string, tags []string, priority string) (*EstimationResult, error) {
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
	resp, err := DefaultFMProvider().Generate(ctx, fullPrompt, 200, 0.2)
	if err != nil {
		return nil, fmt.Errorf("FM generate failed: %w", err)
	}

	// Parse the response
	var parsed struct {
		EstimateHours float64 `json:"estimate_hours"`
		Confidence    float64 `json:"confidence"`
		Complexity    int     `json:"complexity"`
		Reasoning     string  `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(resp), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse FM response: %w", err)
	}

	// Validate and cap the estimate
	estimate := parsed.EstimateHours
	if estimate < 0.5 {
		estimate = 0.5
	}
	if estimate > 20 {
		estimate = 20
	}

	confidence := parsed.Confidence
	if confidence < 0.1 {
		confidence = 0.1
	}
	if confidence > 1.0 {
		confidence = 1.0
	}

	return &EstimationResult{
		EstimateHours: math.Round(estimate*10) / 10,
		Confidence:    confidence,
		Method:        "apple_fm",
		LowerBound:    estimate * 0.5,
		UpperBound:    estimate * 1.5,
		Metadata: map[string]interface{}{
			"complexity": parsed.Complexity,
			"reasoning":  parsed.Reasoning,
		},
	}, nil
}

// HandleEstimationNative is the shared dispatcher for estimation.
func HandleEstimationNative(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	action := "estimate"
	if actionStr, ok := params["action"].(string); ok && actionStr != "" {
		action = actionStr
	}

	switch action {
	case "estimate":
		return handleEstimationEstimateShared(ctx, projectRoot, params)
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

// handleEstimationEstimateShared is the shared estimation handler using runtime feature detection.
func handleEstimationEstimateShared(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	name, _ := params["name"].(string)
	details, _ := params["details"].(string)
	tagsRaw := params["tags"]
	tagListRaw := params["tag_list"]

	priority, _ := params["priority"].(string)
	if priority == "" {
		priority = "medium"
	}

	useHistorical := ParamBool(params, "use_historical", true)

	// local_ai_backend: fm (Apple), ollama, or empty (auto). Legacy "mlx" is ignored (same as auto).
	backend, _ := params["local_ai_backend"].(string)
	backend = strings.TrimSpace(strings.ToLower(backend))
	if backend == "mlx" {
		backend = ""
	}

	// Determine if we should use Apple FM (only if built with CGO)
	useAppleFM := FMAvailable() && (backend == "" || backend == "fm")
	if useAFM := ParamBool(params, "use_apple_fm", true); backend == "" {
		useAppleFM = useAFM && FMAvailable()
	}

	appleFMWeight := ParamFloat64(params, "apple_fm_weight", 0.3)

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

	// Get statistical estimate (always available)
	statisticalResult, err := estimateStatistically(ctx, projectRoot, name, details, tags, priority, useHistorical)
	if err != nil {
		return "", fmt.Errorf("statistical estimation failed: %w", err)
	}

	// Try LLM estimate based on backend preference
	var llmResult *EstimationResult
	llmMethod := ""

	switch backend {
	case "ollama":
		res, err := EstimateWithOllama(ctx, name, details, tags, priority, "")
		if err == nil {
			llmResult = res
			llmMethod = "ollama"
		} else {
			if statisticalResult.Metadata == nil {
				statisticalResult.Metadata = make(map[string]interface{})
			}
			statisticalResult.Metadata["ollama_error"] = err.Error()
		}
	default:
		// Try Apple FM if available and not explicitly disabled
		if useAppleFM {
			afmResult, err := EstimateWithFM(ctx, name, details, tags, priority)
			if err == nil {
				llmResult = afmResult
				llmMethod = "apple_fm"
			} else {
				if statisticalResult.Metadata == nil {
					statisticalResult.Metadata = make(map[string]interface{})
				}
				statisticalResult.Metadata["apple_fm_error"] = err.Error()
			}
		}
	}

	// Combine results
	var finalResult *EstimationResult
	if llmResult != nil && appleFMWeight > 0 {
		statisticalWeight := 1.0 - appleFMWeight
		combinedEstimate := statisticalResult.EstimateHours*statisticalWeight + llmResult.EstimateHours*appleFMWeight
		combinedConfidence := statisticalResult.Confidence*statisticalWeight + llmResult.Confidence*appleFMWeight

		finalResult = &EstimationResult{
			EstimateHours: math.Round(combinedEstimate*10) / 10,
			Confidence:    math.Min(0.95, combinedConfidence),
			Method:        "hybrid_" + llmMethod,
			LowerBound:    statisticalResult.LowerBound,
			UpperBound:    llmResult.UpperBound,
			Metadata: map[string]interface{}{
				"statistical_estimate": statisticalResult.EstimateHours,
				"llm_estimate":         llmResult.EstimateHours,
				"llm_backend":          llmMethod,
				"statistical_weight":   statisticalWeight,
				"llm_weight":           appleFMWeight,
			},
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

// handleEstimationNative delegates to HandleEstimationNative (single symbol for all builds; replaces build-tagged shims).
func handleEstimationNative(ctx context.Context, projectRoot string, params map[string]interface{}) (string, error) {
	return HandleEstimationNative(ctx, projectRoot, params)
}
