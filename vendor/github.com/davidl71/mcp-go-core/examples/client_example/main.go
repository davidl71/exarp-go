// Package main demonstrates using the mcp-go-core client wrapper.
//
// This example shows how to:
//   - Create a client wrapper
//   - Initialize a connection to an MCP server
//   - List tools, resources, and prompts
//   - Call tools on the server
//
// To run this example:
//
//	go run examples/client_example/main.go /path/to/mcp/server
//
// Or build and run:
//
//	go build -o client-example examples/client_example/main.go
//	./client-example /path/to/mcp/server
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/davidl71/mcp-go-core/pkg/mcp/client"
	"github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
)

func main() {
	// Parse command line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <server-command>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s ./bin/my-server\n", os.Args[0])
	}

	var timeout = flag.Duration("timeout", 10*time.Second, "Request timeout")
	var clientName = flag.String("client-name", "example-client", "Client name")
	var clientVersion = flag.String("client-version", "1.0.0", "Client version")

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	serverCommand := flag.Arg(0)

	// Create client
	clientInfo := protocol.ClientInfo{
		Name:    *clientName,
		Version: *clientVersion,
	}

	fmt.Printf("Creating client wrapper for server: %s\n", serverCommand)
	c, err := client.NewClient(serverCommand, clientInfo)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer c.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Initialize client
	fmt.Println("\n1. Initializing client...")
	initResult, err := c.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	fmt.Printf("   Connected to server: %s v%s\n", initResult.ServerInfo.Name, initResult.ServerInfo.Version)
	fmt.Printf("   Protocol version: %s\n", initResult.ProtocolVersion)

	// List tools
	fmt.Println("\n2. Listing tools...")
	tools, err := c.ListTools(ctx)
	if err != nil {
		log.Printf("   Warning: Failed to list tools: %v", err)
	} else {
		fmt.Printf("   Found %d tool(s):\n", len(tools))
		for i, tool := range tools {
			fmt.Printf("   %d. %s\n", i+1, tool.Name)
			if tool.Description != "" {
				fmt.Printf("      Description: %s\n", tool.Description)
			}
			if tool.Schema.Type != "" {
				fmt.Printf("      Schema type: %s\n", tool.Schema.Type)
			}
		}
	}

	// Try calling a tool (if available)
	if len(tools) > 0 {
		fmt.Println("\n3. Calling a tool...")
		toolName := tools[0].Name
		fmt.Printf("   Calling tool: %s\n", toolName)

		// Use minimal arguments (tool may have optional parameters)
		args := map[string]interface{}{}

		result, err := c.CallTool(ctx, toolName, args)
		if err != nil {
			fmt.Printf("   Warning: Tool call failed (may require specific args): %v\n", err)
		} else {
			fmt.Printf("   Tool call succeeded, got %d result(s):\n", len(result))
			for i, content := range result {
				fmt.Printf("   Result %d:\n", i+1)
				fmt.Printf("      Type: %s\n", content.Type)
				if content.Text != "" {
					// Truncate long text
					text := content.Text
					if len(text) > 200 {
						text = text[:200] + "... (truncated)"
					}
					fmt.Printf("      Text: %s\n", text)
				}
			}
		}
	}

	// List resources
	fmt.Println("\n4. Listing resources...")
	resources, err := c.ListResources(ctx)
	if err != nil {
		fmt.Printf("   Warning: Failed to list resources: %v\n", err)
	} else {
		fmt.Printf("   Found %d resource(s):\n", len(resources))
		for i, resource := range resources {
			fmt.Printf("   %d. %s\n", i+1, resource.URI)
			if resource.Name != "" {
				fmt.Printf("      Name: %s\n", resource.Name)
			}
			if resource.Description != "" {
				fmt.Printf("      Description: %s\n", resource.Description)
			}
		}
	}

	// List prompts
	fmt.Println("\n5. Listing prompts...")
	prompts, err := c.ListPrompts(ctx)
	if err != nil {
		fmt.Printf("   Warning: Failed to list prompts: %v\n", err)
	} else {
		fmt.Printf("   Found %d prompt(s):\n", len(prompts))
		for i, prompt := range prompts {
			fmt.Printf("   %d. %s\n", i+1, prompt.Name)
			if prompt.Description != "" {
				fmt.Printf("      Description: %s\n", prompt.Description)
			}
		}
	}

	// Test server capabilities
	fmt.Println("\n6. Testing server capabilities...")
	caps, err := client.TestServerCapabilities(ctx, c)
	if err != nil {
		fmt.Printf("   Warning: Failed to test capabilities: %v\n", err)
	} else {
		fmt.Printf("   Tools available: %v (%d tools)\n", caps.ToolsAvailable, caps.ToolCount)
		fmt.Printf("   Resources available: %v (%d resources)\n", caps.ResourcesAvailable, caps.ResourceCount)
		fmt.Printf("   Prompts available: %v (%d prompts)\n", caps.PromptsAvailable, caps.PromptCount)
	}

	fmt.Println("\nDone!")
}
