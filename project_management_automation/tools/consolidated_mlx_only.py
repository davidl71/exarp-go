"""
MLX-only re-export for bridge. Bridge imports this so it only loads the mlx path.
All other tools are native Go; do not add more exports here.
"""
from .consolidated_ai import mlx

__all__ = ["mlx"]
