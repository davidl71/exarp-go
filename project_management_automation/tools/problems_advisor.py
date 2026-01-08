"""
Problems Advisor MCP Tool

Analyzes IDE diagnostics (linter errors, warnings) and provides:
- Intelligent resolution hints
- Categorization by severity and type
- Metrics tracking
- Auto-fix suggestions where possible

Works with Cursor's read_lints tool output to provide actionable advice.

Memory Integration:
- Saves problem resolutions for pattern matching
"""

import json
import logging
import re
import time
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)


def _save_problems_memory(analysis: dict[str, Any]) -> dict[str, Any]:
    """Save problem analysis as memory for pattern matching."""
    try:
        from .session_memory import save_session_insight

        total = analysis.get('total_problems', 0)

        # Only save if there were problems
        if total == 0:
            return {"success": True, "skipped": "no_problems"}

        by_severity = analysis.get('by_severity', {})
        by_category = analysis.get('by_category', {})

        content = f"""Problems analysis completed.

## Total Problems: {total}

### By Severity
{chr(10).join(f'- {sev}: {count}' for sev, count in by_severity.items()) or '- None'}

### By Category
{chr(10).join(f'- {cat}: {count}' for cat, count in by_category.items()) or '- None'}

### Top Files
{chr(10).join(f'- {f}: {c} problems' for f, c in list(analysis.get('by_file', {}).items())[:5]) or '- None'}

### Resolution Hints Used
{analysis.get('hints_provided', 0)} hints provided
"""

        return save_session_insight(
            title=f"Problems: {total} found",
            content=content,
            category="debug",
            metadata={"type": "problems_analysis", "total_problems": total}
        )
    except ImportError:
        logger.debug("Session memory not available for saving problems analysis")
        return {"success": False, "error": "Memory system not available"}

# Import error handler
try:
    from ..error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
except ImportError:

    def format_success_response(data, message=None):
        return {"success": True, "data": data, "timestamp": time.time()}

    def format_error_response(error, error_code, include_traceback=False):
        return {"success": False, "error": {"code": str(error_code), "message": str(error)}}

    def log_automation_execution(name, duration, success, error=None):
        logger.info(f"{name}: {duration:.2f}s, success={success}")

    class ErrorCode:
        AUTOMATION_ERROR = "AUTOMATION_ERROR"


class ProblemCategory(Enum):
    """Categories of problems with resolution strategies."""

    SPELLING = "spelling"
    TYPE_ERROR = "type_error"
    IMPORT_ERROR = "import_error"
    UNUSED_CODE = "unused_code"
    SYNTAX_ERROR = "syntax_error"
    STYLE = "style"
    SECURITY = "security"
    DEPRECATION = "deprecation"
    PERFORMANCE = "performance"
    UNKNOWN = "unknown"


@dataclass
class ProblemHint:
    """A hint for resolving a problem."""

    category: str
    pattern: str
    description: str
    resolution: str
    auto_fixable: bool
    severity_weight: int  # 1-10, higher = more critical
    tools: list[str]  # Tools that can help


# Problem resolution hints database
PROBLEM_HINTS: list[ProblemHint] = [
    # Spelling
    ProblemHint(
        category="spelling",
        pattern=r"Unknown word|cspell|spell",
        description="Spelling error detected",
        resolution="Add to .cspell/project-words.txt or fix the spelling",
        auto_fixable=True,
        severity_weight=1,
        tools=["cspell", "Code Spell Checker extension"],
    ),
    # Python Type Errors
    ProblemHint(
        category="type_error",
        pattern=r"Argument of type|Type .* cannot be assigned|Incompatible type",
        description="Type mismatch in Python code",
        resolution="Check function signature and argument types. Use type: ignore comment if intentional.",
        auto_fixable=False,
        severity_weight=6,
        tools=["pyright", "mypy", "pylance"],
    ),
    # Import Errors
    ProblemHint(
        category="import_error",
        pattern=r"Import .* could not be resolved|Cannot find module|No module named",
        description="Module import cannot be resolved",
        resolution="1. Check if package is installed (pip install)\n2. Verify PYTHONPATH\n3. Check for typos in import path",
        auto_fixable=False,
        severity_weight=8,
        tools=["pip", "pyright"],
    ),
    # Unused Code
    ProblemHint(
        category="unused_code",
        pattern=r"is not accessed|unused|never read|Unused import",
        description="Unused variable, import, or function",
        resolution="Remove unused code or prefix with _ if intentionally unused",
        auto_fixable=True,
        severity_weight=2,
        tools=["ruff", "pylint", "autoflake"],
    ),
    # Rust Errors
    ProblemHint(
        category="type_error",
        pattern=r"expected .*, found|mismatched types|E0308",
        description="Rust type mismatch",
        resolution="Check expected vs actual types. Use .into() or explicit casting.",
        auto_fixable=False,
        severity_weight=7,
        tools=["rust-analyzer", "cargo check"],
    ),
    # Rust Borrow Checker
    ProblemHint(
        category="type_error",
        pattern=r"borrow|lifetime|E0502|E0499|E0597",
        description="Rust ownership/borrowing error",
        resolution="Review ownership rules. Consider: Clone, Rc/Arc, or restructuring lifetimes.",
        auto_fixable=False,
        severity_weight=9,
        tools=["rust-analyzer"],
    ),
    # Security
    ProblemHint(
        category="security",
        pattern=r"security|vulnerability|CVE|unsafe|injection|XSS",
        description="Potential security issue",
        resolution="Review security implications. Consult OWASP guidelines.",
        auto_fixable=False,
        severity_weight=10,
        tools=["semgrep", "bandit", "cargo-audit"],
    ),
    # Deprecation
    ProblemHint(
        category="deprecation",
        pattern=r"deprecated|will be removed|obsolete",
        description="Using deprecated API",
        resolution="Update to recommended replacement API",
        auto_fixable=False,
        severity_weight=4,
        tools=["documentation", "changelog"],
    ),
    # ESLint/TypeScript
    ProblemHint(
        category="style",
        pattern=r"eslint|prettier|@typescript-eslint",
        description="Code style/linting issue",
        resolution="Run formatter or fix per rule documentation",
        auto_fixable=True,
        severity_weight=2,
        tools=["eslint --fix", "prettier --write"],
    ),
    # C++ Errors
    ProblemHint(
        category="type_error",
        pattern=r"no matching function|cannot convert|undefined reference",
        description="C++ compilation error",
        resolution="Check function signatures, includes, and linking",
        auto_fixable=False,
        severity_weight=8,
        tools=["clangd", "cmake"],
    ),
    # Missing Files
    ProblemHint(
        category="import_error",
        pattern=r"file not found|No such file|ENOENT",
        description="Referenced file does not exist",
        resolution="Create the file or fix the path reference",
        auto_fixable=False,
        severity_weight=7,
        tools=["file system"],
    ),
    # JSON/YAML Syntax
    ProblemHint(
        category="syntax_error",
        pattern=r"JSON|YAML|parse error|Unexpected token",
        description="Configuration file syntax error",
        resolution="Validate JSON/YAML syntax. Check for trailing commas, missing quotes.",
        auto_fixable=False,
        severity_weight=6,
        tools=["jsonlint", "yamllint"],
    ),
]


def categorize_problem(message: str) -> ProblemHint:
    """Match a problem message to a hint category."""
    for hint in PROBLEM_HINTS:
        if re.search(hint.pattern, message, re.IGNORECASE):
            return hint

    # Return unknown category
    return ProblemHint(
        category="unknown",
        pattern="",
        description="Uncategorized problem",
        resolution="Review the error message and consult language documentation",
        auto_fixable=False,
        severity_weight=5,
        tools=["language documentation"],
    )


def analyze_problems(problems: list[dict[str, Any]], include_hints: bool = True) -> dict[str, Any]:
    """
    Analyze a list of problems and provide resolution hints.

    Args:
        problems: List of problem dicts with 'message', 'severity', 'file', 'line' keys
        include_hints: Whether to include resolution hints

    Returns:
        Analysis results with categorization and hints
    """
    results = {
        "total_count": len(problems),
        "by_severity": {"error": 0, "warning": 0, "info": 0, "hint": 0},
        "by_category": {},
        "auto_fixable_count": 0,
        "critical_count": 0,
        "problems_with_hints": [],
    }

    for problem in problems:
        message = problem.get("message", "")
        severity = problem.get("severity", "warning").lower()

        # Count by severity
        if severity in results["by_severity"]:
            results["by_severity"][severity] += 1

        # Categorize and get hint
        hint = categorize_problem(message)

        # Count by category
        cat = hint.category
        results["by_category"][cat] = results["by_category"].get(cat, 0) + 1

        # Count auto-fixable and critical
        if hint.auto_fixable:
            results["auto_fixable_count"] += 1
        if hint.severity_weight >= 8:
            results["critical_count"] += 1

        # Add hint to problem if requested
        if include_hints:
            problem_with_hint = {
                **problem,
                "category": hint.category,
                "hint": {
                    "description": hint.description,
                    "resolution": hint.resolution,
                    "auto_fixable": hint.auto_fixable,
                    "tools": hint.tools,
                    "severity_weight": hint.severity_weight,
                },
            }
            results["problems_with_hints"].append(problem_with_hint)

    return results


def get_quick_fixes(category: str) -> list[str]:
    """Get quick fix commands for a category."""
    fixes = {
        "spelling": ["Add word to .cspell/project-words.txt", "npx cspell --words-only <file>"],
        "unused_code": ["ruff check --fix .", "autoflake --in-place --remove-unused-variables <file>"],
        "style": ["npx prettier --write <file>", "npx eslint --fix <file>", "ruff format ."],
        "import_error": ["pip install <package>", "cargo add <crate>", "npm install <package>"],
        "type_error": ["Review type annotations", "Check function signatures", "Use explicit type casting"],
    }
    return fixes.get(category, ["Consult error message and documentation"])


def analyze_problems_tool(problems_json: str, include_hints: bool = True, output_path: Optional[str] = None) -> str:
    """
    [HINT: Problems advisor. Analyzes linter errors, provides resolution hints, tracks metrics.]

    Analyze IDE problems/diagnostics and provide intelligent resolution hints.

    Args:
        problems_json: JSON string of problems array from read_lints tool
                      Format: [{"message": "...", "severity": "error|warning", "file": "...", "line": 1}]
        include_hints: Include resolution hints for each problem (default: True)
        output_path: Optional path to save detailed report

    Returns:
        JSON string with analysis results and hints

    Example input:
        [
            {"message": "Unknown word: Hcoma", "severity": "warning", "file": "README.md", "line": 10},
            {"message": "Import could not be resolved", "severity": "error", "file": "main.py", "line": 1}
        ]
    """
    start_time = time.time()

    try:
        # Parse problems JSON
        problems = json.loads(problems_json)
        if not isinstance(problems, list):
            problems = [problems]

        # Analyze
        analysis = analyze_problems(problems, include_hints)

        # Add summary
        analysis["summary"] = {
            "total": analysis["total_count"],
            "errors": analysis["by_severity"]["error"],
            "warnings": analysis["by_severity"]["warning"],
            "auto_fixable": analysis["auto_fixable_count"],
            "critical": analysis["critical_count"],
            "top_categories": sorted(analysis["by_category"].items(), key=lambda x: x[1], reverse=True)[:5],
        }

        # Add quick fix suggestions for top categories
        analysis["quick_fixes"] = {}
        for cat, count in analysis["by_category"].items():
            if count > 0:
                analysis["quick_fixes"][cat] = get_quick_fixes(cat)

        # Save report if requested
        if output_path:
            report_path = Path(output_path)
            report_content = generate_problems_report(analysis)
            report_path.write_text(report_content)
            analysis["report_path"] = str(report_path.absolute())

        duration = time.time() - start_time
        log_automation_execution("analyze_problems", duration, True)

        # ═══ MEMORY INTEGRATION: Save problems analysis ═══
        memory_result = _save_problems_memory(analysis)
        if memory_result.get('success') and not memory_result.get('skipped'):
            analysis['memory_saved'] = memory_result.get('memory_id')

        return json.dumps(format_success_response(analysis, "Problems analysis completed"), indent=2)

    except json.JSONDecodeError as e:
        return json.dumps(format_error_response(f"Invalid JSON input: {e}", ErrorCode.AUTOMATION_ERROR), indent=2)
    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("analyze_problems", duration, False, str(e))
        logger.error(f"Error analyzing problems: {e}", exc_info=True)

        return json.dumps(
            format_error_response(
                f"Problems analysis failed: {str(e)}", ErrorCode.AUTOMATION_ERROR, include_traceback=True
            ),
            indent=2,
        )


def generate_problems_report(analysis: dict[str, Any]) -> str:
    """Generate a markdown report from analysis results."""
    lines = [
        "# Problems Analysis Report",
        "",
        f"**Generated:** {time.strftime('%Y-%m-%d %H:%M:%S')}",
        "",
        "## Summary",
        "",
        "| Metric | Count |",
        "|--------|-------|",
        f"| Total Problems | {analysis['total_count']} |",
        f"| Errors | {analysis['by_severity']['error']} |",
        f"| Warnings | {analysis['by_severity']['warning']} |",
        f"| Auto-fixable | {analysis['auto_fixable_count']} |",
        f"| Critical (severity >= 8) | {analysis['critical_count']} |",
        "",
        "## By Category",
        "",
    ]

    for cat, count in sorted(analysis["by_category"].items(), key=lambda x: x[1], reverse=True):
        lines.append(f"- **{cat}**: {count}")
        if cat in analysis.get("quick_fixes", {}):
            for fix in analysis["quick_fixes"][cat]:
                lines.append(f"  - `{fix}`")

    lines.extend(
        [
            "",
            "## Problems with Hints",
            "",
        ]
    )

    for i, problem in enumerate(analysis.get("problems_with_hints", [])[:20], 1):
        lines.append(f"### {i}. {problem.get('file', 'Unknown')}:{problem.get('line', '?')}")
        lines.append(f"**Message:** {problem.get('message', 'No message')}")
        lines.append(f"**Severity:** {problem.get('severity', 'unknown')}")
        lines.append(f"**Category:** {problem.get('category', 'unknown')}")
        if "hint" in problem:
            lines.append(f"**Resolution:** {problem['hint']['resolution']}")
            lines.append(f"**Tools:** {', '.join(problem['hint']['tools'])}")
        lines.append("")

    return "\n".join(lines)


def list_problem_categories() -> str:
    """
    [HINT: List problem categories. Shows all recognized categories with resolution strategies.]

    List all recognized problem categories and their resolution strategies.

    Returns:
        JSON string with categories and hints
    """
    categories = []
    for hint in PROBLEM_HINTS:
        categories.append(
            {
                "category": hint.category,
                "description": hint.description,
                "resolution": hint.resolution,
                "auto_fixable": hint.auto_fixable,
                "severity_weight": hint.severity_weight,
                "tools": hint.tools,
            }
        )

    return json.dumps(
        format_success_response({"categories": categories, "total": len(categories)}, "Problem categories listed"),
        indent=2,
    )


# CLI testing
if __name__ == "__main__":
    # Test with sample problems
    test_problems = [
        {"message": "Unknown word: Raagiosl", "severity": "warning", "file": "README.md", "line": 10},
        {"message": "Import 'foo' could not be resolved", "severity": "error", "file": "main.py", "line": 1},
        {"message": "Variable 'x' is not accessed", "severity": "warning", "file": "utils.py", "line": 25},
        {"message": "expected i32, found &str", "severity": "error", "file": "main.rs", "line": 42},
    ]

    result = analyze_problems_tool(json.dumps(test_problems))
    print(result)
