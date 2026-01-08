"""
Agentic-Tools MCP Client Helper

Provides Python interface to agentic-tools MCP server.
Uses connection pooling to reuse sessions across multiple calls for better performance.
"""

import asyncio
import json
import logging
from pathlib import Path
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)

# Try to import MCP client library
try:
    from mcp import ClientSession, StdioServerParameters
    from mcp.client.stdio import stdio_client
    MCP_CLIENT_AVAILABLE = True
except ImportError:
    MCP_CLIENT_AVAILABLE = False
    logger.debug("MCP client library not available (optional - server uses FastMCP). Install with: uv sync (or uv pip install mcp>=1.0.0)")

# Import session pool from mcp_client if available
try:
    from project_management_automation.scripts.base.mcp_client import _session_pool
    POOL_AVAILABLE = True
except ImportError:
    POOL_AVAILABLE = False
    _session_pool = None


def _load_mcp_config(project_root: Path) -> dict:
    """Load MCP configuration from .cursor/mcp.json or ~/.cursor/mcp.json"""
    # Try project-level config first
    mcp_config_path = project_root / '.cursor' / 'mcp.json'
    if not mcp_config_path.exists():
        # Try user-level config
        from pathlib import Path as PathLib
        home = PathLib.home()
        mcp_config_path = home / '.cursor' / 'mcp.json'
        if not mcp_config_path.exists():
            return {}
    
    try:
        with open(mcp_config_path) as f:
            config = json.load(f)
            return config.get('mcpServers', {})
    except Exception as e:
        logger.warning(f"Failed to load MCP config: {e}")
        return {}


async def _call_agentic_tools_tool(tool_name: str, arguments: dict, project_root: Path) -> Optional[dict]:
    """Call an agentic-tools function via MCP with connection pooling."""
    if not MCP_CLIENT_AVAILABLE:
        logger.debug("MCP client library not available (optional - server uses FastMCP)")
        return None
    
    mcp_config = _load_mcp_config(project_root)
    if 'agentic-tools' not in mcp_config:
        logger.warning("agentic-tools MCP server not configured in MCP config")
        return None
    
    agentic_tools_config = mcp_config.get('agentic-tools', {})
    
    # Agentic-tools uses uvx wrapper, need to extract command and args
    command = agentic_tools_config.get('command', 'uvx')
    args = agentic_tools_config.get('args', [])
    env = agentic_tools_config.get('env', {})
    
    try:
        server_params = StdioServerParameters(
            command=command,
            args=args,
            env=env
        )
        
        # Use connection pool if available, otherwise fall back to direct connection
        if POOL_AVAILABLE and _session_pool is not None:
            async with _session_pool.get_session('agentic-tools', server_params) as session:
                result = await session.call_tool(tool_name, arguments)
                
                # Parse JSON response
                if result.content and len(result.content) > 0:
                    response_text = result.content[0].text
                    try:
                        return json.loads(response_text)
                    except json.JSONDecodeError:
                        # Handle non-JSON responses (e.g., plain text messages)
                        # Return as a dict with the message
                        logger.debug(f"Agentic-tools returned non-JSON response: {response_text}")
                        return {"message": response_text, "success": True}
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
                            logger.error(f"Failed to parse agentic-tools response: {response_text}")
                            return None
                    return None
    except Exception as e:
        logger.error(f"Failed to call agentic-tools tool {tool_name}: {e}", exc_info=True)
        return None


def _call_agentic_tools_tool_sync(tool_name: str, arguments: dict, project_root: Path) -> Optional[dict]:
    """Synchronous wrapper for async agentic-tools tool calls."""
    try:
        return asyncio.run(_call_agentic_tools_tool(tool_name, arguments, project_root))
    except Exception as e:
        logger.error(f"Failed to call agentic-tools tool {tool_name} (sync): {e}")
        return None


# Public API functions

def infer_task_progress_mcp(
    project_id: Optional[str] = None,
    scan_depth: int = 3,
    file_extensions: Optional[List[str]] = None,
    auto_update_tasks: bool = False,
    confidence_threshold: float = 0.7,
    project_root: Optional[Path] = None
) -> Optional[Dict[str, Any]]:
    """
    Infer task progress by analyzing codebase for implementation evidence.
    
    Uses agentic-tools `infer_task_progress` to analyze codebase and detect
    completed tasks based on code changes, file creation, and implementation evidence.
    
    Args:
        project_id: Filter to specific project (optional)
        scan_depth: Directory depth to scan (1-5, default: 3)
        file_extensions: File types to analyze (default: [.js, .ts, .jsx, .tsx, .py, .java, .cs, .go, .rs])
        auto_update_tasks: Whether to automatically update task status (default: False)
        confidence_threshold: Minimum confidence for auto-updating (0-1, default: 0.7)
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        Dictionary with analysis results including inferred completions and confidence scores
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Prepare arguments
    # Note: agentic-tools requires workingDirectory (absolute path) as a required parameter
    arguments = {
        'workingDirectory': str(project_root.absolute())  # Required parameter
    }
    if project_id:
        arguments['projectId'] = project_id
    if scan_depth:
        arguments['scanDepth'] = scan_depth
    if file_extensions:
        arguments['fileExtensions'] = file_extensions
    if auto_update_tasks is not None:
        arguments['autoUpdateTasks'] = auto_update_tasks
    if confidence_threshold is not None:
        arguments['confidenceThreshold'] = confidence_threshold
    
    # Call agentic-tools MCP (tool name is just the base name, not prefixed)
    result = _call_agentic_tools_tool_sync(
        'infer_task_progress',
        arguments,
        project_root
    )
    
    return result


def generate_research_queries_mcp(
    task_id: str,
    query_types: Optional[List[str]] = None,
    include_advanced: bool = False,
    target_year: int = 2025,
    project_root: Optional[Path] = None
) -> Optional[Dict[str, Any]]:
    """
    Generate intelligent, targeted web search queries for task research.
    
    Uses agentic-tools `generate_research_queries` to create optimized search
    queries for different research types.
    
    Args:
        task_id: ID of the task to generate research queries for
        query_types: Types of queries to generate (implementation, best_practices, 
                     troubleshooting, alternatives, performance, security, examples, tools)
        include_advanced: Include advanced search operators (default: False)
        target_year: Target year for recent information (default: 2025)
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        Dictionary with structured list of optimized search queries
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Prepare arguments
    arguments = {
        'taskId': task_id,
    }
    if query_types:
        arguments['queryTypes'] = query_types
    if include_advanced is not None:
        arguments['includeAdvanced'] = include_advanced
    if target_year:
        arguments['targetYear'] = target_year
    
    # Call agentic-tools MCP (tool name is just the base name, not prefixed)
    result = _call_agentic_tools_tool_sync(
        'generate_research_queries',
        arguments,
        project_root
    )
    
    return result


def get_next_task_recommendation_mcp(
    project_id: Optional[str] = None,
    max_recommendations: int = 3,
    consider_complexity: bool = True,
    preferred_tags: Optional[List[str]] = None,
    exclude_blocked: bool = True,
    project_root: Optional[Path] = None
) -> Optional[Dict[str, Any]]:
    """
    Get intelligent recommendations for the next task to work on.
    
    Uses agentic-tools `get_next_task_recommendation` to suggest tasks based on
    dependencies, priorities, complexity, and current project status.
    
    Args:
        project_id: Filter to specific project (optional)
        max_recommendations: Maximum number of recommendations (1-10, default: 3)
        consider_complexity: Whether to factor in task complexity (default: True)
        preferred_tags: Preferred task tags to prioritize (optional)
        exclude_blocked: Whether to exclude blocked tasks (default: True)
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        Dictionary with recommended tasks and rationale
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Prepare arguments
    arguments = {}
    if project_id:
        arguments['projectId'] = project_id
    if max_recommendations:
        arguments['maxRecommendations'] = max_recommendations
    if consider_complexity is not None:
        arguments['considerComplexity'] = consider_complexity
    if preferred_tags:
        arguments['preferredTags'] = preferred_tags
    if exclude_blocked is not None:
        arguments['excludeBlocked'] = exclude_blocked
    
    # Call agentic-tools MCP (tool name is just the base name, not prefixed)
    result = _call_agentic_tools_tool_sync(
        'get_next_task_recommendation',
        arguments,
        project_root
    )
    
    return result


def parse_prd_mcp(
    prd_content: str,
    project_id: Optional[str] = None,
    generate_subtasks: bool = True,
    default_priority: int = 5,
    estimate_complexity: bool = True,
    project_root: Optional[Path] = None
) -> Optional[Dict[str, Any]]:
    """
    Parse a Product Requirements Document and automatically generate structured tasks.
    
    Uses agentic-tools `parse_prd` to transform PRD content into actionable tasks
    with dependencies, priorities, and complexity estimates.
    
    Args:
        prd_content: Content of the Product Requirements Document to parse
        project_id: ID of the project to add tasks to
        generate_subtasks: Whether to generate subtasks for complex tasks (default: True)
        default_priority: Default priority for generated tasks (1-10, default: 5)
        estimate_complexity: Whether to estimate complexity for tasks (default: True)
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        Dictionary with generated tasks and metadata
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Prepare arguments
    arguments = {
        'prdContent': prd_content,
    }
    if project_id:
        arguments['projectId'] = project_id
    if generate_subtasks is not None:
        arguments['generateSubtasks'] = generate_subtasks
    if default_priority:
        arguments['defaultPriority'] = default_priority
    if estimate_complexity is not None:
        arguments['estimateComplexity'] = estimate_complexity
    
    # Call agentic-tools MCP (tool name is just the base name, not prefixed)
    result = _call_agentic_tools_tool_sync(
        'parse_prd',
        arguments,
        project_root
    )
    
    return result


def analyze_task_complexity_mcp(
    task_id: str,
    project_id: Optional[str] = None,
    complexity_threshold: int = 7,
    suggest_breakdown: bool = True,
    auto_create_subtasks: bool = False,
    project_root: Optional[Path] = None
) -> Optional[Dict[str, Any]]:
    """
    Analyze task complexity and suggest breaking down overly complex tasks.
    
    Uses agentic-tools `analyze_task_complexity` to identify tasks that should
    be split for better productivity and progress tracking.
    
    Args:
        task_id: Specific task ID to analyze (or all tasks if not provided)
        project_id: Filter analysis to a specific project
        complexity_threshold: Complexity threshold above which tasks should be broken down (1-10, default: 7)
        suggest_breakdown: Whether to suggest specific task breakdowns (default: True)
        auto_create_subtasks: Whether to automatically create suggested subtasks (default: False)
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        Dictionary with complexity analysis and breakdown suggestions
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Prepare arguments
    arguments = {}
    if task_id:
        arguments['taskId'] = task_id
    if project_id:
        arguments['projectId'] = project_id
    if complexity_threshold:
        arguments['complexityThreshold'] = complexity_threshold
    if suggest_breakdown is not None:
        arguments['suggestBreakdown'] = suggest_breakdown
    if auto_create_subtasks is not None:
        arguments['autoCreateSubtasks'] = auto_create_subtasks
    
    # Call agentic-tools MCP (tool name is just the base name, not prefixed)
    result = _call_agentic_tools_tool_sync(
        'analyze_task_complexity',
        arguments,
        project_root
    )
    
    return result


def research_task_mcp(
    task_id: str,
    project_id: Optional[str] = None,
    research_areas: Optional[List[str]] = None,
    research_depth: str = 'standard',
    save_to_memories: bool = True,
    check_existing_memories: bool = True,
    project_root: Optional[Path] = None
) -> Optional[Dict[str, Any]]:
    """
    Perform comprehensive web research for a task with intelligent suggestions.
    
    Uses agentic-tools `research_task` to conduct research and automatically
    store findings in memory for future reference.
    
    Args:
        task_id: ID of the task to research
        project_id: ID of the project this task belongs to
        research_areas: Specific areas to research (auto-generated if not provided)
        research_depth: Depth of research - "quick", "standard", or "comprehensive" (default: "standard")
        save_to_memories: Whether to save research findings to memories (default: True)
        check_existing_memories: Whether to check existing memories first (default: True)
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        Dictionary with research findings and metadata
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Prepare arguments
    arguments = {
        'taskId': task_id,
    }
    if project_id:
        arguments['projectId'] = project_id
    if research_areas:
        arguments['researchAreas'] = research_areas
    if research_depth:
        arguments['researchDepth'] = research_depth
    if save_to_memories is not None:
        arguments['saveToMemories'] = save_to_memories
    if check_existing_memories is not None:
        arguments['checkExistingMemories'] = check_existing_memories
    
    # Call agentic-tools MCP (tool name is just the base name, not prefixed)
    result = _call_agentic_tools_tool_sync(
        'research_task',
        arguments,
        project_root
    )
    
    return result


__all__ = [
    'infer_task_progress_mcp',
    'generate_research_queries_mcp',
    'get_next_task_recommendation_mcp',
    'parse_prd_mcp',
    'analyze_task_complexity_mcp',
    'research_task_mcp',
]

