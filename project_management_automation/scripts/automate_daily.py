#!/usr/bin/env python3
"""
Daily Automation Orchestrator

Runs routine daily maintenance tasks and generates a combined summary report.
"""

import argparse
import json
import logging
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

# Add project root to path
# Project root will be passed to __init__

# Configure logging (will be configured after project_root is set)
logger = logging.getLogger(__name__)


# Available daily tasks
DAILY_TASKS = {
    'docs_health': {
        'name': 'Documentation Health Check',
        'script': 'project_management_automation/scripts/automate_docs_health_v2.py',
        'mcp_tool': 'check_documentation_health',
        'quick': True,
        'description': 'Monitor documentation quality and structure'
    },
    'todo2_alignment': {
        'name': 'Todo2 Alignment Analysis',
        'script': 'project_management_automation/scripts/automate_todo2_alignment_v2.py',
        'mcp_tool': 'analyze_todo2_alignment',
        'quick': True,
        'description': 'Ensure tasks align with project strategy'
    },
    'duplicate_detection': {
        'name': 'Duplicate Task Detection',
        'script': 'project_management_automation/scripts/automate_todo2_duplicate_detection.py',
        'mcp_tool': 'detect_duplicate_tasks',
        'quick': True,
        'description': 'Detect and report duplicate tasks'
    },
    'dependency_security': {
        'name': 'Dependency Security Scan',
        'script': 'project_management_automation/scripts/automate_dependency_security.py',
        'mcp_tool': 'scan_dependency_security',
        'quick': False,
        'description': 'Check for vulnerable dependencies'
    },
    # NOTE: external_tool_hints removed from daily automation - it's a one-time setup tool
    # Run manually when new documentation files are added: add_external_tool_hints(dry_run=False)
    # NOTE: tool_count_health removed - module never existed; use exarp-go automation(daily) for native Go check (health action=tools)
    'handoff_check': {
        'name': 'Handoff Check',
        'script': None,  # Uses direct function call
        'mcp_tool': 'session_handoff',
        'quick': True,
        'description': 'Check for handoff notes from other developers/machines',
        'function': 'project_management_automation.tools.session_handoff:get_latest_handoff'
    },
    'task_progress_inference': {
        'name': 'Task Progress Inference',
        'script': None,  # Uses direct function call
        'mcp_tool': 'auto_update_task_status',
        'quick': True,
        'description': 'Infer task completion from codebase analysis',
        'function': 'project_management_automation.tools.auto_update_task_status:auto_update_task_status'
    },
    'stale_task_cleanup': {
        'name': 'Stale Task Cleanup',
        'script': 'project_management_automation/scripts/automate_stale_task_cleanup.py',
        'mcp_tool': 'cleanup_stale_tasks',
        'quick': True,
        'description': 'Move stale In Progress tasks back to Todo for accurate time tracking'
    },
    'duplicate_test_names': {
        'name': 'Duplicate Test Name Check',
        'script': 'project_management_automation/scripts/automate_check_duplicate_test_names.py',
        'mcp_tool': None,  # Not exposed as MCP tool yet
        'quick': True,
        'description': 'Check for duplicate test function names across test files'
    }
}


class DailyAutomation:
    """Orchestrates daily maintenance tasks."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        self.config = config
        self.project_root = project_root
        self.tasks_to_run = config.get('tasks', ['handoff_check', 'stale_task_cleanup', 'docs_health', 'todo2_alignment', 'duplicate_detection', 'duplicate_test_names'])
        self.dry_run = config.get('dry_run', False)
        self.output_path = config.get('output_path', 'docs/DAILY_AUTOMATION_REPORT.md')
        self.include_slow = config.get('include_slow', False)

        # Results
        self.results = {
            'timestamp': datetime.now().isoformat(),
            'tasks_run': [],
            'tasks_succeeded': [],
            'tasks_failed': [],
            'summary': {},
            'duration_seconds': 0
        }

    def run(self) -> dict:
        """Run all selected daily tasks."""
        start_time = time.time()
        logger.info(f"Starting daily automation (dry_run={self.dry_run})")

        # Filter tasks based on configuration
        tasks = self._filter_tasks()

        if not tasks:
            logger.warning("No tasks selected to run")
            return {
                'status': 'error',
                'error': 'No tasks selected to run',
                'results': self.results
            }

        logger.info(f"Running {len(tasks)} daily tasks: {', '.join(tasks)}")

        # Run each task
        for task_id in tasks:
            task_info = DAILY_TASKS[task_id]
            logger.info(f"Running: {task_info['name']}")

            task_result = self._run_task(task_id, task_info)
            self.results['tasks_run'].append({
                'task_id': task_id,
                'task_name': task_info['name'],
                'status': task_result['status'],
                'duration_seconds': task_result.get('duration', 0),
                'error': task_result.get('error'),
                'summary': task_result.get('summary', {})
            })

            if task_result['status'] == 'success':
                self.results['tasks_succeeded'].append(task_id)
            else:
                self.results['tasks_failed'].append(task_id)

        # Get task recommendations (if agentic-tools available)
        recommendations = self._get_task_recommendations()
        if recommendations:
            self.results['task_recommendations'] = recommendations
        
        # Get progress inference results (T-13)
        progress_inference = self._get_progress_inference()
        if progress_inference:
            self.results['progress_inference'] = progress_inference

        # Generate summary
        self.results['summary'] = self._generate_summary()
        self.results['duration_seconds'] = time.time() - start_time

        # Generate report
        report = self._generate_report()

        # Save report
        if self.output_path:
            report_path = Path(self.output_path)
            report_path.parent.mkdir(parents=True, exist_ok=True)
            with open(report_path, 'w') as f:
                f.write(report)
            logger.info(f"Report saved to: {report_path}")

        # Generate Agile daily standup MD file
        standup_path = self._generate_daily_standup()
        if standup_path:
            self.results['standup_path'] = str(standup_path)
            logger.info(f"Daily standup saved to: {standup_path}")

        logger.info(f"Daily automation completed in {self.results['duration_seconds']:.2f}s")
        return {
            'status': 'success',
            'results': self.results,
            'report_path': str(self.output_path),
            'standup_path': str(standup_path) if standup_path else None
        }

    def _filter_tasks(self) -> list[str]:
        """Filter tasks based on configuration."""
        tasks = []

        for task_id in self.tasks_to_run:
            if task_id not in DAILY_TASKS:
                logger.warning(f"Unknown task: {task_id}")
                continue

            task_info = DAILY_TASKS[task_id]

            # Skip slow tasks unless explicitly included
            if not task_info['quick'] and not self.include_slow:
                logger.info(f"Skipping slow task: {task_id} (use --include-slow to include)")
                continue

            tasks.append(task_id)

        return tasks

    def _run_task(self, task_id: str, task_info: dict) -> dict[str, Any]:
        """Run a single task."""
        start_time = time.time()

        try:
            # Try to run via script first
            script_path = self.project_root / task_info['script']
            if script_path.exists():
                result = self._run_script(script_path, task_id)
            else:
                # Fallback: Try MCP tool (would need MCP client)
                logger.warning(f"Script not found: {script_path}, skipping")
                result = {'status': 'skipped', 'error': 'Script not found'}

            duration = float(time.time() - start_time)
            result_with_duration: dict[str, Any] = {**result, 'duration': duration}
            return result_with_duration

        except Exception as e:
            logger.error(f"Error running task {task_id}: {e}", exc_info=True)
            return {
                'status': 'error',
                'error': str(e),
                'duration': float(time.time() - start_time)
            }

    def _run_script(self, script_path: Path, task_id: str) -> dict[str, Any]:
        """Run a Python script as a module."""
        import subprocess

        try:
            # Convert script path to module name
            # e.g., project_management_automation/scripts/automate_docs_health_v2.py
            #    -> project_management_automation.scripts.automate_docs_health_v2
            # Use the relative path from DAILY_TASKS, not the full filesystem path
            relative_script = DAILY_TASKS[task_id]['script']
            module_name = relative_script.replace('/', '.').replace('.py', '')

            # Build command - run as module
            cmd = [sys.executable, '-m', module_name]

            # Note: Most scripts don't support --dry-run flag
            # Only add task-specific arguments that are actually supported
            # (duplicate_detection's --auto-fix is a boolean flag, don't pass 'false')

            # Run script
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=300  # 5 minute timeout
            )

            if result.returncode == 0:
                # Try to parse JSON output
                try:
                    output_data = json.loads(result.stdout)
                    return {
                        'status': 'success',
                        'summary': output_data.get('data', {})
                    }
                except json.JSONDecodeError:
                    return {
                        'status': 'success',
                        'summary': {'output': str(result.stdout[:500])}  # First 500 chars
                    }
            else:
                return {
                    'status': 'error',
                    'error': result.stderr[:200] if result.stderr else 'Script failed'
                }

        except subprocess.TimeoutExpired:
            return {
                'status': 'error',
                'error': 'Task timed out after 5 minutes'
            }
        except Exception as e:
            return {
                'status': 'error',
                'error': str(e)
            }

    def _get_task_recommendations(self) -> Optional[dict[str, Any]]:
        """Get intelligent task recommendations using agentic-tools."""
        try:
            from project_management_automation.utils.agentic_tools_client import get_next_task_recommendation_mcp
            
            recommendations = get_next_task_recommendation_mcp(
                project_root=self.project_root,
                max_recommendations=3,
                consider_complexity=True,
                exclude_blocked=True
            )
            
            if recommendations:
                logger.info(f"Received {len(recommendations.get('recommendations', []))} task recommendations")
                return recommendations
        except Exception as e:
            logger.debug(f"Task recommendations not available: {e}")
        return None

    def _get_progress_inference(self) -> Optional[dict[str, Any]]:
        """Get task progress inference using auto_update_task_status (T-13)."""
        try:
            from project_management_automation.tools.auto_update_task_status import auto_update_task_status
            import json
            
            # Call with dry_run=True for safety (only report, don't update)
            result_json = auto_update_task_status(
                confidence_threshold=0.7,
                auto_update=False,  # Don't auto-update in daily automation
                output_path=None,  # Don't create separate report
                codebase_path=str(self.project_root)
            )
            
            if result_json:
                result = json.loads(result_json)
                if result.get('success') and result.get('data'):
                    data = result['data']
                    logger.info(f"Progress inference: {data.get('inferences_made', 0)} inferences made")
                    return data
        except Exception as e:
            logger.debug(f"Progress inference not available: {e}")
        return None

    def _generate_summary(self) -> dict:
        """Generate summary statistics."""
        total = len(self.results['tasks_run'])
        succeeded = len(self.results['tasks_succeeded'])
        failed = len(self.results['tasks_failed'])

        return {
            'total_tasks': total,
            'succeeded': succeeded,
            'failed': failed,
            'success_rate': (succeeded / total * 100) if total > 0 else 0,
            'duration_seconds': self.results['duration_seconds']
        }

    def _generate_daily_standup(self) -> Optional[Path]:
        """Generate Agile daily standup markdown file."""
        try:
            # Determine output path for standup file
            if self.output_path:
                output_dir = Path(self.output_path).parent
                date_str = datetime.now().strftime('%Y-%m-%d')
                standup_path = output_dir / f'daily-standup-{date_str}.md'
            else:
                # Default location if no output_path specified
                if self.project_root:
                    output_dir = self.project_root / 'docs'
                else:
                    from project_management_automation.utils import find_project_root
                    output_dir = find_project_root() / 'docs'
                output_dir.mkdir(parents=True, exist_ok=True)
                date_str = datetime.now().strftime('%Y-%m-%d')
                standup_path = output_dir / f'daily-standup-{date_str}.md'

            # Get completed tasks from yesterday (for "What I did yesterday")
            yesterday_tasks = self._get_recent_completed_tasks(days=1)
            
            # Get tasks for today (for "What I'm doing today")
            today_tasks = self._get_tasks_for_today()
            
            # Get blockers/impediments
            blockers = self._get_blockers()

            # Generate standup content
            standup_content = self._build_standup_template(
                date_str=date_str,
                yesterday_tasks=yesterday_tasks,
                today_tasks=today_tasks,
                blockers=blockers,
                summary=self.results.get('summary', {})
            )

            # Write standup file
            standup_path.parent.mkdir(parents=True, exist_ok=True)
            with open(standup_path, 'w') as f:
                f.write(standup_content)

            return standup_path

        except Exception as e:
            logger.error(f"Failed to generate daily standup: {e}", exc_info=True)
            return None

    def _get_recent_completed_tasks(self, days: int = 1) -> list[dict]:
        """Get recently completed tasks from Todo2."""
        try:
            from project_management_automation.utils.todo2_utils import (
                filter_tasks_by_project,
                get_repo_project_id
            )
            from project_management_automation.utils import find_project_root
            from datetime import datetime, timedelta
            
            project_root = self.project_root or find_project_root()
            todo2_file = project_root / '.todo2' / 'state.todo2.json'
            
            if not todo2_file.exists():
                return []
            
            with open(todo2_file) as f:
                data = json.load(f)
            
            todos = data.get('todos', [])
            project_id = get_repo_project_id(project_root)
            
            # Filter to project tasks
            project_todos = filter_tasks_by_project(todos, project_id, include_unassigned=True)
            
            # Get completed tasks (status = "Done")
            completed_tasks = [
                t for t in project_todos
                if t.get('status', '').lower() in ['done', 'completed']
            ]
            
            # Filter by completion date if available
            cutoff_date = datetime.now() - timedelta(days=days)
            recent_completed = []
            for t in completed_tasks:
                completed_at_str = t.get('completed_at') or t.get('updated_at')
                if completed_at_str:
                    try:
                        # Parse ISO format datetime
                        completed_at = datetime.fromisoformat(completed_at_str.replace('Z', '+00:00'))
                        if completed_at >= cutoff_date:
                            recent_completed.append(t)
                    except (ValueError, AttributeError):
                        # If date parsing fails, include the task anyway
                        recent_completed.append(t)
                else:
                    # If no date, include it (might be recently completed)
                    recent_completed.append(t)
            
            # Sort by completion date (most recent first)
            recent_completed.sort(
                key=lambda t: t.get('completed_at') or t.get('updated_at') or '',
                reverse=True
            )
            
            return [
                {
                    'id': t.get('id', 'N/A'),
                    'name': t.get('name', 'Unnamed'),
                    'completed_at': t.get('completed_at') or t.get('updated_at', 'N/A')
                }
                for t in recent_completed[:10]  # Limit to 10 most recent
            ]
        except Exception as e:
            logger.debug(f"Failed to get recent completed tasks: {e}")
            return []

    def _get_tasks_for_today(self) -> list[dict]:
        """Get tasks planned for today (In Progress or high priority Todo)."""
        try:
            from project_management_automation.utils.todo2_utils import (
                filter_tasks_by_project,
                get_repo_project_id
            )
            from project_management_automation.utils import find_project_root
            
            project_root = self.project_root or find_project_root()
            todo2_file = project_root / '.todo2' / 'state.todo2.json'
            
            if not todo2_file.exists():
                return []
            
            with open(todo2_file) as f:
                data = json.load(f)
            
            todos = data.get('todos', [])
            project_id = get_repo_project_id(project_root)
            
            # Filter to project tasks
            project_todos = filter_tasks_by_project(todos, project_id, include_unassigned=True)
            
            # Get In Progress tasks or high priority Todo tasks
            today_tasks = []
            for t in project_todos:
                status = t.get('status', '').lower()
                priority = t.get('priority', 'medium').lower()
                
                if status in ['in progress', 'in_progress', 'working']:
                    today_tasks.append({
                        'id': t.get('id', 'N/A'),
                        'name': t.get('name', 'Unnamed'),
                        'status': t.get('status', 'In Progress'),
                        'priority': t.get('priority', 'medium')
                    })
                elif status in ['todo', 'pending'] and priority in ['high', 'critical']:
                    today_tasks.append({
                        'id': t.get('id', 'N/A'),
                        'name': t.get('name', 'Unnamed'),
                        'status': t.get('status', 'Todo'),
                        'priority': t.get('priority', 'medium')
                    })
            
            # Sort by priority (high first), then by status
            priority_order = {'critical': 0, 'high': 1, 'medium': 2, 'low': 3}
            today_tasks.sort(key=lambda x: (
                priority_order.get(x.get('priority', 'medium').lower(), 2),
                x.get('status', '')
            ))
            
            return today_tasks[:10]  # Limit to 10 tasks
        except Exception as e:
            logger.debug(f"Failed to get tasks for today: {e}")
            return []

    def _get_blockers(self) -> list[dict]:
        """Get blockers/impediments (blocked tasks or failed automation tasks)."""
        blockers = []
        
        # Check for blocked tasks in Todo2
        try:
            from project_management_automation.utils.todo2_utils import (
                filter_tasks_by_project,
                get_repo_project_id
            )
            from project_management_automation.utils import find_project_root
            
            project_root = self.project_root or find_project_root()
            todo2_file = project_root / '.todo2' / 'state.todo2.json'
            
            if todo2_file.exists():
                with open(todo2_file) as f:
                    data = json.load(f)
                
                todos = data.get('todos', [])
                project_id = get_repo_project_id(project_root)
                
                # Filter to project tasks
                project_todos = filter_tasks_by_project(todos, project_id, include_unassigned=True)
                
                blocked_tasks = [
                    t for t in project_todos
                    if t.get('status', '').lower() in ['blocked', 'block']
                ]
                
                for t in blocked_tasks:
                    blockers.append({
                        'type': 'blocked_task',
                        'id': t.get('id', 'N/A'),
                        'name': t.get('name', 'Unnamed'),
                        'reason': t.get('blocker_reason') or t.get('notes', 'No reason provided')
                    })
        except Exception as e:
            logger.debug(f"Failed to get blocked tasks: {e}")
        
        # Add failed automation tasks as blockers
        for task_id in self.results.get('tasks_failed', []):
            task_info = DAILY_TASKS.get(task_id, {})
            blockers.append({
                'type': 'automation_failure',
                'task': task_info.get('name', task_id),
                'error': 'See daily automation report for details'
            })
        
        return blockers

    def _build_standup_template(
        self,
        date_str: str,
        yesterday_tasks: list[dict],
        today_tasks: list[dict],
        blockers: list[dict],
        summary: dict
    ) -> str:
        """Build Agile daily standup markdown template."""
        lines = [
            f"# Daily Standup - {date_str}",
            "",
            "Generated by exarp-go automation tool",
            "",
            "---",
            "",
            "## 🎯 What I Did Yesterday",
            "",
        ]
        
        if yesterday_tasks:
            lines.append("**Completed Tasks:**")
            lines.append("")
            for task in yesterday_tasks:
                task_name = task.get('name', 'Unnamed')
                task_id = task.get('id', 'N/A')
                lines.append(f"- ✅ {task_name} (`{task_id}`)")
            lines.append("")
        else:
            lines.append("*No tasks completed yesterday*")
            lines.append("")
        
        lines.extend([
            "---",
            "",
            "## 🚀 What I'm Doing Today",
            "",
        ])
        
        if today_tasks:
            lines.append("**Planned Tasks:**")
            lines.append("")
            for task in today_tasks:
                task_name = task.get('name', 'Unnamed')
                task_id = task.get('id', 'N/A')
                status = task.get('status', 'Todo')
                priority = task.get('priority', 'medium')
                priority_emoji = {
                    'critical': '🔴',
                    'high': '🟠',
                    'medium': '🟡',
                    'low': '🟢'
                }.get(priority.lower(), '⚪')
                
                lines.append(f"- {priority_emoji} **{task_name}** (`{task_id}`) - {status} (Priority: {priority})")
            lines.append("")
        else:
            lines.append("*No tasks planned for today*")
            lines.append("")
        
        # Add task recommendations if available
        if self.results.get('task_recommendations'):
            recommendations = self.results['task_recommendations']
            rec_list = recommendations.get('recommendations', [])
            if rec_list:
                lines.extend([
                    "",
                    "**Recommended Tasks:**",
                    ""
                ])
                for rec in rec_list[:3]:  # Show top 3 recommendations
                    task_name = rec.get('taskName', 'Unknown')
                    task_id = rec.get('taskId', 'N/A')
                    reason = rec.get('reason', 'No reason provided')
                    lines.append(f"- 💡 **{task_name}** (`{task_id}`)")
                    lines.append(f"  - *Reason*: {reason}")
                lines.append("")
        
        lines.extend([
            "---",
            "",
            "## 🚧 Blockers / Impediments",
            "",
        ])
        
        if blockers:
            for blocker in blockers:
                blocker_type = blocker.get('type', 'unknown')
                if blocker_type == 'blocked_task':
                    lines.append(f"- 🚫 **{blocker.get('name', 'Unnamed')}** (`{blocker.get('id', 'N/A')}`)")
                    lines.append(f"  - *Reason*: {blocker.get('reason', 'No reason provided')}")
                elif blocker_type == 'automation_failure':
                    lines.append(f"- ⚠️ **{blocker.get('task', 'Unknown')}**")
                    lines.append(f"  - *Issue*: {blocker.get('error', 'Automation task failed')}")
            lines.append("")
        else:
            lines.append("*No blockers identified*")
            lines.append("")
        
        # Add automation summary
        lines.extend([
            "---",
            "",
            "## 📊 Automation Summary",
            "",
            f"- **Tasks Run**: {summary.get('total_tasks', 0)}",
            f"- **Succeeded**: {summary.get('succeeded', 0)}",
            f"- **Failed**: {summary.get('failed', 0)}",
            f"- **Success Rate**: {summary.get('success_rate', 0):.1f}%",
            f"- **Duration**: {summary.get('duration_seconds', 0):.2f}s",
            "",
            "---",
            "",
            "*This standup file was automatically generated by exarp-go. You can edit it to add additional notes or context.*",
            ""
        ])
        
        return '\n'.join(lines)

    def _generate_report(self) -> str:
        """Generate markdown report."""
        summary = self.results['summary']

        report_lines = [
            "# Daily Automation Report",
            "",
            f"**Date**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
            f"**Mode**: {'DRY RUN' if self.dry_run else 'APPLIED'}",
            "",
            "## Summary",
            "",
            f"- **Total Tasks**: {summary['total_tasks']}",
            f"- **Succeeded**: {summary['succeeded']}",
            f"- **Failed**: {summary['failed']}",
            f"- **Success Rate**: {summary['success_rate']:.1f}%",
            f"- **Duration**: {summary['duration_seconds']:.2f} seconds",
            "",
            "## Task Results",
            "",
        ]

        for task_result in self.results['tasks_run']:
            status_icon = "✅" if task_result['status'] == 'success' else "❌"
            report_lines.append(f"### {status_icon} {task_result['task_name']}")
            report_lines.append("")
            report_lines.append(f"- **Status**: {task_result['status']}")
            report_lines.append(f"- **Duration**: {task_result['duration_seconds']:.2f}s")

            if task_result.get('error'):
                report_lines.append(f"- **Error**: {task_result['error']}")

            if task_result.get('summary'):
                report_lines.append("- **Summary**:")
                for key, value in task_result['summary'].items():
                    if isinstance(value, (int, float, str)):
                        report_lines.append(f"  - {key}: {value}")

            report_lines.append("")

        # Add task recommendations if available
        if self.results.get('task_recommendations'):
            recommendations = self.results['task_recommendations']
            rec_list = recommendations.get('recommendations', [])
            if rec_list:
                report_lines.extend([
                    "",
                    "## Task Recommendations",
                    "",
                    "Intelligent task recommendations based on dependencies, priorities, and complexity:",
                    ""
                ])
                for i, rec in enumerate(rec_list[:3], 1):  # Show top 3
                    task_id = rec.get('taskId', 'N/A')
                    task_name = rec.get('taskName', 'Unknown')
                    reason = rec.get('reason', 'No reason provided')
                    report_lines.extend([
                        f"### {i}. {task_name} ({task_id})",
                        "",
                        f"**Reason**: {reason}",
                        ""
                    ])
        
        # Add progress inference if available (T-13)
        if self.results.get('progress_inference'):
            inference = self.results['progress_inference']
            inferences_made = inference.get('inferences_made', 0)
            tasks_analyzed = inference.get('total_tasks_analyzed', 0)
            if inferences_made > 0:
                report_lines.extend([
                    "",
                    "## Progress Inference",
                    "",
                    f"- **Tasks Analyzed**: {tasks_analyzed}",
                    f"- **Inferences Made**: {inferences_made}",
                    "",
                    "Tasks with inferred status changes:",
                    ""
                ])
                for result in inference.get('inferred_results', [])[:5]:  # Show top 5
                    task_name = result.get('task_name', 'N/A')
                    current = result.get('current_status', 'N/A')
                    inferred = result.get('inferred_status', 'N/A')
                    confidence = result.get('confidence', 0.0)
                    report_lines.append(f"- **{task_name}**: {current} → {inferred} (Confidence: {confidence:.0%})")
                report_lines.append("")

        return '\n'.join(report_lines)


def main():
    """Main entry point."""
    from project_management_automation.utils import find_project_root

    parser = argparse.ArgumentParser(description='Run daily maintenance tasks')
    parser.add_argument('--tasks', nargs='+',
                       choices=list(DAILY_TASKS.keys()),
                       default=['docs_health', 'todo2_alignment', 'duplicate_detection'],
                       help='Tasks to run (default: quick tasks only)')
    parser.add_argument('--include-slow', action='store_true',
                       help='Include slow tasks (e.g., dependency security scan)')
    parser.add_argument('--dry-run', action='store_true',
                       help='Preview changes without applying')
    parser.add_argument('--output-path', type=str,
                       default='docs/DAILY_AUTOMATION_REPORT.md',
                       help='Path for report output')

    args = parser.parse_args()

    project_root = find_project_root()

    config = {
        'tasks': args.tasks,
        'include_slow': args.include_slow,
        'dry_run': args.dry_run,
        'output_path': args.output_path
    }

    automation = DailyAutomation(config, project_root=project_root)
    results = automation.run()

    print(json.dumps(results, indent=2))
    return 0 if results['status'] == 'success' else 1


if __name__ == '__main__':
    sys.exit(main())
