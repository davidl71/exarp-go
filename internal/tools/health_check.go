package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
)

// handleHealthNative handles the health tool with native Go implementation
// Currently implements "server" action, falls back to Python bridge for others
func handleHealthNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get action (default: "server")
	action := "server"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	switch action {
	case "server":
		return handleHealthServer(ctx, params)
	default:
		// For other actions (git, docs, dod, cicd), return error to trigger Python bridge fallback
		return nil, fmt.Errorf("health action '%s' not yet implemented in native Go, using Python bridge", action)
	}
}

// handleHealthServer handles the "server" action for health tool
func handleHealthServer(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get project root
	projectRoot, err := security.GetProjectRoot(".")
	if err != nil {
		// Fallback to PROJECT_ROOT env var or current directory
		if envRoot := os.Getenv("PROJECT_ROOT"); envRoot != "" {
			projectRoot = envRoot
		} else {
			wd, _ := os.Getwd()
			projectRoot = wd
		}
	}

	// Try to get version from go.mod or pyproject.toml
	version := "unknown"
	
	// Check go.mod first (Go project)
	goModPath := filepath.Join(projectRoot, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Read go.mod to find module version or use git tag
		content, err := os.ReadFile(goModPath)
		if err == nil {
			// Look for module declaration
			moduleRegex := regexp.MustCompile(`module\s+([^\s]+)`)
			matches := moduleRegex.FindStringSubmatch(string(content))
			if len(matches) > 1 {
				// Try to get version from git
				gitCmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--always")
				gitCmd.Dir = projectRoot
				if output, err := gitCmd.Output(); err == nil {
					version = string(output)
					// Remove trailing newline
					if len(version) > 0 && version[len(version)-1] == '\n' {
						version = version[:len(version)-1]
					}
				}
			}
		}
	}

	// Fallback to pyproject.toml (Python project)
	if version == "unknown" {
		pyprojectPath := filepath.Join(projectRoot, "pyproject.toml")
		if _, err := os.Stat(pyprojectPath); err == nil {
			content, err := os.ReadFile(pyprojectPath)
			if err == nil {
				// Look for version = "x.y.z"
				versionRegex := regexp.MustCompile(`version\s*=\s*["']([^"']+)["']`)
				matches := versionRegex.FindStringSubmatch(string(content))
				if len(matches) > 1 {
					version = matches[1]
				}
			}
		}
	}

	// Build result
	result := map[string]interface{}{
		"status":      "operational",
		"version":     version,
		"project_root": projectRoot,
		"timestamp":   time.Now().Unix(),
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

