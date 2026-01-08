"""
Todo2 utilities for project-scoped automation.

Encapsulates helpers that:
- infer the git-backed project ID (owner/repo) for Todo2 ownership metadata,
- filter task lists to the current project ID, and
- annotate tasks with ownership metadata.
- validate project ownership on startup
"""

import json
import logging
import subprocess
from collections.abc import Iterable
from pathlib import Path
from typing import Iterable, List, Optional, Tuple

# Remove circular import
# from . import find_project_root

logger = logging.getLogger(__name__)


def _normalize_git_remote(url: str) -> Optional[str]:
    """Normalize git remote to owner/repo format."""
    if not url:
        return None

    url = url.strip()
    if url.endswith(".git"):
        url = url[: -len(".git")]

    if url.startswith("git@"):
        # git@github.com:owner/repo
        parts = url.split(":", 1)
        if len(parts) == 2:
            return parts[1]
    elif "://" in url:
        # https://github.com/owner/repo
        parts = url.split("://", 1)[1].split("/")
        if len(parts) >= 2:
            return "/".join(parts[1:3])

    # Fallback for other formats
    if "/" in url:
        return "/".join(url.split("/")[-2:])
    return None


def get_repo_project_id(project_root: Optional[Path] = None) -> Optional[str]:
    """Return the git owner/repo identifier for the current project."""
    # Local import to avoid circular dependency
    from .project_root import find_project_root
    root = find_project_root(project_root)
    try:
        result = subprocess.run(
            ["git", "-C", str(root), "remote", "get-url", "origin"],
            capture_output=True,
            text=True,
            check=True,
        )
        return _normalize_git_remote(result.stdout)
    except subprocess.CalledProcessError as exc:
        logger.debug("Could not read git remote: %s", exc)
        return None
    except FileNotFoundError:
        logger.debug("Git executable not found while determining project ID")
        return None


def task_belongs_to_project(task: dict, project_id: Optional[str]) -> bool:
    """Return True if the task is owned by the given project_id."""
    if not project_id:
        return True

    task_project = task.get("project_id")
    if not task_project:
        return True

    return task_project == project_id


def filter_tasks_by_project(
    tasks: Iterable[dict],
    project_id: Optional[str],
    include_unassigned: bool = True,
    logger: Optional[logging.Logger] = None,
) -> List[dict]:
    """Return only the tasks that belong to the requested project."""
    filtered = []
    for task in tasks:
        task_project = task.get("project_id")
        if task_project:
            if project_id and task_project != project_id:
                if logger:
                    logger.debug("Skipping task %s owned by %s", task.get("id"), task_project)
                continue
        elif not include_unassigned and project_id:
            if logger:
                logger.debug("Skipping unassigned task %s", task.get("id"))
            continue
        filtered.append(task)
    return filtered


def annotate_task_project(task: dict, project_id: Optional[str]) -> dict:
    """Ensure the task has ownership metadata."""
    if project_id:
        task["project_id"] = project_id
    return task


def load_todo2_project_info(project_root: Optional[Path] = None) -> Optional[dict]:
    """
    Load project information from Todo2 state file.
    
    Args:
        project_root: Optional project root path
        
    Returns:
        Project dict with id, name, path, repository, added_at, or None if not found
    """
    from .project_root import find_project_root
    root = find_project_root(project_root)
    todo2_file = root / ".todo2" / "state.todo2.json"
    
    if not todo2_file.exists():
        return None
    
    try:
        with open(todo2_file) as f:
            state = json.load(f)
        return state.get("project")
    except Exception as e:
        logger.debug(f"Error loading Todo2 project info: {e}")
        return None


def validate_project_ownership(
    project_root: Optional[Path] = None,
    warn_only: bool = True
) -> Tuple[bool, Optional[str]]:
    """
    Validate that the current PROJECT_ROOT matches Todo2's project.path.
    
    Args:
        project_root: Optional project root path (defaults to find_project_root)
        warn_only: If True, only warn on mismatch; if False, raise error
        
    Returns:
        Tuple of (is_valid, error_message)
        - is_valid: True if validation passes or project info not found
        - error_message: None if valid, otherwise description of mismatch
    """
    from .project_root import find_project_root
    import os
    
    current_root = find_project_root(project_root)
    current_root_str = str(current_root.resolve())
    
    # Get PROJECT_ROOT env var if set
    env_root = os.getenv("PROJECT_ROOT") or os.getenv("WORKSPACE_PATH")
    if env_root:
        env_root_path = Path(env_root).resolve()
        if env_root_path.exists():
            current_root_str = str(env_root_path)
    
    # Load Todo2 project info
    project_info = load_todo2_project_info(current_root)
    if not project_info:
        logger.debug("No Todo2 project info found - skipping validation")
        return True, None
    
    todo2_path = project_info.get("path")
    if not todo2_path:
        logger.debug("Todo2 project info missing path field")
        return True, None
    
    # Normalize paths for comparison
    todo2_path_resolved = str(Path(todo2_path).resolve())
    
    if current_root_str != todo2_path_resolved:
        error_msg = (
            f"Project path mismatch detected!\n"
            f"  Current PROJECT_ROOT: {current_root_str}\n"
            f"  Todo2 project.path:   {todo2_path_resolved}\n"
            f"  Todo2 project.id:     {project_info.get('id', 'unknown')}"
        )
        
        if warn_only:
            logger.warning(error_msg)
            return True, error_msg
        else:
            logger.error(error_msg)
            return False, error_msg
    
    logger.debug(f"Project ownership validated: {project_info.get('id')} at {current_root_str}")
    return True, None


def get_current_project_id(project_root: Optional[Path] = None) -> Optional[str]:
    """
    Get the current project ID from Todo2 state or git remote.
    
    Tries Todo2 project.id first, falls back to git remote.
    
    Args:
        project_root: Optional project root path
        
    Returns:
        Project ID (owner/repo format) or None if not found
    """
    project_info = load_todo2_project_info(project_root)
    if project_info and project_info.get("id"):
        return project_info.get("id")
    
    # Fallback to git remote
    return get_repo_project_id(project_root)


# ═══════════════════════════════════════════════════════════════
# STATUS NORMALIZATION
# ═══════════════════════════════════════════════════════════════

def normalize_status(status: str) -> str:
    """
    Normalize task status to canonical lowercase form.
    
    Handles case-insensitive matching and variant status values.
    Maps common variants to canonical forms for consistent processing.
    
    Canonical statuses:
    - 'todo': Pending/not started
    - 'in_progress': Currently being worked on
    - 'review': Awaiting review/approval
    - 'completed': Finished (normalizes 'done' to 'completed')
    - 'blocked': Cannot proceed
    - 'cancelled': Cancelled/abandoned
    
    Args:
        status: Raw status value (case-insensitive, handles variants)
    
    Returns:
        Canonical lowercase status value (defaults to 'todo' if empty/invalid)
    
    Examples:
        >>> normalize_status('Todo')
        'todo'
        >>> normalize_status('DONE')
        'completed'
        >>> normalize_status('in-progress')
        'in_progress'
        >>> normalize_status('')
        'todo'
    """
    if not status:
        return 'todo'
    
    status_lower = status.lower().strip()
    
    # Map variants to canonical forms
    status_map = {
        # Pending/Todo variants
        'pending': 'todo',
        'not started': 'todo',
        'new': 'todo',
        
        # In Progress variants
        'in progress': 'in_progress',
        'in-progress': 'in_progress',
        'in_progress': 'in_progress',
        'working': 'in_progress',
        'active': 'in_progress',
        
        # Review variants
        'review': 'review',
        'needs review': 'review',
        'awaiting review': 'review',
        
        # Completed variants (normalize 'done' to 'completed')
        'completed': 'completed',
        'done': 'completed',  # Normalize 'done' to 'completed'
        'finished': 'completed',
        'closed': 'completed',
        
        # Blocked variants
        'blocked': 'blocked',
        'waiting': 'blocked',
        
        # Cancelled variants
        'cancelled': 'cancelled',
        'canceled': 'cancelled',  # US spelling
        'abandoned': 'cancelled',
    }
    
    return status_map.get(status_lower, status_lower)


def is_pending_status(status: str) -> bool:
    """
    Check if status represents a pending task.
    
    Args:
        status: Task status value
    
    Returns:
        True if status is 'todo' (normalized)
    """
    normalized = normalize_status(status)
    return normalized == 'todo'


def is_completed_status(status: str) -> bool:
    """
    Check if status represents a completed task.
    
    Args:
        status: Task status value
    
    Returns:
        True if status is 'completed' or 'cancelled' (normalized)
    """
    normalized = normalize_status(status)
    return normalized in ['completed', 'cancelled']


def is_active_status(status: str) -> bool:
    """
    Check if status represents an active (non-completed) task.
    
    Args:
        status: Task status value
    
    Returns:
        True if status is 'todo', 'in_progress', 'review', or 'blocked' (normalized)
    """
    normalized = normalize_status(status)
    return normalized in ['todo', 'in_progress', 'review', 'blocked']


def is_review_status(status: str) -> bool:
    """
    Check if status represents a task awaiting review.
    
    Args:
        status: Task status value
    
    Returns:
        True if status is 'review' (normalized)
    """
    normalized = normalize_status(status)
    return normalized == 'review'


def normalize_status_to_title_case(status: str) -> str:
    """
    Normalize task status to Title Case for storage/display.
    
    This ensures consistent capitalization in Todo2 files:
    - "Todo" (not "todo", "TODO", "pending")
    - "In Progress" (not "in_progress", "in-progress", "In Progress")
    - "Review" (not "review", "Review")
    - "Done" (not "done", "DONE", "completed")
    
    Args:
        status: Raw status value (case-insensitive, handles variants)
    
    Returns:
        Title Case status value (defaults to 'Todo' if empty/invalid)
    
    Examples:
        >>> normalize_status_to_title_case('todo')
        'Todo'
        >>> normalize_status_to_title_case('DONE')
        'Done'
        >>> normalize_status_to_title_case('in_progress')
        'In Progress'
        >>> normalize_status_to_title_case('')
        'Todo'
    """
    if not status:
        return 'Todo'
    
    # First normalize to canonical lowercase form
    normalized = normalize_status(status)
    
    # Map canonical lowercase to Title Case
    title_case_map = {
        'todo': 'Todo',
        'in_progress': 'In Progress',
        'review': 'Review',
        'completed': 'Done',  # Map 'completed' to 'Done' for consistency
        'done': 'Done',
        'blocked': 'Blocked',
        'cancelled': 'Cancelled',
    }
    
    return title_case_map.get(normalized, status.title())


__all__ = [
    "get_repo_project_id",
    "task_belongs_to_project",
    "filter_tasks_by_project",
    "annotate_task_project",
    "load_todo2_project_info",
    "validate_project_ownership",
    "get_current_project_id",
    # Status normalization
    "normalize_status",
    "normalize_status_to_title_case",
    "is_pending_status",
    "is_completed_status",
    "is_active_status",
    "is_review_status",
]

