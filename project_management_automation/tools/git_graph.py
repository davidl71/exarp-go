"""
Git Graph Visualization

Generates commit graph visualizations for task management.
"""

import json
import logging
from typing import Optional
from pathlib import Path
from datetime import datetime

from ..utils import find_project_root
from ..utils.commit_tracking import get_commit_tracker

logger = logging.getLogger(__name__)


def generate_commit_graph(
    branch: Optional[str] = None,
    task_id: Optional[str] = None,
    format: str = "text",
    output_path: Optional[Path] = None,
    max_commits: int = 50,
) -> str:
    """
    Generate commit graph visualization.
    
    Args:
        branch: Filter by branch
        task_id: Filter by task ID
        format: Output format ("text" or "dot")
        output_path: Optional path to save graph
        max_commits: Maximum commits to include
    
    Returns graph string in requested format.
    """
    try:
        tracker = get_commit_tracker()
        
        if task_id:
            commits = tracker.get_commits_for_task(task_id, branch)
        elif branch:
            commits = tracker.get_commits_for_branch(branch)
        else:
            commits = tracker.get_all_commits()
        
        # Sort by timestamp (oldest first for graph)
        commits.sort(key=lambda c: c.timestamp)
        
        # Limit results
        commits = commits[:max_commits]
        
        if format == "dot":
            graph = _generate_dot_graph(commits)
        else:
            graph = _generate_text_graph(commits)
        
        # Save to file if requested
        if output_path:
            output_path.write_text(graph)
        
        return graph
    except Exception as e:
        logger.error(f"Error generating commit graph: {e}", exc_info=True)
        return f"Error generating graph: {e}"


def _generate_text_graph(commits) -> str:
    """Generate text-based commit graph."""
    lines = []
    lines.append("Commit Graph")
    lines.append("=" * 50)
    
    for commit in commits:
        timestamp = commit.timestamp.strftime("%Y-%m-%d %H:%M:%S")
        lines.append(f"* {commit.id[:8]} - {commit.message[:50]}")
        lines.append(f"  Branch: {commit.branch}, Author: {commit.author}")
        lines.append(f"  Time: {timestamp}")
        if commit.task_id:
            lines.append(f"  Task: {commit.task_id}")
        lines.append("")
    
    return "\n".join(lines)


def _generate_dot_graph(commits) -> str:
    """Generate Graphviz DOT format graph."""
    lines = []
    lines.append("digraph commit_graph {")
    lines.append("  rankdir=LR;")
    lines.append("  node [shape=box];")
    
    # Create nodes
    for i, commit in enumerate(commits):
        node_id = f"c{i}"
        label = f"{commit.id[:8]}\\n{commit.message[:30]}"
        lines.append(f'  {node_id} [label="{label}"];')
    
    # Create edges (simple linear for now)
    for i in range(len(commits) - 1):
        lines.append(f"  c{i} -> c{i+1};")
    
    lines.append("}")
    return "\n".join(lines)


__all__ = ["generate_commit_graph"]

