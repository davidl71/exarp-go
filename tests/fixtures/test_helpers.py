"""Test helpers for Python tests."""

import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Optional, Dict, Any


class JSONRPCClient:
    """JSON-RPC 2.0 client for stdio testing."""
    
    def __init__(self, server_process: subprocess.Popen):
        """Initialize client with server process."""
        self.process = server_process
        self.request_id = 1
    
    def call(self, method: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Send a JSON-RPC request and read response."""
        request = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params or {}
        }
        self.request_id += 1
        
        # Write request
        request_json = json.dumps(request) + "\n"
        self.process.stdin.write(request_json.encode())
        self.process.stdin.flush()
        
        # Read response (may need to skip notifications)
        while True:
            response_line = self.process.stdout.readline()
            if not response_line:
                raise RuntimeError("No response from server")
            
            response = json.loads(response_line.decode())
            
            # Skip notifications (they don't have 'id' field)
            if "id" in response:
                return response
            
            # If it's a notification, continue reading
            if "method" in response and response.get("method", "").startswith("notifications/"):
                continue
            
            # If no 'id' and not a notification, return anyway (may be error)
            return response
    
    def close(self):
        """Close the client."""
        if self.process.stdin:
            self.process.stdin.close()
        if self.process.stdout:
            self.process.stdout.close()
        if self.process:
            self.process.terminate()
            self.process.wait()


def spawn_server(server_path: Optional[str] = None, project_root: Optional[str] = None) -> subprocess.Popen:
    """Spawn the MCP server binary."""
    if server_path is None:
        # Default to bin/exarp-go
        exarp_root = Path(__file__).parent.parent.parent
        server_path = exarp_root / "bin" / "exarp-go"
    
    env = os.environ.copy()
    if project_root is None:
        # Default to exarp-go project root for testing
        project_root = str(Path(__file__).parent.parent.parent)
    env["PROJECT_ROOT"] = project_root
    
    return subprocess.Popen(
        [str(server_path)],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=env,
        cwd=Path(__file__).parent.parent.parent
    )


def generate_test_args(**kwargs):
    """Generate test arguments dict."""
    return kwargs


def parse_json_response(response_str: str) -> Dict[str, Any]:
    """Parse JSON response string."""
    return json.loads(response_str)


def assert_jsonrpc_response(response: Dict[str, Any], expected_id: int = None):
    """Assert response is valid JSON-RPC 2.0."""
    assert "jsonrpc" in response
    assert response["jsonrpc"] == "2.0"
    assert "id" in response
    if expected_id is not None:
        assert response["id"] == expected_id
    assert "result" in response or "error" in response
    return response
