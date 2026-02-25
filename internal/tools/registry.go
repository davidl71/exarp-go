// registry.go — MCP tool/prompt/resource registration (3 batches of tools).
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// RegisterAllTools registers all tools with the server.
func RegisterAllTools(server framework.MCPServer) error {
	// Batch 1: Simple tools (T-22 through T-27)
	if err := registerBatch1Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 1 tools: %w", err)
	}

	// Batch 2: Medium tools (T-28 through T-35)
	if err := registerBatch2Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 2 tools: %w", err)
	}

	// Batch 3: Advanced tools (T-37 through T-44)
	if err := registerBatch3Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 3 tools: %w", err)
	}

	// Batch 4: mcp-generic-tools migration (2 native Go tools)
	if err := registerBatch4Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 4 tools: %w", err)
	}

	// Batch 5: Phase 3 migration - remaining unified tools (4 tools)
	if err := registerBatch5Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 5 tools: %w", err)
	}

	return nil
}
