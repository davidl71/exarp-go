"""
MCP Tool Wrapper for External Tool Hints Automation

Wraps ExternalToolHintsAutomation to expose as MCP tool.
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


def add_external_tool_hints(
    dry_run: bool = False,
    output_path: Optional[str] = None,
    min_file_size: int = 50
) -> str:
    """
    [HINT: External tool hints automation. Returns files scanned, modified, hints added, report path.]

    Automatically detect where Context7/external tool hints should be added to documentation
    and insert them following the standard pattern.

    Args:
        dry_run: Preview changes without applying (default: False)
        output_path: Path for report output (default: docs/EXTERNAL_TOOL_HINTS_REPORT.md)
        min_file_size: Minimum file size in lines to process (default: 50)

    Returns:
        JSON string with analysis results
    """
    start_time = time.time()

    try:
        # Import from package
        from project_management_automation.scripts.automate_external_tool_hints import ExternalToolHintsAutomation
        from project_management_automation.utils import find_project_root

        # Find project root
        project_root = find_project_root()

        # Build config
        config = {
            'dry_run': dry_run,
            'output_path': output_path or 'docs/EXTERNAL_TOOL_HINTS_REPORT.md',
            'min_file_size': min_file_size
        }

        # Create analyzer and run
        analyzer = ExternalToolHintsAutomation(config, project_root)
        results = analyzer.run()

        # Extract key metrics - analysis results are nested in results['results']
        analysis_results = results.get('results', {})
        files_scanned = analysis_results.get('files_scanned', 0)
        files_modified = analysis_results.get('files_modified', 0)
        files_skipped = analysis_results.get('files_skipped', 0)
        hints_added = analysis_results.get('hints_added', [])
        hints_skipped = analysis_results.get('hints_skipped', [])

        # Format response
        response_data = {
            'files_scanned': files_scanned,
            'files_modified': files_modified,
            'files_skipped': files_skipped,
            'hints_added_count': len(hints_added),
            'hints_skipped_count': len(hints_skipped),
            'report_path': str(Path(config['output_path']).absolute()),
            'dry_run': dry_run,
            'status': results.get('status', 'unknown'),
            'hints_added': hints_added[:10],  # First 10 for preview
            'hints_skipped_sample': hints_skipped[:10]  # First 10 for preview
        }

        duration = time.time() - start_time
        log_automation_execution('add_external_tool_hints', duration, True)

        return json.dumps(
            format_success_response(response_data, "External tool hints automation completed"),
            indent=2
        )

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('add_external_tool_hints', duration, False, str(e))
        logger.error(f"Error in external tool hints automation: {e}", exc_info=True)

        return json.dumps(
            format_error_response(
                f"External tool hints automation failed: {str(e)}",
                ErrorCode.AUTOMATION_ERROR,
                include_traceback=True
            ),
            indent=2
        )
