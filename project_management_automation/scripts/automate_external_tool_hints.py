#!/usr/bin/env python3
"""
Automated External Tool Hints Script

Automatically detects where Context7/external tool hints should be added to documentation
and inserts them following the standard pattern from docs/DOCUMENTATION_EXTERNAL_TOOL_HINTS.md.
"""

import argparse
import json
import logging
import re
import sys
from datetime import datetime
from pathlib import Path
from typing import Optional

# Add project root to path
# Project root will be passed to __init__
# Import base class
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase

# Configure logging (will be configured after project_root is set)
logger = logging.getLogger(__name__)


# Common external libraries/frameworks to detect
EXTERNAL_LIBRARIES = {
    'FastAPI', 'React', 'Vue', 'Angular', 'TypeScript', 'JavaScript',
    'Rust', 'Python', 'C++', 'CMake', 'NATS', 'RabbitMQ', 'Kafka',
    'PostgreSQL', 'MySQL', 'MongoDB', 'Redis', 'Docker', 'Kubernetes',
    'TensorFlow', 'PyTorch', 'NumPy', 'Pandas', 'Scikit-learn',
    'Axum', 'Tokio', 'Serde', 'Clap', 'FTXUI', 'Textual',
    'Pydantic', 'SQLAlchemy', 'Alembic', 'Uvicorn', 'Gunicorn'
}

# Context7 hint template
CONTEXT7_HINT_TEMPLATE = """> 💡 **AI Assistant Hint:** For up-to-date, version-specific documentation on {libraries}, use the Context7 MCP server by appending `use context7` to your prompts. For example:
> - "How do I use {example_lib} {example_feature}? use context7"
> - "Show me {example_lib} {example_pattern} examples use context7"
> - "{example_lib} best practices 2025 use context7"
>
> Context7 provides current documentation (2025), version-specific API references, and real code examples without hallucinations."""

# Inline hint template (for specific library mentions)
INLINE_HINT_TEMPLATE = """For more information, see [{lib} Documentation](link) or query Context7: "{lib} patterns use context7\""""


class ExternalToolHintsAutomation(IntelligentAutomationBase):
    """Intelligent automation for adding external tool hints to documentation."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "External Tool Hints Automation", project_root)
        self.docs_path = self.project_root / 'docs'
        self.pattern_guide_path = self.project_root / 'docs' / 'DOCUMENTATION_EXTERNAL_TOOL_HINTS.md'

        # Configuration
        self.dry_run = config.get('dry_run', False)
        self.output_path = config.get('output_path', None)
        self.min_file_size = config.get('min_file_size', 50)  # Skip files < 50 lines

        # Results
        self.files_scanned = 0
        self.files_modified = 0
        self.files_skipped = 0
        self.hints_added = []
        self.hints_skipped = []

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What are external tool hints?"""
        return "What are external tool hints? External Tool Hints = Library Detection × Appropriate Placement × Standard Format × AI Discoverability"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we add external tool hints?"""
        return "How do we systematically detect external libraries and add Context7 hints to documentation?"

    def _identify_followup_tasks(self, analysis_results: dict) -> list[dict]:
        """Identify follow-up tasks based on analysis results."""
        followups = []

        # If many files were skipped, suggest reviewing them
        files_skipped = analysis_results.get('files_skipped', 0)
        if files_skipped > 10:
            followups.append({
                'name': 'Review skipped files for manual hint addition',
                'description': f'{files_skipped} files were skipped - review for manual hint addition if needed',
                'priority': 'low',
                'tags': ['documentation', 'manual-review']
            })

        return followups

    def _format_findings(self, analysis_results: dict) -> str:
        """Format findings for Todo2 result comment."""
        findings = []

        files_modified = analysis_results.get('files_modified', 0)
        hints_added = analysis_results.get('hints_added', [])

        if files_modified > 0:
            findings.append(f"- Added hints to {files_modified} files")
            for hint in hints_added[:5]:  # First 5
                findings.append(f"  - {hint['file']}: {', '.join(hint['libraries'][:3])}")

        return '\n'.join(findings) if findings else "No findings to report"

    def _execute_analysis(self) -> dict:
        """Execute external tool hints analysis and insertion."""
        logger.info("Executing external tool hints automation...")

        if not self.docs_path.exists():
            logger.error(f"Documentation path not found: {self.docs_path}")
            return {'status': 'error', 'error': 'Documentation path not found'}

        # Find all markdown files
        md_files = list(self.docs_path.rglob('*.md'))
        logger.info(f"Found {len(md_files)} markdown files to scan")

        # Process each file
        for md_file in md_files:
            self.files_scanned += 1
            result = self._process_file(md_file)

            if result['modified']:
                self.files_modified += 1
                self.hints_added.append({
                    'file': str(md_file.relative_to(self.project_root)),
                    'libraries': result['libraries'],
                    'hint_type': result['hint_type']
                })
            elif result['skipped']:
                self.files_skipped += 1
                self.hints_skipped.append({
                    'file': str(md_file.relative_to(self.project_root)),
                    'reason': result['reason']
                })

        # Prepare results dict (will be used by base class for insights and report)
        results = {
            'status': 'success',
            'files_scanned': self.files_scanned,
            'files_modified': self.files_modified,
            'files_skipped': self.files_skipped,
            'hints_added': self.hints_added,
            'hints_skipped': self.hints_skipped,
            'report_path': str(self.output_path) if self.output_path else None,
            'dry_run': self.dry_run
        }

        return results

    def _process_file(self, file_path: Path) -> dict:
        """Process a single markdown file."""
        try:
            content = file_path.read_text(encoding='utf-8')

            # Skip very short files
            if len(content.split('\n')) < self.min_file_size:
                return {'modified': False, 'skipped': True, 'reason': 'File too short'}

            # Check if hint already exists
            if self._has_existing_hint(content):
                return {'modified': False, 'skipped': True, 'reason': 'Hint already exists'}

            # Detect external libraries
            libraries = self._detect_libraries(content)

            if not libraries:
                return {'modified': False, 'skipped': True, 'reason': 'No external libraries detected'}

            # Determine hint type and placement
            hint_type, insertion_point = self._determine_hint_placement(content, libraries)

            if not insertion_point:
                return {'modified': False, 'skipped': True, 'reason': 'Could not determine insertion point'}

            # Generate hint
            hint_text = self._generate_hint(libraries, hint_type)

            # Insert hint
            if not self.dry_run:
                new_content = self._insert_hint(content, hint_text, insertion_point)
                file_path.write_text(new_content, encoding='utf-8')
                logger.info(f"Added hint to: {file_path.relative_to(self.project_root)}")

            return {
                'modified': True,
                'skipped': False,
                'libraries': list(libraries),
                'hint_type': hint_type
            }

        except Exception as e:
            logger.error(f"Error processing {file_path}: {e}")
            return {'modified': False, 'skipped': True, 'reason': f'Error: {str(e)}'}

    def _has_existing_hint(self, content: str) -> bool:
        """Check if file already has a Context7 hint."""
        # Look for Context7 hint patterns
        patterns = [
            r'AI Assistant Hint.*Context7',
            r'use context7',
            r'Context7 MCP server'
        ]

        for pattern in patterns:
            if re.search(pattern, content, re.IGNORECASE):
                return True

        return False

    def _detect_libraries(self, content: str) -> set[str]:
        """Detect external libraries mentioned in content."""
        found_libraries = set()

        # Check for library mentions (case-insensitive)
        content_lower = content.lower()

        for lib in EXTERNAL_LIBRARIES:
            # Look for library name (word boundary to avoid partial matches)
            pattern = r'\b' + re.escape(lib.lower()) + r'\b'
            if re.search(pattern, content_lower):
                found_libraries.add(lib)

        return found_libraries

    def _determine_hint_placement(self, content: str, libraries: set[str]) -> tuple[str, Optional[int]]:
        """Determine where to place the hint and what type."""
        lines = content.split('\n')

        # Check if this is an API documentation file (has "API" in title or early content)
        is_api_doc = any('api' in line.lower() for line in lines[:10])

        # Check if this is an integration guide
        is_integration = any('integration' in line.lower() for line in lines[:20])

        # For API docs or files starting with #, insert after title
        if is_api_doc or (lines and lines[0].startswith('#')):
            # Find first non-title line after title
            for i, line in enumerate(lines[1:10], start=1):
                if line.strip() and not line.startswith('#'):
                    return 'top', i

        # For integration guides, insert after first section
        if is_integration:
            for i, line in enumerate(lines[5:30], start=5):
                if line.startswith('##'):
                    return 'section', i + 1

        # Default: insert after title (first # line)
        for i, line in enumerate(lines[1:10], start=1):
            if line.strip() and not line.startswith('#'):
                return 'top', i

        return 'top', None

    def _generate_hint(self, libraries: set[str], hint_type: str) -> str:
        """Generate hint text based on libraries and type."""
        lib_list = ', '.join(sorted(libraries)[:3])  # Top 3 libraries
        example_lib = sorted(libraries)[0] if libraries else 'Library'

        if hint_type == 'top':
            # Full hint for top of file
            return CONTEXT7_HINT_TEMPLATE.format(
                libraries=lib_list,
                example_lib=example_lib,
                example_feature='async endpoints' if 'FastAPI' in libraries else 'patterns',
                example_pattern='hooks' if 'React' in libraries else 'examples'
            )
        else:
            # Shorter hint for sections
            return f"> 💡 **AI Assistant Hint:** Query Context7 for up-to-date {lib_list} documentation: \"{example_lib} patterns use context7\""

    def _insert_hint(self, content: str, hint_text: str, insertion_point: int) -> str:
        """Insert hint at the specified line."""
        lines = content.split('\n')

        # Insert hint with blank lines around it
        lines.insert(insertion_point, '')
        lines.insert(insertion_point, hint_text)
        lines.insert(insertion_point, '')

        return '\n'.join(lines)

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from analysis."""
        insights = []

        files_modified = analysis_results.get('files_modified', 0)
        files_scanned = analysis_results.get('files_scanned', 0)
        hints_added = analysis_results.get('hints_added', [])

        if files_modified > 0:
            insights.append(f"✅ Added Context7 hints to {files_modified} documentation files")
            insights.append(f"📚 Libraries detected: {len({lib for h in hints_added for lib in h.get('libraries', [])})} unique libraries")

        if files_scanned > 0:
            coverage = (files_modified / files_scanned * 100) if files_scanned > 0 else 0
            insights.append(f"📊 Coverage: {coverage:.1f}% of files received hints")

        if not hints_added:
            insights.append("ℹ️ No hints added - files may already have hints or no external libraries detected")

        return '\n'.join(insights) if insights else "No insights generated"

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate report of automation execution."""
        hints_added = analysis_results.get('hints_added', [])
        hints_skipped = analysis_results.get('hints_skipped', [])

        report_lines = [
            "# External Tool Hints Automation Report",
            "",
            f"**Date**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
            f"**Mode**: {'DRY RUN' if self.dry_run else 'APPLIED'}",
            "",
            "## Summary",
            "",
            f"- **Files Scanned**: {analysis_results.get('files_scanned', 0)}",
            f"- **Files Modified**: {analysis_results.get('files_modified', 0)}",
            f"- **Files Skipped**: {analysis_results.get('files_skipped', 0)}",
            "",
            "## Insights",
            "",
            insights,
            "",
        ]

        if hints_added:
            report_lines.extend([
                "## Hints Added",
                "",
            ])
            for hint in hints_added:
                report_lines.append(f"- **{hint['file']}**")
                report_lines.append(f"  - Libraries: {', '.join(hint['libraries'])}")
                report_lines.append(f"  - Type: {hint['hint_type']}")
                report_lines.append("")

        if hints_skipped:
            report_lines.extend([
                "## Files Skipped (Sample)",
                "",
            ])
            for skip in hints_skipped[:20]:  # Limit to first 20
                report_lines.append(f"- **{skip['file']}**: {skip['reason']}")

        # Save report if output path specified
        if self.output_path:
            report_path = Path(self.output_path)
            report_path.parent.mkdir(parents=True, exist_ok=True)
            with open(report_path, 'w') as f:
                f.write('\n'.join(report_lines))
            logger.info(f"Report saved to: {report_path}")

        return '\n'.join(report_lines)


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Add external tool hints to documentation')
    parser.add_argument('--dry-run', action='store_true', help='Preview changes without applying')
    parser.add_argument('--output-path', type=str, help='Path for report output')
    parser.add_argument('--min-file-size', type=int, default=50, help='Minimum file size in lines')

    args = parser.parse_args()

    config = {
        'dry_run': args.dry_run,
        'output_path': args.output_path or 'docs/EXTERNAL_TOOL_HINTS_REPORT.md',
        'min_file_size': args.min_file_size
    }

    automation = ExternalToolHintsAutomation(config)
    results = automation.run()

    print(json.dumps(results, indent=2))
    return 0 if results['status'] == 'success' else 1


if __name__ == '__main__':
    sys.exit(main())
