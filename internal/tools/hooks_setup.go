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
// Implements both "git" and "patterns" actions - fully native Go
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
		return handleSetupPatternHooks(ctx, params)
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

# Suppress INFO logs in git hook context (reduces token usage)
export GIT_HOOK=1

exarp-go health action=docs || exit 1
exarp-go security action=scan || exit 1
`,
		"pre-push": `#!/bin/sh
# Exarp pre-push hook
# Run task alignment and comprehensive security check

# Suppress INFO logs in git hook context (reduces token usage)
export GIT_HOOK=1

exarp-go analyze_alignment action=todo2 || exit 1
exarp-go security action=scan || exit 1
`,
		"post-commit": `#!/bin/sh
# Exarp post-commit hook
# Run automation discovery (non-blocking)

# Suppress INFO logs in git hook context (reduces token usage)
export GIT_HOOK=1

exarp-go automation action=discover || true
`,
		"post-merge": `#!/bin/sh
# Exarp post-merge hook
# Run duplicate detection and task sync (non-blocking)

# Suppress INFO logs in git hook context (reduces token usage)
export GIT_HOOK=1

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

// handleSetupPatternHooks handles the "patterns" action for setup_hooks
// Sets up pattern-based automation triggers
func handleSetupPatternHooks(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
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

	// Get optional parameters
	dryRun := false
	if dryRunRaw, ok := params["dry_run"].(bool); ok {
		dryRun = dryRunRaw
	}

	// Get patterns (optional - can be JSON string or will use defaults)
	var patterns map[string]interface{}
	if patternsRaw, ok := params["patterns"].(string); ok && patternsRaw != "" {
		if err := json.Unmarshal([]byte(patternsRaw), &patterns); err != nil {
			return nil, fmt.Errorf("failed to parse patterns JSON: %w", err)
		}
	} else {
		// Use default patterns
		patterns = getDefaultPatterns()
	}

	// Get config path (optional)
	configPath := ""
	if configPathRaw, ok := params["config_path"].(string); ok && configPathRaw != "" {
		configPath = configPathRaw
		// Load patterns from config file if it exists
		if data, err := os.ReadFile(configPath); err == nil {
			var filePatterns map[string]interface{}
			if err := json.Unmarshal(data, &filePatterns); err == nil {
				// Merge file patterns into existing patterns
				for k, v := range filePatterns {
					if existing, ok := patterns[k].(map[string]interface{}); ok {
						// Merge maps
						if newMap, ok := v.(map[string]interface{}); ok {
							for k2, v2 := range newMap {
								existing[k2] = v2
							}
						}
					} else {
						patterns[k] = v
					}
				}
			}
		}
	}

	// Configuration file location
	cursorDir := filepath.Join(projectRoot, ".cursor")
	configFilePath := filepath.Join(cursorDir, "exarp_patterns.json")

	results := map[string]interface{}{
		"status":             "success",
		"patterns_configured": []map[string]interface{}{},
		"patterns_skipped":    []map[string]interface{}{},
		"config_file":         configFilePath,
		"dry_run":             dryRun,
	}

	if dryRun {
		// Just report what would be configured
		for category, patternConfig := range patterns {
			if configMap, ok := patternConfig.(map[string]interface{}); ok {
				patternNames := []string{}
				tools := extractToolsFromPatterns(configMap)
				for patternName := range configMap {
					patternNames = append(patternNames, patternName)
				}
				results["patterns_configured"] = append(results["patterns_configured"].([]map[string]interface{}), map[string]interface{}{
					"category": category,
					"patterns": patternNames,
					"tools":    tools,
				})
			}
		}
	} else {
		// Write configuration file
		if err := os.MkdirAll(cursorDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create .cursor directory: %w", err)
		}

		configJSON, err := json.MarshalIndent(patterns, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal patterns: %w", err)
		}

		if err := os.WriteFile(configFilePath, configJSON, 0644); err != nil {
			return nil, fmt.Errorf("failed to write config file: %w", err)
		}

		// Build results
		for category, patternConfig := range patterns {
			if configMap, ok := patternConfig.(map[string]interface{}); ok {
				tools := extractToolsFromPatterns(configMap)
				results["patterns_configured"] = append(results["patterns_configured"].([]map[string]interface{}), map[string]interface{}{
					"category":     category,
					"pattern_count": len(configMap),
					"tools":        tools,
				})
			}
		}

		// Setup integration points
		setupGitHooksIntegration(projectRoot, patterns, results)
		setupFileWatcherIntegration(projectRoot, patterns, results)
		setupTaskStatusIntegration(projectRoot, patterns, results)
	}

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return []framework.TextContent{
		{Type: "text", Text: string(resultJSON)},
	}, nil
}

// getDefaultPatterns returns default pattern configurations
func getDefaultPatterns() map[string]interface{} {
	return map[string]interface{}{
		"file_patterns": map[string]interface{}{
			"docs/**/*.md": map[string]interface{}{
				"on_change": "check_documentation_health_tool",
				"on_create": "add_external_tool_hints_tool",
				"description": "Documentation files",
			},
			"requirements.txt|Cargo.toml|package.json|pyproject.toml": map[string]interface{}{
				"on_change": "scan_dependency_security_tool",
				"description": "Dependency files",
			},
			".todo2/state.todo2.json": map[string]interface{}{
				"on_change": "detect_duplicate_tasks_tool",
				"description": "Todo2 state file",
			},
			"CMakeLists.txt|CMakePresets.json": map[string]interface{}{
				"on_change": "validate_ci_cd_workflow_tool",
				"description": "CMake configuration",
			},
		},
		"git_events": map[string]interface{}{
			"pre_commit": map[string]interface{}{
				"tools": []string{
					"check_documentation_health_tool --quick",
					"scan_dependency_security_tool --quick",
				},
				"blocking": true,
				"description": "Quick checks before commit",
			},
			"pre_push": map[string]interface{}{
				"tools": []string{
					"analyze_todo2_alignment_tool",
					"scan_dependency_security_tool",
					"check_documentation_health_tool",
				},
				"blocking": true,
				"description": "Comprehensive checks before push",
			},
			"post_commit": map[string]interface{}{
				"tools": []string{
					"find_automation_opportunities_tool --quick",
				},
				"blocking": false,
				"description": "Non-blocking checks after commit",
			},
			"post_merge": map[string]interface{}{
				"tools": []string{
					"detect_duplicate_tasks_tool",
					"sync_todo_tasks_tool",
				},
				"blocking": false,
				"description": "Checks after merge",
			},
		},
		"task_status_changes": map[string]interface{}{
			"Todo → In Progress": map[string]interface{}{
				"tools": []string{"analyze_todo2_alignment_tool"},
				"description": "Verify alignment when starting work",
			},
			"In Progress → Review": map[string]interface{}{
				"tools": []string{
					"analyze_todo2_alignment_tool",
					"detect_duplicate_tasks_tool",
				},
				"description": "Quality checks before review",
			},
			"Review → Done": map[string]interface{}{
				"tools": []string{"detect_duplicate_tasks_tool"},
				"description": "Final checks on completion",
			},
		},
	}
}

// extractToolsFromPatterns extracts tool names from pattern configuration
func extractToolsFromPatterns(patternConfig map[string]interface{}) []string {
	tools := []string{}
	toolSet := make(map[string]bool)

	for _, config := range patternConfig {
		if configMap, ok := config.(map[string]interface{}); ok {
			if toolsList, ok := configMap["tools"].([]interface{}); ok {
				for _, tool := range toolsList {
					if toolStr, ok := tool.(string); ok && !toolSet[toolStr] {
						tools = append(tools, toolStr)
						toolSet[toolStr] = true
					}
				}
			}
			if onChange, ok := configMap["on_change"].(string); ok && !toolSet[onChange] {
				tools = append(tools, onChange)
				toolSet[onChange] = true
			}
			if onCreate, ok := configMap["on_create"].(string); ok && !toolSet[onCreate] {
				tools = append(tools, onCreate)
				toolSet[onCreate] = true
			}
		}
	}

	return tools
}

// setupGitHooksIntegration sets up git hooks integration for pattern triggers
func setupGitHooksIntegration(projectRoot string, patterns map[string]interface{}, results map[string]interface{}) {
	if gitEvents, ok := patterns["git_events"].(map[string]interface{}); ok && len(gitEvents) > 0 {
		hooksDir := filepath.Join(projectRoot, ".git", "hooks")
		if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
			results["patterns_skipped"] = append(results["patterns_skipped"].([]map[string]interface{}), map[string]interface{}{
				"category": "git_events",
				"reason":   ".git/hooks directory not found",
			})
			return
		}

		results["git_hooks_integration"] = map[string]interface{}{
			"status": "configured",
			"note":   "Use setup_hooks action=git to install actual hooks",
		}
	}
}

// setupFileWatcherIntegration sets up file watcher integration for pattern triggers
func setupFileWatcherIntegration(projectRoot string, patterns map[string]interface{}, results map[string]interface{}) {
	if filePatterns, ok := patterns["file_patterns"].(map[string]interface{}); ok && len(filePatterns) > 0 {
		watcherScript := filepath.Join(projectRoot, ".cursor", "exarp_file_watcher.py")

		watcherContent := generateFileWatcherScript()

		if err := os.WriteFile(watcherScript, []byte(watcherContent), 0755); err != nil {
			results["patterns_skipped"] = append(results["patterns_skipped"].([]map[string]interface{}), map[string]interface{}{
				"category": "file_patterns",
				"reason":   fmt.Sprintf("Failed to create watcher script: %v", err),
			})
			return
		}

		results["file_watcher_integration"] = map[string]interface{}{
			"status": "configured",
			"script": watcherScript,
			"note":   "Run manually or via cron: python3 .cursor/exarp_file_watcher.py",
		}
	}
}

// setupTaskStatusIntegration sets up task status change integration
func setupTaskStatusIntegration(projectRoot string, patterns map[string]interface{}, results map[string]interface{}) {
	if taskStatus, ok := patterns["task_status_changes"].(map[string]interface{}); ok && len(taskStatus) > 0 {
		results["task_status_integration"] = map[string]interface{}{
			"status": "configured",
			"note":   "Task status triggers handled by Todo2 MCP server hooks",
		}
	}
}

// generateFileWatcherScript generates the file watcher script content
func generateFileWatcherScript() string {
	return `#!/usr/bin/env python3
"""
Exarp File Watcher

Monitors file changes and triggers exarp tools based on patterns.
Run manually or via cron job.
"""

import json
import sys
from pathlib import Path
from typing import Dict, List, Optional

# Load pattern configuration
CONFIG_FILE = Path(__file__).parent.parent / ".cursor" / "exarp_patterns.json"

def load_patterns() -> Dict:
    """Load pattern configuration."""
    if not CONFIG_FILE.exists():
        return {}

    with open(CONFIG_FILE, 'r') as f:
        config = json.load(f)
        return config.get("file_patterns", {})

def check_file_changes() -> List[Dict]:
    """Check for file changes and return matching patterns."""
    # This is a placeholder - implement actual file watching logic
    # For now, returns empty list
    return []

def trigger_tool(tool_name: str) -> bool:
    """Trigger an exarp tool."""
    # This is a placeholder - implement actual tool triggering
    # For now, just prints what would be triggered
    print(f"Would trigger: {tool_name}")
    return True

if __name__ == "__main__":
    patterns = load_patterns()
    changes = check_file_changes()

    for change in changes:
        file_path = change["file"]
        for pattern, config in patterns.items():
            # Simple pattern matching (implement proper glob/regex matching)
            if pattern in file_path or file_path.endswith(pattern.split("/")[-1]):
                if "on_change" in config:
                    trigger_tool(config["on_change"])
                if "on_create" in config and change.get("created"):
                    trigger_tool(config["on_create"])
`
}
