package main

import (
	"context"
	"log"

	"github.com/davidl/mcp-stdio-tools/internal/cli"
	"github.com/davidl/mcp-stdio-tools/internal/config"
	"github.com/davidl/mcp-stdio-tools/internal/factory"
	"github.com/davidl/mcp-stdio-tools/internal/framework"
	"github.com/davidl/mcp-stdio-tools/internal/prompts"
	"github.com/davidl/mcp-stdio-tools/internal/resources"
	"github.com/davidl/mcp-stdio-tools/internal/tools"
)

func main() {
	// Detect if running in CLI mode (TTY) or MCP server mode (stdio)
	if cli.IsTTY() {
		// CLI mode - run command line interface
		if err := cli.Run(); err != nil {
			log.Fatalf("CLI error: %v", err)
		}
		return
	}

	// MCP server mode - run as stdio server
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create server using configured framework
	server, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Register components (framework-agnostic)
	if err := tools.RegisterAllTools(server); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	if err := prompts.RegisterAllPrompts(server); err != nil {
		log.Fatalf("Failed to register prompts: %v", err)
	}

	if err := resources.RegisterAllResources(server); err != nil {
		log.Fatalf("Failed to register resources: %v", err)
	}

	// Run server with stdio transport
	ctx := context.Background()
	transport := framework.NewStdioTransport()
	if err := server.Run(ctx, transport); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

