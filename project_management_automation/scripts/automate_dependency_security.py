#!/usr/bin/env python3
"""
Dependency Security Scan Automation

Scans Python, Rust, and npm dependencies for known vulnerabilities.
Uses osv-scanner, pip-audit, cargo-audit, and npm audit.

Usage:
    python3 scripts/automate_dependency_security.py [--config config.json] [--dry-run]
"""

import argparse
import json
import logging
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

# Add project root to path
# Project root will be passed to __init__
# Import base class
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase

# Configure logging (will be configured after project_root is set)
logger = logging.getLogger(__name__)


class DependencySecurityAnalyzer(IntelligentAutomationBase):
    """Intelligent dependency security scanning automation."""

    def __init__(self, config_path: str, project_root: Optional[Path] = None):
        with open(config_path) as f:
            config = json.load(f)
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Dependency Security Scan", project_root)
        self.config = config
        self.scan_configs = config.get('scan_configs', {})
        self.severity_levels = config.get('severity_levels', {})
        self.output_file = self.project_root / config.get('output_file', 'docs/DEPENDENCY_SECURITY_REPORT.md')
        self.history_file = self.project_root / config.get('trend_tracking', {}).get('history_file', 'scripts/.dependency_security_history.json')
        self.create_tasks_config = config.get('create_todo2_tasks', {})

        # Scan results
        self.scan_results = {
            'python': [],
            'rust': [],
            'npm': [],
            'summary': {
                'total_vulnerabilities': 0,
                'by_severity': {},
                'by_language': {},
                'critical_vulnerabilities': []
            }
        }

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is dependency security?"""
        return "What is dependency security? Dependency Security = Vulnerability Detection × Severity Assessment × Update Tracking × Risk Mitigation"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we scan dependencies for security?"""
        return "How do we systematically scan Python, Rust, and npm dependencies for known vulnerabilities?"

    def _execute_analysis(self) -> dict:
        """Execute dependency security scanning."""
        logger.info("Starting dependency security scan...")

        # Scan each language
        if self.scan_configs.get('python', {}).get('enabled', False):
            self._scan_python()

        if self.scan_configs.get('rust', {}).get('enabled', False):
            self._scan_rust()

        if self.scan_configs.get('npm', {}).get('enabled', False):
            self._scan_npm()

        # Generate summary
        self._generate_summary()

        # Track trends
        if self.config.get('trend_tracking', {}).get('enabled', False):
            self._track_trends()

        return self.scan_results

    def _scan_python(self):
        """Scan Python dependencies."""
        logger.info("Scanning Python dependencies...")
        python_config = self.scan_configs.get('python', {})
        files = python_config.get('files', [])
        tools = python_config.get('tools', {})

        vulnerabilities = []

        # Try osv-scanner (already in Trunk config)
        if tools.get('osv_scanner', {}).get('enabled', False):
            vulns = self._run_osv_scanner(files)
            vulnerabilities.extend(vulns)

        # Try pip-audit
        if tools.get('pip_audit', {}).get('enabled', False):
            vulns = self._run_pip_audit()
            vulnerabilities.extend(vulns)

        self.scan_results['python'] = vulnerabilities
        logger.info(f"Found {len(vulnerabilities)} Python vulnerabilities")

    def _scan_rust(self):
        """Scan Rust dependencies."""
        logger.info("Scanning Rust dependencies...")
        rust_config = self.scan_configs.get('rust', {})
        files = rust_config.get('files', [])
        tools = rust_config.get('tools', {})

        vulnerabilities = []

        # Try cargo-audit
        if tools.get('cargo_audit', {}).get('enabled', False):
            vulns = self._run_cargo_audit()
            vulnerabilities.extend(vulns)

        # Try osv-scanner
        if tools.get('osv_scanner', {}).get('enabled', False):
            vulns = self._run_osv_scanner(files)
            vulnerabilities.extend(vulns)

        self.scan_results['rust'] = vulnerabilities
        logger.info(f"Found {len(vulnerabilities)} Rust vulnerabilities")

    def _scan_npm(self):
        """Scan npm dependencies."""
        logger.info("Scanning npm dependencies...")
        npm_config = self.scan_configs.get('npm', {})
        files = npm_config.get('files', [])
        tools = npm_config.get('tools', {})

        vulnerabilities = []

        # Try npm audit
        if tools.get('npm_audit', {}).get('enabled', False):
            vulns = self._run_npm_audit()
            vulnerabilities.extend(vulns)

        # Try osv-scanner
        if tools.get('osv_scanner', {}).get('enabled', False):
            vulns = self._run_osv_scanner(files)
            vulnerabilities.extend(vulns)

        self.scan_results['npm'] = vulnerabilities
        logger.info(f"Found {len(vulnerabilities)} npm vulnerabilities")

    def _run_osv_scanner(self, files: list[str]) -> list[dict]:
        """Run osv-scanner on specified files."""
        vulnerabilities = []

        if not self._command_exists('osv-scanner'):
            logger.warning("osv-scanner not found. Install: brew install osv-scanner or trunk install osv-scanner")
            return vulnerabilities

        try:
            for file_path in files:
                full_path = self.project_root / file_path
                if not full_path.exists():
                    logger.debug(f"Skipping {file_path} (not found)")
                    continue

                cmd = ['osv-scanner', '--format', 'json', str(full_path)]
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)

                if result.returncode == 0:
                    try:
                        data = json.loads(result.stdout)
                        # Parse osv-scanner JSON output
                        if 'results' in data:
                            for result_item in data['results']:
                                if 'packages' in result_item:
                                    for pkg in result_item['packages']:
                                        if 'vulnerabilities' in pkg:
                                            for vuln in pkg['vulnerabilities']:
                                                vulnerabilities.append({
                                                    'package': pkg.get('package', {}).get('name', 'unknown'),
                                                    'version': pkg.get('package', {}).get('version', 'unknown'),
                                                    'vulnerability': vuln.get('id', 'unknown'),
                                                    'severity': self._extract_severity(vuln),
                                                    'summary': vuln.get('summary', ''),
                                                    'source': 'osv-scanner',
                                                    'file': file_path
                                                })
                    except json.JSONDecodeError:
                        logger.warning(f"Failed to parse osv-scanner output for {file_path}")
                else:
                    logger.debug(f"osv-scanner returned non-zero for {file_path}: {result.stderr}")

        except subprocess.TimeoutExpired:
            logger.error("osv-scanner timed out")
        except Exception as e:
            logger.error(f"Error running osv-scanner: {e}")

        return vulnerabilities

    def _run_pip_audit(self) -> list[dict]:
        """Run pip-audit on Python dependencies."""
        vulnerabilities = []

        if not self._command_exists('pip-audit'):
            logger.warning("pip-audit not found. Install: pip install pip-audit")
            return vulnerabilities

        try:
            req_file = self.project_root / 'requirements.txt'
            if not req_file.exists():
                logger.debug("requirements.txt not found")
                return vulnerabilities

            cmd = ['pip-audit', '--format', 'json', '--desc', '--requirement', str(req_file)]
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=300)

            if result.returncode == 0:
                try:
                    data = json.loads(result.stdout)
                    if 'vulnerabilities' in data:
                        for vuln in data['vulnerabilities']:
                            vulnerabilities.append({
                                'package': vuln.get('name', 'unknown'),
                                'version': vuln.get('installed_version', 'unknown'),
                                'vulnerability': vuln.get('id', 'unknown'),
                                'severity': self._extract_severity(vuln),
                                'summary': vuln.get('description', ''),
                                'source': 'pip-audit',
                                'file': 'requirements.txt'
                            })
                except json.JSONDecodeError:
                    logger.warning("Failed to parse pip-audit output")
            else:
                logger.debug(f"pip-audit returned non-zero: {result.stderr}")

        except subprocess.TimeoutExpired:
            logger.error("pip-audit timed out")
        except Exception as e:
            logger.error(f"Error running pip-audit: {e}")

        return vulnerabilities

    def _run_cargo_audit(self) -> list[dict]:
        """Run cargo-audit on Rust dependencies."""
        vulnerabilities = []

        if not self._command_exists('cargo-audit'):
            logger.warning("cargo-audit not found. Install: cargo install cargo-audit")
            return vulnerabilities

        try:
            cargo_toml = self.project_root / 'agents' / 'backend' / 'Cargo.toml'
            if not cargo_toml.exists():
                logger.debug("Cargo.toml not found")
                return vulnerabilities

            # Run from backend directory
            cmd = ['cargo', 'audit', '--json']
            result = subprocess.run(
                cmd,
                cwd=cargo_toml.parent,
                capture_output=True,
                text=True,
                timeout=300
            )

            if result.returncode == 0 or result.returncode == 1:  # cargo-audit returns 1 on vulnerabilities
                try:
                    # cargo-audit outputs line-delimited JSON
                    for line in result.stdout.split('\n'):
                        if not line.strip():
                            continue
                        try:
                            data = json.loads(line)
                            if data.get('type') == 'vulnerability':
                                vulnerabilities.append({
                                    'package': data.get('package', {}).get('name', 'unknown'),
                                    'version': data.get('package', {}).get('version', 'unknown'),
                                    'vulnerability': data.get('advisory', {}).get('id', 'unknown'),
                                    'severity': self._extract_severity(data.get('advisory', {})),
                                    'summary': data.get('advisory', {}).get('title', ''),
                                    'source': 'cargo-audit',
                                    'file': 'agents/backend/Cargo.toml'
                                })
                        except json.JSONDecodeError:
                            continue
                except Exception as e:
                    logger.warning(f"Failed to parse cargo-audit output: {e}")
            else:
                logger.debug(f"cargo-audit returned {result.returncode}: {result.stderr}")

        except subprocess.TimeoutExpired:
            logger.error("cargo-audit timed out")
        except Exception as e:
            logger.error(f"Error running cargo-audit: {e}")

        return vulnerabilities

    def _run_npm_audit(self) -> list[dict]:
        """Run npm audit on npm dependencies."""
        vulnerabilities = []

        if not self._command_exists('npm'):
            logger.warning("npm not found")
            return vulnerabilities

        try:
            package_json = self.project_root / 'web' / 'package.json'
            if not package_json.exists():
                logger.debug("package.json not found")
                return vulnerabilities

            # Run from web directory
            cmd = ['npm', 'audit', '--json']
            result = subprocess.run(
                cmd,
                cwd=package_json.parent,
                capture_output=True,
                text=True,
                timeout=300
            )

            # npm audit returns non-zero on vulnerabilities, but still outputs JSON
            try:
                data = json.loads(result.stdout)
                if 'vulnerabilities' in data:
                    for name, vuln_data in data['vulnerabilities'].items():
                        if isinstance(vuln_data, dict) and 'via' in vuln_data:
                            for via_item in vuln_data['via']:
                                if isinstance(via_item, dict) and 'vulnerability' in via_item:
                                    vulnerabilities.append({
                                        'package': name,
                                        'version': vuln_data.get('version', 'unknown'),
                                        'vulnerability': via_item.get('vulnerability', 'unknown'),
                                        'severity': self._extract_severity(via_item),
                                        'summary': via_item.get('title', ''),
                                        'source': 'npm-audit',
                                        'file': 'web/package.json'
                                    })
            except json.JSONDecodeError:
                logger.warning("Failed to parse npm audit output")

        except subprocess.TimeoutExpired:
            logger.error("npm audit timed out")
        except Exception as e:
            logger.error(f"Error running npm audit: {e}")

        return vulnerabilities

    def _command_exists(self, command: str) -> bool:
        """Check if a command exists."""
        try:
            result = subprocess.run(
                ['which', command],
                capture_output=True,
                timeout=5
            )
            return result.returncode == 0
        except Exception:
            return False

    def _extract_severity(self, vuln_data: dict) -> str:
        """Extract severity from vulnerability data."""
        # Try various severity fields
        severity = (
            vuln_data.get('severity') or
            vuln_data.get('severity') or
            vuln_data.get('database_specific', {}).get('severity') or
            'unknown'
        )

        # Normalize severity
        severity_lower = str(severity).lower()
        if 'critical' in severity_lower:
            return 'critical'
        elif 'high' in severity_lower:
            return 'high'
        elif 'medium' in severity_lower:
            return 'medium'
        elif 'low' in severity_lower:
            return 'low'
        else:
            return 'info'

    def _generate_summary(self):
        """Generate summary statistics."""
        summary = self.scan_results['summary']

        # Count by severity
        all_vulns = (
            self.scan_results['python'] +
            self.scan_results['rust'] +
            self.scan_results['npm']
        )

        summary['total_vulnerabilities'] = len(all_vulns)

        # Count by severity
        for vuln in all_vulns:
            severity = vuln.get('severity', 'unknown')
            summary['by_severity'][severity] = summary['by_severity'].get(severity, 0) + 1

            if severity == 'critical':
                summary['critical_vulnerabilities'].append(vuln)

        # Count by language
        summary['by_language'] = {
            'python': len(self.scan_results['python']),
            'rust': len(self.scan_results['rust']),
            'npm': len(self.scan_results['npm'])
        }

    def _track_trends(self):
        """Track vulnerability trends over time."""
        if not self.history_file.parent.exists():
            self.history_file.parent.mkdir(parents=True, exist_ok=True)

        # Load existing history
        history = []
        if self.history_file.exists():
            try:
                with open(self.history_file) as f:
                    history = json.load(f)
            except Exception as e:
                logger.warning(f"Failed to load history: {e}")

        # Add current scan
        history.append({
            'timestamp': datetime.now(timezone.utc).isoformat(),
            'summary': self.scan_results['summary']
        })

        # Keep only last N days
        max_days = self.config.get('trend_tracking', {}).get('max_history_days', 90)
        cutoff_date = datetime.now(timezone.utc).timestamp() - (max_days * 24 * 60 * 60)
        history = [
            h for h in history
            if datetime.fromisoformat(h['timestamp'].replace('Z', '+00:00')).timestamp() > cutoff_date
        ]

        # Save history
        try:
            with open(self.history_file, 'w') as f:
                json.dump(history, f, indent=2)
        except Exception as e:
            logger.error(f"Failed to save history: {e}")

    def _generate_insights(self, results: dict) -> str:
        """Generate insights from scan results."""
        summary = results.get('summary', {
            'total_vulnerabilities': 0,
            'critical_vulnerabilities': [],
            'by_severity': {},
            'by_language': {}
        })
        total = summary.get('total_vulnerabilities', 0)
        critical = len(summary.get('critical_vulnerabilities', []))

        insights = []

        if total == 0:
            insights.append("✅ No vulnerabilities found in dependencies!")
        else:
            insights.append(f"⚠️ Found {total} total vulnerabilities")

            if critical > 0:
                insights.append(f"🚨 {critical} CRITICAL vulnerabilities require immediate attention")

            # Language breakdown
            by_lang = summary['by_language']
            for lang, count in by_lang.items():
                if count > 0:
                    insights.append(f"  - {lang}: {count} vulnerabilities")

            # Severity breakdown
            by_sev = summary.get('by_severity', {})
            for sev in ['critical', 'high', 'medium', 'low']:
                count = by_sev.get(sev, 0)
                if count > 0:
                    insights.append(f"  - {sev}: {count} vulnerabilities")

            # Language breakdown
            by_lang = summary.get('by_language', {})
            for lang, count in by_lang.items():
                if count > 0:
                    insights.append(f"  - {lang}: {count} vulnerabilities")

        return "\n".join(insights)

    def _generate_report(self, results: dict, insights: str) -> str:
        """Generate markdown report (required by base class)."""
        return self._generate_report_document(results, insights)

    def _generate_report_document(self, results: dict, insights: str) -> str:
        """Generate markdown report."""
        summary = results.get('summary', {
            'total_vulnerabilities': 0,
            'critical_vulnerabilities': [],
            'by_severity': {},
            'by_language': {}
        })

        report = f"""# Dependency Security Report

**Generated**: {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M:%S UTC')}
**Automation**: Dependency Security Scan

---

## Executive Summary

{insights}

---

## Summary Statistics

- **Total Vulnerabilities**: {summary['total_vulnerabilities']}
- **Critical**: {len(summary['critical_vulnerabilities'])}
- **High**: {summary['by_severity'].get('high', 0)}
- **Medium**: {summary['by_severity'].get('medium', 0)}
- **Low**: {summary['by_severity'].get('low', 0)}

### By Language

"""
        for lang, count in summary['by_language'].items():
            report += f"- **{lang}**: {count} vulnerabilities\n"

        report += "\n---\n\n## Critical Vulnerabilities\n\n"

        if summary['critical_vulnerabilities']:
            for vuln in summary['critical_vulnerabilities'][:10]:  # Limit to 10
                report += f"### {vuln.get('package', 'unknown')} ({vuln.get('version', 'unknown')})\n\n"
                report += f"- **Vulnerability**: {vuln.get('vulnerability', 'unknown')}\n"
                report += f"- **Severity**: {vuln.get('severity', 'unknown')}\n"
                report += f"- **Source**: {vuln.get('source', 'unknown')}\n"
                report += f"- **File**: {vuln.get('file', 'unknown')}\n"
                if vuln.get('summary'):
                    report += f"- **Description**: {vuln.get('summary')}\n"
                report += "\n"
        else:
            report += "No critical vulnerabilities found.\n\n"

        report += "\n---\n\n## All Vulnerabilities by Language\n\n"

        # Python
        python_vulns = results.get('python', [])
        if python_vulns:
            report += "### Python\n\n"
            for vuln in python_vulns[:20]:  # Limit to 20 per language
                report += f"- **{vuln.get('package', 'unknown')}** ({vuln.get('version', 'unknown')}): {vuln.get('vulnerability', 'unknown')} [{vuln.get('severity', 'unknown')}]\n"
            if len(python_vulns) > 20:
                report += f"\n*... and {len(python_vulns) - 20} more*\n"
            report += "\n"

        # Rust
        rust_vulns = results.get('rust', [])
        if rust_vulns:
            report += "### Rust\n\n"
            for vuln in rust_vulns[:20]:
                report += f"- **{vuln.get('package', 'unknown')}** ({vuln.get('version', 'unknown')}): {vuln.get('vulnerability', 'unknown')} [{vuln.get('severity', 'unknown')}]\n"
            if len(rust_vulns) > 20:
                report += f"\n*... and {len(rust_vulns) - 20} more*\n"
            report += "\n"

        # npm
        npm_vulns = results.get('npm', [])
        if npm_vulns:
            report += "### npm\n\n"
            for vuln in npm_vulns[:20]:
                report += f"- **{vuln.get('package', 'unknown')}** ({vuln.get('version', 'unknown')}): {vuln.get('vulnerability', 'unknown')} [{vuln.get('severity', 'unknown')}]\n"
            if len(npm_vulns) > 20:
                report += f"\n*... and {len(npm_vulns) - 20} more*\n"
            report += "\n"

        report += "\n---\n\n## Recommendations\n\n"

        if summary['total_vulnerabilities'] == 0:
            report += "✅ All dependencies are secure. Continue regular scanning.\n\n"
        else:
            report += "1. **Review critical vulnerabilities immediately**\n"
            report += "2. **Update vulnerable packages** to patched versions\n"
            report += "3. **Check for alternative packages** if updates are unavailable\n"
            report += "4. **Monitor vulnerability databases** for new advisories\n"
            report += "5. **Set up automated scanning** (this automation) to catch issues early\n\n"

        report += "\n---\n\n*This report was automatically generated by the Dependency Security Scan automation.*\n"

        return report

    def _create_follow_up_tasks(self, results: dict) -> list[dict]:
        """Create Todo2 tasks for critical/high vulnerabilities."""
        if not self.create_tasks_config.get('enabled', False):
            return []

        tasks = []
        results['summary']
        min_severity = self.create_tasks_config.get('min_severity', 'high')
        max_tasks = self.create_tasks_config.get('max_tasks', 10)

        # Get vulnerabilities to create tasks for
        all_vulns = (
            results.get('python', []) +
            results.get('rust', []) +
            results.get('npm', [])
        )

        severity_order = {'critical': 4, 'high': 3, 'medium': 2, 'low': 1, 'info': 0}
        min_severity_level = severity_order.get(min_severity, 3)

        # Filter and sort
        filtered_vulns = [
            v for v in all_vulns
            if severity_order.get(v.get('severity', 'info'), 0) >= min_severity_level
        ]
        filtered_vulns.sort(key=lambda v: severity_order.get(v.get('severity', 'info'), 0), reverse=True)

        # Create tasks (limit to max_tasks)
        for vuln in filtered_vulns[:max_tasks]:
            tasks.append({
                'name': f"Fix {vuln.get('severity', 'unknown').upper()} vulnerability: {vuln.get('package', 'unknown')}",
                'description': f"Vulnerability: {vuln.get('vulnerability', 'unknown')}\nPackage: {vuln.get('package', 'unknown')} ({vuln.get('version', 'unknown')})\nFile: {vuln.get('file', 'unknown')}\n\n{vuln.get('summary', '')}",
                'priority': 'high' if vuln.get('severity') in ['critical', 'high'] else 'medium',
                'tags': ['security', 'dependencies', vuln.get('severity', 'unknown'), vuln.get('source', 'unknown')]
            })

        return tasks


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Dependency Security Scan Automation')
    parser.add_argument('--config', default='scripts/dependency_security_config.json',
                       help='Path to configuration file')
    parser.add_argument('--dry-run', action='store_true',
                       help='Run in dry-run mode (no file changes)')

    args = parser.parse_args()

    # Load config
    config_path = Path(args.config)
    if not config_path.exists():
        logger.error(f"Configuration file not found: {config_path}")
        sys.exit(1)

    # Create analyzer
    analyzer = DependencySecurityAnalyzer(str(config_path))

    # Run automation
    try:
        results = analyzer.run()

        # Write report
        report = analyzer._generate_report_document(results, analyzer._generate_insights(results))
        analyzer.output_file.parent.mkdir(parents=True, exist_ok=True)
        with open(analyzer.output_file, 'w') as f:
            f.write(report)

        logger.info(f"Report written to {analyzer.output_file}")
        summary = results.get('summary', {})
        total = summary.get('total_vulnerabilities', 0)
        logger.info(f"Found {total} total vulnerabilities")

        return 0
    except Exception as e:
        logger.error(f"Automation failed: {e}", exc_info=True)
        return 1


if __name__ == '__main__':
    sys.exit(main())
