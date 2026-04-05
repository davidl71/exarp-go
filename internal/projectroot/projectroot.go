// Package projectroot provides canonical project root resolution with configurable markers.
// Supports exarp markers (.exarp, .todo2) and go.mod for generic Go projects.
package projectroot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Exarp markers: .exarp or .todo2 directory.
var MarkersExarp = []string{".exarp", ".todo2"}

// MarkersGoMod: go.mod file (for generic Go project root).
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

// IsExarpGoSourceRoot reports whether root looks like a checkout of the exarp-go
// repository itself (the MCP server / CLI source), not an arbitrary consumer project.
// Used to avoid a stale shell PROJECT_ROOT pointing at exarp-go while cwd is another repo
// that also has .todo2/todo2.db.
func IsExarpGoSourceRoot(root string) bool {
	goMod := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goMod)
	if err != nil {
		return false
	}
	if !strings.Contains(string(data), "module github.com/davidl71/exarp-go") {
		return false
	}
	if fi, err := os.Stat(filepath.Join(root, "cmd", "server")); err != nil || !fi.IsDir() {
		return false
	}
	return true
}

// Find returns the exarp project root by checking (in order):
//  1. PROJECT_ROOT env (from Cursor {{PROJECT_ROOT}}), if valid and not placeholder —
//     except: if env points at the exarp-go *source* tree and cwd resolves to a different
//     project that has .exarp/.todo2, prefer cwd (stale PROJECT_ROOT in developer shells).
//  2. Walk up from cwd for .exarp or .todo2 directory.
//
// When PROJECT_ROOT is the literal placeholder "{{PROJECT_ROOT}}", it is treated as unset
// so that cwd-based resolution is used (server can still start when client does not substitute).
// When PROJECT_ROOT is set to a valid path, we usually trust it; see exception above.
func Find() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	cwdRoot, cwdErr := FindFromWithMarkers(dir, MarkersExarp)

	if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" && !strings.Contains(envRoot, "{{PROJECT_ROOT}}") {
		absPath, err := filepath.Abs(envRoot)
		if err != nil {
			return "", fmt.Errorf("PROJECT_ROOT: %w", err)
		}

		envNorm, errEnvSym := filepath.EvalSymlinks(absPath)
		if errEnvSym != nil {
			envNorm = absPath
		}

		strict := os.Getenv("EXARP_STRICT_PROJECT_ROOT") == "1" ||
			strings.EqualFold(os.Getenv("EXARP_STRICT_PROJECT_ROOT"), "true")
		if !strict && cwdErr == nil && cwdRoot != "" {
			cwdNorm, errCwdSym := filepath.EvalSymlinks(cwdRoot)
			if errCwdSym != nil {
				cwdNorm = cwdRoot
			}
			if envNorm != cwdNorm && IsExarpGoSourceRoot(envNorm) {
				return cwdRoot, nil
			}
		}

		if _, err := os.Stat(filepath.Join(absPath, ".exarp")); err == nil {
			return absPath, nil
		}

		if _, err := os.Stat(filepath.Join(absPath, ".todo2")); err == nil {
			return absPath, nil
		}

		// Trust env when no markers (e.g. IDE-supplied root or project not yet initialized)
		return absPath, nil
	}

	if cwdErr != nil {
		return "", cwdErr
	}
	return cwdRoot, nil
}

// FindFrom walks up from startPath looking for .exarp or .todo2.
func FindFrom(startPath string) (string, error) {
	return FindFromWithMarkers(startPath, MarkersExarp)
}

// FindGoMod walks up from startPath looking for go.mod. Compatible with mcp-go-core GetProjectRoot behavior.
func FindGoMod(startPath string) (string, error) {
	return FindFromWithMarkers(startPath, MarkersGoMod)
}
