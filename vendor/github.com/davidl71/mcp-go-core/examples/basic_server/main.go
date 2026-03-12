// Package main demonstrates a basic MCP server using mcp-go-core.
//
// This example shows how to:
//   - Create an MCP server
//   - Register tools, prompts, and resources
//   - Run the server with stdio transport
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/davidl71/mcp-go-core/pkg/mcp/cli"
	"github.com/davidl71/mcp-go-core/pkg/mcp/config"
	"github.com/davidl71/mcp-go-core/pkg/mcp/factory"
	"github.com/davidl71/mcp-go-core/pkg/mcp/framework"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

func main() {
	// Detect execution mode (CLI vs MCP server)
	if cli.IsTTY() {
		// CLI mode - run command line interface
		if err := runCLI(); err != nil {
			log.Fatalf("CLI error: %v", err)
		}
		return
	}

	// MCP server mode - run as stdio server
	if err := runMCPServer(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func runMCPServer() error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadBaseConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg.Name = "example-server"
	cfg.Version = "1.0.0"

	// Create server using factory
	server, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Register tools
	if err := registerTools(server); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Register prompts
	if err := registerPrompts(server); err != nil {
		return fmt.Errorf("failed to register prompts: %w", err)
	}

	// Register resources
	if err := registerResources(server); err != nil {
		return fmt.Errorf("failed to register resources: %w", err)
	}

	// Run server with stdio transport
	transport := &framework.StdioTransport{}
	return server.Run(ctx, transport)
}

func runCLI() error {
	// Parse command-line arguments
	args := cli.ParseArgs(os.Args[1:])

	// Load configuration
	cfg, err := config.LoadBaseConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg.Name = "example-server"
	cfg.Version = "1.0.0"

	// Create server
	server, err := factory.NewServerFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Register tools
	if err := registerTools(server); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Handle CLI commands
	switch args.Command {
	case "list":
		return listTools(server)
	case "call":
		return callTool(server, args)
	default:
		fmt.Println("Usage:")
		fmt.Println("  example-server list              - List all tools")
		fmt.Println("  example-server call <tool> <args> - Call a tool")
		return nil
	}
}

func registerTools(server framework.MCPServer) error {
	// Register a simple echo tool
	echoSchema := types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Message to echo",
			},
		},
		Required: []string{"message"},
	}

	echoHandler := func(ctx context.Context, args json.RawMessage) ([]types.TextContent, error) {
		var params map[string]interface{}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		message, ok := params["message"].(string)
		if !ok {
			return nil, fmt.Errorf("message parameter is required")
		}

		return []types.TextContent{
			{Type: "text", Text: fmt.Sprintf("Echo: %s", message)},
		}, nil
	}

	if err := server.RegisterTool("echo", "Echo a message back", echoSchema, echoHandler); err != nil {
		return fmt.Errorf("failed to register echo tool: %w", err)
	}

	// Register a math tool
	mathSchema := types.ToolSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
				"description": "Math operation to perform",
			},
			"a": map[string]interface{}{
				"type":        "number",
				"description": "First number",
			},
			"b": map[string]interface{}{
				"type":        "number",
				"description": "Second number",
			},
		},
		Required: []string{"operation", "a", "b"},
	}

	mathHandler := func(ctx context.Context, args json.RawMessage) ([]types.TextContent, error) {
		var params map[string]interface{}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, fmt.Errorf("failed to parse arguments: %w", err)
		}

		operation, _ := params["operation"].(string)
		a, _ := params["a"].(float64)
		b, _ := params["b"].(float64)

		var result float64
		switch operation {
		case "add":
			result = a + b
		case "subtract":
			result = a - b
		case "multiply":
			result = a * b
		case "divide":
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result = a / b
		default:
			return nil, fmt.Errorf("unknown operation: %s", operation)
		}

		return []types.TextContent{
			{Type: "text", Text: fmt.Sprintf("%.2f %s %.2f = %.2f", a, operation, b, result)},
		}, nil
	}

	if err := server.RegisterTool("math", "Perform math operations", mathSchema, mathHandler); err != nil {
		return fmt.Errorf("failed to register math tool: %w", err)
	}

	return nil
}

func registerPrompts(server framework.MCPServer) error {
	greetingHandler := func(ctx context.Context, args map[string]interface{}) (string, error) {
		name, _ := args["name"].(string)
		if name == "" {
			name = "World"
		}
		return fmt.Sprintf("Hello, %s! Welcome to the MCP server.", name), nil
	}

	if err := server.RegisterPrompt("greeting", "Generate a greeting", greetingHandler); err != nil {
		return fmt.Errorf("failed to register greeting prompt: %w", err)
	}

	return nil
}

func registerResources(server framework.MCPServer) error {
	infoHandler := func(ctx context.Context, uri string) ([]byte, string, error) {
		data := fmt.Sprintf("Resource URI: %s\nServer: example-server v1.0.0", uri)
		return []byte(data), "text/plain", nil
	}

	if err := server.RegisterResource(
		"example://info",
		"Server Information",
		"Information about the example server",
		"text/plain",
		infoHandler,
	); err != nil {
		return fmt.Errorf("failed to register info resource: %w", err)
	}

	return nil
}

func listTools(server framework.MCPServer) error {
	tools := server.ListTools()
	fmt.Printf("Available tools (%d):\n\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  %s - %s\n", tool.Name, tool.Description)
	}
	return nil
}

func callTool(server framework.MCPServer, args *cli.Args) error {
	if args.Subcommand == "" {
		return fmt.Errorf("tool name required")
	}

	toolName := args.Subcommand
	argsJSON := args.GetFlag("args", "")
	if argsJSON == "" {
		argsJSON = "{}"
	}

	var toolArgs map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &toolArgs); err != nil {
		return fmt.Errorf("invalid JSON arguments: %w", err)
	}

	argsBytes, err := json.Marshal(toolArgs)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}

	ctx := context.Background()
	result, err := server.CallTool(ctx, toolName, argsBytes)
	if err != nil {
		return fmt.Errorf("tool execution failed: %w", err)
	}

	for _, content := range result {
		fmt.Println(content.Text)
	}

	return nil
}
