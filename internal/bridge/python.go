package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/davidl71/exarp-go/internal/security"
	"github.com/davidl71/exarp-go/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// getPythonCommand returns the executable and leading args for running Python.
// Requires uv (uv run python); project assumes uv is always used.
func getPythonCommand() (executable string, runArgs []string, err error) {
	path, lookErr := exec.LookPath("uv")
	if lookErr != nil || path == "" {
		return "", nil, fmt.Errorf("uv not found: install uv (https://docs.astral.sh/uv/) or add it to PATH for Python bridge")
	}
	return path, []string{"run", "python"}, nil
}

// getWorkspaceRoot determines the workspace root where bridge scripts are located
func getWorkspaceRoot() string {
	// Check environment variable first
	projectRoot := os.Getenv("PROJECT_ROOT")

	// If PROJECT_ROOT contains placeholder (not substituted), ignore it
	if strings.Contains(projectRoot, "{{PROJECT_ROOT}}") || projectRoot == "" {
		// Try to find workspace root relative to executable location
		// Get the executable path
		execPath, err := os.Executable()
		if err == nil {
			// If executable is in bin/, workspace root is parent directory
			execDir := filepath.Dir(execPath)
			if filepath.Base(execDir) == "bin" {
				return filepath.Dir(execDir)
			}
			// Otherwise, assume executable is in workspace root
			return execDir
		}

		// Fallback: try to find workspace root using __file__ equivalent
		// In Go, we can use runtime.Caller to find the source file location
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			// internal/bridge/python.go -> workspace root
			return filepath.Join(filepath.Dir(filename), "..", "..")
		}
	}

	// If PROJECT_ROOT was set and valid, use it
	if projectRoot != "" && !strings.Contains(projectRoot, "{{") {
		return projectRoot
	}

	// Last resort: assume current directory or relative path
	return "."
}

// ExecutePythonTool executes a Python tool via subprocess (execute_tool.py).
// Supports both protobuf binary and JSON formats for backward compatibility.
// Pool/daemon (execute_tool_daemon.py) removed 2026-01-29; one-shot subprocess only.
func ExecutePythonTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	return executePythonToolSubprocess(ctx, toolName, args)
}

// executePythonToolSubprocess executes a Python tool via subprocess
// This is the original implementation, kept for backward compatibility
func executePythonToolSubprocess(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
	// Get workspace root where bridge scripts are located
	workspaceRoot := getWorkspaceRoot()

	// Validate workspace root path
	validatedRoot, err := security.ValidatePath(workspaceRoot, workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("invalid workspace root: %w", err)
	}
	workspaceRoot = validatedRoot

	// Path to Python bridge script (bridge scripts are in workspace root)
	bridgeScript := filepath.Join(workspaceRoot, "bridge", "execute_tool.py")

	// Validate bridge script path is within workspace root
	_, err = security.ValidatePath(bridgeScript, workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("invalid bridge script path: %w", err)
	}

	// Marshal arguments to JSON (for protobuf message)
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Generate request ID for tracking
	requestID := fmt.Sprintf("%s-%d", toolName, time.Now().UnixNano())

	// Create protobuf ToolRequest
	req := &proto.ToolRequest{
		ToolName:       toolName,
		ArgumentsJson:  string(argsJSON), // JSON string for now (can be replaced with protobuf later)
		ProjectRoot:    workspaceRoot,
		TimeoutSeconds: 30,
		RequestId:      requestID,
	}

	// Marshal protobuf to binary
	protobufData, err := protobuf.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal protobuf request: %w", err)
	}

	// Try protobuf format first (if Python protobuf code is available)
	// Fall back to JSON format for backward compatibility
	var stdout, stderr bytes.Buffer

	// Set timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pyExec, pyRunArgs, err := getPythonCommand()
	if err != nil {
		return "", err
	}
	cmdArgs := append(append(pyRunArgs, bridgeScript, toolName), "--protobuf")
	// Try protobuf mode first
	cmd := exec.CommandContext(ctxWithTimeout, pyExec, cmdArgs...)
	cmd.Stdin = bytes.NewReader(protobufData)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = workspaceRoot
	cmd.Env = append(os.Environ(), fmt.Sprintf("PROJECT_ROOT=%s", workspaceRoot))

	err = cmd.Run()
	if err == nil {
		// Protobuf mode succeeded - try to parse protobuf ToolResponse
		outputBytes := stdout.Bytes()

		// Try to parse as protobuf ToolResponse
		var resp proto.ToolResponse
		if err := protobuf.Unmarshal(outputBytes, &resp); err == nil {
			// Successfully parsed protobuf response
			if !resp.Success {
				// Tool execution failed
				return "", fmt.Errorf("python tool failed: %s", resp.Error)
			}
			// Return the result (JSON string from protobuf message)
			return resp.Result, nil
		}

		// If protobuf parsing failed, treat as JSON (backward compatibility)
		return string(outputBytes), nil
	}

	// Protobuf mode failed (likely Python protobuf code not available or error)
	// Fall back to JSON format (backward compatible)
	cmd = exec.CommandContext(ctxWithTimeout, pyExec, append(append(pyRunArgs, bridgeScript, toolName), string(argsJSON))...)
	cmd.Dir = workspaceRoot
	cmd.Env = append(os.Environ(), fmt.Sprintf("PROJECT_ROOT=%s", workspaceRoot))

	// Execute command (JSON fallback)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("python tool execution failed: %w, output: %s", err, output)
	}

	return string(output), nil
}
