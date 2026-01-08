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
Consolidated Workflow

Consolidated tools: workflow_mode, recommend, tool_catalog
"""

def workflow_mode(
    action: str = "focus",
    # focus action params
    mode: Optional[str] = None,
    enable_group: Optional[str] = None,
    disable_group: Optional[str] = None,
    status: bool = False,
    # suggest action params
    text: Optional[str] = None,
    auto_switch: bool = False,
) -> str:
    """
    Unified workflow mode management tool.

    Consolidates workflow mode operations: focus, suggestions, and usage statistics.

    Args:
        action: "focus" to manage modes/groups, "suggest" to get mode suggestions, "stats" for usage analytics
        mode: Workflow mode to switch to (focus action)
        enable_group: Specific group to enable (focus action)
        disable_group: Specific group to disable (focus action)
        status: If True, return current status without changes (focus action)
        text: Optional text to analyze for mode suggestion (suggest action)
        auto_switch: If True, automatically switch to suggested mode (suggest action)

    Returns:
        JSON with workflow mode operation results
    """
    if action == "focus":
        from .dynamic_tools import focus_mode
        return focus_mode(mode, enable_group, disable_group, status)

    elif action == "suggest":
        from .dynamic_tools import suggest_mode
        return suggest_mode(text, auto_switch)

    elif action == "stats":
        from .dynamic_tools import get_tool_usage_stats
        return get_tool_usage_stats()

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown workflow_mode action: {action}. Use 'focus', 'suggest', or 'stats'.",
        }, indent=2)



def recommend(
    action: str = "model",
    # model action params
    task_description: Optional[str] = None,
    task_type: Optional[str] = None,
    optimize_for: str = "quality",
    include_alternatives: bool = True,
    # workflow action params
    task_id: Optional[str] = None,
    include_rationale: bool = True,
    # advisor action params
    metric: Optional[str] = None,
    tool: Optional[str] = None,
    stage: Optional[str] = None,
    score: float = 50.0,
    context: str = "",
    log: bool = True,
    session_mode: Optional[str] = None,
) -> str:
    """
    Unified recommendation tool.

    Consolidates model recommendations, workflow mode suggestions, and advisor consultations.

    Args:
        action: "model" for AI model recommendations, "workflow" for mode suggestions, "advisor" for wisdom
        task_description: Description of the task (model/workflow actions)
        task_type: Optional explicit task type (model action)
        optimize_for: "quality", "speed", or "cost" (model action)
        include_alternatives: Include alternative recommendations (model action)
        task_id: Optional Todo2 task ID to analyze (workflow action)
        include_rationale: Whether to include detailed reasoning (workflow action)
        metric: Scorecard metric to get advice for (advisor action)
        tool: Tool to get advice for (advisor action)
        stage: Workflow stage to get advice for (advisor action)
        score: Current score for wisdom tier selection (advisor action, 0-100)
        context: What you're working on (advisor action)
        log: Whether to log consultation (advisor action)
        session_mode: Inferred session mode for mode-aware guidance (advisor action)

    Returns:
        JSON with recommendation results (model/workflow) or dict (advisor)
    """
    if action == "model":
        from .model_recommender import recommend_model
        result = recommend_model(task_description, task_type, optimize_for, include_alternatives)
        # Ensure we always return a JSON string
        if isinstance(result, str):
            return result
        elif isinstance(result, dict):
            return json.dumps(result, indent=2)
        else:
            return json.dumps({"result": str(result)}, indent=2)

    elif action == "workflow":
        from .workflow_recommender import recommend_workflow_mode
        result = recommend_workflow_mode(task_description, task_id, include_rationale)
        # Ensure we always return a JSON string
        if isinstance(result, str):
            return result
        elif isinstance(result, dict):
            return json.dumps(result, indent=2)
        else:
            return json.dumps({"result": str(result)}, indent=2)

    elif action == "advisor":
        # Use devwisdom-go MCP server instead of direct import
        from ..utils.wisdom_client import consult_advisor
        # Get session mode if not provided
        if session_mode is None:
            try:
                import json as json_lib

                from ..resources.session import get_session_mode_resource
                mode_resource_json = get_session_mode_resource()
                mode_data = json_lib.loads(mode_resource_json)
                session_mode = mode_data.get("mode") or mode_data.get("inferred_mode")
            except Exception:
                pass  # Fallback gracefully if mode inference unavailable

        result = consult_advisor(
            metric=metric,
            tool=tool,
            stage=stage,
            score=score,
            context=context
        )
        # Convert dict to JSON string if needed
        if isinstance(result, dict):
            return json.dumps(result, indent=2)
        return result if result else json.dumps({"error": "Failed to consult advisor"}, indent=2)

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown recommend action: {action}. Use 'model', 'workflow', or 'advisor'.",
        }, indent=2)



def tool_catalog(
    action: str = "list",
    # list action params
    category: Optional[str] = None,
    persona: Optional[str] = None,
    include_examples: bool = True,
    # help action params
    tool_name: Optional[str] = None,
) -> str:
    """
    Unified tool catalog tool.

    Consolidates tool catalog browsing and help operations.

    Args:
        action: "list" for tool catalog, "help" for specific tool documentation
        category: Filter by category (list action)
        persona: Filter by persona (list action)
        include_examples: Include example prompts (list action)
        tool_name: Name of tool to get help for (help action)

    Returns:
        JSON with tool catalog results
    """
    if action == "list":
        from .hint_catalog import list_tools
        return list_tools(category, persona, include_examples)

    elif action == "help":
        if not tool_name:
            return json.dumps({
                "status": "error",
                "error": "tool_name parameter required for help action",
            }, indent=2)
        from .hint_catalog import get_tool_help
        return get_tool_help(tool_name)

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown tool_catalog action: {action}. Use 'list' or 'help'.",
        }, indent=2)


