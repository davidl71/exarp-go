// task_workflow_ai_run.go — Task workflow: summarize and run-with-AI handlers.
// See also: task_workflow_create_ai.go
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/spf13/cast"
)

// generateWithBackend handles the common backend selection and generation logic.
func generateWithBackend(ctx context.Context, prompt, backend, operation string, maxTokens int, temperature float32) (string, error) {
	backend = strings.TrimSpace(strings.ToLower(backend))
	if backend == "mlx" {
		backend = "fm"
	}

	var result string
	var err error

	switch backend {
	case "ollama":
		result, err = func() (string, error) {
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
				return "", fmt.Errorf("%s: FM and Ollama both unavailable", operation)
			}

			var genResp map[string]interface{}
			if e2 := json.Unmarshal([]byte(tc[0].Text), &genResp); e2 == nil {
				if resp, ok := genResp["response"].(string); ok {
					result = strings.TrimSpace(resp)
				}
			}

			if result == "" {
				result = strings.TrimSpace(tc[0].Text)
			}
		} else {
			result, err = gen.Generate(ctx, prompt, maxTokens, temperature)
		}
	}

	if err != nil {
		return "", fmt.Errorf("%s: generation failed: %w", operation, err)
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return "", fmt.Errorf("%s: empty response from %s backend", operation, backend)
	}

	return result, nil
}

// ─── Contents ───────────────────────────────────────────────────────────────
//   handleTaskWorkflowSummarize — handleTaskWorkflowSummarize generates an AI summary of a task using the preferred local backend
//   handleTaskWorkflowRunWithAI — handleTaskWorkflowRunWithAI loads a task, builds a prompt from its name and description,
// ────────────────────────────────────────────────────────────────────────────

// ─── handleTaskWorkflowSummarize ────────────────────────────────────────────
// handleTaskWorkflowSummarize generates an AI summary of a task using the preferred local backend
// (fm|ollama) and saves it as a comment. Uses BuildEstimationPrompt-style prompt building.
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
		if b == "fm" || b == "ollama" {
			backend = b
		} else if b == "mlx" {
			backend = "fm"
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
	summaryText, err := generateWithBackend(ctx, prompt, backend, "summarize", 256, 0.3)
	if err != nil {
		return nil, err
	}

	// Save summary as a comment (default: save)
	saveComment := ParamBool(params, "save_comment", true)

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

			return framework.FormatResult(result, "")
		}
	}

	result := map[string]interface{}{
		"success":       true,
		"task_id":       taskID,
		"backend":       backend,
		"summary":       summaryText,
		"comment_saved": saveComment,
	}

	return framework.FormatResult(result, "")
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
		if b == "fm" || b == "ollama" {
			backend = b
		} else if b == "mlx" {
			backend = "fm"
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
	outputText, err := generateWithBackend(ctx, prompt, backend, "run_with_ai", 512, 0.5)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"success":     true,
		"task_id":     taskID,
		"task_name":   task.Content,
		"backend":     backend,
		"instruction": instruction,
		"output":      outputText,
	}

	return framework.FormatResult(result, "")
}
