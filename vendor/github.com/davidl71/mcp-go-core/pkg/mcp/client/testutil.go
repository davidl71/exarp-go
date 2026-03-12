// Package client provides testing utilities for mcp-go-core servers.

package client

import (
	"context"
	"fmt"

	"github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// TestServerConfig configures a test client for a server.
type TestServerConfig struct {
	// ServerCommand is the command to launch the server (path to binary)
	ServerCommand string

	// ServerArgs are arguments to pass to the server
	ServerArgs []string

	// ClientInfo identifies this test client
	ClientInfo protocol.ClientInfo
}

// NewTestClient creates a client configured for testing mcp-go-core servers.
//
// This is a convenience function that creates a client with default test settings.
func NewTestClient(config TestServerConfig) (*Client, error) {
	if config.ServerCommand == "" {
		return nil, fmt.Errorf("server command is required")
	}

	if config.ClientInfo.Name == "" {
		config.ClientInfo.Name = "test-client"
	}
	if config.ClientInfo.Version == "" {
		config.ClientInfo.Version = "1.0.0"
	}

	return NewClientWithArgs(config.ServerCommand, config.ServerArgs, config.ClientInfo)
}

// TestToolExecution tests a tool execution end-to-end.
//
// This helper function:
// 1. Initializes the client (if not already initialized)
// 2. Calls the tool with the given arguments
// 3. Returns the results
//
// This is useful for integration tests.
func TestToolExecution(ctx context.Context, c *Client, toolName string, args map[string]interface{}) ([]types.TextContent, error) {
	if !c.IsInitialized() {
		if _, err := c.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize client: %w", err)
		}
	}

	result, err := c.CallTool(ctx, toolName, args)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}

// AssertToolExists asserts that a tool exists in the server's tool list
// and optionally validates its schema.
//
// Returns an error if:
// - The tool doesn't exist
// - The schema doesn't match (if expectedSchema is provided)
func AssertToolExists(ctx context.Context, c *Client, toolName string, expectedSchema *types.ToolSchema) error {
	if !c.IsInitialized() {
		if _, err := c.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize client: %w", err)
		}
	}

	tools, err := c.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	var foundTool *types.ToolInfo
	for i := range tools {
		if tools[i].Name == toolName {
			foundTool = &tools[i]
			break
		}
	}

	if foundTool == nil {
		return fmt.Errorf("tool %q not found in server", toolName)
	}

	// If expected schema is provided, validate it
	if expectedSchema != nil {
		// Basic schema validation - compare key fields
		// More sophisticated validation could be added
		if foundTool.Schema.Type != expectedSchema.Type {
			return fmt.Errorf("tool %q schema type mismatch: got %q, expected %q",
				toolName, foundTool.Schema.Type, expectedSchema.Type)
		}
	}

	return nil
}

// TestServerCapabilities tests basic server capabilities.
//
// This function tests:
// 1. Server initialization
// 2. Tool listing
// 3. Resource listing (if available)
// 4. Prompt listing (if available)
//
// Returns a summary of capabilities found.
type ServerCapabilities struct {
	ToolsAvailable     bool
	ResourcesAvailable bool
	PromptsAvailable   bool
	ToolCount          int
	ResourceCount      int
	PromptCount        int
}

// TestServerCapabilities tests the server's capabilities and returns a summary.
func TestServerCapabilities(ctx context.Context, c *Client) (*ServerCapabilities, error) {
	if !c.IsInitialized() {
		if _, err := c.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize client: %w", err)
		}
	}

	capabilities := &ServerCapabilities{}

	// Test tools
	tools, err := c.ListTools(ctx)
	if err == nil {
		capabilities.ToolsAvailable = true
		capabilities.ToolCount = len(tools)
	}

	// Test resources
	resources, err := c.ListResources(ctx)
	if err == nil {
		capabilities.ResourcesAvailable = true
		capabilities.ResourceCount = len(resources)
	}

	// Test prompts
	prompts, err := c.ListPrompts(ctx)
	if err == nil {
		capabilities.PromptsAvailable = true
		capabilities.PromptCount = len(prompts)
	}

	return capabilities, nil
}
