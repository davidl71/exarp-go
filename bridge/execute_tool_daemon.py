#!/usr/bin/env python3
"""
Python Bridge Tool Executor Daemon

Persistent process that executes Python tools via JSON-RPC over stdin/stdout.
This daemon keeps models and caches in memory between requests, significantly
reducing overhead compared to spawning new processes for each call.

Protocol: JSON-RPC 2.0 over line-delimited stdin/stdout
"""

import json
import sys

# Import the existing execute_tool function
# This allows us to reuse all existing tool logic
from execute_tool import execute_tool

def handle_jsonrpc_request(request: dict) -> dict:
    """
    Handle a JSON-RPC 2.0 request.
    
    Args:
        request: JSON-RPC request object
        
    Returns:
        JSON-RPC response object
    """
    # Extract request ID (required for responses)
    request_id = request.get("id")
    
    # Verify JSON-RPC version
    if request.get("jsonrpc") != "2.0":
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": -32600,
                "message": "Invalid Request",
                "data": "jsonrpc must be '2.0'"
            }
        }
    
    # Extract method and params
    method = request.get("method")
    params = request.get("params", {})
    
    if method != "execute_tool":
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": -32601,
                "message": "Method Not Found",
                "data": f"Unknown method: {method}"
            }
        }
    
    # Extract tool_name and args from params
    tool_name = params.get("tool_name")
    args = params.get("args", {})
    
    if not tool_name:
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": -32602,
                "message": "Invalid Params",
                "data": "tool_name is required"
            }
        }
    
    # Execute tool using existing function
    # Convert args dict to JSON string (as expected by execute_tool)
    args_json = json.dumps(args) if args else "{}"
    
    try:
        # Call existing execute_tool function
        # Use JSON mode (not protobuf) for simplicity
        result = execute_tool(tool_name, args_json, use_protobuf=False, protobuf_data=None, request_id=None)
        
        # execute_tool returns a JSON string (from execute_tool.py line 431)
        # We need to return it as-is in the JSON-RPC result
        # The result is already a JSON string, so we'll return it as a string value
        # The Go code will parse it as needed
        
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "result": result  # Return the JSON string directly
        }
    except Exception as e:
        # Handle execution errors
        return {
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": -32000,
                "message": "Tool Execution Error",
                "data": str(e)
            }
        }

def main():
    """
    Main daemon loop - reads JSON-RPC requests from stdin, executes, writes to stdout.
    
    Protocol:
    - Each request is a JSON object on a single line (newline-delimited)
    - Each response is a JSON object on a single line (newline-delimited)
    - Process stays alive between requests
    """
    # Ensure stdout is unbuffered for line-by-line output
    sys.stdout.reconfigure(line_buffering=True)
    sys.stderr.reconfigure(line_buffering=True)
    
    # Read requests from stdin line by line
    for line in sys.stdin:
        # Skip empty lines
        line = line.strip()
        if not line:
            continue
        
        try:
            # Parse JSON-RPC request
            request = json.loads(line)
            
            # Handle request
            response = handle_jsonrpc_request(request)
            
            # Write response to stdout (single line, no indentation)
            response_json = json.dumps(response, separators=(',', ':'))
            print(response_json, flush=True)
            
        except json.JSONDecodeError as e:
            # Invalid JSON - send parse error
            error_response = {
                "jsonrpc": "2.0",
                "id": None,
                "error": {
                    "code": -32700,
                    "message": "Parse error",
                    "data": str(e)
                }
            }
            print(json.dumps(error_response, separators=(',', ':')), flush=True)
            
        except Exception as e:
            # Unexpected error - send internal error
            error_response = {
                "jsonrpc": "2.0",
                "id": None,
                "error": {
                    "code": -32603,
                    "message": "Internal error",
                    "data": str(e)
                }
            }
            print(json.dumps(error_response, separators=(',', ':')), flush=True)

if __name__ == "__main__":
    main()
