package framework

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

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
//
// SSE transport requires an HTTP server to be set up separately.
// The transport manages the SSE connection lifecycle and message handling.
type SSETransport struct {
	// Server is the HTTP server that will handle SSE connections
	Server *http.Server

	// Endpoint is the path where SSE connections will be accepted
	Endpoint string

	// Port is the port number for the HTTP server (default: 8080)
	Port int

	// mu protects the server state
	mu sync.RWMutex

	// started indicates if the transport has been started
	started bool

	// connections tracks active SSE connections
	connections map[*http.Request]http.ResponseWriter
}

// NewSSETransport creates a new SSE transport with the given endpoint and port
func NewSSETransport(endpoint string, port int) *SSETransport {
	if endpoint == "" {
		endpoint = "/sse"
	}
	if port == 0 {
		port = 8080
	}

	return &SSETransport{
		Endpoint:    endpoint,
		Port:        port,
		connections: make(map[*http.Request]http.ResponseWriter),
	}
}

// Start initializes the SSE transport and starts the HTTP server
func (t *SSETransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.started {
		return fmt.Errorf("SSE transport already started")
	}

	// Create HTTP server if not already set
	if t.Server == nil {
		mux := http.NewServeMux()
		mux.HandleFunc(t.Endpoint, t.handleSSE)

		t.Server = &http.Server{
			Addr:    fmt.Sprintf(":%d", t.Port),
			Handler: mux,
		}
	}

	// Start server in a goroutine
	go func() {
		if err := t.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Log error (would need logger integration)
			_ = err
		}
	}()

	t.started = true
	return nil
}

// Stop shuts down the SSE transport and closes all connections
func (t *SSETransport) Stop(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.started {
		return nil
	}

	// Close all active connections
	for req, w := range t.connections {
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		// Connection will be closed when request context is cancelled
		_ = req
	}
	t.connections = make(map[*http.Request]http.ResponseWriter)

	// Shutdown HTTP server
	if t.Server != nil {
		if err := t.Server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown SSE transport server: %w", err)
		}
	}

	t.started = false
	return nil
}

// Type returns the transport type
func (t *SSETransport) Type() string {
	return "sse"
}

// handleSSE handles incoming SSE connection requests
func (t *SSETransport) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow CORS (can be configured)

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Register connection
	t.mu.Lock()
	t.connections[r] = w
	t.mu.Unlock()

	// Cleanup on disconnect
	defer func() {
		t.mu.Lock()
		delete(t.connections, r)
		t.mu.Unlock()
	}()

	// Send initial connection message
	fmt.Fprintf(w, "data: {\"type\":\"connection\",\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Keep connection alive and wait for context cancellation
	<-r.Context().Done()
}

// WriteMessage sends a message to all connected SSE clients
func (t *SSETransport) WriteMessage(data []byte) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.started {
		return fmt.Errorf("SSE transport not started")
	}

	message := fmt.Sprintf("data: %s\n\n", string(data))

	for req, w := range t.connections {
		// Check if connection is still alive
		select {
		case <-req.Context().Done():
			// Connection closed, skip
			continue
		default:
			// Write message
			if _, err := fmt.Fprintf(w, message); err != nil {
				// Connection error, will be cleaned up on next request
				continue
			}

			// Flush if possible
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}

	return nil
}

// ConnectionCount returns the number of active SSE connections
func (t *SSETransport) ConnectionCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.connections)
}
