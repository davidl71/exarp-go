#!/usr/bin/env python3
"""
Shared TODO Table Synchronization Automation

Syncs between agents/shared/TODO_OVERVIEW.md and Todo2 (.todo2/state.todo2.json)
- Bidirectional sync of task status
- Creates new Todo2 tasks for new shared TODO items
- Handles conflicts intelligently
- Tracks sync history

Usage:
    python3 scripts/automate_todo_sync.py [--config config.json] [--dry-run]
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

# Configure logging (will be configured after project_root is set)
logger = logging.getLogger(__name__)


class TodoSyncAutomation(IntelligentAutomationBase):
    """Intelligent TODO synchronization using base class."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Shared TODO Table Synchronization", project_root)
        self.shared_todo_path = self.project_root / 'agents' / 'shared' / 'TODO_OVERVIEW.md'
        self.todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
        self.sync_history_path = self.project_root / 'scripts' / '.todo_sync_history.json'
        self.dry_run = config.get('dry_run', False)

        # Status mapping
        self.status_map = {
            'pending': 'Todo',
            'in_progress': 'In Progress',
            'completed': 'Done',
            'Todo': 'pending',
            'In Progress': 'in_progress',
            'Done': 'completed',
            'Review': 'in_progress',  # Review maps to in_progress in shared TODO
            'Cancelled': 'completed'  # Cancelled maps to completed
        }

        # Sync results
        self.sync_results = {
            'shared_todos': [],
            'todo2_tasks': [],
            'matches': [],
            'conflicts': [],
            'new_shared_todos': [],
            'new_todo2_tasks': [],
            'updates': []
        }

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is TODO synchronization?"""
        return "What is TODO synchronization? TODO Sync = Shared TODO Table × Todo2 Tasks × Status Mapping × Conflict Resolution × Bidirectional Updates"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we synchronize TODO systems?"""
        return "How do we systematically synchronize shared TODO table with Todo2 tasks bidirectionally?"

    def _execute_analysis(self) -> dict:
        """Execute TODO synchronization."""
        logger.info("Executing TODO synchronization...")

        # Load both systems
        shared_todos = self._load_shared_todos()
        todo2_tasks = self._load_todo2_tasks()

        logger.info(f"Loaded {len(shared_todos)} shared TODOs and {len(todo2_tasks)} Todo2 tasks")

        # Find matches
        matches = self._find_matches(shared_todos, todo2_tasks)
        logger.info(f"Found {len(matches)} matches")

        # Detect conflicts
        conflicts = self._detect_conflicts(matches)
        logger.info(f"Found {len(conflicts)} conflicts")

        # Find new items
        new_shared = self._find_new_shared_todos(shared_todos, matches)
        new_todo2 = self._find_new_todo2_tasks(todo2_tasks, matches)
        logger.info(f"Found {len(new_shared)} new shared TODOs and {len(new_todo2)} new Todo2 tasks")

        # Perform sync (if not dry run)
        if not self.dry_run:
            updates = self._perform_sync(matches, conflicts, new_shared, new_todo2)
            logger.info(f"Performed {len(updates)} updates")
        else:
            updates = self._simulate_sync(matches, conflicts, new_shared, new_todo2)
            logger.info(f"Simulated {len(updates)} updates (dry run)")

        return {
            'shared_todos': shared_todos,
            'todo2_tasks': todo2_tasks,
            'matches': matches,
            'conflicts': conflicts,
            'new_shared_todos': new_shared,
            'new_todo2_tasks': new_todo2,
            'updates': updates,
            'dry_run': self.dry_run
        }

    def _load_shared_todos(self) -> list[dict]:
        """Load shared TODO table from markdown."""
        if not self.shared_todo_path.exists():
            logger.warning(f"Shared TODO file not found: {self.shared_todo_path}")
            return []

        try:
            content = self.shared_todo_path.read_text(encoding='utf-8')

            # Parse markdown table
            todos = []
            table_pattern = re.compile(r'^\| (\d+) \| (.+?) \| (.+?) \| (.+?) \|', re.MULTILINE)

            for match in table_pattern.finditer(content):
                todo_id = match.group(1)
                description = match.group(2).strip()
                owner = match.group(3).strip()
                status = match.group(4).strip().lower()

                todos.append({
                    'id': todo_id,
                    'description': description,
                    'owner': owner,
                    'status': status,
                    'source': 'shared'
                })

            return todos
        except Exception as e:
            logger.error(f"Error loading shared TODOs: {e}")
            return []

    def _load_todo2_tasks(self) -> list[dict]:
        """Load Todo2 tasks from MCP (preferred) or JSON file (fallback)."""
        from project_management_automation.utils.todo2_mcp_client import list_todos_mcp
        
        # Try Todo2 MCP first (preferred)
        try:
            mcp_tasks = list_todos_mcp(project_root=self.project_root)
            if mcp_tasks:
                # Add source marker
                for task in mcp_tasks:
                    task['source'] = 'todo2'
                logger.info(f"Loaded {len(mcp_tasks)} tasks from Todo2 MCP")
                return mcp_tasks
        except Exception as e:
            logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")
        
        # Fallback to direct file access
        if not self.todo2_path.exists():
            logger.warning(f"Todo2 file not found: {self.todo2_path}")
            return []

        try:
            with open(self.todo2_path) as f:
                data = json.load(f)
                tasks = data.get('todos', [])

                # Add source marker
                for task in tasks:
                    task['source'] = 'todo2'

                return tasks
        except FileNotFoundError:
            logger.info("No Todo2 state file found - no tasks to sync")
            return []
        except Exception as e:
            logger.warning(f"Could not load Todo2 tasks: {e}")
            return []

    def _find_matches(self, shared_todos: list[dict], todo2_tasks: list[dict]) -> list[dict]:
        """Find matching tasks between systems."""
        matches = []

        for shared in shared_todos:
            for todo2 in todo2_tasks:
                # Match by description similarity or explicit ID reference
                similarity = self._calculate_similarity(
                    shared['description'],
                    todo2.get('name', '') or todo2.get('content', '')
                )

                # Check if Todo2 task references shared TODO ID
                todo2_desc = str(todo2.get('long_description', '')) + ' ' + str(todo2.get('name', ''))
                has_id_ref = f"TODO {shared['id']}" in todo2_desc or f"#{shared['id']}" in todo2_desc

                if similarity > 0.7 or has_id_ref:
                    matches.append({
                        'shared': shared,
                        'todo2': todo2,
                        'similarity': similarity,
                        'has_id_ref': has_id_ref
                    })

        return matches

    def _calculate_similarity(self, str1: str, str2: str) -> float:
        """Calculate string similarity (simple word overlap)."""
        words1 = set(str1.lower().split())
        words2 = set(str2.lower().split())

        if not words1 or not words2:
            return 0.0

        intersection = words1 & words2
        union = words1 | words2

        return len(intersection) / len(union) if union else 0.0

    def _detect_conflicts(self, matches: list[dict]) -> list[dict]:
        """Detect status conflicts between matched tasks."""
        conflicts = []

        for match in matches:
            shared_status = match['shared']['status']
            todo2_status = match['todo2'].get('status', 'Todo')

            # Map to common status
            shared_mapped = self.status_map.get(shared_status, shared_status)
            todo2_mapped = self.status_map.get(todo2_status, todo2_status)

            # Check if statuses are different
            if shared_mapped != todo2_mapped:
                conflicts.append({
                    'match': match,
                    'shared_status': shared_status,
                    'todo2_status': todo2_status,
                    'shared_mapped': shared_mapped,
                    'todo2_mapped': todo2_mapped
                })

        return conflicts

    def _find_new_shared_todos(self, shared_todos: list[dict], matches: list[dict]) -> list[dict]:
        """Find shared TODOs without Todo2 matches."""
        matched_shared_ids = {m['shared']['id'] for m in matches}
        return [t for t in shared_todos if t['id'] not in matched_shared_ids]

    def _find_new_todo2_tasks(self, todo2_tasks: list[dict], matches: list[dict]) -> list[dict]:
        """Find Todo2 tasks without shared TODO matches."""
        matched_todo2_ids = {m['todo2']['id'] for m in matches}
        return [t for t in todo2_tasks if t['id'] not in matched_todo2_ids]

    def _perform_sync(self, matches: list[dict], conflicts: list[dict],
                     new_shared: list[dict], new_todo2: list[dict]) -> list[dict]:
        """Perform actual synchronization."""
        updates = []

        # Resolve conflicts (prefer Todo2 as source of truth for now)
        for conflict in conflicts:
            match = conflict['match']
            shared = match['shared']
            todo2 = match['todo2']

            # Update shared TODO to match Todo2
            new_status = self.status_map.get(todo2.get('status', 'Todo'), 'pending')
            if new_status != shared['status']:
                updates.append({
                    'type': 'update_shared',
                    'todo_id': shared['id'],
                    'old_status': shared['status'],
                    'new_status': new_status,
                    'reason': 'Sync from Todo2'
                })
                self._update_shared_todo_status(shared['id'], new_status)

        # Create Todo2 tasks for new shared TODOs
        for shared in new_shared:
            todo2_task = self._create_todo2_from_shared(shared)
            updates.append({
                'type': 'create_todo2',
                'shared_id': shared['id'],
                'todo2_id': todo2_task.get('id'),
                'description': shared['description']
            })

        # Update Todo2 tasks for new matches (sync status from shared)
        for match in matches:
            if match not in [c['match'] for c in conflicts]:
                # No conflict, but ensure status is synced
                shared_status = self.status_map.get(match['shared']['status'], 'Todo')
                todo2_status = match['todo2'].get('status', 'Todo')

                if shared_status != todo2_status:
                    updates.append({
                        'type': 'update_todo2',
                        'todo2_id': match['todo2']['id'],
                        'old_status': todo2_status,
                        'new_status': shared_status,
                        'reason': 'Sync from shared TODO'
                    })
                    self._update_todo2_status(match['todo2']['id'], shared_status)

        return updates

    def _simulate_sync(self, matches: list[dict], conflicts: list[dict],
                      new_shared: list[dict], new_todo2: list[dict]) -> list[dict]:
        """Simulate synchronization (dry run)."""
        updates = []

        for conflict in conflicts:
            match = conflict['match']
            shared = match['shared']
            todo2 = match['todo2']

            new_status = self.status_map.get(todo2.get('status', 'Todo'), 'pending')
            updates.append({
                'type': 'update_shared (simulated)',
                'todo_id': shared['id'],
                'old_status': shared['status'],
                'new_status': new_status
            })

        for shared in new_shared:
            updates.append({
                'type': 'create_todo2 (simulated)',
                'shared_id': shared['id'],
                'description': shared['description']
            })

        return updates

    def _update_shared_todo_status(self, todo_id: str, new_status: str) -> bool:
        """Update status in shared TODO markdown file."""
        try:
            content = self.shared_todo_path.read_text(encoding='utf-8')

            # Replace status for this TODO ID
            pattern = rf'^\| {re.escape(todo_id)} \| (.+?) \| (.+?) \| \w+ \|'
            replacement = rf'| {todo_id} | \1 | \2 | {new_status} |'

            new_content = re.sub(pattern, replacement, content, flags=re.MULTILINE)

            if new_content != content:
                self.shared_todo_path.write_text(new_content, encoding='utf-8')
                return True
        except Exception as e:
            logger.error(f"Error updating shared TODO {todo_id}: {e}")

        return False

    def _update_todo2_status(self, todo2_id: str, new_status: str) -> bool:
        """Update status in Todo2 via MCP (preferred) or JSON file (fallback)."""
        from project_management_automation.utils.todo2_mcp_client import update_todos_mcp
        
        # Try Todo2 MCP first (preferred)
        try:
            success = update_todos_mcp(
                updates=[{
                    'id': todo2_id,
                    'status': new_status
                }],
                project_root=self.project_root
            )
            if success:
                logger.debug(f"Updated Todo2 task {todo2_id} status via MCP")
                return True
        except Exception as e:
            logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")
        
        # Fallback to direct file access
        try:
            with open(self.todo2_path) as f:
                data = json.load(f)

            for task in data.get('todos', []):
                if task.get('id') == todo2_id:
                    task['status'] = new_status
                    task['lastModified'] = datetime.now(timezone.utc).isoformat()
                    break

            with open(self.todo2_path, 'w') as f:
                json.dump(data, f, indent=2)

            return True
        except Exception as e:
            logger.error(f"Error updating Todo2 task {todo2_id}: {e}")

        return False

    def _create_todo2_from_shared(self, shared: dict) -> dict:
        """Create Todo2 task from shared TODO via MCP (preferred) or JSON file (fallback)."""
        from project_management_automation.utils.todo2_mcp_client import create_todos_mcp, list_todos_mcp
        
        # Prepare task data
        task_name = shared['description']
        task_description = f"Synced from shared TODO {shared['id']}\n\nOwner: {shared['owner']}"
        task_status = self.status_map.get(shared['status'], 'Todo')
        task_tags = ['shared-todo', 'synced', shared.get('owner', 'unknown')]
        
        # Try Todo2 MCP first (preferred)
        try:
            created_ids = create_todos_mcp(
                todos=[{
                    'name': task_name,
                    'long_description': task_description,
                    'status': task_status,
                    'priority': 'medium',
                    'tags': task_tags
                }],
                project_root=self.project_root
            )
            if created_ids and len(created_ids) > 0:
                todo2_id = created_ids[0]
                # Fetch the created task to return full data
                from project_management_automation.utils.todo2_mcp_client import get_todo_details_mcp
                created_tasks = get_todo_details_mcp([todo2_id], project_root=self.project_root)
                if created_tasks:
                    logger.debug(f"Created Todo2 task {todo2_id} via MCP")
                    return created_tasks[0]
        except Exception as e:
            logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")
        
        # Fallback to direct file access
        try:
            with open(self.todo2_path) as f:
                data = json.load(f)

            # Generate new Todo2 ID
            existing_ids = [t.get('id', '') for t in data.get('todos', [])]
            todo2_id = f"SHARED-{shared['id']}"
            if todo2_id in existing_ids:
                todo2_id = f"SHARED-{shared['id']}-{datetime.now().strftime('%Y%m%d')}"

            new_task = {
                'id': todo2_id,
                'name': task_name,
                'long_description': task_description,
                'status': task_status,
                'created': datetime.now(timezone.utc).isoformat(),
                'lastModified': datetime.now(timezone.utc).isoformat(),
                'priority': 'medium',
                'tags': task_tags
            }

            data.setdefault('todos', []).append(new_task)

            with open(self.todo2_path, 'w') as f:
                json.dump(data, f, indent=2)

            return new_task
        except Exception as e:
            logger.error(f"Error creating Todo2 task from shared TODO: {e}")
            return {}

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from sync results."""
        insights = []

        matches = len(analysis_results.get('matches', []))
        conflicts = len(analysis_results.get('conflicts', []))
        new_shared = len(analysis_results.get('new_shared_todos', []))
        new_todo2 = len(analysis_results.get('new_todo2_tasks', []))
        updates = len(analysis_results.get('updates', []))

        insights.append("**Synchronization Summary:**")
        insights.append(f"- {matches} tasks matched between systems")
        insights.append(f"- {conflicts} status conflicts detected")
        insights.append(f"- {new_shared} new shared TODOs (will create Todo2 tasks)")
        insights.append(f"- {new_todo2} new Todo2 tasks (not in shared TODO)")
        insights.append(f"- {updates} updates performed")

        if conflicts > 0:
            insights.append(f"\n⚠️ {conflicts} conflicts resolved (Todo2 status preferred)")

        if new_shared > 0:
            insights.append(f"\n✅ {new_shared} new Todo2 tasks will be created")

        return '\n'.join(insights)

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate sync report."""
        timestamp = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
        dry_run = analysis_results.get('dry_run', False)

        matches = analysis_results.get('matches', [])
        conflicts = analysis_results.get('conflicts', [])
        new_shared = analysis_results.get('new_shared_todos', [])
        new_todo2 = analysis_results.get('new_todo2_tasks', [])
        updates = analysis_results.get('updates', [])

        conflict_details = ""
        if conflicts:
            conflict_details = "\n### Conflicts Resolved\n\n"
            for conflict in conflicts[:10]:  # Show first 10
                match = conflict['match']
                conflict_details += f"- **TODO {match['shared']['id']}**: {match['shared']['description'][:50]}...\n"
                conflict_details += f"  - Shared: {conflict['shared_status']} → Todo2: {conflict['todo2_status']}\n"
                conflict_details += "  - Resolution: Updated shared TODO to match Todo2\n\n"

        new_tasks_details = ""
        if new_shared:
            new_tasks_details = "\n### New Todo2 Tasks Created\n\n"
            for shared in new_shared[:10]:  # Show first 10
                new_tasks_details += f"- **TODO {shared['id']}**: {shared['description']}\n"
                new_tasks_details += f"  - Owner: {shared['owner']}\n"
                new_tasks_details += f"  - Status: {shared['status']}\n\n"

        return f"""# Shared TODO Table Synchronization Report

*Generated: {timestamp}*
*Generated By: Intelligent TODO Synchronization*
*Mode: {'DRY RUN' if dry_run else 'LIVE SYNC'}*

## Executive Summary

**Synchronization Status:** {'Simulated' if dry_run else 'Completed'} ✅

**Key Metrics:**
- Tasks Matched: {len(matches)}
- Conflicts Detected: {len(conflicts)}
- New Shared TODOs: {len(new_shared)}
- New Todo2 Tasks: {len(new_todo2)}
- Updates Performed: {len(updates)}

---

## Insights

{insights}

{conflict_details}

{new_tasks_details}

---

## Detailed Results

### Matches Found
{len(matches)} tasks matched between shared TODO and Todo2

### Conflicts Resolved
{len(conflicts)} status conflicts detected and resolved

### New Items
- {len(new_shared)} shared TODOs without Todo2 matches
- {len(new_todo2)} Todo2 tasks without shared TODO matches

### Updates
{len(updates)} synchronization updates performed

---

*This report was generated using intelligent automation with Tractatus Thinking, Sequential Thinking, and bidirectional sync logic.*
"""

    def _needs_networkx(self) -> bool:
        """NetworkX not needed for sync (simple matching)."""
        return False

    def _identify_followup_tasks(self, analysis_results: dict) -> list[dict]:
        """Identify follow-up tasks."""
        followups = []

        conflicts = len(analysis_results.get('conflicts', []))
        if conflicts > 5:
            followups.append({
                'name': 'Review TODO synchronization conflicts',
                'description': f'Review {conflicts} conflicts between shared TODO and Todo2',
                'priority': 'medium',
                'tags': ['todo-sync', 'review']
            })

        new_todo2 = len(analysis_results.get('new_todo2_tasks', []))
        if new_todo2 > 10:
            followups.append({
                'name': 'Review Todo2 tasks not in shared TODO',
                'description': f'Review {new_todo2} Todo2 tasks that are not in shared TODO table',
                'priority': 'low',
                'tags': ['todo-sync', 'review']
            })

        return followups


def load_config(config_path: Optional[Path] = None) -> dict:
    """Load configuration."""
    from project_management_automation.utils import find_project_root
    if config_path is None:
        config_path = find_project_root() / 'scripts' / 'todo_sync_config.json'

    default_config = {
        'dry_run': False,
        'output_path': 'docs/TODO_SYNC_REPORT.md'
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
    parser = argparse.ArgumentParser(description='Intelligent TODO Synchronization')
    parser.add_argument('--config', type=Path, help='Path to config file')
    parser.add_argument('--output', type=Path, help='Output path for report')
    parser.add_argument('--dry-run', action='store_true', help='Simulate sync without making changes')
    args = parser.parse_args()

    config = load_config(args.config)
    if args.dry_run:
        config['dry_run'] = True

    analyzer = TodoSyncAutomation(config)

    try:
        results = analyzer.run()

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
        logger.info(f"Sync {'simulated' if config['dry_run'] else 'completed'}: {len(results.get('results', {}).get('updates', []))} updates")

        sys.exit(0)
    except Exception as e:
        logger.error(f"Error running sync: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()
