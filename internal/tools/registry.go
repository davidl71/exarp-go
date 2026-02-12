package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// RegisterAllTools registers all tools with the server
func RegisterAllTools(server framework.MCPServer) error {
	// Batch 1: Simple tools (T-22 through T-27)
	if err := registerBatch1Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 1 tools: %w", err)
	}

	// Batch 2: Medium tools (T-28 through T-35)
	if err := registerBatch2Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 2 tools: %w", err)
	}

	// Batch 3: Advanced tools (T-37 through T-44)
	if err := registerBatch3Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 3 tools: %w", err)
	}

	// Batch 4: mcp-generic-tools migration (2 native Go tools)
	if err := registerBatch4Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 4 tools: %w", err)
	}

	// Batch 5: Phase 3 migration - remaining unified tools (4 tools)
	if err := registerBatch5Tools(server); err != nil {
		return fmt.Errorf("failed to register Batch 5 tools: %w", err)
	}

	return nil
}

// registerBatch1Tools registers Batch 1 tools (6 simple tools)
func registerBatch1Tools(server framework.MCPServer) error {
	// T-22: analyze_alignment
	if err := server.RegisterTool(
		"analyze_alignment",
		"[HINT: Alignment analysis. action=todo2|prd. Unified alignment analysis tool.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"todo2", "prd"},
					"default": "todo2",
				},
				"create_followup_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
			},
		},
		handleAnalyzeAlignment,
	); err != nil {
		return fmt.Errorf("failed to register analyze_alignment: %w", err)
	}

	// T-23: generate_config
	if err := server.RegisterTool(
		"generate_config",
		"[HINT: Config generation. action=rules|ignore|simplify. Creates IDE config files.]",
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

	// T-24: health
	if err := server.RegisterTool(
		"health",
		"[HINT: Health check. action=server|git|docs|dod|cicd|tools. Status and health metrics; tools=Tool count vs design limit.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"server", "git", "docs", "dod", "cicd", "tools"},
					"default": "server",
				},
				"agent_name": map[string]interface{}{
					"type": "string",
				},
				"check_remote": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"create_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"changed_files": map[string]interface{}{
					"type": "string",
				},
				"auto_check": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"workflow_path": map[string]interface{}{
					"type": "string",
				},
				"check_runners": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		handleHealth,
	); err != nil {
		return fmt.Errorf("failed to register health: %w", err)
	}

	// T-25: setup_hooks
	if err := server.RegisterTool(
		"setup_hooks",
		"[HINT: Hooks setup. action=git|patterns. Install automation hooks.]",
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

	// T-26: check_attribution
	if err := server.RegisterTool(
		"check_attribution",
		"[HINT: Attribution compliance check. Verify proper attribution for all third-party components.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"create_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		handleCheckAttribution,
	); err != nil {
		return fmt.Errorf("failed to register check_attribution: %w", err)
	}

	// T-27: add_external_tool_hints
	if err := server.RegisterTool(
		"add_external_tool_hints",
		"[HINT: Tool hints. Files scanned, modified, hints added.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"min_file_size": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
			},
		},
		handleAddExternalToolHints,
	); err != nil {
		return fmt.Errorf("failed to register add_external_tool_hints: %w", err)
	}

	return nil
}

// registerBatch2Tools registers Batch 2 tools (8 medium tools)
func registerBatch2Tools(server framework.MCPServer) error {
	// T-28: memory
	if err := server.RegisterTool(
		"memory",
		"[HINT: Memory tool. action=save|recall|search. Persist and retrieve AI discoveries.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"save", "recall", "search"},
					"default": "search",
				},
				"title": map[string]interface{}{
					"type": "string",
				},
				"content": map[string]interface{}{
					"type": "string",
				},
				"category": map[string]interface{}{
					"type":    "string",
					"default": "insight",
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"metadata": map[string]interface{}{
					"type": "string",
				},
				"include_related": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"query": map[string]interface{}{
					"type": "string",
				},
				"limit": map[string]interface{}{
					"type":    "integer",
					"default": 10,
				},
			},
		},
		handleMemory,
	); err != nil {
		return fmt.Errorf("failed to register memory: %w", err)
	}

	// T-29: memory_maint
	if err := server.RegisterTool(
		"memory_maint",
		"[HINT: Memory maintenance. action=health|gc|prune|consolidate|dream. Lifecycle management.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"health", "gc", "prune", "consolidate", "dream"},
					"default": "health",
				},
				"max_age_days": map[string]interface{}{
					"type":    "integer",
					"default": 90,
				},
				"delete_orphaned": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"delete_duplicates": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"scorecard_max_age_days": map[string]interface{}{
					"type":    "integer",
					"default": 7,
				},
				"value_threshold": map[string]interface{}{
					"type":    "number",
					"default": 0.3,
				},
				"keep_minimum": map[string]interface{}{
					"type":    "integer",
					"default": 50,
				},
				"similarity_threshold": map[string]interface{}{
					"type":    "number",
					"default": 0.85,
				},
				"merge_strategy": map[string]interface{}{
					"type":    "string",
					"default": "newest",
				},
				"scope": map[string]interface{}{
					"type":    "string",
					"default": "week",
				},
				"advisors": map[string]interface{}{
					"type": "string",
				},
				"generate_insights": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"save_dream": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"interactive": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		handleMemoryMaint,
	); err != nil {
		return fmt.Errorf("failed to register memory_maint: %w", err)
	}

	// T-30: report
	if err := server.RegisterTool(
		"report",
		"[HINT: Report generation. action=overview|scorecard|briefing|prd|plan|scorecard_plans. plan generates Cursor-style .plan.md. scorecard_plans writes one improvement plan per scorecard dimension (testing, security, documentation, completion) to .cursor/plans/improve-<dim>.plan.md. For scorecard: fast_mode (default true) or full.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"overview", "scorecard", "briefing", "prd", "plan", "scorecard_plans"},
					"default":     "overview",
					"description": "plan: write Cursor .plan.md; scorecard_plans: write improve-<dim>.plan.md per dimension",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"default": "text",
				},
				"output_path": map[string]interface{}{
					"type":        "string",
					"description": "For action=plan: path for plan file (default: .cursor/plans/<project-slug>.plan.md so Cursor shows Build)",
				},
				"plan_title": map[string]interface{}{
					"type":        "string",
					"description": "For action=plan: title for the plan (default: project name from go.mod or directory)",
				},
				"include_planning": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "If true, overview includes critical path and suggested backlog order (first 10)",
				},
				"fast_mode": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_recommendations": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"overall_score": map[string]interface{}{
					"type":    "number",
					"default": 50.0,
				},
				"security_score": map[string]interface{}{
					"type":    "number",
					"default": 50.0,
				},
				"testing_score": map[string]interface{}{
					"type":    "number",
					"default": 50.0,
				},
				"documentation_score": map[string]interface{}{
					"type":    "number",
					"default": 50.0,
				},
				"completion_score": map[string]interface{}{
					"type":    "number",
					"default": 50.0,
				},
				"alignment_score": map[string]interface{}{
					"type":    "number",
					"default": 50.0,
				},
				"project_name": map[string]interface{}{
					"type": "string",
				},
				"include_architecture": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_metrics": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		handleReport,
	); err != nil {
		return fmt.Errorf("failed to register report: %w", err)
	}

	// T-31: security
	if err := server.RegisterTool(
		"security",
		"[HINT: Security. action=scan|alerts|report. Vulnerabilities, remediation.]",
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

	// T-32: task_analysis
	if err := server.RegisterTool(
		"task_analysis",
		"[HINT: Task analysis. action=duplicates|tags|discover_tags|hierarchy|dependencies|dependencies_summary|parallelization|validate|execution_plan|complexity|conflicts. Task quality and structure. conflicts detects task-overlap (In Progress tasks with dependent also In Progress). complexity classifies tasks (simple/medium/complex) per Model-Assisted Workflow. discover_tags scans MD files for tag hints and uses LLM (Apple FM/Ollama) for semantic inference.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"duplicates", "tags", "discover_tags", "hierarchy", "dependencies", "dependencies_summary", "parallelization", "fix_missing_deps", "validate", "execution_plan", "complexity", "conflicts"},
					"default": "duplicates",
				},
				"similarity_threshold": map[string]interface{}{
					"type":    "number",
					"default": 0.85,
				},
				"auto_fix": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"use_canonical_rules": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "Apply built-in canonical tag rules (scorecard-aligned: testing, docs, security, build, performance, bug, feature, refactor, migration, config, cli, mcp, llm, database)",
				},
				"custom_rules": map[string]interface{}{
					"type":        "string",
					"description": "JSON object of additional tag rename rules (oldTag -> newTag)",
				},
				"remove_tags": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of tags to remove",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"default": "text",
				},
				"include_recommendations": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"include_hierarchy": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=validate: optionally run hierarchy dry-run and report hierarchy_warning (e.g. response_snippet) if FM returns non-JSON",
				},
				"use_llm": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "For action=discover_tags: use Apple FM or Ollama for semantic tag inference",
				},
				"doc_path": map[string]interface{}{
					"type":        "string",
					"default":     "docs",
					"description": "For action=discover_tags: path to scan for markdown files (relative to project root)",
				},
				"use_cache": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "For action=discover_tags: use SQLite cache for discovered tags (speeds up subsequent runs)",
				},
				"clear_cache": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=discover_tags: clear tag cache before scanning (force re-scan)",
				},
				"timeout_seconds": map[string]interface{}{
					"type":        "integer",
					"default":     300,
					"description": "For action=discover_tags: total operation timeout in seconds (default: 300s/5min). LLM calls have per-file timeout of 10s.",
				},
				"llm_batch_size": map[string]interface{}{
					"type":        "integer",
					"default":     0,
					"description": "For action=discover_tags: max files per LLM call. For action=tags with use_llm_semantic: max tasks per LLM call. 0=default 15.",
				},
				"backlog_only": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=discover_tags: when true, only match and apply tags to Todo2 backlog tasks (status Todo or In Progress). Parse todo2 backlog and update tags.",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"default":     0,
					"description": "For action=tags or discover_tags: max number of tasks to process in this run (batch size). 0 = no limit. With prioritize_untagged, untagged tasks fill the batch first.",
				},
				"prioritize_untagged": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=tags or discover_tags: when true, process/return only tasks that have no tags first; with limit, fill batch with untagged tasks.",
				},
				"use_llm_semantic": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=tags: use Apple FM or Ollama to suggest additional tags from task title and content (batched). Quick tag addition from semantic analysis.",
				},
				"match_existing_only": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=tags with use_llm_semantic: quick Apple FM tag inference matching only from existing tags (canonical + project/cache). Constrained output, faster.",
				},
				"use_tiny_tag_model": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=tags with use_llm_semantic: try Ollama with tinyllama then MLX with TinyLlama (1.1B) for faster tag inference before Apple FM.",
				},
				"filter_tag": map[string]interface{}{
					"type":        "string",
					"description": "For action=execution_plan: restrict backlog to tasks with this tag.",
				},
				"filter_tags": map[string]interface{}{
					"type":        "string",
					"description": "For action=execution_plan: restrict backlog to tasks with any of these tags (comma-separated).",
				},
			},
		},
		handleTaskAnalysis,
	); err != nil {
		return fmt.Errorf("failed to register task_analysis: %w", err)
	}

	// T-33: task_discovery
	if err := server.RegisterTool(
		"task_discovery",
		"[HINT: Task discovery. action=comments|markdown|orphans|git_json|planning_links|all. Scans for TODO/FIXME comments, extracts hashtags as tags (e.g., #refactor, #bug). Returns discoveries with tags array. Use create_tasks=true to create Todo2 tasks with extracted tags.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"comments", "markdown", "orphans", "git_json", "planning_links", "all"},
					"default": "all",
				},
				"file_patterns": map[string]interface{}{
					"type": "string",
				},
				"include_fixme": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"doc_path": map[string]interface{}{
					"type": "string",
				},
				"json_pattern": map[string]interface{}{
					"type":    "string",
					"default": "**/.todo2/state.todo2.json",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"create_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleTaskDiscovery,
	); err != nil {
		return fmt.Errorf("failed to register task_discovery: %w", err)
	}

	// T-34: task_workflow
	if err := server.RegisterTool(
		"task_workflow",
		"[HINT: Task workflow. action=sync|approve|clarify|clarity|cleanup|create|fix_dates|fix_empty_descriptions|fix_invalid_ids|link_planning|request_approval|sync_approvals|apply_approval_result|sanity_check|sync_from_plan|sync_plan_status. Manage task lifecycle. ⚠️ CRITICAL: PREFER convenience commands (exarp-go task ...) for common operations. FALLBACK to this tool for advanced operations (clarity, cleanup, complex filters). NEVER edit .todo2/state.todo2.json directly. Use action=approve with task_ids for batch updates. Use action=create to create new tasks. Use action=link_planning with task_id/task_ids and planning_doc/epic_id to set planning hints on Todo or In Progress tasks only. Sync is SQLite↔JSON only; external sync is a future nice-to-have (ignored if passed).]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"sync", "approve", "clarify", "clarity", "cleanup", "create", "delete", "fix_dates", "fix_empty_descriptions", "fix_invalid_ids", "link_planning", "request_approval", "sync_approvals", "apply_approval_result", "sanity_check", "sync_from_plan", "sync_plan_status", "update"},
					"default": "sync",
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"confirm_via_elicitation": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true and client supports MCP elicitation, prompt user to confirm before approve or delete (form: proceed, optional dry_run for approve)",
				},
				"external": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "Future nice-to-have: sync with external sources (e.g. infer_task_progress). Currently ignored; SQLite↔JSON sync is performed.",
				},
				"status": map[string]interface{}{
					"type":    "string",
					"default": "Review",
				},
				"new_status": map[string]interface{}{
					"type":    "string",
					"default": "Todo",
				},
				"clarification_none": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"filter_tag": map[string]interface{}{
					"type": "string",
				},
				"task_ids": map[string]interface{}{
					"type": "string",
				},
				"sub_action": map[string]interface{}{
					"type":    "string",
					"default": "list",
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"form_id": map[string]interface{}{
					"type":        "string",
					"description": "For request_approval/sync_approvals: gotoHuman form ID from list-forms (optional)",
				},
				"result": map[string]interface{}{
					"type":        "string",
					"description": "For apply_approval_result: 'approved' or 'rejected' (from gotoHuman decision)",
				},
				"feedback": map[string]interface{}{
					"type":        "string",
					"description": "For apply_approval_result: optional feedback when result=rejected (appended to task)",
				},
				"clarification_text": map[string]interface{}{
					"type": "string",
				},
				"decision": map[string]interface{}{
					"type": "string",
				},
				"decisions_json": map[string]interface{}{
					"type": "string",
				},
				"move_to_todo": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"auto_apply": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"order": map[string]interface{}{
					"type":        "string",
					"description": "For sub_action=list: order results by 'execution' or 'dependency' (backlog dependency order)",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"default": "text",
				},
				"stale_threshold_hours": map[string]interface{}{
					"type":    "number",
					"default": 2.0,
				},
				"include_legacy": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "If true, also identify and remove legacy tasks with old sequential IDs (T-1, T-2, etc.)",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Task name (required for create action)",
				},
				"long_description": map[string]interface{}{
					"type":        "string",
					"description": "Task description (required for create action)",
				},
				"tags": map[string]interface{}{
					"type":        []interface{}{"array", "string"},
					"description": "Task tags (array of strings or comma-separated string)",
				},
				"remove_tags": map[string]interface{}{
					"type":        []interface{}{"array", "string"},
					"description": "Tags to remove from task(s). For action=update: array or comma-separated string.",
				},
				"dependencies": map[string]interface{}{
					"type":        []interface{}{"array", "string"},
					"description": "Task dependencies (array of task IDs or comma-separated string)",
				},
				"auto_estimate": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Automatically estimate task duration and add as comment (default: true)",
				},
				"planning_doc": map[string]interface{}{
					"type":        "string",
					"description": "Path to planning document. For link_planning: optional, stored in task metadata. For sync_from_plan/sync_plan_status: required (.plan.md path).",
				},
				"write_plan": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "For sync_from_plan: when true (default), update plan file checkboxes and frontmatter from Todo2 status (bidirectional).",
				},
				"epic_id": map[string]interface{}{
					"type":        "string",
					"description": "Epic task ID if this task is part of an epic (optional, stored in task metadata and parent_id)",
				},
				"parent_id": map[string]interface{}{
					"type":        "string",
					"description": "Parent task ID for hierarchy (optional; separate from blocking dependencies). For create/update/link_planning.",
				},
			},
		},
		handleTaskWorkflow,
	); err != nil {
		return fmt.Errorf("failed to register task_workflow: %w", err)
	}

	// T-34b: infer_task_progress
	if err := server.RegisterTool(
		"infer_task_progress",
		"[HINT: Task completion inference. Analyzes In Progress tasks against codebase; returns inferred completions. dry_run, auto_update_tasks, output_path.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"project_root": map[string]interface{}{
					"type": "string",
				},
				"scan_depth": map[string]interface{}{
					"type":    "number",
					"default": 3,
				},
				"file_extensions": map[string]interface{}{
					"type": "array",
				},
				"confidence_threshold": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"auto_update_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"use_fm": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
			},
		},
		handleInferTaskProgress,
	); err != nil {
		return fmt.Errorf("failed to register infer_task_progress: %w", err)
	}

	// T-35: testing
	if err := server.RegisterTool(
		"testing",
		"[HINT: Testing tool. action=run|coverage|suggest|validate. Execute tests, analyze coverage, suggest test cases, or validate test structure.]",
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

	return nil
}

// registerBatch3Tools registers Batch 3 tools (8 advanced tools)
func registerBatch3Tools(server framework.MCPServer) error {
	// T-37: automation
	if err := server.RegisterTool(
		"automation",
		"[HINT: Automation. action=daily|nightly|sprint|discover. Unified automation tool.]",
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
			},
		},
		handleAutomation,
	); err != nil {
		return fmt.Errorf("failed to register automation: %w", err)
	}

	// T-38: tool_catalog (help action only - list action converted to stdio://tools resources)
	if err := server.RegisterTool(
		"tool_catalog",
		"[HINT: Tool catalog. action=help. Get help for a specific tool. Use stdio://tools resources for listing tools.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"help"},
					"default": "help",
				},
				"tool_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the tool to get help for (required)",
				},
			},
			Required: []string{"tool_name"},
		},
		handleToolCatalog,
	); err != nil {
		return fmt.Errorf("failed to register tool_catalog: %w", err)
	}

	// T-39: workflow_mode
	if err := server.RegisterTool(
		"workflow_mode",
		"[HINT: Workflow mode management. action=focus|suggest|stats. Unified workflow operations.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"focus", "suggest", "stats"},
					"default": "focus",
				},
				"mode": map[string]interface{}{
					"type": "string",
				},
				"enable_group": map[string]interface{}{
					"type": "string",
				},
				"disable_group": map[string]interface{}{
					"type": "string",
				},
				"status": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"text": map[string]interface{}{
					"type": "string",
				},
				"auto_switch": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleWorkflowMode,
	); err != nil {
		return fmt.Errorf("failed to register workflow_mode: %w", err)
	}

	// T-40: lint
	if err := server.RegisterTool(
		"lint",
		"[HINT: Linting tool. action=run|analyze. Run linter or analyze problems. For markdown/docs, includes broken link check via gomarklint (enableLinkCheck in .gomarklint.json).]",
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
					"type":    "string",
					"default": "golangci-lint",
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

	// T-41: estimation
	if err := server.RegisterTool(
		"estimation",
		"[HINT: Estimation. action=estimate|analyze|stats|estimate_batch. Unified task duration estimation tool.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"estimate", "analyze", "stats", "estimate_batch"},
					"default": "estimate",
				},
				"task_ids": map[string]interface{}{
					"type":        "array",
					"description": "For estimate_batch: list of task IDs to estimate (or omit with status_filter for all matching)",
				},
				"status_filter": map[string]interface{}{
					"type":        "string",
					"description": "For estimate_batch: e.g. 'Todo' to estimate all Todo tasks (max 50)",
				},
				"name": map[string]interface{}{
					"type": "string",
				},
				"details": map[string]interface{}{
					"type":    "string",
					"default": "",
				},
				"tags": map[string]interface{}{
					"type": "string",
				},
				"tag_list": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
				},
				"priority": map[string]interface{}{
					"type":    "string",
					"default": "medium",
				},
				"use_historical": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"detailed": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"use_mlx": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"mlx_weight": map[string]interface{}{
					"type":    "number",
					"default": 0.3,
				},
			},
		},
		handleEstimation,
	); err != nil {
		return fmt.Errorf("failed to register estimation: %w", err)
	}

	// T-42: git_tools
	if err := server.RegisterTool(
		"git_tools",
		"[HINT: Git tools. action=commits|local_commits|branches|tasks|diff|graph|merge|set_branch. Unified git-inspired tools.]",
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

	// T-43: session
	if err := server.RegisterTool(
		"session",
		"[HINT: Session. action=prime|handoff|prompts|assignee. Use ask_preferences=true for MCP elicitation at prime. Unified session management tools.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"prime", "handoff", "prompts", "assignee"},
					"default": "prime",
				},
				"include_hints": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"ask_preferences": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true and client supports elicitation, prompt user for include_tasks/include_hints preferences at prime time",
				},
				"override_mode": map[string]interface{}{
					"type": "string",
				},
				"task_id": map[string]interface{}{
					"type": "string",
				},
				"summary": map[string]interface{}{
					"type": "string",
				},
				"blockers": map[string]interface{}{
					"type": "string",
				},
				"next_steps": map[string]interface{}{
					"type": "string",
				},
				"unassign_my_tasks": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"include_git_status": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"limit": map[string]interface{}{
					"type":    "integer",
					"default": 5,
				},
				"dry_run": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"direction": map[string]interface{}{
					"type":    "string",
					"default": "both",
				},
				"prefer_agentic_tools": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"auto_commit": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"mode": map[string]interface{}{
					"type": "string",
				},
				"category": map[string]interface{}{
					"type": "string",
				},
				"keywords": map[string]interface{}{
					"type": "string",
				},
				"assignee_name": map[string]interface{}{
					"type": "string",
				},
				"assignee_type": map[string]interface{}{
					"type":    "string",
					"default": "agent",
				},
				"hostname": map[string]interface{}{
					"type": "string",
				},
				"status_filter": map[string]interface{}{
					"type": "string",
				},
				"sub_action": map[string]interface{}{
					"type": "string",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"export_latest": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"priority_filter": map[string]interface{}{
					"type": "string",
				},
				"include_unassigned": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"max_tasks_per_agent": map[string]interface{}{
					"type":    "integer",
					"default": 5,
				},
			},
		},
		handleSession,
	); err != nil {
		return fmt.Errorf("failed to register session: %w", err)
	}

	// T-44: infer_session_mode
	if err := server.RegisterTool(
		"infer_session_mode",
		"[HINT: Session mode inference. Returns AGENT/ASK/MANUAL with confidence.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"force_recompute": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleInferSessionMode,
	); err != nil {
		return fmt.Errorf("failed to register infer_session_mode: %w", err)
	}

	// T-6: MLX Integration tools (ollama and mlx)
	// ollama
	if err := server.RegisterTool(
		"ollama",
		"[HINT: LLM abstraction. ollama. action=status|models|generate|pull|hardware|docs|quality|summary. Native then bridge (DefaultOllama()).]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"status", "models", "generate", "pull", "hardware", "docs", "quality", "summary"},
					"default": "status",
				},
				"host": map[string]interface{}{
					"type": "string",
				},
				"prompt": map[string]interface{}{
					"type": "string",
				},
				"model": map[string]interface{}{
					"type":    "string",
					"default": "llama3.2",
				},
				"stream": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
				"options": map[string]interface{}{
					"type": "string",
				},
				"num_gpu": map[string]interface{}{
					"type": "integer",
				},
				"num_threads": map[string]interface{}{
					"type": "integer",
				},
				"context_size": map[string]interface{}{
					"type": "integer",
				},
				"file_path": map[string]interface{}{
					"type": "string",
				},
				"output_path": map[string]interface{}{
					"type": "string",
				},
				"style": map[string]interface{}{
					"type":    "string",
					"default": "google",
				},
				"include_suggestions": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
				"data": map[string]interface{}{
					"type": "string",
				},
				"level": map[string]interface{}{
					"type":    "string",
					"default": "brief",
				},
			},
		},
		handleOllama,
	); err != nil {
		return fmt.Errorf("failed to register ollama: %w", err)
	}

	// mlx
	if err := server.RegisterTool(
		"mlx",
		"[HINT: LLM abstraction (MLX). action=status|hardware|models|generate. Native status/hardware/models; generate via shared path. text_generate provider=mlx uses DefaultMLXProvider(); report insights use DefaultReportInsight() (FM then MLX).]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"status", "hardware", "models", "generate"},
					"default": "status",
				},
				"prompt": map[string]interface{}{
					"type": "string",
				},
				"model": map[string]interface{}{
					"type":    "string",
					"default": "mlx-community/Phi-3.5-mini-instruct-4bit",
				},
				"max_tokens": map[string]interface{}{
					"type":    "integer",
					"default": 512,
				},
				"temperature": map[string]interface{}{
					"type":    "number",
					"default": 0.7,
				},
				"verbose": map[string]interface{}{
					"type":    "boolean",
					"default": false,
				},
			},
		},
		handleMlx,
	); err != nil {
		return fmt.Errorf("failed to register mlx: %w", err)
	}

	// Apple Foundation Models tool (platform-specific, conditional compilation)
	if err := registerAppleFoundationModelsTool(server); err != nil {
		return err
	}

	return nil
}

// registerBatch4Tools registers Batch 4 tools (2 native Go tools from mcp-generic-tools migration)
// Note: Individual Python bridge tools (context_summarize, context_batch, prompt_log, prompt_analyze,
// recommend_model, recommend_workflow) were removed in favor of unified tools in Batch 5:
// - context(action=summarize|budget|batch)
// - prompt_tracking(action=log|analyze)
// - recommend(action=model|workflow|advisor)
func registerBatch4Tools(server framework.MCPServer) error {
	// Native Go tools (2)

	// context_budget
	if err := server.RegisterTool(
		"context_budget",
		"[HINT: Context budget. Estimate token usage and suggest context reduction strategy.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"items": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of items to analyze",
				},
				"budget_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     4000,
					"description": "Target token budget",
				},
			},
			Required: []string{"items"},
		},
		handleContextBudget,
	); err != nil {
		return fmt.Errorf("failed to register context_budget: %w", err)
	}

	// Note: list_models tool removed - converted to stdio://models resource
	// See internal/resources/models.go for resource implementation

	return nil
}

// registerBatch5Tools registers Batch 5 tools (Phase 3 migration - 4 unified tools)
// Note: demonstrate_elicit and interactive_task_create were removed (required FastMCP Context)
func registerBatch5Tools(server framework.MCPServer) error {
	// context - Unified context management (summarize/budget/batch actions)
	if err := server.RegisterTool(
		"context",
		"[HINT: Context management. action=summarize|budget|batch. Unified context operations.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"summarize", "budget", "batch"},
					"default":     "summarize",
					"description": "Action to perform",
				},
				"data": map[string]interface{}{
					"type":        "string",
					"description": "JSON string to summarize (summarize action)",
				},
				"level": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"brief", "detailed", "key_metrics", "actionable"},
					"default":     "brief",
					"description": "Summarization level (summarize action)",
				},
				"tool_type": map[string]interface{}{
					"type":        "string",
					"description": "Tool type hint for smarter summarization (summarize action)",
				},
				"max_tokens": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum tokens for output (summarize action)",
				},
				"include_raw": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "Include original data in response (summarize action)",
				},
				"items": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of items to analyze (budget/batch actions)",
				},
				"budget_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     4000,
					"description": "Target token budget (budget action)",
				},
				"combine": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Merge summaries into combined view (batch action)",
				},
			},
		},
		handleContext,
	); err != nil {
		return fmt.Errorf("failed to register context: %w", err)
	}

	// text_generate - Unified generate-text via FM, ReportInsight, or MLX (LLM abstraction)
	// When task_type or task_description is provided, uses ResolveModelForTask (recommend + router) for model selection (T-207).
	if err := server.RegisterTool(
		"text_generate",
		"[HINT: Unified generate-text. provider=fm|insight|mlx|auto (default fm). task_type/task_description enable model selection (recommend+router).]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"provider": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"fm", "insight", "mlx", "auto"},
					"default":     "fm",
					"description": "Backend: fm|insight|mlx (explicit), or auto (model selection from task_type/task_description)",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Prompt for text generation (required)",
				},
				"task_type": map[string]interface{}{
					"type":        "string",
					"description": "Task type hint for model selection (e.g. code_analysis, quick_fix). Used with provider=auto.",
				},
				"task_description": map[string]interface{}{
					"type":        "string",
					"description": "Task description for model selection. Used with provider=auto.",
				},
				"optimize_for": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"quality", "speed", "cost"},
					"default":     "quality",
					"description": "Optimization target for model selection (when provider=auto)",
				},
				"max_tokens": map[string]interface{}{
					"type":        "integer",
					"default":     512,
					"description": "Maximum tokens to generate",
				},
				"temperature": map[string]interface{}{
					"type":        "number",
					"default":     0.7,
					"description": "Sampling temperature",
				},
			},
			Required: []string{"prompt"},
		},
		handleTextGenerate,
	); err != nil {
		return fmt.Errorf("failed to register text_generate: %w", err)
	}

	// prompt_tracking - Unified prompt tracking (log/analyze actions)
	if err := server.RegisterTool(
		"prompt_tracking",
		"[HINT: Prompt tracking. action=log|analyze. Track and analyze prompts.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"log", "analyze"},
					"default":     "analyze",
					"description": "Action to perform",
				},
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Prompt text to log (log action)",
				},
				"task_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional task ID",
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Optional mode",
				},
				"outcome": map[string]interface{}{
					"type":        "string",
					"description": "Optional outcome",
				},
				"iteration": map[string]interface{}{
					"type":        "integer",
					"default":     1,
					"description": "Iteration number (log action)",
				},
				"days": map[string]interface{}{
					"type":        "integer",
					"default":     7,
					"description": "Number of days to analyze (analyze action)",
				},
			},
		},
		handlePromptTracking,
	); err != nil {
		return fmt.Errorf("failed to register prompt_tracking: %w", err)
	}

	// recommend - Unified recommendation tool (model/workflow/advisor actions)
	if err := server.RegisterTool(
		"recommend",
		"[HINT: Recommend. action=model|workflow|advisor. Unified recommendation tool.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"model", "workflow", "advisor"},
					"default":     "model",
					"description": "Action to perform",
				},
				"task_description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the task",
				},
				"tags": map[string]interface{}{
					"type":        "string",
					"description": "Optional JSON string of tags to consider",
				},
				"include_rationale": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Whether to include detailed reasoning",
				},
				"task_type": map[string]interface{}{
					"type":        "string",
					"description": "Optional explicit task type (model action)",
				},
				"optimize_for": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"quality", "speed", "cost"},
					"default":     "quality",
					"description": "Optimization target (model action)",
				},
				"include_alternatives": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Include alternative recommendations (model action)",
				},
			},
		},
		handleRecommend,
	); err != nil {
		return fmt.Errorf("failed to register recommend: %w", err)
	}

	// T-224: research_aggregator - runs multiple research tools and combines outputs
	if err := server.RegisterTool(
		"research_aggregator",
		"[HINT: Research result aggregator. Runs task_analysis, analyze_alignment, etc. and combines outputs.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"tools": map[string]interface{}{
					"type": "array",
					"description": "Tool configs: [{tool, action}] or tool names. Default: duplicates, dependencies, todo2.",
				},
			},
		},
		handleResearchAggregator,
	); err != nil {
		return fmt.Errorf("failed to register research_aggregator: %w", err)
	}

	// Note: server_status tool removed - converted to stdio://server/status resource
	// See internal/resources/server.go for resource implementation

	// Note: demonstrate_elicit and interactive_task_create removed
	// These tools require FastMCP Context (not available in stdio mode)
	// They were demonstration tools that don't work in exarp-go's primary stdio mode

	return nil
}
