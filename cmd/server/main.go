package main

import (
	"context"
	"flag"
	"os"

	"github.com/davidl71/exarp-go/internal/cli"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/logging"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/internal/resources"
	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	logging.Init()

	// Normalize "exarp-go tool_name key=value ..." (e.g. from git hooks) to -tool and -args
	// so we run CLI mode instead of MCP server (which would read stdin and fail on non-JSON).
	if normalized, ok := cli.NormalizeToolArgs(os.Args); ok {
		os.Args = normalized
	}

	// In git hook context (GIT_HOOK=1), never run MCP: stdin has refs, not JSON-RPC.
	// If no CLI flags, exit with message so we don't try to parse stdin ("invalid character 'r'").
	if os.Getenv("GIT_HOOK") == "1" && !cli.HasCLIFlags(os.Args) {
		logging.Warn("In git hooks use: exarp-go -tool <name> -args '<json>' (e.g. -tool analyze_alignment -args '{\"action\":\"todo2\"}')")
		os.Exit(1)
	}

	// Check for CLI flags (completion, list, -tool, task, config, tui, etc.) - work without TTY
	// Use DetectMode for TTY-based mode; explicit flags take precedence.
	if cli.HasCLIFlags(os.Args) || cli.DetectMode() == cli.ModeCLI {
		// CLI mode - run command line interface
		if err := cli.Run(); err != nil {
			logging.Fatal("CLI error: %v", err)
		}
		return
	}

	// Reset flag parsing for MCP mode (in case it was partially parsed)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Initialize database (before server creation)
	projectRoot, err := tools.FindProjectRoot()
	if err != nil {
		logging.Warn("Could not find project root: %v (database unavailable, will use JSON fallback)", err)
	} else {
		if err := database.Init(projectRoot); err != nil {
			logging.Warn("Database initialization failed: %v (fallback to JSON)", err)
		} else {
			defer func() {
				if err := database.Close(); err != nil {
					logging.Warn("Error closing database: %v", err)
				}
			}()
		}
	}

	// MCP server mode - run as stdio server
	cfg, err := config.Load()
	if err != nil {
		logging.Fatal("Failed to load config: %v", err)
	}

	server, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		logging.Fatal("Failed to create server: %v", err)
	}

	if err := tools.RegisterAllTools(server); err != nil {
		logging.Fatal("Failed to register tools: %v", err)
	}

	if err := prompts.RegisterAllPrompts(server); err != nil {
		logging.Fatal("Failed to register prompts: %v", err)
	}

	if err := resources.RegisterAllResources(server); err != nil {
		logging.Fatal("Failed to register resources: %v", err)
	}

	ctx := context.Background()
	if err := server.Run(ctx, nil); err != nil {
		logging.Fatal("Server error: %v", err)
	}
}
