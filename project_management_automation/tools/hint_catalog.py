"""
Enhanced HINT System Catalog

Provides rich tool descriptions, examples, and usage guidance.
Based on Cursor IDE Best Practice #5 (Detailed Prompts).
"""

from typing import Optional
import json
import logging
import time

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


# Enhanced tool catalog with rich descriptions
TOOL_CATALOG = {
    # Project Health
    "project_scorecard": {
        "hint": "Scorecard. Overall score, component scores, production readiness.",
        "category": "Project Health",
        "description": "Comprehensive project health assessment with scores across multiple dimensions",
        "outputs": ["Overall score 0-100", "Component breakdowns", "Production readiness"],
        "inputs": {"include_recommendations": "Include improvement suggestions"},
        "side_effects": "None (read-only)",
        "runtime": "5-15 seconds",
        "examples": [
            "Check my project's health score",
            "Is this project production ready?",
            "What areas need improvement?",
        ],
        "related_tools": ["project_overview", "check_documentation_health"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "project_overview": {
        "hint": "Overview. One-page summary with scores, metrics, tasks, risks.",
        "category": "Project Health",
        "description": "Executive summary of project status for stakeholders",
        "outputs": ["Project info", "Health scores", "Task summary", "Risks", "Roadmap"],
        "inputs": {"output_format": "text or json"},
        "side_effects": "None (read-only)",
        "runtime": "3-10 seconds",
        "examples": [
            "Give me a project overview",
            "Summarize project status",
            "What's the state of this project?",
        ],
        "related_tools": ["project_scorecard"],
        "recommended_model": "claude-haiku",
        "persona": "project_manager",
    },

    # Task Management
    "analyze_todo2_alignment": {
        "hint": "Task alignment. Misaligned count, avg score, follow-up tasks.",
        "category": "Task Management",
        "description": "Analyzes task alignment with project goals (PROJECT_GOALS.md)",
        "outputs": ["Misaligned tasks", "Alignment score", "Priority breakdown"],
        "inputs": {"create_followup_tasks": "Auto-create remediation tasks"},
        "side_effects": "May create tasks if create_followup_tasks=True",
        "runtime": "2-5 seconds",
        "examples": [
            "Check if tasks align with project goals",
            "Find misaligned tasks",
            "What tasks don't fit our strategy?",
        ],
        "related_tools": ["analyze_prd_alignment", "detect_duplicate_tasks"],
        "recommended_model": "claude-sonnet",
        "persona": "project_manager",
    },
    "analyze_prd_alignment": {
        "hint": "PRD alignment. Task-to-persona mapping, advisor assignments.",
        "category": "Task Management",
        "description": "Analyzes task alignment with PRD personas and user stories",
        "outputs": ["Persona coverage", "Unaligned tasks", "Recommendations"],
        "inputs": {"output_path": "Optional report path"},
        "side_effects": "None (read-only)",
        "runtime": "1-5 seconds",
        "examples": [
            "How well do tasks align with PRD?",
            "Which personas are underserved?",
            "Map tasks to user personas",
        ],
        "related_tools": ["generate_prd", "analyze_todo2_alignment"],
        "recommended_model": "claude-sonnet",
        "persona": "project_manager",
    },
    "detect_duplicate_tasks": {
        "hint": "Duplicate tasks. Count, groups, auto_fix available.",
        "category": "Task Management",
        "description": "Finds duplicate or similar tasks using fuzzy matching",
        "outputs": ["Duplicate groups", "Similarity scores", "Merge suggestions"],
        "inputs": {"similarity_threshold": "Match threshold 0-1", "auto_fix": "Auto-merge"},
        "side_effects": "May merge tasks if auto_fix=True",
        "runtime": "2-10 seconds",
        "examples": [
            "Find duplicate tasks",
            "Are there any redundant tasks?",
            "Clean up task duplicates",
        ],
        "related_tools": ["analyze_todo2_alignment"],
        "recommended_model": "claude-haiku",
        "persona": "project_manager",
    },

    # Code Quality
    "check_documentation_health": {
        "hint": "Docs health. Score 0-100, broken links, stale content.",
        "category": "Code Quality",
        "description": "Analyzes documentation quality and identifies issues",
        "outputs": ["Health score", "Broken links", "Stale docs", "Coverage gaps"],
        "inputs": {"create_tasks": "Create tasks for issues"},
        "side_effects": "May create tasks if create_tasks=True",
        "runtime": "5-20 seconds",
        "examples": [
            "Check documentation health",
            "Find broken links in docs",
            "Is my documentation up to date?",
        ],
        "related_tools": ["project_scorecard"],
        "recommended_model": "claude-haiku",
        "persona": "technical_writer",
    },
    "run_tests": {
        "hint": "Test runner. pytest/unittest, pass/fail counts, coverage.",
        "category": "Code Quality",
        "description": "Runs project tests and reports results",
        "outputs": ["Pass/fail counts", "Coverage %", "Failed test details"],
        "inputs": {"test_path": "Specific tests to run", "coverage": "Include coverage"},
        "side_effects": "Runs tests (may take time)",
        "runtime": "10-120 seconds",
        "examples": [
            "Run all tests",
            "Check test coverage",
            "What tests are failing?",
        ],
        "related_tools": ["analyze_test_coverage", "check_definition_of_done"],
        "recommended_model": "claude-haiku",
        "persona": "qa_engineer",
    },

    # Security
    "scan_dependency_security": {
        "hint": "Security scan. Vulnerabilities by severity, remediation.",
        "category": "Security",
        "description": "Scans dependencies for known vulnerabilities",
        "outputs": ["Vulnerability count", "Severity breakdown", "Remediation steps"],
        "inputs": {"languages": "Languages to scan"},
        "side_effects": "None (read-only)",
        "runtime": "5-30 seconds",
        "examples": [
            "Check dependencies for vulnerabilities",
            "Are my dependencies secure?",
            "Find security issues in packages",
        ],
        "related_tools": ["project_scorecard"],
        "recommended_model": "claude-sonnet",
        "persona": "security_engineer",
    },

    # PRD & Planning
    "generate_prd": {
        "hint": "PRD generation. Creates Product Requirements Document from codebase.",
        "category": "Planning",
        "description": "Generates PRD with personas, user stories, and requirements",
        "outputs": ["PRD markdown", "User stories", "Features", "Risks"],
        "inputs": {"output_path": "Where to save PRD"},
        "side_effects": "Creates PRD file",
        "runtime": "2-10 seconds",
        "examples": [
            "Generate a PRD for this project",
            "Create product requirements document",
            "What are the user stories?",
        ],
        "related_tools": ["analyze_prd_alignment", "analyze_todo2_alignment"],
        "recommended_model": "claude-sonnet",
        "persona": "project_manager",
    },

    # Workflow Optimization
    "recommend_workflow_mode": {
        "hint": "Workflow mode. AGENT vs ASK recommendation.",
        "category": "Workflow",
        "description": "Recommends AGENT or ASK mode based on task complexity",
        "outputs": ["Recommended mode", "Confidence %", "Rationale"],
        "inputs": {"task_description": "What you want to do"},
        "side_effects": "None (advisory)",
        "runtime": "<1 second",
        "examples": [
            "Should I use AGENT or ASK for this task?",
            "What mode for implementing authentication?",
            "AGENT or ASK for fixing this bug?",
        ],
        "related_tools": ["recommend_model"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "recommend_model": {
        "hint": "Model selection. Optimal AI model for task type.",
        "category": "Workflow",
        "description": "Recommends optimal AI model based on task",
        "outputs": ["Recommended model", "Confidence", "Alternatives"],
        "inputs": {"task_description": "What you want to do", "optimize_for": "quality/speed/cost"},
        "side_effects": "None (advisory)",
        "runtime": "<1 second",
        "examples": [
            "What model for complex refactoring?",
            "Best model for quick bug fix?",
            "Which AI for architecture review?",
        ],
        "related_tools": ["recommend_workflow_mode"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },

    # Definition of Done
    "check_definition_of_done": {
        "hint": "Definition of Done. Validates task completion criteria.",
        "category": "Code Quality",
        "description": "Checks if task meets completion criteria",
        "outputs": ["Checklist status", "Pass/fail/skip counts", "Ready for review?"],
        "inputs": {"task_id": "Task to check", "auto_check": "Run automated checks"},
        "side_effects": "May run tests/linter",
        "runtime": "5-30 seconds",
        "examples": [
            "Is my task ready for review?",
            "Check definition of done",
            "Am I missing anything before PR?",
        ],
        "related_tools": ["run_tests", "project_scorecard"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },

    # Configuration
    "generate_cursor_rules": {
        "hint": "Cursor rules. Generates .mdc rules from project analysis.",
        "category": "Configuration",
        "description": "Creates Cursor IDE rules based on project structure",
        "outputs": ["Generated rules", "Language detection", "Frameworks found"],
        "inputs": {"rules": "Specific rules to generate", "analyze_only": "Just analyze"},
        "side_effects": "Creates .cursor/rules/*.mdc files",
        "runtime": "1-3 seconds",
        "examples": [
            "Generate Cursor rules for my project",
            "What rules should I use?",
            "Create Python development rules",
        ],
        "related_tools": ["generate_cursorignore"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "generate_cursorignore": {
        "hint": "Cursorignore. Generates ignore files for AI context optimization.",
        "category": "Configuration",
        "description": "Creates .cursorignore for faster AI responses",
        "outputs": ["Detected patterns", "Files created"],
        "inputs": {"include_indexing": "Also create .cursorindexingignore"},
        "side_effects": "Creates ignore files",
        "runtime": "1-3 seconds",
        "examples": [
            "Generate cursorignore files",
            "Optimize AI context for my project",
            "What files should be ignored?",
        ],
        "related_tools": ["generate_cursor_rules"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },

    # Consolidated Tools (Action-Based)
    "add_external_tool_hints": {
        "hint": "External tool hints. Adds Context7 hints to documentation.",
        "category": "Code Quality",
        "description": "Adds Context7/external tool hints to documentation files",
        "outputs": ["Files updated", "Hints added", "Report"],
        "inputs": {"dry_run": "Preview changes without applying"},
        "side_effects": "Modifies documentation files",
        "runtime": "2-10 seconds",
        "examples": [
            "Add Context7 hints to my docs",
            "Update documentation with tool hints",
        ],
        "related_tools": ["check_documentation_health"],
        "recommended_model": "claude-haiku",
        "persona": "technical_writer",
    },
    "check_attribution": {
        "hint": "Attribution check. Validates AI-generated code attribution compliance.",
        "category": "Code Quality",
        "description": "Checks attribution compliance for AI-generated code",
        "outputs": ["Compliance status", "Issues found", "Recommendations"],
        "inputs": {"create_tasks": "Create tasks for issues"},
        "side_effects": "May create tasks",
        "runtime": "3-10 seconds",
        "examples": [
            "Check attribution compliance",
            "Are AI attributions correct?",
        ],
        "related_tools": ["project_scorecard"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "context": {
        "hint": "Context management. action=summarize|budget|batch. Unified context operations.",
        "category": "Workflow",
        "description": "Manages context: summarize, estimate budget, batch operations",
        "outputs": ["Summaries", "Token estimates", "Batch results"],
        "inputs": {"action": "summarize/budget/batch"},
        "side_effects": "None (read-only)",
        "runtime": "<1 second",
        "examples": [
            "Summarize this context",
            "Estimate context budget",
        ],
        "related_tools": ["tool_catalog"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "tool_catalog": {
        "hint": "Tool catalog. action=list|help. Unified tool catalog and help.",
        "category": "Workflow",
        "description": "Browses tool catalog and provides help for available tools",
        "outputs": ["Tool catalog", "Tool documentation", "Examples"],
        "inputs": {"action": "list/help", "category": "Filter by category", "tool_name": "Tool name for help"},
        "side_effects": "None (read-only)",
        "runtime": "<1 second",
        "examples": [
            "List all available tools",
            "Help with project_scorecard",
            "What tools are available?",
        ],
        "related_tools": [],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "generate_config": {
        "hint": "Config generation. action=rules|ignore|simplify. Creates IDE config files.",
        "category": "Configuration",
        "description": "Generates configuration files: Cursor rules, ignore files, simplified rules",
        "outputs": ["Generated configs", "Files created"],
        "inputs": {"action": "rules/ignore/simplify"},
        "side_effects": "Creates config files",
        "runtime": "1-5 seconds",
        "examples": [
            "Generate Cursor rules",
            "Create cursorignore files",
        ],
        "related_tools": ["generate_cursor_rules", "generate_cursorignore"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "health": {
        "hint": "Health check. action=server|git|docs|dod|cicd. Status and health metrics.",
        "category": "Project Health",
        "description": "Unified health checks: server, git, docs, definition of done, CI/CD",
        "outputs": ["Health status", "Metrics", "Issues"],
        "inputs": {"action": "server/git/docs/dod/cicd"},
        "side_effects": "None (read-only)",
        "runtime": "2-10 seconds",
        "examples": [
            "Check server health",
            "Validate CI/CD workflow",
        ],
        "related_tools": ["project_scorecard"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "improve_task_clarity": {
        "hint": "Task clarity improvement. Analyzes and improves task clarity metrics.",
        "category": "Task Management",
        "description": "Improves task clarity by adding estimates, renaming, removing dependencies",
        "outputs": ["Clarity score", "Improvements applied", "Recommendations"],
        "inputs": {"auto_apply": "Apply improvements automatically"},
        "side_effects": "Modifies tasks if auto_apply=True",
        "runtime": "2-5 seconds",
        "examples": [
            "Improve task clarity",
            "Add time estimates to tasks",
            "Fix task naming",
        ],
        "related_tools": ["task_analysis", "analyze_todo2_alignment"],
        "recommended_model": "claude-haiku",
        "persona": "project_manager",
    },
    "lint": {
        "hint": "Linting tool. action=run|analyze. Run linter or analyze problems.",
        "category": "Code Quality",
        "description": "Runs linter and analyzes code quality issues",
        "outputs": ["Lint results", "Issues found", "Suggestions"],
        "inputs": {"action": "run/analyze", "fix": "Auto-fix issues"},
        "side_effects": "May modify code if fix=True",
        "runtime": "5-30 seconds",
        "examples": [
            "Run linter",
            "Analyze lint problems",
        ],
        "related_tools": ["testing", "check_definition_of_done"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "memory": {
        "hint": "Memory tool. action=save|recall|search. Persist and retrieve AI discoveries.",
        "category": "Workflow",
        "description": "Manages AI session memories: save, recall, search",
        "outputs": ["Memory saved", "Search results", "Memory details"],
        "inputs": {"action": "save/recall/search"},
        "side_effects": "Creates/updates memory files",
        "runtime": "<1 second",
        "examples": [
            "Save this insight",
            "Search memories",
        ],
        "related_tools": ["memory_maint"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "memory_maint": {
        "hint": "Memory maintenance. action=health|gc|prune|consolidate|dream. Lifecycle management.",
        "category": "Workflow",
        "description": "Manages memory lifecycle: health, garbage collection, pruning, consolidation",
        "outputs": ["Health metrics", "Maintenance results"],
        "inputs": {"action": "health/gc/prune/consolidate/dream"},
        "side_effects": "Modifies memory files",
        "runtime": "2-10 seconds",
        "examples": [
            "Check memory health",
            "Clean up old memories",
        ],
        "related_tools": ["memory"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "prompt_tracking": {
        "hint": "Prompt tracking. action=log|analyze. Track and analyze prompts.",
        "category": "Workflow",
        "description": "Tracks and analyzes prompt iterations for optimization",
        "outputs": ["Tracking results", "Analysis", "Recommendations"],
        "inputs": {"action": "log/analyze"},
        "side_effects": "Creates tracking records",
        "runtime": "<1 second",
        "examples": [
            "Log this prompt",
            "Analyze prompt patterns",
        ],
        "related_tools": [],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "recommend": {
        "hint": "Recommendations. action=model|workflow|advisor. Unified recommendation system.",
        "category": "Workflow",
        "description": "Provides recommendations: model selection, workflow mode, advisor consultation",
        "outputs": ["Recommendations", "Rationale", "Alternatives"],
        "inputs": {"action": "model/workflow/advisor"},
        "side_effects": "None (advisory)",
        "runtime": "<1 second",
        "examples": [
            "Recommend a model for this task",
            "What workflow mode should I use?",
        ],
        "related_tools": ["recommend_model", "recommend_workflow_mode"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "run_daily_automation": {
        "hint": "Daily automation. Run daily checks (docs_health, alignment, duplicates, security).",
        "category": "Automation",
        "description": "Runs daily maintenance automation: docs health, alignment, duplicates, security",
        "outputs": ["Automation results", "Summary report"],
        "inputs": {"include_slow": "Include slow checks", "dry_run": "Preview only"},
        "side_effects": "May create tasks, modify files",
        "runtime": "30-120 seconds",
        "examples": [
            "Run daily automation",
            "Daily health checks",
        ],
        "related_tools": ["health", "project_scorecard"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "run_discover_automation": {
        "hint": "Discover automation. Find automation opportunities in codebase.",
        "category": "Automation",
        "description": "Discovers new automation opportunities in the codebase",
        "outputs": ["Opportunities found", "Value scores", "Recommendations"],
        "inputs": {"min_value_score": "Minimum value threshold"},
        "side_effects": "None (read-only)",
        "runtime": "10-30 seconds",
        "examples": [
            "Find automation opportunities",
            "What can be automated?",
        ],
        "related_tools": ["run_daily_automation"],
        "recommended_model": "claude-sonnet",
        "persona": "developer",
    },
    "run_nightly_automation": {
        "hint": "Nightly automation. Process tasks automatically with host limits.",
        "category": "Automation",
        "description": "Processes tasks automatically during nightly runs",
        "outputs": ["Tasks processed", "Results", "Summary"],
        "inputs": {"max_tasks_per_host": "Task limit per host"},
        "side_effects": "Modifies tasks, executes automation",
        "runtime": "60-300 seconds",
        "examples": [
            "Run nightly automation",
            "Process pending tasks",
        ],
        "related_tools": ["run_daily_automation"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "run_sprint_automation": {
        "hint": "Sprint automation. Full sprint automation with subtask extraction.",
        "category": "Automation",
        "description": "Full sprint automation workflow with task processing",
        "outputs": ["Sprint results", "Tasks processed", "Summary"],
        "inputs": {"max_iterations": "Max automation iterations"},
        "side_effects": "Modifies tasks, executes automation",
        "runtime": "120-600 seconds",
        "examples": [
            "Run sprint automation",
            "Process sprint tasks",
        ],
        "related_tools": ["run_daily_automation", "run_nightly_automation"],
        "recommended_model": "claude-sonnet",
        "persona": "project_manager",
    },
    "security": {
        "hint": "Security tool. action=scan|alerts|report. Unified security analysis.",
        "category": "Security",
        "description": "Unified security analysis: dependency scanning, alerts, reports",
        "outputs": ["Vulnerabilities", "Alerts", "Security report"],
        "inputs": {"action": "scan/alerts/report"},
        "side_effects": "None (read-only)",
        "runtime": "5-60 seconds",
        "examples": [
            "Scan dependencies for vulnerabilities",
            "Check security alerts",
        ],
        "related_tools": ["scan_dependency_security", "project_scorecard"],
        "recommended_model": "claude-sonnet",
        "persona": "security_engineer",
    },
    "setup_hooks": {
        "hint": "Hooks setup. action=git|patterns. Install automation hooks.",
        "category": "Configuration",
        "description": "Sets up git hooks and pattern triggers for automation",
        "outputs": ["Hooks installed", "Configuration"],
        "inputs": {"action": "git/patterns"},
        "side_effects": "Creates hook files",
        "runtime": "1-3 seconds",
        "examples": [
            "Setup git hooks",
            "Install automation hooks",
        ],
        "related_tools": ["generate_config"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "task_analysis": {
        "hint": "Task analysis. action=duplicates|tags|hierarchy|dependencies|parallelization. Task quality and structure.",
        "category": "Task Management",
        "description": "Analyzes tasks: duplicates, tags, hierarchy, dependencies, parallelization",
        "outputs": ["Duplicate groups", "Tag analysis", "Hierarchy structure", "Dependency chains", "Parallelization groups"],
        "inputs": {"action": "duplicates/tags/hierarchy/dependencies/parallelization"},
        "side_effects": "May modify tasks if auto_fix=True",
        "runtime": "2-10 seconds",
        "examples": [
            "Find duplicate tasks",
            "Analyze task hierarchy",
        ],
        "related_tools": ["detect_duplicate_tasks", "improve_task_clarity"],
        "recommended_model": "claude-haiku",
        "persona": "project_manager",
    },
    "task_discovery": {
        "hint": "Task discovery. action=comments|markdown|orphans|all. Find tasks from various sources.",
        "category": "Task Management",
        "description": "Discovers tasks from code comments, markdown, orphaned tasks",
        "outputs": ["Discovered tasks", "Locations", "Counts"],
        "inputs": {"action": "comments/markdown/orphans/all"},
        "side_effects": "May create tasks if create_tasks=True",
        "runtime": "5-20 seconds",
        "examples": [
            "Find TODO comments",
            "Discover tasks in markdown",
        ],
        "related_tools": ["task_analysis"],
        "recommended_model": "claude-haiku",
        "persona": "developer",
    },
    "task_workflow": {
        "hint": "Task workflow. action=sync|approve|clarify|clarity|cleanup. Manage task lifecycle. ⚠️ CRITICAL: ALWAYS use this tool for task updates - NEVER edit .todo2/state.todo2.json directly.",
        "category": "Task Management",
        "description": "Manages task workflow: sync, approve, clarify, and cleanup operations. Use action=approve with task_ids for batch status updates. Automatically calculates actualHours, normalizes status, and tracks history.",
        "outputs": ["Sync results", "Approvals", "Clarifications", "Cleanup results"],
        "inputs": {"action": "sync/approve/clarify/clarity/cleanup", "task_ids": "JSON array string for batch updates"},
        "side_effects": "Modifies task status (with automatic features: actualHours, status normalization, timestamps)",
        "runtime": "2-10 seconds",
        "examples": [
            "Batch update tasks to Done: task_workflow(action='approve', status='Todo', new_status='Done', task_ids='[\"T-0\", \"T-1\"]')",
            "Never edit .todo2/state.todo2.json directly - use this tool instead",
            "Sync tasks between systems",
            "Approve pending tasks",
        ],
        "related_tools": ["task_analysis"],
        "recommended_model": "claude-haiku",
        "persona": "project_manager",
    },
    "testing": {
        "hint": "Testing tool. action=run|coverage|suggest|validate. Execute tests, analyze coverage, suggest test cases, or validate test structure.",
        "category": "Code Quality",
        "description": "Unified testing: run tests, coverage analysis, test suggestions, validation",
        "outputs": ["Test results", "Coverage report", "Suggestions"],
        "inputs": {"action": "run/coverage/suggest/validate"},
        "side_effects": "Runs tests",
        "runtime": "10-300 seconds",
        "examples": [
            "Run tests",
            "Check test coverage",
        ],
        "related_tools": ["run_tests", "check_definition_of_done"],
        "recommended_model": "claude-haiku",
        "persona": "qa_engineer",
    },
}


def list_tools(
    category: Optional[str] = None,
    persona: Optional[str] = None,
    include_examples: bool = True,
) -> str:
    """
    [HINT: Tool catalog. Lists all tools with rich descriptions and examples.]

    📊 Output: Filtered tool catalog with usage guidance
    🔧 Side Effects: None
    📁 Static catalog data
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "What tools are available for task management?"
    "List tools for developers"

    Categories:
    - Project Health, Task Management, Code Quality
    - Security, Planning, Workflow, Configuration

    Args:
        category: Filter by category
        persona: Filter by target persona
        include_examples: Include example prompts

    Returns:
        JSON with tool catalog
    """
    start_time = time.time()

    try:
        tools = []
        for tool_id, tool_data in TOOL_CATALOG.items():
            # Apply filters
            if category and tool_data.get("category", "").lower() != category.lower():
                continue
            if persona and tool_data.get("persona", "").lower() != persona.lower():
                continue

            tool_entry = {
                "tool": tool_id,
                "hint": tool_data["hint"],
                "category": tool_data["category"],
                "description": tool_data["description"],
                "recommended_model": tool_data.get("recommended_model", "any"),
            }

            if include_examples:
                tool_entry["examples"] = tool_data.get("examples", [])[:2]

            tools.append(tool_entry)

        # Get unique categories
        categories = sorted({t.get("category", "Other") for t in TOOL_CATALOG.values()})
        personas = sorted({t.get("persona", "any") for t in TOOL_CATALOG.values()})

        result = {
            "tools": tools,
            "count": len(tools),
            "available_categories": categories,
            "available_personas": personas,
            "filters_applied": {
                "category": category,
                "persona": persona,
            },
        }

        duration = time.time() - start_time
        log_automation_execution("list_tools", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("list_tools", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def get_tool_help(tool_name: str) -> str:
    """
    [HINT: Tool help. Detailed help for a specific tool with examples.]

    📊 Output: Full tool documentation with examples and related tools
    🔧 Side Effects: None
    📁 Static catalog data
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "How do I use the project_scorecard tool?"
    "Help with analyze_prd_alignment"

    Args:
        tool_name: Name of the tool to get help for

    Returns:
        JSON with detailed tool documentation
    """
    start_time = time.time()

    try:
        # Find tool (case-insensitive)
        tool_data = None
        tool_id = None
        for tid, data in TOOL_CATALOG.items():
            if tid.lower() == tool_name.lower():
                tool_data = data
                tool_id = tid
                break

        if not tool_data:
            # Fuzzy match
            matches = [
                tid for tid in TOOL_CATALOG.keys()
                if tool_name.lower() in tid.lower()
            ]
            if matches:
                return json.dumps(format_success_response({
                    "error": f"Tool '{tool_name}' not found",
                    "did_you_mean": matches[:3],
                }), indent=2)
            else:
                return json.dumps(format_success_response({
                    "error": f"Tool '{tool_name}' not found",
                    "available_tools": list(TOOL_CATALOG.keys()),
                }), indent=2)

        result = {
            "tool": tool_id,
            "hint": tool_data["hint"],
            "category": tool_data["category"],
            "description": tool_data["description"],
            "outputs": tool_data.get("outputs", []),
            "inputs": tool_data.get("inputs", {}),
            "side_effects": tool_data.get("side_effects", "None"),
            "runtime": tool_data.get("runtime", "Unknown"),
            "examples": tool_data.get("examples", []),
            "related_tools": tool_data.get("related_tools", []),
            "recommended_model": tool_data.get("recommended_model", "any"),
            "best_for_persona": tool_data.get("persona", "any"),
        }

        duration = time.time() - start_time
        log_automation_execution("get_tool_help", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("get_tool_help", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

