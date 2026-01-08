"""
MCP Tool Wrapper for CI/CD Workflow Validation

Validates GitHub Actions workflows and runner configurations.
"""

import json
import logging
import time
from pathlib import Path
from typing import Any, Optional

import yaml

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


def validate_ci_cd_workflow(
    workflow_path: Optional[str] = None,
    check_runners: bool = True,
    output_path: Optional[str] = None
) -> str:
    """
    Validate CI/CD workflows and runner configurations.

    Args:
        workflow_path: Path to workflow file (default: .github/workflows/parallel-agents-ci.yml)
        check_runners: Whether to validate runner configurations (default: true)
        output_path: Path for validation report (default: docs/CI_CD_VALIDATION_REPORT.md)

    Returns:
        JSON string with validation results
    """
    start_time = time.time()

    try:
        # Import here to avoid circular dependencies
        from project_management_automation.utils import find_project_root
        project_root = find_project_root(Path(__file__).parent.parent.parent)

        # Default paths
        if workflow_path:
            workflow_file = Path(workflow_path)
        else:
            # Try common workflow files
            workflows_dir = project_root / '.github' / 'workflows'
            candidates = ['ci.yml', 'main.yml', 'build.yml', 'parallel-agents-ci.yml']
            workflow_file = None
            for candidate in candidates:
                if (workflows_dir / candidate).exists():
                    workflow_file = workflows_dir / candidate
                    break
            # Fall back to first .yml file found
            if not workflow_file and workflows_dir.exists():
                yml_files = list(workflows_dir.glob('*.yml'))
                if yml_files:
                    workflow_file = yml_files[0]
            # Ultimate fallback
            if not workflow_file:
                workflow_file = workflows_dir / 'ci.yml'
        report_path = Path(output_path) if output_path else project_root / 'docs' / 'CI_CD_VALIDATION_REPORT.md'

        # Validation results
        validation_results = {
            "workflow_file": str(workflow_file),
            "workflow_valid": False,
            "workflow_syntax_errors": [],
            "runner_config_valid": False,
            "runner_issues": [],
            "job_dependencies_valid": False,
            "job_dependency_issues": [],
            "matrix_builds_valid": False,
            "matrix_build_issues": [],
            "triggers_valid": False,
            "trigger_issues": [],
            "artifacts_valid": False,
            "artifact_issues": [],
            "overall_status": "unknown"
        }

        # Check if workflow file exists
        if not workflow_file.exists():
            validation_results["workflow_syntax_errors"].append(f"Workflow file not found: {workflow_file}")
            validation_results["overall_status"] = "failed"

            # Generate report
            _generate_validation_report(validation_results, report_path)

            duration = time.time() - start_time
            log_automation_execution('validate_ci_cd_workflow', duration, False)

            response_data = {
                "workflow_valid": False,
                "runner_config_valid": False,
                "overall_status": "failed",
                "issues": validation_results["workflow_syntax_errors"],
                "report_path": str(report_path.absolute())
            }

            return json.dumps(format_success_response(response_data), indent=2)

        # Load and validate YAML syntax
        try:
            with open(workflow_file) as f:
                workflow_content = yaml.safe_load(f)
            validation_results["workflow_valid"] = True
        except yaml.YAMLError as e:
            validation_results["workflow_syntax_errors"].append(f"YAML syntax error: {str(e)}")
            validation_results["overall_status"] = "failed"

            # Generate report
            _generate_validation_report(validation_results, report_path)

            duration = time.time() - start_time
            log_automation_execution('validate_ci_cd_workflow', duration, False)

            response_data = {
                "workflow_valid": False,
                "runner_config_valid": False,
                "overall_status": "failed",
                "issues": validation_results["workflow_syntax_errors"],
                "report_path": str(report_path.absolute())
            }

            return json.dumps(format_success_response(response_data), indent=2)
        except Exception as e:
            validation_results["workflow_syntax_errors"].append(f"Error loading workflow: {str(e)}")
            validation_results["overall_status"] = "failed"

            # Generate report
            _generate_validation_report(validation_results, report_path)

            duration = time.time() - start_time
            log_automation_execution('validate_ci_cd_workflow', duration, False)

            response_data = {
                "workflow_valid": False,
                "runner_config_valid": False,
                "overall_status": "failed",
                "issues": validation_results["workflow_syntax_errors"],
                "report_path": str(report_path.absolute())
            }

            return json.dumps(format_success_response(response_data), indent=2)

        # Validate workflow structure
        if workflow_content:
            # Validate runner configurations
            if check_runners:
                runner_results = _validate_runner_configs(workflow_content)
                validation_results.update(runner_results)

            # Validate job dependencies
            dep_results = _validate_job_dependencies(workflow_content)
            validation_results.update(dep_results)

            # Validate matrix builds
            matrix_results = _validate_matrix_builds(workflow_content)
            validation_results.update(matrix_results)

            # Validate triggers
            trigger_results = _validate_triggers(workflow_content)
            validation_results.update(trigger_results)

            # Validate artifacts
            artifact_results = _validate_artifacts(workflow_content)
            validation_results.update(artifact_results)

        # Determine overall status
        all_valid = (
            validation_results["workflow_valid"] and
            (not check_runners or validation_results["runner_config_valid"]) and
            validation_results["job_dependencies_valid"] and
            validation_results["matrix_builds_valid"] and
            validation_results["triggers_valid"] and
            validation_results["artifacts_valid"]
        )

        validation_results["overall_status"] = "valid" if all_valid else "issues_found"

        # Generate report
        _generate_validation_report(validation_results, report_path)

        # Format response
        response_data = {
            "workflow_valid": validation_results["workflow_valid"],
            "runner_config_valid": validation_results.get("runner_config_valid", False),
            "job_dependencies_valid": validation_results["job_dependencies_valid"],
            "matrix_builds_valid": validation_results["matrix_builds_valid"],
            "triggers_valid": validation_results["triggers_valid"],
            "artifacts_valid": validation_results["artifacts_valid"],
            "overall_status": validation_results["overall_status"],
            "issues": (
                validation_results["workflow_syntax_errors"] +
                validation_results["runner_issues"] +
                validation_results["job_dependency_issues"] +
                validation_results["matrix_build_issues"] +
                validation_results["trigger_issues"] +
                validation_results["artifact_issues"]
            ),
            "report_path": str(report_path.absolute())
        }

        duration = time.time() - start_time
        log_automation_execution('validate_ci_cd_workflow', duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution('validate_ci_cd_workflow', duration, False, e)

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def _validate_runner_configs(workflow: dict[str, Any]) -> dict[str, Any]:
    """Validate runner configurations in workflow."""
    results = {
        "runner_config_valid": True,
        "runner_issues": []
    }

    jobs = workflow.get("jobs", {})

    for job_name, job_config in jobs.items():
        runs_on = job_config.get("runs-on", "")
        labels = job_config.get("labels", [])

        # Check for self-hosted runners
        if isinstance(runs_on, str) and "self-hosted" in runs_on:
            # Self-hosted runners should have labels
            if not labels and "self-hosted" == runs_on:
                results["runner_issues"].append(
                    f"Job '{job_name}': Self-hosted runner without labels"
                )
                results["runner_config_valid"] = False

        # Check labels format
        if labels:
            if not isinstance(labels, list):
                results["runner_issues"].append(
                    f"Job '{job_name}': Labels must be a list"
                )
                results["runner_config_valid"] = False

    return results


def _validate_job_dependencies(workflow: dict[str, Any]) -> dict[str, Any]:
    """Validate job dependencies."""
    results = {
        "job_dependencies_valid": True,
        "job_dependency_issues": []
    }

    jobs = workflow.get("jobs", {})
    job_names = set(jobs.keys())

    for job_name, job_config in jobs.items():
        needs = job_config.get("needs", [])

        if isinstance(needs, str):
            needs = [needs]

        for dep in needs:
            if dep not in job_names:
                results["job_dependency_issues"].append(
                    f"Job '{job_name}': Depends on non-existent job '{dep}'"
                )
                results["job_dependencies_valid"] = False

    return results


def _validate_matrix_builds(workflow: dict[str, Any]) -> dict[str, Any]:
    """Validate matrix build configurations."""
    results = {
        "matrix_builds_valid": True,
        "matrix_build_issues": []
    }

    jobs = workflow.get("jobs", {})

    for job_name, job_config in jobs.items():
        strategy = job_config.get("strategy", {})
        matrix = strategy.get("matrix", {})

        if matrix:
            # Matrix should have at least one dimension
            if not isinstance(matrix, dict) or len(matrix) == 0:
                results["matrix_build_issues"].append(
                    f"Job '{job_name}': Matrix build has no dimensions"
                )
                results["matrix_builds_valid"] = False

    return results


def _validate_triggers(workflow: dict[str, Any]) -> dict[str, Any]:
    """Validate workflow triggers."""
    results = {
        "triggers_valid": True,
        "trigger_issues": []
    }

    on = workflow.get("on", {})

    if not on:
        results["trigger_issues"].append("Workflow has no triggers defined")
        results["triggers_valid"] = False
        return results

    # Check for valid trigger types
    valid_triggers = ["push", "pull_request", "workflow_dispatch", "schedule", "release", "deployment"]

    for trigger_type in on.keys():
        if trigger_type not in valid_triggers and not trigger_type.startswith("workflow_call"):
            results["trigger_issues"].append(f"Unknown trigger type: {trigger_type}")

    return results


def _validate_artifacts(workflow: dict[str, Any]) -> dict[str, Any]:
    """Validate artifact uploads/downloads."""
    results = {
        "artifacts_valid": True,
        "artifact_issues": []
    }

    jobs = workflow.get("jobs", {})

    for job_name, job_config in jobs.items():
        steps = job_config.get("steps", [])

        artifact_names = set()

        for step in steps:
            # Check for artifact uploads
            if "uses" in step and "upload-artifact" in step["uses"]:
                artifact_name = step.get("with", {}).get("name", "")
                if artifact_name:
                    artifact_names.add(artifact_name)

            # Check for artifact downloads
            if "uses" in step and "download-artifact" in step["uses"]:
                artifact_name = step.get("with", {}).get("name", "")
                if artifact_name and artifact_name not in artifact_names:
                    # Check if artifact is uploaded in a previous job
                    results["artifact_issues"].append(
                        f"Job '{job_name}': Downloads artifact '{artifact_name}' that may not exist"
                    )

    return results


def _generate_validation_report(results: dict[str, Any], report_path: Path) -> str:
    """Generate validation report markdown."""
    report_lines = [
        "# CI/CD Workflow Validation Report",
        "",
        f"*Generated: {time.strftime('%Y-%m-%d %H:%M:%S')}*",
        "",
        "## Executive Summary",
        "",
        f"**Workflow File:** `{results['workflow_file']}`",
        f"**Overall Status:** {results['overall_status'].upper()}",
        "",
        "## Validation Results",
        "",
        f"- ✅ **Workflow Syntax:** {'Valid' if results['workflow_valid'] else '❌ Invalid'}",
    ]

    if results.get('runner_config_valid') is not None:
        report_lines.append(
            f"- {'✅' if results['runner_config_valid'] else '❌'} **Runner Config:** {'Valid' if results['runner_config_valid'] else 'Issues Found'}"
        )

    report_lines.extend([
        f"- {'✅' if results['job_dependencies_valid'] else '❌'} **Job Dependencies:** {'Valid' if results['job_dependencies_valid'] else 'Issues Found'}",
        f"- {'✅' if results['matrix_builds_valid'] else '❌'} **Matrix Builds:** {'Valid' if results['matrix_builds_valid'] else 'Issues Found'}",
        f"- {'✅' if results['triggers_valid'] else '❌'} **Triggers:** {'Valid' if results['triggers_valid'] else 'Issues Found'}",
        f"- {'✅' if results['artifacts_valid'] else '❌'} **Artifacts:** {'Valid' if results['artifacts_valid'] else 'Issues Found'}",
        "",
    ])

    # Add issues sections
    all_issues = (
        results.get('workflow_syntax_errors', []) +
        results.get('runner_issues', []) +
        results.get('job_dependency_issues', []) +
        results.get('matrix_build_issues', []) +
        results.get('trigger_issues', []) +
        results.get('artifact_issues', [])
    )

    if all_issues:
        report_lines.extend([
            "## Issues Found",
            ""
        ])

        for issue in all_issues:
            report_lines.append(f"- {issue}")

        report_lines.append("")
    else:
        report_lines.extend([
            "## Issues",
            "",
            "✅ No issues found. Workflow configuration is valid.",
            ""
        ])

    report_content = "\n".join(report_lines)

    # Write report
    report_path.parent.mkdir(parents=True, exist_ok=True)
    with open(report_path, 'w') as f:
        f.write(report_content)

    return report_content
