// Package tools: task_execute integrates the model-assisted execution flow with Todo2 (T-215).
// Loads a task, renders the task_execution prompt template, calls the model router, parses the
// response, optionally applies file changes, and adds a result comment to the task.
// See docs/MODEL_ASSISTED_WORKFLOW.md Phase 4.

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/prompts"
)

const (
	defaultMinConfidence = 0.5
	defaultMaxTokens     = 4096
	defaultTemperature  = 0.3
)

// handleTaskExecute runs the execution flow for a Todo2 task and optionally applies changes.
func handleTaskExecute(ctx context.Context, args json.RawMessage) ([]framework.TextContent, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	taskID, _ := params["task_id"].(string)
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required")
	}

	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, fmt.Errorf("project root: %w", err)
	}
	if root, ok := params["project_root"].(string); ok && root != "" {
		projectRoot = root
	}

	apply := true
	if v, ok := params["apply"].(bool); ok {
		apply = v
	}

	minConfidence := defaultMinConfidence
	if v, ok := params["min_confidence"].(float64); ok && v >= 0 && v <= 1 {
		minConfidence = v
	}

	result, err := RunTaskExecutionFlow(ctx, RunTaskExecutionFlowParams{
		TaskID:         taskID,
		ProjectRoot:    projectRoot,
		Apply:          apply,
		MinConfidence:  minConfidence,
	})
	if err != nil {
		return nil, err
	}

	out, _ := json.Marshal(result)
	return []framework.TextContent{{Type: "text", Text: string(out)}}, nil
}

// RunTaskExecutionFlowParams holds parameters for RunTaskExecutionFlow.
type RunTaskExecutionFlowParams struct {
	TaskID        string
	ProjectRoot   string
	Apply         bool
	MinConfidence float64
}

// RunTaskExecutionFlowResult is the result of running the execution flow.
type RunTaskExecutionFlowResult struct {
	TaskID       string   `json:"task_id"`
	Applied      []string `json:"applied,omitempty"`
	Explanation  string   `json:"explanation,omitempty"`
	Confidence   float64  `json:"confidence"`
	CommentAdded bool     `json:"comment_added"`
	ParseError   string   `json:"parse_error,omitempty"`
	ApplyError   string   `json:"apply_error,omitempty"`
}

// RunTaskExecutionFlow loads the task, generates an execution plan via the model router,
// parses the response, optionally applies file changes, and adds a result comment to the task.
func RunTaskExecutionFlow(ctx context.Context, p RunTaskExecutionFlowParams) (*RunTaskExecutionFlowResult, error) {
	result := &RunTaskExecutionFlowResult{TaskID: p.TaskID}

	store := NewDefaultTaskStore(p.ProjectRoot)
	task, err := store.GetTask(ctx, p.TaskID)
	if err != nil || task == nil {
		return nil, fmt.Errorf("task %s: %w", p.TaskID, err)
	}

	tpl, err := prompts.GetPromptTemplate("task_execution")
	if err != nil {
		return nil, fmt.Errorf("task_execution template: %w", err)
	}

	args := map[string]interface{}{
		"task_name":        task.Name,
		"task_description": task.Content,
		"code_context":     "",
		"guidelines":       "",
	}
	if task.Content == "" {
		args["task_description"] = task.Name
	}
	prompt := prompts.SubstituteTemplate(tpl, args)

	modelType := DefaultModelRouter.SelectModel("code_generation", ModelRequirements{})
	text, err := DefaultModelRouter.Generate(ctx, modelType, prompt, defaultMaxTokens, defaultTemperature)
	if err != nil {
		return nil, fmt.Errorf("model generate: %w", err)
	}

	parsed, err := ParseExecutionResponse([]byte(text))
	if err != nil {
		result.ParseError = err.Error()
		result.Explanation = "Execution response could not be parsed as JSON."
		addExecutionResultComment(ctx, p.TaskID, result)
		return result, nil
	}

	result.Confidence = parsed.Confidence
	result.Explanation = parsed.Explanation

	if p.Apply && parsed.Confidence >= p.MinConfidence && len(parsed.Changes) > 0 {
		applied, applyErr := ApplyChanges(p.ProjectRoot, parsed.Changes)
		if applyErr != nil {
			result.ApplyError = applyErr.Error()
		} else {
			result.Applied = applied
		}
	}

	addExecutionResultComment(ctx, p.TaskID, result)
	return result, nil
}

func addExecutionResultComment(ctx context.Context, taskID string, result *RunTaskExecutionFlowResult) {
	var b strings.Builder
	b.WriteString("**Auto-execution result**\n\n")
	b.WriteString(result.Explanation)
	if len(result.Applied) > 0 {
		b.WriteString("\n\n**Applied files:**\n")
		for _, f := range result.Applied {
			b.WriteString("- ")
			b.WriteString(f)
			b.WriteString("\n")
		}
	}
	if result.ApplyError != "" {
		b.WriteString("\n**Apply error:** ")
		b.WriteString(result.ApplyError)
	}
	if result.ParseError != "" {
		b.WriteString("\n**Parse error:** ")
		b.WriteString(result.ParseError)
	}

	comment := database.Comment{
		Type:    database.CommentTypeResult,
		Content: strings.TrimSpace(b.String()),
	}
	if err := database.AddComments(ctx, taskID, []database.Comment{comment}); err != nil {
		// DB may be unavailable (e.g. CLI without init); result is still in tool response
		return
	}
	result.CommentAdded = true
}
