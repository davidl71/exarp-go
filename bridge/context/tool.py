"""
Context Management Tool

Unified context management tool consolidating summarization, budgeting, and batch operations.
"""

import json
import logging
from typing import Optional

logger = logging.getLogger(__name__)

from .summarizer import summarize_context, batch_summarize, estimate_context_budget


def context(
    action: str = "summarize",
    data: Optional[str] = None,
    level: str = "brief",
    tool_type: Optional[str] = None,
    max_tokens: Optional[int] = None,
    include_raw: bool = False,
    items: Optional[str] = None,
    budget_tokens: int = 4000,
    combine: bool = True,
) -> str:
    """
    Unified context management tool.

    Args:
        action: "summarize" for single item, "budget" for token analysis, "batch" for multiple items
        data: JSON string to summarize (summarize action)
        level: Summarization level - "brief", "detailed", "key_metrics", "actionable"
        tool_type: Tool type hint for smarter summarization
        max_tokens: Maximum tokens for output
        include_raw: Include original data in response
        items: JSON array of items to analyze (budget/batch actions)
        budget_tokens: Target token budget (budget action)
        combine: Merge summaries into combined view (batch action)

    Returns:
        JSON string with context operation results
    """
    if action == "summarize":
        if not data:
            return json.dumps({
                "status": "error",
                "error": "data parameter required for summarize action",
            }, indent=2)
        result = summarize_context(data, level, tool_type, max_tokens, include_raw)
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "budget":
        if not items:
            return json.dumps({
                "status": "error",
                "error": "items parameter required for budget action",
            }, indent=2)
        parsed_items = json.loads(items) if isinstance(items, str) else items
        result = estimate_context_budget(parsed_items, budget_tokens)
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "batch":
        if not items:
            return json.dumps({
                "status": "error",
                "error": "items parameter required for batch action",
            }, indent=2)
        parsed_items = json.loads(items) if isinstance(items, str) else items
        result = batch_summarize(parsed_items, level, combine)
        return result if isinstance(result, str) else json.dumps(result, indent=2)

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown context action: {action}. Use 'summarize', 'budget', or 'batch'.",
        }, indent=2)

