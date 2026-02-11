package main

import (
	"fmt"
	"os"

	"github.com/davidl71/exarp-go/internal/config"
	"github.com/davidl71/exarp-go/internal/factory"
	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/prompts"
	"github.com/davidl71/exarp-go/internal/resources"
	"github.com/davidl71/exarp-go/internal/tools"
)

// Expected counts
// Tools = 30 base (RegisterAllTools in registry.go) + 1 conditional (Apple Foundation Models on darwin/arm64/cgo) = 31 total on Mac Silicon
// Prompts = 36 (19 original + 16 migrated from Python + 1 tractatus_decompose)
// Resources = 23 (scorecard, memories, prompts, session/mode, server/status, models, cursor/skills, tools, tasks)
const (
	EXPECTED_TOOLS     = 30 // Base tools (31 with conditional Apple Foundation Models on darwin/arm64/cgo)
	EXPECTED_PROMPTS   = 36
	EXPECTED_RESOURCES = 23
)

// Counting wrapper to track registrations
type countingServer struct {
	framework.MCPServer
	toolCount     int
	promptCount   int
	resourceCount int
}

func (c *countingServer) RegisterTool(name, description string, schema framework.ToolSchema, handler framework.ToolHandler) error {
	c.toolCount++
	return c.MCPServer.RegisterTool(name, description, schema, handler)
}

func (c *countingServer) RegisterPrompt(name, description string, handler framework.PromptHandler) error {
	c.promptCount++
	return c.MCPServer.RegisterPrompt(name, description, handler)
}

func (c *countingServer) RegisterResource(uri, name, description, mimeType string, handler framework.ResourceHandler) error {
	c.resourceCount++
	return c.MCPServer.RegisterResource(uri, name, description, mimeType, handler)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create server using configured framework
	baseServer, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Wrap with counter
	server := &countingServer{
		MCPServer: baseServer,
	}

	// Register all components (this will increment counters)
	if err := tools.RegisterAllTools(server); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register tools: %v\n", err)
		os.Exit(1)
	}

	if err := prompts.RegisterAllPrompts(server); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register prompts: %v\n", err)
		os.Exit(1)
	}

	if err := resources.RegisterAllResources(server); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to register resources: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Println("=== MCP Server Sanity Check ===")
	fmt.Println()

	// Tools (allow +1 for conditional Apple Foundation Models tool)
	toolCount := server.toolCount
	toolMatch := toolCount == EXPECTED_TOOLS || toolCount == EXPECTED_TOOLS+1
	if toolMatch {
		fmt.Printf("✅ Tools: %d/%d (or %d with Apple FM)\n", toolCount, EXPECTED_TOOLS, EXPECTED_TOOLS+1)
	} else {
		fmt.Printf("❌ Tools: %d/%d (MISMATCH)\n", toolCount, EXPECTED_TOOLS)
	}

	// Prompts
	promptCount := server.promptCount
	promptMatch := promptCount == EXPECTED_PROMPTS
	if promptMatch {
		fmt.Printf("✅ Prompts: %d/%d\n", promptCount, EXPECTED_PROMPTS)
	} else {
		fmt.Printf("❌ Prompts: %d/%d (MISMATCH)\n", promptCount, EXPECTED_PROMPTS)
	}

	// Resources
	resourceCount := server.resourceCount
	resourceMatch := resourceCount == EXPECTED_RESOURCES
	if resourceMatch {
		fmt.Printf("✅ Resources: %d/%d\n", resourceCount, EXPECTED_RESOURCES)
	} else {
		fmt.Printf("❌ Resources: %d/%d (MISMATCH)\n", resourceCount, EXPECTED_RESOURCES)
	}

	// Summary
	fmt.Println()
	if toolMatch && promptMatch && resourceMatch {
		fmt.Println("✅ All counts match!")
		fmt.Println()
		fmt.Printf("Summary: %d tools, %d prompts, %d resources\n", toolCount, promptCount, resourceCount)
		os.Exit(0)
	} else {
		fmt.Println("❌ Count mismatches detected!")
		fmt.Println()
		if !toolMatch {
			fmt.Printf("  Tools: Expected %d, got %d (difference: %d)\n",
				EXPECTED_TOOLS, toolCount, toolCount-EXPECTED_TOOLS)
		}
		if !promptMatch {
			fmt.Printf("  Prompts: Expected %d, got %d (difference: %d)\n",
				EXPECTED_PROMPTS, promptCount, promptCount-EXPECTED_PROMPTS)
		}
		if !resourceMatch {
			fmt.Printf("  Resources: Expected %d, got %d (difference: %d)\n",
				EXPECTED_RESOURCES, resourceCount, resourceCount-EXPECTED_RESOURCES)
		}
		os.Exit(1)
	}
}
