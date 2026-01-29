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
from typing import Any, Optional

logger = logging.getLogger(__name__)


"""
Consolidated Quality

Consolidated tools: testing_async, testing, lint, health
"""

async def testing_async(
    action: str = "run",
    # run params
    test_path: Optional[str] = None,
    test_framework: str = "auto",
    verbose: bool = True,
    coverage: bool = False,
    # coverage params
    coverage_file: Optional[str] = None,
    min_coverage: int = 80,
    format: str = "html",
    # suggest params
    target_file: Optional[str] = None,
    min_confidence: float = 0.7,
    # generate params
    use_mlx: bool = True,
    use_coreml: bool = True,
    coreml_model_path: Optional[str] = None,
    compute_units: str = "all",
    # validate params
    framework: Optional[str] = None,
    # common
    output_path: Optional[str] = None,
    ctx: Optional[Any] = None,
) -> str:
    """
    Unified testing tool (async with progress).

    Args:
        action: "run" to execute tests, "coverage" to analyze coverage, "suggest" to suggest test cases (templates), "generate" to generate actual test code (AI-powered), "validate" to validate test structure
        test_path: Path to test file/directory (run action)
        test_framework: pytest, unittest, ctest, or auto (run action)
        verbose: Show detailed output (run action)
        coverage: Generate coverage during test run (run action)
        coverage_file: Path to coverage file (coverage action)
        min_coverage: Minimum coverage threshold (coverage action)
        format: Report format - html, json, terminal (coverage action)
        target_file: File to analyze for suggestions (suggest action)
        min_confidence: Minimum confidence threshold for suggestions (suggest action, default: 0.7)
        use_mlx: Use MLX for test generation (generate action, default: True)
        use_coreml: Use Core ML for test generation (generate action, default: True)
        coreml_model_path: Path to Core ML model file (generate action, optional)
        compute_units: Core ML compute units - all, cpu_and_ane, cpu_and_gpu, cpu_only (generate action, default: all)
        framework: Expected framework for validation (validate action, default: auto)
        output_path: Save results to file
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with test, coverage, suggestion, or validation results
    """
    if action == "run":
        from .run_tests import run_tests_async
        result = await run_tests_async(test_path, test_framework, verbose, coverage, output_path, ctx)
        # Result should already be a string, but ensure it is
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    elif action == "coverage":
        from .test_coverage import analyze_test_coverage
        result = analyze_test_coverage(coverage_file, min_coverage, output_path, format)
        # Result should already be a string, but ensure it is
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    elif action == "suggest":
        from .test_suggestions import suggest_test_cases
        result = suggest_test_cases(target_file, test_framework, min_confidence, output_path)
        # Result should already be a string, but ensure it is
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    elif action == "generate":
        # Generate actual test code using AI (Core ML NPU or MLX GPU)
        from .test_suggestions import generate_test_code
        result = generate_test_code(
            target_file=target_file,
            test_framework=test_framework,
            use_coreml=use_coreml,  # Try Core ML first (NPU)
            coreml_model_path=coreml_model_path,  # Path to Core ML model
            use_mlx=use_mlx,  # Fallback to MLX (GPU)
            compute_units=compute_units,  # Core ML compute units
            output_path=output_path,
        )
        # Result should already be a string, but ensure it is
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    elif action == "validate":
        from .test_validation import validate_test_structure
        result = validate_test_structure(test_path, framework, output_path)
        # Result should already be a string, but ensure it is
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown testing action: {action}. Use 'run', 'coverage', 'suggest', 'generate', or 'validate'.",
        }, indent=2)



def testing(
    action: str = "run",
    # run params
    test_path: Optional[str] = None,
    test_framework: str = "auto",
    verbose: bool = True,
    coverage: bool = False,
    # coverage params
    coverage_file: Optional[str] = None,
    min_coverage: int = 80,
    format: str = "html",
    # suggest params
    target_file: Optional[str] = None,
    min_confidence: float = 0.7,
    # generate params
    use_mlx: bool = True,
    use_coreml: bool = True,
    coreml_model_path: Optional[str] = None,
    compute_units: str = "all",
    # validate params
    framework: Optional[str] = None,
    # common
    output_path: Optional[str] = None,
    ctx: Optional[Any] = None,
) -> str:
    """
    Unified testing tool (sync wrapper).

    Args:
        action: "run" to execute tests, "coverage" to analyze coverage, "suggest" to suggest test cases (templates), "generate" to generate actual test code (AI-powered), "validate" to validate test structure
        test_path: Path to test file/directory (run/validate action)
        test_framework: pytest, unittest, ctest, or auto (run/suggest action)
        verbose: Show detailed output (run action)
        coverage: Generate coverage during test run (run action)
        coverage_file: Path to coverage file (coverage action)
        min_coverage: Minimum coverage threshold (coverage action)
        format: Report format - html, json, terminal (coverage action)
        target_file: File to analyze for suggestions (suggest action)
        min_confidence: Minimum confidence threshold for suggestions (suggest action, default: 0.7)
        use_mlx: Use MLX for test generation (generate action, default: True)
        use_coreml: Use Core ML for test generation (generate action, default: True)
        coreml_model_path: Path to Core ML model file (generate action, optional)
        compute_units: Core ML compute units - all, cpu_and_ane, cpu_and_gpu, cpu_only (generate action, default: all)
        framework: Expected framework for validation (validate action, default: auto)
        output_path: Save results to file
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with test, coverage, suggestion, or validation results
    """
    # Check if we're in an async context using try/except/else pattern
    # This ensures the intentional RuntimeError is raised in else block, not caught
    try:
        asyncio.get_running_loop()
        # If we get here, we're in an async context - raise helpful error
        raise RuntimeError(
            "testing() cannot be called from an async context. "
            "Use testing_async() instead and await it."
        )
    except RuntimeError as e:
        # Re-raise if it's our intentional error
        if "testing_async()" in str(e) or "async context" in str(e).lower():
            raise
        # Otherwise, no running loop - safe to use asyncio.run()
        result = asyncio.run(testing_async(action, test_path, test_framework, verbose, coverage, coverage_file, min_coverage, format, target_file, min_confidence, use_mlx, use_coreml, coreml_model_path, compute_units, framework, output_path, ctx))
    # Convert dict to JSON string
    return json.dumps(result, indent=2) if isinstance(result, dict) else result



def lint(
    action: str = "run",
    # run params
    path: Optional[str] = None,
    linter: str = "ruff",
    fix: bool = False,
    analyze: bool = True,
    select: Optional[str] = None,
    ignore: Optional[str] = None,
    # analyze params
    problems_json: Optional[str] = None,
    include_hints: bool = True,
    # common
    output_path: Optional[str] = None,
) -> str:
    """
    Unified linting tool.

    Args:
        action: "run" to execute linter, "analyze" to analyze problems
        path: File or directory to lint (run action)
        linter: "ruff" or "flake8" (run action)
        fix: Auto-fix issues (run action)
        analyze: Run analyze_problems on results (run action)
        select: Rule codes to enable (run action)
        ignore: Rule codes to ignore (run action)
        problems_json: JSON string of problems to analyze (analyze action)
        include_hints: Include resolution hints (analyze action)
        output_path: Save results to file

    Returns:
        JSON string with linting or analysis results
    """
    if action == "run":
        from .linter import run_linter
        result = run_linter(path, linter, fix, analyze, select, ignore)
        # Result might already be a string, or might be a dict
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    elif action == "analyze":
        if not problems_json:
            return json.dumps({
                "status": "error",
                "error": "problems_json is required for analyze action",
            }, indent=2)
        from .problems_advisor import analyze_problems_tool
        result = analyze_problems_tool(problems_json, include_hints, output_path)
        # Result should already be a string, but ensure it is
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown lint action: {action}. Use 'run' or 'analyze'.",
        }, indent=2)



def health(
    action: str = "server",
    # git params
    agent_name: Optional[str] = None,
    check_remote: bool = True,
    # docs params
    output_path: Optional[str] = None,
    create_tasks: bool = True,
    # dod params
    task_id: Optional[str] = None,
    changed_files: Optional[str] = None,
    auto_check: bool = True,
    # cicd params
    workflow_path: Optional[str] = None,
    check_runners: bool = True,
) -> str:
    """
    Unified health check tool.

    Args:
        action: "server" for server status, "git" for working copy, "docs" for documentation, "dod" for definition of done, "cicd" for CI/CD validation
        agent_name: Agent name filter (git action)
        check_remote: Check remote sync status (git action)
        output_path: Save results to file (docs, cicd actions)
        create_tasks: Create tasks for issues (docs action)
        task_id: Task to check completion for (dod action)
        changed_files: Files changed as JSON (dod action)
        auto_check: Auto-run checks (dod action)
        workflow_path: Path to workflow file (cicd action)
        check_runners: Validate runner configs (cicd action)

    Returns:
        JSON string with health check results
    """
    if action == "server":
        import time

        from ..utils import find_project_root

        project_root = find_project_root()
        version = "unknown"
        pyproject = project_root / "pyproject.toml"
        if pyproject.exists():
            import re
            content = pyproject.read_text()
            match = re.search(r'version\s*=\s*["\']([^"\']+)["\']', content)
            if match:
                version = match.group(1)

        result = {
            "status": "operational",
            "version": version,
            "project_root": str(project_root),
            "timestamp": time.time(),
        }
        return json.dumps(result, indent=2)
    elif action == "git":
        from .working_copy_health import check_working_copy_health
        result = check_working_copy_health(agent_name=agent_name, check_remote=check_remote)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "docs":
        from .docs_health import check_documentation_health
        result = check_documentation_health(output_path, create_tasks)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "dod":
        from .definition_of_done import check_definition_of_done
        result = check_definition_of_done(task_id, changed_files, auto_check, output_path)
        return json.dumps(result, indent=2) if isinstance(result, dict) else result
    elif action == "cicd":
        from .ci_cd_validation import validate_ci_cd_workflow
        result = validate_ci_cd_workflow(workflow_path, check_runners, output_path)
        # Result might already be a string, or might be a dict
        if isinstance(result, str):
            return result
        return json.dumps(result, indent=2)
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown health action: {action}. Use 'server', 'git', 'docs', 'dod', or 'cicd'.",
        }, indent=2)


