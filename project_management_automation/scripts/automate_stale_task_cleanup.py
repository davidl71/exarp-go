#!/usr/bin/env python3
"""
Stale Task Cleanup Automation

Automatically moves tasks that are "In Progress" but haven't been updated
recently back to "Todo" status. This ensures accurate time tracking by only
counting actual work time, not idle time.

Part of daily automation suite.
"""

import json
import logging
import sys
import time
from datetime import datetime, timezone, timedelta
from pathlib import Path
from typing import Any, Dict, List, Optional

# Add project root to path
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase
from project_management_automation.utils import find_project_root

logger = logging.getLogger(__name__)


class StaleTaskCleanupAutomation(IntelligentAutomationBase):
    """Automation to move stale 'In Progress' tasks back to 'Todo'."""
    
    def __init__(self, config: dict, project_root: Optional[Path] = None):
        """Initialize stale task cleanup automation."""
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Stale Task Cleanup", project_root)
        
        # Configuration (dry_run is set by base class)
        self.stale_threshold_hours = config.get('stale_threshold_hours', 2)
        self.state_file = project_root / ".todo2" / "state.todo2.json"
        
        # Ensure dry_run is set (base class should set it, but ensure it exists)
        if not hasattr(self, 'dry_run'):
            self.dry_run = config.get('dry_run', False)
        
        # Results
        self.moved_tasks = []
        self.active_tasks = []
    
    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is stale task cleanup?"""
        return "What is stale task cleanup? Stale Cleanup = In Progress Tasks × Time Threshold × Status Tracking × Accurate Time Calculation"
    
    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we clean up stale tasks?"""
        return "How do we systematically identify and move stale In Progress tasks back to Todo status?"
    
    def _execute_analysis(self) -> dict:
        """Execute stale task cleanup analysis."""
        return self.run()
    
    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from cleanup results."""
        return self.get_summary()
    
    def _generate_report(self, output_path: Optional[str] = None) -> str:
        """Generate cleanup report."""
        return self.get_summary()
        
    def run(self) -> Dict[str, Any]:
        """Run stale task cleanup."""
        start_time = time.time()
        logger.info(f"Starting stale task cleanup (threshold: {self.stale_threshold_hours}h)")
        
        try:
            # Load tasks
            if not self.state_file.exists():
                logger.warning(f"Todo2 state file not found: {self.state_file}")
                return {
                    'status': 'error',
                    'error': f'State file not found: {self.state_file}',
                    'duration_seconds': time.time() - start_time
                }
            
            with open(self.state_file) as f:
                data = json.load(f)
            
            todos = data.get('todos', [])
            logger.info(f"Found {len(todos)} total tasks")
            
            # Find stale "In Progress" tasks
            now = datetime.now(timezone.utc)
            stale_tasks = []
            active_tasks = []
            
            for task in todos:
                if task.get('status') != 'In Progress':
                    continue
                
                last_modified_str = task.get('lastModified') or task.get('created')
                if not last_modified_str:
                    continue
                
                try:
                    last_modified = datetime.fromisoformat(
                        last_modified_str.replace('Z', '+00:00')
                    )
                    hours_since_update = (now - last_modified).total_seconds() / 3600.0
                    
                    task_info = {
                        'task': task,
                        'id': task.get('id', 'Unknown'),
                        'name': task.get('name', 'Unknown'),
                        'hours_since_update': hours_since_update,
                        'last_modified': last_modified
                    }
                    
                    if hours_since_update > self.stale_threshold_hours:
                        stale_tasks.append(task_info)
                    else:
                        active_tasks.append(task_info)
                        
                except (ValueError, TypeError) as e:
                    logger.debug(f"Could not parse timestamp for task {task.get('id')}: {e}")
                    continue
            
            logger.info(f"Found {len(stale_tasks)} stale tasks and {len(active_tasks)} active tasks")
            
            # Move stale tasks back to "Todo"
            if not self.dry_run and stale_tasks:
                for task_info in stale_tasks:
                    task = task_info['task']
                    task['status'] = 'Todo'
                    
                    # Add change record
                    if 'changes' not in task:
                        task['changes'] = []
                    task['changes'].append({
                        'field': 'status',
                        'oldValue': 'In Progress',
                        'newValue': 'Todo',
                        'timestamp': datetime.now(timezone.utc).isoformat(),
                        'reason': f'Stale task cleanup (no update in {task_info["hours_since_update"]:.1f}h)'
                    })
                    
                    self.moved_tasks.append({
                        'id': task_info['id'],
                        'name': task_info['name'],
                        'hours_since_update': round(task_info['hours_since_update'], 1)
                    })
                
                # Save updated tasks
                with open(self.state_file, 'w') as f:
                    json.dump(data, f, indent=2)
                
                logger.info(f"Moved {len(self.moved_tasks)} stale tasks back to Todo")
            
            # Prepare results
            self.active_tasks = [
                {
                    'id': t['id'],
                    'name': t['name'],
                    'hours_since_update': round(t['hours_since_update'], 1)
                }
                for t in active_tasks
            ]
            
            duration = time.time() - start_time
            
            result = {
                'status': 'success',
                'stale_threshold_hours': self.stale_threshold_hours,
                'stale_tasks_found': len(stale_tasks),
                'tasks_moved': len(self.moved_tasks) if not self.dry_run else 0,
                'active_tasks': len(active_tasks),
                'moved_tasks': self.moved_tasks,
                'active_tasks_list': self.active_tasks,
                'dry_run': self.dry_run,
                'duration_seconds': duration
            }
            
            return result
            
        except Exception as e:
            logger.error(f"Error in stale task cleanup: {e}", exc_info=True)
            return {
                'status': 'error',
                'error': str(e),
                'duration_seconds': time.time() - start_time
            }
    
    def get_summary(self) -> str:
        """Get human-readable summary."""
        if self.dry_run:
            mode = "DRY RUN"
        else:
            mode = "APPLIED"
        
        summary = f"Stale Task Cleanup ({mode})\n"
        summary += f"Threshold: {self.stale_threshold_hours} hours\n"
        summary += f"Moved to Todo: {len(self.moved_tasks)} tasks\n"
        summary += f"Active tasks: {len(self.active_tasks)} tasks\n"
        
        if self.moved_tasks:
            summary += "\nTasks moved back to Todo:\n"
            for task in self.moved_tasks:
                summary += f"  - {task['id']}: {task['name'][:50]} ({task['hours_since_update']:.1f}h stale)\n"
        
        return summary


def main():
    """CLI entry point."""
    import argparse
    
    parser = argparse.ArgumentParser(description='Clean up stale In Progress tasks')
    parser.add_argument('--dry-run', action='store_true', help='Preview changes without applying')
    parser.add_argument('--threshold', type=float, default=2.0, help='Hours before task is considered stale (default: 2.0)')
    parser.add_argument('--output', type=str, help='Output file for report')
    
    args = parser.parse_args()
    
    # Setup logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Find project root
    project_root = find_project_root()
    
    # Create config
    config = {
        'dry_run': args.dry_run,
        'stale_threshold_hours': args.threshold,
        'output_path': args.output
    }
    
    # Run automation
    automation = StaleTaskCleanupAutomation(config, project_root)
    results = automation.run()
    
    # Print summary
    print(automation.get_summary())
    
    # Save report if requested
    if args.output:
        output_path = Path(args.output)
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, 'w') as f:
            json.dump(results, f, indent=2)
        print(f"\nReport saved to: {output_path}")
    
    # Exit code
    sys.exit(0 if results['status'] == 'success' else 1)


if __name__ == '__main__':
    main()

