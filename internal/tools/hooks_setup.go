package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davidl71/exarp-go/internal/framework"
	"github.com/davidl71/exarp-go/internal/security"
)

// handleSetupHooksNative handles the setup_hooks tool with native Go implementation
// Currently implements "git" action, falls back to Python bridge for "patterns"
func handleSetupHooksNative(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
	// Get action (default: "git")
	action := "git"
	if actionRaw, ok := params["action"].(string); ok && actionRaw != "" {
		action = actionRaw
	}

	switch action {
	case "git":
		return handleSetupGitHooks(ctx, params)
	case "patterns":
		// Pattern triggers are complex, use Python bridge
		return nil, fmt.Errorf("pattern triggers not yet implemented in native Go, using Python bridge")
	default:
		return nil, fmt.Errorf("unknown hooks action: %s. Use 'git' or 'patterns'", action)
	}
}

// handleSetupGitHooks handles the "git" action for setup_hooks
func handleSetupGitHooks(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
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

	// Check if .git/hooks directory exists
	hooksDir := filepath.Join(projectRoot, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return nil, fmt.Errorf(".git/hooks directory not found. Are you in a git repository? Run 'git init' first")
	}

	// Get optional parameters
	install := true
	if installRaw, ok := params["install"].(bool); ok {
		install = installRaw
	}

	dryRun := false
	if dryRunRaw, ok := params["dry_run"].(bool); ok {
		dryRun = dryRunRaw
	}

	// Get hooks to install
	var hooksToInstall []string
	if hooksRaw, ok := params["hooks"]; ok && hooksRaw != nil {
		switch v := hooksRaw.(type) {
		case []interface{}:
			for _, h := range v {
				if hook, ok := h.(string); ok {
					hooksToInstall = append(hooksToInstall, hook)
				}
			}
		case []string:
			hooksToInstall = v
		}
	}

	// Default hooks if not specified
	if len(hooksToInstall) == 0 {
		hooksToInstall = []string{"pre-commit", "pre-push", "post-commit", "post-merge"}
	}

	// Hook configurations
	hookConfigs := map[string]string{
		"pre-commit": `#!/bin/sh
# Exarp pre-commit hook
# Run documentation health check and security scan

exarp-go health action=docs || exit 1
exarp-go security action=scan || exit 1
`,
		"pre-push": `#!/bin/sh
# Exarp pre-push hook
# Run task alignment and comprehensive security check

exarp-go analyze_alignment action=todo2 || exit 1
exarp-go security action=scan || exit 1
`,
		"post-commit": `#!/bin/sh
# Exarp post-commit hook
# Run automation discovery (non-blocking)

exarp-go automation action=discover || true
`,
		"post-merge": `#!/bin/sh
# Exarp post-merge hook
# Run duplicate detection and task sync (non-blocking)

exarp-go task_analysis action=duplicates || true
exarp-go task_workflow action=sync || true
`,
	}

	results := map[string]interface{}{
		"status":    "success",
		"action":    "git",
		"hooks_dir": hooksDir,
		"installed": []string{},
		"skipped":   []string{},
		"errors":    []string{},
		"dry_run":   dryRun,
	}

	if dryRun {
		// Just report what would be installed
		for _, hook := range hooksToInstall {
			if _, ok := hookConfigs[hook]; ok {
				results["installed"] = append(results["installed"].([]string), hook)
			} else {
				results["skipped"] = append(results["skipped"].([]string), hook)
			}
		}
	} else {
		// Actually install hooks
		for _, hook := range hooksToInstall {
			config, ok := hookConfigs[hook]
			if !ok {
				results["skipped"] = append(results["skipped"].([]string), hook)
				continue
			}

			hookPath := filepath.Join(hooksDir, hook)

			if install {
				// Write hook file with config content
				if err := os.WriteFile(hookPath, []byte(config), 0755); err != nil {
					results["errors"] = append(results["errors"].([]string), fmt.Sprintf("%s: %v", hook, err))
					continue
				}
				results["installed"] = append(results["installed"].([]string), hook)
			} else {
				// Uninstall: remove hook file
				if err := os.Remove(hookPath); err != nil && !os.IsNotExist(err) {
					results["errors"] = append(results["errors"].([]string), fmt.Sprintf("%s: %v", hook, err))
				} else {
					results["installed"] = append(results["installed"].([]string), fmt.Sprintf("removed: %s", hook))
				}
			}
		}
	}

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}
