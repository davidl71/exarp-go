"""
Context Primer for Workflow Modes

Provides context priming functionality for different workflow modes.
Supplies hints, tasks, prompts, and goals based on the current workflow mode.
"""

import json
import logging
from typing import Any, Dict, List, Optional

logger = logging.getLogger("exarp.resources.context_primer")

# Prompt manifest (aligned with stdio://prompts / Go; was in prompt_discovery.py)
PROMPT_MODE_MAPPING: Dict[str, List[str]] = {
    "daily_checkin": ["advisor_consult", "advisor_briefing"],
    "security_review": ["persona_security"],
    "task_management": ["task_review", "weekly_maintenance"],
    "code_review": ["project_health", "persona_code_reviewer"],
    "sprint_planning": ["automation_setup", "automation_discovery", "persona_project_manager"],
    "debugging": ["persona_developer"],
    "development": ["persona_developer", "weekly_maintenance"],
}
PROMPT_PERSONA_MAPPING: Dict[str, List[str]] = {
    "developer": ["persona_developer"],
    "project_manager": ["persona_project_manager", "task_review"],
    "code_reviewer": ["persona_code_reviewer", "project_health"],
    "security_engineer": ["persona_security"],
    "qa_engineer": ["persona_qa", "project_health"],
    "architect": ["persona_architect"],
    "tech_writer": ["persona_tech_writer", "doc_health_check"],
    "executive": ["persona_executive"],
}
PROMPT_CATEGORIES: Dict[str, List[str]] = {
    "documentation": ["doc_health_check"],
    "automation": ["automation_discovery", "automation_setup"],
    "workflow": ["weekly_maintenance", "task_review", "project_health"],
    "wisdom": ["advisor_consult", "advisor_briefing"],
    "persona": [
        "persona_developer", "persona_project_manager", "persona_code_reviewer",
        "persona_executive", "persona_security", "persona_architect", "persona_qa", "persona_tech_writer",
    ],
}
PROMPT_DESCRIPTIONS: Dict[str, str] = {
    "doc_health_check": "Analyze documentation health and create tasks for issues",
    "automation_discovery": "Discover new automation opportunities in codebase",
    "automation_setup": "One-time automation setup workflow",
    "weekly_maintenance": "Weekly maintenance workflow",
    "task_review": "Comprehensive task review workflow for backlog hygiene",
    "project_health": "Comprehensive project health assessment",
    "advisor_consult": "Consult a trusted advisor for wisdom on current work",
    "advisor_briefing": "Get morning briefing from trusted advisors",
    "persona_developer": "Developer daily workflow for writing quality code",
    "persona_project_manager": "Project Manager workflow for delivery tracking",
    "persona_code_reviewer": "Code Reviewer workflow for quality gates",
    "persona_executive": "Executive/Stakeholder workflow for strategic view",
    "persona_security": "Security Engineer workflow for risk management",
    "persona_architect": "Architect workflow for system design",
    "persona_qa": "QA Engineer workflow for quality assurance",
    "persona_tech_writer": "Technical Writer workflow for documentation",
}

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
    prompts = PROMPT_MODE_MAPPING.get(mode, PROMPT_MODE_MAPPING.get("development", []))
    return {
        "recommended": [{"name": p, "description": PROMPT_DESCRIPTIONS.get(p, "")} for p in prompts],
        "count": len(prompts),
        "mode": mode,
    }


def discover_prompts(
    mode: Optional[str] = None,
    persona: Optional[str] = None,
    category: Optional[str] = None,
    keywords: Optional[List[str]] = None,
) -> Dict[str, Any]:
    """
    Discover prompts by mode/persona/category/keywords.
    """
    all_prompts = set(PROMPT_DESCRIPTIONS.keys())
    if mode:
        mode_prompts = set(PROMPT_MODE_MAPPING.get(mode, []))
        all_prompts = all_prompts.intersection(mode_prompts) if mode_prompts else all_prompts
    if persona:
        persona_prompts = set(PROMPT_PERSONA_MAPPING.get(persona, []))
        all_prompts = all_prompts.intersection(persona_prompts) if persona_prompts else all_prompts
    if category:
        category_prompts = set(PROMPT_CATEGORIES.get(category, []))
        all_prompts = all_prompts.intersection(category_prompts) if category_prompts else all_prompts
    if keywords:
        keyword_matches = {
            p for p in all_prompts
            if any(kw.lower() in PROMPT_DESCRIPTIONS.get(p, "").lower() for kw in keywords)
        }
        all_prompts = keyword_matches if keyword_matches else all_prompts
    return {
        "prompts": [{"name": p, "description": PROMPT_DESCRIPTIONS.get(p, "")} for p in sorted(all_prompts)],
        "count": len(all_prompts),
        "filters_applied": {"mode": mode, "persona": persona, "category": category, "keywords": keywords},
    }


def discover_prompts_tool(
    mode: Optional[str] = None,
    persona: Optional[str] = None,
    category: Optional[str] = None,
    keywords: Optional[str] = None,
) -> str:
    """Discover prompts (returns JSON string for consolidated_git / find_prompts)."""
    kw_list = [k.strip() for k in keywords.split(",")] if keywords else None
    return json.dumps(discover_prompts(mode=mode, persona=persona, category=category, keywords=kw_list), separators=(",", ":"))


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
    "discover_prompts",
    "discover_prompts_tool",
]

