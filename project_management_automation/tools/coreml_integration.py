"""
Core ML Integration Tools

Provides tools for using Apple's Core ML framework to leverage Neural Engine (NPU)
on Apple Silicon devices. Core ML automatically uses the Neural Engine when available
and when models support it.

Features:
- Model loading and inference
- Neural Engine detection
- Hardware optimization
- Model conversion utilities
"""

import json
import logging
import platform
import subprocess
import time
from pathlib import Path
from typing import Any, Optional

logger = logging.getLogger(__name__)

# Try to import Core ML
try:
    import coremltools as ct
    CORE_ML_AVAILABLE = True
except ImportError:
    CORE_ML_AVAILABLE = False
    logger.debug("Core ML not available. Install with: uv pip install coremltools")

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


def get_chip_model() -> str:
    """Get Apple Silicon chip model."""
    try:
        chip = subprocess.check_output(
            ["sysctl", "-n", "machdep.cpu.brand_string"],
            stderr=subprocess.DEVNULL,
            text=True,
            timeout=1,
        ).strip()
        return chip
    except Exception:
        return "Unknown"


def check_coreml_availability() -> dict[str, Any]:
    """
    Check Core ML availability and hardware support.
    
    Returns:
        Dict with availability status and hardware info
    """
    result = {
        "coreml_available": CORE_ML_AVAILABLE,
        "apple_silicon": is_apple_silicon(),
        "chip_model": get_chip_model(),
        "neural_engine_support": False,
        "notes": [],
    }
    
    if not CORE_ML_AVAILABLE:
        result["notes"].append("Install coremltools: uv pip install coremltools")
        return result
    
    if not is_apple_silicon():
        result["notes"].append("Core ML works best on Apple Silicon")
        return result
    
    # Core ML automatically uses Neural Engine when:
    # 1. Model supports it (ANE-compatible operations)
    # 2. Hardware supports it (M1+ chips have Neural Engine)
    # 3. Model is optimized for Neural Engine
    
    chip = get_chip_model()
    if "M1" in chip or "M2" in chip or "M3" in chip or "M4" in chip or "M5" in chip:
        result["neural_engine_support"] = True
        result["notes"].append(
            f"{chip} has Neural Engine. Core ML will use it automatically "
            "when models support ANE operations."
        )
    
    return result


def list_coreml_models(model_dir: Optional[str] = None) -> str:
    """
    List available Core ML models.
    
    Args:
        model_dir: Directory to search for .mlmodel or .mlpackage files
        
    Returns:
        JSON string with available models
    """
    start_time = time.time()
    
    if not CORE_ML_AVAILABLE:
        error_response = format_error_response(
            "Core ML not available. Install with: uv pip install coremltools",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)
    
    if not is_apple_silicon():
        error_response = format_error_response(
            "Core ML works best on Apple Silicon devices",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)
    
    try:
        models = []
        
        # Search for models
        if model_dir:
            search_path = Path(model_dir)
        else:
            # Default locations
            home = Path.home()
            search_paths = [
                home / "Downloads",
                home / "Documents" / "CoreMLModels",
                Path(".") / "models",
            ]
            # Use first existing path
            search_path = next((p for p in search_paths if p.exists()), None)
        
        if search_path and search_path.exists():
            # Find .mlmodel and .mlpackage files
            mlmodels = list(search_path.rglob("*.mlmodel"))
            mlpackages = list(search_path.rglob("*.mlpackage"))
            
            for model_path in mlmodels + mlpackages:
                try:
                    model = ct.models.MLModel(str(model_path))
                    spec = model.get_spec()
                    
                    models.append({
                        "name": model_path.name,
                        "path": str(model_path),
                        "type": "mlpackage" if model_path.suffix == ".mlpackage" else "mlmodel",
                        "description": spec.description,
                        "compute_units": "automatic",  # Core ML chooses best (CPU/GPU/ANE)
                    })
                except Exception as e:
                    logger.debug(f"Error loading model {model_path}: {e}")
        
        result = format_success_response({
            "models": models,
            "count": len(models),
            "search_path": str(search_path) if search_path else None,
        }, f"Found {len(models)} Core ML model(s)")
        
        duration = time.time() - start_time
        log_automation_execution("list_coreml_models", duration, True)
        
        return json.dumps(result, indent=2)
    
    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("list_coreml_models", duration, False, e)
        error_response = format_error_response(
            f"Error listing Core ML models: {str(e)}",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)


def predict_with_coreml(
    model_path: str,
    input_data: dict[str, Any],
    compute_units: str = "all",  # "all", "cpu_and_gpu", "cpu_and_ane", "cpu_only"
) -> str:
    """
    Run inference with a Core ML model.
    
    Core ML automatically uses the best available compute unit:
    - Neural Engine (ANE) when model supports it
    - GPU (Metal) when available
    - CPU as fallback
    
    Args:
        model_path: Path to .mlmodel or .mlpackage file
        input_data: Input data dict matching model's input specification
        compute_units: Preferred compute units (all, cpu_and_gpu, cpu_and_ane, cpu_only)
        
    Returns:
        JSON string with predictions
    """
    start_time = time.time()
    
    if not CORE_ML_AVAILABLE:
        error_response = format_error_response(
            "Core ML not available. Install with: uv pip install coremltools",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)
    
    if not is_apple_silicon():
        error_response = format_error_response(
            "Core ML works best on Apple Silicon devices",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)
    
    try:
        # Map compute units string to Core ML enum
        compute_unit_map = {
            "all": ct.ComputeUnit.ALL,
            "cpu_and_gpu": ct.ComputeUnit.CPU_AND_GPU,
            "cpu_and_ane": ct.ComputeUnit.CPU_AND_NE,  # Note: CPU_AND_NE is the correct enum name
            "cpu_only": ct.ComputeUnit.CPU_ONLY,
        }
        
        compute_unit = compute_unit_map.get(compute_units.lower(), ct.ComputeUnit.ALL)
        
        # Load model
        model = ct.models.MLModel(
            model_path,
            compute_units=compute_unit
        )
        
        # Get model info
        spec = model.get_spec()
        input_names = [inp.name for inp in spec.description.input]
        
        # Prepare input
        # Core ML expects inputs as dict with proper types
        coreml_input = {}
        for name, value in input_data.items():
            if name in input_names:
                coreml_input[name] = value
        
        # Run prediction
        predictions = model.predict(coreml_input)
        
        # Convert predictions to JSON-serializable format
        if isinstance(predictions, dict):
            predictions_dict = {}
            for key, value in predictions.items():
                # Handle numpy arrays and other types
                if hasattr(value, 'tolist'):
                    predictions_dict[key] = value.tolist()
                elif hasattr(value, '__dict__'):
                    predictions_dict[key] = str(value)
                else:
                    predictions_dict[key] = value
            predictions = predictions_dict
        
        result = format_success_response({
            "predictions": predictions,
            "model_path": model_path,
            "compute_units": compute_units,
            "input_names": input_names,
            "neural_engine_used": compute_units in ["all", "cpu_and_ane"],
        }, "Core ML prediction successful")
        
        duration = time.time() - start_time
        log_automation_execution("predict_with_coreml", duration, True)
        
        return json.dumps(result, indent=2)
    
    except FileNotFoundError:
        error_response = format_error_response(
            f"Model file not found: {model_path}",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)
    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("predict_with_coreml", duration, False, e)
        error_response = format_error_response(
            f"Error running Core ML prediction: {str(e)}",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)


def get_coreml_hardware_info() -> str:
    """
    Get Core ML hardware information and Neural Engine support.
    
    Returns:
        JSON string with hardware info
    """
    start_time = time.time()
    
    if not CORE_ML_AVAILABLE:
        error_response = format_error_response(
            "Core ML not available. Install with: uv pip install coremltools",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)
    
    try:
        info = check_coreml_availability()
        
        # Add more details
        info["chip_model"]
        info["compute_units_available"] = {
            "cpu": True,  # Always available
            "gpu": is_apple_silicon(),  # Metal GPU on Apple Silicon
            "neural_engine": info["neural_engine_support"],
        }
        
        info["recommendations"] = []
        if info["neural_engine_support"]:
            info["recommendations"].append(
                "Use Core ML models optimized for Neural Engine for best performance"
            )
            info["recommendations"].append(
                "Set compute_units='all' or 'cpu_and_ane' to prefer Neural Engine"
            )
        else:
            info["recommendations"].append(
                "Core ML will use CPU and GPU (Metal) for acceleration"
            )
        
        result = format_success_response(info, "Core ML hardware info retrieved")
        
        duration = time.time() - start_time
        log_automation_execution("get_coreml_hardware_info", duration, True)
        
        return json.dumps(result, indent=2)
    
    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("get_coreml_hardware_info", duration, False, e)
        error_response = format_error_response(
            f"Error getting Core ML hardware info: {str(e)}",
            ErrorCode.AUTOMATION_ERROR,
        )
        return json.dumps(error_response, indent=2)

