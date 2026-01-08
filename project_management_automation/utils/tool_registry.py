"""
Tool Registry - Single Source of Truth

Consolidates tool metadata for both FastMCP and stdio server.
Eliminates duplication between @mcp.tool() decorators and manual Tool() definitions.
"""

import inspect
import json
import re
from typing import Any, Dict, List, Optional

logger = None
try:
    import logging
    logger = logging.getLogger(__name__)
except:
    pass


def extract_tool_metadata(func) -> Dict[str, Any]:
    """
    Extract tool metadata from a function's docstring and signature.
    
    Args:
        func: The tool function
        
    Returns:
        Dictionary with name, description, and inputSchema
    """
    name = func.__name__
    docstring = func.__doc__ or ""
    
    # Extract HINT from docstring
    hint_match = re.search(r'\[HINT:(.*?)\]', docstring, re.DOTALL)
    hint = hint_match.group(1).strip() if hint_match else ""
    
    # Extract description (everything after HINT or first line)
    description_lines = []
    in_hint = False
    for line in docstring.split('\n'):
        if '[HINT:' in line:
            in_hint = True
            # Extract text after HINT
            if ']' in line:
                hint_end = line.index(']') + 1
                remaining = line[hint_end:].strip()
                if remaining:
                    description_lines.append(remaining)
                in_hint = False
            continue
        if in_hint and ']' in line:
            in_hint = False
            remaining = line[line.index(']') + 1:].strip()
            if remaining:
                description_lines.append(remaining)
            continue
        if not in_hint and line.strip():
            description_lines.append(line.strip())
    
    description = ' '.join(description_lines[:3]) if description_lines else name
    
    # Extract inputSchema from function signature
    sig = inspect.signature(func)
    properties = {}
    
    for param_name, param in sig.parameters.items():
        if param_name == 'self':
            continue
            
        param_type = "string"
        param_desc = ""
        param_default = None
        
        # Get type from annotation
        if param.annotation != inspect.Parameter.empty:
            if param.annotation == bool:
                param_type = "boolean"
            elif param.annotation == int:
                param_type = "integer"
            elif param.annotation == float:
                param_type = "number"
            elif param.annotation == list or (hasattr(param.annotation, '__origin__') and param.annotation.__origin__ is list):
                param_type = "array"
            elif param.annotation == dict or (hasattr(param.annotation, '__origin__') and param.annotation.__origin__ is dict):
                param_type = "object"
        
        # Get default value
        if param.default != inspect.Parameter.empty:
            param_default = param.default
        
        # Try to extract description from docstring
        param_pattern = rf'{param_name}:\s*(.+?)(?:\n|$)'
        param_match = re.search(param_pattern, docstring)
        if param_match:
            param_desc = param_match.group(1).strip()
        
        prop = {"type": param_type}
        if param_desc:
            prop["description"] = param_desc
        if param_default is not None:
            prop["default"] = param_default
        
        properties[param_name] = prop
    
    input_schema = {
        "type": "object",
        "properties": properties,
    }
    
    return {
        "name": name,
        "description": description,
        "hint": hint,
        "inputSchema": input_schema,
    }


def build_tool_from_function(func, Tool) -> Any:
    """
    Build a Tool object from a function for stdio server.
    
    Args:
        func: The tool function
        Tool: The Tool class from mcp.types
        
    Returns:
        Tool object
    """
    metadata = extract_tool_metadata(func)
    
    return Tool(
        name=metadata["name"],
        description=metadata["description"],
        inputSchema=metadata["inputSchema"],
    )


def get_fastmcp_tools(mcp_instance) -> List[Dict[str, Any]]:
    """
    Extract tool metadata from FastMCP instance.
    
    Note: FastMCP doesn't expose its tool registry directly,
    so we need to use introspection or maintain a separate registry.
    
    Args:
        mcp_instance: FastMCP instance
        
    Returns:
        List of tool metadata dictionaries
    """
    # FastMCP doesn't expose tools directly, so we'll need to
    # maintain a registry or use AST parsing
    # For now, return empty list - will be populated by register_tool_metadata
    return []


# Tool metadata registry (populated during registration)
_TOOL_REGISTRY: Dict[str, Dict[str, Any]] = {}


def register_tool_metadata(name: str, func, description: Optional[str] = None, input_schema: Optional[Dict] = None):
    """
    Register tool metadata for use by stdio server.
    
    Args:
        name: Tool name
        func: Tool function
        description: Optional override description
        input_schema: Optional override inputSchema
    """
    metadata = extract_tool_metadata(func)
    
    if description:
        metadata["description"] = description
    if input_schema:
        metadata["inputSchema"] = input_schema
    
    _TOOL_REGISTRY[name] = metadata


def get_tool_metadata(name: str) -> Optional[Dict[str, Any]]:
    """Get registered tool metadata."""
    return _TOOL_REGISTRY.get(name)


def build_stdio_tools(Tool, additional_tools: Optional[List[Any]] = None) -> List[Any]:
    """
    Build Tool objects for stdio server from registry.
    
    Args:
        Tool: Tool class from mcp.types
        additional_tools: Additional Tool objects to include (e.g., server_status)
        
    Returns:
        List of Tool objects
    """
    tools = []
    
    # Add additional tools first (e.g., server_status)
    if additional_tools:
        tools.extend(additional_tools)
    
    # Add tools from registry
    for name, metadata in sorted(_TOOL_REGISTRY.items()):
        tools.append(
            Tool(
                name=metadata["name"],
                description=metadata["description"],
                inputSchema=metadata["inputSchema"],
            )
        )
    
    return tools
