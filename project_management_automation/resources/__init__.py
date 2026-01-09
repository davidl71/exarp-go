# project_management_automation/resources/__init__.py
"""
Resources package for project management automation.

Exports:
    - context_primer: Workflow mode context and session priming
"""

from .context_primer import WORKFLOW_MODE_CONTEXT, get_context_primer

__all__ = [
    'WORKFLOW_MODE_CONTEXT',
    'get_context_primer',
]

