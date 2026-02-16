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

// SuggestionItem is one improvement suggestion from the suggestions template (T-218).
type SuggestionItem struct {
	Dimension      string `json:"dimension"`
	Issue          string `json:"issue"`
	Recommendation string `json:"recommendation"`
}

// SuggestionsResponse is the JSON response from prompt_optimization_suggestions.
type SuggestionsResponse struct {
	Suggestions []SuggestionItem `json:"suggestions"`
}

// GenerateSuggestions generates improvement suggestions for a prompt given its analysis (T-218).
// Uses the prompt_optimization_suggestions template. Returns suggestions JSON for use with RefinePrompt.
func GenerateSuggestions(ctx context.Context, prompt string, analysis *PromptAnalysis, contextHint, taskType string, gen TextGenerator) (*SuggestionsResponse, error) {
	if gen == nil || !gen.Supported() {
		return nil, fmt.Errorf("text generator not available for suggestions")
	}

	tmpl, err := prompts.GetPromptTemplate("prompt_optimization_suggestions")
	if err != nil {
		return nil, fmt.Errorf("get prompt_optimization_suggestions template: %w", err)
	}

	analysisJSON, _ := json.Marshal(analysis)
	args := map[string]interface{}{
		"prompt":    prompt,
		"analysis":  string(analysisJSON),
		"context":   contextHint,
		"task_type": taskType,
	}
	substituted := prompts.SubstituteTemplate(tmpl, args)

	text, err := gen.Generate(ctx, substituted, 1024, 0.4)
	if err != nil {
		return nil, fmt.Errorf("generate suggestions: %w", err)
	}

	return parseSuggestionsResponse(text)
}

func parseSuggestionsResponse(text string) (*SuggestionsResponse, error) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimSpace(text)
	}

	var out SuggestionsResponse
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		return nil, fmt.Errorf("parse suggestions response: %w", err)
	}

	return &out, nil
}

// RefinePrompt applies improvement suggestions to a prompt using the prompt_optimization_refinement template (T-218).
// Returns the refined prompt text only (template instructs model to return no JSON).
func RefinePrompt(ctx context.Context, prompt string, suggestions *SuggestionsResponse, contextHint, taskType string, gen TextGenerator) (string, error) {
	if gen == nil || !gen.Supported() {
		return "", fmt.Errorf("text generator not available for refinement")
	}

	tmpl, err := prompts.GetPromptTemplate("prompt_optimization_refinement")
	if err != nil {
		return "", fmt.Errorf("get prompt_optimization_refinement template: %w", err)
	}

	suggestionsJSON, _ := json.Marshal(suggestions)
	args := map[string]interface{}{
		"prompt":      prompt,
		"suggestions": string(suggestionsJSON),
		"context":     contextHint,
		"task_type":   taskType,
	}
	substituted := prompts.SubstituteTemplate(tmpl, args)

	refined, err := gen.Generate(ctx, substituted, 2048, 0.4)
	if err != nil {
		return "", fmt.Errorf("generate refinement: %w", err)
	}

	return strings.TrimSpace(refined), nil
}

// RefinePromptLoopOptions configures the iterative refinement loop (T-218).
type RefinePromptLoopOptions struct {
	MaxIterations     int     // max analyze→suggest→refine cycles (default 3)
	MinScoreThreshold float64 // stop when all dimensions >= this (default 0.8)
}

// MinScore returns the minimum of the five dimension scores (for threshold check).
func (p *PromptAnalysis) MinScore() float64 {
	min := p.Clarity
	if p.Specificity < min {
		min = p.Specificity
	}

	if p.Completeness < min {
		min = p.Completeness
	}

	if p.Structure < min {
		min = p.Structure
	}

	if p.Actionability < min {
		min = p.Actionability
	}

	return min
}

// RefinePromptLoop runs analyze → suggestions → refine repeatedly until MinScore >= threshold or max iterations (T-218).
// Returns the final prompt, the last analysis, and any error.
func RefinePromptLoop(ctx context.Context, initialPrompt, contextHint, taskType string, gen TextGenerator, opts *RefinePromptLoopOptions) (finalPrompt string, lastAnalysis *PromptAnalysis, err error) {
	if initialPrompt == "" {
		return "", nil, fmt.Errorf("initial prompt is required")
	}

	if gen == nil || !gen.Supported() {
		return "", nil, fmt.Errorf("text generator not available for refinement loop")
	}

	maxIter := 3
	threshold := 0.8

	if opts != nil {
		if opts.MaxIterations > 0 {
			maxIter = opts.MaxIterations
		}

		if opts.MinScoreThreshold > 0 {
			threshold = opts.MinScoreThreshold
		}
	}

	current := initialPrompt

	var analysis *PromptAnalysis

	for i := 0; i < maxIter; i++ {
		analysis, err = AnalyzePrompt(ctx, current, contextHint, taskType, gen)
		if err != nil {
			return current, analysis, err
		}

		if analysis.MinScore() >= threshold {
			return current, analysis, nil
		}

		suggestions, err := GenerateSuggestions(ctx, current, analysis, contextHint, taskType, gen)
		if err != nil {
			return current, analysis, err
		}

		if len(suggestions.Suggestions) == 0 {
			return current, analysis, nil
		}

		current, err = RefinePrompt(ctx, current, suggestions, contextHint, taskType, gen)
		if err != nil {
			return current, analysis, err
		}
	}

	return current, analysis, nil
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
