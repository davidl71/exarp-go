#!/usr/bin/env python3
"""
Utility script to update Todo2 tasks using proper MCP tools.

This script uses the proper abstraction layer instead of direct JSON editing.
"""

import sys
from pathlib import Path

# Add project_management_automation to path
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))

from project_management_automation.utils.todo2_mcp_client import (
    update_todos_mcp,
    add_comments_mcp,
)


def update_task_status(task_ids, new_status="Done", old_status="Todo"):
    """
    Update task status using proper MCP tools.
    
    Args:
        task_ids: List of task IDs to update
        new_status: New status (default: "Done")
        old_status: Old status to filter by (default: "Todo")
    """
    updates = [{"id": task_id, "status": new_status} for task_id in task_ids]
    
    success = update_todos_mcp(updates=updates, project_root=project_root)
    
    if success:
        print(f"✅ Successfully updated {len(task_ids)} tasks to {new_status}")
        return True
    else:
        print(f"❌ Failed to update tasks")
        return False


def add_result_comment(task_id, content):
    """
    Add a result comment to a task using proper MCP tools.
    
    Args:
        task_id: Task ID to add comment to
        content: Comment content
    """
    comment = {
        "type": "result",
        "content": content,
    }
    
    success = add_comments_mcp(todo_id=task_id, comments=[comment], project_root=project_root)
    
    if success:
        print(f"✅ Successfully added result comment to {task_id}")
        return True
    else:
        print(f"❌ Failed to add comment to {task_id}")
        return False


if __name__ == "__main__":
    import json
    
    if len(sys.argv) < 2:
        print("Usage: python scripts/update_todo2_tasks.py <command> [args...]")
        print("\nCommands:")
        print("  status <task_id1> [task_id2...] [--status=Done] [--old-status=Todo]")
        print("  comment <task_id> <comment_file>")
        sys.exit(1)
    
    command = sys.argv[1]
    
    if command == "status":
        task_ids = [arg for arg in sys.argv[2:] if not arg.startswith("--")]
        new_status = "Done"
        old_status = "Todo"
        
        for arg in sys.argv[2:]:
            if arg.startswith("--status="):
                new_status = arg.split("=", 1)[1]
            elif arg.startswith("--old-status="):
                old_status = arg.split("=", 1)[1]
        
        update_task_status(task_ids, new_status, old_status)
    
    elif command == "comment":
        if len(sys.argv) < 4:
            print("Usage: python scripts/update_todo2_tasks.py comment <task_id> <comment_file>")
            sys.exit(1)
        
        task_id = sys.argv[2]
        comment_file = Path(sys.argv[3])
        
        if not comment_file.exists():
            print(f"❌ Comment file not found: {comment_file}")
            sys.exit(1)
        
        content = comment_file.read_text()
        add_result_comment(task_id, content)
    
    else:
        print(f"Unknown command: {command}")
        sys.exit(1)

