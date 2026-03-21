// Package security provides file path validation and sanitization.
package security

import (
	"fmt"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/projectroot"
	"github.com/davidl71/mcp-go-core/pkg/mcp/security"
)

// GetProjectRoot finds the project root by walking up from startPath looking for go.mod.
// Uses projectroot.FindGoMod (unified with exarp's projectroot package).
var GetProjectRoot = projectroot.FindGoMod

// ValidatePath re-exported from mcp-go-core for backward compatibility.
var ValidatePath = security.ValidatePath

// ValidatePathExists re-exported from mcp-go-core.
var ValidatePathExists = security.ValidatePathExists

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
