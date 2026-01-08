"""
AI Session Memory MCP Tools

Provides tools for AI sessions to save and retrieve memories:
- save_session_insight: Save discoveries during work
- recall_task_context: Get memories related to a task
- search_session_memories: Search past insights
- generate_session_summary: End-of-session summary

Trusted Advisor: 🎓 Confucius (Documentation)
"Choose a job you love, and you will never have to work a day in your life." - Passion sustains.
"""

import json
import logging
from datetime import datetime
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)


def _get_memory_functions():
    """Import memory functions from resources module."""
    try:
        from ..resources.memories import (
            MEMORY_CATEGORIES,
            create_memory,
            get_memories_by_task_resource,
            get_memories_resource,
            get_memory_by_id,
            get_session_memories_resource,
            get_wisdom_resource,
            search_memories,
        )
        return {
            'create_memory': create_memory,
            'get_memory_by_id': get_memory_by_id,
            'search_memories': search_memories,
            'get_memories_resource': get_memories_resource,
            'get_memories_by_task_resource': get_memories_by_task_resource,
            'get_session_memories_resource': get_session_memories_resource,
            'get_wisdom_resource': get_wisdom_resource,
            'MEMORY_CATEGORIES': MEMORY_CATEGORIES,
        }
    except ImportError as e:
        logger.error(f"Failed to import memory functions: {e}")
        return None


# ═══════════════════════════════════════════════════════════════════════════════
# MCP TOOL IMPLEMENTATIONS
# ═══════════════════════════════════════════════════════════════════════════════

def save_session_insight(
    title: str,
    content: str,
    category: str = "insight",
    task_id: Optional[str] = None,
    metadata: Optional[Dict[str, Any]] = None,
) -> dict[str, Any]:
    """
    Save a session insight/discovery to persistent memory.

    Use this to record:
    - Debug solutions ("Fixed ImportError by adding __init__.py")
    - Research findings ("Compared approaches A vs B, chose A because...")
    - Architecture discoveries ("Module X depends on Y through Z")
    - User preferences ("User prefers verbose logging")
    - Sprint insights ("Tasks without estimates tend to stall")

    Args:
        title: Short descriptive title (max 100 chars)
        content: Full insight content (detailed description)
        category: One of: debug, research, architecture, preference, insight
        task_id: Optional task ID to link this memory to
        metadata: Optional additional key-value metadata

    Returns:
        Result dict with created memory or error
    """
    funcs = _get_memory_functions()
    if not funcs:
        return {
            "success": False,
            "error": "Memory system not available",
        }

    try:
        # Validate category
        if category not in funcs['MEMORY_CATEGORIES']:
            return {
                "success": False,
                "error": f"Invalid category '{category}'. Must be one of: {', '.join(funcs['MEMORY_CATEGORIES'])}",
            }

        # Truncate title if too long
        if len(title) > 100:
            title = title[:97] + "..."

        # Create linked_tasks list
        linked_tasks = [task_id] if task_id else []

        # Create the memory
        memory = funcs['create_memory'](
            title=title,
            content=content,
            category=category,
            linked_tasks=linked_tasks,
            metadata=metadata,
        )

        return {
            "success": True,
            "memory_id": memory['id'],
            "title": memory['title'],
            "category": memory['category'],
            "linked_tasks": memory['linked_tasks'],
            "created_at": memory['created_at'],
            "message": f"✅ Memory saved: {title}",
        }

    except Exception as e:
        logger.error(f"Error saving session insight: {e}")
        return {
            "success": False,
            "error": str(e),
        }


def recall_task_context(
    task_id: str,
    include_related: bool = True,
) -> dict[str, Any]:
    """
    Recall all memories related to a task.

    Use this before starting work on a task to:
    - See what was previously discovered
    - Review past approaches tried
    - Understand decisions already made
    - Find related debug solutions

    Args:
        task_id: Task ID to get context for
        include_related: Whether to include memories from related tasks

    Returns:
        Result dict with memories and summary
    """
    funcs = _get_memory_functions()
    if not funcs:
        return {
            "success": False,
            "error": "Memory system not available",
        }

    try:
        # Get memories for this task
        result_json = funcs['get_memories_by_task_resource'](task_id)
        result = json.loads(result_json)

        memories = result.get('memories', [])

        # Group by category
        by_category = {}
        for m in memories:
            cat = m.get('category', 'unknown')
            if cat not in by_category:
                by_category[cat] = []
            by_category[cat].append(m)

        # Generate context summary
        summary_parts = []
        if 'debug' in by_category:
            summary_parts.append(f"🔧 {len(by_category['debug'])} debug solution(s)")
        if 'research' in by_category:
            summary_parts.append(f"📚 {len(by_category['research'])} research finding(s)")
        if 'architecture' in by_category:
            summary_parts.append(f"🏗️ {len(by_category['architecture'])} architecture note(s)")
        if 'preference' in by_category:
            summary_parts.append(f"⚙️ {len(by_category['preference'])} preference(s)")
        if 'insight' in by_category:
            summary_parts.append(f"💡 {len(by_category['insight'])} insight(s)")

        return {
            "success": True,
            "task_id": task_id,
            "total_memories": len(memories),
            "summary": ", ".join(summary_parts) if summary_parts else "No memories for this task",
            "by_category": {k: len(v) for k, v in by_category.items()},
            "memories": memories,
        }

    except Exception as e:
        logger.error(f"Error recalling task context: {e}")
        return {
            "success": False,
            "task_id": task_id,
            "error": str(e),
        }


def search_session_memories(
    query: str,
    category: Optional[str] = None,
    limit: int = 10,
) -> dict[str, Any]:
    """
    Search past session memories.

    Use this to find:
    - Similar problems and their solutions
    - Past research on a topic
    - Previous decisions about similar features

    Args:
        query: Search query text
        category: Optional category filter
        limit: Maximum results to return

    Returns:
        Result dict with matching memories
    """
    funcs = _get_memory_functions()
    if not funcs:
        return {
            "success": False,
            "error": "Memory system not available",
        }

    try:
        # Search memories
        memories = funcs['search_memories'](query, limit=limit)

        # Apply category filter if specified
        if category:
            memories = [m for m in memories if m.get('category') == category]

        return {
            "success": True,
            "query": query,
            "category_filter": category,
            "total_results": len(memories),
            "memories": memories,
        }

    except Exception as e:
        logger.error(f"Error searching memories: {e}")
        return {
            "success": False,
            "query": query,
            "error": str(e),
        }


def generate_session_summary(
    date: Optional[str] = None,
    include_consultations: bool = True,
) -> dict[str, Any]:
    """
    Generate a summary of a session's learnings.

    Use this at end of session to:
    - Review what was learned
    - See all insights captured
    - Get combined wisdom (memories + advisor consultations)

    Args:
        date: Session date in YYYY-MM-DD format (default: today)
        include_consultations: Whether to include advisor consultations

    Returns:
        Result dict with session summary
    """
    funcs = _get_memory_functions()
    if not funcs:
        return {
            "success": False,
            "error": "Memory system not available",
        }

    try:
        # Get session memories
        session_json = funcs['get_session_memories_resource'](date)
        session = json.loads(session_json)

        memories = session.get('memories', [])
        session_date = session.get('session_date', date or datetime.now().strftime("%Y-%m-%d"))

        # Get wisdom if requested (deprecated - use devwisdom-go MCP server directly)
        wisdom = None
        if include_consultations:
            # Note: get_wisdom_resource is deprecated, use devwisdom-go MCP server directly
            wisdom = {"consultations": {"recent": 0, "items": []}}

        # Build narrative summary
        narrative = _build_session_narrative(memories, wisdom)

        return {
            "success": True,
            "session_date": session_date,
            "memories_count": len(memories),
            "by_category": session.get('by_category', {}),
            "consultations_count": wisdom.get('consultations', {}).get('recent', 0) if wisdom else 0,
            "narrative": narrative,
            "memories": memories,
            "advisor_wisdom": wisdom.get('consultations', {}).get('items', [])[:5] if wisdom else [],
        }

    except Exception as e:
        logger.error(f"Error generating session summary: {e}")
        return {
            "success": False,
            "error": str(e),
        }


def _build_session_narrative(
    memories: list[dict[str, Any]],
    wisdom: Optional[Dict[str, Any]] = None,
) -> str:
    """Build a narrative summary of the session."""
    parts = []

    if not memories and not wisdom:
        return "No memories or consultations recorded for this session."

    if memories:
        # Group memories
        debug = [m for m in memories if m.get('category') == 'debug']
        research = [m for m in memories if m.get('category') == 'research']
        architecture = [m for m in memories if m.get('category') == 'architecture']
        insights = [m for m in memories if m.get('category') in ('insight', 'preference')]

        if debug:
            parts.append(f"🔧 Solved {len(debug)} issue(s): " +
                        ", ".join(m.get('title', '')[:50] for m in debug[:3]))

        if research:
            parts.append(f"📚 Researched {len(research)} topic(s): " +
                        ", ".join(m.get('title', '')[:50] for m in research[:3]))

        if architecture:
            parts.append(f"🏗️ Documented {len(architecture)} architecture insight(s)")

        if insights:
            parts.append(f"💡 Captured {len(insights)} insight(s)/preference(s)")

    if wisdom:
        consultations = wisdom.get('consultations', {}).get('items', [])
        if consultations:
            advisors = list({c.get('advisor_name', '') for c in consultations})
            parts.append(f"🧙‍♂️ Consulted {len(advisors)} advisor(s): " +
                        ", ".join(advisors[:4]))

    return " | ".join(parts) if parts else "Session in progress..."


def link_memory_to_task(
    memory_id: str,
    task_id: str,
) -> dict[str, Any]:
    """
    Link an existing memory to a task.

    Args:
        memory_id: ID of memory to link
        task_id: Task ID to link to

    Returns:
        Result dict with updated memory
    """
    funcs = _get_memory_functions()
    if not funcs:
        return {
            "success": False,
            "error": "Memory system not available",
        }

    try:
        memory = funcs['get_memory_by_id'](memory_id)
        if not memory:
            return {
                "success": False,
                "error": f"Memory not found: {memory_id}",
            }

        # Add task to linked_tasks if not already there
        linked = memory.get('linked_tasks', [])
        if task_id not in linked:
            linked.append(task_id)
            memory['linked_tasks'] = linked

            # Save updated memory
            from ..resources.memories import _save_memory
            _save_memory(memory)

        return {
            "success": True,
            "memory_id": memory_id,
            "task_id": task_id,
            "linked_tasks": memory['linked_tasks'],
            "message": f"✅ Linked memory to task {task_id}",
        }

    except Exception as e:
        logger.error(f"Error linking memory to task: {e}")
        return {
            "success": False,
            "error": str(e),
        }


def get_memories_for_sprint() -> dict[str, Any]:
    """
    Get memories useful for sprint planning/review.

    Returns recent insights, debug solutions, and patterns
    that could inform sprint decisions.

    Returns:
        Result dict with sprint-relevant memories
    """
    funcs = _get_memory_functions()
    if not funcs:
        return {
            "success": False,
            "error": "Memory system not available",
        }

    try:
        # Get recent memories (last 7 days)
        all_memories = json.loads(funcs['get_memories_resource'](limit=100))
        memories = all_memories.get('memories', [])

        # Filter to last 7 days
        from datetime import timedelta
        cutoff = (datetime.now() - timedelta(days=7)).isoformat()
        recent = [m for m in memories if m.get('created_at', '') >= cutoff]

        # Categorize for sprint
        blockers = [m for m in recent if 'block' in m.get('content', '').lower() or
                   'block' in m.get('title', '').lower()]
        solutions = [m for m in recent if m.get('category') == 'debug']
        patterns = [m for m in recent if m.get('category') == 'insight']

        return {
            "success": True,
            "period": "last_7_days",
            "total_memories": len(recent),
            "blockers_mentioned": len(blockers),
            "debug_solutions": len(solutions),
            "patterns_observed": len(patterns),
            "highlights": {
                "blockers": [m.get('title', '') for m in blockers[:5]],
                "solutions": [m.get('title', '') for m in solutions[:5]],
                "patterns": [m.get('title', '') for m in patterns[:5]],
            },
            "memories": recent[:20],
        }

    except Exception as e:
        logger.error(f"Error getting sprint memories: {e}")
        return {
            "success": False,
            "error": str(e),
        }


__all__ = [
    'save_session_insight',
    'recall_task_context',
    'search_session_memories',
    'generate_session_summary',
    'link_memory_to_task',
    'get_memories_for_sprint',
]


