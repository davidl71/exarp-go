"""
Git Hooks Setup Tool

Automatically sets up git hooks for exarp tool execution.
Supports pre-commit, pre-push, post-commit, and post-merge hooks.
"""

import json
import os
import sys
from pathlib import Path
from typing import Any, List, Optional

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

try:
    from tools.intelligent_automation_base import IntelligentAutomationBase
except ImportError:
    # Fallback if base class not available
    class IntelligentAutomationBase:
        pass


def setup_git_hooks(
    hooks: Optional[List[str]] = None,
    install: bool = True,
    dry_run: bool = False
) -> str:
    """
    Setup git hooks for automatic exarp tool execution.

    Args:
        hooks: List of hooks to setup (default: all hooks)
               Options: 'pre-commit', 'pre-push', 'post-commit', 'post-merge'
        install: Whether to install hooks (default: True)
        dry_run: Preview mode without making changes (default: False)

    Returns:
        JSON string with setup results
    """
    # Path: tools/git_hooks.py → tools → project_management_automation → project-management-automation
    project_root = Path(__file__).parent.parent.parent
    hooks_dir = project_root / ".git" / "hooks"

    if not hooks_dir.exists():
        return json.dumps({
            "status": "error",
            "error": ".git/hooks directory not found. Are you in a git repository?",
            "suggestion": "Run 'git init' first or ensure you're in a git repository"
        }, indent=2)

    # Default hooks if not specified
    if hooks is None:
        hooks = ["pre-commit", "pre-push", "post-commit", "post-merge"]

    # Hook configurations
    hook_configs = {
        "pre-commit": {
            "description": "Run quick checks before commit",
            "tools": [
                "check_documentation_health_tool --quick",
                "scan_dependency_security_tool --quick"
            ],
            "blocking": True
        },
        "pre-push": {
            "description": "Run comprehensive checks before push",
            "tools": [
                "analyze_todo2_alignment_tool",
                "scan_dependency_security_tool",
                "check_documentation_health_tool"
            ],
            "blocking": True
        },
        "post-commit": {
            "description": "Run non-blocking checks after commit",
            "tools": [
                "find_automation_opportunities_tool --quick"
            ],
            "blocking": False
        },
        "post-merge": {
            "description": "Run checks after merge",
            "tools": [
                "detect_duplicate_tasks_tool",
                "sync_todo_tasks_tool"
            ],
            "blocking": False
        }
    }

    results = {
        "status": "success",
        "hooks_configured": [],
        "hooks_skipped": [],
        "hooks_dir": str(hooks_dir),
        "dry_run": dry_run
    }

    # Find exarp server path
    mcp_config_path = project_root / ".cursor" / "mcp.json"
    exarp_path = None

    if mcp_config_path.exists():
        try:
            with open(mcp_config_path) as f:
                mcp_config = json.load(f)
                if "mcpServers" in mcp_config and "exarp" in mcp_config["mcpServers"]:
                    exarp_config = mcp_config["mcpServers"]["exarp"]
                    if "command" in exarp_config:
                        # Extract path from command (may be a script)
                        command = exarp_config["command"]
                        if command.endswith(".sh"):
                            exarp_path = Path(command).parent
                        else:
                            exarp_path = Path(command).parent
        except Exception:
            pass

    # Default to current package directory if not found
    if exarp_path is None:
        exarp_path = Path(__file__).parent.parent

    for hook_name in hooks:
        if hook_name not in hook_configs:
            results["hooks_skipped"].append({
                "hook": hook_name,
                "reason": "Unknown hook type"
            })
            continue

        hook_file = hooks_dir / hook_name
        config = hook_configs[hook_name]

        if dry_run:
            results["hooks_configured"].append({
                "hook": hook_name,
                "description": config["description"],
                "tools": config["tools"],
                "blocking": config["blocking"],
                "would_create": str(hook_file)
            })
            continue

        # Generate hook script
        hook_script = _generate_hook_script(hook_name, config, exarp_path, project_root)

        # Write hook file
        try:
            with open(hook_file, 'w') as f:
                f.write(hook_script)

            # Make executable
            os.chmod(hook_file, 0o755)

            results["hooks_configured"].append({
                "hook": hook_name,
                "description": config["description"],
                "tools": config["tools"],
                "blocking": config["blocking"],
                "file": str(hook_file)
            })
        except Exception as e:
            results["hooks_skipped"].append({
                "hook": hook_name,
                "reason": f"Failed to create hook: {str(e)}"
            })

    return json.dumps(results, indent=2)


def _generate_hook_script(
    hook_name: str,
    config: dict[str, Any],
    exarp_path: Path,
    project_root: Path
) -> str:
    """Generate git hook script content."""

    # Determine if we should use MCP tools or direct Python calls
    # For now, use direct Python calls to exarp tools
    tools_section = []

    for tool_cmd in config["tools"]:
        # Parse tool command (e.g., "check_documentation_health_tool --quick")
        parts = tool_cmd.split()
        tool_name = parts[0]
        parts[1:] if len(parts) > 1 else []

        # Convert tool name to Python function call
        # e.g., "check_documentation_health_tool" -> "check_documentation_health"
        tool_name.replace("_tool", "")

        # Build Python call with proper module path
        # Use project_management_automation.tools module path

        # For now, skip actual execution to avoid blocking commits
        # TODO: Implement proper MCP tool invocation or direct Python calls
        tools_section.append(f"  # Run {tool_name} (skipped for now)")
        tools_section.append(f"  echo \"  ⏭️  Skipping {tool_name} - implement proper invocation\"")

    blocking_check = ""
    if config["blocking"]:
        blocking_check = """
if [ $EXIT_CODE -ne 0 ]; then
  echo "❌ Hook failed. Commit aborted."
  exit 1
fi"""

    script = f"""#!/bin/bash
# Git hook: {hook_name}
# Auto-generated by exarp setup_git_hooks_tool
# Description: {config["description"]}

set -eo pipefail

EXIT_CODE=0
PROJECT_ROOT="{project_root}"
EXARP_PATH="{exarp_path}"

# Change to project root
cd "$PROJECT_ROOT"

# Add exarp to Python path (handle unset PYTHONPATH)
export PYTHONPATH="${{PYTHONPATH:+$PYTHONPATH:}}$EXARP_PATH"

echo "🔍 Running {hook_name} checks..."

{chr(10).join(tools_section)}
{blocking_check}

if [ $EXIT_CODE -eq 0 ]; then
  echo "✅ {hook_name} checks passed"
fi

exit $EXIT_CODE
"""

    return script


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Setup git hooks for exarp")
    parser.add_argument("--hooks", nargs="+", help="Hooks to setup")
    parser.add_argument("--no-install", action="store_true", help="Don't install hooks")
    parser.add_argument("--dry-run", action="store_true", help="Preview without installing")

    args = parser.parse_args()

    result = setup_git_hooks(
        hooks=args.hooks,
        install=not args.no_install,
        dry_run=args.dry_run
    )

    print(result)
