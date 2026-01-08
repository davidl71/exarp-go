#!/usr/bin/env python3
"""
Set dependencies for migration plan tasks
"""

import sys
from pathlib import Path

# Add project-management-automation to path
project_management_automation = Path("/Users/davidl/Projects/project-management-automation")
sys.path.insert(0, str(project_management_automation))

from project_management_automation.utils.todo2_mcp_client import update_todos_mcp

# Task IDs created
tasks = {
    "T-58": "Phase 1: Analyze and inventory remaining tools to migrate",
    "T-59": "Phase 2: Design migration strategy and architecture",
    "T-60": "Phase 3: Migrate high-priority tools (batch 1)",
    "T-61": "Phase 4: Migrate remaining tools (batch 2)",
    "T-62": "Phase 5: Migrate prompts and resources",
    "T-63": "Phase 6: Testing, validation, and cleanup"
}

# Dependencies: each phase depends on the previous one
dependencies = {
    "T-59": ["T-58"],  # Phase 2 depends on Phase 1
    "T-60": ["T-59"],  # Phase 3 depends on Phase 2
    "T-61": ["T-60"],  # Phase 4 depends on Phase 3
    "T-62": ["T-61"],  # Phase 5 depends on Phase 4
    "T-63": ["T-62"],  # Phase 6 depends on Phase 5
}

exarp_go_root = Path("/Users/davidl/Projects/exarp-go")

def main():
    print("Setting dependencies for migration tasks...")
    print()
    
    updates = []
    for task_id, deps in dependencies.items():
        updates.append({
            "id": task_id,
            "dependencies": deps
        })
        print(f"  {task_id}: depends on {', '.join(deps)}")
    
    print()
    result = update_todos_mcp(updates=updates, project_root=exarp_go_root)
    
    if result:
        print("✅ Dependencies set successfully")
    else:
        print("⚠️  Dependencies may not have been set (check Todo2)")

if __name__ == "__main__":
    main()

