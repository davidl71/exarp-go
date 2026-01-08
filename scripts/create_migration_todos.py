#!/usr/bin/env python3
"""
Create Todo2 tasks for migration plan from project-management-automation to exarp-go
"""

import sys
import json
from pathlib import Path

# Add project-management-automation to path for Todo2 MCP client
project_management_automation = Path("/Users/davidl/Projects/project-management-automation")
sys.path.insert(0, str(project_management_automation))

try:
    from project_management_automation.utils.todo2_mcp_client import create_todos_mcp
except ImportError as e:
    print(f"Error importing Todo2 MCP client: {e}")
    print("Make sure project-management-automation is available")
    sys.exit(1)

# Project root for exarp-go
exarp_go_root = Path("/Users/davidl/Projects/exarp-go")

# Define migration tasks
migration_todos = [
    {
        "name": "Phase 1: Analyze and inventory remaining tools to migrate",
        "long_description": """🎯 **Objective:** Complete analysis of what tools, prompts, and resources remain in project-management-automation that need migration to exarp-go.

📋 **Acceptance Criteria:**
- Complete inventory of all tools in project-management-automation/server.py
- Identify which tools are already migrated to exarp-go (24 tools currently)
- Identify gaps: tools, prompts, and resources still in Python
- Document dependencies and relationships between tools
- Create migration priority matrix (high/medium/low priority)

🚫 **Scope Boundaries (CRITICAL):**
- **Included:** Tool inventory, gap analysis, dependency mapping, priority assessment
- **Excluded:** Actual migration implementation, testing, documentation updates
- **Clarification Required:** None

🔧 **Technical Requirements:**
- Analyze Python server.py for all @mcp.tool() registrations
- Compare with exarp-go/internal/tools/registry.go (24 tools migrated)
- Identify FastMCP dependencies vs standalone functions
- Map tool dependencies (which tools call other tools)

📁 **Files/Components:**
- Analyze: /Users/davidl/Projects/project-management-automation/project_management_automation/server.py
- Compare: /Users/davidl/Projects/exarp-go/internal/tools/registry.go
- Create: /Users/davidl/Projects/exarp-go/docs/MIGRATION_INVENTORY.md

🧪 **Testing Requirements:**
- Verify tool count matches actual registrations
- Validate no duplicate tools between projects
- Check for broken imports or missing dependencies

⚠️ **Edge Cases:**
- Tools that depend on FastMCP Context (may need special handling)
- Tools with complex Python dependencies
- Tools that are deprecated or unused

📚 **Dependencies:** None""",
        "priority": "high",
        "tags": ["migration", "analysis", "inventory", "phase-1"]
    },
    {
        "name": "Phase 2: Design migration strategy and architecture",
        "long_description": """🎯 **Objective:** Design comprehensive migration strategy for remaining Python tools to exarp-go Go implementation.

📋 **Acceptance Criteria:**
- Migration strategy document created (native Go vs Python bridge)
- Decision framework for each tool (when to use Go vs bridge)
- Architecture design for new tool implementations
- Dependency migration plan (Python deps → Go alternatives)
- Testing strategy for migrated tools

🚫 **Scope Boundaries (CRITICAL):**
- **Included:** Strategy design, architecture decisions, migration patterns, testing approach
- **Excluded:** Actual implementation, tool-by-tool migration
- **Clarification Required:** User preference on Go vs Python bridge (performance vs speed)

🔧 **Technical Requirements:**
- Define Go implementation patterns (follow existing 24 tools)
- Design Python bridge patterns (for complex Python deps)
- Plan dependency migration (FastMCP → stdio, Python libs → Go libs)
- Create tool registration templates

📁 **Files/Components:**
- Create: /Users/davidl/Projects/exarp-go/docs/MIGRATION_STRATEGY.md
- Create: /Users/davidl/Projects/exarp-go/docs/MIGRATION_PATTERNS.md
- Update: /Users/davidl/Projects/exarp-go/internal/tools/registry.go (document patterns)

🧪 **Testing Requirements:**
- Strategy document review
- Architecture validation
- Pattern examples

⚠️ **Edge Cases:**
- Tools requiring FastMCP Context (elicit, interactive features)
- Tools with heavy Python dependencies (MLX, Ollama wrappers)
- Tools with complex state management

📚 **Dependencies:** Phase 1 (analysis complete)""",
        "priority": "high",
        "tags": ["migration", "strategy", "architecture", "phase-2"],
        "dependencies": []  # Will be set after Phase 1 is created
    },
    {
        "name": "Phase 3: Migrate high-priority tools (batch 1)",
        "long_description": """🎯 **Objective:** Migrate first batch of high-priority tools from project-management-automation to exarp-go.

📋 **Acceptance Criteria:**
- 5-10 high-priority tools migrated
- All tools registered in internal/tools/registry.go
- Python bridge handlers implemented (or native Go if appropriate)
- Integration tests passing
- Documentation updated

🚫 **Scope Boundaries (CRITICAL):**
- **Included:** Tool migration, registration, bridge handlers, basic tests
- **Excluded:** Complex tool refactoring, performance optimization, full test coverage
- **Clarification Required:** Which tools are highest priority (user input or based on usage)

🔧 **Technical Requirements:**
- Follow existing migration patterns from 24 tools
- Use Python bridge for complex Python dependencies
- Register tools in registry.go (Batch 4)
- Implement Go handlers calling bridge scripts

📁 **Files/Components:**
- Update: /Users/davidl/Projects/exarp-go/internal/tools/registry.go
- Create: /Users/davidl/Projects/exarp-go/internal/tools/handlers.go (if new handlers)
- Update: /Users/davidl/Projects/exarp-go/bridge/execute_tool.py (add new tools)
- Create: /Users/davidl/Projects/exarp-go/tests/integration/tools/ (test files)

🧪 **Testing Requirements:**
- Integration tests for each migrated tool
- Verify tool registration in MCP server
- Test tool execution via bridge
- Validate error handling

⚠️ **Edge Cases:**
- Tools with different parameter types
- Tools returning complex data structures
- Tools with async operations

📚 **Dependencies:** Phase 2 (strategy complete)""",
        "priority": "high",
        "tags": ["migration", "implementation", "batch-1", "phase-3"],
        "dependencies": []  # Will be set after Phase 2 is created
    },
    {
        "name": "Phase 4: Migrate remaining tools (batch 2)",
        "long_description": """🎯 **Objective:** Complete migration of remaining tools from project-management-automation to exarp-go.

📋 **Acceptance Criteria:**
- All remaining tools migrated (excluding deprecated/unused)
- All tools registered and tested
- Python bridge updated for all new tools
- Integration tests passing
- No regressions in existing functionality

🚫 **Scope Boundaries (CRITICAL):**
- **Included:** Tool migration, testing, documentation
- **Excluded:** Performance optimization, refactoring, new features
- **Clarification Required:** Which tools to deprecate vs migrate

🔧 **Technical Requirements:**
- Complete tool migration following established patterns
- Register all tools in registry.go (Batch 5+)
- Update bridge scripts for all new tools
- Comprehensive integration testing

📁 **Files/Components:**
- Update: /Users/davidl/Projects/exarp-go/internal/tools/registry.go
- Update: /Users/davidl/Projects/exarp-go/bridge/execute_tool.py
- Create: /Users/davidl/Projects/exarp-go/tests/integration/tools/ (additional tests)
- Update: /Users/davidl/Projects/exarp-go/README.md (tool count)

🧪 **Testing Requirements:**
- Full integration test suite
- Verify all tools work via MCP interface
- Test tool combinations and workflows
- Validate error handling across all tools

⚠️ **Edge Cases:**
- Tools with unique requirements
- Tools that conflict with existing implementations
- Tools requiring special configuration

📚 **Dependencies:** Phase 3 (batch 1 complete)""",
        "priority": "medium",
        "tags": ["migration", "implementation", "batch-2", "phase-4"],
        "dependencies": []  # Will be set after Phase 3 is created
    },
    {
        "name": "Phase 5: Migrate prompts and resources",
        "long_description": """🎯 **Objective:** Migrate all prompts and resources from project-management-automation to exarp-go.

📋 **Acceptance Criteria:**
- All prompts migrated (15+ prompts)
- All resources migrated (6+ resources)
- Prompt handlers implemented in Go
- Resource handlers implemented in Go
- Integration tests passing

🚫 **Scope Boundaries (CRITICAL):**
- **Included:** Prompt migration, resource migration, handlers, tests
- **Excluded:** Tool migration (handled in previous phases)
- **Clarification Required:** None

🔧 **Technical Requirements:**
- Follow existing prompt/resource patterns from 15 prompts and 6 resources
- Use Python bridge for prompt/resource retrieval
- Register in internal/prompts/registry.go and internal/resources/handlers.go
- Update bridge scripts (get_prompt.py, execute_resource.py)

📁 **Files/Components:**
- Update: /Users/davidl/Projects/exarp-go/internal/prompts/registry.go
- Update: /Users/davidl/Projects/exarp-go/internal/resources/handlers.go
- Update: /Users/davidl/Projects/exarp-go/bridge/get_prompt.py
- Update: /Users/davidl/Projects/exarp-go/bridge/execute_resource.py
- Create: /Users/davidl/Projects/exarp-go/tests/integration/prompts/
- Create: /Users/davidl/Projects/exarp-go/tests/integration/resources/

🧪 **Testing Requirements:**
- Integration tests for all prompts
- Integration tests for all resources
- Verify prompt/resource registration
- Test prompt/resource retrieval via MCP

⚠️ **Edge Cases:**
- Prompts with dynamic content
- Resources with complex URIs
- Prompts/resources with dependencies

📚 **Dependencies:** Phase 4 (tools migration complete)""",
        "priority": "medium",
        "tags": ["migration", "prompts", "resources", "phase-5"],
        "dependencies": []  # Will be set after Phase 4 is created
    },
    {
        "name": "Phase 6: Testing, validation, and cleanup",
        "long_description": """🎯 **Objective:** Complete testing, validation, and cleanup of migration from project-management-automation to exarp-go.

📋 **Acceptance Criteria:**
- All migrated tools tested and working
- All prompts and resources tested
- Performance validation (Go vs Python bridge)
- Documentation updated
- Migration complete summary created
- Old Python code marked for deprecation

🚫 **Scope Boundaries (CRITICAL):**
- **Included:** Testing, validation, documentation, cleanup, summary
- **Excluded:** New features, performance optimization, refactoring
- **Clarification Required:** None

🔧 **Technical Requirements:**
- Comprehensive integration test suite
- Performance benchmarks (if applicable)
- Documentation updates (README, migration docs)
- Code cleanup (remove unused code, update comments)

📁 **Files/Components:**
- Update: /Users/davidl/Projects/exarp-go/README.md
- Update: /Users/davidl/Projects/exarp-go/docs/MIGRATION_COMPLETE.md
- Create: /Users/davidl/Projects/exarp-go/docs/MIGRATION_VALIDATION.md
- Update: /Users/davidl/Projects/project-management-automation/README.md (deprecation notice)

🧪 **Testing Requirements:**
- Full integration test coverage
- MCP interface validation
- Tool/prompt/resource functionality verification
- Error handling validation
- Performance testing (if needed)

⚠️ **Edge Cases:**
- Tools with edge case behavior
- Performance regressions
- Compatibility issues

📚 **Dependencies:** Phase 5 (prompts/resources complete)""",
        "priority": "medium",
        "tags": ["migration", "testing", "validation", "cleanup", "phase-6"],
        "dependencies": []  # Will be set after Phase 5 is created
    }
]

def main():
    """Create Todo2 tasks for migration plan"""
    print("=" * 60)
    print("Creating Todo2 tasks for migration plan")
    print("=" * 60)
    print()
    
    # Create todos
    created_todos = []
    for i, todo in enumerate(migration_todos, 1):
        print(f"Creating task {i}/6: {todo['name']}")
        
        try:
            # Create todo using Todo2 MCP client
            result = create_todos_mcp(
                todos=[{
                    "name": todo["name"],
                    "long_description": todo["long_description"],
                    "priority": todo["priority"],
                    "tags": todo["tags"]
                }],
                project_root=exarp_go_root
            )
            
            if result and len(result) > 0:
                created_id = result[0]
                created_todos.append(created_id)
                print(f"  ✅ Created: {created_id}")
            else:
                print(f"  ⚠️  Created but no ID returned")
                created_todos.append(None)
                
        except Exception as e:
            print(f"  ❌ Error: {e}")
            created_todos.append(None)
    
    print()
    print("=" * 60)
    print("Task Creation Summary")
    print("=" * 60)
    
    # Update dependencies
    for i, (todo, todo_id) in enumerate(zip(migration_todos, created_todos)):
        if todo_id and i > 0:
            prev_id = created_todos[i-1]
            if prev_id:
                print(f"Setting dependency: {todo_id} depends on {prev_id}")
                # Note: Dependency setting would need to be done via update_todos
                # For now, we'll just note it
    
    print()
    print(f"Created {len([t for t in created_todos if t])} tasks")
    print()
    print("Next steps:")
    print("1. Review created tasks in Todo2")
    print("2. Set dependencies between phases")
    print("3. Switch to plan mode to create execution plan")
    print()

if __name__ == "__main__":
    main()

