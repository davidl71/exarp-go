"""
Context Primer for Workflow Modes

Provides context priming functionality for different workflow modes.
Supplies hints, tasks, prompts, and goals based on the current workflow mode.
"""

import json
import logging
from typing import Any, Dict, Optional

logger = logging.getLogger("exarp.resources.context_primer")

# Workflow mode context definitions
WORKFLOW_MODE_CONTEXT: Dict[str, Dict[str, str]] = {
    "daily_checkin": {
        "description": "Daily check-in mode: Overview + health checks. Focus on project status, health scores, and quick updates.",
        "goals": [
            "Get project health overview",
            "Check recent task status",
            "Identify blockers or urgent issues",
            "Review daily metrics and status",
        ],
    },
    "security_review": {
        "description": "Security review mode: Security-focused tools. Scan dependencies, check vulnerabilities, review security posture.",
        "goals": [
            "Scan dependencies for vulnerabilities",
            "Review security alerts",
            "Assess security posture",
            "Address critical security issues",
        ],
    },
    "task_management": {
        "description": "Task management mode: Task tools only. Focus on task organization, alignment, and workflow.",
        "goals": [
            "Manage task backlog",
            "Review task alignment",
            "Resolve task issues",
            "Optimize task workflow",
        ],
    },
    "sprint_planning": {
        "description": "Sprint planning mode: Tasks + automation + PRD. Plan sprints, analyze requirements, set up automation.",
        "goals": [
            "Plan sprint tasks",
            "Analyze PRD alignment",
            "Set up automation",
            "Prepare sprint backlog",
        ],
    },
    "code_review": {
        "description": "Code review mode: Testing + linting. Focus on code quality, tests, and quality gates.",
        "goals": [
            "Run tests and check coverage",
            "Review code quality",
            "Validate definition of done",
            "Ensure quality standards",
        ],
    },
    "development": {
        "description": "Development mode: Balanced set. General development workflow with core tools available.",
        "goals": [
            "Implement features",
            "Write quality code",
            "Manage tasks and context",
            "Maintain project health",
        ],
    },
    "debugging": {
        "description": "Debugging mode: Memory + testing. Focus on problem-solving, memory recall, and investigation.",
        "goals": [
            "Investigate issues",
            "Recall relevant context",
            "Debug problems",
            "Find root causes",
        ],
    },
    "all": {
        "description": "All tools mode: Full tool access. Complete access to all available tools and features.",
        "goals": [
            "Access all tools",
            "Complete flexibility",
            "Full context access",
        ],
    },
}


def get_hints_for_mode(mode: str) -> Dict[str, str]:
    """
    Get tool hints for a specific workflow mode.
    
    Args:
        mode: Workflow mode name
        
    Returns:
        Dict mapping tool names to their hints
    """
    try:
        from ..tools.hint_catalog import TOOL_CATALOG
        from ..tools.dynamic_tools import (
            WorkflowMode,
            WORKFLOW_TOOL_GROUPS,
            TOOL_GROUP_MAPPING,
        )
        
        # Get tool groups for this mode
        try:
            workflow_mode = WorkflowMode(mode)
            tool_groups = WORKFLOW_TOOL_GROUPS.get(workflow_mode, set())
        except ValueError:
            # Invalid mode, default to development
            workflow_mode = WorkflowMode.DEVELOPMENT
            tool_groups = WORKFLOW_TOOL_GROUPS.get(workflow_mode, set())
        
        # Collect tools in these groups
        relevant_tools = []
        for tool_name, tool_group in TOOL_GROUP_MAPPING.items():
            if tool_group in tool_groups:
                relevant_tools.append(tool_name)
        
        # Build hints dict from catalog
        hints = {}
        for tool_name in relevant_tools:
            if tool_name in TOOL_CATALOG:
                tool_info = TOOL_CATALOG[tool_name]
                hints[tool_name] = tool_info.get("hint", "")
        
        return hints
        
    except Exception as e:
        logger.warning(f"Failed to get hints for mode {mode}: {e}")
        return {}


def get_tasks_summary() -> Dict[str, Any]:
    """
    Get a summary of recent tasks.
    
    Returns:
        Dict with task summary information
    """
    try:
        # Try to use Todo2 MCP client if available
        from ..utils.todo2_mcp_client import Todo2Client
        
        client = Todo2Client()
        todos = client.list_todos(limit=10, include_completed=False)
        
        return {
            "recent": todos.get("todos", [])[:5],
            "count": len(todos.get("todos", [])),
            "status_counts": todos.get("status_counts", {}),
        }
    except Exception as e:
        logger.debug(f"Could not get tasks summary: {e}")
        # Fallback: return empty structure
        return {
            "recent": [],
            "count": 0,
            "status_counts": {},
            "note": "Task summary unavailable",
        }


def get_prompts_for_mode(mode: str) -> Dict[str, Any]:
    """
    Get prompts relevant to a workflow mode.
    
    Args:
        mode: Workflow mode name
        
    Returns:
        Dict with recommended prompts for the mode
    """
    try:
        from .prompt_discovery import get_prompts_for_mode as get_mode_prompts
        
        result = get_mode_prompts(mode)
        
        return {
            "recommended": result.get("prompts", []),
            "count": result.get("count", 0),
            "mode": mode,
        }
    except Exception as e:
        logger.warning(f"Failed to get prompts for mode {mode}: {e}")
        return {
            "recommended": [],
            "count": 0,
            "mode": mode,
        }


def get_context_primer(
    mode: str,
    include_hints: bool = True,
    include_tasks: bool = True,
    include_goals: bool = True,
    include_prompts: bool = True,
) -> str:
    """
    Get context primer for a workflow mode.
    
    Args:
        mode: Workflow mode name
        include_hints: Include tool hints for the mode
        include_tasks: Include task summary
        include_goals: Include mode goals
        include_prompts: Include recommended prompts
        
    Returns:
        JSON string with context primer data
    """
    # Get mode context
    mode_context = WORKFLOW_MODE_CONTEXT.get(mode, WORKFLOW_MODE_CONTEXT.get("development", {}))
    
    # Build primer dict
    primer: Dict[str, Any] = {
        "workflow": {
            "mode": mode,
            "description": mode_context.get("description", ""),
        },
    }
    
    # Add goals if requested
    if include_goals:
        primer["goals"] = mode_context.get("goals", [])
    
    # Add hints if requested
    if include_hints:
        primer["hints"] = get_hints_for_mode(mode)
    
    # Add tasks if requested
    if include_tasks:
        primer["tasks"] = get_tasks_summary()
    
    # Add prompts if requested
    if include_prompts:
        primer["prompts"] = get_prompts_for_mode(mode)
    
    # Return as JSON string
    return json.dumps(primer, indent=2)


__all__ = [
    "WORKFLOW_MODE_CONTEXT",
    "get_context_primer",
    "get_hints_for_mode",
    "get_tasks_summary",
    "get_prompts_for_mode",
]

