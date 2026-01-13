package bridge

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/davidl71/exarp-go/internal/security"
)

const (
	// Default idle timeout before restarting process
	defaultIdleTimeout = 5 * time.Minute

	// Default health check interval (every N calls)
	defaultHealthCheckInterval = 10

	// Default request timeout
	defaultRequestTimeout = 30 * time.Second
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PythonProcessPool manages a persistent Python process for tool execution
type PythonProcessPool struct {
	mu                    sync.RWMutex
	process               *exec.Cmd
	stdin                 io.WriteCloser
	stdout                io.ReadCloser
	scanner               *bufio.Scanner
	lastUsed              time.Time
	created               time.Time
	healthy               bool
	requestCount          int
	idleTimeout           time.Duration
	healthCheckInterval   int
	workspaceRoot         string
	bridgeScript          string
}

var (
	globalPool     *PythonProcessPool
	globalPoolOnce sync.Once
	poolEnabled    = os.Getenv("EXARP_PYTHON_POOL_ENABLED") != "false" // Default: enabled
)

// GetGlobalPool returns the global Python process pool singleton
func GetGlobalPool() *PythonProcessPool {
	globalPoolOnce.Do(func() {
		globalPool = NewPythonProcessPool()
	})
	return globalPool
}

// NewPythonProcessPool creates a new Python process pool
func NewPythonProcessPool() *PythonProcessPool {
	workspaceRoot := getWorkspaceRoot()
	bridgeScript := filepath.Join(workspaceRoot, "bridge", "execute_tool_daemon.py")

	return &PythonProcessPool{
		healthy:             false,
		workspaceRoot:       workspaceRoot,
		bridgeScript:        bridgeScript,
		idleTimeout:         defaultIdleTimeout,
		healthCheckInterval: defaultHealthCheckInterval,
	}
}

// ExecuteTool executes a Python tool via the persistent process pool
func (p *PythonProcessPool) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	if !poolEnabled {
		return "", fmt.Errorf("pool disabled")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Ensure process is healthy
	if err := p.ensureHealthy(ctx); err != nil {
		return "", fmt.Errorf("failed to ensure healthy process: %w", err)
	}

	// Check idle timeout
	if time.Since(p.lastUsed) > p.idleTimeout {
		// Restart process if idle too long
		if err := p.restart(ctx); err != nil {
			return "", fmt.Errorf("failed to restart idle process: %w", err)
		}
	}

	// Increment request count
	p.requestCount++
	p.lastUsed = time.Now()

	// Generate request ID
	requestID := fmt.Sprintf("req-%d-%d", time.Now().UnixNano(), p.requestCount)

	// Create JSON-RPC request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      requestID,
		Method:  "execute_tool",
		Params: map[string]interface{}{
			"tool_name": toolName,
			"args":       args,
		},
	}

	// Marshal request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Write request to stdin (line-delimited)
	if _, err := fmt.Fprintf(p.stdin, "%s\n", reqJSON); err != nil {
		// Process may have died - restart and retry once
		if err := p.restart(ctx); err != nil {
			return "", fmt.Errorf("failed to write request (process dead): %w", err)
		}
		// Retry once after restart
		if _, err := fmt.Fprintf(p.stdin, "%s\n", reqJSON); err != nil {
			return "", fmt.Errorf("failed to write request after restart: %w", err)
		}
	}

	// Read response from stdout (line-delimited)
	if !p.scanner.Scan() {
		// Scanner failed - process may have died
		if err := p.restart(ctx); err != nil {
			return "", fmt.Errorf("failed to read response (process dead): %w", err)
		}
		// Retry once after restart
		if !p.scanner.Scan() {
			return "", fmt.Errorf("failed to read response after restart: %w", err)
		}
	}

	responseLine := p.scanner.Text()
	if err := p.scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON-RPC response
	var resp JSONRPCResponse
	if err := json.Unmarshal([]byte(responseLine), &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for JSON-RPC error
	if resp.Error != nil {
		return "", fmt.Errorf("python tool error [%d]: %s: %v", resp.Error.Code, resp.Error.Message, resp.Error.Data)
	}

	// Extract result
	if resp.Result == nil {
		return "", fmt.Errorf("response missing result")
	}

	// Result is already a JSON string from Python daemon
	// If it's a string, return it directly; otherwise marshal it
	switch v := resp.Result.(type) {
	case string:
		return v, nil
	default:
		// If result is not a string, marshal it to JSON
		resultJSON, err := json.Marshal(resp.Result)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %w", err)
		}
		return string(resultJSON), nil
	}
}

// ensureHealthy ensures the process is running and healthy
func (p *PythonProcessPool) ensureHealthy(ctx context.Context) error {
	// Check if process exists and is running
	if p.process == nil || p.process.Process == nil {
		return p.restart(ctx)
	}

	// Check process state
	if p.process.ProcessState != nil && p.process.ProcessState.Exited() {
		// Process has exited - restart
		return p.restart(ctx)
	}

	// If not healthy, do a health check
	if !p.healthy {
		// Simple health check: try a ping request
		// For now, just mark as healthy if process is running
		// TODO: Add actual ping/health check request
		p.healthy = true
	}

	return nil
}

// restart restarts the Python process
func (p *PythonProcessPool) restart(ctx context.Context) error {
	// Close existing process if any
	if p.process != nil {
		if p.stdin != nil {
			_ = p.stdin.Close()
		}
		if p.stdout != nil {
			_ = p.stdout.Close()
		}
		if p.process.Process != nil {
			_ = p.process.Process.Kill()
		}
		_ = p.process.Wait()
	}

	// Validate workspace root
	validatedRoot, err := security.ValidatePath(p.workspaceRoot, p.workspaceRoot)
	if err != nil {
		return fmt.Errorf("invalid workspace root: %w", err)
	}

	// Validate bridge script path
	_, err = security.ValidatePath(p.bridgeScript, validatedRoot)
	if err != nil {
		return fmt.Errorf("invalid bridge script path: %w", err)
	}

	// Create new process
	cmd := exec.CommandContext(ctx, "python3", p.bridgeScript)
	cmd.Dir = validatedRoot
	cmd.Env = append(os.Environ(), fmt.Sprintf("PROJECT_ROOT=%s", validatedRoot))

	// Set up stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return fmt.Errorf("failed to start Python process: %w", err)
	}

	// Store process and pipes
	p.process = cmd
	p.stdin = stdin
	p.stdout = stdout
	p.scanner = bufio.NewScanner(stdout)
	p.created = time.Now()
	p.lastUsed = time.Now()
	p.healthy = true
	p.requestCount = 0

	return nil
}

// Close closes the process pool
func (p *PythonProcessPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.process == nil {
		return nil
	}

	var errs []error

	if p.stdin != nil {
		if err := p.stdin.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close stdin: %w", err))
		}
	}

	if p.stdout != nil {
		if err := p.stdout.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close stdout: %w", err))
		}
	}

	if p.process.Process != nil {
		if err := p.process.Process.Kill(); err != nil {
			errs = append(errs, fmt.Errorf("failed to kill process: %w", err))
		}
	}

	if err := p.process.Wait(); err != nil {
		errs = append(errs, fmt.Errorf("failed to wait for process: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing pool: %v", errs)
	}

	return nil
}
