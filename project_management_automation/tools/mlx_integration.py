"""
MLX Integration Tools

Provides tools for interacting with local MLX models on Apple Silicon:
- List available models
- Generate text with MLX
- Check MLX/Metal availability
- Load models from Hugging Face
"""

import json
import logging
import os
import platform
import subprocess
import time
from typing import Any, TYPE_CHECKING

if TYPE_CHECKING:
    pass

logger = logging.getLogger(__name__)

def get_system_ram_gb() -> float:
    """
    Get total system RAM in GB.

    Returns:
        Total RAM in GB, or 8.0 as default if detection fails
    """
    try:
        system = platform.system()
        if system == "Darwin":  # macOS
            result = subprocess.check_output(
                ["sysctl", "-n", "hw.memsize"],
                stderr=subprocess.DEVNULL,
                text=True,
                timeout=1,
            )
            ram_bytes = int(result.strip())
            ram_gb = ram_bytes / (1024 ** 3)
            return ram_gb
        elif system == "Linux":
            with open("/proc/meminfo") as f:
                for line in f:
                    if line.startswith("MemTotal:"):
                        parts = line.split()
                        ram_kb = int(parts[1])
                        ram_gb = ram_kb / (1024 ** 2)
                        return ram_gb
        return 8.0
    except Exception:
        return 8.0

# Try to import MLX, handle gracefully if not available
try:
    import mlx.core as mx
    from mlx_lm import generate, load
    MLX_AVAILABLE = True
except ImportError:
    MLX_AVAILABLE = False
    logger.warning("MLX package not available. Install with: uv sync")

# Model cache: store loaded models to avoid reloading
_MLX_MODEL_CACHE: dict[str, tuple[Any, Any]] = {}  # model_name -> (model_obj, tokenizer)

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


def is_apple_silicon() -> bool:
    """Check if running on Apple Silicon."""
    return platform.machine() == "arm64" and platform.system() == "Darwin"


def check_metal_available() -> bool:
    """Check if Metal GPU is available."""
    if not MLX_AVAILABLE:
        return False
    try:
        return mx.metal.is_available()
    except Exception:
        return False


def detect_hardware_config() -> dict[str, Any]:
    """
    Detect hardware platform and return optimal MLX configuration.

    MLX only works on Apple Silicon, so this is focused on that platform.

    Returns:
        Dict with hardware info and recommended settings
    """
    system = platform.system()
    arch = platform.machine()
    cpu_cores = os.cpu_count() or 4
    ram_gb = get_system_ram_gb()

    config = {
        "architecture": arch,
        "cpu_cores": cpu_cores,
        "ram_gb": ram_gb,
        "mlx_supported": False,
        "metal_available": False,
        "platform": "unknown",
        "recommended_model_size": "medium",
        "recommended_context_size": 4096,
        "notes": [],
    }

    # MLX only works on Apple Silicon
    if system == "Darwin" and arch == "arm64":
        config["platform"] = "apple_silicon"
        config["mlx_supported"] = True
        config["metal_available"] = check_metal_available()

        # Try to detect chip model
        try:
            chip_model = subprocess.check_output(
                ["sysctl", "-n", "machdep.cpu.brand_string"],
                stderr=subprocess.DEVNULL,
                text=True,
                timeout=1,
            ).strip().upper()

            if "M4" in chip_model:
                config["chip_model"] = "M4"
                config["recommended_context_size"] = 8192
                config["recommended_model_size"] = "large"
            elif "M3" in chip_model:
                config["chip_model"] = "M3"
                config["recommended_context_size"] = 8192
                config["recommended_model_size"] = "large"
            elif "M2" in chip_model:
                config["chip_model"] = "M2"
                config["recommended_context_size"] = 6144
                config["recommended_model_size"] = "medium"
            elif "M1" in chip_model:
                config["chip_model"] = "M1"
                config["recommended_context_size"] = 4096
                config["recommended_model_size"] = "medium"
            else:
                config["chip_model"] = "Apple Silicon (unknown)"
                config["recommended_context_size"] = 4096
        except Exception:
            config["chip_model"] = "Apple Silicon (detected)"

        # RAM-based recommendations
        if ram_gb >= 32:
            config["recommended_model_size"] = "large"
            config["recommended_context_size"] = 8192
        elif ram_gb >= 16:
            config["recommended_model_size"] = "medium"
            config["recommended_context_size"] = 6144
        elif ram_gb >= 8:
            config["recommended_model_size"] = "small"
            config["recommended_context_size"] = 4096
        else:
            config["recommended_model_size"] = "tiny"
            config["recommended_context_size"] = 2048
            config["notes"].append("Limited RAM - consider smaller models")
    else:
        config["notes"].append("MLX only works on Apple Silicon (arm64 macOS)")
        if system != "Darwin":
            config["notes"].append(f"Current platform: {system}")
        if arch != "arm64":
            config["notes"].append(f"Current architecture: {arch}")

    return config


def check_mlx_status() -> str:
    """
    [HINT: MLX status. Check if MLX is available and Metal GPU is accessible.]

    Check MLX availability and Metal GPU status.

    📁 Checks: MLX installation, Metal GPU availability

    Use cases:
    "Is MLX available? Check MLX status"

    Returns:
        JSON with MLX and Metal status
    """
    start_time = time.time()

    if not MLX_AVAILABLE:
        error_response = format_error_response(
            "MLX package not installed. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR,
        )
        error_response["available"] = False
        error_response["metal_available"] = False
        return json.dumps(error_response, indent=2)

    try:
        metal_available = check_metal_available()
        is_apple = is_apple_silicon()
        hw_config = detect_hardware_config()

        result = format_success_response({
            "mlx_available": True,
            "metal_available": metal_available,
            "apple_silicon": is_apple,
            "platform": hw_config.get("platform", "unknown"),
            "chip_model": hw_config.get("chip_model", "unknown"),
            "architecture": hw_config.get("architecture", "unknown"),
            "cpu_cores": hw_config.get("cpu_cores", 0),
            "ram_gb": hw_config.get("ram_gb", 0),
            "mlx_supported": hw_config.get("mlx_supported", False),
            "recommended_model_size": hw_config.get("recommended_model_size", "medium"),
            "recommended_context_size": hw_config.get("recommended_context_size", 4096),
            "notes": hw_config.get("notes", []),
        }, "MLX status checked successfully")

        duration = time.time() - start_time
        log_automation_execution("check_mlx_status", duration, True)

        return json.dumps(result, indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("check_mlx_status", duration, False, e)
        error_response = format_error_response(
            f"Error checking MLX status: {str(e)}",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)


def list_mlx_models() -> str:
    """
    [HINT: MLX models. List recommended MLX models available on Hugging Face.]

    List recommended MLX models available for download.

    Note: This lists recommended models, not locally installed ones.
    MLX models are downloaded from Hugging Face on first use.

    📁 Lists: MLX model recommendations from Hugging Face

    Use cases:
    "What MLX models are available?"

    Returns:
        JSON with recommended MLX models
    """
    start_time = time.time()

    if not MLX_AVAILABLE:
        error_response = format_error_response(
            "MLX package not installed. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)

    try:
        # Recommended MLX models from Hugging Face mlx-community
        # Verified model names as of 2025-01-25
        recommended_models = {
            "code": [
                {
                    "id": "mlx-community/CodeLlama-7b-mlx",
                    "size": "7B",
                    "best_for": "Code analysis, documentation, code generation",
                    "size_gb": "~4GB",
                },
                {
                    "id": "mlx-community/CodeLlama-7b-Python-mlx",
                    "size": "7B",
                    "best_for": "Python-focused code generation",
                    "size_gb": "~4GB",
                },
            ],
            "general": [
                {
                    "id": "mlx-community/Meta-Llama-3.1-8B-Instruct-bf16",
                    "size": "8B",
                    "best_for": "General purpose, high quality responses (full precision)",
                    "size_gb": "~16GB",
                },
                {
                    "id": "mlx-community/Meta-Llama-3.1-8B-Instruct-4bit",
                    "size": "8B",
                    "best_for": "General purpose, high quality responses (quantized)",
                    "size_gb": "~4.5GB",
                },
                {
                    "id": "mlx-community/Phi-3.5-mini-instruct-4bit",
                    "size": "3.8B",
                    "best_for": "Fast, efficient general tasks",
                    "size_gb": "~2.3GB",
                },
                {
                    "id": "mlx-community/Mistral-7B-Instruct-v0.2",
                    "size": "7B",
                    "best_for": "Fast inference, good balance",
                    "size_gb": "~4GB",
                },
            ],
            "small": [
                {
                    "id": "mlx-community/TinyLlama-1.1B-Chat-v1.0-mlx",
                    "size": "1.1B",
                    "best_for": "Very fast tasks, low memory",
                    "size_gb": "~0.7GB",
                },
                {
                    "id": "mlx-community/Phi-3-mini-128k-instruct-4bit",
                    "size": "3.8B",
                    "best_for": "Fast with long context (128k)",
                    "size_gb": "~2.3GB",
                },
            ],
        }

        result = format_success_response({
            "models": recommended_models,
            "total_count": sum(len(models) for models in recommended_models.values()),
            "note": "Models are downloaded from Hugging Face on first use",
            "tip": "Use generate_with_mlx to generate text with a model",
        }, "MLX models listed successfully")

        duration = time.time() - start_time
        log_automation_execution("list_mlx_models", duration, True)

        return json.dumps(result, indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("list_mlx_models", duration, False, e)
        error_response = format_error_response(
            f"Error listing MLX models: {str(e)}",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)


def generate_with_mlx(
    prompt: str,
    model: str = "mlx-community/Phi-3.5-mini-instruct-4bit",
    max_tokens: int = 512,
    temperature: float = 0.7,
    verbose: bool = False,
) -> str:
    """
    [HINT: MLX generation. Generate text using a local MLX model on Apple Silicon.]

    Generate text using a local MLX model.

    ⚠️ Requirements: Apple Silicon (M1/M2/M3/M4) Mac
    🔧 Side Effects: Downloads model on first use, uses GPU/CPU resources
    📁 Uses: Local MLX model or Hugging Face MLX models

    Use cases:
    "Generate a summary using MLX Phi-3.5 model"
    "Analyze this code with MLX CodeLlama"

    Args:
        prompt: Text prompt to send to the model
        model: Model identifier (Hugging Face repo ID or local path)
        max_tokens: Maximum tokens to generate
        temperature: Sampling temperature (0.0-1.0, higher = more creative)
        verbose: Print generation progress

    Returns:
        JSON with generated text
    """
    start_time = time.time()

    if not MLX_AVAILABLE:
        error_response = format_error_response(
            "MLX package not installed. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)

    if not is_apple_silicon():
        error_response = format_error_response(
            "MLX only works on Apple Silicon (M1/M2/M3/M4) Macs",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)

    try:
        # Check cache first
        if model in _MLX_MODEL_CACHE:
            if verbose:
                logger.info(f"Using cached MLX model: {model}")
            model_obj, tokenizer = _MLX_MODEL_CACHE[model]
        else:
            # Load model (will download from Hugging Face if not cached)
            if verbose:
                logger.info(f"Loading MLX model: {model} (first time, will cache)")

            model_obj, tokenizer = load(model)
            
            # Cache the model for reuse
            _MLX_MODEL_CACHE[model] = (model_obj, tokenizer)
            
            if verbose:
                logger.info("Model cached for future use")

        if verbose:
            logger.info(f"Generating with max_tokens={max_tokens}, temperature={temperature}")

        # Generate text
        # Note: mlx_lm.generate() accepts max_tokens but temperature control
        # is not available in the current mlx_lm API. Temperature parameter
        # is accepted but ignored (documented for API compatibility).
        response = generate(
            model_obj,
            tokenizer,
            prompt=prompt,
            max_tokens=max_tokens,
            verbose=verbose,
        )

        result = format_success_response({
            "generated_text": response,
            "model": model,
            "prompt": prompt,
            "max_tokens": max_tokens,
            "temperature": temperature,
        }, "Text generated successfully with MLX")

        duration = time.time() - start_time
        log_automation_execution("generate_with_mlx", duration, True)

        return json.dumps(result, indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("generate_with_mlx", duration, False, e)
        error_response = format_error_response(
            f"Error generating with MLX: {str(e)}",
            ErrorCode.AUTOMATION_ERROR,
        )
        if "not found" in str(e).lower() or "model" in str(e).lower():
            error_response["tip"] = f"Model '{model}' may need to be downloaded. Try a recommended model from list_mlx_models."
        return json.dumps(error_response, indent=2)


def get_mlx_hardware_info() -> str:
    """
    [HINT: Hardware detection. Get hardware information and recommended MLX settings.]

    Get hardware information and recommended MLX configuration.

    ⚠️ Requirements: Apple Silicon (M1/M2/M3/M4) Mac

    Use cases:
    "What hardware do I have? Check hardware for MLX"

    Returns:
        JSON with hardware information and recommended MLX settings
    """
    start_time = time.time()

    try:
        hw_config = detect_hardware_config()

        result = format_success_response(hw_config, "Hardware information retrieved successfully")

        duration = time.time() - start_time
        log_automation_execution("get_mlx_hardware_info", duration, True)

        return json.dumps(result, indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("get_mlx_hardware_info", duration, False, e)
        error_response = format_error_response(
            f"Error getting hardware info: {str(e)}",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)


def register_mlx_tools(mcp):
    """
    Register MLX tools with FastMCP server.

    Args:
        mcp: FastMCP server instance
    """
    if not MLX_AVAILABLE:
        logger.warning("MLX tools not registered: MLX package not available")
        return

    if not is_apple_silicon():
        logger.debug("MLX tools not registered: Not running on Apple Silicon")
        return

    try:
        @mcp.tool()
        def check_mlx_status_tool() -> str:
            """Check if MLX is available and Metal GPU is accessible."""
            return check_mlx_status()

        @mcp.tool()
        def get_mlx_hardware_info_tool() -> str:
            """Get hardware information and recommended MLX performance settings."""
            return get_mlx_hardware_info()

        @mcp.tool()
        def list_mlx_models_tool() -> str:
            """List recommended MLX models available for download."""
            return list_mlx_models()

        @mcp.tool()
        def generate_with_mlx_tool(
            prompt: str,
            model: str = "mlx-community/Phi-3.5-mini-instruct-4bit",
            max_tokens: int = 512,
            temperature: float = 0.7,
            verbose: bool = False,
        ) -> str:
            """
            Generate text using a local MLX model.

            Args:
                prompt: Text prompt to send to the model
                model: Model identifier (default: mlx-community/Phi-3.5-mini-instruct-4bit)
                max_tokens: Maximum tokens to generate (default: 512)
                temperature: Sampling temperature 0.0-1.0 (default: 0.7)
                verbose: Print generation progress (default: False)
            """
            return generate_with_mlx(prompt, model, max_tokens, temperature, verbose)

        logger.info("✅ MLX tools registered")

    except Exception as e:
        logger.error(f"Failed to register MLX tools: {e}", exc_info=True)

