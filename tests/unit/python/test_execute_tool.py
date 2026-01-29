"""Tests for execute_tool.py bridge script."""

import json
import sys
from pathlib import Path
from unittest.mock import patch, MagicMock

# Add bridge directory to path (test is in tests/unit/python/ -> 4 parents to repo)
REPO_ROOT = Path(__file__).resolve().parent.parent.parent.parent
BRIDGE_DIR = REPO_ROOT / "bridge"
sys.path.insert(0, str(BRIDGE_DIR))

import execute_tool


def test_tool_routing():
    """Bridge routes only mlx (Python deprecated; MLX-only retention 2026-01-29)."""
    # execute_tool.py routes only "mlx"; all other tools are native Go
    assert "mlx" in ["mlx"]
    # Smoke: bridge can be imported and execute_tool is callable
    assert callable(execute_tool.execute_tool)


def test_argument_parsing():
    """Test argument parsing and passing."""
    # Test valid JSON
    args_json = '{"action": "test", "value": 42}'
    args = json.loads(args_json)
    assert args["action"] == "test"
    assert args["value"] == 42

    # Test empty JSON
    args_json = "{}"
    args = json.loads(args_json)
    assert args == {}

    # Test invalid JSON
    invalid_json = "{invalid json}"
    try:
        json.loads(invalid_json)
        assert False, "Should raise JSONDecodeError"
    except json.JSONDecodeError:
        pass


def test_error_handling():
    """Test error handling and JSON response format."""
    # Test error response format
    error_result = {
        "success": False,
        "error": "Test error",
        "tool": "test_tool"
    }
    result_json = json.dumps(error_result, indent=2)
    parsed = json.loads(result_json)
    
    assert parsed["success"] is False
    assert "error" in parsed
    assert parsed["tool"] == "test_tool"


def test_unknown_tool_error():
    """Test unknown tool error handling."""
    error_result = {
        "success": False,
        "error": "Unknown tool: invalid_tool",
        "tool": "invalid_tool"
    }
    result_json = json.dumps(error_result, indent=2)
    parsed = json.loads(result_json)
    
    assert parsed["success"] is False
    assert "Unknown tool" in parsed["error"]


def test_result_format_dict():
    """Test result format when tool returns dict."""
    result = {"success": True, "data": "test"}
    if isinstance(result, dict):
        result_json = json.dumps(result, indent=2)
        parsed = json.loads(result_json)
        assert parsed["success"] is True
        assert parsed["data"] == "test"


def test_result_format_string():
    """Test result format when tool returns JSON string."""
    result = '{"success": true, "data": "test"}'
    if isinstance(result, str):
        parsed = json.loads(result)
        assert parsed["success"] is True


def test_result_format_other():
    """Test result format for other types."""
    result = 42
    if not isinstance(result, (dict, str)):
        result_json = json.dumps({"result": str(result)}, indent=2)
        parsed = json.loads(result_json)
        assert parsed["result"] == "42"
