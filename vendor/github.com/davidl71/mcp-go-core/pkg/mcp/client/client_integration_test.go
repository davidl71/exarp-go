// +build integration

// Integration tests for the client wrapper.
// These tests require:
//   - github.com/metoro-io/mcp-golang to be installed
//   - An MCP server to test against
//
// Run with: go test -tags integration ./pkg/mcp/client

package client

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// TestClientInitialization tests basic client initialization.
func TestClientInitialization(t *testing.T) {
	serverCommand := os.Getenv("MCP_TEST_SERVER")
	if serverCommand == "" {
		t.Skip("Skipping integration test: MCP_TEST_SERVER not set")
	}

	clientInfo := protocol.ClientInfo{
		Name:    "integration-test-client",
		Version: "1.0.0",
	}

	c, err := NewClient(serverCommand, clientInfo)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	initResult, err := c.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize client: %v", err)
	}

	if initResult.ServerInfo.Name == "" {
		t.Error("Server info name is empty")
	}

	if initResult.ProtocolVersion == "" {
		t.Error("Protocol version is empty")
	}

	t.Logf("Initialized with server: %s v%s",
		initResult.ServerInfo.Name,
		initResult.ServerInfo.Version)
}

// TestListTools tests listing tools from the server.
func TestListTools(t *testing.T) {
	serverCommand := os.Getenv("MCP_TEST_SERVER")
	if serverCommand == "" {
		t.Skip("Skipping integration test: MCP_TEST_SERVER not set")
	}

	clientInfo := protocol.ClientInfo{
		Name:    "integration-test-client",
		Version: "1.0.0",
	}

	c, err := NewClient(serverCommand, clientInfo)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize
	_, err = c.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// List tools
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) == 0 {
		t.Log("No tools found (this may be expected for some servers)")
		return
	}

	t.Logf("Found %d tools:", len(tools))
	for _, tool := range tools {
		t.Logf("  - %s: %s", tool.Name, tool.Description)
		if tool.Schema.Type == "" {
			t.Errorf("Tool %s has empty schema type", tool.Name)
		}
	}
}

// TestCallTool tests calling a tool on the server.
func TestCallTool(t *testing.T) {
	serverCommand := os.Getenv("MCP_TEST_SERVER")
	if serverCommand == "" {
		t.Skip("Skipping integration test: MCP_TEST_SERVER not set")
	}

	clientInfo := protocol.ClientInfo{
		Name:    "integration-test-client",
		Version: "1.0.0",
	}

	c, err := NewClient(serverCommand, clientInfo)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize
	_, err = c.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// List tools first to find one to call
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) == 0 {
		t.Skip("No tools available to test")
	}

	// Try calling the first tool with minimal args
	toolName := tools[0].Name
	t.Logf("Testing tool: %s", toolName)

	// Use minimal/empty args - tool may have optional parameters
	args := map[string]interface{}{}

	result, err := c.CallTool(ctx, toolName, args)
	if err != nil {
		// Some tools may require specific args - log but don't fail
		t.Logf("Tool call failed (may require specific args): %v", err)
		return
	}

	t.Logf("Tool call succeeded, got %d result(s)", len(result))
	for i, content := range result {
		t.Logf("Result %d: %s", i, content.Text)
	}
}

// TestListResources tests listing resources from the server.
func TestListResources(t *testing.T) {
	serverCommand := os.Getenv("MCP_TEST_SERVER")
	if serverCommand == "" {
		t.Skip("Skipping integration test: MCP_TEST_SERVER not set")
	}

	clientInfo := protocol.ClientInfo{
		Name:    "integration-test-client",
		Version: "1.0.0",
	}

	c, err := NewClient(serverCommand, clientInfo)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize
	_, err = c.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// List resources
	resources, err := c.ListResources(ctx)
	if err != nil {
		t.Fatalf("Failed to list resources: %v", err)
	}

	if len(resources) == 0 {
		t.Log("No resources found (this may be expected for some servers)")
		return
	}

	t.Logf("Found %d resources:", len(resources))
	for _, resource := range resources {
		t.Logf("  - %s: %s", resource.URI, resource.Name)
	}
}

// TestTestServerCapabilities tests the TestServerCapabilities utility.
func TestTestServerCapabilities(t *testing.T) {
	serverCommand := os.Getenv("MCP_TEST_SERVER")
	if serverCommand == "" {
		t.Skip("Skipping integration test: MCP_TEST_SERVER not set")
	}

	config := TestServerConfig{
		ServerCommand: serverCommand,
		ClientInfo: protocol.ClientInfo{
			Name:    "integration-test-client",
			Version: "1.0.0",
		},
	}

	c, err := NewTestClient(config)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	caps, err := TestServerCapabilities(ctx, c)
	if err != nil {
		t.Fatalf("Failed to test capabilities: %v", err)
	}

	t.Logf("Server capabilities:")
	t.Logf("  Tools available: %v (%d)", caps.ToolsAvailable, caps.ToolCount)
	t.Logf("  Resources available: %v (%d)", caps.ResourcesAvailable, caps.ResourceCount)
	t.Logf("  Prompts available: %v (%d)", caps.PromptsAvailable, caps.PromptCount)
}

// TestAssertToolExists tests the AssertToolExists utility.
func TestAssertToolExists(t *testing.T) {
	serverCommand := os.Getenv("MCP_TEST_SERVER")
	if serverCommand == "" {
		t.Skip("Skipping integration test: MCP_TEST_SERVER not set")
	}

	config := TestServerConfig{
		ServerCommand: serverCommand,
		ClientInfo: protocol.ClientInfo{
			Name:    "integration-test-client",
			Version: "1.0.0",
		},
	}

	c, err := NewTestClient(config)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, list tools to find one to test
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) == 0 {
		t.Skip("No tools available to test")
	}

	toolName := tools[0].Name

	// Test without schema validation
	err = AssertToolExists(ctx, c, toolName, nil)
	if err != nil {
		t.Errorf("Tool should exist: %v", err)
	}

	// Test with schema validation (if tool has schema)
	if tools[0].Schema.Type != "" {
		err = AssertToolExists(ctx, c, toolName, &tools[0].Schema)
		if err != nil {
			t.Errorf("Tool schema validation failed: %v", err)
		}
	}
}

// TestToolExecution tests the TestToolExecution utility.
func TestToolExecution(t *testing.T) {
	serverCommand := os.Getenv("MCP_TEST_SERVER")
	if serverCommand == "" {
		t.Skip("Skipping integration test: MCP_TEST_SERVER not set")
	}

	config := TestServerConfig{
		ServerCommand: serverCommand,
		ClientInfo: protocol.ClientInfo{
			Name:    "integration-test-client",
			Version: "1.0.0",
		},
	}

	c, err := NewTestClient(config)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// List tools first
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("Failed to list tools: %v", err)
	}

	if len(tools) == 0 {
		t.Skip("No tools available to test")
	}

	toolName := tools[0].Name
	args := map[string]interface{}{}

	result, err := TestToolExecution(ctx, c, toolName, args)
	if err != nil {
		// Some tools may require specific args
		t.Logf("Tool execution failed (may require specific args): %v", err)
		return
	}

	if len(result) == 0 {
		t.Log("Tool execution succeeded but returned no results")
		return
	}

	t.Logf("Tool execution succeeded, got %d result(s)", len(result))
	for i, content := range result {
		if content.Type == "" {
			t.Errorf("Result %d has empty type", i)
		}
		if content.Text == "" {
			t.Logf("Result %d has empty text (may be valid)", i)
		}
	}
}

// Example: Test with exarp-go server
// Set MCP_TEST_SERVER=/path/to/exarp-go/bin/exarp-go
// Set MCP_TEST_SERVER_ARGS="" (optional)

// Example: Test with devwisdom-go server
// Set MCP_TEST_SERVER=/path/to/devwisdom-go/devwisdom
// Set MCP_TEST_SERVER_ARGS="" (optional)
