#!/usr/bin/env python3
"""
Sprint Automation Script

Systematically processes all background-capable tasks with minimal prompts.
Extracts subtasks, auto-approves safe tasks, runs analysis/testing tools,
and identifies blockers.
"""

import argparse
import json
import logging
import re
import subprocess
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from datetime import datetime
from pathlib import Path
from typing import Optional

# Ensure repo root is on sys.path even when running script directly
script_dir = Path(__file__).resolve().parent
repo_root = script_dir
for candidate in script_dir.parents:
    if (candidate / '.git').exists() or (candidate / '.todo2').exists() or (candidate / 'pyproject.toml').exists():
        repo_root = candidate
        break
else:
    repo_root = script_dir.parent.parent

if str(repo_root) not in sys.path:
    sys.path.insert(0, str(repo_root))

# Import status normalization utilities
try:
    from project_management_automation.utils.todo2_utils import normalize_status, is_pending_status, is_completed_status
except ImportError:
    # Fallback if running as standalone script
    sys.path.insert(0, str(repo_root / 'project_management_automation'))
    from utils.todo2_utils import normalize_status, is_pending_status, is_completed_status

# Import base class (after path setup above)
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase  # noqa: E402

logger = logging.getLogger(__name__)


class SprintAutomation(IntelligentAutomationBase):
    """Sprint automation orchestrator."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Sprint Automation", project_root)

        self.max_iterations = config.get('max_iterations', 10)
        self.auto_approve = config.get('auto_approve', True)
        self.extract_subtasks = config.get('extract_subtasks', True)
        self.run_analysis_tools = config.get('run_analysis_tools', True)
        self.run_testing_tools = config.get('run_testing_tools', True)
        self.test_coverage = config.get('test_coverage', False)  # Default False for speed
        self.cleanup_stale_tasks = config.get('cleanup_stale_tasks', True)  # Default True to prevent stuck tasks
        self.stale_threshold_hours = config.get('stale_threshold_hours', 2.0)  # Default 2 hours
        self.priority_filter = config.get('priority_filter')
        self.tag_filter = config.get('tag_filter')
        self.dry_run = config.get('dry_run', False)
        self.execute_tasks = config.get('execute_tasks', True)  # Default True to actually execute tasks

        # Sprint results
        self.sprint_results = {
            'subtasks_extracted': 0,
            'tasks_auto_approved': 0,
            'tasks_processed': 0,
            'tasks_completed': 0,
            'stale_tasks_cleaned': 0,
            'blockers_identified': [],
            'human_contributions': [],
            'ai_wishlist': [],
            'human_wishlist': [],
            'analysis_results': {},
            'testing_results': {}
        }

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is sprint automation?"""
        return "What is sprint automation? Sprint Automation = Subtask Extraction × Auto-Approval × Tool Orchestration × Background Processing × Blocker Identification"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we sprint through a project?"""
        return "How do we systematically process all background-capable tasks with minimal prompts?"

    def _execute_analysis(self) -> dict:
        """Execute sprint automation workflow."""
        logger.info("Starting sprint automation...")

        iteration = 0
        while iteration < self.max_iterations:
            logger.info(f"Sprint iteration {iteration + 1}/{self.max_iterations}")

            # Step 1: Extract subtasks (only once at start)
            if self.extract_subtasks and iteration == 0:
                self._extract_subtasks()

            # Step 2: Auto-approve safe tasks (only once at start to prevent loop)
            if self.auto_approve and iteration == 0:
                self._auto_approve_tasks()

            # Step 3: Run analysis tools
            if self.run_analysis_tools and iteration == 0:  # Only run once
                self._run_analysis_tools()

            # Step 4: Run testing tools
            if self.run_testing_tools and iteration == 0:  # Only run once
                self._run_testing_tools()

            # Step 5: Clean up stale tasks (before processing new ones)
            if self.cleanup_stale_tasks and iteration == 0:  # Only run once at start
                self._cleanup_stale_tasks()

            # Step 6: Generate wishlists
            if iteration == 0:  # Only generate once
                self._generate_ai_wishlist()
                self._parse_human_wishlist()
                self._identify_human_contributions()

            # Step 7: Process "In Progress" tasks (execute them)
            if self.execute_tasks:
                self._process_in_progress_tasks()

            # Step 8: Process background tasks (mark new tasks as In Progress)
            processed = self._process_background_tasks()
            if processed == 0:
                logger.info("No more background tasks to process")
                break

            iteration += 1

        # Step 9: Identify blockers
        self._identify_blockers()

        # Step 10: Run progress inference at sprint boundaries (T-14)
        self._run_progress_inference()

        # Generate sprint report
        report = self._generate_sprint_report()

        return {
            'status': 'success',
            'iterations': iteration,
            'results': self.sprint_results,
            'report': report
        }

    def _extract_subtasks(self):
        """Extract background-capable subtasks from parent tasks."""
        logger.info("Extracting subtasks...")

        try:
            # Try to use agentic-tools MCP if available
            if hasattr(self, 'mcp_client') and self.mcp_client:
                # Use MCP to list subtasks
                # For now, fallback to file-based approach
                pass

            # Fallback: Load tasks from file
            todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
            if not todo2_path.exists():
                logger.warning("Todo2 state file not found, skipping subtask extraction")
                return

            with open(todo2_path) as f:
                data = json.load(f)

            todos = data.get('todos', [])
            subtasks_extracted = 0

            for task in todos:
                # Check if task has subtasks in description
                long_desc = task.get('long_description', '') or task.get('details', '')
                if 'subtask' in long_desc.lower() or 'sub-task' in long_desc.lower():
                    # Try to extract subtask list
                    subtasks = self._parse_subtasks_from_description(long_desc)
                    for subtask in subtasks:
                        if self._is_background_capable({'name': subtask, 'long_description': ''}):
                            subtasks_extracted += 1
                            logger.info(f"Extracted subtask: {subtask}")

            self.sprint_results['subtasks_extracted'] = subtasks_extracted

        except Exception as e:
            logger.error(f"Error extracting subtasks: {e}", exc_info=True)

    def _parse_subtasks_from_description(self, description: str) -> list[str]:
        """Parse subtasks from task description."""
        subtasks = []

        # Look for bullet points or numbered lists
        lines = description.split('\n')
        for line in lines:
            line = line.strip()
            # Match: - subtask, * subtask, 1. subtask, etc.
            if re.match(r'^[-*•]\s+(.+)', line) or re.match(r'^\d+\.\s+(.+)', line):
                match = re.search(r'[-*•]\s+(.+)|^\d+\.\s+(.+)', line)
                if match:
                    subtask = match.group(1) or match.group(2)
                    if subtask and len(subtask) > 5:  # Filter out very short items
                        subtasks.append(subtask)

        return subtasks

    def _auto_approve_tasks(self):
        """Auto-approve tasks that don't need clarification."""
        logger.info("Auto-approving safe tasks...")

        try:
            # Use batch_approve_tasks tool
            from project_management_automation.tools.batch_task_approval import batch_approve_tasks

            result = batch_approve_tasks(
                status="Review",
                new_status="Todo",
                clarification_none=True,
                dry_run=self.dry_run
            )

            if result.get('success'):
                approved_count = result.get('approved_count', 0)
                self.sprint_results['tasks_auto_approved'] = approved_count
                logger.info(f"Auto-approved {approved_count} tasks")

        except Exception as e:
            logger.error(f"Error auto-approving tasks: {e}", exc_info=True)

    def _run_analysis_tools(self):
        """Run all analysis tools in parallel for better performance."""
        logger.info("Running analysis tools in parallel...")
        start_time = time.time()
        results = {}

        def run_docs_health():
            """Run documentation health check."""
        try:
            from project_management_automation.tools.docs_health import check_documentation_health
            docs_result = json.loads(check_documentation_health(create_tasks=False))
            if docs_result.get('success'):
                    return ('documentation_health', docs_result.get('data', {}))
        except Exception as e:
            logger.warning(f"Documentation health check failed: {e}")
            return None

        def run_task_alignment():
            """Run task alignment analysis."""
        try:
            from project_management_automation.tools.todo2_alignment import analyze_todo2_alignment
            alignment_result = json.loads(analyze_todo2_alignment(create_followup_tasks=True))
            if alignment_result.get('success'):
                    return ('task_alignment', alignment_result.get('data', {}))
        except Exception as e:
            logger.warning(f"Task alignment analysis failed: {e}")
            return None

        def run_duplicate_detection():
            """Run duplicate detection."""
        try:
            from project_management_automation.tools.duplicate_detection import detect_duplicate_tasks
            dup_result = json.loads(detect_duplicate_tasks(auto_fix=True))
            if dup_result.get('success'):
                    return ('duplicate_detection', dup_result.get('data', {}))
        except Exception as e:
            logger.warning(f"Duplicate detection failed: {e}")
            return None

        def run_automation_opportunities():
            """Run automation opportunities analysis."""
        try:
            from project_management_automation.tools.automation_opportunities import find_automation_opportunities
            auto_result = json.loads(find_automation_opportunities(min_value_score=0.8))
            if auto_result.get('success'):
                    return ('automation_opportunities', auto_result.get('data', {}))
        except Exception as e:
            logger.warning(f"Automation opportunities failed: {e}")
            return None

        # Run all analysis tools in parallel
        with ThreadPoolExecutor(max_workers=4) as executor:
            futures = {
                executor.submit(run_docs_health): 'documentation_health',
                executor.submit(run_task_alignment): 'task_alignment',
                executor.submit(run_duplicate_detection): 'duplicate_detection',
                executor.submit(run_automation_opportunities): 'automation_opportunities'
            }

            # Collect results as they complete
            for future in as_completed(futures):
                tool_name = futures[future]
                try:
                    result = future.result()
                    if result:
                        key, data = result
                        results[key] = data
                        logger.info(f"✅ {tool_name} completed")
                except Exception as e:
                    logger.error(f"Error running {tool_name}: {e}", exc_info=True)

        duration = time.time() - start_time
        logger.info(f"Analysis tools completed in {duration:.2f}s (parallel execution)")
        self.sprint_results['analysis_results'] = results

    def _run_testing_tools(self):
        """Run testing tools in parallel for better performance.
        
        Prefers Makefile targets when available (make test, make test-coverage).
        Falls back to Python tools if Makefile is not available.
        """
        logger.info("Running testing tools in parallel...")
        start_time = time.time()
        results = {}
        
        # Check if Makefile exists and has test targets
        makefile_path = self.project_root / 'Makefile'
        use_makefile = makefile_path.exists()

        def run_tests():
            """Run test execution using Makefile or Python tool."""
            try:
                if use_makefile:
                    # Use Makefile target
                    logger.info("Using Makefile target: make test")
                    result = subprocess.run(
                        ['make', 'test'],
                        cwd=self.project_root,
                        capture_output=True,
                        text=True,
                        timeout=300
                    )
                    if result.returncode == 0:
                        return ('test_execution', {
                            'success': True,
                            'method': 'makefile',
                            'output': result.stdout,
                            'message': 'Tests passed via make test'
                        })
                    else:
                        logger.warning(f"make test failed: {result.stderr}")
                        return None
                else:
                    # Fallback to Python tool
                    from project_management_automation.tools.run_tests import run_tests
                    test_result = json.loads(run_tests(coverage=self.test_coverage))
                    if test_result.get('success'):
                        return ('test_execution', test_result.get('data', {}))
            except Exception as e:
                logger.warning(f"Test execution failed: {e}")
                return None

        def analyze_coverage():
            """Analyze test coverage using Makefile or Python tool."""
            try:
                if use_makefile:
                    # Use Makefile target
                    logger.info("Using Makefile target: make test-coverage")
                    result = subprocess.run(
                        ['make', 'test-coverage'],
                        cwd=self.project_root,
                        capture_output=True,
                        text=True,
                        timeout=300
                    )
                    if result.returncode == 0:
                        return ('test_coverage', {
                            'success': True,
                            'method': 'makefile',
                            'output': result.stdout,
                            'message': 'Coverage analyzed via make test-coverage'
                        })
                    else:
                        logger.warning(f"make test-coverage failed: {result.stderr}")
                        return None
                else:
                    # Fallback to Python tool
                    from project_management_automation.tools.test_coverage import analyze_test_coverage
                    coverage_result = json.loads(analyze_test_coverage(min_coverage=80))
                    if coverage_result.get('success'):
                        return ('test_coverage', coverage_result.get('data', {}))
            except Exception as e:
                logger.warning(f"Coverage analysis failed: {e}")
                return None

        # Run tests (always)
        test_future = None
        coverage_future = None
        
        with ThreadPoolExecutor(max_workers=2) as executor:
            # Always run tests
            test_future = executor.submit(run_tests)
            
            # Only run coverage analysis if coverage is enabled
            if self.test_coverage:
                coverage_future = executor.submit(analyze_coverage)
            
            # Collect results as they complete
            futures = {'test_execution': test_future}
            if coverage_future:
                futures['test_coverage'] = coverage_future
            
            for tool_name, future in futures.items():
                try:
                    result = future.result()
                    if result:
                        key, data = result
                        results[key] = data
                        logger.info(f"✅ {tool_name} completed")
                except Exception as e:
                    logger.error(f"Error running {tool_name}: {e}", exc_info=True)

        duration = time.time() - start_time
        coverage_status = "with coverage" if self.test_coverage else "without coverage"
        logger.info(f"Testing tools completed in {duration:.2f}s ({coverage_status})")
        self.sprint_results['testing_results'] = results

    def _generate_ai_wishlist(self):
        """Generate AI wishlist of tasks it wants to work on."""
        logger.info("Generating AI wishlist...")

        wishlist = []

        # Based on analysis results
        if 'automation_opportunities' in self.sprint_results.get('analysis_results', {}):
            opps = self.sprint_results['analysis_results']['automation_opportunities'].get('opportunities', [])
            for opp in opps[:10]:  # Top 10
                wishlist.append({
                    'type': 'automation_opportunity',
                    'name': opp.get('name', ''),
                    'description': opp.get('description', ''),
                    'value_score': opp.get('score', 0),
                    'priority': 'high' if opp.get('score', 0) >= 8 else 'medium'
                })

        # Based on test coverage gaps
        if 'test_coverage' in self.sprint_results.get('testing_results', {}):
            coverage_data = self.sprint_results['testing_results']['test_coverage']
            gaps = coverage_data.get('gaps', [])
            for gap in gaps[:5]:  # Top 5 gaps
                wishlist.append({
                    'type': 'test_coverage_gap',
                    'filename': gap.get('filename', ''),
                    'coverage': gap.get('coverage', 0),
                    'missing_lines': gap.get('missing', 0),
                    'priority': 'high' if gap.get('coverage', 0) < 50 else 'medium'
                })

        # Based on documentation health
        if 'documentation_health' in self.sprint_results.get('analysis_results', {}):
            docs_data = self.sprint_results['analysis_results']['documentation_health']
            if docs_data.get('health_score', 100) < 80:
                wishlist.append({
                    'type': 'documentation_improvement',
                    'health_score': docs_data.get('health_score', 0),
                    'broken_links': docs_data.get('link_validation', {}).get('broken_internal', 0),
                    'priority': 'high' if docs_data.get('health_score', 100) < 70 else 'medium'
                })

        self.sprint_results['ai_wishlist'] = wishlist
        logger.info(f"Generated {len(wishlist)} AI wishlist items")

    def _parse_human_wishlist(self):
        """Parse human wishlist from external sources."""
        logger.info("Parsing human wishlist...")

        wishlist = []

        # 1. Parse from WISHLIST.md or TODO.md files
        for filename in ['WISHLIST.md', 'TODO.md', 'WISH_LIST.md', 'docs/WISHLIST.md']:
            wishlist_path = self.project_root / filename
            if wishlist_path.exists():
                items = self._parse_wishlist_file(wishlist_path)
                wishlist.extend(items)
                logger.info(f"Found {len(items)} items in {filename}")

        # 2. Parse from git commit messages (recent commits with "wish" or "want")
        git_wishlist = self._parse_git_wishlist()
        wishlist.extend(git_wishlist)

        # 3. Parse from code comments (TODO/FIXME with wish indicators)
        comment_wishlist = self._parse_comment_wishlist()
        wishlist.extend(comment_wishlist)

        self.sprint_results['human_wishlist'] = wishlist
        logger.info(f"Parsed {len(wishlist)} human wishlist items")

    def _parse_wishlist_file(self, file_path: Path) -> list[dict]:
        """Parse wishlist from markdown file."""
        items = []

        try:
            content = file_path.read_text()
            lines = content.split('\n')

            for line in lines:
                line = line.strip()
                # Match: - wish item, * wish item, etc.
                if re.match(r'^[-*•]\s+(.+)', line):
                    match = re.search(r'[-*•]\s+(.+)', line)
                    if match:
                        item_text = match.group(1)
                        # Extract priority if present
                        priority = 'medium'
                        if 'high' in item_text.lower() or 'urgent' in item_text.lower():
                            priority = 'high'
                        elif 'low' in item_text.lower() or 'nice' in item_text.lower():
                            priority = 'low'

                        items.append({
                            'type': 'human_wishlist',
                            'source': file_path.name,
                            'text': item_text,
                            'priority': priority
                        })
        except Exception as e:
            logger.warning(f"Error parsing wishlist file {file_path}: {e}")

        return items

    def _parse_git_wishlist(self) -> list[dict]:
        """Parse wishlist items from git commit messages."""
        items = []

        try:
            # Get last 50 commit messages
            result = subprocess.run(
                ['git', 'log', '--pretty=format:%s', '-50'],
                cwd=str(self.project_root),
                capture_output=True,
                text=True,
                timeout=10
            )

            if result.returncode == 0:
                commits = result.stdout.split('\n')
                for commit in commits:
                    commit_lower = commit.lower()
                    if any(word in commit_lower for word in ['wish', 'want', 'would like', 'hope to', 'plan to add']):
                        items.append({
                            'type': 'git_commit',
                            'source': 'git_commits',
                            'text': commit,
                            'priority': 'medium'
                        })
        except Exception as e:
            logger.warning(f"Error parsing git wishlist: {e}")

        return items

    def _parse_comment_wishlist(self) -> list[dict]:
        """Parse wishlist items from code comments."""
        items = []

        try:
            # Search for TODO/FIXME comments with wish indicators
            for ext in ['*.py', '*.js', '*.ts', '*.cpp', '*.h', '*.md']:
                for file_path in self.project_root.rglob(ext):
                    if file_path.is_file():
                        try:
                            content = file_path.read_text(errors='ignore')
                            lines = content.split('\n')
                            for i, line in enumerate(lines, 1):
                                line_lower = line.lower()
                                if ('todo' in line_lower or 'fixme' in line_lower) and \
                                   any(word in line_lower for word in ['wish', 'want', 'would like', 'hope']):
                                    items.append({
                                        'type': 'code_comment',
                                        'source': str(file_path.relative_to(self.project_root)),
                                        'line': i,
                                        'text': line.strip(),
                                        'priority': 'medium'
                                    })
                        except Exception:
                            continue
        except Exception as e:
            logger.warning(f"Error parsing comment wishlist: {e}")

        return items[:20]  # Limit to 20 items

    def _cleanup_stale_tasks(self):
        """Clean up stale In Progress tasks by moving them back to Todo."""
        logger.info(f"Cleaning up stale tasks (threshold: {self.stale_threshold_hours} hours)...")

        try:
            from project_management_automation.tools.stale_task_cleanup import cleanup_stale_tasks

            result_json = cleanup_stale_tasks(
                stale_threshold_hours=self.stale_threshold_hours,
                dry_run=self.dry_run
            )

            if result_json:
                result = json.loads(result_json)
                if result.get('success') and result.get('data'):
                    data = result['data']
                    tasks_moved = data.get('tasks_moved', 0)
                    self.sprint_results['stale_tasks_cleaned'] = tasks_moved
                    logger.info(f"Cleaned up {tasks_moved} stale task(s)")
                else:
                    logger.warning(f"Stale task cleanup returned no data: {result}")
        except Exception as e:
            logger.error(f"Error cleaning up stale tasks: {e}", exc_info=True)

    def _identify_human_contributions(self):
        """Identify tasks better handled by humans."""
        logger.info("Identifying human contribution opportunities...")

        contributions = []

        # Load tasks
        todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
        if todo2_path.exists():
            with open(todo2_path) as f:
                data = json.load(f)

            todos = data.get('todos', [])

            for task in todos:
                name = task.get('name', '').lower()
                long_desc = task.get('long_description', '').lower() or task.get('details', '').lower()
                status = task.get('status', '')

                # Human contribution indicators
                is_design = 'design' in name and any(x in name for x in ['framework', 'system', 'strategy', 'architecture', 'ui', 'ux'])
                is_creative = any(x in name for x in ['creative', 'art', 'design', 'visual', 'branding'])
                is_user_preference = 'user preference' in long_desc or 'user choice' in long_desc or 'user input' in long_desc
                is_strategy = 'strategy' in name or 'strategy' in long_desc
                is_decision = any(x in name for x in ['decide', 'choose', 'select', 'recommend']) and 'implement' not in name
                needs_clarification = 'clarification required' in long_desc

                if (is_design or is_creative or is_user_preference or is_strategy or is_decision) and \
                   (status == 'Review' or needs_clarification):
                    contributions.append({
                        'task_id': task.get('id', ''),
                        'name': task.get('name', ''),
                        'reason': self._get_human_contribution_reason(task),
                        'priority': task.get('priority', 'medium'),
                        'status': status
                    })

        self.sprint_results['human_contributions'] = contributions
        logger.info(f"Identified {len(contributions)} human contribution opportunities")

    def _get_human_contribution_reason(self, task: dict) -> str:
        """Get reason why task needs human contribution."""
        name = task.get('name', '').lower()
        long_desc = task.get('long_description', '').lower() or task.get('details', '').lower()

        if 'design' in name and 'framework' in name:
            return "Design decision required (framework selection)"
        elif 'design' in name and 'ui' in name:
            return "UI/UX design requires human creativity"
        elif 'user preference' in long_desc:
            return "User preference needed"
        elif 'strategy' in name:
            return "Strategic decision required"
        elif 'clarification required' in long_desc:
            return "Clarification needed from human"
        else:
            return "Human judgment required"

    def _process_background_tasks(self) -> int:
        """Process background-capable tasks."""
        logger.info("Processing background tasks...")

        # Load tasks
        todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
        if not todo2_path.exists():
            return 0

        with open(todo2_path) as f:
            data = json.load(f)

        todos = data.get('todos', [])

        # Filter background-capable tasks
        background_tasks = [t for t in todos if self._is_background_capable(t)]
        logger.info(f"Found {len(background_tasks)} background-capable tasks out of {len(todos)} total tasks")

        # Apply filters
        if self.priority_filter:
            before_filter = len(background_tasks)
            background_tasks = [t for t in background_tasks if t.get('priority') == self.priority_filter]
            logger.info(f"After priority filter ({self.priority_filter}): {len(background_tasks)} tasks (was {before_filter})")

        if self.tag_filter:
            before_filter = len(background_tasks)
            background_tasks = [t for t in background_tasks if
                              any(tag in t.get('tags', []) for tag in self.tag_filter)]
            logger.info(f"After tag filter ({self.tag_filter}): {len(background_tasks)} tasks (was {before_filter})")

        # Sort by priority
        priority_order = {'high': 3, 'medium': 2, 'low': 1}
        background_tasks.sort(key=lambda t: priority_order.get(t.get('priority', 'medium'), 2), reverse=True)

        # Process up to 10 tasks
        processed = 0
        tasks_to_update = []
        
        logger.info(f"Evaluating top {min(10, len(background_tasks))} tasks for processing")
        for task in background_tasks[:10]:
            # Use normalized status check (already imported)
            task_status = task.get('status', '')
            
            # Skip completed tasks (Done, Cancelled, etc.) - double-check even though _is_background_capable filters them
            if is_completed_status(task_status):
                task_name = task.get('name') or task.get('content', 'Unknown')
                logger.debug(f"Skipping completed task: {task_name} (status: {task_status})")
                continue
                
            if is_pending_status(task_status):
                task_name = task.get('name') or task.get('content', 'Unknown')
                task_id = task.get('id')
                
                # Mark as in progress first
                if not self.dry_run:
                    tasks_to_update.append({
                        'id': task_id,
                        'status': 'In Progress'
                    })
                
                # Note: Task execution happens via progress inference at the end of the sprint
                # (See _run_progress_inference() method which analyzes all tasks)
                if self.execute_tasks:
                    logger.info(f"Processing task: {task_name} (ID: {task_id}) - will be checked for completion via progress inference")
                else:
                    logger.info(f"Processing task: {task_name} (ID: {task_id}, status: {task_status})")
                
                processed += 1
            else:
                task_name = task.get('name') or task.get('content', 'Unknown')
                logger.debug(f"Skipping task (not pending): {task_name} (status: {task_status})")

        # Actually save the status changes to prevent infinite loop
        if tasks_to_update and not self.dry_run:
            logger.info(f"Attempting to update {len(tasks_to_update)} task(s) to 'In Progress'")
            try:
                from project_management_automation.utils.todo2_mcp_client import update_todos_mcp
                task_ids = [t['id'] for t in tasks_to_update]
                logger.debug(f"Task IDs to update: {task_ids}")
                success = update_todos_mcp(
                    updates=tasks_to_update,
                    project_root=self.project_root
                )
                if success:
                    logger.info(f"Successfully updated {len(tasks_to_update)} task(s) to 'In Progress': {task_ids}")
                else:
                    logger.warning(f"update_todos_mcp returned False for {len(tasks_to_update)} task(s): {task_ids}")
            except Exception as e:
                logger.error(f"Error updating task statuses: {e}", exc_info=True)
        elif tasks_to_update and self.dry_run:
            logger.info(f"DRY RUN: Would update {len(tasks_to_update)} task(s) to 'In Progress'")
        elif not tasks_to_update:
            logger.info("No tasks to update (all background-capable tasks are already processed or filtered out)")

        self.sprint_results['tasks_processed'] += processed
        return processed

    def _process_in_progress_tasks(self) -> int:
        """Process tasks that are already 'In Progress' - attempt execution and check for completion."""
        logger.info("Processing 'In Progress' tasks...")

        # Load tasks
        todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
        if not todo2_path.exists():
            return 0

        with open(todo2_path) as f:
            data = json.load(f)

        todos = data.get('todos', [])

        # Filter to "In Progress" tasks only
        in_progress_tasks = [
            t for t in todos 
            if normalize_status(t.get('status', '')) == 'in_progress'
            and not t.get('id', '').startswith('AUTO-')  # Skip automation tasks
        ]
        
        logger.info(f"Found {len(in_progress_tasks)} 'In Progress' tasks to process")

        if not in_progress_tasks:
            return 0

        # Attempt to execute tasks (if enabled)
        executed_count = 0
        if self.execute_tasks and not self.dry_run:
            for task in in_progress_tasks[:5]:  # Limit to 5 tasks per iteration
                task_id = task.get('id')
                task_name = task.get('name', '')
                task.get('long_description', '') or task.get('details', '')
                
                logger.info(f"Attempting to execute task: {task_name} (ID: {task_id})")
                execution_result = self._execute_task(task)
                
                if execution_result.get('executed'):
                    executed_count += 1
                    logger.info(f"Task execution attempted: {task_name}")
                else:
                    logger.debug(f"Task execution skipped or failed: {task_name} - {execution_result.get('reason', 'unknown')}")

        # Run progress inference to detect completed tasks
        try:
            from project_management_automation.tools.auto_update_task_status import auto_update_task_status
            from project_management_automation.utils.todo2_utils import get_repo_project_id

            project_id = get_repo_project_id(self.project_root)
            
            # Run progress inference to detect completed tasks
            logger.info(f"Checking {len(in_progress_tasks)} 'In Progress' tasks for completion...")
            result_json = auto_update_task_status(
                project_id=project_id,
                scan_depth=3,
                auto_update_tasks=True,  # Actually update task status if completed
                confidence_threshold=0.7,  # Standard threshold
                dry_run=self.dry_run,
                output_path=None
            )

            if result_json:
                result = json.loads(result_json)
                if result.get('success') and result.get('data'):
                    data = result['data']
                    data.get('tasks_updated', 0)
                    inferred_results = data.get('inferred_results', [])
                    
                    # Count how many of our "In Progress" tasks were completed
                    in_progress_task_ids = {t.get('id') for t in in_progress_tasks}
                    completed_count = sum(
                        1 for r in inferred_results 
                        if r.get('task_id') in in_progress_task_ids 
                        and r.get('inferred_status') == 'done'
                    )
                    
                    if completed_count > 0:
                        logger.info(f"Detected {completed_count} completed task(s) from 'In Progress' tasks")
                        self.sprint_results['tasks_completed'] += completed_count
                    else:
                        logger.info(f"Checked {len(in_progress_tasks)} 'In Progress' tasks - none completed yet")
                    
                    return completed_count
        except Exception as e:
            logger.error(f"Error processing 'In Progress' tasks: {e}", exc_info=True)
            return 0

        return 0

    def _execute_task(self, task: dict) -> dict:
        """
        Execute a task by analyzing its description and attempting implementation.
        
        This method attempts to execute tasks by:
        1. Parsing the task description to identify what needs to be done
        2. Generating execution instructions or code based on the task
        3. Logging the execution attempt (actual implementation happens via progress inference)
        
        Note: Full automated code generation/execution requires AI agent integration.
        For now, this prepares tasks for execution and relies on progress inference
        to detect when tasks are actually completed.
        """
        task_id = task.get('id')
        task_name = task.get('name', '')
        task_desc = task.get('long_description', '') or task.get('details', '')
        
        # Skip tasks that need clarification or user input
        if 'clarification required' in task_desc.lower():
            return {'executed': False, 'reason': 'Needs clarification'}
        
        if 'user input' in task_desc.lower() or 'user interaction' in task_desc.lower():
            return {'executed': False, 'reason': 'Requires user input'}
        
        # For now, task execution relies on:
        # 1. Tasks being marked as "In Progress" (already done)
        # 2. AI agent (Cursor) picking them up and implementing them
        # 3. Progress inference detecting completion from codebase changes
        
        # Log execution attempt
        logger.debug(f"Task {task_id} ({task_name}) - execution prepared, waiting for AI agent implementation")
        
        return {
            'executed': True,
            'task_id': task_id,
            'method': 'progress_inference',  # Execution is detected, not directly executed
            'note': 'Task marked for execution - completion will be detected via progress inference'
        }

    def _is_background_capable(self, task: dict) -> bool:
        """Check if task is background-capable."""
        # Support both 'name' and 'content' fields (different todo formats)
        name = (task.get('name', '') or task.get('content', '')).lower()
        long_desc = (task.get('long_description', '') or task.get('details', '')).lower()
        status = task.get('status', '')
        task_id = task.get('id', '')

        # Skip automation tasks (they shouldn't process themselves)
        if task_id.startswith('AUTO-'):
            return False

        # Skip completed tasks (Done, Cancelled, etc.)
        if is_completed_status(status):
            return False

        # Skip if not in actionable status (normalized)
        if not is_pending_status(status):
            return False

        # Interactive indicators (exclude)
        needs_clarification = 'clarification required' in long_desc
        needs_user_input = 'user input' in long_desc or 'user interaction' in long_desc
        is_design = 'design' in name and any(x in name for x in ['framework', 'system', 'strategy'])
        is_decision = any(x in name for x in ['decide', 'choose', 'select']) and 'implement' not in name

        is_interactive = needs_clarification or needs_user_input or (is_design and 'implement' not in name) or is_decision

        # Background indicators (include)
        is_research = 'research' in name
        is_implementation = any(x in name for x in ['implement', 'create', 'add', 'update', 'fix', 'refactor'])
        is_testing = 'test' in name or 'testing' in name
        is_documentation = 'document' in name or 'documentation' in name
        is_configuration = 'config' in name or 'configure' in name

        is_background = (is_research or is_implementation or is_testing or is_documentation or is_configuration) and not is_interactive

        return is_background

    def _run_progress_inference(self):
        """Run progress inference at sprint boundaries to detect completed tasks."""
        logger.info("Running progress inference to detect completed tasks...")
        
        try:
            from project_management_automation.tools.auto_update_task_status import auto_update_task_status
            from project_management_automation.utils.todo2_utils import get_repo_project_id
            import json
            
            # Get project_id from project_root
            project_id = get_repo_project_id(self.project_root)
            
            # Run progress inference and auto-update if execute_tasks is enabled
            result_json = auto_update_task_status(
                project_id=project_id,
                confidence_threshold=0.8 if self.execute_tasks else 0.7,  # Higher threshold when executing
                auto_update_tasks=self.execute_tasks and not self.dry_run,  # Auto-update if executing
                dry_run=self.dry_run,
                output_path=None
            )
            
            if result_json:
                result = json.loads(result_json)
                if result.get('success') and result.get('data'):
                    data = result['data']
                    inferences_made = data.get('inferences_made', 0)
                    tasks_analyzed = data.get('total_tasks_analyzed', 0)
                    tasks_updated = data.get('tasks_updated', 0)
                    
                    self.sprint_results['progress_inference'] = {
                        'tasks_analyzed': tasks_analyzed,
                        'inferences_made': inferences_made,
                        'tasks_updated': tasks_updated,
                        'inferred_results': data.get('inferred_results', [])
                    }
                    
                    if tasks_updated > 0:
                        logger.info(f"Progress inference: Updated {tasks_updated} task(s) based on codebase analysis")
                    else:
                        logger.info(f"Progress inference: Analyzed {tasks_analyzed} tasks, {inferences_made} completion(s) detected")
        except Exception as e:
            logger.debug(f"Progress inference not available: {e}")

    def _identify_blockers(self):
        """Identify tasks that are blocked."""
        logger.info("Identifying blockers...")

        blockers = []

        # Load tasks
        todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
        if todo2_path.exists():
            with open(todo2_path) as f:
                data = json.load(f)

            todos = data.get('todos', [])

            for task in todos:
                if task.get('status') == 'Review':
                    task.get('long_description', '') or task.get('details', '')
                    blockers.append({
                        'task_id': task.get('id', ''),
                        'name': task.get('name', ''),
                        'reason': self._get_blocker_reason(task),
                        'priority': task.get('priority', 'medium')
                    })

        self.sprint_results['blockers_identified'] = blockers
        logger.info(f"Identified {len(blockers)} blockers")

    def _get_blocker_reason(self, task: dict) -> str:
        """Get reason why task is blocked."""
        long_desc = task.get('long_description', '').lower() or task.get('details', '').lower()
        name = task.get('name', '').lower()

        if 'clarification required' in long_desc:
            return "Clarification needed"
        elif 'design' in name:
            return "Design decision required"
        elif 'user input' in long_desc:
            return "User input required"
        elif 'strategy' in name:
            return "Strategic decision needed"
        else:
            return "Review status - awaiting approval"

    def _generate_sprint_report(self) -> str:
        """Generate comprehensive sprint report."""
        report_lines = [
            "# Sprint Automation Report",
            "",
            f"*Generated: {datetime.now().isoformat()}*",
            "",
            "## Summary",
            "",
            f"- **Subtasks Extracted:** {self.sprint_results['subtasks_extracted']}",
            f"- **Tasks Auto-Approved:** {self.sprint_results['tasks_auto_approved']}",
            f"- **Stale Tasks Cleaned:** {self.sprint_results['stale_tasks_cleaned']}",
            f"- **Tasks Processed:** {self.sprint_results['tasks_processed']}",
            f"- **Tasks Completed:** {self.sprint_results['tasks_completed']}",
            f"- **Blockers Identified:** {len(self.sprint_results['blockers_identified'])}",
            "",
        ]

        # Analysis Results
        if self.sprint_results.get('analysis_results'):
            report_lines.extend([
                "## Analysis Results",
                ""
            ])

            if 'documentation_health' in self.sprint_results['analysis_results']:
                docs = self.sprint_results['analysis_results']['documentation_health']
                report_lines.append(f"- **Documentation Health:** {docs.get('health_score', 0)}/100")

            if 'task_alignment' in self.sprint_results['analysis_results']:
                alignment = self.sprint_results['analysis_results']['task_alignment']
                misaligned = alignment.get('misaligned_count', 0)
                report_lines.append(f"- **Task Alignment:** {misaligned} misaligned tasks")

            if 'duplicate_detection' in self.sprint_results['analysis_results']:
                dupes = self.sprint_results['analysis_results']['duplicate_detection']
                dup_count = dupes.get('total_duplicates_found', 0)
                report_lines.append(f"- **Duplicates Found:** {dup_count} (auto-fixed)")

            report_lines.append("")
        
        # Progress Inference Results (T-14)
        if self.sprint_results.get('progress_inference'):
            inference = self.sprint_results['progress_inference']
            report_lines.extend([
                "## Progress Inference",
                "",
                f"- **Tasks Analyzed**: {inference.get('tasks_analyzed', 0)}",
                f"- **Inferences Made**: {inference.get('inferences_made', 0)}",
                ""
            ])
            
            inferred_results = inference.get('inferred_results', [])
            if inferred_results:
                report_lines.append("Tasks with inferred status changes:")
                report_lines.append("")
                for result in inferred_results[:5]:  # Show top 5
                    task_name = result.get('task_name', 'N/A')
                    current = result.get('current_status', 'N/A')
                    inferred = result.get('inferred_status', 'N/A')
                    confidence = result.get('confidence', 0.0)
                    report_lines.append(f"- **{task_name}**: {current} → {inferred} (Confidence: {confidence:.0%})")
                report_lines.append("")

        # Testing Results
        if self.sprint_results.get('testing_results'):
            report_lines.extend([
                "## Testing Results",
                ""
            ])

            if 'test_coverage' in self.sprint_results['testing_results']:
                coverage = self.sprint_results['testing_results']['test_coverage']
                total_coverage = coverage.get('total_coverage', 0)
                report_lines.append(f"- **Test Coverage:** {total_coverage:.1f}%")
                if not coverage.get('meets_threshold', True):
                    report_lines.append("  - ⚠️ Below 80% threshold")

            report_lines.append("")

        # AI Wishlist
        if self.sprint_results.get('ai_wishlist'):
            report_lines.extend([
                "## AI Wishlist",
                "",
                "Tasks the AI wants to work on:",
                ""
            ])

            for item in self.sprint_results['ai_wishlist'][:10]:
                report_lines.append(f"- **{item.get('type', 'unknown')}**: {item.get('name', item.get('text', 'Unknown'))} (Priority: {item.get('priority', 'medium')})")

            report_lines.append("")

        # Human Wishlist
        if self.sprint_results.get('human_wishlist'):
            report_lines.extend([
                "## Human Wishlist",
                "",
                "Items from external sources:",
                ""
            ])

            for item in self.sprint_results['human_wishlist'][:10]:
                report_lines.append(f"- **{item.get('source', 'unknown')}**: {item.get('text', 'Unknown')[:100]}...")

            report_lines.append("")

        # Human Contributions
        if self.sprint_results.get('human_contributions'):
            report_lines.extend([
                "## Human Contribution Opportunities",
                "",
                "Tasks better handled by humans:",
                ""
            ])

            for contrib in self.sprint_results['human_contributions'][:10]:
                report_lines.append(f"- **{contrib.get('task_id', 'Unknown')}**: {contrib.get('name', 'Unknown')}")
                report_lines.append(f"  - Reason: {contrib.get('reason', 'Unknown')}")

            report_lines.append("")

        # Blockers
        if self.sprint_results.get('blockers_identified'):
            report_lines.extend([
                "## Blockers",
                "",
                "Tasks requiring human input:",
                ""
            ])

            for blocker in self.sprint_results['blockers_identified'][:10]:
                report_lines.append(f"- **{blocker.get('task_id', 'Unknown')}**: {blocker.get('name', 'Unknown')}")
                report_lines.append(f"  - Reason: {blocker.get('reason', 'Unknown')}")

            report_lines.append("")

        # Next Actions
        report_lines.extend([
            "## Next Actions",
            "",
            f"- Review {len(self.sprint_results['blockers_identified'])} blockers",
            f"- Consider {len(self.sprint_results['human_contributions'])} human contribution opportunities",
            f"- Process {len(self.sprint_results.get('ai_wishlist', []))} AI wishlist items",
            ""
        ])

        return '\n'.join(report_lines)

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from sprint results."""
        insights = []

        # Insights from results
        results = analysis_results.get('results', {})

        if results.get('subtasks_extracted', 0) > 0:
            insights.append(f"Extracted {results['subtasks_extracted']} background-capable subtasks from parent tasks")

        if results.get('tasks_auto_approved', 0) > 0:
            insights.append(f"Auto-approved {results['tasks_auto_approved']} tasks that don't need clarification")

        if results.get('stale_tasks_cleaned', 0) > 0:
            insights.append(f"Cleaned up {results['stale_tasks_cleaned']} stale In Progress tasks (moved back to Todo)")

        if results.get('tasks_processed', 0) > 0:
            insights.append(f"Processed {results['tasks_processed']} background tasks")

        blockers = results.get('blockers_identified', [])
        if blockers:
            insights.append(f"Identified {len(blockers)} blockers requiring human input")

        human_contribs = results.get('human_contributions', [])
        if human_contribs:
            insights.append(f"Found {len(human_contribs)} tasks better handled by humans")

        ai_wishlist = results.get('ai_wishlist', [])
        if ai_wishlist:
            insights.append(f"Generated {len(ai_wishlist)} AI wishlist items")

        human_wishlist = results.get('human_wishlist', [])
        if human_wishlist:
            insights.append(f"Parsed {len(human_wishlist)} human wishlist items from external sources")

        return '\n'.join(insights) if insights else "Sprint automation completed successfully"

    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate sprint report."""
        # Use the existing _generate_sprint_report method
        return self._generate_sprint_report()

    def _format_findings(self, analysis_results: dict) -> str:
        """Format sprint results."""
        return json.dumps(analysis_results, indent=2)


def load_config(config_path: Optional[Path] = None) -> dict:
    """Load configuration."""
    default_config = {
        'max_iterations': 10,
        'auto_approve': True,
        'extract_subtasks': True,
        'run_analysis_tools': True,
        'run_testing_tools': True,
        'test_coverage': False,  # Default False for speed (coverage adds 10-15+ minutes)
        'cleanup_stale_tasks': True,  # Default True to prevent stuck tasks
        'stale_threshold_hours': 2.0,  # Default 2 hours before task is considered stale
        'execute_tasks': True,  # Default True to actually execute tasks (not just mark as In Progress)
        'priority_filter': None,
        'tag_filter': None,
        'dry_run': False
    }

    if config_path and config_path.exists():
        try:
            with open(config_path) as f:
                user_config = json.load(f)
                default_config.update(user_config)
        except json.JSONDecodeError:
            pass

    return default_config


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Sprint Automation')
    parser.add_argument('--max-iterations', type=int, default=10)
    parser.add_argument('--auto-approve', action='store_true', default=True)
    parser.add_argument('--no-auto-approve', dest='auto_approve', action='store_false')
    parser.add_argument('--extract-subtasks', action='store_true', default=True)
    parser.add_argument('--no-extract-subtasks', dest='extract_subtasks', action='store_false')
    parser.add_argument('--run-analysis', action='store_true', default=True)
    parser.add_argument('--no-analysis', dest='run_analysis_tools', action='store_false')
    parser.add_argument('--run-testing', dest='run_testing_tools', action='store_true', default=True)
    parser.add_argument('--no-testing', dest='run_testing_tools', action='store_false')
    parser.add_argument('--coverage', dest='test_coverage', action='store_true', default=False, help='Enable test coverage analysis (adds 10-15+ minutes, default: disabled for speed)')
    parser.add_argument('--cleanup-stale', dest='cleanup_stale_tasks', action='store_true', default=True, help='Clean up stale In Progress tasks (default: enabled)')
    parser.add_argument('--no-cleanup-stale', dest='cleanup_stale_tasks', action='store_false', help='Disable stale task cleanup')
    parser.add_argument('--stale-threshold', type=float, dest='stale_threshold_hours', default=2.0, help='Hours of inactivity before task is considered stale (default: 2.0)')
    parser.add_argument('--execute-tasks', dest='execute_tasks', action='store_true', default=True, help='Actually execute tasks using progress inference (default: enabled)')
    parser.add_argument('--no-execute-tasks', dest='execute_tasks', action='store_false', help='Only mark tasks as In Progress, do not execute them')
    parser.add_argument('--priority', type=str, choices=['high', 'medium', 'low'])
    parser.add_argument('--tag', type=str, action='append')
    parser.add_argument('--dry-run', action='store_true')
    parser.add_argument('--output', type=str)
    parser.add_argument('--config', type=Path)
    args = parser.parse_args()

    config = load_config(args.config)
    config.update({
        'max_iterations': args.max_iterations,
        'auto_approve': args.auto_approve,
        'extract_subtasks': args.extract_subtasks,
        'run_analysis_tools': args.run_analysis,
        'run_testing_tools': args.run_testing_tools,
        'test_coverage': args.test_coverage,
        'cleanup_stale_tasks': args.cleanup_stale_tasks,
        'stale_threshold_hours': args.stale_threshold_hours,
        'execute_tasks': args.execute_tasks,
        'priority_filter': args.priority,
        'tag_filter': args.tag,
        'dry_run': args.dry_run
    })

    automation = SprintAutomation(config)

    try:
        results = automation.run()

        # Write report
        if args.output:
            output_path = Path(args.output)
        else:
            output_path = automation.project_root / 'docs' / 'SPRINT_AUTOMATION_REPORT.md'

        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, 'w') as f:
            f.write(results.get('report', ''))

        print("✅ Sprint automation complete")
        print(f"📄 Report written to: {output_path}")
        print(json.dumps(results.get('results', {}), indent=2))

        sys.exit(0)
    except Exception as e:
        logger.error(f"Error running sprint automation: {e}", exc_info=True)
        sys.exit(1)


if __name__ == '__main__':
    main()

