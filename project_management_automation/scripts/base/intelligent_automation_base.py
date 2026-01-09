#!/usr/bin/env python3
"""
Intelligent Automation Base Class

This base class integrates:
- Tractatus Thinking: Understand WHAT to analyze (structure)
- Sequential Thinking: Plan HOW to analyze (workflow)
- Todo2: Track execution and create follow-up tasks
- NetworkX: Understand relationships and dependencies

All automation scripts should inherit from this class.
"""

import json
import logging
from abc import ABC, abstractmethod
from datetime import datetime
from pathlib import Path
from typing import Optional

from project_management_automation.utils.logging_config import configure_logging
from project_management_automation.utils.todo2_utils import (
    annotate_task_project,
    get_repo_project_id,
)

# Configure logging for this module
logger = configure_logging(__name__, level=logging.INFO)

# Import MCP client (relative import for package)
try:
    from .mcp_client import get_mcp_client
    MCP_AVAILABLE = True
except ImportError:
    MCP_AVAILABLE = False
    logger.debug("MCP client not available, using fallback implementations")

# Project root detection - will be set by subclasses or tools
project_root = None


class IntelligentAutomationBase(ABC):
    """Base class for intelligent automation scripts."""

    def __init__(self, config: dict, automation_name: str, project_root: Optional[Path] = None):
        self.config = config
        self.automation_name = automation_name
        # Use provided project_root or try to detect it
        if project_root:
            self.project_root = Path(project_root)
        else:
            # Try to find project root by looking for .git, .todo2, or CMakeLists.txt
            current = Path(__file__).parent.parent.parent.parent
            while current != current.parent:
                if (current / '.git').exists() or (current / '.todo2').exists() or (current / 'CMakeLists.txt').exists():
                    self.project_root = current
                    break
                current = current.parent
            else:
                # Fallback to current working directory
                self.project_root = Path.cwd()

        # Initialize components
        self.tractatus_session = None
        self.sequential_session = None
        self.todo2_task = None
        self.networkx_graph = None

        self.project_id = get_repo_project_id(self.project_root)

        # Results storage
        self.results = {
            'automation_name': automation_name,
            'timestamp': datetime.now().isoformat(),
            'workflow_steps': [],
            'findings': [],
            'recommendations': [],
            'followup_tasks': []
        }

    def run(self) -> dict:
        """Main execution method - follows intelligent automation pattern."""
        logger.info(f"Starting intelligent automation: {self.automation_name}")

        try:
            # Step 1: Use Tractatus Thinking to understand structure
            self._tractatus_analysis()

            # Step 2: Use Sequential Thinking to plan workflow
            self._sequential_planning()

            # Step 3: Create Todo2 task for tracking
            self._create_todo2_task()

            # Step 4: Use NetworkX for dependency analysis (if applicable)
            self._networkx_analysis()

            # Step 5: Execute analysis (implemented by subclasses)
            analysis_results = self._execute_analysis()

            # Store analysis results for tool wrappers to access
            self.results['results'] = analysis_results

            # Step 6: Generate insights using Tractatus
            insights = self._generate_insights(analysis_results)

            # Step 7: Store results in Todo2
            self._store_todo2_results(analysis_results, insights)

            # Step 8: Create follow-up tasks
            self._create_followup_tasks(analysis_results)

            # Step 9: Generate report
            report = self._generate_report(analysis_results, insights)

            # Step 10: Update Todo2 task
            self._update_todo2_complete()

            self.results['status'] = 'success'
            self.results['report'] = report

            logger.info(f"Intelligent automation completed: {self.automation_name}")
            return self.results

        except Exception as e:
            logger.error(f"Error in intelligent automation: {e}", exc_info=True)
            self.results['status'] = 'error'
            self.results['error'] = str(e)
            self._update_todo2_error(e)
            raise

    def _tractatus_analysis(self) -> None:
        """Use Tractatus Thinking to understand what to analyze."""
        logger.info("Starting Tractatus analysis...")

        concept = self._get_tractatus_concept()

        # Try to use Tractatus Thinking MCP server
        if MCP_AVAILABLE:
            try:
                mcp_client = get_mcp_client(self.project_root)
                result = mcp_client.call_tractatus_thinking('start', concept=concept)

                if result:
                    self.tractatus_session = {
                        'concept': concept,
                        'session_id': result.get('session_id'),
                        'components': result.get('components', []),
                        'dependencies': self._identify_dependencies(concept)
                    }
                    logger.info(f"Tractatus analysis complete: {len(self.tractatus_session['components'])} components identified")
                    return
            except Exception as e:
                logger.warning(f"Tractatus MCP call failed: {e}, using fallback")

        # Fallback to simplified analysis
        self.tractatus_session = {
            'concept': concept,
            'components': self._extract_components_from_concept(concept),
            'dependencies': self._identify_dependencies(concept)
        }
        logger.info(f"Tractatus analysis complete (fallback): {len(self.tractatus_session['components'])} components identified")

    def _sequential_planning(self) -> None:
        """Use Sequential Thinking to plan workflow."""
        logger.info("Starting Sequential planning...")

        problem = self._get_sequential_problem()

        # Try to use Sequential Thinking MCP server
        if MCP_AVAILABLE:
            try:
                mcp_client = get_mcp_client(self.project_root)
                result = mcp_client.call_sequential_thinking('start', problem=problem)

                if result:
                    self.sequential_session = {
                        'problem': problem,
                        'session_id': result.get('session_id'),
                        'steps': result.get('steps', []),
                        'current_step': 0
                    }
                    logger.info(f"Sequential planning complete: {len(self.sequential_session['steps'])} steps planned")
                    return
            except Exception as e:
                logger.warning(f"Sequential MCP call failed: {e}, using fallback")

        # Fallback to simplified planning
        self.sequential_session = {
            'problem': problem,
            'steps': self._plan_workflow_steps(problem),
            'current_step': 0
        }
        logger.info(f"Sequential planning complete (fallback): {len(self.sequential_session['steps'])} steps planned")

    def _create_todo2_task(self) -> None:
        """Create or reuse Todo2 task for tracking automation execution.

        Prevents duplicate task creation by:
        1. Checking for existing task with same name
        2. Reusing in-progress tasks
        3. Creating new task only if needed with unique ID
        """
        logger.info("Creating/reusing Todo2 task...")

        try:
            # Load Todo2 state
            todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
            if todo2_path.exists():
                with open(todo2_path) as f:
                    todo2_data = json.load(f)

                task_name = f"Automation: {self.automation_name}"

                # Check for existing task with same name
                existing_tasks = [
                    t for t in todo2_data.get('todos', [])
                    if t.get('name') == task_name
                ]

                # Look for reusable task (In Progress or Todo - handle both Title Case and lowercase)
                from project_management_automation.utils.todo2_utils import (
                    normalize_status,
                    normalize_status_to_title_case,
                )
                
                reusable_task = None
                for t in existing_tasks:
                    # Normalize status for comparison (handles both Title Case and lowercase)
                    status_normalized = normalize_status(t.get('status', ''))
                    if status_normalized in ('in_progress', 'todo'):
                        reusable_task = t
                        break

                if reusable_task:
                    # Reuse existing task
                    reusable_task['status'] = normalize_status_to_title_case('in_progress')
                    reusable_task['lastModified'] = datetime.now().isoformat()
                    annotate_task_project(reusable_task, self.project_id)
                    self.todo2_task = reusable_task
                    logger.info(f"Reusing Todo2 task: {reusable_task['id']}")

                    # Save updated state
                    with open(todo2_path, 'w') as f:
                        json.dump(todo2_data, f, indent=2)
                    return

                # Create new task with unique ID (include microseconds + counter)
                import random
                unique_suffix = f"{datetime.now().strftime('%f')[:4]}{random.randint(10, 99)}"
                task_id = f"AUTO-{datetime.now().strftime('%Y%m%d%H%M%S')}-{unique_suffix}"

                task = annotate_task_project({
                    'id': task_id,
                    'name': task_name,
                    'content': f"Automated {self.automation_name} execution",
                    'status': normalize_status_to_title_case('in_progress'),  # Use normalized Title Case
                    'priority': 'medium',
                    'tags': ['automation', self.automation_name.lower().replace(' ', '-')],
                    'created': datetime.now().isoformat(),
                    'lastModified': datetime.now().isoformat(),
                    'dependencies': []
                }, self.project_id)

                if 'todos' not in todo2_data:
                    todo2_data['todos'] = []

                todo2_data['todos'].append(task)

                # Save back
                with open(todo2_path, 'w') as f:
                    json.dump(todo2_data, f, indent=2)

                self.todo2_task = task
                logger.info(f"Todo2 task created: {task_id}")
            else:
                logger.warning("Todo2 state file not found, skipping task creation")
        except Exception as e:
            logger.warning(f"Failed to create Todo2 task: {e}")

    def _networkx_analysis(self) -> None:
        """Use NetworkX for dependency analysis if applicable."""
        if not self._needs_networkx():
            return

        logger.info("Starting NetworkX analysis...")

        try:
            import networkx as nx
            self.networkx_graph = self._build_networkx_graph()

            if self.networkx_graph and len(self.networkx_graph.nodes()) > 0:
                # Analyze graph
                analysis = {
                    'nodes': len(self.networkx_graph.nodes()),
                    'edges': len(self.networkx_graph.edges()),
                    'critical_path': self._find_critical_path(),
                    'bottlenecks': self._find_bottlenecks(),
                    'orphans': self._find_orphans(),
                    'density': nx.density(self.networkx_graph) if isinstance(self.networkx_graph, nx.Graph) else 0,
                    'is_dag': nx.is_directed_acyclic_graph(self.networkx_graph) if isinstance(self.networkx_graph, nx.DiGraph) else False
                }

                # Find strongly connected components if not DAG
                if isinstance(self.networkx_graph, nx.DiGraph) and not analysis['is_dag']:
                    try:
                        cycles = list(nx.simple_cycles(self.networkx_graph))
                        analysis['cycles'] = len(cycles)
                        analysis['cycle_details'] = cycles[:5]  # First 5 cycles
                    except Exception:
                        analysis['cycles'] = 0

                self.results['networkx_analysis'] = analysis
                logger.info(f"NetworkX analysis complete: {analysis['nodes']} nodes, {analysis['edges']} edges")
            else:
                logger.warning("NetworkX graph is empty, skipping analysis")
        except ImportError:
            logger.warning("NetworkX not available, install with: pip install networkx")
        except Exception as e:
            logger.warning(f"NetworkX analysis failed: {e}")

    def _store_todo2_results(self, analysis_results: dict, insights: str) -> None:
        """Store results in Todo2 task."""
        if not self.todo2_task:
            return

        try:
            todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
            if todo2_path.exists():
                with open(todo2_path) as f:
                    todo2_data = json.load(f)

                # Find task
                for task in todo2_data.get('todos', []):
                    if task['id'] == self.todo2_task['id']:
                        # Add result comment
                        if 'comments' not in task:
                            task['comments'] = []

                        result_comment = {
                            'id': f"{task['id']}-C-{len(task['comments']) + 1}",
                            'todoId': task['id'],
                            'type': 'result',
                            'content': f"**Automation Results:**\n\n{insights}\n\n**Key Findings:**\n{self._format_findings(analysis_results)}",
                            'created': datetime.now().isoformat(),
                            'lastModified': datetime.now().isoformat()
                        }

                        task['comments'].append(result_comment)
                        break

                # Save back
                with open(todo2_path, 'w') as f:
                    json.dump(todo2_data, f, indent=2)

                logger.info("Results stored in Todo2")
        except Exception as e:
            logger.warning(f"Failed to store Todo2 results: {e}")

    def _create_followup_tasks(self, analysis_results: dict) -> None:
        """Create follow-up tasks based on findings.

        Prevents duplicates by checking if task with same name already exists.
        Uses Todo2 MCP client (preferred) or falls back to direct file access.
        """
        if not self.todo2_task:
            return

        followup_tasks = self._identify_followup_tasks(analysis_results)

        if not followup_tasks:
            return

        try:
            # Try Todo2 MCP first (preferred)
            from project_management_automation.utils.todo2_mcp_client import (
                create_todos_mcp,
                list_todos_mcp,
            )
            
            # Get existing tasks to check for duplicates
            existing_tasks = list_todos_mcp(project_root=self.project_root)
            existing_names = {t.get('name') for t in existing_tasks}
            
            # Prepare tasks for creation (skip duplicates)
            todos_to_create = []
            for followup in followup_tasks:
                task_name = followup['name']
                if task_name in existing_names:
                    logger.debug(f"Skipping duplicate follow-up task: {task_name}")
                    continue
                
                todos_to_create.append({
                    'name': task_name,
                    'long_description': followup.get('description', task_name),
                    'status': 'Todo',
                    'priority': followup.get('priority', 'medium'),
                    'tags': followup.get('tags', ['automation', 'followup']),
                    'dependencies': [self.todo2_task['id']] if self.todo2_task else [],
                })
            
            if todos_to_create:
                created_ids = create_todos_mcp(todos_to_create, project_root=self.project_root)
                if created_ids:
                    self.results['followup_tasks'].extend(created_ids)
                    logger.info(f"Created {len(created_ids)} follow-up tasks via Todo2 MCP")
                    return
            
            # Fallback to direct file access
            todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
            if todo2_path.exists():
                with open(todo2_path) as f:
                    todo2_data = json.load(f)

                # Get existing task names for duplicate checking
                existing_names = {
                    t.get('name') for t in todo2_data.get('todos', [])
                }

                created_count = 0
                skipped_count = 0

                for followup in followup_tasks:
                    task_name = followup['name']

                    # Skip if task with same name already exists
                    if task_name in existing_names:
                        logger.debug(f"Skipping duplicate follow-up task: {task_name}")
                        skipped_count += 1
                        continue

                    # Create unique ID with random suffix
                    import random
                    unique_suffix = f"{datetime.now().strftime('%f')[:4]}{random.randint(10, 99)}"
                    task_id = f"T-{datetime.now().strftime('%Y%m%d%H%M%S')}-{unique_suffix}"

                    task = annotate_task_project({
                        'id': task_id,
                        'name': task_name,
                        'content': followup.get('description', task_name),
                        'status': 'todo',
                        'priority': followup.get('priority', 'medium'),
                        'tags': followup.get('tags', ['automation', 'followup']),
                        'dependencies': [self.todo2_task['id']] if self.todo2_task else [],
                        'created': datetime.now().isoformat(),
                        'lastModified': datetime.now().isoformat()
                    }, self.project_id)

                    if 'todos' not in todo2_data:
                        todo2_data['todos'] = []

                    todo2_data['todos'].append(task)
                    existing_names.add(task_name)  # Track newly created
                    self.results['followup_tasks'].append(task['id'])
                    created_count += 1

                # Save back
                with open(todo2_path, 'w') as f:
                    json.dump(todo2_data, f, indent=2)

                if created_count > 0:
                    logger.info(f"Created {created_count} follow-up tasks (skipped {skipped_count} duplicates)")
                elif skipped_count > 0:
                    logger.info(f"All {skipped_count} follow-up tasks already exist")
        except Exception as e:
            logger.warning(f"Failed to create follow-up tasks: {e}")

    def _update_todo2_complete(self) -> None:
        """Update Todo2 task as complete."""
        if not self.todo2_task:
            return

        try:
            # Use MCP client (preferred) with fallback to file access
            from project_management_automation.utils.todo2_mcp_client import update_todos_mcp
            
            success = update_todos_mcp(
                updates=[{
                    'id': self.todo2_task['id'],
                    'status': 'Done'
                }],
                project_root=self.project_root
            )
            
            if success:
                logger.info(f"Todo2 task {self.todo2_task['id']} marked as complete via MCP")
            else:
                # Fallback to direct file access
                logger.debug("MCP update failed, falling back to file access")
                todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
                if todo2_path.exists():
                    with open(todo2_path) as f:
                        todo2_data = json.load(f)

                    from project_management_automation.utils.todo2_utils import normalize_status_to_title_case
                    for task in todo2_data.get('todos', []):
                        if task['id'] == self.todo2_task['id']:
                            task['status'] = normalize_status_to_title_case('Done')
                            task['lastModified'] = datetime.now().isoformat()
                            break

                    with open(todo2_path, 'w') as f:
                        json.dump(todo2_data, f, indent=2)

                    logger.info(f"Todo2 task {self.todo2_task['id']} marked as complete via file access")
        except Exception as e:
            logger.warning(f"Failed to update Todo2 task: {e}", exc_info=True)

    def _update_todo2_error(self, error: Exception) -> None:
        """Update Todo2 task with error."""
        if not self.todo2_task:
            return

        try:
            todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
            if todo2_path.exists():
                with open(todo2_path) as f:
                    todo2_data = json.load(f)

                for task in todo2_data.get('todos', []):
                    if task['id'] == self.todo2_task['id']:
                        from project_management_automation.utils.todo2_utils import normalize_status_to_title_case
                        task['status'] = normalize_status_to_title_case('todo')
                        task['lastModified'] = datetime.now().isoformat()

                        if 'comments' not in task:
                            task['comments'] = []

                        error_comment = {
                            'id': f"{task['id']}-C-{len(task['comments']) + 1}",
                            'todoId': task['id'],
                            'type': 'note',
                            'content': f"**Error:** {str(error)}",
                            'created': datetime.now().isoformat(),
                            'lastModified': datetime.now().isoformat()
                        }

                        task['comments'].append(error_comment)
                        break

                with open(todo2_path, 'w') as f:
                    json.dump(todo2_data, f, indent=2)
        except Exception as e:
            logger.warning(f"Failed to update Todo2 task with error: {e}")

    # Abstract methods to be implemented by subclasses

    @abstractmethod
    def _get_tractatus_concept(self) -> str:
        """Return the concept to analyze with Tractatus Thinking."""
        pass

    @abstractmethod
    def _get_sequential_problem(self) -> str:
        """Return the problem to solve with Sequential Thinking."""
        pass

    @abstractmethod
    def _execute_analysis(self) -> dict:
        """Execute the actual analysis - implemented by subclasses."""
        pass

    @abstractmethod
    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from analysis results."""
        pass

    @abstractmethod
    def _generate_report(self, analysis_results: dict, insights: str) -> str:
        """Generate final report."""
        pass

    # Helper methods with default implementations

    def _extract_components_from_concept(self, concept: str) -> list[str]:
        """Extract atomic components from concept (simplified Tractatus)."""
        # This is a simplified version - real implementation would use MCP server
        keywords = ['automation', 'analysis', 'validation', 'monitoring', 'tracking', 'synchronization']
        return [kw for kw in keywords if kw in concept.lower()]

    def _identify_dependencies(self, concept: str) -> list[str]:
        """Identify dependencies (simplified)."""
        return []

    def _plan_workflow_steps(self, problem: str) -> list[str]:
        """Plan workflow steps (simplified Sequential)."""
        return [
            "Load and analyze data",
            "Identify patterns and opportunities",
            "Generate recommendations",
            "Create follow-up tasks"
        ]

    def _needs_networkx(self) -> bool:
        """Determine if NetworkX analysis is needed."""
        return False

    def _build_networkx_graph(self):
        """Build NetworkX graph - override if needed."""
        return None

    def _find_critical_path(self) -> list[str]:
        """Find critical path in graph."""
        if not self.networkx_graph:
            return []

        try:
            import networkx as nx

            if not isinstance(self.networkx_graph, nx.DiGraph):
                return []

            # Find longest path (critical path)
            sources = [n for n in self.networkx_graph.nodes() if self.networkx_graph.in_degree(n) == 0]
            sinks = [n for n in self.networkx_graph.nodes() if self.networkx_graph.out_degree(n) == 0]

            longest_path = []
            for source in sources:
                for sink in sinks:
                    try:
                        paths = list(nx.all_simple_paths(self.networkx_graph, source, sink))
                        if paths:
                            longest = max(paths, key=len)
                            if len(longest) > len(longest_path):
                                longest_path = longest
                    except (nx.NetworkXNoPath, nx.NetworkXError):
                        continue

            return longest_path
        except Exception as e:
            logger.warning(f"Critical path analysis failed: {e}")
            return []

    def _find_bottlenecks(self) -> list[str]:
        """Find bottlenecks in graph (nodes with high out-degree)."""
        if not self.networkx_graph:
            return []

        try:

            # Find nodes with high out-degree (many dependents)
            out_degrees = dict(self.networkx_graph.out_degree())
            if not out_degrees:
                return []

            # Nodes with out-degree > 3 are potential bottlenecks
            threshold = max(3, max(out_degrees.values()) * 0.3) if out_degrees.values() else 3
            bottlenecks = [node for node, degree in out_degrees.items() if degree >= threshold]

            return sorted(bottlenecks, key=lambda n: out_degrees[n], reverse=True)[:10]
        except Exception as e:
            logger.warning(f"Bottleneck analysis failed: {e}")
            return []

    def _find_orphans(self) -> list[str]:
        """Find orphaned nodes in graph (no incoming edges)."""
        if not self.networkx_graph:
            return []

        try:
            import networkx as nx

            if isinstance(self.networkx_graph, nx.DiGraph):
                # Nodes with no incoming edges
                orphans = [n for n in self.networkx_graph.nodes() if self.networkx_graph.in_degree(n) == 0]
            else:
                # Nodes with no edges
                orphans = [n for n in self.networkx_graph.nodes() if self.networkx_graph.degree(n) == 0]

            return orphans
        except Exception as e:
            logger.warning(f"Orphan analysis failed: {e}")
            return []

    def _identify_followup_tasks(self, analysis_results: dict) -> list[dict]:
        """Identify follow-up tasks from analysis results."""
        return []

    def _format_findings(self, analysis_results: dict) -> str:
        """Format findings for Todo2 comment.

        Filters out large lists to prevent bloating the todo2 file.
        Max comment size: 10KB
        """
        # Filter out large lists (like full task dumps)
        filtered = {}
        for key, value in analysis_results.items():
            if isinstance(value, list):
                if len(value) > 10:
                    # Summarize large lists
                    filtered[key] = f"[{len(value)} items - see logs for details]"
                elif len(json.dumps(value)) > 5000:
                    # Truncate lists with large items
                    filtered[key] = f"[{len(value)} items - truncated]"
                else:
                    filtered[key] = value
            elif isinstance(value, dict) and len(json.dumps(value)) > 5000:
                # Truncate large dicts
                filtered[key] = "{...truncated...}"
            else:
                filtered[key] = value

        # Use compact JSON, limit total size
        result = json.dumps(filtered, separators=(',', ':'))
        if len(result) > 10000:
            return result[:10000] + "...truncated"
        return result

    def _fallback_component_extraction(self, concept: str) -> list[str]:
        """Fallback component extraction."""
        return self._extract_components_from_concept(concept)

    def _fallback_workflow_steps(self) -> list[str]:
        """Fallback workflow steps."""
        return self._plan_workflow_steps("")
