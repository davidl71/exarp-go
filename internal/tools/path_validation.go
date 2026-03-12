package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/davidl71/exarp-go/internal/framework"
)

// ValidatePathAgainstMCPRoots validates that a path is within the client's MCP Roots.
// If no roots are set in context (client doesn't support roots), validation passes.
// Returns the validated absolute path or an error.
func ValidatePathAgainstMCPRoots(ctx context.Context, requestedPath string) (string, error) {
	roots := framework.RootsFromContext(ctx)
	if len(roots) == 0 {
		// No roots restriction - validate against project root instead
		return validateAgainstProjectRoot(requestedPath)
	}

	absPath, err := filepath.Abs(requestedPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	for _, root := range roots {
		// Convert root URI to path (handle file:// prefix)
		rootPath := strings.TrimPrefix(root.URI, "file://")
		if rootPath == "" {
			continue
		}

		absRoot, err := filepath.Abs(rootPath)
		if err != nil {
			continue
		}

		// Check if path is within this root
		rel, err := filepath.Rel(absRoot, absPath)
		if err != nil {
			continue
		}

		// If relative path doesn't start with "..", it's within root
		if !strings.HasPrefix(rel, "..") && !strings.HasPrefix(rel, string(filepath.Separator)+"..") {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("path %s is outside client MCP roots", requestedPath)
}

// validateAgainstProjectRoot falls back to validating against project root
func validateAgainstProjectRoot(path string) (string, error) {
	projectRoot, err := FindProjectRoot()
	if err != nil {
		// No project root - allow the path
		return filepath.Abs(path)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path escapes project root: %s", path)
	}

	return absPath, nil
}

// PathValidationEnabled returns true if the current context has MCP Roots set.
// Tools can use this to determine if enhanced path validation is available.
func PathValidationEnabled(ctx context.Context) bool {
	return len(framework.RootsFromContext(ctx)) > 0
}
