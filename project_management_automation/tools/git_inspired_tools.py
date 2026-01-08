"""
Git-Inspired Task Management Tools

Provides MCP-compatible tools for Git-inspired task management features:
- Commit tracking and history
- Branch management
- Task diff/version comparison
- Git graph visualization
- Branch merging

Inspired by concepts from GitTask (https://github.com/Bengerthelorf/gittask)
Licensed under GPL-3.0. This implementation is original Python code.
See ATTRIBUTIONS.md for details.
"""

from typing import Optional
import json
import logging
from datetime import datetime
from pathlib import Path

from ..utils import find_project_root
from ..utils.branch_utils import (
    filter_tasks_by_branch,
    get_all_branch_statistics,
    get_all_branches,
    get_task_branch,
    set_task_branch,
)
from ..utils.commit_tracking import (
    get_commit_tracker,
    track_task_update,
)
from .branch_merge import merge_branches, preview_merge
from .git_graph import generate_commit_graph
from .task_diff import compare_task_versions

logger = logging.getLogger(__name__)


def get_task_commits(task_id: str, branch: Optional[str] = None, limit: int = 50) -> str:
    """
    Get commit history for a task.

    Returns JSON string with commit list.
    """
    try:
        tracker = get_commit_tracker()
        commits = tracker.get_commits_for_task(task_id, branch)

        # Sort by timestamp (newest first)
        commits.sort(key=lambda c: c.timestamp, reverse=True)

        # Limit results
        commits = commits[:limit]

        result = {
            "task_id": task_id,
            "branch": branch,
            "total_commits": len(commits),
            "commits": [
                {
                    "id": c.id,
                    "message": c.message,
                    "timestamp": c.timestamp.isoformat(),
                    "author": c.author,
                    "branch": c.branch,
                }
                for c in commits
            ],
        }

        return json.dumps(result, indent=2)
    except Exception as e:
        logger.error(f"Error getting task commits: {e}")
        return json.dumps({"error": str(e)}, indent=2)


def get_branch_commits(branch: str, limit: int = 100) -> str:
    """
    Get all commits for a branch.

    Returns JSON string with commit list.
    """
    try:
        tracker = get_commit_tracker()
        commits = tracker.get_commits_for_branch(branch)

        # Sort by timestamp (newest first)
        commits.sort(key=lambda c: c.timestamp, reverse=True)

        # Limit results
        commits = commits[:limit]

        result = {
            "branch": branch,
            "total_commits": len(commits),
            "commits": [
                {
                    "id": c.id,
                    "task_id": c.task_id,
                    "message": c.message,
                    "timestamp": c.timestamp.isoformat(),
                    "author": c.author,
                }
                for c in commits
            ],
        }

        return json.dumps(result, indent=2)
    except Exception as e:
        logger.error(f"Error getting branch commits: {e}")
        return json.dumps({"error": str(e)}, indent=2)


def list_branches() -> str:
    """
    List all branches (from tasks).

    Returns JSON string with branch list and statistics.
    """
    try:
        project_root = find_project_root()
        todo2_file = project_root / ".todo2" / "state.todo2.json"

        if not todo2_file.exists():
            return json.dumps({"branches": [], "statistics": {}}, indent=2)

        with open(todo2_file) as f:
            data = json.load(f)

        tasks = data.get("todos", [])
        branches = get_all_branches(tasks)
        statistics = get_all_branch_statistics(tasks)

        result = {
            "branches": sorted(branches),
            "statistics": dict(statistics.items()),
        }

        return json.dumps(result, indent=2)
    except Exception as e:
        logger.error(f"Error listing branches: {e}")
        return json.dumps({"error": str(e)}, indent=2)


def get_branch_tasks(branch: str) -> str:
    """
    Get all tasks in a branch.

    Returns JSON string with task list.
    """
    try:
        project_root = find_project_root()
        todo2_file = project_root / ".todo2" / "state.todo2.json"

        if not todo2_file.exists():
            return json.dumps({"branch": branch, "tasks": []}, indent=2)

        with open(todo2_file) as f:
            data = json.load(f)

        tasks = data.get("todos", [])
        branch_tasks = filter_tasks_by_branch(tasks, branch)

        result = {
            "branch": branch,
            "task_count": len(branch_tasks),
            "tasks": branch_tasks,
        }

        return json.dumps(result, indent=2)
    except Exception as e:
        logger.error(f"Error getting branch tasks: {e}")
        return json.dumps({"error": str(e)}, indent=2)


def compare_task_diff(
    task_id: str,
    commit1: Optional[str] = None,
    commit2: Optional[str] = None,
    time1: Optional[str] = None,
    time2: Optional[str] = None,
) -> str:
    """
    Compare two versions of a task.

    Returns JSON string with diff results.
    """
    try:
        result = compare_task_versions(
            task_id=task_id,
            version1_commit_id=commit1,
            version2_commit_id=commit2,
            version1_time=datetime.fromisoformat(time1) if time1 else None,
            version2_time=datetime.fromisoformat(time2) if time2 else None,
        )

        return json.dumps(result, indent=2, default=str)
    except Exception as e:
        logger.error(f"Error comparing task diff: {e}")
        return json.dumps({"error": str(e)}, indent=2)


def generate_graph(
    branch: Optional[str] = None,
    task_id: Optional[str] = None,
    format: str = "text",
    output_path: Optional[str] = None,
    max_commits: int = 50,
) -> str:
    """
    Generate commit graph visualization.

    Returns graph string (text or DOT format).
    """
    try:
        path = Path(output_path) if output_path else None
        graph = generate_commit_graph(
            branch=branch,
            task_id=task_id,
            format=format,
            output_path=path,
            max_commits=max_commits,
        )

        result = {
            "format": format,
            "branch": branch,
            "task_id": task_id,
            "graph": graph,
            "output_path": str(path) if path else None,
        }

        return json.dumps(result, indent=2)
    except Exception as e:
        logger.error(f"Error generating graph: {e}")
        return json.dumps({"error": str(e)}, indent=2)


def merge_branch_tools(
    source_branch: str,
    target_branch: str,
    conflict_strategy: str = "newer",
    author: str = "system",
    dry_run: bool = False,
) -> str:
    """
    Merge tasks from source branch to target branch.

    Returns JSON string with merge results.
    """
    try:
        if dry_run:
            result = preview_merge(source_branch, target_branch)
            result["dry_run"] = True
        else:
            result = merge_branches(
                source_branch=source_branch,
                target_branch=target_branch,
                conflict_strategy=conflict_strategy,
                author=author,
            )
            result["dry_run"] = False

        return json.dumps(result, indent=2, default=str)
    except Exception as e:
        logger.error(f"Error merging branches: {e}")
        return json.dumps({"error": str(e)}, indent=2)


def set_task_branch_tool(task_id: str, branch: str) -> str:
    """
    Set branch for a task.

    Returns JSON string with result.
    """
    try:
        project_root = find_project_root()
        todo2_file = project_root / ".todo2" / "state.todo2.json"

        if not todo2_file.exists():
            return json.dumps({"error": "Todo2 state file not found"}, indent=2)

        with open(todo2_file) as f:
            data = json.load(f)

        tasks = data.get("todos", [])
        task_index = next((i for i, t in enumerate(tasks) if t.get("id") == task_id), None)

        if task_index is None:
            return json.dumps({"error": f"Task {task_id} not found"}, indent=2)

        old_task = tasks[task_index].copy()
        new_task = set_task_branch(tasks[task_index], branch)
        tasks[task_index] = new_task

        # Save
        data["todos"] = tasks
        with open(todo2_file, "w") as f:
            json.dump(data, f, indent=2)

        # Track commit
        track_task_update(
            task_id=task_id,
            old_state=old_task,
            new_state=new_task,
            author="system",
            branch=branch,
        )

        result = {
            "task_id": task_id,
            "old_branch": get_task_branch(old_task),
            "new_branch": branch,
            "success": True,
        }

        return json.dumps(result, indent=2)
    except Exception as e:
        logger.error(f"Error setting task branch: {e}")
        return json.dumps({"error": str(e)}, indent=2)


__all__ = [
    "get_task_commits",
    "get_branch_commits",
    "list_branches",
    "get_branch_tasks",
    "compare_task_diff",
    "generate_graph",
    "merge_branch_tools",
    "set_task_branch_tool",
]
