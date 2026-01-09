#!/usr/bin/env python3
"""
Python Bridge Resource Executor

Executes Python resource handlers from the Go MCP server via subprocess.
This bridge allows the Go server to access existing Python resources.
"""

import json
import sys
import os
from pathlib import Path

# Add local project_management_automation to path (copied to exarp-go)
PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

def execute_resource(uri: str):
    """Execute a Python resource handler with given URI."""
    try:
        # Import resource handlers
        from project_management_automation.resources.memories import (
            get_memories_resource,
            get_memories_by_category_resource,
            get_memories_by_task_resource,
            get_recent_memories_resource,
            get_session_memories_resource,
        )
        from project_management_automation.resources.prompt_discovery import (
            get_all_prompts_compact,
            get_prompts_for_mode,
            get_prompts_for_persona,
            get_prompts_for_category,
        )
        from project_management_automation.resources.session import (
            get_session_mode_resource,
        )
        from project_management_automation.tools.project_scorecard import (
            generate_project_scorecard,
        )
        
        # Route to appropriate resource
        if uri == "stdio://scorecard":
            result = generate_project_scorecard("json", True, None)
        elif uri == "stdio://memories":
            result = get_memories_resource()
        elif uri.startswith("stdio://memories/category/"):
            category = uri.split("/")[-1]
            result = get_memories_by_category_resource(category)
        elif uri.startswith("stdio://memories/task/"):
            task_id = uri.split("/")[-1]
            result = get_memories_by_task_resource(task_id)
        elif uri == "stdio://memories/recent":
            result = get_recent_memories_resource()
        elif uri.startswith("stdio://memories/session/"):
            date = uri.split("/")[-1]
            result = get_session_memories_resource(date)
        elif uri == "stdio://prompts":
            result = get_all_prompts_compact()
        elif uri.startswith("stdio://prompts/mode/"):
            mode = uri.split("/")[-1]
            result = get_prompts_for_mode(mode)
        elif uri.startswith("stdio://prompts/persona/"):
            persona = uri.split("/")[-1]
            result = get_prompts_for_persona(persona)
        elif uri.startswith("stdio://prompts/category/"):
            category = uri.split("/")[-1]
            result = get_prompts_for_category(category)
        elif uri == "stdio://session/mode":
            result = get_session_mode_resource()
        else:
            error_result = {
                "success": False,
                "error": f"Unknown resource: {uri}",
                "uri": uri
            }
            print(json.dumps(error_result, indent=2))
            return 1
        
        # Handle result - resources may return dict or JSON string
        if isinstance(result, dict):
            result_json = json.dumps(result, indent=2)
        elif isinstance(result, str):
            result_json = result
        else:
            result_json = json.dumps({"result": str(result)}, indent=2)
        
        print(result_json)
        return 0
        
    except Exception as e:
        error_result = {
            "success": False,
            "error": str(e),
            "uri": uri
        }
        print(json.dumps(error_result, indent=2))
        return 1

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Resource URI required"}, indent=2))
        sys.exit(1)
    
    uri = sys.argv[1]
    
    sys.exit(execute_resource(uri))

