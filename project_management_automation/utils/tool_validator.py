"""
Tool Validation Utilities

Validates MCP tools to ensure they follow FastMCP best practices and avoid
conditional logic patterns that can cause "object dict can't be used in 'await' expression" errors.
"""

import ast
import re
from pathlib import Path
from typing import Any, Optional


class ToolValidationError(Exception):
    """Raised when a tool fails validation."""
    pass


class ToolValidator:
    """Validates MCP tool implementations for FastMCP compatibility."""
    
    # Patterns that indicate problematic conditional logic
    PROBLEMATIC_PATTERNS = [
        r'if\s+action\s*==',  # Conditional on action parameter
        r'elif\s+action\s*==',  # Elif with action
        r'if\s+\w+\s*==\s*["\']',  # String comparison in if (might be action-based)
    ]
    
    def __init__(self, server_file: Optional[Path] = None):
        """Initialize validator with server file path."""
        if server_file is None:
            # Default to project server.py
            script_dir = Path(__file__).resolve().parent
            # Go up from utils/ to project_management_automation/ to find server.py
            server_file = script_dir.parent / 'server.py'
        
        self.server_file = Path(server_file).resolve()
        self.issues: list[dict[str, Any]] = []
    
    def validate_all_tools(self) -> dict[str, Any]:
        """Validate all MCP tools in server.py."""
        if not self.server_file.exists():
            raise ToolValidationError(f"Server file not found: {self.server_file}")
        
        tools = self._find_mcp_tools()
        results = {
            'total_tools': len(tools),
            'valid_tools': [],
            'invalid_tools': [],
            'warnings': [],
            'errors': []
        }
        
        for tool in tools:
            validation = self._validate_tool(tool)
            
            if validation['is_valid']:
                results['valid_tools'].append(tool['name'])
            else:
                results['invalid_tools'].append({
                    'name': tool['name'],
                    'line': tool['line'],
                    'issues': validation['issues']
                })
            
            if validation.get('warnings'):
                results['warnings'].extend([
                    {'tool': tool['name'], 'warning': w}
                    for w in validation['warnings']
                ])
        
        return results
    
    def _find_mcp_tools(self) -> list[dict[str, Any]]:
        """Find all @mcp.tool() decorated functions."""
        with open(self.server_file, 'r') as f:
            content = f.read()
        
        tree = ast.parse(content, filename=str(self.server_file))
        tools = []
        
        for node in ast.walk(tree):
            if isinstance(node, ast.FunctionDef):
                # Check for @mcp.tool() decorator
                for decorator in node.decorator_list:
                    if isinstance(decorator, ast.Call):
                        if isinstance(decorator.func, ast.Attribute):
                            if (isinstance(decorator.func.value, ast.Name) and
                                decorator.func.value.id == 'mcp' and
                                decorator.func.attr == 'tool'):
                                tools.append({
                                    'name': node.name,
                                    'line': node.lineno,
                                    'node': node
                                })
                                break
        
        return tools
    
    def _validate_tool(self, tool: dict[str, Any]) -> dict[str, Any]:
        """Validate a single tool."""
        tool_name = tool['name']
        line_num = tool['line']
        
        # Get function body
        with open(self.server_file, 'r') as f:
            lines = f.readlines()
        
        # Find function body
        func_body, func_start, func_end = self._extract_function_body(
            lines, tool_name, line_num
        )
        
        if not func_body:
            return {
                'is_valid': False,
                'issues': [f'Could not extract function body']
            }
        
        issues = []
        warnings = []
        
        # Check 1: No conditional logic based on action parameter
        if self._has_action_based_conditionals(func_body):
            issues.append(
                f"Tool has conditional logic based on 'action' parameter. "
                f"Split into separate tools to avoid FastMCP issues."
            )
        
        # Check 2: Simple return pattern (preferred)
        if not self._has_simple_return(func_body):
            warnings.append(
                "Tool does not follow simple return pattern. "
                "Consider simplifying to: return _underlying_function(...)"
            )
        
        # Check 3: Has @ensure_json_string decorator
        if not self._has_ensure_json_string_decorator(lines, func_start):
            warnings.append(
                "Tool missing @ensure_json_string decorator. "
                "This decorator ensures FastMCP receives JSON strings."
            )
        
        # Check 4: No multiple if/elif/else branches
        if_count = len(re.findall(r'\bif\s+', func_body))
        elif_count = len(re.findall(r'\belif\s+', func_body))
        else_count = len(re.findall(r'\belse\s*:', func_body))
        
        if if_count > 2 or (if_count + elif_count) > 3:
            issues.append(
                f"Tool has complex conditional logic ({if_count} if, {elif_count} elif, {else_count} else). "
                f"FastMCP may have issues with complex control flow."
            )
        
        # Check 5: Function length (should be simple wrapper)
        func_length = func_end - func_start
        if func_length > 50:
            warnings.append(
                f"Tool function is long ({func_length} lines). "
                f"Consider moving logic to helper functions."
            )
        
        return {
            'is_valid': len(issues) == 0,
            'issues': issues,
            'warnings': warnings
        }
    
    def _extract_function_body(
        self, lines: list[str], func_name: str, line_num: int
    ) -> tuple[Optional[str], Optional[int], Optional[int]]:
        """Extract function body from lines."""
        func_start = None
        func_end = None
        
        # Find function start
        for i in range(line_num - 1, len(lines)):
            if f'def {func_name}(' in lines[i]:
                func_start = i
                break
        
        if func_start is None:
            return None, None, None
        
        # Calculate indent
        indent_level = len(lines[func_start]) - len(lines[func_start].lstrip())
        
        # Find function end
        for i in range(func_start + 1, len(lines)):
            line = lines[i]
            current_indent = len(line) - len(line.lstrip())
            
            if line.strip() and current_indent <= indent_level:
                if (line.strip().startswith('def ') or
                    line.strip().startswith('@') or
                    line.strip().startswith('class ')):
                    func_end = i
                    break
        
        if func_end is None:
            func_end = len(lines)
        
        func_body = ''.join(lines[func_start:func_end])
        return func_body, func_start, func_end
    
    def _has_action_based_conditionals(self, func_body: str) -> bool:
        """Check if function has conditional logic based on action parameter."""
        for pattern in self.PROBLEMATIC_PATTERNS:
            if re.search(pattern, func_body, re.IGNORECASE):
                return True
        return False
    
    def _has_simple_return(self, func_body: str) -> bool:
        """Check if function has simple return pattern."""
        # Simple pattern: return _function_name(...)
        simple_pattern = r'return\s+_\w+\([^)]*\)'
        returns = re.findall(r'return\s+', func_body)
        
        if len(returns) == 1:
            if re.search(simple_pattern, func_body):
                return True
        
        return False
    
    def _has_ensure_json_string_decorator(
        self, lines: list[str], func_start: int
    ) -> bool:
        """Check if function has @ensure_json_string decorator."""
        # Look backwards from function start for decorators
        for i in range(max(0, func_start - 5), func_start):
            if '@ensure_json_string' in lines[i]:
                return True
        return False
    
    def generate_report(self, output_path: Optional[Path] = None) -> str:
        """Generate validation report."""
        results = self.validate_all_tools()
        
        report_lines = [
            "# Tool Validation Report",
            "",
            "## Summary",
            "",
            f"- **Total Tools:** {results['total_tools']}",
            f"- **Valid Tools:** {len(results['valid_tools'])}",
            f"- **Invalid Tools:** {len(results['invalid_tools'])}",
            f"- **Warnings:** {len(results['warnings'])}",
            "",
            "## Validation Rules",
            "",
            "1. ✅ No conditional logic based on 'action' parameter",
            "2. ✅ Simple return pattern (direct call to underlying function)",
            "3. ✅ @ensure_json_string decorator applied",
            "4. ✅ Minimal conditional logic (< 3 if/elif branches)",
            "5. ✅ Function length < 50 lines",
            "",
        ]
        
        if results['invalid_tools']:
            report_lines.extend([
                "## ❌ Invalid Tools",
                ""
            ])
            
            for tool in results['invalid_tools']:
                report_lines.extend([
                    f"### {tool['name']} (line {tool['line']})",
                    ""
                ])
                
                for issue in tool['issues']:
                    report_lines.append(f"- ❌ {issue}")
                
                report_lines.append("")
        
        if results['warnings']:
            report_lines.extend([
                "## ⚠️  Warnings",
                ""
            ])
            
            for warning in results['warnings']:
                report_lines.append(
                    f"- ⚠️  **{warning['tool']}**: {warning['warning']}"
                )
            
            report_lines.append("")
        
        if results['valid_tools']:
            report_lines.extend([
                "## ✅ Valid Tools",
                ""
            ])
            
            for tool_name in sorted(results['valid_tools']):
                report_lines.append(f"- ✅ {tool_name}")
            
            report_lines.append("")
        
        report = '\n'.join(report_lines)
        
        if output_path:
            output_path.write_text(report)
            return f"Report written to: {output_path}"
        
        return report


def validate_tools(server_file: Optional[Path] = None) -> dict[str, Any]:
    """Convenience function to validate all tools."""
    validator = ToolValidator(server_file)
    return validator.validate_all_tools()


if __name__ == '__main__':
    import sys
    
    validator = ToolValidator()
    results = validator.validate_all_tools()
    
    print("🔍 Tool Validation Results")
    print("=" * 80)
    print(f"Total tools: {results['total_tools']}")
    print(f"✅ Valid: {len(results['valid_tools'])}")
    print(f"❌ Invalid: {len(results['invalid_tools'])}")
    print(f"⚠️  Warnings: {len(results['warnings'])}")
    
    has_errors = False
    
    if results['invalid_tools']:
        has_errors = True
        print("\n❌ Invalid Tools:")
        for tool in results['invalid_tools']:
            print(f"  - {tool['name']} (line {tool['line']})")
            for issue in tool['issues']:
                print(f"    • {issue}")
    
    if results['warnings']:
        print("\n⚠️  Warnings:")
        for warning in results['warnings']:
            print(f"  - {warning['tool']}: {warning['warning']}")
    
    # Generate report
    script_dir = Path(__file__).resolve().parent
    # Go from utils/ -> project_management_automation/ -> repo root
    repo_root = script_dir.parent.parent
    report_path = repo_root / 'docs' / 'TOOL_VALIDATION_REPORT.md'
    validator.generate_report(report_path)
    print(f"\n📄 Report written to: {report_path}")
    
    # Exit with error code if validation failed
    if has_errors:
        print("\n❌ Validation failed! Fix invalid tools before committing.")
        print("   See docs/FASTMCP_TOOL_CONSTRAINTS.md for guidelines.")
        sys.exit(1)
    else:
        print("\n✅ Validation passed!")
        sys.exit(0)

