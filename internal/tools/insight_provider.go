// Package tools: Report insight provider abstraction for AI-generated report/scorecard insights.
// Implementations: MLX via Python bridge, or DefaultFM (Apple FM when available).
// Report code uses DefaultReportInsight so it does not depend on the bridge or MLX by name.

package tools

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/davidl71/exarp-go/internal/bridge"
)

// ReportInsightProvider generates long-form AI insights for report/scorecard content.
// Implementations: MLX via bridge, or DefaultFM when MLX unavailable.
type ReportInsightProvider interface {
	// Supported reports whether this provider can generate insights (e.g. bridge available or FM available).
	Supported() bool
	// Generate runs the model with the given prompt and returns generated text.
	Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error)
}

// defaultReportInsight is the shared default; set by init to composite (MLX then FM).
var defaultReportInsight ReportInsightProvider

func init() {
	defaultReportInsight = &compositeReportInsight{}
}

// DefaultReportInsight returns the default report insight provider (tries MLX then FM).
func DefaultReportInsight() ReportInsightProvider {
	if defaultReportInsight == nil {
		return &compositeReportInsight{}
	}
	return defaultReportInsight
}

// compositeReportInsight tries MLX (bridge) first, then DefaultFM.
type compositeReportInsight struct{}

func (c *compositeReportInsight) Supported() bool {
	return true // We always "support" by trying MLX then FM
}

func (c *compositeReportInsight) Generate(ctx context.Context, prompt string, maxTokens int, temperature float32) (string, error) {
	// Try MLX via bridge first
	if text, err := tryMLXReportInsight(ctx, prompt, maxTokens, temperature); err == nil && text != "" {
		return text, nil
	}
	// Fall back to DefaultFM when available
	if FMAvailable() {
		return DefaultFM.Generate(ctx, prompt, maxTokens, temperature)
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

// executeMLXViaBridge invokes the mlx tool via the Python bridge.
// Extracted so report_mlx and insight_provider can share or so we have a single bridge call site for "mlx" generate.
func executeMLXViaBridge(ctx context.Context, params map[string]interface{}) (string, error) {
	return bridge.ExecutePythonTool(ctx, "mlx", params)
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
