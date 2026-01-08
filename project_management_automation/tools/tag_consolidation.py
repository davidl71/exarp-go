"""
Tag Consolidation Tool for Todo2 Tasks.

Analyzes, validates, and consolidates tags in Todo2 tasks.
"""

import json
import logging
from collections import defaultdict
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

from ..utils.todo2_utils import (
    filter_tasks_by_project,
    get_repo_project_id,
    task_belongs_to_project,
)

# Standard consolidation rules
DEFAULT_CONSOLIDATION_RULES = {
    # Plural → Singular
    "tools": "tool",
    "tests": "testing",
    "test": "testing",
    "docs": "docs",
    "doc": "docs",
    "documentation": "docs",
    "features": "enhancement",
    "feature": "enhancement",
    "bugs": "bug",
    "refactoring": "refactor",

    # Case fixes
    "Apache-license": "apache-license",
    "MIT-license": "mit-license",

    # Long tag shortening
    "automation-opportunity-finder": "automation-finder",
    "documentation-health-analysis": "docs-health",
    "external-tool-hints-automation": "tool-hints",
    "shared-todo-table-synchronization": "todo-sync",
    "test-coverage-analyzer": "coverage-analyzer",
    "automation-opportunities": "automation-opps",
    "dependency-security-scan": "security-scan",
    "information-disclosure": "info-disclosure",
    "todo2-alignment-analysis": "task-alignment",
    "todo2-duplicate-detection": "duplicate-detect",
}

# Tags that should be removed (too generic or obsolete)
TAGS_TO_REMOVE = set()

# Maximum recommended tag length
MAX_TAG_LENGTH = 20


logger = logging.getLogger(__name__)


def find_project_root() -> Path:
    """Find project root by looking for .todo2 directory."""
    current = Path.cwd()
    while current != current.parent:
        if (current / '.git').exists() or (current / '.todo2').exists() or (current / 'CMakeLists.txt').exists() or (current / 'go.mod').exists():
            return current
        current = current.parent
    return Path.cwd()


def load_todo2_tasks(project_root: Path) -> tuple[dict[str, Any], Path]:
    """Load Todo2 tasks from state file."""
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    if not todo2_file.exists():
        raise FileNotFoundError(f"Todo2 state file not found: {todo2_file}")

    with open(todo2_file) as f:
        data = json.load(f)

    return data, todo2_file


def analyze_tags(todos: list[dict[str, Any]]) -> dict[str, Any]:
    """Analyze current tag usage."""
    all_tags = []
    tag_counts = defaultdict(int)
    tag_to_tasks = defaultdict(list)

    for task in todos:
        tags = task.get('tags', [])
        task_id = task.get('id', 'unknown')
        for tag in tags:
            all_tags.append(tag)
            tag_counts[tag] += 1
            tag_to_tasks[tag].append(task_id)

    unique_tags = sorted(set(all_tags))

    # Identify issues
    issues = {
        'not_lowercase': [],
        'too_long': [],
        'has_spaces': [],
        'rare_tags': [],  # Used 1-2 times
    }

    for tag in unique_tags:
        if tag != tag.lower():
            issues['not_lowercase'].append(tag)
        if len(tag) > MAX_TAG_LENGTH:
            issues['too_long'].append(tag)
        if ' ' in tag:
            issues['has_spaces'].append(tag)
        if tag_counts[tag] <= 2:
            issues['rare_tags'].append((tag, tag_counts[tag]))

    return {
        'total_tasks': len(todos),
        'total_tag_usages': len(all_tags),
        'unique_tags': unique_tags,
        'tag_counts': dict(tag_counts),
        'tag_to_tasks': dict(tag_to_tasks),
        'issues': issues,
    }


def plan_consolidations(
    analysis: dict[str, Any],
    rules: dict[str, str],
    remove_tags: set
) -> dict[str, Any]:
    """Plan tag consolidations without applying them."""
    tag_counts = analysis['tag_counts']
    tag_to_tasks = analysis['tag_to_tasks']

    renames = []  # (old_tag, new_tag, affected_task_ids)
    removals = []  # (tag, affected_task_ids)

    for old_tag, new_tag in rules.items():
        if old_tag in tag_counts:
            renames.append({
                'old': old_tag,
                'new': new_tag,
                'count': tag_counts[old_tag],
                'tasks': tag_to_tasks[old_tag],
            })

    for tag in remove_tags:
        if tag in tag_counts:
            removals.append({
                'tag': tag,
                'count': tag_counts[tag],
                'tasks': tag_to_tasks[tag],
            })

    # Calculate stats
    tags_before = len(analysis['unique_tags'])
    tags_consolidated = len(renames)
    tags_removed = len(removals)

    # Some renames might create new tags, some consolidate existing
    new_tags_from_renames = {r['new'] for r in renames}
    existing_tags = set(analysis['unique_tags'])
    truly_new = new_tags_from_renames - existing_tags

    tags_after = tags_before - tags_consolidated - tags_removed + len(truly_new)

    return {
        'renames': renames,
        'removals': removals,
        'stats': {
            'tags_before': tags_before,
            'tags_consolidated': tags_consolidated,
            'tags_removed': tags_removed,
            'tags_after': tags_after,
            'reduction': tags_before - tags_after,
        }
    }


def apply_consolidations(
    data: dict[str, Any],
    plan: dict[str, Any],
    todo2_file: Path,
    dry_run: bool = True,
    project_id: Optional[str] = None,
) -> dict[str, Any]:
    """Apply tag consolidations to Todo2 tasks."""
    todos = data.get('todos', [])

    # Build rename map
    rename_map = {r['old']: r['new'] for r in plan['renames']}
    remove_set = {r['tag'] for r in plan['removals']}

    changes = []

    for task in todos:
        if not task_belongs_to_project(task, project_id):
            continue
        old_tags = task.get('tags', [])
        if not old_tags:
            continue

        new_tags = []
        task_changes = []

        for tag in old_tags:
            if tag in remove_set:
                task_changes.append(f"removed '{tag}'")
            elif tag in rename_map:
                new_tag = rename_map[tag]
                if new_tag not in new_tags:  # Avoid duplicates
                    new_tags.append(new_tag)
                task_changes.append(f"'{tag}' → '{new_tag}'")
            else:
                if tag not in new_tags:  # Avoid duplicates
                    new_tags.append(tag)

        if task_changes:
            changes.append({
                'task_id': task.get('id'),
                'task_content': task.get('content', '')[:50],
                'changes': task_changes,
                'old_tags': old_tags,
                'new_tags': new_tags,
            })

            if not dry_run:
                task['tags'] = new_tags

    if not dry_run:
        # Save changes
        with open(todo2_file, 'w') as f:
            json.dump(data, f, indent=2)

    return {
        'dry_run': dry_run,
        'tasks_modified': len(changes),
        'changes': changes,
        'stats': plan['stats'],
    }


def format_report(
    analysis: dict[str, Any],
    plan: dict[str, Any],
    result: dict[str, Any]
) -> str:
    """Format a human-readable report."""
    lines = []

    # Header
    mode = "DRY RUN" if result['dry_run'] else "APPLIED"
    lines.append(f"{'=' * 70}")
    lines.append(f"📊 TAG CONSOLIDATION REPORT ({mode})")
    lines.append(f"{'=' * 70}")
    lines.append(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")

    # Stats
    stats = result['stats']
    lines.append("\n📈 STATISTICS")
    lines.append("─" * 40)
    lines.append(f"  Tags before:      {stats['tags_before']}")
    lines.append(f"  Tags consolidated:{stats['tags_consolidated']}")
    lines.append(f"  Tags removed:     {stats['tags_removed']}")
    lines.append(f"  Tags after:       {stats['tags_after']}")
    lines.append(f"  Net reduction:    {stats['reduction']} ({stats['reduction']/stats['tags_before']*100:.1f}%)")

    # Renames
    if plan['renames']:
        lines.append(f"\n🔄 TAG RENAMES ({len(plan['renames'])})")
        lines.append("─" * 40)
        for r in sorted(plan['renames'], key=lambda x: -x['count']):
            lines.append(f"  '{r['old']}' → '{r['new']}' ({r['count']} tasks)")

    # Removals
    if plan['removals']:
        lines.append(f"\n🗑️  TAG REMOVALS ({len(plan['removals'])})")
        lines.append("─" * 40)
        for r in plan['removals']:
            lines.append(f"  '{r['tag']}' ({r['count']} tasks)")

    # Task changes (first 20)
    if result['changes']:
        lines.append(f"\n📝 TASK CHANGES ({result['tasks_modified']} tasks)")
        lines.append("─" * 40)
        for change in result['changes'][:20]:
            lines.append(f"  {change['task_id']}: {change['task_content'][:40]}...")
            for c in change['changes']:
                lines.append(f"      • {c}")
        if len(result['changes']) > 20:
            lines.append(f"  ... and {len(result['changes']) - 20} more tasks")

    # Issues remaining
    issues = analysis['issues']
    remaining_issues = []
    if issues['not_lowercase']:
        fixed = [r['old'] for r in plan['renames'] if r['old'] in issues['not_lowercase']]
        remaining = [t for t in issues['not_lowercase'] if t not in fixed]
        if remaining:
            remaining_issues.append(f"Not lowercase: {remaining}")
    if issues['too_long']:
        fixed = [r['old'] for r in plan['renames'] if r['old'] in issues['too_long']]
        remaining = [t for t in issues['too_long'] if t not in fixed]
        if remaining:
            remaining_issues.append(f"Too long: {remaining}")

    if remaining_issues:
        lines.append("\n⚠️  REMAINING ISSUES")
        lines.append("─" * 40)
        for issue in remaining_issues:
            lines.append(f"  • {issue}")

    # Footer
    lines.append(f"\n{'=' * 70}")
    if result['dry_run']:
        lines.append("💡 This was a DRY RUN. No changes were made.")
        lines.append("   Run with dry_run=False to apply changes.")
    else:
        lines.append("✅ Changes have been applied to .todo2/state.todo2.json")
    lines.append(f"{'=' * 70}")

    return '\n'.join(lines)


def consolidate_tags(
    dry_run: bool = True,
    custom_rules: Optional[Dict[str, str]] = None,
    remove_tags: Optional[List[str]] = None,
    output_path: Optional[str] = None,
) -> str:
    """
    Main function to consolidate Todo2 tags.

    Args:
        dry_run: If True, only report what would change without applying
        custom_rules: Additional consolidation rules (old_tag → new_tag)
        remove_tags: Tags to remove entirely
        output_path: Optional path to save report

    Returns:
        Formatted report string
    """
    # Find project and load data
    project_root = find_project_root()
    data, todo2_file = load_todo2_tasks(project_root)
    project_id = get_repo_project_id(project_root)
    todos = filter_tasks_by_project(data.get('todos', []), project_id, logger=logger)

    # Merge rules
    rules = DEFAULT_CONSOLIDATION_RULES.copy()
    if custom_rules:
        rules.update(custom_rules)

    tags_to_remove = TAGS_TO_REMOVE.copy()
    if remove_tags:
        tags_to_remove.update(remove_tags)

    # Analyze
    analysis = analyze_tags(todos)

    # Plan
    plan = plan_consolidations(analysis, rules, tags_to_remove)

    # Apply (or simulate)
    result = apply_consolidations(data, plan, todo2_file, dry_run, project_id)

    # Format report
    report = format_report(analysis, plan, result)

    # Save report if requested
    if output_path:
        Path(output_path).write_text(report)

    return report


# MCP Tool wrapper
def tag_consolidation_tool(
    dry_run: bool = True,
    custom_rules: Optional[str] = None,
    remove_tags: Optional[str] = None,
    output_path: Optional[str] = None,
) -> str:
    """
    [HINT: Tag consolidation. Returns renames, removals, stats, task changes.]

    Analyze and consolidate Todo2 task tags.

    Applies standard consolidation rules:
    - Plural → singular (tools → tool, tests → testing)
    - Case normalization (Apache-license → apache-license)
    - Long tag shortening (documentation-health-analysis → docs-health)

    Args:
        dry_run: If True, only report what would change (default: True)
        custom_rules: JSON string of additional rules {"old": "new", ...}
        remove_tags: JSON array of tags to remove ["tag1", "tag2"]
        output_path: Optional path to save report

    Returns:
        Formatted consolidation report
    """
    parsed_rules = None
    parsed_remove = None

    if custom_rules:
        try:
            parsed_rules = json.loads(custom_rules)
        except json.JSONDecodeError:
            return f"Error: Invalid JSON in custom_rules: {custom_rules}"

    if remove_tags:
        try:
            parsed_remove = json.loads(remove_tags)
        except json.JSONDecodeError:
            return f"Error: Invalid JSON in remove_tags: {remove_tags}"

    return consolidate_tags(
        dry_run=dry_run,
        custom_rules=parsed_rules,
        remove_tags=parsed_remove,
        output_path=output_path,
    )


if __name__ == "__main__":
    import sys

    dry_run = "--apply" not in sys.argv
    report = consolidate_tags(dry_run=dry_run)
    print(report)

