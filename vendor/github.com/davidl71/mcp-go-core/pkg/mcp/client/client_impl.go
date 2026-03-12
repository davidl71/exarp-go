//go:build !no_mcp_client
// +build !no_mcp_client

// Package client implementation using github.com/metoro-io/mcp-golang
//
// This file contains the actual implementation using the external client library.
// To use this package, ensure github.com/metoro-io/mcp-golang is available:
//
//	go get github.com/metoro-io/mcp-golang
//
// To build without the client wrapper (to avoid the dependency), use:
//
//	go build -tags no_mcp_client

package client

import (
	"context"
	"encoding/base64"
	"fmt"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"

	"github.com/davidl71/mcp-go-core/pkg/mcp/protocol"
	"github.com/davidl71/mcp-go-core/pkg/mcp/types"
)

// initUnderlyingClient creates and initializes the underlying mcp-golang client.
func (c *Client) initUnderlyingClient() error {
	if c.underlying != nil {
		return nil // Already initialized
	}

	// Create stdio transport
	transport := stdio.NewStdioServerTransport()

	// Create underlying client
	underlyingClient := mcp.NewClient(transport)

	c.underlying = underlyingClient
	return nil
}

// Initialize initializes the client session with the MCP server.
func (c *Client) Initialize(ctx context.Context) (*protocol.InitializeResult, error) {
	if err := c.initUnderlyingClient(); err != nil {
		return nil, fmt.Errorf("failed to initialize underlying client: %w", err)
	}

	client := c.underlying.(*mcp.Client)

	// Call underlying Initialize
	// Note: mcp-golang's Initialize may take different parameters
	// This is a placeholder based on the expected API
	response, err := client.Initialize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}

	// Convert response to protocol.InitializeResult
	result := &protocol.InitializeResult{
		ProtocolVersion: response.ProtocolVersion,
		Capabilities: protocol.ServerCapabilities{
			Tools:     convertToolsCapability(response.Capabilities),
			Resources: convertResourcesCapability(response.Capabilities),
		},
		ServerInfo: protocol.ServerInfo{
			Name:    response.ServerInfo.Name,
			Version: response.ServerInfo.Version,
		},
	}

	c.initialized = true
	return result, nil
}

// ListTools lists all available tools from the server.
func (c *Client) ListTools(ctx context.Context) ([]types.ToolInfo, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client must be initialized before listing tools")
	}

	client := c.underlying.(*mcp.Client)

	// Call underlying ListTools
	toolsResponse, err := client.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	// Convert to mcp-go-core types
	tools := make([]types.ToolInfo, 0, len(toolsResponse.Tools))
	for _, tool := range toolsResponse.Tools {
		converted, err := ConvertExternalToolToToolInfo(tool)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool %q: %w", tool.Name, err)
		}
		tools = append(tools, converted)
	}

	return tools, nil
}

// CallTool calls a tool on the server with the given arguments.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) ([]types.TextContent, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client must be initialized before calling tools")
	}
	if name == "" {
		return nil, fmt.Errorf("tool name cannot be empty")
	}

	client := c.underlying.(*mcp.Client)

	// Call underlying CallTool
	response, err := client.CallTool(ctx, name, args)
	if err != nil {
		return nil, fmt.Errorf("failed to call tool %q: %w", name, err)
	}

	// Convert response content
	result := make([]types.TextContent, 0, len(response.Content))
	for _, content := range response.Content {
		if content.TextContent != nil {
			result = append(result, types.TextContent{
				Type: "text",
				Text: content.TextContent.Text,
			})
		}
	}

	return result, nil
}

// ListResources lists all available resources from the server.
func (c *Client) ListResources(ctx context.Context) ([]protocol.Resource, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client must be initialized before listing resources")
	}

	client := c.underlying.(*mcp.Client)

	// Call underlying ListResources
	resourcesResponse, err := client.ListResources(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	// Convert to protocol.Resource
	resources := make([]protocol.Resource, 0, len(resourcesResponse.Resources))
	for _, resource := range resourcesResponse.Resources {
		desc := ""
		if resource.Description != nil {
			desc = *resource.Description
		}
		mimeType := ""
		if resource.MimeType != nil {
			mimeType = *resource.MimeType
		}
		resources = append(resources, protocol.Resource{
			URI:         resource.Uri,
			Name:        resource.Name,
			Description: desc,
			MimeType:    mimeType,
		})
	}

	return resources, nil
}

// ReadResource reads a resource from the server by URI.
func (c *Client) ReadResource(ctx context.Context, uri string) ([]byte, string, error) {
	if !c.initialized {
		return nil, "", fmt.Errorf("client must be initialized before reading resources")
	}
	if uri == "" {
		return nil, "", fmt.Errorf("resource URI cannot be empty")
	}

	client := c.underlying.(*mcp.Client)

	// Call underlying ReadResource
	resource, err := client.ReadResource(ctx, uri)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read resource %q: %w", uri, err)
	}

	if resource == nil || len(resource.Contents) == 0 {
		return nil, "", fmt.Errorf("resource response contained no contents")
	}
	content := resource.Contents[0]
	if content.TextResourceContents != nil {
		mimeType := ""
		if content.TextResourceContents.MimeType != nil {
			mimeType = *content.TextResourceContents.MimeType
		}
		return []byte(content.TextResourceContents.Text), mimeType, nil
	}
	if content.BlobResourceContents != nil {
		mimeType := ""
		if content.BlobResourceContents.MimeType != nil {
			mimeType = *content.BlobResourceContents.MimeType
		}
		data, err := base64.StdEncoding.DecodeString(content.BlobResourceContents.Blob)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode blob resource %q: %w", uri, err)
		}
		return data, mimeType, nil
	}
	return nil, "", fmt.Errorf("resource response contained unsupported content type")
}

// ListPrompts lists all available prompts from the server.
func (c *Client) ListPrompts(ctx context.Context) ([]PromptInfo, error) {
	if !c.initialized {
		return nil, fmt.Errorf("client must be initialized before listing prompts")
	}

	client := c.underlying.(*mcp.Client)

	// Call underlying ListPrompts
	promptsResponse, err := client.ListPrompts(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	// Convert to PromptInfo
	prompts := make([]PromptInfo, 0, len(promptsResponse.Prompts))
	for _, prompt := range promptsResponse.Prompts {
		desc := ""
		if prompt.Description != nil {
			desc = *prompt.Description
		}
		prompts = append(prompts, PromptInfo{
			Name:        prompt.Name,
			Description: desc,
		})
	}

	return prompts, nil
}

// GetPrompt gets a prompt template from the server.
func (c *Client) GetPrompt(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	if !c.initialized {
		return "", fmt.Errorf("client must be initialized before getting prompts")
	}
	if name == "" {
		return "", fmt.Errorf("prompt name cannot be empty")
	}

	client := c.underlying.(*mcp.Client)

	// Call underlying GetPrompt
	response, err := client.GetPrompt(ctx, name, args)
	if err != nil {
		return "", fmt.Errorf("failed to get prompt %q: %w", name, err)
	}

	// Extract text from messages
	if len(response.Messages) > 0 {
		if response.Messages[0].Content.TextContent != nil {
			return response.Messages[0].Content.TextContent.Text, nil
		}
	}

	return "", fmt.Errorf("prompt response had no text content")
}

// Close closes the client connection and cleans up resources.
func (c *Client) Close() error {
	if c.underlying == nil {
		return nil
	}

	// The underlying client may have a Close method
	// For stdio transport, cleanup is typically automatic
	c.underlying = nil
	c.initialized = false
	return nil
}

// Helper functions for capability conversion

func convertToolsCapability(capabilities interface{}) *protocol.ToolsCapability {
	// Check if tools capability exists in response
	// This is a placeholder - actual implementation depends on mcp-golang's capability structure
	return &protocol.ToolsCapability{}
}

func convertResourcesCapability(capabilities interface{}) *protocol.ResourcesCapability {
	// Check if resources capability exists in response
	// This is a placeholder - actual implementation depends on mcp-golang's capability structure
	return &protocol.ResourcesCapability{}
}
