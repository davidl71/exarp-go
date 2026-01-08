"""
Context Management Tool

Unified context management tool consolidating summarization, budgeting, and batch operations.
Extracted from consolidated.py for better modularity and to test FastMCP static analysis.
"""

import json
import logging
from typing import Optional, cast

logger = logging.getLogger(__name__)


def context(
    action: str = "summarize",
    # summarize action params
    data: Optional[str] = None,
    level: str = "brief",
    tool_type: Optional[str] = None,
    max_tokens: Optional[int] = None,
    include_raw: bool = False,
    # budget action params
    items: Optional[str] = None,
    budget_tokens: int = 4000,
    # batch action params
    combine: bool = True,
) -> str:
    """
    Unified context management tool.

    Consolidates context summarization, budgeting, and batch operations.

    Args:
        action: "summarize" for single item, "budget" for token analysis, "batch" for multiple items
        data: JSON string to summarize (summarize action)
        level: Summarization level - "brief", "detailed", "key_metrics", "actionable" (summarize action)
        tool_type: Tool type hint for smarter summarization (summarize action)
        max_tokens: Maximum tokens for output (summarize action)
        include_raw: Include original data in response (summarize action)
        items: JSON array of items to analyze (budget/batch actions)
        budget_tokens: Target token budget (budget action)
        combine: Merge summaries into combined view (batch action)

    Returns:
        JSON string with context operation results (FastMCP requires strings)
    """
    if action == "summarize":
        if not data:
            return json.dumps({
                "status": "error",
                "error": "data parameter required for summarize action",
            }, indent=2)
        from .context_summarizer import summarize_context
        result = summarize_context(data, level, tool_type, max_tokens, include_raw)
        # Explicit type cast to help static analysis understand this is always str
        return cast(str, result) if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "budget":
        if not items:
            return json.dumps({
                "status": "error",
                "error": "items parameter required for budget action",
            }, indent=2)
        import json as json_lib
        parsed_items = json_lib.loads(items) if isinstance(items, str) else items
        from .context_summarizer import estimate_context_budget
        result = estimate_context_budget(parsed_items, budget_tokens)
        # Explicit type cast to help static analysis understand this is always str
        return cast(str, result) if isinstance(result, str) else json.dumps(result, indent=2)

    elif action == "batch":
        if not items:
            return json.dumps({
                "status": "error",
                "error": "items parameter required for batch action",
            }, indent=2)
        import json as json_lib
        parsed_items = json_lib.loads(items) if isinstance(items, str) else items
        from .context_summarizer import batch_summarize
        result = batch_summarize(parsed_items, level, combine)
        # Explicit type cast to help static analysis understand this is always str
        return cast(str, result) if isinstance(result, str) else json.dumps(result, indent=2)

    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown context action: {action}. Use 'summarize', 'budget', or 'batch'.",
        }, indent=2)

