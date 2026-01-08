"""
Context Summarization Tool

Strategically summarizes and reduces context for LLM interactions.
Compresses verbose tool outputs into key metrics while preserving actionable information.

Features:
- Multi-level summarization (brief, detailed, key_metrics)
- Tool-aware compression using hint patterns
- Batch summarization for multiple results
- Token estimation for context budgeting
"""

import json
import logging
import time
from typing import Any, Optional, Union, List, Dict

logger = logging.getLogger(__name__)


# ═══════════════════════════════════════════════════════════════════════════════
# SUMMARIZATION PATTERNS (Tool-specific extraction rules)
# ═══════════════════════════════════════════════════════════════════════════════

TOOL_PATTERNS = {
    # Health tools
    "health": {
        "key_fields": ["health_score", "score", "overall_score"],
        "count_fields": ["broken_links", "broken_internal", "broken_external", "stale_files", "format_errors"],
        "action_fields": ["tasks_created", "followup_tasks", "recommendations"],
        "brief_template": "Health: {score}/100, {issues} issues, {actions} actions",
    },
    "scorecard": {
        "key_fields": ["overall_score", "production_ready"],
        "count_fields": ["security_score", "testing_score", "documentation_score", "completion_score"],
        "action_fields": ["recommendations", "critical_issues"],
        "brief_template": "Score: {overall_score}/100, Production Ready: {production_ready}",
    },
    # Security tools
    "security": {
        "key_fields": ["status", "total_vulnerabilities"],
        "count_fields": ["critical", "high", "medium", "low", "vulnerabilities"],
        "action_fields": ["remediation", "recommendations", "fixes_available"],
        "brief_template": "Security: {critical} critical, {high} high, {medium} medium vulns",
    },
    # Task tools
    "task": {
        "key_fields": ["status", "tasks_analyzed", "total_tasks"],
        "count_fields": ["duplicates", "misaligned", "pending", "completed", "blocked"],
        "action_fields": ["tasks_created", "recommendations", "followup_tasks"],
        "brief_template": "Tasks: {total} total, {pending} pending, {actions} actions",
    },
    # Testing tools
    "testing": {
        "key_fields": ["status", "passed", "failed", "coverage"],
        "count_fields": ["total_tests", "skipped", "errors"],
        "action_fields": ["failures", "recommendations"],
        "brief_template": "Tests: {passed} passed, {failed} failed, {coverage}% coverage",
    },
    # Generic fallback
    "generic": {
        "key_fields": ["status", "success", "result"],
        "count_fields": ["count", "total", "found"],
        "action_fields": ["recommendations", "actions", "tasks"],
        "brief_template": "Result: {status}, {count} items",
    },
}

# Estimate tokens per character (rough approximation)
TOKENS_PER_CHAR = 0.25


def summarize_context(
    data: Union[str, Dict[str, Any], List[Any]],
    level: str = "brief",
    tool_type: Optional[str] = None,
    max_tokens: Optional[int] = None,
    include_raw: bool = False,
) -> str:
    """
    Strategically summarizes tool outputs for efficient context usage.

    Args:
        data: JSON string, dict, or list to summarize
        level: Summarization level - "brief", "detailed", "key_metrics", "actionable"
        tool_type: Tool type hint for smarter summarization
        max_tokens: Maximum tokens for output (truncates if needed)
        include_raw: Include original data in response

    Returns:
        JSON string with summarized content and metadata
    """
    start_time = time.time()

    try:
        # Parse input
        if isinstance(data, str):
            try:
                parsed = json.loads(data)
            except json.JSONDecodeError:
                parsed = {"text": data}
        else:
            parsed = data

        # Auto-detect tool type if not provided
        if not tool_type:
            tool_type = _detect_tool_type(parsed)

        # Get pattern for this tool type
        pattern = TOOL_PATTERNS.get(tool_type, TOOL_PATTERNS["generic"])

        # Extract based on level
        if level == "brief":
            summary = _extract_brief(parsed, pattern)
        elif level == "detailed":
            summary = _extract_detailed(parsed, pattern)
        elif level == "key_metrics":
            summary = _extract_key_metrics(parsed, pattern)
        elif level == "actionable":
            summary = _extract_actionable(parsed, pattern)
        else:
            summary = _extract_brief(parsed, pattern)

        # Calculate token estimates
        original_tokens = _estimate_tokens(json.dumps(parsed))
        summary_tokens = _estimate_tokens(json.dumps(summary) if isinstance(summary, dict) else summary)
        reduction = round((1 - summary_tokens / max(original_tokens, 1)) * 100, 1)

        # Truncate if max_tokens specified
        if max_tokens and summary_tokens > max_tokens:
            summary = _truncate_to_tokens(summary, max_tokens)
            summary_tokens = max_tokens

        result = {
            "summary": summary,
            "level": level,
            "tool_type": tool_type,
            "token_estimate": {
                "original": original_tokens,
                "summarized": summary_tokens,
                "reduction_percent": reduction,
            },
            "duration_ms": round((time.time() - start_time) * 1000, 2),
        }

        if include_raw:
            result["raw_data"] = parsed

        return json.dumps(result, indent=2)

    except Exception as e:
        logger.error(f"Summarization error: {e}")
        return json.dumps({
            "error": str(e),
            "summary": str(data)[:200] + "..." if len(str(data)) > 200 else str(data),
            "level": level,
        }, indent=2)


def batch_summarize(
    items: List[Dict[str, Any]],
    level: str = "brief",
    combine: bool = True,
) -> str:
    """Summarize multiple tool results efficiently."""
    summaries = []
    total_original = 0
    total_summarized = 0

    for item in items:
        data = item.get("data", item)
        tool_type = item.get("tool_type")

        result = json.loads(summarize_context(data, level=level, tool_type=tool_type))
        summaries.append(result)

        total_original += result.get("token_estimate", {}).get("original", 0)
        total_summarized += result.get("token_estimate", {}).get("summarized", 0)

    if combine:
        combined = {
            "combined_summary": [s.get("summary") for s in summaries],
            "total_items": len(summaries),
            "token_estimate": {
                "original": total_original,
                "summarized": total_summarized,
                "reduction_percent": round((1 - total_summarized / max(total_original, 1)) * 100, 1),
            },
        }
        return json.dumps(combined, indent=2)

    return json.dumps({"summaries": summaries}, indent=2)


def estimate_context_budget(
    items: List[Any],
    budget_tokens: int = 4000,
) -> str:
    """Estimate token usage and suggest context reduction strategy."""
    analysis = []
    total_tokens = 0

    for i, item in enumerate(items):
        item_str = json.dumps(item) if isinstance(item, (dict, list)) else str(item)
        tokens = _estimate_tokens(item_str)
        total_tokens += tokens

        analysis.append({
            "index": i,
            "tokens": tokens,
            "percent_of_budget": round(tokens / budget_tokens * 100, 1),
            "recommendation": _get_budget_recommendation(tokens, budget_tokens),
        })

    analysis.sort(key=lambda x: x["tokens"], reverse=True)

    over_budget = total_tokens > budget_tokens

    return json.dumps({
        "total_tokens": total_tokens,
        "budget_tokens": budget_tokens,
        "over_budget": over_budget,
        "reduction_needed": max(0, total_tokens - budget_tokens),
        "items": analysis,
        "strategy": _suggest_reduction_strategy(analysis, total_tokens, budget_tokens),
    }, indent=2)


# Helper functions (abbreviated - full implementation in original file)
def _detect_tool_type(data: dict) -> str:
    """Auto-detect tool type from data structure."""
    data_str = json.dumps(data).lower()
    if any(k in data_str for k in ["vulnerability", "cve", "security"]):
        return "security"
    if any(k in data_str for k in ["health_score", "broken_link"]):
        return "health"
    if any(k in data_str for k in ["overall_score", "scorecard"]):
        return "scorecard"
    if any(k in data_str for k in ["test", "passed", "failed"]):
        return "testing"
    if any(k in data_str for k in ["task", "todo"]):
        return "task"
    return "generic"


def _extract_brief(data: dict, pattern: dict) -> str:
    """Extract one-line brief summary."""
    values = {}
    for field in pattern["key_fields"]:
        val = _deep_get(data, field)
        if val is not None:
            values["score"] = val
            break
    
    issue_count = sum(
        len(val) if isinstance(val, list) else (val if isinstance(val, (int, float)) else 0)
        for field in pattern["count_fields"]
        for val in [_deep_get(data, field)]
        if val is not None
    )
    values["issues"] = issue_count
    values["count"] = issue_count
    
    try:
        return pattern["brief_template"].format(**values)
    except KeyError:
        key_items = [f"{k}: {v}" for k, v in values.items() if v and v != "N/A" and v != 0][:5]
        return ", ".join(key_items)


def _extract_detailed(data: dict, pattern: dict) -> dict:
    """Extract multi-line detailed summary."""
    result = {"key_metrics": {}, "counts": {}, "actions": []}
    for field in pattern["key_fields"]:
        val = _deep_get(data, field)
        if val is not None:
            result["key_metrics"][field] = val
    for field in pattern["count_fields"]:
        val = _deep_get(data, field)
        if val is not None:
            result["counts"][field] = len(val) if isinstance(val, list) else val
    return result


def _extract_key_metrics(data: dict, pattern: dict) -> dict:
    """Extract only numerical metrics."""
    metrics = {}
    for field in pattern["key_fields"] + pattern["count_fields"]:
        val = _deep_get(data, field)
        if isinstance(val, (int, float)):
            metrics[field] = val
        elif isinstance(val, list):
            metrics[f"{field}_count"] = len(val)
    return metrics


def _extract_actionable(data: dict, pattern: dict) -> dict:
    """Extract only actionable items."""
    actions = {"recommendations": [], "tasks": [], "fixes": []}
    for field in pattern["action_fields"]:
        val = _deep_get(data, field)
        if isinstance(val, list):
            if "recommend" in field.lower():
                actions["recommendations"].extend(val[:3])
            elif "task" in field.lower():
                actions["tasks"].extend(val[:3])
            else:
                actions["fixes"].extend(val[:3])
    return {k: v for k, v in actions.items() if v}


def _deep_get(data: dict, key: str, default=None) -> Any:
    """Get value from nested dict, searching recursively."""
    if isinstance(data, dict):
        if key in data:
            return data[key]
        for v in data.values():
            if isinstance(v, dict):
                result = _deep_get(v, key, None)
                if result is not None:
                    return result
    return default


def _estimate_tokens(text: str) -> int:
    """Estimate token count for text."""
    return int(len(text) * TOKENS_PER_CHAR)


def _truncate_to_tokens(data: Any, max_tokens: int) -> Any:
    """Truncate data to fit within token limit."""
    data_str = json.dumps(data) if isinstance(data, (dict, list)) else str(data)
    max_chars = int(max_tokens / TOKENS_PER_CHAR)
    if len(data_str) <= max_chars:
        return data
    return data_str[:max_chars - 20] + "... [truncated]"


def _get_budget_recommendation(tokens: int, budget: int) -> str:
    """Get recommendation for a single item."""
    ratio = tokens / budget
    if ratio > 0.5:
        return "summarize_brief"
    elif ratio > 0.25:
        return "summarize_key_metrics"
    elif ratio > 0.1:
        return "keep_detailed"
    else:
        return "keep_full"


def _suggest_reduction_strategy(analysis: List[Dict[str, Any]], total: int, budget: int) -> str:
    """Suggest overall reduction strategy."""
    if total <= budget:
        return "Within budget - no reduction needed"
    reduction_needed = total - budget
    to_summarize = [a for a in analysis if a["recommendation"].startswith("summarize")]
    if not to_summarize:
        return f"Reduce largest items to fit. Need to remove ~{reduction_needed} tokens."
    return f"Summarize {len(to_summarize)} items using 'brief' level. Estimated savings: {sum(a['tokens'] * 0.7 for a in to_summarize):.0f} tokens."

