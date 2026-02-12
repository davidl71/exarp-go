// Package projectroot provides canonical project root resolution with configurable markers.
// Supports exarp markers (.exarp, .todo2) and go.mod for generic Go projects.
package projectroot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Exarp markers: .exarp or .todo2 directory
var MarkersExarp = []string{".exarp", ".todo2"}

// MarkersGoMod: go.mod file (for generic Go project root)
var MarkersGoMod = []string{"go.mod"}

// FindFromWithMarkers walks up from startPath looking for any of the markers.
// Markers are path components (file or dir) to find in the candidate directory.
// startPath can be a file; it will use the containing directory.
func FindFromWithMarkers(startPath string, markers []string) (string, error) {
	if len(markers) == 0 {
		return "", fmt.Errorf("at least one marker required")
	}
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve start path: %w", err)
	}
	dir := absPath
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}
	for {
		for _, m := range markers {
			p := filepath.Join(dir, m)
			if _, err := os.Stat(p); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("project root not found from %s (markers: %v)", startPath, markers)
}

// Find returns the exarp project root by checking (in order):
// 1. PROJECT_ROOT env (from Cursor {{PROJECT_ROOT}}), if valid and not placeholder
// 2. Walk up from cwd for .exarp or .todo2 directory
func Find() (string, error) {
	if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
		if !strings.Contains(envRoot, "{{PROJECT_ROOT}}") {
			absPath, err := filepath.Abs(envRoot)
			if err == nil {
				if _, err := os.Stat(filepath.Join(absPath, ".exarp")); err == nil {
					return absPath, nil
				}
				if _, err := os.Stat(filepath.Join(absPath, ".todo2")); err == nil {
					return absPath, nil
				}
				return absPath, nil
			}
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
