"""
Todo2 MCP Client Helper

Provides Python interface to Todo2 MCP server.
Uses connection pooling to reuse sessions across multiple calls for better performance.
Falls back to direct file access if MCP unavailable.
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


def _load_todo2_file(project_root: Path) -> Optional[dict]:
    """Fallback: Load Todo2 state directly from file."""
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    if not todo2_file.exists():
        return None
    
    try:
        with open(todo2_file) as f:
            return json.load(f)
    except Exception as e:
        logger.error(f"Failed to load Todo2 file: {e}")
        return None


async def _call_todo2_tool(tool_name: str, arguments: dict, project_root: Path) -> Optional[dict]:
    """Call a Todo2 tool via MCP with connection pooling."""
    if not MCP_CLIENT_AVAILABLE:
        logger.debug("MCP client library not available, will use file fallback")
        return None
    
    mcp_config = _load_mcp_config(project_root)
    if 'todo2' not in mcp_config:
        logger.debug("Todo2 MCP server not configured, will use file fallback")
        return None
    
    todo2_config = mcp_config.get('todo2', {})
    
    # Check if Todo2 uses URL-based connection (HTTP/SSE)
    if 'url' in todo2_config:
        logger.warning("Todo2 URL-based MCP server not yet supported for direct calls. Use file fallback.")
        return None
    
    # For stdio-based servers, we'd need command/args
    command = todo2_config.get('command')
    if not command:
        logger.debug("Todo2 MCP server command not found, will use file fallback")
        return None
    
    args = todo2_config.get('args', [])
    env = todo2_config.get('env', {})
    
    try:
        server_params = StdioServerParameters(
            command=command,
            args=args,
            env=env
        )
        
        # Use connection pool if available, otherwise fall back to direct connection
        if POOL_AVAILABLE and _session_pool is not None:
            async with _session_pool.get_session('todo2', server_params) as session:
                result = await session.call_tool(tool_name, arguments)
                
                # Parse JSON response
                if result.content and len(result.content) > 0:
                    response_text = result.content[0].text
                    try:
                        return json.loads(response_text)
                    except json.JSONDecodeError:
                        logger.error(f"Failed to parse Todo2 tool response: {response_text}")
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
                            logger.error(f"Failed to parse Todo2 tool response: {response_text}")
                            return None
                    return None
    except Exception as e:
        logger.error(f"Failed to call Todo2 tool {tool_name}: {e}", exc_info=True)
        return None


def _call_todo2_tool_sync(tool_name: str, arguments: dict, project_root: Path) -> Optional[dict]:
    """Synchronous wrapper for async Todo2 tool calls."""
    try:
        return asyncio.run(_call_todo2_tool(tool_name, arguments, project_root))
    except Exception as e:
        logger.error(f"Failed to call Todo2 tool {tool_name} (sync): {e}")
        return None


# Public API functions

def list_todos_mcp(
    project_root: Optional[Path] = None,
    status: Optional[str] = None,
    priority: Optional[str] = None,
    tags: Optional[List[str]] = None,
    ready_only: bool = False
) -> List[dict]:
    """
    List todos using Todo2 MCP, with fallback to file access.
    
    Args:
        project_root: Project root path (defaults to find_project_root)
        status: Filter by status (Todo, In Progress, Done)
        priority: Filter by priority (low, medium, high, critical)
        tags: Filter by tags (list of tag strings)
        ready_only: Only return tasks ready to start (dependencies completed)
    
    Returns:
        List of task dictionaries
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Try MCP first
    arguments = {}
    if status:
        arguments['status'] = status
    if priority:
        arguments['priority'] = priority
    if tags:
        arguments['tags'] = tags
    if ready_only:
        arguments['ready_only'] = ready_only
    
    result = _call_todo2_tool_sync('mcp_extension-todo2_list_todos', arguments, project_root)
    
    if result is not None:
        # Parse MCP response format
        # Assuming result contains 'todos' key or is a list
        if isinstance(result, list):
            return result
        elif isinstance(result, dict) and 'todos' in result:
            return result['todos']
        elif isinstance(result, dict) and 'data' in result:
            return result['data']
        else:
            logger.warning(f"Unexpected Todo2 MCP response format: {result}")
    
    # Fallback to file access
    logger.debug("Falling back to direct file access for list_todos")
    state = _load_todo2_file(project_root)
    if state is None:
        return []
    
    tasks = state.get('todos', [])
    
    # Apply filters manually
    filtered = []
    for task in tasks:
        if status and task.get('status') != status:
            continue
        if priority and task.get('priority') != priority:
            continue
        if tags:
            task_tags = task.get('tags', [])
            if not all(tag in task_tags for tag in tags):
                continue
        if ready_only:
            # Check if dependencies are completed
            deps = task.get('dependencies', [])
            if deps:
                from .todo2_utils import is_completed_status
                # Extract dependency IDs (handle both string and dict formats)
                dep_ids = [d if isinstance(d, str) else d.get('id') for d in deps]
                # Check if all dependencies are completed (using normalized status check)
                completed_deps = [
                    t for t in tasks 
                    if t.get('id') in dep_ids and is_completed_status(t.get('status', ''))
                ]
                # Task is only ready if ALL dependencies are completed
                if len(completed_deps) < len(dep_ids):
                    continue
        filtered.append(task)
    
    return filtered


def create_todos_mcp(
    todos: List[dict],
    project_root: Optional[Path] = None
) -> Optional[List[str]]:
    """
    Create todos using Todo2 MCP, with fallback to file access.
    
    Args:
        todos: List of todo dictionaries with name, long_description, priority, tags, etc.
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        List of created todo IDs, or None if failed
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Try MCP first
    arguments = {'todos': todos}
    result = _call_todo2_tool_sync('mcp_extension-todo2_create_todos', arguments, project_root)
    
    if result is not None:
        # Parse MCP response format
        if isinstance(result, dict):
            if 'todos' in result:
                return [t.get('id') for t in result['todos']]
            elif 'ids' in result:
                return result['ids']
            elif 'data' in result:
                data = result['data']
                if isinstance(data, list):
                    return [t.get('id') if isinstance(t, dict) else str(t) for t in data]
        logger.warning(f"Unexpected Todo2 MCP response format: {result}")
    
    # Fallback to file access
    logger.debug("Falling back to direct file access for create_todos")
    state = _load_todo2_file(project_root)
    if state is None:
        logger.error("Cannot create todos: Todo2 state file not found")
        return None
    
    todos_list = state.get('todos', [])
    created_ids = []
    
    # Generate IDs and add tasks
    for todo in todos:
        # Generate simple ID (in real Todo2, IDs are more complex)
        task_id = f"T-{len(todos_list) + len(created_ids) + 1}"
        task = {
            'id': task_id,
            'name': todo.get('name', ''),
            'long_description': todo.get('long_description', ''),
            'status': todo.get('status', 'Todo'),
            'priority': todo.get('priority', 'medium'),
            'tags': todo.get('tags', []),
            'dependencies': todo.get('dependencies', []),
        }
        todos_list.append(task)
        created_ids.append(task_id)
    
    # Write back to file
    state['todos'] = todos_list
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    try:
        with open(todo2_file, 'w') as f:
            json.dump(state, f, indent=2)
        return created_ids
    except Exception as e:
        logger.error(f"Failed to write Todo2 file: {e}")
        return None


def update_todos_mcp(
    updates: List[dict],
    project_root: Optional[Path] = None
) -> bool:
    """
    Update todos using Todo2 MCP, with fallback to file access.
    
    Automatically normalizes status values to Title Case (Todo, In Progress, Done, etc.)
    to ensure consistency across all updates.
    
    Args:
        updates: List of update dictionaries with 'id' and fields to update
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        True if successful, False otherwise
    """
    from .project_root import find_project_root
    from .todo2_utils import normalize_status_to_title_case
    
    if project_root is None:
        project_root = find_project_root()
    
    # Normalize status values to Title Case before updating
    normalized_updates = []
    for update in updates:
        normalized_update = update.copy()
        if 'status' in normalized_update:
            normalized_update['status'] = normalize_status_to_title_case(normalized_update['status'])
        normalized_updates.append(normalized_update)
    
    # Try MCP first
    arguments = {'updates': normalized_updates}
    result = _call_todo2_tool_sync('mcp_extension-todo2_update_todos', arguments, project_root)
    
    if result is not None:
        # Check if update was successful
        if isinstance(result, dict):
            return result.get('success', False) or 'updated' in result or 'data' in result
        return True
    
    # Fallback to file access
    logger.debug("Falling back to direct file access for update_todos")
    state = _load_todo2_file(project_root)
    if state is None:
        logger.error("Cannot update todos: Todo2 state file not found")
        return False
    
    todos_list = state.get('todos', [])
    task_dict = {t.get('id'): t for t in todos_list}
    
    # Apply updates
    for update in updates:
        task_id = update.get('id')
        if task_id not in task_dict:
            logger.warning(f"Task {task_id} not found for update")
            continue
        
        task = task_dict[task_id]
        # Update fields (normalize status if present)
        for key, value in update.items():
            if key != 'id':
                if key == 'status':
                    from .todo2_utils import normalize_status_to_title_case
                    value = normalize_status_to_title_case(value)
                task[key] = value
    
    # Write back to file
    state['todos'] = list(task_dict.values())
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    try:
        with open(todo2_file, 'w') as f:
            json.dump(state, f, indent=2)
        return True
    except Exception as e:
        logger.error(f"Failed to write Todo2 file: {e}")
        return False


def get_todo_details_mcp(
    ids: List[str],
    project_root: Optional[Path] = None
) -> List[dict]:
    """
    Get detailed todo information using Todo2 MCP, with fallback to file access.
    
    Args:
        ids: List of todo IDs to get details for
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        List of detailed todo dictionaries with comments
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Try MCP first
    arguments = {'ids': ids}
    result = _call_todo2_tool_sync('mcp_extension-todo2_get_todo_details', arguments, project_root)
    
    if result is not None:
        # Parse MCP response format
        if isinstance(result, list):
            return result
        elif isinstance(result, dict):
            if 'todos' in result:
                return result['todos']
            elif 'data' in result:
                return result['data'] if isinstance(result['data'], list) else [result['data']]
        logger.warning(f"Unexpected Todo2 MCP response format: {result}")
    
    # Fallback to file access
    logger.debug("Falling back to direct file access for get_todo_details")
    state = _load_todo2_file(project_root)
    if state is None:
        return []
    
    todos_list = state.get('todos', [])
    task_dict = {t.get('id'): t for t in todos_list}
    
    # Return requested tasks
    result_tasks = []
    for task_id in ids:
        if task_id in task_dict:
            task = task_dict[task_id].copy()
            # Add comments if available in state
            if 'comments' in state:
                task['comments'] = [c for c in state.get('comments', []) if c.get('todoId') == task_id]
            result_tasks.append(task)
    
    return result_tasks


def add_comments_mcp(
    todo_id: str,
    comments: List[dict],
    project_root: Optional[Path] = None
) -> bool:
    """
    Add comments to a todo using Todo2 MCP, with fallback to file access.
    
    Args:
        todo_id: ID of the todo to add comments to
        comments: List of comment dictionaries with 'type' and 'content'
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        True if successful, False otherwise
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Try MCP first
    arguments = {'todoId': todo_id, 'comments': comments}
    result = _call_todo2_tool_sync('mcp_extension-todo2_add_comments', arguments, project_root)
    
    if result is not None:
        # Check if add was successful
        if isinstance(result, dict):
            return result.get('success', False) or 'added' in result or 'data' in result
        return True
    
    # Fallback to file access
    logger.debug("Falling back to direct file access for add_comments")
    state = _load_todo2_file(project_root)
    if state is None:
        logger.error("Cannot add comments: Todo2 state file not found")
        return False
    
    # Ensure comments structure exists
    if 'comments' not in state:
        state['comments'] = []
    
    # Add comments
    for comment in comments:
        comment['todoId'] = todo_id
        state['comments'].append(comment)
    
    # Write back to file
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    try:
        with open(todo2_file, 'w') as f:
            json.dump(state, f, indent=2)
        return True
    except Exception as e:
        logger.error(f"Failed to write Todo2 file: {e}")
        return False


def delete_todos_mcp(
    ids: List[str],
    project_root: Optional[Path] = None
) -> bool:
    """
    Delete todos using Todo2 MCP, with fallback to file access.
    
    Args:
        ids: List of todo IDs to delete
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        True if successful, False otherwise
    """
    from .project_root import find_project_root
    
    if project_root is None:
        project_root = find_project_root()
    
    # Try MCP first
    arguments = {'ids': ids}
    result = _call_todo2_tool_sync('mcp_extension-todo2_delete_todos', arguments, project_root)
    
    if result is not None:
        # Check if delete was successful
        if isinstance(result, dict):
            return result.get('success', False) or 'deleted' in result or 'data' in result
        return True
    
    # Fallback to file access
    logger.debug("Falling back to direct file access for delete_todos")
    state = _load_todo2_file(project_root)
    if state is None:
        logger.error("Cannot delete todos: Todo2 state file not found")
        return False
    
    todos_list = state.get('todos', [])
    # Filter out deleted tasks
    state['todos'] = [t for t in todos_list if t.get('id') not in ids]
    
    # Also remove associated comments
    if 'comments' in state:
        state['comments'] = [c for c in state.get('comments', []) if c.get('todoId') not in ids]
    
    # Write back to file
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    try:
        with open(todo2_file, 'w') as f:
            json.dump(state, f, indent=2)
        return True
    except Exception as e:
        logger.error(f"Failed to write Todo2 file: {e}")
        return False


__all__ = [
    'list_todos_mcp',
    'create_todos_mcp',
    'update_todos_mcp',
    'get_todo_details_mcp',
    'add_comments_mcp',
    'delete_todos_mcp',
]

