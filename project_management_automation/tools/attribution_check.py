"""
MCP Tool Wrapper for Attribution Compliance Check

Scans codebase to verify proper attribution for all third-party components,
concepts, external services, and dependencies.

Inspired by concepts from GitTask (https://github.com/Bengerthelorf/gittask)
Licensed under GPL-3.0. This implementation is original Python code.
See ATTRIBUTIONS.md for details.
"""

from typing import Optional
import json
import logging
import time
from pathlib import Path

logger = logging.getLogger(__name__)

# Import error handler at module level
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
        def format_success_response(data, message=None):
            return {"success": True, "data": data, "timestamp": time.time()}
        def format_error_response(error, error_code, include_traceback=False):
            return {"success": False, "error": {"code": str(error_code), "message": str(error)}}
        def log_automation_execution(name, duration, success, error=None):
            logger.info(f"{name}: {duration:.2f}s, success={success}")
        class ErrorCode:
            AUTOMATION_ERROR = "AUTOMATION_ERROR"


def check_attribution_compliance(
    output_path: Optional[str] = None,
    create_tasks: bool = True,
) -> str:
    """
    Check attribution compliance across the codebase.

    Scans for:
    - Missing attribution in file headers
    - Missing entries in ATTRIBUTIONS.md
    - Uncredited third-party references
    - Dependency license compliance

    Args:
        output_path: Path for report output (default: docs/ATTRIBUTION_COMPLIANCE_REPORT.md)
        create_tasks: Whether to create Todo2 tasks for issues found

    Returns:
        JSON string with compliance check results
    """
    start_time = time.time()

    try:
        # Import from automation script
        from project_management_automation.scripts.automate_attribution_check import AttributionComplianceChecker

        # Find project root
        from project_management_automation.utils import find_project_root
        project_root = find_project_root()

        # Build config
        config = {
            'output_path': output_path or 'docs/ATTRIBUTION_COMPLIANCE_REPORT.md',
            'create_tasks': create_tasks
        }

        # Create checker and run
        checker = AttributionComplianceChecker(config, project_root)
        results = checker.run()

        # Extract key metrics from results
        analysis_results = results.get('results', {})
        attribution_score = analysis_results.get('attribution_score', 0)
        status = analysis_results.get('status', 'unknown')
        
        # Write report to file if generated
        report_path = Path(config['output_path'])
        if 'report' in results:
            report_path.parent.mkdir(parents=True, exist_ok=True)
            report_path.write_text(results['report'], encoding='utf-8')
            logger.info(f"Report written to {report_path}")

        # Format response
        response_data = {
            'attribution_score': attribution_score,
            'status': status,
            'compliant_files': len(analysis_results.get('compliant_files', [])),
            'issues_found': len(analysis_results.get('issues', [])),
            'warnings': len(analysis_results.get('warnings', [])),
            'missing_attribution': len(analysis_results.get('missing_attribution', [])),
            'tasks_created': len(results.get('followup_tasks', [])) if create_tasks else 0,
            'report_path': str(report_path.absolute()),
        }

        duration = time.time() - start_time
        log_automation_execution('check_attribution_compliance', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('check_attribution_compliance', duration, False, e)

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)
