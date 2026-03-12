// +build no_mcp_client

// Package client stub implementation when client wrapper is disabled.
//
// This file provides stub implementations that return errors indicating
// the client wrapper is not available. It is compiled when building with
// the "no_mcp_client" build tag.
//
// To build without the client wrapper (to avoid the dependency):
//
//	go build -tags no_mcp_client

package client

import (
	"context"
	"fmt"

	"github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// Initialize returns an error indicating the client wrapper is not available.
func (c *Client) Initialize(ctx context.Context) (*protocol.InitializeResult, error) {
	return nil, fmt.Errorf("client wrapper not available: build without -tags no_mcp_client and ensure github.com/metoro-io/mcp-golang is installed")
}

// ListTools returns an error indicating the client wrapper is not available.
func (c *Client) ListTools(ctx context.Context) ([]types.ToolInfo, error) {
	return nil, fmt.Errorf("client wrapper not available: build without -tags no_mcp_client and ensure github.com/metoro-io/mcp-golang is installed")
}

// CallTool returns an error indicating the client wrapper is not available.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) ([]types.TextContent, error) {
	return nil, fmt.Errorf("client wrapper not available: build without -tags no_mcp_client and ensure github.com/metoro-io/mcp-golang is installed")
}

// ListResources returns an error indicating the client wrapper is not available.
func (c *Client) ListResources(ctx context.Context) ([]protocol.Resource, error) {
	return nil, fmt.Errorf("client wrapper not available: build without -tags no_mcp_client and ensure github.com/metoro-io/mcp-golang is installed")
}

// ReadResource returns an error indicating the client wrapper is not available.
func (c *Client) ReadResource(ctx context.Context, uri string) ([]byte, string, error) {
	return nil, "", fmt.Errorf("client wrapper not available: build without -tags no_mcp_client and ensure github.com/metoro-io/mcp-golang is installed")
}

// ListPrompts returns an error indicating the client wrapper is not available.
func (c *Client) ListPrompts(ctx context.Context) ([]PromptInfo, error) {
	return nil, fmt.Errorf("client wrapper not available: build without -tags no_mcp_client and ensure github.com/metoro-io/mcp-golang is installed")
}

// GetPrompt returns an error indicating the client wrapper is not available.
func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	return "", fmt.Errorf("client wrapper not available: build without -tags no_mcp_client and ensure github.com/metoro-io/mcp-golang is installed")
}
