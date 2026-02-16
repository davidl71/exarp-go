// Package tools - task_analyzer implements LLM-backed task breakdown per MODEL_ASSISTED_WORKFLOW Phase 3 (T-208).
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/davidl71/exarp-go/internal/prompts"
)

// TaskBreakdownSubtask represents one subtask from the breakdown (matches task_breakdown template JSON).
type TaskBreakdownSubtask struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptance_criteria,omitempty"`
	Complexity         string   `json:"complexity"` // simple | medium | complex
	Dependencies       []string `json:"dependencies,omitempty"`
}

// TaskBreakdownResult is the parsed result of AnalyzeTask (T-208).
type TaskBreakdownResult struct {
	Subtasks []TaskBreakdownSubtask `json:"subtasks"`
}

// AnalyzeTask uses the task_breakdown prompt template and the configured text generator to break a task
// into 3-8 subtasks (T-208). Per docs/MODEL_ASSISTED_WORKFLOW.md Phase 3.
func AnalyzeTask(ctx context.Context, taskDescription, acceptanceCriteria, contextHint string, gen TextGenerator) (*TaskBreakdownResult, error) {
	if taskDescription == "" {
		return nil, fmt.Errorf("task description is required for breakdown")
	}

	if gen == nil || !gen.Supported() {
		return nil, fmt.Errorf("text generator not available for task breakdown")
	}

	tmpl, err := prompts.GetPromptTemplate("task_breakdown")
	if err != nil {
		return nil, fmt.Errorf("get task_breakdown template: %w", err)
	}

	args := map[string]interface{}{
		"task_description":    taskDescription,
		"acceptance_criteria": acceptanceCriteria,
		"context":             contextHint,
	}
	substituted := prompts.SubstituteTemplate(tmpl, args)

	text, err := gen.Generate(ctx, substituted, 1024, 0.3)
	if err != nil {
		return nil, fmt.Errorf("generate breakdown: %w", err)
	}

	return parseTaskBreakdown(text)
}

// parseTaskBreakdown parses LLM output into TaskBreakdownResult. Tries JSON first, then fallback.
func parseTaskBreakdown(text string) (*TaskBreakdownResult, error) {
	text = extractJSON(text)

	var result TaskBreakdownResult

	if err := json.Unmarshal([]byte(text), &result); err == nil && len(result.Subtasks) > 0 {
		return &result, nil
	}
	// Fallback: build minimal result from raw text (one "subtask" per line that looks like a task)
	result = TaskBreakdownResult{}

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "{") || strings.HasPrefix(line, "}") {
			continue
		}

		if len(line) > 10 {
			result.Subtasks = append(result.Subtasks, TaskBreakdownSubtask{
				Name:        line,
				Description: line,
				Complexity:  "medium",
			})
		}
	}

	if len(result.Subtasks) == 0 {
		return nil, fmt.Errorf("could not parse task breakdown from model output")
	}

	return &result, nil
}

// extractJSON returns the first JSON object (or array) found in text, for robust parsing.
func extractJSON(text string) string {
	start := strings.Index(text, "{")
	if start < 0 {
		return text
	}

	depth := 0

	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}
	// Unbalanced; try regex for "subtasks": [...]
	re := regexp.MustCompile(`\{[\s\S]*"subtasks"[\s\S]*\}`)
	if m := re.FindString(text); m != "" {
		return m
	}

	return text[start:]
}
