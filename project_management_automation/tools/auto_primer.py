"""
Auto-Priming Tool for AI Session Start

Automatically primes AI context based on:
- Agent type (detected from cursor-agent.json or environment)
- Time of day (morning = daily_checkin mode)
- Previous session state
- Current workflow mode

This tool should be called at the start of each AI session to provide
optimal context with minimal token usage.

Usage:
    # At session start, call auto_prime()
    # Returns compact context tailored to current situation
"""

import json
import logging
import os
import time
from datetime import datetime
from pathlib import Path
from typing import Any, List, Optional

logger = logging.getLogger("exarp.auto_primer")


def _find_project_root() -> Path:
    """Find project root by looking for markers."""
    env_root = os.getenv("PROJECT_ROOT") or os.getenv("WORKSPACE_PATH")
    if env_root:
        return Path(env_root).resolve()

    current = Path.cwd()
    for _ in range(5):
        if (current / ".git").exists() or (current / ".todo2").exists() or (current / "CMakeLists.txt").exists() or (current / "go.mod").exists():
            return current.resolve()
        if current.parent == current:
            break
        current = current.parent

    return Path.cwd().resolve()


# ═══════════════════════════════════════════════════════════════════════════════
# AGENT DETECTION
# ═══════════════════════════════════════════════════════════════════════════════

def detect_agent_type() -> dict[str, Any]:
    """
    Detect the current agent type from environment or config.

    Detection sources (in order):
    1. EXARP_AGENT env var
    2. cursor-agent.json in current directory
    3. .cursor/mcp.json configuration
    4. Default to "general"
    """
    # 1. Environment variable
    agent = os.getenv("EXARP_AGENT")
    if agent:
        return {
            "agent": agent,
            "source": "environment",
            "config": {},
        }

    # 2. cursor-agent.json
    project_root = _find_project_root()
    agent_config = project_root / "cursor-agent.json"
    if agent_config.exists():
        try:
            config = json.loads(agent_config.read_text())
            return {
                "agent": config.get("name", config.get("agent", "unknown")),
                "source": "cursor-agent.json",
                "config": config,
            }
        except Exception as e:
            logger.debug(f"Error reading cursor-agent.json: {e}")

    # 3. Check agents directory structure
    cwd = Path.cwd()
    if "agents" in str(cwd):
        # Extract agent name from path like "agents/backend/"
        parts = cwd.parts
        try:
            agent_idx = parts.index("agents")
            if agent_idx + 1 < len(parts):
                return {
                    "agent": parts[agent_idx + 1],
                    "source": "path",
                    "config": {},
                }
        except ValueError:
            pass

    # 4. Default
    return {
        "agent": "general",
        "source": "default",
        "config": {},
    }


# ═══════════════════════════════════════════════════════════════════════════════
# TIME-BASED MODE SUGGESTION
# ═══════════════════════════════════════════════════════════════════════════════

def suggest_mode_by_time() -> dict[str, Any]:
    """
    Suggest workflow mode based on time of day.

    Time-based suggestions:
    - 6am-9am: daily_checkin (morning review)
    - 9am-12pm: development (prime working hours)
    - 12pm-1pm: task_management (lunch review)
    - 1pm-5pm: development (afternoon coding)
    - 5pm-6pm: code_review (end of day review)
    - After 6pm: development (evening work)
    """
    hour = datetime.now().hour

    if 6 <= hour < 9:
        return {
            "mode": "daily_checkin",
            "reason": "Morning hours - start with a health check",
            "confidence": 0.7,
        }
    elif 9 <= hour < 12:
        return {
            "mode": "development",
            "reason": "Prime working hours - development focus",
            "confidence": 0.8,
        }
    elif 12 <= hour < 13:
        return {
            "mode": "task_management",
            "reason": "Lunch time - good for task review",
            "confidence": 0.5,
        }
    elif 13 <= hour < 17:
        return {
            "mode": "development",
            "reason": "Afternoon - continued development",
            "confidence": 0.8,
        }
    elif 17 <= hour < 18:
        return {
            "mode": "code_review",
            "reason": "End of day - review and wrap up",
            "confidence": 0.6,
        }
    else:
        return {
            "mode": "development",
            "reason": "Evening/night - default development mode",
            "confidence": 0.5,
        }


# ═══════════════════════════════════════════════════════════════════════════════
# AGENT-SPECIFIC CONTEXT
# ═══════════════════════════════════════════════════════════════════════════════

# Agent type to recommended mode mapping
AGENT_MODE_MAPPING: dict[str, str] = {
    "backend": "development",
    "backend-agent": "development",
    "web": "development",
    "web-agent": "development",
    "tui": "development",
    "desktop": "development",
    "security": "security_review",
    "qa": "code_review",
    "devops": "sprint_planning",
    "pm": "task_management",
    "general": "development",
}

# Agent-specific context hints
AGENT_CONTEXT: dict[str, dict[str, Any]] = {
    "backend": {
        "focus_areas": ["API development", "database", "services", "Rust"],
        "relevant_tools": ["run_tests", "scan_dependency_security", "analyze_test_coverage", "session_handoff", "task_assignee"],
        "relevant_prompts": ["persona_developer", "security_scan_rust", "end_of_day"],
        "startup_hint": "Check session_handoff(action='resume') for handoffs from other developers",
    },
    "web": {
        "focus_areas": ["Frontend", "React", "TypeScript"],
        "relevant_tools": ["run_tests", "generate_cursorignore", "session_handoff", "task_assignee"],
        "relevant_prompts": ["persona_developer", "end_of_day"],
        "startup_hint": "Check session_handoff(action='resume') for handoffs from other developers",
    },
    "security": {
        "focus_areas": ["Vulnerability scanning", "Dependency audit", "Security review"],
        "relevant_tools": ["scan_dependency_security", "fetch_dependabot_alerts", "generate_security_report", "session_handoff"],
        "relevant_prompts": ["persona_security", "security_scan_all", "end_of_day"],
        "startup_hint": "Check session_handoff(action='resume') for handoffs from other developers",
    },
    "general": {
        "focus_areas": ["General development", "Project management"],
        "relevant_tools": ["project_scorecard", "check_documentation_health", "session_handoff", "task_assignee"],
        "relevant_prompts": ["daily_checkin", "persona_developer", "end_of_day", "resume_session"],
        "startup_hint": "Check session_handoff(action='resume') for handoffs from other developers",
    },
}


def get_agent_context(agent: str) -> dict[str, Any]:
    """Get context specific to an agent type."""
    # Normalize agent name
    agent_lower = agent.lower().replace("-agent", "")

    context = AGENT_CONTEXT.get(agent_lower, AGENT_CONTEXT["general"])
    mode = AGENT_MODE_MAPPING.get(agent_lower, AGENT_MODE_MAPPING.get(agent, "development"))

    return {
        "agent": agent,
        "normalized": agent_lower,
        "recommended_mode": mode,
        **context,
    }


# ═══════════════════════════════════════════════════════════════════════════════
# RECOMMENDED COMPANION MCPs
# ═══════════════════════════════════════════════════════════════════════════════

RECOMMENDED_COMPANIONS: dict[str, dict[str, Any]] = {
    "human": {
        "description": "Human interaction for confirmations and clarifications",
        "benefit": "Enables confirmations for batch operations, task clarification",
        "install": "npx -y @anthropic/mcp-human-in-the-loop",
        "detection_tools": ["ask_human", "confirm", "wait_for_input", "request_input"],
        "priority": "recommended",
        "use_cases": [
            "Confirm before batch task operations (merge duplicates, bulk assign)",
            "Ask clarifying questions when task descriptions are ambiguous",
            "Get human approval before auto-fixing issues",
            "Interactive task clarification resolution",
        ],
    },
    "context7": {
        "description": "Up-to-date library documentation lookup",
        "benefit": "Get current API docs when implementing automation scripts",
        "install": "npx -y @context7/mcp-server",
        "detection_tools": ["resolve-library-id", "get-library-docs"],
        "priority": "recommended",
        "use_cases": [
            "Look up FastMCP patterns",
            "Verify tool patterns match library best practices",
        ],
    },
    "sequential-thinking": {
        "description": "Step-by-step reasoning for complex workflows",
        "benefit": "Plan multi-step automation sequences",
        "install": "See docs/SEQUENTIAL_THINKING_MCP_INTEGRATION_STATUS.md",
        "detection_tools": ["sequential_thinking"],
        "priority": "optional",
        "use_cases": [
            "Break down complex refactoring tasks",
            "Create implementation workflows",
        ],
    },
    "tractatus-thinking": {
        "description": "Logical decomposition of complex problems",
        "benefit": "Analyze multiplicative dependencies in requirements",
        "install": "See docs/MCP_SERVERS.md",
        "detection_tools": ["tractatus_thinking"],
        "priority": "optional",
        "use_cases": [
            "Find the ONE missing element preventing task completion",
            "Decompose fuzzy requirements into measurable components",
        ],
    },
}


def detect_companion_mcps(available_tools: Optional[List[str]] = None) -> dict[str, Any]:
    """
    Detect which recommended companion MCPs are available.

    Args:
        available_tools: List of available tool names (if known)

    Returns:
        Dict with detected and missing companions
    """
    detected = []
    missing = []
    suggestions = []

    for mcp_name, mcp_info in RECOMMENDED_COMPANIONS.items():
        # Check if any detection tool is available
        is_detected = False
        if available_tools:
            for tool in mcp_info["detection_tools"]:
                if tool in available_tools or any(tool in t for t in available_tools):
                    is_detected = True
                    break

        if is_detected:
            detected.append({
                "name": mcp_name,
                "description": mcp_info["description"],
            })
        else:
            missing.append({
                "name": mcp_name,
                "description": mcp_info["description"],
                "benefit": mcp_info["benefit"],
                "install": mcp_info["install"],
                "priority": mcp_info["priority"],
            })

            if mcp_info["priority"] == "recommended":
                suggestions.append(
                    f"💡 Recommended MCP '{mcp_name}' not detected. "
                    f"{mcp_info['benefit']}. Install: {mcp_info['install']}"
                )

    return {
        "detected": detected,
        "missing": missing,
        "suggestions": suggestions,
        "all_recommended_present": all(
            m["priority"] != "recommended" for m in missing
        ),
    }


# ═══════════════════════════════════════════════════════════════════════════════
# AUTO-PRIME TOOL
# ═══════════════════════════════════════════════════════════════════════════════

def auto_prime(
    include_hints: bool = True,
    include_tasks: bool = True,
    include_prompts: bool = True,
    include_handoff: bool = True,
    override_mode: Optional[str] = None,
    compact: bool = True,
) -> str:
    """
    [HINT: Auto-primer. Returns optimal context for session start. Detects agent/time/mode. Checks for handoffs.]

    Automatically prime AI context at session start.

    📊 Output: Compact context tailored to agent type and time of day
    🔧 Side Effects: None (read-only)
    ⏱️ Typical Runtime: < 100ms

    This should be called at the start of each AI session to provide
    optimal context with minimal token usage.

    **Multi-Dev Coordination**: Checks for handoff notes from other developers
    to ensure continuity across machines.

    Args:
        include_hints: Include tool hints for detected mode
        include_tasks: Include recent task summary
        include_prompts: Include relevant prompts for agent/mode
        include_handoff: Check for handoff notes from other developers (default: True)
        override_mode: Override auto-detected mode
        compact: Return compact format (fewer tokens)

    Returns:
        Dict with primed context (FastMCP will serialize)

    Example:
        # At session start
        auto_prime()
        → Returns context tailored to agent type and time of day
        → Includes any handoff notes from previous developer

        # Force specific mode
        auto_prime(override_mode="security_review")
    """
    start_time = time.time()

    try:
        # 1. Detect agent type
        agent_info = detect_agent_type()
        agent_context = get_agent_context(agent_info["agent"])

        # 2. Determine mode
        if override_mode:
            mode = override_mode
            mode_source = "override"
        else:
            # Combine agent and time suggestions
            time_suggestion = suggest_mode_by_time()
            agent_mode = agent_context["recommended_mode"]

            # Prefer agent-specific mode in working hours, time-based otherwise
            if time_suggestion["mode"] == "daily_checkin":
                mode = time_suggestion["mode"]
                mode_source = "time_of_day"
            else:
                mode = agent_mode
                mode_source = "agent_type"

        # 3. Get context primer with detected mode
        from ..resources.context_primer import WORKFLOW_MODE_CONTEXT, get_context_primer

        primer_json = get_context_primer(
            mode=mode,
            include_hints=include_hints,
            include_tasks=include_tasks,
            include_goals=True,
            include_prompts=include_prompts,
        )
        primer = json.loads(primer_json)

        # 4. Add auto-prime specific info
        result = {
            "auto_primed": True,
            "timestamp": datetime.now().isoformat(),
            "duration_ms": round((time.time() - start_time) * 1000, 2),
            "detection": {
                "agent": agent_info["agent"],
                "agent_source": agent_info["source"],
                "mode": mode,
                "mode_source": mode_source,
                "time_of_day": datetime.now().strftime("%H:%M"),
            },
            "agent_context": {
                "focus_areas": agent_context["focus_areas"],
                "relevant_tools": agent_context["relevant_tools"][:5] if compact else agent_context["relevant_tools"],
                "relevant_prompts": agent_context["relevant_prompts"][:3] if compact else agent_context["relevant_prompts"],
            },
        }

        # 5. Check for companion MCPs (suggest if missing)
        companion_status = detect_companion_mcps()
        if companion_status["suggestions"]:
            result["companion_suggestions"] = companion_status["suggestions"][:2]  # Top 2

        # 6. Check for handoff notes from other developers (multi-dev coordination)
        if include_handoff:
            try:
                import socket

                from .session_handoff import get_latest_handoff

                handoff_json = get_latest_handoff()
                handoff_data = json.loads(handoff_json)

                if handoff_data.get("has_handoff"):
                    handoff = handoff_data.get("handoff", {})
                    current_host = socket.gethostname()

                    # Only show if from a different host (not our own handoff)
                    if handoff.get("host") != current_host:
                        result["handoff_alert"] = {
                            "from_host": handoff.get("host"),
                            "timestamp": handoff.get("timestamp"),
                            "summary": handoff.get("summary", "")[:100],
                            "blockers": handoff.get("blockers", []),
                            "next_steps": handoff.get("next_steps", [])[:3],
                        }
                        result["action_required"] = "📋 Review handoff from previous developer before starting work"
                    elif handoff.get("blockers"):
                        # Our own handoff - remind about blockers
                        result["reminder"] = {
                            "blockers": handoff.get("blockers", []),
                            "message": "You noted these blockers in your last session",
                        }
            except Exception as e:
                logger.debug(f"Could not check handoff: {e}")

        # 7. Merge with primer (compact or full)
        if compact:
            result["workflow"] = {
                "mode": mode,
                "description": WORKFLOW_MODE_CONTEXT.get(mode, {}).get("description", ""),
            }
            if include_hints:
                # Only include first 10 hints in compact mode
                hints = primer.get("hints", {})
                result["hints_count"] = len(hints)
                result["top_hints"] = dict(list(hints.items())[:10])
            if include_tasks:
                result["tasks"] = primer.get("tasks", {})
            if include_prompts:
                result["prompts"] = primer.get("prompts", {}).get("recommended", [])[:3]
        else:
            result.update(primer)

        # Return JSON string - FastMCP requires strings, not dicts
        return json.dumps(result, indent=2)

    except Exception as e:
        logger.error(f"Auto-prime error: {e}", exc_info=True)
        # Return error response as JSON string
        return json.dumps({
            "auto_primed": False,
            "error": str(e),
            "fallback_mode": "development",
        }, indent=2)


def get_session_context(task_id: Optional[str] = None) -> dict[str, Any]:
    """
    [HINT: Session context. Get context for a specific task or general session.]

    Get context optimized for working on a specific task.

    Args:
        task_id: Optional task ID to focus context on

    Returns:
        Dict with task-focused context (FastMCP will serialize)
    """
    try:
        # Get base auto-prime context
        base = json.loads(auto_prime(compact=True))

        if task_id:
            # Add task-specific context
            from ..resources.templates import get_task_by_id
            task_data = get_task_by_id(task_id)

            if task_data.get("found"):
                task = task_data.get("task", {})
                base["task_context"] = {
                    "id": task_id,
                    "name": task.get("name", ""),
                    "status": task.get("status", ""),
                    "tags": task.get("tags", []),
                    "description": task.get("description", "")[:200],  # Truncate
                }

                # Try to recall relevant memories
                try:
                    from ..resources.templates import get_memories_by_category
                    memories = get_memories_by_category("research")
                    if memories.get("count", 0) > 0:
                        base["related_memories"] = memories.get("memories", [])[:3]
                except Exception:
                    pass

        # Return dict directly - FastMCP will handle serialization
        return base

    except Exception as e:
        logger.error(f"Session context error: {e}", exc_info=True)
        # Return error response as dict
        return {"error": str(e)}


def prime_for_mode(mode: str) -> dict[str, Any]:
    """
    [HINT: Mode primer. Prime context for a specific workflow mode.]

    Quick primer for switching to a specific workflow mode.

    Args:
        mode: Workflow mode (daily_checkin, security_review, task_management, etc.)

    Returns:
        Dict with mode-specific context (FastMCP will serialize)
    """
    return auto_prime(override_mode=mode, compact=True)


# Register with MCP
def register_auto_primer_tools(mcp) -> None:
    """Register auto-primer tools with MCP server."""
    try:
        @mcp.tool()
        def auto_prime_session(
            include_hints: bool = True,
            include_tasks: bool = True,
            override_mode: Optional[str] = None,
        ) -> str:
            """
            [HINT: Auto-primer. Returns optimal context for session start. Detects agent/time/mode.]

            Auto-prime AI context at session start.
            Call this at the beginning of each session for optimal context.
            """
            try:
                result = auto_prime(
                    include_hints=include_hints,
                    include_tasks=include_tasks,
                    override_mode=override_mode,
                )
                # Ensure we always return a string (JSON)
                if isinstance(result, str):
                    return result
                elif isinstance(result, dict):
                    # Defensive: if somehow a dict is returned, convert to JSON
                    return json.dumps(result, indent=2)
                else:
                    # Fallback: convert to JSON string
                    return json.dumps({"result": str(result)}, indent=2)
            except Exception as e:
                logger.error(f"Error in auto_prime_session: {e}", exc_info=True)
                return json.dumps({
                    "auto_primed": False,
                    "error": str(e),
                    "fallback_mode": "development",
                }, indent=2)

        @mcp.tool()
        def get_task_context(task_id: Optional[str] = None) -> str:
            """
            [HINT: Task context. Get optimized context for working on a task.]

            Get context optimized for a specific task.
            """
            try:
                result = get_session_context(task_id=task_id)
                # Ensure we always return a string (JSON)
                if isinstance(result, str):
                    return result
                elif isinstance(result, dict):
                    # Defensive: if somehow a dict is returned, convert to JSON
                    return json.dumps(result, indent=2)
                else:
                    # Fallback: convert to JSON string
                    return json.dumps({"result": str(result)}, indent=2)
            except Exception as e:
                logger.error(f"Error in get_task_context: {e}", exc_info=True)
                return json.dumps({
                    "error": str(e),
                    "task_id": task_id,
                }, indent=2)

        logger.info("✅ Registered 2 auto-primer tools")

    except Exception as e:
        logger.warning(f"Could not register auto-primer tools: {e}")


__all__ = [
    "detect_agent_type",
    "suggest_mode_by_time",
    "get_agent_context",
    "auto_prime",
    "get_session_context",
    "prime_for_mode",
    "register_auto_primer_tools",
    "detect_companion_mcps",
    "AGENT_MODE_MAPPING",
    "AGENT_CONTEXT",
    "RECOMMENDED_COMPANIONS",
]

