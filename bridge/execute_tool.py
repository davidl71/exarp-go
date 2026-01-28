#!/usr/bin/env python3
"""
Python Bridge Tool Executor

Executes Python tools from the Go MCP server via subprocess.
This bridge allows the Go server to execute existing Python tools.

Supports both protobuf binary and JSON formats for backward compatibility.
"""

import json
import sys
from pathlib import Path

# Try to import protobuf (required for protobuf support)
try:
    from bridge.proto import bridge_pb2
    HAVE_PROTOBUF = True
except ImportError:
    HAVE_PROTOBUF = False
    # Fall back to JSON-only mode if protobuf not available

# Add local project_management_automation to path (copied to exarp-go)
PROJECT_ROOT = Path(__file__).parent.parent
sys.path.insert(0, str(PROJECT_ROOT))

# Add bridge directory to path for mcp-generic-tools modules
BRIDGE_ROOT = Path(__file__).parent
sys.path.insert(0, str(BRIDGE_ROOT))

def execute_tool(tool_name: str, args_json: str, use_protobuf: bool = False, protobuf_data: bytes = None, request_id: str = None):
    """Execute a Python tool with given arguments.
    
    Args:
        tool_name: Name of the tool to execute
        args_json: JSON string with arguments (for JSON mode)
        use_protobuf: Whether to use protobuf format
        protobuf_data: Protobuf binary data (if use_protobuf is True)
        request_id: Request ID for tracking (from protobuf request)
    
    Returns:
        If use_protobuf and HAVE_PROTOBUF: Returns protobuf ToolResponse binary
        Otherwise: Returns JSON string
    """
    import time
    start_time = time.time()
    
    try:
        # Parse arguments - try protobuf first if available, fall back to JSON
        args = {}
        if use_protobuf and HAVE_PROTOBUF and protobuf_data:
            try:
                # Parse protobuf ToolRequest
                request = bridge_pb2.ToolRequest()
                request.ParseFromString(protobuf_data)
                
                # Extract tool name, arguments, and request ID from protobuf
                tool_name = request.tool_name
                args_json = request.arguments_json
                request_id = request.request_id if request.request_id else None
                # Parse JSON arguments from protobuf message
                args = json.loads(args_json) if args_json else {}
            except Exception:
                # If protobuf parsing fails, fall back to JSON
                args = json.loads(args_json) if args_json else {}
        else:
            # Use JSON format (backward compatibility)
            args = json.loads(args_json) if args_json else {}
        
        # Import tool functions from project_management_automation
        # Note: memory, task_discovery - migrated to native Go (no Python handler, removed 2026-01-28)
        # Note: tool_catalog, workflow_mode, git_tools, infer_session_mode, health, automation,
        # generate_config, add_external_tool_hints, setup_hooks, check_attribution, session, memory_maint,
        # analyze_alignment, estimation, task_analysis - fully native Go with no Python handler
        from project_management_automation.tools.consolidated import (
            report as _report,
            security as _security,
            task_workflow as _task_workflow,
            testing as _testing,
            lint as _lint,
            mlx as _mlx,
            ollama as _ollama,
        )
        
        # Import mcp-generic-tools modules (from bridge directory)
        from recommend.tool import recommend as _recommend
        # Import context_tool for unified context wrapper (used for "context" tool)
        from project_management_automation.tools.context_tool import context as _context_unified
        
        # Route to appropriate tool
        # memory, task_discovery: migrated to native Go (removed 2026-01-28)
        if tool_name == "report":
            result = _report(
                action=args.get("action", "overview"),
                output_format=args.get("output_format", "text"),
                output_path=args.get("output_path"),
                include_recommendations=args.get("include_recommendations", True),
                overall_score=args.get("overall_score", 50.0),
                security_score=args.get("security_score", 50.0),
                testing_score=args.get("testing_score", 50.0),
                documentation_score=args.get("documentation_score", 50.0),
                completion_score=args.get("completion_score", 50.0),
                alignment_score=args.get("alignment_score", 50.0),
                project_name=args.get("project_name"),
                include_architecture=args.get("include_architecture", True),
                include_metrics=args.get("include_metrics", True),
                include_tasks=args.get("include_tasks", True),
            )
        elif tool_name == "security":
            result = _security(
                action=args.get("action", "report"),
                repo=args.get("repo", "davidl71/exarp-go"),
                languages=args.get("languages"),
                config_path=args.get("config_path"),
                state=args.get("state", "open"),
                include_dismissed=args.get("include_dismissed", False),
            )
        elif tool_name == "task_workflow":
            result = _task_workflow(
                action=args.get("action", "sync"),
                dry_run=args.get("dry_run", False),
                status=args.get("status", "Review"),
                new_status=args.get("new_status", "Todo"),
                clarification_none=args.get("clarification_none", True),
                filter_tag=args.get("filter_tag"),
                task_ids=args.get("task_ids"),
                sub_action=args.get("sub_action", "list"),
                task_id=args.get("task_id"),
                clarification_text=args.get("clarification_text"),
                decision=args.get("decision"),
                decisions_json=args.get("decisions_json"),
                move_to_todo=args.get("move_to_todo", True),
                auto_apply=args.get("auto_apply", False),
                output_format=args.get("output_format", "text"),
                stale_threshold_hours=args.get("stale_threshold_hours", 2.0),
                output_path=args.get("output_path"),
            )
        elif tool_name == "testing":
            result = _testing(
                action=args.get("action", "run"),
                test_path=args.get("test_path"),
                test_framework=args.get("test_framework", "auto"),
                verbose=args.get("verbose", True),
                coverage=args.get("coverage", False),
                coverage_file=args.get("coverage_file"),
                min_coverage=args.get("min_coverage", 80),
                format=args.get("format", "html"),
                target_file=args.get("target_file"),
                min_confidence=args.get("min_confidence", 0.7),
                framework=args.get("framework"),
                output_path=args.get("output_path"),
            )
        elif tool_name == "lint":
            result = _lint(
                action=args.get("action", "run"),
                path=args.get("path"),
                linter=args.get("linter", "ruff"),
                fix=args.get("fix", False),
                analyze=args.get("analyze", True),
                select=args.get("select"),
                ignore=args.get("ignore"),
                problems_json=args.get("problems_json"),
                include_hints=args.get("include_hints", True),
                output_path=args.get("output_path"),
            )
        # estimation: fully native Go, no Python handler (removed 2026-01-27)
        elif tool_name == "ollama":
            result = _ollama(
                action=args.get("action", "status"),
                host=args.get("host"),
                prompt=args.get("prompt"),
                model=args.get("model", "llama3.2"),
                stream=args.get("stream", False),
                options=args.get("options"),
                num_gpu=args.get("num_gpu"),
                num_threads=args.get("num_threads"),
                context_size=args.get("context_size"),
                file_path=args.get("file_path"),
                output_path=args.get("output_path"),
                style=args.get("style", "google"),
                include_suggestions=args.get("include_suggestions", True),
                data=args.get("data"),
                level=args.get("level", "brief"),
            )
        elif tool_name == "mlx":
            result = _mlx(
                action=args.get("action", "status"),
                prompt=args.get("prompt"),
                model=args.get("model", "mlx-community/Phi-3.5-mini-instruct-4bit"),
                max_tokens=args.get("max_tokens", 512),
                temperature=args.get("temperature", 0.7),
                verbose=args.get("verbose", False),
            )
        # Phase 3 Migration: Unified tools
        # Note: Individual tools (context_summarize, context_batch, prompt_log, prompt_analyze,
        # recommend_model, recommend_workflow) were removed in favor of unified tools below
        elif tool_name == "context":
            result = _context_unified(
                action=args.get("action", "summarize"),
                data=args.get("data"),
                level=args.get("level", "brief"),
                tool_type=args.get("tool_type"),
                max_tokens=args.get("max_tokens"),
                include_raw=args.get("include_raw", False),
                items=args.get("items"),
                budget_tokens=args.get("budget_tokens", 4000),
                combine=args.get("combine", True),
            )
        elif tool_name == "recommend":
            result = _recommend(
                action=args.get("action", "model"),
                task_description=args.get("task_description"),
                tags=args.get("tags"),
                include_rationale=args.get("include_rationale", True),
                task_type=args.get("task_type"),
                optimize_for=args.get("optimize_for", "quality"),
                include_alternatives=args.get("include_alternatives", True),
            )
        # Note: prompt_tracking and server_status removed - fully native Go with no Python fallback
        # Note: demonstrate_elicit and interactive_task_create removed
        # These tools required FastMCP Context (not available in stdio mode)
        else:
            execution_time_ms = int((time.time() - start_time) * 1000)
            error_result = {
                "success": False,
                "error": f"Unknown tool: {tool_name}",
                "tool": tool_name
            }
            
            # If protobuf mode, return protobuf ToolResponse
            if use_protobuf and HAVE_PROTOBUF:
                try:
                    response = bridge_pb2.ToolResponse()
                    response.success = False
                    response.error = error_result["error"]
                    response.execution_time_ms = execution_time_ms
                    response.exit_code = 1
                    if request_id:
                        response.request_id = request_id
                    sys.stdout.buffer.write(response.SerializeToString())
                    return 1
                except Exception:
                    # Fall back to JSON on error
                    print(json.dumps(error_result, indent=2))
                    return 1
            
            # Return JSON (backward compatible)
            print(json.dumps(error_result, indent=2))
            return 1
        
        # Calculate execution time
        execution_time_ms = int((time.time() - start_time) * 1000)
        
        # Handle result - tools may return dict or JSON string
        if isinstance(result, dict):
            result_json = json.dumps(result, indent=2)
        elif isinstance(result, str):
            result_json = result
        else:
            result_json = json.dumps({"result": str(result)}, indent=2)
        
        # If protobuf mode and protobuf is available, return protobuf ToolResponse
        if use_protobuf and HAVE_PROTOBUF:
            try:
                # Create protobuf ToolResponse
                response = bridge_pb2.ToolResponse()
                response.success = True
                response.result = result_json
                response.execution_time_ms = execution_time_ms
                response.exit_code = 0
                if request_id:
                    response.request_id = request_id
                
                # Serialize to binary and return
                return response.SerializeToString()
            except Exception as e:
                # If protobuf serialization fails, fall back to JSON
                # Include error in JSON response
                error_result = {
                    "success": False,
                    "error": f"Failed to serialize protobuf response: {str(e)}",
                    "result": result_json
                }
                return json.dumps(error_result, indent=2)
        
        # Return JSON (backward compatible)
        return result_json
        
    except Exception as e:
        # Calculate execution time even on error
        execution_time_ms = int((time.time() - start_time) * 1000)
        
        # If protobuf mode, return protobuf ToolResponse
        if use_protobuf and HAVE_PROTOBUF:
            try:
                response = bridge_pb2.ToolResponse()
                response.success = False
                response.error = str(e)
                response.execution_time_ms = execution_time_ms
                response.exit_code = 1
                if request_id:
                    response.request_id = request_id
                sys.stdout.buffer.write(response.SerializeToString())
                return 1
            except Exception:
                # Fall back to JSON if protobuf fails
                pass
        
        error_result = {
            "success": False,
            "error": str(e),
            "tool": tool_name
        }
        print(json.dumps(error_result, indent=2))
        return 1

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Tool name required"}, indent=2))
        sys.exit(1)
    
    tool_name = sys.argv[1]
    
    # Check if protobuf format is being used (--protobuf flag)
    use_protobuf = False
    protobuf_data = None
    
    if len(sys.argv) > 2 and sys.argv[2] == "--protobuf":
        # Protobuf mode: read binary data from stdin
        if HAVE_PROTOBUF:
            use_protobuf = True
            protobuf_data = sys.stdin.buffer.read()
        else:
            # Protobuf not available, fall back to JSON
            print(json.dumps({"error": "Protobuf support not available"}, indent=2), file=sys.stderr)
            sys.exit(1)
    else:
        # JSON mode: get JSON string from command line
        args_json = sys.argv[2] if len(sys.argv) > 2 else "{}"
    
    # Extract request_id from protobuf request if available
    request_id = None
    if use_protobuf and HAVE_PROTOBUF and protobuf_data:
        try:
            request = bridge_pb2.ToolRequest()
            request.ParseFromString(protobuf_data)
            request_id = request.request_id if request.request_id else None
        except Exception:
            # If parsing fails, continue without request_id
            pass
    
    # Execute tool
    result = execute_tool(tool_name, args_json if not use_protobuf else "", use_protobuf, protobuf_data, request_id)
    
    # If protobuf mode and result is bytes (protobuf binary), write to stdout.buffer
    if use_protobuf and HAVE_PROTOBUF and isinstance(result, bytes):
        sys.stdout.buffer.write(result)
        sys.exit(0)
    
    # Otherwise, result is JSON string or exit code
    if isinstance(result, str):
        print(result)
        sys.exit(0)
    else:
        sys.exit(result)

