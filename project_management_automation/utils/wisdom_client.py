"""
Wisdom MCP Client Helper

Provides Python interface to devwisdom-go MCP server.
Uses connection pooling to reuse sessions across multiple calls for better performance.
"""

import asyncio
import json
import logging
from pathlib import Path
from typing import Any, Dict, Optional

logger = logging.getLogger(__name__)

# Try to import MCP client library
try:
    from mcp import ClientSession, StdioServerParameters
    from mcp.client.stdio import stdio_client
    MCP_CLIENT_AVAILABLE = True
except ImportError:
    MCP_CLIENT_AVAILABLE = False
    logger.debug("MCP client library not available (optional - server uses FastMCP). Install with: uv sync (or uv pip install mcp>=1.0.0)")

# Import session pool from mcp_client
try:
    from project_management_automation.scripts.base.mcp_client import _session_pool
    POOL_AVAILABLE = True
except ImportError:
    POOL_AVAILABLE = False
    _session_pool = None


def _load_mcp_config(project_root: Path) -> dict:
    """Load MCP configuration from .cursor/mcp.json"""
    mcp_config_path = project_root / '.cursor' / 'mcp.json'
    if not mcp_config_path.exists():
        return {}
    
    try:
        with open(mcp_config_path) as f:
            config = json.load(f)
            return config.get('mcpServers', {})
    except Exception as e:
        logger.warning(f"Failed to load MCP config: {e}")
        return {}


async def _call_wisdom_tool(tool_name: str, arguments: dict, project_root: Path) -> Optional[dict]:
    """Call a wisdom tool via MCP with connection pooling."""
    if not MCP_CLIENT_AVAILABLE:
        logger.debug("MCP client library not available (optional - server uses FastMCP)")
        return None
    
    mcp_config = _load_mcp_config(project_root)
    if 'devwisdom' not in mcp_config:
        logger.warning("devwisdom MCP server not configured in .cursor/mcp.json")
        return None
    
    try:
        devwisdom_config = mcp_config.get('devwisdom', {})
        command = devwisdom_config.get('command', '/Users/davidl/Projects/devwisdom-go/devwisdom')
        args = devwisdom_config.get('args', [])
        
        server_params = StdioServerParameters(
            command=command,
            args=args
        )
        
        # Use connection pool if available, otherwise fall back to direct connection
        if POOL_AVAILABLE and _session_pool is not None:
            async with _session_pool.get_session('devwisdom', server_params) as session:
                result = await session.call_tool(tool_name, arguments)
                
                # Parse JSON response
                if result.content and len(result.content) > 0:
                    response_text = result.content[0].text
                    try:
                        return json.loads(response_text)
                    except json.JSONDecodeError:
                        logger.error(f"Failed to parse wisdom tool response: {response_text}")
                        return None
                return None
        else:
            # Fallback to direct connection
            async with stdio_client(server_params) as (read, write):
                async with ClientSession(read, write) as session:
                    await session.initialize()
                    
                    # Call the tool
                    result = await session.call_tool(tool_name, arguments)
                    
                    # Parse JSON response
                    if result.content and len(result.content) > 0:
                        response_text = result.content[0].text
                        try:
                            return json.loads(response_text)
                        except json.JSONDecodeError:
                            logger.error(f"Failed to parse wisdom tool response: {response_text}")
                            return None
                    return None
    except Exception as e:
        logger.error(f"Failed to call wisdom tool {tool_name}: {e}", exc_info=True)
        return None


async def _read_wisdom_resource(uri: str, project_root: Path) -> Optional[str]:
    """Read a wisdom resource via MCP with connection pooling."""
    if not MCP_CLIENT_AVAILABLE:
        logger.debug("MCP client library not available (optional - server uses FastMCP)")
        return None
    
    mcp_config = _load_mcp_config(project_root)
    if 'devwisdom' not in mcp_config:
        logger.warning("devwisdom MCP server not configured in .cursor/mcp.json")
        return None
    
    try:
        devwisdom_config = mcp_config.get('devwisdom', {})
        command = devwisdom_config.get('command', '/Users/davidl/Projects/devwisdom-go/devwisdom')
        args = devwisdom_config.get('args', [])
        
        server_params = StdioServerParameters(
            command=command,
            args=args
        )
        
        # Use connection pool if available, otherwise fall back to direct connection
        if POOL_AVAILABLE and _session_pool is not None:
            async with _session_pool.get_session('devwisdom', server_params) as session:
                # Read the resource
                result = await session.read_resource(uri)
                
                # Return resource content
                if result.contents and len(result.contents) > 0:
                    return result.contents[0].text
                return None
        else:
            # Fallback to direct connection
            async with stdio_client(server_params) as (read, write):
                async with ClientSession(read, write) as session:
                    await session.initialize()
                    
                    # Read the resource
                    result = await session.read_resource(uri)
                    
                    # Return resource content
                    if result.contents and len(result.contents) > 0:
                        return result.contents[0].text
                    return None
    except Exception as e:
        logger.error(f"Failed to read wisdom resource {uri}: {e}", exc_info=True)
        return None


# Synchronous wrappers for easier integration
def call_wisdom_tool_sync(tool_name: str, arguments: dict, project_root: Optional[Path] = None) -> Optional[dict]:
    """Synchronous wrapper for calling wisdom tools."""
    if project_root is None:
        from .project_root import find_project_root
        project_root = find_project_root()
    
    try:
        return asyncio.run(_call_wisdom_tool(tool_name, arguments, project_root))
    except Exception as e:
        logger.error(f"Failed to call wisdom tool synchronously: {e}")
        return None


def read_wisdom_resource_sync(uri: str, project_root: Optional[Path] = None) -> Optional[str]:
    """Synchronous wrapper for reading wisdom resources."""
    if project_root is None:
        from .project_root import find_project_root
        project_root = find_project_root()
    
    try:
        return asyncio.run(_read_wisdom_resource(uri, project_root))
    except Exception as e:
        logger.error(f"Failed to read wisdom resource synchronously: {e}")
        return None


# Convenience functions matching the old wisdom API
def get_wisdom(score: float, source: Optional[str] = None, **kwargs) -> Optional[dict]:
    """Get wisdom quote - calls devwisdom-go MCP server."""
    from .project_root import find_project_root
    project_root = find_project_root()
    
    arguments = {"score": score}
    if source:
        arguments["source"] = source
    
    result = call_wisdom_tool_sync("get_wisdom", arguments, project_root)
    return result


def consult_advisor(metric: Optional[str] = None, tool: Optional[str] = None, 
                    stage: Optional[str] = None, score: float = 50.0, 
                    context: Optional[str] = None, **kwargs) -> Optional[dict]:
    """Consult advisor - calls devwisdom-go MCP server."""
    from .project_root import find_project_root
    project_root = find_project_root()
    
    arguments = {"score": score}
    if metric:
        arguments["metric"] = metric
    if tool:
        arguments["tool"] = tool
    if stage:
        arguments["stage"] = stage
    if context:
        arguments["context"] = context
    
    result = call_wisdom_tool_sync("consult_advisor", arguments, project_root)
    return result


def get_daily_briefing(overall_score: float, metric_scores: Optional[dict] = None, **kwargs) -> Optional[dict]:
    """Get daily briefing - calls devwisdom-go MCP server."""
    from .project_root import find_project_root
    project_root = find_project_root()
    
    arguments = {"score": overall_score}
    result = call_wisdom_tool_sync("get_daily_briefing", arguments, project_root)
    return result


def format_text(wisdom: Optional[dict]) -> str:
    """Format wisdom quote as text - compatibility wrapper."""
    if not wisdom:
        return ""
    
    quote = wisdom.get('quote', '')
    source = wisdom.get('source', '')
    encouragement = wisdom.get('encouragement', '')
    
    if not quote:
        return ""
    
    lines = [f'"{quote}"']
    if encouragement:
        lines.append(f"— {encouragement}")
    if source:
        lines.append(f"Source: {source}")
    
    return "\n".join(lines)


def list_sources() -> list[str]:
    """List available wisdom sources - calls devwisdom-go MCP server."""
    from .project_root import find_project_root
    project_root = find_project_root()
    
    result = read_wisdom_resource_sync("wisdom://sources", project_root)
    if result:
        try:
            data = json.loads(result)
            if isinstance(data, list):
                return [item.get('id', '') for item in data if isinstance(item, dict)]
            elif isinstance(data, dict) and 'sources' in data:
                return data['sources']
        except json.JSONDecodeError:
            pass
    
    return []

