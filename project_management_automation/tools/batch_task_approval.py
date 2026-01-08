"""
Batch Task Approval Tool

MCP Tool wrapper for batch approving TODO2 tasks using the batch update script.
"""

import subprocess
import sys
from typing import Any, List, Optional

from ..utils import find_project_root


def batch_approve_tasks(
    status: str = "Review",
    new_status: str = "Todo",
    clarification_none: bool = True,
    filter_tag: Optional[str] = None,
    task_ids: Optional[List[str]] = None,
    dry_run: bool = False
) -> dict[str, Any]:
    """
    Batch approve TODO2 tasks using the batch update script.

    Args:
        status: Current status to filter (default: "Review")
        new_status: New status after approval (default: "Todo")
        clarification_none: Only approve tasks with no clarification needed (default: True)
        filter_tag: Filter by tag (optional)
        task_ids: List of specific task IDs to approve (optional)
        dry_run: If True, don't actually approve, just report (default: False)

    Returns:
        Dictionary with approval results including count, task IDs, and status
    """
    project_root = find_project_root()
    
    # Try Todo2 MCP tools directly (fallback from native Go or when script missing)
    try:
        from ..utils.todo2_mcp_client import list_todos_mcp, update_todos_mcp
        
        # Get tasks filtered by status (MCP handles status normalization)
        todos = list_todos_mcp(status=status, project_root=project_root) or []
        
        # Filter tasks by additional criteria
        candidates = []
        for todo in todos:
            # list_todos_mcp already filtered by status, but verify
            todo_status = todo.get("status", "")
            if todo_status.title() != status.title():
                continue
            
            # Filter by specific task IDs if provided
            if task_ids:
                if todo.get("id") not in task_ids:
                    continue
            
            # Filter by tag if provided
            if filter_tag:
                todo_tags = todo.get("tags", [])
                if filter_tag not in todo_tags:
                    continue
            
            # Filter by clarification requirement if needed
            if clarification_none:
                long_desc = todo.get("long_description", "")
                if not long_desc or len(long_desc) < 50:
                    continue
            
            candidates.append(todo)
        
        if dry_run:
            task_list = [{"id": t["id"], "content": t.get("name", ""), "status": t.get("status", "")} for t in candidates]
            return {
                "success": True,
                "method": "todo2_mcp",
                "dry_run": True,
                "approved_count": len(candidates),
                "task_ids": [t["id"] for t in candidates],
                "tasks": task_list
            }
        
        # Update tasks via MCP
        updates = [{"id": t["id"], "status": new_status} for t in candidates]
        success = update_todos_mcp(updates=updates, project_root=project_root)
        
        if success:
            return {
                "success": True,
                "method": "todo2_mcp",
                "approved_count": len(candidates),
                "task_ids": [t["id"] for t in candidates]
            }
    except Exception as e:
        # If MCP fails, try batch script (legacy fallback)
        pass
    
    # Legacy: Try batch script if MCP failed
    batch_script = project_root / "scripts" / "batch_update_todos.py"
    if not batch_script.exists():
        return {
            "success": False,
            "error": f"Batch script not found: {batch_script}. Todo2 MCP fallback also unavailable.",
            "approved_count": 0,
            "task_ids": []
        }

    # Build command
    cmd = [
        sys.executable,
        str(batch_script),
        'approve',
        '--status', status,
        '--new-status', new_status
    ]

    if clarification_none:
        cmd.append('--clarification-none')

    if filter_tag:
        cmd.extend(['--filter-tag', filter_tag])

    if task_ids:
        cmd.extend(['--task-ids', ','.join(task_ids)])

    if dry_run:
        # For dry run, use list command instead
        cmd = [
            sys.executable,
            str(batch_script),
            'list',
            '--status', status
        ]
        if clarification_none:
            cmd.append('--clarification-none')
        if filter_tag:
            cmd.extend(['--filter-tag', filter_tag])
        if task_ids:
            cmd.extend(['--task-ids', ','.join(task_ids)])
    else:
        cmd.append('--yes')  # Skip confirmation

    try:
        result = subprocess.run(
            cmd,
            cwd=str(project_root),
            capture_output=True,
            text=True,
            timeout=30
        )

        if result.returncode != 0:
            return {
                "success": False,
                "error": f"Script execution failed: {result.stderr}",
                "approved_count": 0,
                "task_ids": []
            }

        # Parse output
        output = result.stdout
        approved_count = 0
        task_ids_approved = []

        if dry_run:
            # Count tasks that would be approved
            lines = output.split('\n')
            for line in lines:
                if line.strip().startswith('•'):
                    # Extract task ID
                    parts = line.split(':')
                    if len(parts) > 0:
                        task_id = parts[0].strip().replace('•', '').strip()
                        if task_id:
                            task_ids_approved.append(task_id)
                            approved_count += 1
        else:
            # Extract approved count from output
            import re
            match = re.search(r'Approved (\d+) tasks', output)
            if match:
                approved_count = int(match.group(1))

            # Try to extract task IDs from output
            lines = output.split('\n')
            for line in lines:
                if line.strip().startswith('•'):
                    parts = line.split(':')
                    if len(parts) > 0:
                        task_id = parts[0].strip().replace('•', '').strip()
                        if task_id:
                            task_ids_approved.append(task_id)

        return {
            "success": True,
            "approved_count": approved_count,
            "task_ids": task_ids_approved,
            "status_from": status,
            "status_to": new_status,
            "dry_run": dry_run,
            "output": output
        }

    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "error": "Script execution timed out",
            "approved_count": 0,
            "task_ids": []
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
            "approved_count": 0,
            "task_ids": []
        }
