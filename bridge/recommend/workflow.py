"""
Workflow Mode Recommender

Recommends AGENT vs ASK mode based on task complexity.
"""

import json
import logging
import re
import time
from typing import Optional

logger = logging.getLogger(__name__)


# Mode suggestion messages
MODE_SUGGESTIONS = {
    "AGENT": {
        "emoji": "🤖",
        "action": "Switch to AGENT mode",
        "instruction": "Click the mode selector (top of chat) → Select 'Agent'",
        "why": "AGENT mode enables autonomous multi-file editing with automatic tool execution",
        "benefits": [
            "Autonomous file creation and modification",
            "Multi-step task execution",
            "Automatic tool calls without confirmation",
            "Better for implementation tasks",
        ],
    },
    "ASK": {
        "emoji": "💬",
        "action": "Switch to ASK mode",
        "instruction": "Click the mode selector (top of chat) → Select 'Ask'",
        "why": "ASK mode provides focused assistance with user control over changes",
        "benefits": [
            "User confirms each change",
            "Better for learning and understanding",
            "Safer for critical code review",
            "Ideal for questions and explanations",
        ],
    },
}

# Complexity indicators for AGENT mode
AGENT_INDICATORS = {
    "keywords": [
        "implement", "create", "build", "develop", "refactor",
        "migrate", "upgrade", "integrate", "deploy", "configure",
        "setup", "install", "automate", "generate", "scaffold",
    ],
    "patterns": [
        r"multi.?file",
        r"cross.?module",
        r"end.?to.?end",
        r"full.?stack",
        r"complete\s+\w+",
    ],
    "tags": [
        "feature", "implementation", "infrastructure", "integration",
        "refactoring", "migration", "automation",
    ],
}

# Simplicity indicators for ASK mode
ASK_INDICATORS = {
    "keywords": [
        "explain", "what", "why", "how", "understand",
        "clarify", "review", "check", "validate", "analyze",
        "debug", "find", "locate", "show", "list",
    ],
    "patterns": [
        r"single\s+file",
        r"quick\s+\w+",
        r"simple\s+\w+",
        r"just\s+\w+",
    ],
    "tags": [
        "question", "documentation", "review", "analysis",
        "debugging", "research",
    ],
}


def recommend_workflow_mode(
    task_description: Optional[str] = None,
    tags: Optional[str] = None,  # JSON string of tags
    include_rationale: bool = True,
) -> str:
    """
    Recommend AGENT vs ASK mode based on task complexity.

    Args:
        task_description: Description of the task
        tags: Optional JSON string of tags to consider
        include_rationale: Whether to include detailed reasoning

    Returns:
        JSON string with mode recommendation
    """
    try:
        content = task_description or ""
        content_lower = content.lower()

        # Parse tags if provided
        tag_list = []
        if tags:
            try:
                tag_list = json.loads(tags) if isinstance(tags, str) else tags
            except json.JSONDecodeError:
                pass

        # Score AGENT indicators
        agent_score = 0
        agent_reasons = []

        for kw in AGENT_INDICATORS["keywords"]:
            if kw in content_lower:
                agent_score += 2
                agent_reasons.append(f"Keyword: '{kw}'")

        for pattern in AGENT_INDICATORS["patterns"]:
            if re.search(pattern, content_lower):
                agent_score += 3
                agent_reasons.append(f"Pattern: '{pattern}'")

        for tag in tag_list:
            if isinstance(tag, str) and tag.lower() in AGENT_INDICATORS["tags"]:
                agent_score += 2
                agent_reasons.append(f"Tag: '{tag}'")

        # Score ASK indicators
        ask_score = 0
        ask_reasons = []

        for kw in ASK_INDICATORS["keywords"]:
            if kw in content_lower:
                ask_score += 2
                ask_reasons.append(f"Keyword: '{kw}'")

        for pattern in ASK_INDICATORS["patterns"]:
            if re.search(pattern, content_lower):
                ask_score += 3
                ask_reasons.append(f"Pattern: '{pattern}'")

        for tag in tag_list:
            if isinstance(tag, str) and tag.lower() in ASK_INDICATORS["tags"]:
                ask_score += 2
                ask_reasons.append(f"Tag: '{tag}'")

        # Determine recommendation
        if agent_score > ask_score:
            mode = "AGENT"
            confidence = min(agent_score / (agent_score + ask_score + 1) * 100, 95)
            reasons = agent_reasons
            description = "Use AGENT mode for autonomous multi-step implementation"
        elif ask_score > agent_score:
            mode = "ASK"
            confidence = min(ask_score / (agent_score + ask_score + 1) * 100, 95)
            reasons = ask_reasons
            description = "Use ASK mode for focused questions and single edits"
        else:
            mode = "ASK"  # Default to ASK when uncertain
            confidence = 50
            reasons = ["No strong indicators - defaulting to ASK for safety"]
            description = "Unclear complexity - start with ASK, escalate to AGENT if needed"

        result = {
            "recommended_mode": mode,
            "confidence": round(confidence, 1),
            "description": description,
            "agent_score": agent_score,
            "ask_score": ask_score,
        }

        if include_rationale:
            result["rationale"] = reasons[:5]  # Top 5 reasons
            result["guidelines"] = {
                "AGENT": "Best for: Multi-file changes, feature implementation, refactoring, scaffolding",
                "ASK": "Best for: Questions, code review, single-file edits, debugging help",
            }

        # Add user-facing suggestion
        suggestion = MODE_SUGGESTIONS[mode]
        result["suggestion"] = {
            "message": _generate_mode_suggestion_message(mode, confidence, content[:100] if content else ""),
            "action": suggestion["action"],
            "instruction": suggestion["instruction"],
            "benefits": suggestion["benefits"][:3],
        }

        return json.dumps({
            "success": True,
            "data": result,
            "timestamp": time.time(),
        }, indent=2)

    except Exception as e:
        logger.error(f"Error recommending workflow mode: {e}")
        return json.dumps({
            "success": False,
            "error": {"code": "AUTOMATION_ERROR", "message": str(e)},
        }, indent=2)


def _generate_mode_suggestion_message(
    mode: str,
    confidence: float,
    task_summary: str = "",
) -> str:
    """Generate a user-friendly mode suggestion message."""
    suggestion = MODE_SUGGESTIONS.get(mode, MODE_SUGGESTIONS["ASK"])
    emoji = suggestion["emoji"]

    # Build confidence qualifier
    if confidence >= 80:
        qualifier = "strongly recommend"
    elif confidence >= 60:
        qualifier = "recommend"
    else:
        qualifier = "suggest"

    # Build the message
    lines = [
        f"{emoji} **Mode Suggestion: {mode}**",
        "",
    ]

    if task_summary:
        lines.append(f"For this task ({task_summary[:50]}{'...' if len(task_summary) > 50 else ''}), I {qualifier} using **{mode}** mode.")
    else:
        lines.append(f"I {qualifier} using **{mode}** mode for this task.")

    lines.extend([
        "",
        f"**Why?** {suggestion['why']}",
        "",
        "**To switch:**",
        f"→ {suggestion['instruction']}",
    ])

    return "\n".join(lines)

