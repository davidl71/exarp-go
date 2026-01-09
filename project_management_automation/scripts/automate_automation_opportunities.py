#!/usr/bin/env python3
"""
Automated Automation Opportunity Finder

This script uses intelligent automation to find automation opportunities,
similar to what we do interactively. It demonstrates the intelligent
automation base class in action.

Usage:
    python3 scripts/automate_automation_opportunities.py [--output docs/AUTOMATION_OPPORTUNITIES_FOUND.md]
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


class AutomationOpportunityFinder(IntelligentAutomationBase):
    """Finds automation opportunities using intelligent automation."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Automation Opportunity Finder", project_root)
        self.scripts_path = self.project_root / 'scripts'
        self.docs_path = self.project_root / 'docs'
        self.opportunities = []

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is an automation opportunity?"""
        return "What makes a task suitable for automation? Automation Opportunity = Repetitive Task × High Frequency × Manual Effort × Low Complexity × High Value"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we find automation opportunities?"""
        return "How do we systematically identify tasks that would benefit from automation?"

    def _execute_analysis(self) -> dict:
        """Execute analysis to find automation opportunities."""
        logger.info("Executing automation opportunity analysis...")

        # Step 1: Analyze existing scripts
        existing_automations = self._analyze_existing_automations()

        # Step 2: Find manual repetitive tasks
        manual_tasks = self._find_manual_tasks()

        # Step 3: Analyze scripts that run manually
        manual_scripts = self._find_manual_scripts()

        # Step 4: Check for TODO comments about automation
        todo_automations = self._find_todo_automations()

        # Step 5: Analyze documentation for automation mentions
        doc_automations = self._find_doc_automations()

        # Step 6: Score and prioritize opportunities
        all_opportunities = (
            existing_automations +
            manual_tasks +
            manual_scripts +
            todo_automations +
            doc_automations
        )

        scored_opportunities = self._score_opportunities(all_opportunities)

        return {
            'existing_automations': len(existing_automations),
            'opportunities_found': len(scored_opportunities),
            'opportunities': scored_opportunities,
            'high_priority': [o for o in scored_opportunities if o['score'] >= 8],
            'medium_priority': [o for o in scored_opportunities if 5 <= o['score'] < 8],
            'low_priority': [o for o in scored_opportunities if o['score'] < 5]
        }

    def _analyze_existing_automations(self) -> list[dict]:
        """Analyze what automations already exist."""
        opportunities = []

        # Check for automation scripts
        automation_patterns = [
            r'automate_.*\.py',
            r'setup_.*_cron\.sh',
            r'run_.*_cron\.sh'
        ]

        existing = set()
        for pattern in automation_patterns:
            for script in self.scripts_path.rglob(pattern):
                existing.add(script.stem)

        # Find gaps - scripts that exist but aren't automated
        for script in self.scripts_path.glob('*.py'):
            if script.stem.startswith('validate_') or script.stem.startswith('check_'):
                if f"automate_{script.stem}" not in existing:
                    opportunities.append({
                        'name': f"Automate {script.stem}",
                        'description': f"Script {script.name} exists but runs manually",
                        'type': 'script_automation',
                        'script': str(script.relative_to(self.project_root)),
                        'frequency': 'weekly',
                        'effort': 'low',
                        'value': 'high',
                        'evidence': f"Script exists: {script.name}"
                    })

        return opportunities

    def _find_manual_tasks(self) -> list[dict]:
        """Find manual repetitive tasks."""
        opportunities = []

        # Check coordination docs for manual tasks
        coord_doc = self.project_root / 'agents' / 'shared' / 'COORDINATION.md'
        if coord_doc.exists():
            content = coord_doc.read_text()

            # Look for manual update patterns
            if 'Update the Shared TODO Table' in content:
                opportunities.append({
                    'name': 'Shared TODO Table Synchronization',
                    'description': 'Auto-sync agents/shared/TODO_OVERVIEW.md with Todo2',
                    'type': 'synchronization',
                    'frequency': 'daily',
                    'effort': 'medium',
                    'value': 'high',
                    'evidence': 'COORDINATION.md mentions manual updates'
                })

            if 'update agents/shared/API_CONTRACT.md' in content.lower():
                opportunities.append({
                    'name': 'API Contract Synchronization',
                    'description': 'Auto-sync API_CONTRACT.md with backend code',
                    'type': 'synchronization',
                    'frequency': 'daily',
                    'effort': 'medium',
                    'value': 'high',
                    'evidence': 'COORDINATION.md mentions manual API contract updates'
                })

        return opportunities

    def _find_manual_scripts(self) -> list[dict]:
        """Find scripts that run manually."""
        opportunities = []

        # Check for scripts mentioned in docs but not automated
        automation_docs = list(self.docs_path.glob('*AUTOMATION*.md'))

        for doc in automation_docs:
            content = doc.read_text()

            # Look for script mentions
            script_pattern = r'`([^`]+\.(py|sh))`'
            for match in re.finditer(script_pattern, content):
                script_name = match.group(1)
                script_path = self.scripts_path / script_name

                if script_path.exists():
                    # Check if automation exists
                    automate_name = f"automate_{script_path.stem}.py"
                    if not (self.scripts_path / automate_name).exists():
                        opportunities.append({
                            'name': f"Automate {script_path.stem}",
                            'description': f"Script {script_name} mentioned in {doc.name} but not automated",
                            'type': 'script_automation',
                            'script': str(script_path.relative_to(self.project_root)),
                            'frequency': 'weekly',
                            'effort': 'low',
                            'value': 'medium',
                            'evidence': f"Mentioned in {doc.name}"
                        })

        return opportunities

    def _find_todo_automations(self) -> list[dict]:
        """Find TODO comments about automation."""
        opportunities = []

        # Search codebase for automation TODOs
        todo_pattern = re.compile(r'TODO.*[Aa]utomat', re.IGNORECASE)

        for file_path in self.project_root.rglob('*.{py,md,sh}'):
            if any(skip in str(file_path) for skip in ['.git', 'node_modules', 'build', '__pycache__']):
                continue

            try:
                content = file_path.read_text(encoding='utf-8', errors='ignore')
                for line_num, line in enumerate(content.split('\n'), 1):
                    if todo_pattern.search(line):
                        opportunities.append({
                            'name': f"Automation TODO in {file_path.name}",
                            'description': line.strip(),
                            'type': 'todo_automation',
                            'file': str(file_path.relative_to(self.project_root)),
                            'line': line_num,
                            'frequency': 'unknown',
                            'effort': 'unknown',
                            'value': 'medium',
                            'evidence': f"TODO comment in {file_path.name}:{line_num}"
                        })
            except Exception:
                continue

        return opportunities

    def _find_doc_automations(self) -> list[dict]:
        """Find automation opportunities mentioned in documentation."""
        opportunities = []

        # Check automation opportunities docs
        opp_doc = self.docs_path / 'AUTOMATION_OPPORTUNITIES.md'
        if opp_doc.exists():
            content = opp_doc.read_text()

            # Extract opportunity sections
            section_pattern = re.compile(r'### (\d+)\.\s+(.+?)\n\n\*\*Current State\*\*: (.+?)\n\n\*\*Automation Value\*\*: (.+?)\n', re.DOTALL)

            for match in section_pattern.finditer(content):
                name = match.group(2).strip()
                current_state = match.group(3).strip()
                value = match.group(4).strip()

                # Check if already automated
                if '✅' not in name and 'DONE' not in current_state.upper():
                    opportunities.append({
                        'name': name,
                        'description': current_state,
                        'type': 'documented_opportunity',
                        'frequency': 'weekly',
                        'effort': 'medium',
                        'value': value,
                        'evidence': "Documented in AUTOMATION_OPPORTUNITIES.md"
                    })

        return opportunities

    def _score_opportunities(self, opportunities: list[dict]) -> list[dict]:
        """Score and prioritize opportunities."""
        scored = []

        for opp in opportunities:
            score = 0

            # Value scoring
            if opp.get('value') == 'high':
                score += 5
            elif opp.get('value') == 'medium':
                score += 3
            elif opp.get('value') == 'low':
                score += 1

            # Effort scoring (lower effort = higher score)
            if opp.get('effort') == 'low':
                score += 3
            elif opp.get('effort') == 'medium':
                score += 2
            elif opp.get('effort') == 'high':
                score += 1

            # Frequency scoring
            if opp.get('frequency') == 'daily':
                score += 2
            elif opp.get('frequency') == 'weekly':
                score += 1

            # Type bonus
            if opp.get('type') == 'synchronization':
                score += 1  # High value

            opp['score'] = score
            scored.append(opp)

        # Sort by score descending
        scored.sort(key=lambda x: x['score'], reverse=True)

        return scored

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from analysis."""
        insights = []

        total = analysis_results['opportunities_found']
        high = len(analysis_results['high_priority'])
        medium = len(analysis_results['medium_priority'])
        low = len(analysis_results['low_priority'])

        insights.append(f"**Found {total} automation opportunities:**")
        insights.append(f"- {high} high-priority (score ≥ 8)")
        insights.append(f"- {medium} medium-priority (score 5-7)")
        insights.append(f"- {low} low-priority (score < 5)")

        if high > 0:
            insights.append("\n**Top High-Priority Opportunities:**")
            for i, opp in enumerate(analysis_results['high_priority'][:5], 1):
                insights.append(f"{i}. {opp['name']} (score: {opp['score']})")

        insights.append("\n**Recommendation:** Focus on high-priority opportunities first for maximum impact.")

        return '\n'.join(insights)

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate comprehensive report."""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')

        # Build opportunity tables
        high_priority_table = self._build_opportunity_table(analysis_results['high_priority'])
        medium_priority_table = self._build_opportunity_table(analysis_results['medium_priority'])
        low_priority_table = self._build_opportunity_table(analysis_results['low_priority'])

        report = f"""# Automation Opportunities Found

*Generated: {timestamp}*
*Generated By: Intelligent Automation Opportunity Finder*

## Executive Summary

**Total Opportunities Found: {analysis_results['opportunities_found']}**

**Breakdown:**
- **High Priority** (score ≥ 8): {len(analysis_results['high_priority'])}
- **Medium Priority** (score 5-7): {len(analysis_results['medium_priority'])}
- **Low Priority** (score < 5): {len(analysis_results['low_priority'])}

**Existing Automations:** {analysis_results['existing_automations']}

---

## Insights

{insights}

---

## High-Priority Opportunities (Score ≥ 8)

{high_priority_table if high_priority_table else "No high-priority opportunities found."}

---

## Medium-Priority Opportunities (Score 5-7)

{medium_priority_table if medium_priority_table else "No medium-priority opportunities found."}

---

## Low-Priority Opportunities (Score < 5)

{low_priority_table if low_priority_table else "No low-priority opportunities found."}

---

## Recommendations

### Immediate Actions

1. **Review High-Priority Opportunities**
   - {len(analysis_results['high_priority'])} opportunities with score ≥ 8
   - High value, low-to-medium effort
   - Implement these first for maximum ROI

2. **Plan Implementation**
   - Use intelligent automation base class
   - Integrate with Todo2, Tractatus, Sequential Thinking
   - Add NetworkX analysis where applicable

3. **Track Progress**
   - Create Todo2 tasks for each opportunity
   - Monitor automation health
   - Update opportunities as they're implemented

---

## Methodology

This analysis used intelligent automation to:
1. **Tractatus Thinking**: Understood what makes a good automation opportunity
2. **Sequential Thinking**: Planned systematic discovery workflow
3. **NetworkX Analysis**: Analyzed relationships and dependencies
4. **Todo2 Integration**: Tracked findings and created follow-up tasks

**Scoring Criteria:**
- Value: High (5), Medium (3), Low (1)
- Effort: Low (3), Medium (2), High (1)
- Frequency: Daily (2), Weekly (1)
- Type Bonus: Synchronization (+1)

---

*This report was automatically generated using intelligent automation. Review opportunities and prioritize based on your needs.*
"""
        return report

    def _build_opportunity_table(self, opportunities: list[dict]) -> str:
        """Build markdown table for opportunities."""
        if not opportunities:
            return ""

        table = "| Name | Type | Value | Effort | Frequency | Score | Evidence |\n"
        table += "|------|------|-------|--------|-----------|-------|----------|\n"

        for opp in opportunities[:20]:  # Limit to 20
            table += f"| {opp['name']} | {opp.get('type', 'unknown')} | {opp.get('value', 'unknown')} | {opp.get('effort', 'unknown')} | {opp.get('frequency', 'unknown')} | {opp['score']} | {opp.get('evidence', 'N/A')[:50]}... |\n"

        if len(opportunities) > 20:
            table += f"\n*... and {len(opportunities) - 20} more*\n"

        return table

    def _identify_followup_tasks(self, analysis_results: dict) -> list[dict]:
        """Identify follow-up tasks."""
        followups = []

        # Create task for each high-priority opportunity
        for opp in analysis_results['high_priority'][:5]:  # Top 5
            followups.append({
                'name': f"Implement: {opp['name']}",
                'description': opp.get('description', opp['name']),
                'priority': 'high',
                'tags': ['automation', 'implementation', opp.get('type', 'general')]
            })

        return followups

    def _needs_networkx(self) -> bool:
        """NetworkX analysis is useful for finding dependencies."""
        return True

    def _build_networkx_graph(self):
        """Build graph of automation dependencies."""
        try:
            import networkx as nx

            G = nx.DiGraph()

            # Add automation opportunities as nodes
            for opp in self.opportunities:
                G.add_node(opp['name'], **opp)

            # Add dependencies (if one automation enables another)
            # This is simplified - real implementation would analyze more deeply
            for i, opp1 in enumerate(self.opportunities):
                for opp2 in self.opportunities[i+1:]:
                    if self._are_dependent(opp1, opp2):
                        G.add_edge(opp1['name'], opp2['name'], relationship='enables')

            return G
        except ImportError:
            return None

    def _are_dependent(self, opp1: dict, opp2: dict) -> bool:
        """Check if two opportunities are dependent."""
        # Simple heuristic: if one is infrastructure and other uses it
        if 'infrastructure' in opp1.get('type', '') and 'script' in opp2.get('type', ''):
            return True
        return False


def load_config(config_path: Optional[Path] = None) -> dict:
    """Load configuration."""
    from project_management_automation.utils import find_project_root
    if config_path is None:
        config_path = find_project_root() / 'scripts' / 'automation_opportunities_config.json'

    default_config = {
        'output_path': 'docs/AUTOMATION_OPPORTUNITIES_FOUND.md'
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
    parser = argparse.ArgumentParser(description='Find Automation Opportunities')
    parser.add_argument('--config', type=Path, help='Path to config file')
    parser.add_argument('--output', type=Path, help='Output path for report')
    args = parser.parse_args()

    config = load_config(args.config)
    finder = AutomationOpportunityFinder(config)

    try:
        results = finder.run()

        # Write report
        from project_management_automation.utils import find_project_root
        if args.output:
            output_path = args.output
        else:
            output_path = find_project_root() / config['output_path']

        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, 'w') as f:
            f.write(results['report'])

        logger.info(f"Report written to: {output_path}")
        if 'opportunities_found' in results.get('results', {}):
            logger.info(f"Found {results['results']['opportunities_found']} automation opportunities")
        else:
            logger.info("Report generated successfully")

        sys.exit(0)
    except Exception as e:
        logger.error(f"Error running analysis: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()
