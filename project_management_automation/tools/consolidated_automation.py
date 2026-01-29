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
Consolidated Automation

Consolidated tools: automation, estimation, task_workflow
"""

def automation(
    action: str = "daily",
    # daily params
    tasks: Optional[List[str]] = None,
    include_slow: bool = False,
    # nightly params
    max_tasks_per_host: int = 5,
    max_parallel_tasks: int = 10,
    priority_filter: Optional[str] = None,
    tag_filter: Optional[List[str]] = None,
    # sprint params
    max_iterations: int = 10,
    auto_approve: bool = True,
    extract_subtasks: bool = True,
    run_analysis_tools: bool = True,
    run_testing_tools: bool = True,
    # discover params
    min_value_score: float = 0.7,
    # common params
    dry_run: bool = False,
    output_path: Optional[str] = None,
) -> str:
    import sys
    print("DEBUG [consolidated.py automation] ENTRY", file=sys.stderr, flush=True)
    print(f"DEBUG [consolidated.py automation] action={action}, tasks={tasks}", file=sys.stderr, flush=True)
    """
    Unified automation tool.

    Args:
        action: "daily" for daily maintenance, "nightly" for task processing,
                "sprint" for sprint automation, "discover" for opportunity discovery
        tasks: List of task IDs to run (daily action)
        include_slow: Include slow tasks (daily action)
        max_tasks_per_host: Max tasks per host (nightly action)
        max_parallel_tasks: Max parallel tasks (nightly action)
        priority_filter: Filter by priority (nightly action)
        tag_filter: Filter by tags (nightly action)
        max_iterations: Max sprint iterations (sprint action)
        auto_approve: Auto-approve tasks (sprint action)
        extract_subtasks: Extract subtasks (sprint action)
        run_analysis_tools: Run analysis tools (sprint action)
        run_testing_tools: Run testing tools (sprint action)
        min_value_score: Min value score threshold (discover action)
        dry_run: Preview without applying
        output_path: Save results to file

    Returns:
        JSON string with automation results
    """
    if action == "daily":
        # Python daily orchestrator removed 2026-01-29; use exarp-go -tool automation (Go) only.
        return json.dumps({
            "status": "error",
            "error": "Daily automation is native Go only. Use: exarp-go -tool automation -args '{\"action\":\"daily\"}'.",
            "native_only": True,
        }, indent=2)

    elif action == "nightly":
        from .nightly_task_automation import run_nightly_task_automation
        result = run_nightly_task_automation(
            max_tasks_per_host=max_tasks_per_host,
            max_parallel_tasks=max_parallel_tasks,
            priority_filter=priority_filter,
            tag_filter=tag_filter,
            dry_run=dry_run,
        )
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "sprint":
        from .sprint_automation import sprint_automation
        result = sprint_automation(
            max_iterations=max_iterations,
            auto_approve=auto_approve,
            extract_subtasks=extract_subtasks,
            run_analysis_tools=run_analysis_tools,
            run_testing_tools=run_testing_tools,
            priority_filter=priority_filter,
            tag_filter=tag_filter,
            dry_run=dry_run,
            output_path=output_path,
        )
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "discover":
        from .automation_opportunities import find_automation_opportunities
        result = find_automation_opportunities(min_value_score, output_path)
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown automation action: {action}. Use 'daily', 'nightly', 'sprint', or 'discover'.",
        }, indent=2)



def estimation(
    action: str = "estimate",
    # estimate params
    name: Optional[str] = None,
    details: str = "",
    tags: Optional[str] = None,
    # BREAKING TEST: Add List[str] parameter like automation has
    tag_list: Optional[List[str]] = None,
    priority: str = "medium",
    use_historical: bool = True,
    detailed: bool = False,
    use_mlx: bool = True,
    mlx_weight: float = 0.3,
    # Core ML params
    use_coreml: bool = False,  # New: Enable Core ML backend
    coreml_model_path: Optional[str] = None,  # New: Path to Core ML model
    coreml_weight: float = 0.3,  # New: Core ML weight in hybrid
    compute_units: str = "all",  # New: Core ML compute units (all, cpu_and_gpu, cpu_and_ane, cpu_only)
    # analyze params (no additional params needed)
    # stats params (no additional params needed)
    # batch params
    output_path: Optional[str] = None,
) -> str:
    """
    Unified task duration estimation tool.

    Args:
        action: "estimate" for duration estimate, "analyze" for accuracy analysis,
                "stats" for statistical summary
        name: Task name (estimate action)
        details: Task details (estimate action)
        tags: Comma-separated tags (estimate action)
        priority: Task priority (estimate action)
        use_historical: Use historical data (estimate action)
        detailed: Return detailed breakdown (estimate action)
        use_mlx: Use MLX enhancement (estimate action)
        mlx_weight: MLX weight in hybrid estimate (estimate action)
        use_coreml: Use Core ML with Neural Engine (estimate action)
        coreml_model_path: Path to Core ML model (.mlpackage or .mlmodel)
        coreml_weight: Core ML weight in hybrid estimate (estimate action)
        compute_units: Core ML compute units (all, cpu_and_gpu, cpu_and_ane, cpu_only)

    Returns:
        JSON string with estimation results
    """
    if action == "estimate":
        if not name:
            return json.dumps({"status": "error", "error": "name parameter required for estimate action"}, indent=2)

        tag_list = [t.strip() for t in tags.split(",")] if tags else []

        # Try Core ML-enhanced estimator first (if enabled and model available)
        if use_coreml and coreml_model_path:
            try:
                from .coreml_task_estimator import (
                    estimate_task_duration_coreml_enhanced as _estimate_coreml_simple,
                )
                from .coreml_task_estimator import (
                    estimate_task_duration_coreml_enhanced_detailed,
                )

                if detailed:
                    result = estimate_task_duration_coreml_enhanced_detailed(
                        name=name,
                        details=details,
                        tags=tag_list,
                        priority=priority,
                        use_historical=use_historical,
                        use_coreml=True,
                        coreml_weight=coreml_weight,
                        coreml_model_path=coreml_model_path,
                        compute_units=compute_units,
                    )
                    return json.dumps(result, indent=2) if isinstance(result, dict) else (result if isinstance(result, str) else json.dumps({"error": "Invalid return type"}, indent=2))
                else:
                    hours = _estimate_coreml_simple(
                        name=name,
                        details=details,
                        tags=tag_list,
                        priority=priority,
                        use_historical=use_historical,
                        use_coreml=True,
                        coreml_weight=coreml_weight,
                        coreml_model_path=coreml_model_path,
                        compute_units=compute_units,
                    )
                    return json.dumps({
                        "estimate_hours": hours,
                        "name": name,
                        "priority": priority,
                        "method": "coreml_neural_engine",
                    }, indent=2)
            except ImportError:
                # Core ML not available, fall through to MLX or statistical
                logger.debug("Core ML not available, falling back to MLX or statistical")
            except Exception as e:
                logger.debug(f"Core ML estimation failed: {e}, falling back to MLX or statistical")

        # Try MLX-enhanced estimator (if enabled and Core ML not used)
        if use_mlx:
            try:
                from .mlx_task_estimator import (
                    estimate_task_duration_mlx_enhanced as _estimate_mlx_simple,
                )
                from .mlx_task_estimator import (
                    estimate_task_duration_mlx_enhanced_detailed,
                )

                if detailed:
                    result = estimate_task_duration_mlx_enhanced_detailed(
                        name=name,
                        details=details,
                        tags=tag_list,
                        priority=priority,
                        use_historical=use_historical,
                        use_mlx=True,
                        mlx_weight=mlx_weight,
                    )
                    return json.dumps(result, indent=2) if isinstance(result, dict) else (result if isinstance(result, str) else json.dumps({"error": "Invalid return type"}, indent=2))
                else:
                    hours = _estimate_mlx_simple(
                        name=name,
                        details=details,
                        tags=tag_list,
                        priority=priority,
                        use_historical=use_historical,
                        use_mlx=True,
                        mlx_weight=mlx_weight,
                    )
                    return json.dumps({
                        "estimate_hours": hours,
                        "name": name,
                        "priority": priority,
                        "method": "mlx_enhanced",
                    }, indent=2)
            except ImportError:
                # MLX not available, fall through to statistical-only
                pass

        # Fallback to statistical-only estimator
        from .task_duration_estimator import (
            estimate_task_duration as _estimate_simple,
        )
        from .task_duration_estimator import (
            estimate_task_duration_detailed,
        )

        if detailed:
            result = estimate_task_duration_detailed(
                name=name,
                details=details,
                tags=tag_list,
                priority=priority,
                use_historical=use_historical,
            )
            return json.dumps(result, indent=2) if isinstance(result, dict) else (result if isinstance(result, str) else json.dumps({"error": "Invalid return type"}, indent=2))
        else:
            hours = _estimate_simple(
                name=name,
                details=details,
                tags=tag_list,
                priority=priority,
                use_historical=use_historical,
            )
            return json.dumps({
                "estimate_hours": hours,
                "name": name,
                "priority": priority,
                "method": "statistical",
            }, indent=2)

    elif action == "analyze":
        from .estimation_learner import EstimationLearner
        from ..utils import find_project_root
        project_root = find_project_root()
        learner = EstimationLearner(project_root)
        result = learner.analyze_estimation_accuracy()
        
        # Auto-save learning data if analysis was successful
        if result.get('success') and result.get('completed_tasks_count', 0) > 0:
            learning_data = {
                'adjustment_factors': learner.get_adjustment_factors(),
                'learning_records': [
                    {
                        'task_id': t.get('id', ''),
                        'estimated_hours': t.get('estimated_hours'),
                        'actual_hours': t.get('actual_hours'),
                        'ratio': t.get('actual_hours', 0) / t.get('estimated_hours', 1) if t.get('estimated_hours') else 0,
                        'error_pct': t.get('error_pct', 0)
                    }
                    for t in result.get('completed_tasks', [])
                    if t.get('estimated_hours') and t.get('actual_hours')
                ],
                'last_updated': result.get('generated', ''),
                'metrics': result.get('accuracy_metrics', {})
            }
            learner.save_learning_data(learning_data)
        
        return json.dumps(result, indent=2) if isinstance(result, dict) else result

    elif action == "stats":
        from .task_duration_estimator import TaskDurationEstimator
        estimator = TaskDurationEstimator()
        stats = estimator.get_statistics()
        # get_statistics now returns JSON string directly
        return stats if isinstance(stats, str) else json.dumps(stats, indent=2)

    elif action == "batch":
        # Batch estimation: process multiple tasks with a single MLX model instance
        # tasks_json should be a JSON array of task objects with: name, details, tags, priority
        from pathlib import Path
        import json as json_lib
        from datetime import datetime
        
        # If tasks_json not provided, load from Todo2
        if not name:  # Using 'name' as tasks_json parameter (hack for now)
            # Load from Todo2
            from ..utils import find_project_root
            from ..utils.todo2_utils import filter_tasks_by_project, get_repo_project_id
            
            project_root = find_project_root()
            todo2_file = project_root / '.todo2' / 'state.todo2.json'
            
            if not todo2_file.exists():
                return json.dumps({
                    "status": "error",
                    "error": "Todo2 state file not found. Provide tasks_json parameter.",
                }, indent=2)
            
            with open(todo2_file) as f:
                data = json_lib.load(f)
            
            todos = data.get('todos', [])
            project_id = get_repo_project_id(project_root)
            project_tasks = filter_tasks_by_project(todos, project_id)
            
            # Filter to non-completed tasks
            tasks_to_estimate = [
                t for t in project_tasks
                if t.get('status', '').lower() not in ['done', 'completed', 'cancelled']
            ]
        else:
            # Parse tasks_json from name parameter
            try:
                tasks_to_estimate = json_lib.loads(name)
            except json_lib.JSONDecodeError:
                return json.dumps({
                    "status": "error",
                    "error": "Invalid tasks_json. Must be JSON array of task objects.",
                }, indent=2)
        
        # Create estimator instance (model loads once)
        # Prefer Core ML if enabled and model available, otherwise use MLX
        estimator = None
        use_coreml_batch = use_coreml and coreml_model_path
        if use_coreml_batch:
            try:
                from .coreml_task_estimator import CoreMLTaskEstimator
                estimator = CoreMLTaskEstimator(
                    use_coreml=True,
                    coreml_weight=coreml_weight,
                    coreml_model_path=coreml_model_path,
                    compute_units=compute_units,
                )
                logger.debug("Using Core ML estimator for batch processing")
            except ImportError:
                logger.debug("Core ML not available for batch, falling back to MLX")
                use_coreml_batch = False
        
        if not estimator:
            # Use MLX estimator (default or fallback)
            from .mlx_task_estimator import MLXEnhancedTaskEstimator
            estimator = MLXEnhancedTaskEstimator(
                use_mlx=use_mlx,
                mlx_weight=mlx_weight
            )
        
        # Also create statistical estimator for comparison if MLX/Core ML is enabled
        statistical_estimator = None
        if use_mlx or use_coreml_batch:
            from .task_duration_estimator import TaskDurationEstimator
            statistical_estimator = TaskDurationEstimator()
        
        results = []
        errors = []
        total_tasks = len(tasks_to_estimate)
        
        # Progress reporting
        import sys
        def report_progress(current: int, total: int, task_name: str, status: str = "processing"):
            """Report progress to stderr (won't interfere with JSON output)."""
            progress_pct = int((current / total) * 100)
            bar_length = 30
            filled = int(bar_length * current / total)
            bar = '█' * filled + '░' * (bar_length - filled)
            short_name = task_name[:40] + '...' if len(task_name) > 40 else task_name
            sys.stderr.write(f"\r[{bar}] {progress_pct:3d}% ({current}/{total}) {status}: {short_name}")
            sys.stderr.flush()
            if current == total:
                sys.stderr.write("\n")  # New line when complete
                sys.stderr.flush()
        
        for i, task in enumerate(tasks_to_estimate, 1):
            task_name = task.get('name') or task.get('id', 'Unnamed')
            task_details = task.get('long_description') or task.get('details') or task.get('content', '')
            task_tags = task.get('tags', [])
            if isinstance(task_tags, str):
                task_tags = [t.strip() for t in task_tags.split(',')]
            task_priority = task.get('priority', priority)
            
            # Report progress
            report_progress(i, total_tasks, task_name, "estimating")
            
            try:
                # Get MLX-enhanced estimate
                result = estimator.estimate(
                    name=task_name,
                    details=task_details[:500] if task_details else '',
                    tags=task_tags,
                    priority=task_priority,
                    use_historical=use_historical
                )
                
                # Get statistical estimate for comparison if MLX is enabled
                statistical_estimate = None
                if statistical_estimator:
                    try:
                        stat_result = statistical_estimator.estimate(
                            task_name,
                            task_details[:500] if task_details else '',
                            task_tags,
                            task_priority,
                            use_historical
                        )
                        statistical_estimate = {
                            'estimate_hours': stat_result.get('estimate_hours'),
                            'confidence': stat_result.get('confidence'),
                            'method': stat_result.get('method', 'statistical'),
                        }
                    except Exception as e:
                        logger.debug(f"Statistical estimate failed for {task_name}: {e}")
                
                results.append({
                    'task_id': task.get('id'),
                    'name': task_name,
                    'status': task.get('status'),
                    'priority': task_priority,
                    'estimate_hours': result.get('estimate_hours'),
                    'confidence': result.get('confidence'),
                    'method': result.get('method'),
                    'lower_bound': result.get('lower_bound'),
                    'upper_bound': result.get('upper_bound'),
                    'statistical_estimate': statistical_estimate,  # For comparison
                    'mlx_improvement': round(
                        ((result.get('estimate_hours', 0) - (statistical_estimate.get('estimate_hours', 0) if statistical_estimate else 0)) / 
                         (statistical_estimate.get('estimate_hours', 1) if statistical_estimate else 1)) * 100, 1
                    ) if statistical_estimate else None,
                })
                # Update progress to show success
                report_progress(i, total_tasks, task_name, f"✓ {result.get('estimate_hours', 0):.1f}h")
            except Exception as e:
                error_msg = f"Failed to estimate {task_name}: {str(e)}"
                logger.warning(error_msg)
                errors.append({
                    'task_id': task.get('id'),
                    'name': task_name,
                    'error': str(e)
                })
                # Update progress to show error
                report_progress(i, total_tasks, task_name, "✗ failed")
                # Continue processing other tasks
        
        total_hours = sum(r['estimate_hours'] for r in results)
        avg_confidence = sum(r['confidence'] for r in results) / len(results) if results else 0
        
        # Calculate comparison stats if we have both estimates
        comparison_stats = None
        if statistical_estimator and results:
            tasks_with_comparison = [r for r in results if r.get('statistical_estimate')]
            if tasks_with_comparison:
                mlx_avg = sum(r['estimate_hours'] for r in tasks_with_comparison) / len(tasks_with_comparison)
                stat_avg = sum(r['statistical_estimate']['estimate_hours'] for r in tasks_with_comparison) / len(tasks_with_comparison)
                comparison_stats = {
                    'tasks_compared': len(tasks_with_comparison),
                    'mlx_average_hours': round(mlx_avg, 1),
                    'statistical_average_hours': round(stat_avg, 1),
                    'average_difference': round(mlx_avg - stat_avg, 1),
                    'average_difference_percent': round(((mlx_avg - stat_avg) / stat_avg * 100) if stat_avg > 0 else 0, 1),
                }
        
        response_data = {
            "status": "success",
            "method": "coreml_neural_engine_batch" if use_coreml_batch else ("mlx_enhanced_batch" if use_mlx else "statistical_batch"),
            "generated": datetime.now().isoformat(),
            "total_tasks": len(results),
            "successful": len(results),
            "failed": len(errors),
            "total_hours": round(total_hours, 1),
            "average_hours": round(total_hours / len(results), 1) if results else 0,
            "average_confidence": round(avg_confidence * 100, 1),
            "comparison_stats": comparison_stats,
            "tasks": results,
            "errors": errors if errors else None,
        }
        
        # Auto-save results if output_path not provided (save to docs/)
        if not output_path:
            from ..utils import find_project_root
            project_root = find_project_root()
            output_path = str(project_root / 'docs' / f'TASK_ESTIMATION_BATCH_{datetime.now().strftime("%Y%m%d_%H%M%S")}.json')
        
        # Save results
        try:
            output_file = Path(output_path)
            output_file.parent.mkdir(parents=True, exist_ok=True)
            with open(output_file, 'w') as f:
                json_lib.dump(response_data, f, indent=2)
            response_data['saved_to'] = str(output_file)
        except Exception as e:
            logger.warning(f"Failed to save results to {output_path}: {e}")
        
        return json.dumps(response_data, indent=2)

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown estimation action: {action}. Use 'estimate', 'analyze', 'stats', or 'batch'.",
        }, indent=2)



def task_workflow(
    action: str = "sync",
    # sync params
    dry_run: bool = False,
    # approve params
    status: str = "Review",
    new_status: str = "Todo",
    clarification_none: bool = True,
    filter_tag: Optional[str] = None,
    task_ids: Optional[str] = None,  # JSON list
    # clarify params
    sub_action: str = "list",  # list, resolve, batch (for clarify action)
    task_id: Optional[str] = None,
    clarification_text: Optional[str] = None,
    decision: Optional[str] = None,
    decisions_json: Optional[str] = None,
    move_to_todo: bool = True,
    # clarity params
    auto_apply: bool = False,
    output_format: str = "text",
    # cleanup params
    stale_threshold_hours: float = 2.0,
    # common
    output_path: Optional[str] = None,
) -> str:
    """
    Unified task workflow management tool.

    Args:
        action: "sync" for TODO↔Todo2 sync, "approve" for bulk approval, "clarify" for clarifications,
                "clarity" for task clarity improvement, "cleanup" for stale task cleanup
        dry_run: Preview changes without applying (sync, approve, cleanup)
        status: Filter tasks by status (approve)
        new_status: Target status (approve)
        clarification_none: Only tasks without clarification (approve)
        filter_tag: Filter by tag (approve)
        task_ids: JSON list of task IDs (approve)
        sub_action: "list", "resolve", or "batch" (clarify action)
        task_id: Task to resolve (clarify)
        clarification_text: Clarification response (clarify)
        decision: Decision made (clarify)
        decisions_json: Batch decisions as JSON (clarify)
        move_to_todo: Move resolved tasks to Todo (clarify)
        auto_apply: Auto-apply improvements (clarity action)
        output_format: Output format (clarity action)
        stale_threshold_hours: Hours before task is stale (cleanup action)
        output_path: Save results to file

    Returns:
        JSON string with workflow operation results
    """
    if action == "sync":
        from .todo_sync import sync_todo_tasks
        result = sync_todo_tasks(dry_run, output_path)
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "approve":
        from .batch_task_approval import batch_approve_tasks
        ids = None
        if task_ids:
            try:
                ids = json.loads(task_ids)
            except json.JSONDecodeError:
                return json.dumps({"status": "error", "error": "Invalid task_ids JSON"}, indent=2)
        result = batch_approve_tasks(
            status=status,
            new_status=new_status,
            clarification_none=clarification_none,
            filter_tag=filter_tag,
            task_ids=ids,
            dry_run=dry_run,
        )
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "clarify":
        from .task_clarification_resolution import (
            list_tasks_awaiting_clarification,
            resolve_multiple_clarifications,
            resolve_task_clarification,
        )

        if sub_action == "list":
            result = list_tasks_awaiting_clarification()
        elif sub_action == "resolve":
            if not task_id:
                return json.dumps({"status": "error", "error": "task_id required for resolve"}, indent=2)
            result = resolve_task_clarification(
                task_id, clarification_text, decision, move_to_todo, dry_run
            )
        elif sub_action == "batch":
            if not decisions_json:
                return json.dumps({"status": "error", "error": "decisions_json required for batch"}, indent=2)
            try:
                decisions = json.loads(decisions_json)
            except json.JSONDecodeError:
                return json.dumps({"status": "error", "error": "Invalid decisions_json"}, indent=2)
            result = resolve_multiple_clarifications(decisions, move_to_todo, dry_run)
        else:
            return json.dumps({"status": "error", "error": f"Unknown sub_action: {sub_action}. Use 'list', 'resolve', or 'batch'."}, indent=2)

        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "clarity":
        from .task_clarity_improver import analyze_task_clarity, improve_task_clarity
        if auto_apply:
            result = improve_task_clarity(auto_apply=True, output_path=output_path)
        else:
            result = analyze_task_clarity(output_format=output_format, output_path=output_path, dry_run=True)

        # Handle text format output
        if output_format == "text" and isinstance(result, dict) and "formatted_output" in result:
            return result["formatted_output"]
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "cleanup":
        from .stale_task_cleanup import cleanup_stale_tasks
        result = cleanup_stale_tasks(stale_threshold_hours, dry_run, output_path)
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown task_workflow action: {action}. Use 'sync', 'approve', 'clarify', 'clarity', or 'cleanup'.",
        }, indent=2)


