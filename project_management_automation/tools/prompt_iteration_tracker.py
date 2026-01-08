"""
Prompt Iteration Tracking Tool

Tracks and analyzes prompt iterations for AI workflow optimization.
Based on Cursor IDE Best Practice #7.
"""

import json
import logging
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)

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


class PromptIterationTracker:
    """Tracks prompt iterations for workflow analysis."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.log_dir = project_root / ".cursor" / "prompt_history"
        self.log_dir.mkdir(parents=True, exist_ok=True)
        self.current_log = self.log_dir / f"session_{datetime.now().strftime('%Y%m%d')}.json"

    def log_prompt(
        self,
        prompt: str,
        task_id: Optional[str] = None,
        mode: Optional[str] = None,
        outcome: Optional[str] = None,
        iteration: int = 1,
    ) -> dict[str, Any]:
        """Log a prompt iteration."""
        entry = {
            "timestamp": datetime.now().isoformat(),
            "prompt": prompt[:500],  # Truncate long prompts
            "task_id": task_id,
            "mode": mode or "unknown",
            "outcome": outcome or "pending",
            "iteration": iteration,
            "prompt_length": len(prompt),
        }

        # Load existing log
        log_data = self._load_log()
        log_data["entries"].append(entry)
        log_data["last_updated"] = datetime.now().isoformat()

        # Save
        self._save_log(log_data)

        return entry

    def analyze_iterations(self, days: int = 7) -> dict[str, Any]:
        """Analyze prompt iterations over time."""
        analysis = {
            "period_days": days,
            "total_prompts": 0,
            "by_mode": {},
            "by_outcome": {},
            "avg_iterations": 0,
            "patterns": [],
            "recommendations": [],
        }

        # Load all logs from period
        all_entries = []
        for log_file in self.log_dir.glob("session_*.json"):
            try:
                log_data = json.loads(log_file.read_text())
                all_entries.extend(log_data.get("entries", []))
            except Exception:
                pass

        analysis["total_prompts"] = len(all_entries)

        if not all_entries:
            analysis["recommendations"].append(
                "No prompt history found. Use log_prompt_iteration to track prompts."
            )
            return analysis

        # Analyze by mode
        for entry in all_entries:
            mode = entry.get("mode", "unknown")
            analysis["by_mode"][mode] = analysis["by_mode"].get(mode, 0) + 1

        # Analyze by outcome
        for entry in all_entries:
            outcome = entry.get("outcome", "unknown")
            analysis["by_outcome"][outcome] = analysis["by_outcome"].get(outcome, 0) + 1

        # Calculate average iterations per task
        task_iterations = {}
        for entry in all_entries:
            task_id = entry.get("task_id")
            if task_id:
                task_iterations[task_id] = max(
                    task_iterations.get(task_id, 0), entry.get("iteration", 1)
                )

        if task_iterations:
            analysis["avg_iterations"] = round(
                sum(task_iterations.values()) / len(task_iterations), 2
            )

        # Generate patterns
        if analysis["avg_iterations"] > 3:
            analysis["patterns"].append("High iteration count - consider more detailed initial prompts")

        agent_count = analysis["by_mode"].get("AGENT", 0)
        ask_count = analysis["by_mode"].get("ASK", 0)
        if agent_count > ask_count * 2:
            analysis["patterns"].append("Heavy AGENT usage - consider ASK for simpler queries")

        failed = analysis["by_outcome"].get("failed", 0)
        if failed > analysis["total_prompts"] * 0.2:
            analysis["patterns"].append("High failure rate - review prompt quality")

        # Generate recommendations
        if analysis["avg_iterations"] > 2:
            analysis["recommendations"].append(
                "Break down complex tasks into smaller, more specific prompts"
            )
        if not analysis["by_mode"]:
            analysis["recommendations"].append(
                "Track workflow mode (AGENT/ASK) to optimize tool selection"
            )

        return analysis

    def _load_log(self) -> dict:
        """Load current session log."""
        if self.current_log.exists():
            try:
                return json.loads(self.current_log.read_text())
            except Exception:
                pass
        return {"created": datetime.now().isoformat(), "entries": []}

    def _save_log(self, data: dict) -> None:
        """Save current session log."""
        self.current_log.write_text(json.dumps(data, indent=2))


def log_prompt_iteration(
    prompt: str,
    task_id: Optional[str] = None,
    mode: Optional[str] = None,
    outcome: Optional[str] = None,
    iteration: int = 1,
) -> str:
    """
    [HINT: Prompt logging. Track prompt iterations for workflow optimization.]

    📊 Output: Logged entry with timestamp, prompt metadata
    🔧 Side Effects: Writes to .cursor/prompt_history/
    📁 Creates: Daily session logs
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "Log this prompt iteration for my current task"

    Tracking helps identify:
    - Which prompts need multiple iterations
    - AGENT vs ASK mode effectiveness
    - Patterns in successful vs failed prompts

    Args:
        prompt: The prompt text to log
        task_id: Optional associated task ID
        mode: AGENT or ASK mode used
        outcome: success/failed/partial
        iteration: Iteration number (1 for first try)

    Returns:
        JSON with logged entry
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()
        tracker = PromptIterationTracker(project_root)

        entry = tracker.log_prompt(
            prompt=prompt,
            task_id=task_id,
            mode=mode,
            outcome=outcome,
            iteration=iteration,
        )

        duration = time.time() - start_time
        log_automation_execution("log_prompt_iteration", duration, True)

        return json.dumps(format_success_response(entry), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("log_prompt_iteration", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def analyze_prompt_iterations(days: int = 7) -> str:
    """
    [HINT: Prompt analysis. Analyzes iteration patterns for workflow optimization.]

    📊 Output: Statistics, patterns, and recommendations
    🔧 Side Effects: None (read-only)
    📁 Analyzes: .cursor/prompt_history/ logs
    ⏱️ Typical Runtime: 1-3 seconds

    Example Prompt:
    "Analyze my prompt iteration patterns from the last week"

    Provides insights on:
    - Average iterations per task
    - Mode distribution (AGENT/ASK)
    - Success/failure patterns
    - Optimization recommendations

    Args:
        days: Number of days to analyze (default: 7)

    Returns:
        JSON with analysis and recommendations
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()
        tracker = PromptIterationTracker(project_root)

        analysis = tracker.analyze_iterations(days=days)

        duration = time.time() - start_time
        log_automation_execution("analyze_prompt_iterations", duration, True)

        return json.dumps(format_success_response(analysis), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("analyze_prompt_iterations", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

