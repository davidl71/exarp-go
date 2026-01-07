"""Mock Python responses for testing bridge scripts."""

import json


class MockPythonTool:
    """Mock Python tool responses."""
    
    @staticmethod
    def success_result(data):
        """Create a successful tool response."""
        if isinstance(data, dict):
            return json.dumps(data, indent=2)
        elif isinstance(data, str):
            return data
        else:
            return json.dumps({"result": str(data)}, indent=2)
    
    @staticmethod
    def error_result(error_msg, tool_name="unknown"):
        """Create an error tool response."""
        return json.dumps({
            "success": False,
            "error": error_msg,
            "tool": tool_name
        }, indent=2)


class MockPythonPrompt:
    """Mock Python prompt responses."""
    
    @staticmethod
    def success_result(prompt_text):
        """Create a successful prompt response."""
        return json.dumps({
            "success": True,
            "prompt": prompt_text
        }, indent=2)
    
    @staticmethod
    def error_result(error_msg, prompt_name="unknown"):
        """Create an error prompt response."""
        return json.dumps({
            "success": False,
            "error": error_msg,
            "prompt": prompt_name
        }, indent=2)


class MockPythonResource:
    """Mock Python resource responses."""
    
    @staticmethod
    def success_result(data):
        """Create a successful resource response."""
        if isinstance(data, dict):
            return json.dumps(data, indent=2)
        elif isinstance(data, str):
            return data
        else:
            return json.dumps({"result": str(data)}, indent=2)
    
    @staticmethod
    def error_result(error_msg, uri="unknown"):
        """Create an error resource response."""
        return json.dumps({
            "success": False,
            "error": error_msg,
            "uri": uri
        }, indent=2)


def get_test_prompt(name):
    """Get a test prompt template."""
    prompts = {
        "align": "Test alignment prompt",
        "discover": "Test discovery prompt",
        "config": "Test config prompt",
        "scan": "Test scan prompt",
        "scorecard": "Test scorecard prompt",
        "overview": "Test overview prompt",
        "dashboard": "Test dashboard prompt",
        "remember": "Test remember prompt",
    }
    return prompts.get(name, f"Test prompt for {name}")


def get_test_tool_result(tool_name, args):
    """Get a mock tool result."""
    return {
        "success": True,
        "tool": tool_name,
        "args": args,
        "result": f"Mock result for {tool_name}"
    }


def get_test_resource_result(uri):
    """Get a mock resource result."""
    return {
        "success": True,
        "uri": uri,
        "data": f"Mock data for {uri}"
    }
