"""
Prompt Tracking Tool

Unified prompt tracking tool consolidating logging and analysis.
"""

import json
import logging
from typing import Optional

logger = logging.getLogger(__name__)

from .tracker import log_prompt_iteration, analyze_prompt_iterations


def prompt_tracking(
    action: str = "analyze",
    prompt: Optional[str] = None,
    task_id: Optional[str] = None,
    mode: Optional[str] = None,
    outcome: Optional[str] = None,
    iteration: int = 1,
    days: int = 7,
    project_root: Optional[str] = None,
) -> str:
    """
    Unified prompt tracking tool.

    Args:
        action: "log" to log a prompt, "analyze" to analyze patterns
        prompt: Prompt text to log (log action, required)
        task_id: Optional task ID (log action)
        mode: AGENT or ASK mode (log action)
        outcome: success/failed/partial (log action)
        iteration: Iteration number (log action)
        days: Days to analyze (analyze action)
        project_root: Optional project root path (auto-detected if not provided)

    Returns:
        JSON string with tracking results
    """
    if action == "log":
        if not prompt:
            return json.dumps({
                "status": "error",
                "error": "prompt parameter required for log action",
            }, indent=2)
        return log_prompt_iteration(
            prompt=prompt,
            task_id=task_id,
            mode=mode,
            outcome=outcome,
            iteration=iteration,
            project_root=project_root,
        )
    elif action == "analyze":
        return analyze_prompt_iterations(
            days=days,
            project_root=project_root,
        )
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown prompt_tracking action: {action}. Use 'log' or 'analyze'.",
        }, indent=2)

