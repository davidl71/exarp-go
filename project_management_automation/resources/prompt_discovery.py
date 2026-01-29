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

# Map prompts to workflow modes (manifest aligned with stdio://prompts / Go; prompts.py retired)
PROMPT_MODE_MAPPING: Dict[str, List[str]] = {
    "daily_checkin": [
        "advisor_consult",
        "advisor_briefing",
    ],
    "security_review": [
        "persona_security",
    ],
    "task_management": [
        "task_review",
        "weekly_maintenance",
    ],
    "code_review": [
        "project_health",
        "persona_code_reviewer",
    ],
    "sprint_planning": [
        "automation_setup",
        "automation_discovery",
        "persona_project_manager",
    ],
    "debugging": [
        "persona_developer",
    ],
    "development": [
        "persona_developer",
        "weekly_maintenance",
    ],
}

# Map prompts to personas (manifest aligned with stdio://prompts / Go)
PROMPT_PERSONA_MAPPING: Dict[str, List[str]] = {
    "developer": [
        "persona_developer",
    ],
    "project_manager": [
        "persona_project_manager",
        "task_review",
    ],
    "code_reviewer": [
        "persona_code_reviewer",
        "project_health",
    ],
    "security_engineer": [
        "persona_security",
    ],
    "qa_engineer": [
        "persona_qa",
        "project_health",
    ],
    "architect": [
        "persona_architect",
    ],
    "tech_writer": [
        "persona_tech_writer",
        "doc_health_check",
    ],
    "executive": [
        "persona_executive",
    ],
}

# Map prompts to categories (manifest aligned with stdio://prompts / Go)
PROMPT_CATEGORIES: Dict[str, List[str]] = {
    "documentation": [
        "doc_health_check",
    ],
    "automation": [
        "automation_discovery",
        "automation_setup",
    ],
    "workflow": [
        "weekly_maintenance",
        "task_review",
        "project_health",
    ],
    "wisdom": [
        "advisor_consult",
        "advisor_briefing",
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

# Compact prompt descriptions for quick discovery (aligned with stdio://prompts / Go)
PROMPT_DESCRIPTIONS: Dict[str, str] = {
    # Documentation
    "doc_health_check": "Analyze documentation health and create tasks for issues",

    # Automation
    "automation_discovery": "Discover new automation opportunities in codebase",
    "automation_setup": "One-time automation setup workflow",

    # Workflows
    "weekly_maintenance": "Weekly maintenance workflow",
    "task_review": "Comprehensive task review workflow for backlog hygiene",
    "project_health": "Comprehensive project health assessment",

    # Wisdom
    "advisor_consult": "Consult a trusted advisor for wisdom on current work",
    "advisor_briefing": "Get morning briefing from trusted advisors",

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

