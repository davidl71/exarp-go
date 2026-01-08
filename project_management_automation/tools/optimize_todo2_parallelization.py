"""
MCP Tool Wrapper for Todo2 Parallelization Optimization

Identifies Todo2 tasks that can run in parallel based on dependency readiness and estimated time.
"""

import json
import logging
import time
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Import error handler
try:
    from ..error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
except ImportError:
    import sys
    from pathlib import Path
    server_dir = Path(__file__).parent.parent
    sys.path.insert(0, str(server_dir))
    try:
        from error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
    except ImportError:
        def format_success_response(data, message=None):
            return {"success": True, "data": data, "timestamp": time.time()}
        def format_error_response(error, error_code, include_traceback=False):
            return {"success": False, "error": {"code": str(error_code), "message": str(error)}}
        def log_automation_execution(name, duration, success, error=None):
            logger.info(f"{name}: {duration:.2f}s, success={success}")
        class ErrorCode:
            AUTOMATION_ERROR = "AUTOMATION_ERROR"


def optimize_todo2_parallelization(
    output_format: str = "markdown",
    output_path: Optional[str] = None
) -> str:
    """
    Identify Todo2 tasks that can run in parallel based on dependency readiness and estimated time.

    Args:
        output_format: Output format - "text", "json", or "markdown" (default: "markdown")
        output_path: Path for report output (default: docs/TODO2_PARALLELIZATION_OPTIMIZATION_REPORT.md)

    Returns:
        JSON string with parallelization analysis results
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root
        from project_management_automation.utils.todo2_mcp_client import (
            list_todos_mcp,
        )

        project_root = find_project_root()

        # Get tasks ready to start (dependencies completed)
        ready_tasks = list_todos_mcp(project_root=project_root, ready_only=True)

        # Get all tasks for analysis
        all_tasks = list_todos_mcp(project_root=project_root)
        task_map = {task.get('id'): task for task in all_tasks}

        # Group tasks by dependency readiness
        parallel_groups = _group_parallelizable_tasks(ready_tasks, all_tasks, task_map)

        # Calculate time savings
        time_savings = _calculate_time_savings(parallel_groups, all_tasks)

        # Generate execution plan
        execution_plan = _generate_execution_plan(parallel_groups, all_tasks)

        # Generate report
        report_path = output_path or 'docs/TODO2_PARALLELIZATION_OPTIMIZATION_REPORT.md'
        if report_path:
            report = _generate_report(
                parallel_groups,
                time_savings,
                execution_plan,
                ready_tasks,
                all_tasks,
                output_format
            )
            report_file = project_root / report_path
            report_file.parent.mkdir(parents=True, exist_ok=True)
            with open(report_file, 'w') as f:
                f.write(report)

        # Format response
        response_data = {
            'total_tasks': len(all_tasks),
            'ready_tasks': len(ready_tasks),
            'parallel_groups': len(parallel_groups),
            'total_parallelizable': sum(len(group) for group in parallel_groups),
            'estimated_time_savings_hours': time_savings['estimated_savings'],
            'sequential_time_hours': time_savings['sequential_time'],
            'parallel_time_hours': time_savings['parallel_time'],
            'report_path': str(Path(report_path).absolute()),
            'execution_plan': execution_plan
        }

        duration = time.time() - start_time
        log_automation_execution('optimize_todo2_parallelization', duration, True)

        if output_format == "json":
            return json.dumps(response_data, indent=2)
        else:
            return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('optimize_todo2_parallelization', duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def _group_parallelizable_tasks(
    ready_tasks: list[dict[str, Any]],
    all_tasks: list[dict[str, Any]],
    task_map: dict[str, dict[str, Any]]
) -> list[list[str]]:
    """Group tasks that can run in parallel."""
    # Simple grouping: all ready tasks can run in parallel
    # More sophisticated: group by estimated time, priority, etc.

    groups = []
    current_group = []

    for task in ready_tasks:
        task_id = task.get('id')
        estimated_hours = task.get('estimatedHours', 0) or task.get('estimated_hours', 0)

        # Group tasks with similar estimated time (within 2 hours)
        if current_group:
            first_task = task_map.get(current_group[0], {})
            first_hours = first_task.get('estimatedHours', 0) or first_task.get('estimated_hours', 0)

            if abs(estimated_hours - first_hours) <= 2:
                current_group.append(task_id)
            else:
                if current_group:
                    groups.append(current_group)
                current_group = [task_id]
        else:
            current_group = [task_id]

    if current_group:
        groups.append(current_group)

    return groups


def _calculate_time_savings(
    parallel_groups: list[list[str]],
    all_tasks: list[dict[str, Any]]
) -> dict[str, float]:
    """Calculate time savings from parallelization."""
    task_map = {task.get('id'): task for task in all_tasks}

    sequential_time = 0.0
    parallel_time = 0.0

    for group in parallel_groups:
        group_times = []
        for task_id in group:
            task = task_map.get(task_id, {})
            hours = task.get('estimatedHours', 0) or task.get('estimated_hours', 0)
            if hours > 0:
                group_times.append(hours)

        if group_times:
            sequential_time += sum(group_times)
            parallel_time += max(group_times)  # Longest task in group

    estimated_savings = sequential_time - parallel_time if sequential_time > 0 else 0

    return {
        'sequential_time': sequential_time,
        'parallel_time': parallel_time,
        'estimated_savings': estimated_savings
    }


def _generate_execution_plan(
    parallel_groups: list[list[str]],
    all_tasks: list[dict[str, Any]]
) -> list[dict[str, Any]]:
    """Generate execution plan with task groups."""
    task_map = {task.get('id'): task for task in all_tasks}

    plan = []
    for i, group in enumerate(parallel_groups, 1):
        group_tasks = []
        total_hours = 0

        for task_id in group:
            task = task_map.get(task_id, {})
            hours = task.get('estimatedHours', 0) or task.get('estimated_hours', 0)
            total_hours = max(total_hours, hours)  # Longest task

            group_tasks.append({
                'id': task_id,
                'name': task.get('name', ''),
                'estimated_hours': hours,
                'priority': task.get('priority', 'medium')
            })

        plan.append({
            'phase': i,
            'tasks': group_tasks,
            'estimated_hours': total_hours,
            'parallel_count': len(group_tasks)
        })

    return plan


def _generate_report(
    parallel_groups: list[list[str]],
    time_savings: dict[str, float],
    execution_plan: list[dict[str, Any]],
    ready_tasks: list[dict[str, Any]],
    all_tasks: list[dict[str, Any]],
    output_format: str
) -> str:
    """Generate markdown report."""
    {task.get('id'): task for task in all_tasks}

    report = f"""# Todo2 Parallelization Optimization Report

**Generated**: {time.strftime('%Y-%m-%d %H:%M:%S')}

## Summary

- **Total Tasks**: {len(all_tasks)}
- **Ready to Start**: {len(ready_tasks)}
- **Parallel Groups**: {len(parallel_groups)}
- **Total Parallelizable**: {sum(len(group) for group in parallel_groups)}
- **Estimated Time Savings**: {time_savings['estimated_savings']:.1f} hours
- **Sequential Time**: {time_savings['sequential_time']:.1f} hours
- **Parallel Time**: {time_savings['parallel_time']:.1f} hours

## Execution Plan

"""

    for phase in execution_plan:
        report += f"""### Phase {phase['phase']} ({phase['parallel_count']} tasks in parallel)

**Estimated Time**: {phase['estimated_hours']:.1f} hours

**Tasks**:
"""
        for task in phase['tasks']:
            report += f"- {task['name']} ({task['id']}) - {task['estimated_hours']:.1f}h - Priority: {task['priority']}\n"
        report += "\n"

    return report

