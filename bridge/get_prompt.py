#!/usr/bin/env python3
"""
Python Bridge Prompt Retriever

Retrieves prompt templates from the Go MCP server via subprocess.
This bridge allows the Go server to access existing Python prompts.
"""

import json
import sys
import os
from pathlib import Path

# Add parent project to path
PROJECT_ROOT = Path(__file__).parent.parent.parent / "project-management-automation"
sys.path.insert(0, str(PROJECT_ROOT))

def get_prompt(prompt_name: str):
    """Get a prompt template by name."""
    try:
        # Import prompts
        try:
            from project_management_automation.prompts import (
                TASK_ALIGNMENT_ANALYSIS,
                TASK_DISCOVERY,
                CONFIG_GENERATION,
                SECURITY_SCAN_ALL,
                PROJECT_SCORECARD,
                PROJECT_OVERVIEW,
                PROJECT_DASHBOARD,
                MEMORY_SYSTEM,
                # High-value workflow prompts
                DAILY_CHECKIN,
                SPRINT_START,
                SPRINT_END,
                PRE_SPRINT_CLEANUP,
                POST_IMPLEMENTATION_REVIEW,
                TASK_SYNC,
                DUPLICATE_TASK_CLEANUP,
            )
        except ImportError:
            # Fallback to absolute import
            sys.path.insert(0, str(PROJECT_ROOT))
            from prompts import (
                TASK_ALIGNMENT_ANALYSIS,
                TASK_DISCOVERY,
                CONFIG_GENERATION,
                SECURITY_SCAN_ALL,
                PROJECT_SCORECARD,
                PROJECT_OVERVIEW,
                PROJECT_DASHBOARD,
                MEMORY_SYSTEM,
                # High-value workflow prompts
                DAILY_CHECKIN,
                SPRINT_START,
                SPRINT_END,
                PRE_SPRINT_CLEANUP,
                POST_IMPLEMENTATION_REVIEW,
                TASK_SYNC,
                DUPLICATE_TASK_CLEANUP,
            )
        
        # Map prompt names to templates
        prompt_map = {
            "align": TASK_ALIGNMENT_ANALYSIS,
            "discover": TASK_DISCOVERY,
            "config": CONFIG_GENERATION,
            "scan": SECURITY_SCAN_ALL,
            "scorecard": PROJECT_SCORECARD,
            "overview": PROJECT_OVERVIEW,
            "dashboard": PROJECT_DASHBOARD,
            "remember": MEMORY_SYSTEM,
            # High-value workflow prompts
            "daily_checkin": DAILY_CHECKIN,
            "sprint_start": SPRINT_START,
            "sprint_end": SPRINT_END,
            "pre_sprint": PRE_SPRINT_CLEANUP,
            "post_impl": POST_IMPLEMENTATION_REVIEW,
            "sync": TASK_SYNC,
            "dups": DUPLICATE_TASK_CLEANUP,
        }
        
        prompt_text = prompt_map.get(prompt_name)
        if prompt_text is None:
            result = {
                "success": False,
                "error": f"Unknown prompt: {prompt_name}",
                "prompt": prompt_name
            }
            print(json.dumps(result, indent=2))
            return 1
        
        result = {
            "success": True,
            "prompt": prompt_text
        }
        print(json.dumps(result, indent=2))
        return 0
        
    except Exception as e:
        result = {
            "success": False,
            "error": str(e),
            "prompt": prompt_name
        }
        print(json.dumps(result, indent=2))
        return 1

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Prompt name required"}, indent=2))
        sys.exit(1)
    
    prompt_name = sys.argv[1]
    
    sys.exit(get_prompt(prompt_name))
