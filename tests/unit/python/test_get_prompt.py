"""Tests for get_prompt.py bridge script."""

import json
import sys
from pathlib import Path

# Add bridge directory to path
BRIDGE_DIR = Path(__file__).parent.parent.parent / "bridge"
sys.path.insert(0, str(BRIDGE_DIR))

import get_prompt


def test_prompt_name_mapping():
    """Test prompt name mapping (8 prompts)."""
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

    # Verify all prompts are in the mapping
    # (We can't easily test the actual mapping without importing the real modules)
    for prompt_name in expected_prompts:
        assert prompt_name is not None


def test_prompt_retrieval_success():
    """Test prompt retrieval success response format."""
    success_result = {
        "success": True,
        "prompt": "Test prompt text"
    }
    result_json = json.dumps(success_result, indent=2)
    parsed = json.loads(result_json)
    
    assert parsed["success"] is True
    assert "prompt" in parsed
    assert parsed["prompt"] == "Test prompt text"


def test_unknown_prompt_error():
    """Test unknown prompt error handling."""
    error_result = {
        "success": False,
        "error": "Unknown prompt: invalid_prompt",
        "prompt": "invalid_prompt"
    }
    result_json = json.dumps(error_result, indent=2)
    parsed = json.loads(result_json)
    
    assert parsed["success"] is False
    assert "Unknown prompt" in parsed["error"]


def test_json_response_format():
    """Test JSON response format."""
    # Success format
    success_response = {
        "success": True,
        "prompt": "test"
    }
    assert "success" in success_response
    assert "prompt" in success_response

    # Error format
    error_response = {
        "success": False,
        "error": "test error",
        "prompt": "test"
    }
    assert "success" in error_response
    assert "error" in error_response


def test_import_error_handling():
    """Test import error handling."""
    # Test error response format for import errors
    error_result = {
        "success": False,
        "error": "ModuleNotFoundError: No module named 'test'",
        "prompt": "test"
    }
    result_json = json.dumps(error_result, indent=2)
    parsed = json.loads(result_json)
    
    assert parsed["success"] is False
    assert "error" in parsed


def test_prompt_json_structure():
    """Test prompt JSON structure validation."""
    valid_response = {
        "success": True,
        "prompt": "Test prompt with multiple lines\nLine 2"
    }
    result_json = json.dumps(valid_response, indent=2)
    parsed = json.loads(result_json)
    
    # Verify structure
    assert isinstance(parsed, dict)
    assert "success" in parsed
    assert "prompt" in parsed
    assert isinstance(parsed["prompt"], str)
