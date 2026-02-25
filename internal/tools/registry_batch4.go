// registry_batch4.go — MCP tool registration batch.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// Note: Individual Python bridge tools (context_summarize, context_batch, prompt_log, prompt_analyze,
// recommend_model, recommend_workflow) were removed in favor of unified tools in Batch 5:
// - context(action=summarize|budget|batch)
// - prompt_tracking(action=log|analyze)
// - recommend(action=model|workflow|advisor).
func registerBatch4Tools(server framework.MCPServer) error {
	// Native Go tools (2)
	// context_budget
	if err := server.RegisterTool(
		"context_budget",
		"[HINT: Estimate token usage and suggest context reduction. Use when managing context window limits. Related: context.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"items": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of items to analyze",
				},
				"budget_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     4000,
					"description": "Target token budget",
				},
			},
			Required: []string{"items"},
		},
		handleContextBudget,
	); err != nil {
		return fmt.Errorf("failed to register context_budget: %w", err)
	}

	// Note: list_models tool removed - converted to stdio://models resource
	// See internal/resources/models.go for resource implementation

	return nil
}
