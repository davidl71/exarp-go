"""
Branch Merge Tools

Provides functions for merging tasks between branches in a Git-inspired task management system.
"""

import json
import logging
from typing import Dict, Any

from ..utils import find_project_root
from ..utils.branch_utils import (
    filter_tasks_by_branch,
    set_task_branch,
    get_task_branch,
)

logger = logging.getLogger(__name__)


def preview_merge(source_branch: str, target_branch: str) -> Dict[str, Any]:
    """
    Preview what would happen if merging source_branch into target_branch.
    
    Returns a dictionary with:
    - conflicts: List of tasks that would conflict
    - new_tasks: Tasks that would be added
    - updated_tasks: Tasks that would be updated
    """
    try:
        project_root = find_project_root()
        todo2_file = project_root / ".todo2" / "state.todo2.json"
        
        if not todo2_file.exists():
            return {
                "source_branch": source_branch,
                "target_branch": target_branch,
                "conflicts": [],
                "new_tasks": [],
                "updated_tasks": [],
                "message": "Todo2 state file not found"
            }
        
        with open(todo2_file) as f:
            data = json.load(f)
        
        tasks = data.get("todos", [])
        source_tasks = filter_tasks_by_branch(tasks, source_branch)
        target_tasks = filter_tasks_by_branch(tasks, target_branch)
        
        # Find conflicts (tasks that exist in both branches with different states)
        target_task_ids = {task.get("id"): task for task in target_tasks}
        conflicts = []
        new_tasks = []
        updated_tasks = []
        
        for source_task in source_tasks:
            task_id = source_task.get("id")
            if task_id in target_task_ids:
                target_task = target_task_ids[task_id]
                # Check if tasks differ
                if source_task != target_task:
                    conflicts.append({
                        "task_id": task_id,
                        "source": source_task,
                        "target": target_task
                    })
            else:
                # New task to be added
                new_tasks.append(source_task)
        
        return {
            "source_branch": source_branch,
            "target_branch": target_branch,
            "conflicts": conflicts,
            "new_tasks": new_tasks,
            "updated_tasks": updated_tasks,
            "conflict_count": len(conflicts),
            "new_task_count": len(new_tasks)
        }
    except Exception as e:
        logger.error(f"Error previewing merge: {e}", exc_info=True)
        return {
            "error": str(e),
            "source_branch": source_branch,
            "target_branch": target_branch
        }


def merge_branches(
    source_branch: str,
    target_branch: str,
    conflict_strategy: str = "newer",
    author: str = "system"
) -> Dict[str, Any]:
    """
    Merge tasks from source_branch into target_branch.
    
    Args:
        source_branch: Branch to merge from
        target_branch: Branch to merge into
        conflict_strategy: How to handle conflicts ("newer", "source", "target")
        author: Author of the merge operation
    
    Returns a dictionary with merge results.
    """
    try:
        project_root = find_project_root()
        todo2_file = project_root / ".todo2" / "state.todo2.json"
        
        if not todo2_file.exists():
            return {
                "error": "Todo2 state file not found",
                "source_branch": source_branch,
                "target_branch": target_branch
            }
        
        with open(todo2_file) as f:
            data = json.load(f)
        
        tasks = data.get("todos", [])
        filter_tasks_by_branch(tasks, source_branch)
        target_task_ids = {task.get("id"): task for task in filter_tasks_by_branch(tasks, target_branch)}
        
        merged_count = 0
        conflict_count = 0
        new_count = 0
        
        # Update tasks in place
        for i, task in enumerate(tasks):
            task_id = task.get("id")
            current_branch = get_task_branch(task)
            
            # If task is in source branch
            if current_branch == source_branch:
                if task_id in target_task_ids:
                    # Conflict - apply strategy
                    target_task = target_task_ids[task_id]
                    if conflict_strategy == "newer":
                        # Use task with newer updated_at timestamp
                        source_updated = task.get("updated_at", "")
                        target_updated = target_task.get("updated_at", "")
                        if source_updated > target_updated:
                            tasks[i] = set_task_branch(task, target_branch)
                            merged_count += 1
                        else:
                            conflict_count += 1
                    elif conflict_strategy == "source":
                        tasks[i] = set_task_branch(task, target_branch)
                        merged_count += 1
                    elif conflict_strategy == "target":
                        conflict_count += 1  # Keep target version
                else:
                    # New task - move to target branch
                    tasks[i] = set_task_branch(task, target_branch)
                    new_count += 1
                    merged_count += 1
        
        # Save updated tasks
        data["todos"] = tasks
        with open(todo2_file, "w") as f:
            json.dump(data, f, indent=2)
        
        return {
            "source_branch": source_branch,
            "target_branch": target_branch,
            "merged_count": merged_count,
            "new_tasks": new_count,
            "conflicts": conflict_count,
            "strategy": conflict_strategy,
            "author": author,
            "success": True
        }
    except Exception as e:
        logger.error(f"Error merging branches: {e}", exc_info=True)
        return {
            "error": str(e),
            "source_branch": source_branch,
            "target_branch": target_branch
        }


__all__ = ["merge_branches", "preview_merge"]

