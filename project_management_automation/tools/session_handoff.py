"""
Session Handoff Tool

Allows developers to "finish work for the day" and ensure the project can be
resumed by other developers on different machines.

Creates a handoff note that includes:
- Current work state (tasks in progress, blockers)
- What was accomplished
- What needs attention next
- Context for resumption

The handoff is stored in .todo2/ so it's visible to all machines via git sync.
"""

import asyncio
import json
import logging
import os
import socket
import subprocess
import time
from datetime import UTC, datetime
from pathlib import Path
from typing import Any, List, Optional

logger = logging.getLogger(__name__)


def _run_async_safe(coro):
    """
    Safely run an async coroutine, handling cases where an event loop may already be running.

    Based on FastMCP best practices: avoid asyncio.run() in running event loops.
    Uses asyncio.ensure_future() when loop exists, asyncio.run() when it doesn't.

    Args:
        coro: The coroutine to run

    Returns:
        The result of the coroutine
    """
    try:
        # Check if there's already a running event loop
        asyncio.get_running_loop()
        # If we get here, there's already a loop - use ensure_future to schedule
        # Then run in a separate thread to avoid blocking
        import concurrent.futures

        def run_in_thread():
            """Run the coroutine in a new event loop in a separate thread"""
            new_loop = asyncio.new_event_loop()
            asyncio.set_event_loop(new_loop)
            try:
                return new_loop.run_until_complete(coro)
            finally:
                new_loop.close()

        with concurrent.futures.ThreadPoolExecutor() as executor:
            future = executor.submit(run_in_thread)
            return future.result()
    except RuntimeError:
        # No running loop, safe to use asyncio.run()
        return asyncio.run(coro)


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


def _get_current_hostname() -> str:
    """Get current machine's hostname."""
    return socket.gethostname()


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
        with open(todo2_file, 'w') as f:
            json.dump(state, f, indent=2)
        return True
    except Exception as e:
        logger.error(f"Error saving Todo2 state: {e}")
        return False


def _get_git_status() -> dict[str, Any]:
    """Get current git status."""
    try:
        project_root = _find_project_root()

        # Check for uncommitted changes
        result = subprocess.run(
            ["git", "status", "--porcelain"],
            cwd=str(project_root),
            capture_output=True,
            text=True,
            timeout=10
        )
        changed_files = [line.strip() for line in result.stdout.strip().split('\n') if line.strip()]

        # Get current branch
        branch_result = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            cwd=str(project_root),
            capture_output=True,
            text=True,
            timeout=5
        )
        branch = branch_result.stdout.strip()

        # Check if ahead/behind remote
        status_result = subprocess.run(
            ["git", "status", "-sb"],
            cwd=str(project_root),
            capture_output=True,
            text=True,
            timeout=5
        )
        status_line = status_result.stdout.split('\n')[0] if status_result.stdout else ""
        ahead = "ahead" in status_line
        behind = "behind" in status_line

        return {
            "branch": branch,
            "uncommitted_files": len(changed_files),
            "changed_files": changed_files[:10],  # First 10
            "ahead_of_remote": ahead,
            "behind_remote": behind,
            "needs_push": ahead and not behind,
            "needs_pull": behind,
        }
    except Exception as e:
        return {"error": str(e)}


def _load_handoff_history() -> list[dict[str, Any]]:
    """Load existing handoff notes."""
    project_root = _find_project_root()
    handoff_file = project_root / ".todo2" / "handoffs.json"

    if not handoff_file.exists():
        return []

    try:
        with open(handoff_file) as f:
            data = json.load(f)
            return data.get("handoffs", [])
    except Exception:
        return []


def _save_handoff(handoff: dict[str, Any]) -> bool:
    """Save handoff note."""
    project_root = _find_project_root()
    handoff_file = project_root / ".todo2" / "handoffs.json"

    try:
        # Load existing
        history = _load_handoff_history()

        # Add new handoff
        history.append(handoff)

        # Keep last 20 handoffs
        history = history[-20:]

        # Save
        handoff_file.parent.mkdir(parents=True, exist_ok=True)
        with open(handoff_file, 'w') as f:
            json.dump({"handoffs": history}, f, indent=2)
        return True
    except Exception as e:
        logger.error(f"Error saving handoff: {e}")
        return False


def end_session(
    summary: Optional[str] = None,
    blockers: Optional[List[str]] = None,
    next_steps: Optional[List[str]] = None,
    unassign_my_tasks: bool = True,
    include_git_status: bool = True,
    dry_run: bool = False
) -> str:
    """
    [HINT: End session. Creates handoff note for other devs. Unassigns tasks, summarizes work, notes blockers.]

    End your work session and create a handoff note for other developers/machines.

    This tool:
    1. Summarizes your work (tasks completed, in progress)
    2. Records any blockers or context
    3. Optionally unassigns your tasks so others can pick them up
    4. Creates a handoff note visible to all machines
    5. Warns about uncommitted changes

    Args:
        summary: Optional summary of what you worked on
        blockers: Optional list of blockers encountered
        next_steps: Optional list of suggested next steps
        unassign_my_tasks: Unassign tasks assigned to this host (default: True)
        include_git_status: Include git status in handoff (default: True)
        dry_run: Preview without saving (default: False)

    Returns:
        JSON with handoff details and recommendations

    Example:
        end_session(
            summary="Implemented task assignee system",
            blockers=["Waiting on API key for external service"],
            next_steps=["Test on Ubuntu server", "Update documentation"]
        )
    """
    start_time = time.time()
    current_host = _get_current_hostname()
    timestamp = datetime.now(UTC).isoformat().replace("+00:00", "Z")

    try:
        # Load state
        state = _load_todo2_state()
        todos = state.get("todos", [])

        # Find my tasks (assigned to this host or human)
        my_in_progress = []
        my_completed_today = []
        my_tasks_to_unassign = []

        today_str = datetime.now().strftime("%Y-%m-%d")

        for task in todos:
            assignee = task.get("assignee", {})
            hostname = assignee.get("hostname", "")
            assignee_name = assignee.get("name", "")

            # Check if task is mine
            is_mine = (
                hostname == current_host or
                assignee_name == current_host or
                assignee_name in ["david", "human"]  # Human assignees
            )

            if is_mine:
                status = task.get("status", "")
                if status == "In Progress":
                    my_in_progress.append({
                        "id": task.get("id"),
                        "name": task.get("name", "")[:60],
                        "priority": task.get("priority"),
                    })
                    my_tasks_to_unassign.append(task)
                elif status == "Done":
                    # Check if completed today
                    last_modified = task.get("lastModified", "")
                    if last_modified.startswith(today_str):
                        my_completed_today.append({
                            "id": task.get("id"),
                            "name": task.get("name", "")[:60],
                        })

        # Get git status
        git_status = None
        if include_git_status:
            git_status = _get_git_status()

        # Build handoff note
        handoff = {
            "id": f"handoff-{int(time.time())}",
            "timestamp": timestamp,
            "host": current_host,
            "summary": summary or f"Session ended on {current_host}",
            "tasks_in_progress": my_in_progress,
            "tasks_completed_today": my_completed_today,
            "blockers": blockers or [],
            "next_steps": next_steps or [],
            "git_status": git_status,
            "tasks_unassigned": len(my_tasks_to_unassign) if unassign_my_tasks else 0,
        }

        # Build warnings
        warnings = []
        if git_status:
            if git_status.get("uncommitted_files", 0) > 0:
                warnings.append(f"⚠️ {git_status['uncommitted_files']} uncommitted file(s) - consider committing before ending session")
            if git_status.get("needs_push"):
                warnings.append("⚠️ Local commits not pushed - other machines won't see your changes")
            if git_status.get("behind_remote"):
                warnings.append("⚠️ Behind remote - pull before resuming work")

        if my_in_progress:
            if not unassign_my_tasks:
                warnings.append(f"⚠️ {len(my_in_progress)} task(s) still assigned to you - other devs can't pick them up")

        handoff["warnings"] = warnings

        # Apply changes (if not dry run)
        if not dry_run:
            # Save handoff as memory for future reference
            try:
                from .session_memory import save_session_insight
                memory_content = f"""Session handoff from {current_host}.

## Summary
{summary or 'Session ended'}

## Tasks In Progress
{chr(10).join('- ' + t['name'] for t in my_in_progress) or 'None'}

## Blockers
{chr(10).join('- ' + b for b in (blockers or [])) or 'None'}

## Next Steps
{chr(10).join('- ' + s for s in (next_steps or [])) or 'None'}
"""
                save_session_insight(
                    title=f"Handoff: {summary[:50] if summary else 'Session ended'}",
                    content=memory_content,
                    category="insight",
                    metadata={"type": "handoff", "host": current_host}
                )
            except Exception as e:
                logger.debug(f"Could not save handoff as memory: {e}")

            # Unassign my tasks
            if unassign_my_tasks:
                for task in my_tasks_to_unassign:
                    task_id = task.get("id")
                    # Find in todos and remove assignee
                    for i, t in enumerate(todos):
                        if t.get("id") == task_id:
                            # Add handoff note to task
                            if "comments" not in todos[i]:
                                todos[i]["comments"] = []
                            todos[i]["comments"].append({
                                "id": f"{task_id}-handoff-{int(time.time())}",
                                "todoId": task_id,
                                "type": "handoff",
                                "content": f"**Handoff from {current_host}**: {summary or 'Session ended'}\n\nBlockers: {', '.join(blockers) if blockers else 'None'}\n\nNext steps: {', '.join(next_steps) if next_steps else 'Continue work'}",
                                "created": timestamp,
                            })
                            # Remove assignee
                            if "assignee" in todos[i]:
                                del todos[i]["assignee"]
                            todos[i]["lastModified"] = timestamp
                            break

                state["todos"] = todos
                _save_todo2_state(state)

            # Save handoff note
            _save_handoff(handoff)

        duration = time.time() - start_time

        result = {
            "success": True,
            "dry_run": dry_run,
            "handoff_id": handoff["id"],
            "host": current_host,
            "timestamp": timestamp,
            "summary": {
                "tasks_in_progress": len(my_in_progress),
                "tasks_completed_today": len(my_completed_today),
                "tasks_unassigned": len(my_tasks_to_unassign) if unassign_my_tasks and not dry_run else 0,
                "blockers_noted": len(blockers or []),
            },
            "in_progress_tasks": my_in_progress,
            "completed_today": my_completed_today,
            "blockers": blockers or [],
            "next_steps": next_steps or [],
            "warnings": warnings,
            "git_status": git_status,
            "duration_ms": round(duration * 1000, 2),
            "message": "✅ Session ended. Handoff note created for other developers." if not dry_run else "Preview mode - no changes made",
        }

        return json.dumps(result, indent=2)

    except Exception as e:
        logger.error(f"Error ending session: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


def get_latest_handoff() -> str:
    """
    [HINT: Latest handoff. Shows the most recent handoff note from another developer.]

    Get the most recent handoff note to understand current project state.

    Use this when starting work to see:
    - What the previous developer was working on
    - Any blockers they encountered
    - Suggested next steps
    - Tasks available to pick up

    Returns:
        JSON with latest handoff details
    """
    try:
        history = _load_handoff_history()

        if not history:
            return json.dumps({
                "success": True,
                "has_handoff": False,
                "message": "No handoff notes found. You may be the first to work on this project!",
            }, indent=2)

        latest = history[-1]
        current_host = _get_current_hostname()

        # Check if it's from this host
        is_own = latest.get("host") == current_host

        return json.dumps({
            "success": True,
            "has_handoff": True,
            "is_from_current_host": is_own,
            "handoff": latest,
            "message": f"Latest handoff from {latest.get('host', 'unknown')} at {latest.get('timestamp', 'unknown')}",
            "action_items": [
                "Review blockers noted",
                "Check suggested next steps",
                "Pick up unassigned tasks if needed",
                "Pull latest changes if behind remote" if latest.get("git_status", {}).get("needs_push") else None,
            ],
        }, indent=2)

    except Exception as e:
        logger.error(f"Error getting latest handoff: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


def list_handoffs(limit: int = 5) -> str:
    """
    [HINT: List handoffs. Shows recent handoff notes from all developers.]

    List recent handoff notes to see project history.

    Args:
        limit: Maximum number of handoffs to return (default: 5)

    Returns:
        JSON with handoff history
    """
    try:
        history = _load_handoff_history()

        # Get last N
        recent = history[-limit:] if len(history) > limit else history
        recent.reverse()  # Most recent first

        # Summarize
        summaries = []
        for h in recent:
            summaries.append({
                "id": h.get("id"),
                "host": h.get("host"),
                "timestamp": h.get("timestamp"),
                "summary": h.get("summary", "")[:80],
                "tasks_in_progress": len(h.get("tasks_in_progress", [])),
                "blockers": len(h.get("blockers", [])),
            })

        return json.dumps({
            "success": True,
            "total_handoffs": len(history),
            "showing": len(summaries),
            "handoffs": summaries,
        }, indent=2)

    except Exception as e:
        logger.error(f"Error listing handoffs: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


def _try_agentic_tools_sync(direction: str = "pull") -> dict[str, Any]:
    """
    Try to sync using agentic-tools MCP.

    Args:
        direction: "pull" to fetch remote state, "push" to sync local state, "both" for pull then push

    Returns:
        Result dict with success status and details
    """
    try:
        from ..scripts.base.mcp_client import get_mcp_client
        from ..utils.todo2_utils import get_repo_project_id

        project_root = _find_project_root()
        project_id = get_repo_project_id(project_root)

        if not project_id:
            return {
                "success": False,
                "method": "agentic-tools",
                "error": "Could not determine project ID",
            }

        mcp_client = get_mcp_client(project_root)
        working_dir = str(project_root)

        # Agentic-tools MCP handles its own state file (.agentic-tools-mcp/tasks/tasks.json)
        # The MCP server automatically syncs when operations are performed
        # For pull: we can list tasks to ensure we have latest
        # For push: operations already update the state file

        if direction in ("pull", "both"):
            # List todos to ensure we have latest state
            try:
                tasks = _run_async_safe(mcp_client.list_todos(project_id, working_dir))
                return {
                    "success": True,
                    "method": "agentic-tools",
                    "direction": direction,
                    "tasks_synced": len(tasks) if isinstance(tasks, list) else 0,
                    "message": "Agentic-tools MCP state synced (uses .agentic-tools-mcp/tasks/tasks.json)",
                }
            except Exception as e:
                logger.debug(f"Agentic-tools MCP pull failed: {e}")
                return {
                    "success": False,
                    "method": "agentic-tools",
                    "error": str(e),
                }

        # For push, agentic-tools MCP operations already update state
        return {
            "success": True,
            "method": "agentic-tools",
            "direction": direction,
            "message": "Agentic-tools MCP operations automatically sync state",
        }

    except ImportError:
        return {
            "success": False,
            "method": "agentic-tools",
            "error": "MCP client not available",
        }
    except Exception as e:
        logger.debug(f"Agentic-tools sync failed: {e}")
        return {
            "success": False,
            "method": "agentic-tools",
            "error": str(e),
        }


def _git_auto_sync(direction: str = "pull", auto_commit: bool = True) -> dict[str, Any]:
    """
    Sync Todo2 state using git with automatic commit/push.

    Args:
        direction: "pull" to fetch remote, "push" to commit and push local, "both" for pull then push
        auto_commit: Whether to auto-commit state changes (default: True)

    Returns:
        Result dict with success status and details
    """
    project_root = _find_project_root()
    project_root / ".todo2" / "state.todo2.json"
    project_root / ".todo2" / "handoffs.json"

    results = {
        "success": True,
        "method": "git-auto",
        "direction": direction,
        "operations": [],
        "errors": [],
    }

    try:
        # Check if we're in a git repo
        git_check = subprocess.run(
            ["git", "rev-parse", "--git-dir"],
            cwd=str(project_root),
            capture_output=True,
            text=True,
            timeout=5
        )
        if git_check.returncode != 0:
            return {
                "success": False,
                "method": "git-auto",
                "error": "Not in a git repository",
            }

        # Pull remote changes first if requested
        if direction in ("pull", "both"):
            try:
                # Fetch latest
                fetch_result = subprocess.run(
                    ["git", "fetch", "origin"],
                    cwd=str(project_root),
                    capture_output=True,
                    text=True,
                    timeout=30
                )

                if fetch_result.returncode == 0:
                    results["operations"].append("fetched")

                    # Check if there are remote changes to merge
                    status_result = subprocess.run(
                        ["git", "status", "-sb"],
                        cwd=str(project_root),
                        capture_output=True,
                        text=True,
                        timeout=5
                    )

                    if "behind" in status_result.stdout:
                        # Get current branch name
                        branch_result = subprocess.run(
                            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
                            cwd=str(project_root),
                            capture_output=True,
                            text=True,
                            timeout=5
                        )
                        current_branch = branch_result.stdout.strip()

                        # Merge remote changes
                        merge_result = subprocess.run(
                            ["git", "merge", f"origin/{current_branch}", "--no-edit"],
                            cwd=str(project_root),
                            capture_output=True,
                            text=True,
                            timeout=30
                        )

                        if merge_result.returncode == 0:
                            results["operations"].append("merged")
                        else:
                            # Check for conflicts
                            if "CONFLICT" in merge_result.stdout or merge_result.returncode != 0:
                                results["errors"].append("Merge conflict detected - manual resolution required")
                                results["success"] = False
                            else:
                                results["errors"].append(f"Merge failed: {merge_result.stderr}")
                                results["success"] = False
                    else:
                        results["operations"].append("already_up_to_date")
                else:
                    results["errors"].append(f"Fetch failed: {fetch_result.stderr}")
                    results["success"] = False

            except subprocess.TimeoutExpired:
                results["errors"].append("Git pull operation timed out")
                results["success"] = False
            except Exception as e:
                results["errors"].append(f"Pull error: {str(e)}")
                results["success"] = False

        # Push local changes if requested
        if direction in ("push", "both") and results["success"]:
            try:
                # Check if there are uncommitted changes to state files
                status_result = subprocess.run(
                    ["git", "status", "--porcelain", ".todo2/"],
                    cwd=str(project_root),
                    capture_output=True,
                    text=True,
                    timeout=5
                )

                has_changes = bool(status_result.stdout.strip())

                if has_changes and auto_commit:
                    # Stage Todo2 files
                    add_result = subprocess.run(
                        ["git", "add", ".todo2/state.todo2.json", ".todo2/handoffs.json"],
                        cwd=str(project_root),
                        capture_output=True,
                        text=True,
                        timeout=10
                    )

                    if add_result.returncode == 0:
                        results["operations"].append("staged")

                        # Auto-commit with descriptive message
                        current_host = _get_current_hostname()
                        commit_msg = f"Auto-sync Todo2 state from {current_host} [{datetime.now(UTC).isoformat().replace('+00:00', 'Z')}]"

                        commit_result = subprocess.run(
                            ["git", "commit", "-m", commit_msg],
                            cwd=str(project_root),
                            capture_output=True,
                            text=True,
                            timeout=10
                        )

                        if commit_result.returncode == 0:
                            results["operations"].append("committed")
                        else:
                            # Check if there's nothing to commit
                            if "nothing to commit" in commit_result.stdout:
                                results["operations"].append("nothing_to_commit")
                            else:
                                results["errors"].append(f"Commit failed: {commit_result.stderr}")
                                results["success"] = False
                                return results

                # Push if we have commits to push
                push_check = subprocess.run(
                    ["git", "status", "-sb"],
                    cwd=str(project_root),
                    capture_output=True,
                    text=True,
                    timeout=5
                )

                if "ahead" in push_check.stdout:
                    push_result = subprocess.run(
                        ["git", "push", "origin", "HEAD"],
                        cwd=str(project_root),
                        capture_output=True,
                        text=True,
                        timeout=30
                    )

                    if push_result.returncode == 0:
                        results["operations"].append("pushed")
                    else:
                        results["errors"].append(f"Push failed: {push_result.stderr}")
                        results["success"] = False
                else:
                    results["operations"].append("nothing_to_push")

            except subprocess.TimeoutExpired:
                results["errors"].append("Git push operation timed out")
                results["success"] = False
            except Exception as e:
                results["errors"].append(f"Push error: {str(e)}")
                results["success"] = False

        return results

    except Exception as e:
        return {
            "success": False,
            "method": "git-auto",
            "error": str(e),
        }


def sync_todo2_state(
    direction: str = "both",
    prefer_agentic_tools: bool = True,
    auto_commit: bool = True,
    dry_run: bool = False
) -> str:
    """
    [HINT: Sync Todo2 state. Syncs task state across agents without manual commits. Directions: pull, push, both.]

    Sync Todo2 state across multiple agents/machines without requiring manual git commits.

    This tool provides multiple sync methods:
    1. **Agentic-Tools MCP** (preferred): Uses MCP server which handles its own state sync
    2. **Git Auto-Sync** (fallback): Automatically commits and pushes state changes

    Directions:
    - pull: Fetch and merge remote state changes
    - push: Auto-commit and push local state changes
    - both: Pull remote changes, then push local changes

    Args:
        direction: Sync direction - "pull", "push", or "both" (default: "both")
        prefer_agentic_tools: Try agentic-tools MCP first (default: True)
        auto_commit: Auto-commit state changes for git sync (default: True)
        dry_run: Preview sync operations without making changes (default: False)

    Returns:
        JSON with sync results and details

    Examples:
        sync_todo2_state(direction="pull")  # Pull latest state from remote
        sync_todo2_state(direction="push")  # Push local state changes
        sync_todo2_state(direction="both")  # Pull then push
    """
    start_time = time.time()
    current_host = _get_current_hostname()

    if dry_run:
        return json.dumps({
            "success": True,
            "dry_run": True,
            "direction": direction,
            "message": "Dry run - no changes made",
            "methods_available": {
                "agentic-tools": "Would try agentic-tools MCP first" if prefer_agentic_tools else "Skipped",
                "git-auto": "Would use git auto-commit/push as fallback",
            },
        }, indent=2)

    results = {
        "success": False,
        "host": current_host,
        "timestamp": datetime.now(UTC).isoformat().replace("+00:00", "Z"),
        "direction": direction,
        "methods_tried": [],
        "final_method": None,
    }

    # Try agentic-tools MCP first if preferred
    if prefer_agentic_tools:
        agentic_result = _try_agentic_tools_sync(direction)
        results["methods_tried"].append(agentic_result)

        if agentic_result.get("success"):
            results["success"] = True
            results["final_method"] = "agentic-tools"
            results["message"] = agentic_result.get("message", "Synced via agentic-tools MCP")
            results["details"] = agentic_result
            duration = time.time() - start_time
            results["duration_ms"] = round(duration * 1000, 2)
            return json.dumps(results, indent=2)

    # Fallback to git auto-sync
    git_result = _git_auto_sync(direction, auto_commit)
    results["methods_tried"].append(git_result)

    if git_result.get("success"):
        results["success"] = True
        results["final_method"] = "git-auto"
        results["message"] = f"Synced via git auto-commit/push: {', '.join(git_result.get('operations', []))}"
        results["details"] = git_result
    else:
        results["success"] = False
        results["final_method"] = "git-auto"
        results["message"] = f"Sync failed: {', '.join(git_result.get('errors', []))}"
        results["errors"] = git_result.get("errors", [])

    duration = time.time() - start_time
    results["duration_ms"] = round(duration * 1000, 2)

    return json.dumps(results, indent=2)


def resume_session() -> str:
    """
    [HINT: Resume session. Primes context from latest handoff and shows available tasks.]

    Start a new work session by reviewing the latest handoff and available tasks.

    This tool:
    1. Shows the latest handoff note
    2. Lists unassigned tasks available to pick up
    3. Warns about any blockers noted
    4. Suggests tasks based on priority

    Returns:
        JSON with session resumption context
    """
    try:
        current_host = _get_current_hostname()

        # Get latest handoff
        history = _load_handoff_history()
        latest_handoff = history[-1] if history else None

        # Get unassigned tasks
        state = _load_todo2_state()
        todos = state.get("todos", [])

        unassigned_todo = []
        unassigned_in_progress = []

        for task in todos:
            if not task.get("assignee"):
                status = task.get("status", "")
                task_info = {
                    "id": task.get("id"),
                    "name": task.get("name", "")[:60],
                    "priority": task.get("priority"),
                    "tags": task.get("tags", [])[:3],
                }
                if status == "Todo":
                    unassigned_todo.append(task_info)
                elif status == "In Progress":
                    # Orphaned in-progress task
                    unassigned_in_progress.append(task_info)

        # Sort by priority
        priority_order = {"high": 0, "medium": 1, "low": 2}
        unassigned_todo.sort(key=lambda t: priority_order.get(t.get("priority", "medium"), 1))

        # Build context
        context = {
            "success": True,
            "current_host": current_host,
            "timestamp": datetime.now(UTC).isoformat().replace("+00:00", "Z"),
        }

        # Add handoff info
        if latest_handoff:
            context["previous_handoff"] = {
                "from_host": latest_handoff.get("host"),
                "timestamp": latest_handoff.get("timestamp"),
                "summary": latest_handoff.get("summary"),
                "blockers": latest_handoff.get("blockers", []),
                "next_steps": latest_handoff.get("next_steps", []),
            }

        # Add available work
        context["available_work"] = {
            "unassigned_todo": unassigned_todo[:10],
            "orphaned_in_progress": unassigned_in_progress,
            "total_unassigned": len(unassigned_todo),
        }

        # Recommendations
        recommendations = []
        if latest_handoff and latest_handoff.get("blockers"):
            recommendations.append(f"Review blockers from {latest_handoff.get('host')}: {', '.join(latest_handoff.get('blockers', []))}")
        if unassigned_in_progress:
            recommendations.append(f"Pick up {len(unassigned_in_progress)} orphaned in-progress task(s)")
        if unassigned_todo:
            high_priority = [t for t in unassigned_todo if t.get("priority") == "high"]
            if high_priority:
                recommendations.append(f"Consider {len(high_priority)} high-priority unassigned task(s)")

        context["recommendations"] = recommendations

        # Check for handoff-related memories
        try:
            from .session_memory import search_session_memories
            memory_result = search_session_memories("handoff", limit=3)
            if memory_result.get("success") and memory_result.get("total_results", 0) > 0:
                context["related_memories"] = [
                    {"title": m.get("title"), "created_at": m.get("created_at")}
                    for m in memory_result.get("memories", [])[:3]
                ]
        except Exception as e:
            logger.debug(f"Could not search handoff memories: {e}")

        context["message"] = "Session resumed. Review handoff and pick up tasks to continue."

        return json.dumps(context, indent=2)

    except Exception as e:
        logger.error(f"Error resuming session: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e),
        }, indent=2)


# Consolidated tool for MCP registration
def session_handoff(
    action: str,
    summary: Optional[str] = None,
    blockers: Optional[List[str]] = None,
    next_steps: Optional[List[str]] = None,
    unassign_my_tasks: bool = True,
    include_git_status: bool = True,
    limit: int = 5,
    dry_run: bool = False,
    # Sync-specific parameters
    direction: str = "both",
    prefer_agentic_tools: bool = True,
    auto_commit: bool = True,
) -> str:
    """
    [HINT: Session handoff. End/resume sessions for multi-dev coordination. Actions: end, resume, latest, list, sync.]

    Manage work sessions for multi-developer coordination across machines.

    Actions:
        - end: End your session and create handoff note
        - resume: Start session by reviewing latest handoff
        - latest: Get most recent handoff note
        - list: List recent handoff notes
        - sync: Sync Todo2 state across agents without manual commits (pull/push/both)

    Args:
        action: One of "end", "resume", "latest", "list", "sync"
        summary: Summary of work (for end action)
        blockers: List of blockers (for end action)
        next_steps: Suggested next steps (for end action)
        unassign_my_tasks: Unassign tasks on end (default: True)
        include_git_status: Include git status (default: True)
        limit: Max handoffs for list action (default: 5)
        dry_run: Preview without changes (default: False)
        direction: Sync direction for sync action - "pull", "push", or "both" (default: "both")
        prefer_agentic_tools: Try agentic-tools MCP first for sync (default: True)
        auto_commit: Auto-commit state changes for git sync (default: True)

    Returns:
        JSON with action results

    Examples:
        session_handoff(action="end", summary="Finished auth module", next_steps=["Add tests"])
        session_handoff(action="resume")
        session_handoff(action="sync", direction="pull")  # Pull latest state
        session_handoff(action="sync", direction="both")  # Pull then push
    """
    action = action.lower()

    if action == "end":
        return end_session(
            summary=summary,
            blockers=blockers,
            next_steps=next_steps,
            unassign_my_tasks=unassign_my_tasks,
            include_git_status=include_git_status,
            dry_run=dry_run,
        )

    elif action == "resume":
        return resume_session()

    elif action == "latest":
        return get_latest_handoff()

    elif action == "list":
        return list_handoffs(limit=limit)

    elif action == "sync":
        return sync_todo2_state(
            direction=direction,
            prefer_agentic_tools=prefer_agentic_tools,
            auto_commit=auto_commit,
            dry_run=dry_run,
        )

    else:
        return json.dumps({
            "error": f"Unknown action: {action}",
            "valid_actions": ["end", "resume", "latest", "list", "sync"],
        }, indent=2)


def register_handoff_tools(mcp) -> None:
    """Register session handoff tool with MCP server."""
    try:
        @mcp.tool()
        def exarp_session_handoff(
            action: str,
            summary: Optional[str] = None,
            blockers: Optional[List[str]] = None,
            next_steps: Optional[List[str]] = None,
            unassign_my_tasks: bool = True,
            include_git_status: bool = True,
            limit: int = 5,
            dry_run: bool = False,
            direction: str = "both",
            prefer_agentic_tools: bool = True,
            auto_commit: bool = True,
        ) -> str:
            """
            [HINT: Exarp session handoff. End/resume sessions for multi-dev coordination with git sync. Actions: end, resume, latest, list, sync.]

            Enhanced session handoff tool that wraps agentic-tools with additional features:
            - Git sync integration (prefers agentic-tools MCP, falls back to git)
            - Multi-device coordination
            - Auto-commit functionality
            - Todo2 state synchronization

            This is an enhanced wrapper around agentic-tools' session_handoff_tool with exarp-specific features.
            """
            return session_handoff(
                action=action,
                summary=summary,
                blockers=blockers,
                next_steps=next_steps,
                unassign_my_tasks=unassign_my_tasks,
                include_git_status=include_git_status,
                limit=limit,
                dry_run=dry_run,
                direction=direction,
                prefer_agentic_tools=prefer_agentic_tools,
                auto_commit=auto_commit,
            )

        logger.info("✅ Registered exarp_session_handoff tool (enhanced wrapper around agentic-tools)")

    except Exception as e:
        logger.warning(f"Could not register handoff tools: {e}")


__all__ = [
    "end_session",
    "get_latest_handoff",
    "list_handoffs",
    "resume_session",
    "session_handoff",
    "sync_todo2_state",
    "register_handoff_tools",
]

