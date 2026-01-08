"""
Consolidated MCP Tools

Combines related tools into unified interfaces with action parameters.
Reduces tool count while maintaining all functionality.

All tools use 'action' as the dispatcher parameter for consistency.

Consolidated tools:
- analyze_alignment(action=todo2|prd) ← analyze_todo2_alignment, analyze_prd_alignment
- automation(action=daily|nightly|sprint|discover) ← run_daily_automation, run_nightly_automation, run_sprint_automation, run_discover_automation
- estimation(action=estimate|analyze|stats) ← estimate_task_duration, analyze_estimation_accuracy, get_estimation_statistics
- security(action=scan|alerts|report) ← scan_dependency_security, fetch_dependabot_alerts, generate_security_report
- generate_config(action=rules|ignore|simplify) ← generate_cursor_rules, generate_cursorignore, simplify_rules
- setup_hooks(action=git|patterns) ← setup_git_hooks, setup_pattern_triggers
- prompt_tracking(action=log|analyze) ← log_prompt_iteration, analyze_prompt_iterations
- health(action=server|git|docs|dod|cicd) ← server_status, check_working_copy_health, check_documentation_health, check_definition_of_done, validate_ci_cd_workflow
- report(action=overview|scorecard|briefing|prd) ← generate_project_overview, generate_project_scorecard, get_daily_briefing, generate_prd
- advisor_audio removed - migrated to devwisdom-go MCP server
- task_analysis(action=duplicates|tags|hierarchy|dependencies|parallelization) ← detect_duplicate_tasks, consolidate_tags, analyze_task_hierarchy, analyze_todo2_dependencies, optimize_todo2_parallelization
- testing(action=run|coverage|suggest|generate|validate) ← run_tests, analyze_test_coverage, suggest_test_cases, generate_test_code (AI), validate_test_structure
- lint(action=run|analyze) ← run_linter, analyze_problems
- memory(action=save|recall|search) ← save_memory, recall_context, search_memories
- memory_maint(action=health|gc|prune|consolidate|dream) ← memory lifecycle management and advisor dreaming
- task_discovery(action=comments|markdown|orphans|all) ← NEW: find tasks from various sources
- task_workflow(action=sync|approve|clarify|clarity|cleanup, sub_action for clarify) ← sync_todo_tasks, batch_approve_tasks, clarification, improve_task_clarity, cleanup_stale_tasks
- context(action=summarize|budget|batch) ← moved to context_tool.py
- tool_catalog(action=list|help) ← list_tools, get_tool_help
- workflow_mode(action=focus|suggest|stats) ← focus_mode, suggest_mode, get_tool_usage_stats
- recommend(action=model|workflow|advisor) ← recommend_model, recommend_workflow_mode, consult_advisor
"""

import asyncio
import json
import logging
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


"""
Consolidated Memory

Consolidated tools: memory, memory_maint
"""

def memory(
    action: str = "search",
    # save params
    title: Optional[str] = None,
    content: Optional[str] = None,
    category: str = "insight",
    task_id: Optional[str] = None,
    metadata: Optional[str] = None,  # JSON string
    # recall params
    include_related: bool = True,
    # search params
    query: Optional[str] = None,
    limit: int = 10,
) -> str:
    """
    Unified memory management tool.

    Args:
        action: "save" to store insight, "recall" to get task context, "search" to find memories
        title: Memory title (save action, required)
        content: Memory content (save action, required)
        category: One of: debug, research, architecture, preference, insight (save action)
        task_id: Task ID to link memory to (save action) or recall context for (recall action)
        metadata: Additional metadata as JSON string (save action)
        include_related: Include related task memories (recall action)
        query: Search query text (search action)
        limit: Maximum results (search action)

    Returns:
        JSON string with memory operation results
    """
    if action == "save":
        if not title or not content:
            return json.dumps({
                "success": False,
                "error": "title and content are required for save action",
            }, indent=2)
        from .session_memory import save_session_insight
        meta = None
        if metadata:
            try:
                meta = json.loads(metadata)
            except json.JSONDecodeError:
                return json.dumps({"success": False, "error": "Invalid metadata JSON"}, indent=2)
        result = save_session_insight(title, content, category, task_id, meta)
        # Ensure we always return a JSON string
        if isinstance(result, str):
            return result
        elif isinstance(result, dict):
            return json.dumps(result, indent=2)
        else:
            return json.dumps({"result": str(result)}, indent=2)
    elif action == "recall":
        if not task_id:
            return json.dumps({
                "success": False,
                "error": "task_id is required for recall action",
            }, indent=2)
        from .session_memory import recall_task_context
        result = recall_task_context(task_id, include_related)
        # Ensure we always return a JSON string
        if isinstance(result, str):
            return result
        elif isinstance(result, dict):
            return json.dumps(result, indent=2)
        else:
            return json.dumps({"result": str(result)}, indent=2)
    elif action == "search":
        if not query:
            return json.dumps({
                "success": False,
                "error": "query is required for search action",
            }, indent=2)
        from .session_memory import search_session_memories
        result = search_session_memories(query, category if category != "insight" else None, limit)
        # Ensure we always return a JSON string
        if isinstance(result, str):
            return result
        elif isinstance(result, dict):
            return json.dumps(result, indent=2)
        else:
            return json.dumps({"result": str(result)}, indent=2)
    else:
        return json.dumps({
            "success": False,
            "error": f"Unknown memory action: {action}. Use 'save', 'recall', or 'search'.",
        }, indent=2)


# context tool moved to context_tool.py for better modularity



def memory_maint(
    action: str = "health",
    # gc params
    max_age_days: int = 90,
    delete_orphaned: bool = True,
    delete_duplicates: bool = True,
    scorecard_max_age_days: int = 7,
    # prune params
    value_threshold: float = 0.3,
    keep_minimum: int = 50,
    # consolidate params
    similarity_threshold: float = 0.85,
    merge_strategy: str = "newest",
    # dream params
    scope: str = "week",
    advisors: Optional[str] = None,
    generate_insights: bool = True,
    save_dream: bool = True,
    # common
    dry_run: bool = True,
    interactive: bool = True,
) -> str:
    """
    Unified memory maintenance tool.

    Args:
        action: "health", "gc", "prune", "consolidate", or "dream"
        max_age_days: Delete memories older than this (gc)
        delete_orphaned: Delete orphaned memories (gc)
        delete_duplicates: Delete duplicates (gc)
        scorecard_max_age_days: Max age for scorecard memories (gc)
        value_threshold: Minimum value score to keep (prune)
        keep_minimum: Always keep at least N memories (prune)
        similarity_threshold: Title similarity threshold (consolidate)
        merge_strategy: newest, oldest, or longest (consolidate)
        scope: day, week, month, or all (dream)
        advisors: JSON list of advisor keys (dream)
        generate_insights: Generate actionable insights (dream)
        save_dream: Save dream as new memory (dream)
        dry_run: Preview without executing (default True)
        interactive: Use interactive MCP for approvals (reserved for future)

    Returns:
        JSON string with maintenance operation results
    """
    if action == "health":
        from .memory_maintenance import memory_health_check
        result = memory_health_check()
        return json.dumps(result, indent=2) if isinstance(result, dict) else result

    elif action == "gc":
        from .memory_maintenance import memory_garbage_collect
        result = memory_garbage_collect(
            max_age_days=max_age_days,
            delete_orphaned=delete_orphaned,
            delete_duplicates=delete_duplicates,
            scorecard_max_age_days=scorecard_max_age_days,
            dry_run=dry_run,
        )
        return json.dumps(result, indent=2) if isinstance(result, dict) else result

    elif action == "prune":
        from .memory_maintenance import memory_prune
        result = memory_prune(
            value_threshold=value_threshold,
            keep_minimum=keep_minimum,
            dry_run=dry_run,
        )
        return json.dumps(result, indent=2) if isinstance(result, dict) else result

    elif action == "consolidate":
        from .memory_maintenance import memory_consolidate
        result = memory_consolidate(
            similarity_threshold=similarity_threshold,
            merge_strategy=merge_strategy,
            dry_run=dry_run,
        )
        return json.dumps(result, indent=2) if isinstance(result, dict) else result

    elif action == "dream":
        from .memory_dreaming import memory_dream
        advisor_list = None
        if advisors:
            try:
                advisor_list = json.loads(advisors)
            except json.JSONDecodeError:
                return json.dumps({"status": "error", "error": "Invalid advisors JSON"}, indent=2)
        result = memory_dream(
            scope=scope,
            advisors=advisor_list,
            generate_insights=generate_insights,
            save_dream=save_dream,
        )
        return json.dumps(result, indent=2) if isinstance(result, dict) else result

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown memory_maint action: {action}. Use 'health', 'gc', 'prune', 'consolidate', or 'dream'.",
        }, indent=2)


