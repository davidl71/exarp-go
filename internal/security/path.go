package security

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/projectroot"
	"github.com/davidl71/mcp-go-core/pkg/mcp/security"
)

// GetProjectRoot finds the project root by walking up from startPath looking for go.mod.
// Uses projectroot.FindGoMod (unified with exarp's projectroot package).
var GetProjectRoot = projectroot.FindGoMod

// ValidatePath re-exported from mcp-go-core for backward compatibility.
var ValidatePath = security.ValidatePath

// ValidatePathExists ensures a path is valid AND exists
// This is a local extension not in mcp-go-core.
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

// ValidatePathWithinRoot is a convenience function that validates a path is within root
// and returns the relative path from root.
func ValidatePathWithinRoot(path, projectRoot string) (string, string, error) {
	absPath, err := ValidatePath(path, projectRoot)
	if err != nil {
		return "", "", err
	}

	absProjectRoot, _ := filepath.Abs(projectRoot)

	relPath, err := filepath.Rel(absProjectRoot, absPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to get relative path: %w", err)
	}

	return absPath, relPath, nil
}
