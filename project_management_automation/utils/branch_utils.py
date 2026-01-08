"""
Branch utilities for Git-inspired task organization.

Provides functions for managing task branches (work streams) using tags.

Inspired by concepts from GitTask (https://github.com/Bengerthelorf/gittask)
Licensed under GPL-3.0. This implementation is original Python code.
See ATTRIBUTIONS.md for details.
"""

import logging
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Branch tag prefix
BRANCH_TAG_PREFIX = "branch:"
MAIN_BRANCH = "main"


def extract_branch_from_tags(tags: list[str]) -> Optional[str]:
    """
    Extract branch name from task tags.
    
    Looks for tags with 'branch:' prefix.
    
    Args:
        tags: List of task tags
        
    Returns:
        Branch name if found, None otherwise
    """
    if not tags:
        return None
    
    for tag in tags:
        if tag.startswith(BRANCH_TAG_PREFIX):
            branch = tag[len(BRANCH_TAG_PREFIX):].strip()
            if branch:
                return branch
    
    return None


def get_task_branch(task: dict[str, Any]) -> str:
    """
    Get branch for a task from its tags.
    
    Args:
        task: Task dictionary
        
    Returns:
        Branch name (defaults to 'main' if not found)
    """
    tags = task.get("tags", [])
    branch = extract_branch_from_tags(tags)
    return branch or MAIN_BRANCH


def set_task_branch(task: dict[str, Any], branch: str) -> dict[str, Any]:
    """
    Set branch for a task by updating its tags.
    
    Removes existing branch tag and adds new one.
    
    Args:
        task: Task dictionary (will be modified)
        branch: Branch name to set
        
    Returns:
        Modified task dictionary
    """
    tags = task.get("tags", [])
    
    # Remove existing branch tags
    tags = [tag for tag in tags if not tag.startswith(BRANCH_TAG_PREFIX)]
    
    # Add new branch tag (skip if main branch)
    if branch != MAIN_BRANCH:
        tags.append(f"{BRANCH_TAG_PREFIX}{branch}")
    
    task["tags"] = tags
    return task


def create_branch_tag(branch: str) -> str:
    """
    Create a branch tag string.
    
    Args:
        branch: Branch name
        
    Returns:
        Tag string (e.g., 'branch:feature-auth')
    """
    if branch == MAIN_BRANCH:
        return ""  # Main branch doesn't need a tag
    return f"{BRANCH_TAG_PREFIX}{branch}"


def get_all_branches(tasks: list[dict[str, Any]]) -> set[str]:
    """
    Extract all unique branches from a list of tasks.
    
    Args:
        tasks: List of task dictionaries
        
    Returns:
        Set of branch names (always includes 'main')
    """
    branches = {MAIN_BRANCH}
    
    for task in tasks:
        branch = get_task_branch(task)
        branches.add(branch)
    
    return branches


def filter_tasks_by_branch(tasks: list[dict[str, Any]], branch: str) -> list[dict[str, Any]]:
    """
    Filter tasks to only those in a specific branch.
    
    Args:
        tasks: List of task dictionaries
        branch: Branch name to filter by
        
    Returns:
        Filtered list of tasks
    """
    return [task for task in tasks if get_task_branch(task) == branch]


def get_branch_statistics(tasks: list[dict[str, Any]], branch: str) -> dict[str, Any]:
    """
    Get statistics for a specific branch.
    
    Args:
        tasks: List of task dictionaries
        branch: Branch name
        
    Returns:
        Dictionary with branch statistics:
        - task_count: Total tasks in branch
        - by_status: Task count by status
        - completed_count: Number of completed tasks
        - completion_rate: Percentage completed (0-100)
    """
    branch_tasks = filter_tasks_by_branch(tasks, branch)
    
    status_counts: dict[str, int] = {}
    completed_count = 0
    
    for task in branch_tasks:
        status = task.get("status", "todo")
        status_counts[status] = status_counts.get(status, 0) + 1
        
        if status.lower() in ("completed", "done"):
            completed_count += 1
    
    total = len(branch_tasks)
    completion_rate = (completed_count / total * 100) if total > 0 else 0.0
    
    return {
        "branch": branch,
        "task_count": total,
        "by_status": status_counts,
        "completed_count": completed_count,
        "completion_rate": round(completion_rate, 2),
    }


def get_all_branch_statistics(tasks: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    """
    Get statistics for all branches.
    
    Args:
        tasks: List of task dictionaries
        
    Returns:
        Dictionary mapping branch name to statistics
    """
    branches = get_all_branches(tasks)
    return {branch: get_branch_statistics(tasks, branch) for branch in branches}


__all__ = [
    "BRANCH_TAG_PREFIX",
    "MAIN_BRANCH",
    "extract_branch_from_tags",
    "get_task_branch",
    "set_task_branch",
    "create_branch_tag",
    "get_all_branches",
    "filter_tasks_by_branch",
    "get_branch_statistics",
    "get_all_branch_statistics",
]
