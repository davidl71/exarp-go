"""
MCP Tool Wrapper for Todo2 Dependency Analysis

Analyzes Todo2 task dependency chains, identifies circular dependencies, and visualizes critical paths.
"""

import json
import logging
import time
from collections import defaultdict
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


def analyze_todo2_dependencies(
    output_format: str = "text",
    output_path: Optional[str] = None
) -> str:
    """
    Analyze Todo2 task dependency chains, identify circular dependencies, and visualize critical paths.

    Args:
        output_format: Output format - "text" or "json" (default: "text")
        output_path: Path for report output (default: docs/TODO2_DEPENDENCY_ANALYSIS_REPORT.md)

    Returns:
        JSON string with dependency analysis results
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root
        from project_management_automation.utils.todo2_mcp_client import (
            list_todos_mcp,
        )

        project_root = find_project_root()

        # Get all tasks
        all_tasks = list_todos_mcp(project_root=project_root)

        # Build dependency graph
        graph = _build_dependency_graph(all_tasks)

        # Detect circular dependencies
        cycles = _detect_cycles(graph, all_tasks)

        # Identify critical paths
        critical_paths = _find_critical_paths(graph, all_tasks)

        # Calculate metrics
        metrics = _calculate_metrics(graph, all_tasks)

        # Generate report
        report_path = output_path or 'docs/TODO2_DEPENDENCY_ANALYSIS_REPORT.md'
        if report_path:
            report = _generate_report(
                graph,
                cycles,
                critical_paths,
                metrics,
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
            'tasks_with_dependencies': metrics['tasks_with_dependencies'],
            'circular_dependencies': len(cycles),
            'critical_paths_count': len(critical_paths),
            'max_depth': metrics['max_depth'],
            'longest_chain': metrics['longest_chain'],
            'report_path': str(Path(report_path).absolute()),
            'cycles': cycles[:10],  # Limit to first 10 for response
            'critical_paths': critical_paths[:5]  # Limit to first 5 for response
        }

        duration = time.time() - start_time
        log_automation_execution('analyze_todo2_dependencies', duration, True)

        if output_format == "json":
            return json.dumps(response_data, indent=2)
        else:
            return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('analyze_todo2_dependencies', duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def _build_dependency_graph(tasks: list[dict[str, Any]]) -> dict[str, list[str]]:
    """Build directed graph from task dependencies."""
    graph = defaultdict(list)
    task_ids = {task.get('id') for task in tasks}

    for task in tasks:
        task_id = task.get('id')
        deps = task.get('dependencies', []) or task.get('dependsOn', [])

        # Normalize dependencies (handle both string and dict formats)
        for dep in deps:
            dep_id = dep if isinstance(dep, str) else dep.get('id')
            if dep_id and dep_id in task_ids:
                graph[dep_id].append(task_id)  # dep_id -> task_id (dependency -> dependent)

    return dict(graph)


def _detect_cycles(graph: dict[str, list[str]], tasks: list[dict[str, Any]]) -> list[list[str]]:
    """Detect circular dependencies using DFS."""
    cycles = []
    visited = set()
    rec_stack = set()
    path = []

    def dfs(node: str) -> bool:
        if node in rec_stack:
            # Found cycle
            cycle_start = path.index(node)
            cycle = path[cycle_start:] + [node]
            cycles.append(cycle)
            return True

        if node in visited:
            return False

        visited.add(node)
        rec_stack.add(node)
        path.append(node)

        for neighbor in graph.get(node, []):
            if dfs(neighbor):
                return True

        rec_stack.remove(node)
        path.pop()
        return False

    task_ids = {task.get('id') for task in tasks}
    for task_id in task_ids:
        if task_id not in visited:
            dfs(task_id)

    return cycles


def _find_critical_paths(graph: dict[str, list[str]], tasks: list[dict[str, Any]]) -> list[list[str]]:
    """Find critical paths (longest dependency chains) using topological sort."""
    # Build reverse graph for topological sort
    in_degree = defaultdict(int)
    task_ids = {task.get('id') for task in tasks}

    for task_id in task_ids:
        in_degree[task_id] = 0

    for _node, neighbors in graph.items():
        for neighbor in neighbors:
            in_degree[neighbor] += 1

    # Find longest paths
    longest_paths = []
    max_length = 0

    def dfs_longest(node: str, path: list[str], length: int):
        nonlocal max_length, longest_paths

        if length > max_length:
            max_length = length
            longest_paths = [path[:]]
        elif length == max_length:
            longest_paths.append(path[:])

        for neighbor in graph.get(node, []):
            if neighbor not in path:  # Avoid cycles
                path.append(neighbor)
                dfs_longest(neighbor, path, length + 1)
                path.pop()

    # Start from nodes with no dependencies (in_degree == 0)
    for task_id in task_ids:
        if in_degree[task_id] == 0:
            dfs_longest(task_id, [task_id], 1)

    return longest_paths


def _calculate_metrics(graph: dict[str, list[str]], tasks: list[dict[str, Any]]) -> dict[str, Any]:
    """Calculate dependency metrics."""
    task_ids = {task.get('id') for task in tasks}

    tasks_with_deps = sum(
        1 for task in tasks
        if task.get('dependencies') or task.get('dependsOn')
    )

    # Calculate max depth
    def get_depth(node: str, visited: set[str] = None) -> int:
        if visited is None:
            visited = set()
        if node in visited:
            return 0
        visited.add(node)

        max_dep_depth = 0
        for dep in graph.get(node, []):
            max_dep_depth = max(max_dep_depth, get_depth(dep, visited.copy()))

        return max_dep_depth + 1

    max_depth = 0
    for task_id in task_ids:
        depth = get_depth(task_id)
        max_depth = max(max_depth, depth)

    # Find longest chain
    longest_chain = []
    longest_paths = _find_critical_paths(graph, tasks)
    if longest_paths:
        longest_chain = max(longest_paths, key=len)

    return {
        'tasks_with_dependencies': tasks_with_deps,
        'max_depth': max_depth,
        'longest_chain': longest_chain,
        'longest_chain_length': len(longest_chain)
    }


def _generate_report(
    graph: dict[str, list[str]],
    cycles: list[list[str]],
    critical_paths: list[list[str]],
    metrics: dict[str, Any],
    tasks: list[dict[str, Any]],
    output_format: str
) -> str:
    """Generate markdown report."""
    task_map = {task.get('id'): task for task in tasks}

    report = f"""# Todo2 Dependency Analysis Report

**Generated**: {time.strftime('%Y-%m-%d %H:%M:%S')}

## Summary

- **Total Tasks**: {len(tasks)}
- **Tasks with Dependencies**: {metrics['tasks_with_dependencies']}
- **Circular Dependencies**: {len(cycles)}
- **Critical Paths**: {len(critical_paths)}
- **Max Depth**: {metrics['max_depth']}
- **Longest Chain**: {metrics['longest_chain_length']} tasks

## Circular Dependencies

"""

    if not cycles:
        report += "✅ **No circular dependencies found!**\n\n"
    else:
        for i, cycle in enumerate(cycles, 1):
            report += f"### Cycle {i}\n\n"
            cycle_names = [task_map.get(tid, {}).get('name', tid) for tid in cycle]
            report += " → ".join(cycle_names) + "\n\n"

    report += "## Critical Paths\n\n"
    if critical_paths:
        for i, path in enumerate(critical_paths[:5], 1):  # Show top 5
            report += f"### Path {i} ({len(path)} tasks)\n\n"
            path_names = [task_map.get(tid, {}).get('name', tid) for tid in path]
            report += " → ".join(path_names) + "\n\n"
    else:
        report += "No critical paths found.\n\n"

    report += "## Dependency Tree\n\n"
    # Build tree structure
    roots = [tid for tid in task_map.keys() if not any(tid in deps for deps in graph.values())]

    def build_tree(node: str, depth: int = 0, visited: set[str] = None) -> str:
        if visited is None:
            visited = set()
        if node in visited:
            return ""
        visited.add(node)

        task = task_map.get(node, {})
        name = task.get('name', node)
        indent = "  " * depth
        result = f"{indent}- {name} ({node})\n"

        for child in graph.get(node, []):
            result += build_tree(child, depth + 1, visited.copy())

        return result

    for root in roots[:10]:  # Show first 10 root tasks
        report += build_tree(root)
        report += "\n"

    return report

