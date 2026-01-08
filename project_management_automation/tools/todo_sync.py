"""
MCP Tool Wrapper for Todo Sync

Wraps TodoSyncAutomation to expose as MCP tool.
"""

from typing import Optional
import json
import logging
import time
from pathlib import Path

logger = logging.getLogger(__name__)

# Import error handler at module level to avoid scoping issues
try:
    from ..error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
except ImportError:
    import sys
    from pathlib import Path
    server_dir = Path(__file__).parent.parent
    sys.path.insert(0, str(server_dir))
    try:
        from error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
    except ImportError:
        # Fallback: define minimal versions if import fails
        def format_success_response(data, message=None):
            return {"success": True, "data": data, "timestamp": time.time()}
        def format_error_response(error, error_code, include_traceback=False):
            return {"success": False, "error": {"code": str(error_code), "message": str(error)}}
        def log_automation_execution(name, duration, success, error=None):
            logger.info(f"{name}: {duration:.2f}s, success={success}")
        class ErrorCode:
            AUTOMATION_ERROR = "AUTOMATION_ERROR"


def sync_todo_tasks(
    dry_run: bool = False,
    output_path: Optional[str] = None
) -> str:
    """
    Synchronize tasks between shared TODO table and Todo2.

    Args:
        dry_run: Whether to simulate sync without making changes (default: False)
        output_path: Path for report output (default: docs/TODO_SYNC_REPORT.md)

    Returns:
        JSON string with sync results
    """
    start_time = time.time()

    try:
        # Import from package
        from project_management_automation.scripts.automate_todo_sync import TodoSyncAutomation
        from project_management_automation.utils import find_project_root

        # Find project root
        project_root = find_project_root()

        # Build config
        config = {
            'output_path': output_path or 'docs/TODO_SYNC_REPORT.md',
            'dry_run': dry_run
        }

        # Create sync automation and run
        sync_automation = TodoSyncAutomation(config, project_root)
        results = sync_automation.run()

        # Extract key metrics
        sync_results = results.get('results', {})
        matches = sync_results.get('matches', [])
        conflicts = sync_results.get('conflicts', [])
        new_shared = sync_results.get('new_shared_todos', [])
        new_todo2 = sync_results.get('new_todo2_tasks', [])
        updates = sync_results.get('updates', [])

        # Format response
        response_data = {
            'dry_run': dry_run,
            'matches_found': len(matches),
            'conflicts_detected': len(conflicts),
            'new_shared_todos': len(new_shared),
            'new_todo2_tasks': len(new_todo2),
            'updates_performed': len(updates),
            'report_path': str(Path(config['output_path']).absolute()),
            'status': results.get('status', 'unknown')
        }

        duration = time.time() - start_time
        log_automation_execution('sync_todo_tasks', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('sync_todo_tasks', duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)
