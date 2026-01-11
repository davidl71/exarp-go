package bridge

import (
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

	// Marshal arguments to JSON
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Create command
	cmd := exec.CommandContext(ctx, "python3", bridgeScript, toolName, string(argsJSON))

	// Pass environment variables to Python subprocess (especially PROJECT_ROOT from Cursor)
	cmd.Env = os.Environ()

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Execute command
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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, "application/json", fmt.Errorf("python resource execution failed: %w, output: %s", err, output)
	}

	// Return as JSON
	return output, "application/json", nil
}
