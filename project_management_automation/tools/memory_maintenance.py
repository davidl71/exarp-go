"""
Memory Maintenance Tools for EXARP

Provides lifecycle management for AI session memories:
- Garbage Collection: Remove stale/orphaned memories
- Pruning: Remove low-value memories based on scoring
- Consolidation: Merge similar/duplicate memories

Trusted Advisor: 📜 Chacham (Wisdom)
"He who has wisdom has everything." - Memory is the foundation of learning.
"""

import logging
import os
from collections import defaultdict
from datetime import datetime, timedelta
from difflib import SequenceMatcher
from typing import Any

from ..resources.memories import (
    _get_memories_dir,
    _load_all_memories,
    _save_memory,
)

logger = logging.getLogger(__name__)


# ═══════════════════════════════════════════════════════════════════════════════
# HELPER FUNCTIONS
# ═══════════════════════════════════════════════════════════════════════════════


def _delete_memory(memory_id: str) -> bool:
    """Delete a memory file by ID."""
    memories_dir = _get_memories_dir()
    memory_file = memories_dir / f"{memory_id}.json"

    if memory_file.exists():
        try:
            os.remove(memory_file)
            logger.info(f"Deleted memory: {memory_id}")
            return True
        except OSError as e:
            logger.error(f"Failed to delete memory {memory_id}: {e}")
            return False
    return False


def _get_task_ids() -> set[str]:
    """Get all valid task IDs from Todo2 system."""
    try:
        from ..resources.tasks import load_tasks
        tasks = load_tasks()
        return {t.get("id") for t in tasks if t.get("id")}
    except Exception as e:
        logger.warning(f"Could not load tasks: {e}")
        return set()


def _calculate_memory_value(memory: dict[str, Any]) -> float:
    """
    Calculate a value score for a memory (0.0 - 1.0).

    Scoring factors:
    - Has linked tasks: +0.2
    - Category weight: debug/architecture=0.3, research=0.2, insight/preference=0.1
    - Recency (< 7 days): +0.2
    - Content length (normalized): up to +0.2
    """
    score = 0.0

    # Linked tasks
    if memory.get("linked_tasks"):
        score += 0.2

    # Category weight
    category_weights = {
        "debug": 0.3,
        "architecture": 0.3,
        "research": 0.2,
        "insight": 0.1,
        "preference": 0.1,
    }
    category = memory.get("category", "insight")
    score += category_weights.get(category, 0.1)

    # Recency boost
    created_at = memory.get("created_at", "")
    if created_at:
        try:
            created = datetime.fromisoformat(created_at)
            if datetime.now() - created < timedelta(days=7):
                score += 0.2
        except ValueError:
            pass

    # Content length (normalized, max +0.2 for >500 chars)
    content_len = len(memory.get("content", ""))
    content_score = min(content_len / 500.0, 1.0) * 0.2
    score += content_score

    return min(score, 1.0)


def _similarity_ratio(a: str, b: str) -> float:
    """Calculate similarity ratio between two strings (0.0 - 1.0)."""
    return SequenceMatcher(None, a.lower(), b.lower()).ratio()


def _find_duplicate_groups(
    memories: list[dict[str, Any]],
    similarity_threshold: float = 0.85,
) -> list[list[dict[str, Any]]]:
    """
    Find groups of similar memories based on title similarity.

    Returns list of groups, where each group has 2+ similar memories.
    """
    # Group by category first
    by_category: dict[str, list[dict[str, Any]]] = defaultdict(list)
    for m in memories:
        by_category[m.get("category", "unknown")].append(m)

    all_groups = []

    for _category, cat_memories in by_category.items():
        # Track which memories have been grouped
        grouped = set()

        for i, m1 in enumerate(cat_memories):
            if m1["id"] in grouped:
                continue

            group = [m1]
            grouped.add(m1["id"])

            for _j, m2 in enumerate(cat_memories[i + 1:], start=i + 1):
                if m2["id"] in grouped:
                    continue

                # Compare titles
                title_sim = _similarity_ratio(
                    m1.get("title", ""),
                    m2.get("title", ""),
                )

                if title_sim >= similarity_threshold:
                    group.append(m2)
                    grouped.add(m2["id"])

            if len(group) > 1:
                all_groups.append(group)

    return all_groups


def _merge_memories(
    memories: list[dict[str, Any]],
    strategy: str = "newest",
) -> dict[str, Any]:
    """
    Merge multiple similar memories into one.

    Strategies:
    - newest: Use newest memory as base, combine content
    - oldest: Use oldest memory as base, combine content
    - longest: Use memory with longest content as base
    """
    if not memories:
        return {}

    if len(memories) == 1:
        return memories[0]

    # Sort by strategy
    if strategy == "newest":
        base = max(memories, key=lambda m: m.get("created_at", ""))
    elif strategy == "oldest":
        base = min(memories, key=lambda m: m.get("created_at", ""))
    elif strategy == "longest":
        base = max(memories, key=lambda m: len(m.get("content", "")))
    else:
        base = memories[0]

    # Combine unique content from all memories
    all_content = [base.get("content", "")]
    for m in memories:
        if m["id"] != base["id"]:
            content = m.get("content", "")
            # Only add if meaningfully different
            if content and _similarity_ratio(content, all_content[0]) < 0.9:
                all_content.append(f"\n--- Merged from {m.get('title', 'untitled')} ---\n{content}")

    # Combine linked tasks
    all_tasks = set()
    for m in memories:
        all_tasks.update(m.get("linked_tasks", []))

    # Create merged memory
    merged = {
        "id": base["id"],
        "title": base.get("title", "Merged Memory"),
        "content": "\n".join(all_content),
        "category": base.get("category", "insight"),
        "linked_tasks": list(all_tasks),
        "metadata": {
            **base.get("metadata", {}),
            "merged_from": [m["id"] for m in memories if m["id"] != base["id"]],
            "merged_at": datetime.now().isoformat(),
        },
        "created_at": base.get("created_at"),
        "session_date": base.get("session_date"),
    }

    return merged


# ═══════════════════════════════════════════════════════════════════════════════
# GARBAGE COLLECTION
# ═══════════════════════════════════════════════════════════════════════════════


def memory_garbage_collect(
    max_age_days: int = 90,
    delete_orphaned: bool = True,
    delete_duplicates: bool = True,
    scorecard_max_age_days: int = 7,
    dry_run: bool = True,
) -> dict[str, Any]:
    """
    Garbage collect stale and invalid memories.

    Removes:
    - Memories older than max_age_days
    - Orphaned memories (linked to non-existent tasks)
    - Exact duplicate titles within same category
    - Auto-generated scorecard memories older than scorecard_max_age_days

    Args:
        max_age_days: Delete memories older than this (default 90)
        delete_orphaned: Delete memories with broken task links
        delete_duplicates: Delete exact title duplicates (keep newest)
        scorecard_max_age_days: Max age for scorecard-type memories (default 7)
        dry_run: Preview without deleting (default True)

    Returns:
        Summary of garbage collection with items found/deleted
    """
    memories = _load_all_memories()
    valid_task_ids = _get_task_ids() if delete_orphaned else set()

    to_delete: list[dict[str, Any]] = []
    reasons: dict[str, list[str]] = defaultdict(list)

    now = datetime.now()
    age_cutoff = (now - timedelta(days=max_age_days)).isoformat()
    scorecard_cutoff = (now - timedelta(days=scorecard_max_age_days)).isoformat()

    # Track titles for duplicate detection
    seen_titles: dict[str, dict[str, Any]] = {}  # (category, title) -> newest memory

    for memory in memories:
        memory_id = memory.get("id", "")
        created_at = memory.get("created_at", "")
        title = memory.get("title", "").strip().lower()
        category = memory.get("category", "")
        linked_tasks = memory.get("linked_tasks", [])

        # Check age
        if created_at < age_cutoff:
            to_delete.append(memory)
            reasons[memory_id].append(f"older than {max_age_days} days")
            continue

        # Check scorecard memories (auto-generated)
        is_scorecard = "scorecard" in title.lower() or "project health" in title.lower()
        if is_scorecard and created_at < scorecard_cutoff:
            to_delete.append(memory)
            reasons[memory_id].append(f"scorecard memory older than {scorecard_max_age_days} days")
            continue

        # Check orphaned task links
        if delete_orphaned and linked_tasks and valid_task_ids:
            orphaned = [t for t in linked_tasks if t not in valid_task_ids]
            if orphaned and len(orphaned) == len(linked_tasks):
                # All linked tasks are invalid
                to_delete.append(memory)
                reasons[memory_id].append(f"all linked tasks orphaned: {orphaned}")
                continue

        # Check duplicates
        if delete_duplicates and title:
            key = (category, title)
            if key in seen_titles:
                existing = seen_titles[key]
                # Keep the newer one
                if created_at > existing.get("created_at", ""):
                    to_delete.append(existing)
                    reasons[existing["id"]].append(f"duplicate title, keeping newer: {memory_id}")
                    seen_titles[key] = memory
                else:
                    to_delete.append(memory)
                    reasons[memory_id].append(f"duplicate title, keeping: {existing['id']}")
            else:
                seen_titles[key] = memory

    # Remove duplicates from to_delete
    unique_to_delete = {m["id"]: m for m in to_delete}.values()

    # Execute deletion if not dry run
    deleted_ids = []
    if not dry_run:
        for memory in unique_to_delete:
            if _delete_memory(memory["id"]):
                deleted_ids.append(memory["id"])

    return {
        "status": "success",
        "dry_run": dry_run,
        "total_memories": len(memories),
        "garbage_found": len(unique_to_delete),
        "deleted_count": len(deleted_ids),
        "deletion_reasons": {
            mid: reasons[mid] for mid in [m["id"] for m in unique_to_delete]
        },
        "deleted_ids": deleted_ids,
        "items_to_delete": [
            {
                "id": m["id"],
                "title": m.get("title", ""),
                "category": m.get("category", ""),
                "created_at": m.get("created_at", ""),
                "reasons": reasons[m["id"]],
            }
            for m in unique_to_delete
        ][:20],  # Limit preview
        "recommendations": [
            f"Run with dry_run=False to delete {len(unique_to_delete)} memories"
        ] if dry_run and unique_to_delete else [],
    }


# ═══════════════════════════════════════════════════════════════════════════════
# PRUNING
# ═══════════════════════════════════════════════════════════════════════════════


def memory_prune(
    value_threshold: float = 0.3,
    keep_minimum: int = 50,
    dry_run: bool = True,
) -> dict[str, Any]:
    """
    Prune low-value memories based on scoring.

    Each memory is scored based on:
    - Linked tasks (+0.2)
    - Category (debug/architecture=0.3, research=0.2, insight/preference=0.1)
    - Recency (< 7 days = +0.2)
    - Content length (normalized, up to +0.2)

    Args:
        value_threshold: Minimum score to keep (0.0-1.0, default 0.3)
        keep_minimum: Always keep at least this many memories (default 50)
        dry_run: Preview without deleting (default True)

    Returns:
        Summary of pruning with scored memories
    """
    memories = _load_all_memories()

    # Score all memories
    scored: list[tuple[float, dict[str, Any]]] = []
    for memory in memories:
        score = _calculate_memory_value(memory)
        scored.append((score, memory))

    # Sort by score ascending (lowest first)
    scored.sort(key=lambda x: x[0])

    # Determine which to prune
    to_prune: list[tuple[float, dict[str, Any]]] = []
    kept = 0

    for score, memory in scored:
        if score < value_threshold and (len(memories) - len(to_prune)) > keep_minimum:
            to_prune.append((score, memory))
        else:
            kept += 1

    # Execute pruning if not dry run
    pruned_ids = []
    if not dry_run:
        for _score, memory in to_prune:
            if _delete_memory(memory["id"]):
                pruned_ids.append(memory["id"])

    # Score distribution for statistics
    score_distribution = {
        "0.0-0.2": sum(1 for s, _ in scored if s < 0.2),
        "0.2-0.4": sum(1 for s, _ in scored if 0.2 <= s < 0.4),
        "0.4-0.6": sum(1 for s, _ in scored if 0.4 <= s < 0.6),
        "0.6-0.8": sum(1 for s, _ in scored if 0.6 <= s < 0.8),
        "0.8-1.0": sum(1 for s, _ in scored if s >= 0.8),
    }

    return {
        "status": "success",
        "dry_run": dry_run,
        "total_memories": len(memories),
        "value_threshold": value_threshold,
        "keep_minimum": keep_minimum,
        "memories_to_prune": len(to_prune),
        "memories_kept": kept,
        "pruned_count": len(pruned_ids),
        "pruned_ids": pruned_ids,
        "score_distribution": score_distribution,
        "items_to_prune": [
            {
                "id": m["id"],
                "title": m.get("title", ""),
                "category": m.get("category", ""),
                "score": round(s, 3),
                "created_at": m.get("created_at", ""),
            }
            for s, m in to_prune[:20]  # Limit preview
        ],
        "recommendations": [
            f"Run with dry_run=False to prune {len(to_prune)} low-value memories"
        ] if dry_run and to_prune else [],
    }


# ═══════════════════════════════════════════════════════════════════════════════
# CONSOLIDATION
# ═══════════════════════════════════════════════════════════════════════════════


def memory_consolidate(
    similarity_threshold: float = 0.85,
    merge_strategy: str = "newest",
    dry_run: bool = True,
) -> dict[str, Any]:
    """
    Consolidate similar memories by merging them.

    Finds memories with similar titles (fuzzy match) within the same category
    and merges them into single, comprehensive memories.

    Args:
        similarity_threshold: Title similarity threshold (0.0-1.0, default 0.85)
        merge_strategy: "newest", "oldest", or "longest" (default "newest")
        dry_run: Preview without merging (default True)

    Returns:
        Summary of consolidation with groups found
    """
    memories = _load_all_memories()

    # Find duplicate groups
    groups = _find_duplicate_groups(memories, similarity_threshold)

    merged_count = 0
    deleted_ids = []
    merged_results = []

    for group in groups:
        if len(group) < 2:
            continue

        if not dry_run:
            # Merge the group
            merged = _merge_memories(group, merge_strategy)

            # Save the merged memory
            _save_memory(merged)

            # Delete the others
            for m in group:
                if m["id"] != merged["id"]:
                    if _delete_memory(m["id"]):
                        deleted_ids.append(m["id"])

            merged_count += 1

        merged_results.append({
            "group_size": len(group),
            "titles": [m.get("title", "") for m in group],
            "category": group[0].get("category", ""),
            "base_id": group[0]["id"],
            "merge_into": max(group, key=lambda m: m.get("created_at", ""))["id"],
        })

    return {
        "status": "success",
        "dry_run": dry_run,
        "total_memories": len(memories),
        "similarity_threshold": similarity_threshold,
        "merge_strategy": merge_strategy,
        "groups_found": len(groups),
        "total_duplicates": sum(len(g) for g in groups),
        "memories_mergeable": sum(len(g) - 1 for g in groups),  # One kept per group
        "merged_groups": merged_count,
        "deleted_count": len(deleted_ids),
        "deleted_ids": deleted_ids,
        "consolidation_groups": merged_results[:10],  # Limit preview
        "recommendations": [
            f"Run with dry_run=False to consolidate {len(groups)} groups of similar memories"
        ] if dry_run and groups else [],
    }


# ═══════════════════════════════════════════════════════════════════════════════
# HEALTH CHECK
# ═══════════════════════════════════════════════════════════════════════════════


def memory_health_check() -> dict[str, Any]:
    """
    Analyze memory system health and provide recommendations.

    Returns:
        Health metrics and maintenance recommendations
    """
    memories = _load_all_memories()
    valid_task_ids = _get_task_ids()

    now = datetime.now()

    # Age distribution
    age_distribution = {
        "last_24h": 0,
        "last_7d": 0,
        "last_30d": 0,
        "last_90d": 0,
        "older": 0,
    }

    # Category distribution
    category_counts = defaultdict(int)

    # Problem counts
    orphaned_count = 0
    duplicate_titles = set()
    seen_titles: dict[str, str] = {}  # (category, title) -> id
    low_value_count = 0
    stale_scorecard_count = 0

    scorecard_cutoff = (now - timedelta(days=7)).isoformat()

    for memory in memories:
        created_at = memory.get("created_at", "")
        title = memory.get("title", "").strip().lower()
        category = memory.get("category", "")
        linked_tasks = memory.get("linked_tasks", [])

        # Age
        if created_at:
            try:
                created = datetime.fromisoformat(created_at)
                age = now - created
                if age < timedelta(hours=24):
                    age_distribution["last_24h"] += 1
                elif age < timedelta(days=7):
                    age_distribution["last_7d"] += 1
                elif age < timedelta(days=30):
                    age_distribution["last_30d"] += 1
                elif age < timedelta(days=90):
                    age_distribution["last_90d"] += 1
                else:
                    age_distribution["older"] += 1
            except ValueError:
                age_distribution["older"] += 1

        # Category
        category_counts[category] += 1

        # Orphaned
        if linked_tasks and valid_task_ids:
            orphaned = [t for t in linked_tasks if t not in valid_task_ids]
            if len(orphaned) == len(linked_tasks):
                orphaned_count += 1

        # Duplicates
        key = (category, title)
        if key in seen_titles:
            duplicate_titles.add(title)
        else:
            seen_titles[key] = memory["id"]

        # Low value
        if _calculate_memory_value(memory) < 0.3:
            low_value_count += 1

        # Stale scorecards
        is_scorecard = "scorecard" in title.lower() or "project health" in title.lower()
        if is_scorecard and created_at < scorecard_cutoff:
            stale_scorecard_count += 1

    # Calculate overall health
    total = len(memories)
    issues = orphaned_count + len(duplicate_titles) + low_value_count + stale_scorecard_count
    health_score = max(0, 100 - (issues / max(total, 1) * 100))

    # Generate recommendations
    recommendations = []
    if orphaned_count > 0:
        recommendations.append(f"Run garbage collection: {orphaned_count} orphaned memories")
    if duplicate_titles:
        recommendations.append(f"Run consolidation: {len(duplicate_titles)} duplicate title groups")
    if low_value_count > total * 0.3:
        recommendations.append(f"Consider pruning: {low_value_count} low-value memories ({low_value_count/total*100:.0f}%)")
    if stale_scorecard_count > 5:
        recommendations.append(f"Clean up {stale_scorecard_count} old scorecard memories")
    if age_distribution["older"] > total * 0.2:
        recommendations.append(f"Consider archiving: {age_distribution['older']} memories older than 90 days")

    return {
        "total_memories": total,
        "health_score": round(health_score, 1),
        "by_category": dict(category_counts),
        "age_distribution": age_distribution,
        "issues": {
            "orphaned": orphaned_count,
            "duplicates": len(duplicate_titles),
            "low_value": low_value_count,
            "stale_scorecards": stale_scorecard_count,
        },
        "estimated_space": f"{total * 2:.1f} KB",  # Rough estimate
        "recommendations": recommendations,
        "timestamp": now.isoformat(),
    }


__all__ = [
    "memory_garbage_collect",
    "memory_prune",
    "memory_consolidate",
    "memory_health_check",
]







