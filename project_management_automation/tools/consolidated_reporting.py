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
import subprocess
from typing import Any, List, Optional

logger = logging.getLogger(__name__)


"""
Consolidated Reporting

Consolidated tools: report, security_async, security
"""

def report(
    action: str = "overview",
    # common params
    output_format: str = "text",
    output_path: Optional[str] = None,
    # scorecard params
    include_recommendations: bool = True,
    # briefing params
    overall_score: float = 50.0,
    security_score: float = 50.0,
    testing_score: float = 50.0,
    documentation_score: float = 50.0,
    completion_score: float = 50.0,
    alignment_score: float = 50.0,
    # prd params
    project_name: Optional[str] = None,
    include_architecture: bool = True,
    include_metrics: bool = True,
    include_tasks: bool = True,
) -> str:
    """
    Unified report generation tool.

    Args:
        action: "overview" for project overview, "scorecard" for health metrics, "briefing" for advisor wisdom, "prd" for requirements doc
        output_format: Output format (text, json, markdown)
        output_path: Save results to file
        include_recommendations: Include recommendations (scorecard action)
        overall_score: Overall project score (briefing action)
        security_score: Security metric score (briefing action)
        testing_score: Testing metric score (briefing action)
        documentation_score: Documentation score (briefing action)
        completion_score: Completion score (briefing action)
        alignment_score: Alignment score (briefing action)
        project_name: Project name override (prd action)
        include_architecture: Include architecture section (prd action)
        include_metrics: Include metrics section (prd action)
        include_tasks: Include tasks section (prd action)

    Returns:
        Generated report
    """
    try:
        if action == "overview":
            from .project_overview import generate_project_overview
            result = generate_project_overview(output_format, output_path)
        elif action == "scorecard":
            from ..utils import find_project_root
            from .project_scorecard import (
                _get_exarp_go_binary,
                _is_go_project,
                generate_project_scorecard,
            )
            project_root = find_project_root()
            if _is_go_project(project_root):
                args_json = json.dumps({
                    "action": "scorecard",
                    "output_format": output_format,
                    "output_path": output_path,
                })
                try:
                    proc = subprocess.run(
                        [_get_exarp_go_binary(project_root), "-tool", "report", "-args", args_json],
                        cwd=str(project_root),
                        capture_output=True,
                        text=True,
                        timeout=120,
                    )
                    if proc.returncode == 0 and (proc.stdout or "").strip():
                        raw = proc.stdout.strip()
                        if "Result:\n" in raw:
                            raw = raw.split("Result:\n", 1)[1].strip()
                        result = json.loads(raw) if output_format == "json" and raw.startswith("{") else raw
                    else:
                        result = generate_project_scorecard(output_format, include_recommendations, output_path)
                except Exception:
                    result = generate_project_scorecard(output_format, include_recommendations, output_path)
            else:
                result = generate_project_scorecard(output_format, include_recommendations, output_path)
        elif action == "briefing":
            # Use devwisdom-go MCP server instead of direct import
            from ..utils.wisdom_client import get_daily_briefing
            metric_scores = {
                "security": security_score,
                "testing": testing_score,
                "documentation": documentation_score,
                "completion": completion_score,
                "alignment": alignment_score,
            }
            result = get_daily_briefing(overall_score, metric_scores)
            # Convert dict to JSON string if needed
            if isinstance(result, dict):
                result = json.dumps(result, indent=2)
        elif action == "prd":
            from .prd_generator import generate_prd
            result = generate_prd(
                project_name=project_name,
                output_path=output_path,
                include_tasks=include_tasks,
                include_architecture=include_architecture,
                include_metrics=include_metrics
            )
        else:
            result = {
                "status": "error",
                "error": f"Unknown report action: {action}. Use 'overview', 'scorecard', 'briefing', or 'prd'.",
            }

        # Ensure we always return a JSON string
        if isinstance(result, str):
            return result
        elif isinstance(result, dict):
            return json.dumps(result, indent=2)
        else:
            return json.dumps({"result": str(result)}, indent=2)
    except Exception as e:
        return json.dumps({"error": str(e)}, indent=2)


# advisor_audio tool removed - migrated to devwisdom-go MCP server
# Use devwisdom MCP server tools directly for voice synthesis and podcast generation



async def security_async(
    action: str = "report",
    repo: str = "davidl71/project-management-automation",
    languages: Optional[List[str]] = None,
    config_path: Optional[str] = None,
    state: str = "open",
    include_dismissed: bool = False,
    ctx: Optional[Any] = None,
) -> str:
    """
    Unified security analysis tool (async with progress).

    Args:
        action: "scan" for local pip-audit, "alerts" for Dependabot, "report" for combined
        repo: GitHub repo for alerts/report (owner/repo format)
        languages: Languages to scan (scan action)
        config_path: Config file path (scan action)
        state: Alert state filter (alerts action)
        include_dismissed: Include dismissed alerts (report action)
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with security scan/report results
    """
    if action == "scan":
        from .dependency_security import scan_dependency_security_async
        result = await scan_dependency_security_async(languages, config_path, ctx)
        # Result should already be a string, but ensure it is
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    elif action == "alerts":
        from .dependabot_integration import fetch_dependabot_alerts
        result = fetch_dependabot_alerts(repo, state)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "report":
        from .dependabot_integration import get_unified_security_report
        result = get_unified_security_report(repo, include_dismissed)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown security action: {action}. Use 'scan', 'alerts', or 'report'.",
        }, indent=2)



def security(
    action: str = "report",
    repo: str = "davidl71/project-management-automation",
    languages: Optional[List[str]] = None,
    config_path: Optional[str] = None,
    state: str = "open",
    include_dismissed: bool = False,
    ctx: Optional[Any] = None,
) -> str:
    """
    Unified security analysis tool (sync wrapper).

    Args:
        action: "scan" for local pip-audit, "alerts" for Dependabot, "report" for combined
        repo: GitHub repo for alerts/report (owner/repo format)
        languages: Languages to scan (scan action)
        config_path: Config file path (scan action)
        state: Alert state filter (alerts action)
        include_dismissed: Include dismissed alerts (report action)
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with security scan/report results
    """
    # Check if we're in an async context using try/except/else pattern
    # This ensures the intentional RuntimeError is raised in else block, not caught
    try:
        asyncio.get_running_loop()
        # If we get here, we're in an async context - raise helpful error
        raise RuntimeError(
            "security() cannot be called from an async context. "
            "Use security_async() instead and await it."
        )
    except RuntimeError as e:
        # Re-raise if it's our intentional error
        if "security_async()" in str(e) or "async context" in str(e).lower():
            raise
        # Otherwise, no running loop - safe to use asyncio.run()
        result = asyncio.run(security_async(action, repo, languages, config_path, state, include_dismissed, ctx))
    # Convert dict to JSON string
    return json.dumps(result, indent=2) if isinstance(result, dict) else result


