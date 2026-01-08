"""
Ollama-Enhanced Tools for Exarp PMA

Examples of how to integrate Ollama/CodeLlama into existing exarp_pma workflows.
These tools enhance existing functionality with LLM-powered analysis.
"""

from typing import Optional
import json
import logging
import time
from pathlib import Path

logger = logging.getLogger(__name__)

# Import Ollama integration
try:
    from .ollama_integration import generate_with_ollama
    OLLAMA_AVAILABLE = True
except ImportError:
    OLLAMA_AVAILABLE = False
    logger.warning("Ollama integration not available")

# Import error handler
try:
    from ..error_handler import (
        ErrorCode,
        format_error_response,
        format_success_response,
        log_automation_execution,
    )
except ImportError:
    def format_success_response(data, message=None):
        return {"success": True, "data": data, "timestamp": time.time()}

    def format_error_response(error, error_code, include_traceback=False):
        return {"success": False, "error": {"code": str(error_code), "message": str(error)}}

    def log_automation_execution(name, duration, success, error=None):
        logger.info(f"{name}: {duration:.2f}s, success={success}")

    class ErrorCode:
        AUTOMATION_ERROR = "AUTOMATION_ERROR"


def generate_code_documentation(
    file_path: str,
    output_path: Optional[str] = None,
    style: str = "google",
    model: str = "codellama",
) -> str:
    """
    [HINT: Code documentation. Generate comprehensive documentation for Python code using CodeLlama.]

    📊 Output: Generated documentation with docstrings
    🔧 Side Effects: Optionally writes documentation to file
    📁 Analyzes: Python code file
    ⏱️ Typical Runtime: 5-15 seconds

    Example Prompt:
    "Generate documentation for project_management_automation/server.py"

    Args:
        file_path: Path to Python file to document
        output_path: Optional path to save generated documentation
        style: Documentation style (google, numpy, sphinx)
        model: Ollama model to use (default: codellama)

    Returns:
        JSON with generated documentation
    """
    start_time = time.time()

    if not OLLAMA_AVAILABLE:
        error_response = format_error_response(
            "Ollama integration not available. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)

    try:
        file = Path(file_path)
        if not file.exists():
            error_response = format_error_response(
                f"File not found: {file_path}",
                ErrorCode.AUTOMATION_ERROR
            )
            return json.dumps(error_response, indent=2)

        code = file.read_text()

        prompt = f"""Generate comprehensive documentation for this Python code.
Use {style} docstring style.

Requirements:
1. Module-level docstring explaining the file's purpose
2. Function/class docstrings with:
   - Clear description
   - Parameters (Args section)
   - Returns section
   - Raises section (if applicable)
   - Examples section (if helpful)
3. Inline comments for complex logic
4. Type hints where appropriate

Code:
{code}

Generate the documented version of this code.
"""

        # Use optimized settings for faster generation
        docs_result = generate_with_ollama(
            prompt,
            model=model,
            stream=True,  # Enable streaming for faster perceived response
        )
        docs_data = json.loads(docs_result)

        if not docs_data.get("success"):
            return docs_result

        documented_code = docs_data.get("data", {}).get("response", "")

        result = {
            "file_path": str(file.absolute()),
            "style": style,
            "documentation": documented_code,
            "original_length": len(code),
            "documented_length": len(documented_code),
        }

        if output_path:
            output_file = Path(output_path)
            output_file.write_text(documented_code)
            result["output_path"] = str(output_file.absolute())

        duration = time.time() - start_time
        log_automation_execution("generate_code_documentation", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("generate_code_documentation", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def analyze_code_quality(
    file_path: str,
    include_suggestions: bool = True,
    model: str = "codellama",
) -> str:
    """
    [HINT: Code quality analysis. Analyze Python code quality using CodeLlama.]

    📊 Output: Quality score, issues, suggestions
    🔧 Side Effects: None (read-only analysis)
    📁 Analyzes: Python code file
    ⏱️ Typical Runtime: 5-15 seconds

    Example Prompt:
    "Analyze code quality of project_management_automation/server.py"

    Args:
        file_path: Path to Python file to analyze
        include_suggestions: Whether to include improvement suggestions
        model: Ollama model to use (default: codellama)

    Returns:
        JSON with quality analysis
    """
    start_time = time.time()

    if not OLLAMA_AVAILABLE:
        error_response = format_error_response(
            "Ollama integration not available. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)

    try:
        file = Path(file_path)
        if not file.exists():
            error_response = format_error_response(
                f"File not found: {file_path}",
                ErrorCode.AUTOMATION_ERROR
            )
            return json.dumps(error_response, indent=2)

        code = file.read_text()

        prompt = f"""Analyze this Python code for quality and provide a structured assessment.

Provide:
1. Overall quality score (0-100) with brief justification
2. Code smells detected (list specific issues)
3. Performance issues (if any)
4. Security concerns (if any)
5. Best practice violations
6. Code maintainability assessment

{"7. Specific refactoring suggestions" if include_suggestions else ""}

Format your response as JSON with these keys:
- quality_score (number)
- code_smells (array of strings)
- performance_issues (array of strings)
- security_concerns (array of strings)
- best_practice_violations (array of strings)
- maintainability (string: "excellent" | "good" | "fair" | "poor")
- suggestions (array of strings, if include_suggestions is true)

Code:
{code}
"""

        # Use optimized settings for faster generation
        analysis_result = generate_with_ollama(
            prompt,
            model=model,
            stream=True,  # Enable streaming for faster perceived response
        )
        analysis_data = json.loads(analysis_result)

        if not analysis_data.get("success"):
            return analysis_result

        # Try to parse the LLM response as JSON
        response_text = analysis_data.get("data", {}).get("response", "")
        try:
            # Extract JSON from response if it's wrapped in markdown
            if "```json" in response_text:
                json_start = response_text.find("```json") + 7
                json_end = response_text.find("```", json_start)
                response_text = response_text[json_start:json_end].strip()
            elif "```" in response_text:
                json_start = response_text.find("```") + 3
                json_end = response_text.find("```", json_start)
                response_text = response_text[json_start:json_end].strip()

            quality_data = json.loads(response_text)
        except json.JSONDecodeError:
            # If JSON parsing fails, return the raw response
            quality_data = {"raw_analysis": response_text}

        result = {
            "file_path": str(file.absolute()),
            "analysis": quality_data,
            "timestamp": time.time(),
        }

        duration = time.time() - start_time
        log_automation_execution("analyze_code_quality", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("analyze_code_quality", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def enhance_context_summary(
    data: str | dict | list,
    level: str = "brief",
    model: str = "codellama",
) -> str:
    """
    [HINT: Enhanced summarization. Use CodeLlama to create intelligent summaries of tool outputs.]

    📊 Output: LLM-enhanced summary
    🔧 Side Effects: None
    📁 Analyzes: Tool output data
    ⏱️ Typical Runtime: 3-10 seconds

    Example Prompt:
    "Summarize this project health report with insights"

    Args:
        data: Data to summarize (JSON string, dict, or list)
        level: Summary level (brief, detailed, actionable)
        model: Ollama model to use (default: codellama)

    Returns:
        JSON with enhanced summary
    """
    start_time = time.time()

    if not OLLAMA_AVAILABLE:
        error_response = format_error_response(
            "Ollama integration not available. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)

    try:
        # Convert data to string if needed
        if isinstance(data, (dict, list)):
            data_str = json.dumps(data, indent=2)
        else:
            data_str = str(data)

        prompt = f"""Summarize this project management data in a {level} format.

Focus on:
- Key metrics and numbers
- Actionable items and recommendations
- Critical issues that need attention
- Next steps or follow-up actions

Provide a concise, well-structured summary that highlights the most important information.

Data:
{data_str}
"""

        # Use optimized settings for faster generation
        summary_result = generate_with_ollama(
            prompt,
            model=model,
            stream=True,  # Enable streaming for faster perceived response
        )
        summary_data = json.loads(summary_result)

        if not summary_data.get("success"):
            return summary_result

        result = {
            "level": level,
            "summary": summary_data.get("data", {}).get("response", ""),
            "original_length": len(data_str),
            "summary_length": len(summary_data.get("data", {}).get("response", "")),
        }

        duration = time.time() - start_time
        log_automation_execution("enhance_context_summary", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("enhance_context_summary", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def register_ollama_enhanced_tools(mcp):
    """
    Register Ollama-enhanced tools with FastMCP server.

    Args:
        mcp: FastMCP server instance
    """
    if not OLLAMA_AVAILABLE:
        logger.warning("Ollama-enhanced tools not registered: ollama package not available")
        return

    try:
        @mcp.tool()
        def generate_code_documentation_tool(
            file_path: str,
            output_path: Optional[str] = None,
            style: str = "google",
            model: str = "codellama",
        ) -> str:
            """Generate comprehensive documentation for Python code using CodeLlama."""
            return generate_code_documentation(file_path, output_path, style, model)

        @mcp.tool()
        def analyze_code_quality_tool(
            file_path: str,
            include_suggestions: bool = True,
            model: str = "codellama",
        ) -> str:
            """Analyze Python code quality using CodeLlama."""
            return analyze_code_quality(file_path, include_suggestions, model)

        @mcp.tool()
        def enhance_context_summary_tool(
            data: str,  # JSON string
            level: str = "brief",
            model: str = "codellama",
        ) -> str:
            """Use CodeLlama to create intelligent summaries of tool outputs."""
            # Parse JSON string
            try:
                parsed_data = json.loads(data)
            except json.JSONDecodeError:
                parsed_data = data

            return enhance_context_summary(parsed_data, level, model)

        logger.info("✅ Ollama-enhanced tools registered")

    except Exception as e:
        logger.error(f"Failed to register Ollama-enhanced tools: {e}", exc_info=True)

