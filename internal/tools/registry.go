// registry.go — MCP tool/prompt/resource registration.
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
