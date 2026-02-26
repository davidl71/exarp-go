// registry.go — MCP tool/prompt/resource registration.
//
// Package tools provides all MCP tool handlers for the exarp-go server.
// Tools are grouped into four semantic registries:
//   - registry_core.go:  task_workflow, task_discovery, task_analysis, session, report, health
//   - registry_ai.go:    memory, estimation, ollama, mlx, llamacpp, text_generate, context, recommend, cursor, FM
//   - registry_infra.go: automation, git_tools, testing, lint, security, generate_config, setup_hooks
//   - registry_misc.go:  analyze_alignment, check_attribution, add_external_tool_hints, tool_catalog,
//                        workflow_mode, infer_session_mode, context_budget
//
// Handler files follow the naming convention: <tool_name>.go or <tool_name>_native.go (platform-specific).
// All handlers accept (ctx context.Context, args json.RawMessage) and return ([]framework.TextContent, error).
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// RegisterAllTools registers all tools with the server.
func RegisterAllTools(server framework.MCPServer) error {
	if err := registerCoreTools(server); err != nil {
		return fmt.Errorf("failed to register core tools: %w", err)
	}
	if err := registerAITools(server); err != nil {
		return fmt.Errorf("failed to register AI tools: %w", err)
	}
	if err := registerInfraTools(server); err != nil {
		return fmt.Errorf("failed to register infra tools: %w", err)
	}
	if err := registerMiscTools(server); err != nil {
		return fmt.Errorf("failed to register misc tools: %w", err)
	}
	return nil
}
