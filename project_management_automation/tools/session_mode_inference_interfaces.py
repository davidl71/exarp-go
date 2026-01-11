"""
Interface Contracts for MODE-002 Parallel Development

This file defines the interfaces that different agents will implement
to ensure compatibility during parallel development.

These are TYPE STUBS and DOCUMENTATION - actual implementations go in:
- FileEditTracker: dynamic_tools.py (Agent A)
- SessionModeInference: session_mode_inference.py (Agent B)
- Storage/Resources: resources/session.py (Agent C)
"""

from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    # Forward reference for ToolUsageTracker (defined in dynamic_tools.py)
    # This is only for type checking, not runtime
    pass

# ═══════════════════════════════════════════════════════════════════════════════
# INTERFACE 1: FileEditTracker (Agent A)
# ═══════════════════════════════════════════════════════════════════════════════

@dataclass
class FileEditTracker:
    """
    Tracks file edits during tool calls to detect multi-file vs single-file patterns.

    Interface Contract (Agent A must implement):
    - Must track edited files in a set
    - Must record timestamps for edit frequency analysis
    - Must provide methods for multi-file detection
    - Must be serializable for persistence

    Usage:
        tracker = FileEditTracker()
        tracker.record_file_edit("/path/to/file.py")
        count = tracker.get_unique_files_count()
        is_multi = tracker.is_multi_file_session(threshold=2)
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
        raise NotImplementedError("Agent A must implement this")

    def get_unique_files_count(self) -> int:
        """
        Get the number of unique files edited.

        Returns:
            Number of unique files edited in this session
        """
        raise NotImplementedError("Agent A must implement this")

    def get_edit_frequency(self, window_seconds: float = 60) -> float:
        """
        Get the frequency of file edits in edits per minute.

        Args:
            window_seconds: Time window to analyze (default: 60 seconds)

        Returns:
            Edits per minute in the specified window
        """
        raise NotImplementedError("Agent A must implement this")

    def is_multi_file_session(self, threshold: int = 2) -> bool:
        """
        Check if this session involves editing multiple files.

        Args:
            threshold: Minimum number of files to consider "multi-file" (default: 2)

        Returns:
            True if more than threshold files have been edited
        """
        raise NotImplementedError("Agent A must implement this")

    def to_dict(self) -> dict[str, Any]:
        """Serialize for persistence."""
        raise NotImplementedError("Agent A must implement this")

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "FileEditTracker":
        """Deserialize from persisted data."""
        raise NotImplementedError("Agent A must implement this")


# ═══════════════════════════════════════════════════════════════════════════════
# INTERFACE 2: SessionMode Enum (Agent B)
# ═══════════════════════════════════════════════════════════════════════════════

class SessionMode(str, Enum):
    """
    Cursor session modes inferred from tool patterns.

    Interface Contract (Agent B must implement):
    - Must be a string enum compatible with JSON serialization
    - Must include all three modes: AGENT, ASK, MANUAL
    - Must include UNKNOWN for insufficient data cases
    """
    AGENT = "agent"      # High tool frequency, multi-file, longer sessions
    ASK = "ask"          # Lower frequency, single queries, shorter
    MANUAL = "manual"    # No tool calls, direct file edits
    UNKNOWN = "unknown"  # Insufficient data


# ═══════════════════════════════════════════════════════════════════════════════
# INTERFACE 3: ModeInferenceResult (Agent B)
# ═══════════════════════════════════════════════════════════════════════════════

@dataclass
class ModeInferenceResult:
    """
    Result of mode inference with confidence and reasoning.

    Interface Contract (Agent B must implement):
    - Must include mode (SessionMode enum)
    - Must include confidence score (0.0 to 1.0)
    - Must include reasoning list (human-readable explanations)
    - Must include metrics dict (raw data used for inference)
    """
    mode: SessionMode
    confidence: float  # 0.0 to 1.0
    reasoning: list[str]
    metrics: dict[str, Any]

    def to_dict(self) -> dict[str, Any]:
        """Serialize for JSON/API responses."""
        return {
            "mode": self.mode.value,
            "confidence": self.confidence,
            "reasoning": self.reasoning,
            "metrics": self.metrics,
        }


# ═══════════════════════════════════════════════════════════════════════════════
# INTERFACE 4: SessionModeInference (Agent B)
# ═══════════════════════════════════════════════════════════════════════════════

class SessionModeInference:
    """
    Analyzes patterns to infer AGENT/ASK/MANUAL modes with confidence scores.

    Interface Contract (Agent B must implement):
    - Must accept ToolUsageTracker and FileEditTracker as inputs
    - Must return ModeInferenceResult with confidence scores
    - Must implement sliding window analysis
    - Must handle edge cases (insufficient data, etc.)

    Usage:
        inference = SessionModeInference()
        result = inference.infer_mode(
            tool_tracker=tool_tracker,
            file_tracker=file_tracker,
            session_duration_seconds=300.0
        )
        print(f"Mode: {result.mode.value}, Confidence: {result.confidence}")
    """

    def infer_mode(
        self,
        tool_tracker: Any,  # ToolUsageTracker from dynamic_tools.py
        file_tracker: FileEditTracker,
        session_duration_seconds: float
    ) -> ModeInferenceResult:
        """
        Infer session mode from tool and file patterns.

        Args:
            tool_tracker: ToolUsageTracker instance with tool call data
            file_tracker: FileEditTracker instance with file edit data
            session_duration_seconds: Current session duration in seconds

        Returns:
            ModeInferenceResult with inferred mode, confidence, and reasoning

        Detection Heuristics:
        - AGENT: High tool frequency (>5/min), multi-file (>2), longer sessions (>5min)
        - ASK: Lower frequency (1-3/min), single/few files (≤2), shorter (<5min)
        - MANUAL: Very low/no tool calls (<1/min), direct file edits
        """
        raise NotImplementedError("Agent B must implement this")

    def analyze_sliding_window(
        self,
        tool_tracker: Any,  # ToolUsageTracker from dynamic_tools.py
        file_tracker: FileEditTracker,
        window_size_seconds: float = 300,  # 5 minutes
        step_size_seconds: float = 60     # 1 minute
    ) -> list[ModeInferenceResult]:
        """
        Analyze mode changes over time using a sliding window.

        Args:
            tool_tracker: ToolUsageTracker instance
            file_tracker: FileEditTracker instance
            window_size_seconds: Size of analysis window (default: 300s = 5min)
            step_size_seconds: Step size between windows (default: 60s = 1min)

        Returns:
            List of ModeInferenceResult for each window
        """
        raise NotImplementedError("Agent B must implement this")


# ═══════════════════════════════════════════════════════════════════════════════
# INTERFACE 5: Storage Interface (Agent C)
# ═══════════════════════════════════════════════════════════════════════════════

class SessionModeStorage:
    """
    Persists inferred mode and makes it available via MCP resources/tools.

    Interface Contract (Agent C must implement):
    - Must store mode in .exarp/session_mode.json
    - Must track mode history per session
    - Must include confidence scores and reasoning
    - Must provide MCP resource: automation://session/mode
    - Must provide MCP tool: infer_session_mode
    """

    def save_mode(
        self,
        mode: SessionMode,
        confidence: float,
        reasoning: list[str],
        metrics: dict[str, Any],
        session_id: Optional[str] = None
    ) -> None:
        """
        Save inferred mode to persistent storage.

        Args:
            mode: Inferred session mode
            confidence: Confidence score (0.0 to 1.0)
            reasoning: Human-readable reasoning
            metrics: Raw metrics used for inference
            session_id: Optional session identifier
        """
        raise NotImplementedError("Agent C must implement this")

    def get_current_mode(self) -> Optional[Dict[str, Any]]:
        """
        Get the current inferred session mode.

        Returns:
            Dict with mode, confidence, reasoning, metrics, or None if not available
        """
        raise NotImplementedError("Agent C must implement this")

    def get_mode_history(self, session_id: Optional[str] = None) -> list[dict[str, Any]]:
        """
        Get mode history for a session.

        Args:
            session_id: Optional session identifier (default: current session)

        Returns:
            List of mode inference results over time
        """
        raise NotImplementedError("Agent C must implement this")


# ═══════════════════════════════════════════════════════════════════════════════
# MCP RESOURCE INTERFACE (Agent C)
# ═══════════════════════════════════════════════════════════════════════════════

def get_session_mode_resource() -> str:
    """
    MCP Resource: automation://session/mode

    Returns current inferred session mode as JSON.

    Expected JSON format:
    {
        "inferred_mode": "AGENT|ASK|MANUAL|UNKNOWN",
        "confidence": 0.0-1.0,
        "signals": {
            "tool_calls": 15,
            "files_edited": 3,
            "session_duration_sec": 300,
            "multi_file_edits": true
        },
        "last_updated": "ISO timestamp",
        "reasoning": ["reason1", "reason2"]
    }
    """
    raise NotImplementedError("Agent C must implement this")


# ═══════════════════════════════════════════════════════════════════════════════
# MCP TOOL INTERFACE (Agent C)
# ═══════════════════════════════════════════════════════════════════════════════

# LEGACY: Removed from bridge, kept for testing/backup
# This tool is now fully native Go (internal/tools/session.go)
# The Python implementation is kept for reference and testing purposes only
def infer_session_mode_tool(force_recompute: bool = False) -> str:
    """
    MCP Tool: infer_session_mode

    [HINT: Session mode inference. Returns AGENT/ASK/MANUAL with confidence.]

    Infer current session mode from tool patterns.

    Args:
        force_recompute: If True, recompute even if recent result exists

    Returns:
        JSON string with mode inference result
    """
    raise NotImplementedError("Agent C must implement this")


# ═══════════════════════════════════════════════════════════════════════════════
# INTEGRATION POINTS
# ═══════════════════════════════════════════════════════════════════════════════

# Agent A will integrate FileEditTracker into DynamicToolManager:
# - Add file_tracker: FileEditTracker field
# - Enhance record_tool_usage() to accept tool_args and extract file paths
# - Add update_inferred_mode() method that calls SessionModeInference

# Agent B will provide SessionModeInference that Agent A can import:
# - from .session_mode_inference import SessionModeInference, SessionMode, ModeInferenceResult

# Agent C will create MCP resource/tool that uses both:
# - from .dynamic_tools import DynamicToolManager
# - from .session_mode_inference import SessionModeInference
# - Register resource and tool in server.py
