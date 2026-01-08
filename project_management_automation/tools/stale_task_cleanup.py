"""
MCP Tool Wrapper for Stale Task Cleanup

Wraps StaleTaskCleanupAutomation to expose as MCP tool.
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


def cleanup_stale_tasks(
    stale_threshold_hours: float = 2.0,
    dry_run: bool = False,
    output_path: Optional[str] = None
) -> str:
    """
    [HINT: Stale task cleanup. Moves In Progress tasks back to Todo if no update in threshold hours.]

    Automatically moves tasks that are "In Progress" but haven't been updated
    recently back to "Todo" status. This ensures accurate time tracking by only
    counting actual work time, not idle time.

    📊 Output: Number of tasks moved, list of moved tasks, active tasks remaining
    🔧 Side Effects: Updates task status from In Progress to Todo

    Args:
        stale_threshold_hours: Hours of inactivity before task is considered stale (default: 2.0)
        dry_run: Preview changes without applying (default: False)
        output_path: Path for report output (optional)

    Returns:
        JSON string with cleanup results
    """
    start_time = time.time()

    try:
        # Import from package
        from project_management_automation.scripts.automate_stale_task_cleanup import StaleTaskCleanupAutomation
        from project_management_automation.utils import find_project_root

        # Find project root
        project_root = find_project_root()

        # Build config
        config = {
            'stale_threshold_hours': stale_threshold_hours,
            'dry_run': dry_run,
            'output_path': output_path
        }

        # Create cleanup automation and run
        cleanup = StaleTaskCleanupAutomation(config, project_root)
        results = cleanup.run()

        # Format response
        response_data = {
            'stale_threshold_hours': stale_threshold_hours,
            'stale_tasks_found': results.get('stale_tasks_found', 0),
            'tasks_moved': results.get('tasks_moved', 0),
            'active_tasks': results.get('active_tasks', 0),
            'moved_tasks': results.get('moved_tasks', []),
            'active_tasks_list': results.get('active_tasks_list', []),
            'dry_run': dry_run,
            'status': results.get('status', 'unknown')
        }

        if output_path:
            response_data['report_path'] = str(Path(output_path).absolute())

        duration = time.time() - start_time
        log_automation_execution('cleanup_stale_tasks', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('cleanup_stale_tasks', duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

