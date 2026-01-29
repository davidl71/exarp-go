"""
Auto-update task status based on codebase analysis.

Uses agentic-tools MCP infer_task_progress to analyze codebase and detect
completed tasks based on code changes, file creation, and implementation evidence.
"""

import json
import logging
from pathlib import Path
from typing import Optional

from ..utils import find_project_root
from ..utils.agentic_tools_client import infer_task_progress_mcp
from ..utils.todo2_utils import get_repo_project_id

logger = logging.getLogger(__name__)


def auto_update_task_status(
    project_id: Optional[str] = None,
    scan_depth: int = 3,
    file_extensions: Optional[list] = None,
    auto_update_tasks: bool = False,
    confidence_threshold: float = 0.7,
    dry_run: bool = False,
    output_path: Optional[str] = None,
    codebase_path: Optional[str] = None,
    project_root: Optional[Path] = None
) -> str:
    """
    Infer task progress by analyzing codebase for implementation evidence.
    
    Analyzes all Todo and In Progress tasks to detect which ones are actually
    completed based on code changes, file creation, and implementation evidence.
    
    Args:
        project_id: Filter to specific project (optional, auto-detected if None)
        scan_depth: Directory depth to scan (1-5, default: 3)
        file_extensions: File types to analyze (default: [.js, .ts, .jsx, .tsx, .py, .java, .cs, .go, .rs])
        auto_update_tasks: Whether to automatically update task status (default: False)
        confidence_threshold: Minimum confidence for auto-updating (0-1, default: 0.7)
        dry_run: Preview mode without making changes (default: False)
        output_path: Path to save report (optional)
        codebase_path: Path to codebase (deprecated, use project_root)
        project_root: Project root path (defaults to find_project_root)
    
    Returns:
        JSON string with analysis results including:
        - total_tasks_analyzed: Number of tasks checked
        - inferences_made: Number of completion inferences
        - tasks_updated: Number of tasks actually updated (if auto_update_tasks=True)
        - inferred_results: List of inferred completions with confidence scores
    """
    if project_root is None:
        project_root = find_project_root()
    
    if codebase_path:
        project_root = Path(codebase_path)
    
    # Auto-detect project_id if not provided
    if project_id is None:
        try:
            project_id = get_repo_project_id(project_root)
        except Exception as e:
            logger.warning(f"Could not auto-detect project_id: {e}")
    
    # Default file extensions if not provided
    if file_extensions is None:
        file_extensions = [".js", ".ts", ".jsx", ".tsx", ".py", ".java", ".cs", ".go", ".rs"]
    
    # Respect dry_run - if dry_run, don't auto-update
    if dry_run:
        auto_update_tasks = False
    
    try:
        # Call agentic-tools MCP infer_task_progress
        result = infer_task_progress_mcp(
            project_id=project_id,
            scan_depth=scan_depth,
            file_extensions=file_extensions,
            auto_update_tasks=auto_update_tasks,
            confidence_threshold=confidence_threshold,
            project_root=project_root
        )
        
        if result is None:
            return json.dumps({
                "success": False,
                "error": "Failed to call infer_task_progress - agentic-tools MCP may not be available"
            }, indent=2)
        
        # Format response
        response = {
            "success": True,
            "data": {
                "total_tasks_analyzed": result.get("total_tasks_analyzed", 0),
                "inferences_made": result.get("inferences_made", 0),
                "tasks_updated": result.get("tasks_updated", 0) if auto_update_tasks else 0,
                "inferred_results": result.get("inferred_results", []),
                "method": "agentic-tools-mcp",
                "dry_run": dry_run,
                "auto_update": auto_update_tasks
            }
        }
        
        # Save report if output_path provided
        if output_path:
            report_path = Path(output_path)
            report_path.parent.mkdir(parents=True, exist_ok=True)
            
            report_content = f"""# Task Completion Check Report

**Generated:** {Path(__file__).stat().st_mtime}

## Summary

- **Total Tasks Analyzed:** {response['data']['total_tasks_analyzed']}
- **Inferences Made:** {response['data']['inferences_made']}
- **Tasks Updated:** {response['data']['tasks_updated']}
- **Method:** {response['data']['method']}
- **Dry Run:** {dry_run}

## Inferred Completions

"""
            for inference in response['data']['inferred_results']:
                task_id = inference.get('task_id', 'Unknown')
                confidence = inference.get('confidence', 0)
                evidence = inference.get('evidence', [])
                report_content += f"### Task {task_id}\n"
                report_content += f"- **Confidence:** {confidence:.1%}\n"
                report_content += f"- **Evidence:** {len(evidence)} items\n"
                for ev in evidence[:3]:  # Show first 3 evidence items
                    report_content += f"  - {ev}\n"
                report_content += "\n"
            
            report_path.write_text(report_content)
            logger.info(f"Report saved to {report_path}")
        
        return json.dumps(response, indent=2)
        
    except Exception as e:
        logger.error(f"Error in auto_update_task_status: {e}", exc_info=True)
        return json.dumps({
            "success": False,
            "error": str(e)
        }, indent=2)

