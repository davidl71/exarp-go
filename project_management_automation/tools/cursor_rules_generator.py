"""
Cursor Rules Generator Tool

Generates .cursor/rules/*.mdc files based on project structure
and coding patterns. Based on Cursor IDE Best Practice #2.
"""

import json
import logging
import time
from pathlib import Path
from typing import Any, List, Optional

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


# Rule templates
RULE_TEMPLATES = {
    "python": {
        "filename": "python.mdc",
        "content": """---
description: Python development rules for AI assistance
globs: ["**/*.py"]
alwaysApply: true
---

# Python Development Rules

## Code Style
- Follow PEP 8 conventions
- Use type hints for function arguments and return values
- Maximum line length: 100 characters
- Use docstrings for all public functions and classes

## Imports
- Group imports: stdlib, third-party, local
- Use absolute imports over relative where possible
- Sort imports alphabetically within groups

## Error Handling
- Use specific exception types
- Always include informative error messages
- Log exceptions with context

## Testing
- Write tests for all new functions
- Use pytest as the testing framework
- Aim for >80% test coverage

## Security
- Never hardcode secrets or credentials
- Use environment variables for configuration
- Validate all external input
""",
    },
    "typescript": {
        "filename": "typescript.mdc",
        "content": """---
description: TypeScript development rules for AI assistance
globs: ["**/*.ts", "**/*.tsx"]
alwaysApply: true
---

# TypeScript Development Rules

## Type Safety
- Always use explicit types, avoid `any`
- Use interfaces for object shapes
- Prefer `const` assertions where applicable

## Code Style
- Use arrow functions for callbacks
- Prefer async/await over raw Promises
- Use destructuring for cleaner code

## React (if applicable)
- Use functional components with hooks
- Prefer named exports over default exports
- Keep components small and focused

## Testing
- Use Jest for unit testing
- Write tests for edge cases
- Mock external dependencies
""",
    },
    "api": {
        "filename": "api.mdc",
        "content": """---
description: API development rules for AI assistance
globs: ["**/api/**", "**/routes/**", "**/endpoints/**"]
alwaysApply: false
---

# API Development Rules

## REST Conventions
- Use proper HTTP methods (GET, POST, PUT, DELETE)
- Return appropriate status codes
- Use consistent response formats

## Error Handling
- Return structured error responses
- Include error codes and messages
- Log errors with request context

## Security
- Validate all input parameters
- Use authentication middleware
- Rate limit sensitive endpoints

## Documentation
- Document all endpoints
- Include request/response examples
- Keep OpenAPI spec updated
""",
    },
    "testing": {
        "filename": "testing.mdc",
        "content": """---
description: Testing rules for AI assistance
globs: ["**/test_*.py", "**/*_test.py", "**/*.test.ts", "**/*.spec.ts"]
alwaysApply: false
---

# Testing Rules

## Test Structure
- Use descriptive test names
- Follow Arrange-Act-Assert pattern
- One assertion per test when possible

## Mocking
- Mock external dependencies
- Use dependency injection for testability
- Reset mocks between tests

## Coverage
- Aim for >80% code coverage
- Cover edge cases and error paths
- Test both happy and unhappy paths

## Performance
- Keep tests fast (<1s each)
- Use fixtures for expensive setup
- Run tests in isolation
""",
    },
    "mcp": {
        "filename": "mcp.mdc",
        "content": """---
description: MCP tool development rules for AI assistance
globs: ["**/tools/**/*.py", "**/mcp/**/*.py"]
alwaysApply: false
---

# MCP Tool Development Rules

## Tool Design
- Each tool should do one thing well
- Use clear, descriptive function names
- Include comprehensive docstrings with HINT

## HINT Format
- Start with: [HINT: Brief description. Key outputs.]
- Include: 📊 Output, 🔧 Side Effects, 📁 Files, ⏱️ Runtime
- Add example prompts

## Error Handling
- Return JSON responses always
- Use format_success_response/format_error_response
- Log execution with log_automation_execution

## Registration
- Register tools in server.py
- Follow existing import patterns
- Include both relative and absolute import fallbacks
""",
    },
}


class CursorRulesGenerator:
    """Generates Cursor rules files based on project analysis."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.rules_dir = project_root / ".cursor" / "rules"

    def analyze_project(self) -> dict[str, Any]:
        """Analyze project to determine which rules to generate."""
        analysis = {
            "languages": [],
            "frameworks": [],
            "patterns": [],
            "recommended_rules": [],
        }

        # Detect languages
        if list(self.project_root.rglob("*.py")):
            analysis["languages"].append("python")
            analysis["recommended_rules"].append("python")

        if list(self.project_root.rglob("*.ts")) or list(self.project_root.rglob("*.tsx")):
            analysis["languages"].append("typescript")
            analysis["recommended_rules"].append("typescript")

        if list(self.project_root.rglob("*.js")) or list(self.project_root.rglob("*.jsx")):
            analysis["languages"].append("javascript")

        if list(self.project_root.rglob("*.go")):
            analysis["languages"].append("go")

        if list(self.project_root.rglob("*.rs")):
            analysis["languages"].append("rust")

        # Detect frameworks/patterns
        if (self.project_root / "package.json").exists():
            try:
                pkg = json.loads((self.project_root / "package.json").read_text())
                deps = {**pkg.get("dependencies", {}), **pkg.get("devDependencies", {})}
                if "react" in deps:
                    analysis["frameworks"].append("react")
                if "next" in deps:
                    analysis["frameworks"].append("nextjs")
                if "express" in deps:
                    analysis["frameworks"].append("express")
                    analysis["recommended_rules"].append("api")
            except Exception:
                pass

        if (self.project_root / "pyproject.toml").exists():
            try:
                content = (self.project_root / "pyproject.toml").read_text()
                if "fastapi" in content.lower():
                    analysis["frameworks"].append("fastapi")
                    analysis["recommended_rules"].append("api")
                if "django" in content.lower():
                    analysis["frameworks"].append("django")
                if "mcp" in content.lower():
                    analysis["frameworks"].append("mcp")
                    analysis["recommended_rules"].append("mcp")
            except Exception:
                pass

        # Detect test files
        test_files = list(self.project_root.rglob("test_*.py")) + list(
            self.project_root.rglob("*.test.ts")
        )
        if test_files:
            analysis["patterns"].append("testing")
            analysis["recommended_rules"].append("testing")

        # Detect API patterns
        api_dirs = ["api", "routes", "endpoints"]
        if any((self.project_root / d).exists() for d in api_dirs):
            analysis["patterns"].append("api")
            if "api" not in analysis["recommended_rules"]:
                analysis["recommended_rules"].append("api")

        return analysis

    def generate_rules(
        self,
        rules: Optional[List[str]] = None,
        overwrite: bool = False,
    ) -> dict[str, Any]:
        """
        Generate Cursor rules files.

        Args:
            rules: Specific rules to generate (None = all recommended)
            overwrite: Overwrite existing rules

        Returns:
            Dict with generation results
        """
        self.rules_dir.mkdir(parents=True, exist_ok=True)

        analysis = self.analyze_project()
        rules_to_generate = rules or analysis["recommended_rules"]

        results = {
            "analysis": analysis,
            "generated": [],
            "skipped": [],
            "errors": [],
        }

        for rule_name in rules_to_generate:
            if rule_name not in RULE_TEMPLATES:
                results["errors"].append(f"Unknown rule: {rule_name}")
                continue

            template = RULE_TEMPLATES[rule_name]
            rule_path = self.rules_dir / template["filename"]

            if rule_path.exists() and not overwrite:
                results["skipped"].append(
                    {"rule": rule_name, "path": str(rule_path), "reason": "exists"}
                )
                continue

            try:
                rule_path.write_text(template["content"])
                results["generated"].append(
                    {"rule": rule_name, "path": str(rule_path)}
                )
            except Exception as e:
                results["errors"].append(f"Error writing {rule_name}: {e}")

        return results


def generate_cursor_rules(
    rules: Optional[str] = None,
    overwrite: bool = False,
    analyze_only: bool = False,
) -> str:
    """
    [HINT: Cursor rules. Generates .mdc rules files from project analysis.]

    📊 Output: Generated rule files, project analysis
    🔧 Side Effects: Creates .cursor/rules/*.mdc files
    📁 Analyzes: Project structure, languages, frameworks
    ⏱️ Typical Runtime: 1-3 seconds

    Example Prompt:
    "Generate Cursor rules for my Python MCP project"

    Available rules:
    - python: PEP 8, type hints, docstrings
    - typescript: Type safety, React patterns
    - api: REST conventions, security
    - testing: Test patterns, coverage
    - mcp: MCP tool development

    Args:
        rules: Comma-separated rule names (None = auto-detect)
        overwrite: Overwrite existing rules
        analyze_only: Just analyze project, don't generate

    Returns:
        JSON with generation results
    """
    start_time = time.time()

    try:
        from project_management_automation.utils import find_project_root

        project_root = find_project_root()
        generator = CursorRulesGenerator(project_root)

        if analyze_only:
            analysis = generator.analyze_project()
            results = {
                "analysis": analysis,
                "available_rules": list(RULE_TEMPLATES.keys()),
                "tip": "Run without analyze_only to generate recommended rules",
            }
        else:
            rules_list = None
            if rules:
                rules_list = [r.strip() for r in rules.split(",")]

            results = generator.generate_rules(
                rules=rules_list,
                overwrite=overwrite,
            )

            results["summary"] = {
                "generated": len(results["generated"]),
                "skipped": len(results["skipped"]),
                "errors": len(results["errors"]),
            }

        duration = time.time() - start_time
        log_automation_execution("generate_cursor_rules", duration, True)

        return json.dumps(format_success_response(results), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("generate_cursor_rules", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

