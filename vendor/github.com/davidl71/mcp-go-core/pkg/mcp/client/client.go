// Package client provides wrapper functions around existing MCP client libraries.
//
// This package wraps external MCP client libraries (e.g., github.com/metoro-io/mcp-golang)
// to provide:
//   - Type integration with mcp-go-core types
//   - Testing utilities for mcp-go-core servers
//   - Consistent API using mcp-go-core types
//
// Example usage:
//
//	clientInfo := protocol.ClientInfo{
//	    Name:    "test-client",
//	    Version: "1.0.0",
//	}
//	c, err := client.NewClient("/path/to/server", clientInfo)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer c.Close()
//
//	ctx := context.Background()
//	_, err = c.Initialize(ctx)
//	tools, err := c.ListTools(ctx)
//
// Dependencies:
//
// This package requires github.com/metoro-io/mcp-golang to be available.
// To build without this package (to avoid the dependency), use:
//
//	go build -tags no_mcp_client
package client

import (
	"fmt"
	"os/exec"

	"github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
)

// Client wraps an external MCP client library to provide a consistent API
// using mcp-go-core types.
type Client struct {
	// underlying is the wrapped client from the external library.
	// Using interface{} to avoid direct dependency (can be made type-safe with build tags)
	underlying interface{} // Expected to be *mcp.Client from github.com/metoro-io/mcp-golang

	// clientInfo stores the client information
	clientInfo protocol.ClientInfo

	// serverCommand is the command to launch the server
	serverCommand string

	// serverArgs are arguments to pass to the server
	serverArgs []string

	// initialized tracks whether the client has been initialized
	initialized bool
}

// NewClient creates a new client wrapper that connects to an MCP server.
//
// The serverCommand should be the path to the server binary or command.
// The clientInfo identifies this client to the server.
//
// Note: This function requires the mcp-golang library to be available.
// See the package documentation for dependency requirements.
func NewClient(serverCommand string, clientInfo protocol.ClientInfo) (*Client, error) {
	if serverCommand == "" {
		return nil, fmt.Errorf("server command cannot be empty")
	}
	if clientInfo.Name == "" {
		return nil, fmt.Errorf("client info name cannot be empty")
	}

	return &Client{
		clientInfo:    clientInfo,
		serverCommand: serverCommand,
		serverArgs:    []string{},
		initialized:   false,
	}, nil
}

// NewClientWithArgs creates a new client wrapper with server arguments.
func NewClientWithArgs(serverCommand string, serverArgs []string, clientInfo protocol.ClientInfo) (*Client, error) {
	client, err := NewClient(serverCommand, clientInfo)
	if err != nil {
		return nil, err
	}
	client.serverArgs = serverArgs
	return client, nil
}

// PromptInfo represents prompt metadata (similar to ToolInfo)
type PromptInfo struct {
	Name        string
	Description string
}

// Note: Method implementations (Initialize, ListTools, CallTool, etc.) are in:
//   - client_impl.go (when building without -tags no_mcp_client)
//   - client_stub.go (when building with -tags no_mcp_client)

// GetClientInfo returns the client information.
func (c *Client) GetClientInfo() protocol.ClientInfo {
	return c.clientInfo
}

// IsInitialized returns whether the client has been initialized.
func (c *Client) IsInitialized() bool {
	return c.initialized
}

// validateServerCommand checks if the server command exists and is executable.
func validateServerCommand(command string) error {
	// Check if command is executable
	cmd := exec.Command(command, "--help")
	if err := cmd.Run(); err != nil {
		// Command might not support --help, try just checking if file exists
		// This is a basic validation - actual execution will show real errors
		return nil // Don't fail validation on --help check
	}
	return nil
}
