// Package tools: Report insight provider abstraction for AI-generated report/scorecard insights.
// Implementations: MLX via Python bridge, or DefaultFMProvider() (Apple FM when available).
// Report code uses DefaultReportInsight so it does not depend on the bridge or MLX by name.

package tools

import (
	"context"
	"encoding/json"
	"errors"
)

// ReportInsightProvider generates long-form AI insights for report/scorecard content.
// Implementations: FM chain (Apple → Ollama) first, then MLX via bridge when unavailable.
// ReportInsightProvider implements TextGenerator.
type ReportInsightProvider interface {
	TextGenerator
}

// defaultReportInsight is the shared default; set by init to composite (FM then MLX).
var defaultReportInsight ReportInsightProvider

func init() {
	defaultReportInsight = &compositeReportInsight{}
}

// DefaultReportInsight returns the default report insight provider (tries FM chain then MLX).
func DefaultReportInsight() ReportInsightProvider {
	if defaultReportInsight == nil {
		return &compositeReportInsight{}
	}

	return defaultReportInsight
}

// compositeReportInsight tries FM chain (Apple → Ollama → stub) first, then MLX (bridge) as last resort.
type compositeReportInsight struct{}

func (c *compositeReportInsight) Supported() bool {
	return true // We always "support" by trying FM then MLX
}

func (c *compositeReportInsight) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	// Try FM chain (native: Apple, then Ollama) first
	var fm TextGenerator = DefaultFMProvider()
	if fm != nil && fm.Supported() {
		if text, err := fm.Generate(ctx, prompt, maxTokens, temperature); err == nil && text != "" {
			return text, nil
		}
	}
	// Fall back to MLX via bridge only when FM chain fails or is unavailable
	if text, err := tryMLXReportInsight(ctx, prompt, maxTokens, temperature); err == nil && text != "" {
		return text, nil
	}

	return "", ErrFMNotSupported
}

// tryMLXReportInsight calls the MLX tool via the Python bridge and returns generated text.
func tryMLXReportInsight(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	params := map[string]interface{}{
		"action":      "generate",
		"prompt":      prompt,
		"model":       "mlx-community/Mistral-7B-Instruct-v0.2-4bit",
		"max_tokens":  maxTokens,
		"temperature": float64(temperature),
	}

	result, err := executeMLXViaBridge(ctx, params)
	if err != nil {
		return "", err
	}

	return parseGeneratedTextFromMLXResponse(result)
}

// executeMLXViaBridge invokes the mlx tool via the shared path (native then bridge).
// Uses InvokeMLXTool so report insights and MLX provider use the same path as the mlx tool.
func executeMLXViaBridge(ctx context.Context, params map[string]interface{}) (string, error) {
	return InvokeMLXTool(ctx, params)
}

// parseGeneratedTextFromMLXResponse extracts generated text from the MLX bridge response.
// Returns ("", err) when no generated text is found so the composite can fall back to FM.
func parseGeneratedTextFromMLXResponse(mlxResult string) (string, error) {
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(mlxResult), &resp); err != nil {
		return "", err
	}
	// data.generated_text (format_success_response)
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if text, ok := data["generated_text"].(string); ok && text != "" {
			return text, nil
		}
	}

	if text, ok := resp["generated_text"].(string); ok && text != "" {
		return text, nil
	}

	if result, ok := resp["result"].(string); ok && result != "" {
		return result, nil
	}

	return "", errNoGeneratedText
}

// errNoGeneratedText is returned when the MLX response contains no generated text.
var errNoGeneratedText = errors.New("no generated text in MLX response")
