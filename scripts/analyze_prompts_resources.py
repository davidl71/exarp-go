#!/usr/bin/env python3
"""
Analyze prompts and resources from both codebases
"""

import re
from pathlib import Path

pma_root = Path("/Users/davidl/Projects/project-management-automation")
exarp_go_root = Path("/Users/davidl/Projects/exarp-go")

def extract_python_prompts():
    """Extract prompt names from Python server.py"""
    server_py = pma_root / "project_management_automation" / "server.py"
    if not server_py.exists():
        return []
    
    prompts = []
    with open(server_py, 'r') as f:
        content = f.read()
        
        # Find @mcp.prompt() decorators
        pattern = r'@mcp\.prompt\(\)\s+def\s+(\w+)\s*\('
        matches = re.findall(pattern, content)
        prompts.extend(matches)
    
    return sorted(set(prompts))

def extract_python_resources():
    """Extract resource URIs from Python server.py"""
    server_py = pma_root / "project_management_automation" / "server.py"
    if not server_py.exists():
        return []
    
    resources = []
    with open(server_py, 'r') as f:
        content = f.read()
        
        # Find @mcp.resource("uri") patterns
        pattern = r'@mcp\.resource\(["\']([^"\']+)["\']'
        matches = re.findall(pattern, content)
        resources.extend(matches)
    
    return sorted(set(resources))

def extract_go_prompts():
    """Extract prompt names from Go registry.go"""
    registry_go = exarp_go_root / "internal" / "prompts" / "registry.go"
    if not registry_go.exists():
        return []
    
    prompts = []
    with open(registry_go, 'r') as f:
        content = f.read()
        
        # Find {"name", "description"} patterns
        pattern = r'\{"(\w+)",\s*"[^"]+"\}'
        matches = re.findall(pattern, content)
        prompts.extend(matches)
    
    return sorted(set(prompts))

def extract_go_resources():
    """Extract resource URIs from Go handlers.go"""
    handlers_go = exarp_go_root / "internal" / "resources" / "handlers.go"
    if not handlers_go.exists():
        return []
    
    resources = []
    with open(handlers_go, 'r') as f:
        content = f.read()
        
        # Find RegisterResource("uri", ...) patterns
        pattern = r'RegisterResource\(\s*["\']([^"\']+)["\']'
        matches = re.findall(pattern, content)
        resources.extend(matches)
    
    return sorted(set(resources))

def main():
    print("=" * 60)
    print("Prompts and Resources Analysis")
    print("=" * 60)
    print()
    
    python_prompts = extract_python_prompts()
    go_prompts = extract_go_prompts()
    python_resources = extract_python_resources()
    go_resources = extract_go_resources()
    
    print(f"Python Prompts: {len(python_prompts)}")
    print(f"Go Prompts: {len(go_prompts)}")
    print(f"Python Resources: {len(python_resources)}")
    print(f"Go Resources: {len(go_resources)}")
    print()
    
    migrated_prompts = sorted(set(python_prompts) & set(go_prompts))
    prompts_to_migrate = sorted(set(python_prompts) - set(go_prompts))
    
    # Resources are harder to compare (different URI formats)
    # Python uses automation://, Go uses stdio://
    migrated_resources_count = len(go_resources)  # Assume all Go resources are migrated
    resources_to_migrate = len(python_resources) - migrated_resources_count
    
    print(f"✅ Migrated Prompts: {len(migrated_prompts)}")
    print(f"📋 Prompts to Migrate: {len(prompts_to_migrate)}")
    print(f"✅ Migrated Resources: {migrated_resources_count}")
    print(f"📋 Resources to Migrate: {resources_to_migrate}")
    print()
    
    if prompts_to_migrate:
        print("Prompts to migrate:")
        for p in prompts_to_migrate:
            print(f"  - {p}")
        print()
    
    print("Python Resources:")
    for r in python_resources[:10]:  # Show first 10
        print(f"  - {r}")
    if len(python_resources) > 10:
        print(f"  ... and {len(python_resources) - 10} more")
    print()
    
    print("Go Resources:")
    for r in go_resources:
        print(f"  - {r}")

if __name__ == "__main__":
    main()

