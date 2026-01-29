"""
Context Management Tool

Context tool is implemented in native Go only (bridge/context removed 2026-01-29).
This module returns a clear message for direct Python callers; use exarp-go for context operations.
"""

import json
from typing import Optional


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
    Context tool is native Go only. Use exarp-go -tool context for summarize/budget/batch.
    """
    return json.dumps({
        "status": "error",
        "error": "Context tool is native Go only. Use exarp-go -tool context (action=summarize|budget|batch).",
        "native_only": True,
    }, indent=2)
