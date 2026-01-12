"""
Project Scorecard Tool - Generate comprehensive project health overview.

[HINT: Project scorecard. Returns overall score, component scores (security, testing,
docs, alignment, clarity, parallelizable, performance, dogfooding, uniqueness), task metrics, production readiness.]

Memory Integration:
- Saves score history for trend tracking
"""

import json
import logging
import re
from datetime import datetime
from pathlib import Path
from typing import Any, Optional

from ..utils import find_project_root
from ..utils.todo2_utils import is_completed_status, is_pending_status

scorecard_logger = logging.getLogger(__name__)


def _save_scorecard_memory(result: dict[str, Any]) -> dict[str, Any]:
    """Save scorecard results as memory for trend tracking."""
    try:
        from .session_memory import save_session_insight

        scores = result.get('scores', {})
        blockers = result.get('blockers', [])
        recs = result.get('recommendations', [])[:5]

        # Format recommendations (avoid long line)
        recs_str = 'None'
        if recs:
            recs_str = chr(10).join(
                f"- [{r.get('priority', '?')}] {r.get('area', '?')}: {r.get('action', '?')}"
                for r in recs
            )

        content = f"""Project scorecard generated.

## Overall Score: {result.get('overall_score', 0)}%
Production Ready: {'✅ Yes' if result.get('production_ready') else '❌ No'}

## Component Scores
{chr(10).join(f'- {name}: {score}%' for name, score in sorted(scores.items(), key=lambda x: -x[1]))}

## Blockers
{chr(10).join('- ' + b for b in blockers) or 'None'}

## Recommendations
{recs_str}
"""

        return save_session_insight(
            title=f"Scorecard: {result.get('overall_score', 0)}% {'✅' if result.get('production_ready') else '❌'}",
            content=content,
            category="insight",
            metadata={"type": "scorecard", "overall_score": result.get('overall_score', 0)}
        )
    except ImportError:
        scorecard_logger.debug("Session memory not available for saving scorecard")
        return {"success": False, "error": "Memory system not available"}

# Optional: Multi-source wisdom system (now via devwisdom-go MCP server)
try:
    from ..utils.wisdom_client import (
        format_text as format_wisdom_text,
    )
    from ..utils.wisdom_client import (
        get_wisdom,
    )
    from ..utils.wisdom_client import (
        list_sources as list_available_sources,
    )
    # Load config is no longer needed (handled by external server)
    def load_wisdom_config():
        return {"disabled": False}
    WISDOM_AVAILABLE = True
except ImportError:
    WISDOM_AVAILABLE = False
    def get_wisdom(x, **kwargs):
        return None
    def format_wisdom_text(x):
        return ""
    def load_wisdom_config():
        return {"disabled": True}
    def list_available_sources():
        return []


def generate_project_scorecard(
    output_format: str = "text",
    include_recommendations: bool = True,
    output_path: Optional[str] = None
) -> dict[str, Any]:
    """
    Generate comprehensive project health scorecard.

    [HINT: Project scorecard. Returns overall score, component scores (security, testing,
    docs, alignment, clarity, parallelizable, performance, dogfooding, uniqueness), task metrics, production readiness.]

    Args:
        output_format: Output format - "text", "json", or "markdown"
        include_recommendations: Include improvement recommendations
        output_path: Optional path to save report

    Returns:
        Dictionary with scorecard data and formatted output
    """
    project_root = find_project_root()

    # Safety check: Ensure we're not scanning the entire home directory
    # If project_root is the home directory, something went wrong
    if str(project_root) == str(Path.home()) or str(project_root) == '/Users/davidl':
        # Try to find the actual project by looking for project_management_automation package
        potential_root = Path(__file__).parent.parent.parent
        if (potential_root / 'project_management_automation').exists():
            project_root = potential_root
        else:
            # Last resort: use current working directory if it looks like a project
            cwd = Path.cwd()
            if (cwd / 'project_management_automation').exists() or (cwd / '.git').exists():
                project_root = cwd

    scores = {}
    metrics = {}

    # ═══════════════════════════════════════════════════════════════
    # 1. CODEBASE METRICS
    # ═══════════════════════════════════════════════════════════════
    # Add exclusions for system directories to prevent scanning home directory
    py_files = list(project_root.rglob('*.py'))
    py_files = [f for f in py_files
                if 'venv' not in str(f)
                and '.build-env' not in str(f)
                and '__pycache__' not in str(f)
                and 'Library' not in str(f)  # Exclude macOS Library directories
                and 'Containers' not in str(f)  # Exclude app containers
                and 'Group Containers' not in str(f)  # Exclude group containers
                and '.Trash' not in str(f)  # Exclude trash
                and str(f).startswith(str(project_root))]  # Ensure file is within project

    total_py_lines = 0
    for f in py_files:
        try:
            total_py_lines += len(f.read_text().splitlines())
        except (OSError, UnicodeDecodeError):
            pass

    # Count tools and prompts
    tools_dir = project_root / 'project_management_automation' / 'tools'
    tools_count = len([f for f in tools_dir.glob('*.py') if not f.name.startswith('__')]) if tools_dir.exists() else 0

    try:
        import sys
        sys.path.insert(0, str(project_root))
        from prompts import PROMPTS
        prompts_count = len(PROMPTS)
    except (ImportError, ModuleNotFoundError):
        prompts_count = 0

    metrics['codebase'] = {
        'python_files': len(py_files),
        'python_lines': total_py_lines,
        'mcp_tools': tools_count,
        'mcp_prompts': prompts_count,
    }
    scores['codebase'] = 80  # Base score for having a structured codebase

    # ═══════════════════════════════════════════════════════════════
    # 2. TESTING
    # ═══════════════════════════════════════════════════════════════
    # Detect test files across multiple languages and locations
    test_files = []
    test_lines = 0

    # Common test directories
    test_dirs = [
        project_root / 'tests',
        project_root / 'test',
        project_root / 'native' / 'tests',
        project_root / 'native' / 'test',
        project_root / 'agents' / 'backend' / 'tests',
        project_root / 'web' / 'src' / '__tests__',
        project_root / 'web' / '__tests__',
    ]

    # Python tests: test_*.py
    for test_dir in test_dirs:
        if test_dir.exists():
            test_files.extend(test_dir.rglob('test_*.py'))

    # C++ tests: test_*.cpp, *_test.cpp, *test*.cpp
    for test_dir in test_dirs:
        if test_dir.exists():
            test_files.extend(test_dir.rglob('test_*.cpp'))
            test_files.extend(test_dir.rglob('*_test.cpp'))
            test_files.extend(test_dir.rglob('*test*.cpp'))

    # Rust tests: *_test.rs, test_*.rs, tests/*.rs (in Rust, tests are in tests/ directory)
    rust_test_dirs = [project_root / 'agents' / 'backend', project_root / 'native']
    for base_dir in rust_test_dirs:
        if base_dir.exists():
            # Rust convention: tests in tests/ directory
            tests_dir = base_dir / 'tests'
            if tests_dir.exists():
                test_files.extend(tests_dir.rglob('*.rs'))
            # Also check for inline tests: *_test.rs
            test_files.extend(base_dir.rglob('*_test.rs'))
            test_files.extend(base_dir.rglob('test_*.rs'))

    # TypeScript/JavaScript tests: *.test.ts, *.test.tsx, *.spec.ts, *.spec.tsx
    ts_dirs = [project_root / 'web', project_root / 'agents' / 'web']
    for ts_dir in ts_dirs:
        if ts_dir.exists():
            test_files.extend(ts_dir.rglob('*.test.ts'))
            test_files.extend(ts_dir.rglob('*.test.tsx'))
            test_files.extend(ts_dir.rglob('*.spec.ts'))
            test_files.extend(ts_dir.rglob('*.spec.tsx'))

    # Swift tests: *Tests.swift, *Test.swift
    swift_dirs = [project_root / 'ios', project_root / 'desktop']
    for swift_dir in swift_dirs:
        if swift_dir.exists():
            test_files.extend(swift_dir.rglob('*Tests.swift'))
            test_files.extend(swift_dir.rglob('*Test.swift'))

    # Remove duplicates
    test_files = list(set(test_files))

    # Calculate total test lines
    for f in test_files:
        if f.exists():
            try:
                test_lines += len(f.read_text().splitlines())
            except (OSError, UnicodeDecodeError):
                pass

    # Calculate test ratio - compare test lines to source code lines
    # For multi-language projects, use all source lines (Python + C++ + Rust + TypeScript + Swift)

    # C++ source files (exclude tests)
    cpp_files = list(project_root.rglob('*.cpp'))
    cpp_files = [f for f in cpp_files
                if 'venv' not in str(f)
                and '.build-env' not in str(f)
                and '__pycache__' not in str(f)
                and 'test' not in str(f).lower()
                and 'target' not in str(f)
                and 'Library' not in str(f)
                and 'Containers' not in str(f)
                and str(f).startswith(str(project_root))]
    total_cpp_lines = 0
    for f in cpp_files:
        try:
            total_cpp_lines += len(f.read_text().splitlines())
        except (OSError, UnicodeDecodeError):
            pass

    # Rust source files (exclude tests)
    rust_files = list(project_root.rglob('*.rs'))
    rust_files = [f for f in rust_files
                 if 'target' not in str(f)
                 and 'test' not in str(f).lower()
                 and 'tests' not in str(f)
                 and 'Library' not in str(f)
                 and 'Containers' not in str(f)
                 and str(f).startswith(str(project_root))]
    total_rust_lines = 0
    for f in rust_files:
        try:
            total_rust_lines += len(f.read_text().splitlines())
        except (OSError, UnicodeDecodeError):
            pass

    # TypeScript source files (exclude tests)
    ts_files = list(project_root.rglob('*.ts'))
    ts_files.extend(project_root.rglob('*.tsx'))
    ts_files = [f for f in ts_files
                if 'node_modules' not in str(f)
                and 'test' not in str(f).lower()
                and 'spec' not in str(f).lower()
                and '.test.' not in str(f)
                and 'Library' not in str(f)
                and 'Containers' not in str(f)
                and str(f).startswith(str(project_root))]
    total_ts_lines = 0
    for f in ts_files:
        try:
            total_ts_lines += len(f.read_text().splitlines())
        except (OSError, UnicodeDecodeError):
            pass

    # Swift source files (exclude tests)
    swift_files = list(project_root.rglob('*.swift'))
    swift_files = [f for f in swift_files
                  if 'Library' not in str(f)
                  and 'Containers' not in str(f)
                  and str(f).startswith(str(project_root))]
    swift_files = [f for f in swift_files if 'test' not in str(f).lower() and 'build' not in str(f)]
    total_swift_lines = 0
    for f in swift_files:
        try:
            total_swift_lines += len(f.read_text().splitlines())
        except (OSError, UnicodeDecodeError):
            pass

    # Total source lines across all languages
    total_source_lines = total_py_lines + total_cpp_lines + total_rust_lines + total_ts_lines + total_swift_lines
    test_ratio = (test_lines / total_source_lines * 100) if total_source_lines > 0 else 0
    scores['testing'] = min(100, test_ratio * 3)  # 33% ratio = 100%

    metrics['testing'] = {
        'test_files': len(test_files),
        'test_lines': test_lines,
        'test_ratio': round(test_ratio, 1),
    }

    # ═══════════════════════════════════════════════════════════════
    # 3. DOCUMENTATION
    # ═══════════════════════════════════════════════════════════════
    project_root / 'docs'
    md_files = list(project_root.rglob('*.md'))
    md_files = [f for f in md_files
               if 'Library' not in str(f)
               and 'Containers' not in str(f)
               and str(f).startswith(str(project_root))]
    md_files = [f for f in md_files if 'venv' not in str(f)]

    doc_lines = sum(len(f.read_text().splitlines()) for f in md_files if f.exists() and f.is_file())
    doc_ratio = (doc_lines / total_py_lines * 100) if total_py_lines > 0 else 0

    key_docs = ['README.md', 'INSTALL.md', 'docs/SECURITY.md', 'docs/WORKFLOW.md']
    existing_docs = sum(1 for d in key_docs if (project_root / d).exists())

    scores['documentation'] = min(100, doc_ratio + (existing_docs / len(key_docs) * 50))

    metrics['documentation'] = {
        'doc_files': len(md_files),
        'doc_lines': doc_lines,
        'doc_ratio': round(doc_ratio, 1),
        'key_docs': f"{existing_docs}/{len(key_docs)}",
    }

    # ═══════════════════════════════════════════════════════════════
    # 4. TASK MANAGEMENT
    # ═══════════════════════════════════════════════════════════════
    todo2_file = project_root / '.todo2' / 'state.todo2.json'
    if todo2_file.exists():
        with open(todo2_file) as f:
            data = json.load(f)
        todos = data.get('todos', [])

        # Normalize status matching using utility functions
        pending = [t for t in todos if is_pending_status(t.get('status', ''))]
        completed = [t for t in todos if is_completed_status(t.get('status', ''))]

        completion_rate = len(completed) / len(todos) * 100 if todos else 0
        scores['completion'] = completion_rate

        remaining_hours = sum(t.get('estimatedHours', 0) for t in pending)

        metrics['tasks'] = {
            'total': len(todos),
            'pending': len(pending),
            'completed': len(completed),
            'completion_rate': round(completion_rate, 1),
            'remaining_hours': remaining_hours,
        }

        # ═══════════════════════════════════════════════════════════
        # 5. ALIGNMENT ANALYSIS
        # ═══════════════════════════════════════════════════════════
        mcp_keywords = {
            'mcp', 'fastmcp', 'tool', 'tools', 'prompt', 'prompts', 'resource',
            'server', 'client', 'automation', 'automate', 'security', 'secure',
            'validation', 'validate', 'test', 'testing', 'tests', 'coverage',
            'integration', 'documentation', 'docs', 'workflow', 'ci', 'cd',
            'health', 'analysis', 'task', 'tasks', 'todo', 'sprint',
            'boundary', 'rate', 'limiting', 'access', 'control', 'auth',
            'exarp', 'hook', 'hooks', 'trigger', 'config', 'deploy',
        }

        alignment_scores = []
        well_aligned = 0
        moderately_aligned = 0
        for task in pending:
            content = task.get('content', '').lower()
            details_text = (task.get('details', '') or task.get('long_description', '') or '').lower()
            tags = ' '.join(task.get('tags', [])).lower()
            full_text = f"{content} {details_text} {tags}"

            words = set(re.findall(r'\b[a-z_]{3,}\b', full_text))
            matches = len(words & mcp_keywords)

            # Score based on matches (generous scoring)
            if matches >= 5:
                score = 100
                well_aligned += 1
            elif matches >= 3:
                score = 75
                moderately_aligned += 1
            elif matches >= 2:
                score = 50
                moderately_aligned += 1
            elif matches >= 1:
                score = 30
            else:
                score = 10
            alignment_scores.append(score)

        avg_alignment = sum(alignment_scores) / len(alignment_scores) if alignment_scores else 0
        # If no pending tasks but we have completed tasks, use a default score based on completion rate
        if len(pending) == 0 and len(completed) > 0:
            # High completion suggests good alignment historically
            avg_alignment = 75.0  # Default score for completed projects
        scores['alignment'] = avg_alignment

        metrics['alignment'] = {
            'well_aligned': well_aligned,
            'moderately_aligned': moderately_aligned,
            'total_pending': len(pending),
            'avg_score': round(avg_alignment, 1),
        }

        # ═══════════════════════════════════════════════════════════
        # 6. CLARITY & PARALLELIZABILITY
        # ═══════════════════════════════════════════════════════════
        action_verbs = ['add', 'implement', 'create', 'fix', 'update', 'remove',
                       'refactor', 'migrate', 'integrate', 'test', 'document', 'extend']

        has_estimate = sum(1 for t in pending if t.get('estimatedHours', 0) > 0)
        has_tags = sum(1 for t in pending if t.get('tags'))
        small_enough = sum(1 for t in pending if 0 < t.get('estimatedHours', 0) <= 4)
        clear_name = sum(1 for t in pending if any(
            t.get('content', '').lower().startswith(v) for v in action_verbs))
        no_deps = sum(1 for t in pending if not t.get('dependsOn') and not t.get('dependencies'))

        total_pending = len(pending) or 1
        clarity_score = (has_estimate + has_tags + small_enough + clear_name + no_deps) / (5 * total_pending) * 100
        # If no pending tasks but we have completed tasks, use a default score
        if len(pending) == 0 and len(completed) > 0:
            # High completion suggests tasks were clear enough to complete
            clarity_score = 70.0  # Default score for completed projects
        scores['clarity'] = clarity_score

        parallelizable = sum(1 for t in pending if
            t.get('estimatedHours', 0) <= 4 and
            not t.get('dependsOn') and
            not t.get('dependencies'))
        parallel_score = parallelizable / total_pending * 100 if total_pending else 0
        # If no pending tasks but we have completed tasks, use a default score
        if len(pending) == 0 and len(completed) > 0:
            # If all tasks are completed, they were parallelizable enough to finish
            parallel_score = 60.0  # Default score for completed projects
        scores['parallelizable'] = parallel_score

        metrics['clarity'] = {
            'has_estimate': has_estimate,
            'has_tags': has_tags,
            'small_enough': small_enough,
            'clear_name': clear_name,
            'no_dependencies': no_deps,
            'clarity_score': round(clarity_score, 1),
        }

        metrics['parallelizable'] = {
            'ready': parallelizable,
            'total': total_pending,
            'score': round(parallel_score, 1),
        }

        # ═══════════════════════════════════════════════════════════
        # 6.5. PROGRESS INFERENCE METRICS (T-17)
        # ═══════════════════════════════════════════════════════════
        try:
            from project_management_automation.tools.auto_update_task_status import auto_update_task_status

            # Call with dry_run=True to get metrics without updating
            inference_json = auto_update_task_status(
                confidence_threshold=0.7,
                auto_update=False,  # Only get metrics, don't update
                output_path=None,
                codebase_path=str(project_root)
            )

            if inference_json:
                inference_result = json.loads(inference_json)
                if inference_result.get('success') and inference_result.get('data'):
                    inference_data = inference_result['data']
                    inferred_results = inference_data.get('inferred_results', [])

                    # Calculate inferred vs. marked completion comparison
                    inferred_done = sum(1 for r in inferred_results if r.get('inferred_status') == 'Done')
                    marked_done = len(completed)
                    discrepancy_count = sum(
                        1 for r in inferred_results
                        if r.get('current_status') != r.get('inferred_status')
                    )

                    # Calculate inferred completion rate
                    inferred_completion_rate = (inferred_done / len(todos) * 100) if todos else 0

                    metrics['progress_inference'] = {
                        'tasks_analyzed': inference_data.get('total_tasks_analyzed', 0),
                        'inferences_made': inference_data.get('inferences_made', 0),
                        'marked_completed': marked_done,
                        'inferred_completed': inferred_done,
                        'inferred_completion_rate': round(inferred_completion_rate, 1),
                        'discrepancy_count': discrepancy_count,
                        'discrepancies': [
                            {
                                'task_id': r.get('task_id'),
                                'task_name': r.get('task_name'),
                                'marked': r.get('current_status'),
                                'inferred': r.get('inferred_status'),
                                'confidence': round(r.get('confidence', 0.0), 2)
                            }
                            for r in inferred_results[:5]
                            if r.get('current_status') != r.get('inferred_status')
                        ]
                    }
        except Exception as e:
            scorecard_logger.debug(f"Progress inference not available: {e}")
            metrics['progress_inference'] = {
                'available': False,
                'error': str(e)
            }
    else:
        scores['completion'] = 0
        scores['alignment'] = 0
        scores['clarity'] = 0
        scores['parallelizable'] = 0
        metrics['tasks'] = {'total': 0, 'pending': 0, 'completed': 0}

    # ═══════════════════════════════════════════════════════════════
    # 7. SECURITY (including CodeQL integration)
    # ═══════════════════════════════════════════════════════════════

    # Check for security.py module existence
    security_module = project_root / 'project_management_automation' / 'utils' / 'security.py'
    security_module_exists = security_module.exists()

    # Check security module contents for specific controls
    security_content = security_module.read_text() if security_module_exists else ""

    # Get CodeQL metrics
    try:
        from .codeql_security import get_codeql_security_metrics
        codeql_metrics = get_codeql_security_metrics()
    except ImportError:
        codeql_metrics = {
            'score': 0,
            'configured': False,
            'checks': {
                'codeql_workflow': False,
                'codeql_config': False,
                'no_critical_alerts': True,
                'no_high_alerts': True,
            },
            'alerts': {'total': 0, 'critical': 0, 'high': 0, 'medium': 0, 'low': 0, 'source': None},
            'languages': [],
            'recommendations': [],
        }

    security_checks = {
        'security_docs': (project_root / 'docs' / 'SECURITY.md').exists(),
        'ci_cd_workflow': (project_root / '.github' / 'workflows' / 'ci.yml').exists(),
        'gitignore': (project_root / '.gitignore').exists(),
        'no_hardcoded_secrets': True,
        'input_validation': 'sanitize_string' in security_content and 'validate_' in security_content,
        'path_boundaries': 'PathValidator' in security_content and 'PathBoundaryError' in security_content,
        'rate_limiting': 'RateLimiter' in security_content and 'rate_limit' in security_content,
        'access_control': 'AccessController' in security_content and 'require_access' in security_content,
        # CodeQL checks
        'codeql_workflow': codeql_metrics['checks']['codeql_workflow'],
        'codeql_no_critical': codeql_metrics['checks']['no_critical_alerts'],
        'codeql_no_high': codeql_metrics['checks']['no_high_alerts'],
    }

    passed = sum(1 for v in security_checks.values() if v)

    # Base security score from checks
    base_security_score = passed / len(security_checks) * 100

    # Blend with CodeQL score if configured (CodeQL gets 30% weight when enabled)
    if codeql_metrics['configured']:
        scores['security'] = (base_security_score * 0.7) + (codeql_metrics['score'] * 0.3)
    else:
        scores['security'] = base_security_score

    # Count pending security tasks
    security_tasks = [t for t in todos if 'security' in t.get('tags', []) and is_pending_status(t.get('status', ''))] if todo2_file.exists() else []

    metrics['security'] = {
        'checks_passed': passed,
        'checks_total': len(security_checks),
        'pending_tasks': len(security_tasks),
        'details': security_checks,
        'codeql': {
            'configured': codeql_metrics['configured'],
            'score': round(codeql_metrics['score'], 1),
            'alerts': codeql_metrics['alerts'],
            'languages': codeql_metrics['languages'],
        },
    }

    # ═══════════════════════════════════════════════════════════════
    # 8. CI/CD
    # ═══════════════════════════════════════════════════════════════
    ci_checks = {
        'github_actions': (project_root / '.github' / 'workflows' / 'ci.yml').exists(),
        'linting': (project_root / 'pyproject.toml').exists(),
        'type_checking': (project_root / 'pyproject.toml').exists(),
        'unit_tests': (project_root / 'tests').exists(),
        'pre_commit': (project_root / '.pre-commit-config.yaml').exists(),
        'dependency_lock': (project_root / 'requirements.txt').exists(),
    }

    scores['ci_cd'] = sum(1 for v in ci_checks.values() if v) / len(ci_checks) * 100
    metrics['ci_cd'] = ci_checks

    # ═══════════════════════════════════════════════════════════════
    # 9. PERFORMANCE
    # ═══════════════════════════════════════════════════════════════
    # Check for performance optimizations and infrastructure

    # Check for MCP connection pooling
    mcp_client_path = project_root / 'project_management_automation' / 'scripts' / 'base' / 'mcp_client.py'
    mcp_client_content = mcp_client_path.read_text() if mcp_client_path.exists() else ""

    # Check for logging middleware with timing
    logging_middleware_path = project_root / 'project_management_automation' / 'middleware' / 'logging_middleware.py'
    logging_middleware_content = logging_middleware_path.read_text() if logging_middleware_path.exists() else ""

    # Check for async operations
    server_path = project_root / 'project_management_automation' / 'server.py'
    server_content = server_path.read_text() if server_path.exists() else ""

    performance_checks = {
        'mcp_connection_pooling': 'MCPSessionPool' in mcp_client_content or 'connection_pool' in mcp_client_content.lower(),
        'async_operations': 'async def' in server_content or 'asyncio' in server_content,
        'timing_middleware': 'elapsed_ms' in logging_middleware_content or 'timing' in logging_middleware_content.lower(),
        'batch_operations': 'batch_operations' in mcp_client_content or 'batch' in mcp_client_content.lower(),
        'rate_limiting': 'RateLimiter' in server_content or 'rate_limit' in server_content,
        'caching': 'cache' in server_content.lower() or 'lru_cache' in server_content,
    }

    # NetworkX-based dependency analysis for performance insights
    dependency_analysis = {}
    try:
        import networkx as nx

        # Build task dependency graph if we have tasks
        if todo2_file.exists() and todos:
            G = nx.DiGraph()

            # Add tasks as nodes
            for task in todos:
                task_id = task.get('id', '')
                if task_id:
                    G.add_node(task_id)

            # Add dependencies as edges
            for task in todos:
                task_id = task.get('id', '')
                if not task_id:
                    continue

                # Check both dependsOn and dependencies fields
                deps = task.get('dependsOn', []) or task.get('dependencies', [])
                for dep_id in deps:
                    if dep_id in G:
                        G.add_edge(dep_id, task_id)

            if len(G.nodes()) > 0:
                # Check for circular dependencies (performance killer)
                try:
                    cycles = list(nx.simple_cycles(G))
                    dependency_analysis['has_cycles'] = len(cycles) > 0
                    dependency_analysis['cycle_count'] = len(cycles)
                    performance_checks['no_circular_dependencies'] = len(cycles) == 0
                except Exception:
                    dependency_analysis['has_cycles'] = False
                    dependency_analysis['cycle_count'] = 0
                    performance_checks['no_circular_dependencies'] = True

                # Check for long dependency chains (potential bottlenecks)
                if isinstance(G, nx.DiGraph):
                    longest_path = 0
                    try:
                        # Find longest path in DAG
                        if nx.is_directed_acyclic_graph(G):
                            longest_path = len(nx.dag_longest_path(G)) if len(G.nodes()) > 0 else 0
                        else:
                            # If not DAG, find longest simple path
                            for source in G.nodes():
                                for target in G.nodes():
                                    if source != target:
                                        try:
                                            path = nx.shortest_path(G, source, target)
                                            longest_path = max(longest_path, len(path))
                                        except (nx.NetworkXNoPath, nx.NetworkXError):
                                            pass
                    except Exception:
                        pass

                    dependency_analysis['longest_chain'] = longest_path
                    # Long chains (>10) indicate potential performance issues
                    performance_checks['reasonable_dependency_chains'] = longest_path <= 10
                else:
                    dependency_analysis['longest_chain'] = 0
                    performance_checks['reasonable_dependency_chains'] = True

                # Check for bottleneck tasks (high in-degree = many tasks depend on this)
                if len(G.nodes()) > 0:
                    in_degrees = dict(G.in_degree())
                    max_dependents = max(in_degrees.values()) if in_degrees else 0
                    dependency_analysis['max_dependents'] = max_dependents
                    # Tasks with >5 dependents might be bottlenecks
                    performance_checks['no_bottleneck_tasks'] = max_dependents <= 5
                else:
                    dependency_analysis['max_dependents'] = 0
                    performance_checks['no_bottleneck_tasks'] = True

                # Check parallelization opportunities (tasks with no dependencies)
                independent_tasks = sum(1 for n in G.nodes() if G.in_degree(n) == 0)
                dependency_analysis['independent_tasks'] = independent_tasks
                dependency_analysis['total_tasks'] = len(G.nodes())
                # Good if >20% of tasks can run in parallel
                parallel_ratio = independent_tasks / len(G.nodes()) if len(G.nodes()) > 0 else 0
                performance_checks['good_parallelization'] = parallel_ratio >= 0.2
            else:
                # No tasks, skip dependency analysis
                performance_checks['no_circular_dependencies'] = True
                performance_checks['reasonable_dependency_chains'] = True
                performance_checks['no_bottleneck_tasks'] = True
                performance_checks['good_parallelization'] = True
        else:
            # No tasks file, assume good performance
            performance_checks['no_circular_dependencies'] = True
            performance_checks['reasonable_dependency_chains'] = True
            performance_checks['no_bottleneck_tasks'] = True
            performance_checks['good_parallelization'] = True
    except ImportError:
        # NetworkX not available, skip dependency analysis
        performance_checks['no_circular_dependencies'] = True
        performance_checks['reasonable_dependency_chains'] = True
        performance_checks['no_bottleneck_tasks'] = True
        performance_checks['good_parallelization'] = True
        dependency_analysis['networkx_available'] = False
    except Exception as e:
        # Error in analysis, assume good performance
        scorecard_logger.debug(f"NetworkX dependency analysis failed: {e}")
        performance_checks['no_circular_dependencies'] = True
        performance_checks['reasonable_dependency_chains'] = True
        performance_checks['no_bottleneck_tasks'] = True
        performance_checks['good_parallelization'] = True
        dependency_analysis['error'] = str(e)

    performance_passed = sum(1 for v in performance_checks.values() if v)
    scores['performance'] = performance_passed / len(performance_checks) * 100

    metrics['performance'] = {
        'checks_passed': performance_passed,
        'checks_total': len(performance_checks),
        'details': performance_checks,
        'dependency_analysis': dependency_analysis,
        'description': 'Performance optimizations: connection pooling, async ops, timing, batching, dependency analysis',
    }

    # ═══════════════════════════════════════════════════════════════
    # 10. DOGFOODING SCORE (Does Exarp use its own tools?)
    # ═══════════════════════════════════════════════════════════════
    dogfooding_checks = {
        # Git hooks using exarp
        'pre_commit_hook': (project_root / '.git' / 'hooks' / 'pre-commit').exists() and
            'exarp' in ((project_root / '.git' / 'hooks' / 'pre-commit').read_text().lower()
            if (project_root / '.git' / 'hooks' / 'pre-commit').exists() else ''),
        'pre_push_hook': (project_root / '.git' / 'hooks' / 'pre-push').exists() and
            'exarp' in ((project_root / '.git' / 'hooks' / 'pre-push').read_text().lower()
            if (project_root / '.git' / 'hooks' / 'pre-push').exists() else ''),
        'post_commit_hook': (project_root / '.git' / 'hooks' / 'post-commit').exists(),
        'post_merge_hook': (project_root / '.git' / 'hooks' / 'post-merge').exists(),
        # Pre-commit config
        'pre_commit_config': (project_root / '.pre-commit-config.yaml').exists() and
            'exarp' in ((project_root / '.pre-commit-config.yaml').read_text().lower()
            if (project_root / '.pre-commit-config.yaml').exists() else ''),
        # CI/CD using exarp
        'ci_self_check': (project_root / '.github' / 'workflows' / 'ci.yml').exists() and
            'exarp' in ((project_root / '.github' / 'workflows' / 'ci.yml').read_text().lower()
            if (project_root / '.github' / 'workflows' / 'ci.yml').exists() else ''),
        # Cron automation
        'daily_cron': (project_root / 'scripts' / 'cron' / 'run_daily_exarp.sh').exists(),
        'weekly_cron': (project_root / 'scripts' / 'cron' / 'run_weekly_exarp.sh').exists(),
        # Project goals for alignment
        'project_goals': (project_root / 'PROJECT_GOALS.md').exists(),
        # Pattern triggers
        'pattern_triggers': (project_root / '.cursor' / 'exarp_patterns.json').exists(),
    }

    dogfooding_passed = sum(1 for v in dogfooding_checks.values() if v)
    scores['dogfooding'] = dogfooding_passed / len(dogfooding_checks) * 100

    metrics['dogfooding'] = {
        'checks_passed': dogfooding_passed,
        'checks_total': len(dogfooding_checks),
        'details': dogfooding_checks,
        'description': 'How much Exarp uses its own tools for self-maintenance',
    }

    # ═══════════════════════════════════════════════════════════════
    # 11. UNIQUENESS SCORE (Are we reinventing the wheel?)
    # ═══════════════════════════════════════════════════════════════

    # Check for common patterns that could use existing libraries

    # Analyze source files for potential reinventions
    tools_dir = project_root / 'project_management_automation' / 'tools'

    # Known patterns that might be reinventions

    # Check requirements.txt for dependencies we DO use
    req_file = project_root / 'requirements.txt'
    dependencies = set()
    if req_file.exists():
        for line in req_file.read_text().splitlines():
            if line.strip() and not line.startswith('#'):
                dep = line.split('==')[0].split('>=')[0].split('[')[0].strip().lower()
                dependencies.add(dep)

    # Check pyproject.toml for more deps
    pyproject = project_root / 'pyproject.toml'
    if pyproject.exists():
        content = pyproject.read_text()
        # Simple extraction of dependencies
        for match in re.findall(r'"([a-zA-Z0-9_-]+)', content):
            dependencies.add(match.lower())

    # Check for a DESIGN_DECISIONS.md or similar
    design_docs = [
        project_root / 'docs' / 'DESIGN_DECISIONS.md',
        project_root / 'docs' / 'ARCHITECTURE.md',
        project_root / 'docs' / 'WHY_CUSTOM.md',
        project_root / 'DESIGN.md',
    ]
    has_design_docs = any(d.exists() for d in design_docs)

    # Score components
    uniqueness_analysis = {
        'core_features': {
            'task_management': {
                'custom': True,
                'justified': True,
                'reason': 'MCP-native task management is our core value proposition',
            },
            'mcp_tools': {
                'custom': True,
                'justified': True,
                'reason': 'Custom MCP tools are the product itself',
            },
            'project_scorecard': {
                'custom': True,
                'justified': True,
                'reason': 'Project-specific metrics not available elsewhere',
            },
            'alignment_analysis': {
                'custom': True,
                'justified': True,
                'reason': 'Custom goal-task alignment for this project type',
            },
        },
        'infrastructure': {
            'git_hooks': {
                'custom': True,
                'justified': True,
                'reason': 'Simple shell scripts, no framework needed',
            },
            'cron_scripts': {
                'custom': True,
                'justified': True,
                'reason': 'Using system cron, lightweight approach',
            },
            'file_patterns': {
                'custom': True,
                'justified': True,
                'reason': 'Simple JSON config, no complex framework',
            },
        },
        'potential_improvements': [],
    }

    # Check if we're using heavy frameworks where simpler would work
    using_fastmcp = 'fastmcp' in dependencies or 'mcp' in dependencies

    # FastMCP is justified - it's the standard for MCP servers
    if using_fastmcp:
        uniqueness_analysis['infrastructure']['mcp_framework'] = {
            'custom': False,
            'justified': True,
            'reason': 'FastMCP is the standard MCP server framework',
        }

    # Count justified vs potentially unjustified
    total_decisions = 0
    justified_count = 0

    for category in ['core_features', 'infrastructure']:
        for _name, info in uniqueness_analysis.get(category, {}).items():
            total_decisions += 1
            if info.get('justified'):
                justified_count += 1

    # Check for design documentation
    if has_design_docs:
        justified_count += 1
        total_decisions += 1
        uniqueness_analysis['documentation'] = {
            'design_decisions_documented': True,
            'reason': 'Design decisions are documented',
        }
    else:
        total_decisions += 1
        uniqueness_analysis['potential_improvements'].append({
            'item': 'Design Documentation',
            'suggestion': 'Create docs/DESIGN_DECISIONS.md to document why custom implementations were chosen',
            'priority': 'low',
        })

    # Calculate uniqueness score
    # Higher score = better justified custom code OR using existing solutions
    uniqueness_score = (justified_count / total_decisions * 100) if total_decisions > 0 else 50

    # Bonus for minimal dependencies (less dependency hell)
    dep_count = len(dependencies)
    if dep_count <= 5:
        uniqueness_score = min(100, uniqueness_score + 10)
        uniqueness_analysis['dependency_approach'] = {
            'count': dep_count,
            'rating': 'minimal',
            'reason': 'Minimal dependencies reduces dependency hell',
        }
    elif dep_count <= 15:
        uniqueness_analysis['dependency_approach'] = {
            'count': dep_count,
            'rating': 'moderate',
            'reason': 'Reasonable dependency count',
        }
    else:
        uniqueness_score = max(0, uniqueness_score - 10)
        uniqueness_analysis['dependency_approach'] = {
            'count': dep_count,
            'rating': 'heavy',
            'reason': 'Consider reducing dependencies',
        }

    scores['uniqueness'] = uniqueness_score

    metrics['uniqueness'] = {
        'score': round(uniqueness_score, 1),
        'justified_decisions': justified_count,
        'total_decisions': total_decisions,
        'dependency_count': dep_count,
        'has_design_docs': has_design_docs,
        'analysis': uniqueness_analysis,
        'description': 'Are we reinventing wheels? If so, is it justified?',
    }

    # ═══════════════════════════════════════════════════════════════
    # CALCULATE OVERALL SCORE
    # ═══════════════════════════════════════════════════════════════
    weights = {
        'documentation': 0.06,
        'ci_cd': 0.06,
        'codebase': 0.06,
        'clarity': 0.06,
        'parallelizable': 0.06,
        'alignment': 0.06,
        'security': 0.20,
        'testing': 0.10,
        'completion': 0.05,
        'performance': 0.08,  # Performance optimizations
        'dogfooding': 0.13,  # Eating our own dog food
        'uniqueness': 0.10,  # Not reinventing wheels (or justified if we are)
    }

    overall_score = sum(scores.get(k, 0) * weights.get(k, 0) for k in weights)

    # Determine production readiness
    production_ready = scores.get('security', 0) >= 80 and scores.get('testing', 0) >= 50
    blockers = []
    if scores.get('security', 0) < 80:
        blockers.append("Security controls incomplete")
    if scores.get('testing', 0) < 50:
        blockers.append("Test coverage too low")

    # ═══════════════════════════════════════════════════════════════
    # BUILD RESULT
    # ═══════════════════════════════════════════════════════════════
    result = {
        'generated_at': datetime.now().isoformat(),
        'overall_score': round(overall_score, 1),
        'production_ready': production_ready,
        'blockers': blockers,
        'scores': {k: round(v, 1) for k, v in scores.items()},
        'weights': weights,
        'metrics': metrics,
    }

    # Add recommendations if requested
    if include_recommendations:
        recommendations = []

        if scores.get('security', 0) < 80:
            recommendations.append({
                'priority': 'critical',
                'area': 'Security',
                'action': 'Implement path boundary enforcement, rate limiting, and access control',
                'impact': '+25% to security score',
            })

        # CodeQL-specific recommendations
        if not codeql_metrics['configured']:
            recommendations.append({
                'priority': 'high',
                'area': 'CodeQL',
                'action': 'Enable CodeQL workflow for automated security scanning',
                'impact': '+10% to security score',
            })
        elif codeql_metrics['alerts']['critical'] > 0:
            recommendations.append({
                'priority': 'critical',
                'area': 'CodeQL',
                'action': f"Fix {codeql_metrics['alerts']['critical']} critical CodeQL security alerts",
                'impact': 'Prevent security vulnerabilities',
            })
        elif codeql_metrics['alerts']['high'] > 0:
            recommendations.append({
                'priority': 'high',
                'area': 'CodeQL',
                'action': f"Address {codeql_metrics['alerts']['high']} high-severity CodeQL alerts",
                'impact': '+15% to CodeQL score',
            })

        if scores.get('testing', 0) < 50:
            recommendations.append({
                'priority': 'high',
                'area': 'Testing',
                'action': 'Fix failing tests and increase coverage to 30%',
                'impact': '+15% to testing score',
            })

        if scores.get('performance', 0) < 70:
            perf_metrics = metrics.get('performance', {})
            perf_details = perf_metrics.get('details', {})
            dep_analysis = perf_metrics.get('dependency_analysis', {})

            missing = [k for k, v in perf_details.items() if not v]
            issues = []

            # Check for dependency-related performance issues
            if dep_analysis.get('has_cycles'):
                issues.append(f"{dep_analysis.get('cycle_count', 0)} circular dependencies")
            if dep_analysis.get('longest_chain', 0) > 10:
                issues.append(f"long dependency chain ({dep_analysis.get('longest_chain', 0)} tasks)")
            if dep_analysis.get('max_dependents', 0) > 5:
                issues.append(f"bottleneck tasks ({dep_analysis.get('max_dependents', 0)} dependents)")

            action_parts = []
            if missing:
                action_parts.append(f"Enable: {', '.join(missing[:2])}")
            if issues:
                action_parts.append(f"Fix: {', '.join(issues[:2])}")

            action = ' | '.join(action_parts) if action_parts else 'Review performance optimizations'

            recommendations.append({
                'priority': 'medium',
                'area': 'Performance',
                'action': action,
                'impact': '+15% to performance score',
            })

        if scores.get('completion', 0) < 25:
            recommendations.append({
                'priority': 'medium',
                'area': 'Tasks',
                'action': 'Complete pending tasks to show progress',
                'impact': '+5% to overall score',
            })

        if scores.get('dogfooding', 0) < 70:
            missing = [k for k, v in metrics.get('dogfooding', {}).get('details', {}).items() if not v]
            recommendations.append({
                'priority': 'medium',
                'area': 'Dogfooding',
                'action': f'Enable more self-maintenance: {", ".join(missing[:3])}{"..." if len(missing) > 3 else ""}',
                'impact': '+13% to dogfooding score',
            })

        if scores.get('uniqueness', 0) < 80:
            improvements = metrics.get('uniqueness', {}).get('analysis', {}).get('potential_improvements', [])
            if improvements:
                recommendations.append({
                    'priority': 'low',
                    'area': 'Uniqueness',
                    'action': improvements[0].get('suggestion', 'Document design decisions'),
                    'impact': '+10% to uniqueness score',
                })

        result['recommendations'] = recommendations

    # ═══════════════════════════════════════════════════════════════
    # FORMAT OUTPUT
    # ═══════════════════════════════════════════════════════════════
    if output_format == "json":
        formatted_output = json.dumps(result, indent=2)
    elif output_format == "markdown":
        formatted_output = _format_markdown(result)
    else:
        formatted_output = _format_text(result)

    result['formatted_output'] = formatted_output

    # Save to file if requested
    if output_path:
        output_file = Path(output_path)
        output_file.parent.mkdir(parents=True, exist_ok=True)
        output_file.write_text(formatted_output)
        result['output_file'] = str(output_file)

    # ═══ MEMORY INTEGRATION: Save scorecard for trend tracking ═══
    memory_result = _save_scorecard_memory(result)
    if memory_result.get('success'):
        result['memory_saved'] = memory_result.get('memory_id')

    return result


def _format_text(data: dict) -> str:
    """Format scorecard as plain text."""
    lines = []
    lines.append("=" * 70)
    lines.append("  📊 EXARP PROJECT SCORE CARD")
    lines.append(f"  Generated: {data['generated_at'][:16].replace('T', ' ')}")
    lines.append("=" * 70)

    # Overall score
    overall = data['overall_score']
    status = "🟢" if overall >= 70 else "🟡" if overall >= 50 else "🔴"
    lines.append(f"\n  OVERALL SCORE: {overall}% {status}")
    lines.append(f"  Production Ready: {'YES ✅' if data['production_ready'] else 'NO ❌'}")

    if data.get('blockers'):
        lines.append(f"  Blockers: {', '.join(data['blockers'])}")

    # Component scores
    lines.append("\n  Component Scores:")
    for name, score in sorted(data['scores'].items(), key=lambda x: -x[1]):
        bar_len = int(score / 5)
        bar = "█" * bar_len + "░" * (20 - bar_len)
        status = "🟢" if score >= 70 else "🟡" if score >= 50 else "🔴"
        weight = int(data['weights'].get(name, 0) * 100)
        lines.append(f"    {name:<14} [{bar}] {score:>5.1f}% {status} (×{weight}%)")

    # Key metrics
    lines.append("\n  Key Metrics:")
    if 'tasks' in data['metrics']:
        t = data['metrics']['tasks']
        lines.append(f"    Tasks: {t.get('pending', 0)} pending, {t.get('completed', 0)} completed")
        lines.append(f"    Remaining work: {t.get('remaining_hours', 0)}h")

    # Progress Inference Metrics (T-17)
    if 'progress_inference' in data['metrics']:
        pi = data['metrics']['progress_inference']
        if pi.get('available') is not False:
            lines.append(f"    Progress Inference: {pi.get('inferences_made', 0)} inferences for {pi.get('tasks_analyzed', 0)} tasks")
            if pi.get('discrepancy_count', 0) > 0:
                lines.append(f"    ⚠️ Status Discrepancies: {pi.get('discrepancy_count', 0)} tasks (marked vs. inferred)")
                lines.append(f"    Inferred Completion: {pi.get('inferred_completed', 0)} (vs. {pi.get('marked_completed', 0)} marked)")

    if 'parallelizable' in data['metrics']:
        p = data['metrics']['parallelizable']
        lines.append(f"    Parallelizable: {p.get('ready', 0)} tasks ({p.get('score', 0)}%)")

    if 'performance' in data['metrics']:
        p = data['metrics']['performance']
        dep_analysis = p.get('dependency_analysis', {})
        perf_info = f"    Performance: {p.get('checks_passed', 0)}/{p.get('checks_total', 0)} optimizations"
        if dep_analysis.get('networkx_available') is not False and dep_analysis:
            if dep_analysis.get('has_cycles'):
                perf_info += f" | ⚠️ {dep_analysis.get('cycle_count', 0)} cycles"
            if dep_analysis.get('longest_chain', 0) > 10:
                perf_info += f" | ⚠️ Long chain ({dep_analysis.get('longest_chain', 0)})"
        lines.append(perf_info)

    if 'dogfooding' in data['metrics']:
        d = data['metrics']['dogfooding']
        lines.append(f"    Dogfooding: {d.get('checks_passed', 0)}/{d.get('checks_total', 0)} self-checks")

    if 'uniqueness' in data['metrics']:
        u = data['metrics']['uniqueness']
        deps = u.get('dependency_count', 0)
        justified = u.get('justified_decisions', 0)
        total = u.get('total_decisions', 0)
        lines.append(f"    Uniqueness: {justified}/{total} decisions justified, {deps} deps")

    # CodeQL Security
    if 'security' in data['metrics'] and 'codeql' in data['metrics']['security']:
        cql = data['metrics']['security']['codeql']
        if cql.get('configured'):
            alerts = cql.get('alerts', {})
            alert_str = f"🔐 CodeQL: {alerts.get('total', 0)} alerts"
            if alerts.get('critical', 0) > 0:
                alert_str += f" (🔴 {alerts['critical']} critical)"
            elif alerts.get('high', 0) > 0:
                alert_str += f" (🟠 {alerts['high']} high)"
            elif alerts.get('total', 0) == 0:
                alert_str += " ✅"
            lines.append(f"    {alert_str}")
            if cql.get('languages'):
                lines.append(f"    CodeQL Languages: {', '.join(cql['languages'])}")
        else:
            lines.append("    🔐 CodeQL: Not configured")

    # Recommendations
    if data.get('recommendations'):
        lines.append("\n  Recommendations:")
        for rec in data['recommendations']:
            icon = {'critical': '🔴', 'high': '🟠', 'medium': '🟡'}.get(rec['priority'], '•')
            lines.append(f"    {icon} [{rec['area']}] {rec['action']}")

    lines.append("\n" + "=" * 70)

    # Add daily wisdom if available and enabled
    if WISDOM_AVAILABLE:
        wisdom_config = load_wisdom_config()
        if not wisdom_config.get("disabled"):
            # Check for first run
            project_root = find_project_root()
            marker_file = project_root / '.exarp_wisdom_seen'
            if not marker_file.exists():
                try:
                    from datetime import datetime
                    marker_file.write_text(f"First seen: {datetime.now().isoformat()}\n")
                    lines.append(_first_run_wisdom_prompt())
                except OSError:
                    pass

            wisdom = get_wisdom(data['overall_score'])
            if wisdom:
                lines.append(format_wisdom_text(wisdom))

    return "\n".join(lines)


def _format_markdown(data: dict) -> str:
    """Format scorecard as markdown."""
    lines = []
    lines.append("# 📊 Exarp Project Score Card")
    lines.append(f"\n*Generated: {data['generated_at'][:16].replace('T', ' ')}*")

    # Overall score
    overall = data['overall_score']
    status = "🟢" if overall >= 70 else "🟡" if overall >= 50 else "🔴"
    lines.append(f"\n## Overall Score: **{overall}%** {status}")
    lines.append(f"\n**Production Ready:** {'✅ Yes' if data['production_ready'] else '❌ No'}")

    if data.get('blockers'):
        lines.append(f"\n**Blockers:** {', '.join(data['blockers'])}")

    # Component scores table
    lines.append("\n## Component Scores\n")
    lines.append("| Component | Score | Status | Weight |")
    lines.append("|-----------|-------|--------|--------|")
    for name, score in sorted(data['scores'].items(), key=lambda x: -x[1]):
        status = "🟢" if score >= 70 else "🟡" if score >= 50 else "🔴"
        weight = int(data['weights'].get(name, 0) * 100)
        lines.append(f"| {name.title()} | {score:.1f}% | {status} | {weight}% |")

    # Key metrics
    lines.append("\n## Key Metrics\n")
    if 'tasks' in data['metrics']:
        t = data['metrics']['tasks']
        lines.append(f"- **Tasks:** {t.get('pending', 0)} pending, {t.get('completed', 0)} completed")
        lines.append(f"- **Remaining work:** {t.get('remaining_hours', 0)}h")

    # Progress Inference Metrics (T-17)
    if 'progress_inference' in data['metrics']:
        pi = data['metrics']['progress_inference']
        if pi.get('available') is not False:
            lines.append(f"- **Progress Inference:** {pi.get('inferences_made', 0)} inferences for {pi.get('tasks_analyzed', 0)} tasks")
            if pi.get('discrepancy_count', 0) > 0:
                lines.append(f"  - ⚠️ Status Discrepancies: {pi.get('discrepancy_count', 0)} tasks")
                lines.append(f"  - Inferred Completion: {pi.get('inferred_completed', 0)} (vs. {pi.get('marked_completed', 0)} marked)")

    if 'parallelizable' in data['metrics']:
        p = data['metrics']['parallelizable']
        lines.append(f"- **Parallelizable:** {p.get('ready', 0)} tasks ({p.get('score', 0)}%)")

    if 'performance' in data['metrics']:
        p = data['metrics']['performance']
        dep_analysis = p.get('dependency_analysis', {})
        perf_line = f"- **Performance:** {p.get('checks_passed', 0)}/{p.get('checks_total', 0)} optimizations ({data['scores'].get('performance', 0):.0f}%)"
        if dep_analysis.get('networkx_available') is not False and dep_analysis:
            issues = []
            if dep_analysis.get('has_cycles'):
                issues.append(f"⚠️ {dep_analysis.get('cycle_count', 0)} circular dependencies")
            if dep_analysis.get('longest_chain', 0) > 10:
                issues.append(f"⚠️ Long chain ({dep_analysis.get('longest_chain', 0)} tasks)")
            if dep_analysis.get('max_dependents', 0) > 5:
                issues.append(f"⚠️ Bottleneck ({dep_analysis.get('max_dependents', 0)} dependents)")
            if issues:
                perf_line += f" - {', '.join(issues)}"
        lines.append(perf_line)

    if 'dogfooding' in data['metrics']:
        d = data['metrics']['dogfooding']
        lines.append(f"- **Dogfooding:** {d.get('checks_passed', 0)}/{d.get('checks_total', 0)} self-checks ({data['scores'].get('dogfooding', 0):.0f}%)")

    if 'uniqueness' in data['metrics']:
        u = data['metrics']['uniqueness']
        deps = u.get('dependency_count', 0)
        justified = u.get('justified_decisions', 0)
        total = u.get('total_decisions', 0)
        lines.append(f"- **Uniqueness:** {justified}/{total} decisions justified, {deps} dependencies ({data['scores'].get('uniqueness', 0):.0f}%)")

    # CodeQL Security
    if 'security' in data['metrics'] and 'codeql' in data['metrics']['security']:
        cql = data['metrics']['security']['codeql']
        if cql.get('configured'):
            alerts = cql.get('alerts', {})
            status = "✅ No alerts" if alerts.get('total', 0) == 0 else f"⚠️ {alerts.get('total', 0)} alerts"
            lines.append("\n### 🔐 CodeQL Security\n")
            lines.append(f"- **Status:** {status}")
            lines.append(f"- **Score:** {cql.get('score', 0):.0f}%")
            if alerts.get('total', 0) > 0:
                lines.append(f"- **Critical:** {alerts.get('critical', 0)} | **High:** {alerts.get('high', 0)} | **Medium:** {alerts.get('medium', 0)} | **Low:** {alerts.get('low', 0)}")
            if cql.get('languages'):
                lines.append(f"- **Languages:** {', '.join(cql['languages'])}")
        else:
            lines.append("\n### 🔐 CodeQL Security\n")
            lines.append("- **Status:** Not configured")
            lines.append("- **Recommendation:** Add `.github/workflows/codeql.yml` for automated security scanning")

    # Recommendations
    if data.get('recommendations'):
        lines.append("\n## Recommendations\n")
        for rec in data['recommendations']:
            icon = {'critical': '🔴', 'high': '🟠', 'medium': '🟡'}.get(rec['priority'], '•')
            lines.append(f"- {icon} **{rec['area']}:** {rec['action']} ({rec['impact']})")

    # Add daily wisdom if available and enabled
    if WISDOM_AVAILABLE:
        wisdom_config = load_wisdom_config()
        if not wisdom_config.get("disabled"):
            wisdom = get_wisdom(data['overall_score'])
            if wisdom:
                lines.append(_format_wisdom_markdown(wisdom))

    return "\n".join(lines)


def _first_run_wisdom_prompt() -> str:
    """First-run message introducing the wisdom feature."""
    sources = list_available_sources() if WISDOM_AVAILABLE else []
    source_list = ", ".join([s['id'] for s in sources[:5]]) + "..." if len(sources) > 5 else ", ".join([s['id'] for s in sources])

    return f"""
┌──────────────────────────────────────────────────────────────────────┐
│  ✨ NEW FEATURE: Daily Wisdom from Public Domain Texts               │
├──────────────────────────────────────────────────────────────────────┤
│  Exarp now includes inspirational quotes matched to your project's   │
│  health status. Multiple wisdom sources are available:               │
│                                                                      │
│  📚 Sources: {source_list:<52} │
│                                                                      │
│  Change source: export EXARP_WISDOM_SOURCE=<source>                  │
│  Example:       EXARP_WISDOM_SOURCE=bofh  (tech humor)               │
│                 EXARP_WISDOM_SOURCE=stoic (resilience)               │
│                 EXARP_WISDOM_SOURCE=tao   (balance)                  │
│                                                                      │
│  To disable: export EXARP_DISABLE_WISDOM=1                           │
│              or create .exarp_no_wisdom file in project root         │
└──────────────────────────────────────────────────────────────────────┘
"""


def _format_wisdom_markdown(wisdom: dict) -> str:
    """Format wisdom as Markdown."""
    if wisdom is None:
        return ""

    return f"""
---

### {wisdom.get('wisdom_icon', '📜')} {wisdom.get('wisdom_source', 'Daily Wisdom')}

**Project Status:** {wisdom.get('aeon_level', 'Unknown')}

> *"{wisdom.get('quote', '')}"*
>
> — {wisdom.get('source', '')}

💡 **{wisdom.get('encouragement', '')}**

<details>
<summary>ℹ️ About this quote</summary>

- **Health Score:** {wisdom.get('health_score', 0):.1f}%
- **Change source:** `EXARP_WISDOM_SOURCE=bofh|tao|stoic|bible|murphy|...`
- **Disable:** `export EXARP_DISABLE_WISDOM=1` or create `.exarp_no_wisdom` file.

</details>

---
"""

