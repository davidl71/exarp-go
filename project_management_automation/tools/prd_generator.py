"""
MCP Tool for PRD (Product Requirements Document) Generation

Generates comprehensive PRDs by analyzing:
- Existing codebase structure and architecture
- Todo2 tasks (extracts user stories)
- PROJECT_GOALS.md (strategic alignment)
- Technical patterns and dependencies

Based on best practices from:
- ChatPRD (chatprd.ai)
- Cursor IDE Best Practices
"""

import json
import logging
import re
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Defined personas from project research with trusted advisors from wisdom system
PERSONAS = {
    "developer": {
        "name": "Developer",
        "role": "Daily Contributor",
        "goal": "Write quality code, stay unblocked, contribute effectively",
        "keywords": ["code", "implement", "fix", "build", "debug", "api", "endpoint", "mcp", "tool"],
        "key_metrics": ["Cyclomatic Complexity <10", "Test Coverage >80%", "Bandit Findings: 0 high/critical"],
        "workflows": ["Morning Checkin", "Before Committing", "Before PR/Push", "Weekly Self-Review"],
        "trusted_advisor": {
            "advisor": "tao_of_programming",
            "icon": "💻",
            "rationale": "The Tao of Programming teaches elegant flow - let code emerge naturally",
        },
    },
    "project_manager": {
        "name": "Project Manager",
        "role": "Delivery Focus",
        "goal": "Track progress, remove blockers, ensure delivery",
        "keywords": ["sprint", "planning", "status", "delivery", "blockers", "progress", "schedule"],
        "key_metrics": ["On-Time Delivery", "Blocked Tasks: 0", "Sprint Velocity"],
        "workflows": ["Daily Standup Prep", "Sprint Planning", "Sprint Review", "Stakeholder Update"],
        "trusted_advisor": {
            "advisor": "art_of_war",
            "icon": "⚔️",
            "rationale": "Sun Tzu teaches strategy and decisive execution - sprints are campaigns",
        },
    },
    "code_reviewer": {
        "name": "Code Reviewer",
        "role": "Quality Assurance",
        "goal": "Ensure code quality and standards compliance",
        "keywords": ["pr", "review", "approve", "merge", "quality", "standards"],
        "key_metrics": ["Review Cycle Time <24h", "Defect Escape Rate <5%"],
        "workflows": ["PR Review", "Architecture Review", "Security Review"],
        "trusted_advisor": {
            "advisor": "stoic",
            "icon": "🏛️",
            "rationale": "Stoics accept harsh truths with equanimity - reviews reveal reality",
        },
    },
    "architect": {
        "name": "Architect",
        "role": "System Design",
        "goal": "Design scalable, maintainable systems",
        "keywords": ["design", "architecture", "coupling", "patterns", "structure", "scalability"],
        "key_metrics": ["Avg Complexity <5", "Max Complexity <15", "Distance from Main Sequence <0.3"],
        "workflows": ["Weekly Architecture Review", "Before Major Changes", "Tech Debt Prioritization"],
        "trusted_advisor": {
            "advisor": "enochian",
            "icon": "🔮",
            "rationale": "Enochian mysticism reveals hidden structure and patterns in architecture",
        },
    },
    "security_engineer": {
        "name": "Security Engineer",
        "role": "Risk Management",
        "goal": "Identify and mitigate security risks",
        "keywords": ["vulnerability", "security", "cve", "audit", "scan", "risk"],
        "key_metrics": ["Critical Vulns: 0", "High Vulns: 0", "Security Score >90%"],
        "workflows": ["Daily Scan", "Weekly Deep Scan", "Security Audit"],
        "trusted_advisor": {
            "advisor": "bofh",
            "icon": "😈",
            "rationale": "BOFH is paranoid about security - expects users to break everything",
        },
    },
    "qa_engineer": {
        "name": "QA Engineer",
        "role": "Quality Assurance",
        "goal": "Ensure product quality through testing",
        "keywords": ["test", "coverage", "quality", "defect", "validation", "qa"],
        "key_metrics": ["Test Coverage >80%", "Tests Passing: 100%", "Defect Density <5/KLOC"],
        "workflows": ["Daily Testing Status", "Sprint Testing Review", "Defect Analysis"],
        "trusted_advisor": {
            "advisor": "stoic",
            "icon": "🏛️",
            "rationale": "Stoics teach discipline through adversity - tests reveal truth",
        },
    },
    "executive": {
        "name": "Executive/Stakeholder",
        "role": "Status Overview",
        "goal": "Understand project health and progress at a glance",
        "keywords": ["status", "summary", "overview", "stakeholder", "report", "dashboard"],
        "key_metrics": ["Overall Health Score", "On-Time Delivery %", "Risk Level"],
        "workflows": ["Weekly Status Review", "Stakeholder Briefing", "Executive Dashboard"],
        "trusted_advisor": {
            "advisor": "pistis_sophia",
            "icon": "📜",
            "rationale": "Pistis Sophia's journey through aeons mirrors understanding project health stages",
        },
    },
    "tech_writer": {
        "name": "Technical Writer",
        "role": "Documentation",
        "goal": "Create and maintain clear documentation",
        "keywords": ["doc", "documentation", "docs", "readme", "guide", "tutorial"],
        "key_metrics": ["Broken Links: 0", "Stale Docs: 0", "Docstring Coverage >90%"],
        "workflows": ["Weekly Doc Health", "Doc Update Cycle", "API Documentation"],
        "trusted_advisor": {
            "advisor": "confucius",
            "icon": "🎓",
            "rationale": "Confucius emphasized teaching and transmitting wisdom to future generations",
        },
    },
    "infrastructure": {
        "name": "Infrastructure/DevOps Engineer",
        "role": "System Operations",
        "goal": "Maintain reliable infrastructure, automate operations, ensure system health",
        "keywords": ["automation", "infrastructure", "devops", "ci/cd", "deployment", "monitoring", "sync", "synchronization", "duplicate", "alignment", "system"],
        "key_metrics": ["Uptime >99.9%", "Automation Coverage >80%", "Deployment Frequency", "MTTR <1h"],
        "workflows": ["Daily System Health Check", "Automation Review", "Infrastructure Updates", "Monitoring Alerts"],
        "trusted_advisor": {
            "advisor": "tao_of_programming",
            "icon": "⚙️",
            "rationale": "The Tao teaches that infrastructure should work invisibly - automation flows naturally",
        },
    },
}

# Import error handler
try:
    from ..error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
except ImportError:

    def format_success_response(data, message=None):
        return {"success": True, "data": data, "timestamp": time.time()}

    def format_error_response(error, error_code, include_traceback=False):
        return {"success": False, "error": {"code": str(error_code), "message": str(error)}}

    def log_automation_execution(name, duration, success, error=None):
        logger.info(f"{name}: {duration:.2f}s, success={success}")

    class ErrorCode:
        AUTOMATION_ERROR = "AUTOMATION_ERROR"


class PRDGenerator:
    """Generates Product Requirements Documents from codebase analysis."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.todo2_path = project_root / ".todo2" / "state.todo2.json"
        self.goals_path = project_root / "PROJECT_GOALS.md"
        self.readme_path = project_root / "README.md"

    def generate(
        self,
        project_name: Optional[str] = None,
        include_tasks: bool = True,
        include_architecture: bool = True,
        include_metrics: bool = True,
    ) -> dict[str, Any]:
        """
        Generate a comprehensive PRD.

        Args:
            project_name: Override project name (auto-detected if not provided)
            include_tasks: Include user stories from Todo2 tasks
            include_architecture: Include technical architecture analysis
            include_metrics: Include success metrics from PROJECT_GOALS.md

        Returns:
            Dict with PRD content and metadata
        """
        # Auto-detect project name
        if not project_name:
            project_name = self._detect_project_name()

        # Gather all sections
        prd_sections = {
            "project_name": project_name,
            "generated_at": datetime.now().isoformat(),
            "overview": self._generate_overview(project_name),
            "problem_statement": self._extract_problem_statement(),
            "target_users": self._extract_personas(),
            "user_stories": self._extract_user_stories() if include_tasks else [],
            "key_features": self._extract_features(),
            "technical_requirements": self._analyze_architecture() if include_architecture else {},
            "success_metrics": self._extract_metrics() if include_metrics else [],
            "risks_dependencies": self._analyze_risks(),
            "timeline": self._extract_timeline() if include_tasks else {},
        }

        # Generate markdown
        prd_markdown = self._format_prd_markdown(prd_sections)

        return {
            "sections": prd_sections,
            "markdown": prd_markdown,
            "stats": {
                "user_stories": len(prd_sections["user_stories"]),
                "features": len(prd_sections["key_features"]),
                "risks": len(prd_sections["risks_dependencies"]),
                "metrics": len(prd_sections["success_metrics"]),
                "personas": len(prd_sections["target_users"]),
            },
        }

    def _detect_project_name(self) -> str:
        """Detect project name from various sources."""
        # Try README
        if self.readme_path.exists():
            content = self.readme_path.read_text()
            # Look for # Title
            match = re.search(r"^#\s+(.+)$", content, re.MULTILINE)
            if match:
                return match.group(1).strip()

        # Try PROJECT_GOALS.md
        if self.goals_path.exists():
            content = self.goals_path.read_text()
            match = re.search(r"\*\*Project\*\*:\s*(.+)$", content, re.MULTILINE)
            if match:
                return match.group(1).strip()

        # Fallback to directory name
        return self.project_root.name

    def _generate_overview(self, project_name: str) -> dict[str, str]:
        """Generate project overview section."""
        overview = {"name": project_name, "type": "Unknown", "vision": "", "target_users": []}

        # Extract from PROJECT_GOALS.md
        if self.goals_path.exists():
            content = self.goals_path.read_text()

            # Project type
            match = re.search(r"\*\*Type\*\*:\s*(.+)$", content, re.MULTILINE)
            if match:
                overview["type"] = match.group(1).strip()

            # Vision section
            vision_match = re.search(r"## Vision\s*\n\n(.+?)(?=\n---|\n##|$)", content, re.DOTALL)
            if vision_match:
                overview["vision"] = vision_match.group(1).strip()

        # Extract from README
        if self.readme_path.exists():
            readme = self.readme_path.read_text()
            # Look for description after title
            desc_match = re.search(r"^#[^#].+?\n\n(.+?)(?=\n##|\n```|$)", readme, re.DOTALL)
            if desc_match and not overview["vision"]:
                overview["vision"] = desc_match.group(1).strip()[:500]

        return overview

    def _extract_problem_statement(self) -> str:
        """Extract or generate problem statement."""
        # Try PROJECT_GOALS.md
        if self.goals_path.exists():
            content = self.goals_path.read_text()
            # Look for Problem section or Vision
            for section in ["Problem", "Vision", "Overview"]:
                match = re.search(rf"## {section}\s*\n\n(.+?)(?=\n---|\n##|$)", content, re.DOTALL)
                if match:
                    return match.group(1).strip()

        return "Problem statement not defined. Please add a ## Vision or ## Problem section to PROJECT_GOALS.md"

    def _extract_personas(self) -> list[dict[str, Any]]:
        """Extract target user personas from project configuration."""
        personas = []

        # Use defined PERSONAS constant, filtering by relevance
        for persona_id, persona_data in PERSONAS.items():
            # Check if this persona is relevant to the project
            is_relevant = self._is_persona_relevant(persona_id, persona_data)
            if is_relevant:
                advisor_info = persona_data.get("trusted_advisor", {})
                personas.append(
                    {
                        "id": persona_id,
                        "name": persona_data["name"],
                        "role": persona_data["role"],
                        "goal": persona_data["goal"],
                        "key_metrics": persona_data["key_metrics"],
                        "workflows": persona_data["workflows"],
                        "relevance": is_relevant,
                        "trusted_advisor": advisor_info.get("advisor", "sage"),
                        "advisor_icon": advisor_info.get("icon", "🧙"),
                        "advisor_rationale": advisor_info.get("rationale", "Provides wisdom and guidance"),
                    }
                )

        return personas

    def _is_persona_relevant(self, persona_id: str, persona_data: dict) -> str:
        """Determine if a persona is relevant to this project and why."""
        relevance_reasons = []

        # Check README for keywords
        readme_content = ""
        if self.readme_path.exists():
            readme_content = self.readme_path.read_text().lower()

        # Check PROJECT_GOALS for keywords
        goals_content = ""
        if self.goals_path.exists():
            goals_content = self.goals_path.read_text().lower()

        combined = readme_content + " " + goals_content

        # Count keyword matches
        keywords = persona_data.get("keywords", [])
        matches = sum(1 for kw in keywords if kw.lower() in combined)

        if matches >= 2:
            relevance_reasons.append(f"{matches} keyword matches")

        # Special relevance checks
        if persona_id == "developer":
            relevance_reasons.append("Primary user of development tools")
        elif persona_id == "security_engineer" and "security" in combined:
            relevance_reasons.append("Project has security focus")
        elif persona_id == "tech_writer" and "documentation" in combined:
            relevance_reasons.append("Documentation is emphasized")
        elif persona_id == "project_manager" and self.todo2_path.exists():
            relevance_reasons.append("Todo2 task management present")
        elif persona_id == "qa_engineer" and any(self.project_root.glob("**/test*.py")):
            relevance_reasons.append("Test suite present")
        elif persona_id == "architect" and "architecture" in combined:
            relevance_reasons.append("Architecture documentation present")
        elif persona_id == "infrastructure" and (matches >= 2 or "automation" in combined or "ci/cd" in combined):
            relevance_reasons.append("Automation and infrastructure tasks present")

        return ", ".join(relevance_reasons) if relevance_reasons else ""

    def _extract_user_stories(self) -> list[dict[str, Any]]:
        """Extract user stories from Todo2 tasks."""
        user_stories = []

        if not self.todo2_path.exists():
            return user_stories

        try:
            state = json.loads(self.todo2_path.read_text())
            tasks = state.get("todos", [])

            for task in tasks:
                # Convert task to user story format
                story = {
                    "id": task.get("id", ""),
                    "title": task.get("name", ""),
                    "description": task.get("long_description", task.get("description", "")),
                    "priority": task.get("priority", "medium"),
                    "status": task.get("status", "Todo"),
                    "tags": task.get("tags", []),
                    # Generate user story format
                    "as_a": self._infer_user_role(task),
                    "i_want": task.get("name", ""),
                    "so_that": self._infer_benefit(task),
                }
                user_stories.append(story)

            # Sort by priority
            priority_order = {"critical": 0, "high": 1, "medium": 2, "low": 3}
            user_stories.sort(key=lambda x: priority_order.get(x["priority"], 2))

        except Exception as e:
            logger.warning(f"Error reading Todo2 tasks: {e}")

        return user_stories

    def _infer_user_role(self, task: dict) -> str:
        """Infer user role/persona from task content using defined personas."""
        content = f"{task.get('name', '')} {task.get('long_description', '')}".lower()

        # Score each persona by keyword matches
        persona_scores = {}
        for persona_id, persona_data in PERSONAS.items():
            score = sum(1 for kw in persona_data["keywords"] if kw.lower() in content)
            if score > 0:
                persona_scores[persona_id] = score

        # Return best matching persona
        if persona_scores:
            best_persona = max(persona_scores.keys(), key=lambda x: persona_scores[x])
            persona_name = PERSONAS[best_persona]["name"]
            return f"{persona_name} ({PERSONAS[best_persona]['role']})"

        # Default to developer
        return "Developer (Daily Contributor)"

    def _infer_benefit(self, task: dict) -> str:
        """Infer benefit from task content."""
        content = f"{task.get('name', '')} {task.get('long_description', '')}".lower()

        if any(kw in content for kw in ["automate", "automation", "automatic"]):
            return "I can save time through automation"
        elif any(kw in content for kw in ["security", "vulnerability", "scan"]):
            return "the project is secure and protected"
        elif any(kw in content for kw in ["test", "coverage"]):
            return "I can be confident the code works correctly"
        elif any(kw in content for kw in ["doc", "documentation"]):
            return "I can understand and use the project effectively"
        elif any(kw in content for kw in ["performance", "optimize", "speed"]):
            return "the system performs efficiently"
        else:
            return "the project capabilities are enhanced"

    def _extract_features(self) -> list[dict[str, Any]]:
        """Extract key features from codebase and docs."""
        features = []

        # Extract from README Features section
        if self.readme_path.exists():
            readme = self.readme_path.read_text()
            features_match = re.search(r"## Features?\s*\n(.+?)(?=\n## |$)", readme, re.DOTALL)
            if features_match:
                features_text = features_match.group(1)
                # Parse feature items
                for line in features_text.split("\n"):
                    line = line.strip()
                    if line.startswith(("- ", "* ", "| ")):
                        # Clean up table rows or list items
                        feature = re.sub(r"^[-*|]\s*", "", line)
                        feature = re.sub(r"\|.*$", "", feature).strip()
                        if feature and not feature.startswith("---"):
                            features.append({"name": feature[:100], "source": "README.md"})

        # Extract from tools directory
        tools_dir = self.project_root / "project_management_automation" / "tools"
        if tools_dir.exists():
            for tool_file in tools_dir.glob("*.py"):
                if tool_file.name.startswith("_"):
                    continue
                # Extract tool name from filename
                tool_name = tool_file.stem.replace("_", " ").title()
                features.append({"name": f"Tool: {tool_name}", "source": f"tools/{tool_file.name}"})

        return features[:30]  # Limit to top 30

    def _analyze_architecture(self) -> dict[str, Any]:
        """Analyze technical architecture from codebase."""
        arch = {"language": "Python", "framework": [], "dependencies": [], "structure": [], "patterns": []}

        # Analyze requirements/dependencies
        req_files = [
            self.project_root / "requirements.txt",
            self.project_root / "pyproject.toml",
            self.project_root / "setup.py",
        ]

        for req_file in req_files:
            if req_file.exists():
                content = req_file.read_text()

                # Detect frameworks
                if "fastmcp" in content.lower():
                    arch["framework"].append("FastMCP")
                if "fastapi" in content.lower():
                    arch["framework"].append("FastAPI")
                if "pydantic" in content.lower():
                    arch["framework"].append("Pydantic")
                if "pytest" in content.lower():
                    arch["framework"].append("pytest")

                # Extract dependencies from requirements.txt
                if req_file.name == "requirements.txt":
                    for line in content.split("\n"):
                        line = line.strip()
                        if line and not line.startswith("#"):
                            dep = line.split(">=")[0].split("==")[0].split("<")[0].strip()
                            if dep:
                                arch["dependencies"].append(dep)

        # Analyze directory structure
        important_dirs = ["src", "lib", "tools", "scripts", "tests", "docs", "app"]
        for dir_name in important_dirs:
            dir_path = self.project_root / dir_name
            if dir_path.exists():
                arch["structure"].append(dir_name)
            # Also check nested
            for nested in self.project_root.glob(f"*/{dir_name}"):
                if nested.is_dir():
                    arch["structure"].append(f"{nested.parent.name}/{dir_name}")

        # Detect patterns
        patterns_detected = []
        if (self.project_root / ".cursor" / "mcp.json").exists():
            patterns_detected.append("MCP Server Pattern")
        if (self.project_root / ".todo2").exists():
            patterns_detected.append("Todo2 Task Management")
        if any(self.project_root.glob("**/test*.py")):
            patterns_detected.append("Unit Testing")
        if (self.project_root / ".github" / "workflows").exists():
            patterns_detected.append("CI/CD with GitHub Actions")

        arch["patterns"] = patterns_detected

        return arch

    def _extract_metrics(self) -> list[dict[str, str]]:
        """Extract success metrics from PROJECT_GOALS.md."""
        metrics = []

        if self.goals_path.exists():
            content = self.goals_path.read_text()

            # Look for Success Metrics table
            metrics_match = re.search(r"## Success Metrics\s*\n\n(.+?)(?=\n---|\n##|$)", content, re.DOTALL)
            if metrics_match:
                table_content = metrics_match.group(1)
                # Parse table rows
                for line in table_content.split("\n"):
                    if "|" in line and not line.strip().startswith("|---"):
                        cols = [c.strip() for c in line.split("|")]
                        if len(cols) >= 3 and cols[1] and not cols[1].startswith("Metric"):
                            metrics.append(
                                {
                                    "metric": cols[1],
                                    "target": cols[2] if len(cols) > 2 else "",
                                    "current": cols[3] if len(cols) > 3 else "",
                                }
                            )

        # Add default metrics if none found
        if not metrics:
            metrics = [
                {"metric": "Code Coverage", "target": "80%", "current": "TBD"},
                {"metric": "Documentation Completeness", "target": "90%", "current": "TBD"},
                {"metric": "Tool Availability", "target": "99%", "current": "TBD"},
            ]

        return metrics

    def _analyze_risks(self) -> list[dict[str, str]]:
        """Analyze risks and dependencies."""
        risks = []

        # Extract from PROJECT_GOALS.md
        if self.goals_path.exists():
            content = self.goals_path.read_text()
            risks_match = re.search(r"## .*Risks.*\s*\n\n(.+?)(?=\n---|\n##|$)", content, re.DOTALL | re.IGNORECASE)
            if risks_match:
                for line in risks_match.group(1).split("\n"):
                    line = line.strip()
                    if line.startswith(("- ", "* ")):
                        risk = line[2:].strip()
                        risks.append(
                            {
                                "risk": risk,
                                "severity": "high" if "🔴" in risk else "medium" if "🟡" in risk else "low",
                                "source": "PROJECT_GOALS.md",
                            }
                        )

        # Analyze Todo2 for blocked tasks
        if self.todo2_path.exists():
            try:
                state = json.loads(self.todo2_path.read_text())
                blocked = [t for t in state.get("todos", []) if t.get("status") == "Blocked"]
                for task in blocked[:5]:  # Top 5 blockers
                    risks.append(
                        {
                            "risk": f"Blocked: {task.get('name', 'Unknown')}",
                            "severity": "high",
                            "source": "Todo2 blocked tasks",
                        }
                    )
            except Exception:
                pass

        # Add common risks if few detected
        if len(risks) < 3:
            common_risks = [
                {"risk": "Dependency updates may introduce breaking changes", "severity": "medium", "source": "Common"},
                {"risk": "External API changes may affect integrations", "severity": "medium", "source": "Common"},
            ]
            risks.extend(common_risks)

        return risks

    def _extract_timeline(self) -> dict[str, Any]:
        """Extract timeline from phases and tasks."""
        timeline = {"phases": [], "estimated_hours": 0, "task_counts": {"Todo": 0, "In Progress": 0, "Done": 0}}

        # Extract phases from PROJECT_GOALS.md
        if self.goals_path.exists():
            content = self.goals_path.read_text()
            phase_pattern = r"### Phase (\d+): ([^\n]+)"
            for match in re.finditer(phase_pattern, content):
                timeline["phases"].append({"number": int(match.group(1)), "name": match.group(2).strip()})

        # Count tasks by status
        if self.todo2_path.exists():
            try:
                state = json.loads(self.todo2_path.read_text())
                for task in state.get("todos", []):
                    status = task.get("status", "Todo")
                    if status in timeline["task_counts"]:
                        timeline["task_counts"][status] += 1
                    else:
                        timeline["task_counts"]["Todo"] += 1

                # Rough estimate: 2 hours per task
                timeline["estimated_hours"] = timeline["task_counts"]["Todo"] * 2
            except Exception:
                pass

        return timeline

    def _format_prd_markdown(self, sections: dict[str, Any]) -> str:
        """Format all sections into a markdown PRD."""
        lines = []

        # Header
        lines.append(f"# PRD: {sections['project_name']}")
        lines.append("")
        lines.append(f"*Generated: {sections['generated_at']}*")
        lines.append("")
        lines.append("---")
        lines.append("")

        # Overview
        overview = sections["overview"]
        lines.append("## 1. Overview")
        lines.append("")
        lines.append(f"**Project:** {overview['name']}")
        lines.append(f"**Type:** {overview['type']}")
        lines.append("")
        if overview.get("vision"):
            lines.append("### Vision")
            lines.append("")
            lines.append(overview["vision"])
            lines.append("")

        # Problem Statement
        lines.append("## 2. Problem Statement")
        lines.append("")
        lines.append(sections["problem_statement"])
        lines.append("")

        # Target Users / Personas
        if sections.get("target_users"):
            lines.append("## 3. Target Users / Personas")
            lines.append("")
            lines.append("| Persona | Role | Trusted Advisor | Goal |")
            lines.append("|---------|------|-----------------|------|")
            for persona in sections["target_users"]:
                advisor_icon = persona.get("advisor_icon", "🧙")
                advisor_name = persona.get("trusted_advisor", "sage").replace("_", " ").title()
                lines.append(
                    f"| **{persona['name']}** | {persona['role']} | {advisor_icon} {advisor_name} | {persona['goal']} |"
                )
            lines.append("")

            # Detailed persona descriptions
            lines.append("### Persona Details")
            lines.append("")
            for persona in sections["target_users"]:
                advisor_icon = persona.get("advisor_icon", "🧙")
                advisor_name = persona.get("trusted_advisor", "sage").replace("_", " ").title()
                lines.append(f"#### 👤 {persona['name']} ({persona['role']})")
                lines.append("")
                lines.append(f"**Goal:** {persona['goal']}")
                lines.append("")
                lines.append(f"**Trusted Advisor:** {advisor_icon} **{advisor_name}**")
                lines.append(f"> *{persona.get('advisor_rationale', 'Provides wisdom and guidance')}*")
                lines.append("")
                if persona.get("key_metrics"):
                    lines.append("**Key Metrics:**")
                    for metric in persona["key_metrics"]:
                        lines.append(f"- {metric}")
                    lines.append("")
                if persona.get("workflows"):
                    lines.append("**Workflows:**")
                    for workflow in persona["workflows"]:
                        lines.append(f"- {workflow}")
                    lines.append("")
                if persona.get("relevance"):
                    lines.append(f"*Relevance to this project: {persona['relevance']}*")
                    lines.append("")

        # User Stories
        if sections["user_stories"]:
            lines.append("## 4. User Stories")
            lines.append("")
            for i, story in enumerate(sections["user_stories"][:20], 1):  # Top 20
                priority_icon = {"critical": "🔴", "high": "🟠", "medium": "🟡", "low": "🟢"}.get(
                    story["priority"], "⚪"
                )
                lines.append(f"### US-{i}: {story['title']}")
                lines.append("")
                lines.append(f"**Priority:** {priority_icon} {story['priority'].upper()}")
                lines.append(f"**Status:** {story['status']}")
                if story.get("tags"):
                    lines.append(f"**Tags:** {', '.join(story['tags'])}")
                lines.append("")
                lines.append(f"*As a* {story['as_a']},")
                lines.append(f"*I want* {story['i_want']},")
                lines.append(f"*So that* {story['so_that']}.")
                lines.append("")
                if story.get("description"):
                    lines.append("**Details:**")
                    lines.append(story["description"][:500])
                    lines.append("")

            if len(sections["user_stories"]) > 20:
                lines.append(f"*... and {len(sections['user_stories']) - 20} more user stories*")
                lines.append("")

        # Key Features
        if sections["key_features"]:
            lines.append("## 5. Key Features")
            lines.append("")
            for feature in sections["key_features"]:
                lines.append(f"- {feature['name']}")
            lines.append("")

        # Technical Requirements
        tech = sections.get("technical_requirements", {})
        if tech:
            lines.append("## 6. Technical Requirements")
            lines.append("")
            lines.append(f"**Language:** {tech.get('language', 'Python')}")
            if tech.get("framework"):
                lines.append(f"**Frameworks:** {', '.join(tech['framework'])}")
            if tech.get("patterns"):
                lines.append(f"**Patterns:** {', '.join(tech['patterns'])}")
            lines.append("")
            if tech.get("dependencies"):
                lines.append("### Dependencies")
                lines.append("")
                for dep in tech["dependencies"][:15]:
                    lines.append(f"- {dep}")
                lines.append("")
            if tech.get("structure"):
                lines.append("### Project Structure")
                lines.append("")
                for dir_name in tech["structure"]:
                    lines.append(f"- `{dir_name}/`")
                lines.append("")

        # Success Metrics
        if sections["success_metrics"]:
            lines.append("## 7. Success Metrics")
            lines.append("")
            lines.append("| Metric | Target | Current |")
            lines.append("|--------|--------|---------|")
            for metric in sections["success_metrics"]:
                lines.append(f"| {metric['metric']} | {metric['target']} | {metric['current']} |")
            lines.append("")

        # Risks & Dependencies
        if sections["risks_dependencies"]:
            lines.append("## 8. Risks & Dependencies")
            lines.append("")
            for risk in sections["risks_dependencies"]:
                severity_icon = {"high": "🔴", "medium": "🟡", "low": "🟢"}.get(risk["severity"], "⚪")
                lines.append(f"- {severity_icon} {risk['risk']}")
            lines.append("")

        # Timeline
        timeline = sections.get("timeline", {})
        if timeline:
            lines.append("## 9. Timeline & Progress")
            lines.append("")
            if timeline.get("phases"):
                lines.append("### Phases")
                lines.append("")
                for phase in timeline["phases"]:
                    lines.append(f"- **Phase {phase['number']}:** {phase['name']}")
                lines.append("")

            counts = timeline.get("task_counts", {})
            if counts:
                total = sum(counts.values())
                done = counts.get("Done", 0)
                progress = (done / total * 100) if total > 0 else 0
                lines.append("### Current Progress")
                lines.append("")
                lines.append(f"- **Total Tasks:** {total}")
                lines.append(f"- **Completed:** {done} ({progress:.0f}%)")
                lines.append(f"- **In Progress:** {counts.get('In Progress', 0)}")
                lines.append(f"- **Remaining:** {counts.get('Todo', 0)}")
                if timeline.get("estimated_hours"):
                    lines.append(f"- **Estimated Hours Remaining:** {timeline['estimated_hours']}h")
                lines.append("")

        # Footer
        lines.append("---")
        lines.append("")
        lines.append("*Generated by Exarp PRD Generator*")
        lines.append("")
        lines.append("## How to Use This PRD")
        lines.append("")
        lines.append("1. **Review and refine** - This is a starting point, customize for your needs")
        lines.append("2. **Iterate with AI** - Use prompts like 'Review this PRD and suggest improvements'")
        lines.append("3. **Keep updated** - Re-run generation as project evolves")
        lines.append("4. **Align tasks** - Use `analyze_todo2_alignment` to verify task alignment with PRD")
        lines.append("")

        return "\n".join(lines)


def generate_prd(
    project_name: Optional[str] = None,
    output_path: Optional[str] = None,
    include_tasks: bool = True,
    include_architecture: bool = True,
    include_metrics: bool = True,
) -> str:
    """
    [HINT: PRD generation. Creates Product Requirements Document from codebase analysis.]

    📊 Output: Structured PRD markdown with user stories, features, requirements
    🔧 Side Effects: Creates/overwrites PRD file at output_path
    📁 Analyzes: PROJECT_GOALS.md, README.md, Todo2 tasks, codebase structure
    ⏱️ Typical Runtime: 2-10 seconds

    Example Prompt:
    "Generate a PRD for this project and save to docs/PRD.md"

    Related Tools:
    - analyze_todo2_alignment (align tasks against PRD)
    - project_scorecard (overall project health)
    - check_documentation_health (docs quality)

    Args:
        project_name: Override project name (auto-detected if not provided)
        output_path: Where to save PRD (default: docs/PRD.md)
        include_tasks: Include user stories from Todo2 tasks
        include_architecture: Include technical architecture analysis
        include_metrics: Include success metrics

    Returns:
        JSON with PRD content and statistics
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()
        generator = PRDGenerator(project_root)

        # Generate PRD
        result = generator.generate(
            project_name=project_name,
            include_tasks=include_tasks,
            include_architecture=include_architecture,
            include_metrics=include_metrics,
        )

        # Determine output path
        if output_path:
            out_path = Path(output_path)
            if not out_path.is_absolute():
                out_path = project_root / out_path
        else:
            out_path = project_root / "docs" / "PRD.md"

        # Ensure directory exists
        out_path.parent.mkdir(parents=True, exist_ok=True)

        # Write PRD
        out_path.write_text(result["markdown"])

        # Parse PRD and create tasks (if agentic-tools available)
        tasks_created = []
        try:
            from project_management_automation.utils.agentic_tools_client import parse_prd_mcp
            from project_management_automation.utils.todo2_mcp_client import create_todos_mcp

            # Parse PRD to generate tasks
            parse_result = parse_prd_mcp(
                prd_content=result["markdown"],
                project_root=project_root,
                generate_subtasks=True,
                default_priority=5,
                estimate_complexity=True
            )

            if parse_result and parse_result.get('tasks'):
                # Create tasks in Todo2
                created_ids = create_todos_mcp(
                    todos=parse_result['tasks'],
                    project_root=project_root
                )
                if created_ids:
                    tasks_created = created_ids
                    logger.info(f"Created {len(tasks_created)} tasks from PRD")
        except Exception as e:
            logger.debug(f"PRD parsing not available: {e}")

        # Prepare response
        response_data = {
            "output_path": str(out_path),
            "project_name": result["sections"]["project_name"],
            "stats": result["stats"],
            "tasks_created": len(tasks_created),
            "task_ids": tasks_created,
            "sections_generated": [
                "overview",
                "problem_statement",
                f"target_users ({result['stats']['personas']})",
                f"user_stories ({result['stats']['user_stories']})",
                f"key_features ({result['stats']['features']})",
                "technical_requirements",
                f"success_metrics ({result['stats']['metrics']})",
                f"risks_dependencies ({result['stats']['risks']})",
                "timeline",
            ],
            "tip": "Use analyze_todo2_alignment to verify tasks align with this PRD",
        }

        duration = time.time() - start_time
        log_automation_execution("generate_prd", duration, True)

        return json.dumps(format_success_response(response_data), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("generate_prd", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)
