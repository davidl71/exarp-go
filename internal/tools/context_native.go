//go:build darwin && arm64 && cgo
// +build darwin,arm64,cgo

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
)

// summarizeWithAppleFM summarizes text using the default FM provider (e.g. Apple Foundation Models when available).
func summarizeWithAppleFM(ctx context.Context, data string, level string, maxTokens int) (string, error) {
	if !FMAvailable() {
		return "", ErrFMNotSupported
	}

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

	return DefaultFMProvider().Generate(ctx, prompt, maxTokens, temperature)
}

// handleContextSummarizeNative handles context summarization using native Go with Apple FM
// Uses protobuf ContextRequest for type-safe parameter handling
func handleContextSummarizeNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	// Get required parameters (data is already a string from protobuf conversion)
	dataRaw, ok := params["data"]
	if !ok || dataRaw == nil {
		return nil, fmt.Errorf("data parameter is required for summarize action")
	}

	// Convert data to string (simplified - protobuf ensures it's already a string)
	var dataStr string
	switch v := dataRaw.(type) {
	case string:
		dataStr = v
	case map[string]interface{}, []interface{}:
		// Convert to JSON string (fallback for JSON format)
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

	// Try default FM provider (e.g. Apple FM when available)
	summary, err := summarizeWithAppleFM(ctx, dataStr, level, maxTokens)
	if err != nil {
		// Fallback to Python bridge if FM not available
		return nil, fmt.Errorf("FM summarization failed, will fallback: %w", err)
	}

	// Estimate tokens (using default tokensPerChar of 0.25)
	originalTokens := estimateTokens(dataStr, 0.25)
	summaryTokens := estimateTokens(summary, 0.25)
	reduction := 0.0
	if originalTokens > 0 {
		reduction = (1.0 - float64(summaryTokens)/float64(originalTokens)) * 100.0
	}

	// Build result
	result := map[string]interface{}{
		"summary": summary,
		"level":   level,
		"method":  "fm_provider",
		"token_estimate": map[string]interface{}{
			"original":          originalTokens,
			"summarized":        summaryTokens,
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

	return response.FormatResult(result, "")
}
