#!/usr/bin/env python3
"""
Python Bridge Tool Executor

Executes Python tools from the Go MCP server via subprocess.
This bridge allows the Go server to execute existing Python tools.
"""

import json
import sys
import os
from pathlib import Path

# Add local project_management_automation to path (copied to exarp-go)
PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Add bridge directory to path for mcp-generic-tools modules
BRIDGE_ROOT = Path(__file__).parent
sys.path.insert(0, str(BRIDGE_ROOT))

def execute_tool(tool_name: str, args_json: str):
    """Execute a Python tool with given arguments."""
    try:
        # Parse arguments
        args = json.loads(args_json) if args_json else {}
        
        # Import tool functions from project_management_automation
        # Note: tool_catalog, workflow_mode, git_tools, and infer_session_mode removed
        # These tools are fully native Go with no Python fallback
        from project_management_automation.tools.consolidated import (
            analyze_alignment as _analyze_alignment,
            generate_config as _generate_config,
            health as _health,
            setup_hooks as _setup_hooks,
            memory as _memory,
            memory_maint as _memory_maint,
            report as _report,
            security as _security,
            task_analysis as _task_analysis,
            task_discovery as _task_discovery,
            task_workflow as _task_workflow,
            testing as _testing,
            automation as _automation,
            estimation as _estimation,
            lint as _lint,
            mlx as _mlx,
            ollama as _ollama,
            session as _session,
        )
        from project_management_automation.tools.external_tool_hints import (
            add_external_tool_hints as _add_external_tool_hints,
        )
        from project_management_automation.tools.attribution_check import (
            check_attribution_compliance as _check_attribution_compliance,
        )
        
        # Import mcp-generic-tools modules (from bridge directory)
        from context.tool import context as _context
        from recommend.tool import recommend as _recommend
        # Note: prompt_tracking removed - fully native Go with no Python fallback
        
        # Import context_tool for unified context wrapper
        from project_management_automation.tools.context_tool import context as _context_unified
        
        # Route to appropriate tool
        if tool_name == "analyze_alignment":
            result = _analyze_alignment(
                action=args.get("action", "todo2"),
                create_followup_tasks=args.get("create_followup_tasks", True),
                output_path=args.get("output_path"),
            )
        elif tool_name == "generate_config":
            result = _generate_config(
                action=args.get("action", "rules"),
                rules=args.get("rules"),
                overwrite=args.get("overwrite", False),
                analyze_only=args.get("analyze_only", False),
                include_indexing=args.get("include_indexing", True),
                analyze_project=args.get("analyze_project", True),
                rule_files=args.get("rule_files"),
                output_dir=args.get("output_dir"),
                dry_run=args.get("dry_run", False),
            )
        elif tool_name == "health":
            result = _health(
                action=args.get("action", "server"),
                agent_name=args.get("agent_name"),
                check_remote=args.get("check_remote", True),
                output_path=args.get("output_path"),
                create_tasks=args.get("create_tasks", True),
                task_id=args.get("task_id"),
                changed_files=args.get("changed_files"),
                auto_check=args.get("auto_check", True),
                workflow_path=args.get("workflow_path"),
                check_runners=args.get("check_runners", True),
            )
        elif tool_name == "setup_hooks":
            result = _setup_hooks(
                action=args.get("action", "git"),
                hooks=args.get("hooks"),
                patterns=args.get("patterns"),
                config_path=args.get("config_path"),
                install=args.get("install", True),
                dry_run=args.get("dry_run", False),
            )
        elif tool_name == "check_attribution":
            result = _check_attribution_compliance(
                output_path=args.get("output_path"),
                create_tasks=args.get("create_tasks", True),
            )
        elif tool_name == "add_external_tool_hints":
            result = _add_external_tool_hints(
                dry_run=args.get("dry_run", False),
                output_path=args.get("output_path"),
                min_file_size=args.get("min_file_size", 50),
            )
        elif tool_name == "memory":
            result = _memory(
                action=args.get("action", "search"),
                title=args.get("title"),
                content=args.get("content"),
                category=args.get("category", "insight"),
                task_id=args.get("task_id"),
                metadata=args.get("metadata"),
                include_related=args.get("include_related", True),
                query=args.get("query"),
                limit=args.get("limit", 10),
            )
        elif tool_name == "memory_maint":
            result = _memory_maint(
                action=args.get("action", "health"),
                max_age_days=args.get("max_age_days", 90),
                delete_orphaned=args.get("delete_orphaned", True),
                delete_duplicates=args.get("delete_duplicates", True),
                scorecard_max_age_days=args.get("scorecard_max_age_days", 7),
                value_threshold=args.get("value_threshold", 0.3),
                keep_minimum=args.get("keep_minimum", 50),
                similarity_threshold=args.get("similarity_threshold", 0.85),
                merge_strategy=args.get("merge_strategy", "newest"),
                scope=args.get("scope", "week"),
                advisors=args.get("advisors"),
                generate_insights=args.get("generate_insights", True),
                save_dream=args.get("save_dream", True),
                dry_run=args.get("dry_run", True),
                interactive=args.get("interactive", True),
            )
        elif tool_name == "report":
            result = _report(
                action=args.get("action", "overview"),
                output_format=args.get("output_format", "text"),
                output_path=args.get("output_path"),
                include_recommendations=args.get("include_recommendations", True),
                overall_score=args.get("overall_score", 50.0),
                security_score=args.get("security_score", 50.0),
                testing_score=args.get("testing_score", 50.0),
                documentation_score=args.get("documentation_score", 50.0),
                completion_score=args.get("completion_score", 50.0),
                alignment_score=args.get("alignment_score", 50.0),
                project_name=args.get("project_name"),
                include_architecture=args.get("include_architecture", True),
                include_metrics=args.get("include_metrics", True),
                include_tasks=args.get("include_tasks", True),
            )
        elif tool_name == "security":
            result = _security(
                action=args.get("action", "report"),
                repo=args.get("repo", "davidl71/project-management-automation"),
                languages=args.get("languages"),
                config_path=args.get("config_path"),
                state=args.get("state", "open"),
                include_dismissed=args.get("include_dismissed", False),
            )
        elif tool_name == "task_analysis":
            result = _task_analysis(
                action=args.get("action", "duplicates"),
                similarity_threshold=args.get("similarity_threshold", 0.85),
                auto_fix=args.get("auto_fix", False),
                dry_run=args.get("dry_run", True),
                custom_rules=args.get("custom_rules"),
                remove_tags=args.get("remove_tags"),
                output_format=args.get("output_format", "text"),
                include_recommendations=args.get("include_recommendations", True),
                output_path=args.get("output_path"),
            )
        elif tool_name == "task_discovery":
            result = _task_discovery(
                action=args.get("action", "all"),
                file_patterns=args.get("file_patterns"),
                include_fixme=args.get("include_fixme", True),
                doc_path=args.get("doc_path"),
                output_path=args.get("output_path"),
                create_tasks=args.get("create_tasks", False),
            )
        elif tool_name == "task_workflow":
            result = _task_workflow(
                action=args.get("action", "sync"),
                dry_run=args.get("dry_run", False),
                status=args.get("status", "Review"),
                new_status=args.get("new_status", "Todo"),
                clarification_none=args.get("clarification_none", True),
                filter_tag=args.get("filter_tag"),
                task_ids=args.get("task_ids"),
                sub_action=args.get("sub_action", "list"),
                task_id=args.get("task_id"),
                clarification_text=args.get("clarification_text"),
                decision=args.get("decision"),
                decisions_json=args.get("decisions_json"),
                move_to_todo=args.get("move_to_todo", True),
                auto_apply=args.get("auto_apply", False),
                output_format=args.get("output_format", "text"),
                stale_threshold_hours=args.get("stale_threshold_hours", 2.0),
                output_path=args.get("output_path"),
            )
        elif tool_name == "testing":
            result = _testing(
                action=args.get("action", "run"),
                test_path=args.get("test_path"),
                test_framework=args.get("test_framework", "auto"),
                verbose=args.get("verbose", True),
                coverage=args.get("coverage", False),
                coverage_file=args.get("coverage_file"),
                min_coverage=args.get("min_coverage", 80),
                format=args.get("format", "html"),
                target_file=args.get("target_file"),
                min_confidence=args.get("min_confidence", 0.7),
                framework=args.get("framework"),
                output_path=args.get("output_path"),
            )
        elif tool_name == "automation":
            result = _automation(
                action=args.get("action", "daily"),
                tasks=args.get("tasks"),
                include_slow=args.get("include_slow", False),
                max_tasks_per_host=args.get("max_tasks_per_host", 5),
                max_parallel_tasks=args.get("max_parallel_tasks", 10),
                priority_filter=args.get("priority_filter"),
                tag_filter=args.get("tag_filter"),
                max_iterations=args.get("max_iterations", 10),
                auto_approve=args.get("auto_approve", True),
                extract_subtasks=args.get("extract_subtasks", True),
                run_analysis_tools=args.get("run_analysis_tools", True),
                run_testing_tools=args.get("run_testing_tools", True),
                min_value_score=args.get("min_value_score", 0.7),
                dry_run=args.get("dry_run", False),
                output_path=args.get("output_path"),
            )
        elif tool_name == "lint":
            result = _lint(
                action=args.get("action", "run"),
                path=args.get("path"),
                linter=args.get("linter", "ruff"),
                fix=args.get("fix", False),
                analyze=args.get("analyze", True),
                select=args.get("select"),
                ignore=args.get("ignore"),
                problems_json=args.get("problems_json"),
                include_hints=args.get("include_hints", True),
                output_path=args.get("output_path"),
            )
        elif tool_name == "estimation":
            result = _estimation(
                action=args.get("action", "estimate"),
                name=args.get("name"),
                details=args.get("details", ""),
                tags=args.get("tags"),
                tag_list=args.get("tag_list"),
                priority=args.get("priority", "medium"),
                use_historical=args.get("use_historical", True),
                detailed=args.get("detailed", False),
                use_mlx=args.get("use_mlx", True),
                mlx_weight=args.get("mlx_weight", 0.3),
            )
        elif tool_name == "session":
            result = _session(
                action=args.get("action", "prime"),
                include_hints=args.get("include_hints", True),
                include_tasks=args.get("include_tasks", True),
                override_mode=args.get("override_mode"),
                task_id=args.get("task_id"),
                summary=args.get("summary"),
                blockers=args.get("blockers"),
                next_steps=args.get("next_steps"),
                unassign_my_tasks=args.get("unassign_my_tasks", True),
                include_git_status=args.get("include_git_status", True),
                limit=args.get("limit", 5),
                dry_run=args.get("dry_run", False),
                direction=args.get("direction", "both"),
                prefer_agentic_tools=args.get("prefer_agentic_tools", True),
                auto_commit=args.get("auto_commit", True),
                mode=args.get("mode"),
                category=args.get("category"),
                keywords=args.get("keywords"),
                assignee_name=args.get("assignee_name"),
                assignee_type=args.get("assignee_type", "agent"),
                hostname=args.get("hostname"),
                status_filter=args.get("status_filter"),
                priority_filter=args.get("priority_filter"),
                include_unassigned=args.get("include_unassigned", False),
                max_tasks_per_agent=args.get("max_tasks_per_agent", 5),
            )
        elif tool_name == "ollama":
            result = _ollama(
                action=args.get("action", "status"),
                host=args.get("host"),
                prompt=args.get("prompt"),
                model=args.get("model", "llama3.2"),
                stream=args.get("stream", False),
                options=args.get("options"),
                num_gpu=args.get("num_gpu"),
                num_threads=args.get("num_threads"),
                context_size=args.get("context_size"),
                file_path=args.get("file_path"),
                output_path=args.get("output_path"),
                style=args.get("style", "google"),
                include_suggestions=args.get("include_suggestions", True),
                data=args.get("data"),
                level=args.get("level", "brief"),
            )
        elif tool_name == "mlx":
            result = _mlx(
                action=args.get("action", "status"),
                prompt=args.get("prompt"),
                model=args.get("model", "mlx-community/Phi-3.5-mini-instruct-4bit"),
                max_tokens=args.get("max_tokens", 512),
                temperature=args.get("temperature", 0.7),
                verbose=args.get("verbose", False),
            )
        # Phase 3 Migration: Unified tools
        # Note: Individual tools (context_summarize, context_batch, prompt_log, prompt_analyze,
        # recommend_model, recommend_workflow) were removed in favor of unified tools below
        elif tool_name == "context":
            result = _context_unified(
                action=args.get("action", "summarize"),
                data=args.get("data"),
                level=args.get("level", "brief"),
                tool_type=args.get("tool_type"),
                max_tokens=args.get("max_tokens"),
                include_raw=args.get("include_raw", False),
                items=args.get("items"),
                budget_tokens=args.get("budget_tokens", 4000),
                combine=args.get("combine", True),
            )
        elif tool_name == "prompt_tracking":
            # prompt_tracking is fully native Go - no Python bridge fallback
            error_result = {
                "success": False,
                "error": "prompt_tracking is fully native Go - Python bridge not available",
                "tool": tool_name
            }
            print(json.dumps(error_result, indent=2))
            return 1
        elif tool_name == "recommend":
            result = _recommend(
                action=args.get("action", "model"),
                task_description=args.get("task_description"),
                tags=args.get("tags"),
                include_rationale=args.get("include_rationale", True),
                task_type=args.get("task_type"),
                optimize_for=args.get("optimize_for", "quality"),
                include_alternatives=args.get("include_alternatives", True),
            )
        elif tool_name == "server_status":
            # server_status is fully native Go - no Python bridge fallback
            error_result = {
                "success": False,
                "error": "server_status is fully native Go - Python bridge not available",
                "tool": tool_name
            }
            print(json.dumps(error_result, indent=2))
            return 1
        # Note: demonstrate_elicit and interactive_task_create removed
        # These tools required FastMCP Context (not available in stdio mode)
        else:
            error_result = {
                "success": False,
                "error": f"Unknown tool: {tool_name}",
                "tool": tool_name
            }
            print(json.dumps(error_result, indent=2))
            return 1
        
        # Handle result - tools may return dict or JSON string
        if isinstance(result, dict):
            result_json = json.dumps(result, indent=2)
        elif isinstance(result, str):
            result_json = result
        else:
            result_json = json.dumps({"result": str(result)}, indent=2)
        
        print(result_json)
        return 0
        
    except Exception as e:
        error_result = {
            "success": False,
            "error": str(e),
            "tool": tool_name
        }
        print(json.dumps(error_result, indent=2))
        return 1

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Tool name required"}, indent=2))
        sys.exit(1)
    
    tool_name = sys.argv[1]
    args_json = sys.argv[2] if len(sys.argv) > 2 else "{}"
    
    sys.exit(execute_tool(tool_name, args_json))

