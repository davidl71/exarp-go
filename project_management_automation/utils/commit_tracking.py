"""
Git-inspired commit tracking for Todo2 tasks.

Tracks all task changes as commits, enabling:
- Complete audit trail
- Task version history
- Branch-based workflows
- Change analytics

Inspired by concepts from GitTask (https://github.com/Bengerthelorf/gittask)
Licensed under GPL-3.0. This implementation is original Python code.
See ATTRIBUTIONS.md for details.
"""

import json
import logging
import uuid
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

from .project_root import find_project_root
from .json_cache import JsonCacheManager

logger = logging.getLogger(__name__)

# Cache manager for commits
_cache_manager = JsonCacheManager.get_instance()


class TaskCommit:
    """Represents a single commit (change) to a task."""

    def __init__(
        self,
        commit_id: Optional[str] = None,
        task_id: str = "",
        message: str = "",
        old_state: Optional[dict[str, Any]] = None,
        new_state: Optional[dict[str, Any]] = None,
        timestamp: Optional[datetime] = None,
        author: str = "system",
        branch: Optional[str] = None,
    ):
        self.id = commit_id or str(uuid.uuid4())
        self.task_id = task_id
        self.message = message
        self.old_state = old_state or {}
        self.new_state = new_state or {}
        self.timestamp = timestamp or datetime.now()
        self.author = author
        self.branch = branch or "main"

    def to_dict(self) -> dict[str, Any]:
        """Convert commit to dictionary for storage."""
        return {
            "id": self.id,
            "task_id": self.task_id,
            "message": self.message,
            "old_state": self.old_state,
            "new_state": self.new_state,
            "timestamp": self.timestamp.isoformat(),
            "author": self.author,
            "branch": self.branch,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "TaskCommit":
        """Create commit from dictionary."""
        return cls(
            commit_id=data.get("id"),
            task_id=data.get("task_id", ""),
            message=data.get("message", ""),
            old_state=data.get("old_state", {}),
            new_state=data.get("new_state", {}),
            timestamp=datetime.fromisoformat(data["timestamp"]) if data.get("timestamp") else None,
            author=data.get("author", "system"),
            branch=data.get("branch", "main"),
        )


class CommitTracker:
    """Manages commit history for tasks."""

    def __init__(self, project_root: Optional[Path] = None):
        self.project_root = project_root or find_project_root()
        self.commits_file = self.project_root / ".todo2" / "commits.json"
        # Use unified JSON cache instead of module-level cache
        self._cache = _cache_manager.get_cache(self.commits_file, enable_stats=True)

    def _ensure_commits_file(self) -> None:
        """Ensure commits file exists with initial structure."""
        self.commits_file.parent.mkdir(parents=True, exist_ok=True)
        if not self.commits_file.exists():
            self.commits_file.write_text(json.dumps({"commits": [], "version": "1.0"}, indent=2))

    def _load_commits(self) -> list[TaskCommit]:
        """Load all commits from storage with caching."""
        self._ensure_commits_file()

        try:
            # Use unified JSON cache (automatically handles mtime invalidation)
            data = self._cache.get_or_load()
            commits_data = data.get("commits", [])
            commits = [TaskCommit.from_dict(c) for c in commits_data]
            return commits
        except Exception as e:
            logger.error(f"Error loading commits: {e}")
            return []

    def _save_commits(self, commits: list[TaskCommit]) -> None:
        """Save commits to storage and invalidate cache."""
        self._ensure_commits_file()

        try:
            commits_data = [c.to_dict() for c in commits]
            data = {"commits": commits_data, "version": "1.0", "last_updated": datetime.now().isoformat()}
            with open(self.commits_file, "w") as f:
                json.dump(data, f, indent=2)
            # Invalidate cache after save (next load will pick up new mtime)
            self._cache.invalidate()
        except Exception as e:
            logger.error(f"Error saving commits: {e}")
            raise

    def create_commit(
        self,
        task_id: str,
        message: str,
        old_state: Optional[dict[str, Any]] = None,
        new_state: Optional[dict[str, Any]] = None,
        author: str = "system",
        branch: Optional[str] = None,
    ) -> TaskCommit:
        """Create and store a new commit."""
        commit = TaskCommit(
            task_id=task_id,
            message=message,
            old_state=old_state or {},
            new_state=new_state or {},
            author=author,
            branch=branch or self._get_task_branch(task_id),
        )

        commits = self._load_commits()
        commits.append(commit)
        self._save_commits(commits)

        logger.debug(f"Created commit {commit.id[:8]} for task {task_id[:8]}: {message}")
        return commit

    def get_commits_for_task(self, task_id: str, branch: Optional[str] = None) -> list[TaskCommit]:
        """Get all commits for a specific task."""
        commits = self._load_commits()
        filtered = [c for c in commits if c.task_id == task_id]
        if branch:
            filtered = [c for c in filtered if c.branch == branch]
        # Sort by timestamp (oldest first)
        filtered.sort(key=lambda c: c.timestamp)
        return filtered

    def get_commits_for_branch(self, branch: str) -> list[TaskCommit]:
        """Get all commits for a specific branch."""
        commits = self._load_commits()
        filtered = [c for c in commits if c.branch == branch]
        filtered.sort(key=lambda c: c.timestamp)
        return filtered

    def get_latest_commit_for_task(self, task_id: str, branch: Optional[str] = None) -> Optional[TaskCommit]:
        """Get the most recent commit for a task."""
        commits = self.get_commits_for_task(task_id, branch)
        return commits[-1] if commits else None

    def get_task_state_at_commit(self, task_id: str, commit_id: str) -> Optional[dict[str, Any]]:
        """Get task state at a specific commit."""
        commits = self.get_commits_for_task(task_id)
        for commit in commits:
            if commit.id == commit_id:
                return commit.new_state
            if commit.new_state:  # Use new_state from the commit
                # Continue to find the target commit
                pass
        return None

    def get_task_state_at_time(self, task_id: str, timestamp: datetime) -> Optional[dict[str, Any]]:
        """Get task state at a specific point in time."""
        commits = self.get_commits_for_task(task_id)
        state = None
        for commit in commits:
            if commit.timestamp <= timestamp:
                state = commit.new_state
            else:
                break
        return state

    def _get_task_branch(self, task_id: str) -> str:
        """Extract branch from task metadata or return 'main'."""
        # Try to load task from Todo2 to get branch tag
        try:
            todo2_file = self.project_root / ".todo2" / "state.todo2.json"
            if todo2_file.exists():
                with open(todo2_file) as f:
                    data = json.load(f)
                    tasks = data.get("todos", [])
                    for task in tasks:
                        if task.get("id") == task_id:
                            tags = task.get("tags", [])
                            # Look for branch: prefix
                            for tag in tags:
                                if tag.startswith("branch:"):
                                    return tag.split(":", 1)[1]
            return "main"
        except Exception:
            return "main"

    def clear_cache(self) -> None:
        """Clear the commits cache (force reload)."""
        self._commits_cache = None


# Global commit tracker instance
_commit_tracker: Optional[CommitTracker] = None


def get_commit_tracker(project_root: Optional[Path] = None) -> CommitTracker:
    """Get or create the global commit tracker instance."""
    global _commit_tracker
    if _commit_tracker is None:
        _commit_tracker = CommitTracker(project_root)
    return _commit_tracker


def track_task_create(
    task_id: str,
    task_data: dict[str, Any],
    author: str = "system",
    branch: Optional[str] = None,
) -> TaskCommit:
    """Track task creation as a commit."""
    message = f"Create task: {task_data.get('name', task_id[:8])}"
    return get_commit_tracker().create_commit(
        task_id=task_id,
        message=message,
        old_state={},
        new_state=task_data,
        author=author,
        branch=branch,
    )


def track_task_update(
    task_id: str,
    old_state: dict[str, Any],
    new_state: dict[str, Any],
    author: str = "system",
    branch: Optional[str] = None,
) -> TaskCommit:
    """Track task update as a commit."""
    task_name = new_state.get("name", old_state.get("name", task_id[:8]))
    message = f"Update task: {task_name}"
    return get_commit_tracker().create_commit(
        task_id=task_id,
        message=message,
        old_state=old_state,
        new_state=new_state,
        author=author,
        branch=branch,
    )


def track_task_delete(
    task_id: str,
    old_state: dict[str, Any],
    author: str = "system",
    branch: Optional[str] = None,
) -> TaskCommit:
    """Track task deletion as a commit."""
    task_name = old_state.get("name", task_id[:8])
    message = f"Delete task: {task_name}"
    return get_commit_tracker().create_commit(
        task_id=task_id,
        message=message,
        old_state=old_state,
        new_state={},
        author=author,
        branch=branch,
    )


def track_task_status_change(
    task_id: str,
    old_status: str,
    new_status: str,
    task_data: dict[str, Any],
    author: str = "system",
    branch: Optional[str] = None,
) -> TaskCommit:
    """Track task status change as a commit."""
    task_name = task_data.get("name", task_id[:8])
    message = f"Change status: {task_name} ({old_status} → {new_status})"
    
    # Create old_state with old status
    old_state = task_data.copy()
    old_state["status"] = old_status
    
    # Create new_state with new status
    new_state = task_data.copy()
    new_state["status"] = new_status
    
    return get_commit_tracker().create_commit(
        task_id=task_id,
        message=message,
        old_state=old_state,
        new_state=new_state,
        author=author,
        branch=branch,
    )


__all__ = [
    "TaskCommit",
    "CommitTracker",
    "get_commit_tracker",
    "track_task_create",
    "track_task_update",
    "track_task_delete",
    "track_task_status_change",
]
