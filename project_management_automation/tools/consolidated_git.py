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
Consolidated Git

Consolidated tools: git_tools, session
"""

def git_tools(
    action: str = "commits",
    task_id: Optional[str] = None,
    branch: Optional[str] = None,
    limit: int = 50,
    commit1: Optional[str] = None,
    commit2: Optional[str] = None,
    time1: Optional[str] = None,
    time2: Optional[str] = None,
    format: str = "text",
    output_path: Optional[str] = None,
    max_commits: int = 50,
    source_branch: Optional[str] = None,
    target_branch: Optional[str] = None,
    conflict_strategy: str = "newer",
    author: str = "system",
    dry_run: bool = False,
) -> str:
    """
    [HINT: Git tools. action=commits|branches|diff|graph|merge|set_branch. Unified git-inspired tools.]

    Unified git-inspired task management tools.

    Actions:
    - action="commits": Get commit history (task_id required for task commits, branch for branch commits)
    - action="branches": List all branches with statistics
    - action="diff": Compare task versions
    - action="graph": Generate commit graph visualization
    - action="merge": Merge tasks from one branch to another
    - action="set_branch": Assign task to a branch

    Returns JSON string (FastMCP requirement).
    """
    try:
        if action == "commits":
            from .git_inspired_tools import get_branch_commits, get_task_commits
            if task_id:
                return get_task_commits(task_id, branch, limit)
            elif branch:
                return get_branch_commits(branch, limit)
            else:
                return json.dumps({"error": "task_id or branch required for commits action"}, indent=2)

        elif action == "branches":
            from .git_inspired_tools import list_branches
            return list_branches()

        elif action == "tasks":
            from .git_inspired_tools import get_branch_tasks
            if not branch:
                return json.dumps({"error": "branch parameter required for tasks action"}, indent=2)
            return get_branch_tasks(branch)

        elif action == "diff":
            from .git_inspired_tools import compare_task_diff
            if not task_id:
                return json.dumps({"error": "task_id parameter required for diff action"}, indent=2)
            return compare_task_diff(task_id, commit1, commit2, time1, time2)

        elif action == "graph":
            from .git_inspired_tools import generate_graph
            return generate_graph(branch, task_id, format, output_path, max_commits)

        elif action == "merge":
            from .git_inspired_tools import merge_branch_tools
            if not source_branch or not target_branch:
                return json.dumps({"error": "source_branch and target_branch required for merge action"}, indent=2)
            return merge_branch_tools(source_branch, target_branch, conflict_strategy, author, dry_run)

        elif action == "set_branch":
            from .git_inspired_tools import set_task_branch_tool
            if not task_id or not branch:
                return json.dumps({"error": "task_id and branch required for set_branch action"}, indent=2)
            return set_task_branch_tool(task_id, branch)

        else:
            return json.dumps({
                "error": f"Unknown git_tools action: {action}. Use 'commits', 'branches', 'tasks', 'diff', 'graph', 'merge', or 'set_branch'."
            }, indent=2)

    except ImportError as e:
        return json.dumps({"error": f"Git tools not available: {e}"}, indent=2)
    except Exception as e:
        logger.error(f"Git tools error: {e}", exc_info=True)
        return json.dumps({"error": str(e)}, indent=2)



def session(
    action: str = "prime",
    include_hints: bool = True,
    include_tasks: bool = True,
    override_mode: Optional[str] = None,
    task_id: Optional[str] = None,
    summary: Optional[str] = None,
    blockers: Optional[str] = None,
    next_steps: Optional[str] = None,
    unassign_my_tasks: bool = True,
    include_git_status: bool = True,
    limit: int = 5,
    dry_run: bool = False,
    direction: str = "both",
    prefer_agentic_tools: bool = True,
    auto_commit: bool = True,
    mode: Optional[str] = None,
    category: Optional[str] = None,
    keywords: Optional[str] = None,
    assignee_name: Optional[str] = None,
    assignee_type: str = "agent",
    hostname: Optional[str] = None,
    status_filter: Optional[str] = None,
    priority_filter: Optional[str] = None,
    include_unassigned: bool = False,
    max_tasks_per_agent: int = 5,
) -> str:
    """
    [HINT: Session. action=prime|handoff|prompts|assignee. Unified session management tools.]

    Unified session management tool consolidating auto-primer, handoff, prompt discovery, and task assignee.

    Actions:
    - action="prime": Auto-prime AI context at session start
    - action="handoff": End/resume sessions for multi-device coordination
    - action="prompts": Find relevant prompts by mode/persona/category/keywords
    - action="assignee": Manage task assignments across agents/humans/hosts

    Returns JSON string (FastMCP requirement).
    """
    try:
        if action == "prime":
            from .auto_primer import auto_prime
            result = auto_prime(
                include_hints=include_hints,
                include_tasks=include_tasks,
                override_mode=override_mode,
            )
            # auto_prime returns JSON string - return it directly
            return result if isinstance(result, str) else json.dumps({"error": "Invalid return type from auto_prime"}, indent=2)

        elif action == "handoff":
            from .session_handoff import exarp_session_handoff
            result = exarp_session_handoff(
                action="end" if summary else "resume",
                summary=summary,
                blockers=json.loads(blockers) if blockers else None,
                next_steps=json.loads(next_steps) if next_steps else None,
                unassign_my_tasks=unassign_my_tasks,
                include_git_status=include_git_status,
                limit=limit,
                dry_run=dry_run,
                direction=direction,
                prefer_agentic_tools=prefer_agentic_tools,
                auto_commit=auto_commit,
            )
            return result if isinstance(result, str) else json.dumps({"error": "Invalid return type"}, indent=2)

        elif action == "prompts":
            from .resources.prompt_discovery import find_prompts
            result = find_prompts(
                mode=mode,
                persona=category,  # Using category for persona
                category=category,
                keywords=keywords,
            )
            return result if isinstance(result, str) else json.dumps({"error": "Invalid return type"}, indent=2)

        elif action == "assignee":
            from .task_assignee import task_assignee_tool
            # Map action to assignee tool actions
            if assignee_name:
                assignee_action = "assign"
            elif task_id:
                assignee_action = "unassign"
            else:
                assignee_action = "list"

            result = task_assignee_tool(
                action=assignee_action,
                task_id=task_id,
                assignee_name=assignee_name,
                assignee_type=assignee_type,
                hostname=hostname,
                status_filter=status_filter,
                priority_filter=priority_filter,
                include_unassigned=include_unassigned,
                max_tasks_per_agent=max_tasks_per_agent,
            )
            return result if isinstance(result, str) else json.dumps({"error": "Invalid return type"}, indent=2)

        else:
            return json.dumps({
                "error": f"Unknown session action: {action}. Use 'prime', 'handoff', 'prompts', or 'assignee'."
            }, indent=2)

    except ImportError as e:
        return json.dumps({"error": f"Session tools not available: {e}"}, indent=2)
    except Exception as e:
        logger.error(f"Session tool error: {e}", exc_info=True)
        return json.dumps({"error": str(e)}, indent=2)


