"""
Base Classes for Automation Scripts

Provides IntelligentAutomationBase and MCP client utilities.
"""

from .intelligent_automation_base import IntelligentAutomationBase
from .mcp_client import MCPClient, get_mcp_client

__all__ = ['IntelligentAutomationBase', 'MCPClient', 'get_mcp_client']
