// Package security provides security utilities for MCP servers.
//
// This package includes path validation, project root detection, access control,
// and rate limiting functionality. All security utilities are designed to be
// framework-agnostic and reusable across different MCP server implementations.
//
// Example:
//
//	root, err := security.GetProjectRoot(".")
//	if err != nil {
//		log.Fatal(err)
//	}
//	validPath, err := security.ValidatePath("subdir/file.txt", root)
//	if err != nil {
//		log.Fatal(err)
//	}
package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetProjectRoot attempts to find the project root by looking for go.mod
// Returns the project root path or an error
func GetProjectRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve start path: %w", err)
	}

	currentPath := absPath
	for {
		// Check if go.mod exists in current directory
		goModPath := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return currentPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			// Reached filesystem root without finding go.mod
			return "", fmt.Errorf("project root (go.mod) not found")
		}
		currentPath = parent
	}
}

// ValidatePath ensures a path is within the project root and prevents directory traversal
// Returns the cleaned absolute path if valid, or an error if the path is invalid
func ValidatePath(path, projectRoot string) (string, error) {
	if projectRoot == "" {
		return "", fmt.Errorf("project root cannot be empty")
	}

	// Clean and make project root absolute
	projectRoot = filepath.Clean(projectRoot)
	absProjectRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to resolve project root: %w", err)
	}

	// Ensure project root exists and is a directory
	info, err := os.Stat(absProjectRoot)
	if err != nil {
		return "", fmt.Errorf("project root does not exist: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("project root is not a directory")
	}

	// Clean the input path
	cleanPath := filepath.Clean(path)

	// Resolve to absolute path
	var absPath string
	if filepath.IsAbs(cleanPath) {
		absPath = cleanPath
	} else {
		// Join with project root
		absPath = filepath.Join(absProjectRoot, cleanPath)
		absPath = filepath.Clean(absPath)
	}

	// Ensure the resolved path is within project root
	// Use filepath.Rel to check if path is within root
	relPath, err := filepath.Rel(absProjectRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve relative path: %w", err)
	}

	// Check for directory traversal attempts
	// If relPath starts with "..", it's trying to escape the root
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path escapes project root: %s", path)
	}

	// Additional check: ensure the path doesn't contain ".." after cleaning
	// This catches cases like "a/../b/../../etc/passwd"
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		if part == ".." {
			return "", fmt.Errorf("path contains directory traversal: %s", path)
		}
	}

	return absPath, nil
}

// ValidatePathExists ensures a path is valid AND exists
func ValidatePathExists(path, projectRoot string) (string, error) {
	absPath, err := ValidatePath(path, projectRoot)
	if err != nil {
		return "", err
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	return absPath, nil
}
