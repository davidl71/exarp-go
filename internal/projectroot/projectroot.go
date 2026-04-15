// Package projectroot provides canonical project root resolution for exarp-go.
// The generic marker-walking logic lives in mcp-go-core/pkg/mcp/projectroot;
// this package adds exarp-specific marker sets and the PROJECT_ROOT env override.
package projectroot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mcpprojectroot "github.com/davidl71/mcp-go-core/pkg/mcp/projectroot"
)

// Exarp markers: .exarp or .todo2 directory.
var MarkersExarp = []string{".exarp", ".todo2"}

// MarkersGoMod: go.mod file (for generic Go project root).
var MarkersGoMod = []string{"go.mod"}

// FindFromWithMarkers walks up from startPath looking for any of the given markers.
// Delegates to mcp-go-core's generic implementation.
func FindFromWithMarkers(startPath string, markers []string) (string, error) {
	return mcpprojectroot.FindFromWithMarkers(startPath, markers)
}

// Find returns the exarp project root by checking (in order):
// 1. PROJECT_ROOT env (from Cursor {{PROJECT_ROOT}}), if valid and not placeholder
// 2. Walk up from cwd for .exarp or .todo2 directory.
// When PROJECT_ROOT is the literal placeholder "{{PROJECT_ROOT}}", it is treated as unset
// so that cwd-based resolution is used (server can still start when client does not substitute).
// When PROJECT_ROOT is set to a valid path, we trust it and return it even if .exarp/.todo2
// are missing (e.g. Cursor/IDE-supplied root or a project not yet initialized).
func Find() (string, error) {
	if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" && !strings.Contains(envRoot, "{{PROJECT_ROOT}}") {
		absPath, err := filepath.Abs(envRoot)
		if err == nil {
			if _, err := os.Stat(filepath.Join(absPath, ".exarp")); err == nil {
				return absPath, nil
			}

			if _, err := os.Stat(filepath.Join(absPath, ".todo2")); err == nil {
				return absPath, nil
			}

			// Trust env when no markers (e.g. IDE-supplied root or project not yet initialized)
			return absPath, nil
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return FindFromWithMarkers(dir, MarkersExarp)
}

// FindFrom walks up from startPath looking for .exarp or .todo2.
func FindFrom(startPath string) (string, error) {
	return FindFromWithMarkers(startPath, MarkersExarp)
}

// FindGoMod walks up from startPath looking for go.mod. Compatible with mcp-go-core GetProjectRoot behavior.
func FindGoMod(startPath string) (string, error) {
	return FindFromWithMarkers(startPath, MarkersGoMod)
}
