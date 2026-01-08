"""
MCP Tool Wrapper for Todo2 Alignment Analysis

Wraps Todo2AlignmentAnalyzerV2 to expose as MCP tool.
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


def analyze_todo2_alignment(
    create_followup_tasks: bool = True,
    output_path: Optional[str] = None
) -> str:
    """
    Analyze task alignment with project goals, find misaligned tasks.

    Args:
        create_followup_tasks: Whether to create Todo2 tasks for misaligned tasks
        output_path: Path for report output (default: docs/TODO2_ALIGNMENT_REPORT.md)

    Returns:
        JSON string with analysis results
    """
    start_time = time.time()

    try:
        # Import from package
        from project_management_automation.scripts.automate_todo2_alignment_v2 import Todo2AlignmentAnalyzerV2
        from project_management_automation.utils import find_project_root

        # Find project root
        project_root = find_project_root()

        # Build config
        config = {
            'create_followup_tasks': create_followup_tasks,
            'output_path': output_path or 'docs/TODO2_ALIGNMENT_REPORT.md'
        }

        # Create analyzer and run
        analyzer = Todo2AlignmentAnalyzerV2(config, project_root)
        results = analyzer.run()

        # Extract key metrics from analysis results
        analysis_results = results.get('results', {})
        misaligned_tasks = analysis_results.get('misaligned_tasks', [])
        infrastructure_tasks = analysis_results.get('infrastructure_tasks', [])
        stale_tasks = analysis_results.get('stale_tasks', [])
        alignment_score = analysis_results.get('alignment_score', 0)

        # Format response with correct keys from analyzer
        response_data = {
            'total_tasks_analyzed': analysis_results.get('total_tasks', 0),
            'misaligned_count': len(misaligned_tasks),
            'infrastructure_count': len(infrastructure_tasks),
            'stale_count': len(stale_tasks),
            'average_alignment_score': alignment_score,
            'by_priority': analysis_results.get('by_priority', {}),
            'by_status': analysis_results.get('by_status', {}),
            'report_path': str(Path(config['output_path']).absolute()),
            'tasks_created': len(results.get('followup_tasks', [])) if create_followup_tasks else 0,
            'status': results.get('status', 'unknown')
        }

        duration = time.time() - start_time
        log_automation_execution('analyze_todo2_alignment', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('analyze_todo2_alignment', duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)
