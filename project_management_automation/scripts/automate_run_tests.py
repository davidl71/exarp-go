#!/usr/bin/env python3
"""
Automated Test Runner Script

Executes test suites with flexible options for pytest, unittest, and ctest.
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


class TestRunner(IntelligentAutomationBase):
    """Test runner using intelligent automation base."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Test Runner", project_root)

        self.test_path = Path(config.get('test_path', 'tests/'))
        self.test_framework = config.get('test_framework', 'auto')
        self.verbose = config.get('verbose', True)
        self.coverage = config.get('coverage', False)
        self.output_path = Path(config.get('output_path', 'test-results/'))

        # Ensure output directory exists
        self.output_path.mkdir(parents=True, exist_ok=True)

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is test execution?"""
        return "What is test execution? Test Execution = Framework Detection × Test Discovery × Execution × Result Reporting"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we run tests?"""
        return "How do we systematically execute tests across different frameworks?"

    def _detect_framework(self) -> str:
        """Auto-detect test framework."""
        if self.test_framework != 'auto':
            return self.test_framework

        # Check for pytest
        if (self.project_root / 'pytest.ini').exists() or \
           (self.project_root / 'pyproject.toml').exists() and 'pytest' in (self.project_root / 'pyproject.toml').read_text():
            return 'pytest'

        # Check for unittest
        if (self.test_path / '__init__.py').exists() or \
           any(f.startswith('test_') for f in self.test_path.glob('*.py')):
            return 'unittest'

        # Check for ctest (CMake)
        if (self.project_root / 'CMakeLists.txt').exists() and \
           (self.project_root / 'build').exists():
            return 'ctest'

        # Default to pytest
        return 'pytest'

    def _execute_analysis(self) -> dict:
        """Execute test suite."""
        logger.info(f"Running tests with framework: {self.test_framework}")

        framework = self._detect_framework()
        logger.info(f"Detected framework: {framework}")

        results = {
            'framework': framework,
            'test_path': str(self.test_path),
            'tests_run': 0,
            'tests_passed': 0,
            'tests_failed': 0,
            'tests_skipped': 0,
            'duration': 0,
            'output_file': None,
            'coverage_file': None,
            'status': 'unknown'
        }

        try:
            if framework == 'pytest':
                results = self._run_pytest()
            elif framework == 'unittest':
                results = self._run_unittest()
            elif framework == 'ctest':
                results = self._run_ctest()
            else:
                raise ValueError(f"Unsupported framework: {framework}")

            results['status'] = 'success' if results['tests_failed'] == 0 else 'failed'

        except Exception as e:
            logger.error(f"Test execution failed: {e}", exc_info=True)
            results['status'] = 'error'
            results['error'] = str(e)

        return results

    def _run_pytest(self) -> dict:
        """Run pytest tests."""
        cmd = [sys.executable, '-m', 'pytest', str(self.test_path)]

        if self.verbose:
            cmd.append('-v')

        if self.coverage:
            cmd.extend(['--cov=project_management_automation', '--cov=tools', '--cov=resources'])
            cmd.append('--cov-report=html:' + str(self.output_path / 'coverage'))
            cmd.append('--cov-report=xml:' + str(self.output_path / 'coverage.xml'))

        # Generate JUnit XML
        junit_xml = self.output_path / 'junit.xml'
        cmd.extend(['--junit-xml', str(junit_xml)])

        logger.info(f"Running: {' '.join(cmd)}")
        result = subprocess.run(
            cmd,
            cwd=str(self.project_root),
            capture_output=True,
            text=True,
            timeout=300
        )

        # Parse results
        results = {
            'framework': 'pytest',
            'tests_run': 0,
            'tests_passed': 0,
            'tests_failed': 0,
            'tests_skipped': 0,
            'output': result.stdout,
            'error': result.stderr,
            'returncode': result.returncode,
            'output_file': str(junit_xml)
        }

        # Parse JUnit XML if available
        if junit_xml.exists():
            try:
                tree = ET.parse(junit_xml)
                root = tree.getroot()
                testsuite = root.find('testsuite')
                if testsuite is not None:
                    results['tests_run'] = int(testsuite.get('tests', 0))
                    results['tests_failed'] = int(testsuite.get('failures', 0))
                    results['tests_skipped'] = int(testsuite.get('skipped', 0))
                    results['tests_passed'] = results['tests_run'] - results['tests_failed'] - results['tests_skipped']
                    results['duration'] = float(testsuite.get('time', 0))
            except Exception as e:
                logger.warning(f"Failed to parse JUnit XML: {e}")

        # Parse stdout for summary
        if 'passed' in result.stdout.lower():
            for line in result.stdout.split('\n'):
                if 'passed' in line.lower() and 'failed' in line.lower():
                    # Try to extract numbers
                    import re
                    nums = re.findall(r'\d+', line)
                    if len(nums) >= 2:
                        results['tests_passed'] = int(nums[0])
                        results['tests_failed'] = int(nums[1]) if len(nums) > 1 else 0
                        results['tests_run'] = results['tests_passed'] + results['tests_failed']
                    break

        if self.coverage:
            coverage_xml = self.output_path / 'coverage.xml'
            if coverage_xml.exists():
                results['coverage_file'] = str(coverage_xml)

        return results

    def _run_unittest(self) -> dict:
        """Run unittest tests."""
        cmd = ['python3', '-m', 'unittest', 'discover', '-s', str(self.test_path), '-p', 'test_*.py']

        if self.verbose:
            cmd.append('-v')

        logger.info(f"Running: {' '.join(cmd)}")
        result = subprocess.run(
            cmd,
            cwd=str(self.project_root),
            capture_output=True,
            text=True,
            timeout=300
        )

        # Parse results
        results = {
            'framework': 'unittest',
            'tests_run': 0,
            'tests_passed': 0,
            'tests_failed': 0,
            'tests_skipped': 0,
            'output': result.stdout,
            'error': result.stderr,
            'returncode': result.returncode
        }

        # Parse unittest output
        if 'OK' in result.stdout:
            # Extract test count
            import re
            match = re.search(r'Ran (\d+) test', result.stdout)
            if match:
                results['tests_run'] = int(match.group(1))
                results['tests_passed'] = results['tests_run']
        elif 'FAILED' in result.stdout:
            # Extract failure count
            import re
            match = re.search(r'Ran (\d+) test', result.stdout)
            if match:
                results['tests_run'] = int(match.group(1))
            match = re.search(r'failures?=(\d+)', result.stdout)
            if match:
                results['tests_failed'] = int(match.group(1))
                results['tests_passed'] = results['tests_run'] - results['tests_failed']

        return results

    def _run_ctest(self) -> dict:
        """Run ctest (CMake tests)."""
        build_dir = self.project_root / 'build'
        if not build_dir.exists():
            raise FileNotFoundError(f"Build directory not found: {build_dir}")

        cmd = ['ctest', '--output-on-failure']

        if self.verbose:
            cmd.append('-V')

        # Generate XML
        self.output_path / 'ctest.xml'
        cmd.extend(['-T', 'Test', '--no-compress-output'])

        logger.info(f"Running: {' '.join(cmd)}")
        result = subprocess.run(
            cmd,
            cwd=str(build_dir),
            capture_output=True,
            text=True,
            timeout=300
        )

        # Parse results
        results = {
            'framework': 'ctest',
            'tests_run': 0,
            'tests_passed': 0,
            'tests_failed': 0,
            'tests_skipped': 0,
            'output': result.stdout,
            'error': result.stderr,
            'returncode': result.returncode
        }

        # Parse ctest output
        import re
        match = re.search(r'Tests run: (\d+)', result.stdout)
        if match:
            results['tests_run'] = int(match.group(1))

        match = re.search(r'(\d+) passed', result.stdout)
        if match:
            results['tests_passed'] = int(match.group(1))

        match = re.search(r'(\d+) failed', result.stdout)
        if match:
            results['tests_failed'] = int(match.group(1))

        return results

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from test results."""
        insights = []

        results = analysis_results.get('results', {})
        framework = results.get('framework', 'unknown')
        tests_run = results.get('tests_run', 0)
        tests_passed = results.get('tests_passed', 0)
        tests_failed = results.get('tests_failed', 0)
        status = results.get('status', 'unknown')

        insights.append(f"**Test Execution ({framework}):** {tests_run} tests run")
        insights.append(f"- Passed: {tests_passed}")
        insights.append(f"- Failed: {tests_failed}")

        if status == 'success':
            insights.append("✅ All tests passed!")
        elif status == 'failed':
            insights.append(f"⚠️ {tests_failed} test(s) failed")

        return '\n'.join(insights)

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate test execution report."""
        results = analysis_results.get('results', {})
        report_lines = [
            "# Test Execution Report",
            "",
            f"*Generated: {datetime.now().isoformat()}*",
            "",
            "## Summary",
            "",
            f"- **Framework:** {results.get('framework', 'unknown')}",
            f"- **Tests Run:** {results.get('tests_run', 0)}",
            f"- **Tests Passed:** {results.get('tests_passed', 0)}",
            f"- **Tests Failed:** {results.get('tests_failed', 0)}",
            f"- **Tests Skipped:** {results.get('tests_skipped', 0)}",
            f"- **Duration:** {results.get('duration', 0):.2f}s",
            f"- **Status:** {results.get('status', 'unknown')}",
            "",
            "## Insights",
            "",
            insights,
            ""
        ]

        if results.get('output_file'):
            report_lines.append(f"**Output File:** {results['output_file']}")

        if results.get('coverage_file'):
            report_lines.append(f"**Coverage File:** {results['coverage_file']}")

        return '\n'.join(report_lines)

    def _format_findings(self, analysis_results: dict) -> str:
        """Format test results."""
        return json.dumps(analysis_results, indent=2)


def load_config(config_path: Optional[Path] = None) -> dict:
    """Load configuration."""
    default_config = {
        'test_path': 'tests/',
        'test_framework': 'auto',
        'verbose': True,
        'coverage': False,
        'output_path': 'test-results/'
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
    parser = argparse.ArgumentParser(description='Run Test Suite')
    parser.add_argument('--test-path', type=str, help='Path to test file/directory')
    parser.add_argument('--framework', type=str, choices=['pytest', 'unittest', 'ctest', 'auto'], default='auto')
    parser.add_argument('--verbose', action='store_true', default=True)
    parser.add_argument('--coverage', action='store_true', help='Generate coverage report')
    parser.add_argument('--output', type=str, help='Output path for results')
    parser.add_argument('--config', type=Path, help='Path to config file')
    args = parser.parse_args()

    config = load_config(args.config)
    if args.test_path:
        config['test_path'] = args.test_path
    if args.framework:
        config['test_framework'] = args.framework
    if args.verbose is not None:
        config['verbose'] = args.verbose
    if args.coverage:
        config['coverage'] = True
    if args.output:
        config['output_path'] = args.output

    runner = TestRunner(config)

    try:
        results = runner.run()
        print(json.dumps(results, indent=2))
        sys.exit(0 if results.get('status') == 'success' else 1)
    except Exception as e:
        logger.error(f"Error running tests: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()

