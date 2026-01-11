package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
		return handleSetupPatternTriggers(ctx, params)
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

// handleSetupPatternTriggers handles the "patterns" action for setup_hooks
// Sets up pattern-based automation triggers for file changes, git events, and task status changes
func handleSetupPatternTriggers(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error) {
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
	install := true
	if installRaw, ok := params["install"].(bool); ok {
		install = installRaw
	}

	dryRun := false
	if dryRunRaw, ok := params["dry_run"].(bool); ok {
		dryRun = dryRunRaw
	}

	// Get patterns from params or use defaults
	patterns := getDefaultPatterns()

	// Parse patterns if provided
	if patternsRaw, ok := params["patterns"]; ok && patternsRaw != nil {
		var parsedPatterns map[string]interface{}
		switch v := patternsRaw.(type) {
		case string:
			if v != "" {
				if err := json.Unmarshal([]byte(v), &parsedPatterns); err != nil {
					return nil, fmt.Errorf("invalid JSON in patterns parameter: %w", err)
				}
				// Merge with defaults
				for k, v := range parsedPatterns {
					if pMap, ok := v.(map[string]interface{}); ok {
						patterns[k] = pMap
					}
				}
			}
		case map[string]interface{}:
			parsedPatterns = v
			// Merge with defaults
			for k, v := range parsedPatterns {
				if pMap, ok := v.(map[string]interface{}); ok {
					patterns[k] = pMap
				}
			}
		}
	}

	// Load from config file if provided
	if configPathRaw, ok := params["config_path"].(string); ok && configPathRaw != "" {
		configData, err := os.ReadFile(configPathRaw)
		if err == nil {
			var filePatterns map[string]interface{}
			if err := json.Unmarshal(configData, &filePatterns); err == nil {
				// Merge with existing patterns
				for k, v := range filePatterns {
					if pMap, ok := v.(map[string]interface{}); ok {
						patterns[k] = pMap
					}
				}
			}
		}
	}

	// Configuration file location
	cursorDir := filepath.Join(projectRoot, ".cursor")
	configFilePath := filepath.Join(cursorDir, "automa_patterns.json")

	results := map[string]interface{}{
		"status":              "success",
		"patterns_configured": []interface{}{},
		"patterns_skipped":    []interface{}{},
		"config_file":         configFilePath,
		"dry_run":             dryRun,
	}

	if dryRun {
		// Just report what would be configured
		configured := []interface{}{}
		for category, patternConfig := range patterns {
			if pMap, ok := patternConfig.(map[string]interface{}); ok {
				tools := extractToolsFromPatterns(pMap)
				configured = append(configured, map[string]interface{}{
					"category":     category,
					"patterns":     getMapKeys(pMap),
					"tools":        tools,
					"pattern_count": len(pMap),
				})
			}
		}
		results["patterns_configured"] = configured

		resultJSON, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return []framework.TextContent{
			{Type: "text", Text: string(resultJSON)},
		}, nil
	}

	// Write configuration file
	if install {
		if err := os.MkdirAll(cursorDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create .cursor directory: %w", err)
		}

		configData, err := json.MarshalIndent(patterns, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal patterns: %w", err)
		}

		if err := os.WriteFile(configFilePath, configData, 0644); err != nil {
			return nil, fmt.Errorf("failed to write config file: %w", err)
		}

		// Setup integration points
		setupPatternIntegrations(projectRoot, patterns, results)

		// Build configured patterns list
		configured := []interface{}{}
		for category, patternConfig := range patterns {
			if pMap, ok := patternConfig.(map[string]interface{}); ok {
				tools := extractToolsFromPatterns(pMap)
				configured = append(configured, map[string]interface{}{
					"category":      category,
					"pattern_count": len(pMap),
					"tools":         tools,
				})
			}
		}
		results["patterns_configured"] = configured
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
				"on_change":   "check_documentation_health_tool",
				"on_create":   "add_external_tool_hints_tool",
				"description": "Documentation files",
			},
			"requirements.txt|Cargo.toml|package.json|pyproject.toml": map[string]interface{}{
				"on_change":   "scan_dependency_security_tool",
				"description": "Dependency files",
			},
			".todo2/state.todo2.json": map[string]interface{}{
				"on_change":   "detect_duplicate_tasks_tool",
				"description": "Todo2 state file",
			},
			"CMakeLists.txt|CMakePresets.json": map[string]interface{}{
				"on_change":   "validate_ci_cd_workflow_tool",
				"description": "CMake configuration",
			},
		},
		"git_events": map[string]interface{}{
			"pre_commit": map[string]interface{}{
				"tools":       []string{"check_documentation_health_tool --quick", "scan_dependency_security_tool --quick"},
				"blocking":    true,
				"description": "Quick checks before commit",
			},
			"pre_push": map[string]interface{}{
				"tools":       []string{"analyze_todo2_alignment_tool", "scan_dependency_security_tool", "check_documentation_health_tool"},
				"blocking":    true,
				"description": "Comprehensive checks before push",
			},
			"post_commit": map[string]interface{}{
				"tools":       []string{"find_automation_opportunities_tool --quick"},
				"blocking":    false,
				"description": "Non-blocking checks after commit",
			},
			"post_merge": map[string]interface{}{
				"tools":       []string{"detect_duplicate_tasks_tool", "sync_todo_tasks_tool"},
				"blocking":    false,
				"description": "Checks after merge",
			},
		},
		"task_status_changes": map[string]interface{}{
			"Todo → In Progress": map[string]interface{}{
				"tools":       []string{"analyze_todo2_alignment_tool"},
				"description": "Verify alignment when starting work",
			},
			"In Progress → Review": map[string]interface{}{
				"tools":       []string{"analyze_todo2_alignment_tool", "detect_duplicate_tasks_tool"},
				"description": "Quality checks before review",
			},
			"Review → Done": map[string]interface{}{
				"tools":       []string{"detect_duplicate_tasks_tool"},
				"description": "Final checks on completion",
			},
		},
	}
}

// extractToolsFromPatterns extracts tool names from pattern configuration
func extractToolsFromPatterns(patternConfig map[string]interface{}) []string {
	tools := make(map[string]bool)

	for _, config := range patternConfig {
		if configMap, ok := config.(map[string]interface{}); ok {
			// Check for "tools" array
			if toolsRaw, ok := configMap["tools"].([]interface{}); ok {
				for _, tool := range toolsRaw {
					if toolStr, ok := tool.(string); ok {
						// Extract tool name (remove args)
						toolName := strings.Fields(toolStr)[0]
						tools[toolName] = true
					}
				}
			}

			// Check for "on_change" string
			if onChange, ok := configMap["on_change"].(string); ok {
				toolName := strings.Fields(onChange)[0]
				tools[toolName] = true
			}

			// Check for "on_create" string
			if onCreate, ok := configMap["on_create"].(string); ok {
				toolName := strings.Fields(onCreate)[0]
				tools[toolName] = true
			}
		}
	}

	// Convert map keys to slice
	result := []string{}
	for tool := range tools {
		result = append(result, tool)
	}

	return result
}

// getMapKeys returns keys of a map as a slice
func getMapKeys(m map[string]interface{}) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// setupPatternIntegrations sets up integration points for pattern triggers
func setupPatternIntegrations(projectRoot string, patterns map[string]interface{}, results map[string]interface{}) {
	// Git hooks integration
	if _, ok := patterns["git_events"]; ok {
		hooksDir := filepath.Join(projectRoot, ".git", "hooks")
		if _, err := os.Stat(hooksDir); err == nil {
			results["git_hooks_integration"] = map[string]interface{}{
				"status": "configured",
				"note":   "Use setup_hooks(action=\"git\") to install actual hooks",
			}
		}
	}

	// File watcher integration
	if _, ok := patterns["file_patterns"]; ok {
		watcherScript := filepath.Join(projectRoot, ".cursor", "automa_file_watcher.py")
		scriptContent := generateFileWatcherScript()

		if err := os.WriteFile(watcherScript, []byte(scriptContent), 0755); err == nil {
			results["file_watcher_integration"] = map[string]interface{}{
				"status": "configured",
				"script": watcherScript,
				"note":   "Run manually or via cron: python3 .cursor/automa_file_watcher.py",
			}
		} else {
			if skipped, ok := results["patterns_skipped"].([]interface{}); ok {
				results["patterns_skipped"] = append(skipped, map[string]interface{}{
					"category": "file_patterns",
					"reason":   fmt.Sprintf("Failed to create watcher script: %v", err),
				})
			}
		}
	}

	// Task status integration
	if _, ok := patterns["task_status_changes"]; ok {
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
CONFIG_FILE = Path(__file__).parent.parent / ".cursor" / "automa_patterns.json"

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
