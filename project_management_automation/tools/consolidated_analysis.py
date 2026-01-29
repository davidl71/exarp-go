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
from typing import Optional

logger = logging.getLogger(__name__)


"""
Consolidated Analysis

Consolidated tools: analyze_alignment, task_analysis, task_discovery
"""

def analyze_alignment(
    action: str = "todo2",
    create_followup_tasks: bool = True,
    output_path: Optional[str] = None,
) -> str:
    """
    Unified alignment analysis tool.

    Args:
        action: "todo2" for task alignment, "prd" for PRD persona mapping
        create_followup_tasks: Create tasks for misaligned items (todo2 only)
        output_path: Optional file to save results

    Returns:
        JSON string with alignment analysis results (FastMCP requires strings)
    """
    if action == "todo2":
        from .todo2_alignment import analyze_todo2_alignment
        # analyze_todo2_alignment already returns a JSON string
        result = analyze_todo2_alignment(create_followup_tasks, output_path)
        # Ensure it's a string (it should already be)
        return result if isinstance(result, str) else json.dumps(result, separators=(",", ":"))
    elif action == "prd":
        from .prd_alignment import analyze_prd_alignment
        result = analyze_prd_alignment(output_path)
        # Ensure we return a JSON string
        if isinstance(result, str):
            return result
        else:
            return json.dumps(result, separators=(",", ":"))
    else:
        error_result = {
            "status": "error",
            "error": f"Unknown alignment action: {action}. Use 'todo2' or 'prd'.",
        }
        return json.dumps(error_result, separators=(",", ":"))



def task_analysis(
    action: str = "duplicates",
    # duplicates params
    similarity_threshold: float = 0.85,
    auto_fix: bool = False,
    # tags params
    dry_run: bool = True,
    custom_rules: Optional[str] = None,
    remove_tags: Optional[str] = None,
    # hierarchy params
    output_format: str = "text",
    include_recommendations: bool = True,
    # dependencies params (uses output_format)
    # parallelization params (uses output_format)
    # common
    output_path: Optional[str] = None,
) -> str:
    """
    Unified task analysis tool.

    Args:
        action: "duplicates" to find duplicates, "tags" for tag cleanup, "hierarchy" for structure analysis,
                "dependencies" for dependency chain analysis, "parallelization" for parallel execution optimization
        similarity_threshold: Threshold for duplicate detection (duplicates action)
        auto_fix: Auto-merge duplicates (duplicates action)
        dry_run: Preview changes without applying (tags action)
        custom_rules: Custom tag rename rules as JSON (tags action)
        remove_tags: Tags to remove as JSON list (tags action)
        output_format: Output format - text, json, or markdown (hierarchy/dependencies/parallelization actions)
        include_recommendations: Include recommendations (hierarchy action)
        output_path: Save results to file

    Returns:
        JSON string with analysis results
    """
    if action == "duplicates":
        from .duplicate_detection import detect_duplicate_tasks
        result = detect_duplicate_tasks(similarity_threshold, auto_fix, output_path)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "tags":
        from .tag_consolidation import consolidate_tags
        result = consolidate_tags(dry_run, custom_rules, remove_tags, output_path)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "hierarchy":
        from .task_hierarchy_analyzer import analyze_task_hierarchy
        result = analyze_task_hierarchy(output_format, output_path, include_recommendations)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "dependencies":
        from .analyze_todo2_dependencies import analyze_todo2_dependencies
        result = analyze_todo2_dependencies(output_format, output_path)
        return result if isinstance(result, str) else json.dumps(result, indent=2)
    elif action == "parallelization":
        from .optimize_todo2_parallelization import optimize_todo2_parallelization
        # Get duration_weight parameter (default: 0.3 for lower weight)
        duration_weight = 0.3  # Lower default weight means duration matters less
        # Note: This can be made configurable via MCP tool parameters in the future
        result = optimize_todo2_parallelization(output_format, output_path, duration_weight)
        return result if isinstance(result, str) else json.dumps(result, indent=2)
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown task_analysis action: {action}. Use 'duplicates', 'tags', 'hierarchy', 'dependencies', or 'parallelization'.",
        }, indent=2)



def task_discovery(
    action: str = "all",
    # comments params
    file_patterns: Optional[str] = None,  # JSON list of glob patterns
    include_fixme: bool = True,
    # markdown params
    doc_path: Optional[str] = None,
    # orphans params (uses task_analysis internally)
    # common
    output_path: Optional[str] = None,
    create_tasks: bool = False,
) -> str:
    """
    Unified task discovery tool.

    Finds tasks from various sources in the codebase.

    Args:
        action: "comments" for TODO/FIXME in code, "markdown" for task lists in docs,
                "orphans" for orphaned Todo2 tasks, "all" for everything
        file_patterns: JSON list of glob patterns for code scanning (comments action)
        include_fixme: Include FIXME comments (comments action)
        doc_path: Path to scan for markdown tasks (markdown action)
        output_path: Save results to file
        create_tasks: Auto-create Todo2 tasks from discoveries

    Returns:
        JSON string with discovery results and found tasks
    """
    import re
    from pathlib import Path

    from ..utils import find_project_root

    project_root = find_project_root()
    results = {
        "action": action,
        "discoveries": [],
        "summary": {},
    }

    def scan_comments():
        """Scan code for TODO/FIXME comments."""
        discoveries = []
        patterns = ["**/*.py", "**/*.js", "**/*.ts", "**/*.tsx", "**/*.jsx"]
        if file_patterns:
            try:
                patterns = json.loads(file_patterns)
            except json.JSONDecodeError:
                pass

        todo_pattern = re.compile(r'#\s*(TODO|FIXME)[\s:]+(.+)', re.IGNORECASE) if include_fixme else re.compile(r'#\s*TODO[\s:]+(.+)', re.IGNORECASE)

        for pattern in patterns:
            for file_path in project_root.glob(pattern):
                if any(skip in str(file_path) for skip in ['.git', 'node_modules', '__pycache__', '.venv', 'build']):
                    continue
                try:
                    content = file_path.read_text(encoding='utf-8', errors='ignore')
                    for line_num, line in enumerate(content.split('\n'), 1):
                        match = todo_pattern.search(line)
                        if match:
                            discoveries.append({
                                "type": match.group(1).upper() if include_fixme else "TODO",
                                "text": match.group(2).strip() if include_fixme else match.group(1).strip(),
                                "file": str(file_path.relative_to(project_root)),
                                "line": line_num,
                                "source": "comment",
                            })
                except Exception:
                    continue
        return discoveries

    def scan_markdown():
        """Scan markdown files for task lists."""
        discoveries = []
        search_path = Path(doc_path) if doc_path else project_root / "docs"
        if not search_path.exists():
            search_path = project_root

        task_pattern = re.compile(r'^[\s]*[-*]\s*\[([ xX])\]\s*(.+)', re.MULTILINE)

        for md_file in search_path.rglob("*.md"):
            if any(skip in str(md_file) for skip in ['.git', 'node_modules']):
                continue
            try:
                content = md_file.read_text(encoding='utf-8', errors='ignore')
                for match in task_pattern.finditer(content):
                    is_done = match.group(1).lower() == 'x'
                    if not is_done:  # Only uncompleted tasks
                        discoveries.append({
                            "type": "MARKDOWN_TASK",
                            "text": match.group(2).strip(),
                            "file": str(md_file.relative_to(project_root)),
                            "completed": is_done,
                            "source": "markdown",
                        })
            except Exception:
                continue
        return discoveries

    def find_orphans():
        """Find orphaned Todo2 tasks."""
        from .task_hierarchy_analyzer import analyze_task_hierarchy
        result = analyze_task_hierarchy(output_format="json", include_recommendations=False)
        orphans = result.get("networkx_analysis", {}).get("orphans", [])
        return [{"type": "ORPHAN", "task_id": o, "source": "todo2"} for o in orphans]

    # Execute based on source
    if action in ["comments", "all"]:
        results["discoveries"].extend(scan_comments())
    if action in ["markdown", "all"]:
        results["discoveries"].extend(scan_markdown())
    if action in ["orphans", "all"]:
        results["discoveries"].extend(find_orphans())

    # Summary
    results["summary"] = {
        "total": len(results["discoveries"]),
        "by_source": {},
        "by_type": {},
    }
    for d in results["discoveries"]:
        src = d.get("source", "unknown")
        typ = d.get("type", "unknown")
        results["summary"]["by_source"][src] = results["summary"]["by_source"].get(src, 0) + 1
        results["summary"]["by_type"][typ] = results["summary"]["by_type"].get(typ, 0) + 1

    # Save if requested
    if output_path:
        Path(output_path).write_text(json.dumps(results, indent=2))

    # Always return JSON string
    return json.dumps(results, indent=2)


