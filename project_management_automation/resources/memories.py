"""
MCP Resource Handler for AI Session Memories

Provides resource access to persistent AI session memories for:
- Session continuity across conversations
- Task-linked discoveries and insights
- Debug solutions and workarounds
- Architecture discoveries
- User preferences

Trusted Advisor: 🔮 Enochian (Architecture)
"Behold the face of your God, the beginning of comfort." - Find joy in revealing hidden patterns.
"""

import json
import logging
import uuid
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any, Optional

from ..utils import find_project_root
from ..utils.json_cache import JsonCacheManager

logger = logging.getLogger(__name__)

# Cache manager for memory files
_cache_manager = JsonCacheManager.get_instance()

# Memory categories
MEMORY_CATEGORIES = [
    "debug",  # Error solutions, workarounds, root causes
    "research",  # Pre-implementation findings, approach comparisons
    "architecture",  # Component relationships, hidden dependencies
    "preference",  # User coding style, workflow preferences
    "insight",  # Sprint patterns, blockers, optimizations
]


def _get_memories_dir() -> Path:
    """Get path to memories storage directory."""
    project_root = find_project_root()
    memories_dir = project_root / ".exarp" / "memories"
    memories_dir.mkdir(parents=True, exist_ok=True)
    return memories_dir


def _load_all_memories() -> list[dict[str, Any]]:
    """Load all memories from storage with per-file caching."""
    memories_dir = _get_memories_dir()
    memories = []

    for memory_file in memories_dir.glob("*.json"):
        try:
            # Use unified JSON cache for each memory file
            cache = _cache_manager.get_cache(memory_file, enable_stats=False)
            memory = cache.get_or_load()
            if isinstance(memory, dict):
                memories.append(memory)
        except Exception as e:
            logger.warning(f"Error loading memory {memory_file}: {e}")

    # Sort by created_at descending (newest first)
    memories.sort(key=lambda m: m.get("created_at", ""), reverse=True)
    return memories


def _save_memory(memory: dict[str, Any]) -> Path:
    """Save a memory to storage."""
    memories_dir = _get_memories_dir()
    memory_file = memories_dir / f"{memory['id']}.json"

    with open(memory_file, "w") as f:
        json.dump(memory, f, indent=2)

    return memory_file


def _filter_memories(
    memories: list[dict[str, Any]],
    category: Optional[str] = None,
    task_id: Optional[str] = None,
    days: Optional[int] = None,
    session_date: Optional[str] = None,
) -> list[dict[str, Any]]:
    """Filter memories by various criteria."""
    filtered = memories

    if category:
        filtered = [m for m in filtered if m.get("category") == category]

    if task_id:
        filtered = [m for m in filtered if task_id in m.get("linked_tasks", [])]

    if days:
        cutoff = (datetime.now() - timedelta(days=days)).isoformat()
        filtered = [m for m in filtered if m.get("created_at", "") >= cutoff]

    if session_date:
        filtered = [m for m in filtered if m.get("session_date") == session_date]

    return filtered


def create_memory(
    title: str,
    content: str,
    category: str,
    linked_tasks: Optional[list[str]] = None,
    metadata: Optional[dict[str, Any]] = None,
) -> dict[str, Any]:
    """
    Create a new memory.

    Args:
        title: Short title for the memory
        content: Full content/insight
        category: One of MEMORY_CATEGORIES
        linked_tasks: Optional list of task IDs this relates to
        metadata: Optional additional metadata

    Returns:
        Created memory dict
    """
    if category not in MEMORY_CATEGORIES:
        logger.warning(f"Unknown category '{category}', using 'insight'")
        category = "insight"

    memory = {
        "id": str(uuid.uuid4()),
        "title": title,
        "content": content,
        "category": category,
        "linked_tasks": linked_tasks or [],
        "metadata": metadata or {},
        "created_at": datetime.now().isoformat(),
        "session_date": datetime.now().strftime("%Y-%m-%d"),
    }

    _save_memory(memory)
    logger.info(f"Created memory: {title} [{category}]")

    return memory


def get_memory_by_id(memory_id: str) -> Optional[dict[str, Any]]:
    """Get a specific memory by ID."""
    memories_dir = _get_memories_dir()
    memory_file = memories_dir / f"{memory_id}.json"

    if not memory_file.exists():
        return None

    try:
        with open(memory_file) as f:
            return json.load(f)
    except Exception as e:
        logger.error(f"Error loading memory {memory_id}: {e}")
        return None


def search_memories(query: str, limit: int = 10) -> list[dict[str, Any]]:
    """
    Search memories by text content.

    Args:
        query: Search query
        limit: Maximum results

    Returns:
        List of matching memories
    """
    memories = _load_all_memories()
    query_lower = query.lower()

    scored = []
    for memory in memories:
        score = 0
        title = memory.get("title", "").lower()
        content = memory.get("content", "").lower()
        category = memory.get("category", "").lower()

        # Title match scores highest
        if query_lower in title:
            score += 10

        # Content match
        if query_lower in content:
            score += 5
            # Bonus for multiple occurrences
            score += content.count(query_lower)

        # Category match
        if query_lower in category:
            score += 3

        if score > 0:
            scored.append((score, memory))

    # Sort by score descending
    scored.sort(key=lambda x: x[0], reverse=True)

    return [m for _, m in scored[:limit]]


# ═══════════════════════════════════════════════════════════════════════════════
# MCP RESOURCE HANDLERS
# ═══════════════════════════════════════════════════════════════════════════════


def get_memories_resource(limit: int = 50) -> str:
    """
    Get all memories as resource.

    Returns:
        JSON string with memory list and statistics
    """
    try:
        memories = _load_all_memories()[:limit]

        # Calculate statistics
        categories = {}
        for m in _load_all_memories():
            cat = m.get("category", "unknown")
            categories[cat] = categories.get(cat, 0) + 1

        result = {
            "memories": memories,
            "total": len(_load_all_memories()),
            "returned": len(memories),
            "categories": categories,
            "available_categories": MEMORY_CATEGORIES,
            "timestamp": datetime.now().isoformat(),
        }

        return json.dumps(result, separators=(',', ':'))

    except Exception as e:
        logger.error(f"Error getting memories resource: {e}")
        return json.dumps({"memories": [], "error": str(e), "timestamp": datetime.now().isoformat()}, separators=(',', ':'))


def get_memories_by_category_resource(category: str, limit: int = 50) -> str:
    """
    Get memories filtered by category.

    Args:
        category: Memory category (debug, research, architecture, preference, insight)

    Returns:
        JSON string with filtered memories
    """
    try:
        memories = _load_all_memories()
        filtered = _filter_memories(memories, category=category)[:limit]

        result = {
            "category": category,
            "memories": filtered,
            "total": len(filtered),
            "timestamp": datetime.now().isoformat(),
        }

        return json.dumps(result, separators=(',', ':'))

    except Exception as e:
        logger.error(f"Error getting memories by category: {e}")
        return json.dumps(
            {"memories": [], "category": category, "error": str(e), "timestamp": datetime.now().isoformat()}, indent=2
        )


def get_memories_by_task_resource(task_id: str) -> str:
    """
    Get memories linked to a specific task.

    Args:
        task_id: Task ID to filter by

    Returns:
        JSON string with task-linked memories
    """
    try:
        memories = _load_all_memories()
        filtered = _filter_memories(memories, task_id=task_id)

        result = {
            "task_id": task_id,
            "memories": filtered,
            "total": len(filtered),
            "timestamp": datetime.now().isoformat(),
        }

        return json.dumps(result, separators=(',', ':'))

    except Exception as e:
        logger.error(f"Error getting memories by task: {e}")
        return json.dumps(
            {"memories": [], "task_id": task_id, "error": str(e), "timestamp": datetime.now().isoformat()}, indent=2
        )


def get_recent_memories_resource(hours: int = 24) -> str:
    """
    Get memories from the last N hours.

    Args:
        hours: Number of hours to look back (default 24)

    Returns:
        JSON string with recent memories
    """
    try:
        memories = _load_all_memories()
        cutoff = (datetime.now() - timedelta(hours=hours)).isoformat()
        recent = [m for m in memories if m.get("created_at", "") >= cutoff]

        result = {
            "hours": hours,
            "memories": recent,
            "total": len(recent),
            "timestamp": datetime.now().isoformat(),
        }

        return json.dumps(result, separators=(',', ':'))

    except Exception as e:
        logger.error(f"Error getting recent memories: {e}")
        return json.dumps(
            {"memories": [], "hours": hours, "error": str(e), "timestamp": datetime.now().isoformat()}, indent=2
        )


def get_session_memories_resource(date: Optional[str] = None) -> str:
    """
    Get memories from a specific session date.

    Args:
        date: Session date in YYYY-MM-DD format (default: today)

    Returns:
        JSON string with session memories
    """
    try:
        if not date:
            date = datetime.now().strftime("%Y-%m-%d")

        memories = _load_all_memories()
        filtered = _filter_memories(memories, session_date=date)

        # Group by category for session summary
        by_category = {}
        for m in filtered:
            cat = m.get("category", "unknown")
            if cat not in by_category:
                by_category[cat] = []
            by_category[cat].append(m)

        result = {
            "session_date": date,
            "memories": filtered,
            "total": len(filtered),
            "by_category": {k: len(v) for k, v in by_category.items()},
            "summary": _generate_session_summary(filtered),
            "timestamp": datetime.now().isoformat(),
        }

        return json.dumps(result, separators=(',', ':'))

    except Exception as e:
        logger.error(f"Error getting session memories: {e}")
        return json.dumps(
            {"memories": [], "session_date": date, "error": str(e), "timestamp": datetime.now().isoformat()}, indent=2
        )


def _generate_session_summary(memories: list[dict[str, Any]]) -> str:
    """Generate a brief summary of session memories."""
    if not memories:
        return "No memories recorded for this session."

    categories = {}
    for m in memories:
        cat = m.get("category", "unknown")
        categories[cat] = categories.get(cat, 0) + 1

    parts = []
    for cat, count in sorted(categories.items()):
        parts.append(f"{count} {cat}")

    tasks_mentioned = set()
    for m in memories:
        tasks_mentioned.update(m.get("linked_tasks", []))

    summary = f"Session recorded {len(memories)} memories: {', '.join(parts)}."
    if tasks_mentioned:
        summary += f" Linked to {len(tasks_mentioned)} task(s)."

    return summary


# ═══════════════════════════════════════════════════════════════════════════════
# WISDOM INTEGRATION
# ═══════════════════════════════════════════════════════════════════════════════


def get_wisdom_resource() -> str:
    """
    Get combined view of memories and advisor consultations.
    
    DEPRECATED: This resource has been migrated to devwisdom-go MCP server.
    Use devwisdom MCP server resources directly instead.
    
    The combined wisdom view is now available via devwisdom-go MCP server.

    Returns:
        JSON string with unified wisdom view
    """
    try:
        # Load memories
        memories = _load_all_memories()

        # Load advisor consultations (if available)
        project_root = find_project_root()
        log_dir = project_root / ".exarp" / "advisor_logs"

        consultations = []
        if log_dir.exists():
            for log_file in sorted(log_dir.glob("consultations_*.jsonl")):
                with open(log_file) as f:
                    for line in f:
                        try:
                            consultations.append(json.loads(line.strip()))
                        except json.JSONDecodeError:
                            continue

        # Recent (last 7 days)
        cutoff = (datetime.now() - timedelta(days=7)).isoformat()
        recent_memories = [m for m in memories if m.get("created_at", "") >= cutoff]
        recent_consultations = [c for c in consultations if c.get("timestamp", "") >= cutoff]

        result = {
            "memories": {
                "total": len(memories),
                "recent": len(recent_memories),
                "items": recent_memories[:20],
            },
            "consultations": {
                "total": len(consultations),
                "recent": len(recent_consultations),
                "items": recent_consultations[:20],
            },
            "combined_insights": _merge_wisdom(recent_memories, recent_consultations),
            "timestamp": datetime.now().isoformat(),
        }

        return json.dumps(result, separators=(',', ':'))

    except Exception as e:
        logger.error(f"Error getting wisdom resource: {e}")
        return json.dumps({"error": str(e), "timestamp": datetime.now().isoformat()}, separators=(',', ':'))


def _merge_wisdom(memories: list[dict[str, Any]], consultations: list[dict[str, Any]]) -> list[dict[str, Any]]:
    """Merge memories and consultations into timeline."""
    timeline = []

    for m in memories:
        timeline.append(
            {
                "type": "memory",
                "timestamp": m.get("created_at", ""),
                "title": m.get("title", ""),
                "category": m.get("category", ""),
                "id": m.get("id", ""),
            }
        )

    for c in consultations:
        timeline.append(
            {
                "type": "consultation",
                "timestamp": c.get("timestamp", ""),
                "advisor": c.get("advisor_name", ""),
                "quote": c.get("quote", "")[:100] + "..." if len(c.get("quote", "")) > 100 else c.get("quote", ""),
                "context": c.get("context", ""),
            }
        )

    # Sort by timestamp
    timeline.sort(key=lambda x: x.get("timestamp", ""), reverse=True)

    return timeline[:30]  # Last 30 items


def get_memories_health_resource() -> str:
    """
    Get memory system health metrics.
    
    Returns:
        JSON string with health analysis and recommendations
    """
    try:
        # Import maintenance functions
        from ..tools.memory_maintenance import memory_health_check

        health = memory_health_check()

        result = {
            "total_memories": health.get("total_memories", 0),
            "health_score": health.get("health_score", 0),
            "by_category": health.get("by_category", {}),
            "age_distribution": health.get("age_distribution", {}),
            "issues": health.get("issues", {}),
            "recommendations": health.get("recommendations", []),
            "timestamp": datetime.now().isoformat(),
        }

        return json.dumps(result, separators=(',', ':'))

    except Exception as e:
        logger.error(f"Error getting memories health: {e}")
        return json.dumps({
            "error": str(e),
            "timestamp": datetime.now().isoformat(),
        }, indent=2)


__all__ = [
    "MEMORY_CATEGORIES",
    "create_memory",
    "get_memory_by_id",
    "search_memories",
    "get_memories_resource",
    "get_memories_by_category_resource",
    "get_memories_by_task_resource",
    "get_recent_memories_resource",
    "get_session_memories_resource",
    "get_wisdom_resource",
    "get_memories_health_resource",
]

