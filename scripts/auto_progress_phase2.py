#!/usr/bin/env python3
"""
Auto-Progress Phase 2 Tasks

Monitors Phase 1 research task completion and automatically starts Phase 2 tasks
when their dependencies are satisfied.

Usage: Run from exarp-go project root
"""
import sys
import json
from pathlib import Path

# Add project root to path
project_root = Path(__file__).parent.parent
sys.path.insert(0, str(project_root))

from project_management_automation.utils.project_root import find_project_root
from project_management_automation.utils.todo2_mcp_client import Todo2Client, update_todos_mcp
from project_management_automation.utils.todo2_utils import is_completed_status

PHASE1_RESEARCH_IDS = ['T-0', 'T-5', 'T-6', 'T-7', 'T-8', 'T-16', 'T-17', 'T-18']

def check_and_start_phase2():
    """Check Phase 1 completion and start Phase 2 tasks if ready."""
    client = Todo2Client()
    all_todos = client.list_todos(include_completed=False)
    all_tasks = {t.get('id'): t for t in all_todos.get('todos', [])}
    
    # Find Phase 2 tasks ready to start
    phase2_ready = []
    for task in all_todos.get('todos', []):
        task_id = task.get('id')
        status = task.get('status', 'Todo')
        deps = task.get('dependencies', [])
        
        # Skip if already completed or in progress
        if is_completed_status(status) or status == 'In Progress':
            continue
        
        # Check if depends on Phase 1 research
        phase1_deps = [d for d in deps if d in PHASE1_RESEARCH_IDS]
        other_deps = [d for d in deps if d not in PHASE1_RESEARCH_IDS]
        
        if not phase1_deps:
            continue  # Doesn't depend on Phase 1
        
        # Priority: Tasks that ONLY depend on Phase 1 (no other dependencies)
        # Will auto-start as soon as Phase 1 completes
        if other_deps:
            # Has other dependencies - check if ALL are satisfied
            all_other_deps_satisfied = all(
                dep_id in all_tasks and 
                is_completed_status(all_tasks[dep_id].get('status', 'Todo'))
                for dep_id in other_deps
            )
            if not all_other_deps_satisfied:
                continue  # Wait for other dependencies
        
        # Check if ALL Phase 1 dependencies are satisfied
        all_phase1_satisfied = all(
            dep_id in all_tasks and 
            is_completed_status(all_tasks[dep_id].get('status', 'Todo'))
            for dep_id in phase1_deps
        )
        
        if all_phase1_satisfied:
            phase2_ready.append({
                'id': task_id,
                'name': task.get('name', '')[:60] or task_id,
                'priority': task.get('priority', 'medium'),
                'hours': task.get('estimatedHours', 0)
            })
    
    if not phase2_ready:
        return {'started': 0, 'message': 'No Phase 2 tasks ready yet'}
    
    # Sort by priority
    priority_order = {'critical': 4, 'high': 3, 'medium': 2, 'low': 1}
    phase2_ready.sort(key=lambda t: (-priority_order.get(t['priority'], 2), -t['hours']))
    
    # Start top priority Phase 2 tasks (up to 8 for parallel execution)
    to_start = phase2_ready[:8]
    
    updates = [{'id': t['id'], 'status': 'In Progress'} for t in to_start]
    project_root = find_project_root()
    success = update_todos_mcp(updates, project_root=project_root)
    
    if success:
        return {
            'started': len(to_start),
            'tasks': [{'id': t['id'], 'name': t['name']} for t in to_start],
            'message': f'Started {len(to_start)} Phase 2 tasks'
        }
    else:
        return {'started': 0, 'message': 'Failed to start tasks'}

if __name__ == '__main__':
    result = check_and_start_phase2()
    print(json.dumps(result, indent=2))
