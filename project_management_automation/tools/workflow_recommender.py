"""
Workflow Mode Recommender Tool

Recommends AGENT vs ASK mode based on task complexity and type.
Based on Cursor IDE Best Practice #3.

Provides user-facing suggestions for mode changes since MCP cannot
programmatically change Cursor's mode.
"""

import json
import logging
import re
import time
from typing import Any, Optional

logger = logging.getLogger(__name__)


# ═══════════════════════════════════════════════════════════════════════════════
# MODE SUGGESTION MESSAGES (User-facing)
# ═══════════════════════════════════════════════════════════════════════════════

MODE_SUGGESTIONS = {
    "AGENT": {
        "emoji": "🤖",
        "action": "Switch to AGENT mode",
        "instruction": "Click the mode selector (top of chat) → Select 'Agent'",
        "keyboard": "No keyboard shortcut available",
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
        "keyboard": "No keyboard shortcut available",
        "why": "ASK mode provides focused assistance with user control over changes",
        "benefits": [
            "User confirms each change",
            "Better for learning and understanding",
            "Safer for critical code review",
            "Ideal for questions and explanations",
        ],
    },
}

# Current mode detection hints (patterns in responses that suggest current mode)
CURRENT_MODE_HINTS = {
    "AGENT": ["I'll make these changes", "Creating file", "Modifying", "I've updated"],
    "ASK": ["Would you like me to", "Should I", "Do you want", "Here's what I suggest"],
}

# Import error handler
try:
    from ..error_handler import (
        ErrorCode,
        format_error_response,
        format_success_response,
        log_automation_execution,
    )
except ImportError:

    def format_success_response(data, message=None):
        return {"success": True, "data": data, "timestamp": time.time()}

    def format_error_response(error, error_code, include_traceback=False):
        return {"success": False, "error": {"code": str(error_code), "message": str(error)}}

    def log_automation_execution(name, duration, success, error=None):
        logger.info(f"{name}: {duration:.2f}s, success={success}")

    class ErrorCode:
        AUTOMATION_ERROR = "AUTOMATION_ERROR"


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
    task_id: Optional[str] = None,
    include_rationale: bool = True,
) -> str:
    """
    [HINT: Workflow mode. AGENT vs ASK recommendation based on task complexity.]

    📊 Output: Recommended mode (AGENT/ASK), confidence, rationale
    🔧 Side Effects: None (advisory only)
    📁 Analyzes: Task description, tags, complexity indicators
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "Should I use AGENT or ASK mode for implementing user authentication?"

    Guidelines from Cursor Best Practices:
    - AGENT mode: Multi-file changes, implementation, refactoring
    - ASK mode: Questions, single-file edits, code review

    Args:
        task_description: Description of the task (or natural language query)
        task_id: Optional Todo2 task ID to analyze
        include_rationale: Whether to include detailed reasoning

    Returns:
        JSON with mode recommendation
    """
    start_time = time.time()

    try:
        content = task_description or ""
        tags = []

        # If task_id provided, load from Todo2
        if task_id and not task_description:
            from project_management_automation.utils import find_project_root

            project_root = find_project_root()
            todo2_path = project_root / ".todo2" / "state.todo2.json"

            if todo2_path.exists():
                state = json.loads(todo2_path.read_text())
                task = next(
                    (t for t in state.get("todos", []) if t.get("id") == task_id),
                    None,
                )
                if task:
                    content = f"{task.get('name', '')} {task.get('long_description', '')}"
                    tags = task.get("tags", [])

        content_lower = content.lower()

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

        for tag in tags:
            if tag.lower() in AGENT_INDICATORS["tags"]:
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

        for tag in tags:
            if tag.lower() in ASK_INDICATORS["tags"]:
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
            "message": generate_mode_suggestion_message(mode, confidence, content[:100] if content else ""),
            "action": suggestion["action"],
            "instruction": suggestion["instruction"],
            "benefits": suggestion["benefits"][:3],
        }

        duration = time.time() - start_time
        log_automation_execution("recommend_workflow_mode", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("recommend_workflow_mode", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


# ═══════════════════════════════════════════════════════════════════════════════
# MODE SUGGESTION HELPERS
# ═══════════════════════════════════════════════════════════════════════════════


def generate_mode_suggestion_message(
    mode: str,
    confidence: float,
    task_summary: str = "",
) -> str:
    """
    Generate a user-friendly mode suggestion message.

    This message is designed to be displayed to the user by the AI agent,
    since MCP cannot programmatically change Cursor's mode.

    Args:
        mode: Recommended mode (AGENT or ASK)
        confidence: Confidence level (0-100)
        task_summary: Brief task description for context

    Returns:
        Formatted suggestion message
    """
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


def get_mode_suggestion_for_task(
    task_description: str,
    current_mode: Optional[str] = None,
) -> dict[str, Any]:
    """
    Quick helper to get mode suggestion for a task.

    Can be called by other tools to check if mode change is recommended.

    Args:
        task_description: What the user wants to do
        current_mode: Current mode if known (AGENT or ASK)

    Returns:
        Dict with recommendation and whether switch is needed
    """
    content_lower = task_description.lower()

    # Quick scoring
    agent_score = sum(2 for kw in AGENT_INDICATORS["keywords"] if kw in content_lower)
    agent_score += sum(3 for p in AGENT_INDICATORS["patterns"] if re.search(p, content_lower))

    ask_score = sum(2 for kw in ASK_INDICATORS["keywords"] if kw in content_lower)
    ask_score += sum(3 for p in ASK_INDICATORS["patterns"] if re.search(p, content_lower))

    recommended = "AGENT" if agent_score > ask_score else "ASK"
    confidence = abs(agent_score - ask_score) / (agent_score + ask_score + 1) * 100

    needs_switch = current_mode and current_mode.upper() != recommended

    return {
        "recommended_mode": recommended,
        "confidence": round(confidence, 1),
        "needs_switch": needs_switch,
        "current_mode": current_mode,
        "suggestion_message": generate_mode_suggestion_message(recommended, confidence, task_description) if needs_switch else None,
    }


def format_mode_switch_prompt(
    from_mode: str,
    to_mode: str,
    reason: str = "",
) -> str:
    """
    Format a clear prompt for the AI to communicate mode switch suggestion.

    Args:
        from_mode: Current mode
        to_mode: Recommended mode
        reason: Why the switch is recommended

    Returns:
        Formatted prompt string for AI to display
    """
    to_suggestion = MODE_SUGGESTIONS.get(to_mode, MODE_SUGGESTIONS["ASK"])

    message = f"""
---
{to_suggestion['emoji']} **Suggested Mode Change: {from_mode} → {to_mode}**

{reason if reason else f"This task would benefit from {to_mode} mode."}

**How to switch:**
{to_suggestion['instruction']}

**Benefits of {to_mode} mode:**
{"".join(f"• {b}" + chr(10) for b in to_suggestion['benefits'][:3])}
---
"""
    return message.strip()


# ═══════════════════════════════════════════════════════════════════════════════
# INLINE SUGGESTION DECORATOR (for other tools)
# ═══════════════════════════════════════════════════════════════════════════════


def suggest_mode_if_needed(task_description: str) -> Optional[str]:
    """
    Check if mode suggestion should be appended to tool output.

    Call this at the end of tool execution to potentially add a mode suggestion.
    Returns None if no suggestion needed, otherwise returns suggestion text.

    Args:
        task_description: Description of what was just done

    Returns:
        Suggestion text or None
    """
    result = get_mode_suggestion_for_task(task_description)

    # Only suggest if confidence is high enough
    if result["confidence"] >= 70:
        return f"\n\n💡 **Tip:** {MODE_SUGGESTIONS[result['recommended_mode']]['action']} for better results with this type of task."

    return None

