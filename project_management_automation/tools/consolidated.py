"""
Consolidated MCP Tools - Backward Compatibility Re-export

This module re-exports all consolidated tools from their new split modules
for backward compatibility. The tools have been split into logical modules:

- consolidated_analysis: analyze_alignment, task_analysis, task_discovery
- consolidated_automation: automation, estimation, task_workflow
- consolidated_quality: testing, lint, health
- consolidated_memory: memory, memory_maint
- consolidated_ai: ollama, mlx, coreml
- consolidated_config: generate_config, setup_hooks, prompt_tracking
- consolidated_reporting: report, security
- consolidated_workflow: workflow_mode, recommend, tool_catalog
- consolidated_git: git_tools, session

All imports from consolidated.py continue to work.
"""

# Analysis tools
from .consolidated_analysis import analyze_alignment, task_analysis, task_discovery

# Automation tools
from .consolidated_automation import automation, estimation, task_workflow

# Quality tools
from .consolidated_quality import testing, testing_async, lint, health

# Memory tools
from .consolidated_memory import memory, memory_maint

# AI tools
from .consolidated_ai import ollama, mlx, coreml

# Config tools
from .consolidated_config import generate_config, setup_hooks, prompt_tracking

# Reporting tools
from .consolidated_reporting import report, security, security_async

# Workflow tools
from .consolidated_workflow import workflow_mode, recommend, tool_catalog

# Git tools
from .consolidated_git import git_tools, session

__all__ = [
    # Analysis
    "analyze_alignment",
    "task_analysis",
    "task_discovery",
    # Automation
    "automation",
    "estimation",
    "task_workflow",
    # Quality
    "testing",
    "testing_async",
    "lint",
    "health",
    # Memory
    "memory",
    "memory_maint",
    # AI
    "ollama",
    "mlx",
    "coreml",
    # Config
    "generate_config",
    "setup_hooks",
    "prompt_tracking",
    # Reporting
    "report",
    "security",
    "security_async",
    # Workflow
    "workflow_mode",
    "recommend",
    "tool_catalog",
    # Git
    "git_tools",
    "session",
]
