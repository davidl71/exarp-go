"""
MCP Tool Wrapper for Duplicate Task Detection

Wraps Todo2DuplicateDetector to expose as MCP tool.

Memory Integration:
- Saves duplicate resolution decisions for consistency
"""

import json
import logging
import time
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)


def _save_duplicate_detection_memory(response_data: dict[str, Any], auto_fix: bool) -> dict[str, Any]:
    """Save duplicate detection results as memory for future reference."""
    try:
        from .session_memory import save_session_insight

        content = f"""Duplicate task detection completed.

## Results
- Total tasks scanned: {response_data.get('total_tasks', 0)}
- Duplicate IDs found: {response_data.get('duplicate_ids', 0)}
- Exact name matches: {response_data.get('exact_name_matches', 0)}
- Similar name matches: {response_data.get('similar_name_matches', 0)}
- Similar description matches: {response_data.get('similar_description_matches', 0)}
- Self dependencies: {response_data.get('self_dependencies', 0)}

## Actions Taken {'(auto_fix applied)' if auto_fix else '(report only)'}
- Tasks removed: {response_data.get('tasks_removed', 0)}
- Tasks merged: {response_data.get('tasks_merged', 0)}
- Dependencies updated: {response_data.get('dependencies_updated', 0)}

## Report
{response_data.get('report_path', 'N/A')}
"""

        return save_session_insight(
            title=f"Duplicates: {response_data.get('total_duplicates_found', 0)} found",
            content=content,
            category="insight",
            metadata={"type": "duplicate_detection", "auto_fix": auto_fix}
        )
    except ImportError:
        logger.debug("Session memory not available for saving duplicate detection")
        return {"success": False, "error": "Memory system not available"}

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


def detect_duplicate_tasks(
    similarity_threshold: float = 0.85,
    auto_fix: bool = False,
    output_path: Optional[str] = None
) -> str:
    """
    Find and consolidate duplicate Todo2 tasks.

    Args:
        similarity_threshold: Similarity threshold for duplicate detection (0.0-1.0)
        auto_fix: Whether to automatically fix duplicates (default: False)
        output_path: Path for report output (default: docs/TODO2_DUPLICATE_DETECTION_REPORT.md)

    Returns:
        JSON string with detection results
    """
    start_time = time.time()

    try:
        # Import from package
        from project_management_automation.scripts.automate_todo2_duplicate_detection import Todo2DuplicateDetector
        from project_management_automation.utils import find_project_root

        # Find project root
        project_root = find_project_root()

        # Build config
        config = {
            'similarity_threshold': similarity_threshold,
            'auto_fix': auto_fix,
            'output_path': output_path or 'docs/TODO2_DUPLICATE_DETECTION_REPORT.md'
        }

        # Create detector and run
        detector = Todo2DuplicateDetector(config, project_root)
        results = detector.run()

        # Extract duplicates from detector instance (they're stored there, not in results)
        # The base class doesn't include duplicates in the returned structure
        duplicates = detector.duplicates

        # Get total_tasks from analysis results by re-running analysis
        # (or we can access it from the detector's _execute_analysis result)
        # For now, load it directly from Todo2 state file
        todo2_path = project_root / '.todo2' / 'state.todo2.json'
        total_tasks = 0
        if todo2_path.exists():
            try:
                with open(todo2_path) as f:
                    data = json.load(f)
                    total_tasks = len(data.get('todos', []))
            except Exception:
                pass

        # Extract auto-fix results from results dict (stored by _execute_analysis)
        auto_fix_applied = results.get('auto_fix_applied', auto_fix)
        tasks_removed = results.get('tasks_removed', 0)
        tasks_merged = results.get('tasks_merged', 0)
        dependencies_updated = results.get('dependencies_updated', 0)

        # Format response
        response_data = {
            'total_tasks': total_tasks,
            'duplicate_ids': len(duplicates.get('duplicate_ids', [])),
            'exact_name_matches': len(duplicates.get('exact_name_matches', [])),
            'similar_name_matches': len(duplicates.get('similar_name_matches', [])),
            'similar_description_matches': len(duplicates.get('similar_description_matches', [])),
            'self_dependencies': len(duplicates.get('self_dependencies', [])),
            'total_duplicates_found': (
                len(duplicates.get('duplicate_ids', [])) +
                len(duplicates.get('exact_name_matches', [])) +
                len(duplicates.get('similar_name_matches', [])) +
                len(duplicates.get('similar_description_matches', [])) +
                len(duplicates.get('self_dependencies', []))
            ),
            'report_path': str((project_root / config['output_path']).absolute()),
            'auto_fix_applied': auto_fix_applied,
            'tasks_removed': tasks_removed,
            'tasks_merged': tasks_merged,
            'dependencies_updated': dependencies_updated,
            'status': results.get('status', 'unknown')
        }

        duration = time.time() - start_time
        log_automation_execution('detect_duplicate_tasks', duration, True)

        # ═══ MEMORY INTEGRATION: Save detection results ═══
        if response_data.get('total_duplicates_found', 0) > 0:
            memory_result = _save_duplicate_detection_memory(response_data, auto_fix)
            if memory_result.get('success'):
                response_data['memory_saved'] = memory_result.get('memory_id')

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('detect_duplicate_tasks', duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)
