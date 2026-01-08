"""
Security utilities for Exarp MCP server.

Provides:
- Path boundary enforcement (prevent path traversal attacks)
- Input validation and sanitization
- Rate limiting for tool calls
- Access control helpers

These controls ensure safe operation when AI assistants execute tools.
"""

import re
import time
from collections import defaultdict
from functools import wraps
from pathlib import Path
from typing import Callable, Optional, Union

from .logging_config import get_logger

logger = get_logger(__name__)


# ═══════════════════════════════════════════════════════════════════════════════
# PATH BOUNDARY ENFORCEMENT
# ═══════════════════════════════════════════════════════════════════════════════

class PathBoundaryError(Exception):
    """Raised when a path operation attempts to escape allowed boundaries."""
    pass


class PathValidator:
    """
    Validates and normalizes paths to prevent directory traversal attacks.

    Usage:
        validator = PathValidator(allowed_roots=[project_root])
        safe_path = validator.validate("/path/to/file")
    """

    def __init__(
        self,
        allowed_roots: Optional[list[Path]] = None,
        allow_symlinks: bool = False,
        blocked_patterns: Optional[list[str]] = None
    ):
        """
        Initialize path validator.

        Args:
            allowed_roots: List of allowed root directories. If None, uses cwd.
            allow_symlinks: Whether to allow symlinks (default: False for security)
            blocked_patterns: Regex patterns for blocked paths (e.g., ".git", "__pycache__")
        """
        self.allowed_roots = [Path(r).resolve() for r in (allowed_roots or [Path.cwd()])]
        self.allow_symlinks = allow_symlinks
        self.blocked_patterns = blocked_patterns or [
            r'\.git(?:/|$)',      # .git directory
            r'\.env',             # Environment files
            r'\.ssh',             # SSH keys
            r'\.aws',             # AWS credentials
            r'\.gnupg',           # GPG keys
            r'id_rsa',            # SSH private keys
            r'\.pem$',            # Certificate files
            r'secrets?\.ya?ml',   # Secrets files
            r'credentials',       # Credential files
        ]
        self._compiled_patterns = [re.compile(p, re.IGNORECASE) for p in self.blocked_patterns]

    def is_within_boundary(self, path: Path) -> bool:
        """Check if path is within allowed boundaries."""
        resolved = path.resolve()
        return any(
            resolved == root or root in resolved.parents
            for root in self.allowed_roots
        )

    def is_blocked(self, path: Path) -> bool:
        """Check if path matches any blocked patterns."""
        path_str = str(path)
        return any(pattern.search(path_str) for pattern in self._compiled_patterns)

    def validate(self, path: Union[str, Path], must_exist: bool = False) -> Path:
        """
        Validate a path is safe to access.

        Args:
            path: Path to validate
            must_exist: Whether the path must exist

        Returns:
            Resolved, validated Path

        Raises:
            PathBoundaryError: If path is outside boundaries or blocked
        """
        try:
            # Convert to Path and resolve
            p = Path(path).expanduser()

            # Check for symlinks if not allowed
            if not self.allow_symlinks and p.is_symlink():
                raise PathBoundaryError(f"Symlinks not allowed: {path}")

            resolved = p.resolve()

            # Check boundary
            if not self.is_within_boundary(resolved):
                raise PathBoundaryError(
                    f"Path outside allowed boundaries: {path}\n"
                    f"Allowed roots: {[str(r) for r in self.allowed_roots]}"
                )

            # Check blocked patterns
            if self.is_blocked(resolved):
                raise PathBoundaryError(f"Path matches blocked pattern: {path}")

            # Check existence if required
            if must_exist and not resolved.exists():
                raise PathBoundaryError(f"Path does not exist: {path}")

            return resolved

        except (OSError, ValueError) as e:
            raise PathBoundaryError(f"Invalid path: {path} - {e}")

    def validate_output_path(self, path: Union[str, Path]) -> Path:
        """
        Validate a path for writing output.
        Creates parent directories if needed.
        """
        validated = self.validate(path, must_exist=False)
        validated.parent.mkdir(parents=True, exist_ok=True)
        return validated


# Global default validator (set during server init)
_default_validator: Optional[PathValidator] = None


def set_default_path_validator(validator: PathValidator) -> None:
    """Set the default path validator for the application."""
    global _default_validator
    _default_validator = validator


def get_default_path_validator() -> PathValidator:
    """Get or create the default path validator."""
    global _default_validator
    if _default_validator is None:
        _default_validator = PathValidator()
    return _default_validator


def validate_path(path: Union[str, Path], must_exist: bool = False) -> Path:
    """Convenience function to validate a path using the default validator."""
    return get_default_path_validator().validate(path, must_exist)


# ═══════════════════════════════════════════════════════════════════════════════
# INPUT VALIDATION
# ═══════════════════════════════════════════════════════════════════════════════

class InputValidationError(Exception):
    """Raised when input validation fails."""
    pass


def sanitize_string(
    value: str,
    max_length: int = 10000,
    allow_newlines: bool = True,
    strip_control_chars: bool = True
) -> str:
    """
    Sanitize a string input.

    Args:
        value: String to sanitize
        max_length: Maximum allowed length
        allow_newlines: Whether to preserve newlines
        strip_control_chars: Remove control characters

    Returns:
        Sanitized string
    """
    if not isinstance(value, str):
        raise InputValidationError(f"Expected string, got {type(value).__name__}")

    if len(value) > max_length:
        raise InputValidationError(f"String exceeds maximum length ({max_length})")

    if strip_control_chars:
        # Remove control characters except tab, newline, carriage return
        allowed_controls = {'\t', '\n', '\r'} if allow_newlines else {'\t'}
        value = ''.join(
            c for c in value
            if c >= ' ' or c in allowed_controls
        )

    return value


def validate_identifier(value: str, pattern: str = r'^[a-zA-Z_][a-zA-Z0-9_-]*$') -> str:
    """
    Validate a string is a safe identifier (no injection risk).

    Args:
        value: String to validate
        pattern: Regex pattern for valid identifiers

    Returns:
        Validated identifier
    """
    if not re.match(pattern, value):
        raise InputValidationError(
            f"Invalid identifier: '{value}'. Must match pattern: {pattern}"
        )
    return value


def validate_enum(value: str, allowed: set[str], param_name: str = "value") -> str:
    """
    Validate a value is one of the allowed options.

    Args:
        value: Value to validate
        allowed: Set of allowed values
        param_name: Parameter name for error message

    Returns:
        Validated value
    """
    if value not in allowed:
        raise InputValidationError(
            f"Invalid {param_name}: '{value}'. Allowed: {sorted(allowed)}"
        )
    return value


def validate_range(
    value: Union[int, float],
    min_val: Optional[Union[int, float]] = None,
    max_val: Optional[Union[int, float]] = None,
    param_name: str = "value"
) -> Union[int, float]:
    """
    Validate a numeric value is within range.
    """
    if min_val is not None and value < min_val:
        raise InputValidationError(f"{param_name} must be >= {min_val}, got {value}")
    if max_val is not None and value > max_val:
        raise InputValidationError(f"{param_name} must be <= {max_val}, got {value}")
    return value


# ═══════════════════════════════════════════════════════════════════════════════
# RATE LIMITING
# ═══════════════════════════════════════════════════════════════════════════════

class RateLimiter:
    """
    Simple in-memory rate limiter using token bucket algorithm.

    Usage:
        limiter = RateLimiter(calls_per_minute=60)
        if limiter.allow("tool_name"):
            # Execute tool
        else:
            # Rate limited
    """

    def __init__(
        self,
        calls_per_minute: int = 60,
        burst_size: int = 10
    ):
        """
        Initialize rate limiter.

        Args:
            calls_per_minute: Sustained rate limit
            burst_size: Allow short bursts up to this size
        """
        self.rate = calls_per_minute / 60.0  # Calls per second
        self.burst_size = burst_size
        self._buckets: dict[str, dict[str, float]] = defaultdict(
            lambda: {'tokens': burst_size, 'last_update': time.time()}
        )

    def allow(self, key: str = "default") -> bool:
        """
        Check if a call is allowed under rate limits.

        Args:
            key: Identifier for rate limiting (e.g., tool name, user ID)

        Returns:
            True if allowed, False if rate limited
        """
        bucket = self._buckets[key]
        now = time.time()

        # Refill tokens based on time elapsed
        elapsed = now - bucket['last_update']
        bucket['tokens'] = min(
            self.burst_size,
            bucket['tokens'] + elapsed * self.rate
        )
        bucket['last_update'] = now

        # Check if we have tokens
        if bucket['tokens'] >= 1:
            bucket['tokens'] -= 1
            return True

        return False

    def get_wait_time(self, key: str = "default") -> float:
        """Get seconds to wait before next allowed call."""
        bucket = self._buckets[key]
        if bucket['tokens'] >= 1:
            return 0.0
        return (1 - bucket['tokens']) / self.rate


# Global rate limiter
_rate_limiter: Optional[RateLimiter] = None


def get_rate_limiter() -> RateLimiter:
    """Get or create the global rate limiter."""
    global _rate_limiter
    if _rate_limiter is None:
        _rate_limiter = RateLimiter()
    return _rate_limiter


def rate_limit(key: Optional[str] = None):
    """
    Decorator to apply rate limiting to a function.

    Args:
        key: Rate limit key (defaults to function name)
    """
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs):
            limiter = get_rate_limiter()
            limit_key = key or func.__name__

            if not limiter.allow(limit_key):
                wait_time = limiter.get_wait_time(limit_key)
                raise InputValidationError(
                    f"Rate limit exceeded for {limit_key}. "
                    f"Try again in {wait_time:.1f} seconds."
                )

            return func(*args, **kwargs)

        @wraps(func)
        async def async_wrapper(*args, **kwargs):
            limiter = get_rate_limiter()
            limit_key = key or func.__name__

            if not limiter.allow(limit_key):
                wait_time = limiter.get_wait_time(limit_key)
                raise InputValidationError(
                    f"Rate limit exceeded for {limit_key}. "
                    f"Try again in {wait_time:.1f} seconds."
                )

            return await func(*args, **kwargs)

        # Return appropriate wrapper based on function type
        import asyncio
        if asyncio.iscoroutinefunction(func):
            return async_wrapper
        return wrapper

    return decorator


# ═══════════════════════════════════════════════════════════════════════════════
# ACCESS CONTROL
# ═══════════════════════════════════════════════════════════════════════════════

class AccessLevel:
    """Access levels for tool operations."""
    READ = "read"           # Read files, list directories
    WRITE = "write"         # Create/modify files
    EXECUTE = "execute"     # Run commands, scripts
    ADMIN = "admin"         # Destructive operations, config changes


class AccessController:
    """
    Controls what operations are allowed.

    Usage:
        controller = AccessController(default_level=AccessLevel.WRITE)
        controller.deny_tool("delete_project")

        if controller.can_execute("create_task"):
            # Execute
    """

    def __init__(
        self,
        default_level: str = AccessLevel.WRITE,
        denied_tools: Optional[set[str]] = None,
        read_only: bool = False
    ):
        """
        Initialize access controller.

        Args:
            default_level: Default access level for tools
            denied_tools: Set of tool names that are always denied
            read_only: If True, deny all write/execute operations
        """
        self.default_level = default_level
        self.denied_tools = denied_tools or set()
        self.read_only = read_only

        # Tool -> required access level mapping
        self._tool_levels: dict[str, str] = {
            # Read operations
            'check_documentation_health': AccessLevel.READ,
            'analyze_todo2_alignment': AccessLevel.READ,
            'detect_duplicate_tasks': AccessLevel.READ,
            'project_scorecard': AccessLevel.READ,
            'project_overview': AccessLevel.READ,
            'list_tasks_awaiting_clarification': AccessLevel.READ,

            # Write operations
            'sync_todo_tasks': AccessLevel.WRITE,
            'resolve_task_clarification': AccessLevel.WRITE,
            'batch_approve_tasks': AccessLevel.WRITE,

            # Admin operations
            'run_nightly_task_automation': AccessLevel.ADMIN,
            'sprint_automation': AccessLevel.ADMIN,
        }

    def set_tool_level(self, tool_name: str, level: str) -> None:
        """Set required access level for a tool."""
        self._tool_levels[tool_name] = level

    def deny_tool(self, tool_name: str) -> None:
        """Explicitly deny a tool."""
        self.denied_tools.add(tool_name)

    def allow_tool(self, tool_name: str) -> None:
        """Remove tool from denied list."""
        self.denied_tools.discard(tool_name)

    def can_execute(self, tool_name: str) -> bool:
        """Check if a tool can be executed."""
        # Check explicit denials
        if tool_name in self.denied_tools:
            return False

        # Get required level
        required_level = self._tool_levels.get(tool_name, self.default_level)

        # In read-only mode, only allow read operations
        if self.read_only and required_level != AccessLevel.READ:
            return False

        return True

    def check_access(self, tool_name: str) -> None:
        """
        Check access and raise if denied.

        Raises:
            InputValidationError: If access is denied
        """
        if not self.can_execute(tool_name):
            raise InputValidationError(f"Access denied for tool: {tool_name}")


# Global access controller
_access_controller: Optional[AccessController] = None


def get_access_controller() -> AccessController:
    """Get or create the global access controller."""
    global _access_controller
    if _access_controller is None:
        _access_controller = AccessController()
    return _access_controller


def set_access_controller(controller: AccessController) -> None:
    """Set the global access controller."""
    global _access_controller
    _access_controller = controller


def require_access(tool_name: Optional[str] = None):
    """
    Decorator to check access control before executing a tool.

    Args:
        tool_name: Tool name (defaults to function name)
    """
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs):
            controller = get_access_controller()
            name = tool_name or func.__name__
            controller.check_access(name)
            return func(*args, **kwargs)

        @wraps(func)
        async def async_wrapper(*args, **kwargs):
            controller = get_access_controller()
            name = tool_name or func.__name__
            controller.check_access(name)
            return await func(*args, **kwargs)

        import asyncio
        if asyncio.iscoroutinefunction(func):
            return async_wrapper
        return wrapper

    return decorator


# ═══════════════════════════════════════════════════════════════════════════════
# SUBPROCESS SANDBOXING
# ═══════════════════════════════════════════════════════════════════════════════

import subprocess
from typing import Any


class SubprocessSecurityError(Exception):
    """Raised when subprocess security validation fails."""
    pass


# Allowed commands with their allowed arguments
# Format: command -> set of allowed first arguments (None = any)
ALLOWED_COMMANDS = {
    'git': {'status', 'log', 'diff', 'show', 'rev-parse', 'ls-files', 'branch', 'remote', None},
    'python': {'-m', '-c', None},
    'pytest': {'-v', '--version', '--co', '--cov', '--junit-xml', None},
    'pip': {'list', 'show', 'install', '--version', None},
    'npm': {'audit', 'list', '--version', None},
    'cargo': {'audit', '--version', None},
    'uv': {'sync', 'run', 'pip', 'install', '--version', None},
    'uvx': {None},  # uvx can run any package, but we validate the package name
    'pip-audit': {'--version', None},
    'ruff': {'check', 'format', '--version', None},
    'black': {'--version', '--check', None},
    'mypy': {'--version', None},
    # System commands (use with caution)
    'ls': {None},
    'cat': {None},
    'head': {None},
    'tail': {None},
    'grep': {None},
    'find': {None},
    'which': {None},
}


def validate_command(command: list[str], project_root: Optional[Path] = None) -> None:
    """
    Validate a command is safe to execute.

    Args:
        command: Command as list (e.g., ['git', 'status'])
        project_root: Project root for path validation

    Raises:
        SubprocessSecurityError: If command is not allowed
    """
    if not command:
        raise SubprocessSecurityError("Empty command")

    cmd_name = command[0]

    # Check if command is in allowlist
    if cmd_name not in ALLOWED_COMMANDS:
        raise SubprocessSecurityError(
            f"Command '{cmd_name}' is not in the allowlist. "
            f"Allowed commands: {sorted(ALLOWED_COMMANDS.keys())}"
        )

    # Check first argument if provided
    allowed_args = ALLOWED_COMMANDS[cmd_name]
    if len(command) > 1:
        first_arg = command[1]
        # None in allowed_args means any argument is allowed
        if None not in allowed_args and first_arg not in allowed_args:
            raise SubprocessSecurityError(
                f"Command '{cmd_name}' with argument '{first_arg}' is not allowed. "
                f"Allowed arguments: {sorted(allowed_args - {None})}"
            )

    # Additional validation for specific commands
    if cmd_name == 'uvx':
        # For uvx, validate the package name (second arg) is safe
        if len(command) > 1:
            package = command[1]
            # Only allow alphanumeric, hyphens, underscores, dots, and == for versions
            if not re.match(r'^[a-zA-Z0-9._-]+(==[0-9.]+)?$', package):
                raise SubprocessSecurityError(
                    f"Invalid package name for uvx: '{package}'. "
                    "Only alphanumeric, dots, hyphens, underscores, and version specifiers allowed."
                )


def safe_subprocess(
    command: list[str],
    cwd: Optional[Union[str, Path]] = None,
    project_root: Optional[Path] = None,
    timeout: Optional[int] = 300,
    capture_output: bool = True,
    text: bool = True,
    **kwargs: Any
) -> subprocess.CompletedProcess:
    """
    Execute a subprocess command with security validation.

    Validates:
    - Command is in allowlist
    - Working directory is within project boundaries
    - Command arguments are safe

    Args:
        command: Command to execute as list
        cwd: Working directory (must be within project_root)
        project_root: Project root for boundary validation
        timeout: Command timeout in seconds
        capture_output: Capture stdout/stderr
        text: Return text instead of bytes
        **kwargs: Additional subprocess.run() arguments

    Returns:
        CompletedProcess result

    Raises:
        SubprocessSecurityError: If validation fails
        PathBoundaryError: If cwd is outside boundaries
    """
    # Validate command
    validate_command(command, project_root)

    # Validate and resolve working directory
    if cwd:
        if project_root:
            validator = get_default_path_validator()
            safe_cwd = validator.validate(cwd, must_exist=True)
        else:
            safe_cwd = Path(cwd).resolve()
            if not safe_cwd.exists():
                raise SubprocessSecurityError(f"Working directory does not exist: {cwd}")
    else:
        safe_cwd = project_root or Path.cwd()

    # Execute with validated parameters
    try:
        result = subprocess.run(
            command,
            cwd=str(safe_cwd),
            timeout=timeout,
            capture_output=capture_output,
            text=text,
            **kwargs
        )
        return result
    except subprocess.TimeoutExpired as e:
        raise SubprocessSecurityError(f"Command timed out after {timeout}s: {' '.join(command)}") from e
    except Exception as e:
        raise SubprocessSecurityError(f"Subprocess execution failed: {e}") from e


# ═══════════════════════════════════════════════════════════════════════════════
# EXPORTS
# ═══════════════════════════════════════════════════════════════════════════════

__all__ = [
    # Path validation
    'PathBoundaryError',
    'PathValidator',
    'set_default_path_validator',
    'get_default_path_validator',
    'validate_path',

    # Input validation
    'InputValidationError',
    'sanitize_string',
    'validate_identifier',
    'validate_enum',
    'validate_range',

    # Rate limiting
    'RateLimiter',
    'get_rate_limiter',
    'rate_limit',

    # Access control
    'AccessLevel',
    'AccessController',
    'get_access_controller',
    'set_access_controller',
    'require_access',

    # Subprocess security
    'SubprocessSecurityError',
    'validate_command',
    'safe_subprocess',
    'ALLOWED_COMMANDS',
]

