//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	fm "github.com/blacktop/go-foundationmodels"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/platform"
)

// summarizeWithAppleFM summarizes text using Apple Foundation Models
func summarizeWithAppleFM(ctx context.Context, data string, level string, maxTokens int) (string, error) {
	// Check platform support
	support := platform.CheckAppleFoundationModelsSupport()
	if !support.Supported {
		return "", fmt.Errorf("Apple Foundation Models not supported: %s", support.Reason)
	}

	// Create session
	sess := fm.NewSession()
	defer sess.Release()

	// Build prompt based on level
	var prompt string
	switch level {
	case "brief":
		prompt = fmt.Sprintf("Summarize the following text in one concise sentence with key metrics:\n\n%s", data)
	case "detailed":
		prompt = fmt.Sprintf("Summarize the following text in multiple paragraphs with categories:\n\n%s", data)
	case "key_metrics":
		prompt = fmt.Sprintf("Extract only the numerical metrics and key numbers from the following text. Return as JSON:\n\n%s", data)
	case "actionable":
		prompt = fmt.Sprintf("Extract only actionable items (recommendations, tasks, fixes) from the following text. Return as JSON:\n\n%s", data)
	default:
		prompt = fmt.Sprintf("Summarize the following text concisely:\n\n%s", data)
	}

	// Set temperature based on level (lower for more deterministic outputs)
	temperature := float32(0.3) // Good for summarization
	if level == "key_metrics" || level == "actionable" {
		temperature = 0.2 // Very deterministic for structured extraction
	}

	// Set max tokens if specified, otherwise use default
	if maxTokens <= 0 {
		maxTokens = 512 // Default for summaries
	}

	// Generate summary
	summary := sess.RespondWithOptions(prompt, maxTokens, temperature)

	return summary, nil
}

// handleContextSummarizeNative handles context summarization using native Go with Apple FM
func handleContextSummarizeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	// Get required parameters
	dataRaw, ok := params["data"]
	if !ok || dataRaw == nil {
		return nil, fmt.Errorf("data parameter is required for summarize action")
	}

	// Convert data to string
	var dataStr string
	switch v := dataRaw.(type) {
	case string:
		dataStr = v
	case map[string]interface{}, []interface{}:
		// Convert to JSON string
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
		dataStr = string(bytes)
	default:
		dataStr = fmt.Sprintf("%v", v)
	}

	// Get optional parameters
	level := "brief"
	if levelRaw, ok := params["level"].(string); ok && levelRaw != "" {
		level = levelRaw
	}

	maxTokens := 0
	if maxTokensRaw, ok := params["max_tokens"]; ok {
		switch v := maxTokensRaw.(type) {
		case int:
			maxTokens = v
		case float64:
			maxTokens = int(v)
		}
	}

	includeRaw := false
	if includeRawRaw, ok := params["include_raw"].(bool); ok {
		includeRaw = includeRawRaw
	}

	toolType := ""
	if toolTypeRaw, ok := params["tool_type"].(string); ok {
		toolType = toolTypeRaw
	}

	// Try Apple Foundation Models first
	summary, err := summarizeWithAppleFM(ctx, dataStr, level, maxTokens)
	if err != nil {
		// Fallback to Python bridge if Apple FM not available
		return nil, fmt.Errorf("Apple FM summarization failed, will fallback: %w", err)
	}

	// Estimate tokens
	originalTokens := estimateTokens(dataStr)
	summaryTokens := estimateTokens(summary)
	reduction := 0.0
	if originalTokens > 0 {
		reduction = (1.0 - float64(summaryTokens)/float64(originalTokens)) * 100.0
	}

	// Build result
	result := map[string]interface{}{
		"summary": summary,
		"level":   level,
		"method":  "apple_foundation_models",
		"token_estimate": map[string]interface{}{
			"original":        originalTokens,
			"summarized":      summaryTokens,
			"reduction_percent": fmt.Sprintf("%.1f", reduction),
		},
		"duration_ms": time.Since(startTime).Milliseconds(),
	}

	if toolType != "" {
		result["tool_type"] = toolType
	}

	if includeRaw {
		// Parse original data back
		var rawData interface{}
		if err := json.Unmarshal([]byte(dataStr), &rawData); err != nil {
			rawData = dataStr
		}
		result["raw_data"] = rawData
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

