"""
MCP Tool Wrapper for Documentation Health Check

Wraps DocumentationHealthAnalyzerV2 to expose as MCP tool.
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


def check_documentation_health(
    output_path: Optional[str] = None,
    create_tasks: bool = True
) -> str:
    """
    Analyze documentation structure, find broken references, identify issues.

    Args:
        output_path: Path for report output (default: docs/DOCUMENTATION_HEALTH_REPORT.md)
        create_tasks: Whether to create Todo2 tasks for issues found

    Returns:
        JSON string with analysis results
    """
    start_time = time.time()

    try:
        # Import from package
        from project_management_automation.scripts.automate_docs_health_v2 import DocumentationHealthAnalyzerV2

        # Find project root
        from project_management_automation.utils import find_project_root
        project_root = find_project_root()

        # Build config
        config = {
            'output_path': output_path or 'docs/DOCUMENTATION_HEALTH_REPORT.md',
            'create_tasks': create_tasks
        }

        # Create analyzer and run
        analyzer = DocumentationHealthAnalyzerV2(config, project_root)
        results = analyzer.run()

        # Extract key metrics
        health_score = results.get('results', {}).get('health_score', 0)
        link_validation = results.get('results', {}).get('link_validation', {})
        format_validation = results.get('results', {}).get('format_validation', {})

        # Format response
        response_data = {
            'health_score': health_score,
            'report_path': str(Path(config['output_path']).absolute()),
            'link_validation': {
                'total_links': link_validation.get('total_links', 0),
                'broken_internal': len(link_validation.get('broken_internal', [])),
                'broken_external': len(link_validation.get('broken_external', []))
            },
            'format_errors': len(format_validation.get('format_errors', [])),
            'tasks_created': len(results.get('followup_tasks', [])) if create_tasks else 0,
            'status': results.get('status', 'unknown')
        }

        duration = time.time() - start_time
        log_automation_execution('check_documentation_health', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('check_documentation_health', duration, False, e)

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)
