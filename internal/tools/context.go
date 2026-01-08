package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
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
func handleContextBudget(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	// Parse arguments
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Get items (required)
	itemsRaw, ok := params["items"]
	if !ok {
		return nil, fmt.Errorf("items parameter is required")
	}

	// Parse items - can be JSON string or array
	var items []interface{}
	switch v := itemsRaw.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &items); err != nil {
			return nil, fmt.Errorf("failed to parse items JSON: %w", err)
		}
	case []interface{}:
		items = v
	default:
		return nil, fmt.Errorf("items must be a JSON string or array")
	}

	// Get budget_tokens (optional, default: 4000)
	budgetTokens := 4000
	if budgetRaw, ok := params["budget_tokens"]; ok {
		if budgetFloat, ok := budgetRaw.(float64); ok {
			budgetTokens = int(budgetFloat)
		} else if budgetInt, ok := budgetRaw.(int); ok {
			budgetTokens = budgetInt
		}
	}

	// Analyze items
	analysis := make([]ItemAnalysis, 0, len(items))
	totalTokens := 0

	for i, item := range items {
		// Convert item to JSON string for token estimation
		itemBytes, err := json.Marshal(item)
		if err != nil {
			continue
		}
		itemStr := string(itemBytes)

		// Estimate tokens
		tokens := estimateTokens(itemStr)
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

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// estimateTokens estimates token count for text
func estimateTokens(text string) int {
	return int(float64(len(text)) * TOKENS_PER_CHAR)
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
