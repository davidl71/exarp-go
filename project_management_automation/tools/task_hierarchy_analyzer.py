"""
Task Hierarchy Analyzer - Recommend hierarchy vs tags for task organization.

[HINT: Hierarchy analysis. Returns component groups, hierarchy recommendations,
extraction candidates, decision matrix.]

Analyzes Todo2 tasks to identify:
- Components that should have explicit hierarchies (T-COMPONENT-*)
- Components where tags are sufficient
- Extraction candidates (potential standalone packages)
"""

import json
import logging
from collections import Counter
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

from ..utils import find_project_root
from ..utils.todo2_utils import filter_tasks_by_project, get_repo_project_id, is_completed_status, is_pending_status

logger = logging.getLogger(__name__)

# Component detection patterns
COMPONENT_PATTERNS = {
    "security": {
        "keywords": ["security", "auth", "rate-limit", "boundary", "access", "credential",
                    "ssrf", "injection", "validation", "sanitize", "encrypt"],
        "extraction_potential": "high",
        "rationale": "Security layer can be extracted as standalone middleware",
    },
    "metrics": {
        "keywords": ["metric", "scorecard", "analysis", "radon", "complexity", "coverage",
                    "score", "health", "measure", "statistics", "collector"],
        "extraction_potential": "high",
        "rationale": "Metrics system can be a reusable analytics package",
    },
    "testing": {
        "keywords": ["test", "coverage", "pytest", "fixture", "mock", "assert",
                    "unittest", "snapshot", "integration"],
        "extraction_potential": "medium",
        "rationale": "Test utilities can be shared across projects",
    },
    "wisdom": {
        "keywords": ["wisdom", "quote", "sefaria", "pistis", "stoic", "inspiration",
                    "devwisdom", "daily"],
        "extraction_potential": "high",
        "rationale": "General-purpose wisdom system, not project-specific",
    },
    "ci_cd": {
        "keywords": ["ci", "cd", "workflow", "github", "hook", "pre-commit", "action",
                    "pipeline", "deploy", "build"],
        "extraction_potential": "low",
        "rationale": "CI/CD is project-specific integration",
    },
    "documentation": {
        "keywords": ["docs", "documentation", "readme", "markdown", "guide", "tutorial",
                    "reference", "api-docs"],
        "extraction_potential": "low",
        "rationale": "Documentation is cross-cutting, not extractable",
    },
    "mcp_core": {
        "keywords": ["mcp", "fastmcp", "tool", "resource", "prompt", "server", "client"],
        "extraction_potential": "none",
        "rationale": "Core MCP functionality IS the project",
    },
    "task_management": {
        "keywords": ["todo2", "task", "alignment", "duplicate", "sprint", "backlog",
                    "priority", "dependency"],
        "extraction_potential": "low",
        "rationale": "Task management is core to Exarp",
    },
}

# Hierarchy recommendation thresholds
HIERARCHY_THRESHOLDS = {
    "min_pending_tasks": 5,      # Need at least 5 pending tasks
    "min_total_tasks": 8,        # Need at least 8 total tasks
    "extraction_potential": ["high", "medium"],  # Only recommend for extractable
}


def analyze_task_hierarchy(
    output_format: str = "text",
    output_path: Optional[str] = None,
    include_recommendations: bool = True,
) -> dict[str, Any]:
    """
    Analyze tasks and recommend hierarchy vs tags organization.

    [HINT: Hierarchy analysis. Returns component groups, hierarchy recommendations,
    extraction candidates, decision matrix.]

    Args:
        output_format: Output format - "text", "json", or "markdown"
        output_path: Optional path to save report
        include_recommendations: Include specific hierarchy recommendations

    Returns:
        Dictionary with analysis results
    """
    project_root = find_project_root()

    # Load tasks
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    if not todo2_file.exists():
        return {
            "success": False,
            "error": "No .todo2/state.todo2.json found",
        }

    with open(todo2_file) as f:
        data = json.load(f)

    todos = data.get('todos', [])
    project_id = get_repo_project_id(project_root)
    todos = filter_tasks_by_project(todos, project_id, logger=logger)

    # ═══════════════════════════════════════════════════════════════
    # 1. TAG ANALYSIS
    # ═══════════════════════════════════════════════════════════════
    all_tags = []
    for t in todos:
        all_tags.extend(t.get('tags', []))

    tag_counts = Counter(all_tags)
    top_tags = tag_counts.most_common(30)

    # ═══════════════════════════════════════════════════════════════
    # 2. COMPONENT DETECTION
    # ═══════════════════════════════════════════════════════════════
    components = {}

    for comp_name, comp_info in COMPONENT_PATTERNS.items():
        keywords = comp_info["keywords"]

        # Find matching tasks
        matching_tasks = []
        for t in todos:
            task_text = ' '.join([
                t.get('content', ''),
                t.get('long_description', '') or '',
                ' '.join(t.get('tags', [])),
            ]).lower()

            if any(kw in task_text for kw in keywords):
                matching_tasks.append(t)

        # Categorize by status (normalized)
        pending = [t for t in matching_tasks if is_pending_status(t.get('status', ''))]
        completed = [t for t in matching_tasks if is_completed_status(t.get('status', ''))]

        # Check for existing hierarchy
        has_hierarchy = any(
            t.get('id', '').upper().startswith(f'T-{comp_name.upper().replace("_", "-")}')
            for t in matching_tasks
        )

        components[comp_name] = {
            "total_tasks": len(matching_tasks),
            "pending_tasks": len(pending),
            "completed_tasks": len(completed),
            "extraction_potential": comp_info["extraction_potential"],
            "rationale": comp_info["rationale"],
            "has_hierarchy": has_hierarchy,
            "sample_tasks": [t.get('content', '')[:50] for t in matching_tasks[:3]],
        }

    # ═══════════════════════════════════════════════════════════════
    # 3. HIERARCHY RECOMMENDATIONS
    # ═══════════════════════════════════════════════════════════════
    recommendations = {
        "use_hierarchy": [],
        "use_tags": [],
        "already_hierarchical": [],
    }

    for comp_name, comp_data in components.items():
        if comp_data["has_hierarchy"]:
            recommendations["already_hierarchical"].append({
                "component": comp_name,
                "reason": "Already has T-{COMPONENT}-* hierarchy",
            })
        elif (
            comp_data["pending_tasks"] >= HIERARCHY_THRESHOLDS["min_pending_tasks"] and
            comp_data["total_tasks"] >= HIERARCHY_THRESHOLDS["min_total_tasks"] and
            comp_data["extraction_potential"] in HIERARCHY_THRESHOLDS["extraction_potential"]
        ):
            recommendations["use_hierarchy"].append({
                "component": comp_name,
                "pending": comp_data["pending_tasks"],
                "total": comp_data["total_tasks"],
                "extraction_potential": comp_data["extraction_potential"],
                "rationale": comp_data["rationale"],
                "suggested_prefix": f"T-{comp_name.upper().replace('_', '-')}",
            })
        else:
            recommendations["use_tags"].append({
                "component": comp_name,
                "reason": _get_tag_reason(comp_data),
            })

    # ═══════════════════════════════════════════════════════════════
    # 4. EXTRACTION CANDIDATES
    # ═══════════════════════════════════════════════════════════════
    extraction_candidates = [
        {
            "component": comp_name,
            "package_name": f"exarp-{comp_name.replace('_', '-')}",
            "tasks": comp_data["total_tasks"],
            "pending": comp_data["pending_tasks"],
            "rationale": comp_data["rationale"],
        }
        for comp_name, comp_data in components.items()
        if comp_data["extraction_potential"] == "high" and comp_data["total_tasks"] >= 5
    ]

    # ═══════════════════════════════════════════════════════════════
    # 5. DECISION MATRIX
    # ═══════════════════════════════════════════════════════════════
    decision_matrix = []
    for comp_name, comp_data in sorted(components.items(), key=lambda x: -x[1]["total_tasks"]):
        use_hierarchy = any(r["component"] == comp_name for r in recommendations["use_hierarchy"])
        already_has = comp_data["has_hierarchy"]

        decision_matrix.append({
            "component": comp_name,
            "total": comp_data["total_tasks"],
            "pending": comp_data["pending_tasks"],
            "hierarchy": "✅ YES" if use_hierarchy else ("✅ Done" if already_has else "❌ Tags"),
            "extraction": comp_data["extraction_potential"],
            "rationale": comp_data["rationale"][:40] + "..." if len(comp_data["rationale"]) > 40 else comp_data["rationale"],
        })

    # ═══════════════════════════════════════════════════════════════
    # BUILD RESULT
    # ═══════════════════════════════════════════════════════════════
    result = {
        "success": True,
        "generated_at": datetime.now().isoformat(),
        "total_tasks": len(todos),
        "total_tags": len(tag_counts),
        "top_tags": top_tags[:15],
        "components": components,
        "recommendations": recommendations,
        "extraction_candidates": extraction_candidates,
        "decision_matrix": decision_matrix,
    }

    # Format output
    if output_format == "json":
        formatted_output = json.dumps(result, indent=2)
    elif output_format == "markdown":
        formatted_output = _format_markdown(result)
    else:
        formatted_output = _format_text(result)

    result["formatted_output"] = formatted_output

    # Save if requested
    if output_path:
        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)
        output_file.write_text(formatted_output)
        result["output_file"] = str(output_file)

    return result


def _get_tag_reason(comp_data: dict) -> str:
    """Get reason why tags are sufficient for this component."""
    if comp_data["extraction_potential"] == "none":
        return "Core project functionality, not extractable"
    elif comp_data["extraction_potential"] == "low":
        return "Cross-cutting concern or project-specific"
    elif comp_data["pending_tasks"] < HIERARCHY_THRESHOLDS["min_pending_tasks"]:
        return f"Too few pending tasks ({comp_data['pending_tasks']})"
    elif comp_data["total_tasks"] < HIERARCHY_THRESHOLDS["min_total_tasks"]:
        return f"Too few total tasks ({comp_data['total_tasks']})"
    else:
        return "Tags sufficient for organization"


def _format_text(data: dict) -> str:
    """Format analysis as plain text."""
    lines = []
    lines.append("=" * 70)
    lines.append("  📊 TASK HIERARCHY ANALYSIS")
    lines.append(f"  Generated: {data['generated_at'][:16].replace('T', ' ')}")
    lines.append(f"  Total Tasks: {data['total_tasks']} | Unique Tags: {data['total_tags']}")
    lines.append("=" * 70)

    # Decision Matrix
    lines.append("\n  📋 DECISION MATRIX")
    lines.append("  " + "-" * 66)
    lines.append(f"  {'Component':<15} {'Total':>6} {'Pending':>8} {'Hierarchy':>12} {'Extract':>8}")
    lines.append("  " + "-" * 66)

    for row in data["decision_matrix"]:
        lines.append(
            f"  {row['component']:<15} {row['total']:>6} {row['pending']:>8} "
            f"{row['hierarchy']:>12} {row['extraction']:>8}"
        )

    # Recommendations
    if data["recommendations"]["use_hierarchy"]:
        lines.append("\n  ✅ RECOMMEND HIERARCHY (T-COMPONENT-*)")
        lines.append("  " + "-" * 66)
        for rec in data["recommendations"]["use_hierarchy"]:
            lines.append(f"    • {rec['component']}: {rec['pending']} pending tasks")
            lines.append(f"      Prefix: {rec['suggested_prefix']}-*")
            lines.append(f"      Reason: {rec['rationale']}")

    if data["recommendations"]["already_hierarchical"]:
        lines.append("\n  ✓ ALREADY HIERARCHICAL")
        for rec in data["recommendations"]["already_hierarchical"]:
            lines.append(f"    • {rec['component']}: {rec['reason']}")

    # Extraction Candidates
    if data["extraction_candidates"]:
        lines.append("\n  📦 EXTRACTION CANDIDATES")
        lines.append("  " + "-" * 66)
        for cand in data["extraction_candidates"]:
            lines.append(f"    • {cand['package_name']}: {cand['tasks']} tasks")
            lines.append(f"      {cand['rationale']}")

    lines.append("\n" + "=" * 70)
    return "\n".join(lines)


def _format_markdown(data: dict) -> str:
    """Format analysis as markdown."""
    lines = []
    lines.append("# 📊 Task Hierarchy Analysis")
    lines.append(f"\n*Generated: {data['generated_at'][:16].replace('T', ' ')}*")
    lines.append(f"\n**Total Tasks:** {data['total_tasks']} | **Unique Tags:** {data['total_tags']}")

    # Decision Matrix
    lines.append("\n## Decision Matrix\n")
    lines.append("| Component | Total | Pending | Hierarchy? | Extraction | Rationale |")
    lines.append("|-----------|-------|---------|------------|------------|-----------|")

    for row in data["decision_matrix"]:
        lines.append(
            f"| {row['component']} | {row['total']} | {row['pending']} | "
            f"{row['hierarchy']} | {row['extraction']} | {row['rationale']} |"
        )

    # Recommendations
    if data["recommendations"]["use_hierarchy"]:
        lines.append("\n## ✅ Recommended Hierarchies\n")
        for rec in data["recommendations"]["use_hierarchy"]:
            lines.append(f"### {rec['component'].title()} (`{rec['suggested_prefix']}-*`)\n")
            lines.append(f"- **Pending tasks:** {rec['pending']}")
            lines.append(f"- **Extraction potential:** {rec['extraction_potential']}")
            lines.append(f"- **Rationale:** {rec['rationale']}")
            lines.append("")

    if data["recommendations"]["already_hierarchical"]:
        lines.append("\n## ✓ Already Hierarchical\n")
        for rec in data["recommendations"]["already_hierarchical"]:
            lines.append(f"- **{rec['component']}:** {rec['reason']}")

    # Extraction Candidates
    if data["extraction_candidates"]:
        lines.append("\n## 📦 Extraction Candidates\n")
        for cand in data["extraction_candidates"]:
            lines.append(f"- **{cand['package_name']}** ({cand['tasks']} tasks): {cand['rationale']}")

    # Top Tags
    lines.append("\n## 🏷️ Top Tags\n")
    lines.append("| Tag | Count |")
    lines.append("|-----|-------|")
    for tag, count in data["top_tags"]:
        lines.append(f"| {tag} | {count} |")

    return "\n".join(lines)


# CLI support
if __name__ == "__main__":
    import sys

    output_format = sys.argv[1] if len(sys.argv) > 1 else "text"
    result = analyze_task_hierarchy(output_format=output_format)

    if result.get("success"):
        print(result["formatted_output"])
    else:
        print(f"Error: {result.get('error')}")

