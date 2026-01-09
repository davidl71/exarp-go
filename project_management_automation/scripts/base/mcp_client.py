#!/usr/bin/env python3
"""
MCP Client Wrapper

Provides Python interface to MCP servers (Tractatus Thinking, Sequential Thinking, Agentic-Tools).
Uses connection pooling to reuse sessions across multiple calls for better performance.
"""

import asyncio
import json
import logging
import time
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Any, AsyncContextManager, Dict, List, Optional, Tuple

logger = logging.getLogger(__name__)

# Retry configuration
MAX_RETRIES = 3
RETRY_DELAY = 0.5  # Base delay in seconds (exponential backoff)

# Connection pool configuration
SESSION_TIMEOUT = 300  # 5 minutes - close idle sessions
MAX_SESSION_AGE = 3600  # 1 hour - maximum session lifetime

# Try to import MCP client library
try:
    from mcp import ClientSession, StdioServerParameters
    from mcp.client.stdio import stdio_client
    MCP_CLIENT_AVAILABLE = True
except ImportError:
    MCP_CLIENT_AVAILABLE = False
    # Create type aliases for when MCP is not available
    from typing import Any
    ClientSession = Any
    StdioServerParameters = Any
    logger.debug("MCP client library not available (optional - server uses FastMCP). Install with: uv sync (or uv pip install mcp>=1.0.0)")


class MCPSessionPool:
    """
    Connection pool for MCP server sessions.
    
    Maintains reusable sessions to avoid the overhead of creating new processes
    for each tool call. Sessions are reused across multiple calls and automatically
    recreated on errors or after timeout.
    """
    
    def __init__(self):
        self._pools: Dict[str, "_ServerPool"] = {}
        self._lock = asyncio.Lock()
    
    def get_session(
        self, 
        server_name: str, 
        server_params: "StdioServerParameters"
    ) -> "AsyncContextManager[ClientSession]":
        """
        Get a reusable session for the specified server.
        
        Args:
            server_name: Name of the MCP server (e.g., 'agentic-tools')
            server_params: Server parameters for connection
            
        Returns:
            Async context manager that yields a ClientSession
        """
        # Create pool entry if it doesn't exist (synchronously)
        # The actual async locking happens inside _ServerPool.get_session()
        if server_name not in self._pools:
            self._pools[server_name] = _ServerPool(server_name, server_params)
        # Return the async context manager directly
        return self._pools[server_name].get_session()
    
    async def close_all(self):
        """Close all sessions in the pool."""
        async with self._lock:
            for pool in self._pools.values():
                await pool.close()
            self._pools.clear()


class _ServerPool:
    """Pool for a single MCP server type."""
    
    def __init__(self, server_name: str, server_params: StdioServerParameters):
        self.server_name = server_name
        self.server_params = server_params
        self._client_context: Optional[Any] = None
        self._session_context: Optional[Any] = None
        self._session: Optional[ClientSession] = None
        self._read_write: Optional[Tuple] = None
        self._lock = asyncio.Lock()
        self._last_used = time.time()
        self._created_at = time.time()
        self._in_use = False
    
    @asynccontextmanager
    async def get_session(self):
        """Get or create a session, reusing if available and healthy."""
        async with self._lock:
            # Check if we need to recreate the session
            if self._session is None or not self._is_healthy():
                await self._recreate_session()
            
            self._last_used = time.time()
            self._in_use = True
        
        try:
            if self._session is None:
                raise RuntimeError(f"Failed to create session for {self.server_name}")
            yield self._session
        finally:
            async with self._lock:
                self._in_use = False
                self._last_used = time.time()
    
    def _is_healthy(self) -> bool:
        """Check if the current session is healthy and should be reused."""
        if self._session is None:
            return False
        
        # Check session age
        age = time.time() - self._created_at
        if age > MAX_SESSION_AGE:
            logger.debug(f"Session for {self.server_name} expired (age: {age:.1f}s)")
            return False
        
        # Check idle timeout
        idle_time = time.time() - self._last_used
        if idle_time > SESSION_TIMEOUT:
            logger.debug(f"Session for {self.server_name} timed out (idle: {idle_time:.1f}s)")
            return False
        
        return True
    
    async def _recreate_session(self):
        """Create a new session, closing the old one if it exists."""
        # Close existing session
        await self._close_session()
        
        # Create new session
        try:
            self._client_context = stdio_client(self.server_params)
            self._read_write = await self._client_context.__aenter__()
            read, write = self._read_write
            
            self._session_context = ClientSession(read, write)
            self._session = await self._session_context.__aenter__()
            await self._session.initialize()
            
            self._created_at = time.time()
            self._last_used = time.time()
            logger.debug(f"Created new session for {self.server_name}")
        except Exception as e:
            logger.error(f"Failed to create session for {self.server_name}: {e}")
            await self._close_session()
            raise
    
    async def _close_session(self):
        """Close the current session and clean up."""
        if self._session_context is not None:
            try:
                await self._session_context.__aexit__(None, None, None)
            except Exception as e:
                logger.warning(f"Error closing session context for {self.server_name}: {e}")
            self._session_context = None
        
        if self._client_context is not None:
            try:
                await self._client_context.__aexit__(None, None, None)
            except Exception as e:
                logger.warning(f"Error closing client context for {self.server_name}: {e}")
            self._client_context = None
        
        self._session = None
        self._read_write = None
    
    async def close(self):
        """Close the session and clean up."""
        async with self._lock:
            await self._close_session()


# Global session pool (exported for use by other modules)
_session_pool = MCPSessionPool()

# Export session pool for use in other modules
__all__ = ['MCPClient', 'get_mcp_client', 'load_json_with_retry', '_session_pool']


class MCPClient:
    """Client for communicating with MCP servers using lazy connection."""

    def __init__(self, project_root: Path):
        self.project_root = project_root
        self.mcp_config_path = project_root / '.cursor' / 'mcp.json'
        # Lazy loading - don't load config until needed
        self._mcp_config: Optional[dict] = None
        self._config_loaded = False
        # Agentic-tools session management
        self.agentic_tools_session = None
        self._agentic_tools_lock = asyncio.Lock() if MCP_CLIENT_AVAILABLE else None

    @property
    def mcp_config(self) -> dict:
        """Lazy load MCP config on first access."""
        if not self._config_loaded:
            self._mcp_config = self._load_mcp_config_with_retry()
            self._config_loaded = True
        return self._mcp_config or {}

    def _load_mcp_config_with_retry(self) -> dict:
        """Load MCP server configuration with retry logic.
        
        Checks both project config (.cursor/mcp.json) and global config (~/.cursor/mcp.json).
        Merges them with project config taking precedence.
        """
        merged_config = {}
        
        # Try global config first (base configuration)
        global_config_path = Path.home() / '.cursor' / 'mcp.json'
        if global_config_path.exists():
            for attempt in range(MAX_RETRIES):
                try:
                    with open(global_config_path) as f:
                        global_config = json.load(f)
                        merged_config.update(global_config.get('mcpServers', {}))
                        break
                except json.JSONDecodeError as e:
                    if attempt < MAX_RETRIES - 1:
                        logger.debug(f"Global MCP config parse error, retry {attempt + 1}/{MAX_RETRIES}: {e}")
                        time.sleep(RETRY_DELAY * (attempt + 1))
                    else:
                        logger.warning(f"Failed to parse global MCP config after {MAX_RETRIES} attempts: {e}")
                except Exception as e:
                    logger.debug(f"Failed to load global MCP config: {e}")
                    break
        
        # Try project config (overrides global)
        if self.mcp_config_path.exists():
            for attempt in range(MAX_RETRIES):
                try:
                    with open(self.mcp_config_path) as f:
                        project_config = json.load(f)
                        # Project config overrides global config
                        merged_config.update(project_config.get('mcpServers', {}))
                        break
                except json.JSONDecodeError as e:
                    if attempt < MAX_RETRIES - 1:
                        logger.debug(f"Project MCP config parse error, retry {attempt + 1}/{MAX_RETRIES}: {e}")
                        time.sleep(RETRY_DELAY * (attempt + 1))
                    else:
                        logger.warning(f"Failed to parse project MCP config after {MAX_RETRIES} attempts: {e}")
                except Exception as e:
                    logger.warning(f"Failed to load project MCP config: {e}")
                    break
        
        return merged_config

    def call_tractatus_thinking(self, operation: str, **kwargs) -> Optional[dict]:
        """Call Tractatus Thinking MCP server."""
        if 'tractatus_thinking' not in self.mcp_config:
            logger.warning("Tractatus Thinking MCP server not configured")
            return None

        try:
            # For now, return a simplified response
            # In a full implementation, this would communicate with the MCP server
            # via stdio or HTTP
            logger.info(f"Tractatus Thinking: {operation}")

            # Simplified response structure
            if operation == "start":
                concept = kwargs.get('concept', '')
                return {
                    'session_id': f"tractatus_{hash(concept)}",
                    'concept': concept,
                    'components': self._extract_components_simple(concept)
                }

            return None
        except Exception as e:
            logger.warning(f"Tractatus Thinking call failed: {e}")
            return None

    def call_sequential_thinking(self, operation: str, **kwargs) -> Optional[dict]:
        """Call Sequential Thinking MCP server."""
        if 'sequential_thinking' not in self.mcp_config:
            logger.warning("Sequential Thinking MCP server not configured")
            return None

        try:
            logger.info(f"Sequential Thinking: {operation}")

            # Simplified response structure
            if operation == "start":
                problem = kwargs.get('problem', '')
                return {
                    'session_id': f"sequential_{hash(problem)}",
                    'problem': problem,
                    'steps': self._plan_steps_simple(problem)
                }

            return None
        except Exception as e:
            logger.warning(f"Sequential Thinking call failed: {e}")
            return None

    def _extract_components_simple(self, concept: str) -> list[str]:
        """Simple component extraction (fallback)."""
        # Look for × or * patterns indicating components
        components = []

        if '×' in concept or '*' in concept:
            # Split by × or *
            parts = concept.replace('×', '*').split('*')
            components = [p.strip() for p in parts if p.strip()]
        else:
            # Extract keywords
            keywords = ['automation', 'analysis', 'validation', 'monitoring',
                       'tracking', 'synchronization', 'health', 'alignment']
            components = [kw for kw in keywords if kw in concept.lower()]

        return components if components else ['general']

    def _plan_steps_simple(self, problem: str) -> list[str]:
        """Simple step planning (fallback)."""
        steps = [
            "Load and analyze data",
            "Identify patterns and opportunities",
            "Generate recommendations",
            "Create follow-up tasks"
        ]

        # Customize based on problem keywords
        if 'find' in problem.lower() or 'discover' in problem.lower():
            steps = [
                "Search for opportunities",
                "Analyze and score findings",
                "Prioritize recommendations",
                "Create implementation tasks"
            ]
        elif 'check' in problem.lower() or 'validate' in problem.lower():
            steps = [
                "Load data to validate",
                "Run validation checks",
                "Identify issues",
                "Generate fix recommendations"
            ]

        return steps

    # Agentic-Tools MCP Support
    def _get_agentic_tools_params(self) -> Optional[StdioServerParameters]:
        """Get agentic-tools server parameters."""
        if not MCP_CLIENT_AVAILABLE:
            return None
        
        if 'agentic-tools' not in self.mcp_config:
            return None
        
        agentic_config = self.mcp_config.get('agentic-tools', {})
        command = agentic_config.get('command', 'npx')
        args = agentic_config.get('args', ['-y', '@modelcontextprotocol/server-agentic-tools'])
        
        return StdioServerParameters(command=command, args=args)
    
    async def _call_agentic_tool(
        self, 
        tool_name: str, 
        arguments: dict,
        retry_on_error: bool = True
    ) -> Optional[dict]:
        """
        Call an agentic-tools MCP tool using connection pooling.
        
        Args:
            tool_name: Name of the tool to call
            arguments: Tool arguments
            retry_on_error: Whether to retry once on connection errors
            
        Returns:
            Parsed JSON response or None on failure
        """
        if not MCP_CLIENT_AVAILABLE:
            logger.debug("MCP client library not available (optional - server uses FastMCP)")
            return None
        
        server_params = self._get_agentic_tools_params()
        if server_params is None:
            logger.warning("Agentic-tools MCP server not configured")
            return None
        
        max_attempts = 2 if retry_on_error else 1
        for attempt in range(max_attempts):
            try:
                async with _session_pool.get_session('agentic-tools', server_params) as session:
                    result = await session.call_tool(tool_name, arguments)
                    
                    # Parse JSON response
                    if result.content and len(result.content) > 0:
                        response_text = result.content[0].text
                        try:
                            return json.loads(response_text)
                        except json.JSONDecodeError:
                            logger.error(f"Failed to parse {tool_name} response: {response_text}")
                            return None
                    return None
            except Exception as e:
                if attempt < max_attempts - 1:
                    logger.warning(f"Error calling {tool_name} (attempt {attempt + 1}), retrying: {e}")
                    # Force session recreation on retry
                    async with _session_pool._lock:
                        if 'agentic-tools' in _session_pool._pools:
                            await _session_pool._pools['agentic-tools'].close()
                            del _session_pool._pools['agentic-tools']
                    await asyncio.sleep(RETRY_DELAY * (attempt + 1))
                else:
                    logger.error(f"Failed to call {tool_name} via agentic-tools MCP: {e}", exc_info=True)
                    return None
        
        return None

    async def list_todos(self, project_id: str, working_directory: str) -> List[Dict]:
        """
        List todos using agentic-tools MCP with connection pooling.
        
        Args:
            project_id: The project ID to list todos for
            working_directory: The working directory for the project
            
        Returns:
            List of task dictionaries
        """
        result = await self._call_agentic_tool("list_todos", {
            "workingDirectory": working_directory,
            "projectId": project_id
        })
        
        if result is None:
            return []
        
        # Extract tasks from response (structure may vary)
        if isinstance(result, list):
            return result
        elif isinstance(result, dict):
            return result.get('todos', result.get('tasks', []))
        return []

    async def create_task(
        self,
        project_id: str,
        working_directory: str,
        name: str,
        details: str,
        **kwargs
    ) -> Optional[Dict]:
        """
        Create task using agentic-tools MCP with connection pooling.
        
        Args:
            project_id: The project ID
            working_directory: The working directory for the project
            name: Task name
            details: Task details/description
            **kwargs: Additional task parameters (priority, tags, etc.)
            
        Returns:
            Created task dictionary or None on failure
        """
        task_data = {
            "workingDirectory": working_directory,
            "projectId": project_id,
            "name": name,
            "details": details,
            **kwargs
        }
        
        return await self._call_agentic_tool("create_task", task_data)

    async def update_task(
        self,
        task_id: str,
        working_directory: str,
        **updates
    ) -> Optional[Dict]:
        """
        Update task using agentic-tools MCP with connection pooling.
        
        Args:
            task_id: The task ID to update
            working_directory: The working directory for the project
            **updates: Task fields to update (name, details, status, priority, etc.)
            
        Returns:
            Updated task dictionary or None on failure
        """
        update_data = {
            "workingDirectory": working_directory,
            "id": task_id,
            **updates
        }
        
        return await self._call_agentic_tool("update_task", update_data)

    async def get_task(self, task_id: str, working_directory: str) -> Optional[Dict]:
        """
        Get task details using agentic-tools MCP with connection pooling.
        
        Args:
            task_id: The task ID
            working_directory: The working directory for the project
            
        Returns:
            Task dictionary or None on failure
        """
        return await self._call_agentic_tool("get_task", {
            "workingDirectory": working_directory,
            "id": task_id
        })

    async def delete_task(self, task_id: str, working_directory: str) -> bool:
        """
        Delete task using agentic-tools MCP with connection pooling.
        
        Args:
            task_id: The task ID to delete
            working_directory: The working directory for the project
            
        Returns:
            True if successful, False otherwise
        """
        result = await self._call_agentic_tool("delete_task", {
            "workingDirectory": working_directory,
            "id": task_id
        })
        
        return result is not None
    
    async def batch_operations(
        self,
        operations: List[Dict[str, Any]],
        working_directory: str
    ) -> List[Optional[Dict]]:
        """
        Execute multiple tool calls in a single session for better performance.
        
        This method reuses the same connection for all operations, significantly
        reducing overhead when making multiple calls.
        
        Args:
            operations: List of operation dicts, each with:
                - 'tool': Tool name (e.g., 'create_task', 'update_task')
                - 'arguments': Tool arguments dict
            working_directory: The working directory for the project
            
        Returns:
            List of results (one per operation), None for failed operations
            
        Example:
            results = await client.batch_operations([
                {
                    'tool': 'create_task',
                    'arguments': {
                        'projectId': 'my-project',
                        'workingDirectory': '/path',
                        'name': 'Task 1',
                        'details': 'Description'
                    }
                },
                {
                    'tool': 'create_task',
                    'arguments': {
                        'projectId': 'my-project',
                        'workingDirectory': '/path',
                        'name': 'Task 2',
                        'details': 'Description'
                    }
                }
            ], '/path')
        """
        if not MCP_CLIENT_AVAILABLE:
            logger.debug("MCP client library not available (optional - server uses FastMCP)")
            return [None] * len(operations)
        
        server_params = self._get_agentic_tools_params()
        if server_params is None:
            logger.warning("Agentic-tools MCP server not configured")
            return [None] * len(operations)
        
        results = []
        try:
            # Use a single session for all operations
            async with _session_pool.get_session('agentic-tools', server_params) as session:
                for op in operations:
                    tool_name = op.get('tool')
                    arguments = op.get('arguments', {})
                    
                    if not tool_name:
                        logger.warning(f"Missing 'tool' in operation: {op}")
                        results.append(None)
                        continue
                    
                    try:
                        result = await session.call_tool(tool_name, arguments)
                        
                        # Parse JSON response
                        if result.content and len(result.content) > 0:
                            response_text = result.content[0].text
                            try:
                                results.append(json.loads(response_text))
                            except json.JSONDecodeError:
                                logger.error(f"Failed to parse {tool_name} response: {response_text}")
                                results.append(None)
                        else:
                            results.append(None)
                    except Exception as e:
                        logger.error(f"Error in batch operation {tool_name}: {e}")
                        results.append(None)
        except Exception as e:
            logger.error(f"Failed batch operations: {e}", exc_info=True)
            # Return None for all operations on connection failure
            return [None] * len(operations)
        
        return results


# Global MCP client instance
_mcp_client = None

def get_mcp_client(project_root: Path) -> MCPClient:
    """Get or create MCP client instance."""
    global _mcp_client
    if _mcp_client is None:
        _mcp_client = MCPClient(project_root)
    return _mcp_client


def load_json_with_retry(
    file_path: Path,
    max_retries: int = MAX_RETRIES,
    retry_delay: float = RETRY_DELAY,
    default: Any = None
) -> Any:
    """
    Load JSON file with retry logic for race conditions.

    Handles cases where file is being written by another process
    (e.g., another MCP server updating its state file).

    Args:
        file_path: Path to JSON file
        max_retries: Maximum number of retry attempts
        retry_delay: Base delay between retries (exponential backoff)
        default: Default value if file cannot be loaded

    Returns:
        Parsed JSON data or default value
    """
    for attempt in range(max_retries):
        if not file_path.exists():
            if attempt < max_retries - 1:
                logger.debug(f"File not found, retry {attempt + 1}/{max_retries}: {file_path}")
                time.sleep(retry_delay * (attempt + 1))
                continue
            logger.debug(f"File not found after {max_retries} attempts: {file_path}")
            return default

        try:
            with open(file_path) as f:
                return json.load(f)
        except json.JSONDecodeError as e:
            # File might be partially written - retry
            if attempt < max_retries - 1:
                logger.debug(f"JSON parse error, retry {attempt + 1}/{max_retries}: {e}")
                time.sleep(retry_delay * (attempt + 1))
            else:
                logger.warning(f"Failed to parse JSON after {max_retries} attempts: {file_path}")
                return default
        except PermissionError as e:
            # File might be locked by another process - retry
            if attempt < max_retries - 1:
                logger.debug(f"File locked, retry {attempt + 1}/{max_retries}: {e}")
                time.sleep(retry_delay * (attempt + 1))
            else:
                logger.warning(f"File permission error after {max_retries} attempts: {file_path}")
                return default
        except Exception as e:
            logger.warning(f"Unexpected error loading {file_path}: {e}")
            return default

    return default
