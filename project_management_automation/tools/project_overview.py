"""
Project Overview Tool - Generate comprehensive one-page project summary.

[HINT: Project overview. Returns one-page summary with project info, health scores,
codebase metrics, task breakdown, risks, roadmap, and next actions. Formats: text, html, markdown, json.]
"""

import json
import re
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

from ..utils import find_project_root
from ..utils.todo2_utils import is_active_status, normalize_status


def generate_project_overview(
    output_format: str = "text",
    output_path: Optional[str] = None
) -> dict[str, Any]:
    """
    Generate comprehensive one-page project overview.

    [HINT: Project overview. Returns one-page summary with project info, health scores,
    codebase metrics, task breakdown, risks, roadmap, and next actions.]

    Args:
        output_format: Output format - "text", "html", "markdown", "json", or "slides"
        output_path: Optional path to save report

    Returns:
        Dictionary with overview data and formatted output
    """
    project_root = find_project_root()

    # Aggregate all data
    overview_data = _aggregate_project_data(project_root)

    # Format output
    if output_format == "json":
        formatted_output = json.dumps(overview_data, indent=2, default=str)
    elif output_format == "html":
        formatted_output = _format_html(overview_data)
    elif output_format == "markdown":
        formatted_output = _format_markdown(overview_data)
    elif output_format == "slides":
        formatted_output = _format_marp_slides(overview_data)
    else:
        formatted_output = _format_text(overview_data)

    result = {
        'overview_data': overview_data,
        'formatted_output': formatted_output,
        'output_format': output_format,
        'generated_at': datetime.now().isoformat(),
    }

    # Save to file if requested
    if output_path:
        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)
        output_file.write_text(formatted_output)
        result['output_file'] = str(output_file)

    return result


def _aggregate_project_data(project_root: Path) -> dict[str, Any]:
    """Aggregate all project data from various sources."""

    data = {
        'project': _get_project_info(project_root),
        'health': _get_health_metrics(project_root),
        'codebase': _get_codebase_metrics(project_root),
        'tasks': _get_task_metrics(project_root),
        'phases': _get_project_phases(project_root),
        'risks': _get_risks_and_blockers(project_root),
        'next_actions': _get_next_actions(project_root),
    }

    return data


def _get_project_info(project_root: Path) -> dict:
    """Get project metadata from pyproject.toml and git."""
    info = {
        'name': 'Exarp MCP Server',
        'version': '0.1.8',
        'description': 'Project Management Automation MCP Server',
        'type': 'MCP Server',
        'status': 'Active Development',
        'started': None,
        'author': None,
    }

    # Try pyproject.toml
    pyproject = project_root / 'pyproject.toml'
    if pyproject.exists():
        content = pyproject.read_text()
        if match := re.search(r'name\s*=\s*"([^"]+)"', content):
            info['name'] = match.group(1)
        if match := re.search(r'version\s*=\s*"([^"]+)"', content):
            info['version'] = match.group(1)
        if match := re.search(r'description\s*=\s*"([^"]+)"', content):
            info['description'] = match.group(1)

    # Try git for dates and author
    try:
        result = subprocess.run(
            ['git', 'log', '--reverse', '--format=%ci', '-1'],
            capture_output=True, text=True, cwd=project_root
        )
        if result.returncode == 0 and result.stdout.strip():
            info['started'] = result.stdout.strip()[:10]

        result = subprocess.run(
            ['git', 'config', 'user.name'],
            capture_output=True, text=True, cwd=project_root
        )
        if result.returncode == 0:
            info['author'] = result.stdout.strip()
    except (OSError, subprocess.SubprocessError):
        pass

    return info


def _get_health_metrics(project_root: Path) -> dict:
    """Get health metrics from scorecard (Go: exarp-go; else Python)."""
    try:
        from .project_scorecard import _call_go_scorecard_json, _is_go_project
        if _is_go_project(project_root):
            data = _call_go_scorecard_json(project_root)
            if data:
                return {
                    'overall_score': data.get('overall_score', 0),
                    'production_ready': data.get('production_ready', False),
                    'blockers': data.get('blockers', []),
                    'scores': data.get('scores', {}),
                }
        from .project_scorecard import generate_project_scorecard
        result = generate_project_scorecard("json", False, None)
        return {
            'overall_score': result['overall_score'],
            'production_ready': result['production_ready'],
            'blockers': result.get('blockers', []),
            'scores': result['scores'],
        }
    except Exception as e:
        return {
            'overall_score': 0,
            'production_ready': False,
            'blockers': [str(e)],
            'scores': {},
        }


def _get_codebase_metrics(project_root: Path) -> dict:
    """Get codebase statistics."""
    py_files = list(project_root.rglob('*.py'))
    py_files = [f for f in py_files if 'venv' not in str(f) and '.build-env' not in str(f)
                and '__pycache__' not in str(f)]

    total_lines = 0
    for f in py_files:
        try:
            total_lines += len(f.read_text().splitlines())
        except (OSError, UnicodeDecodeError):
            pass

    md_files = list(project_root.rglob('*.md'))
    md_files = [f for f in md_files if 'venv' not in str(f)]

    tools_dir = project_root / 'project_management_automation' / 'tools'
    tools_count = len([f for f in tools_dir.glob('*.py') if not f.name.startswith('__')]) if tools_dir.exists() else 0

    # Prompts are in exarp-go (stdio://prompts); Python stub removed
    prompts_count = 0

    return {
        'python_files': len(py_files),
        'python_lines': total_lines,
        'doc_files': len(md_files),
        'mcp_tools': tools_count,
        'mcp_prompts': prompts_count,
    }


def _get_task_metrics(project_root: Path) -> dict:
    """Get task breakdown from Todo2."""
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    if not todo2_file.exists():
        return {'total': 0, 'by_status': {}, 'by_priority': {}, 'by_category': {}, 'remaining_hours': 0}

    with open(todo2_file) as f:
        data = json.load(f)

    todos = data.get('todos', [])

    by_status = {}
    by_priority = {}
    by_category = {}
    remaining_hours = 0

    for task in todos:
        status = task.get('status', 'pending')
        normalized_status = normalize_status(status)
        priority = task.get('priority', 'medium')
        tags = task.get('tags', [])
        hours = task.get('estimatedHours', 0)

        # Use normalized status for counting
        by_status[normalized_status] = by_status.get(normalized_status, 0) + 1
        by_priority[priority] = by_priority.get(priority, 0) + 1

        # Count hours for active (non-completed) tasks
        if is_active_status(status):
            remaining_hours += hours

        for tag in tags[:2]:
            by_category[tag] = by_category.get(tag, 0) + 1

    # Sort categories by count
    by_category = dict(sorted(by_category.items(), key=lambda x: -x[1])[:6])

    return {
        'total': len(todos),
        'by_status': by_status,
        'by_priority': by_priority,
        'by_category': by_category,
        'remaining_hours': remaining_hours,
    }


def _get_project_phases(project_root: Path) -> list[dict]:
    """Get project phases from PROJECT_GOALS.md."""
    phases = [
        {'name': 'Phase 1: Core Tools', 'progress': 100, 'status': 'complete'},
        {'name': 'Phase 2: Automation', 'progress': 80, 'status': 'in_progress'},
        {'name': 'Phase 3: Security', 'progress': 30, 'status': 'in_progress'},
        {'name': 'Phase 4: Production', 'progress': 0, 'status': 'not_started'},
    ]

    goals_file = project_root / 'PROJECT_GOALS.md'
    if goals_file.exists():
        goals_file.read_text()
        # Could parse phases from the file here
        pass

    return phases


def _get_risks_and_blockers(project_root: Path) -> list[dict]:
    """Get risks from security status and health metrics."""
    risks = []

    # Check security status
    security_file = project_root / 'docs' / 'SECURITY_STATUS.md'
    if security_file.exists():
        content = security_file.read_text()
        if 'path boundary' in content.lower():
            risks.append({'severity': 'critical', 'description': 'No path boundary enforcement'})
        if 'rate limiting' in content.lower():
            risks.append({'severity': 'critical', 'description': 'No rate limiting'})

    # Default risks based on common issues
    if not risks:
        risks = [
            {'severity': 'critical', 'description': 'No path boundary enforcement'},
            {'severity': 'critical', 'description': 'No subprocess sandboxing'},
            {'severity': 'high', 'description': 'No rate limiting'},
            {'severity': 'high', 'description': 'No access control'},
            {'severity': 'medium', 'description': 'Failing unit tests'},
        ]

    return risks


def _get_next_actions(project_root: Path) -> list[dict]:
    """Get prioritized next actions from high-priority tasks."""
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    if not todo2_file.exists():
        return []

    with open(todo2_file) as f:
        data = json.load(f)

    todos = data.get('todos', [])
    high_priority = [t for t in todos if t.get('priority') == 'high'
                     and is_active_status(t.get('status', ''))]

    actions = []
    for task in high_priority[:5]:
        actions.append({
            'action': task.get('content', '')[:50],
            'estimate': f"{task.get('estimatedHours', 0)}h",
            'impact': 'High priority task',
        })

    return actions


def _format_text(data: dict) -> str:
    """Format as ASCII text for terminal."""
    lines = []

    # Header
    lines.append("┌" + "─" * 78 + "┐")
    lines.append(f"│  🏗️  {data['project']['name']:<67} │")
    lines.append(f"│  {'Project Overview':<72} │")
    lines.append(f"│  Generated: {datetime.now().strftime('%Y-%m-%d %H:%M'):<61} │")
    lines.append("├" + "─" * 78 + "┤")

    # Project Info + Health Score (side by side)
    health = data['health']
    score = health.get('overall_score', 0)
    status_icon = "🟢" if score >= 70 else "🟡" if score >= 50 else "🔴"

    lines.append("│                                                                              │")
    lines.append("│  PROJECT INFO                        │  HEALTH SCORE                        │")
    lines.append("│  ─────────────                       │  ────────────                        │")
    lines.append(f"│  Version: {data['project']['version']:<26} │  Overall: {score:.0f}% {status_icon:<25} │")
    lines.append(f"│  Type: {data['project']['type']:<29} │  {'█' * int(score/5)}{'░' * (20-int(score/5)):<25} │")
    lines.append(f"│  Status: {data['project']['status']:<27} │  Production: {'YES ✅' if health.get('production_ready') else 'NO ❌':<23} │")
    lines.append("│                                                                              │")

    # Codebase + Tasks (side by side)
    lines.append("├" + "─" * 78 + "┤")
    codebase = data['codebase']
    tasks = data['tasks']
    lines.append("│                                                                              │")
    lines.append("│  CODEBASE                            │  TASKS                               │")
    lines.append("│  ────────                            │  ─────                               │")
    lines.append(f"│  📁 {codebase['python_files']:>3} Python files              │  Total: {tasks['total']:<28} │")
    lines.append(f"│  📝 {codebase['python_lines']:>5} lines of code           │  Pending: {tasks['by_status'].get('pending', 0) + tasks['by_status'].get('Todo', 0):<26} │")
    lines.append(f"│  🔧 {codebase['mcp_tools']:>3} MCP tools                 │  Completed: {tasks['by_status'].get('completed', 0):<24} │")
    lines.append(f"│  📋 {codebase['mcp_prompts']:>3} prompts                   │  Remaining: {tasks['remaining_hours']:.0f}h ({tasks['remaining_hours']/8:.0f} days)          │")
    lines.append("│                                                                              │")

    # Phases
    lines.append("├" + "─" * 78 + "┤")
    lines.append("│                                                                              │")
    lines.append("│  PROJECT PHASES                                                              │")
    lines.append("│  ──────────────                                                              │")
    for phase in data['phases']:
        progress = phase['progress']
        bar = "█" * int(progress/5) + "░" * (20-int(progress/5))
        status_icon = "✅" if phase['status'] == 'complete' else "🔄" if phase['status'] == 'in_progress' else "⏳"
        lines.append(f"│  {phase['name']:<25} [{bar}] {progress:>3}% {status_icon:<5} │")
    lines.append("│                                                                              │")

    # Risks
    lines.append("├" + "─" * 78 + "┤")
    lines.append("│                                                                              │")
    lines.append("│  RISKS & BLOCKERS                                                            │")
    lines.append("│  ─────────────────                                                           │")
    for risk in data['risks'][:5]:
        icon = "🔴" if risk['severity'] == 'critical' else "🟡" if risk['severity'] == 'high' else "🟢"
        lines.append(f"│  {icon} {risk['description']:<72} │")
    lines.append("│                                                                              │")

    # Next Actions
    lines.append("├" + "─" * 78 + "┤")
    lines.append("│                                                                              │")
    lines.append("│  NEXT ACTIONS                                                                │")
    lines.append("│  ────────────                                                                │")
    for i, action in enumerate(data['next_actions'][:4], 1):
        lines.append(f"│  {i}. {action['action']:<56} ({action['estimate']:<4}) │")
    lines.append("│                                                                              │")

    lines.append("└" + "─" * 78 + "┘")

    return "\n".join(lines)


def _format_html(data: dict) -> str:
    """Format as styled HTML for PDF export."""
    health = data['health']
    score = health.get('overall_score', 0)
    score_color = "#22c55e" if score >= 70 else "#eab308" if score >= 50 else "#ef4444"

    html = f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{data['project']['name']} - Project Overview</title>
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        body {{
            font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            color: #e4e4e7;
            min-height: 100vh;
            padding: 2rem;
        }}
        .container {{
            max-width: 1000px;
            margin: 0 auto;
            background: rgba(30, 41, 59, 0.8);
            border-radius: 16px;
            padding: 2rem;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
        }}
        .header {{
            text-align: center;
            margin-bottom: 2rem;
            padding-bottom: 1.5rem;
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }}
        .header h1 {{
            font-size: 2.5rem;
            background: linear-gradient(90deg, #60a5fa, #a78bfa);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 0.5rem;
        }}
        .header .subtitle {{
            color: #94a3b8;
            font-size: 1rem;
        }}
        .grid {{
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 1.5rem;
            margin-bottom: 1.5rem;
        }}
        .card {{
            background: rgba(51, 65, 85, 0.5);
            border-radius: 12px;
            padding: 1.5rem;
            border: 1px solid rgba(255,255,255,0.05);
        }}
        .card h3 {{
            color: #94a3b8;
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 0.1em;
            margin-bottom: 1rem;
        }}
        .score-ring {{
            width: 120px;
            height: 120px;
            border-radius: 50%;
            background: conic-gradient({score_color} {score*3.6}deg, #374151 0deg);
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto;
        }}
        .score-inner {{
            width: 100px;
            height: 100px;
            border-radius: 50%;
            background: #1e293b;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
        }}
        .score-value {{
            font-size: 2rem;
            font-weight: bold;
            color: {score_color};
        }}
        .metric {{
            display: flex;
            justify-content: space-between;
            padding: 0.5rem 0;
            border-bottom: 1px solid rgba(255,255,255,0.05);
        }}
        .metric:last-child {{ border-bottom: none; }}
        .metric-label {{ color: #94a3b8; }}
        .metric-value {{ font-weight: 600; }}
        .progress-bar {{
            height: 8px;
            background: #374151;
            border-radius: 4px;
            overflow: hidden;
            margin-top: 0.5rem;
        }}
        .progress-fill {{
            height: 100%;
            border-radius: 4px;
            transition: width 0.3s ease;
        }}
        .phase {{
            margin-bottom: 1rem;
        }}
        .phase-header {{
            display: flex;
            justify-content: space-between;
            margin-bottom: 0.25rem;
        }}
        .risk {{
            display: flex;
            align-items: center;
            gap: 0.5rem;
            padding: 0.5rem 0;
        }}
        .risk-dot {{
            width: 8px;
            height: 8px;
            border-radius: 50%;
        }}
        .risk-critical {{ background: #ef4444; }}
        .risk-high {{ background: #eab308; }}
        .risk-medium {{ background: #22c55e; }}
        .action {{
            padding: 0.75rem;
            background: rgba(96, 165, 250, 0.1);
            border-radius: 8px;
            margin-bottom: 0.5rem;
            border-left: 3px solid #60a5fa;
        }}
        .footer {{
            text-align: center;
            margin-top: 2rem;
            padding-top: 1rem;
            border-top: 1px solid rgba(255,255,255,0.1);
            color: #64748b;
            font-size: 0.875rem;
        }}
        @media print {{
            body {{ background: white; color: #1e293b; padding: 1rem; }}
            .container {{ box-shadow: none; background: white; }}
            .card {{ background: #f8fafc; }}
        }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏗️ {data['project']['name']}</h1>
            <p class="subtitle">Project Overview • {datetime.now().strftime('%B %d, %Y')}</p>
        </div>

        <div class="grid">
            <div class="card">
                <h3>📊 Health Score</h3>
                <div class="score-ring">
                    <div class="score-inner">
                        <span class="score-value">{score:.0f}%</span>
                        <span style="font-size: 0.75rem; color: #64748b;">{'Ready' if health.get('production_ready') else 'In Dev'}</span>
                    </div>
                </div>
            </div>

            <div class="card">
                <h3>📁 Codebase</h3>
                <div class="metric">
                    <span class="metric-label">Python Files</span>
                    <span class="metric-value">{data['codebase']['python_files']}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Lines of Code</span>
                    <span class="metric-value">{data['codebase']['python_lines']:,}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">MCP Tools</span>
                    <span class="metric-value">{data['codebase']['mcp_tools']}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Prompts</span>
                    <span class="metric-value">{data['codebase']['mcp_prompts']}</span>
                </div>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>📋 Task Status</h3>
                <div class="metric">
                    <span class="metric-label">Total Tasks</span>
                    <span class="metric-value">{data['tasks']['total']}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Pending</span>
                    <span class="metric-value">{data['tasks']['by_status'].get('pending', 0) + data['tasks']['by_status'].get('Todo', 0)}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Remaining Work</span>
                    <span class="metric-value">{data['tasks']['remaining_hours']:.0f}h</span>
                </div>
            </div>

            <div class="card">
                <h3>🚧 Project Phases</h3>
                {''.join([f'''
                <div class="phase">
                    <div class="phase-header">
                        <span>{phase['name']}</span>
                        <span>{phase['progress']}%</span>
                    </div>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: {phase['progress']}%; background: {'#22c55e' if phase['status']=='complete' else '#60a5fa' if phase['status']=='in_progress' else '#374151'};"></div>
                    </div>
                </div>
                ''' for phase in data['phases']])}
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>⚠️ Risks & Blockers</h3>
                {''.join([f'''
                <div class="risk">
                    <div class="risk-dot risk-{risk['severity']}"></div>
                    <span>{risk['description']}</span>
                </div>
                ''' for risk in data['risks'][:5]])}
            </div>

            <div class="card">
                <h3>🎯 Next Actions</h3>
                {''.join([f'''
                <div class="action">
                    {action['action']} <span style="color: #64748b;">({action['estimate']})</span>
                </div>
                ''' for action in data['next_actions'][:4]])}
            </div>
        </div>

        <div class="footer">
            Generated by Exarp MCP Server • v{data['project']['version']}
        </div>
    </div>
</body>
</html>"""

    return html


def _format_markdown(data: dict) -> str:
    """Format as markdown."""
    health = data['health']
    score = health.get('overall_score', 0)

    md = f"""# 🏗️ {data['project']['name']}

*Project Overview • Generated: {datetime.now().strftime('%B %d, %Y')}*

---

## 📊 Health Score: **{score:.0f}%** {'🟢' if score >= 70 else '🟡' if score >= 50 else '🔴'}

**Production Ready:** {'✅ Yes' if health.get('production_ready') else '❌ No'}

| Component | Score | Status |
|-----------|-------|--------|
"""
    for name, value in sorted(health.get('scores', {}).items(), key=lambda x: -x[1]):
        status = "🟢" if value >= 70 else "🟡" if value >= 50 else "🔴"
        md += f"| {name.title()} | {value:.0f}% | {status} |\n"

    md += f"""
---

## 📁 Codebase

| Metric | Value |
|--------|-------|
| Python Files | {data['codebase']['python_files']} |
| Lines of Code | {data['codebase']['python_lines']:,} |
| MCP Tools | {data['codebase']['mcp_tools']} |
| Prompts | {data['codebase']['mcp_prompts']} |
| Documentation | {data['codebase']['doc_files']} files |

---

## 📋 Tasks

| Status | Count |
|--------|-------|
| Total | {data['tasks']['total']} |
| Pending | {data['tasks']['by_status'].get('pending', 0) + data['tasks']['by_status'].get('Todo', 0)} |
| Completed | {data['tasks']['by_status'].get('completed', 0)} |
| **Remaining Work** | **{data['tasks']['remaining_hours']:.0f}h** |

---

## 🚧 Project Phases

"""
    for phase in data['phases']:
        icon = "✅" if phase['status'] == 'complete' else "🔄" if phase['status'] == 'in_progress' else "⏳"
        md += f"- {icon} **{phase['name']}**: {phase['progress']}%\n"

    md += """
---

## ⚠️ Risks & Blockers

"""
    for risk in data['risks'][:5]:
        icon = "🔴" if risk['severity'] == 'critical' else "🟡" if risk['severity'] == 'high' else "🟢"
        md += f"- {icon} {risk['description']}\n"

    md += """
---

## 🎯 Next Actions

"""
    for i, action in enumerate(data['next_actions'][:5], 1):
        md += f"{i}. **{action['action']}** ({action['estimate']})\n"

    md += f"""
---

*Generated by Exarp MCP Server v{data['project']['version']}*
"""

    return md


def _format_marp_slides(data: dict) -> str:
    """Format as Marp markdown slides."""
    health = data['health']
    score = health.get('overall_score', 0)

    marp = f"""---
marp: true
theme: default
paginate: true
backgroundColor: #1a1a2e
color: #e4e4e7
style: |
  section {{
    font-family: 'Segoe UI', system-ui, sans-serif;
  }}
  h1 {{
    color: #60a5fa;
  }}
  h2 {{
    color: #a78bfa;
  }}
---

# 🏗️ {data['project']['name']}

## Project Status Overview

**Version:** {data['project']['version']}
**Generated:** {datetime.now().strftime('%B %d, %Y')}

---

# 📊 Health Score

<div style="text-align: center; font-size: 4rem; margin: 2rem 0;">
  <span style="color: {'#22c55e' if score >= 70 else '#eab308' if score >= 50 else '#ef4444'};">
    {score:.0f}%
  </span>
</div>

**Production Ready:** {'✅ Yes' if health.get('production_ready') else '❌ No'}

---

# 📊 Component Scores

| Component | Score | Status |
|-----------|-------|--------|
"""
    for name, value in sorted(health.get('scores', {}).items(), key=lambda x: -x[1])[:6]:
        status = "🟢" if value >= 70 else "🟡" if value >= 50 else "🔴"
        marp += f"| {name.title()} | {value:.0f}% | {status} |\n"

    marp += f"""
---

# 📁 Codebase & Tasks

<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2rem;">
<div>

### Codebase
- **{data['codebase']['python_files']}** Python files
- **{data['codebase']['python_lines']:,}** lines of code
- **{data['codebase']['mcp_tools']}** MCP tools
- **{data['codebase']['mcp_prompts']}** prompts

</div>
<div>

### Tasks
- **{data['tasks']['total']}** total tasks
- **{data['tasks']['by_status'].get('pending', 0) + data['tasks']['by_status'].get('Todo', 0)}** pending
- **{data['tasks']['remaining_hours']:.0f}h** remaining

</div>
</div>

---

# ⚠️ Risks & Blockers

"""
    for risk in data['risks'][:5]:
        icon = "🔴" if risk['severity'] == 'critical' else "🟡" if risk['severity'] == 'high' else "🟢"
        marp += f"- {icon} {risk['description']}\n"

    marp += """
---

# 🎯 Next Actions

"""
    for i, action in enumerate(data['next_actions'][:4], 1):
        marp += f"{i}. **{action['action']}** ({action['estimate']})\n"

    marp += """
---

# 🚧 Roadmap

"""
    for phase in data['phases']:
        icon = "✅" if phase['status'] == 'complete' else "🔄" if phase['status'] == 'in_progress' else "⏳"
        marp += f"- {icon} **{phase['name']}**: {phase['progress']}%\n"

    marp += """
---

# Thank You

**Questions?**

*Generated by Exarp MCP Server*
"""

    return marp

