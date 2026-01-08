"""
Pattern-Based Triggers Tool

Setup pattern-based automation triggers for automatic tool execution.
Supports file patterns, git events, and task status changes.
"""

import json
import os
import sys
from pathlib import Path

from ..utils import find_project_root

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

try:
    from tools.intelligent_automation_base import IntelligentAutomationBase
except ImportError:
    # Fallback if base class not available
    class IntelligentAutomationBase:
        pass


def setup_pattern_triggers(
    patterns: Optional[Dict[str, dict]] = None,
    config_path: Optional[str] = None,
    install: bool = True,
    dry_run: bool = False
) -> str:
    """
    Setup pattern-based automation triggers.

    Args:
        patterns: Dictionary of pattern configurations (optional)
        config_path: Path to pattern configuration file (optional)
        install: Whether to install triggers (default: True)
        dry_run: Preview mode without making changes (default: False)

    Returns:
        JSON string with setup results
    """
    project_root = find_project_root()

    # Default patterns if not provided
    if patterns is None:
        patterns = _get_default_patterns()

    # Load from config file if provided
    if config_path:
        config_file = Path(config_path)
        if config_file.exists():
            with open(config_file) as f:
                file_patterns = json.load(f)
                patterns.update(file_patterns)

    # Configuration file location
    config_file_path = project_root / ".cursor" / "automa_patterns.json"

    results = {
        "status": "success",
        "patterns_configured": [],
        "patterns_skipped": [],
        "config_file": str(config_file_path),
        "dry_run": dry_run
    }

    if dry_run:
        results["patterns_configured"] = [
            {
                "category": category,
                "patterns": list(pattern_config.keys()),
                "tools": _extract_tools(pattern_config)
            }
            for category, pattern_config in patterns.items()
        ]
        return json.dumps(results, indent=2)

    # Write configuration file
    try:
        config_file_path.parent.mkdir(parents=True, exist_ok=True)
        with open(config_file_path, 'w') as f:
            json.dump(patterns, f, indent=2)

        results["patterns_configured"] = [
            {
                "category": category,
                "pattern_count": len(pattern_config),
                "tools": _extract_tools(pattern_config)
            }
            for category, pattern_config in patterns.items()
        ]

        # Setup integration points
        _setup_git_hooks_integration(project_root, patterns, results)
        _setup_file_watcher_integration(project_root, patterns, results)
        _setup_task_status_integration(project_root, patterns, results)

    except Exception as e:
        results["status"] = "error"
        results["error"] = str(e)

    return json.dumps(results, indent=2)


def _get_default_patterns() -> dict[str, dict]:
    """Get default pattern configurations."""
    return {
        "file_patterns": {
            "docs/**/*.md": {
                "on_change": "check_documentation_health_tool",
                "on_create": "add_external_tool_hints_tool",
                "description": "Documentation files"
            },
            "requirements.txt|Cargo.toml|package.json|pyproject.toml": {
                "on_change": "scan_dependency_security_tool",
                "description": "Dependency files"
            },
            ".todo2/state.todo2.json": {
                "on_change": "detect_duplicate_tasks_tool",
                "description": "Todo2 state file"
            },
            "CMakeLists.txt|CMakePresets.json": {
                "on_change": "validate_ci_cd_workflow_tool",
                "description": "CMake configuration"
            }
        },
        "git_events": {
            "pre_commit": {
                "tools": [
                    "check_documentation_health_tool --quick",
                    "scan_dependency_security_tool --quick"
                ],
                "blocking": True,
                "description": "Quick checks before commit"
            },
            "pre_push": {
                "tools": [
                    "analyze_todo2_alignment_tool",
                    "scan_dependency_security_tool",
                    "check_documentation_health_tool"
                ],
                "blocking": True,
                "description": "Comprehensive checks before push"
            },
            "post_commit": {
                "tools": [
                    "find_automation_opportunities_tool --quick"
                ],
                "blocking": False,
                "description": "Non-blocking checks after commit"
            },
            "post_merge": {
                "tools": [
                    "detect_duplicate_tasks_tool",
                    "sync_todo_tasks_tool"
                ],
                "blocking": False,
                "description": "Checks after merge"
            }
        },
        "task_status_changes": {
            "Todo → In Progress": {
                "tools": ["analyze_todo2_alignment_tool"],
                "description": "Verify alignment when starting work"
            },
            "In Progress → Review": {
                "tools": [
                    "analyze_todo2_alignment_tool",
                    "detect_duplicate_tasks_tool"
                ],
                "description": "Quality checks before review"
            },
            "Review → Done": {
                "tools": ["detect_duplicate_tasks_tool"],
                "description": "Final checks on completion"
            }
        }
    }


def _extract_tools(pattern_config: dict) -> list[str]:
    """Extract tool names from pattern configuration."""
    tools = []
    for _pattern, config in pattern_config.items():
        if isinstance(config, dict):
            if "tools" in config:
                tools.extend(config["tools"])
            elif "on_change" in config:
                tools.append(config["on_change"])
            elif "on_create" in config:
                tools.append(config["on_create"])
    return list(set(tools))  # Remove duplicates


def _setup_git_hooks_integration(
    project_root: Path,
    patterns: dict[str, dict],
    results: dict
) -> None:
    """Setup git hooks integration for pattern triggers."""
    if "git_events" not in patterns:
        return

    hooks_dir = project_root / ".git" / "hooks"
    if not hooks_dir.exists():
        results["patterns_skipped"].append({
            "category": "git_events",
            "reason": ".git/hooks directory not found"
        })
        return

    # Note: Git hooks are handled by setup_git_hooks_tool
    # This just documents the integration
    results["git_hooks_integration"] = {
        "status": "configured",
        "note": "Use setup_git_hooks_tool to install actual hooks"
    }


def _setup_file_watcher_integration(
    project_root: Path,
    patterns: dict[str, dict],
    results: dict
) -> None:
    """Setup file watcher integration for pattern triggers."""
    if "file_patterns" not in patterns:
        return

    # Create file watcher script
    watcher_script = project_root / ".cursor" / "automa_file_watcher.py"

    try:
        watcher_content = _generate_file_watcher_script(patterns["file_patterns"])
        with open(watcher_script, 'w') as f:
            f.write(watcher_content)

        os.chmod(watcher_script, 0o755)

        results["file_watcher_integration"] = {
            "status": "configured",
            "script": str(watcher_script),
            "note": "Run manually or via cron: python3 .cursor/automa_file_watcher.py"
        }
    except Exception as e:
        results["patterns_skipped"].append({
            "category": "file_patterns",
            "reason": f"Failed to create watcher script: {str(e)}"
        })


def _setup_task_status_integration(
    project_root: Path,
    patterns: dict[str, dict],
    results: dict
) -> None:
    """Setup task status change integration."""
    if "task_status_changes" not in patterns:
        return

    # Note: Task status triggers are handled by Todo2 MCP server
    # This just documents the integration
    results["task_status_integration"] = {
        "status": "configured",
        "note": "Task status triggers handled by Todo2 MCP server hooks"
    }


def _generate_file_watcher_script(file_patterns: dict) -> str:
    """Generate file watcher script content."""
    return """#!/usr/bin/env python3
\"\"\"
Exarp File Watcher

Monitors file changes and triggers exarp tools based on patterns.
Run manually or via cron job.
\"\"\"

import json
import sys
from pathlib import Path
from typing import Dict, List, Optional

# Load pattern configuration
CONFIG_FILE = Path(__file__).parent.parent / ".cursor" / "automa_patterns.json"

def load_patterns() -> Dict:
    \"\"\"Load pattern configuration.\"\"\"
    if not CONFIG_FILE.exists():
        return {}

    with open(CONFIG_FILE, 'r') as f:
        config = json.load(f)
        return config.get("file_patterns", {})

def check_file_changes() -> List[Dict]:
    \"\"\"Check for file changes and return matching patterns.\"\"\"
    # This is a placeholder - implement actual file watching logic
    # For now, returns empty list
    return []

def trigger_tool(tool_name: str) -> bool:
    \"\"\"Trigger an exarp tool.\"\"\"
    # This is a placeholder - implement actual tool triggering
    # For now, just prints what would be triggered
    print(f"Would trigger: {tool_name}")
    return True

if __name__ == "__main__":
    patterns = load_patterns()
    changes = check_file_changes()

    for change in changes:
        file_path = change["file"]
        for pattern, config in patterns.items():
            # Simple pattern matching (implement proper glob/regex matching)
            if pattern in file_path or file_path.endswith(pattern.split("/")[-1]):
                if "on_change" in config:
                    trigger_tool(config["on_change"])
                if "on_create" in config and change.get("created"):
                    trigger_tool(config["on_create"])
"""


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Setup pattern-based triggers")
    parser.add_argument("--config", help="Path to pattern configuration file")
    parser.add_argument("--no-install", action="store_true", help="Don't install triggers")
    parser.add_argument("--dry-run", action="store_true", help="Preview without installing")

    args = parser.parse_args()

    result = setup_pattern_triggers(
        config_path=args.config,
        install=not args.no_install,
        dry_run=args.dry_run
    )

    print(result)
