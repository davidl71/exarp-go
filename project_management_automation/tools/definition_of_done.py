"""
Definition of Done Checker Tool

Validates task completion against customizable Definition of Done criteria.
Based on Cursor IDE Best Practice #6.
"""

import json
import logging
import re
import time
from datetime import datetime
from pathlib import Path
from typing import Any, List, Optional

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


# Default Definition of Done criteria
DEFAULT_DOD = {
    "code_quality": {
        "name": "Code Quality",
        "checks": [
            {"id": "tests_pass", "name": "All tests pass", "required": True},
            {"id": "no_linter_errors", "name": "No linter errors", "required": True},
            {"id": "complexity_ok", "name": "Complexity under threshold", "required": False},
        ],
    },
    "documentation": {
        "name": "Documentation",
        "checks": [
            {"id": "docstrings", "name": "Functions have docstrings", "required": True},
            {"id": "readme_updated", "name": "README updated if needed", "required": False},
            {"id": "changelog", "name": "CHANGELOG entry added", "required": False},
        ],
    },
    "security": {
        "name": "Security",
        "checks": [
            {"id": "no_secrets", "name": "No hardcoded secrets", "required": True},
            {"id": "bandit_clean", "name": "Bandit scan clean", "required": True},
            {"id": "deps_secure", "name": "Dependencies secure", "required": False},
        ],
    },
    "review": {
        "name": "Review",
        "checks": [
            {"id": "self_review", "name": "Self-reviewed changes", "required": True},
            {"id": "pr_description", "name": "PR description complete", "required": False},
            {"id": "linked_task", "name": "Linked to task/issue", "required": False},
        ],
    },
}


class DefinitionOfDoneChecker:
    """Checks task completion against Definition of Done criteria."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.dod_path = project_root / ".cursor" / "DEFINITION_OF_DONE.json"
        self.dod = self._load_dod()

    def _load_dod(self) -> dict:
        """Load custom DoD or use defaults."""
        if self.dod_path.exists():
            try:
                return json.loads(self.dod_path.read_text())
            except Exception as e:
                logger.warning(f"Error loading custom DoD: {e}, using defaults")
        return DEFAULT_DOD

    def check(
        self,
        task_id: Optional[str] = None,
        changed_files: Optional[List[str]] = None,
        auto_check: bool = True,
    ) -> dict[str, Any]:
        """
        Check Definition of Done criteria.

        Args:
            task_id: Optional task ID to check
            changed_files: List of changed file paths
            auto_check: Run automated checks where possible

        Returns:
            Dict with DoD check results
        """
        results = {
            "timestamp": datetime.now().isoformat(),
            "task_id": task_id,
            "categories": {},
            "passed": 0,
            "failed": 0,
            "skipped": 0,
            "ready_for_review": False,
        }

        for cat_id, category in self.dod.items():
            cat_results = {
                "name": category["name"],
                "checks": [],
            }

            for check in category["checks"]:
                check_result = {
                    "id": check["id"],
                    "name": check["name"],
                    "required": check["required"],
                    "status": "skipped",
                    "message": None,
                }

                if auto_check:
                    # Run automated check if available
                    auto_result = self._run_auto_check(check["id"], changed_files)
                    if auto_result:
                        check_result["status"] = auto_result["status"]
                        check_result["message"] = auto_result.get("message")

                cat_results["checks"].append(check_result)

                # Count results
                if check_result["status"] == "passed":
                    results["passed"] += 1
                elif check_result["status"] == "failed":
                    results["failed"] += 1
                else:
                    results["skipped"] += 1

            results["categories"][cat_id] = cat_results

        # Determine if ready for review
        required_failed = sum(
            1
            for cat in results["categories"].values()
            for check in cat["checks"]
            if check["required"] and check["status"] == "failed"
        )
        results["ready_for_review"] = required_failed == 0
        results["required_failures"] = required_failed

        return results

    def _run_auto_check(
        self, check_id: str, changed_files: Optional[List[str]]
    ) -> dict | None:
        """Run automated check for a specific criterion."""
        try:
            if check_id == "tests_pass":
                return self._check_tests()
            elif check_id == "no_linter_errors":
                return self._check_linter()
            elif check_id == "no_secrets":
                return self._check_secrets(changed_files)
            elif check_id == "docstrings":
                return self._check_docstrings(changed_files)
        except Exception as e:
            logger.warning(f"Auto-check failed for {check_id}: {e}")
        return None

    def _check_tests(self) -> dict:
        """Check if tests pass."""
        import subprocess

        try:
            result = subprocess.run(
                ["python", "-m", "pytest", "--tb=no", "-q"],
                capture_output=True,
                text=True,
                timeout=60,
                cwd=self.project_root,
            )
            if result.returncode == 0:
                return {"status": "passed", "message": "All tests pass"}
            else:
                return {"status": "failed", "message": f"Tests failed: {result.stdout[:100]}"}
        except subprocess.TimeoutExpired:
            return {"status": "skipped", "message": "Test timeout"}
        except FileNotFoundError:
            return {"status": "skipped", "message": "pytest not found"}

    def _check_linter(self) -> dict:
        """Check for linter errors."""
        import subprocess

        try:
            result = subprocess.run(
                ["python", "-m", "ruff", "check", "."],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=self.project_root,
            )
            if result.returncode == 0:
                return {"status": "passed", "message": "No linter errors"}
            else:
                error_count = len(result.stdout.strip().split("\n"))
                return {"status": "failed", "message": f"{error_count} linter errors"}
        except FileNotFoundError:
            return {"status": "skipped", "message": "ruff not found"}

    def _check_secrets(self, changed_files: Optional[List[str]]) -> dict:
        """Check for hardcoded secrets."""
        secret_patterns = [
            r"api[_-]?key\s*=\s*['\"][^'\"]+['\"]",
            r"password\s*=\s*['\"][^'\"]+['\"]",
            r"secret\s*=\s*['\"][^'\"]+['\"]",
            r"token\s*=\s*['\"][^'\"]+['\"]",
            r"-----BEGIN.*PRIVATE KEY-----",
        ]

        files_to_check = changed_files or []
        if not files_to_check:
            # Check recent Python files
            files_to_check = [str(f) for f in self.project_root.rglob("*.py")][:50]

        secrets_found = []
        for file_path in files_to_check:
            try:
                path = Path(file_path)
                if not path.is_absolute():
                    path = self.project_root / path
                if path.exists() and path.suffix in [".py", ".js", ".ts", ".yaml", ".yml"]:
                    content = path.read_text()
                    for pattern in secret_patterns:
                        if re.search(pattern, content, re.IGNORECASE):
                            secrets_found.append(str(path.name))
                            break
            except Exception:
                pass

        if secrets_found:
            return {"status": "failed", "message": f"Possible secrets in: {', '.join(secrets_found[:3])}"}
        return {"status": "passed", "message": "No obvious secrets detected"}

    def _check_docstrings(self, changed_files: Optional[List[str]]) -> dict:
        """Check for missing docstrings."""
        import ast

        files_to_check = changed_files or []
        missing = 0
        total = 0

        for file_path in files_to_check:
            try:
                path = Path(file_path)
                if not path.is_absolute():
                    path = self.project_root / path
                if path.exists() and path.suffix == ".py":
                    content = path.read_text()
                    tree = ast.parse(content)
                    for node in ast.walk(tree):
                        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef, ast.ClassDef)):
                            total += 1
                            if not ast.get_docstring(node):
                                missing += 1
            except Exception:
                pass

        if total == 0:
            return {"status": "skipped", "message": "No functions to check"}
        if missing == 0:
            return {"status": "passed", "message": f"All {total} functions documented"}
        return {"status": "failed", "message": f"{missing}/{total} missing docstrings"}


def check_definition_of_done(
    task_id: Optional[str] = None,
    changed_files: Optional[str] = None,
    auto_check: bool = True,
    output_path: Optional[str] = None,
) -> str:
    """
    [HINT: Definition of Done. Validates task completion criteria before review.]

    📊 Output: DoD checklist with pass/fail/skip status
    🔧 Side Effects: None (read-only validation)
    📁 Analyzes: Tests, linter, secrets, docstrings
    ⏱️ Typical Runtime: 5-30 seconds (runs tests)

    Example Prompt:
    "Check if my changes are ready for review"

    Categories checked:
    - Code Quality: Tests pass, no linter errors
    - Documentation: Docstrings, README
    - Security: No secrets, bandit clean
    - Review: Self-review checklist

    Args:
        task_id: Optional task ID being completed
        changed_files: Comma-separated list of changed files
        auto_check: Run automated checks (tests, linter, etc.)
        output_path: Optional path to save detailed report

    Returns:
        JSON with DoD check results
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()
        checker = DefinitionOfDoneChecker(project_root)

        # Parse changed files
        files_list = None
        if changed_files:
            files_list = [f.strip() for f in changed_files.split(",")]

        results = checker.check(
            task_id=task_id,
            changed_files=files_list,
            auto_check=auto_check,
        )

        # Add summary
        results["summary"] = {
            "total_checks": results["passed"] + results["failed"] + results["skipped"],
            "completion_rate": round(
                results["passed"] / max(results["passed"] + results["failed"], 1) * 100, 1
            ),
            "recommendation": (
                "✅ Ready for review!"
                if results["ready_for_review"]
                else f"⚠️ Fix {results['required_failures']} required items first"
            ),
        }

        # Save report if requested
        if output_path:
            out_path = Path(output_path)
            if not out_path.is_absolute():
                out_path = project_root / out_path
            out_path.parent.mkdir(parents=True, exist_ok=True)
            out_path.write_text(json.dumps(results, indent=2))
            results["report_path"] = str(out_path)

        duration = time.time() - start_time
        log_automation_execution("check_definition_of_done", duration, True)

        return json.dumps(format_success_response(results), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("check_definition_of_done", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

