// Package tools: plan-and-execute flow using DefaultFMProvider (Apple FM → Ollama → stub).
// Tag hints for Todo2: #feature #mcp
//
// This implements a coordinator/planner/worker pattern in Go: one "plan" generation
// to break a task into subtasks, then parallel "worker" generations, then combine.
// See docs/APPLE_FM_PLANNER_WORKER.md and the Natasha the Robot article on Apple FM agents.

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/davidl71/exarp-go/internal/framework"
)

const (
	defaultMaxSubtasks = 5
	defaultPlanTokens  = 512
	defaultWorkerTokens = 512
)

// handleFMPlanAndExecute runs a plan-and-execute flow: plan (one FM call) → workers (parallel FM calls) → combine.
func handleFMPlanAndExecute(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	task, _ := params["task"].(string)
	if task == "" {
		return nil, fmt.Errorf("task is required for fm_plan_and_execute")
	}

	gen := DefaultFMProvider()
	if gen == nil || !gen.Supported() {
		return []framework.TextContent{{
			Type: "text",
			Text: "fm_plan_and_execute requires a supported backend (Apple Foundation Models or Ollama). None available.",
		}}, nil
	}

	maxSubtasks := defaultMaxSubtasks
	if n, ok := params["max_subtasks"].(float64); ok && n > 0 && n <= 20 {
		maxSubtasks = int(n)
	} else if n, ok := params["max_subtasks"].(int); ok && n > 0 && n <= 20 {
		maxSubtasks = n
	}

	planTokens := defaultPlanTokens
	if n, ok := params["plan_max_tokens"].(float64); ok && n > 0 {
		planTokens = int(n)
	} else if n, ok := params["plan_max_tokens"].(int); ok && n > 0 {
		planTokens = n
	}

	workerTokens := defaultWorkerTokens
	if n, ok := params["worker_max_tokens"].(float64); ok && n > 0 {
		workerTokens = int(n)
	} else if n, ok := params["worker_max_tokens"].(int); ok && n > 0 {
		workerTokens = n
	}

	// Phase 1: Planner — break task into subtasks
	planPrompt := fmt.Sprintf(`Break down the following task into %d or fewer clear, actionable subtasks.
Return ONLY a JSON array of strings, one string per subtask, no other text. Example: ["First do X", "Then do Y"]
Task: %s`, maxSubtasks, task)

	planOut, err := gen.Generate(ctx, planPrompt, planTokens, 0.3)
	if err != nil {
		return nil, fmt.Errorf("planner failed: %w", err)
	}

	subtasks := parseSubtasks(planOut, maxSubtasks)
	if len(subtasks) == 0 {
		// No parseable subtasks: run the whole task as one "worker"
		subtasks = []string{task}
	}

	// Phase 2: Workers — execute each subtask in parallel
	results := make([]string, len(subtasks))
	var wg sync.WaitGroup
	for i, st := range subtasks {
		wg.Add(1)
		go func(idx int, subtask string) {
			defer wg.Done()
			workerPrompt := fmt.Sprintf("Complete this subtask concisely. Return only the result, no preamble.\n\nSubtask: %s", subtask)
			out, _ := gen.Generate(ctx, workerPrompt, workerTokens, 0.4)
			if out != "" {
				results[idx] = strings.TrimSpace(out)
			}
		}(i, st)
	}
	wg.Wait()

	// Phase 3: Combine in order
	var b strings.Builder
	b.WriteString("## Plan-and-execute result\n\n")
	for i, st := range subtasks {
		b.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, st))
		if results[i] != "" {
			b.WriteString(results[i])
			b.WriteString("\n\n")
		}
	}
	combined := strings.TrimSpace(b.String())

	return []framework.TextContent{{Type: "text", Text: combined}}, nil
}

// parseSubtasks tries JSON array first, then numbered/bullet lines.
func parseSubtasks(planOut string, max int) []string {
	planOut = strings.TrimSpace(planOut)
	// Try JSON array
	var arr []string
	if err := json.Unmarshal([]byte(planOut), &arr); err == nil && len(arr) > 0 {
		if len(arr) > max {
			arr = arr[:max]
		}
		return arr
	}
	// Strip markdown code block if present
	if strings.HasPrefix(planOut, "```") {
		planOut = regexp.MustCompile("(?s)^```\\w*\\n?").ReplaceAllString(planOut, "")
		planOut = regexp.MustCompile("\\n?```\\s*$").ReplaceAllString(planOut, "")
		planOut = strings.TrimSpace(planOut)
		if err := json.Unmarshal([]byte(planOut), &arr); err == nil && len(arr) > 0 {
			if len(arr) > max {
				arr = arr[:max]
			}
			return arr
		}
	}
	// Fallback: numbered or bullet lines (1. 2. - *)
	lines := strings.Split(planOut, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove leading "1. ", "2. ", "- ", "* "
		for _, re := range []*regexp.Regexp{
			regexp.MustCompile(`^\d+[.)]\s*`),
			regexp.MustCompile(`^[-*]\s*`),
		} {
			line = re.ReplaceAllString(line, "")
		}
		line = strings.TrimSpace(line)
		if line != "" {
			arr = append(arr, line)
			if len(arr) >= max {
				break
			}
		}
	}
	return arr
}
