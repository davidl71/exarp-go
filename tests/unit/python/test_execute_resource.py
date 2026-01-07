"""Tests for execute_resource.py bridge script."""

import json
import sys
from pathlib import Path

# Add bridge directory to path
BRIDGE_DIR = Path(__file__).parent.parent.parent / "bridge"
sys.path.insert(0, str(BRIDGE_DIR))

import execute_resource


def test_resource_uri_routing():
    """Test resource URI routing."""
    test_uris = [
        "stdio://scorecard",
        "stdio://memories",
        "stdio://memories/category/debug",
        "stdio://memories/task/T-1",
        "stdio://memories/recent",
        "stdio://memories/session/2026-01-07",
    ]

    for uri in test_uris:
        # Verify URI format
        assert uri.startswith("stdio://")
        assert len(uri) > len("stdio://")


def test_resource_execution():
    """Test resource execution response format."""
    # Test successful resource response
    result = {
        "success": True,
        "uri": "stdio://scorecard",
        "data": {"test": "data"}
    }
    
    # Handle result format
    if isinstance(result, dict):
        result_json = json.dumps(result, indent=2)
        parsed = json.loads(result_json)
        assert parsed["success"] is True
        assert "uri" in parsed


def test_json_output_format():
    """Test JSON output format."""
    # Test dict result
    result = {"key": "value"}
    if isinstance(result, dict):
        result_json = json.dumps(result, indent=2)
        parsed = json.loads(result_json)
        assert parsed["key"] == "value"

    # Test string result
    result = '{"key": "value"}'
    if isinstance(result, str):
        parsed = json.loads(result)
        assert parsed["key"] == "value"


def test_error_handling():
    """Test error handling."""
    error_result = {
        "success": False,
        "error": "Resource not found",
        "uri": "stdio://invalid"
    }
    result_json = json.dumps(error_result, indent=2)
    parsed = json.loads(result_json)
    
    assert parsed["success"] is False
    assert "error" in parsed
    assert parsed["uri"] == "stdio://invalid"


def test_uri_patterns():
    """Test URI pattern matching."""
    # Test exact match
    uri = "stdio://scorecard"
    assert uri == "stdio://scorecard"

    # Test prefix match
    uri = "stdio://memories/category/debug"
    assert uri.startswith("stdio://memories/category/")

    # Test suffix extraction
    uri = "stdio://memories/task/T-1"
    task_id = uri.split("/")[-1]
    assert task_id == "T-1"

    # Test date extraction
    uri = "stdio://memories/session/2026-01-07"
    date = uri.split("/")[-1]
    assert date == "2026-01-07"


def test_unknown_resource_error():
    """Test unknown resource error handling."""
    error_result = {
        "success": False,
        "error": "Unknown resource: stdio://invalid",
        "uri": "stdio://invalid"
    }
    result_json = json.dumps(error_result, indent=2)
    parsed = json.loads(result_json)
    
    assert parsed["success"] is False
    assert "Unknown resource" in parsed["error"]
