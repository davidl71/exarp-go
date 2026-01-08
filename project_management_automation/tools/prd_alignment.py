"""
PRD-Based Alignment Tool

Enhances task alignment by checking against:
1. PROJECT_GOALS.md (strategic phases)
2. PRD.md (user stories, personas, features)

This tool provides persona-aware alignment scoring.
"""

import json
import logging
import re
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


# Import personas from PRD generator
try:
    from .prd_generator import PERSONAS
except ImportError:
    PERSONAS = {}


class PRDAlignmentAnalyzer:
    """Analyzes task alignment against PRD personas and user stories."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.prd_path = project_root / "docs" / "PRD.md"
        self.goals_path = project_root / "PROJECT_GOALS.md"
        self.todo2_path = project_root / ".todo2" / "state.todo2.json"

    def analyze(self) -> dict[str, Any]:
        """
        Analyze all tasks against PRD.

        Returns:
            Dict with alignment analysis results
        """
        results = {
            "timestamp": datetime.now().isoformat(),
            "prd_exists": self.prd_path.exists(),
            "goals_exists": self.goals_path.exists(),
            "tasks_analyzed": 0,
            "persona_coverage": {},
            "unaligned_tasks": [],
            "alignment_by_persona": {},
            "recommendations": [],
        }

        if not self.prd_path.exists():
            results["recommendations"].append(
                "Run generate_prd to create PRD.md for persona-based alignment"
            )
            return results

        # Load tasks
        tasks = self._load_tasks()
        results["tasks_analyzed"] = len(tasks)

        # Parse PRD for user stories
        prd_user_stories = self._parse_prd_user_stories()
        results["prd_user_stories"] = len(prd_user_stories)

        # Analyze each task
        persona_counts = dict.fromkeys(PERSONAS.keys(), 0)
        aligned_tasks = []
        unaligned_tasks = []

        for task in tasks:
            alignment = self._analyze_task_alignment(task, prd_user_stories)

            if alignment["persona"]:
                persona_counts[alignment["persona"]] = (
                    persona_counts.get(alignment["persona"], 0) + 1
                )
                aligned_tasks.append(
                    {
                        "id": task.get("id"),
                        "name": task.get("name"),
                        "persona": alignment["persona"],
                        "advisor": alignment["advisor"],
                        "alignment_score": alignment["score"],
                    }
                )
            else:
                unaligned_tasks.append(
                    {
                        "id": task.get("id"),
                        "name": task.get("name"),
                        "reason": alignment["reason"],
                    }
                )

        results["persona_coverage"] = persona_counts
        results["aligned_count"] = len(aligned_tasks)
        results["unaligned_count"] = len(unaligned_tasks)
        results["unaligned_tasks"] = unaligned_tasks[:10]  # Top 10

        # Calculate alignment score
        if tasks:
            results["overall_alignment_score"] = round(
                len(aligned_tasks) / len(tasks) * 100, 1
            )
        else:
            results["overall_alignment_score"] = 0

        # Generate recommendations
        results["recommendations"] = self._generate_recommendations(
            persona_counts, unaligned_tasks
        )

        # Add persona details
        results["alignment_by_persona"] = {
            pid: {
                "name": PERSONAS.get(pid, {}).get("name", pid),
                "count": count,
                "advisor": PERSONAS.get(pid, {}).get("trusted_advisor", {}).get("advisor", "sage"),
            }
            for pid, count in persona_counts.items()
            if count > 0
        }

        return results

    def _load_tasks(self) -> list[dict]:
        """Load tasks from Todo2."""
        if not self.todo2_path.exists():
            return []

        try:
            state = json.loads(self.todo2_path.read_text())
            return state.get("todos", [])
        except Exception as e:
            logger.warning(f"Error loading tasks: {e}")
            return []

    def _parse_prd_user_stories(self) -> list[dict]:
        """Parse user stories from PRD.md."""
        user_stories = []

        if not self.prd_path.exists():
            return user_stories

        try:
            content = self.prd_path.read_text()

            # Find User Stories section
            stories_match = re.search(
                r"## \d+\. User Stories\s*\n(.+?)(?=\n## \d+\.|\Z)",
                content,
                re.DOTALL,
            )

            if stories_match:
                stories_text = stories_match.group(1)

                # Parse individual stories
                story_pattern = r"### US-(\d+): (.+?)\n"
                for match in re.finditer(story_pattern, stories_text):
                    user_stories.append(
                        {
                            "id": f"US-{match.group(1)}",
                            "title": match.group(2).strip(),
                        }
                    )

        except Exception as e:
            logger.warning(f"Error parsing PRD: {e}")

        return user_stories

    def _analyze_task_alignment(
        self, task: dict, prd_user_stories: list[dict]
    ) -> dict[str, Any]:
        """Analyze a single task's alignment with personas."""
        content = f"{task.get('name', '')} {task.get('long_description', '')}".lower()
        tags = [t.lower() for t in task.get("tags", [])]

        # Score each persona
        best_persona = None
        best_score = 0
        best_advisor = None

        for persona_id, persona_data in PERSONAS.items():
            score = 0

            # Keyword matching
            keywords = persona_data.get("keywords", [])
            keyword_matches = sum(1 for kw in keywords if kw.lower() in content)
            score += keyword_matches * 2

            # Tag matching
            tag_matches = sum(1 for kw in keywords if kw.lower() in tags)
            score += tag_matches * 3

            if score > best_score:
                best_score = score
                best_persona = persona_id
                best_advisor = persona_data.get("trusted_advisor", {}).get(
                    "advisor", "sage"
                )

        if best_score >= 2:  # Threshold for alignment
            return {
                "persona": best_persona,
                "advisor": best_advisor,
                "score": best_score,
                "reason": None,
            }
        else:
            return {
                "persona": None,
                "advisor": None,
                "score": 0,
                "reason": "No strong persona match found",
            }

    def _generate_recommendations(
        self, persona_counts: dict, unaligned_tasks: list
    ) -> list[str]:
        """Generate alignment recommendations."""
        recommendations = []

        # Check for underrepresented personas
        for persona_id, count in persona_counts.items():
            if count == 0:
                persona_name = PERSONAS.get(persona_id, {}).get("name", persona_id)
                recommendations.append(
                    f"No tasks aligned with {persona_name} persona - consider their needs"
                )

        # Check unaligned ratio
        total = sum(persona_counts.values()) + len(unaligned_tasks)
        if total > 0:
            unaligned_ratio = len(unaligned_tasks) / total
            if unaligned_ratio > 0.2:
                recommendations.append(
                    f"{len(unaligned_tasks)} tasks ({unaligned_ratio:.0%}) lack persona alignment - add relevant tags"
                )

        # Check for imbalance
        if persona_counts:
            max_count = max(persona_counts.values())
            min_count = min(persona_counts.values())
            if max_count > 0 and min_count == 0:
                recommendations.append(
                    "Task distribution is unbalanced across personas"
                )

        return recommendations


def analyze_prd_alignment(output_path: Optional[str] = None) -> str:
    """
    [HINT: PRD alignment. Task-to-persona mapping, advisor assignments, recommendations.]

    📊 Output: Alignment scores by persona, unaligned tasks, recommendations
    🔧 Side Effects: None (read-only analysis)
    📁 Analyzes: PRD.md, Todo2 tasks, PROJECT_GOALS.md
    ⏱️ Typical Runtime: 1-5 seconds

    Example Prompt:
    "Analyze how well my tasks align with PRD personas"

    Related Tools:
    - generate_prd (create PRD first)
    - analyze_todo2_alignment (strategic alignment)
    - consult_advisor (get persona-specific wisdom)

    Args:
        output_path: Optional path to save detailed report

    Returns:
        JSON with alignment analysis
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()
        analyzer = PRDAlignmentAnalyzer(project_root)
        results = analyzer.analyze()

        # Save report if requested
        if output_path:
            out_path = Path(output_path)
            if not out_path.is_absolute():
                out_path = project_root / out_path
            out_path.parent.mkdir(parents=True, exist_ok=True)
            out_path.write_text(json.dumps(results, indent=2))
            results["report_path"] = str(out_path)

        duration = time.time() - start_time
        log_automation_execution("analyze_prd_alignment", duration, True)

        return json.dumps(format_success_response(results), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("analyze_prd_alignment", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

