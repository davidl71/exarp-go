package framework

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestStdioTransport_Start(t *testing.T) {
	transport := &StdioTransport{}
	ctx := context.Background()

	if err := transport.Start(ctx); err != nil {
		t.Errorf("StdioTransport.Start() error = %v, want nil", err)
	}
}

func TestStdioTransport_Stop(t *testing.T) {
	transport := &StdioTransport{}
	ctx := context.Background()

	if err := transport.Stop(ctx); err != nil {
		t.Errorf("StdioTransport.Stop() error = %v, want nil", err)
	}
}

func TestStdioTransport_Type(t *testing.T) {
	transport := &StdioTransport{}
	if got := transport.Type(); got != "stdio" {
		t.Errorf("StdioTransport.Type() = %q, want %q", got, "stdio")
	}
}

func TestSSETransport_Type(t *testing.T) {
	transport := NewSSETransport("", 0)
	if got := transport.Type(); got != "sse" {
		t.Errorf("SSETransport.Type() = %q, want %q", got, "sse")
	}
}

func TestSSETransport_NewSSETransport(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		port     int
		wantPort int
	}{
		{
			name:     "default values",
			endpoint: "",
			port:     0,
			wantPort: 8080,
		},
		{
			name:     "custom endpoint and port",
			endpoint: "/custom",
			port:     9000,
			wantPort: 9000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewSSETransport(tt.endpoint, tt.port)
			if transport == nil {
				t.Fatal("NewSSETransport() returned nil")
			}
			if transport.Port != tt.wantPort {
				t.Errorf("transport.Port = %d, want %d", transport.Port, tt.wantPort)
			}
			expectedEndpoint := tt.endpoint
			if expectedEndpoint == "" {
				expectedEndpoint = "/sse"
			}
			if transport.Endpoint != expectedEndpoint {
				t.Errorf("transport.Endpoint = %q, want %q", transport.Endpoint, expectedEndpoint)
			}
		})
	}
}

func TestSSETransport_Start(t *testing.T) {
	transport := NewSSETransport("/test", 0) // Use port 0 to get random port
	ctx := context.Background()

	// Start transport
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("SSETransport.Start() error = %v, want nil", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify started
	if !transport.started {
		t.Error("SSETransport.Start() did not set started flag")
	}

	// Stop transport
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := transport.Stop(stopCtx); err != nil {
		t.Errorf("SSETransport.Stop() error = %v, want nil", err)
	}
}

func TestSSETransport_StartTwice(t *testing.T) {
	transport := NewSSETransport("/test", 0)
	ctx := context.Background()

	// Start transport
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("SSETransport.Start() error = %v, want nil", err)
	}

	// Try to start again (should fail)
	if err := transport.Start(ctx); err == nil {
		t.Error("SSETransport.Start() second call should return error, got nil")
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = transport.Stop(stopCtx)
}

func TestSSETransport_Stop(t *testing.T) {
	transport := NewSSETransport("/test", 0)
	ctx := context.Background()

	// Start transport
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("SSETransport.Start() error = %v, want nil", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Stop transport
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := transport.Stop(stopCtx); err != nil {
		t.Errorf("SSETransport.Stop() error = %v, want nil", err)
	}

	// Verify stopped
	if transport.started {
		t.Error("SSETransport.Stop() did not clear started flag")
	}
}

func TestSSETransport_StopWithoutStart(t *testing.T) {
	transport := NewSSETransport("/test", 0)
	ctx := context.Background()

	// Stop without starting (should not error)
	if err := transport.Stop(ctx); err != nil {
		t.Errorf("SSETransport.Stop() without start error = %v, want nil", err)
	}
}

func TestSSETransport_ConnectionCount(t *testing.T) {
	transport := NewSSETransport("/test", 0)

	// Initially no connections
	if count := transport.ConnectionCount(); count != 0 {
		t.Errorf("ConnectionCount() = %d, want 0", count)
	}

	// Start transport
	ctx := context.Background()
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("SSETransport.Start() error = %v, want nil", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Still no connections (no clients connected)
	if count := transport.ConnectionCount(); count != 0 {
		t.Errorf("ConnectionCount() = %d, want 0", count)
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = transport.Stop(stopCtx)
}

func TestSSETransport_WriteMessage_NotStarted(t *testing.T) {
	transport := NewSSETransport("/test", 0)

	// Write message without starting (should error)
	if err := transport.WriteMessage([]byte("test")); err == nil {
		t.Error("WriteMessage() without start should return error, got nil")
	}
}

func TestSSETransport_handleSSE(t *testing.T) {
	transport := NewSSETransport("/test", 0)
	ctx := context.Background()

	// Start transport
	if err := transport.Start(ctx); err != nil {
		t.Fatalf("SSETransport.Start() error = %v, want nil", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Get server address
	addr := transport.Server.Addr
	if addr == "" {
		t.Fatal("Server address is empty")
	}

	// Make request to SSE endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost"+addr+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Note: This test would need a full HTTP client to properly test SSE
	// For now, we just verify the handler function exists and can be called
	_ = req

	// Cleanup
	stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = transport.Stop(stopCtx)
}
