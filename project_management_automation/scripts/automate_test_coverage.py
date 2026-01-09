#!/usr/bin/env python3
"""
Automated Test Coverage Analysis Script

Generates coverage reports and identifies gaps in test coverage.
"""

import argparse
import json
import logging
import subprocess
import sys
import xml.etree.ElementTree as ET
from datetime import datetime
from pathlib import Path
from typing import Optional

# Import base class
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase

logger = logging.getLogger(__name__)


class TestCoverageAnalyzer(IntelligentAutomationBase):
    """Test coverage analyzer using intelligent automation base."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Test Coverage Analyzer", project_root)

        self.coverage_file = config.get('coverage_file')
        self.min_coverage = config.get('min_coverage', 80)
        self.output_path = Path(config.get('output_path', 'coverage-report/'))
        self.format = config.get('format', 'html')

        # Ensure output directory exists
        self.output_path.mkdir(parents=True, exist_ok=True)

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is test coverage?"""
        return "What is test coverage? Test Coverage = Lines Covered × Functions Covered × Branches Covered × Statements Covered"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we analyze coverage?"""
        return "How do we systematically analyze test coverage and identify gaps?"

    def _detect_coverage_file(self) -> Optional[Path]:
        """Auto-detect coverage file."""
        if self.coverage_file:
            return Path(self.coverage_file)

        # Look for common coverage files
        candidates = [
            self.output_path / 'coverage.xml',
            self.project_root / '.coverage',
            self.project_root / 'coverage.xml',
            self.project_root / 'coverage.json'
        ]

        for candidate in candidates:
            if candidate.exists():
                return candidate

        return None

    def _execute_analysis(self) -> dict:
        """Execute coverage analysis."""
        logger.info("Analyzing test coverage...")

        coverage_file = self._detect_coverage_file()

        if not coverage_file:
            # Try to generate coverage first
            logger.info("No coverage file found, attempting to generate...")
            self._generate_coverage()
            coverage_file = self._detect_coverage_file()

        if not coverage_file:
            return {
                'status': 'error',
                'error': 'No coverage file found and unable to generate',
                'coverage_data': None
            }

        # Parse coverage based on format
        if coverage_file.suffix == '.xml':
            coverage_data = self._parse_coverage_xml(coverage_file)
        elif coverage_file.suffix == '.json':
            coverage_data = self._parse_coverage_json(coverage_file)
        else:
            # Try coverage.py report
            coverage_data = self._parse_coverage_py(coverage_file)

        # Generate report
        report_path = self._generate_coverage_report_file(coverage_data)

        # Identify gaps
        gaps = self._identify_gaps(coverage_data)

        return {
            'status': 'success',
            'coverage_file': str(coverage_file),
            'coverage_data': coverage_data,
            'report_path': str(report_path),
            'gaps': gaps,
            'meets_threshold': coverage_data.get('total_coverage', 0) >= self.min_coverage
        }

    def _check_pytest_cov_installed(self) -> bool:
        """Check if pytest-cov is installed."""
        try:
            import pytest_cov
            return True
        except ImportError:
            return False

    def _generate_coverage(self):
        """Generate coverage by running tests with coverage."""
        # Check if pytest-cov is installed
        if not self._check_pytest_cov_installed():
            error_msg = (
                "pytest-cov is not installed. Install it with:\n"
                "  pip install pytest-cov\n"
                "Or install all dev dependencies:\n"
                "  pip install -e '.[dev]'\n"
                "Or from requirements.txt:\n"
                "  pip install -r requirements.txt"
            )
            logger.error(error_msg)
            raise ImportError(error_msg)
        
        try:
            cmd = [sys.executable, '-m', 'pytest', '--cov=project_management_automation', '--cov-report=xml:' + str(self.output_path / 'coverage.xml')]
            result = subprocess.run(
                cmd,
                cwd=str(self.project_root),
                capture_output=True,
                text=True,
                timeout=300
            )
            if result.returncode != 0:
                logger.warning(f"Coverage generation had issues: {result.stderr}")
                # Check if it's because pytest-cov isn't recognized
                if 'unrecognized arguments: --cov' in result.stderr:
                    error_msg = (
                        "pytest-cov plugin not loaded. Try:\n"
                        "  1. pip install pytest-cov\n"
                        "  2. Verify pytest-cov is installed: python -c 'import pytest_cov'\n"
                        "  3. Check pytest plugins: pytest --collect-only"
                    )
                    logger.error(error_msg)
                    raise RuntimeError(error_msg)
        except Exception as e:
            logger.warning(f"Failed to generate coverage: {e}")
            raise

    def _parse_coverage_xml(self, coverage_file: Path) -> dict:
        """Parse coverage.xml (coverage.py format)."""
        try:
            tree = ET.parse(coverage_file)
            root = tree.getroot()

            coverage_data = {
                'format': 'xml',
                'files': [],
                'total_lines': 0,
                'covered_lines': 0,
                'total_coverage': 0
            }

            for package in root.findall('.//package'):
                for class_elem in package.findall('.//class'):
                    filename = class_elem.get('filename', '')
                    lines = class_elem.findall('.//line')

                    file_data = {
                        'filename': filename,
                        'lines': len(lines),
                        'covered': sum(1 for line in lines if line.get('hits', '0') != '0')
                    }

                    coverage_data['files'].append(file_data)
                    coverage_data['total_lines'] += file_data['lines']
                    coverage_data['covered_lines'] += file_data['covered']

            if coverage_data['total_lines'] > 0:
                coverage_data['total_coverage'] = (coverage_data['covered_lines'] / coverage_data['total_lines']) * 100

            return coverage_data
        except Exception as e:
            logger.error(f"Failed to parse coverage XML: {e}")
            return {'format': 'xml', 'error': str(e)}

    def _parse_coverage_json(self, coverage_file: Path) -> dict:
        """Parse coverage.json."""
        try:
            with open(coverage_file) as f:
                data = json.load(f)

            coverage_data = {
                'format': 'json',
                'files': [],
                'total_lines': 0,
                'covered_lines': 0,
                'total_coverage': data.get('totals', {}).get('percent_covered', 0)
            }

            for filename, file_data in data.get('files', {}).items():
                lines = file_data.get('summary', {}).get('num_statements', 0)
                covered = file_data.get('summary', {}).get('covered_lines', 0)

                coverage_data['files'].append({
                    'filename': filename,
                    'lines': lines,
                    'covered': covered
                })
                coverage_data['total_lines'] += lines
                coverage_data['covered_lines'] += covered

            return coverage_data
        except Exception as e:
            logger.error(f"Failed to parse coverage JSON: {e}")
            return {'format': 'json', 'error': str(e)}

    def _parse_coverage_py(self, coverage_file: Path) -> dict:
        """Parse .coverage file using coverage.py."""
        try:
            # Use coverage.py to generate report
            cmd = [sys.executable, '-m', 'coverage', 'json', '-o', str(self.output_path / 'coverage.json')]
            result = subprocess.run(
                cmd,
                cwd=str(self.project_root),
                capture_output=True,
                text=True,
                timeout=60
            )

            if result.returncode == 0:
                json_file = self.output_path / 'coverage.json'
                if json_file.exists():
                    return self._parse_coverage_json(json_file)

            return {'format': 'coverage.py', 'error': 'Failed to parse coverage.py file'}
        except Exception as e:
            logger.error(f"Failed to parse coverage.py: {e}")
            return {'format': 'coverage.py', 'error': str(e)}

    def _generate_coverage_report_file(self, coverage_data: dict) -> Path:
        """Generate coverage report."""
        if self.format == 'html':
            return self._generate_html_report(coverage_data)
        elif self.format == 'json':
            return self._generate_json_report(coverage_data)
        else:
            return self._generate_terminal_report(coverage_data)

    def _generate_html_report(self, coverage_data: dict) -> Path:
        """Generate HTML coverage report."""
        report_path = self.output_path / 'coverage_report.html'

        html = f"""<!DOCTYPE html>
<html>
<head>
    <title>Test Coverage Report</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; }}
        h1 {{ color: #333; }}
        .summary {{ background: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0; }}
        .coverage {{ font-size: 24px; font-weight: bold; }}
        .good {{ color: green; }}
        .warning {{ color: orange; }}
        .bad {{ color: red; }}
        table {{ border-collapse: collapse; width: 100%; }}
        th, td {{ border: 1px solid #ddd; padding: 8px; text-align: left; }}
        th {{ background-color: #4CAF50; color: white; }}
    </style>
</head>
<body>
    <h1>Test Coverage Report</h1>
    <div class="summary">
        <h2>Summary</h2>
        <p>Total Coverage: <span class="coverage {'good' if coverage_data.get('total_coverage', 0) >= self.min_coverage else 'bad'}">{coverage_data.get('total_coverage', 0):.1f}%</span></p>
        <p>Total Lines: {coverage_data.get('total_lines', 0)}</p>
        <p>Covered Lines: {coverage_data.get('covered_lines', 0)}</p>
        <p>Minimum Threshold: {self.min_coverage}%</p>
        <p>Status: {'✅ Meets Threshold' if coverage_data.get('total_coverage', 0) >= self.min_coverage else '❌ Below Threshold'}</p>
    </div>
    <h2>File Coverage</h2>
    <table>
        <tr>
            <th>File</th>
            <th>Lines</th>
            <th>Covered</th>
            <th>Coverage %</th>
        </tr>
"""

        for file_data in coverage_data.get('files', [])[:50]:  # Limit to 50 files
            coverage_pct = (file_data['covered'] / file_data['lines'] * 100) if file_data['lines'] > 0 else 0
            html += f"""
        <tr>
            <td>{file_data['filename']}</td>
            <td>{file_data['lines']}</td>
            <td>{file_data['covered']}</td>
            <td>{coverage_pct:.1f}%</td>
        </tr>
"""

        html += """
    </table>
    <p><em>Report generated: """ + datetime.now().isoformat() + """</em></p>
</body>
</html>
"""

        with open(report_path, 'w') as f:
            f.write(html)

        return report_path

    def _generate_json_report(self, coverage_data: dict) -> Path:
        """Generate JSON coverage report."""
        report_path = self.output_path / 'coverage_report.json'

        report = {
            'timestamp': datetime.now().isoformat(),
            'min_coverage': self.min_coverage,
            'coverage_data': coverage_data
        }

        with open(report_path, 'w') as f:
            json.dump(report, f, indent=2)

        return report_path

    def _generate_terminal_report(self, coverage_data: dict) -> Path:
        """Generate terminal-friendly coverage report."""
        report_path = self.output_path / 'coverage_report.txt'

        lines = [
            "Test Coverage Report",
            "=" * 60,
            f"Total Coverage: {coverage_data.get('total_coverage', 0):.1f}%",
            f"Total Lines: {coverage_data.get('total_lines', 0)}",
            f"Covered Lines: {coverage_data.get('covered_lines', 0)}",
            f"Minimum Threshold: {self.min_coverage}%",
            f"Status: {'✅ Meets Threshold' if coverage_data.get('total_coverage', 0) >= self.min_coverage else '❌ Below Threshold'}",
            "",
            "File Coverage:",
            "-" * 60
        ]

        for file_data in coverage_data.get('files', [])[:20]:  # Limit to 20 files
            coverage_pct = (file_data['covered'] / file_data['lines'] * 100) if file_data['lines'] > 0 else 0
            lines.append(f"{file_data['filename']}: {coverage_pct:.1f}% ({file_data['covered']}/{file_data['lines']})")

        lines.append(f"\nReport generated: {datetime.now().isoformat()}")

        with open(report_path, 'w') as f:
            f.write('\n'.join(lines))

        return report_path

    def _identify_gaps(self, coverage_data: dict) -> list[dict]:
        """Identify files with low coverage."""
        gaps = []

        for file_data in coverage_data.get('files', []):
            if file_data['lines'] > 0:
                coverage_pct = (file_data['covered'] / file_data['lines']) * 100
                if coverage_pct < self.min_coverage:
                    gaps.append({
                        'filename': file_data['filename'],
                        'coverage': coverage_pct,
                        'lines': file_data['lines'],
                        'covered': file_data['covered'],
                        'missing': file_data['lines'] - file_data['covered']
                    })

        # Sort by coverage (lowest first)
        gaps.sort(key=lambda x: x['coverage'])

        return gaps

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from coverage analysis."""
        insights = []

        results = analysis_results.get('results', {})
        coverage_data = results.get('coverage_data', {})
        total_coverage = coverage_data.get('total_coverage', 0)
        min_coverage = self.min_coverage
        meets_threshold = results.get('meets_threshold', False)
        gaps = results.get('gaps', [])

        insights.append(f"**Test Coverage: {total_coverage:.1f}%**")
        insights.append(f"- Minimum Threshold: {min_coverage}%")

        if meets_threshold:
            insights.append("✅ Coverage meets threshold!")
        else:
            insights.append(f"⚠️ Coverage below threshold (gap: {min_coverage - total_coverage:.1f}%)")

        if gaps:
            insights.append(f"⚠️ {len(gaps)} files have low coverage")
            top_gaps = sorted(gaps, key=lambda x: x.get('coverage', 100))[:5]
            for gap in top_gaps:
                insights.append(f"  - {gap.get('filename', 'Unknown')}: {gap.get('coverage', 0):.1f}%")

        return '\n'.join(insights)

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate coverage report."""
        results = analysis_results.get('results', {})
        coverage_data = results.get('coverage_data', {})
        report_path = results.get('report_path', '')

        report_lines = [
            "# Test Coverage Report",
            "",
            f"*Generated: {datetime.now().isoformat()}*",
            "",
            "## Summary",
            "",
            f"- **Total Coverage:** {coverage_data.get('total_coverage', 0):.1f}%",
            f"- **Total Lines:** {coverage_data.get('total_lines', 0)}",
            f"- **Covered Lines:** {coverage_data.get('covered_lines', 0)}",
            f"- **Minimum Threshold:** {self.min_coverage}%",
            f"- **Meets Threshold:** {'✅ Yes' if results.get('meets_threshold', False) else '❌ No'}",
            "",
            "## Insights",
            "",
            insights,
            ""
        ]

        if report_path:
            report_lines.append(f"**Report File:** {report_path}")

        gaps = results.get('gaps', [])
        if gaps:
            report_lines.extend([
                "",
                "## Coverage Gaps",
                ""
            ])
            for gap in gaps[:10]:
                report_lines.append(f"- **{gap.get('filename', 'Unknown')}**: {gap.get('coverage', 0):.1f}% ({gap.get('missing', 0)} lines missing)")

        return '\n'.join(report_lines)

    def _format_findings(self, analysis_results: dict) -> str:
        """Format coverage analysis results."""
        return json.dumps(analysis_results, indent=2)


def load_config(config_path: Optional[Path] = None) -> dict:
    """Load configuration."""
    default_config = {
        'coverage_file': None,
        'min_coverage': 80,
        'output_path': 'coverage-report/',
        'format': 'html'
    }

    if config_path and config_path.exists():
        try:
            with open(config_path) as f:
                user_config = json.load(f)
                default_config.update(user_config)
        except json.JSONDecodeError:
            pass

    return default_config


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Analyze Test Coverage')
    parser.add_argument('--coverage-file', type=str, help='Path to coverage file')
    parser.add_argument('--min-coverage', type=int, default=80, help='Minimum coverage threshold')
    parser.add_argument('--output', type=str, help='Output path for report')
    parser.add_argument('--format', type=str, choices=['html', 'json', 'terminal'], default='html')
    parser.add_argument('--config', type=Path, help='Path to config file')
    args = parser.parse_args()

    config = load_config(args.config)
    if args.coverage_file:
        config['coverage_file'] = args.coverage_file
    if args.min_coverage:
        config['min_coverage'] = args.min_coverage
    if args.output:
        config['output_path'] = args.output
    if args.format:
        config['format'] = args.format

    analyzer = TestCoverageAnalyzer(config)

    try:
        results = analyzer.run()
        print(json.dumps(results, indent=2))
        sys.exit(0 if results.get('status') == 'success' else 1)
    except Exception as e:
        logger.error(f"Error analyzing coverage: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()

