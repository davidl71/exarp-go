#!/usr/bin/env python3
"""
Automated Todo2 Task Alignment Analysis Script (v2 - Intelligent)

Refactored to use IntelligentAutomationBase with:
- Tractatus Thinking for structure analysis
- Sequential Thinking for workflow planning
- Todo2 integration for tracking
- NetworkX for task dependency graph analysis
- Generic PROJECT_GOALS.md support for any project type
"""

import argparse
import json
import logging
import re
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

# Add project root to path
# Project root will be passed to __init__
# Import base class
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase
from project_management_automation.utils.todo2_utils import (
    filter_tasks_by_project,
    get_repo_project_id,
)

# Configure logging (will be configured after project_root is set)
logger = logging.getLogger(__name__)


class Todo2AlignmentAnalyzerV2(IntelligentAutomationBase):
    """Intelligent Todo2 alignment analyzer using base class."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root

        if project_root is None:
            project_root = find_project_root()

        super().__init__(config, "Todo2 Alignment Analysis", project_root)

        # Configure logging after project_root is set
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler(self.project_root / 'scripts' / 'todo2_alignment.log'),
                logging.StreamHandler()
            ],
            force=True
        )
        # Support both agentic-tools MCP format (preferred) and legacy Todo2 format
        self.agentic_tools_path = self.project_root / '.agentic-tools-mcp' / 'tasks' / 'tasks.json'
        self.todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
        self.docs_path = self.project_root / 'docs'

        # Load goals from PROJECT_GOALS.md (generic) or fall back to defaults
        self.goals_path = self.project_root / 'PROJECT_GOALS.md'
        self.strategy_phases, self.infrastructure_keywords = self._load_project_goals()

    def _load_project_goals(self) -> tuple:
        """Load project goals from PROJECT_GOALS.md file.

        Returns:
            Tuple of (strategy_phases dict, infrastructure_keywords list)
        """
        default_infrastructure = [
            'research', 'analysis', 'review', 'investigation',
            'config', 'configuration', 'setup', 'infrastructure',
            'testing', 'test', 'unittest', 'pytest',
            'documentation', 'docs', 'readme',
            'build', 'package', 'release', 'version',
            'refactor', 'cleanup', 'optimization',
            'migration', 'upgrade', 'deprecation'
        ]

        if not self.goals_path.exists():
            logger.warning(f"PROJECT_GOALS.md not found at {self.goals_path}, using defaults")
            return self._get_default_phases(), default_infrastructure

        try:
            with open(self.goals_path) as f:
                content = f.read()

            phases = {}
            infrastructure = default_infrastructure.copy()

            # Parse phases: ### Phase N: Name
            phase_pattern = r'###\s+Phase\s+(\d+):\s+([^\n]+)'
            keyword_pattern = r'\*\*Keywords\*\*:\s*([^\n]+)'

            # Find all phase sections
            phase_matches = list(re.finditer(phase_pattern, content))

            for i, match in enumerate(phase_matches):
                phase_num = match.group(1)
                phase_name = match.group(2).strip()
                phase_key = f'phase{phase_num}'

                # Find keywords after this phase header
                start_pos = match.end()
                end_pos = phase_matches[i + 1].start() if i + 1 < len(phase_matches) else len(content)
                section_content = content[start_pos:end_pos]

                keyword_match = re.search(keyword_pattern, section_content)
                if keyword_match:
                    keywords_str = keyword_match.group(1).strip()
                    keywords = [k.strip().lower() for k in keywords_str.split(',')]
                    phases[phase_key] = {
                        'name': phase_name,
                        'keywords': keywords
                    }
                    logger.debug(f"Loaded {phase_key}: {phase_name} with {len(keywords)} keywords")

            # Parse infrastructure keywords section
            infra_section = re.search(
                r'##\s+Infrastructure Keywords[^\n]*\n(.*?)(?=\n##|\Z)',
                content,
                re.DOTALL | re.IGNORECASE
            )
            if infra_section:
                # Extract bullet points or comma-separated keywords
                infra_content = infra_section.group(1)
                # Look for bullet points
                bullet_keywords = re.findall(r'-\s*([^,\n]+(?:,\s*[^,\n]+)*)', infra_content)
                for bullet in bullet_keywords:
                    for kw in bullet.split(','):
                        kw = kw.strip().lower()
                        if kw and kw not in infrastructure:
                            infrastructure.append(kw)

            if phases:
                logger.info(f"Loaded {len(phases)} phases from PROJECT_GOALS.md")
                return phases, infrastructure
            else:
                logger.warning("No phases found in PROJECT_GOALS.md, using defaults")
                return self._get_default_phases(), infrastructure

        except Exception as e:
            logger.error(f"Error loading PROJECT_GOALS.md: {e}")
            return self._get_default_phases(), default_infrastructure

    def _get_current_tool_count(self) -> int:
        """Get current tool count from tool_count_health module."""
        try:
            from project_management_automation.tools.tool_count_health import _count_registered_tools
            tool_info = _count_registered_tools()
            return tool_info.get('count', 0)
        except Exception as e:
            logger.debug(f"Could not get tool count: {e}")
            # Fallback: try to count from server.py or estimate
            return 30  # Default estimate

    def _get_default_phases(self) -> dict:
        """Return default phases for generic projects."""
        return {
            'phase1': {
                'name': 'Foundation',
                'keywords': ['setup', 'infrastructure', 'configuration', 'foundation', 'core'],
            },
            'phase2': {
                'name': 'Core Features',
                'keywords': ['feature', 'implementation', 'integration', 'api', 'service'],
            },
            'phase3': {
                'name': 'Enhancement',
                'keywords': ['enhancement', 'optimization', 'improvement', 'refactor'],
            },
            'phase4': {
                'name': 'Quality',
                'keywords': ['testing', 'quality', 'coverage', 'ci', 'cd', 'validation'],
            },
            'phase5': {
                'name': 'Documentation',
                'keywords': ['documentation', 'docs', 'guide', 'readme', 'api'],
            }
        }

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is task alignment?"""
        return "What is task alignment? Task Alignment = Project Goals Relevance × Priority Match × Dependency Completeness × Currency"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we analyze task alignment?"""
        return "How do we systematically analyze Todo2 task alignment with project goals framework?"

    def _execute_analysis(self) -> dict:
        """Execute Todo2 alignment analysis."""
        logger.info("Executing Todo2 alignment analysis...")

        # Load tasks
        tasks = self._load_todo2_tasks()
        logger.info(f"Loaded {len(tasks)} tasks")

        # Analyze alignment
        analysis = self._analyze_task_alignment(tasks)

        # Calculate alignment score
        alignment_score = self._calculate_alignment_score(analysis)
        analysis['alignment_score'] = alignment_score

        return analysis

    def _normalize_priority(self, priority) -> str:
        """Normalize priority to text format (high/medium/low/critical)."""
        if isinstance(priority, int):
            # agentic-tools uses 1-10 scale
            if priority >= 10:
                return 'critical'
            elif priority >= 8:
                return 'high'
            elif priority >= 5:
                return 'medium'
            else:
                return 'low'
        elif isinstance(priority, str):
            return priority.lower()
        return 'medium'

    def _normalize_task(self, task: dict) -> dict:
        """Normalize agentic-tools task format to expected format."""
        return {
            'id': task.get('id', 'unknown'),
            'content': task.get('name', task.get('content', '')),
            'long_description': task.get('details', task.get('long_description', '')),
            'priority': self._normalize_priority(task.get('priority', 5)),
            'status': task.get('status', 'pending').replace('-', '_'),  # in-progress -> in_progress
            'tags': task.get('tags', []),
            'completed': task.get('completed', False),
        }

    def _load_todo2_tasks(self) -> list[dict]:
        """Load tasks from Todo2 MCP (preferred), agentic-tools MCP, or legacy Todo2 format.

        Priority:
        1. Todo2 MCP server (via todo2_mcp_client)
        2. agentic-tools MCP format (with retry)
        3. Legacy Todo2 format (with retry)
        """
        from .base.mcp_client import load_json_with_retry
        from project_management_automation.utils.todo2_mcp_client import list_todos_mcp

        project_id = get_repo_project_id(self.project_root)

        # Try Todo2 MCP first (preferred)
        try:
            mcp_tasks = list_todos_mcp(project_root=self.project_root)
            if mcp_tasks:
                # Normalize MCP tasks to expected format
                tasks = [self._normalize_task(t) for t in mcp_tasks]
                filtered = filter_tasks_by_project(tasks, project_id, logger=logger)
                logger.info(
                    "Loaded %d tasks from Todo2 MCP (%d matched project)",
                    len(tasks),
                    len(filtered),
                )
                return filtered
        except Exception as e:
            logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")

        # Fall back to agentic-tools MCP format with retry
        data = load_json_with_retry(self.agentic_tools_path, default=None)
        if data is not None:
            raw_tasks = data.get('tasks', [])
            tasks = [self._normalize_task(t) for t in raw_tasks]
            filtered = filter_tasks_by_project(tasks, project_id, logger=logger)
            logger.info(
                "Loaded %d tasks from agentic-tools MCP (%d matched project)",
                len(tasks),
                len(filtered),
            )
            return filtered

        # Fall back to legacy Todo2 format with retry
        data = load_json_with_retry(self.todo2_path, default=None)
        if data is not None:
            tasks = data.get('todos', [])
            filtered = filter_tasks_by_project(tasks, project_id, logger=logger)
            logger.info(
                "Loaded %d tasks from legacy Todo2 format (%d matched project)",
                len(tasks),
                len(filtered),
            )
            return filtered

        logger.info("No task files found - no tasks to analyze")
        return []

    def _analyze_task_alignment(self, tasks: list[dict]) -> dict:
        """Analyze task alignment."""
        # Get current tool count to check constraint
        current_tool_count = self._get_current_tool_count()
        tool_limit = 30
        
        analysis = {
            'total_tasks': len(tasks),
            'by_priority': {'high': 0, 'medium': 0, 'low': 0, 'critical': 0},
            'by_status': {'todo': 0, 'in_progress': 0, 'review': 0, 'done': 0},
            'by_phase': {phase: {'total': 0, 'high_priority': 0, 'aligned': 0}
                        for phase in self.strategy_phases},
            'strategy_critical': [],
            'misaligned_tasks': [],
            'stale_tasks': [],
            'blocked_tasks': [],
            'infrastructure_tasks': [],
            'constraint_violations': [],
            'current_tool_count': current_tool_count,
            'tool_limit': tool_limit
        }

        for task in tasks:
            content = str(task.get('content', '')).lower()
            long_desc = str(task.get('long_description', '')).lower()
            tags = [tag.lower() for tag in task.get('tags', [])]
            priority = task.get('priority', 'medium').lower()
            status = task.get('status', 'todo').lower()
            task_id = task.get('id', 'unknown')

            # Count by priority
            if priority in analysis['by_priority']:
                analysis['by_priority'][priority] += 1

            # Count by status
            if 'todo' in status:
                analysis['by_status']['todo'] += 1
            elif 'progress' in status:
                analysis['by_status']['in_progress'] += 1
            elif 'review' in status:
                analysis['by_status']['review'] += 1
            elif 'done' in status:
                analysis['by_status']['done'] += 1

            # Check alignment with strategy phases
            task_text = f"{content} {long_desc} {' '.join(tags)}"
            aligned_phases = []

            for phase_key, phase_info in self.strategy_phases.items():
                if any(keyword in task_text for keyword in phase_info['keywords']):
                    aligned_phases.append(phase_key)
                    analysis['by_phase'][phase_key]['total'] += 1
                    if priority == 'high':
                        analysis['by_phase'][phase_key]['high_priority'] += 1
                    analysis['by_phase'][phase_key]['aligned'] += 1

            # Identify strategy-critical tasks
            if aligned_phases and priority == 'high':
                analysis['strategy_critical'].append({
                    'id': task_id,
                    'content': task.get('content', ''),
                    'phases': aligned_phases,
                    'priority': priority,
                    'status': status
                })

            # Identify misaligned or infrastructure tasks
            if priority == 'high' and not aligned_phases:
                # Use infrastructure keywords loaded from PROJECT_GOALS.md
                if any(keyword in task_text for keyword in self.infrastructure_keywords):
                    analysis['infrastructure_tasks'].append({
                        'id': task_id,
                        'content': task.get('content', ''),
                        'priority': priority,
                        'status': status
                    })
                else:
                    analysis['misaligned_tasks'].append({
                        'id': task_id,
                        'content': task.get('content', ''),
                        'priority': priority,
                        'status': status
                    })

            # Check for stale tasks
            last_modified = task.get('lastModified', '')
            if last_modified:
                try:
                    modified_date = datetime.fromisoformat(last_modified.replace('Z', '+00:00'))
                    days_old = (datetime.now(timezone.utc) - modified_date).days
                    if days_old > 30 and status not in ['done', 'cancelled']:
                        analysis['stale_tasks'].append({
                            'id': task_id,
                            'content': task.get('content', ''),
                            'days_old': days_old,
                            'status': status
                        })
                except Exception:
                    pass

            # Check for blocked tasks
            dependencies = task.get('dependencies', [])
            if dependencies:
                dep_tasks = {t.get('id'): t for t in tasks}
                blocked = False
                for dep_id in dependencies:
                    dep_task = dep_tasks.get(dep_id)
                    if dep_task and dep_task.get('status', '').lower() not in ['done', 'completed']:
                        blocked = True
                        break
                if blocked:
                    analysis['blocked_tasks'].append({
                        'id': task_id,
                        'content': task.get('content', ''),
                        'dependencies': dependencies,
                        'status': status
                    })

            # Check Tool Count Limit constraint (≤30 tools)
            # Detect tasks that would create new tools
            tool_creation_keywords = [
                'new tool', 'create tool', 'add tool', 'implement tool', 'register tool',
                'tool for', 'mcp tool', 'tool function', 'tool handler', 'tool endpoint',
                '@mcp.tool', 'register.*tool', 'def.*tool', 'tool.*function'
            ]
            task_text_lower = task_text.lower()
            would_create_tool = any(keyword in task_text_lower for keyword in tool_creation_keywords)
            
            if would_create_tool and status not in ['done', 'completed']:
                # Check if this would violate the constraint
                if current_tool_count >= tool_limit:
                    analysis['constraint_violations'].append({
                        'id': task_id,
                        'content': task.get('content', ''),
                        'constraint': 'Tool Count Limit (≤30)',
                        'violation': f'Would create new tool when already at limit ({current_tool_count}/{tool_limit})',
                        'priority': priority,
                        'status': status,
                        'recommendation': 'Consider consolidating with existing tools or using resources instead'
                    })

        return analysis

    def _calculate_alignment_score(self, analysis: dict) -> float:
        """Calculate alignment score."""
        if analysis['total_tasks'] == 0:
            return 0.0

        strategy_critical_ratio = len(analysis['strategy_critical']) / max(analysis['by_priority']['high'], 1)
        high_priority_aligned = (analysis['by_priority']['high'] - len(analysis['misaligned_tasks'])) / max(analysis['by_priority']['high'], 1)
        not_stale_ratio = 1.0 - (len(analysis['stale_tasks']) / max(analysis['total_tasks'], 1))
        not_blocked_ratio = 1.0 - (len(analysis['blocked_tasks']) / max(analysis['total_tasks'], 1))

        score = (
            strategy_critical_ratio * 0.4 +
            high_priority_aligned * 0.3 +
            not_stale_ratio * 0.2 +
            not_blocked_ratio * 0.1
        ) * 100

        return round(score, 1)

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights."""
        insights = []

        alignment_score = analysis_results.get('alignment_score', 0)
        insights.append(f"**Alignment Score: {alignment_score}%**")

        if alignment_score < 70:
            insights.append("⚠️ Alignment score is below target (80%+)")

        misaligned = len(analysis_results.get('misaligned_tasks', []))
        if misaligned > 0:
            insights.append(f"⚠️ {misaligned} high-priority tasks are not strategy-aligned")

        blocked = len(analysis_results.get('blocked_tasks', []))
        if blocked > 0:
            insights.append(f"⚠️ {blocked} tasks are blocked by incomplete dependencies")

        # Check Tool Count Limit constraint
        constraint_violations = analysis_results.get('constraint_violations', [])
        current_tool_count = analysis_results.get('current_tool_count', 0)
        tool_limit = analysis_results.get('tool_limit', 30)
        
        if constraint_violations:
            insights.append(f"🚨 {len(constraint_violations)} tasks violate Tool Count Limit constraint (≤{tool_limit})")
            insights.append(f"   Current tool count: {current_tool_count}/{tool_limit}")
            insights.append("   Recommendation: Consolidate tools or use resources instead")
        elif current_tool_count >= tool_limit:
            insights.append(f"⚠️ Tool count at limit ({current_tool_count}/{tool_limit}) - no new tools should be created")
        elif current_tool_count >= tool_limit - 5:
            insights.append(f"⚠️ Tool count approaching limit ({current_tool_count}/{tool_limit})")

        if alignment_score >= 80 and not constraint_violations:
            insights.append("✅ Task alignment is good!")

        return '\n'.join(insights)

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate report."""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        alignment_score = analysis_results.get('alignment_score', 0)

        # Include NetworkX analysis if available
        networkx_section = ""
        if 'networkx_analysis' in self.results:
            nx_analysis = self.results['networkx_analysis']
            networkx_section = f"""
## NetworkX Task Dependency Analysis

**Graph Statistics:**
- Tasks (Nodes): {nx_analysis.get('nodes', 0)}
- Dependencies (Edges): {nx_analysis.get('edges', 0)}
- Is DAG: {nx_analysis.get('is_dag', False)}

**Critical Path:** {len(nx_analysis.get('critical_path', []))} tasks
**Bottlenecks:** {len(nx_analysis.get('bottlenecks', []))} identified
**Orphaned Tasks:** {len(nx_analysis.get('orphans', []))} found

"""

        # Generate phase summary
        phase_summary = self._generate_phase_summary(analysis_results)
        
        # Generate constraint violations section
        constraint_violations = analysis_results.get('constraint_violations', [])
        constraint_section = ""
        if constraint_violations:
            constraint_section = f"""
## Design Constraint Violations

**Tool Count Limit (≤{analysis_results.get('tool_limit', 30)} tools)**

The following tasks would create new tools and violate the design constraint:

"""
            for violation in constraint_violations:
                constraint_section += f"- **{violation['id']}**: {violation['content']}\n"
                constraint_section += f"  - Priority: {violation['priority']}\n"
                constraint_section += f"  - Issue: {violation['violation']}\n"
                constraint_section += f"  - Recommendation: {violation.get('recommendation', 'Review and consolidate')}\n\n"

        return f"""# Todo2 Task Alignment Analysis

*Generated: {timestamp}*
*Goals Source: {self.goals_path.name if self.goals_path.exists() else 'Default phases'}*

## Executive Summary

**Overall Alignment: {alignment_score}%** {'✅' if alignment_score >= 80 else '⚠️'}

**Key Metrics:**
- Total Tasks: {analysis_results.get('total_tasks', 0)}
- High Priority: {analysis_results.get('by_priority', {}).get('high', 0)}
- Goal-Aligned Critical: {len(analysis_results.get('strategy_critical', []))}
- Misaligned: {len(analysis_results.get('misaligned_tasks', []))}
- Infrastructure: {len(analysis_results.get('infrastructure_tasks', []))}
- Blocked: {len(analysis_results.get('blocked_tasks', []))}
- Constraint Violations: {len(analysis_results.get('constraint_violations', []))}
- Current Tool Count: {analysis_results.get('current_tool_count', 0)}/{analysis_results.get('tool_limit', 30)}

---

## Phase Alignment

{phase_summary}

---

## Insights

{insights}

{constraint_section}{networkx_section}
---

*This report was generated using intelligent automation with Tractatus Thinking, Sequential Thinking, and NetworkX analysis.*
"""

    def _generate_phase_summary(self, analysis_results: dict) -> str:
        """Generate phase-by-phase summary."""
        lines = []
        by_phase = analysis_results.get('by_phase', {})

        for phase_key, phase_info in self.strategy_phases.items():
            phase_data = by_phase.get(phase_key, {})
            total = phase_data.get('total', 0)
            high_priority = phase_data.get('high_priority', 0)

            status_icon = '✅' if total > 0 else '⬜'
            lines.append(f"| {status_icon} **{phase_info['name']}** | {total} tasks | {high_priority} high-priority |")

        if lines:
            header = "| Phase | Tasks | High Priority |\n|-------|-------|---------------|\n"
            return header + '\n'.join(lines)
        return "*No phase data available*"

    def _needs_networkx(self) -> bool:
        """NetworkX is useful for task dependency analysis."""
        return True

    def _build_networkx_graph(self):
        """Build task dependency graph."""
        try:
            import networkx as nx

            G = nx.DiGraph()
            tasks = self._load_todo2_tasks()

            # Add tasks as nodes
            for task in tasks:
                G.add_node(
                    task.get('id', 'unknown'),
                    name=task.get('content', ''),
                    priority=task.get('priority', 'medium'),
                    status=task.get('status', 'todo')
                )

            # Add dependencies as edges
            for task in tasks:
                task_id = task.get('id', 'unknown')
                dependencies = task.get('dependencies', [])
                for dep_id in dependencies:
                    if dep_id in G:
                        G.add_edge(dep_id, task_id, relationship='depends_on')

            return G
        except ImportError:
            return None

    def _identify_followup_tasks(self, analysis_results: dict) -> list[dict]:
        """Identify follow-up tasks."""
        followups = []

        # Create tasks for misaligned tasks
        misaligned = analysis_results.get('misaligned_tasks', [])
        if misaligned:
            followups.append({
                'name': 'Review misaligned high-priority tasks',
                'description': f'Review {len(misaligned)} high-priority tasks that are not strategy-aligned',
                'priority': 'high',
                'tags': ['todo2', 'alignment', 'review']
            })

        # Create tasks for blocked tasks
        blocked = analysis_results.get('blocked_tasks', [])
        if blocked:
            followups.append({
                'name': 'Unblock tasks by completing dependencies',
                'description': f'Complete dependencies for {len(blocked)} blocked tasks',
                'priority': 'medium',
                'tags': ['todo2', 'dependencies']
            })

        return followups


def load_config(config_path: Optional[Path] = None) -> dict:
    """Load configuration."""
    if config_path is None:
        from project_management_automation.utils import find_project_root
        project_root = find_project_root()
        config_path = project_root / 'scripts' / 'todo2_alignment_config.json'

    default_config = {
        'output_path': 'docs/TODO2_PRIORITY_ALIGNMENT_ANALYSIS.md'
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
    parser = argparse.ArgumentParser(description='Intelligent Todo2 Alignment Analysis')
    parser.add_argument('--config', type=Path, help='Path to config file')
    parser.add_argument('--output', type=Path, help='Output path for report')
    args = parser.parse_args()

    config = load_config(args.config)
    analyzer = Todo2AlignmentAnalyzerV2(config)

    try:
        results = analyzer.run()

        # Write report
        if args.output:
            output_path = args.output
        else:
            from project_management_automation.utils import find_project_root
            output_path = find_project_root() / config['output_path']

        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, 'w') as f:
            f.write(results['report'])

        logger.info(f"Report written to: {output_path}")
        logger.info(f"Alignment score: {results.get('results', {}).get('alignment_score', 0)}%")

        sys.exit(0)
    except Exception as e:
        logger.error(f"Error running analysis: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()
