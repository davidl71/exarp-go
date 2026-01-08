"""
Linter Tool - Run external linters and analyze results.

Bridges the gap between external linters (ruff, flake8) and the analyze_problems tool.
Runs linters, converts output to problems_json format, and optionally analyzes.

Memory Integration:
- Saves linting results for pattern tracking
"""

import json
import logging
import subprocess
import time
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Import error handler
try:
    from ..error_handler import (
        ErrorCode,
        format_error_response,
        format_success_response,
        log_automation_execution,
    )
except ImportError:
    def format_success_response(data, message=None):
        return {"success": True, "data": data, "timestamp": time.time()}

    def format_error_response(error, error_code, include_traceback=False):
        return {"success": False, "error": {"code": str(error_code), "message": str(error)}}

    def log_automation_execution(name, duration, success, error=None):
        logger.info(f"{name}: {duration:.2f}s, success={success}")

    class ErrorCode:
        AUTOMATION_ERROR = "AUTOMATION_ERROR"


def _save_linter_memory(response_data: dict[str, Any]) -> dict[str, Any]:
    """Save linter results as memory for pattern tracking."""
    try:
        from .session_memory import save_session_insight

        total = response_data.get('total_issues', 0)

        # Only save if there were issues
        if total == 0:
            return {"success": True, "skipped": "no_issues"}

        by_severity = response_data.get('by_severity', {})
        by_category = response_data.get('by_category', {})

        content = f"""Linter run completed.

## Total Issues: {total}

### By Severity
{chr(10).join(f'- {sev}: {count}' for sev, count in by_severity.items()) or '- None'}

### By Category
{chr(10).join(f'- {cat}: {count}' for cat, count in by_category.items()) or '- None'}

### Tool
{response_data.get('linter', 'unknown')}

### Files Checked
{response_data.get('files_checked', 0)} files
"""

        return save_session_insight(
            title=f"Linter: {total} issues",
            content=content,
            category="debug" if total > 0 else "insight",
            metadata={"type": "linter_run", "total_issues": total}
        )
    except ImportError:
        logger.debug("Session memory not available for saving linter results")
        return {"success": False, "error": "Memory system not available"}


def _parse_ruff_output(output: str) -> list[dict[str, Any]]:
    """Parse ruff JSON output into problems format."""
    problems = []

    try:
        # Ruff with --output-format=json returns JSON array
        issues = json.loads(output) if output.strip() else []

        for issue in issues:
            # Map ruff severity to standard severity
            # Ruff doesn't have severity, use code prefix
            code = issue.get('code', '')
            if code.startswith('E') or code.startswith('F'):
                severity = 'error'
            elif code.startswith('W'):
                severity = 'warning'
            else:
                severity = 'information'

            problems.append({
                'file': issue.get('filename', ''),
                'line': issue.get('location', {}).get('row', 0),
                'column': issue.get('location', {}).get('column', 0),
                'message': issue.get('message', ''),
                'code': code,
                'severity': severity,
                'source': 'ruff',
                'fix_available': issue.get('fix') is not None,
            })
    except json.JSONDecodeError:
        # Fallback: parse text output
        for line in output.strip().split('\n'):
            if ':' in line and line.strip():
                parts = line.split(':')
                if len(parts) >= 4:
                    problems.append({
                        'file': parts[0],
                        'line': int(parts[1]) if parts[1].isdigit() else 0,
                        'column': int(parts[2]) if parts[2].isdigit() else 0,
                        'message': ':'.join(parts[3:]).strip(),
                        'severity': 'warning',
                        'source': 'ruff',
                    })

    return problems


def run_linter(
    path: Optional[str] = None,
    linter: str = "ruff",
    fix: bool = False,
    analyze: bool = True,
    select: Optional[str] = None,
    ignore: Optional[str] = None,
) -> str:
    """
    Run external linter and optionally analyze results.

    Args:
        path: File or directory to lint (default: current directory)
        linter: Linter to use - "ruff" or "flake8" (default: ruff)
        fix: Auto-fix issues where possible (default: false)
        analyze: Run analyze_problems on results (default: true)
        select: Comma-separated rule codes to enable (e.g., "E,F,W")
        ignore: Comma-separated rule codes to ignore (e.g., "E501")

    Returns:
        JSON string with linting results and optional analysis
    """
    start_time = time.time()

    try:
        from ..utils import find_project_root

        project_root = find_project_root()
        target_path = path or str(project_root)

        # Build command based on linter
        if linter == "ruff":
            cmd = ["ruff", "check", "--output-format=json"]
            if fix:
                cmd.append("--fix")
            if select:
                cmd.extend(["--select", select])
            if ignore:
                cmd.extend(["--ignore", ignore])
            cmd.append(target_path)
        elif linter == "flake8":
            cmd = ["flake8", "--format=json"]
            if select:
                cmd.extend(["--select", select])
            if ignore:
                cmd.extend(["--ignore", ignore])
            cmd.append(target_path)
        else:
            return json.dumps(format_error_response(
                f"Unknown linter: {linter}. Use 'ruff' or 'flake8'",
                ErrorCode.AUTOMATION_ERROR
            ), indent=2)

        # Run linter
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=120,
            cwd=str(project_root)
        )

        # Parse output
        if linter == "ruff":
            problems = _parse_ruff_output(result.stdout)
        else:
            # For other linters, try to parse as JSON or text
            problems = _parse_ruff_output(result.stdout)

        # Categorize issues
        by_severity: dict[str, int] = {}
        by_category: dict[str, int] = {}
        by_file: dict[str, int] = {}

        for p in problems:
            sev = p.get('severity', 'unknown')
            by_severity[sev] = by_severity.get(sev, 0) + 1

            # Categorize by code prefix
            code = p.get('code', '')
            if code.startswith('E'):
                cat = 'error'
            elif code.startswith('F'):
                cat = 'fatal'
            elif code.startswith('W'):
                cat = 'warning'
            elif code.startswith('I'):
                cat = 'import'
            elif code.startswith('UP'):
                cat = 'upgrade'
            else:
                cat = 'other'
            by_category[cat] = by_category.get(cat, 0) + 1

            file = p.get('file', 'unknown')
            by_file[file] = by_file.get(file, 0) + 1

        # Count files checked (approximation from path)
        target = Path(target_path)
        if target.is_file():
            files_checked = 1
        elif target.is_dir():
            files_checked = len(list(target.rglob("*.py")))
        else:
            files_checked = 0

        response_data = {
            'linter': linter,
            'path': target_path,
            'total_issues': len(problems),
            'by_severity': by_severity,
            'by_category': by_category,
            'by_file': dict(list(by_file.items())[:10]),  # Top 10 files
            'files_checked': files_checked,
            'fix_applied': fix,
            'problems': problems[:50],  # Limit to 50 in response
            'problems_json': json.dumps(problems),  # Full list for analyze_problems
        }

        # Optionally run analyze_problems
        if analyze and problems:
            try:
                from .problems_advisor import analyze_problems_tool
                analysis_result = json.loads(analyze_problems_tool(
                    json.dumps(problems),
                    include_hints=True
                ))
                if analysis_result.get('success'):
                    response_data['analysis'] = analysis_result.get('data', {})
            except Exception as e:
                logger.warning(f"Could not run analyze_problems: {e}")

        duration = time.time() - start_time
        log_automation_execution('run_linter', duration, True)

        # ═══ MEMORY INTEGRATION: Save linter results ═══
        memory_result = _save_linter_memory(response_data)
        if memory_result.get('success') and not memory_result.get('skipped'):
            response_data['memory_saved'] = memory_result.get('memory_id')

        return json.dumps(format_success_response(response_data), indent=2)

    except FileNotFoundError:
        return json.dumps(format_error_response(
            f"Linter '{linter}' not found. Install with: pip install {linter}",
            ErrorCode.AUTOMATION_ERROR
        ), indent=2)
    except subprocess.TimeoutExpired:
        return json.dumps(format_error_response(
            "Linter timed out after 120 seconds",
            ErrorCode.AUTOMATION_ERROR
        ), indent=2)
    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('run_linter', duration, False, e)
        return json.dumps(format_error_response(str(e), ErrorCode.AUTOMATION_ERROR), indent=2)


def get_linter_status() -> str:
    """
    Check which linters are available.

    Returns:
        JSON string with available linters and their versions
    """
    linters = {}

    for linter in ['ruff', 'flake8', 'pylint', 'mypy', 'black']:
        try:
            result = subprocess.run(
                [linter, '--version'],
                capture_output=True,
                text=True,
                timeout=5
            )
            if result.returncode == 0:
                version = result.stdout.strip().split('\n')[0]
                linters[linter] = {'available': True, 'version': version}
            else:
                linters[linter] = {'available': False}
        except FileNotFoundError:
            linters[linter] = {'available': False}
        except Exception as e:
            linters[linter] = {'available': False, 'error': str(e)}

    return json.dumps(format_success_response({
        'linters': linters,
        'recommended': 'ruff' if linters.get('ruff', {}).get('available') else 'flake8',
    }), indent=2)

