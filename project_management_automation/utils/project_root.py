"""Project root detection utilities."""

from pathlib import Path
from typing import Optional

# Module-level cache for project root (only invalidates on process restart)
_cached_project_root: Optional[Path] = None


def find_project_root(start_path: Optional[Path] = None) -> Path:
    """
    Find project root by looking for marker files.

    Looks for .git, .todo2, CMakeLists.txt, go.mod, or pyproject.toml to identify project root.

    Search order:
    1. PROJECT_ROOT or WORKSPACE_PATH environment variable (if set)
    2. If start_path provided, search up from there
    3. Search up from current working directory
    4. Search up from package location (for MCP server context)

    Args:
        start_path: Starting path for search (optional)

    Returns:
        Path to project root, or current working directory if not found
    """
    import os
    
    def _search_up(path: Path) -> Optional[Path]:
        """Search upward for project markers."""
        current = path.resolve()
        while current != current.parent:
            # Check for common project markers
            if ((current / '.git').exists() or 
                (current / '.todo2').exists() or 
                (current / 'CMakeLists.txt').exists() or 
                (current / 'go.mod').exists() or
                (current / 'pyproject.toml').exists()):
                return current
            current = current.parent
        return None
    
    # Check environment variable first (highest priority)
    env_root = os.getenv("PROJECT_ROOT") or os.getenv("WORKSPACE_PATH")
    if env_root:
        env_path = Path(env_root)
        if env_path.exists():
            return env_path.resolve()

    global _cached_project_root

    # If explicit start_path, don't use cache
    if start_path is not None:
        result = _search_up(Path(start_path))
        if result:
            return result
        return Path(start_path).resolve()

    # Use cached value if available
    if _cached_project_root is not None:
        return _cached_project_root

    # Try current working directory first
    result = _search_up(Path.cwd())
    if result:
        _cached_project_root = result
        return result

    # Try package location (for MCP server context)
    # Go up from utils/project_root.py to project root
    package_path = Path(__file__).parent.parent.parent  # utils -> project_management_automation -> project root
    result = _search_up(package_path)
    if result:
        _cached_project_root = result
        return result

    # Fallback to current working directory
    _cached_project_root = Path.cwd()
    return _cached_project_root

