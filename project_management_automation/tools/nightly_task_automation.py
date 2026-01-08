"""
Nightly Task Automation Tool

Automatically executes background-capable TODO2 tasks in parallel across multiple hosts.
Moves interactive tasks to Review status and proceeds to next tasks.
Also includes batch approval of research tasks that don't need clarification.

Memory Integration:
- Recalls task context before execution
- Saves execution results for future reference
"""

import json
import logging
import os
import socket
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Any, List, Optional

nightly_logger = logging.getLogger(__name__)

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

try:
    from tools.intelligent_automation_base import IntelligentAutomationBase
    from utils.task_locking import atomic_assign_task
    from utils.todo2_utils import is_pending_status, is_review_status
except ImportError:
    # Fallback if base class not available
    class IntelligentAutomationBase:
        pass
    # Fallback for atomic assignment
    def atomic_assign_task(*args, **kwargs):
        return (False, "Import error")


def _get_local_ip_addresses() -> list[str]:
    """Get all local IP addresses (excluding localhost)."""
    local_ips = []

    # Get hostname
    try:
        hostname = socket.gethostname()
        local_ips.append(hostname)
        # Also add FQDN if available
        fqdn = socket.getfqdn()
        if fqdn != hostname:
            local_ips.append(fqdn)
    except Exception:
        pass

    # Get IP addresses from all interfaces
    try:
        result = subprocess.run(
            ["ifconfig"],
            capture_output=True,
            text=True,
            timeout=5
        )
        if result.returncode == 0:
            for line in result.stdout.split('\n'):
                if 'inet ' in line and '127.0.0.1' not in line:
                    parts = line.split()
                    # Find the IP address (usually after 'inet')
                    for i, part in enumerate(parts):
                        if part == 'inet' and i + 1 < len(parts):
                            ip = parts[i + 1]
                            # Remove netmask if present (e.g., "192.168.1.1/24" -> "192.168.1.1")
                            ip = ip.split('/')[0]
                            if ip not in local_ips and ip != '127.0.0.1':
                                local_ips.append(ip)
    except Exception:
        pass

    # Also try to get primary IP via socket (fallback)
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        local_ip = s.getsockname()[0]
        if local_ip not in local_ips:
            local_ips.append(local_ip)
        s.close()
    except Exception:
        pass

    return local_ips


def _is_local_host(hostname: str) -> bool:
    """
    Determine if a hostname/IP refers to the local machine.

    Args:
        hostname: Hostname or IP address (may include user@ prefix)

    Returns:
        True if hostname refers to local machine
    """
    # Remove user@ prefix if present
    host = hostname.split('@')[-1] if '@' in hostname else hostname

    # Check if it's localhost
    if host in ['localhost', '127.0.0.1', '::1']:
        return True

    # Get local IPs and hostname
    local_ips = _get_local_ip_addresses()
    local_hostname = socket.gethostname()

    # Check if host matches local IP or hostname
    if host in local_ips or host == local_hostname:
        return True

    # Try to resolve hostname to IP and compare
    try:
        resolved_ip = socket.gethostbyname(host)
        if resolved_ip in local_ips or resolved_ip == '127.0.0.1':
            return True
    except Exception:
        pass

    return False


def _find_project_root(start_path: Path) -> Path:
    """
    Find project root by looking for .git directory or other markers.
    Falls back to relative path detection if markers not found.
    """
    # Try environment variable first
    env_root = os.getenv('PROJECT_ROOT') or os.getenv('WORKSPACE_PATH')
    if env_root:
        root_path = Path(env_root)
        if root_path.exists():
            return root_path.resolve()

    # Try relative path detection (assumes standard structure)
    current = start_path
    for _ in range(5):  # Go up max 5 levels
        # Check for project markers
        if (current / '.git').exists() or (current / '.todo2').exists() or (current / 'CMakeLists.txt').exists() or (current / 'go.mod').exists():
            return current.resolve()
        if current.parent == current:  # Reached filesystem root
            break
        current = current.parent

    # Fallback to relative path (assumes project-management-automation/project_management_automation/tools/file.py)
    return start_path.parent.parent.parent.resolve()


class NightlyTaskAutomation(IntelligentAutomationBase):
    """Automated nightly task execution across parallel hosts."""

    def __init__(self):
        self.project_root = _find_project_root(Path(__file__))
        self.todo2_state_file = self.project_root / ".todo2" / "state.todo2.json"
        self.batch_script = self.project_root / "scripts" / "batch_update_todos.py"
        self.agent_hostnames = self._load_agent_hostnames()

    def _load_agent_hostnames(self) -> dict[str, str]:
        """Load agent hostname configuration from environment or config file."""
        import os

        # Load from environment variable (JSON format)
        # Format: EXARP_AGENT_HOSTNAMES='{"ubuntu": {"hostname": "user@host", "project_path": "~/project", "type": "ubuntu"}}'
        env_hostnames = os.environ.get("EXARP_AGENT_HOSTNAMES", "{}")
        try:
            default_hostnames = json.loads(env_hostnames)
        except json.JSONDecodeError:
            default_hostnames = {}

        # Try to read from file if it exists
        hostnames_file = self.project_root / "docs" / "AGENT_HOSTNAMES.md"
        if hostnames_file.exists():
            try:
                hostnames_file.read_text()
                # Parse markdown file (simple extraction)
                # In production, use proper markdown parser
                pass
            except Exception:
                pass

        # Auto-detect local agents and mark them appropriately
        for _agent_name, agent_config in default_hostnames.items():
            hostname = agent_config.get("hostname", "")
            if hostname and _is_local_host(hostname):
                # This is the local machine, mark as local
                agent_config["type"] = "local"
                # Use current project root for local agents
                agent_config["project_path"] = str(self.project_root)

        return default_hostnames

    def _load_todo2_state(self) -> dict[str, Any]:
        """Load TODO2 state file."""
        if not self.todo2_state_file.exists():
            return {"todos": []}

        try:
            with open(self.todo2_state_file) as f:
                return json.load(f)
        except Exception as e:
            return {"todos": [], "error": str(e)}

    def _save_todo2_state(self, state: dict[str, Any]) -> bool:
        """Save TODO2 state file."""
        try:
            # Create backup
            if self.todo2_state_file.exists():
                backup_file = self.todo2_state_file.with_suffix('.json.bak')
                with open(self.todo2_state_file) as f:
                    backup_file.write_text(f.read())

            with open(self.todo2_state_file, 'w') as f:
                json.dump(state, f, indent=2)
            return True
        except Exception:
            return False

    def _is_background_capable(self, task: dict[str, Any]) -> bool:
        """Determine if task can run in background."""
        task_id = task.get('id', '')
        name = task.get('name', '').lower()
        long_desc = task.get('long_description', '').lower()
        status = task.get('status', '')

        # Skip if not in Todo status (normalized)
        if not is_pending_status(status):
            return False

        # Skip Review status (already reviewed)
        if is_review_status(status):
            return False

        # Interactive indicators (exclude)
        is_review = is_review_status(status)
        needs_clarification = 'clarification required' in long_desc
        needs_user_input = 'user input' in long_desc or 'user interaction' in long_desc
        is_design = 'design' in name and any(x in name for x in ['framework', 'system', 'strategy', 'allocation'])
        is_decision = any(x in name for x in ['decide', 'choose', 'select', 'recommend', 'suggest', 'propose'])
        is_strategy = 'strategy' in name or 'strategy' in long_desc or ('plan' in name and 'workflow' in long_desc)

        is_interactive = is_review or needs_clarification or needs_user_input or (is_design and 'implement' not in name) or (is_decision and 'implement' not in name) or (is_strategy and 'implement' not in name)

        # Background indicators (include)
        is_mcp_extension = task_id.startswith('MCP-EXT')
        is_research = 'research' in name
        is_implementation = any(x in name for x in ['implement', 'create', 'add', 'update', 'fix', 'refactor'])
        is_testing = 'test' in name or 'testing' in name or 'validate' in name
        is_documentation = 'document' in name or 'documentation' in name
        is_configuration = 'config' in name or 'configure' in name or 'setup' in name

        is_background = (is_mcp_extension or is_research or is_implementation or is_testing or is_documentation or is_configuration) and not is_interactive

        return is_background

    def _move_to_review(self, task: dict[str, Any], reason: str) -> dict[str, Any]:
        """Move task to Review status."""
        from project_management_automation.utils.todo2_utils import normalize_status_to_title_case
        task['status'] = normalize_status_to_title_case('Review')

        # Add note comment about why moved to review
        if 'comments' not in task:
            task['comments'] = []

        task['comments'].append({
            'id': f"{task['id']}-C-{int(time.time())}",
            'todoId': task['id'],
            'type': 'note',
            'content': f"**Automated Review:** Moved to Review status by nightly automation. Reason: {reason}",
            'created': datetime.utcnow().isoformat() + 'Z'
        })

        # Update last modified
        task['lastModified'] = datetime.utcnow().isoformat() + 'Z'

        # Add status change to changes array
        if 'changes' not in task:
            task['changes'] = []

        task['changes'].append({
            'field': 'status',
            'oldValue': task.get('status', 'Todo'),
            'newValue': 'Review',
            'timestamp': datetime.utcnow().isoformat() + 'Z'
        })

        return task

    def _recall_task_memories(self, task_id: str) -> dict[str, Any]:
        """Recall memories related to a task before execution."""
        try:
            from .session_memory import recall_task_context
            return recall_task_context(task_id, include_related=True)
        except ImportError:
            nightly_logger.debug(f"Session memory not available for task {task_id}")
            return {"success": False, "error": "Memory system not available"}

    def _save_task_execution_memory(self, task: dict[str, Any], result: dict[str, Any]) -> dict[str, Any]:
        """Save task execution result as memory."""
        try:
            from .session_memory import save_session_insight

            task_id = task.get('id', 'unknown')
            task_title = task.get('title', 'Untitled task')
            status = result.get('status', 'unknown')

            content = f"""Nightly automation executed task.

## Task
- ID: {task_id}
- Title: {task_title}
- Status: {status}
- Host: {result.get('host', 'N/A')}
- Timestamp: {result.get('timestamp', 'N/A')}

## Notes
{result.get('note', 'No additional notes')}
"""

            return save_session_insight(
                title=f"Nightly: {task_title[:50]}",
                content=content,
                category="insight",
                task_id=task_id,
                metadata={"automation_type": "nightly", "status": status}
            )
        except ImportError:
            nightly_logger.debug("Session memory not available for saving execution result")
            return {"success": False, "error": "Memory system not available"}

    def _execute_task_on_host(self, task: dict[str, Any], host_info: dict[str, str]) -> dict[str, Any]:
        """Execute a task on a specific host via SSH."""
        task_id = task.get('id', '')
        hostname = host_info.get('hostname', '')
        host_info.get('project_path', '')
        host_info.get('type', 'ubuntu')

        # ═══ MEMORY INTEGRATION: Recall task context ═══
        task_context = self._recall_task_memories(task_id)
        if task_context.get('success') and task_context.get('memories'):
            nightly_logger.info(f"Task {task_id}: {len(task_context.get('memories', []))} memories recalled")

        # Construct SSH command
        if '@' in hostname:
            # User@host format
            pass
        else:
            # Just hostname (assume current user)
            pass

        # Navigate to project and execute task
        # For now, we'll just mark as in progress and return
        # In production, would use actual task execution via Cursor agent or script

        result = {
            'task_id': task_id,
            'host': hostname,
            'status': 'started',
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'note': 'Task execution would happen here via SSH/Cursor agent',
            'memories_used': task_context.get('memory_count', 0) if task_context.get('success') else 0
        }

        # ═══ MEMORY INTEGRATION: Save execution result ═══
        self._save_task_execution_memory(task, result)

        return result

    def _update_task_status(self, task: dict[str, Any], new_status: str, result_comment: Optional[str] = None) -> dict[str, Any]:
        """Update task status in TODO2 state."""
        from project_management_automation.utils.todo2_utils import normalize_status_to_title_case
        old_status = task.get('status', 'Todo')
        task['status'] = normalize_status_to_title_case(new_status)
        task['lastModified'] = datetime.utcnow().isoformat() + 'Z'

        # Add status change
        if 'changes' not in task:
            task['changes'] = []

        task['changes'].append({
            'field': 'status',
            'oldValue': old_status,
            'newValue': new_status,
            'timestamp': datetime.utcnow().isoformat() + 'Z'
        })

        # Add result comment if provided
        if result_comment:
            if 'comments' not in task:
                task['comments'] = []

            task['comments'].append({
                'id': f"{task['id']}-C-{int(time.time())}",
                'todoId': task['id'],
                'type': 'result',
                'content': result_comment,
                'created': datetime.utcnow().isoformat() + 'Z'
            })

        return task

    def _check_working_copy_health(self) -> dict[str, Any]:
        """Check working copy health before task execution."""
        try:
            from tools.working_copy_health import check_working_copy_health
            return check_working_copy_health(check_remote=True)
        except Exception as e:
            return {
                "status": "error",
                "error": f"Failed to check working copy health: {e}",
                "summary": {}
            }

    def run_nightly_automation(
        self,
        max_tasks_per_host: int = 5,
        max_parallel_tasks: int = 10,
        priority_filter: Optional[str] = None,
        tag_filter: Optional[List[str]] = None,
        dry_run: bool = False
    ) -> dict[str, Any]:
        """
        Run nightly automation across parallel hosts.

        Args:
            max_tasks_per_host: Maximum tasks to assign per host
            max_parallel_tasks: Maximum total parallel tasks
            priority_filter: Filter by priority ('high', 'medium', 'low')
            tag_filter: Filter by tags (list of tag strings)
            dry_run: If True, don't actually execute, just report

        Returns:
            Dictionary with execution results
        """
        # Check working copy health before execution
        working_copy_status = self._check_working_copy_health()

        state = self._load_todo2_state()
        todos = state.get('todos', [])

        # Filter tasks
        background_tasks = []
        interactive_tasks = []

        for task in todos:
            if self._is_background_capable(task):
                # Apply filters
                if priority_filter and task.get('priority') != priority_filter:
                    continue

                if tag_filter:
                    task_tags = task.get('tags', [])
                    if not any(tag in task_tags for tag in tag_filter):
                        continue

                background_tasks.append(task)
            else:
                # Check if interactive task should be moved to review
                if is_pending_status(task.get('status', '')):
                    # Check if it needs user input
                    long_desc = task.get('long_description', '').lower()
                    needs_clarification = 'clarification required' in long_desc
                    needs_user_input = 'user input' in long_desc or 'user interaction' in long_desc

                    if needs_clarification or needs_user_input:
                        interactive_tasks.append(task)

        # Move interactive tasks to Review (if not dry run)
        moved_to_review = []
        if not dry_run:
            for task in interactive_tasks[:max_parallel_tasks]:  # Limit moves
                task = self._move_to_review(task, "Requires user input or clarification")
                moved_to_review.append(task['id'])

                # Update in state
                for i, t in enumerate(todos):
                    if t.get('id') == task['id']:
                        todos[i] = task
                        break

        # Assign background tasks to hosts
        assigned_tasks = []
        task_assignments = {}

        available_hosts = list(self.agent_hostnames.keys())
        host_index = 0

        for task in background_tasks[:max_parallel_tasks]:
            if len(assigned_tasks) >= max_parallel_tasks:
                break

            # Round-robin assignment to hosts
            host_key = available_hosts[host_index % len(available_hosts)]
            host_info = self.agent_hostnames[host_key]

            # Check host task limit
            host_task_count = sum(1 for t in task_assignments.values() if t['host'] == host_key)
            if host_task_count >= max_tasks_per_host:
                # Skip to next host
                host_index += 1
                host_key = available_hosts[host_index % len(available_hosts)]
                host_info = self.agent_hostnames[host_key]
                host_task_count = sum(1 for t in task_assignments.values() if t['host'] == host_key)
                if host_task_count >= max_tasks_per_host:
                    continue  # All hosts at capacity

            # Assign task
            assignment = {
                'task_id': task['id'],
                'task_name': task.get('name', ''),
                'host': host_key,
                'hostname': host_info['hostname'],
                'project_path': host_info['project_path'],
                'status': 'assigned'
            }

            task_assignments[task['id']] = assignment

            # Update task status to In Progress (if not dry run)
            if not dry_run:
                # Use atomic assignment to prevent race conditions
                success, error = atomic_assign_task(
                    task_id=task['id'],
                    assignee_name=host_key,
                    assignee_type='host',
                    hostname=host_info['hostname'],
                    assigned_by='nightly_automation',
                    timeout=5.0,
                )

                if not success:
                    # Task was assigned by another agent or assignment failed
                    nightly_logger.warning(
                        f"Failed to assign task {task['id']} to {host_key}: {error}. "
                        "Task may have been assigned by another agent."
                    )
                    # Skip this task and try next
                    host_index += 1
                    continue

                # Reload task to get updated state
                state = self._load_todo2_state()
                todos = state.get('todos', [])
                for _i, t in enumerate(todos):
                    if t.get('id') == task['id']:
                        task = t
                        break

                # Update task status to In Progress
                task = self._update_task_status(task, 'In Progress',
                    f"Assigned to {host_key} agent for automated execution")

                # Update in state
                for i, t in enumerate(todos):
                    if t.get('id') == task['id']:
                        todos[i] = task
                        break

            assigned_tasks.append(task)
            host_index += 1

        # Batch approve research tasks that don't need clarification (if not dry run)
        batch_approved_count = 0
        if not dry_run and self.batch_script.exists():
            try:
                # Count Review tasks with no clarification before approval
                review_tasks_before = [t for t in todos if is_review_status(t.get('status', ''))]

                # Run batch approval script for Review tasks with no clarification needed
                result = subprocess.run(
                    [
                        sys.executable,
                        str(self.batch_script),
                        'approve',
                        '--status', 'Review',
                        '--clarification-none',
                        '--new-status', 'Todo',
                        '--yes'  # Skip confirmation in automated runs
                    ],
                    cwd=str(self.project_root),
                    capture_output=True,
                    text=True,
                    timeout=30
                )

                if result.returncode == 0:
                    # Reload state to get updated task list
                    state = self._load_todo2_state()
                    todos = state.get('todos', [])

                    # Count Review tasks after approval
                    review_tasks_after = [t for t in todos if is_review_status(t.get('status', ''))]
                    batch_approved_count = len(review_tasks_before) - len(review_tasks_after)
            except Exception as e:
                # Log error but don't fail the automation
                print(f"Warning: Batch approval failed: {e}", file=sys.stderr)

        # Save state (if not dry run)
        if not dry_run:
            state['todos'] = todos
            self._save_todo2_state(state)

        # Prepare results
        results = {
            'timestamp': datetime.utcnow().isoformat() + 'Z',
            'dry_run': dry_run,
            'working_copy_status': working_copy_status,
            'summary': {
                'background_tasks_found': len(background_tasks),
                'interactive_tasks_found': len(interactive_tasks),
                'tasks_assigned': len(assigned_tasks),
                'tasks_moved_to_review': len(moved_to_review),
                'tasks_batch_approved': batch_approved_count,
                'hosts_used': len({a['host'] for a in task_assignments.values()}),
                'working_copy_warnings': working_copy_status.get('summary', {}).get('warning_agents', 0)
            },
            'assigned_tasks': [
                {
                    'task_id': a['task_id'],
                    'task_name': a['task_name'],
                    'host': a['host'],
                    'hostname': a['hostname']
                }
                for a in task_assignments.values()
            ],
            'moved_to_review': moved_to_review,
            'background_tasks_remaining': len(background_tasks) - len(assigned_tasks)
        }

        # ═══ MEMORY INTEGRATION: Save overall nightly results ═══
        if not dry_run:
            self._save_nightly_summary(results)

        return results

    def _save_nightly_summary(self, results: dict[str, Any]) -> dict[str, Any]:
        """Save nightly automation summary as memory."""
        try:
            from .session_memory import save_session_insight

            summary = results.get('summary', {})
            content = f"""Nightly automation run completed.

## Summary
- Background tasks found: {summary.get('background_tasks_found', 0)}
- Tasks assigned: {summary.get('tasks_assigned', 0)}
- Tasks moved to review: {summary.get('tasks_moved_to_review', 0)}
- Tasks batch approved: {summary.get('tasks_batch_approved', 0)}
- Hosts used: {summary.get('hosts_used', 0)}

## Assigned Tasks
{chr(10).join('- ' + t.get('task_name', 'Untitled') + ' → ' + t.get('host', 'N/A') for t in results.get('assigned_tasks', [])[:10]) or 'None'}

## Remaining Background Tasks
{results.get('background_tasks_remaining', 0)} tasks remaining for future runs.
"""

            return save_session_insight(
                title=f"Nightly: {summary.get('tasks_assigned', 0)} assigned, {summary.get('tasks_batch_approved', 0)} approved",
                content=content,
                category="insight",
                metadata={"automation_type": "nightly_summary"}
            )
        except ImportError:
            nightly_logger.debug("Session memory not available for nightly summary")
            return {"success": False, "error": "Memory system not available"}


def run_nightly_task_automation(
    max_tasks_per_host: int = 5,
    max_parallel_tasks: int = 10,
    priority_filter: Optional[str] = None,
    tag_filter: Optional[List[str]] = None,
    dry_run: bool = False
) -> dict[str, Any]:
    """
    MCP Tool: Run nightly task automation across parallel hosts.

    Automatically executes background-capable TODO2 tasks in parallel across multiple hosts.
    Moves interactive tasks requiring user input to Review status.

    Args:
        max_tasks_per_host: Maximum tasks to assign per host (default: 5)
        max_parallel_tasks: Maximum total parallel tasks (default: 10)
        priority_filter: Filter by priority - 'high', 'medium', or 'low' (optional)
        tag_filter: Filter by tags - list of tag strings (optional)
        dry_run: If true, don't execute, just report what would happen (default: false)

    Returns:
        Dictionary with execution results including assigned tasks, moved tasks, and summary
    """
    automation = NightlyTaskAutomation()
    return automation.run_nightly_automation(
        max_tasks_per_host=max_tasks_per_host,
        max_parallel_tasks=max_parallel_tasks,
        priority_filter=priority_filter,
        tag_filter=tag_filter,
        dry_run=dry_run
    )
