# project_management_automation/utils/__init__.py
"""
Utility functions for project management automation.

Exports:
    - find_project_root: Locate project root directory
    - split_output: Separate human/AI output for token efficiency
    - progress_wrapper: Progress reporting wrapper
    - compact_json: Compact JSON serialization
    - configure_logging: MCP-aware logging configuration
    - get_logger: Get MCP-aware logger
    - is_mcp_mode: Check if running as MCP server
"""

from .logging_config import configure_logging, get_logger, is_mcp_mode, suppress_noisy_loggers
from .output import compact_json, output_to_human_and_ai, progress_wrapper, split_output
from .project_root import find_project_root
from .wisdom_client import (
    call_wisdom_tool_sync,
    consult_advisor,
    format_text as format_wisdom_text,
    get_daily_briefing,
    get_wisdom,
    list_sources as list_wisdom_sources,
    read_wisdom_resource_sync,
)
from .security import (
    AccessController,
    # Access control
    AccessLevel,
    # Input validation
    InputValidationError,
    # Path validation
    PathBoundaryError,
    PathValidator,
    # Rate limiting
    RateLimiter,
    get_access_controller,
    get_default_path_validator,
    get_rate_limiter,
    rate_limit,
    require_access,
    sanitize_string,
    set_access_controller,
    set_default_path_validator,
    validate_enum,
    validate_identifier,
    validate_path,
    validate_range,
)
from .branch_utils import (
    BRANCH_TAG_PREFIX,
    MAIN_BRANCH,
    create_branch_tag,
    extract_branch_from_tags,
    filter_tasks_by_branch,
    get_all_branch_statistics,
    get_all_branches,
    get_branch_statistics,
    get_task_branch,
    set_task_branch,
)
from .commit_tracking import (
    CommitTracker,
    TaskCommit,
    get_commit_tracker,
    track_task_create,
    track_task_delete,
    track_task_status_change,
    track_task_update,
)
from .todo2_utils import (
    annotate_task_project,
    filter_tasks_by_project,
    get_current_project_id,
    get_repo_project_id,
    load_todo2_project_info,
    normalize_status_to_title_case,
    task_belongs_to_project,
    validate_project_ownership,
)
from .todo2_mcp_client import (
    add_comments_mcp,
    create_todos_mcp,
    delete_todos_mcp,
    get_todo_details_mcp,
    list_todos_mcp,
    update_todos_mcp,
)
from .agentic_tools_client import (
    generate_research_queries_mcp,
    infer_task_progress_mcp,
)


__all__ = [
    # Project utilities
    'find_project_root',
    'split_output',
    'progress_wrapper',
    'compact_json',
    'output_to_human_and_ai',
    # Logging
    'configure_logging',
    'get_logger',
    'is_mcp_mode',
    'suppress_noisy_loggers',
    # Security - Path validation
    'PathBoundaryError',
    'PathValidator',
    'set_default_path_validator',
    'get_default_path_validator',
    'validate_path',
    # Security - Input validation
    'InputValidationError',
    'sanitize_string',
    'validate_identifier',
    'validate_enum',
    'validate_range',
    # Security - Rate limiting
    'RateLimiter',
    'get_rate_limiter',
    'rate_limit',
    # Security - Access control
    'AccessLevel',
    'AccessController',
    'get_access_controller',
    'set_access_controller',
    'require_access',
    'annotate_task_project',
    'filter_tasks_by_project',
    'get_current_project_id',
    'get_repo_project_id',
    'load_todo2_project_info',
    'normalize_status_to_title_case',
    'task_belongs_to_project',
    'validate_project_ownership',
    # Todo2 MCP client
    'list_todos_mcp',
    'create_todos_mcp',
    'update_todos_mcp',
    'get_todo_details_mcp',
    'add_comments_mcp',
    'delete_todos_mcp',
    # Agentic-Tools MCP client
    'infer_task_progress_mcp',
    'generate_research_queries_mcp',
    'get_next_task_recommendation_mcp',
    'parse_prd_mcp',
    'analyze_task_complexity_mcp',
    'research_task_mcp',
    # Branch utilities
    'BRANCH_TAG_PREFIX',
    'MAIN_BRANCH',
    'create_branch_tag',
    'extract_branch_from_tags',
    'filter_tasks_by_branch',
    'get_all_branch_statistics',
    'get_all_branches',
    'get_branch_statistics',
    'get_task_branch',
    'set_task_branch',
    # Commit tracking
    'CommitTracker',
    'TaskCommit',
    'get_commit_tracker',
    'track_task_create',
    'track_task_delete',
    'track_task_status_change',
    'track_task_update',
    # Wisdom MCP client (devwisdom-go)
    'call_wisdom_tool_sync',
    'consult_advisor',
    'format_wisdom_text',
    'get_daily_briefing',
    'get_wisdom',
    'list_wisdom_sources',
    'read_wisdom_resource_sync',
]

