package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/davidl71/exarp-go/internal/bridge"
	"github.com/davidl71/exarp-go/internal/cli"
	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/database"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/internal/resources"
	"github.com/davidl71/exarp-go/internal/tools"
)

func main() {
	// Normalize "exarp-go tool_name key=value ..." (e.g. from git hooks) to -tool and -args
	// so we run CLI mode instead of MCP server (which would read stdin and fail on non-JSON).
	if len(os.Args) >= 2 && len(os.Args[1]) > 0 && os.Args[1][0] != '-' {
		first := os.Args[1]
		if first != "task" && first != "config" && first != "tui" && first != "tui3270" &&
			!strings.Contains(first, "/") {
			// Treat as tool name with key=value args
			argsMap := make(map[string]interface{})
			for _, arg := range os.Args[2:] {
				if idx := strings.IndexByte(arg, '='); idx >= 0 {
					key := strings.TrimSpace(arg[:idx])
					val := strings.TrimSpace(arg[idx+1:])
					if key != "" {
						argsMap[key] = val
					}
				}
			}
			argsJSON, err := json.Marshal(argsMap)
			if err == nil {
				os.Args = []string{os.Args[0], "-tool", first, "-args", string(argsJSON)}
			}
		}
	}

	// Check for CLI flags first (completion, list, etc.) - these should work even without TTY
	// Also check for task and config subcommands
	hasCLIFlags := false
	if len(os.Args) > 1 && (os.Args[1] == "task" || os.Args[1] == "config" || os.Args[1] == "tui") {
		hasCLIFlags = true
	} else {
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
		// Only log warnings/errors (not INFO) for non-MCP operations
		log.Printf("Warning: Could not find project root: %v (database unavailable, will use JSON fallback)", err)
	} else {
		if err := database.Init(projectRoot); err != nil {
			// Only log warnings/errors (not INFO) for non-MCP operations
			log.Printf("Warning: Database initialization failed: %v (fallback to JSON)", err)
		} else {
			defer func() {
				if err := database.Close(); err != nil {
					log.Printf("Warning: Error closing database: %v", err)
				}
			}()
			// Suppress INFO log for database initialization (not MCP-related)
			// Only log at WARN level for important database messages
		}
	}

	// Cleanup Python process pool on shutdown
	defer func() {
		pool := bridge.GetGlobalPool()
		if err := pool.Close(); err != nil {
			log.Printf("Warning: Error closing Python process pool: %v", err)
		}
	}()

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
	// Note: Transport parameter is ignored by Go SDK adapter (always uses stdio)
	ctx := context.Background()
	if err := server.Run(ctx, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
