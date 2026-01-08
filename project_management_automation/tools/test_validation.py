"""
Test Structure Validation Tool

Validates test organization and patterns.
Checks naming conventions, validates organization, and identifies missing test files.
"""

import logging
import re
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Import error handler
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
        import time
        def format_success_response(data, message=None):
            return {"success": True, "data": data, "timestamp": time.time()}
        def format_error_response(error, error_code, include_traceback=False):
            return {"success": False, "error": {"code": str(error_code), "message": str(error)}}
        def log_automation_execution(name, duration, success, error=None):
            logger.info(f"{name}: {duration:.2f}s, success={success}")
        class ErrorCode:
            AUTOMATION_ERROR = "AUTOMATION_ERROR"


def validate_test_structure(
    test_path: Optional[str] = None,
    framework: Optional[str] = None,
    output_path: Optional[str] = None,
) -> dict[str, Any]:
    """
    Validate test organization and patterns.

    Args:
        test_path: Path to test directory (default: tests/)
        framework: Expected framework - pytest, unittest, or auto (default: auto)
        output_path: Path for validation report (optional)

    Returns:
        Dictionary with validation results
    """
    import time
    start_time = time.time()

    try:
        from ..utils import find_project_root

        project_root = find_project_root()

        # Determine test directory
        if test_path:
            test_dir = Path(test_path)
            if not test_dir.is_absolute():
                test_dir = project_root / test_dir
        else:
            test_dir = project_root / "tests"

        if not test_dir.exists():
            return format_error_response(
                ValueError(f"Test directory not found: {test_dir}"),
                ErrorCode.AUTOMATION_ERROR
            )

        # Auto-detect framework if not specified
        if framework == "auto" or framework is None:
            framework = _detect_framework(test_dir)

        # Run validations
        results = {
            "test_directory": str(test_dir.relative_to(project_root)),
            "framework": framework,
            "validations": {},
            "issues": [],
            "warnings": [],
        }

        # Check naming conventions
        naming_issues = _check_naming_conventions(test_dir, framework)
        results["validations"]["naming"] = {
            "passed": len(naming_issues) == 0,
            "issues": naming_issues,
        }
        results["issues"].extend(naming_issues)

        # Check test organization
        org_issues = _check_organization(test_dir, framework)
        results["validations"]["organization"] = {
            "passed": len(org_issues) == 0,
            "issues": org_issues,
        }
        results["issues"].extend(org_issues)

        # Check for missing test files
        missing_tests = _check_missing_tests(project_root, test_dir)
        results["validations"]["coverage"] = {
            "passed": len(missing_tests) == 0,
            "missing_tests": missing_tests,
        }
        if missing_tests:
            results["warnings"].extend([
                f"Missing test file for: {m}" for m in missing_tests[:10]
            ])

        # Overall status
        total_issues = len(results["issues"])
        results["status"] = "pass" if total_issues == 0 else "fail"
        results["summary"] = {
            "total_issues": total_issues,
            "total_warnings": len(results["warnings"]),
            "validations_passed": sum(
                1 for v in results["validations"].values()
                if v.get("passed", False)
            ),
            "total_validations": len(results["validations"]),
        }

        # Save report if requested
        if output_path:
            output_file = Path(output_path)
            if not output_file.is_absolute():
                output_file = project_root / output_file
            output_file.parent.mkdir(parents=True, exist_ok=True)

            # Generate markdown report
            report = _generate_report(results)
            with open(output_file, "w") as f:
                f.write(report)

            results["report_path"] = str(output_file)

        duration = time.time() - start_time
        log_automation_execution("validate_test_structure", duration, True)

        return format_success_response(results)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("validate_test_structure", duration, False, e)
        logger.error(f"Error validating test structure: {e}", exc_info=True)

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return error_response


def _detect_framework(test_dir: Path) -> str:
    """Auto-detect test framework."""
    # Check for pytest markers
    for test_file in test_dir.rglob("test_*.py"):
        try:
            content = test_file.read_text(encoding="utf-8")
            if "@pytest.fixture" in content or "import pytest" in content:
                return "pytest"
        except Exception:
            pass

    # Check for unittest
    for test_file in test_dir.rglob("test_*.py"):
        try:
            content = test_file.read_text(encoding="utf-8")
            if "import unittest" in content or "unittest.TestCase" in content:
                return "unittest"
        except Exception:
            pass

    return "pytest"  # Default


def _check_naming_conventions(test_dir: Path, framework: str) -> list[str]:
    """Check test file and function naming conventions."""
    issues = []

    for test_file in test_dir.rglob("*.py"):
        # Check file naming
        if not test_file.name.startswith("test_") and not test_file.name.endswith("_test.py"):
            issues.append(f"File naming: {test_file.name} should start with 'test_' or end with '_test.py'")

        # Check function naming
        try:
            content = test_file.read_text(encoding="utf-8")
            if framework == "pytest":
                # pytest: functions should start with test_
                func_pattern = r"def (test_\w+)"
                matches = re.findall(func_pattern, content)
                for match in matches:
                    if not match.startswith("test_"):
                        issues.append(f"Function naming in {test_file.name}: {match} should start with 'test_'")
        except Exception:
            pass

    return issues


def _check_organization(test_dir: Path, framework: str) -> list[str]:
    """Check test organization."""
    issues = []

    # Check for __init__.py in test directories
    if not (test_dir / "__init__.py").exists():
        issues.append(f"Missing __init__.py in {test_dir}")

    # Check for conftest.py if pytest
    if framework == "pytest" and not (test_dir / "conftest.py").exists():
        issues.append("Missing conftest.py for pytest (recommended)")

    return issues


def _check_missing_tests(project_root: Path, test_dir: Path) -> list[str]:
    """Check for missing test files for source files."""
    missing = []

    # Find source Python files
    source_files = []
    for py_file in project_root.rglob("*.py"):
        # Skip test files, venv, hidden dirs
        if "test" in str(py_file).lower():
            continue
        if "venv" in str(py_file) or py_file.name.startswith("."):
            continue
        if py_file.parent.name.startswith("."):
            continue
        if py_file.parent.name == "__pycache__":
            continue

        source_files.append(py_file)

    # Check for corresponding test files
    for source_file in source_files:
        rel_path = source_file.relative_to(project_root)
        # Look for test file
        test_name = f"test_{source_file.stem}.py"
        test_file = test_dir / rel_path.parent / test_name

        if not test_file.exists():
            missing.append(str(rel_path))

    return missing[:20]  # Limit to 20


def _generate_report(results: dict[str, Any]) -> str:
    """Generate markdown validation report."""
    report = f"""# Test Structure Validation Report

**Framework**: {results['framework']}
**Test Directory**: {results['test_directory']}
**Status**: {results['status'].upper()}

## Summary

- **Total Issues**: {results['summary']['total_issues']}
- **Warnings**: {results['summary']['total_warnings']}
- **Validations Passed**: {results['summary']['validations_passed']}/{results['summary']['total_validations']}

## Validations

"""
    for name, validation in results['validations'].items():
        status = "✅" if validation.get("passed") else "❌"
        report += f"### {name.title()}\n\n{status} {'Passed' if validation.get('passed') else 'Failed'}\n\n"
        if validation.get("issues"):
            report += "**Issues:**\n"
            for issue in validation.get("issues", [])[:10]:
                report += f"- {issue}\n"
        report += "\n"

    if results.get("warnings"):
        report += "## Warnings\n\n"
        for warning in results["warnings"][:10]:
            report += f"- {warning}\n"
        report += "\n"

    return report

