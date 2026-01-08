"""
MCP Tool Wrapper for Test Coverage Analysis

Wraps TestCoverageAnalyzer to expose as MCP tool.
"""

from typing import Optional
import json
import logging
import time
from pathlib import Path

logger = logging.getLogger(__name__)

# Import error handler
try:
    from ..error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
except ImportError:
    import sys
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


def analyze_test_coverage(
    coverage_file: Optional[str] = None,
    min_coverage: int = 80,
    output_path: Optional[str] = None,
    format: str = "html"
) -> str:
    """
    Generate coverage reports and identify gaps.

    Args:
        coverage_file: Path to coverage file (default: auto-detect)
        min_coverage: Minimum coverage threshold (default: 80)
        output_path: Path for coverage report (default: coverage-report/)
        format: Report format: html, json, or terminal (default: html)

    Returns:
        JSON string with coverage analysis results
    """
    start_time = time.time()

    try:
        from project_management_automation.scripts.automate_test_coverage import TestCoverageAnalyzer
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()

        config = {
            'coverage_file': coverage_file,
            'min_coverage': min_coverage,
            'output_path': output_path or 'coverage-report/',
            'format': format
        }

        analyzer = TestCoverageAnalyzer(config, project_root)
        results = analyzer.run()

        # Format response
        coverage_data = results.get('results', {}).get('coverage_data', {})
        response_data = {
            'total_coverage': coverage_data.get('total_coverage', 0),
            'total_lines': coverage_data.get('total_lines', 0),
            'covered_lines': coverage_data.get('covered_lines', 0),
            'min_coverage': min_coverage,
            'meets_threshold': results.get('results', {}).get('meets_threshold', False),
            'report_path': results.get('results', {}).get('report_path'),
            'gaps_count': len(results.get('results', {}).get('gaps', [])),
            'gaps': results.get('results', {}).get('gaps', [])[:10]  # Top 10 gaps
        }

        duration = time.time() - start_time
        log_automation_execution('analyze_test_coverage', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('analyze_test_coverage', duration, False, e)

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

