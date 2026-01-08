"""
Task Assignee Management Tool

Manages task assignments across:
- Background agents (backend-agent, web-agent, tui-agent, etc.)
- Human developer (you)
- Remote hosts (ubuntu server, macbook-pro, etc.)

Provides cross-host visibility so different machines understand what others are working on.

Assignee Format:
    {
        "type": "agent" | "human" | "host",
        "name": "backend-agent" | "david" | "ubuntu-server",
        "hostname": "192.168.1.100" | null,  # For remote hosts
        "assigned_at": "2025-11-27T10:00:00Z",
        "assigned_by": "nightly_automation" | "manual"
    }
"""

import json
import logging
import os
import socket
import time
from datetime import datetime
from pathlib import Path
from typing import Any, List, Literal, Optional

from ..utils.task_locking import atomic_assign_task, atomic_batch_assign
from ..utils.todo2_utils import is_pending_status, normalize_status

logger = logging.getLogger(__name__)

# ═══════════════════════════════════════════════════════════════════════════════
# ASSIGNEE TYPES
# ═══════════════════════════════════════════════════════════════════════════════

AssigneeType = Literal["agent", "human", "host"]

# Well-known assignees
KNOWN_AGENTS = [
    "backend-agent",
    "backend-data-agent",
    "backend-market-data-agent",
    "backend-mock-agent",
    "web-agent",
    "tui-agent",
    "desktop-agent",
    "ipad-agent",
]

KNOWN_HUMANS = [
    "david",  # Primary developer
    "human",  # Generic human assignee
]

# Hosts are loaded from EXARP_AGENT_HOSTNAMES environment variable


def _find_project_root() -> Path:
    """Find project root by looking for markers."""
    env_root = os.getenv("PROJECT_ROOT") or os.getenv("WORKSPACE_PATH")
    if env_root:
        return Path(env_root).resolve()

    current = Path.cwd()
    for _ in range(5):
        if (current / ".git").exists() or (current / ".todo2").exists() or (current / "CMakeLists.txt").exists() or (current / "go.mod").exists():
            return current.resolve()
        if current.parent == current:
            break
        current = current.parent

    return Path.cwd().resolve()


def _load_todo2_state() -> dict[str, Any]:
    """Load Todo2 state file."""
    project_root = _find_project_root()
    todo2_file = project_root / ".todo2" / "state.todo2.json"

    if not todo2_file.exists():
        return {"todos": []}

    try:
        with open(todo2_file) as f:
            return json.load(f)
    except Exception as e:
        logger.error(f"Error loading Todo2 state: {e}")
        return {"todos": [], "error": str(e)}


def _save_todo2_state(state: dict[str, Any]) -> bool:
    """Save Todo2 state file."""
    project_root = _find_project_root()
    todo2_file = project_root / ".todo2" / "state.todo2.json"

    try:
        # Create backup
        if todo2_file.exists():
            backup_file = todo2_file.with_suffix('.json.bak')
            backup_file.write_text(todo2_file.read_text())

        with open(todo2_file, 'w') as f:
            json.dump(state, f, indent=2)
        return True
    except Exception as e:
        logger.error(f"Error saving Todo2 state: {e}")
        return False


def _get_current_hostname() -> str:
    """Get current machine's hostname."""
    return socket.gethostname()


def _get_known_hosts() -> dict[str, dict]:
    """Load known hosts from environment."""
    env_hostnames = os.environ.get("EXARP_AGENT_HOSTNAMES", "{}")
    try:
        return json.loads(env_hostnames)
    except json.JSONDecodeError:
        return {}


def _detect_agents_from_project() -> list[str]:
    """Detect available agents from project structure."""
    project_root = _find_project_root()
    agents_dir = project_root / "agents"

    agents = []
    if agents_dir.exists():
        for agent_dir in agents_dir.iterdir():
            if agent_dir.is_dir():
                cursor_agent_file = agent_dir / "cursor-agent.json"
                if cursor_agent_file.exists():
                    try:
                        config = json.loads(cursor_agent_file.read_text())
                        agents.append(config.get("name", f"{agent_dir.name}-agent"))
                    except Exception:
                        agents.append(f"{agent_dir.name}-agent")

    return agents


# ═══════════════════════════════════════════════════════════════════════════════
# ASSIGNEE MANAGEMENT
# ═══════════════════════════════════════════════════════════════════════════════

def assign_task(
    task_id: str,
    assignee_name: str,
    assignee_type: AssigneeType = "agent",
    hostname: Optional[str] = None,
    dry_run: bool = False
) -> str:
    """
    [HINT: Assign task. Sets assignee on a todo2 task. Enables cross-host visibility.]

    Assign a todo2 task to an agent, human, or remote host.

    Args:
        task_id: The task ID to assign (e.g., "T-123", "SHARED-4")
        assignee_name: Name of assignee (e.g., "backend-agent", "david", "ubuntu-server")
        assignee_type: Type of assignee - "agent", "human", or "host"
        hostname: IP or hostname for remote hosts (optional)
        dry_run: Preview without making changes

    Returns:
        JSON with assignment result

    Example:
        assign_task("T-123", "backend-agent", "agent")
        assign_task("T-456", "david", "human")
        assign_task("T-789", "ubuntu-server", "host", hostname="192.168.1.100")
    """
    start_time = time.time()

    try:
        state = _load_todo2_state()
        todos = state.get("todos", [])

        # Find the task
        task = None
        task_index = -1
        for i, t in enumerate(todos):
            if t.get("id") == task_id:
                task = t
                task_index = i
                break

        if task is None:
            return json.dumps({
                "success": False,
                "error": f"Task {task_id} not found",
            }, indent=2)

        # Build assignee object
        assignee = {
            "type": assignee_type,
            "name": assignee_name,
            "hostname": hostname or (_get_current_hostname() if assignee_type == "host" else None),
            "assigned_at": datetime.utcnow().isoformat() + "Z",
            "assigned_by": "manual",
        }

        # Preview or apply
        if dry_run:
            return json.dumps({
                "success": True,
                "dry_run": True,
                "task_id": task_id,
                "task_name": task.get("name", ""),
                "current_assignee": task.get("assignee"),
                "new_assignee": assignee,
            }, indent=2)

        # Use atomic assignment to prevent race conditions
        old_assignee = task.get("assignee")

        # Check if already assigned (atomic check)
        if old_assignee:
            existing_name = old_assignee.get("name", "unknown")
            existing_type = old_assignee.get("type", "unknown")
            return json.dumps({
                "success": False,
                "error": f"Task already assigned to {existing_type}:{existing_name}",
                "task_id": task_id,
                "current_assignee": old_assignee,
            }, indent=2)

        # Atomically assign task (prevents concurrent assignment)
        success, error = atomic_assign_task(
            task_id=task_id,
            assignee_name=assignee_name,
            assignee_type=assignee_type,
            hostname=hostname or (_get_current_hostname() if assignee_type == "host" else None),
            assigned_by="manual",
            timeout=5.0,
        )

        if not success:
            return json.dumps({
                "success": False,
                "error": error or "Failed to assign task",
                "task_id": task_id,
            }, indent=2)

        # Reload task to add change tracking (assignment already done atomically)
        state = _load_todo2_state()
        todos = state.get("todos", [])
        for i, t in enumerate(todos):
            if t.get("id") == task_id:
                task = t
                task_index = i
                break

        # Add change record
        if "changes" not in task:
            task["changes"] = []
        task["changes"].append({
            "field": "assignee",
            "oldValue": old_assignee,
            "newValue": task.get("assignee"),
            "timestamp": datetime.utcnow().isoformat() + "Z",
        })

        # Save state with change tracking
        todos[task_index] = task
        state["todos"] = todos
        _save_todo2_state(state)

        duration = time.time() - start_time
        return json.dumps({
            "success": True,
            "task_id": task_id,
            "task_name": task.get("name", ""),
            "assignee": task.get("assignee"),
            "previous_assignee": old_assignee,
            "duration_ms": round(duration * 1000, 2),
            "locked": True,  # Indicates atomic assignment was used
        }, indent=2)

    except Exception as e:
        logger.error(f"Error assigning task: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


def unassign_task(task_id: str, dry_run: bool = False) -> str:
    """
    [HINT: Unassign task. Removes assignee from a todo2 task.]

    Remove the assignee from a todo2 task.

    Args:
        task_id: The task ID to unassign
        dry_run: Preview without making changes

    Returns:
        JSON with unassignment result
    """
    try:
        state = _load_todo2_state()
        todos = state.get("todos", [])

        # Find the task
        task = None
        task_index = -1
        for i, t in enumerate(todos):
            if t.get("id") == task_id:
                task = t
                task_index = i
                break

        if task is None:
            return json.dumps({
                "success": False,
                "error": f"Task {task_id} not found",
            }, indent=2)

        old_assignee = task.get("assignee")

        if dry_run:
            return json.dumps({
                "success": True,
                "dry_run": True,
                "task_id": task_id,
                "current_assignee": old_assignee,
            }, indent=2)

        # Remove assignee
        if "assignee" in task:
            del task["assignee"]
        task["lastModified"] = datetime.utcnow().isoformat() + "Z"

        # Add change record
        if "changes" not in task:
            task["changes"] = []
        task["changes"].append({
            "field": "assignee",
            "oldValue": old_assignee,
            "newValue": None,
            "timestamp": datetime.utcnow().isoformat() + "Z",
        })

        # Save state
        todos[task_index] = task
        state["todos"] = todos
        _save_todo2_state(state)

        return json.dumps({
            "success": True,
            "task_id": task_id,
            "previous_assignee": old_assignee,
        }, indent=2)

    except Exception as e:
        logger.error(f"Error unassigning task: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


def list_tasks_by_assignee(
    assignee_name: Optional[str] = None,
    assignee_type: AssigneeType | None = None,
    status_filter: Optional[str] = None,
    include_unassigned: bool = False
) -> str:
    """
    [HINT: List by assignee. Shows tasks grouped by who they're assigned to. Cross-host visibility.]

    List tasks grouped by assignee for cross-host visibility.

    Args:
        assignee_name: Filter to specific assignee (optional)
        assignee_type: Filter by type - "agent", "human", "host" (optional)
        status_filter: Filter by status - "Todo", "In Progress", etc. (optional)
        include_unassigned: Include unassigned tasks (default: False)

    Returns:
        JSON with tasks grouped by assignee

    Example:
        list_tasks_by_assignee()  # All assigned tasks grouped
        list_tasks_by_assignee(assignee_name="backend-agent")  # Just backend tasks
        list_tasks_by_assignee(assignee_type="host")  # All host-assigned tasks
    """
    try:
        state = _load_todo2_state()
        todos = state.get("todos", [])

        # Group by assignee
        by_assignee: dict[str, list[dict]] = {}
        unassigned: list[dict] = []

        for task in todos:
            # Apply status filter
            if status_filter and task.get("status") != status_filter:
                continue

            assignee = task.get("assignee")

            if assignee is None:
                if include_unassigned:
                    unassigned.append({
                        "id": task.get("id"),
                        "name": task.get("name", "")[:60],
                        "status": task.get("status"),
                        "priority": task.get("priority"),
                    })
                continue

            # Apply filters
            if assignee_name and assignee.get("name") != assignee_name:
                continue
            if assignee_type and assignee.get("type") != assignee_type:
                continue

            # Group key
            key = f"{assignee.get('type', 'unknown')}:{assignee.get('name', 'unknown')}"
            if key not in by_assignee:
                by_assignee[key] = []

            by_assignee[key].append({
                "id": task.get("id"),
                "name": task.get("name", "")[:60],
                "status": task.get("status"),
                "priority": task.get("priority"),
                "assigned_at": assignee.get("assigned_at"),
                "hostname": assignee.get("hostname"),
            })

        # Build result
        result = {
            "by_assignee": {},
            "summary": {
                "total_assigned": sum(len(tasks) for tasks in by_assignee.values()),
                "assignee_count": len(by_assignee),
                "unassigned_count": len(unassigned),
            },
            "current_host": _get_current_hostname(),
            "timestamp": datetime.utcnow().isoformat() + "Z",
        }

        # Format by_assignee with metadata
        for key, tasks in by_assignee.items():
            assignee_type_parsed, assignee_name_parsed = key.split(":", 1)
            result["by_assignee"][key] = {
                "type": assignee_type_parsed,
                "name": assignee_name_parsed,
                "task_count": len(tasks),
                "tasks": tasks,
            }

        if include_unassigned and unassigned:
            result["unassigned"] = unassigned

        return json.dumps(result, indent=2)

    except Exception as e:
        logger.error(f"Error listing tasks by assignee: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


def get_workload_summary() -> str:
    """
    [HINT: Workload summary. Shows what each agent/host/human is working on. Cross-machine coordination.]

    Get a summary of task distribution across all assignees.
    Useful for understanding workload and coordinating across machines.

    Returns:
        JSON with workload distribution by assignee type and name
    """
    try:
        state = _load_todo2_state()
        todos = state.get("todos", [])

        # Categorize
        agents: dict[str, dict] = {}
        humans: dict[str, dict] = {}
        hosts: dict[str, dict] = {}
        unassigned = {"total": 0, "in_progress": 0, "todo": 0}

        for task in todos:
            assignee = task.get("assignee")
            status = task.get("status", "Unknown")

            if assignee is None:
                unassigned["total"] += 1
                normalized_status = normalize_status(status)
                if normalized_status == "in_progress":
                    unassigned["in_progress"] += 1
                elif normalized_status == "todo":
                    unassigned["todo"] += 1
                continue

            atype = assignee.get("type", "unknown")
            aname = assignee.get("name", "unknown")

            # Select bucket
            if atype == "agent":
                bucket = agents
            elif atype == "human":
                bucket = humans
            else:
                bucket = hosts

            if aname not in bucket:
                bucket[aname] = {
                    "total": 0,
                    "in_progress": 0,
                    "todo": 0,
                    "done": 0,
                    "review": 0,
                    "hostname": assignee.get("hostname"),
                }

            bucket[aname]["total"] += 1
            normalized_status = normalize_status(status)
            if normalized_status == "in_progress":
                bucket[aname]["in_progress"] += 1
            elif normalized_status == "todo":
                bucket[aname]["todo"] += 1
            elif normalized_status == "completed":
                bucket[aname]["done"] += 1
            elif normalized_status == "review":
                bucket[aname]["review"] += 1

        # Get available agents from project
        available_agents = _detect_agents_from_project()
        known_hosts = _get_known_hosts()

        result = {
            "workload": {
                "agents": agents,
                "humans": humans,
                "hosts": hosts,
                "unassigned": unassigned,
            },
            "available": {
                "agents": available_agents,
                "hosts": list(known_hosts.keys()),
                "humans": KNOWN_HUMANS,
            },
            "current_host": _get_current_hostname(),
            "summary": {
                "total_tasks": len(todos),
                "assigned_tasks": len(todos) - unassigned["total"],
                "agent_count": len(agents),
                "host_count": len(hosts),
                "human_count": len(humans),
            },
            "timestamp": datetime.utcnow().isoformat() + "Z",
        }

        return json.dumps(result, indent=2)

    except Exception as e:
        logger.error(f"Error getting workload summary: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


def bulk_assign_tasks(
    task_ids: list[str],
    assignee_name: str,
    assignee_type: AssigneeType = "agent",
    hostname: Optional[str] = None,
    dry_run: bool = False
) -> str:
    """
    [HINT: Bulk assign. Assigns multiple tasks to one assignee at once.]

    Assign multiple tasks to the same assignee atomically (all or nothing).

    Args:
        task_ids: List of task IDs to assign
        assignee_name: Name of assignee
        assignee_type: Type of assignee - "agent", "human", or "host"
        hostname: IP or hostname for remote hosts (optional)
        dry_run: Preview without making changes

    Returns:
        JSON with bulk assignment results
    """
    if dry_run:
        # Preview mode: check each task individually
        results = []
        for task_id in task_ids:
            state = _load_todo2_state()
            todos = state.get("todos", [])
            task = next((t for t in todos if t.get("id") == task_id), None)
            if task:
                results.append({
                    "task_id": task_id,
                    "current_assignee": task.get("assignee"),
                    "would_assign_to": assignee_name,
                })
            else:
                results.append({
                    "task_id": task_id,
                    "error": "Task not found",
                })

        return json.dumps({
            "success": True,
            "dry_run": True,
            "assignee": {
                "type": assignee_type,
                "name": assignee_name,
                "hostname": hostname,
            },
            "summary": {
                "total": len(task_ids),
                "previewed": len(results),
            },
            "results": results,
        }, indent=2)

    # Use atomic batch assignment for thread-safe bulk assignment
    batch_result = atomic_batch_assign(
        task_ids=task_ids,
        assignee_name=assignee_name,
        assignee_type=assignee_type,
        hostname=hostname,
        timeout=10.0,
    )

    # Format results to match expected structure
    results = []
    for task_id in task_ids:
        if task_id in batch_result["assigned"]:
            results.append({
                "task_id": task_id,
                "success": True,
            })
        else:
            # Find failure reason
            failed_info = next((f for f in batch_result["failed"] if f["task_id"] == task_id), None)
            results.append({
                "task_id": task_id,
                "success": False,
                "error": failed_info["reason"] if failed_info else "Unknown error",
            })

    return json.dumps({
        "success": batch_result["success"],
        "dry_run": False,
        "assignee": {
            "type": assignee_type,
            "name": assignee_name,
            "hostname": hostname,
        },
        "summary": {
            "total": len(task_ids),
            "success": len(batch_result["assigned"]),
            "failed": len(batch_result["failed"]),
        },
        "results": results,
        "locked": True,  # Indicates atomic batch assignment was used
    }, indent=2)


def auto_assign_background_tasks(
    max_tasks_per_agent: int = 5,
    priority_filter: Optional[str] = None,
    dry_run: bool = False
) -> str:
    """
    [HINT: Auto-assign. Distributes background-capable tasks across available agents.]

    Automatically assign unassigned background-capable tasks to available agents
    using round-robin distribution.

    Args:
        max_tasks_per_agent: Maximum tasks to assign per agent (default: 5)
        priority_filter: Only assign tasks with this priority (optional)
        dry_run: Preview without making changes

    Returns:
        JSON with auto-assignment results
    """
    try:
        state = _load_todo2_state()
        todos = state.get("todos", [])

        # Get available agents
        agents = _detect_agents_from_project()
        if not agents:
            return json.dumps({
                "success": False,
                "error": "No agents detected in project",
            }, indent=2)

        # Find unassigned, background-capable tasks
        unassigned_tasks = []
        for task in todos:
            # Skip if assigned
            if task.get("assignee"):
                continue

            # Skip if not in Todo status (normalized)
            if not is_pending_status(task.get("status", "")):
                continue

            # Apply priority filter
            if priority_filter and task.get("priority") != priority_filter:
                continue

            # Check if background-capable (same logic as nightly automation)
            name = task.get("name", "").lower()
            long_desc = task.get("long_description", "").lower()

            # Interactive indicators (exclude)
            needs_clarification = "clarification required" in long_desc
            needs_user_input = "user input" in long_desc or "user interaction" in long_desc
            is_design = "design" in name and any(x in name for x in ["framework", "system", "strategy", "allocation"])
            is_decision = any(x in name for x in ["decide", "choose", "select", "propose"])

            if needs_clarification or needs_user_input or is_design or is_decision:
                continue

            # Background indicators (include)
            task_id = task.get("id", "")
            is_mcp_extension = task_id.startswith("MCP-EXT")
            is_research = "research" in name
            is_implementation = any(x in name for x in ["implement", "create", "add", "update", "fix", "refactor"])
            is_testing = "test" in name or "testing" in name or "validate" in name
            is_documentation = "document" in name or "documentation" in name
            is_configuration = "config" in name or "configure" in name or "setup" in name

            if is_mcp_extension or is_research or is_implementation or is_testing or is_documentation or is_configuration:
                unassigned_tasks.append(task)

        # Round-robin assignment
        assignments = []
        agent_task_counts = dict.fromkeys(agents, 0)
        agent_index = 0

        for task in unassigned_tasks:
            # Find next agent with capacity
            attempts = 0
            while attempts < len(agents):
                agent = agents[agent_index % len(agents)]
                if agent_task_counts[agent] < max_tasks_per_agent:
                    break
                agent_index += 1
                attempts += 1

            if attempts >= len(agents):
                # All agents at capacity
                break

            agent = agents[agent_index % len(agents)]
            agent_task_counts[agent] += 1
            agent_index += 1

            assignments.append({
                "task_id": task.get("id"),
                "task_name": task.get("name", "")[:50],
                "agent": agent,
            })

            if not dry_run:
                # Use atomic assignment to prevent race conditions
                success, error = atomic_assign_task(
                    task_id=task.get("id"),
                    assignee_name=agent,
                    assignee_type="agent",
                    hostname=None,
                    assigned_by="auto_assign",
                    timeout=3.0,
                )

                if not success:
                    # Task was assigned by another agent or assignment failed
                    logger.warning(f"Failed to assign {task.get('id')} to {agent}: {error}")
                    # Remove from assignments list
                    assignments = [a for a in assignments if a["task_id"] != task.get("id")]
                    continue

        # Note: atomic_assign_task already saves state, so we don't need to save again
        # But we reload to get updated state for the response
        if not dry_run:
            state = _load_todo2_state()

        return json.dumps({
            "success": True,
            "dry_run": dry_run,
            "summary": {
                "unassigned_background_tasks": len(unassigned_tasks),
                "tasks_assigned": len(assignments),
                "agents_used": len([a for a, c in agent_task_counts.items() if c > 0]),
            },
            "agent_distribution": {agent: count for agent, count in agent_task_counts.items() if count > 0},
            "assignments": assignments,
        }, indent=2)

    except Exception as e:
        logger.error(f"Error auto-assigning tasks: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


# ═══════════════════════════════════════════════════════════════════════════════
# MCP REGISTRATION
# ═══════════════════════════════════════════════════════════════════════════════

def task_assignee(
    action: str,
    task_id: Optional[str] = None,
    task_ids: Optional[List[str]] = None,
    assignee_name: Optional[str] = None,
    assignee_type: str = "agent",
    hostname: Optional[str] = None,
    status_filter: Optional[str] = None,
    include_unassigned: bool = False,
    max_tasks_per_agent: int = 5,
    priority_filter: Optional[str] = None,
    dry_run: bool = False
) -> str:
    """
    [HINT: Task assignee. Manages task assignments across agents/humans/hosts. Actions: assign, unassign, list, workload, bulk_assign, auto_assign.]

    Consolidated tool for managing todo2 task assignments.
    Enables cross-host visibility of what each agent/human/host is working on.

    Actions:
        - assign: Assign a task to an agent/human/host
        - unassign: Remove assignee from a task
        - list: List tasks grouped by assignee
        - workload: Get workload distribution summary
        - bulk_assign: Assign multiple tasks at once
        - auto_assign: Auto-distribute background tasks to agents

    Args:
        action: One of "assign", "unassign", "list", "workload", "bulk_assign", "auto_assign"
        task_id: Task ID (for assign/unassign)
        task_ids: List of task IDs (for bulk_assign)
        assignee_name: Name of assignee (e.g., "backend-agent", "david")
        assignee_type: Type - "agent", "human", or "host" (default: "agent")
        hostname: IP/hostname for remote hosts (optional)
        status_filter: Filter by status for list action (optional)
        include_unassigned: Include unassigned tasks in list (default: False)
        max_tasks_per_agent: Max tasks per agent for auto_assign (default: 5)
        priority_filter: Priority filter for auto_assign (optional)
        dry_run: Preview without changes (default: False)

    Returns:
        JSON with action results

    Examples:
        task_assignee(action="assign", task_id="T-123", assignee_name="backend-agent")
        task_assignee(action="unassign", task_id="T-123")
        task_assignee(action="list", assignee_name="backend-agent")
        task_assignee(action="workload")
        task_assignee(action="bulk_assign", task_ids=["T-1", "T-2"], assignee_name="web-agent")
        task_assignee(action="auto_assign", max_tasks_per_agent=5, dry_run=True)
    """
    action = action.lower()

    if action == "assign":
        if not task_id or not assignee_name:
            return json.dumps({"error": "assign requires task_id and assignee_name"}, indent=2)
        return assign_task(task_id, assignee_name, assignee_type, hostname, dry_run)

    elif action == "unassign":
        if not task_id:
            return json.dumps({"error": "unassign requires task_id"}, indent=2)
        return unassign_task(task_id, dry_run)

    elif action == "list":
        return list_tasks_by_assignee(assignee_name, assignee_type, status_filter, include_unassigned)

    elif action == "workload":
        return get_workload_summary()

    elif action == "bulk_assign":
        if not task_ids or not assignee_name:
            return json.dumps({"error": "bulk_assign requires task_ids and assignee_name"}, indent=2)
        return bulk_assign_tasks(task_ids, assignee_name, assignee_type, hostname, dry_run)

    elif action == "auto_assign":
        return auto_assign_background_tasks(max_tasks_per_agent, priority_filter, dry_run)

    else:
        return json.dumps({
            "error": f"Unknown action: {action}",
            "valid_actions": ["assign", "unassign", "list", "workload", "bulk_assign", "auto_assign"],
        }, indent=2)


def register_assignee_tools(mcp) -> None:
    """Register assignee management tool with MCP server."""
    try:
        @mcp.tool()
        def task_assignee_tool(
            action: str,
            task_id: Optional[str] = None,
            task_ids: Optional[List[str]] = None,
            assignee_name: Optional[str] = None,
            assignee_type: str = "agent",
            hostname: Optional[str] = None,
            status_filter: Optional[str] = None,
            include_unassigned: bool = False,
            max_tasks_per_agent: int = 5,
            priority_filter: Optional[str] = None,
            dry_run: bool = False
        ) -> str:
            """
            [HINT: Task assignee. Manages task assignments across agents/humans/hosts. Actions: assign, unassign, list, workload, bulk_assign, auto_assign.]

            Consolidated tool for managing todo2 task assignments.
            Enables cross-host visibility of what each agent/human/host is working on.

            Actions:
                - assign: Assign a task to an agent/human/host
                - unassign: Remove assignee from a task
                - list: List tasks grouped by assignee
                - workload: Get workload distribution summary
                - bulk_assign: Assign multiple tasks at once
                - auto_assign: Auto-distribute background tasks to agents
            """
            return task_assignee(
                action=action,
                task_id=task_id,
                task_ids=task_ids,
                assignee_name=assignee_name,
                assignee_type=assignee_type,
                hostname=hostname,
                status_filter=status_filter,
                include_unassigned=include_unassigned,
                max_tasks_per_agent=max_tasks_per_agent,
                priority_filter=priority_filter,
                dry_run=dry_run,
            )

        logger.info("✅ Registered 1 consolidated assignee tool (replaces 6)")

    except Exception as e:
        logger.warning(f"Could not register assignee tools: {e}")


__all__ = [
    "assign_task",
    "unassign_task",
    "list_tasks_by_assignee",
    "get_workload_summary",
    "bulk_assign_tasks",
    "auto_assign_background_tasks",
    "task_assignee",  # Consolidated tool
    "register_assignee_tools",
    "KNOWN_AGENTS",
    "KNOWN_HUMANS",
]

