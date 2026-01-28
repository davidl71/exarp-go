package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// TOKENS_PER_CHAR is the estimated tokens per character (rough approximation)
const TOKENS_PER_CHAR = 0.25

// ItemAnalysis represents analysis of a single item in context budget
type ItemAnalysis struct {
	Index           int     `json:"index"`
	Tokens          int     `json:"tokens"`
	PercentOfBudget float64 `json:"percent_of_budget"`
	Recommendation  string  `json:"recommendation"`
}

// handleContextBudget handles the context_budget tool
// Estimates token usage and suggests context reduction strategy
// Uses protobuf parsing for type-safe argument handling
func handleContextBudget(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Try protobuf first, fall back to JSON for backward compatibility
	req, params, err := ParseContextRequest(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert protobuf request to params map if needed
	if req != nil {
		params = ContextRequestToParams(req)
		if req.BudgetTokens == 0 {
			params["budget_tokens"] = config.DefaultContextBudget()
		}
	}

	// Get items (required)
	itemsRaw, ok := params["items"]
	if !ok {
		return nil, fmt.Errorf("items parameter is required")
	}

	// Parse items using protobuf helper (simplifies type conversion)
	contextItems, err := ParseContextItems(itemsRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	// Get budget_tokens (optional, use config default)
	budgetTokens := config.DefaultContextBudget()
	if budgetRaw, ok := params["budget_tokens"]; ok {
		if budgetFloat, ok := budgetRaw.(float64); ok {
			budgetTokens = int(budgetFloat)
		} else if budgetInt, ok := budgetRaw.(int); ok {
			budgetTokens = budgetInt
		}
	}

	// Analyze items using protobuf ContextItem (type-safe)
	analysis := make([]ItemAnalysis, 0, len(contextItems))
	totalTokens := 0

	for i, contextItem := range contextItems {
		// Extract data string from ContextItem (type-safe, no marshaling needed)
		itemStr := ContextItemToDataString(contextItem)

		// Estimate tokens using config ratio
		tokens := estimateTokens(itemStr, config.TokensPerChar())
		totalTokens += tokens

		// Calculate percentage of budget
		percentOfBudget := float64(tokens) / float64(budgetTokens) * 100

		// Get recommendation
		recommendation := getBudgetRecommendation(tokens, budgetTokens)

		analysis = append(analysis, ItemAnalysis{
			Index:           i,
			Tokens:          tokens,
			PercentOfBudget: percentOfBudget,
			Recommendation:  recommendation,
		})
	}

	// Sort by tokens (descending)
	for i := 0; i < len(analysis)-1; i++ {
		for j := i + 1; j < len(analysis); j++ {
			if analysis[i].Tokens < analysis[j].Tokens {
				analysis[i], analysis[j] = analysis[j], analysis[i]
			}
		}
	}

	// Build result
	overBudget := totalTokens > budgetTokens
	reductionNeeded := 0
	if overBudget {
		reductionNeeded = totalTokens - budgetTokens
	}

	strategy := suggestReductionStrategy(analysis, totalTokens, budgetTokens)

	result := map[string]interface{}{
		"total_tokens":     totalTokens,
		"budget_tokens":    budgetTokens,
		"over_budget":      overBudget,
		"reduction_needed": reductionNeeded,
		"items":            analysis,
		"strategy":         strategy,
	}

	return response.FormatResult(result, "")
}

// estimateTokens estimates token count for text using provided ratio
func estimateTokens(text string, tokensPerChar float64) int {
	return int(float64(len(text)) * tokensPerChar)
}

// getBudgetRecommendation gets recommendation for a single item
func getBudgetRecommendation(tokens, budget int) string {
	ratio := float64(tokens) / float64(budget)
	if ratio > 0.5 {
		return "summarize_brief"
	} else if ratio > 0.25 {
		return "summarize_key_metrics"
	} else if ratio > 0.1 {
		return "keep_detailed"
	}
	return "keep_full"
}

// suggestReductionStrategy suggests overall reduction strategy
func suggestReductionStrategy(analysis []ItemAnalysis, total, budget int) string {
	if total <= budget {
		return "Within budget - no reduction needed"
	}

	reductionNeeded := total - budget
	toSummarize := 0
	estimatedSavings := 0

	for _, a := range analysis {
		if a.Recommendation == "summarize_brief" || a.Recommendation == "summarize_key_metrics" {
			toSummarize++
			estimatedSavings += int(float64(a.Tokens) * 0.7)
		}
	}

	if toSummarize == 0 {
		return fmt.Sprintf("Reduce largest items to fit. Need to remove ~%d tokens.", reductionNeeded)
	}

	return fmt.Sprintf("Summarize %d items using 'brief' level. Estimated savings: %d tokens.", toSummarize, estimatedSavings)
}

// handleContextBatchNative handles the context batch action using native Go
// Summarizes multiple items and optionally combines them
// Uses protobuf ContextItem for type-safe item processing
func handleContextBatchNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get items (required)
	itemsRaw, ok := params["items"]
	if !ok || itemsRaw == nil {
		return nil, fmt.Errorf("items parameter is required for batch action")
	}

	// Parse items using protobuf helper (simplifies type conversion)
	contextItems, err := ParseContextItems(itemsRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	// Get optional parameters
	level := "brief"
	if levelRaw, ok := params["level"].(string); ok && levelRaw != "" {
		level = levelRaw
	}

	combine := true
	if combineRaw, ok := params["combine"].(bool); ok {
		combine = combineRaw
	}

	// Summarize each item using protobuf ContextItem (type-safe)
	summaries := make([]map[string]interface{}, 0, len(contextItems))
	totalOriginal := 0
	totalSummarized := 0

	for _, contextItem := range contextItems {
		// Extract data string from ContextItem (type-safe, no assertions needed)
		dataStr := ContextItemToDataString(contextItem)
		toolType := contextItem.ToolType

		// Create params for summarize
		summarizeParams := map[string]interface{}{
			"data":      dataStr,
			"level":     level,
			"tool_type": toolType,
		}

		// Try to use native summarize (with default FM if available)
		var summaryResult map[string]interface{}

		if DefaultFM != nil && DefaultFM.Supported() {
			// Try native Go with Apple FM
			result, summarizeErr := handleContextSummarizeNative(ctx, summarizeParams)
			if summarizeErr == nil && len(result) > 0 {
				// Parse the result JSON
				if parseErr := json.Unmarshal([]byte(result[0].Text), &summaryResult); parseErr == nil {
					// Successfully got native summary
				} else {
					// Failed to parse, fall back to simple summarization
					summaryResult = createSimpleSummary(dataStr, level, toolType)
				}
			} else {
				// Native failed, use simple summarization
				summaryResult = createSimpleSummary(dataStr, level, toolType)
			}
		} else {
			// FM not available, use simple summarization
			summaryResult = createSimpleSummary(dataStr, level, toolType)
		}

		summaries = append(summaries, summaryResult)

		// Accumulate token estimates
		if tokenEst, ok := summaryResult["token_estimate"].(map[string]interface{}); ok {
			if orig, ok := tokenEst["original"].(float64); ok {
				totalOriginal += int(orig)
			} else if orig, ok := tokenEst["original"].(int); ok {
				totalOriginal += orig
			}
			if summ, ok := tokenEst["summarized"].(float64); ok {
				totalSummarized += int(summ)
			} else if summ, ok := tokenEst["summarized"].(int); ok {
				totalSummarized += summ
			}
		}
	}

	// Build result
	var result map[string]interface{}
	if combine {
		// Extract summaries for combined view
		combinedSummaries := make([]interface{}, 0, len(summaries))
		for _, s := range summaries {
			if summary, ok := s["summary"]; ok {
				combinedSummaries = append(combinedSummaries, summary)
			}
		}

		reduction := 0.0
		if totalOriginal > 0 {
			reduction = (1.0 - float64(totalSummarized)/float64(totalOriginal)) * 100.0
		}

		result = map[string]interface{}{
			"combined_summary": combinedSummaries,
			"total_items":      len(summaries),
			"token_estimate": map[string]interface{}{
				"original":          totalOriginal,
				"summarized":        totalSummarized,
				"reduction_percent": fmt.Sprintf("%.1f", reduction),
			},
		}
	} else {
		result = map[string]interface{}{
			"summaries": summaries,
		}
	}

	return response.FormatResult(result, "")
}

// createSimpleSummary creates a simple summary without Apple FM
// This is a fallback when Apple FM is not available
func createSimpleSummary(dataStr string, level string, toolType string) map[string]interface{} {
	originalTokens := estimateTokens(dataStr, config.TokensPerChar())

	// Create a simple summary based on level
	var summary interface{}
	switch level {
	case "brief":
		// Extract first 200 chars as brief summary
		if len(dataStr) > 200 {
			summary = dataStr[:200] + "..."
		} else {
			summary = dataStr
		}
	case "key_metrics":
		// Try to extract numbers from JSON
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err == nil {
			metrics := make(map[string]interface{})
			extractNumbers(data, metrics)
			summary = metrics
		} else {
			summary = "Key metrics extraction failed"
		}
	case "actionable":
		// Try to extract actionable items
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err == nil {
			actions := make(map[string]interface{})
			if recs, ok := data["recommendations"].([]interface{}); ok {
				actions["recommendations"] = recs
			}
			if tasks, ok := data["tasks"].([]interface{}); ok {
				actions["tasks"] = tasks
			}
			summary = actions
		} else {
			summary = "Actionable items extraction failed"
		}
	default:
		// Detailed: return first 500 chars
		if len(dataStr) > 500 {
			summary = dataStr[:500] + "..."
		} else {
			summary = dataStr
		}
	}

	summaryStr := ""
	if s, ok := summary.(string); ok {
		summaryStr = s
	} else {
		bytes, _ := json.Marshal(summary)
		summaryStr = string(bytes)
	}

	summaryTokens := estimateTokens(summaryStr, config.TokensPerChar())
	reduction := 0.0
	if originalTokens > 0 {
		reduction = (1.0 - float64(summaryTokens)/float64(originalTokens)) * 100.0
	}

	return map[string]interface{}{
		"summary": summary,
		"level":   level,
		"method":  "simple",
		"token_estimate": map[string]interface{}{
			"original":          originalTokens,
			"summarized":        summaryTokens,
			"reduction_percent": fmt.Sprintf("%.1f", reduction),
		},
	}
}

// extractNumbers recursively extracts numeric values from a map
func extractNumbers(data interface{}, result map[string]interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			switch val := val.(type) {
			case float64, int:
				result[k] = val
			case map[string]interface{}:
				extractNumbers(val, result)
			case []interface{}:
				if len(val) > 0 {
					// Check if it's a numeric array
					if _, ok := val[0].(float64); ok || len(val) < 10 {
						result[k+"_count"] = len(val)
					}
				}
			}
		}
	case []interface{}:
		for i, val := range v {
			if i < 10 { // Limit depth
				extractNumbers(val, result)
			}
		}
	}
}
