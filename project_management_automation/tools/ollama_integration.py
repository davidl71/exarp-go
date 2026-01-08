"""
Ollama Integration Tools

Provides tools for interacting with local Ollama models:
- List available models
- Generate text with Ollama
- Check Ollama server status
- Pull/download models
"""

import json
import logging
import os
import platform
import subprocess
import time
from typing import Any, Optional

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
            # Use sysctl to get total memory
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
            # Read from /proc/meminfo
            with open("/proc/meminfo") as f:
                for line in f:
                    if line.startswith("MemTotal:"):
                        parts = line.split()
                        ram_kb = int(parts[1])
                        ram_gb = ram_kb / (1024 ** 2)
                        return ram_gb
        # Windows or unknown - return default
        return 8.0
    except Exception:
        return 8.0  # Default fallback

# Try to import ollama, handle gracefully if not available
try:
    import ollama
    OLLAMA_AVAILABLE = True
except ImportError:
    OLLAMA_AVAILABLE = False
    logger.warning("Ollama package not available. Install with: uv sync")

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


def detect_hardware_config() -> dict[str, Any]:
    """
    Detect hardware platform and return optimal Ollama configuration.

    Detects:
    - CPU architecture (Intel x86_64 vs Apple Silicon arm64)
    - CPU core count
    - Total system RAM
    - GPU availability (Metal on Apple Silicon, CUDA on NVIDIA, ROCm on AMD)
    - Optimal performance settings based on available resources

    Returns:
        Dict with hardware info and recommended settings:
        {
            "platform": "apple_silicon" | "intel" | "linux" | "windows",
            "architecture": "arm64" | "x86_64",
            "cpu_cores": int,
            "ram_gb": float,
            "gpu_available": bool,
            "gpu_type": "metal" | "cuda" | "rocm" | None,
            "recommended_num_threads": int,
            "recommended_num_gpu": Optional[int],
            "recommended_context_size": int,
            "ram_optimizations": dict with RAM-based optimization recommendations,
        }
    """
    system = platform.system()
    arch = platform.machine()
    cpu_cores = os.cpu_count() or 4  # Fallback to 4 if detection fails
    ram_gb = get_system_ram_gb()

    config = {
        "architecture": arch,
        "cpu_cores": cpu_cores,
        "ram_gb": ram_gb,
        "gpu_available": False,
        "gpu_type": None,
        "recommended_num_threads": max(1, cpu_cores - 1),  # Leave one core free
        "recommended_num_gpu": None,
        "recommended_context_size": 4096,  # Default
        "ram_optimizations": {},
    }

    # RAM-based optimizations
    # With more RAM, we can use larger context windows and higher quality models
    if ram_gb >= 32:
        # 32GB+ RAM: Can use very large contexts
        config["ram_optimizations"] = {
            "max_context_size": 16384,
            "recommended_context_size": 8192,
            "can_use_larger_models": True,
            "enable_flash_attention": True,
        }
    elif ram_gb >= 16:
        # 16-32GB RAM: Can use large contexts
        config["ram_optimizations"] = {
            "max_context_size": 8192,
            "recommended_context_size": 6144,
            "can_use_larger_models": True,
            "enable_flash_attention": True,
        }
    elif ram_gb >= 8:
        # 8-16GB RAM: Medium contexts
        config["ram_optimizations"] = {
            "max_context_size": 4096,
            "recommended_context_size": 4096,
            "can_use_larger_models": False,
            "enable_flash_attention": True,
        }
    else:
        # <8GB RAM: Smaller contexts
        config["ram_optimizations"] = {
            "max_context_size": 2048,
            "recommended_context_size": 2048,
            "can_use_larger_models": False,
            "enable_flash_attention": False,
        }

    # macOS detection
    if system == "Darwin":
        if arch in ("arm64", "aarch64"):
            config["platform"] = "apple_silicon"
            config["gpu_available"] = True
            config["gpu_type"] = "metal"
            # Apple Silicon: Metal GPU acceleration available
            # Recommend using GPU for most layers
            # Common Apple Silicon chips: M1 (8-core), M1 Pro/Max (8-10 core GPU), M2 (8-10 core GPU), M3/M4 (up to 40 GPU cores)

            # Try to detect chip model for better optimization
            try:
                chip_model = subprocess.check_output(
                    ["sysctl", "-n", "machdep.cpu.brand_string"],
                    stderr=subprocess.DEVNULL,
                    text=True,
                    timeout=1,
                ).strip().upper()

                # Detect M-series chip and estimate GPU cores
                if "M4" in chip_model:
                    # M4 has up to 40 GPU cores (depending on model)
                    config["recommended_num_gpu"] = 40
                    config["recommended_context_size"] = 8192
                    config["chip_model"] = "M4"
                elif "M3" in chip_model:
                    # M3 has 14-20 GPU cores (depending on model)
                    config["recommended_num_gpu"] = 40
                    config["recommended_context_size"] = 8192
                    config["chip_model"] = "M3"
                elif "M2" in chip_model:
                    # M2 has 8-10 GPU cores
                    config["recommended_num_gpu"] = 35
                    config["recommended_context_size"] = 6144
                    config["chip_model"] = "M2"
                elif "M1" in chip_model:
                    # M1 has 7-8 GPU cores
                    config["recommended_num_gpu"] = 30
                    config["recommended_context_size"] = 4096
                    config["chip_model"] = "M1"
                else:
                    # Default for Apple Silicon (unknown chip)
                    config["recommended_num_gpu"] = 35
                    config["chip_model"] = "Apple Silicon (unknown)"
            except (subprocess.CalledProcessError, FileNotFoundError, subprocess.TimeoutExpired):
                # Fallback: reasonable defaults for Apple Silicon
                config["recommended_num_gpu"] = 35
                config["chip_model"] = "Apple Silicon (detected)"

        else:
            # Intel Mac
            config["platform"] = "intel"
            config["gpu_available"] = False
            config["gpu_type"] = None
            config["recommended_num_gpu"] = None
            # Use RAM-based context size (larger if more RAM available)
            config["recommended_context_size"] = config["ram_optimizations"]["recommended_context_size"]
            # Intel Macs: CPU-only, but can use larger contexts if RAM available

    # Linux detection
    elif system == "Linux":
        config["platform"] = "linux"
        gpu_detected = False

        # Check for NVIDIA GPU (CUDA) first
        try:
            result = subprocess.run(
                ["nvidia-smi", "--query-gpu=count", "--format=csv,noheader"],
                capture_output=True,
                text=True,
                timeout=2,
            )
            if result.returncode == 0 and result.stdout.strip():
                config["gpu_available"] = True
                config["gpu_type"] = "cuda"
                config["recommended_num_gpu"] = 40  # Can use more on NVIDIA GPUs
                config["recommended_context_size"] = 8192
                gpu_detected = True
        except (subprocess.CalledProcessError, FileNotFoundError, subprocess.TimeoutExpired):
            pass  # Try AMD GPU next

        # Check for AMD GPU (ROCm) if NVIDIA not found
        if not gpu_detected:
            try:
                result = subprocess.run(
                    ["rocminfo"],  # ROCm info command
                    capture_output=True,
                    text=True,
                    timeout=2,
                )
                if result.returncode == 0:
                    # Check if output contains GPU information
                    if "gfx" in result.stdout.lower() or "GPU" in result.stdout:
                        config["gpu_available"] = True
                        config["gpu_type"] = "rocm"
                        config["recommended_num_gpu"] = 40  # Similar to NVIDIA
                        config["recommended_context_size"] = 8192
                        gpu_detected = True
            except (subprocess.CalledProcessError, FileNotFoundError, subprocess.TimeoutExpired):
                pass  # No AMD GPU or rocminfo not available

        # Fallback to CPU-only if no GPU detected
        if not gpu_detected:
            config["gpu_available"] = False
            config["recommended_num_gpu"] = None
            # Use RAM-based context size (larger if more RAM available)
            config["recommended_context_size"] = config["ram_optimizations"]["recommended_context_size"]

    elif system == "Windows":
        config["platform"] = "windows"
        # Check for AMD GPU (ROCm) on Windows (ROCm v6.1+)
        # Note: NVIDIA CUDA detection on Windows would need different approach
        try:
            result = subprocess.run(
                ["rocminfo"],
                capture_output=True,
                text=True,
                timeout=2,
            )
            if result.returncode == 0 and ("gfx" in result.stdout.lower() or "GPU" in result.stdout):
                config["gpu_available"] = True
                config["gpu_type"] = "rocm"
                config["recommended_num_gpu"] = 40
                config["recommended_context_size"] = 8192
            else:
                config["gpu_available"] = False
                config["recommended_num_gpu"] = None
                config["recommended_context_size"] = 2048
        except (subprocess.CalledProcessError, FileNotFoundError, subprocess.TimeoutExpired):
            # No AMD GPU or rocminfo not available on Windows
            config["gpu_available"] = False
            config["recommended_num_gpu"] = None
            config["recommended_context_size"] = 2048

    else:
        # Unknown platform
        config["platform"] = "unknown"
        config["recommended_num_gpu"] = None
        config["recommended_context_size"] = 2048

    return config


def get_optimized_ollama_options(
    num_gpu: Optional[int] = None,
    num_threads: Optional[int] = None,
    context_size: Optional[int] = None,
    use_auto_detect: bool = True,
) -> dict[str, Any]:
    """
    Get optimized Ollama options based on hardware detection or provided parameters.

    Args:
        num_gpu: Override GPU layers (None = auto-detect)
        num_threads: Override CPU threads (None = auto-detect)
        context_size: Override context size (None = auto-detect)
        use_auto_detect: If True, auto-detect hardware if parameters not provided

    Returns:
        Dict with optimized options for Ollama
    """
    options = {}

    if use_auto_detect:
        hw_config = detect_hardware_config()

        # GPU layers
        if num_gpu is None:
            if hw_config["recommended_num_gpu"] is not None:
                options["num_gpu"] = hw_config["recommended_num_gpu"]
                logger.info(
                    f"Auto-detected hardware: {hw_config['platform']} with {hw_config['gpu_type']} GPU, "
                    f"using {hw_config['recommended_num_gpu']} GPU layers"
                )
        else:
            options["num_gpu"] = num_gpu

        # CPU threads
        if num_threads is None:
            options["num_threads"] = hw_config["recommended_num_threads"]
            logger.info(
                f"Auto-detected {hw_config['cpu_cores']} CPU cores, "
                f"using {hw_config['recommended_num_threads']} threads"
            )
        else:
            options["num_threads"] = num_threads

        # Context size
        if context_size is None:
            options["num_ctx"] = hw_config["recommended_context_size"]
        else:
            options["num_ctx"] = context_size
    else:
        # Manual mode - only use provided parameters
        if num_gpu is not None:
            options["num_gpu"] = num_gpu
        if num_threads is not None:
            options["num_threads"] = num_threads
        if context_size is not None:
            options["num_ctx"] = context_size

    return options


def check_ollama_status(host: Optional[str] = None) -> str:
    """
    [HINT: Ollama status. Check if Ollama server is running and accessible.]

    📊 Output: Server status, version, available models count
    🔧 Side Effects: None (read-only check)
    📁 Checks: Ollama server connection
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "Is Ollama running? Check Ollama status"

    Args:
        host: Optional Ollama host URL (default: http://localhost:11434)

    Returns:
        JSON with Ollama server status
    """
    start_time = time.time()

    if not OLLAMA_AVAILABLE:
        error_response = format_error_response(
            "Ollama package not installed. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)

    try:
        # Configure client if host provided
        if host:
            client = ollama.Client(host=host)
        else:
            client = ollama.Client()

        # Try to list models to check connection
        models_response = client.list()
        model_list = models_response.models if hasattr(models_response, 'models') else models_response.get("models", [])

        # Extract model names
        model_names = []
        for model_obj in model_list[:10]:  # First 10
            if hasattr(model_obj, 'model'):
                model_names.append(model_obj.model)
            elif hasattr(model_obj, 'model_dump'):
                model_dict = model_obj.model_dump()
                model_names.append(model_dict.get("model", ""))
            elif isinstance(model_obj, dict):
                model_names.append(model_obj.get("model", model_obj.get("name", "")))
            else:
                model_names.append(str(model_obj))

        result = {
            "status": "running",
            "host": host or "http://localhost:11434",
            "model_count": len(model_list),
            "models": model_names,
        }

        duration = time.time() - start_time
        log_automation_execution("check_ollama_status", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("check_ollama_status", duration, False, e)

        # Check if it's a connection error
        error_msg = str(e)
        if "connection" in error_msg.lower() or "refused" in error_msg.lower():
            error_msg = "Ollama server not running. Start it with: ollama serve"

        error_response = format_error_response(error_msg, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def list_ollama_models(host: Optional[str] = None) -> str:
    """
    [HINT: Ollama models. List all available Ollama models on the local server.]

    📊 Output: List of models with details (name, size, modified date)
    🔧 Side Effects: None (read-only)
    📁 Queries: Ollama server
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "What Ollama models do I have installed?"

    Args:
        host: Optional Ollama host URL (default: http://localhost:11434)

    Returns:
        JSON with list of available models
    """
    start_time = time.time()

    if not OLLAMA_AVAILABLE:
        error_response = format_error_response(
            "Ollama package not installed. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)

    try:
        # Configure client if host provided
        if host:
            client = ollama.Client(host=host)
        else:
            client = ollama.Client()

        models = client.list()
        model_list = models.models if hasattr(models, 'models') else models.get("models", [])

        # Format model information
        formatted_models = []
        for model_obj in model_list:
            # Convert Pydantic model to dict if needed
            if hasattr(model_obj, 'model_dump'):
                model_dict = model_obj.model_dump()
            elif hasattr(model_obj, 'dict'):
                model_dict = model_obj.dict()
            elif isinstance(model_obj, dict):
                model_dict = model_obj
            else:
                # Fallback: access attributes directly
                model_dict = {
                    "model": getattr(model_obj, "model", ""),
                    "size": getattr(model_obj, "size", 0),
                    "modified_at": getattr(model_obj, "modified_at", None),
                    "digest": getattr(model_obj, "digest", ""),
                }

            # Convert datetime to ISO string if present
            modified_at = model_dict.get("modified_at")
            if modified_at and hasattr(modified_at, 'isoformat'):
                modified_at_str = modified_at.isoformat()
            elif modified_at:
                modified_at_str = str(modified_at)
            else:
                modified_at_str = ""

            formatted_models.append({
                "name": model_dict.get("model", ""),  # Note: Ollama uses "model" not "name"
                "size": model_dict.get("size", 0),
                "modified_at": modified_at_str,
                "digest": model_dict.get("digest", "")[:12] if model_dict.get("digest") else "",  # Short digest
            })

        result = {
            "models": formatted_models,
            "count": len(formatted_models),
            "tip": "Use generate_with_ollama to generate text with a model",
        }

        duration = time.time() - start_time
        log_automation_execution("list_ollama_models", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("list_ollama_models", duration, False, e)

        error_msg = str(e)
        if "connection" in error_msg.lower() or "refused" in error_msg.lower():
            error_msg = "Ollama server not running. Start it with: ollama serve"

        error_response = format_error_response(error_msg, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def generate_with_ollama(
    prompt: str,
    model: str = "llama3.2",
    host: Optional[str] = None,
    stream: bool = False,
    options: dict | None = None,
    num_gpu: Optional[int] = None,
    num_threads: Optional[int] = None,
    context_size: Optional[int] = None,
    use_flash_attention: Optional[bool] = None,
    use_ram_optimizations: bool = True,
) -> str:
    """
    [HINT: Ollama generation. Generate text using a local Ollama model.]

    📊 Output: Generated text response
    🔧 Side Effects: Calls Ollama API (may use GPU/CPU resources)
    📁 Uses: Local Ollama server
    ⏱️ Typical Runtime: 1-30 seconds (depends on model and prompt length)

    Example Prompt:
    "Generate a summary using Ollama llama3.2 model"

    Performance Optimization Tips:
    - Hardware is auto-detected (Intel vs Apple Silicon/M4) and optimized automatically
    - Set num_gpu to offload layers to GPU (e.g., 40 layers) - auto-detected on Apple Silicon
    - Set num_threads for CPU parallelization (match CPU cores) - auto-detected
    - Reduce context_size for faster inference (smaller context = faster) - optimized per hardware
    - Use smaller/quantized models (phi3, llama3.2:1b) for speed
    - Enable streaming for faster perceived response time

    Auto-Detection:
    - Apple Silicon (M1/M2/M3/M4): Automatically uses Metal GPU acceleration (30-40 layers)
    - Intel Mac: Optimized for CPU-only inference
    - Linux: Detects NVIDIA GPU (CUDA) or AMD GPU (ROCm) if available
    - Windows: Detects AMD GPU (ROCm v6.1+) if available
    - Parameters can override auto-detection

    Args:
        prompt: Text prompt to send to the model
        model: Model name (default: llama3.2)
        host: Optional Ollama host URL (default: http://localhost:11434)
        stream: Whether to stream the response (default: False) - enables faster perceived response
        options: Optional model parameters (temperature, top_p, top_k, repeat_penalty, etc.)
        num_gpu: Number of layers to offload to GPU (None = auto-detect based on hardware)
        num_threads: Number of CPU threads to use (None = auto-detect based on CPU cores)
        context_size: Context window size (None = auto-optimize based on hardware)

    Returns:
        JSON with generated text
    """
    start_time = time.time()

    if not OLLAMA_AVAILABLE:
        error_response = format_error_response(
            "Ollama package not installed. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)

    try:
        # Configure client if host provided
        if host:
            client = ollama.Client(host=host)
        else:
            client = ollama.Client()

        # Prepare generation parameters with performance optimizations
        gen_options = options.copy() if options else {}

        # Auto-detect hardware and apply optimizations (unless explicitly overridden)
        # Priority: explicit args > env vars > auto-detection
        auto_detect = True

        # Check if any performance params are explicitly set (via args or env)
        has_explicit_gpu = num_gpu is not None or os.getenv("OLLAMA_NUM_GPU")
        has_explicit_threads = num_threads is not None or os.getenv("OLLAMA_NUM_THREADS")
        has_explicit_ctx = context_size is not None or os.getenv("OLLAMA_NUM_CTX")

        # GPU layers: explicit arg > env var > auto-detect
        if num_gpu is not None:
            gen_options["num_gpu"] = num_gpu
        elif "num_gpu" not in gen_options:
            if os.getenv("OLLAMA_NUM_GPU"):
                try:
                    gen_options["num_gpu"] = int(os.getenv("OLLAMA_NUM_GPU"))
                except ValueError:
                    logger.warning(f"Invalid OLLAMA_NUM_GPU value: {os.getenv('OLLAMA_NUM_GPU')}")
            elif auto_detect and not has_explicit_gpu:
                # Auto-detect GPU settings
                hw_config = detect_hardware_config()
                if hw_config["recommended_num_gpu"] is not None:
                    gen_options["num_gpu"] = hw_config["recommended_num_gpu"]
                    logger.info(
                        f"🚀 Auto-configured GPU: {hw_config['platform']} with {hw_config['gpu_type']} GPU, "
                        f"using {hw_config['recommended_num_gpu']} layers"
                    )

        # CPU threads: explicit arg > env var > auto-detect
        if num_threads is not None:
            gen_options["num_threads"] = num_threads
        elif "num_threads" not in gen_options:
            if os.getenv("OLLAMA_NUM_THREADS"):
                try:
                    gen_options["num_threads"] = int(os.getenv("OLLAMA_NUM_THREADS"))
                except ValueError:
                    logger.warning(f"Invalid OLLAMA_NUM_THREADS value: {os.getenv('OLLAMA_NUM_THREADS')}")
            elif auto_detect and not has_explicit_threads:
                # Auto-detect CPU threads
                hw_config = detect_hardware_config()
                gen_options["num_threads"] = hw_config["recommended_num_threads"]
                logger.info(
                    f"🚀 Auto-configured CPU: {hw_config['cpu_cores']} cores, "
                    f"using {hw_config['recommended_num_threads']} threads"
                )

        # Context size: explicit arg > env var > auto-detect (with RAM optimization)
        if context_size is not None:
            gen_options["num_ctx"] = context_size
        elif "num_ctx" not in gen_options:
            if os.getenv("OLLAMA_NUM_CTX"):
                try:
                    gen_options["num_ctx"] = int(os.getenv("OLLAMA_NUM_CTX"))
                except ValueError:
                    logger.warning(f"Invalid OLLAMA_NUM_CTX value: {os.getenv('OLLAMA_NUM_CTX')}")
            elif auto_detect and not has_explicit_ctx and use_ram_optimizations:
                # Auto-detect context size based on RAM
                hw_config = detect_hardware_config()
                gen_options["num_ctx"] = hw_config["recommended_context_size"]
                if hw_config.get("ram_gb", 0) >= 16:
                    logger.info(f"💾 Large RAM detected ({hw_config['ram_gb']:.1f}GB) - using larger context size ({hw_config['recommended_context_size']})")

        # Flash Attention: explicit > env var > auto-detect (based on RAM)
        if use_flash_attention is not None:
            if use_flash_attention:
                # Flash Attention is enabled via environment variable
                os.environ["OLLAMA_FLASH_ATTENTION"] = "1"
        elif use_ram_optimizations and auto_detect:
            # Auto-enable Flash Attention if RAM allows
            hw_config = detect_hardware_config()
            if hw_config.get("ram_optimizations", {}).get("enable_flash_attention", False):
                if not os.getenv("OLLAMA_FLASH_ATTENTION"):
                    os.environ["OLLAMA_FLASH_ATTENTION"] = "1"
                    logger.info("💾 Flash Attention enabled (RAM optimization)")

        # KV Cache quantization for memory efficiency (optional)
        if use_ram_optimizations and auto_detect:
            hw_config = detect_hardware_config()
            if hw_config.get("ram_gb", 0) < 16 and not os.getenv("OLLAMA_KV_CACHE_TYPE"):
                # For systems with <16GB RAM, use quantized KV cache to save memory
                gen_options.setdefault("kv_cache_type", "q8_0")  # 8-bit quantization
                logger.info("💾 Using quantized KV cache for memory efficiency")

        # Generate response
        if stream:
            # Stream response
            response_text = ""
            for chunk in client.generate(model=model, prompt=prompt, stream=True, options=gen_options):
                if "response" in chunk:
                    response_text += chunk["response"]
        else:
            # Single response
            response = client.generate(model=model, prompt=prompt, stream=False, options=gen_options)
            response_text = response.get("response", "")

        # Get hardware info for response (if auto-detected)
        hw_info = None
        if auto_detect and (gen_options.get("num_gpu") or gen_options.get("num_threads") or gen_options.get("num_ctx")):
            try:
                hw_config = detect_hardware_config()
                hw_info = {
                    "platform": hw_config["platform"],
                    "gpu_type": hw_config["gpu_type"],
                    "cpu_cores": hw_config["cpu_cores"],
                }
            except Exception:
                pass  # Don't fail if hardware detection fails

        # Get RAM info if available
        try:
            hw_config = detect_hardware_config()
            ram_info = {
                "total_ram_gb": hw_config.get("ram_gb"),
                "ram_optimizations_applied": hw_config.get("ram_optimizations", {}),
            }
        except Exception:
            ram_info = None

        result = {
            "model": model,
            "prompt": prompt[:200] + "..." if len(prompt) > 200 else prompt,  # Truncate for display
            "response": response_text,
            "length": len(response_text),
            "performance_options": {
                "num_gpu": gen_options.get("num_gpu"),
                "num_threads": gen_options.get("num_threads"),
                "context_size": gen_options.get("num_ctx"),
                "streaming": stream,
                "flash_attention": os.getenv("OLLAMA_FLASH_ATTENTION") == "1",
                "kv_cache_type": gen_options.get("kv_cache_type"),
            },
            "hardware_detected": hw_info,
            "ram_info": ram_info,
            "duration_seconds": round(time.time() - start_time, 2),
        }

        duration = time.time() - start_time
        log_automation_execution("generate_with_ollama", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("generate_with_ollama", duration, False, e)

        error_msg = str(e)
        if "connection" in error_msg.lower() or "refused" in error_msg.lower():
            error_msg = "Ollama server not running. Start it with: ollama serve"
        elif "model" in error_msg.lower() and "not found" in error_msg.lower():
            error_msg = f"Model '{model}' not found. Pull it with: ollama pull {model}"

        error_response = format_error_response(error_msg, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def pull_ollama_model(model: str, host: Optional[str] = None) -> str:
    """
    [HINT: Ollama pull. Download/pull an Ollama model from the registry.]

    📊 Output: Pull status and progress
    🔧 Side Effects: Downloads model (may take time and disk space)
    📁 Downloads: Model from Ollama registry
    ⏱️ Typical Runtime: 30 seconds - 10 minutes (depends on model size)

    Example Prompt:
    "Pull the llama3.2 model from Ollama"

    Args:
        model: Model name to pull (e.g., "llama3.2", "mistral", "codellama")
        host: Optional Ollama host URL (default: http://localhost:11434)

    Returns:
        JSON with pull status
    """
    start_time = time.time()

    if not OLLAMA_AVAILABLE:
        error_response = format_error_response(
            "Ollama package not installed. Install with: uv sync",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)

    try:
        # Configure client if host provided
        if host:
            client = ollama.Client(host=host)
        else:
            client = ollama.Client()

        # Pull model (this may take a while)
        logger.info(f"Pulling model: {model} (this may take several minutes)")
        client.pull(model)

        result = {
            "model": model,
            "status": "completed",
            "message": f"Model {model} pulled successfully",
        }

        duration = time.time() - start_time
        log_automation_execution("pull_ollama_model", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("pull_ollama_model", duration, False, e)

        error_msg = str(e)
        if "connection" in error_msg.lower() or "refused" in error_msg.lower():
            error_msg = "Ollama server not running. Start it with: ollama serve"

        error_response = format_error_response(error_msg, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def get_hardware_info() -> str:
    """
    [HINT: Hardware detection. Get hardware information and recommended Ollama settings.]

    📊 Output: Hardware platform, GPU availability, recommended settings
    🔧 Side Effects: None (read-only hardware detection)
    📁 Detects: CPU architecture, GPU type, optimal performance settings
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "What hardware do I have? Check hardware configuration for Ollama"

    Returns:
        JSON with hardware information and recommended Ollama settings
    """
    try:
        hw_config = detect_hardware_config()

        result = {
            "platform": hw_config["platform"],
            "architecture": hw_config["architecture"],
            "cpu_cores": hw_config["cpu_cores"],
            "ram_gb": round(hw_config.get("ram_gb", 0), 1),
            "gpu_available": hw_config["gpu_available"],
            "gpu_type": hw_config["gpu_type"],
            "chip_model": hw_config.get("chip_model"),
            "recommended_settings": {
                "num_threads": hw_config["recommended_num_threads"],
                "num_gpu": hw_config["recommended_num_gpu"],
                "context_size": hw_config["recommended_context_size"],
            },
            "ram_optimizations": hw_config.get("ram_optimizations", {}),
            "notes": []
        }

        # Add RAM-specific notes
        ram_gb = hw_config.get("ram_gb", 0)
        if ram_gb >= 32:
            result["notes"].append(f"Large RAM ({ram_gb:.1f}GB) - Can use very large contexts (up to 16K)")
        elif ram_gb >= 16:
            result["notes"].append(f"Good RAM ({ram_gb:.1f}GB) - Can use large contexts (up to 8K)")
        elif ram_gb >= 8:
            result["notes"].append(f"Moderate RAM ({ram_gb:.1f}GB) - Medium contexts recommended")
        else:
            result["notes"].append(f"Limited RAM ({ram_gb:.1f}GB) - Use smaller contexts and models")

        # Add helpful notes
        if hw_config["platform"] == "apple_silicon":
            chip_info = f" ({hw_config.get('chip_model', '')})" if hw_config.get("chip_model") else ""
            result["notes"].append(f"Apple Silicon{chip_info} detected - Metal GPU acceleration available")
            if hw_config["recommended_num_gpu"]:
                result["notes"].append(f"Recommended: Use {hw_config['recommended_num_gpu']} GPU layers for optimal performance")
        elif hw_config["platform"] == "intel":
            result["notes"].append("Intel Mac detected - CPU-only inference (no GPU acceleration)")
            result["notes"].append("Recommendation: Use smaller models (phi3, llama3.2:1b) for faster inference")
        elif hw_config["gpu_available"] and hw_config["gpu_type"] == "cuda":
            result["notes"].append("NVIDIA GPU detected - CUDA acceleration available")
            result["notes"].append(f"Recommended: Use {hw_config['recommended_num_gpu']} GPU layers")
        elif hw_config["gpu_available"] and hw_config["gpu_type"] == "rocm":
            result["notes"].append("AMD GPU detected - ROCm acceleration available")
            result["notes"].append(f"Recommended: Use {hw_config['recommended_num_gpu']} GPU layers")
            result["notes"].append("Note: Check Ollama docs for supported AMD GPU models")
        else:
            result["notes"].append("CPU-only inference recommended")
            result["notes"].append("Recommendation: Use smaller models for faster inference")

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        error_response = format_error_response(
            f"Error detecting hardware: {str(e)}",
            ErrorCode.AUTOMATION_ERROR
        )
        return json.dumps(error_response, indent=2)


def register_ollama_tools(mcp):
    """
    Register Ollama tools with FastMCP server.

    Args:
        mcp: FastMCP server instance
    """
    if not OLLAMA_AVAILABLE:
        logger.warning("Ollama tools not registered: ollama package not available")
        return

    try:
        @mcp.tool()
        def check_ollama_status_tool(host: Optional[str] = None) -> str:
            """Check if Ollama server is running and accessible."""
            return check_ollama_status(host)

        @mcp.tool()
        def get_hardware_info_tool() -> str:
            """Get hardware information and recommended Ollama performance settings."""
            return get_hardware_info()

        @mcp.tool()
        def list_ollama_models_tool(host: Optional[str] = None) -> str:
            """List all available Ollama models on the local server."""
            return list_ollama_models(host)

        @mcp.tool()
        def generate_with_ollama_tool(
            prompt: str,
            model: str = "llama3.2",
            host: Optional[str] = None,
            stream: bool = False,
            options: Optional[str] = None,  # JSON string
            num_gpu: Optional[int] = None,
            num_threads: Optional[int] = None,
            context_size: Optional[int] = None,
        ) -> str:
            """
            Generate text using a local Ollama model.

            Performance optimization parameters:
            - num_gpu: Number of layers to offload to GPU (speeds up inference)
            - num_threads: Number of CPU threads (match your CPU cores)
            - context_size: Context window size (smaller = faster)
            - stream: Enable streaming for faster perceived response time

            Args:
                prompt: Text prompt to send to the model
                model: Model name (default: llama3.2)
                host: Optional Ollama host URL
                stream: Whether to stream the response (faster perceived speed)
                options: Optional JSON string with model parameters (temperature, top_p, etc.)
                num_gpu: Number of GPU layers (default: from OLLAMA_NUM_GPU env var)
                num_threads: Number of CPU threads (default: from OLLAMA_NUM_THREADS env var)
                context_size: Context window size (default: model default)
            """
            # Parse options if provided
            parsed_options = None
            if options:
                try:
                    parsed_options = json.loads(options)
                except json.JSONDecodeError:
                    logger.warning(f"Invalid options JSON: {options}")

            return generate_with_ollama(
                prompt, model, host, stream, parsed_options,
                num_gpu=num_gpu,
                num_threads=num_threads,
                context_size=context_size,
            )

        @mcp.tool()
        def pull_ollama_model_tool(model: str, host: Optional[str] = None) -> str:
            """Download/pull an Ollama model from the registry."""
            return pull_ollama_model(model, host)

        logger.info("✅ Ollama tools registered")

    except Exception as e:
        logger.error(f"Failed to register Ollama tools: {e}", exc_info=True)

