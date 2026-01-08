# project_management_automation/utils/output.py
"""
Output separation utilities for MCP tools.

Separates human-readable formatted output from AI-processable structured data.
Uses FastMCP Context methods (ctx.info(), ctx.report_progress()) to stream
human output without consuming AI tokens.

Design Decision: See docs/DESIGN_DECISIONS.md "Output Separation Pattern"
"""

import json
from collections.abc import AsyncIterator, Iterator
from typing import Any, Optional, Union


async def split_output(
    ctx,
    human: str,
    ai: Any,
    stream_human: bool = False
) -> dict:
    """
    Separate human-readable output from AI-processable data.

    Human output is sent via ctx.info() (visible in client UI but not
    consuming AI context tokens). AI output is returned as structured data.

    Args:
        ctx: FastMCP Context object (must have .info() method)
        human: Formatted text for human consumption (→ ctx.info())
        ai: Structured data for AI (→ return value)
        stream_human: If True, stream human output line by line

    Returns:
        The 'ai' parameter as a dict (wraps non-dict values)

    Example:
        @mcp.tool()
        async def my_tool(ctx: Context) -> str:
            result = do_work()
            return json.dumps(await split_output(ctx,
                human=result['formatted_report'],  # Human sees formatted
                ai={'score': result['score']}       # AI gets structured
            ), separators=(',', ':'))
    """
    if human and hasattr(ctx, 'info'):
        if stream_human:
            for line in human.split('\n'):
                await ctx.info(line)
        else:
            await ctx.info(human)

    # Ensure return is always a dict
    if isinstance(ai, dict):
        return ai
    elif ai is None:
        return {"status": "ok"}
    else:
        return {"result": ai}


async def progress_wrapper(
    ctx,
    iterable: Union[list, Iterator, AsyncIterator],
    total: Optional[int] = None,
    desc: str = "Processing"
) -> AsyncIterator:
    """
    Wrap an iterable with progress reporting via ctx.report_progress().

    Reports progress to the MCP client, visible to users in supported clients.
    Does not affect AI token consumption.

    Args:
        ctx: FastMCP Context object (must have .report_progress() method)
        iterable: Items to iterate (list, iterator, or async iterator)
        total: Total count if known (required for iterators)
        desc: Description shown in progress indicator

    Yields:
        Items from the iterable

    Example:
        @mcp.tool()
        async def scan_files(ctx: Context) -> str:
            files = list(Path('.').rglob('*.py'))
            results = []
            async for f in progress_wrapper(ctx, files, desc="Scanning"):
                results.append(analyze(f))
            return json.dumps({'files_scanned': len(results)})
    """
    # Convert to list if total unknown (needed for percentage calc)
    if total is None:
        if hasattr(iterable, '__len__'):
            items = iterable
            count = len(items)
        else:
            items = list(iterable)
            count = len(items)
    else:
        items = iterable
        count = total

    has_progress = hasattr(ctx, 'report_progress')

    for i, item in enumerate(items):
        if has_progress and count > 0:
            progress = i / count
            await ctx.report_progress(
                progress=progress,
                total=count,
                message=f"{desc}: {i + 1}/{count}"
            )
        yield item

    # Final progress report
    if has_progress and count > 0:
        await ctx.report_progress(
            progress=1.0,
            total=count,
            message=f"{desc}: Complete"
        )


def compact_json(data: Any) -> str:
    """
    Convert data to compact JSON (no whitespace).

    Saves ~20-30% tokens compared to indent=2.

    Args:
        data: JSON-serializable data

    Returns:
        Compact JSON string
    """
    return json.dumps(data, separators=(',', ':'))


# Backwards compatibility alias
async def output_to_human_and_ai(ctx, human: str, ai: Any) -> dict:
    """Alias for split_output (backwards compatibility)."""
    return await split_output(ctx, human, ai)


__all__ = [
    'split_output',
    'progress_wrapper',
    'compact_json',
    'output_to_human_and_ai'
]

