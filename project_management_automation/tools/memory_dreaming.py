"""
Memory Dreaming - Advisor-Based Reflection System

Uses wisdom advisors to reflect on stored memories and generate insights.
The "dreaming" process analyzes patterns across memories and produces
consolidated wisdom with advisor quotes.

Trusted Advisor: 📜 Chacham (Wisdom)
"In dreams, we process the day's learning." - Reflection deepens understanding.
"""

import logging
from collections import defaultdict
from datetime import datetime, timedelta
from typing import Any, List, Optional

from ..resources.memories import (
    MEMORY_CATEGORIES,
    _load_all_memories,
    create_memory,
)

# Use devwisdom-go MCP server instead of direct import
from ..utils.wisdom_client import consult_advisor

# METRIC_ADVISORS no longer needed (handled by external server)

logger = logging.getLogger(__name__)


# ═══════════════════════════════════════════════════════════════════════════════
# DREAM ADVISORS
# ═══════════════════════════════════════════════════════════════════════════════

# Advisors specialized for dreaming/reflection
DREAM_ADVISORS = {
    "chacham": {
        "name": "Chacham (The Sage)",
        "icon": "📜",
        "focus": "Deep pattern analysis and wisdom extraction",
        "analyzes": ["insight", "preference"],
    },
    "confucius": {
        "name": "Confucius",
        "icon": "🎓",
        "focus": "Learning and knowledge transmission",
        "analyzes": ["research", "architecture"],
    },
    "enochian": {
        "name": "Enochian Mysteries",
        "icon": "🔮",
        "focus": "Hidden patterns and architecture",
        "analyzes": ["architecture", "debug"],
    },
    "stoic": {
        "name": "Stoic Philosophy",
        "icon": "🏛️",
        "focus": "Debug patterns and persistence",
        "analyzes": ["debug"],
    },
    "tao": {
        "name": "Tao",
        "icon": "☯️",
        "focus": "Balance and flow of work",
        "analyzes": ["insight", "preference"],
    },
}


# ═══════════════════════════════════════════════════════════════════════════════
# PATTERN DETECTION
# ═══════════════════════════════════════════════════════════════════════════════


def _detect_patterns(memories: list[dict[str, Any]]) -> dict[str, Any]:
    """
    Analyze memories to detect patterns.

    Returns dict with:
    - recurring_topics: Topics that appear multiple times
    - blockers: Mentions of blocking issues
    - solutions: Debug solutions that were found
    - architecture_insights: Architecture-related discoveries
    """
    patterns = {
        "recurring_topics": defaultdict(int),
        "blockers": [],
        "solutions": [],
        "architecture_insights": [],
        "preferences_learned": [],
    }

    # Common blocker words
    blocker_words = ["block", "stuck", "fail", "error", "broken", "cannot", "unable"]

    for memory in memories:
        title = memory.get("title", "").lower()
        content = memory.get("content", "").lower()
        category = memory.get("category", "")

        # Extract topics (words from title)
        for word in title.split():
            if len(word) > 3 and word.isalpha():
                patterns["recurring_topics"][word] += 1

        # Detect blockers
        for word in blocker_words:
            if word in content or word in title:
                patterns["blockers"].append({
                    "title": memory.get("title", ""),
                    "id": memory.get("id", ""),
                    "keyword": word,
                })
                break

        # Categorize by type
        if category == "debug":
            patterns["solutions"].append({
                "title": memory.get("title", ""),
                "id": memory.get("id", ""),
                "summary": content[:100] + "..." if len(content) > 100 else content,
            })
        elif category == "architecture":
            patterns["architecture_insights"].append({
                "title": memory.get("title", ""),
                "id": memory.get("id", ""),
            })
        elif category == "preference":
            patterns["preferences_learned"].append({
                "title": memory.get("title", ""),
                "id": memory.get("id", ""),
            })

    # Filter to truly recurring topics (3+ occurrences)
    patterns["recurring_topics"] = {
        k: v for k, v in patterns["recurring_topics"].items() if v >= 3
    }

    return patterns


def _generate_dream_narrative(
    patterns: dict[str, Any],
    memories: list[dict[str, Any]],
    advisor_reflections: list[dict[str, Any]],
) -> str:
    """Generate a narrative summary of the dream."""
    parts = []

    # Opening
    parts.append("🌙 **Memory Dream Summary**\n")
    parts.append(f"Reflecting on {len(memories)} memories...\n")

    # Recurring themes
    if patterns["recurring_topics"]:
        topics = sorted(patterns["recurring_topics"].items(), key=lambda x: -x[1])[:5]
        parts.append("\n📌 **Recurring Themes:**")
        for topic, count in topics:
            parts.append(f"  - {topic}: mentioned {count} times")

    # Blockers encountered
    if patterns["blockers"]:
        parts.append(f"\n🚧 **Blockers Encountered:** {len(patterns['blockers'])}")
        for blocker in patterns["blockers"][:3]:
            parts.append(f"  - {blocker['title']}")

    # Solutions found
    if patterns["solutions"]:
        parts.append(f"\n✅ **Solutions Discovered:** {len(patterns['solutions'])}")
        for solution in patterns["solutions"][:3]:
            parts.append(f"  - {solution['title']}")

    # Architecture insights
    if patterns["architecture_insights"]:
        parts.append(f"\n🏗️ **Architecture Insights:** {len(patterns['architecture_insights'])}")

    # Advisor wisdom
    if advisor_reflections:
        parts.append("\n\n🧙‍♂️ **Advisor Wisdom:**")
        for reflection in advisor_reflections[:3]:
            parts.append(f"\n  {reflection.get('advisor_icon', '📜')} **{reflection.get('advisor_name', 'Advisor')}:**")
            parts.append(f"  \"{reflection.get('quote', '')}\"")
            parts.append(f"  — {reflection.get('quote_source', '')}")

    return "\n".join(parts)


# ═══════════════════════════════════════════════════════════════════════════════
# MAIN DREAMING FUNCTION
# ═══════════════════════════════════════════════════════════════════════════════


def memory_dream(
    scope: str = "week",
    advisors: Optional[List[str]] = None,
    generate_insights: bool = True,
    save_dream: bool = True,
) -> dict[str, Any]:
    """
    Dream on memories - reflect with advisor wisdom.

    This process:
    1. Loads memories from the specified time scope
    2. Analyzes patterns across memories
    3. Consults relevant advisors for wisdom
    4. Generates consolidated insights
    5. Optionally saves the dream as a new memory

    Args:
        scope: "day", "week", "month", or "all" (default "week")
        advisors: Specific advisors to consult, or None for auto-select
        generate_insights: Whether to extract actionable insights
        save_dream: Whether to save the dream as a new memory

    Returns:
        Dream results with patterns, advisor wisdom, and insights
    """
    # Load memories based on scope
    all_memories = _load_all_memories()

    scope_days = {
        "day": 1,
        "week": 7,
        "month": 30,
        "all": 365 * 10,  # Effectively all
    }

    days = scope_days.get(scope, 7)
    cutoff = (datetime.now() - timedelta(days=days)).isoformat()

    memories = [m for m in all_memories if m.get("created_at", "") >= cutoff]

    if not memories:
        return {
            "status": "no_memories",
            "message": f"No memories found in the last {days} day(s)",
            "scope": scope,
            "patterns": {},
            "reflections": [],
            "insights": [],
        }

    # Detect patterns
    patterns = _detect_patterns(memories)

    # Group memories by category for analysis
    by_category: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for m in memories:
        by_category[m.get("category", "unknown")].append(m)

    # Select advisors
    if advisors is None:
        # Auto-select based on memory categories present
        selected_advisors = []

        # Always include Chacham for wisdom
        selected_advisors.append("chacham")

        # Add category-specific advisors
        if by_category.get("debug"):
            selected_advisors.append("stoic")
        if by_category.get("architecture"):
            selected_advisors.append("enochian")
        if by_category.get("research"):
            selected_advisors.append("confucius")
        if by_category.get("insight") or by_category.get("preference"):
            selected_advisors.append("tao")

        advisors = list(set(selected_advisors))[:4]  # Max 4 advisors

    # Consult advisors
    reflections = []
    for advisor_key in advisors:
        advisor_info = DREAM_ADVISORS.get(advisor_key)
        if not advisor_info:
            continue

        # Find a relevant metric for this advisor from devwisdom-go MCP server
        metric = None
        try:
            from ..utils.project_root import find_project_root
            from ..utils.wisdom_client import read_wisdom_resource_sync
            project_root = find_project_root()
            advisors_json = read_wisdom_resource_sync("wisdom://advisors", project_root)
            if advisors_json:
                import json
                advisors_data = json.loads(advisors_json) if isinstance(advisors_json, str) else advisors_json
                METRIC_ADVISORS = advisors_data.get("by_metric", {})
                for m, info in METRIC_ADVISORS.items():
                    if info.get("advisor") == advisor_key:
                        metric = m
                        break
        except Exception as e:
            logger.warning(f"Failed to get METRIC_ADVISORS from devwisdom-go MCP server: {e}")
            metric = None

        # Calculate a "score" based on memory health
        score = min(100, len(memories) * 5)  # More memories = higher score

        consultation_result = consult_advisor(
            metric=metric,
            stage="retrospective",  # Dreaming is like a retrospective
            score=score,
            context=f"Dreaming on {len(memories)} memories from last {days} days",
        )

        # consult_advisor now returns dict (from MCP server), convert to JSON string if needed
        if isinstance(consultation_result, dict):
            import json
            consultation_json = json.dumps(consultation_result, indent=2)
        else:
            consultation_json = consultation_result or "{}"

        # Parse JSON string to dict
        import json
        consultation = json.loads(consultation_json)

        reflections.append({
            **consultation,
            "dream_focus": advisor_info.get("focus", ""),
            "categories_analyzed": advisor_info.get("analyzes", []),
        })

    # Generate insights
    insights = []
    if generate_insights:
        # Insight from recurring topics
        if patterns["recurring_topics"]:
            top_topics = sorted(patterns["recurring_topics"].items(), key=lambda x: -x[1])[:3]
            insights.append({
                "type": "recurring_theme",
                "content": f"Focus areas: {', '.join(t[0] for t in top_topics)}",
                "evidence_count": sum(t[1] for t in top_topics),
            })

        # Insight from blockers
        if patterns["blockers"]:
            insights.append({
                "type": "blocker_pattern",
                "content": f"Encountered {len(patterns['blockers'])} blockers - consider preventive measures",
                "evidence_count": len(patterns["blockers"]),
            })

        # Insight from solutions
        if len(patterns["solutions"]) >= 3:
            insights.append({
                "type": "solution_bank",
                "content": f"Built up {len(patterns['solutions'])} debug solutions - valuable knowledge base",
                "evidence_count": len(patterns["solutions"]),
            })

        # Insight from architecture
        if patterns["architecture_insights"]:
            insights.append({
                "type": "architecture_growth",
                "content": f"Documented {len(patterns['architecture_insights'])} architecture insights",
                "evidence_count": len(patterns["architecture_insights"]),
            })

    # Generate narrative
    narrative = _generate_dream_narrative(patterns, memories, reflections)

    # Save dream as memory
    dream_memory_id = None
    if save_dream:
        dream_memory = create_memory(
            title=f"Dream: {scope.title()} Reflection ({datetime.now().strftime('%Y-%m-%d')})",
            content=narrative,
            category="insight",
            metadata={
                "type": "dream",
                "scope": scope,
                "memories_analyzed": len(memories),
                "advisors_consulted": advisors,
                "patterns_found": {
                    "recurring_topics": len(patterns["recurring_topics"]),
                    "blockers": len(patterns["blockers"]),
                    "solutions": len(patterns["solutions"]),
                },
            },
        )
        dream_memory_id = dream_memory["id"]

    return {
        "status": "success",
        "scope": scope,
        "days_analyzed": days,
        "memories_analyzed": len(memories),
        "memories_by_category": {k: len(v) for k, v in by_category.items()},
        "patterns": {
            "recurring_topics": dict(patterns["recurring_topics"]),
            "blockers_found": len(patterns["blockers"]),
            "solutions_found": len(patterns["solutions"]),
            "architecture_insights": len(patterns["architecture_insights"]),
            "preferences_learned": len(patterns["preferences_learned"]),
        },
        "advisors_consulted": advisors,
        "reflections": [
            {
                "advisor": r.get("advisor_name", ""),
                "icon": r.get("advisor_icon", ""),
                "quote": r.get("quote", ""),
                "source": r.get("quote_source", ""),
                "encouragement": r.get("encouragement", ""),
            }
            for r in reflections
        ],
        "insights": insights,
        "narrative": narrative,
        "dream_saved": save_dream,
        "dream_memory_id": dream_memory_id,
    }


def dream_with_focus(
    focus_category: str,
    scope: str = "month",
    save_dream: bool = True,
) -> dict[str, Any]:
    """
    Dream with focus on a specific memory category.

    Args:
        focus_category: Category to focus on (debug, research, architecture, preference, insight)
        scope: Time scope for memories
        save_dream: Whether to save the dream

    Returns:
        Focused dream results
    """
    if focus_category not in MEMORY_CATEGORIES:
        return {
            "status": "error",
            "error": f"Invalid category: {focus_category}. Use one of: {', '.join(MEMORY_CATEGORIES)}",
        }

    # Select advisor based on category
    category_advisors = {
        "debug": ["stoic", "enochian"],
        "research": ["confucius", "chacham"],
        "architecture": ["enochian", "tao"],
        "preference": ["chacham", "tao"],
        "insight": ["chacham", "confucius"],
    }

    return memory_dream(
        scope=scope,
        advisors=category_advisors.get(focus_category, ["chacham"]),
        generate_insights=True,
        save_dream=save_dream,
    )


__all__ = [
    "memory_dream",
    "dream_with_focus",
    "DREAM_ADVISORS",
]







