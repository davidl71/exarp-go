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

import json
import logging
from typing import List, Optional

logger = logging.getLogger(__name__)


"""
Consolidated Config

Consolidated tools: generate_config, setup_hooks, prompt_tracking
"""

def generate_config(
    action: str = "rules",
    # rules params
    rules: Optional[str] = None,
    overwrite: bool = False,
    analyze_only: bool = False,
    # ignore params
    include_indexing: bool = True,
    analyze_project: bool = True,
    # simplify params
    rule_files: Optional[str] = None,
    output_dir: Optional[str] = None,
    # common
    dry_run: bool = False,
) -> str:
    """
    Unified config generation tool.

    Args:
        action: "rules" for .mdc files, "ignore" for .cursorignore, "simplify" for rule simplification
        rules: Specific rules to generate (rules action)
        overwrite: Overwrite existing rules (rules action)
        analyze_only: Only analyze, don't generate (rules action)
        include_indexing: Include indexing ignore (ignore action)
        analyze_project: Analyze project structure (ignore action)
        rule_files: Rule files to simplify (simplify action)
        output_dir: Output directory (simplify action)
        dry_run: Preview changes without writing

    Returns:
        JSON string with config generation results
    """
    if action == "rules":
        from .cursor_rules_generator import generate_cursor_rules
        result = generate_cursor_rules(rules, overwrite, analyze_only)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "ignore":
        from .cursorignore_generator import generate_cursorignore
        result = generate_cursorignore(include_indexing, analyze_project, dry_run)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "simplify":
        from .simplify_rules import simplify_rules
        parsed_files = None
        if rule_files:
            try:
                parsed_files = json.loads(rule_files)
            except json.JSONDecodeError:
                return json.dumps({"status": "error", "error": "Invalid JSON in rule_files parameter"}, indent=2)
        result = simplify_rules(parsed_files, dry_run, output_dir)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown config action: {action}. Use 'rules', 'ignore', or 'simplify'.",
        }, indent=2)



def setup_hooks(
    action: str = "git",
    # git hooks params
    hooks: Optional[List[str]] = None,
    # pattern params
    patterns: Optional[str] = None,
    config_path: Optional[str] = None,
    # common
    install: bool = True,
    dry_run: bool = False,
) -> str:
    """
    Unified hooks setup tool.

    Args:
        action: "git" for git hooks, "patterns" for pattern triggers
        hooks: Specific git hooks to install (git action)
        patterns: Pattern trigger definitions as JSON (patterns action)
        config_path: Config file path (patterns action)
        install: Install hooks (vs uninstall)
        dry_run: Preview changes without writing

    Returns:
        JSON string with hook setup results
    """
    if action == "git":
        from .git_hooks import setup_git_hooks
        result = setup_git_hooks(hooks, install, dry_run)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "patterns":
        from .pattern_triggers import setup_pattern_triggers
        parsed_patterns = None
        if patterns:
            try:
                parsed_patterns = json.loads(patterns)
            except json.JSONDecodeError:
                return json.dumps({"status": "error", "error": "Invalid JSON in patterns parameter"}, indent=2)
        result = setup_pattern_triggers(parsed_patterns, config_path, install, dry_run)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown hooks action: {action}. Use 'git' or 'patterns'.",
        }, indent=2)



def prompt_tracking(
    action: str = "analyze",
    # log params
    prompt: Optional[str] = None,
    task_id: Optional[str] = None,
    mode: Optional[str] = None,
    outcome: Optional[str] = None,
    iteration: int = 1,
    # analyze params
    days: int = 7,
) -> str:
    """
    Unified prompt tracking tool.

    Args:
        action: "log" to record a prompt, "analyze" to view patterns
        prompt: Prompt text to log (log action, required)
        task_id: Associated task ID (log action)
        mode: Workflow mode used (log action)
        outcome: Result of prompt (log action)
        iteration: Iteration number (log action)
        days: Days of history to analyze (analyze action)

    Returns:
        JSON string with log confirmation or analysis results
    """
    if action == "log":
        if not prompt:
            return json.dumps({"status": "error", "error": "prompt parameter required for log action"}, indent=2)
        from .prompt_iteration_tracker import log_prompt_iteration
        result = log_prompt_iteration(prompt, task_id, mode, outcome, iteration)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "analyze":
        from .prompt_iteration_tracker import analyze_prompt_iterations
        result = analyze_prompt_iterations(days)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown prompt tracking action: {action}. Use 'log' or 'analyze'.",
        }, indent=2)


