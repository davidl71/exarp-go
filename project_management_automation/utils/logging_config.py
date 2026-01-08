"""
Centralized logging configuration for Exarp.

When running as MCP server, logs go to stderr which Cursor shows as [error].
This module provides MCP-aware logging that:
- Suppresses console output in MCP mode (detected via EXARP_MCP_MODE env var)
- Optionally logs to file for debugging
- Logs normally to console in CLI mode
"""

import logging
import os
import sys
from pathlib import Path
from typing import Optional


def is_mcp_mode() -> bool:
    """Check if running as MCP server."""
    return os.environ.get("EXARP_MCP_MODE", "").lower() in ("1", "true", "yes")


def configure_logging(
    name: str = "exarp",
    level: int = logging.INFO,
    log_file: Optional[Path] = None,
    force_quiet: bool = False
) -> logging.Logger:
    """
    Configure logging appropriately for MCP or CLI mode.

    Args:
        name: Logger name
        level: Log level (default: INFO)
        log_file: Optional file path for logging (useful for debugging MCP)
        force_quiet: Force quiet mode even in CLI

    Returns:
        Configured logger
    """
    logger = logging.getLogger(name)
    logger.setLevel(level)

    # Clear existing handlers to avoid duplicates
    logger.handlers.clear()

    # In MCP mode or force_quiet, only log to file (if specified)
    if is_mcp_mode() or force_quiet:
        if log_file:
            file_handler = logging.FileHandler(log_file)
            file_handler.setFormatter(logging.Formatter(
                "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
            ))
            logger.addHandler(file_handler)
        else:
            # Add NullHandler to prevent "No handler found" warnings
            logger.addHandler(logging.NullHandler())
    else:
        # CLI mode: log to stderr as normal
        console_handler = logging.StreamHandler(sys.stderr)
        console_handler.setFormatter(logging.Formatter(
            "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        ))
        logger.addHandler(console_handler)

        # Also log to file if specified
        if log_file:
            file_handler = logging.FileHandler(log_file)
            file_handler.setFormatter(logging.Formatter(
                "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
            ))
            logger.addHandler(file_handler)

    return logger


def suppress_noisy_loggers():
    """Suppress verbose logs from MCP framework and dependencies."""
    noisy_loggers = [
        "mcp", "mcp.server", "mcp.server.lowlevel",
        "mcp.server.lowlevel.server", "mcp.server.stdio",
        "fastmcp", "httpx", "httpcore", "asyncio"
    ]

    for logger_name in noisy_loggers:
        logging.getLogger(logger_name).setLevel(logging.WARNING)

    # Also suppress any dynamically created MCP loggers
    for logger_name in list(logging.Logger.manager.loggerDict.keys()):
        if any(x in logger_name for x in ['mcp', 'fastmcp', 'stdio']):
            logging.getLogger(logger_name).setLevel(logging.WARNING)


def get_logger(name: str) -> logging.Logger:
    """
    Get a logger configured for MCP/CLI mode.

    Use this instead of logging.getLogger() for consistent behavior.
    """
    logger = logging.getLogger(name)

    # If no handlers, configure with defaults
    if not logger.handlers:
        configure_logging(name)

    return logger


# Export for convenience
__all__ = [
    'is_mcp_mode',
    'configure_logging',
    'suppress_noisy_loggers',
    'get_logger'
]

