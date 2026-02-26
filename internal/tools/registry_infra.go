// registry_infra.go — Tool registrations: automation, git, testing, lint, security, config, hooks.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

func registerInfraTools(server framework.MCPServer) error {
	// automation
	if err := server.RegisterTool(
		"automation",
		"[HINT: action=daily|nightly|sprint|discover. Scheduled automation workflows. Use for routine maintenance, sprint automation, or discovering actionable tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"daily", "nightly", "sprint", "discover"},
					"default": "daily",
				},
				"tasks": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"include_slow": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"max_tasks_per_host": map[string]interface{}{
					"type":    "integer",
					"default": 5,
				},
				"max_parallel_tasks": map[string]interface{}{
					"type":    "integer",
					"default": 10,
				},
				"priority_filter": map[string]interface{}{
					"type": "string",
				},
				"tag_filter": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"max_iterations": map[string]interface{}{
					"type":    "integer",
					"default": 10,
				},
				"auto_approve": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"extract_subtasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"run_analysis_tools": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"run_testing_tools": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"min_value_score": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"use_cursor_agent": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true, run Cursor CLI agent -p in project root and attach output to result (daily/nightly/sprint). Requires agent on PATH.",
				},
				"cursor_agent_prompt": map[string]interface{}{
					"type":        "string",
					"description": "Prompt for Cursor agent step when use_cursor_agent is true. Default: \"Review the backlog and suggest which task to do next\".",
				},
			},
		},
		handleAutomation,
	); err != nil {
		return fmt.Errorf("failed to register automation: %w", err)
	}

	// git_tools
	if err := server.RegisterTool(
		"git_tools",
		"[HINT: action=commits|local_commits|branches|tasks|diff|graph|merge|set_branch. Git operations and task-commit linking. Use when reviewing changes or linking commits to tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"commits", "local_commits", "branches", "tasks", "diff", "graph", "merge", "set_branch"},
					"default": "commits",
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"branch": map[string]interface{}{
					"type": "string",
				},
				"limit": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
				"commit1": map[string]interface{}{
					"type": "string",
				},
				"commit2": map[string]interface{}{
					"type": "string",
				},
				"time1": map[string]interface{}{
					"type": "string",
				},
				"time2": map[string]interface{}{
					"type": "string",
				},
				"format": map[string]interface{}{
					"type":    "string",
					"default": "text",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"max_commits": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
				"source_branch": map[string]interface{}{
					"type": "string",
				},
				"target_branch": map[string]interface{}{
					"type": "string",
				},
				"conflict_strategy": map[string]interface{}{
					"type":    "string",
					"default": "newer",
				},
				"author": map[string]interface{}{
					"type":    "string",
					"default": "system",
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleGitTools,
	); err != nil {
		return fmt.Errorf("failed to register git_tools: %w", err)
	}

	// testing
	if err := server.RegisterTool(
		"testing",
		"[HINT: action=run|coverage|suggest|validate. Execute tests, analyze coverage, suggest tests. Use when running tests or checking coverage.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"run", "coverage", "suggest", "validate"},
					"default": "run",
				},
				"test_path": map[string]interface{}{
					"type": "string",
				},
				"test_framework": map[string]interface{}{
					"type":    "string",
					"default": "auto",
				},
				"verbose": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"coverage": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"coverage_file": map[string]interface{}{
					"type": "string",
				},
				"min_coverage": map[string]interface{}{
					"type":    "integer",
					"default": 80,
				},
				"format": map[string]interface{}{
					"type":    "string",
					"default": "html",
				},
				"target_file": map[string]interface{}{
					"type": "string",
				},
				"min_confidence": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"framework": map[string]interface{}{
					"type": "string",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
			},
		},
		handleTesting,
	); err != nil {
		return fmt.Errorf("failed to register testing: %w", err)
	}

	// lint
	if err := server.RegisterTool(
		"lint",
		"[HINT: action=run|analyze. Run linters or analyze results. Go: golangci-lint, go-vet, gofmt, goimports, deadcode. Markdown: markdownlint. Shell: shellcheck. C/C++: clang-tidy, cppcheck, clang-format. Python: ruff, flake8, pylint. Rust: clippy, rustfmt. PHP: phpcs, phpstan, php-cs-fixer. LaTeX: chktex, lacheck. linter=auto detects from file extension.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"run", "analyze"},
					"default": "run",
				},
				"path": map[string]interface{}{
					"type": "string",
				},
				"linter": map[string]interface{}{
					"type": "string",
					"enum": []string{
						"auto",
						"golangci-lint", "go-vet", "gofmt", "goimports", "deadcode",
						"markdownlint", "shellcheck",
						"clang-tidy", "cppcheck", "clang-format",
						"ruff", "flake8", "pylint",
						"clippy", "rustfmt",
						"phpcs", "phpstan", "php-cs-fixer",
						"chktex", "lacheck",
					},
					"default": "auto",
				},
				"fix": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"analyze": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"select": map[string]interface{}{
					"type": "string",
				},
				"ignore": map[string]interface{}{
					"type": "string",
				},
				"problems_json": map[string]interface{}{
					"type": "string",
				},
				"include_hints": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
			},
		},
		handleLint,
	); err != nil {
		return fmt.Errorf("failed to register lint: %w", err)
	}

	// security
	if err := server.RegisterTool(
		"security",
		"[HINT: action=scan|alerts|report. Security scanning and vulnerability reports. Use when checking for vulnerabilities or reviewing alerts.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"scan", "alerts", "report"},
					"default": "report",
				},
				"repo": map[string]interface{}{
					"type":    "string",
					"default": "davidl71/exarp-go",
				},
				"languages": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"config_path": map[string]interface{}{
					"type": "string",
				},
				"state": map[string]interface{}{
					"type":    "string",
					"default": "open",
				},
				"include_dismissed": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleSecurity,
	); err != nil {
		return fmt.Errorf("failed to register security: %w", err)
	}

	// generate_config
	if err := server.RegisterTool(
		"generate_config",
		"[HINT: action=rules|ignore|simplify. Generate Cursor config files (.cursor/rules/*.mdc, .cursorignore). Use when setting up Cursor IDE. Not for Claude Code.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"rules", "ignore", "simplify"},
					"default": "rules",
				},
				"rules": map[string]interface{}{
					"type": "string",
				},
				"overwrite": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"analyze_only": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"include_indexing": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"analyze_project": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"rule_files": map[string]interface{}{
					"type": "string",
				},
				"output_dir": map[string]interface{}{
					"type": "string",
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleGenerateConfig,
	); err != nil {
		return fmt.Errorf("failed to register generate_config: %w", err)
	}

	// setup_hooks
	if err := server.RegisterTool(
		"setup_hooks",
		"[HINT: action=git|patterns. Install git hooks and automation patterns. Use when setting up dev environment or CI hooks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"git", "patterns"},
					"default": "git",
				},
				"hooks": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"patterns": map[string]interface{}{
					"type": "string",
				},
				"config_path": map[string]interface{}{
					"type": "string",
				},
				"install": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleSetupHooks,
	); err != nil {
		return fmt.Errorf("failed to register setup_hooks: %w", err)
	}

	return nil
}
