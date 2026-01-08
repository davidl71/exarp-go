"""
MCP Tool Wrapper for Test Runner

Wraps TestRunner to expose as MCP tool.

Memory Integration:
- Saves test failures for debugging patterns

FastMCP Integration:
- Progress reporting via ctx.report_progress()
- Client logging via ctx.info()
"""

import asyncio
import json
import logging
import time
from pathlib import Path
from typing import Any, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from mcp.server.fastmcp import Context

logger = logging.getLogger(__name__)


# ═══════════════════════════════════════════════════════════════════════════════
# PROGRESS REPORTING HELPERS
# ═══════════════════════════════════════════════════════════════════════════════

async def _report_progress(ctx: Optional["Context"], progress: float, total: float, message: str) -> None:
    """Report progress if context is available."""
    if ctx is None:
        return
    try:
        from ..context_helpers import report_progress
        await report_progress(ctx, progress, total, message)
    except Exception:
        pass  # Progress reporting is optional


async def _log_info(ctx: Optional["Context"], message: str) -> None:
    """Log info to client if context is available."""
    if ctx is None:
        return
    try:
        from ..context_helpers import log_info
        await log_info(ctx, message)
    except Exception:
        pass  # Logging is optional


def _save_test_run_memory(response_data: dict[str, Any]) -> dict[str, Any]:
    """Save test run results as memory for debugging patterns."""
    try:
        from .session_memory import save_session_insight

        passed = response_data.get('tests_passed', 0)
        failed = response_data.get('tests_failed', 0)
        skipped = response_data.get('tests_skipped', 0)
        total = passed + failed + skipped

        # Only save if there were failures or it's a significant run
        if failed == 0 and total < 10:
            return {"success": True, "skipped": "no_failures_small_run"}

        content = f"""Test run completed.

## Results
- Total: {total}
- Passed: {passed} ✅
- Failed: {failed} {'❌' if failed > 0 else ''}
- Skipped: {skipped}
- Duration: {response_data.get('duration', 0):.2f}s

## Framework
{response_data.get('framework', 'unknown')}

## Output
{response_data.get('output_file', 'N/A')}
"""

        category = "debug" if failed > 0 else "insight"
        status = f"❌ {failed} failed" if failed > 0 else f"✅ {passed} passed"

        return save_session_insight(
            title=f"Tests: {status}",
            content=content,
            category=category,
            metadata={"type": "test_run", "passed": passed, "failed": failed}
        )
    except ImportError:
        logger.debug("Session memory not available for saving test run")
        return {"success": False, "error": "Memory system not available"}

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


async def run_tests_async(
    test_path: Optional[str] = None,
    test_framework: str = "auto",
    verbose: bool = True,
    coverage: bool = False,
    output_path: Optional[str] = None,
    ctx: Optional["Context"] = None,
) -> str:
    """
    Execute test suites with flexible options (async with progress).

    Args:
        test_path: Path to test file/directory (default: tests/)
        test_framework: pytest, unittest, ctest, or auto (default: auto)
        verbose: Show detailed output (default: true)
        coverage: Generate coverage report (default: false)
        output_path: Path for test results (default: test-results/)
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with test execution results
    """
    start_time = time.time()

    try:
        # ═══ PROGRESS: Step 1/4 - Initialize ═══
        await _report_progress(ctx, 1, 4, "Initializing test runner...")
        await _log_info(ctx, "🧪 Starting test execution")

        from project_management_automation.scripts.automate_run_tests import TestRunner
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()

        config = {
            'test_path': test_path or 'tests/',
            'test_framework': test_framework,
            'verbose': verbose,
            'coverage': coverage,
            'output_path': output_path or 'test-results/'
        }

        # ═══ PROGRESS: Step 2/4 - Detect framework ═══
        await _report_progress(ctx, 2, 4, f"Configuring {test_framework} framework...")
        await _log_info(ctx, f"📁 Test path: {config['test_path']}")

        runner = TestRunner(config, project_root)

        # ═══ PROGRESS: Step 3/4 - Run tests ═══
        await _report_progress(ctx, 3, 4, "Running tests...")

        # Run in executor to avoid blocking
        loop = asyncio.get_event_loop()
        results = await loop.run_in_executor(None, runner.run)

        # ═══ PROGRESS: Step 4/4 - Process results ═══
        await _report_progress(ctx, 4, 4, "Processing test results...")

        # Format response
        response_data = {
            'framework': results.get('results', {}).get('framework', 'unknown'),
            'tests_run': results.get('results', {}).get('tests_run', 0),
            'tests_passed': results.get('results', {}).get('tests_passed', 0),
            'tests_failed': results.get('results', {}).get('tests_failed', 0),
            'tests_skipped': results.get('results', {}).get('tests_skipped', 0),
            'duration': results.get('results', {}).get('duration', 0),
            'output_file': results.get('results', {}).get('output_file'),
            'coverage_file': results.get('results', {}).get('coverage_file'),
            'status': results.get('results', {}).get('status', 'unknown')
        }

        # Log summary
        passed = response_data['tests_passed']
        failed = response_data['tests_failed']
        skipped = response_data['tests_skipped']

        if failed == 0:
            await _log_info(ctx, f"✅ All {passed} tests passed!")
        else:
            await _log_info(ctx, f"⚠️ {failed} failed, {passed} passed, {skipped} skipped")

        duration = time.time() - start_time
        log_automation_execution('run_tests', duration, True)

        # ═══ MEMORY INTEGRATION: Save test results ═══
        memory_result = _save_test_run_memory(response_data)
        if memory_result.get('success'):
            response_data['memory_saved'] = memory_result.get('memory_id')

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('run_tests', duration, False, e)
        if ctx:
            await _log_info(ctx, f"❌ Test run failed: {e}")

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def run_tests(
    test_path: Optional[str] = None,
    test_framework: str = "auto",
    verbose: bool = True,
    coverage: bool = False,
    output_path: Optional[str] = None,
    ctx: Optional["Context"] = None,
) -> str:
    """
    Execute test suites with flexible options (sync wrapper).

    For async version with progress reporting, use run_tests_async().

    Args:
        test_path: Path to test file/directory (default: tests/)
        test_framework: pytest, unittest, ctest, or auto (default: auto)
        verbose: Show detailed output (default: true)
        coverage: Generate coverage report (default: false)
        output_path: Path for test results (default: test-results/)
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with test execution results
    """
    # Toggle: check if we're in an async context
    in_async = False
    try:
        asyncio.get_running_loop()
        in_async = True
    except RuntimeError:
        pass

    if in_async:
        raise RuntimeError("Use run_tests_async() in async context, or call from sync code")
    return asyncio.run(run_tests_async(test_path, test_framework, verbose, coverage, output_path, ctx))

