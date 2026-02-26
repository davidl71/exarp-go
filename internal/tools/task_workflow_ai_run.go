// task_workflow_ai_run.go — Task workflow: summarize and run-with-AI handlers.
// See also: task_workflow_create_ai.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/response"
	"github.com/spf13/cast"
	"strings"
)

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleTaskWorkflowSummarize — handleTaskWorkflowSummarize generates an AI summary of a task using the preferred local backend
//   handleTaskWorkflowRunWithAI — handleTaskWorkflowRunWithAI loads a task, builds a prompt from its name and description,
// ────────────────────────────────────────────────────────────────────────────

// ─── handleTaskWorkflowSummarize ────────────────────────────────────────────
// handleTaskWorkflowSummarize generates an AI summary of a task using the preferred local backend
// (fm|mlx|ollama) and saves it as a comment. Uses BuildEstimationPrompt-style prompt building.
// Params: task_id (required), local_ai_backend (optional, overrides task metadata preferred_backend).
func handleTaskWorkflowSummarize(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := cast.ToString(params["task_id"])
	if taskID == "" {
		return nil, fmt.Errorf("summarize action requires task_id")
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("summarize: failed to get task store: %w", err)
	}

	task, err := store.GetTask(ctx, taskID)
	if err != nil || task == nil {
		return nil, fmt.Errorf("summarize: task %s not found: %w", taskID, err)
	}

	// Determine backend: param overrides task metadata
	backend := ""

	if b, ok := params["local_ai_backend"].(string); ok && b != "" {
		b = strings.TrimSpace(strings.ToLower(b))
		if b == "fm" || b == "mlx" || b == "ollama" {
			backend = b
		}
	}

	if backend == "" {
		backend = GetPreferredBackend(task.Metadata)
	}

	if backend == "" {
		backend = "fm" // default to FM chain (Apple → Ollama → stub)
	}

	// Build summarization prompt
	tagsStr := "none"
	if len(task.Tags) > 0 {
		tagsStr = strings.Join(task.Tags, ", ")
	}

	prompt := fmt.Sprintf(`You are a technical project assistant. Summarize the following software task in 2-3 sentences, highlighting the goal, key steps, and expected outcome.

TASK:
- Name: %s
- Description: %s
- Priority: %s
- Tags: %s

Respond with a concise, plain-text summary only. No JSON, no bullet points.`,
		task.Content, task.LongDescription, task.Priority, tagsStr)

	// Generate summary using selected backend
	var summaryText string

	switch backend {
	case "ollama":
		summaryText, err = func() (string, error) {
			p := map[string]interface{}{
				"action": "generate",
				"prompt": prompt,
				"model":  "llama3.2",
				"stream": false,
			}

			tc, e := DefaultOllama().Invoke(ctx, p)
			if e != nil || len(tc) == 0 {
				return "", fmt.Errorf("ollama generate failed: %w", e)
			}

			var genResp map[string]interface{}
			if e2 := json.Unmarshal([]byte(tc[0].Text), &genResp); e2 == nil {
				if resp, ok := genResp["response"].(string); ok {
					return strings.TrimSpace(resp), nil
				}
			}

			return strings.TrimSpace(tc[0].Text), nil
		}()
	case "mlx":
		gen := DefaultMLXProvider()
		if gen == nil || !gen.Supported() {
			return nil, fmt.Errorf("summarize: MLX provider not available")
		}

		summaryText, err = gen.Generate(ctx, prompt, 256, 0.3)
	default: // "fm"
		gen := DefaultFMProvider()
		if gen == nil || !gen.Supported() {
			// Fall through to ollama if FM unavailable
			p := map[string]interface{}{
				"action": "generate",
				"prompt": prompt,
				"model":  "llama3.2",
				"stream": false,
			}

			tc, e := DefaultOllama().Invoke(ctx, p)
			if e != nil || len(tc) == 0 {
				return nil, fmt.Errorf("summarize: FM and Ollama both unavailable")
			}

			var genResp map[string]interface{}
			if e2 := json.Unmarshal([]byte(tc[0].Text), &genResp); e2 == nil {
				if resp, ok := genResp["response"].(string); ok {
					summaryText = strings.TrimSpace(resp)
				}
			}

			if summaryText == "" {
				summaryText = strings.TrimSpace(tc[0].Text)
			}
		} else {
			summaryText, err = gen.Generate(ctx, prompt, 256, 0.3)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("summarize: generation failed: %w", err)
	}

	summaryText = strings.TrimSpace(summaryText)
	if summaryText == "" {
		return nil, fmt.Errorf("summarize: empty response from %s backend", backend)
	}

	// Save summary as a comment
	saveComment, _ := params["save_comment"].(bool)
	if v, ok := params["save_comment"]; !ok {
		_ = v
		saveComment = true // default: save
	}

	if saveComment {
		commentContent := fmt.Sprintf("## AI Summary (%s)\n\n%s", backend, summaryText)

		comment := database.Comment{
			TaskID:  task.ID,
			Type:    "note",
			Content: commentContent,
		}
		if err2 := database.AddComments(ctx, task.ID, []database.Comment{comment}); err2 != nil {
			// Non-fatal: return summary with warning
			result := map[string]interface{}{
				"success": true,
				"task_id": taskID,
				"backend": backend,
				"summary": summaryText,
				"warning": fmt.Sprintf("summary generated but comment not saved: %v", err2),
			}

			return response.FormatResult(result, "")
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"task_id":       taskID,
		"backend":       backend,
		"summary":       summaryText,
		"comment_saved": saveComment,
	}

	return response.FormatResult(result, "")
}

// ─── handleTaskWorkflowRunWithAI ────────────────────────────────────────────
// handleTaskWorkflowRunWithAI loads a task, builds a prompt from its name and description,
// calls the preferred local AI backend, and returns the output — without applying any file changes.
// Params: task_id (required), local_ai_backend (optional), instruction (optional extra instruction).
func handleTaskWorkflowRunWithAI(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	taskID := cast.ToString(params["task_id"])
	if taskID == "" {
		return nil, fmt.Errorf("run_with_ai action requires task_id")
	}

	store, err := getTaskStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("run_with_ai: failed to get task store: %w", err)
	}

	task, err := store.GetTask(ctx, taskID)
	if err != nil || task == nil {
		return nil, fmt.Errorf("run_with_ai: task %s not found: %w", taskID, err)
	}

	// Determine backend
	backend := ""

	if b, ok := params["local_ai_backend"].(string); ok && b != "" {
		b = strings.TrimSpace(strings.ToLower(b))
		if b == "fm" || b == "mlx" || b == "ollama" {
			backend = b
		}
	}

	if backend == "" {
		backend = GetPreferredBackend(task.Metadata)
	}

	if backend == "" {
		backend = "fm"
	}

	// Optional extra instruction from caller
	instruction := cast.ToString(params["instruction"])
	if instruction == "" {
		instruction = "Analyze this task and provide: 1) a brief implementation plan, 2) key risks or blockers, 3) suggested next steps. Keep your response concise and actionable."
	}

	tagsStr := "none"
	if len(task.Tags) > 0 {
		tagsStr = strings.Join(task.Tags, ", ")
	}

	prompt := fmt.Sprintf(`You are a software engineering assistant. You are working on the following task.

TASK:
- ID: %s
- Name: %s
- Description: %s
- Priority: %s
- Tags: %s
- Status: %s

INSTRUCTION:
%s`,
		task.ID, task.Content, task.LongDescription, task.Priority, tagsStr, task.Status, instruction)

	// Generate using selected backend
	var outputText string

	switch backend {
	case "ollama":
		outputText, err = func() (string, error) {
			p := map[string]interface{}{
				"action": "generate",
				"prompt": prompt,
				"model":  "llama3.2",
				"stream": false,
			}

			tc, e := DefaultOllama().Invoke(ctx, p)
			if e != nil || len(tc) == 0 {
				return "", fmt.Errorf("ollama generate failed: %w", e)
			}

			var genResp map[string]interface{}
			if e2 := json.Unmarshal([]byte(tc[0].Text), &genResp); e2 == nil {
				if resp, ok := genResp["response"].(string); ok {
					return strings.TrimSpace(resp), nil
				}
			}

			return strings.TrimSpace(tc[0].Text), nil
		}()
	case "mlx":
		gen := DefaultMLXProvider()
		if gen == nil || !gen.Supported() {
			return nil, fmt.Errorf("run_with_ai: MLX provider not available")
		}

		outputText, err = gen.Generate(ctx, prompt, 512, 0.5)
	default: // "fm"
		gen := DefaultFMProvider()
		if gen == nil || !gen.Supported() {
			p := map[string]interface{}{
				"action": "generate",
				"prompt": prompt,
				"model":  "llama3.2",
				"stream": false,
			}

			tc, e := DefaultOllama().Invoke(ctx, p)
			if e != nil || len(tc) == 0 {
				return nil, fmt.Errorf("run_with_ai: FM and Ollama both unavailable")
			}

			var genResp map[string]interface{}
			if e2 := json.Unmarshal([]byte(tc[0].Text), &genResp); e2 == nil {
				if resp, ok := genResp["response"].(string); ok {
					outputText = strings.TrimSpace(resp)
				}
			}

			if outputText == "" {
				outputText = strings.TrimSpace(tc[0].Text)
			}
		} else {
			outputText, err = gen.Generate(ctx, prompt, 512, 0.5)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("run_with_ai: generation failed: %w", err)
	}

	outputText = strings.TrimSpace(outputText)
	if outputText == "" {
		return nil, fmt.Errorf("run_with_ai: empty response from %s backend", backend)
	}

	result := map[string]interface{}{
		"success":     true,
		"task_id":     taskID,
		"task_name":   task.Content,
		"backend":     backend,
		"instruction": instruction,
		"output":      outputText,
	}

	return response.FormatResult(result, "")
}
