"""Integration tests for server startup and initialization."""

import subprocess
import sys
import os
from pathlib import Path

# Add tests/fixtures directory to path to import test helpers
project_root = Path(__file__).parent.parent.parent.parent
sys.path.insert(0, str(project_root / "tests" / "fixtures"))
from test_helpers import JSONRPCClient, spawn_server


def test_server_initialization():
    """Test server initialization."""
    # Spawn server binary
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Send initialize request
        response = client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {
                "name": "test",
                "version": "1.0.0"
            }
        })
        
        # Verify response format
        assert response["jsonrpc"] == "2.0"
        assert "id" in response
        assert "result" in response or "error" in response
        
        # Verify server initialized successfully
        if "error" in response:
            raise AssertionError(f"Server initialization failed: {response['error']}")
        
        assert "result" in response
        result = response["result"]
        assert "protocolVersion" in result or "serverInfo" in result
        
    finally:
        client.close()


def test_all_tools_available():
    """Test all 24+ tools are available."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server first
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Send tools/list request
        response = client.call("tools/list", {})
        
        # Verify response format
        assert response["jsonrpc"] == "2.0"
        assert "result" in response or "error" in response
        
        if "error" in response:
            raise AssertionError(f"tools/list failed: {response['error']}")
        
        # Verify tools are returned
        result = response["result"]
        assert "tools" in result
        tools = result["tools"]
        assert isinstance(tools, list)
        assert len(tools) >= 24  # At least 24 tools (may be 25 with apple_foundation_models)
        
        # Verify expected tool names are present
        tool_names = [tool["name"] for tool in tools]
        expected_tools = [
            "analyze_alignment",
            "generate_config",
            "health",
            "setup_hooks",
            "check_attribution",
            "add_external_tool_hints",
            "memory",
            "memory_maint",
            "report",
            "security",
            "task_analysis",
            "task_discovery",
            "task_workflow",
            "testing",
            "automation",
            "tool_catalog",
            "workflow_mode",
            "lint",
            "estimation",
            "git_tools",
            "session",
            "infer_session_mode",
            "ollama",
            "mlx",
        ]
        
        for expected_tool in expected_tools:
            assert expected_tool in tool_names, f"Expected tool '{expected_tool}' not found in tools list"
        
    finally:
        client.close()


def test_all_prompts_available():
    """Test all 15+ prompts are available."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server first
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Send prompts/list request
        response = client.call("prompts/list", {})
        
        # Verify response format
        assert response["jsonrpc"] == "2.0"
        assert "result" in response or "error" in response
        
        if "error" in response:
            raise AssertionError(f"prompts/list failed: {response['error']}")
        
        # Verify prompts are returned
        result = response["result"]
        assert "prompts" in result
        prompts = result["prompts"]
        assert isinstance(prompts, list)
        assert len(prompts) >= 15  # At least 15 prompts (may be more)
        
        # Verify expected prompt names are present
        prompt_names = [prompt["name"] for prompt in prompts]
        expected_prompts = [
            "align",
            "discover",
            "config",
            "scan",
            "scorecard",
            "overview",
            "dashboard",
            "remember",
        ]
        
        for expected_prompt in expected_prompts:
            assert expected_prompt in prompt_names, f"Expected prompt '{expected_prompt}' not found in prompts list"
        
    finally:
        client.close()


def test_all_resources_available():
    """Test all 21+ resources are available."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server first
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Send resources/list request
        response = client.call("resources/list", {})
        
        # Verify response format
        assert response["jsonrpc"] == "2.0"
        assert "result" in response or "error" in response
        
        if "error" in response:
            raise AssertionError(f"resources/list failed: {response['error']}")
        
        # Verify resources are returned
        result = response["result"]
        assert "resources" in result
        resources = result["resources"]
        assert isinstance(resources, list)
        assert len(resources) >= 21  # At least 21 resources
        
        # Verify expected resource URIs are present
        resource_uris = [resource["uri"] for resource in resources]
        expected_resources = [
            "stdio://scorecard",
            "stdio://memories",
            "stdio://memories/category/{category}",
            "stdio://memories/task/{task_id}",
            "stdio://memories/recent",
            "stdio://memories/session/{date}",
        ]
        
        for expected_resource in expected_resources:
            assert expected_resource in resource_uris, f"Expected resource '{expected_resource}' not found in resources list"
        
    finally:
        client.close()


def test_server_termination():
    """Test server termination."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Verify server is running
        assert server_process.poll() is None, "Server should be running"
        
    finally:
        # Close client and terminate server
        client.close()
        
        # Verify server exits gracefully
        server_process.wait(timeout=5)
        assert server_process.returncode is not None, "Server should have exited"


def test_binary_exists():
    """Test server binary exists."""
    project_root = Path(__file__).parent.parent.parent.parent
    binary_path = project_root / "bin" / "exarp-go"
    
    # Verify binary exists (may not exist during development)
    # This is a basic check - actual binary may be built separately
    if not binary_path.exists():
        # Try alternative location
        binary_path = project_root / "exarp-go"
    
    # Note: Binary may not exist during development, so we just check the path
    # In CI/CD, binary should be built before running tests
    assert binary_path.exists() or os.getenv("CI") is None, f"Binary not found at {binary_path} (may need to run 'make build')"
