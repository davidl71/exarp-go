"""
Task locking utilities for preventing concurrent task assignment.

Provides atomic task assignment operations that prevent two agents from
working on the same task simultaneously.
"""

import json
import logging
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

from .file_lock import state_file_lock, task_lock
from .project_root import find_project_root

logger = logging.getLogger(__name__)


def atomic_assign_task(
    task_id: str,
    assignee_name: str,
    assignee_type: str = "agent",
    hostname: Optional[str] = None,
    assigned_by: str = "auto",
    timeout: float = 5.0,
) -> tuple[bool, Optional[str]]:
    """
    Atomically assign a task to an agent/host (with locking).

    Prevents race conditions where multiple agents try to assign the same task.

    Args:
        task_id: Task ID to assign
        assignee_name: Name of assignee (agent/human/host)
        assignee_type: Type of assignee ("agent", "human", "host")
        hostname: Optional hostname for remote hosts
        assigned_by: Who/what is making the assignment
        timeout: Lock timeout in seconds

    Returns:
        Tuple of (success: bool, error_message: Optional[str])
        - (True, None) if assignment successful
        - (False, "Task already assigned") if task already has assignee
        - (False, "Task not found") if task doesn't exist
        - (False, error_message) for other errors

    Example:
        success, error = atomic_assign_task("T-123", "backend-agent", "agent")
        if success:
            print("Task assigned successfully")
        else:
            print(f"Assignment failed: {error}")
    """
    try:
        with task_lock(task_id=task_id, timeout=timeout):
            # Load state
            project_root = find_project_root()
            state_file = project_root / ".todo2" / "state.todo2.json"

            if not state_file.exists():
                return (False, "State file not found")

            # Load state (bypass cache during locked operation)
            with open(state_file) as f:
                state = json.load(f)

            # Find task
            task = None
            task_index = -1
            for i, t in enumerate(state.get("todos", [])):
                if t.get("id") == task_id:
                    task = t
                    task_index = i
                    break

            if task is None:
                return (False, f"Task {task_id} not found")

            # Check if already assigned
            existing_assignee = task.get("assignee")
            if existing_assignee:
                existing_name = existing_assignee.get("name", "unknown")
                existing_type = existing_assignee.get("type", "unknown")
                return (
                    False,
                    f"Task already assigned to {existing_type}:{existing_name}"
                )

            # Assign task
            task["assignee"] = {
                "type": assignee_type,
                "name": assignee_name,
                "hostname": hostname,
                "assigned_at": datetime.utcnow().isoformat() + "Z",
                "assigned_by": assigned_by,
            }
            task["lastModified"] = datetime.utcnow().isoformat() + "Z"

            # Update in state
            state["todos"][task_index] = task

            # Save state (atomic write)
            # Create backup first
            if state_file.exists():
                backup_file = state_file.with_suffix('.json.bak')
                backup_file.write_text(state_file.read_text())

            # Write new state
            with open(state_file, 'w') as f:
                json.dump(state, f, indent=2)

            logger.info(f"Atomically assigned task {task_id} to {assignee_type}:{assignee_name}")
            return (True, None)

    except TimeoutError:
        return (False, f"Lock timeout after {timeout}s")
    except Exception as e:
        logger.error(f"Error in atomic_assign_task: {e}", exc_info=True)
        return (False, str(e))


def atomic_check_and_assign(
    task_id: str,
    assignee_name: str,
    assignee_type: str = "agent",
    hostname: Optional[str] = None,
    timeout: float = 5.0,
) -> dict[str, Any]:
    """
    Check if task is available and assign it atomically.

    Returns detailed result with status information.

    Args:
        task_id: Task ID to check and assign
        assignee_name: Name of assignee
        assignee_type: Type of assignee
        hostname: Optional hostname
        timeout: Lock timeout

    Returns:
        Dict with:
        - success: bool
        - assigned: bool (whether assignment was made)
        - reason: str (why assignment succeeded/failed)
        - existing_assignee: Optional[dict] (if already assigned)
    """
    try:
        with task_lock(task_id=task_id, timeout=timeout):
            project_root = find_project_root()
            state_file = project_root / ".todo2" / "state.todo2.json"

            if not state_file.exists():
                return {
                    "success": False,
                    "assigned": False,
                    "reason": "State file not found",
                }

            with open(state_file) as f:
                state = json.load(f)

            # Find task
            task = None
            task_index = -1
            for i, t in enumerate(state.get("todos", [])):
                if t.get("id") == task_id:
                    task = t
                    task_index = i
                    break

            if task is None:
                return {
                    "success": False,
                    "assigned": False,
                    "reason": f"Task {task_id} not found",
                }

            existing_assignee = task.get("assignee")

            if existing_assignee:
                return {
                    "success": True,
                    "assigned": False,
                    "reason": "Task already assigned",
                    "existing_assignee": existing_assignee,
                }

            # Assign task
            task["assignee"] = {
                "type": assignee_type,
                "name": assignee_name,
                "hostname": hostname,
                "assigned_at": datetime.utcnow().isoformat() + "Z",
                "assigned_by": "atomic_check_and_assign",
            }
            task["lastModified"] = datetime.utcnow().isoformat() + "Z"

            state["todos"][task_index] = task

            # Save
            if state_file.exists():
                backup_file = state_file.with_suffix('.json.bak')
                backup_file.write_text(state_file.read_text())

            with open(state_file, 'w') as f:
                json.dump(state, f, indent=2)

            return {
                "success": True,
                "assigned": True,
                "reason": "Task assigned successfully",
                "assignee": task["assignee"],
            }

    except TimeoutError:
        return {
            "success": False,
            "assigned": False,
            "reason": f"Lock timeout after {timeout}s",
        }
    except Exception as e:
        logger.error(f"Error in atomic_check_and_assign: {e}", exc_info=True)
        return {
            "success": False,
            "assigned": False,
            "reason": str(e),
        }


def atomic_batch_assign(
    task_ids: list[str],
    assignee_name: str,
    assignee_type: str = "agent",
    hostname: Optional[str] = None,
    timeout: float = 10.0,
) -> dict[str, Any]:
    """
    Atomically assign multiple tasks (all or nothing).

    Uses state file lock to ensure atomic batch assignment.

    Args:
        task_ids: List of task IDs to assign
        assignee_name: Name of assignee
        assignee_type: Type of assignee
        hostname: Optional hostname
        timeout: Lock timeout

    Returns:
        Dict with:
        - success: bool
        - assigned: list[str] (task IDs successfully assigned)
        - failed: list[dict] (task IDs that failed with reasons)
        - total: int (total tasks attempted)
    """
    try:
        with state_file_lock(timeout=timeout):
            project_root = find_project_root()
            state_file = project_root / ".todo2" / "state.todo2.json"

            if not state_file.exists():
                return {
                    "success": False,
                    "assigned": [],
                    "failed": [{"task_id": tid, "reason": "State file not found"} for tid in task_ids],
                    "total": len(task_ids),
                }

            with open(state_file) as f:
                state = json.load(f)

            assigned = []
            failed = []

            for task_id in task_ids:
                # Find task
                task = None
                task_index = -1
                for i, t in enumerate(state.get("todos", [])):
                    if t.get("id") == task_id:
                        task = t
                        task_index = i
                        break

                if task is None:
                    failed.append({"task_id": task_id, "reason": "Task not found"})
                    continue

                if task.get("assignee"):
                    existing = task.get("assignee", {})
                    existing_name = existing.get("name", "unknown")
                    failed.append({
                        "task_id": task_id,
                        "reason": f"Already assigned to {existing_name}",
                    })
                    continue

                # Assign
                task["assignee"] = {
                    "type": assignee_type,
                    "name": assignee_name,
                    "hostname": hostname,
                    "assigned_at": datetime.utcnow().isoformat() + "Z",
                    "assigned_by": "atomic_batch_assign",
                }
                task["lastModified"] = datetime.utcnow().isoformat() + "Z"
                state["todos"][task_index] = task
                assigned.append(task_id)

            # Save if any assignments made
            if assigned:
                if state_file.exists():
                    backup_file = state_file.with_suffix('.json.bak')
                    backup_file.write_text(state_file.read_text())

                with open(state_file, 'w') as f:
                    json.dump(state, f, indent=2)

            return {
                "success": len(failed) == 0,
                "assigned": assigned,
                "failed": failed,
                "total": len(task_ids),
            }

    except TimeoutError:
        return {
            "success": False,
            "assigned": [],
            "failed": [{"task_id": tid, "reason": f"Lock timeout"} for tid in task_ids],
            "total": len(task_ids),
        }
    except Exception as e:
        logger.error(f"Error in atomic_batch_assign: {e}", exc_info=True)
        return {
            "success": False,
            "assigned": [],
            "failed": [{"task_id": tid, "reason": str(e)} for tid in task_ids],
            "total": len(task_ids),
        }
