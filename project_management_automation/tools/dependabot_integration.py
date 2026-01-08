"""
Dependabot integration for exarp security scanning.

Fetches and compares Dependabot alerts with local pip-audit results.
Provides a unified view of all security vulnerabilities.

Usage:
    from project_management_automation.tools.dependabot_integration import (
        fetch_dependabot_alerts,
        compare_security_findings,
        get_unified_security_report,
    )
"""

import json
import subprocess
from dataclasses import dataclass
from datetime import datetime
from typing import Any, Optional


@dataclass
class SecurityVulnerability:
    """Unified vulnerability representation."""

    package: str
    severity: str  # critical, high, medium, low
    cve: Optional[str]
    source: str  # dependabot, pip-audit, cargo-audit, npm-audit
    state: str  # open, fixed, dismissed
    ecosystem: str  # python, rust, npm
    description: Optional[str] = None
    fix_available: bool = False
    fixed_version: Optional[str] = None


def fetch_dependabot_alerts(
    repo: str = "davidl71/project-management-automation",
    state: str = "open",
) -> dict[str, Any]:
    """
    Fetch Dependabot alerts from GitHub API.

    Requires: `gh` CLI installed and authenticated.

    Args:
        repo: GitHub repo in owner/repo format
        state: Alert state filter (open, fixed, dismissed, all)

    Returns:
        Dict with alerts, counts by severity, and status
    """
    try:
        # Build jq query for structured output
        jq_query = '.[] | {package: .security_vulnerability.package.name, severity: .security_vulnerability.severity, cve: .security_advisory.cve_id, state: .state, ecosystem: .security_vulnerability.package.ecosystem, description: .security_advisory.summary, fix_available: .security_vulnerability.first_patched_version != null, fixed_version: .security_vulnerability.first_patched_version.identifier}'

        result = subprocess.run(
            ["gh", "api", f"repos/{repo}/dependabot/alerts", "--jq", jq_query],
            capture_output=True,
            text=True,
            timeout=30,
        )

        if result.returncode != 0:
            return {
                "success": False,
                "error": result.stderr or "gh CLI failed",
                "hint": "Install gh CLI: brew install gh && gh auth login",
            }

        # Parse JSONL output
        alerts = []
        for line in result.stdout.strip().split("\n"):
            if line:
                try:
                    alert = json.loads(line)
                    alerts.append(alert)
                except json.JSONDecodeError:
                    continue

        # Filter by state if not "all"
        if state != "all":
            alerts = [a for a in alerts if a.get("state") == state]

        # Count by severity
        by_severity = {"critical": 0, "high": 0, "medium": 0, "low": 0}
        by_ecosystem = {}

        for alert in alerts:
            sev = alert.get("severity", "unknown").lower()
            if sev in by_severity:
                by_severity[sev] += 1

            eco = alert.get("ecosystem", "unknown")
            by_ecosystem[eco] = by_ecosystem.get(eco, 0) + 1

        return {
            "success": True,
            "total_alerts": len(alerts),
            "by_severity": by_severity,
            "by_ecosystem": by_ecosystem,
            "alerts": alerts,
            "repo": repo,
            "state_filter": state,
        }

    except FileNotFoundError:
        return {
            "success": False,
            "error": "gh CLI not found",
            "hint": "Install: brew install gh && gh auth login",
        }
    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "error": "GitHub API request timed out",
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e),
        }


def run_pip_audit() -> dict[str, Any]:
    """
    Run pip-audit on current environment.

    Returns:
        Dict with vulnerabilities found
    """
    try:
        result = subprocess.run(
            ["pip-audit", "--format=json"],
            capture_output=True,
            text=True,
            timeout=120,
        )

        data = json.loads(result.stdout)

        vulnerabilities = []
        for dep in data.get("dependencies", []):
            for vuln in dep.get("vulns", []):
                vulnerabilities.append({
                    "package": dep["name"],
                    "version": dep["version"],
                    "vuln_id": vuln.get("id"),
                    "fix_version": vuln.get("fix_versions", [None])[0],
                })

        return {
            "success": True,
            "total_vulnerabilities": len(vulnerabilities),
            "vulnerabilities": vulnerabilities,
            "total_packages": len(data.get("dependencies", [])),
        }

    except FileNotFoundError:
        return {
            "success": False,
            "error": "pip-audit not found",
            "hint": "Install: pip install pip-audit",
        }
    except json.JSONDecodeError as e:
        return {
            "success": False,
            "error": f"Failed to parse pip-audit output: {e}",
        }
    except subprocess.TimeoutExpired:
        return {
            "success": False,
            "error": "pip-audit timed out",
        }


def compare_security_findings(
    dependabot: dict[str, Any],
    pip_audit: dict[str, Any],
) -> dict[str, Any]:
    """
    Compare Dependabot and pip-audit findings.

    Returns:
        Dict with comparison results and discrepancies
    """
    comparison = {
        "dependabot_total": dependabot.get("total_alerts", 0),
        "pip_audit_total": pip_audit.get("total_vulnerabilities", 0),
        "discrepancies": [],
        "only_in_dependabot": [],
        "only_in_pip_audit": [],
        "in_both": [],
    }

    # Get Python-only Dependabot alerts
    dependabot_python = [
        a for a in dependabot.get("alerts", [])
        if a.get("ecosystem") == "pip"
    ]

    pip_packages = {v["package"] for v in pip_audit.get("vulnerabilities", [])}
    dependabot_packages = {a["package"] for a in dependabot_python}

    comparison["only_in_dependabot"] = list(dependabot_packages - pip_packages)
    comparison["only_in_pip_audit"] = list(pip_packages - dependabot_packages)
    comparison["in_both"] = list(dependabot_packages & pip_packages)

    # Note discrepancies
    if comparison["only_in_dependabot"]:
        comparison["discrepancies"].append({
            "type": "dependabot_only",
            "note": "These packages have alerts in Dependabot but not in pip-audit (may be outdated requirements.txt)",
            "packages": comparison["only_in_dependabot"],
        })

    if comparison["only_in_pip_audit"]:
        comparison["discrepancies"].append({
            "type": "pip_audit_only",
            "note": "These packages have vulnerabilities in pip-audit but not in Dependabot",
            "packages": comparison["only_in_pip_audit"],
        })

    return comparison


def get_unified_security_report(
    repo: str = "davidl71/project-management-automation",
    include_dismissed: bool = False,
) -> dict[str, Any]:
    """
    Get unified security report combining Dependabot and local scans.

    Returns:
        Comprehensive security report
    """
    state = "all" if include_dismissed else "open"

    # Fetch from both sources
    dependabot = fetch_dependabot_alerts(repo, state)
    pip_audit = run_pip_audit()

    # Compare findings
    comparison = compare_security_findings(dependabot, pip_audit)

    # Build unified report
    report = {
        "timestamp": datetime.now().isoformat(),
        "repo": repo,
        "summary": {
            "total_issues": (
                dependabot.get("total_alerts", 0) +
                pip_audit.get("total_vulnerabilities", 0)
            ),
            "dependabot_alerts": dependabot.get("total_alerts", 0),
            "pip_audit_vulnerabilities": pip_audit.get("total_vulnerabilities", 0),
            "by_severity": dependabot.get("by_severity", {}),
            "by_ecosystem": dependabot.get("by_ecosystem", {}),
        },
        "dependabot": dependabot,
        "pip_audit": pip_audit,
        "comparison": comparison,
        "recommendations": [],
    }

    # Add recommendations
    if dependabot.get("by_severity", {}).get("critical", 0) > 0:
        report["recommendations"].append({
            "priority": "critical",
            "action": "Fix critical vulnerabilities immediately",
            "packages": [
                a["package"] for a in dependabot.get("alerts", [])
                if a.get("severity") == "critical"
            ],
        })

    if dependabot.get("by_ecosystem", {}).get("npm", 0) > 0:
        report["recommendations"].append({
            "priority": "high",
            "action": "Review npm dependencies - may be accidentally committed files",
            "hint": "Check if node_modules or package-lock.json should be gitignored",
        })

    if comparison.get("only_in_dependabot"):
        report["recommendations"].append({
            "priority": "medium",
            "action": "Sync requirements.txt with installed packages",
            "hint": "pip freeze > requirements.txt",
        })

    return report


def dismiss_dependabot_alert(
    alert_number: int,
    reason: str = "tolerable_risk",
    comment: Optional[str] = None,
    repo: str = "davidl71/project-management-automation",
) -> dict[str, Any]:
    """
    Dismiss a Dependabot alert.

    Args:
        alert_number: The alert number to dismiss
        reason: One of: fix_started, inaccurate, no_bandwidth, not_used, tolerable_risk
        comment: Optional comment explaining the dismissal
        repo: GitHub repo

    Returns:
        Result of the dismissal
    """
    try:
        data = {"dismissed_reason": reason}
        if comment:
            data["dismissed_comment"] = comment

        result = subprocess.run(
            [
                "gh", "api",
                "-X", "PATCH",
                f"repos/{repo}/dependabot/alerts/{alert_number}",
                "-f", "state=dismissed",
                "-f", f"dismissed_reason={reason}",
            ] + (["-f", f"dismissed_comment={comment}"] if comment else []),
            capture_output=True,
            text=True,
            timeout=30,
        )

        if result.returncode != 0:
            return {"success": False, "error": result.stderr}

        return {"success": True, "alert_number": alert_number, "dismissed": True}

    except Exception as e:
        return {"success": False, "error": str(e)}

