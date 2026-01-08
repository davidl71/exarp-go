"""
Task Clarity Improver Tool

Analyzes and improves task clarity by:
1. Adding time estimates (1-4 hours for parallelizable tasks)
2. Renaming tasks to start with action verbs
3. Removing unnecessary dependencies
4. Breaking down large tasks

Improves clarity score metrics:
- has_estimate: Tasks with estimatedHours > 0
- small_enough: Tasks with estimatedHours <= 4
- clear_name: Tasks starting with action verbs
- no_dependencies: Tasks without dependencies
"""

import json
import logging
from pathlib import Path
from typing import Any, Optional

from ..utils import find_project_root
from .task_duration_estimator import estimate_task_duration, estimate_task_duration_detailed

logger = logging.getLogger(__name__)

# Action verbs that improve clarity
ACTION_VERBS = [
    'add', 'implement', 'create', 'fix', 'update', 'remove',
    'refactor', 'migrate', 'integrate', 'test', 'document', 'extend',
    'improve', 'optimize', 'enhance', 'replace', 'delete', 'build',
    'setup', 'configure', 'deploy', 'verify', 'validate', 'review'
]


def analyze_task_clarity(
    output_format: str = "json",
    output_path: Optional[str] = None,
    dry_run: bool = True
) -> dict[str, Any]:
    """
    Analyze task clarity and suggest improvements.

    Args:
        output_format: Output format - "json" or "text"
        output_path: Optional path to save results
        dry_run: If True, only analyze without making changes

    Returns:
        Dictionary with analysis results and suggestions
    """
    project_root = find_project_root()

    # Try Todo2 MCP first (preferred)
    from project_management_automation.utils.todo2_mcp_client import list_todos_mcp

    try:
        todos = list_todos_mcp(project_root=project_root)
        if not todos:
            # Fallback to file access
            state_file = project_root / ".todo2" / "state.todo2.json"
            if not state_file.exists():
                return {
                    "status": "error",
                    "error": f"State file not found: {state_file}"
                }
            with open(state_file) as f:
                data = json.load(f)
            todos = data.get('todos', [])
    except Exception as e:
        logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")
        # Fallback to file access
        state_file = project_root / ".todo2" / "state.todo2.json"
        if not state_file.exists():
            return {
                "status": "error",
                "error": f"State file not found: {state_file}"
            }
        try:
            with open(state_file) as f:
                data = json.load(f)
            todos = data.get('todos', [])
        except Exception as e2:
            logger.error(f"Failed to read state file: {e2}")
            return {
                "status": "error",
                "error": f"Failed to read state file: {e2}"
            }

    # Filter to pending tasks
    pending = [
        t for t in todos
        if t.get('status', '').lower() not in ['done', 'completed', 'cancelled']
    ]

    improvements = []
    stats = {
        'total_pending': len(pending),
        'needs_estimate': 0,
        'needs_rename': 0,
        'needs_dependency_removal': 0,
        'needs_breakdown': 0,
        'ready_to_improve': 0
    }

    for task in pending:
        task_id = task.get('id', '')
        name = task.get('name', '') or task.get('content', '')
        estimated_hours = task.get('estimatedHours', 0)
        depends_on = task.get('dependsOn', []) or task.get('dependencies', [])

        task_improvements = {
            'task_id': task_id,
            'current_name': name,
            'improvements': []
        }

        # 1. Check for time estimate
        if estimated_hours == 0:
            stats['needs_estimate'] += 1
            # Estimate based on task complexity using statistical methods
            suggested_hours = _estimate_task_hours(name, task.get('details', ''), task)

            # Get detailed estimation info for better recommendations
            try:
                detailed = estimate_task_duration_detailed(
                    name=name,
                    details=task.get('details', '') or task.get('long_description', ''),
                    tags=task.get('tags', []),
                    priority=task.get('priority', 'medium'),
                    use_historical=True
                )
                confidence = detailed.get('confidence', 0.5)
                method = detailed.get('method', 'unknown')
                reason = f'Task needs time estimate (suggested: {suggested_hours}h, confidence: {confidence:.0%}, method: {method})'
            except Exception:
                reason = f'Task needs time estimate for clarity (suggested: {suggested_hours}h)'

            task_improvements['improvements'].append({
                'type': 'add_estimate',
                'current': 0,
                'suggested': suggested_hours,
                'reason': reason
            })

        # 2. Check if estimate is too large (>4 hours) or complexity analysis (T-19)
        elif estimated_hours > 4:
            stats['needs_breakdown'] += 1
            # Try complexity analysis from agentic-tools
            complexity_suggestion = None
            try:
                from project_management_automation.utils.agentic_tools_client import analyze_task_complexity_mcp
                complexity_result = analyze_task_complexity_mcp(
                    task_id=task_id,
                    project_root=project_root,
                    complexity_threshold=7,
                    suggest_breakdown=True,
                    auto_create_subtasks=False
                )
                if complexity_result and complexity_result.get('complexity'):
                    complexity_score = complexity_result['complexity'].get('score', 0)
                    if complexity_score >= 7:
                        breakdown_suggestions = complexity_result.get('breakdown_suggestions', [])
                        complexity_suggestion = {
                            'complexity_score': complexity_score,
                            'breakdown_suggestions': breakdown_suggestions,
                            'reason': f'Task complexity score {complexity_score}/10 suggests breakdown needed'
                        }
            except Exception as e:
                logger.debug(f"Complexity analysis not available: {e}")

            if complexity_suggestion:
                task_improvements['improvements'].append({
                    'type': 'breakdown',
                    'current': estimated_hours,
                    'suggested': 'Split into smaller tasks',
                    'reason': complexity_suggestion['reason'],
                    'complexity_score': complexity_suggestion['complexity_score'],
                    'breakdown_suggestions': complexity_suggestion['breakdown_suggestions']
                })
            else:
                task_improvements['improvements'].append({
                    'type': 'breakdown',
                    'current': estimated_hours,
                    'suggested': 'Split into smaller tasks',
                    'reason': f'Task estimated at {estimated_hours}h exceeds 4h limit for parallelization'
                })

        # 3. Check for action verb in name
        name_lower = name.lower().strip()
        starts_with_verb = any(name_lower.startswith(verb) for verb in ACTION_VERBS)

        if not starts_with_verb:
            stats['needs_rename'] += 1
            suggested_name = _suggest_action_verb_name(name)
            if suggested_name != name:
                task_improvements['improvements'].append({
                    'type': 'rename',
                    'current': name,
                    'suggested': suggested_name,
                    'reason': 'Task name should start with action verb for clarity'
                })

        # 4. Check for dependencies
        if depends_on:
            stats['needs_dependency_removal'] += 1
            # Check if dependencies are actually needed
            can_remove = _can_remove_dependencies(task, todos)
            if can_remove['can_remove']:
                task_improvements['improvements'].append({
                    'type': 'remove_dependencies',
                    'current': depends_on,
                    'suggested': [],
                    'reason': can_remove['reason']
                })

        if task_improvements['improvements']:
            improvements.append(task_improvements)
            stats['ready_to_improve'] += 1

    # Calculate potential clarity score improvement
    current_clarity = _calculate_clarity_score(pending)
    potential_clarity = _calculate_potential_clarity_score(pending, improvements)
    improvement_potential = potential_clarity - current_clarity

    recommendations = _generate_recommendations(stats, improvement_potential)

    result = {
        'status': 'success',
        'analysis': {
            'current_clarity_score': current_clarity,
            'potential_clarity_score': potential_clarity,
            'improvement_potential': improvement_potential,
            'stats': stats,
            'improvements': improvements,
            'recommendations': recommendations
        }
    }

    # Format output
    if output_format == "text":
        result['formatted_output'] = _format_text_output(result)

    # Save if path provided
    if output_path:
        output_file = Path(output_path)
        with open(output_file, 'w') as f:
            json.dump(result, f, indent=2)
        result['output_path'] = str(output_file)

    return result


def _estimate_task_hours(name: str, details: str, task: dict | None = None) -> float:
    """
    Estimate task hours using MLX-enhanced statistical methods.

    Uses MLX semantic understanding combined with historical data when available,
    falls back to keyword heuristics if MLX unavailable.
    """
    try:
        # Extract task metadata for better estimation
        tags = task.get('tags', []) if task else []
        priority = task.get('priority', 'medium') if task else 'medium'

        # Try MLX-enhanced estimator first (better accuracy)
        try:
            from .mlx_task_estimator import estimate_task_duration_mlx_enhanced
            return estimate_task_duration_mlx_enhanced(
                name=name,
                details=details or task.get('details', '') or task.get('long_description', '') if task else '',
                tags=tags,
                priority=priority,
                use_historical=True,
                use_mlx=True,
                mlx_weight=0.3,  # 30% MLX, 70% statistical
            )
        except ImportError:
            # MLX not available, fall back to statistical-only
            pass

        # Use statistical estimator (fallback or if MLX disabled)
        return estimate_task_duration(
            name=name,
            details=details or task.get('details', '') or task.get('long_description', '') if task else '',
            tags=tags,
            priority=priority,
            use_historical=True
        )
    except Exception as e:
        logger.warning(f"Estimation failed, using fallback: {e}")
        # Fallback to simple heuristic
        text = (name + " " + details).lower()
        if any(word in text for word in ['quick', 'simple', 'minor', 'small', 'fix typo']):
            return 1.0
        elif any(word in text for word in ['implement', 'add', 'create', 'setup']):
            return 2.0
        elif any(word in text for word in ['refactor', 'migrate', 'integrate', 'update']):
            return 3.0
        elif any(word in text for word in ['complex', 'major', 'rewrite', 'redesign']):
            return 4.0
        else:
            return 2.0  # Default estimate


def _suggest_action_verb_name(name: str) -> str:
    """Suggest a name starting with an action verb."""
    name_lower = name.lower().strip()

    # Check if it already starts with a verb
    if any(name_lower.startswith(verb) for verb in ACTION_VERBS):
        return name

    # Try to find appropriate verb based on content
    if any(word in name_lower for word in ['bug', 'error', 'issue', 'fix', 'broken']):
        return f"Fix {name}"
    elif any(word in name_lower for word in ['new', 'add', 'create']):
        return f"Add {name}"
    elif any(word in name_lower for word in ['update', 'change', 'modify']):
        return f"Update {name}"
    elif any(word in name_lower for word in ['remove', 'delete', 'cleanup']):
        return f"Remove {name}"
    elif any(word in name_lower for word in ['test', 'testing']):
        return f"Test {name}"
    elif any(word in name_lower for word in ['doc', 'documentation', 'readme']):
        return f"Document {name}"
    else:
        # Default: use "Implement" as it's generic
        return f"Implement {name}"


def _can_remove_dependencies(task: dict, all_tasks: list[dict]) -> dict[str, Any]:
    """Check if task dependencies can be safely removed."""
    depends_on = task.get('dependsOn', []) or task.get('dependencies', [])

    if not depends_on:
        return {'can_remove': False, 'reason': 'No dependencies'}

    # Check if dependencies exist
    missing_deps = []
    for dep_id in depends_on:
        if not any(t.get('id') == dep_id for t in all_tasks):
            missing_deps.append(dep_id)

    if missing_deps:
        return {
            'can_remove': True,
            'reason': f'Dependencies {missing_deps} do not exist - can be removed'
        }

    # Check if dependencies are completed
    completed_deps = []
    for dep_id in depends_on:
        dep_task = next((t for t in all_tasks if t.get('id') == dep_id), None)
        if dep_task and dep_task.get('status', '').lower() in ['done', 'completed']:
            completed_deps.append(dep_id)

    if completed_deps:
        return {
            'can_remove': True,
            'reason': f'Dependencies {completed_deps} are already completed - can be removed'
        }

    return {'can_remove': False, 'reason': 'Dependencies are valid and needed'}


def _calculate_clarity_score(tasks: list[dict]) -> float:
    """Calculate current clarity score."""
    if not tasks:
        return 0.0

    action_verbs = ACTION_VERBS

    has_estimate = sum(1 for t in tasks if t.get('estimatedHours', 0) > 0)
    has_tags = sum(1 for t in tasks if t.get('tags'))
    small_enough = sum(1 for t in tasks if 0 < t.get('estimatedHours', 0) <= 4)
    clear_name = sum(1 for t in tasks if any(
        (t.get('name', '') or t.get('content', '')).lower().startswith(v)
        for v in action_verbs
    ))
    no_deps = sum(1 for t in tasks if not (t.get('dependsOn') or t.get('dependencies')))

    total = len(tasks)
    return (has_estimate + has_tags + small_enough + clear_name + no_deps) / (5 * total) * 100


def _calculate_potential_clarity_score(tasks: list[dict], improvements: list[dict]) -> float:
    """Calculate potential clarity score after improvements."""
    if not tasks:
        return 0.0

    # Create improvement map
    improvement_map = {imp['task_id']: imp for imp in improvements}

    # Simulate improvements
    improved_tasks = []
    for task in tasks:
        task_id = task.get('id')
        improved = task.copy()

        if task_id in improvement_map:
            for improvement in improvement_map[task_id]['improvements']:
                if improvement['type'] == 'add_estimate':
                    improved['estimatedHours'] = improvement['suggested']
                elif improvement['type'] == 'rename':
                    improved['name'] = improvement['suggested']
                elif improvement['type'] == 'remove_dependencies':
                    improved['dependsOn'] = []
                    improved['dependencies'] = []

        improved_tasks.append(improved)

    return _calculate_clarity_score(improved_tasks)


def _generate_recommendations(stats: dict, improvement_potential: float) -> list[str]:
    """Generate recommendations based on analysis."""
    recommendations = []

    if stats['needs_estimate'] > 0:
        recommendations.append(
            f"Add time estimates to {stats['needs_estimate']} tasks "
            f"(target: 1-4 hours for parallelization)"
        )

    if stats['needs_rename'] > 0:
        recommendations.append(
            f"Rename {stats['needs_rename']} tasks to start with action verbs "
            f"(e.g., 'Implement X', 'Fix Y', 'Add Z')"
        )

    if stats['needs_dependency_removal'] > 0:
        recommendations.append(
            f"Review and remove unnecessary dependencies from "
            f"{stats['needs_dependency_removal']} tasks"
        )

    if stats['needs_breakdown'] > 0:
        recommendations.append(
            f"Break down {stats['needs_breakdown']} large tasks (>4h) "
            f"into smaller, parallelizable tasks"
        )

    if improvement_potential > 20:
        recommendations.append(
            f"High improvement potential: +{improvement_potential:.1f}% clarity score "
            f"if all improvements are applied"
        )

    return recommendations


def _format_text_output(result: dict) -> str:
    """Format analysis results as text."""
    analysis = result['analysis']
    stats = analysis['stats']

    lines = [
        "=" * 70,
        "TASK CLARITY ANALYSIS",
        "=" * 70,
        "",
        f"Current Clarity Score: {analysis['current_clarity_score']:.1f}%",
        f"Potential Clarity Score: {analysis['potential_clarity_score']:.1f}%",
        f"Improvement Potential: +{analysis['improvement_potential']:.1f}%",
        "",
        "Statistics:",
        f"  Total Pending Tasks: {stats['total_pending']}",
        f"  Needs Estimate: {stats['needs_estimate']}",
        f"  Needs Rename: {stats['needs_rename']}",
        f"  Needs Dependency Removal: {stats['needs_dependency_removal']}",
        f"  Needs Breakdown: {stats['needs_breakdown']}",
        f"  Ready to Improve: {stats['ready_to_improve']}",
        "",
    ]

    if analysis['recommendations']:
        lines.append("Recommendations:")
        for i, rec in enumerate(analysis['recommendations'], 1):
            lines.append(f"  {i}. {rec}")
        lines.append("")

    if analysis['improvements']:
        lines.append("Task Improvements Needed:")
        lines.append("")
        for task_imp in analysis['improvements'][:10]:  # Show first 10
            lines.append(f"Task: {task_imp['current_name'][:50]}")
            lines.append(f"  ID: {task_imp['task_id']}")
            for imp in task_imp['improvements']:
                lines.append(f"  - {imp['type']}: {imp['reason']}")
                if imp['type'] in ['add_estimate', 'rename']:
                    lines.append(f"    Current: {imp['current']}")
                    lines.append(f"    Suggested: {imp['suggested']}")
            lines.append("")

        if len(analysis['improvements']) > 10:
            lines.append(f"... and {len(analysis['improvements']) - 10} more tasks")

    lines.append("=" * 70)

    return "\n".join(lines)


async def improve_task_clarity_async(
    auto_apply: bool = False,
    output_path: Optional[str] = None
) -> dict[str, Any]:
    """
    Improve task clarity by applying suggested improvements using MCP.

    Args:
        auto_apply: If True, automatically apply improvements
        output_path: Optional path to save results

    Returns:
        Dictionary with improvement results
    """
    # First analyze
    analysis_result = analyze_task_clarity(output_format="json", dry_run=not auto_apply)

    if analysis_result['status'] != 'success':
        return analysis_result

    if not auto_apply:
        return {
            'status': 'success',
            'message': 'Analysis complete. Set auto_apply=True to apply improvements.',
            'analysis': analysis_result['analysis']
        }

    # Apply improvements using MCP
    try:
        from ..scripts.base.mcp_client import get_mcp_client

        project_root = find_project_root()
        client = get_mcp_client(project_root)
        working_directory = str(project_root)

        # Get project ID from git remote
        from ..utils.todo2_utils import get_current_project_id
        project_id = get_current_project_id(project_root)

        if not project_id:
            # Fallback to file-based approach if no project ID
            return _apply_improvements_file_based(analysis_result, output_path)

        improvements = analysis_result['analysis']['improvements']
        applied = 0
        failed = 0

        # Use batch operations for efficiency
        batch_ops = []
        for task_imp in improvements:
            task_id = task_imp['task_id']
            updates = {}

            for improvement in task_imp['improvements']:
                if improvement['type'] == 'add_estimate':
                    updates['estimatedHours'] = improvement['suggested']
                elif improvement['type'] == 'rename':
                    updates['name'] = improvement['suggested']
                elif improvement['type'] == 'remove_dependencies':
                    updates['dependsOn'] = []

            if updates:
                batch_ops.append({
                    'tool': 'update_task',
                    'arguments': {
                        'workingDirectory': working_directory,
                        'id': task_id,
                        **updates
                    }
                })

        # Apply in batches
        if batch_ops:
            results = await client.batch_operations(batch_ops, working_directory)
            applied = sum(1 for r in results if r is not None)
            failed = len(results) - applied

        return {
            'status': 'success',
            'message': f'Applied {applied} improvements to tasks ({failed} failed)',
            'improvements_applied': applied,
            'improvements_failed': failed,
            'analysis': analysis_result['analysis']
        }

    except Exception as e:
        logger.error(f"Failed to apply improvements via MCP: {e}", exc_info=True)
        # Fallback to file-based approach
        return _apply_improvements_file_based(analysis_result, output_path)


def _apply_improvements_file_based(analysis_result: dict, output_path: Optional[str]) -> dict[str, Any]:
    """Fallback: Apply improvements via MCP (preferred) or file directly (fallback)."""
    from project_management_automation.utils.todo2_mcp_client import update_todos_mcp

    project_root = find_project_root()
    improvements = analysis_result['analysis']['improvements']

    # Prepare updates for MCP
    updates = []
    applied = 0

    for task_imp in improvements:
        task_id = task_imp['task_id']
        task_updates = {'id': task_id}

        for improvement in task_imp['improvements']:
            if improvement['type'] == 'add_estimate':
                task_updates['estimatedHours'] = improvement['suggested']
                applied += 1
            elif improvement['type'] == 'rename':
                task_updates['name'] = improvement['suggested']
                applied += 1
            elif improvement['type'] == 'remove_dependencies':
                task_updates['dependencies'] = []
                applied += 1

        if len(task_updates) > 1:  # Has updates beyond just ID
            updates.append(task_updates)

    # Try Todo2 MCP first (preferred)
    try:
        if updates:
            success = update_todos_mcp(updates, project_root=project_root)
            if success:
                return {
                    'status': 'success',
                    'message': f'Applied {applied} improvements to tasks (via MCP)',
                    'improvements_applied': applied,
                    'analysis': analysis_result['analysis']
                }
    except Exception as e:
        logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")

    # Fallback to direct file access
    state_file = project_root / ".todo2" / "state.todo2.json"

    try:
        with open(state_file) as f:
            data = json.load(f)
        todos = data.get('todos', [])

        applied = 0
        for task_imp in improvements:
            task_id = task_imp['task_id']
            task = next((t for t in todos if t.get('id') == task_id), None)

            if not task:
                continue

            for improvement in task_imp['improvements']:
                if improvement['type'] == 'add_estimate':
                    task['estimatedHours'] = improvement['suggested']
                    applied += 1
                elif improvement['type'] == 'rename':
                    task['name'] = improvement['suggested']
                    applied += 1
                elif improvement['type'] == 'remove_dependencies':
                    task['dependsOn'] = []
                    task['dependencies'] = []
                    applied += 1

        # Save updated state
        data['todos'] = todos
        with open(state_file, 'w') as f:
            json.dump(data, f, indent=2)

        return {
            'status': 'success',
            'message': f'Applied {applied} improvements to tasks (file-based)',
            'improvements_applied': applied,
            'analysis': analysis_result['analysis']
        }

    except Exception as e:
        logger.error(f"Failed to apply improvements: {e}", exc_info=True)
        return {
            "status": "error",
            "error": str(e)
        }


def improve_task_clarity(
    auto_apply: bool = False,
    output_path: Optional[str] = None,
    use_mcp: bool = False
) -> dict[str, Any]:
    """
    Improve task clarity by applying suggested improvements.

    Args:
        auto_apply: If True, automatically apply improvements
        output_path: Optional path to save results
        use_mcp: If True, use MCP for updates (slower but follows best practices)

    Returns:
        Dictionary with improvement results
    """
    # First analyze
    analysis_result = analyze_task_clarity(output_format="json", dry_run=not auto_apply)

    if analysis_result['status'] != 'success':
        return analysis_result

    if not auto_apply:
        return {
            'status': 'success',
            'message': 'Analysis complete. Set auto_apply=True to apply improvements.',
            'analysis': analysis_result['analysis']
        }

    # Use MCP if requested, otherwise use file-based approach
    if use_mcp:
        import asyncio
        try:
            return asyncio.run(improve_task_clarity_async(auto_apply, output_path))
        except Exception as e:
            logger.warning(f"MCP approach failed, falling back to file-based: {e}")
            return _apply_improvements_file_based(analysis_result, output_path)
    else:
        return _apply_improvements_file_based(analysis_result, output_path)
