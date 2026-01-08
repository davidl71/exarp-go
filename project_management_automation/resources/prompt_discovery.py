"""
Mode-Aware Prompt Discovery Resource

Provides intelligent prompt discovery based on:
- Current workflow mode
- Agent type
- Task context
- Recent activity

This helps AI assistants quickly find relevant prompts
without searching through the entire prompt catalog.

Usage:
    Resource URI: automation://prompts/{mode}
    Resource URI: automation://prompts/persona/{persona}
    Resource URI: automation://prompts/category/{category}
"""

import json
import logging
from typing import Any, Dict, List, Optional

logger = logging.getLogger("exarp.prompt_discovery")


# ═══════════════════════════════════════════════════════════════════════════════
# PROMPT METADATA (Extracted from prompts.py)
# ═══════════════════════════════════════════════════════════════════════════════

# Map prompts to workflow modes
PROMPT_MODE_MAPPING: Dict[str, List[str]] = {
    "daily_checkin": [
        "daily_checkin",
        "project_scorecard",
        "project_overview",
        "advisor_consult",
        "advisor_briefing",
    ],
    "security_review": [
        "security_scan_all",
        "security_scan_python",
        "security_scan_rust",
        "persona_security",
    ],
    "task_management": [
        "task_alignment",
        "duplicate_cleanup",
        "task_sync",
        "task_discovery",
        "task_review",
        "pre_sprint_cleanup",
    ],
    "code_review": [
        "project_health",
        "persona_code_reviewer",
        "post_implementation_review",
    ],
    "sprint_planning": [
        "sprint_start",
        "sprint_end",
        "pre_sprint_cleanup",
        "automation_setup",
        "automation_discovery",
        "automation_high_value",
        "persona_project_manager",
    ],
    "debugging": [
        "memory_system",
        "persona_developer",
        "context_management",
    ],
    "development": [
        "persona_developer",
        "mode_suggestion",
        "context_management",
        "config_generation",
        "weekly_maintenance",
    ],
}

# Map prompts to personas
PROMPT_PERSONA_MAPPING: Dict[str, List[str]] = {
    "developer": [
        "persona_developer",
        "mode_suggestion",
        "context_management",
        "config_generation",
        "memory_system",
    ],
    "project_manager": [
        "persona_project_manager",
        "sprint_start",
        "sprint_end",
        "task_alignment",
        "task_review",
        "project_overview",
    ],
    "code_reviewer": [
        "persona_code_reviewer",
        "post_implementation_review",
        "project_health",
    ],
    "security_engineer": [
        "persona_security",
        "security_scan_all",
        "security_scan_python",
        "security_scan_rust",
    ],
    "qa_engineer": [
        "persona_qa",
        "project_health",
    ],
    "architect": [
        "persona_architect",
        "project_scorecard",
    ],
    "tech_writer": [
        "persona_tech_writer",
        "doc_health_check",
        "doc_quick_check",
    ],
    "executive": [
        "persona_executive",
        "project_overview",
        "project_scorecard",
    ],
}

# Map prompts to categories
PROMPT_CATEGORIES: Dict[str, List[str]] = {
    "documentation": [
        "doc_health_check",
        "doc_quick_check",
    ],
    "tasks": [
        "task_alignment",
        "duplicate_cleanup",
        "task_sync",
        "task_discovery",
    ],
    "security": [
        "security_scan_all",
        "security_scan_python",
        "security_scan_rust",
    ],
    "automation": [
        "automation_discovery",
        "automation_high_value",
        "automation_setup",
    ],
    "config": [
        "config_generation",
    ],
    "workflow": [
        "pre_sprint_cleanup",
        "post_implementation_review",
        "weekly_maintenance",
        "daily_checkin",
        "sprint_start",
        "sprint_end",
        "task_review",
        "project_health",
        "mode_suggestion",
        "context_management",
    ],
    "reports": [
        "project_scorecard",
        "project_overview",
    ],
    "wisdom": [
        "advisor_consult",
        "advisor_briefing",
        # advisor_audio removed - migrated to devwisdom-go MCP server
    ],
    "memory": [
        "memory_system",
    ],
    "persona": [
        "persona_developer",
        "persona_project_manager",
        "persona_code_reviewer",
        "persona_executive",
        "persona_security",
        "persona_architect",
        "persona_qa",
        "persona_tech_writer",
    ],
}

# Compact prompt descriptions for quick discovery
PROMPT_DESCRIPTIONS: Dict[str, str] = {
    # Documentation
    "doc_health_check": "Analyze documentation health and create tasks for issues",
    "doc_quick_check": "Quick documentation health check without creating tasks",

    # Task Management
    "task_alignment": "Analyze Todo2 task alignment with project goals",
    "duplicate_cleanup": "Find and consolidate duplicate Todo2 tasks",
    "task_sync": "Synchronize tasks between shared TODO table and Todo2",
    "task_discovery": "Discover tasks from code comments, markdown, and orphans",

    # Security
    "security_scan_all": "Scan all dependencies for security vulnerabilities",
    "security_scan_python": "Scan Python dependencies for vulnerabilities",
    "security_scan_rust": "Scan Rust dependencies for vulnerabilities",

    # Automation
    "automation_discovery": "Discover new automation opportunities in codebase",
    "automation_high_value": "Find only high-value automation opportunities",
    "automation_setup": "One-time automation setup workflow",

    # Config
    "config_generation": "Generate IDE configuration files",

    # Workflows
    "daily_checkin": "Daily check-in workflow for project health monitoring",
    "pre_sprint_cleanup": "Pre-sprint cleanup workflow",
    "post_implementation_review": "Post-implementation review workflow",
    "weekly_maintenance": "Weekly maintenance workflow",
    "sprint_start": "Sprint start workflow for preparing clean backlog",
    "sprint_end": "Sprint end workflow for quality assurance",
    "task_review": "Comprehensive task review workflow for backlog hygiene",
    "project_health": "Comprehensive project health assessment",

    # Reports
    "project_scorecard": "Generate comprehensive project health scorecard",
    "project_overview": "Generate one-page project overview for stakeholders",

    # Wisdom
    "advisor_consult": "Consult a trusted advisor for wisdom on current work",
    "advisor_briefing": "Get morning briefing from trusted advisors",
    # advisor_audio removed - migrated to devwisdom-go MCP server

    # Memory
    "memory_system": "Use AI session memory to persist insights across sessions",

    # Workflow Helpers
    "mode_suggestion": "Suggest optimal Cursor IDE mode (Agent vs Ask)",
    "context_management": "Strategically manage LLM context to reduce token usage",

    # Personas
    "persona_developer": "Developer daily workflow for writing quality code",
    "persona_project_manager": "Project Manager workflow for delivery tracking",
    "persona_code_reviewer": "Code Reviewer workflow for quality gates",
    "persona_executive": "Executive/Stakeholder workflow for strategic view",
    "persona_security": "Security Engineer workflow for risk management",
    "persona_architect": "Architect workflow for system design",
    "persona_qa": "QA Engineer workflow for quality assurance",
    "persona_tech_writer": "Technical Writer workflow for documentation",
}


# ═══════════════════════════════════════════════════════════════════════════════
# DISCOVERY FUNCTIONS
# ═══════════════════════════════════════════════════════════════════════════════

def get_prompts_for_mode(mode: str) -> Dict[str, Any]:
    """
    Get prompts relevant to a workflow mode.
    
    Args:
        mode: Workflow mode (daily_checkin, security_review, etc.)
    
    Returns:
        Dict with mode info and relevant prompts
    """
    prompts = PROMPT_MODE_MAPPING.get(mode, PROMPT_MODE_MAPPING.get("development", []))

    return {
        "mode": mode,
        "prompts": [
            {
                "name": p,
                "description": PROMPT_DESCRIPTIONS.get(p, ""),
            }
            for p in prompts
        ],
        "count": len(prompts),
        "available_modes": list(PROMPT_MODE_MAPPING.keys()),
    }


def get_prompts_for_persona(persona: str) -> Dict[str, Any]:
    """
    Get prompts relevant to a persona.
    
    Args:
        persona: Target persona (developer, project_manager, etc.)
    
    Returns:
        Dict with persona info and relevant prompts
    """
    prompts = PROMPT_PERSONA_MAPPING.get(persona, [])

    return {
        "persona": persona,
        "prompts": [
            {
                "name": p,
                "description": PROMPT_DESCRIPTIONS.get(p, ""),
            }
            for p in prompts
        ],
        "count": len(prompts),
        "available_personas": list(PROMPT_PERSONA_MAPPING.keys()),
    }


def get_prompts_for_category(category: str) -> Dict[str, Any]:
    """
    Get prompts in a category.
    
    Args:
        category: Prompt category (documentation, tasks, security, etc.)
    
    Returns:
        Dict with category info and prompts
    """
    prompts = PROMPT_CATEGORIES.get(category, [])

    return {
        "category": category,
        "prompts": [
            {
                "name": p,
                "description": PROMPT_DESCRIPTIONS.get(p, ""),
            }
            for p in prompts
        ],
        "count": len(prompts),
        "available_categories": list(PROMPT_CATEGORIES.keys()),
    }


def get_all_prompts_compact() -> Dict[str, Any]:
    """
    Get all prompts in compact format.
    
    Returns:
        Dict with all prompts organized by category
    """
    by_category = {}
    for category, prompts in PROMPT_CATEGORIES.items():
        by_category[category] = [
            {
                "name": p,
                "description": PROMPT_DESCRIPTIONS.get(p, ""),
            }
            for p in prompts
        ]

    return {
        "total_prompts": len(PROMPT_DESCRIPTIONS),
        "categories": list(PROMPT_CATEGORIES.keys()),
        "by_category": by_category,
    }


def discover_prompts(
    mode: Optional[str] = None,
    persona: Optional[str] = None,
    category: Optional[str] = None,
    keywords: Optional[List[str]] = None,
) -> Dict[str, Any]:
    """
    Intelligent prompt discovery based on multiple filters.
    
    Args:
        mode: Filter by workflow mode
        persona: Filter by persona
        category: Filter by category
        keywords: Filter by keywords in description
    
    Returns:
        Dict with discovered prompts
    """
    # Start with all prompts
    all_prompts = set(PROMPT_DESCRIPTIONS.keys())

    # Filter by mode
    if mode:
        mode_prompts = set(PROMPT_MODE_MAPPING.get(mode, []))
        all_prompts = all_prompts.intersection(mode_prompts) if mode_prompts else all_prompts

    # Filter by persona
    if persona:
        persona_prompts = set(PROMPT_PERSONA_MAPPING.get(persona, []))
        all_prompts = all_prompts.intersection(persona_prompts) if persona_prompts else all_prompts

    # Filter by category
    if category:
        category_prompts = set(PROMPT_CATEGORIES.get(category, []))
        all_prompts = all_prompts.intersection(category_prompts) if category_prompts else all_prompts

    # Filter by keywords
    if keywords:
        keyword_matches = set()
        for p in all_prompts:
            desc = PROMPT_DESCRIPTIONS.get(p, "").lower()
            if any(kw.lower() in desc for kw in keywords):
                keyword_matches.add(p)
        all_prompts = keyword_matches if keyword_matches else all_prompts

    return {
        "prompts": [
            {
                "name": p,
                "description": PROMPT_DESCRIPTIONS.get(p, ""),
            }
            for p in sorted(all_prompts)
        ],
        "count": len(all_prompts),
        "filters_applied": {
            "mode": mode,
            "persona": persona,
            "category": category,
            "keywords": keywords,
        },
    }


# ═══════════════════════════════════════════════════════════════════════════════
# MCP RESOURCE REGISTRATION
# ═══════════════════════════════════════════════════════════════════════════════

def register_prompt_discovery_resources(mcp) -> None:
    """Register prompt discovery resources with MCP server."""
    try:
        @mcp.resource("automation://prompts")
        def all_prompts_resource() -> str:
            """Get all prompts in compact format."""
            return json.dumps(get_all_prompts_compact(), separators=(',', ':'))

        @mcp.resource("automation://prompts/mode/{mode}")
        def prompts_by_mode_resource(mode: str) -> str:
            """Get prompts for a workflow mode."""
            return json.dumps(get_prompts_for_mode(mode), separators=(',', ':'))

        @mcp.resource("automation://prompts/persona/{persona}")
        def prompts_by_persona_resource(persona: str) -> str:
            """Get prompts for a persona."""
            return json.dumps(get_prompts_for_persona(persona), separators=(',', ':'))

        @mcp.resource("automation://prompts/category/{category}")
        def prompts_by_category_resource(category: str) -> str:
            """Get prompts in a category."""
            return json.dumps(get_prompts_for_category(category), separators=(',', ':'))

        logger.info("✅ Registered 4 prompt discovery resources")

    except Exception as e:
        logger.warning(f"Could not register prompt discovery resources: {e}")


# Tool for interactive prompt discovery
def discover_prompts_tool(
    mode: Optional[str] = None,
    persona: Optional[str] = None,
    category: Optional[str] = None,
    keywords: Optional[str] = None,
) -> str:
    """
    [HINT: Prompt discovery. Find relevant prompts by mode/persona/category/keywords.]
    
    Discover relevant prompts based on filters.
    
    Args:
        mode: Filter by workflow mode
        persona: Filter by persona
        category: Filter by category
        keywords: Comma-separated keywords to search
    
    Returns:
        JSON with matching prompts
    """
    kw_list = [k.strip() for k in keywords.split(",")] if keywords else None
    result = discover_prompts(
        mode=mode,
        persona=persona,
        category=category,
        keywords=kw_list,
    )
    return json.dumps(result, separators=(',', ':'))


def register_prompt_discovery_tools(mcp) -> None:
    """Register prompt discovery tools with MCP server."""
    try:
        @mcp.tool()
        def find_prompts(
            mode: Optional[str] = None,
            persona: Optional[str] = None,
            category: Optional[str] = None,
            keywords: Optional[str] = None,
        ) -> str:
            """
            [HINT: Prompt discovery. Find relevant prompts by mode/persona/category/keywords.]
            
            Discover relevant prompts based on filters.
            """
            return discover_prompts_tool(mode, persona, category, keywords)

        logger.info("✅ Registered 1 prompt discovery tool")

    except Exception as e:
        logger.warning(f"Could not register prompt discovery tools: {e}")


__all__ = [
    "PROMPT_MODE_MAPPING",
    "PROMPT_PERSONA_MAPPING",
    "PROMPT_CATEGORIES",
    "PROMPT_DESCRIPTIONS",
    "get_prompts_for_mode",
    "get_prompts_for_persona",
    "get_prompts_for_category",
    "get_all_prompts_compact",
    "discover_prompts",
    "discover_prompts_tool",
    "register_prompt_discovery_resources",
    "register_prompt_discovery_tools",
]

