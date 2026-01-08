#!/usr/bin/env python3
"""
Analyze and inventory tools from both codebases for migration planning
"""

import re
import json
from pathlib import Path
from collections import defaultdict

# Paths
pma_root = Path("/Users/davidl/Projects/project-management-automation")
exarp_go_root = Path("/Users/davidl/Projects/exarp-go")

# Extract tools from Python server.py
def extract_python_tools():
    """Extract all tool names from Python server.py"""
    server_py = pma_root / "project_management_automation" / "server.py"
    if not server_py.exists():
        return []
    
    tools = []
    with open(server_py, 'r') as f:
        content = f.read()
        
        # Find @mcp.tool() decorators followed by function definitions
        pattern = r'@mcp\.tool\(\)\s+def\s+(\w+)\s*\('
        matches = re.findall(pattern, content)
        tools.extend(matches)
        
        # Also check for consolidated tools in manual tool list
        # Look for Tool(name="...") patterns
        tool_pattern = r'Tool\(\s*name=["\'](\w+)["\']'
        tool_matches = re.findall(tool_pattern, content)
        tools.extend(tool_matches)
    
    return sorted(set(tools))

# Extract tools from Go registry.go
def extract_go_tools():
    """Extract all tool names from Go registry.go"""
    registry_go = exarp_go_root / "internal" / "tools" / "registry.go"
    if not registry_go.exists():
        return []
    
    tools = []
    with open(registry_go, 'r') as f:
        content = f.read()
        
        # Find RegisterTool("tool_name", ...) patterns
        pattern = r'RegisterTool\(\s*["\'](\w+)["\']'
        matches = re.findall(pattern, content)
        tools.extend(matches)
    
    return sorted(set(tools))

# Extract tools from bridge scripts
def extract_bridge_tools():
    """Extract tools available in Python bridge"""
    bridge_py = exarp_go_root / "bridge" / "execute_tool.py"
    if not bridge_py.exists():
        return []
    
    tools = []
    with open(bridge_py, 'r') as f:
        content = f.read()
        
        # Find tool_name == "..." patterns
        pattern = r'tool_name\s*==\s*["\'](\w+)["\']'
        matches = re.findall(pattern, content)
        tools.extend(matches)
    
    return sorted(set(tools))

def main():
    print("=" * 60)
    print("Tool Inventory Analysis")
    print("=" * 60)
    print()
    
    # Extract tools
    print("Extracting tools from Python server.py...")
    python_tools = extract_python_tools()
    print(f"  Found {len(python_tools)} tools")
    
    print("Extracting tools from Go registry.go...")
    go_tools = extract_go_tools()
    print(f"  Found {len(go_tools)} tools")
    
    print("Extracting tools from bridge scripts...")
    bridge_tools = extract_bridge_tools()
    print(f"  Found {len(bridge_tools)} tools")
    
    print()
    print("=" * 60)
    print("Analysis Results")
    print("=" * 60)
    print()
    
    # Find migrated tools (in both Python and Go)
    migrated = sorted(set(python_tools) & set(go_tools))
    print(f"✅ Migrated tools ({len(migrated)}):")
    for tool in migrated:
        print(f"  - {tool}")
    
    print()
    
    # Find tools only in Python (need migration)
    python_only = sorted(set(python_tools) - set(go_tools))
    print(f"📋 Tools to migrate ({len(python_only)}):")
    for tool in python_only:
        print(f"  - {tool}")
    
    print()
    
    # Find tools only in Go (new tools)
    go_only = sorted(set(go_tools) - set(python_tools))
    if go_only:
        print(f"🆕 New tools in Go ({len(go_only)}):")
        for tool in go_only:
            print(f"  - {tool}")
        print()
    
    # Create inventory document
    inventory = {
        "python_tools": python_tools,
        "go_tools": go_tools,
        "bridge_tools": bridge_tools,
        "migrated": migrated,
        "to_migrate": python_only,
        "new_in_go": go_only,
        "statistics": {
            "total_python": len(python_tools),
            "total_go": len(go_tools),
            "migrated_count": len(migrated),
            "to_migrate_count": len(python_only),
            "new_in_go_count": len(go_only),
            "migration_progress": f"{len(migrated)}/{len(python_tools)} ({len(migrated)/len(python_tools)*100:.1f}%)" if python_tools else "N/A"
        }
    }
    
    # Write JSON inventory
    inventory_file = exarp_go_root / "docs" / "MIGRATION_INVENTORY.json"
    with open(inventory_file, 'w') as f:
        json.dump(inventory, f, indent=2)
    print(f"✅ Inventory saved to: {inventory_file}")
    
    # Create markdown inventory
    md_file = exarp_go_root / "docs" / "MIGRATION_INVENTORY.md"
    with open(md_file, 'w') as f:
        f.write("# Migration Inventory: project-management-automation → exarp-go\n\n")
        f.write(f"**Generated:** 2026-01-07\n")
        f.write(f"**Status:** Phase 1 Analysis\n\n")
        f.write("---\n\n")
        f.write("## Statistics\n\n")
        f.write(f"- **Total Python Tools:** {len(python_tools)}\n")
        f.write(f"- **Total Go Tools:** {len(go_tools)}\n")
        f.write(f"- **Migrated:** {len(migrated)}\n")
        f.write(f"- **To Migrate:** {len(python_only)}\n")
        f.write(f"- **New in Go:** {len(go_only)}\n")
        f.write(f"- **Migration Progress:** {inventory['statistics']['migration_progress']}\n\n")
        f.write("---\n\n")
        f.write("## Migrated Tools\n\n")
        for tool in migrated:
            f.write(f"- ✅ `{tool}`\n")
        f.write("\n---\n\n")
        f.write("## Tools to Migrate\n\n")
        for tool in python_only:
            f.write(f"- 📋 `{tool}`\n")
        f.write("\n---\n\n")
        if go_only:
            f.write("## New Tools in Go\n\n")
            for tool in go_only:
                f.write(f"- 🆕 `{tool}`\n")
            f.write("\n---\n\n")
        f.write("## All Python Tools\n\n")
        for tool in python_tools:
            status = "✅" if tool in migrated else "📋"
            f.write(f"- {status} `{tool}`\n")
        f.write("\n---\n\n")
        f.write("## All Go Tools\n\n")
        for tool in go_tools:
            status = "✅" if tool in migrated else "🆕"
            f.write(f"- {status} `{tool}`\n")
    
    print(f"✅ Markdown inventory saved to: {md_file}")
    print()
    print("=" * 60)
    print("Summary")
    print("=" * 60)
    print(f"Total tools in Python: {len(python_tools)}")
    print(f"Total tools in Go: {len(go_tools)}")
    print(f"Already migrated: {len(migrated)}")
    print(f"Remaining to migrate: {len(python_only)}")
    print(f"Migration progress: {inventory['statistics']['migration_progress']}")

if __name__ == "__main__":
    main()

