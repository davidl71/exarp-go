"""
Session Mode Resources for MODE-002.

Provides MCP resources and tools for accessing inferred session mode information.
"""

import json
import logging
import os
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

from ..tools.dynamic_tools import DynamicToolManager, get_tool_manager
from ..tools.session_mode_inference import (
    ModeInferenceResult,
    SessionMode,
    SessionModeInference,
)
from ..utils.json_cache import JsonCacheManager

logger = logging.getLogger("exarp.resources.session")

# Cache manager for session data
_cache_manager = JsonCacheManager.get_instance()

# Storage path for session mode data
SESSION_MODE_STORAGE = Path(".exarp") / "session_mode.json"


class SessionModeStorage:
    """
    Persists inferred mode and makes it available via MCP resources/tools.
    
    Stores mode data in .exarp/session_mode.json with history tracking.
    """
    
    def __init__(self, storage_path: Optional[Path] = None):
        """Initialize storage with optional custom path."""
        self.storage_path = storage_path or SESSION_MODE_STORAGE
        self.storage_path.parent.mkdir(parents=True, exist_ok=True)
    
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
        try:
            # Load existing data
            data = self._load_data()
            
            # Create mode entry
            mode_entry = {
                "mode": mode.value,
                "confidence": confidence,
                "reasoning": reasoning,
                "metrics": metrics,
                "timestamp": datetime.now().isoformat() + "Z",
                "session_id": session_id,
            }
            
            # Update current mode
            data["current"] = mode_entry
            
            # Add to history
            if "history" not in data:
                data["history"] = []
            data["history"].append(mode_entry)
            
            # Limit history size
            if len(data["history"]) > 100:
                data["history"] = data["history"][-100:]
            
            # Save
            self._save_data(data)
            
        except Exception as e:
            logger.warning(f"Failed to save session mode: {e}")
    
    def get_current_mode(self) -> Optional[dict[str, Any]]:
        """
        Get the current inferred session mode.
        
        Returns:
            Dict with mode, confidence, reasoning, metrics, or None if not available
        """
        try:
            data = self._load_data()
            return data.get("current")
        except Exception as e:
            logger.debug(f"Failed to get current mode: {e}")
            return None
    
    def get_mode_history(self, session_id: Optional[str] = None) -> list[dict[str, Any]]:
        """
        Get mode history for a session.
        
        Args:
            session_id: Optional session identifier (default: current session)
            
        Returns:
            List of mode inference results over time
        """
        try:
            data = self._load_data()
            history = data.get("history", [])
            
            if session_id:
                return [entry for entry in history if entry.get("session_id") == session_id]
            
            return history
            
        except Exception as e:
            logger.debug(f"Failed to get mode history: {e}")
            return []
    
    def _load_data(self) -> dict[str, Any]:
        """Load data from storage file with caching."""
        if not self.storage_path.exists():
            return {"current": None, "history": []}
        
        try:
            # Use unified JSON cache
            cache = _cache_manager.get_cache(self.storage_path, enable_stats=False)
            data = cache.get_or_load()
            if not isinstance(data, dict):
                return {"current": None, "history": []}
            return data
        except Exception as e:
            logger.warning(f"Failed to load session mode data: {e}")
            return {"current": None, "history": []}
    
    def _save_data(self, data: dict[str, Any]) -> None:
        """Save data to storage file and invalidate cache."""
        try:
            with open(self.storage_path, "w") as f:
                json.dump(data, f, indent=2)
            # Invalidate cache after save
            _cache_manager.invalidate_file(self.storage_path)
        except Exception as e:
            logger.error(f"Failed to save session mode data: {e}")


# Global storage instance
_storage = SessionModeStorage()


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
    def ensure_json_string(result: Any) -> str:
        """Ensure result is always a JSON string for FastMCP compatibility."""
        if isinstance(result, str):
            return result
        elif isinstance(result, dict):
            return json.dumps(result, separators=(',', ':'))
        else:
            return json.dumps({"result": str(result)}, separators=(',', ':'))
    
    try:
        # Try to get from storage first
        current = _storage.get_current_mode()
        if current:
            return ensure_json_string(current)
        
        # Fallback: compute from current manager state
        manager = get_tool_manager()
        if hasattr(manager, "mode_inference") and hasattr(manager, "file_tracker"):
            inference_engine = SessionModeInference()
            from datetime import datetime
            session_start = datetime.fromisoformat(manager.usage_tracker.session_start)
            session_duration = (datetime.now() - session_start).total_seconds()
            
            result = inference_engine.infer_mode(
                tool_tracker=manager.usage_tracker,
                file_tracker=manager.file_tracker,
                session_duration_seconds=session_duration
            )
            
            return ensure_json_string(result.to_dict())
        
        # No data available
        return ensure_json_string({
            "inferred_mode": "UNKNOWN",
            "confidence": 0.0,
            "signals": {},
            "last_updated": datetime.now().isoformat() + "Z",
            "reasoning": ["No session data available"]
        })
        
    except Exception as e:
        logger.warning(f"Failed to get session mode resource: {e}")
        return ensure_json_string({
            "inferred_mode": "UNKNOWN",
            "confidence": 0.0,
            "signals": {},
            "last_updated": datetime.now().isoformat() + "Z",
            "reasoning": [f"Error: {str(e)}"]
        })


def infer_session_mode_tool(force_recompute: bool = False) -> dict:
    """
    MCP Tool: infer_session_mode
    
    [HINT: Session mode inference. Returns AGENT/ASK/MANUAL with confidence.]
    
    Infer current session mode from tool patterns.
    
    Args:
        force_recompute: If True, recompute even if recent result exists
        
    Returns:
        Dict with mode inference result (let FastMCP handle serialization)
    """
    try:
        manager = get_tool_manager()
        
        # Check if we have the required components
        if not hasattr(manager, "mode_inference") or not hasattr(manager, "file_tracker"):
            return {
                "error": "Mode inference not initialized. Ensure DynamicToolManager has file_tracker and mode_inference.",
                "mode": "UNKNOWN",
                "confidence": 0.0
            }
        
        # Get cached result if available and not forcing recompute
        if not force_recompute:
            current = _storage.get_current_mode()
            if current:
                # Check if recent (within last 2 minutes)
                timestamp_str = current.get("timestamp", "")
                if timestamp_str:
                    try:
                        timestamp = datetime.fromisoformat(timestamp_str.replace("Z", "+00:00"))
                        age_seconds = (datetime.now(timestamp.tzinfo) - timestamp).total_seconds()
                        if age_seconds < 120:  # 2 minutes
                            return current
                    except Exception:
                        pass  # Continue to recompute
        
        # Compute new inference
        inference_engine = SessionModeInference()
        from datetime import datetime
        session_start = datetime.fromisoformat(manager.usage_tracker.session_start)
        session_duration = (datetime.now() - session_start).total_seconds()
        
        result = inference_engine.infer_mode(
            tool_tracker=manager.usage_tracker,
            file_tracker=manager.file_tracker,
            session_duration_seconds=session_duration
        )
        
        # Save to storage
        _storage.save_mode(
            mode=result.mode,
            confidence=result.confidence,
            reasoning=result.reasoning,
            metrics=result.metrics,
            session_id=None  # Could extract from context if available
        )
        
        # Return dict directly (let FastMCP handle serialization)
        return result.to_dict()
        
    except Exception as e:
        logger.error(f"Failed to infer session mode: {e}", exc_info=True)
        return {
            "error": str(e),
            "mode": "UNKNOWN",
            "confidence": 0.0
        }


def register_session_resources(mcp) -> None:
    """
    Register session mode resources and tools with the MCP server.
    
    Usage:
        from project_management_automation.resources.session import register_session_resources
        register_session_resources(mcp)
    """
    try:
        @mcp.resource("automation://session/mode")
        def session_mode_resource() -> str:
            """Get current inferred session mode."""
            return get_session_mode_resource()
        
        @mcp.tool()
        def infer_session_mode(force_recompute: bool = False) -> dict:
            """
            [HINT: Session mode inference. Returns AGENT/ASK/MANUAL with confidence.]
            
            Infer current session mode from tool patterns.
            
            Args:
                force_recompute: If True, recompute even if recent result exists
            """
            return infer_session_mode_tool(force_recompute=force_recompute)
        
        logger.info("✅ Registered session mode resource and tool")
        
    except Exception as e:
        logger.warning(f"Could not register session resources: {e}")


__all__ = [
    "SessionModeStorage",
    "get_session_mode_resource",
    "infer_session_mode_tool",
    "register_session_resources",
]
