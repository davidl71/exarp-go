package framework

import "context"

// Transport abstracts transport mechanism for MCP servers
type Transport interface {
	// Start initializes the transport
	Start(ctx context.Context) error

	// Stop shuts down the transport
	Stop(ctx context.Context) error

	// Type returns the transport type (stdio, sse, http, etc.)
	Type() string
}

// StdioTransport represents standard I/O transport
// This is the default transport for MCP servers using stdin/stdout
type StdioTransport struct{}

// Start initializes the stdio transport
func (t *StdioTransport) Start(ctx context.Context) error {
	// Stdio transport doesn't need explicit initialization
	return nil
}

// Stop shuts down the stdio transport
func (t *StdioTransport) Stop(ctx context.Context) error {
	// Stdio transport doesn't need explicit cleanup
	return nil
}

// Type returns the transport type
func (t *StdioTransport) Type() string {
	return "stdio"
}

// SSETransport represents Server-Sent Events transport
// This transport is used for HTTP-based MCP servers
type SSETransport struct{}

// Start initializes the SSE transport
func (t *SSETransport) Start(ctx context.Context) error {
	// SSE transport initialization would be implemented here
	// For now, this is a placeholder
	return nil
}

// Stop shuts down the SSE transport
func (t *SSETransport) Stop(ctx context.Context) error {
	// SSE transport cleanup would be implemented here
	// For now, this is a placeholder
	return nil
}

// Type returns the transport type
func (t *SSETransport) Type() string {
	return "sse"
}
