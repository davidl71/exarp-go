#!/usr/bin/env python3
"""
Statistics Bridge - Python wrapper for Go statistics functions

Provides Python-compatible API for Go statistics helper functions.
Calls Go CLI tool via subprocess for statistical calculations.
"""

import json
import subprocess
from pathlib import Path
from typing import List, Optional


def _get_go_binary_path() -> Path:
    """Get path to statistics Go binary."""
    # Get project root (bridge is in project_root/bridge/)
    project_root = Path(__file__).parent.parent
    
    # Binary is in bin/statistics (after build)
    binary_path = project_root / "bin" / "statistics"
    
    # If binary doesn't exist, try to find it in PATH or build it
    if not binary_path.exists():
        # Try to find in PATH
        import shutil
        path_binary = shutil.which("statistics")
        if path_binary:
            return Path(path_binary)
        
        # If not found, we'll need to build it
        # For now, return the expected path
        # Caller should handle the error
        return binary_path
    
    return binary_path


def _call_go_statistics(function: str, data: Optional[List[float]] = None, 
                        p: Optional[float] = None, value: Optional[float] = None,
                        decimals: Optional[int] = None) -> float:
    """
    Call Go statistics function via subprocess.
    
    Args:
        function: Function name ('mean', 'median', 'stdev', 'quantile', 'round')
        data: List of floats (for mean, median, stdev, quantile)
        p: Quantile value 0.0-1.0 (for quantile)
        value: Value to round (for round)
        decimals: Number of decimal places (for round)
    
    Returns:
        float: Statistical result
    
    Raises:
        RuntimeError: If Go binary not found or execution fails
        ValueError: If invalid parameters provided
    """
    binary_path = _get_go_binary_path()
    
    if not binary_path.exists():
        raise RuntimeError(
            f"Statistics Go binary not found at {binary_path}. "
            "Please build it with: make build"
        )
    
    # Build request
    request = {"function": function}
    
    if data is not None:
        request["data"] = data
    if p is not None:
        request["p"] = p
    if value is not None:
        request["value"] = value
    if decimals is not None:
        request["decimals"] = decimals
    
    # Call Go binary
    try:
        result = subprocess.run(
            [str(binary_path), json.dumps(request)],
            capture_output=True,
            text=True,
            check=True,
            timeout=5.0
        )
        
        response = json.loads(result.stdout)
        
        if not response.get("success", False):
            error_msg = response.get("error", "Unknown error")
            raise RuntimeError(f"Go statistics error: {error_msg}")
        
        result_value = response.get("result")
        if result_value is None:
            raise RuntimeError("Go statistics returned no result")
        
        return float(result_value)
    
    except subprocess.TimeoutExpired:
        raise RuntimeError("Statistics calculation timed out")
    except subprocess.CalledProcessError as e:
        raise RuntimeError(f"Statistics calculation failed: {e.stderr}")
    except json.JSONDecodeError as e:
        raise RuntimeError(f"Failed to parse statistics response: {e}")


def mean(data: List[float]) -> float:
    """
    Calculate the arithmetic mean of a list of floats.
    
    Args:
        data: List of float values
    
    Returns:
        float: Mean value (0.0 for empty list)
    
    Example:
        >>> mean([1.0, 2.0, 3.0, 4.0, 5.0])
        3.0
    """
    if not data:
        return 0.0
    return _call_go_statistics("mean", data=data)


def median(data: List[float]) -> float:
    """
    Calculate the median (50th percentile) of a list of floats.
    
    Args:
        data: List of float values
    
    Returns:
        float: Median value (0.0 for empty list)
    
    Example:
        >>> median([1.0, 2.0, 3.0, 4.0, 5.0])
        3.0
    """
    if not data:
        return 0.0
    return _call_go_statistics("median", data=data)


def stdev(data: List[float]) -> float:
    """
    Calculate the sample standard deviation of a list of floats.
    
    Args:
        data: List of float values
    
    Returns:
        float: Standard deviation (0.0 for empty or single-element list)
    
    Example:
        >>> stdev([1.0, 2.0, 3.0, 4.0, 5.0])
        1.5811388300841898
    """
    if len(data) <= 1:
        return 0.0
    return _call_go_statistics("stdev", data=data)


def quantile(data: List[float], p: float) -> float:
    """
    Calculate the specified quantile (0.0 to 1.0) of a list of floats.
    
    Args:
        data: List of float values
        p: Quantile value between 0.0 and 1.0
    
    Returns:
        float: Quantile value (0.0 for empty list or invalid p)
    
    Example:
        >>> quantile([1.0, 2.0, 3.0, 4.0, 5.0], 0.25)
        2.0
        >>> quantile([1.0, 2.0, 3.0, 4.0, 5.0], 0.5)
        3.0
    """
    if not data:
        return 0.0
    if p < 0.0 or p > 1.0:
        return 0.0
    return _call_go_statistics("quantile", data=data, p=p)


def round(value: float, decimals: int = 2) -> float:
    """
    Round a float to the specified number of decimal places.
    
    Args:
        value: Float value to round
        decimals: Number of decimal places (default: 2)
    
    Returns:
        float: Rounded value
    
    Example:
        >>> round(3.14159, 2)
        3.14
        >>> round(2.5, 0)
        3.0
    """
    return _call_go_statistics("round", value=value, decimals=decimals)


# Compatibility: Provide StatisticsError exception for Python statistics compatibility
class StatisticsError(Exception):
    """Exception raised for statistics calculation errors."""
    pass

