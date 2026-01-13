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

// ExecutePythonTool executes a Python tool via subprocess
// Supports both protobuf binary and JSON formats for backward compatibility
func ExecutePythonTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error) {
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
	
	// Try protobuf mode first
	cmd := exec.CommandContext(ctxWithTimeout, "python3", bridgeScript, toolName, "--protobuf")
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
	cmd = exec.CommandContext(ctxWithTimeout, "python3", bridgeScript, toolName, string(argsJSON))
	cmd.Dir = workspaceRoot
	cmd.Env = append(os.Environ(), fmt.Sprintf("PROJECT_ROOT=%s", workspaceRoot))

	// Execute command (JSON fallback)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("python tool execution failed: %w, output: %s", err, output)
	}

	return string(output), nil
}

// ExecutePythonResource executes a Python resource handler via subprocess
func ExecutePythonResource(ctx context.Context, uri string) ([]byte, string, error) {
	// Get workspace root where bridge scripts are located
	workspaceRoot := getWorkspaceRoot()

	// Validate workspace root path
	validatedRoot, err := security.ValidatePath(workspaceRoot, workspaceRoot)
	if err != nil {
		return nil, "", fmt.Errorf("invalid workspace root: %w", err)
	}
	workspaceRoot = validatedRoot

	// Path to Python bridge script for resources (bridge scripts are in workspace root)
	bridgeScript := filepath.Join(workspaceRoot, "bridge", "execute_resource.py")

	// Validate bridge script path is within workspace root
	_, err = security.ValidatePath(bridgeScript, workspaceRoot)
	if err != nil {
		return nil, "", fmt.Errorf("invalid bridge script path: %w", err)
	}

	// Create command
	cmd := exec.CommandContext(ctx, "python3", bridgeScript, uri)

	// Pass environment variables to Python subprocess (especially PROJECT_ROOT from Cursor)
	cmd.Env = os.Environ()

	// Set timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctxWithTimeout, "python3", bridgeScript, uri)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "application/json", fmt.Errorf("python resource execution failed: %w, output: %s", err, output)
	}

	// Return as JSON
	return output, "application/json", nil
}
