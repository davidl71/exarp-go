#!/usr/bin/env python3
"""
Python Bridge Tool Executor

Executes Python tools from the Go MCP server via subprocess.
This bridge allows the Go server to execute existing Python tools.

Supports both protobuf binary and JSON formats for backward compatibility.
"""

import json
import sys
from pathlib import Path

# Try to import protobuf (required for protobuf support)
try:
    from bridge.proto import bridge_pb2
    HAVE_PROTOBUF = True
except ImportError:
    HAVE_PROTOBUF = False
    # Fall back to JSON-only mode if protobuf not available

# Add local project_management_automation to path (copied to exarp-go)
PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Add bridge directory to path for mcp-generic-tools modules
BRIDGE_ROOT = Path(__file__).parent
sys.path.insert(0, str(BRIDGE_ROOT))

def execute_tool(tool_name: str, args_json: str, use_protobuf: bool = False, protobuf_data: bytes = None, request_id: str = None):
    """Execute a Python tool with given arguments.
    
    Args:
        tool_name: Name of the tool to execute
        args_json: JSON string with arguments (for JSON mode)
        use_protobuf: Whether to use protobuf format
        protobuf_data: Protobuf binary data (if use_protobuf is True)
        request_id: Request ID for tracking (from protobuf request)
    
    Returns:
        If use_protobuf and HAVE_PROTOBUF: Returns protobuf ToolResponse binary
        Otherwise: Returns JSON string
    """
    import time
    start_time = time.time()
    
    try:
        # Parse arguments - try protobuf first if available, fall back to JSON
        args = {}
        if use_protobuf and HAVE_PROTOBUF and protobuf_data:
            try:
                # Parse protobuf ToolRequest
                request = bridge_pb2.ToolRequest()
                request.ParseFromString(protobuf_data)
                
                # Extract tool name, arguments, and request ID from protobuf
                tool_name = request.tool_name
                args_json = request.arguments_json
                request_id = request.request_id if request.request_id else None
                # Parse JSON arguments from protobuf message
                args = json.loads(args_json) if args_json else {}
            except Exception:
                # If protobuf parsing fails, fall back to JSON
                args = json.loads(args_json) if args_json else {}
        else:
            # Use JSON format (backward compatibility)
            args = json.loads(args_json) if args_json else {}
        
        # Bridge only routes mlx; use slim import so we don't load all consolidated tools
        from project_management_automation.tools.consolidated_mlx_only import mlx as _mlx

        # Route to appropriate tool (bridge only handles mlx; Go uses native for all others)
        if tool_name == "mlx":
            result = _mlx(
                action=args.get("action", "status"),
                prompt=args.get("prompt"),
                model=args.get("model", "mlx-community/Phi-3.5-mini-instruct-4bit"),
                max_tokens=args.get("max_tokens", 512),
                temperature=args.get("temperature", 0.7),
                verbose=args.get("verbose", False),
            )
        # task_workflow, context, recommend: native Go only (removed from bridge)
        else:
            execution_time_ms = int((time.time() - start_time) * 1000)
            error_result = {
                "success": False,
                "error": f"Unknown tool: {tool_name}",
                "tool": tool_name
            }
            
            # If protobuf mode, return protobuf ToolResponse
            if use_protobuf and HAVE_PROTOBUF:
                try:
                    response = bridge_pb2.ToolResponse()
                    response.success = False
                    response.error = error_result["error"]
                    response.execution_time_ms = execution_time_ms
                    response.exit_code = 1
                    if request_id:
                        response.request_id = request_id
                    sys.stdout.buffer.write(response.SerializeToString())
                    return 1
                except Exception:
                    # Fall back to JSON on error
                    print(json.dumps(error_result, indent=2))
                    return 1
            
            # Return JSON (backward compatible)
            print(json.dumps(error_result, indent=2))
            return 1
        
        # Calculate execution time
        execution_time_ms = int((time.time() - start_time) * 1000)
        
        # Handle result - tools may return dict or JSON string
        if isinstance(result, dict):
            result_json = json.dumps(result, indent=2)
        elif isinstance(result, str):
            result_json = result
        else:
            result_json = json.dumps({"result": str(result)}, indent=2)
        
        # If protobuf mode and protobuf is available, return protobuf ToolResponse
        if use_protobuf and HAVE_PROTOBUF:
            try:
                # Create protobuf ToolResponse
                response = bridge_pb2.ToolResponse()
                response.success = True
                response.result = result_json
                response.execution_time_ms = execution_time_ms
                response.exit_code = 0
                if request_id:
                    response.request_id = request_id
                
                # Serialize to binary and return
                return response.SerializeToString()
            except Exception as e:
                # If protobuf serialization fails, fall back to JSON
                # Include error in JSON response
                error_result = {
                    "success": False,
                    "error": f"Failed to serialize protobuf response: {str(e)}",
                    "result": result_json
                }
                return json.dumps(error_result, indent=2)
        
        # Return JSON (backward compatible)
        return result_json
        
    except Exception as e:
        # Calculate execution time even on error
        execution_time_ms = int((time.time() - start_time) * 1000)
        
        # If protobuf mode, return protobuf ToolResponse
        if use_protobuf and HAVE_PROTOBUF:
            try:
                response = bridge_pb2.ToolResponse()
                response.success = False
                response.error = str(e)
                response.execution_time_ms = execution_time_ms
                response.exit_code = 1
                if request_id:
                    response.request_id = request_id
                sys.stdout.buffer.write(response.SerializeToString())
                return 1
            except Exception:
                # Fall back to JSON if protobuf fails
                pass
        
        error_result = {
            "success": False,
            "error": str(e),
            "tool": tool_name
        }
        print(json.dumps(error_result, indent=2))
        return 1

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Tool name required"}, indent=2))
        sys.exit(1)
    
    tool_name = sys.argv[1]
    
    # Check if protobuf format is being used (--protobuf flag)
    use_protobuf = False
    protobuf_data = None
    
    if len(sys.argv) > 2 and sys.argv[2] == "--protobuf":
        # Protobuf mode: read binary data from stdin
        if HAVE_PROTOBUF:
            use_protobuf = True
            protobuf_data = sys.stdin.buffer.read()
        else:
            # Protobuf not available, fall back to JSON
            print(json.dumps({"error": "Protobuf support not available"}, indent=2), file=sys.stderr)
            sys.exit(1)
    else:
        # JSON mode: get JSON string from command line
        args_json = sys.argv[2] if len(sys.argv) > 2 else "{}"
    
    # Extract request_id from protobuf request if available
    request_id = None
    if use_protobuf and HAVE_PROTOBUF and protobuf_data:
        try:
            request = bridge_pb2.ToolRequest()
            request.ParseFromString(protobuf_data)
            request_id = request.request_id if request.request_id else None
        except Exception:
            # If parsing fails, continue without request_id
            pass
    
    # Execute tool
    result = execute_tool(tool_name, args_json if not use_protobuf else "", use_protobuf, protobuf_data, request_id)
    
    # If protobuf mode and result is bytes (protobuf binary), write to stdout.buffer
    if use_protobuf and HAVE_PROTOBUF and isinstance(result, bytes):
        sys.stdout.buffer.write(result)
        sys.exit(0)
    
    # Otherwise, result is JSON string or exit code
    if isinstance(result, str):
        print(result)
        sys.exit(0)
    else:
        sys.exit(result)

