"""
Rule Simplification Tool

Automatically simplifies rules based on exarp automation capabilities.
Replaces manual processes with exarp tool references.
"""

from typing import Optional, List
import json
import re
import sys
from pathlib import Path

from ..utils import find_project_root

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))


def simplify_rules(
    rule_files: Optional[List[str]] = None,
    dry_run: bool = True,
    output_dir: Optional[str] = None
) -> str:
    """
    Automatically simplify rules based on exarp automation capabilities.

    Args:
        rule_files: List of rule files to simplify (default: all .cursorrules and .cursor/rules/*.mdc)
        dry_run: Preview mode without making changes (default: True)
        output_dir: Directory to write simplified rules (default: same as source)

    Returns:
        JSON string with simplification results
    """
    project_root = find_project_root()

    # Default rule files if not specified
    if rule_files is None:
        rule_files = []
        # Add .cursorrules
        cursorrules_file = project_root / ".cursorrules"
        if cursorrules_file.exists():
            rule_files.append(str(cursorrules_file))

        # Add .cursor/rules/*.mdc files
        rules_dir = project_root / ".cursor" / "rules"
        if rules_dir.exists():
            for rule_file in rules_dir.glob("*.mdc"):
                rule_files.append(str(rule_file))

    results = {
        "status": "success",
        "files_processed": [],
        "files_skipped": [],
        "simplifications": [],
        "dry_run": dry_run
    }

    # Patterns to simplify
    simplification_patterns = _get_simplification_patterns()

    for rule_file_path in rule_files:
        rule_file = Path(rule_file_path)

        if not rule_file.exists():
            results["files_skipped"].append({
                "file": rule_file_path,
                "reason": "File not found"
            })
            continue

        try:
            with open(rule_file, encoding='utf-8') as f:
                content = f.read()

            simplifications = []

            # Apply simplification patterns
            for _pattern_name, pattern_config in simplification_patterns.items():
                matches = _apply_pattern(content, pattern_config)
                if matches:
                    content = _replace_pattern(content, pattern_config, matches)
                    simplifications.extend(matches)

            if simplifications:
                if not dry_run:
                    # Write simplified content
                    output_path = rule_file
                    if output_dir:
                        output_path = Path(output_dir) / rule_file.name
                        output_path.parent.mkdir(parents=True, exist_ok=True)

                    with open(output_path, 'w', encoding='utf-8') as f:
                        f.write(content)

                results["files_processed"].append({
                    "file": rule_file_path,
                    "simplifications_count": len(simplifications),
                    "simplifications": simplifications
                })
            else:
                results["files_skipped"].append({
                    "file": rule_file_path,
                    "reason": "No simplifications found"
                })

        except Exception as e:
            results["files_skipped"].append({
                "file": rule_file_path,
                "reason": f"Error processing file: {str(e)}"
            })

    return json.dumps(results, indent=2)


def _get_simplification_patterns() -> dict[str, dict]:
    """Get patterns for rule simplification."""
    return {
        "manual_linting": {
            "pattern": r"Run linters:\s*`\./scripts/run_linters\.sh`",
            "replacement": "Run linters: `lint:run` (or `./scripts/run_linters.sh`)",
            "description": "Replace manual linting command with command reference"
        },
        "manual_testing": {
            "pattern": r"Run:\s*`ctest --output-on-failure`",
            "replacement": "Run: `test:run` (or `ctest --output-on-failure`)",
            "description": "Replace manual testing command with command reference"
        },
        "manual_build": {
            "pattern": r"Verify build:\s*`cmake --build build`",
            "replacement": "Verify build: `build:debug` (or `cmake --build build`)",
            "description": "Replace manual build command with command reference"
        },
        "before_committing_manual": {
            "pattern": r"## Before Committing\n\n1\. Run linters:.*?\n6\. Add static analysis annotations where appropriate",
            "replacement": """## Before Committing

1. Run linters: `lint:run` (or `./scripts/run_linters.sh`)
2. Run tests: `test:run` (or `ctest --output-on-failure`)
3. Verify build: `build:debug` (or `cmake --build build`)
4. Check for credentials/secrets (manual - no automation)
5. Update documentation if needed
6. Add static analysis annotations where appropriate

**Automated checks (via exarp - see `.cursor/rules/project-automation.mdc`):**
- Documentation health (runs automatically on git hooks)
- Security scanning (runs automatically on dependency changes)
- Task alignment (runs automatically on task status changes)
- Duplicate detection (runs automatically on task creation)

**To configure automated checks:** Use exarp tools (git hooks, file watchers, cron jobs)""",
            "description": "Add automated checks section to Before Committing",
            "flags": re.DOTALL
        },
        "manual_docs_check": {
            "pattern": r"Check documentation health after recent changes",
            "replacement": "Check documentation health after recent changes (also runs automatically on git hooks)",
            "description": "Add note about automatic execution"
        },
        "manual_security_scan": {
            "pattern": r"Scan dependencies for security vulnerabilities",
            "replacement": "Scan dependencies for security vulnerabilities (also runs automatically on dependency file changes)",
            "description": "Add note about automatic execution"
        },
        "manual_task_alignment": {
            "pattern": r"Analyze task alignment before starting work",
            "replacement": "Analyze task alignment before starting work (also runs automatically on task status changes)",
            "description": "Add note about automatic execution"
        }
    }


def _apply_pattern(content: str, pattern_config: dict) -> list[dict]:
    """Apply a simplification pattern and return matches."""
    pattern = pattern_config["pattern"]
    flags = pattern_config.get("flags", 0)

    matches = []
    for match in re.finditer(pattern, content, flags):
        matches.append({
            "start": match.start(),
            "end": match.end(),
            "match": match.group(0),
            "description": pattern_config["description"]
        })

    return matches


def _replace_pattern(content: str, pattern_config: dict, matches: list[dict]) -> str:
    """Replace pattern matches with simplified versions."""
    pattern = pattern_config["pattern"]
    replacement = pattern_config["replacement"]
    flags = pattern_config.get("flags", 0)

    # Replace all matches
    content = re.sub(pattern, replacement, content, flags=flags)

    return content


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Simplify rules based on exarp automation")
    parser.add_argument("--files", nargs="+", help="Rule files to simplify")
    parser.add_argument("--no-dry-run", action="store_true", help="Actually modify files")
    parser.add_argument("--output-dir", help="Directory to write simplified rules")

    args = parser.parse_args()

    result = simplify_rules(
        rule_files=args.files,
        dry_run=not args.no_dry_run,
        output_dir=args.output_dir
    )

    print(result)
