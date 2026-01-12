package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/davidl71/exarp-go/internal/cli"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/internal/resources"
	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	// Check for CLI flags first (completion, list, etc.) - these should work even without TTY
	hasCLIFlags := false
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-completion" || arg == "--completion" ||
			arg == "-list" || arg == "--list" ||
			arg == "-tool" || arg == "--tool" ||
			arg == "-test" || arg == "--test" ||
			arg == "-i" || arg == "--interactive" ||
			arg == "-args" || arg == "--args" ||
			arg == "-h" || arg == "--help" || arg == "help" {
			hasCLIFlags = true
			break
		}
		// Stop at first non-flag argument
		if len(arg) > 0 && arg[0] != '-' {
			break
		}
	}

	// If CLI flags are present or we're in a TTY, run CLI mode
	if hasCLIFlags || cli.IsTTY() {
		// CLI mode - run command line interface
		if err := cli.Run(); err != nil {
			log.Fatalf("CLI error: %v", err)
		}
		return
	}

	// Reset flag parsing for MCP mode (in case it was partially parsed)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Initialize database (before server creation)
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		log.Printf("Warning: Could not find project root: %v (database unavailable, will use JSON fallback)", err)
	} else {
		if err := database.Init(projectRoot); err != nil {
			log.Printf("Warning: Database initialization failed: %v (fallback to JSON)", err)
		} else {
			defer func() {
				if err := database.Close(); err != nil {
					log.Printf("Warning: Error closing database: %v", err)
				}
			}()
			log.Printf("Database initialized: %s/.todo2/todo2.db", projectRoot)
		}
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
