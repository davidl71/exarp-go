"""
Dynamic Tool Management for MCP Server.

Implements context-aware tool loading based on the MCP listChanged notification.
Instead of exposing 40+ tools at all times (which pollutes LLM context),
this module enables intelligent tool curation:

1. CORE tools (always visible): server_status, list_tools, focus_mode
2. TOOL_CATALOG tools: help LLM understand what's available
3. CONTEXTUAL tools: activated based on workflow/focus mode

Philosophy (from https://www.jlowin.dev/blog/stop-converting-rest-apis-to-mcp):
- "An API built for humans will poison your AI agent"
- "Context pollution is the silent killer of agentic workflows"
- "Ruthlessly curated and minimalist" > "rich, composable, atomic"

MCP Spec Reference (2025-06-18):
- tools/list_changed notification enables dynamic tool lists
- Clients re-fetch via tools/list when notified

Features:
- Workflow modes with pre-defined tool sets
- Adaptive mode inference from conversation context
- Tool usage tracking for learning patterns
"""

import json
import logging
import re
import time
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from pathlib import Path
from typing import Any, Dict, List, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    try:
        from fastmcp import Context
    except ImportError:
        Context = Any

logger = logging.getLogger("exarp.dynamic_tools")


# ═══════════════════════════════════════════════════════════════════════════════
# ADAPTIVE MODE INFERENCE
# ═══════════════════════════════════════════════════════════════════════════════

# Keywords that suggest specific workflow modes
MODE_KEYWORDS: dict[str, list[str]] = {
    "daily_checkin": [
        "daily", "check", "status", "morning", "standup", "overview",
        "health", "how is", "project status", "quick check",
    ],
    "security_review": [
        "security", "vulnerability", "vulnerabilities", "cve", "cves", "audit",
        "dependabot", "dependency", "dependencies", "scan", "secure",
        "exploit", "risk", "compliance", "scan for",
    ],
    "task_management": [
        "task", "tasks", "todo", "backlog", "ticket", "tickets",
        "duplicate", "alignment", "clarification", "assign", "priority",
        "sprint backlog", "grooming", "groom", "manage tasks", "manage task",
    ],
    "sprint_planning": [
        "sprint", "planning", "plan", "roadmap", "milestone",
        "prd", "requirements", "epic", "story", "stories",
        "automation", "automate",
    ],
    "code_review": [
        "review", "pr", "pull request", "code review", "lint", "linter",
        "test", "tests", "testing", "coverage", "quality", "ci", "cd",
        "pipeline", "build",
    ],
    "debugging": [
        "debug", "bug", "fix", "error", "issue", "problem", "broken",
        "crash", "failing", "failure", "investigate", "trace", "memory",
        "recall", "what did", "previous",
    ],
    "development": [
        "develop", "implement", "build", "create", "feature", "code",
        "write", "add", "new",
    ],
}

# Tool name patterns that suggest workflow modes
TOOL_USAGE_MODE_HINTS: dict[str, str] = {
    "scan_dependency_security": "security_review",
    "generate_security_report": "security_review",
    "fetch_dependabot_alerts": "security_review",
    "analyze_todo2_alignment": "task_management",
    "detect_duplicate_tasks": "task_management",
    "batch_approve_tasks": "task_management",
    "run_tests": "code_review",
    "analyze_test_coverage": "code_review",
    "check_definition_of_done": "code_review",
    "generate_prd": "sprint_planning",
    "analyze_prd_alignment": "sprint_planning",
    "run_automation": "sprint_planning",
    "save_memory": "debugging",
    "recall_context": "debugging",
    "search_memories": "debugging",
    "generate_project_scorecard": "daily_checkin",
    "generate_project_overview": "daily_checkin",
}


class ToolGroup(str, Enum):
    """Tool groups for lazy loading."""

    # Always visible
    CORE = "core"           # server_status
    TOOL_CATALOG = "tool_catalog" # list_tools, get_tool_help, focus_mode

    # Contextually loaded
    HEALTH = "health"       # project_scorecard, project_overview, docs_health
    TASKS = "tasks"         # alignment, duplicates, clarification, hierarchy
    SECURITY = "security"   # dependency scan, security report, dependabot
    AUTOMATION = "automation"  # run_automation, sprint_automation, batch_approve
    CONFIG = "config"       # cursor_rules, cursorignore, simplify_rules
    TESTING = "testing"     # run_tests, analyze_coverage, definition_of_done
    ADVISORS = "advisors"   # consult_advisor, get_advisor_briefing, list_advisors
    MEMORY = "memory"       # save_memory, recall_context, search_memories
    WORKFLOW = "workflow"   # recommend_workflow_mode, recommend_model, log_prompt
    PRD = "prd"             # generate_prd, analyze_prd_alignment


class WorkflowMode(str, Enum):
    """Pre-defined workflow contexts with curated tool sets."""

    # Minimal modes (5-8 tools each)
    DAILY_CHECKIN = "daily_checkin"       # Overview + health checks
    SECURITY_REVIEW = "security_review"   # Security-focused tools
    TASK_MANAGEMENT = "task_management"   # Task tools only
    SPRINT_PLANNING = "sprint_planning"   # Tasks + automation
    CODE_REVIEW = "code_review"           # Testing + linting

    # Standard modes (10-15 tools)
    DEVELOPMENT = "development"           # Balanced set
    DEBUGGING = "debugging"               # Memory + testing

    # Full access
    ALL = "all"                           # All tools (legacy behavior)


# Tool group definitions
TOOL_GROUP_MAPPING: dict[str, ToolGroup] = {
    # Core (always)
    "server_status": ToolGroup.CORE,
    # Discovery (always)
    "list_tools": ToolGroup.TOOL_CATALOG,
    "get_tool_help": ToolGroup.TOOL_CATALOG,
    "focus_mode": ToolGroup.TOOL_CATALOG,  # The mode switcher itself
    "suggest_mode": ToolGroup.TOOL_CATALOG,  # Adaptive mode suggestion
    "tool_usage_stats": ToolGroup.TOOL_CATALOG,  # Usage analytics

    # Health
    "check_documentation_health": ToolGroup.HEALTH,
    "generate_project_scorecard": ToolGroup.HEALTH,
    "generate_project_overview": ToolGroup.HEALTH,
    "check_working_copy_health": ToolGroup.HEALTH,

    # Tasks
    "analyze_todo2_alignment": ToolGroup.TASKS,
    "detect_duplicate_tasks": ToolGroup.TASKS,
    "sync_todo_tasks": ToolGroup.TASKS,
    "clarification": ToolGroup.TASKS,
    "batch_approve_tasks": ToolGroup.TASKS,
    "analyze_task_hierarchy": ToolGroup.TASKS,
    "consolidate_tags": ToolGroup.TASKS,

    # Security
    "scan_dependency_security": ToolGroup.SECURITY,
    "fetch_dependabot_alerts": ToolGroup.SECURITY,
    "generate_security_report": ToolGroup.SECURITY,

    # Automation
    "run_automation": ToolGroup.AUTOMATION,
    "find_automation_opportunities": ToolGroup.AUTOMATION,
    "setup_git_hooks": ToolGroup.AUTOMATION,
    "setup_pattern_triggers": ToolGroup.AUTOMATION,

    # Config
    "generate_cursor_rules": ToolGroup.CONFIG,
    "generate_cursorignore": ToolGroup.CONFIG,
    "simplify_rules": ToolGroup.CONFIG,
    "add_external_tool_hints": ToolGroup.CONFIG,
    "validate_ci_cd_workflow": ToolGroup.CONFIG,

    # Testing
    "run_tests": ToolGroup.TESTING,
    "analyze_test_coverage": ToolGroup.TESTING,
    "check_definition_of_done": ToolGroup.TESTING,
    "analyze_problems": ToolGroup.TESTING,
    "list_problem_categories": ToolGroup.TESTING,

    # Advisors
    "consult_advisor": ToolGroup.ADVISORS,
    "get_advisor_briefing": ToolGroup.ADVISORS,
    "list_advisors": ToolGroup.ADVISORS,
    # Audio tools removed - migrated to devwisdom-go MCP server

    # Memory
    "save_memory": ToolGroup.MEMORY,
    "recall_context": ToolGroup.MEMORY,
    "search_memories": ToolGroup.MEMORY,
    "get_session_summary": ToolGroup.MEMORY,

    # Workflow
    "recommend_workflow_mode": ToolGroup.WORKFLOW,
    "recommend_model": ToolGroup.WORKFLOW,
    "list_available_models": ToolGroup.WORKFLOW,
    "log_prompt_iteration": ToolGroup.WORKFLOW,
    "analyze_prompt_iterations": ToolGroup.WORKFLOW,

    # PRD
    "generate_prd": ToolGroup.PRD,
    "analyze_prd_alignment": ToolGroup.PRD,
}


# Workflow mode → tool groups mapping
WORKFLOW_TOOL_GROUPS: dict[WorkflowMode, set[ToolGroup]] = {
    # Minimal focused modes
    WorkflowMode.DAILY_CHECKIN: {
        ToolGroup.CORE, ToolGroup.TOOL_CATALOG, ToolGroup.HEALTH
    },
    WorkflowMode.SECURITY_REVIEW: {
        ToolGroup.CORE, ToolGroup.TOOL_CATALOG, ToolGroup.SECURITY, ToolGroup.HEALTH
    },
    WorkflowMode.TASK_MANAGEMENT: {
        ToolGroup.CORE, ToolGroup.TOOL_CATALOG, ToolGroup.TASKS
    },
    WorkflowMode.SPRINT_PLANNING: {
        ToolGroup.CORE, ToolGroup.TOOL_CATALOG, ToolGroup.TASKS, ToolGroup.AUTOMATION, ToolGroup.PRD
    },
    WorkflowMode.CODE_REVIEW: {
        ToolGroup.CORE, ToolGroup.TOOL_CATALOG, ToolGroup.TESTING, ToolGroup.HEALTH
    },

    # Balanced modes
    WorkflowMode.DEVELOPMENT: {
        ToolGroup.CORE, ToolGroup.TOOL_CATALOG, ToolGroup.HEALTH,
        ToolGroup.TASKS, ToolGroup.TESTING, ToolGroup.MEMORY
    },
    WorkflowMode.DEBUGGING: {
        ToolGroup.CORE, ToolGroup.TOOL_CATALOG, ToolGroup.MEMORY,
        ToolGroup.TESTING, ToolGroup.HEALTH
    },

    # Full access
    WorkflowMode.ALL: set(ToolGroup),
}


@dataclass
class ToolUsageTracker:
    """
    Tracks tool usage patterns for adaptive recommendations.

    Persists usage data to enable learning across sessions.
    """

    # Tool usage counts
    tool_counts: dict[str, int] = field(default_factory=lambda: defaultdict(int))

    # Tool co-occurrence (which tools are used together)
    tool_pairs: dict[str, dict[str, int]] = field(default_factory=lambda: defaultdict(lambda: defaultdict(int)))

    # Recent tool sequence (for pattern detection)
    recent_tools: list[str] = field(default_factory=list)
    max_recent: int = 20

    # Mode usage counts
    mode_counts: dict[str, int] = field(default_factory=lambda: defaultdict(int))

    # Session start time
    session_start: str = field(default_factory=lambda: datetime.now().isoformat())

    def record_tool_call(self, tool_name: str) -> None:
        """Record a tool being called."""
        self.tool_counts[tool_name] += 1

        # Track co-occurrence with recent tools
        for recent in self.recent_tools[-5:]:  # Last 5 tools
            if recent != tool_name:
                self.tool_pairs[tool_name][recent] += 1
                self.tool_pairs[recent][tool_name] += 1

        # Update recent list
        self.recent_tools.append(tool_name)
        if len(self.recent_tools) > self.max_recent:
            self.recent_tools.pop(0)

    def record_mode_switch(self, mode: str) -> None:
        """Record a mode switch."""
        self.mode_counts[mode] += 1

    def get_related_tools(self, tool_name: str, limit: int = 5) -> list[tuple[str, int]]:
        """Get tools commonly used with the given tool."""
        pairs = self.tool_pairs.get(tool_name, {})
        sorted_pairs = sorted(pairs.items(), key=lambda x: x[1], reverse=True)
        return sorted_pairs[:limit]

    def get_most_used_tools(self, limit: int = 10) -> list[tuple[str, int]]:
        """Get the most frequently used tools."""
        sorted_tools = sorted(self.tool_counts.items(), key=lambda x: x[1], reverse=True)
        return sorted_tools[:limit]

    def get_preferred_mode(self) -> Optional[str]:
        """Get the most commonly used mode."""
        if not self.mode_counts:
            return None
        return max(self.mode_counts.items(), key=lambda x: x[1])[0]

    def to_dict(self) -> dict[str, Any]:
        """Serialize for persistence."""
        return {
            "tool_counts": dict(self.tool_counts),
            "tool_pairs": {k: dict(v) for k, v in self.tool_pairs.items()},
            "recent_tools": self.recent_tools,
            "mode_counts": dict(self.mode_counts),
            "session_start": self.session_start,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "ToolUsageTracker":
        """Deserialize from persisted data."""
        tracker = cls()
        tracker.tool_counts = defaultdict(int, data.get("tool_counts", {}))
        tracker.tool_pairs = defaultdict(
            lambda: defaultdict(int),
            {k: defaultdict(int, v) for k, v in data.get("tool_pairs", {}).items()}
        )
        tracker.recent_tools = data.get("recent_tools", [])
        tracker.mode_counts = defaultdict(int, data.get("mode_counts", {}))
        tracker.session_start = data.get("session_start", datetime.now().isoformat())
        return tracker


@dataclass
class FileEditTracker:
    """
    Tracks file edits during tool calls to detect multi-file vs single-file patterns.

    Used by SessionModeInference to determine if a session involves editing
    multiple files (AGENT mode) vs single file (ASK/MANUAL mode).
    """

    edited_files: set[str] = field(default_factory=set)
    edit_timestamps: list[tuple[str, float]] = field(default_factory=list)
    max_tracked: int = 100

    def record_file_edit(self, file_path: str) -> None:
        """
        Record that a file was edited.

        Args:
            file_path: Path to the edited file (can be relative or absolute)
        """
        # Normalize path (store as string for JSON serialization)
        normalized_path = str(Path(file_path).resolve())
        self.edited_files.add(normalized_path)

        # Record timestamp
        current_time = time.time()
        self.edit_timestamps.append((normalized_path, current_time))

        # Limit tracked timestamps to prevent memory bloat
        if len(self.edit_timestamps) > self.max_tracked:
            self.edit_timestamps = self.edit_timestamps[-self.max_tracked:]

    def get_unique_files_count(self) -> int:
        """
        Get the number of unique files edited.

        Returns:
            Number of unique files edited in this session
        """
        return len(self.edited_files)

    def get_edit_frequency(self, window_seconds: float = 60) -> float:
        """
        Get the frequency of file edits in edits per minute.

        Args:
            window_seconds: Time window to analyze (default: 60 seconds)

        Returns:
            Edits per minute in the specified window
        """
        if not self.edit_timestamps:
            return 0.0

        current_time = time.time()
        window_start = current_time - window_seconds

        # Count edits in the window
        edits_in_window = sum(
            1 for _, timestamp in self.edit_timestamps
            if timestamp >= window_start
        )

        # Convert to edits per minute
        return (edits_in_window / window_seconds) * 60.0

    def is_multi_file_session(self, threshold: int = 2) -> bool:
        """
        Check if this session involves editing multiple files.

        Args:
            threshold: Minimum number of files to consider "multi-file" (default: 2)

        Returns:
            True if more than threshold files have been edited
        """
        return len(self.edited_files) > threshold

    def to_dict(self) -> dict[str, Any]:
        """Serialize for persistence."""
        return {
            "edited_files": list(self.edited_files),
            "edit_timestamps": self.edit_timestamps,
            "max_tracked": self.max_tracked,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "FileEditTracker":
        """Deserialize from persisted data."""
        tracker = cls()
        tracker.edited_files = set(data.get("edited_files", []))
        tracker.edit_timestamps = data.get("edit_timestamps", [])
        tracker.max_tracked = data.get("max_tracked", 100)
        return tracker


@dataclass
class DynamicToolManager:
    """
    Manages dynamic tool visibility based on context.

    Usage:
        manager = DynamicToolManager()

        # Switch to security review mode
        await manager.set_workflow_mode(ctx, WorkflowMode.SECURITY_REVIEW)

        # Or enable specific groups
        await manager.enable_group(ctx, ToolGroup.MEMORY)

        # Check if tool should be visible
        if manager.is_tool_visible("scan_dependency_security"):
            # Include in tools/list response

        # Infer mode from conversation
        suggested = manager.infer_mode_from_text("I need to check for vulnerabilities")
        # Returns: ("security_review", 0.85, ["vulnerability"])
    """

    # Current workflow mode
    current_mode: WorkflowMode = WorkflowMode.DEVELOPMENT

    # Explicitly enabled groups (in addition to mode defaults)
    extra_groups: set[ToolGroup] = field(default_factory=set)

    # Explicitly disabled groups (override mode defaults)
    disabled_groups: set[ToolGroup] = field(default_factory=set)

    # Tool usage tracking
    usage_tracker: ToolUsageTracker = field(default_factory=ToolUsageTracker)

    # File edit tracking (MODE-002)
    file_tracker: FileEditTracker = field(default_factory=FileEditTracker)

    # Session mode inference (MODE-002)
    mode_inference: Optional[Any] = None  # SessionModeInference - lazy import to avoid circular deps
    inferred_mode: Optional[Any] = None  # ModeInferenceResult - lazy import
    mode_history: list[Any] = field(default_factory=list)  # List[ModeInferenceResult]
    last_mode_update: Optional[float] = None  # Timestamp of last mode update

    # Persistence path (optional)
    persistence_path: Optional[Path] = None

    def get_active_groups(self) -> set[ToolGroup]:
        """Get currently active tool groups."""
        base_groups = WORKFLOW_TOOL_GROUPS.get(self.current_mode, set())
        active = (base_groups | self.extra_groups) - self.disabled_groups

        # Core and Discovery are always active
        active.add(ToolGroup.CORE)
        active.add(ToolGroup.TOOL_CATALOG)

        return active

    def get_visible_tools(self) -> list[str]:
        """Get list of currently visible tool names."""
        active_groups = self.get_active_groups()
        visible = []

        for tool_name, group in TOOL_GROUP_MAPPING.items():
            if group in active_groups:
                visible.append(tool_name)

        return sorted(visible)

    def is_tool_visible(self, tool_name: str) -> bool:
        """Check if a specific tool should be visible."""
        group = TOOL_GROUP_MAPPING.get(tool_name)
        if group is None:
            # Unknown tool - show it (fail open)
            logger.warning(f"Unknown tool '{tool_name}' - defaulting to visible")
            return True

        return group in self.get_active_groups()

    def record_tool_usage(self, tool_name: str, tool_args: Optional[Dict[str, Any]] = None) -> None:
        """
        Record tool usage for adaptive recommendations.

        Args:
            tool_name: Name of the tool being called
            tool_args: Optional tool arguments (used to extract file paths for MODE-002)
        """
        self.usage_tracker.record_tool_call(tool_name)

        # Extract file paths from tool arguments (MODE-002)
        if tool_args:
            file_paths = self._extract_file_paths(tool_name, tool_args)
            for file_path in file_paths:
                self.file_tracker.record_file_edit(file_path)

        # Check if this tool suggests a mode switch
        suggested_mode = TOOL_USAGE_MODE_HINTS.get(tool_name)
        if suggested_mode and suggested_mode != self.current_mode.value:
            logger.debug(f"Tool {tool_name} suggests mode: {suggested_mode}")

    def _extract_file_paths(self, tool_name: str, tool_args: dict[str, Any]) -> list[str]:
        """
        Extract file paths from tool arguments.

        Args:
            tool_name: Name of the tool
            tool_args: Tool arguments dictionary

        Returns:
            List of file paths found in arguments
        """
        file_paths = []

        # Common file path argument names
        file_arg_names = [
            "file_path", "target_file", "file", "path",
            "file_paths", "files", "paths"  # plural forms
        ]

        # Check for direct file path arguments
        for arg_name in file_arg_names:
            if arg_name in tool_args:
                value = tool_args[arg_name]
                if isinstance(value, str):
                    file_paths.append(value)
                elif isinstance(value, list):
                    file_paths.extend([v for v in value if isinstance(v, str)])

        # Tool-specific extraction
        if tool_name in ["search_replace", "write", "edit_file", "read_file"]:
            # These tools typically have file_path or target_file
            if "file_path" in tool_args:
                file_paths.append(tool_args["file_path"])
            if "target_file" in tool_args:
                file_paths.append(tool_args["target_file"])

        return file_paths

    def update_inferred_mode(self) -> Optional[Any]:
        """
        Update inferred session mode based on current tool and file patterns.

        Returns:
            ModeInferenceResult or None if inference not available
        """
        try:
            # Lazy import to avoid circular dependencies
            if self.mode_inference is None:
                from .session_mode_inference import SessionModeInference
                self.mode_inference = SessionModeInference()

            # Calculate session duration
            from datetime import datetime
            session_start = datetime.fromisoformat(self.usage_tracker.session_start)
            session_duration = (datetime.now() - session_start).total_seconds()

            # Infer mode
            result = self.mode_inference.infer_mode(
                tool_tracker=self.usage_tracker,
                file_tracker=self.file_tracker,
                session_duration_seconds=session_duration
            )

            # Store result
            self.inferred_mode = result
            self.mode_history.append(result)

            # Limit history size
            if len(self.mode_history) > 50:
                self.mode_history = self.mode_history[-50:]

            self.last_mode_update = time.time()

            logger.debug(
                f"Inferred session mode: {result.mode.value} "
                f"(confidence: {result.confidence:.1%})"
            )

            return result

        except Exception as e:
            logger.warning(f"Failed to update inferred mode: {e}")
            return None

    def get_current_mode(self) -> Optional[Any]:
        """
        Get current inferred session mode.

        Returns:
            ModeInferenceResult or None
        """
        # Update if needed (every 2 minutes or if never updated)
        current_time = time.time()
        if (
            self.last_mode_update is None
            or (current_time - self.last_mode_update) > 120  # 2 minutes
        ):
            self.update_inferred_mode()

        return self.inferred_mode

    def get_recommended_groups(self) -> list[ToolGroup]:
        """Get group recommendations based on usage patterns."""
        group_scores: dict[ToolGroup, int] = {}

        for tool_name, count in self.usage_tracker.tool_counts.items():
            group = TOOL_GROUP_MAPPING.get(tool_name)
            if group:
                group_scores[group] = group_scores.get(group, 0) + count

        # Return groups sorted by usage
        return sorted(group_scores.keys(), key=lambda g: group_scores[g], reverse=True)

    # ═══════════════════════════════════════════════════════════════════════
    # ADAPTIVE MODE INFERENCE
    # ═══════════════════════════════════════════════════════════════════════

    def infer_mode_from_text(
        self,
        text: str,
        threshold: float = 0.3
    ) -> tuple[Optional[str], float, list[str]]:
        """
        Infer the best workflow mode from conversation text.

        Uses keyword matching with scoring to suggest the most
        appropriate mode based on what the user is discussing.

        Args:
            text: Conversation text to analyze
            threshold: Minimum confidence score (0-1) to return a suggestion

        Returns:
            Tuple of (suggested_mode, confidence, matched_keywords)
            If no mode exceeds threshold, returns (None, 0, [])
        """
        text_lower = text.lower()

        mode_scores: dict[str, tuple[float, list[str]]] = {}

        for mode, keywords in MODE_KEYWORDS.items():
            matches = []
            for keyword in keywords:
                # Use word boundary matching for better accuracy
                pattern = r'\b' + re.escape(keyword) + r'\b'
                if re.search(pattern, text_lower):
                    matches.append(keyword)

            if matches:
                # Score based on number and specificity of matches
                # More specific/longer keywords get higher weight
                score = sum(len(kw.split()) for kw in matches) / 10.0
                score = min(score, 1.0)  # Cap at 1.0
                mode_scores[mode] = (score, matches)

        if not mode_scores:
            return (None, 0.0, [])

        # Get highest scoring mode
        best_mode = max(mode_scores.items(), key=lambda x: x[1][0])
        mode_name, (score, matches) = best_mode

        if score >= threshold:
            return (mode_name, score, matches)

        return (None, 0.0, [])

    def get_mode_suggestion(
        self,
        text: Optional[str] = None,
        include_rationale: bool = True
    ) -> dict[str, Any]:
        """
        Get a mode suggestion based on text and/or usage patterns.

        Combines text inference with usage history for better recommendations.

        Args:
            text: Optional conversation text to analyze
            include_rationale: Include explanation of the suggestion

        Returns:
            Dict with suggestion, confidence, and rationale
        """
        suggestions = []

        # Text-based inference
        if text:
            mode, confidence, keywords = self.infer_mode_from_text(text)
            if mode:
                suggestions.append({
                    "mode": mode,
                    "confidence": confidence,
                    "source": "text_analysis",
                    "keywords": keywords,
                })

        # Usage pattern inference
        recent_tools = self.usage_tracker.recent_tools[-5:]
        mode_hints = {}
        for tool in recent_tools:
            hint = TOOL_USAGE_MODE_HINTS.get(tool)
            if hint:
                mode_hints[hint] = mode_hints.get(hint, 0) + 1

        if mode_hints:
            top_hint = max(mode_hints.items(), key=lambda x: x[1])
            suggestions.append({
                "mode": top_hint[0],
                "confidence": min(top_hint[1] / 3.0, 0.8),  # Cap at 0.8
                "source": "usage_pattern",
                "tools": [t for t in recent_tools if TOOL_USAGE_MODE_HINTS.get(t) == top_hint[0]],
            })

        # Historical preference
        preferred = self.usage_tracker.get_preferred_mode()
        if preferred:
            suggestions.append({
                "mode": preferred,
                "confidence": 0.3,  # Low confidence for historical
                "source": "historical_preference",
            })

        if not suggestions:
            return {
                "suggested_mode": None,
                "confidence": 0.0,
                "rationale": "No clear mode suggestion based on available context",
                "current_mode": self.current_mode.value,
            }

        # Pick best suggestion
        best = max(suggestions, key=lambda x: x["confidence"])

        result = {
            "suggested_mode": best["mode"],
            "confidence": round(best["confidence"], 2),
            "current_mode": self.current_mode.value,
            "would_change": best["mode"] != self.current_mode.value,
        }

        if include_rationale:
            if best["source"] == "text_analysis":
                result["rationale"] = f"Keywords detected: {', '.join(best['keywords'])}"
            elif best["source"] == "usage_pattern":
                result["rationale"] = f"Recent tools suggest this mode: {', '.join(best.get('tools', []))}"
            else:
                result["rationale"] = "Based on your historical usage patterns"

        return result

    # ═══════════════════════════════════════════════════════════════════════
    # PERSISTENCE
    # ═══════════════════════════════════════════════════════════════════════

    def save_usage_data(self, path: Optional[Path] = None) -> bool:
        """Save usage tracking data to disk."""
        save_path = path or self.persistence_path
        if not save_path:
            # Default to .exarp directory
            save_path = Path(".exarp/tool_usage.json")

        try:
            save_path.parent.mkdir(parents=True, exist_ok=True)
            data = {
                "version": 1,
                "current_mode": self.current_mode.value,
                "tracker": self.usage_tracker.to_dict(),
            }
            with open(save_path, "w") as f:
                json.dump(data, f, indent=2)
            logger.debug(f"Saved usage data to {save_path}")
            return True
        except Exception as e:
            logger.warning(f"Failed to save usage data: {e}")
            return False

    def load_usage_data(self, path: Optional[Path] = None) -> bool:
        """Load usage tracking data from disk."""
        load_path = path or self.persistence_path
        if not load_path:
            load_path = Path(".exarp/tool_usage.json")

        try:
            if not load_path.exists():
                return False

            with open(load_path) as f:
                data = json.load(f)

            if data.get("version") == 1:
                self.usage_tracker = ToolUsageTracker.from_dict(data.get("tracker", {}))
                # Optionally restore mode
                # self.current_mode = WorkflowMode(data.get("current_mode", "development"))
                logger.debug(f"Loaded usage data from {load_path}")
                return True
        except Exception as e:
            logger.warning(f"Failed to load usage data: {e}")

        return False

    async def set_workflow_mode(
        self,
        ctx: Optional["Context"],
        mode: WorkflowMode,
    ) -> dict[str, Any]:
        """
        Switch to a workflow mode and notify client.

        Args:
            ctx: FastMCP context for notification
            mode: Target workflow mode

        Returns:
            Status dict with mode info and tool counts
        """
        old_mode = self.current_mode
        old_tools = set(self.get_visible_tools())

        self.current_mode = mode
        self.extra_groups.clear()
        self.disabled_groups.clear()

        new_tools = set(self.get_visible_tools())

        # Notify client of tool list change
        if ctx:
            await self._notify_tools_changed(ctx)

        added = new_tools - old_tools
        removed = old_tools - new_tools

        return {
            "success": True,
            "mode": mode.value,
            "previous_mode": old_mode.value,
            "tool_count": len(new_tools),
            "groups_active": [g.value for g in self.get_active_groups()],
            "tools_added": sorted(added),
            "tools_removed": sorted(removed),
        }

    async def enable_group(
        self,
        ctx: Optional["Context"],
        group: ToolGroup,
    ) -> dict[str, Any]:
        """Enable a tool group."""
        if group in self.disabled_groups:
            self.disabled_groups.remove(group)
        self.extra_groups.add(group)

        if ctx:
            await self._notify_tools_changed(ctx)

        return {
            "success": True,
            "group": group.value,
            "action": "enabled",
            "visible_tools": self.get_visible_tools(),
        }

    async def disable_group(
        self,
        ctx: Optional["Context"],
        group: ToolGroup,
    ) -> dict[str, Any]:
        """Disable a tool group."""
        # Never disable core/tool_catalog
        if group in (ToolGroup.CORE, ToolGroup.TOOL_CATALOG):
            return {
                "success": False,
                "error": f"Cannot disable {group.value} group - always required",
            }

        if group in self.extra_groups:
            self.extra_groups.remove(group)
        self.disabled_groups.add(group)

        if ctx:
            await self._notify_tools_changed(ctx)

        return {
            "success": True,
            "group": group.value,
            "action": "disabled",
            "visible_tools": self.get_visible_tools(),
        }

    async def _notify_tools_changed(self, ctx: "Context") -> None:
        """Send tools/list_changed notification to client."""
        try:
            from ..context_helpers import notify_tools_changed
            await notify_tools_changed(ctx)
            logger.info(f"Notified client of tool list change (mode={self.current_mode.value})")
        except Exception as e:
            logger.warning(f"Failed to notify tools changed: {e}")

    def get_status(self) -> dict[str, Any]:
        """Get current status for introspection."""
        active_groups = self.get_active_groups()
        visible_tools = self.get_visible_tools()
        all_tools = list(TOOL_GROUP_MAPPING.keys())

        return {
            "mode": self.current_mode.value,
            "active_groups": sorted([g.value for g in active_groups]),
            "extra_groups": sorted([g.value for g in self.extra_groups]),
            "disabled_groups": sorted([g.value for g in self.disabled_groups]),
            "visible_tool_count": len(visible_tools),
            "total_tool_count": len(all_tools),
            "reduction_percent": round((1 - len(visible_tools) / len(all_tools)) * 100, 1),
            "visible_tools": visible_tools,
            "hidden_tools": sorted(set(all_tools) - set(visible_tools)),
            "available_modes": [m.value for m in WorkflowMode],
            "available_groups": [g.value for g in ToolGroup],
        }


# Global instance (singleton pattern for MCP server)
_manager: Optional[DynamicToolManager] = None


def get_tool_manager() -> DynamicToolManager:
    """Get or create the global tool manager."""
    global _manager
    if _manager is None:
        _manager = DynamicToolManager()
    return _manager


def reset_tool_manager() -> None:
    """Reset tool manager (for testing)."""
    global _manager
    _manager = None


# ═══════════════════════════════════════════════════════════════════════════════
# MCP TOOL IMPLEMENTATIONS
# ═══════════════════════════════════════════════════════════════════════════════

def focus_mode(
    mode: Optional[str] = None,
    enable_group: Optional[str] = None,
    disable_group: Optional[str] = None,
    status: bool = False,
) -> str:
    """
    [HINT: Tool curation. Dynamic tool visibility based on workflow mode.]

    🎯 Output: Mode status, visible tools, context reduction metrics
    🔧 Side Effects: Updates tool visibility, sends list_changed notification
    ⏱️ Typical Runtime: <100ms

    Philosophy: "An API built for humans will poison your AI agent."
    Instead of 40+ tools polluting context, focus on what's relevant NOW.

    Modes:
    - daily_checkin: Health + overview (5-8 tools)
    - security_review: Security-focused (8-10 tools)
    - task_management: Task tools only (8-10 tools)
    - sprint_planning: Tasks + automation + PRD (12-15 tools)
    - code_review: Testing + linting (8-10 tools)
    - development: Balanced set (15-18 tools) [default]
    - debugging: Memory + testing (12-15 tools)
    - all: Full tool access (40+ tools)

    Example Prompts:
    "Switch to security review mode"
    "Focus on task management"
    "Enable the advisors tools"
    "Show current tool focus status"

    Args:
        mode: Workflow mode to switch to (see modes above)
        enable_group: Specific group to enable (health, tasks, security, etc.)
        disable_group: Specific group to disable
        status: If True, return current status without changes

    Returns:
        JSON with mode info, visible tools, and context reduction metrics
    """
    import json

    manager = get_tool_manager()

    # Status only
    if status or (mode is None and enable_group is None and disable_group is None):
        return json.dumps(manager.get_status(), indent=2)

    # This is a sync function - for async notification, wrap in tool registration
    # The actual notification happens via the async tool wrapper

    result = {}

    if mode:
        try:
            workflow_mode = WorkflowMode(mode.lower())
            # Note: notification handled by async wrapper
            old_mode = manager.current_mode
            manager.current_mode = workflow_mode
            manager.extra_groups.clear()
            manager.disabled_groups.clear()

            result = {
                "success": True,
                "action": "mode_changed",
                "mode": workflow_mode.value,
                "previous_mode": old_mode.value,
                **manager.get_status(),
            }
        except ValueError:
            result = {
                "success": False,
                "error": f"Unknown mode: {mode}",
                "available_modes": [m.value for m in WorkflowMode],
            }

    elif enable_group:
        try:
            group = ToolGroup(enable_group.lower())
            if group in manager.disabled_groups:
                manager.disabled_groups.remove(group)
            manager.extra_groups.add(group)
            result = {
                "success": True,
                "action": "group_enabled",
                "group": group.value,
                **manager.get_status(),
            }
        except ValueError:
            result = {
                "success": False,
                "error": f"Unknown group: {enable_group}",
                "available_groups": [g.value for g in ToolGroup],
            }

    elif disable_group:
        try:
            group = ToolGroup(disable_group.lower())
            if group in (ToolGroup.CORE, ToolGroup.TOOL_CATALOG):
                result = {
                    "success": False,
                    "error": f"Cannot disable {group.value} - always required",
                }
            else:
                if group in manager.extra_groups:
                    manager.extra_groups.remove(group)
                manager.disabled_groups.add(group)
                result = {
                    "success": True,
                    "action": "group_disabled",
                    "group": group.value,
                    **manager.get_status(),
                }
        except ValueError:
            result = {
                "success": False,
                "error": f"Unknown group: {disable_group}",
                "available_groups": [g.value for g in ToolGroup],
            }

    # Record mode switch for tracking
    if mode and result.get("success"):
        manager.usage_tracker.record_mode_switch(mode.lower())

    return json.dumps(result, indent=2)


def suggest_mode(
    text: Optional[str] = None,
    auto_switch: bool = False,
) -> str:
    """
    [HINT: Adaptive mode suggestion. Infers best mode from context/usage patterns.]

    🎯 Output: Suggested mode, confidence score, rationale
    🔧 Side Effects: If auto_switch=True, changes mode and notifies client
    ⏱️ Typical Runtime: <50ms

    Uses keyword analysis and usage patterns to suggest the best workflow mode.
    Call without arguments to get suggestion based on your usage history.

    Example Prompts:
    "What mode should I use for security work?"
    "Suggest a mode based on my recent activity"
    "Auto-switch to the best mode for vulnerability scanning"

    Args:
        text: Optional text to analyze for mode suggestion
        auto_switch: If True, automatically switch to suggested mode

    Returns:
        JSON with suggested mode, confidence, and rationale
    """
    manager = get_tool_manager()
    suggestion = manager.get_mode_suggestion(text, include_rationale=True)

    # Add usage stats
    suggestion["usage_stats"] = {
        "total_tool_calls": sum(manager.usage_tracker.tool_counts.values()),
        "most_used_tools": manager.usage_tracker.get_most_used_tools(5),
        "recent_tools": manager.usage_tracker.recent_tools[-5:],
    }

    # Auto-switch if requested and confident
    if auto_switch and suggestion.get("suggested_mode") and suggestion.get("confidence", 0) >= 0.5:
        try:
            workflow_mode = WorkflowMode(suggestion["suggested_mode"])
            old_mode = manager.current_mode
            manager.current_mode = workflow_mode
            manager.extra_groups.clear()
            manager.disabled_groups.clear()
            manager.usage_tracker.record_mode_switch(workflow_mode.value)

            suggestion["auto_switched"] = True
            suggestion["previous_mode"] = old_mode.value
            suggestion.update(manager.get_status())
        except ValueError:
            suggestion["auto_switched"] = False
            suggestion["error"] = "Could not auto-switch to suggested mode"
    else:
        suggestion["auto_switched"] = False

    return json.dumps(suggestion, indent=2)


def get_tool_usage_stats() -> str:
    """
    [HINT: Tool usage analytics. Shows usage patterns and co-occurrence data.]

    🎯 Output: Usage statistics, tool relationships, mode history
    🔧 Side Effects: None
    ⏱️ Typical Runtime: <10ms

    Returns:
        JSON with comprehensive usage analytics
    """
    manager = get_tool_manager()
    tracker = manager.usage_tracker

    stats = {
        "session_start": tracker.session_start,
        "total_tool_calls": sum(tracker.tool_counts.values()),
        "unique_tools_used": len(tracker.tool_counts),
        "most_used_tools": tracker.get_most_used_tools(10),
        "recent_tools": tracker.recent_tools,
        "mode_switches": dict(tracker.mode_counts),
        "preferred_mode": tracker.get_preferred_mode(),
        "current_mode": manager.current_mode.value,
        "tool_relationships": {
            tool: tracker.get_related_tools(tool, 3)
            for tool in list(tracker.tool_counts.keys())[:5]
        },
    }

    return json.dumps(stats, indent=2)


__all__ = [
    "ToolGroup",
    "WorkflowMode",
    "DynamicToolManager",
    "ToolUsageTracker",
    "get_tool_manager",
    "reset_tool_manager",
    "focus_mode",
    "suggest_mode",
    "get_tool_usage_stats",
    "TOOL_GROUP_MAPPING",
    "WORKFLOW_TOOL_GROUPS",
    "MODE_KEYWORDS",
]

