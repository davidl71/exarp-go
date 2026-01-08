"""
MCP Tool Wrapper for Automation Opportunity Finder

Wraps AutomationOpportunityFinder to expose as MCP tool.
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


def find_automation_opportunities(
    min_value_score: float = 0.7,
    output_path: Optional[str] = None
) -> str:
    """
    Discover new automation opportunities in the codebase.

    Args:
        min_value_score: Minimum value score threshold (0.0-1.0, default: 0.7)
        output_path: Path for report output (default: docs/AUTOMATION_OPPORTUNITIES_REPORT.md)

    Returns:
        JSON string with opportunity analysis results
    """
    start_time = time.time()

    try:
        # Import from package
        from project_management_automation.scripts.automate_automation_opportunities import AutomationOpportunityFinder
        from project_management_automation.utils import find_project_root

        # Find project root
        project_root = find_project_root()

        # Build config
        config = {
            'output_path': output_path or 'docs/AUTOMATION_OPPORTUNITIES_REPORT.md',
            'min_value_score': min_value_score
        }

        # Create finder and run
        finder = AutomationOpportunityFinder(config, project_root)
        results = finder.run()

        # Extract key metrics
        analysis_results = results.get('results', {})
        opportunities = analysis_results.get('opportunities', [])
        high_priority = analysis_results.get('high_priority', [])
        medium_priority = analysis_results.get('medium_priority', [])

        # Filter by min_value_score
        filtered_opportunities = [
            o for o in opportunities
            if o.get('score', 0) >= (min_value_score * 10)  # Score is 0-10, threshold is 0-1
        ]

        # Format response
        response_data = {
            'total_opportunities': len(opportunities),
            'filtered_opportunities': len(filtered_opportunities),
            'high_priority_count': len(high_priority),
            'medium_priority_count': len(medium_priority),
            'low_priority_count': len(analysis_results.get('low_priority', [])),
            'existing_automations': analysis_results.get('existing_automations', 0),
            'report_path': str(Path(config['output_path']).absolute()),
            'top_opportunities': filtered_opportunities[:10],  # Top 10
            'status': results.get('status', 'unknown')
        }

        duration = time.time() - start_time
        log_automation_execution('find_automation_opportunities', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('find_automation_opportunities', duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)
