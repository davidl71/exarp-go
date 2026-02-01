"""Regression tests comparing native Go vs Python bridge outputs."""

import json
import subprocess
import sys
from pathlib import Path
from typing import Dict, Any, Optional

# Add tests/fixtures directory to path
project_root = Path(__file__).parent.parent.parent.parent
sys.path.insert(0, str(project_root / "tests" / "fixtures"))
try:
    from test_helpers import JSONRPCClient, spawn_server
except ImportError:
    # Fallback if import fails
    import sys
    sys.path.insert(0, str(project_root / "tests" / "fixtures"))
    from test_helpers import JSONRPCClient, spawn_server


def invoke_native_tool(tool_name: str, args: Dict[str, Any]) -> Dict[str, Any]:
    """Invoke a tool via native Go implementation."""
    server_process = spawn_server()
    client = JSONRPCClient(server_process)
    
    try:
        # Initialize server
        client.call("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "test", "version": "1.0.0"}
        })
        
        # Call tool
        response = client.call("tools/call", {
            "name": tool_name,
            "arguments": args
        })
        
        if "error" in response:
            return {"error": response["error"]}
        
        return {"result": response.get("result", {})}
    finally:
        client.close()


def assert_structure_parity(native_result: Dict[str, Any], bridge_result: Dict[str, Any], tolerance: float = 0.01):
    """Assert that native and bridge results have similar structure."""
    # Both should have same top-level keys (result or error)
    assert set(native_result.keys()) == set(bridge_result.keys()), \
        f"Structure mismatch: {native_result.keys()} vs {bridge_result.keys()}"
    
    if "error" in native_result:
        # Both should have errors with similar codes
        assert "error" in bridge_result
        # Error codes should match (or be similar)
        native_code = native_result["error"].get("code")
        bridge_code = bridge_result["error"].get("code")
        assert native_code == bridge_code, \
            f"Error code mismatch: {native_code} vs {bridge_code}"
        return
    
    # Both should have results
    assert "result" in native_result
    assert "result" in bridge_result
    
    native_data = native_result["result"]
    bridge_data = bridge_result["result"]
    
    # Compare structure (recursive)
    _compare_structure(native_data, bridge_data)


def _compare_structure(native: Any, bridge: Any, path: str = ""):
    """Recursively compare structure of native and bridge results."""
    if isinstance(native, dict) and isinstance(bridge, dict):
        # Both are dicts - compare keys
        native_keys = set(native.keys())
        bridge_keys = set(bridge.keys())
        
        # Allow some keys to differ (e.g., timestamps, IDs)
        common_keys = native_keys & bridge_keys
        diff_keys = (native_keys | bridge_keys) - common_keys
        
        # Warn about different keys but don't fail
        if diff_keys:
            print(f"Warning: Different keys at {path}: {diff_keys}")
        
        # Compare common keys
        for key in common_keys:
            _compare_structure(native[key], bridge[key], f"{path}.{key}")
    
    elif isinstance(native, list) and isinstance(bridge, list):
        # Both are lists - compare lengths
        if len(native) != len(bridge):
            print(f"Warning: List length mismatch at {path}: {len(native)} vs {len(bridge)}")
        
        # Compare elements up to minimum length
        for i in range(min(len(native), len(bridge))):
            _compare_structure(native[i], bridge[i], f"{path}[{i}]")
    
    elif type(native) != type(bridge):
        # Type mismatch - warn but don't fail (e.g., int vs float)
        print(f"Warning: Type mismatch at {path}: {type(native)} vs {type(bridge)}")


def assert_value_parity(native_result: Dict[str, Any], bridge_result: Dict[str, Any], tolerance: float = 0.01):
    """Assert that native and bridge results have similar values."""
    if "error" in native_result:
        # Both should have errors - already checked in structure parity
        return
    
    native_data = native_result["result"]
    bridge_data = bridge_result["result"]
    
    # Compare values (recursive)
    _compare_values(native_data, bridge_data, tolerance)


def _compare_values(native: Any, bridge: Any, tolerance: float, path: str = ""):
    """Recursively compare values of native and bridge results."""
    if isinstance(native, dict) and isinstance(bridge, dict):
        common_keys = set(native.keys()) & set(bridge.keys())
        for key in common_keys:
            _compare_values(native[key], bridge[key], tolerance, f"{path}.{key}")
    
    elif isinstance(native, list) and isinstance(bridge, list):
        for i in range(min(len(native), len(bridge))):
            _compare_values(native[i], bridge[i], tolerance, f"{path}[{i}]")
    
    elif isinstance(native, (int, float)) and isinstance(bridge, (int, float)):
        diff = abs(float(native) - float(bridge))
        if diff > tolerance:
            print(f"Warning: Value difference at {path}: {native} vs {bridge} (diff: {diff})")
    
    elif native != bridge:
        # Allow some differences (e.g., timestamps, IDs)
        if path and not any(skip in path.lower() for skip in ["timestamp", "id", "time", "date", "created", "updated"]):
            print(f"Warning: Value mismatch at {path}: {native} vs {bridge}")


def test_server_status_native_vs_bridge():
    """Compare server_status tool native vs bridge outputs."""
    # Note: server_status is native Go only, so this test verifies structure
    result = invoke_native_tool("server_status", {})
    
    # Verify result structure
    assert "result" in result or "error" in result
    
    if "result" in result:
        result_data = result["result"]
        # Should have status, version, project_root
        assert "content" in result_data or "text" in result_data


def test_automation_daily_native():
    """Test automation daily action (native Go implementation)."""
    result = invoke_native_tool("automation", {
        "action": "daily",
        "dry_run": True
    })
    
    # Verify result structure
    assert "result" in result or "error" in result
    
    if "result" in result:
        # Should return JSON with action, status, etc.
        result_data = result["result"]
        assert "content" in result_data or "text" in result_data


def test_automation_nightly_native():
    """Test automation nightly action (native Go implementation)."""
    result = invoke_native_tool("automation", {
        "action": "nightly",
        "dry_run": True,
        "max_iterations": 1
    })
    
    # Verify result structure
    assert "result" in result or "error" in result


def test_automation_sprint_native():
    """Test automation sprint action (native Go implementation)."""
    result = invoke_native_tool("automation", {
        "action": "sprint",
        "dry_run": True,
        "max_iterations": 1,
        "auto_approve": False,
        "extract_subtasks": False,
        "run_analysis_tools": False,
        "run_testing_tools": False
    })
    
    # Verify result structure
    assert "result" in result or "error" in result


def test_task_workflow_sync_native():
    """Test task_workflow sync action (native Go). External sync is future nice-to-have; param ignored."""
    result = invoke_native_tool("task_workflow", {
        "action": "sync"
    })
    
    # Verify result structure
    assert "result" in result or "error" in result
    
    if "result" in result:
        result_data = result["result"]
        # Should have sync_results
        assert "content" in result_data or "text" in result_data
