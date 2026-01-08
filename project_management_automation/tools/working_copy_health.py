"""
Working Copy Health Check Tool

MCP Tool for checking git working copy status across all agents and runners.
"""

import json
import os
import socket
import subprocess
from pathlib import Path
from typing import Any, Optional


def _ssh_command(host: str, command: str, timeout: int = 10) -> subprocess.CompletedProcess:
    """
    Execute SSH command with optimized options.

    Args:
        host: SSH host (user@host or just host)
        command: Command to execute on remote host
        timeout: Command timeout in seconds

    Returns:
        CompletedProcess with stdout, stderr, returncode
    """
    ssh_opts = [
        "-o", "ConnectTimeout=5",
        "-o", "StrictHostKeyChecking=accept-new",
        "-o", "IdentitiesOnly=yes",
        "-o", "PreferredAuthentications=publickey",
        "-o", "PasswordAuthentication=no",
        "-o", "BatchMode=yes"
    ]

    return subprocess.run(
        ["ssh"] + ssh_opts + [host, command],
        capture_output=True,
        text=True,
        timeout=timeout
    )


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


def check_working_copy_health(
    agent_name: Optional[str] = None,
    check_remote: bool = True
) -> dict[str, Any]:
    """
    Check working copy health across agents.

    Args:
        agent_name: Specific agent to check (optional, checks all if None)
        check_remote: Whether to check remote agents (default: True)

    Returns:
        Dictionary with working copy status for each agent
    """
    project_root = _find_project_root(Path(__file__))

    # Agent configurations
    agents = {
        "local": {
            "path": str(project_root),
            "type": "local"
        }
    }

    if check_remote:
        # Load remote agents from environment or config
        # Format: EXARP_REMOTE_AGENTS='{"ubuntu": {"host": "user@host", "path": "~/project"}}'
        import os
        remote_agents_json = os.environ.get("EXARP_REMOTE_AGENTS", "{}")
        try:
            remote_agents = json.loads(remote_agents_json)
            for _name, config in remote_agents.items():
                config["type"] = "remote"
            agents.update(remote_agents)
        except json.JSONDecodeError:
            pass  # No remote agents configured

    results = {}

    # Filter to specific agent if requested
    if agent_name and agent_name in agents:
        agents = {agent_name: agents[agent_name]}

    for agent_name, agent_config in agents.items():
        agent_type = agent_config.get("type", "local")
        host = agent_config.get("host", "")

        # Auto-detect if agent is local (even if marked as remote)
        if agent_type == "remote" and host:
            if _is_local_host(host):
                # This is actually the local machine, treat as local
                agent_type = "local"
                agent_config["type"] = "local"
                # Use current project root for local agents if path not set
                if "path" not in agent_config or not agent_config["path"]:
                    agent_config["path"] = str(project_root)

        if agent_type == "local":
            # Check local working copy
            # Use project_root if path not specified
            agent_path = agent_config.get("path", str(project_root))
            # Expand ~ in path
            if agent_path.startswith("~"):
                agent_path = os.path.expanduser(agent_path)

            try:
                result = subprocess.run(
                    ["git", "status", "--porcelain"],
                    cwd=agent_path,
                    capture_output=True,
                    text=True,
                    timeout=5
                )

                has_changes = bool(result.stdout.strip())

                # Get branch and commit info
                branch_result = subprocess.run(
                    ["git", "branch", "--show-current"],
                    cwd=agent_path,
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                branch = branch_result.stdout.strip() or "unknown"

                # Get latest commit
                commit_result = subprocess.run(
                    ["git", "log", "-1", "--oneline"],
                    cwd=agent_path,
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                latest_commit = commit_result.stdout.strip() or "unknown"

                # Check sync status
                subprocess.run(
                    ["git", "fetch", "--quiet"],
                    cwd=agent_path,
                    capture_output=True,
                    timeout=10
                )

                behind_result = subprocess.run(
                    ["git", "rev-list", "--count", "HEAD..origin/main"],
                    cwd=agent_path,
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                behind = int(behind_result.stdout.strip() or "0")

                ahead_result = subprocess.run(
                    ["git", "rev-list", "--count", "origin/main..HEAD"],
                    cwd=agent_path,
                    capture_output=True,
                    text=True,
                    timeout=5
                )
                ahead = int(ahead_result.stdout.strip() or "0")

                results[agent_name] = {
                    "status": "ok" if not has_changes and behind == 0 and ahead == 0 else "warning",
                    "has_uncommitted_changes": has_changes,
                    "uncommitted_files": result.stdout.strip().split('\n') if has_changes else [],
                    "branch": branch,
                    "latest_commit": latest_commit,
                    "behind_remote": behind,
                    "ahead_remote": ahead,
                    "in_sync": behind == 0 and ahead == 0,
                    "type": "local"
                }

            except Exception as e:
                results[agent_name] = {
                    "status": "error",
                    "error": str(e),
                    "type": "local"
                }

        else:
            # Check remote agent
            host = agent_config["host"]
            path = agent_config["path"]

            try:
                # Check SSH connectivity with better options
                ssh_test = subprocess.run(
                    [
                        "ssh",
                        "-o", "ConnectTimeout=5",
                        "-o", "BatchMode=yes",
                        "-o", "StrictHostKeyChecking=accept-new",
                        "-o", "IdentitiesOnly=yes",
                        "-o", "PreferredAuthentications=publickey",
                        "-o", "PasswordAuthentication=no",
                        host, "exit"
                    ],
                    capture_output=True,
                    timeout=10
                )

                if ssh_test.returncode != 0:
                    error_msg = ssh_test.stderr.decode('utf-8', errors='ignore').strip()
                    if not error_msg:
                        error_msg = f"Cannot connect to {host}"
                    results[agent_name] = {
                        "status": "error",
                        "error": error_msg,
                        "type": "remote"
                    }
                    continue

                # Get git status
                status_cmd = f"cd {path} && git status --porcelain 2>/dev/null || echo ''"
                status_result = _ssh_command(host, status_cmd)

                has_changes = bool(status_result.stdout.strip())
                uncommitted_files = status_result.stdout.strip().split('\n') if has_changes else []

                # Get branch
                branch_cmd = f"cd {path} && git branch --show-current 2>/dev/null || echo 'unknown'"
                branch_result = _ssh_command(host, branch_cmd)
                branch = branch_result.stdout.strip() or "unknown"

                # Get latest commit
                commit_cmd = f"cd {path} && git log -1 --oneline 2>/dev/null || echo 'unknown'"
                commit_result = _ssh_command(host, commit_cmd)
                latest_commit = commit_result.stdout.strip() or "unknown"

                # Check sync status
                fetch_cmd = f"cd {path} && git fetch --quiet 2>/dev/null || true"
                _ssh_command(host, fetch_cmd, timeout=15)

                behind_cmd = f"cd {path} && git rev-list --count HEAD..origin/main 2>/dev/null || echo '0'"
                behind_result = _ssh_command(host, behind_cmd)
                behind = int(behind_result.stdout.strip() or "0")

                ahead_cmd = f"cd {path} && git rev-list --count origin/main..HEAD 2>/dev/null || echo '0'"
                ahead_result = _ssh_command(host, ahead_cmd)
                ahead = int(ahead_result.stdout.strip() or "0")

                results[agent_name] = {
                    "status": "ok" if not has_changes and behind == 0 and ahead == 0 else "warning",
                    "has_uncommitted_changes": has_changes,
                    "uncommitted_files": uncommitted_files,
                    "branch": branch,
                    "latest_commit": latest_commit,
                    "behind_remote": behind,
                    "ahead_remote": ahead,
                    "in_sync": behind == 0 and ahead == 0,
                    "type": "remote",
                    "host": host
                }

            except Exception as e:
                results[agent_name] = {
                    "status": "error",
                    "error": str(e),
                    "type": "remote",
                    "host": host
                }

    # Calculate summary
    total_agents = len(results)
    ok_agents = sum(1 for r in results.values() if r.get("status") == "ok")
    warning_agents = sum(1 for r in results.values() if r.get("status") == "warning")
    error_agents = sum(1 for r in results.values() if r.get("status") == "error")

    agents_with_changes = sum(1 for r in results.values() if r.get("has_uncommitted_changes", False))
    agents_behind = sum(1 for r in results.values() if r.get("behind_remote", 0) > 0)

    return {
        "timestamp": __import__("datetime").datetime.utcnow().isoformat() + "Z",
        "summary": {
            "total_agents": total_agents,
            "ok_agents": ok_agents,
            "warning_agents": warning_agents,
            "error_agents": error_agents,
            "agents_with_uncommitted_changes": agents_with_changes,
            "agents_behind_remote": agents_behind
        },
        "agents": results,
        "recommendations": _generate_recommendations(results)
    }


def _generate_recommendations(results: dict[str, Any]) -> list[str]:
    """Generate recommendations based on working copy status."""
    recommendations = []

    for agent_name, agent_data in results.items():
        if agent_data.get("status") == "error":
            recommendations.append(f"{agent_name}: Fix connection/access issues")
            continue

        if agent_data.get("has_uncommitted_changes", False):
            file_count = len(agent_data.get("uncommitted_files", []))
            recommendations.append(f"{agent_name}: Commit {file_count} uncommitted file(s)")

        if agent_data.get("behind_remote", 0) > 0:
            behind = agent_data.get("behind_remote", 0)
            recommendations.append(f"{agent_name}: Pull {behind} commit(s) from remote")

        if agent_data.get("ahead_remote", 0) > 0:
            ahead = agent_data.get("ahead_remote", 0)
            recommendations.append(f"{agent_name}: Push {ahead} commit(s) to remote")

    if not recommendations:
        recommendations.append("All agents have clean working copies and are in sync")

    return recommendations
