"""
Task Diff Tools

Compare different versions of tasks.
"""

import logging
from typing import Dict, Any, Optional
from datetime import datetime

from ..utils.commit_tracking import get_commit_tracker

logger = logging.getLogger(__name__)


def compare_task_versions(
    task_id: str,
    version1_commit_id: Optional[str] = None,
    version2_commit_id: Optional[str] = None,
    version1_time: Optional[datetime] = None,
    version2_time: Optional[datetime] = None,
) -> Dict[str, Any]:
    """
    Compare two versions of a task.
    
    Args:
        task_id: Task ID to compare
        version1_commit_id: First version commit ID
        version2_commit_id: Second version commit ID
        version1_time: First version timestamp
        version2_time: Second version timestamp
    
    Returns dictionary with diff results.
    """
    try:
        tracker = get_commit_tracker()
        
        # Get commits for the task
        commits = tracker.get_commits_for_task(task_id)
        commits.sort(key=lambda c: c.timestamp)
        
        # Find versions to compare
        version1 = None
        version2 = None
        
        if version1_commit_id:
            version1 = next((c for c in commits if c.id == version1_commit_id), None)
        elif version1_time:
            # Find closest commit to timestamp
            version1 = min(commits, key=lambda c: abs((c.timestamp - version1_time).total_seconds()))
        
        if version2_commit_id:
            version2 = next((c for c in commits if c.id == version2_commit_id), None)
        elif version2_time:
            version2 = min(commits, key=lambda c: abs((c.timestamp - version2_time).total_seconds()))
        
        # Default to first and last if not specified
        if not version1 and commits:
            version1 = commits[0]
        if not version2 and commits:
            version2 = commits[-1]
        
        if not version1 or not version2:
            return {
                "error": "Could not find versions to compare",
                "task_id": task_id
            }
        
        # Extract task states from commits (if stored)
        # For now, return basic comparison
        diff = {
            "task_id": task_id,
            "version1": {
                "commit_id": version1.id,
                "timestamp": version1.timestamp.isoformat(),
                "message": version1.message,
                "author": version1.author,
            },
            "version2": {
                "commit_id": version2.id,
                "timestamp": version2.timestamp.isoformat(),
                "message": version2.message,
                "author": version2.author,
            },
            "changes": _compute_diff(version1, version2),
        }
        
        return diff
    except Exception as e:
        logger.error(f"Error comparing task versions: {e}", exc_info=True)
        return {
            "error": str(e),
            "task_id": task_id
        }


def _compute_diff(commit1, commit2) -> Dict[str, Any]:
    """Compute differences between two commits."""
    changes = []
    
    if commit1.message != commit2.message:
        changes.append({
            "field": "message",
            "old": commit1.message,
            "new": commit2.message
        })
    
    if commit1.branch != commit2.branch:
        changes.append({
            "field": "branch",
            "old": commit1.branch,
            "new": commit2.branch
        })
    
    if commit1.author != commit2.author:
        changes.append({
            "field": "author",
            "old": commit1.author,
            "new": commit2.author
        })
    
    return {
        "change_count": len(changes),
        "changes": changes
    }


__all__ = ["compare_task_versions"]

