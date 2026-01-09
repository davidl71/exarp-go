#!/usr/bin/env python3
"""
Automated Duplicate Test Name Checker

Uses IntelligentAutomationBase to check for duplicate test function names
across test files. This prevents confusion and maintenance issues.

Usage:
    python3 -m project_management_automation.scripts.automate_check_duplicate_test_names
"""

import argparse
import json
import logging
import re
import sys
from collections import defaultdict
from datetime import datetime
from pathlib import Path
from typing import Optional

# Import base class
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase

logger = logging.getLogger(__name__)


class DuplicateTestNameChecker(IntelligentAutomationBase):
    """Intelligent duplicate test name checker using base class."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Duplicate Test Name Check", project_root)

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is duplicate test name checking?"""
        return "What is duplicate test name checking? Duplicate Test Name Check = Test Function Discovery × Cross-File Comparison × Duplicate Detection × Reporting"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we check for duplicate test names?"""
        return "How do we systematically scan all test files, extract function names, detect cross-file duplicates, and report findings?"

    def _execute_analysis(self) -> dict:
        """Execute duplicate test name check."""
        logger.info("Checking for duplicate test names...")

        tests_dir = self.project_root / 'tests'

        if not tests_dir.exists():
            return {
                'success': False,
                'error': f"Tests directory not found: {tests_dir}",
                'duplicates': [],
                'total_tests': 0
            }

        # Collect all test functions with file and class context
        test_functions = defaultdict(list)

        for test_file in sorted(tests_dir.glob('test_*.py')):
            functions = self._find_test_functions_with_context(test_file)
            file_path = str(test_file.relative_to(self.project_root))
            for func_name, test_class in functions:
                test_functions[func_name].append((file_path, test_class))

        # Find duplicates across different files (same name in different files = bad)
        # Same name in different classes within same file is acceptable
        cross_file_duplicates = {}
        for name, locations in test_functions.items():
            files = set(loc[0] for loc in locations)
            if len(files) > 1:  # Same name in different files
                cross_file_duplicates[name] = locations

        total_test_functions = sum(len(locs) for locs in test_functions.values())

        return {
            'success': True,
            'duplicates': cross_file_duplicates,
            'duplicate_count': len(cross_file_duplicates),
            'total_test_functions': total_test_functions,
            'total_test_files': len(list(tests_dir.glob('test_*.py')))
        }

    def _find_test_functions_with_context(self, test_file: Path) -> list[tuple[str, str]]:
        """Extract test function names with their class context."""
        content = test_file.read_text()
        functions = []
        current_class = None

        for line in content.split('\n'):
            # Match class definitions
            class_match = re.match(r'class (\w+):', line)
            if class_match:
                current_class = class_match.group(1)

            # Match function definitions
            func_match = re.match(r'\s+def (test_\w+)\(', line)
            if func_match:
                func_name = func_match.group(1)
                functions.append((func_name, current_class))

        return functions

    def _generate_insights(self, analysis_results: dict) -> list[str]:
        """Generate insights from analysis results."""
        insights = []

        if analysis_results.get('success'):
            duplicate_count = analysis_results.get('duplicate_count', 0)
            total_test_functions = analysis_results.get('total_test_functions', 0)

            if duplicate_count == 0:
                insights.append(f"✅ No duplicate test names found across {analysis_results.get('total_test_files', 0)} test files")
            else:
                insights.append(f"⚠️  Found {duplicate_count} duplicate test name(s) across different files")
                insights.append(f"Total test functions scanned: {total_test_functions}")
                insights.append("Recommendation: Rename duplicate tests to be more specific (include module/tool name)")

        return insights

    def _generate_report(self, analysis_results: dict, insights: list[str]) -> str:
        """Generate report from analysis results."""
        lines = [
            "# Duplicate Test Name Check Report",
            "",
            f"*Generated: {datetime.now().isoformat()}*",
            "",
            "## Summary",
            ""
        ]

        if analysis_results.get('success'):
            duplicate_count = analysis_results.get('duplicate_count', 0)
            total_test_functions = analysis_results.get('total_test_functions', 0)
            total_test_files = analysis_results.get('total_test_files', 0)

            lines.extend([
                f"- **Test Files Scanned**: {total_test_files}",
                f"- **Total Test Functions**: {total_test_functions}",
                f"- **Duplicate Names Found**: {duplicate_count}",
                ""
            ])

            if duplicate_count > 0:
                lines.extend([
                    "## Duplicate Test Names",
                    "",
                    "The following test function names appear in multiple files:",
                    ""
                ])

                duplicates = analysis_results.get('duplicates', {})
                for name, locations in sorted(duplicates.items()):
                    lines.append(f"### {name}")
                    for file, test_class in locations:
                        lines.append(f"- {file} (class: {test_class})")
                    lines.append("")

                lines.extend([
                    "## Recommendation",
                    "",
                    "Rename duplicate tests to be more specific:",
                    "- Include module/tool name in test name",
                    "- Add context (e.g., `test_mcp_client_no_config_basic` vs `test_mcp_client_no_config_agentic`)",
                    "- See `docs/TEST_ORGANIZATION_GUIDELINES.md` for naming conventions",
                    ""
                ])
            else:
                lines.append("✅ **No duplicate test names found!**")
                lines.append("")

        lines.extend([
            "## Insights",
            ""
        ])
        for insight in insights:
            lines.append(f"- {insight}")
        lines.append("")

        return "\n".join(lines)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Check for duplicate test names')
    parser.add_argument('--output', type=str, help='Output path for report')
    parser.add_argument('--config', type=Path, help='Config file path')
    args = parser.parse_args()

    # Load config
    config = {}
    if args.config and args.config.exists():
        with open(args.config) as f:
            config = json.load(f)

    if args.output:
        config['output_path'] = args.output

    # Create checker and run
    checker = DuplicateTestNameChecker(config)
    results = checker.run()

    # Print JSON output for programmatic use
    print(json.dumps(results, indent=2))

    # Exit with appropriate code
    analysis_results = results.get('results', {})
    if analysis_results.get('duplicate_count', 0) > 0:
        sys.exit(1)
    else:
        sys.exit(0)


if __name__ == '__main__':
    main()

