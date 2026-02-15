// Package tools - prompt_analyzer implements prompt quality analysis per MODEL_ASSISTED_WORKFLOW Phase 5 (T-216).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/davidl71/exarp-go/internal/prompts"
)

// PromptAnalysis holds the result of analyzing a prompt (per PROMPT_OPTIMIZATION_TEMPLATE_SPEC.md).
type PromptAnalysis struct {
	Clarity       float64  `json:"clarity"`
	Specificity   float64  `json:"specificity"`
	Completeness  float64  `json:"completeness"`
	Structure     float64  `json:"structure"`
	Actionability float64  `json:"actionability"`
	Summary       string   `json:"summary"`
	Notes         []string `json:"notes"`
}

// AnalyzePrompt analyzes a prompt for quality using the prompt_optimization_analysis template (T-216).
// Uses the configured text generator (FM, MLX, or Ollama) to evaluate clarity, specificity,
// completeness, structure, and actionability. Per docs/PROMPT_OPTIMIZATION_TEMPLATE_SPEC.md.
func AnalyzePrompt(ctx context.Context, prompt, contextHint, taskType string, gen TextGenerator) (*PromptAnalysis, error) {
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required for analysis")
	}
	if gen == nil || !gen.Supported() {
		return nil, fmt.Errorf("text generator not available for prompt analysis")
	}

	tmpl, err := prompts.GetPromptTemplate("prompt_optimization_analysis")
	if err != nil {
		return nil, fmt.Errorf("get prompt_optimization_analysis template: %w", err)
	}

	args := map[string]interface{}{
		"prompt":    prompt,
		"context":   contextHint,
		"task_type": taskType,
	}
	substituted := prompts.SubstituteTemplate(tmpl, args)

	text, err := gen.Generate(ctx, substituted, 512, 0.3)
	if err != nil {
		return nil, fmt.Errorf("generate analysis: %w", err)
	}

	return parsePromptAnalysis(text)
}

// parsePromptAnalysis parses LLM output into PromptAnalysis. Handles JSON and fallback plain-text.
func parsePromptAnalysis(text string) (*PromptAnalysis, error) {
	// Try JSON parse first
	var result PromptAnalysis
	if err := json.Unmarshal([]byte(text), &result); err == nil {
		return &result, nil
	}

	// Fallback: parse line-by-line for "dimension: score" (per PROMPT_OPTIMIZATION_TEMPLATE_SPEC)
	result = PromptAnalysis{}
	scoreRe := regexp.MustCompile(`(?i)(clarity|specificity|completeness|structure|actionability)\s*[:=]\s*([0-9.]+)`)
	for _, line := range regexp.MustCompile(`\r?\n`).Split(text, -1) {
		matches := scoreRe.FindStringSubmatch(line)
		if len(matches) == 3 {
			v, _ := strconv.ParseFloat(matches[2], 64)
			switch strings.ToLower(matches[1]) {
			case "clarity":
				result.Clarity = v
			case "specificity":
				result.Specificity = v
			case "completeness":
				result.Completeness = v
			case "structure":
				result.Structure = v
			case "actionability":
				result.Actionability = v
			}
		}
	}
	return &result, nil
}
