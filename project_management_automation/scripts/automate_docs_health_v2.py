#!/usr/bin/env python3
"""
Automated Documentation Health Monitoring Script (v2 - Intelligent)

Refactored to use IntelligentAutomationBase with:
- Tractatus Thinking for structure analysis
- Sequential Thinking for workflow planning
- Todo2 integration for tracking
- NetworkX for cross-reference graph analysis
"""

import argparse
import json
import logging
import re
import subprocess
import sys
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

# Import base class (relative import for package)
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase

# Configure logging (will be configured after project_root is set)
logger = logging.getLogger(__name__)


class DocumentationHealthAnalyzerV2(IntelligentAutomationBase):
    """Intelligent documentation health analyzer using base class."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        # Detect project root if not provided
        if project_root is None:
            # Try to find project root by looking for .git, .todo2, or CMakeLists.txt
            current = Path(__file__).parent.parent.parent.parent
            while current != current.parent:
                if (current / '.git').exists() or (current / '.todo2').exists() or (current / 'CMakeLists.txt').exists():
                    project_root = current
                    break
                current = current.parent
            else:
                # Fallback to current working directory
                project_root = Path.cwd()

        super().__init__(config, "Documentation Health Analysis", project_root)

        # Configure logging after project_root is set
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler(self.project_root / 'scripts' / 'docs_health.log'),
                logging.StreamHandler()
            ],
            force=True
        )

        self.docs_path = self.project_root / 'docs'
        self.history_path = self.project_root / 'scripts' / '.docs_health_history.json'
        self.history = self._load_history()

        # Analysis results
        self.analysis_results = {
            'link_validation': {'total_links': 0, 'broken_internal': [], 'broken_external': []},
            'format_validation': {'format_errors': [], 'missing_required_fields': []},
            'date_currency': {'stale_files': [], 'missing_dates': []},
            'cross_references': {'broken_references': [], 'orphaned_files': []}
        }

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is documentation health?"""
        return "What is documentation health? Documentation Health = Link Validity × Format Compliance × Content Completeness × Currency × Cross-Reference Integrity"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we check documentation health?"""
        return "How do we systematically check documentation health across all dimensions?"

    def _load_history(self) -> dict:
        """Load historical health data."""
        if self.history_path.exists():
            try:
                with open(self.history_path) as f:
                    return json.load(f)
            except json.JSONDecodeError:
                pass
        return {'runs': []}

    def _execute_analysis(self) -> dict:
        """Execute documentation health analysis."""
        logger.info("Executing documentation health analysis...")

        # Use Tractatus components to guide analysis
        components = self.tractatus_session.get('components', [])

        # Execute checks based on components
        if 'link' in str(components).lower() or 'validity' in str(components).lower():
            self._validate_links()

        if 'format' in str(components).lower() or 'compliance' in str(components).lower():
            self._validate_format()

        if 'currency' in str(components).lower():
            self._check_date_currency()

        if 'cross' in str(components).lower() or 'reference' in str(components).lower():
            self._validate_cross_references()

        # Calculate health score
        health_score = self._calculate_health_score()
        self.analysis_results['health_score'] = health_score

        # Save history
        self._save_history()

        return self.analysis_results

    def _validate_links(self) -> None:
        """Validate all links in documentation."""
        logger.info("Validating links...")

        md_files = list(self.docs_path.rglob('*.md'))
        skip_dirs = {'archive', 'indices', 'message_schemas', 'resource-summaries', 'video-summaries'}
        md_files = [f for f in md_files if not any(skip in str(f) for skip in skip_dirs)]

        link_pattern = re.compile(r'\[([^\]]+)\]\(([^)]+)\)')
        all_internal_files = {f.relative_to(self.docs_path) for f in md_files}

        for md_file in md_files:
            try:
                content = md_file.read_text(encoding='utf-8')

                for match in link_pattern.finditer(content):
                    url = match.group(2)
                    self._validate_link(url, md_file, match.start(), all_internal_files)
            except Exception as e:
                logger.error(f"Error processing {md_file}: {e}")

    def _validate_link(self, url: str, source_file: Path, position: int, all_internal_files: set[Path]) -> None:
        """Validate a single link."""
        self.analysis_results['link_validation']['total_links'] += 1

        skip_patterns = ['mailto:', '#', 'docs/', 'github.com.*blob']
        if any(pattern in url for pattern in skip_patterns):
            return

        if not url.startswith('http://') and not url.startswith('https://'):
            # Internal link
            source_dir = source_file.parent
            target_path = (source_dir / url).resolve()

            if not target_path.exists():
                self.analysis_results['link_validation']['broken_internal'].append({
                    'url': url,
                    'source': str(source_file.relative_to(self.project_root))
                })
        else:
            # External link - simplified check (full implementation would use urllib)
            pass  # Skip external link checking for speed

    def _validate_format(self) -> None:
        """Validate documentation format."""
        logger.info("Validating format...")

        index_file = self.docs_path / 'API_DOCUMENTATION_INDEX.md'
        if index_file.exists():
            try:
                result = subprocess.run(
                    [sys.executable, str(self.project_root / 'scripts' / 'validate_docs_format.py')],
                    capture_output=True,
                    text=True,
                    cwd=self.project_root
                )

                if result.returncode != 0:
                    for line in result.stdout.split('\n'):
                        if 'Error' in line or 'Missing' in line:
                            self.analysis_results['format_validation']['format_errors'].append(line)
            except Exception as e:
                logger.error(f"Error running format validation: {e}")

    def _check_date_currency(self) -> None:
        """Check documentation currency."""
        logger.info("Checking date currency...")

        date_patterns = [
            re.compile(r'\*\*Last Updated\*\*:\s*(\d{4}-\d{2}-\d{2})', re.IGNORECASE),
            re.compile(r'Last Updated:\s*(\d{4}-\d{2}-\d{2})', re.IGNORECASE),
        ]

        stale_threshold_days = self.config.get('stale_threshold_days', 90)
        threshold_date = datetime.now() - timedelta(days=stale_threshold_days)

        for doc_file in self.docs_path.glob('*.md'):
            try:
                content = doc_file.read_text(encoding='utf-8')

                for pattern in date_patterns:
                    match = pattern.search(content)
                    if match:
                        date_str = match.group(1)
                        try:
                            doc_date = datetime.strptime(date_str, '%Y-%m-%d')
                            if doc_date < threshold_date:
                                self.analysis_results['date_currency']['stale_files'].append({
                                    'file': str(doc_file.relative_to(self.project_root)),
                                    'last_updated': date_str,
                                    'days_old': (datetime.now() - doc_date).days
                                })
                        except ValueError:
                            pass
                        break
            except Exception:
                continue

    def _validate_cross_references(self) -> None:
        """Validate cross-references (will use NetworkX in _networkx_analysis)."""
        logger.info("Validating cross-references...")
        # Actual validation happens in NetworkX analysis
        pass

    def _calculate_health_score(self) -> float:
        """Calculate overall health score."""
        total_issues = (
            len(self.analysis_results['link_validation']['broken_internal']) +
            len(self.analysis_results['link_validation']['broken_external']) +
            len(self.analysis_results['format_validation']['format_errors']) +
            len(self.analysis_results['date_currency']['stale_files']) +
            len(self.analysis_results['cross_references']['broken_references']) +
            len(self.analysis_results['cross_references']['orphaned_files'])
        )

        # Simple scoring (can be enhanced)
        max_issues = 100
        score = max(0, 100 - (total_issues / max_issues * 100))
        return round(score, 1)

    def _save_history(self) -> None:
        """Save current run to history."""
        if 'runs' not in self.history:
            self.history['runs'] = []

        run_data = {
            'timestamp': datetime.now().isoformat(),
            'health_score': self.analysis_results.get('health_score', 0),
            'link_validation': {
                'broken': len(self.analysis_results['link_validation']['broken_internal']) +
                         len(self.analysis_results['link_validation']['broken_external'])
            },
            'format_errors': len(self.analysis_results['format_validation']['format_errors']),
            'stale_files': len(self.analysis_results['date_currency']['stale_files'])
        }

        self.history['runs'].append(run_data)
        if len(self.history['runs']) > 50:
            self.history['runs'] = self.history['runs'][-50:]

        with open(self.history_path, 'w') as f:
            json.dump(self.history, f, indent=2)

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from analysis."""
        insights = []

        health_score = analysis_results.get('health_score', 0)
        insights.append(f"**Documentation Health Score: {health_score}%**")

        broken_links = (
            len(analysis_results['link_validation']['broken_internal']) +
            len(analysis_results['link_validation']['broken_external'])
        )
        if broken_links > 0:
            insights.append(f"⚠️ {broken_links} broken links need attention")

        stale = len(analysis_results['date_currency']['stale_files'])
        if stale > 0:
            insights.append(f"⚠️ {stale} documents are stale (>90 days)")

        format_errors = len(analysis_results['format_validation']['format_errors'])
        if format_errors > 0:
            insights.append(f"⚠️ {format_errors} format errors found")

        if health_score >= 80:
            insights.append("✅ Documentation health is good!")

        return '\n'.join(insights)

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate comprehensive report."""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        health_score = analysis_results.get('health_score', 0)

        # Include NetworkX analysis if available
        networkx_section = ""
        if 'networkx_analysis' in self.results:
            nx_analysis = self.results['networkx_analysis']
            networkx_section = f"""
## NetworkX Analysis

**Graph Statistics:**
- Nodes: {nx_analysis.get('nodes', 0)}
- Edges: {nx_analysis.get('edges', 0)}
- Density: {nx_analysis.get('density', 0):.3f}
- Is DAG: {nx_analysis.get('is_dag', False)}

**Critical Path:** {len(nx_analysis.get('critical_path', []))} nodes
**Bottlenecks:** {len(nx_analysis.get('bottlenecks', []))} identified
**Orphaned Documents:** {len(nx_analysis.get('orphans', []))} found

"""

        return f"""# Documentation Health Report

*Generated: {timestamp}*
*Generated By: Intelligent Documentation Health Analyzer*

## Executive Summary

**Overall Health Score: {health_score}%** {'✅' if health_score >= 80 else '⚠️' if health_score >= 60 else '❌'}

**Key Metrics:**
- Broken Links: {len(analysis_results['link_validation']['broken_internal']) + len(analysis_results['link_validation']['broken_external'])}
- Format Errors: {len(analysis_results['format_validation']['format_errors'])}
- Stale Documents: {len(analysis_results['date_currency']['stale_files'])}
- Cross-Reference Issues: {len(analysis_results['cross_references']['broken_references']) + len(analysis_results['cross_references']['orphaned_files'])}

---

## Insights

{insights}

{networkx_section}
---

## Detailed Results

### Link Validation
- Total Links Checked: {analysis_results['link_validation']['total_links']}
- Broken Internal: {len(analysis_results['link_validation']['broken_internal'])}
- Broken External: {len(analysis_results['link_validation']['broken_external'])}

### Format Validation
- Format Errors: {len(analysis_results['format_validation']['format_errors'])}

### Date Currency
- Stale Files: {len(analysis_results['date_currency']['stale_files'])}

### Cross-References
- Broken References: {len(analysis_results['cross_references']['broken_references'])}
- Orphaned Files: {len(analysis_results['cross_references']['orphaned_files'])}

---

*This report was generated using intelligent automation with Tractatus Thinking, Sequential Thinking, and NetworkX analysis.*
"""

    def _needs_networkx(self) -> bool:
        """NetworkX is useful for cross-reference analysis."""
        return True

    def _build_networkx_graph(self):
        """Build documentation cross-reference graph."""
        try:
            import networkx as nx

            G = nx.DiGraph()

            # Find all markdown files
            md_files = list(self.docs_path.rglob('*.md'))
            skip_dirs = {'archive', 'indices', 'message_schemas', 'resource-summaries', 'video-summaries'}
            md_files = [f for f in md_files if not any(skip in str(f) for skip in skip_dirs)]

            # Build reference map
            referenced_files = set()
            link_pattern = re.compile(r'\[([^\]]+)\]\(([^)]+)\)')

            for md_file in md_files:
                try:
                    content = md_file.read_text(encoding='utf-8')
                    rel_path = str(md_file.relative_to(self.docs_path))
                    G.add_node(rel_path)

                    for match in link_pattern.finditer(content):
                        url = match.group(2)

                        if not url.startswith('http') and not url.startswith('mailto') and not url.startswith('#'):
                            source_dir = md_file.parent
                            target_path = (source_dir / url).resolve()

                            if target_path.exists() and target_path.suffix == '.md':
                                target_rel = str(target_path.relative_to(self.docs_path))
                                G.add_edge(rel_path, target_rel)
                                referenced_files.add(target_rel)
                            else:
                                self.analysis_results['cross_references']['broken_references'].append({
                                    'source': rel_path,
                                    'target': url
                                })
                except Exception:
                    continue

            # Find orphaned files
            all_files = {str(f.relative_to(self.docs_path)) for f in md_files}
            orphaned = all_files - referenced_files

            # Filter out index files
            self.analysis_results['cross_references']['orphaned_files'] = [
                f for f in orphaned
                if not any(skip in f for skip in ['INDEX', 'README', 'SUMMARY', 'TEMPLATE'])
            ]

            return G
        except ImportError:
            return None

    def _identify_followup_tasks(self, analysis_results: dict) -> list[dict]:
        """Identify follow-up tasks."""
        followups = []

        # Create tasks for critical issues
        broken_links = len(analysis_results['link_validation']['broken_internal']) + len(analysis_results['link_validation']['broken_external'])
        if broken_links > 10:
            followups.append({
                'name': 'Fix broken documentation links',
                'description': f'Fix {broken_links} broken links in documentation',
                'priority': 'high',
                'tags': ['documentation', 'fix', 'links']
            })

        stale = len(analysis_results['date_currency']['stale_files'])
        if stale > 5:
            followups.append({
                'name': 'Update stale documentation',
                'description': f'Update {stale} stale documents (>90 days old)',
                'priority': 'medium',
                'tags': ['documentation', 'update']
            })

        return followups


def load_config(config_path: Optional[Path] = None) -> dict:
    """Load configuration."""
    if config_path is None:
        # Find project root
        from project_management_automation.utils import find_project_root
        project_root = find_project_root(Path(__file__).parent.parent.parent)
        config_path = project_root / 'project_management_automation' / 'scripts' / 'docs_health_config.json'

    default_config = {
        'stale_threshold_days': 90,
        'output_path': 'docs/DOCUMENTATION_HEALTH_REPORT.md'
    }

    if config_path.exists():
        try:
            with open(config_path) as f:
                user_config = json.load(f)
                default_config.update(user_config)
        except json.JSONDecodeError:
            pass

    return default_config


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Intelligent Documentation Health Analysis')
    parser.add_argument('--config', type=Path, help='Path to config file')
    parser.add_argument('--output', type=Path, help='Output path for report')
    args = parser.parse_args()

    # Find project root
    from project_management_automation.utils import find_project_root
    project_root = find_project_root(Path(__file__).parent.parent.parent)

    config = load_config(args.config)
    analyzer = DocumentationHealthAnalyzerV2(config, project_root)

    try:
        results = analyzer.run()

        # Write report
        if args.output:
            output_path = args.output
        else:
            output_path = project_root / config['output_path']

        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, 'w') as f:
            f.write(results['report'])

        logger.info(f"Report written to: {output_path}")
        logger.info(f"Health score: {results.get('results', {}).get('health_score', 0)}%")

        sys.exit(0)
    except Exception as e:
        logger.error(f"Error running analysis: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()
