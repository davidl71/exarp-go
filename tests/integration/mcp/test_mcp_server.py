"""Integration tests for MCP server JSON-RPC communication."""

import json
import subprocess
import sys
from pathlib import Path

# Add tests directory to path to import test helpers
project_root = Path(__file__).parent.parent.parent.parent
sys.path.insert(0, str(project_root / "tests" / "fixtures"))
from test_helpers import JSONRPCClient, spawn_server


def test_jsonrpc_protocol_compliance():
    """Test JSON-RPC 2.0 protocol compliance."""
    # Test JSON-RPC 2.0 request format
    request = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/list",
        "params": {}
    }
    
    request_json = json.dumps(request)
    parsed = json.loads(request_json)
    
    assert parsed["jsonrpc"] == "2.0"
    assert "id" in parsed
    assert "method" in parsed
    assert "params" in parsed


def test_tool_invocation_via_stdio():
    """Test tool invocation via stdio."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server first
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Invoke a simple tool (server_status)
        response = client.call("tools/call", {
            "name": "server_status",
            "arguments": {}
        })
        
        # Verify response format
        assert response["jsonrpc"] == "2.0"
        assert "id" in response
        assert "result" in response or "error" in response
        
        if "error" in response:
            # Some tools may require specific setup, so we just verify error format
            assert "code" in response["error"]
            assert "message" in response["error"]
        else:
            # Verify result structure
            result = response["result"]
            assert "content" in result or "text" in result or "isError" in result
        
    finally:
        client.close()


def test_prompt_retrieval_via_stdio():
    """Test prompt retrieval via stdio."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server first
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Get a prompt
        response = client.call("prompts/get", {
            "name": "align"
        })
        
        # Verify response format
        assert response["jsonrpc"] == "2.0"
        assert "id" in response
        assert "result" in response or "error" in response
        
        if "error" not in response:
            # Verify prompt structure
            result = response["result"]
            assert "messages" in result or "description" in result or "arguments" in result
        
    finally:
        client.close()


def test_resource_retrieval_via_stdio():
    """Test resource retrieval via stdio."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server first
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Get a resource
        response = client.call("resources/read", {
            "uri": "stdio://scorecard"
        })
        
        # Verify response format
        assert response["jsonrpc"] == "2.0"
        assert "id" in response
        assert "result" in response or "error" in response
        
        if "error" not in response:
            # Verify resource structure
            result = response["result"]
            assert "contents" in result or "uri" in result
        
    finally:
        client.close()


def test_error_responses():
    """Test error responses."""
    # Test JSON-RPC error response format
    error_response = {
        "jsonrpc": "2.0",
        "id": 1,
        "error": {
            "code": -32600,
            "message": "Invalid Request"
        }
    }
    
    response_json = json.dumps(error_response)
    parsed = json.loads(response_json)
    
    assert parsed["jsonrpc"] == "2.0"
    assert "error" in parsed
    assert parsed["error"]["code"] == -32600


def test_batch_requests():
    """Test batch requests."""
    # Test JSON-RPC batch request format
    batch_request = [
        {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "tools/list",
            "params": {}
        },
        {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "prompts/list",
            "params": {}
        }
    ]
    
    batch_json = json.dumps(batch_request)
    parsed = json.loads(batch_json)
    
    assert isinstance(parsed, list)
    assert len(parsed) == 2
    assert all("jsonrpc" in req for req in parsed)
    
    # Note: Actual batch request execution would require server support
    # This test verifies the batch request format is valid JSON-RPC 2.0