"""
Test Case Suggestions Tool

Suggests test cases based on code analysis.
Analyzes function signatures, identifies edge cases, and generates test templates.

Enhanced with MLX CodeLlama for AI-powered test code generation.
"""

import ast
import json
import logging
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Import error handler
try:
    from ..error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
except ImportError:
    import sys
    from pathlib import Path
    server_dir = Path(__file__).parent.parent
    sys.path.insert(0, str(server_dir))
    try:
        from error_handler import ErrorCode, format_error_response, format_success_response, log_automation_execution
    except ImportError:
        import time
        def format_success_response(data, message=None):
            return {"success": True, "data": data, "timestamp": time.time()}
        def format_error_response(error, error_code, include_traceback=False):
            return {"success": False, "error": {"code": str(error_code), "message": str(error)}}
        def log_automation_execution(name, duration, success, error=None):
            logger.info(f"{name}: {duration:.2f}s, success={success}")
        class ErrorCode:
            AUTOMATION_ERROR = "AUTOMATION_ERROR"


def suggest_test_cases(
    target_file: Optional[str] = None,
    test_framework: str = "pytest",
    min_confidence: float = 0.7,
    output_path: Optional[str] = None,
) -> dict[str, Any]:
    """
    Suggest test cases based on code analysis.

    Args:
        target_file: File to analyze (optional - analyzes all Python files if not provided)
        test_framework: Framework for suggestions (default: pytest)
        min_confidence: Minimum confidence threshold (default: 0.7)
        output_path: Path for suggestions output (optional)

    Returns:
        Dictionary with suggested test cases
    """
    import time
    start_time = time.time()

    try:
        from ..utils import find_project_root

        project_root = find_project_root()
        suggestions = []

        if target_file:
            # Analyze specific file
            file_path = Path(target_file)
            if not file_path.is_absolute():
                file_path = project_root / file_path

            if file_path.exists() and file_path.suffix == ".py":
                suggestions.extend(_analyze_file(file_path, test_framework, min_confidence))
        else:
            # Analyze all Python files in project
            for py_file in project_root.rglob("*.py"):
                # Skip test files, venv, and hidden directories
                if "test" in str(py_file).lower() or "venv" in str(py_file) or py_file.name.startswith("."):
                    continue
                if py_file.parent.name.startswith("."):
                    continue

                file_suggestions = _analyze_file(py_file, test_framework, min_confidence)
                if file_suggestions:
                    suggestions.extend(file_suggestions)

        # Filter by confidence
        filtered_suggestions = [
            s for s in suggestions
            if s.get("confidence", 0.0) >= min_confidence
        ]

        # Sort by confidence (highest first)
        filtered_suggestions.sort(key=lambda x: x.get("confidence", 0.0), reverse=True)

        result = {
            "suggestions_count": len(filtered_suggestions),
            "total_analyzed": len(suggestions),
            "min_confidence": min_confidence,
            "framework": test_framework,
            "suggestions": filtered_suggestions[:50],  # Top 50
        }

        # Save to file if requested
        if output_path:
            output_file = Path(output_path)
            if not output_file.is_absolute():
                output_file = project_root / output_file
            output_file.parent.mkdir(parents=True, exist_ok=True)

            with open(output_file, "w") as f:
                json.dump(result, f, indent=2)

            result["output_file"] = str(output_file)

        duration = time.time() - start_time
        log_automation_execution("suggest_test_cases", duration, True)

        return format_success_response(result)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("suggest_test_cases", duration, False, e)
        logger.error(f"Error suggesting test cases: {e}", exc_info=True)

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return error_response


def _analyze_file(file_path: Path, framework: str, min_confidence: float) -> list[dict[str, Any]]:
    """Analyze a Python file and suggest test cases."""
    suggestions = []

    try:
        with open(file_path, encoding="utf-8") as f:
            content = f.read()

        tree = ast.parse(content, filename=str(file_path))

        for node in ast.walk(tree):
            if isinstance(node, ast.FunctionDef):
                # Skip private/internal functions
                if node.name.startswith("_"):
                    continue

                # Analyze function signature
                func_suggestions = _analyze_function(node, file_path, framework)
                if func_suggestions:
                    suggestions.extend(func_suggestions)

    except SyntaxError:
        logger.debug(f"Skipping {file_path} - syntax error")
    except Exception as e:
        logger.debug(f"Error analyzing {file_path}: {e}")

    return suggestions


def _analyze_function(func_node: ast.FunctionDef, file_path: Path, framework: str) -> list[dict[str, Any]]:
    """Analyze a function and suggest test cases."""
    suggestions = []

    # Basic test case for function existence
    suggestions.append({
        "function": func_node.name,
        "file": str(file_path.relative_to(file_path.parents[len(file_path.parts) - 3])),
        "type": "basic",
        "confidence": 0.8,
        "test_name": f"test_{func_node.name}",
        "description": f"Test that {func_node.name} can be called",
        "framework": framework,
    })

    # Analyze parameters for edge cases
    args = func_node.args
    if args.args:
        # Suggest tests for None, empty, and boundary values
        for arg in args.args:
            arg_name = arg.arg
            if arg_name == "self":
                continue

            # Edge case: None
            suggestions.append({
                "function": func_node.name,
                "file": str(file_path.relative_to(file_path.parents[len(file_path.parts) - 3])),
                "type": "edge_case",
                "confidence": 0.75,
                "test_name": f"test_{func_node.name}_with_none_{arg_name}",
                "description": f"Test {func_node.name} with None for {arg_name}",
                "framework": framework,
                "parameter": arg_name,
            })

    # Check for return statements
    has_return = any(isinstance(node, ast.Return) for node in ast.walk(func_node))
    if has_return:
        suggestions.append({
            "function": func_node.name,
            "file": str(file_path.relative_to(file_path.parents[len(file_path.parts) - 3])),
            "type": "return_value",
            "confidence": 0.85,
            "test_name": f"test_{func_node.name}_return_value",
            "description": f"Test return value of {func_node.name}",
            "framework": framework,
        })

    return suggestions


def generate_test_code(
    target_file: Optional[str] = None,
    test_framework: str = "pytest",
    use_mlx: bool = True,
    use_coreml: bool = True,  # Try Core ML first (NPU acceleration)
    coreml_model_path: Optional[str] = None,  # Path to Core ML model
    model: str = "mlx-community/Phi-3.5-mini-instruct-4bit",  # MLX fallback model
    max_tokens: int = 512,
    temperature: float = 0.3,
    compute_units: str = "all",  # Core ML compute units (all, cpu_and_ane, etc.)
    output_path: Optional[str] = None,
) -> dict[str, Any]:
    """
    Generate actual test code using AI models (Core ML NPU or MLX GPU).

    Uses AI models to generate complete, runnable test code from function signatures.
    This is an enhancement over suggest_test_cases() which only provides templates.

    Priority order:
    1. Core ML model (if available) - Uses Neural Engine (NPU) for fastest inference
    2. MLX model - Uses Metal GPU acceleration
    3. Template fallback - Basic test templates

    Args:
        target_file: File to analyze (required for code generation)
        test_framework: Framework for test generation (default: pytest)
        use_coreml: Try Core ML first for NPU acceleration (default: True)
        coreml_model_path: Path to Core ML model file (.mlpackage or .mlmodel)
        use_mlx: Use MLX as fallback if Core ML unavailable (default: True)
        model: MLX model to use (default: Phi-3.5-mini-instruct-4bit)
        max_tokens: Maximum tokens to generate (default: 512)
        temperature: Generation temperature, lower = more deterministic (default: 0.3)
        compute_units: Core ML compute units (all, cpu_and_ane, cpu_and_gpu, cpu_only)
        output_path: Path to save generated test code (optional)

    Returns:
        Dictionary with generated test code
    """
    import time
    start_time = time.time()

    try:
        from ..utils import find_project_root

        if not target_file:
            return format_error_response(
                "target_file parameter required for test code generation",
                ErrorCode.AUTOMATION_ERROR
            )

        project_root = find_project_root()
        file_path = Path(target_file)
        if not file_path.is_absolute():
            file_path = project_root / file_path

        if not file_path.exists() or file_path.suffix != ".py":
            return format_error_response(
                f"File not found or not a Python file: {target_file}",
                ErrorCode.AUTOMATION_ERROR
            )

        # Extract functions from file
        functions = _extract_functions_with_code(file_path)
        if not functions:
            return format_error_response(
                f"No functions found in {target_file}",
                ErrorCode.AUTOMATION_ERROR
            )

        generated_tests = []

        # Generate test code for each function
        # Priority: Core ML (NPU) > MLX (GPU) > Template
        for func_info in functions:
            test_code = None
            method_used = "template"
            
            # Try Core ML first (NPU acceleration)
            if use_coreml and coreml_model_path:
                try:
                    test_code = _generate_test_with_coreml(
                        func_info=func_info,
                        test_framework=test_framework,
                        model_path=coreml_model_path,
                        compute_units=compute_units,
                        max_tokens=max_tokens,
                    )
                    if test_code:
                        method_used = "coreml_neural_engine"
                except Exception as e:
                    logger.debug(f"Core ML generation failed for {func_info['name']}: {e}")
            
            # Fallback to MLX (GPU acceleration)
            if not test_code and use_mlx:
                try:
                    test_code = _generate_test_with_mlx(
                        func_info=func_info,
                        test_framework=test_framework,
                        model=model,
                        max_tokens=max_tokens,
                        temperature=temperature,
                    )
                    if test_code and not test_code.startswith('{"success": false'):  # Check for error JSON
                        method_used = "mlx_gpu"
                    else:
                        test_code = None  # MLX returned error
                except Exception as e:
                    logger.debug(f"MLX generation failed for {func_info['name']}: {e}")
            
            # Final fallback to template
            if not test_code:
                test_code = _generate_test_template(func_info, test_framework)
                method_used = "template_fallback" if (use_coreml or use_mlx) else "template"
            
            generated_tests.append({
                "function": func_info["name"],
                "test_code": test_code,
                "framework": test_framework,
                "method": method_used,
            })

        # Combine all test code
        combined_test_code = "\n\n".join([
            f"# Test for {test['function']}\n{test['test_code']}"
            for test in generated_tests
        ])

        result = {
            "target_file": str(file_path.relative_to(project_root)),
            "functions_tested": len(generated_tests),
            "test_framework": test_framework,
            "generated_tests": generated_tests,
            "combined_test_code": combined_test_code,
            "method": "coreml_neural_engine" if (use_coreml and coreml_model_path) else ("mlx_gpu" if use_mlx else "template"),
        }

        # Save to file if requested
        if output_path:
            output_file = Path(output_path)
            if not output_file.is_absolute():
                output_file = project_root / output_file
            output_file.parent.mkdir(parents=True, exist_ok=True)

            with open(output_file, "w") as f:
                f.write(combined_test_code)

            result["output_file"] = str(output_file)

        duration = time.time() - start_time
        log_automation_execution("generate_test_code", duration, True)

        return format_success_response(result)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("generate_test_code", duration, False, e)
        logger.error(f"Error generating test code: {e}", exc_info=True)

        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return error_response


def _extract_functions_with_code(file_path: Path) -> list[dict[str, Any]]:
    """Extract functions with their code and signatures."""
    functions = []

    try:
        with open(file_path, encoding="utf-8") as f:
            content = f.read()

        tree = ast.parse(content, filename=str(file_path))

        for node in ast.walk(tree):
            if isinstance(node, ast.FunctionDef):
                # Skip private/internal functions
                if node.name.startswith("_"):
                    continue

                # Get function source code
                func_lines = content.split('\n')
                start_line = node.lineno - 1
                end_line = node.end_lineno if hasattr(node, 'end_lineno') else start_line + 10
                func_code = '\n'.join(func_lines[start_line:end_line])

                # Extract function signature
                args = []
                for arg in node.args.args:
                    if arg.arg != "self":
                        args.append(arg.arg)

                # Get docstring if available
                docstring = ast.get_docstring(node) or ""

                functions.append({
                    "name": node.name,
                    "code": func_code,
                    "signature": f"{node.name}({', '.join(args)})",
                    "args": args,
                    "docstring": docstring,
                    "line": node.lineno,
                })

    except SyntaxError:
        logger.debug(f"Skipping {file_path} - syntax error")
    except Exception as e:
        logger.debug(f"Error extracting functions from {file_path}: {e}")

    return functions


def _generate_test_with_mlx(
    func_info: dict[str, Any],
    test_framework: str = "pytest",
    model: str = "mlx-community/Phi-3.5-mini-instruct-4bit",  # Use Phi-3.5 as default (more reliable)
    max_tokens: int = 512,
    temperature: float = 0.3,
) -> Optional[str]:
    """Generate test code using MLX CodeLlama or Phi-3.5."""
    try:
        from .mlx_integration import generate_with_mlx

        # Create prompt for CodeLlama
        prompt = f"""<|begin_of_text|><|start_header_id|>system<|end_header_id|>

You are a Python testing expert. Generate comprehensive {test_framework} test code for the given function.

Requirements:
- Use {test_framework} framework
- Include edge cases (None, empty, boundary values)
- Test return values
- Include docstring explaining what the test does
- Make tests runnable and complete

<|eot_id|><|start_header_id|>user<|end_header_id|>

Generate a {test_framework} test for this function:

```python
{func_info['code']}
```

Function signature: {func_info['signature']}
Docstring: {func_info['docstring']}

Generate the test code:<|eot_id|><|start_header_id|>assistant<|end_header_id|>

```python
"""

        # Generate test code
        test_code = generate_with_mlx(
            prompt=prompt,
            model=model,
            max_tokens=max_tokens,
            temperature=temperature,
            verbose=False,
        )

        # Clean up generated code (remove prompt artifacts)
        if test_code:
            # Extract just the Python code block if present
            if "```python" in test_code:
                parts = test_code.split("```python")
                if len(parts) > 1:
                    code_part = parts[1].split("```")[0]
                    test_code = code_part.strip()
            elif "```" in test_code:
                parts = test_code.split("```")
                if len(parts) > 1:
                    test_code = parts[1].strip()
                    if test_code.startswith("python"):
                        test_code = test_code[6:].strip()

            return test_code.strip()

        return None

    except ImportError:
        logger.warning("MLX not available for test generation")
        return None
    except Exception as e:
        logger.debug(f"MLX test generation error: {e}")
        return None


def _generate_test_with_coreml(
    func_info: dict[str, Any],
    test_framework: str = "pytest",
    model_path: str = "",
    compute_units: str = "all",
    max_tokens: int = 512,
) -> Optional[str]:
    """Generate test code using Core ML model (Neural Engine acceleration)."""
    try:
        from .coreml_integration import predict_with_coreml
        
        # Create prompt for test generation
        prompt = f"""Generate a comprehensive {test_framework} test for this function:

```python
{func_info['code']}
```

Function signature: {func_info['signature']}
Docstring: {func_info['docstring']}

Requirements:
- Use {test_framework} framework
- Include edge cases (None, empty, boundary values)
- Test return values
- Include docstring
- Make tests runnable and complete

Generate the test code:"""

        # Prepare input for Core ML model
        # Note: The exact input format depends on the Core ML model's specification
        # This is a generic approach - may need adjustment for specific models
        input_data = {
            "prompt": prompt,
            "text": prompt,  # Some models use "text" instead of "prompt"
            "max_tokens": max_tokens,
        }
        
        # Run Core ML prediction
        result_json = predict_with_coreml(
            model_path=model_path,
            input_data=input_data,
            compute_units=compute_units,
        )
        
        result = json.loads(result_json)
        
        if not result.get("success"):
            logger.debug(f"Core ML prediction failed: {result.get('error', {}).get('message', 'Unknown error')}")
            return None
        
        # Extract generated text from predictions
        predictions = result.get("data", {}).get("predictions", {})
        
        # Try different possible output keys
        test_code = (
            predictions.get("generated_text") or
            predictions.get("text") or
            predictions.get("output") or
            predictions.get("completion") or
            str(predictions)  # Fallback to string representation
        )
        
        if test_code and isinstance(test_code, str):
            # Clean up generated code
            if "```python" in test_code:
                parts = test_code.split("```python")
                if len(parts) > 1:
                    code_part = parts[1].split("```")[0]
                    test_code = code_part.strip()
            elif "```" in test_code:
                parts = test_code.split("```")
                if len(parts) > 1:
                    test_code = parts[1].strip()
                    if test_code.startswith("python"):
                        test_code = test_code[6:].strip()
            
            return test_code.strip()
        
        return None
        
    except ImportError:
        logger.debug("Core ML not available for test generation")
        return None
    except Exception as e:
        logger.debug(f"Core ML test generation error: {e}")
        return None


def _generate_test_template(
    func_info: dict[str, Any],
    test_framework: str = "pytest",
) -> str:
    """Generate a basic test template (fallback when MLX unavailable)."""
    func_name = func_info["name"]
    args = func_info["args"]

    if test_framework == "pytest":
        test_code = f'''def test_{func_name}():
    """Test {func_name} function."""
    # TODO: Import the function
    # from module import {func_name}
    
    # Test basic functionality
    # result = {func_name}({', '.join(['None' if arg else '' for arg in args])})
    # assert result is not None
    pass
'''
    else:  # unittest
        test_code = f'''class Test{func_name.title()}(unittest.TestCase):
    """Test {func_name} function."""
    
    def test_{func_name}_basic(self):
        """Test basic functionality."""
        # TODO: Import and test
        pass
'''

    return test_code
