"""
MCP Tool Wrapper for Dependency Security Scan

Wraps DependencySecurityAnalyzer to expose as MCP tool.

Memory Integration:
- Saves vulnerability findings for tracking remediation

FastMCP Integration:
- Progress reporting via ctx.report_progress()
- Client logging via ctx.info()
"""

import asyncio
import json
import logging
import time
from pathlib import Path
from typing import Any, List, Optional, TYPE_CHECKING

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


def _save_security_scan_memory(response_data: dict[str, Any]) -> dict[str, Any]:
    """Save security scan results as memory for remediation tracking."""
    try:
        from .session_memory import save_session_insight

        total_vulns = response_data.get('total_vulnerabilities', 0)

        content = f"""Dependency security scan completed.

## Vulnerabilities Found: {total_vulns}

### By Severity
- Critical: {response_data.get('critical_count', 0)}
- High: {response_data.get('high_count', 0)}
- Medium: {response_data.get('medium_count', 0)}
- Low: {response_data.get('low_count', 0)}

### By Language
- Python: {response_data.get('python_vulnerabilities', 0)}
- Rust: {response_data.get('rust_vulnerabilities', 0)}
- NPM: {response_data.get('npm_vulnerabilities', 0)}

## Report
{response_data.get('report_path', 'N/A')}
"""

        severity = "debug" if total_vulns == 0 else "insight"
        title_suffix = "✅ Clean" if total_vulns == 0 else f"⚠️ {total_vulns} vulns"

        return save_session_insight(
            title=f"Security Scan: {title_suffix}",
            content=content,
            category=severity,
            metadata={"type": "security_scan", "total_vulnerabilities": total_vulns}
        )
    except ImportError:
        logger.debug("Session memory not available for saving security scan")
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


async def scan_dependency_security_async(
    languages: Optional[List[str]] = None,
    config_path: Optional[str] = None,
    ctx: Optional["Context"] = None,
) -> str:
    """
    Scan project dependencies for security vulnerabilities (async with progress).

    Args:
        languages: List of languages to scan (python, rust, npm). If None, scans all.
        config_path: Path to dependency security config file (default: scripts/dependency_security_config.json)
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with scan results
    """
    start_time = time.time()

    try:
        # ═══ PROGRESS: Step 1/5 - Initialize ═══
        await _report_progress(ctx, 1, 5, "Initializing security scanner...")
        await _log_info(ctx, "🔒 Starting dependency security scan")

        # Import from package
        from project_management_automation.scripts.automate_dependency_security import DependencySecurityAnalyzer
        from project_management_automation.utils import find_project_root

        # Find project root
        project_root = find_project_root()

        # ═══ PROGRESS: Step 2/5 - Load config ═══
        await _report_progress(ctx, 2, 5, "Loading configuration...")

        # Use default config if not provided
        if not config_path:
            config_path = str(project_root / 'scripts' / 'dependency_security_config.json')

        # Load and modify config if languages specified
        if languages:
            await _log_info(ctx, f"📋 Scanning languages: {', '.join(languages)}")
            import json as json_module
            with open(config_path) as f:
                config_data = json_module.load(f)

            # Enable only specified languages
            scan_configs = config_data.get('scan_configs', {})
            for lang in ['python', 'rust', 'npm']:
                if lang in scan_configs:
                    scan_configs[lang]['enabled'] = lang in languages

            # Write temporary config
            temp_config_path = project_root / 'scripts' / '.temp_dependency_security_config.json'
            with open(temp_config_path, 'w') as f:
                json_module.dump(config_data, f, indent=2)
            config_path = str(temp_config_path)
        else:
            await _log_info(ctx, "📋 Scanning all languages: python, rust, npm")

        # ═══ PROGRESS: Step 3/5 - Run scanner ═══
        await _report_progress(ctx, 3, 5, "Running vulnerability scanner...")

        # Create analyzer and run (this is the slow part)
        analyzer = DependencySecurityAnalyzer(config_path, project_root)

        # Run in executor to avoid blocking
        loop = asyncio.get_event_loop()
        results = await loop.run_in_executor(None, analyzer.run)

        # ═══ PROGRESS: Step 4/5 - Process results ═══
        await _report_progress(ctx, 4, 5, "Processing scan results...")

        # Extract key metrics - scan_results are nested in results['results']
        scan_results = results.get('results', {})
        summary = scan_results.get('summary', {})

        # Format response
        response_data = {
            'total_vulnerabilities': summary.get('total_vulnerabilities', 0),
            'by_severity': summary.get('by_severity', {}),
            'by_language': summary.get('by_language', {}),
            'critical_vulnerabilities': len(summary.get('critical_vulnerabilities', [])),
            'python_vulnerabilities': len(scan_results.get('python', [])),
            'rust_vulnerabilities': len(scan_results.get('rust', [])),
            'npm_vulnerabilities': len(scan_results.get('npm', [])),
            'report_path': str(analyzer.output_file.absolute()),
            'status': results.get('status', 'unknown')
        }

        # ═══ PROGRESS: Step 5/5 - Finalize ═══
        await _report_progress(ctx, 5, 5, "Finalizing...")

        total_vulns = response_data['total_vulnerabilities']
        critical_count = response_data.get('critical_vulnerabilities', 0)

        if total_vulns == 0:
            await _log_info(ctx, "✅ No vulnerabilities found!")
        else:
            await _log_info(ctx, f"⚠️ Found {total_vulns} vulnerabilities")

        duration = time.time() - start_time
        log_automation_execution('scan_dependency_security', duration, True)

        # ═══ MEMORY INTEGRATION: Save scan results ═══
        memory_result = _save_security_scan_memory(response_data)
        if memory_result.get('success'):
            response_data['memory_saved'] = memory_result.get('memory_id')

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('scan_dependency_security', duration, False, e)
        if ctx:
            await _log_info(ctx, f"❌ Security scan failed: {e}")
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def scan_dependency_security(
    languages: Optional[List[str]] = None,
    config_path: Optional[str] = None,
    ctx: Optional["Context"] = None,
) -> str:
    """
    Scan project dependencies for security vulnerabilities (sync wrapper).

    For async version with progress reporting, use scan_dependency_security_async().

    Args:
        languages: List of languages to scan (python, rust, npm). If None, scans all.
        config_path: Path to dependency security config file
        ctx: FastMCP Context for progress reporting (optional)

    Returns:
        JSON string with scan results
    """
    # Check if we're in an async context using try/except/else pattern
    # This ensures the intentional RuntimeError is raised in else block, not caught
    try:
        asyncio.get_running_loop()
        # If we get here, we're in an async context - raise helpful error
        raise RuntimeError(
            "scan_dependency_security() cannot be called from an async context. "
            "Use scan_dependency_security_async() instead and await it."
        )
    except RuntimeError as e:
        # Re-raise if it's our intentional error
        if "scan_dependency_security_async()" in str(e) or "async context" in str(e).lower():
            raise
        # Otherwise, no running loop - safe to use asyncio.run()
        return asyncio.run(scan_dependency_security_async(languages, config_path, ctx))
