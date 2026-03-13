// context_shared.go — Shared context handling logic for both CGO and no-CGO builds.
// Uses runtime FMAvailable() check for Apple FM features.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
)

// SummarizeWithFM summarizes text using Apple FM if available, otherwise returns error.
// This is the runtime-checked version that can be used by both CGO and no-CGO builds.
func SummarizeWithFM(ctx context.Context, data string, level string, maxTokens int) (string, error) {
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

// HandleContextSummarizeShared is the shared summarization handler used by both CGO and no-CGO builds.
// It checks FMAvailable() at runtime and returns an appropriate error if Apple FM is not available.
func HandleContextSummarizeShared(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	startTime := time.Now()

	// Check if Apple FM is available
	if !FMAvailable() {
		return nil, fmt.Errorf("Apple Foundation Models not available on this platform or build (darwin/arm64 with CGO required)")
	}

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

	includeRaw := ParamBool(params, "include_raw", false)
	toolType := ""
	if toolTypeRaw, ok := params["tool_type"].(string); ok {
		toolType = toolTypeRaw
	}

	// Use Apple FM for summarization
	summary, err := SummarizeWithFM(ctx, dataStr, level, maxTokens)
	if err != nil {
		return nil, fmt.Errorf("FM summarization failed: %w", err)
	}

	// Estimate tokens
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
		var rawData interface{}
		if err := json.Unmarshal([]byte(dataStr), &rawData); err != nil {
			rawData = dataStr
		}
		result["raw_data"] = rawData
	}

	return framework.FormatResult(result, "")
}
