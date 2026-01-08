"""
Recommendations Tool

Unified recommendations tool consolidating model, workflow, and advisor recommendations.
"""

import json
import logging
from typing import Optional

logger = logging.getLogger(__name__)

from .model import recommend_model, list_available_models
from .workflow import recommend_workflow_mode


def recommend(
    action: str = "model",
    task_description: Optional[str] = None,
    task_type: Optional[str] = None,
    optimize_for: str = "quality",
    include_alternatives: bool = True,
    tags: Optional[str] = None,
    include_rationale: bool = True,
) -> str:
    """
    Unified recommendations tool.

    Args:
        action: "model" for AI model recommendations, "workflow" for mode suggestions
        task_description: Description of the task
        task_type: Optional explicit task type (model action)
        optimize_for: "quality", "speed", or "cost" (model action)
        include_alternatives: Include alternative recommendations (model action)
        tags: Optional JSON string of tags (workflow action)
        include_rationale: Whether to include detailed reasoning (workflow action)

    Returns:
        JSON string with recommendation results
    """
    if action == "model":
        return recommend_model(
            task_description=task_description,
            task_type=task_type,
            optimize_for=optimize_for,
            include_alternatives=include_alternatives,
        )
    elif action == "workflow":
        return recommend_workflow_mode(
            task_description=task_description,
            tags=tags,
            include_rationale=include_rationale,
        )
    elif action == "models":
        return list_available_models()
    else:
        return json.dumps({
            "status": "error",
            "error": f"Unknown recommend action: {action}. Use 'model', 'workflow', or 'models'.",
        }, indent=2)

