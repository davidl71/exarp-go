"""
Consolidated MCP Tools

Combines related tools into unified interfaces with action parameters.
Reduces tool count while maintaining all functionality.

All tools use 'action' as the dispatcher parameter for consistency.

Consolidated tools:
- analyze_alignment(action=todo2|prd) ← analyze_todo2_alignment, analyze_prd_alignment
- automation(action=daily|nightly|sprint|discover) ← run_daily_automation, run_nightly_automation, run_sprint_automation, run_discover_automation
- estimation(action=estimate|analyze|stats) ← estimate_task_duration, analyze_estimation_accuracy, get_estimation_statistics
- security(action=scan|alerts|report) ← scan_dependency_security, fetch_dependabot_alerts, generate_security_report
- generate_config(action=rules|ignore|simplify) ← generate_cursor_rules, generate_cursorignore, simplify_rules
- setup_hooks(action=git|patterns) ← setup_git_hooks, setup_pattern_triggers
- prompt_tracking(action=log|analyze) ← log_prompt_iteration, analyze_prompt_iterations
- health(action=server|git|docs|dod|cicd) ← server_status, check_working_copy_health, check_documentation_health, check_definition_of_done, validate_ci_cd_workflow
- report(action=overview|scorecard|briefing|prd) ← generate_project_overview, generate_project_scorecard, get_daily_briefing, generate_prd
- advisor_audio removed - migrated to devwisdom-go MCP server
- task_analysis(action=duplicates|tags|hierarchy|dependencies|parallelization) ← detect_duplicate_tasks, consolidate_tags, analyze_task_hierarchy, analyze_todo2_dependencies, optimize_todo2_parallelization
- testing(action=run|coverage|suggest|generate|validate) ← run_tests, analyze_test_coverage, suggest_test_cases, generate_test_code (AI), validate_test_structure
- lint(action=run|analyze) ← run_linter, analyze_problems
- memory(action=save|recall|search) ← save_memory, recall_context, search_memories
- memory_maint(action=health|gc|prune|consolidate|dream) ← memory lifecycle management and advisor dreaming
- task_discovery(action=comments|markdown|orphans|all) ← NEW: find tasks from various sources
- task_workflow(action=sync|approve|clarify|clarity|cleanup, sub_action for clarify) ← sync_todo_tasks, batch_approve_tasks, clarification, improve_task_clarity, cleanup_stale_tasks
- context(action=summarize|budget|batch) ← moved to context_tool.py
- tool_catalog(action=list|help) ← list_tools, get_tool_help
- workflow_mode(action=focus|suggest|stats) ← focus_mode, suggest_mode, get_tool_usage_stats
- recommend(action=model|workflow|advisor) ← recommend_model, recommend_workflow_mode, consult_advisor
"""

import asyncio
import json
import logging
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)


"""
Consolidated Ai

Consolidated tools: ollama, mlx, coreml
"""

def ollama(
    action: str = "status",
    host: Optional[str] = None,
    prompt: Optional[str] = None,
    model: str = "llama3.2",
    stream: bool = False,
    options: Optional[str] = None,
    num_gpu: Optional[int] = None,
    num_threads: Optional[int] = None,
    context_size: Optional[int] = None,
    file_path: Optional[str] = None,
    output_path: Optional[str] = None,
    style: str = "google",
    include_suggestions: bool = True,
    data: Optional[str] = None,
    level: str = "brief",
) -> str:
    """
    [HINT: Ollama. action=status|models|generate|pull|hardware|docs|quality|summary. Unified Ollama tool.]

    Unified Ollama tool consolidating integration and enhanced tools.

    Actions:
    - action="status": Check if Ollama server is running
    - action="models": List available models
    - action="generate": Generate text with Ollama
    - action="pull": Download/pull a model
    - action="hardware": Get hardware info and recommended settings
    - action="docs": Generate code documentation
    - action="quality": Analyze code quality
    - action="summary": Enhance context summary

    Returns JSON string (FastMCP requirement).
    """
    try:
        if action == "status":
            from .ollama_integration import check_ollama_status
            return check_ollama_status(host)

        elif action == "models":
            from .ollama_integration import list_ollama_models
            return list_ollama_models(host)

        elif action == "generate":
            from .ollama_integration import generate_with_ollama
            if not prompt:
                return json.dumps({"error": "prompt parameter required for generate action"}, indent=2)
            parsed_options = None
            if options:
                try:
                    parsed_options = json.loads(options)
                except json.JSONDecodeError:
                    pass
            return generate_with_ollama(
                prompt, model, host, stream, parsed_options,
                num_gpu=num_gpu, num_threads=num_threads, context_size=context_size
            )

        elif action == "pull":
            from .ollama_integration import pull_ollama_model
            if not model:
                return json.dumps({"error": "model parameter required for pull action"}, indent=2)
            return pull_ollama_model(model, host)

        elif action == "hardware":
            from .ollama_integration import get_hardware_info
            return get_hardware_info()

        elif action == "docs":
            from .ollama_enhanced_tools import generate_code_documentation
            if not file_path:
                return json.dumps({"error": "file_path parameter required for docs action"}, indent=2)
            return generate_code_documentation(file_path, output_path, style, model)

        elif action == "quality":
            from .ollama_enhanced_tools import analyze_code_quality
            if not file_path:
                return json.dumps({"error": "file_path parameter required for quality action"}, indent=2)
            return analyze_code_quality(file_path, include_suggestions, model)

        elif action == "summary":
            from .ollama_enhanced_tools import enhance_context_summary
            if not data:
                return json.dumps({"error": "data parameter required for summary action"}, indent=2)
            try:
                parsed_data = json.loads(data)
            except json.JSONDecodeError:
                parsed_data = data
            return enhance_context_summary(parsed_data, level, model)

        else:
            return json.dumps({
                "error": f"Unknown ollama action: {action}. Use 'status', 'models', 'generate', 'pull', 'hardware', 'docs', 'quality', or 'summary'."
            }, indent=2)

    except ImportError as e:
        return json.dumps({"error": f"Ollama tools not available: {e}"}, indent=2)
    except Exception as e:
        logger.error(f"Ollama tool error: {e}", exc_info=True)
        return json.dumps({"error": str(e)}, indent=2)



def mlx(
    action: str = "status",
    prompt: Optional[str] = None,
    model: str = "mlx-community/Phi-3.5-mini-instruct-4bit",
    max_tokens: int = 512,
    temperature: float = 0.7,
    verbose: bool = False,
) -> str:
    """
    [HINT: MLX. action=status|hardware|models|generate. Unified MLX tool.]

    Unified MLX tool for Apple Silicon GPU acceleration.

    Actions:
    - action="status": Check if MLX is available
    - action="hardware": Get hardware info and recommended settings
    - action="models": List recommended MLX models
    - action="generate": Generate text with MLX

    Returns JSON string (FastMCP requirement).
    """
    try:
        if action == "status":
            from .mlx_integration import check_mlx_status
            return check_mlx_status()

        elif action == "hardware":
            from .mlx_integration import get_mlx_hardware_info
            return get_mlx_hardware_info()

        elif action == "models":
            from .mlx_integration import list_mlx_models
            return list_mlx_models()

        elif action == "generate":
            from .mlx_integration import generate_with_mlx
            if not prompt:
                return json.dumps({"error": "prompt parameter required for generate action"}, indent=2)
            return generate_with_mlx(prompt, model, max_tokens, temperature, verbose)

        else:
            return json.dumps({
                "error": f"Unknown mlx action: {action}. Use 'status', 'hardware', 'models', or 'generate'."
            }, indent=2)

    except ImportError as e:
        return json.dumps({"error": f"MLX tools not available: {e}"}, indent=2)
    except Exception as e:
        logger.error(f"MLX tool error: {e}", exc_info=True)
        return json.dumps({"error": str(e)}, indent=2)



def coreml(
    action: str = "info",
    model_path: Optional[str] = None,
    input_data: Optional[str] = None,  # JSON string
    compute_units: str = "all",
    model_dir: Optional[str] = None,
) -> str:
    """
    [HINT: Core ML. action=info|list|predict. Unified Core ML tool with Neural Engine support.]

    Unified Core ML tool for Neural Engine (NPU) access on Apple Silicon.

    Actions:
    - action="info": Get Core ML hardware info and Neural Engine support
    - action="list": List available Core ML models
    - action="predict": Run inference with Core ML model (uses Neural Engine when available)

    📊 Output: Hardware info, model list, or predictions
    🔧 Side Effects: Loads and runs Core ML models
    ⚡ Neural Engine: Automatically used when models support ANE operations

    Returns JSON string (FastMCP requirement).
    """
    try:
        if action == "info":
            from .coreml_integration import get_coreml_hardware_info
            return get_coreml_hardware_info()

        elif action == "list":
            from .coreml_integration import list_coreml_models
            return list_coreml_models(model_dir)

        elif action == "predict":
            from .coreml_integration import predict_with_coreml
            if not model_path:
                return json.dumps({"error": "model_path parameter required for predict action"}, indent=2)
            if not input_data:
                return json.dumps({"error": "input_data parameter required for predict action"}, indent=2)
            
            # Parse input_data JSON
            try:
                parsed_input = json.loads(input_data)
            except json.JSONDecodeError:
                return json.dumps({"error": "input_data must be valid JSON"}, indent=2)
            
            return predict_with_coreml(model_path, parsed_input, compute_units)

        else:
            return json.dumps({
                "error": f"Unknown coreml action: {action}. Use 'info', 'list', or 'predict'."
            }, indent=2)

    except ImportError as e:
        return json.dumps({"error": f"Core ML tools not available: {e}. Install with: uv pip install coremltools"}, indent=2)
    except Exception as e:
        logger.error(f"Core ML tool error: {e}", exc_info=True)
        return json.dumps({"error": str(e)}, indent=2)


