// registry_batch2.go — MCP tool registration batch.
package tools

import (
	"fmt"

	"github.com/davidl71/exarp-go/internal/framework"
)

// registerBatch2Tools registers Batch 2 tools (8 medium tools).
func registerBatch2Tools(server framework.MCPServer) error {
	// T-28: memory
	if err := server.RegisterTool(
		"memory",
		"[HINT: action=save|recall|search. Persist and retrieve AI discoveries. Use when saving learnings or searching past context.]",
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
		"[HINT: action=health|gc|prune|consolidate|dream. Memory lifecycle management. Use when cleaning up old memories or consolidating insights.]",
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
		"[HINT: action=overview|scorecard|briefing|prd|plan. Project reports and plans. Use for project status, scorecard, or generating .plan.md files. Related: task_analysis.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"overview", "scorecard", "briefing", "prd", "plan", "scorecard_plans", "parallel_execution_plan", "update_waves_from_plan"},
					"default":     "overview",
					"description": "plan: write .plan.md; scorecard_plans: improve-<dim>.plan.md; parallel_execution_plan: parallel-execution-subagents.plan.md",
				},
				"output_format": map[string]interface{}{
					"type":    "string",
					"default": "text",
				},
				"compact": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true and output_format=json, return compact JSON to reduce context size (overview, scorecard, briefing)",
				},
				"output_path": map[string]interface{}{
					"type":        "string",
					"description": "For action=plan: path for plan file (default: .cursor/plans/<project-slug>.plan.md)",
				},
				"plan_title": map[string]interface{}{
					"type":        "string",
					"description": "For action=plan: title for the plan (default: project name from go.mod or directory)",
				},
				"include_subagents": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=plan: when true, also update .cursor/plans/parallel-execution-subagents.plan.md from waves",
				},
				"repair": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=plan: when true, repair existing plan file (restore frontmatter and ## 3. Iterative Milestones) without overwriting rest of body",
				},
				"plan_path": map[string]interface{}{
					"type":        "string",
					"description": "For action=plan with repair=true: path to plan file to repair (default: same as output_path)",
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
				"skip_scorecard_cache": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=scorecard: when true, bypass 5-minute result cache and regenerate",
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

	// T-32: task_analysis
	if err := server.RegisterTool(
		"task_analysis",
		"[HINT: action=duplicates|tags|discover_tags|dependencies|execution_plan|complexity|conflicts|noise. Analyze task backlog. Use when planning sprints, detecting duplicates, or generating execution waves. Related: task_workflow.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"duplicates", "tags", "discover_tags", "hierarchy", "dependencies", "dependencies_summary", "suggest_dependencies", "parallelization", "fix_missing_deps", "validate", "execution_plan", "complexity", "conflicts", "noise"},
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
					"description": "For action=noise: filter to tasks with this tag (default: discovered when empty). For action=execution_plan: restrict backlog to tasks with this tag.",
				},
				"filter_tags": map[string]interface{}{
					"type":        "string",
					"description": "For action=execution_plan: restrict backlog to tasks with any of these tags (comma-separated).",
				},
				"include_planning_docs": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "For action=suggest_dependencies: when true, also extract dependency hints from .cursor/plans and docs/*plan*.md (Depends on: T-XXX, milestone order).",
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
		"[HINT: action=comments|markdown|orphans|git_json|planning_links|all. Discover tasks from code TODOs and docs. Use when scanning for undocumented work. create_tasks=true to auto-create.]",
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
		"[HINT: action=sync|approve|create|update|delete|clarify|cleanup|summarize|run_with_ai|link_planning. Task lifecycle management. Use for CRUD, batch status updates (approve+task_ids), AI summaries. Prefer exarp-go task CLI for simple ops. Related: task_analysis, session.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"action": map[string]interface{}{
					"type":    "string",
					"enum":    []string{"sync", "approve", "clarify", "clarity", "cleanup", "create", "delete", "enrich_tool_hints", "fix_dates", "fix_empty_descriptions", "fix_invalid_ids", "link_planning", "request_approval", "sync_approvals", "apply_approval_result", "sanity_check", "sync_from_plan", "sync_plan_status", "update", "summarize", "run_with_ai"},
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
				"compact": map[string]interface{}{
					"type":        "boolean",
					"default":     false,
					"description": "When true and output_format=json, return compact JSON (no indentation) to reduce context size",
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
					"description": "Task name (required for single create; omit when using tasks array)",
				},
				"long_description": map[string]interface{}{
					"type":        "string",
					"description": "Task description (for single create; omit when using tasks array)",
				},
				"tasks": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of tasks for batch create. Each element: {name, priority?, tags?, long_description?, dependencies?}. Example: [{\"name\":\"Task A\",\"priority\":\"high\"},{\"name\":\"Task B\"}]",
				},
				"tags": map[string]interface{}{
					"type":        "string",
					"description": "Task tags as comma-separated values (e.g. 'backend,urgent') or JSON array encoded as string (e.g. '[\"backend\",\"urgent\"]')",
				},
				"remove_tags": map[string]interface{}{
					"type":        "string",
					"description": "Tags to remove from task(s). For action=update: comma-separated values or JSON array encoded as string.",
				},
				"dependencies": map[string]interface{}{
					"type":        "string",
					"description": "Task dependencies as comma-separated task IDs or JSON array encoded as string (e.g. '[\"T-1\",\"T-2\"]')",
				},
				"auto_estimate": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "Automatically estimate task duration and add as comment (default: true)",
				},
				"local_ai_backend": map[string]interface{}{
					"type":        "string",
					"description": "For create/update: preferred local LLM for estimation (fm|mlx|ollama). Stored in task metadata as preferred_backend. For summarize/run_with_ai: overrides task metadata to select backend.",
					"enum":        []string{"", "fm", "mlx", "ollama"},
				},
				"recommended_tools": map[string]interface{}{
					"type":        "string",
					"description": "For create/update: comma-separated MCP tool IDs to suggest for this task (e.g. report, task_workflow). Stored in task metadata as recommended_tools; exposed in task show and session prime suggested_next.",
				},
				"instruction": map[string]interface{}{
					"type":        "string",
					"description": "For run_with_ai: custom instruction/question for the LLM about the task. Defaults to implementation plan + risks + next steps.",
				},
				"save_comment": map[string]interface{}{
					"type":        "boolean",
					"default":     true,
					"description": "For summarize: when true (default), save generated summary as a task comment.",
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
		"[HINT: Analyze tasks against codebase to infer completions. status_filter (default: In Progress) can be set to Todo, Review, etc. Use when checking if tasks are already done. dry_run, auto_update_tasks.]",
		framework.ToolSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"project_root": map[string]interface{}{
					"type": "string",
				},
				"status_filter": map[string]interface{}{
					"type":        "string",
					"default":     "In Progress",
					"description": "Filter tasks by status (Todo, In Progress, Review, Done, Cancelled). Default: In Progress",
				},
				"scan_depth": map[string]interface{}{
					"type":    "number",
					"default": 3,
				},
				"file_extensions": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "string"},
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

	return nil
}
