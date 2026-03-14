// main.go — CLI tool to verify MCP tool/prompt/resource registration counts.
//
// Package main provides the entry point for the sanity-check tool.
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
// Tools = 36 base (RegisterAllTools in registry.go) + 1 conditional (Apple Foundation Models on darwin/arm64/cgo) = 37 on Mac Silicon
// Prompts = 36 (19 original + 16 migrated from Python + 1 tractatus_decompose)
// Resources = 38 (27 RegisterResource + 11 RegisterResourceTemplate calls):
// config(2) + scorecard + memories(6) + prompts(4) + session(2) + server + models + cursor/skills + tools(2) + tasks(7) + agent/card = 27 resources
// + 11 templates (memories/category, memories/task, memories/session, prompts/mode, prompts/persona, prompts/category, tools/{category}, tasks/{task_id}, tasks/status, tasks/priority, tasks/tag)
const (
	ExpectedTools     = 37 // Base tools (38 with conditional Apple Foundation Models on darwin/arm64/cgo)
	ExpectedPrompts   = 36
	ExpectedResources = 38
)

// Counting wrapper to track registrations.
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

func (c *countingServer) RegisterResourceTemplate(uriTemplate, name, description, mimeType string, handler framework.ResourceHandler) error {
	c.resourceCount++
	if registrar, ok := c.MCPServer.(framework.ResourceTemplateRegistrar); ok {
		return registrar.RegisterResourceTemplate(uriTemplate, name, description, mimeType, handler)
	}
	return nil
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
	toolMatch := toolCount == ExpectedTools || toolCount == ExpectedTools+1

	if toolMatch {
		fmt.Printf("✅ Tools: %d/%d (or %d with Apple FM)\n", toolCount, ExpectedTools, ExpectedTools+1)
	} else {
		fmt.Printf("❌ Tools: %d/%d (MISMATCH)\n", toolCount, ExpectedTools)
	}

	// Prompts
	promptCount := server.promptCount
	promptMatch := promptCount == ExpectedPrompts

	if promptMatch {
		fmt.Printf("✅ Prompts: %d/%d\n", promptCount, ExpectedPrompts)
	} else {
		fmt.Printf("❌ Prompts: %d/%d (MISMATCH)\n", promptCount, ExpectedPrompts)
	}

	// Resources
	resourceCount := server.resourceCount
	resourceMatch := resourceCount == ExpectedResources

	if resourceMatch {
		fmt.Printf("✅ Resources: %d/%d\n", resourceCount, ExpectedResources)
	} else {
		fmt.Printf("❌ Resources: %d/%d (MISMATCH)\n", resourceCount, ExpectedResources)
	}

	// Summary
	fmt.Println()

	if toolMatch && promptMatch && resourceMatch {
		fmt.Println("✅ All counts match!")
		fmt.Println()
		fmt.Printf("Summary: %d tools, %d prompts, %d resources\n", toolCount, promptCount, resourceCount)
		os.Exit(0)
	}
	fmt.Println("❌ Count mismatches detected!")
	fmt.Println()

	if !toolMatch {
		fmt.Printf("  Tools: Expected %d, got %d (difference: %d)\n",
			ExpectedTools, toolCount, toolCount-ExpectedTools)
	}

	if !promptMatch {
		fmt.Printf("  Prompts: Expected %d, got %d (difference: %d)\n",
			ExpectedPrompts, promptCount, promptCount-ExpectedPrompts)
	}

	if !resourceMatch {
		fmt.Printf("  Resources: Expected %d, got %d (difference: %d)\n",
			ExpectedResources, resourceCount, resourceCount-ExpectedResources)
	}

	os.Exit(1)
}
