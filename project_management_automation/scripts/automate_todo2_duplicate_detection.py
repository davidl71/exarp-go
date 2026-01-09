#!/usr/bin/env python3
"""
Automated Todo2 Duplicate Task Detection Script

Detects duplicate tasks in Todo2 by analyzing:
- Task names (exact and fuzzy matches)
- Task descriptions (similarity)
- Task IDs (should be unique)
- Task content (long descriptions)

Uses IntelligentAutomationBase for consistency with other automation tools.
"""

import argparse
import json
import logging
import sys
from collections import Counter, defaultdict
from datetime import datetime
from difflib import SequenceMatcher
from pathlib import Path
from typing import Optional

# Add project root to path
# Project root will be passed to __init__
# Import base class
from project_management_automation.scripts.base.intelligent_automation_base import IntelligentAutomationBase
from project_management_automation.utils.todo2_utils import (
    filter_tasks_by_project,
    get_repo_project_id,
)

# Configure logging (will be configured after project_root is set)
logger = logging.getLogger(__name__)


class Todo2DuplicateDetector(IntelligentAutomationBase):
    """Intelligent Todo2 duplicate task detector using base class."""

    def __init__(self, config: dict, project_root: Optional[Path] = None):
        from project_management_automation.utils import find_project_root
        if project_root is None:
            project_root = find_project_root()
        super().__init__(config, "Todo2 Duplicate Detection", project_root)
        self.todo2_path = self.project_root / '.todo2' / 'state.todo2.json'
        self.output_path = self.project_root / config.get('output_path', 'docs/TODO2_DUPLICATE_DETECTION_REPORT.md')
        self.similarity_threshold = config.get('similarity_threshold', 0.85)
        self.auto_fix = config.get('auto_fix', False)

        # Detection results
        self.duplicates = {
            'duplicate_ids': [],  # Multiple tasks with same ID
            'exact_name_matches': [],  # Tasks with identical names
            'similar_name_matches': [],  # Tasks with similar names (>threshold)
            'similar_description_matches': [],  # Tasks with similar descriptions
            'self_dependencies': []  # Tasks that depend on themselves
        }

    def _get_tractatus_concept(self) -> str:
        """Tractatus concept: What is duplicate detection?"""
        return "What is duplicate detection? Duplicate Detection = Unique ID × Name Uniqueness × Description Similarity × Dependency Validity"

    def _get_sequential_problem(self) -> str:
        """Sequential problem: How do we detect duplicates?"""
        return "How do we systematically detect duplicate tasks in Todo2?"

    def _execute_analysis(self) -> dict:
        """Execute duplicate detection analysis."""
        logger.info("Executing Todo2 duplicate detection...")

        # Load tasks
        tasks = self._load_todo2_tasks()
        if not tasks:
            logger.error("No tasks found in Todo2 state file")
            return {'error': 'No tasks found'}

        logger.info(f"Analyzing {len(tasks)} tasks...")

        # Detect duplicates
        self._detect_duplicate_ids(tasks)
        self._detect_exact_name_matches(tasks)
        self._detect_similar_name_matches(tasks)
        self._detect_similar_descriptions(tasks)
        self._detect_self_dependencies(tasks)

        # Summary
        total_duplicates = (
            len(self.duplicates['duplicate_ids']) +
            len(self.duplicates['exact_name_matches']) +
            len(self.duplicates['similar_name_matches']) +
            len(self.duplicates['similar_description_matches']) +
            len(self.duplicates['self_dependencies'])
        )

        results = {
            'total_tasks': len(tasks),
            'duplicates_found': total_duplicates,
            'duplicate_ids': self.duplicates['duplicate_ids'],
            'exact_name_matches': self.duplicates['exact_name_matches'],
            'similar_name_matches': self.duplicates['similar_name_matches'],
            'similar_description_matches': self.duplicates['similar_description_matches'],
            'self_dependencies': self.duplicates['self_dependencies']
        }

        logger.info(f"Found {total_duplicates} duplicate issues")

        # Apply auto-fix if enabled
        if self.auto_fix and total_duplicates > 0:
            fix_results = self._apply_auto_fix(tasks)
            results['auto_fix_applied'] = fix_results['applied']
            results['tasks_removed'] = fix_results['tasks_removed']
            results['tasks_merged'] = fix_results['tasks_merged']
            results['dependencies_updated'] = fix_results['dependencies_updated']
            if fix_results['applied']:
                logger.info(f"Auto-fix applied: {fix_results['tasks_removed']} tasks removed, {fix_results['tasks_merged']} tasks merged")
            # Store in self.results for access by wrapper
            self.results['auto_fix_applied'] = fix_results['applied']
            self.results['tasks_removed'] = fix_results['tasks_removed']
            self.results['tasks_merged'] = fix_results['tasks_merged']
            self.results['dependencies_updated'] = fix_results['dependencies_updated']
        else:
            results['auto_fix_applied'] = False
            results['tasks_removed'] = 0
            results['tasks_merged'] = 0
            results['dependencies_updated'] = 0

        # Store in results for reporting
        results['duplicates_found'] = total_duplicates
        return results

    def _load_todo2_tasks(self) -> list[dict]:
        """Load tasks from Todo2 MCP (preferred) or state file (fallback)."""
        from project_management_automation.utils.todo2_mcp_client import list_todos_mcp
        
        # Try Todo2 MCP first (preferred)
        try:
            mcp_tasks = list_todos_mcp(project_root=self.project_root)
            if mcp_tasks:
                project_id = get_repo_project_id(self.project_root)
                filtered = filter_tasks_by_project(mcp_tasks, project_id, logger=logger)
                logger.info("Loaded %d tasks from Todo2 MCP (%d matched project)", len(mcp_tasks), len(filtered))
                return filtered
        except Exception as e:
            logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")
        
        # Fallback to direct file access
        try:
            with open(self.todo2_path) as f:
                data = json.load(f)
            tasks = data.get('todos', [])
            project_id = get_repo_project_id(self.project_root)
            filtered = filter_tasks_by_project(tasks, project_id, logger=logger)
            logger.info("Loaded %d tasks from file (%d matched project)", len(tasks), len(filtered))
            return filtered
        except FileNotFoundError:
            logger.error(f"Todo2 state file not found: {self.todo2_path}")
            return []
        except json.JSONDecodeError as e:
            logger.error(f"Invalid JSON in Todo2 state file: {e}")
            return []

    def _detect_duplicate_ids(self, tasks: list[dict]):
        """Detect tasks with duplicate IDs (should never happen)."""
        id_counts = Counter([t['id'] for t in tasks])
        duplicates = {id: count for id, count in id_counts.items() if count > 1}

        for task_id, count in duplicates.items():
            duplicate_tasks = [t for t in tasks if t['id'] == task_id]
            self.duplicates['duplicate_ids'].append({
                'id': task_id,
                'count': count,
                'tasks': [
                    {
                        'id': t['id'],
                        'name': t.get('name', ''),
                        'status': t.get('status', 'unknown'),
                        'created': t.get('created', 'unknown')
                    }
                    for t in duplicate_tasks
                ]
            })

    def _should_skip_task(self, task: dict) -> bool:
        """Check if task should be skipped in duplicate detection."""
        task_id = task.get('id', '')
        name = task.get('name', '').lower()
        
        # Skip AUTO- tasks (automation execution records, not duplicates)
        if task_id.startswith('AUTO-'):
            return True
        
        # Skip tasks with "Automation: " prefix that are done (execution records)
        if name.startswith('automation:') and task.get('status', '').lower() in ['done', 'completed']:
            return True
            
        return False

    def _detect_exact_name_matches(self, tasks: list[dict]):
        """Detect tasks with identical names."""
        name_to_tasks = defaultdict(list)
        for task in tasks:
            # Skip AUTO- tasks and automation execution records
            if self._should_skip_task(task):
                continue
                
            name = task.get('name', '').strip().lower()
            if name:
                name_to_tasks[name].append(task)

        for _name, task_list in name_to_tasks.items():
            if len(task_list) > 1:
                # Filter out tasks with different IDs (true duplicates)
                task_ids = [t['id'] for t in task_list]
                if len(set(task_ids)) > 1:  # Different IDs = potential duplicates
                    self.duplicates['exact_name_matches'].append({
                        'name': task_list[0].get('name', ''),
                        'count': len(task_list),
                        'tasks': [
                            {
                                'id': t['id'],
                                'name': t.get('name', ''),
                                'status': t.get('status', 'unknown'),
                                'created': t.get('created', 'unknown'),
                                'priority': t.get('priority', 'none')
                            }
                            for t in task_list
                        ]
                    })

    def _is_template_similarity(self, name1: str, name2: str) -> bool:
        """Check if two names only differ by a single word (template pattern)."""
        words1 = name1.split()
        words2 = name2.split()
        
        # If same number of words and all but one match, it's a template pattern
        if len(words1) == len(words2):
            differences = sum(1 for w1, w2 in zip(words1, words2) if w1 != w2)
            # If only 1-2 words differ (and they're at same positions), it's a template
            if differences <= 2:
                # Check if the differences are substitutions (not additions/deletions)
                # Count word-by-word differences
                diff_count = 0
                for w1, w2 in zip(words1, words2):
                    if w1 != w2:
                        diff_count += 1
                
                # If only 1-2 word substitutions, likely a template (e.g., "consolidated_analysis" vs "consolidated_memory")
                if diff_count <= 2:
                    # Additional check: if they share a common prefix pattern
                    # Extract common prefix/suffix
                    common_prefix_len = 0
                    common_suffix_len = 0
                    
                    # Find common prefix
                    for i in range(min(len(words1), len(words2))):
                        if words1[i] == words2[i]:
                            common_prefix_len += 1
                        else:
                            break
                    
                    # Find common suffix
                    for i in range(1, min(len(words1), len(words2)) + 1):
                        if words1[-i] == words2[-i]:
                            common_suffix_len += 1
                        else:
                            break
                    
                    # If they share significant prefix/suffix, it's a template
                    total_words = len(words1)
                    if common_prefix_len + common_suffix_len >= total_words - 2:
                        return True
        
        # Fallback: check word sets for single word difference
        words1_set = set(words1)
        words2_set = set(words2)
        intersection = words1_set & words2_set
        union = words1_set | words2_set
        
        # If they share most words (>75%) and only differ by 1-2 words, likely template
        if len(union) > 0:
            overlap_ratio = len(intersection) / len(union)
            diff1 = words1_set - words2_set
            diff2 = words2_set - words1_set
            
            # High overlap (>75%) with 1-2 word differences = template pattern
            if overlap_ratio > 0.75 and len(diff1) <= 2 and len(diff2) <= 2:
                return True
        
        return False

    def _detect_similar_name_matches(self, tasks: list[dict]):
        """Detect tasks with similar names (fuzzy matching)."""
        for i, task1 in enumerate(tasks):
            # Skip AUTO- tasks and automation execution records
            if self._should_skip_task(task1):
                continue
                
            name1 = task1.get('name', '').strip().lower()
            if not name1 or len(name1) < 10:  # Skip very short names
                continue

            for task2 in tasks[i + 1:]:
                # Skip AUTO- tasks and automation execution records
                if self._should_skip_task(task2):
                    continue
                    
                name2 = task2.get('name', '').strip().lower()
                if not name2 or task1['id'] == task2['id']:
                    continue

                # Skip template patterns (tasks that only differ by module/file names)
                if self._is_template_similarity(name1, name2):
                    continue

                similarity = SequenceMatcher(None, name1, name2).ratio()
                if similarity >= self.similarity_threshold:
                    self.duplicates['similar_name_matches'].append({
                        'similarity': similarity,
                        'tasks': [
                            {
                                'id': task1['id'],
                                'name': task1.get('name', ''),
                                'status': task1.get('status', 'unknown')
                            },
                            {
                                'id': task2['id'],
                                'name': task2.get('name', ''),
                                'status': task2.get('status', 'unknown')
                            }
                        ]
                    })

    def _detect_similar_descriptions(self, tasks: list[dict]):
        """Detect tasks with similar long descriptions."""
        for i, task1 in enumerate(tasks):
            # Skip AUTO- tasks and automation execution records
            if self._should_skip_task(task1):
                continue
                
            desc1 = task1.get('long_description', '').strip()
            if not desc1 or len(desc1) < 50:  # Skip very short descriptions
                continue

            for task2 in tasks[i + 1:]:
                # Skip AUTO- tasks and automation execution records
                if self._should_skip_task(task2):
                    continue
                    
                desc2 = task2.get('long_description', '').strip()
                if not desc2 or task1['id'] == task2['id']:
                    continue

                # Skip if task names match template pattern (descriptions will too)
                name1 = task1.get('name', '').strip().lower()
                name2 = task2.get('name', '').strip().lower()
                if name1 and name2 and self._is_template_similarity(name1, name2):
                    continue

                similarity = SequenceMatcher(None, desc1, desc2).ratio()
                if similarity >= self.similarity_threshold:
                    self.duplicates['similar_description_matches'].append({
                        'similarity': similarity,
                        'tasks': [
                            {
                                'id': task1['id'],
                                'name': task1.get('name', ''),
                                'status': task1.get('status', 'unknown')
                            },
                            {
                                'id': task2['id'],
                                'name': task2.get('name', ''),
                                'status': task2.get('status', 'unknown')
                            }
                        ]
                    })

    def _detect_self_dependencies(self, tasks: list[dict]):
        """Detect tasks that depend on themselves (invalid)."""
        for task in tasks:
            deps = task.get('dependencies', [])
            task_id = task['id']
            if task_id in deps:
                self.duplicates['self_dependencies'].append({
                    'id': task_id,
                    'name': task.get('name', ''),
                    'dependencies': deps
                })

    def _apply_auto_fix(self, tasks: list[dict]) -> dict:
        """Apply auto-fix to consolidate duplicates."""
        tasks_removed = 0
        tasks_merged = 0
        dependencies_updated = 0
        tasks_to_remove = set()
        tasks_to_keep = {}  # Map: keep_id -> task dict

        # Load current state
        try:
            with open(self.todo2_path) as f:
                state = json.load(f)
        except Exception as e:
            logger.error(f"Failed to load Todo2 state: {e}")
            return {'applied': False, 'tasks_removed': 0, 'tasks_merged': 0, 'dependencies_updated': 0}

        current_tasks = state.get('todos', [])

        # Process exact name matches
        for match in self.duplicates['exact_name_matches']:
            match_tasks = match['tasks']
            # Keep the "best" task (prefer "In Progress" > most comments > most recent)
            best_task = self._select_best_task(match_tasks, current_tasks)
            best_id = best_task['id']
            tasks_to_keep[best_id] = best_task

            # Mark others for removal and merge their data
            for task in match_tasks:
                if task['id'] != best_id:
                    tasks_to_remove.add(task['id'])
                    self._merge_task_data(best_id, task['id'], current_tasks)
                    tasks_merged += 1

        # Process similar name matches (only if very similar, threshold >= 0.95)
        for match in self.duplicates['similar_name_matches']:
            if match['similarity'] >= 0.95:  # Only very similar tasks
                match_tasks = match['tasks']
                task1_id = match_tasks[0]['id']
                task2_id = match_tasks[1]['id']

                # Skip if already processed
                if task1_id in tasks_to_remove or task2_id in tasks_to_remove:
                    continue

                # Select best task
                task1 = next((t for t in current_tasks if t['id'] == task1_id), None)
                task2 = next((t for t in current_tasks if t['id'] == task2_id), None)
                if not task1 or not task2:
                    continue

                best_task = self._select_best_task([task1, task2], current_tasks)
                best_id = best_task['id']
                tasks_to_keep[best_id] = best_task

                # Remove the other
                other_id = task2_id if best_id == task1_id else task1_id
                tasks_to_remove.add(other_id)
                self._merge_task_data(best_id, other_id, current_tasks)
                tasks_merged += 1

        # Fix self-dependencies
        for self_dep in self.duplicates['self_dependencies']:
            task_id = self_dep['id']
            task = next((t for t in current_tasks if t['id'] == task_id), None)
            if task:
                deps = task.get('dependencies', [])
                if task_id in deps:
                    deps.remove(task_id)
                    task['dependencies'] = deps
                    dependencies_updated += 1

        # Update dependencies: replace references to removed tasks with kept tasks
        id_mapping = {}  # Map: removed_id -> kept_id (for exact name matches)
        for match in self.duplicates['exact_name_matches']:
            match_tasks = match['tasks']
            best_id = self._select_best_task(match_tasks, current_tasks)['id']
            for task in match_tasks:
                if task['id'] != best_id:
                    id_mapping[task['id']] = best_id

        for task in current_tasks:
            deps = task.get('dependencies', [])
            updated = False
            new_deps = []
            for dep_id in deps:
                if dep_id in tasks_to_remove:
                    # Map to kept task if available
                    if dep_id in id_mapping:
                        new_deps.append(id_mapping[dep_id])
                        updated = True
                        dependencies_updated += 1
                    # Otherwise, skip (remove) this dependency
                else:
                    # Keep dependency
                    new_deps.append(dep_id)

            if updated:
                task['dependencies'] = list(set(new_deps))  # Remove duplicates
                task['lastModified'] = datetime.now().isoformat()

        # Remove duplicate tasks
        current_tasks = [t for t in current_tasks if t['id'] not in tasks_to_remove]
        tasks_removed = len(tasks_to_remove)

        # Update via MCP (preferred) or file (fallback)
        from project_management_automation.utils.todo2_mcp_client import (
            delete_todos_mcp,
            update_todos_mcp,
        )
        
        # Try Todo2 MCP first (preferred)
        try:
            # Delete removed tasks
            if tasks_to_remove:
                delete_todos_mcp(list(tasks_to_remove), project_root=self.project_root)
            
            # Update modified tasks (those with merged data or fixed dependencies)
            updates = []
            for task in current_tasks:
                # Check if task was modified (has merge comment or dependency update)
                if task.get('lastModified') and task.get('lastModified') != task.get('created'):
                    updates.append({
                        'id': task['id'],
                        'dependencies': task.get('dependencies', []),
                        'tags': task.get('tags', []),
                        'comments': task.get('comments', [])
                    })
            
            if updates:
                update_todos_mcp(updates, project_root=self.project_root)
            
            logger.info(f"Applied auto-fix via MCP: {tasks_removed} removed, {tasks_merged} merged")
            return {
                'applied': True,
                'tasks_removed': tasks_removed,
                'tasks_merged': tasks_merged,
                'dependencies_updated': dependencies_updated
            }
        except Exception as e:
            logger.debug(f"Todo2 MCP not available: {e}, falling back to file access")
            
            # Fallback to direct file access
            state['todos'] = current_tasks
            state['lastModified'] = datetime.now().isoformat()
            
            try:
                with open(self.todo2_path, 'w') as f:
                    json.dump(state, f, indent=2)
                logger.info(f"Todo2 state updated: {tasks_removed} tasks removed")
                return {
                    'applied': True,
                    'tasks_removed': tasks_removed,
                    'tasks_merged': tasks_merged,
                    'dependencies_updated': dependencies_updated
                }
            except Exception as e:
                logger.error(f"Failed to save Todo2 state: {e}")
                return {'applied': False, 'tasks_removed': 0, 'tasks_merged': 0, 'dependencies_updated': 0}

    def _select_best_task(self, tasks: list[dict], all_tasks: list[dict]) -> dict:
        """Select the best task to keep from a list of duplicates."""
        # Get full task data
        full_tasks = []
        for task in tasks:
            full_task = next((t for t in all_tasks if t['id'] == task['id']), None)
            if full_task:
                full_tasks.append(full_task)

        if not full_tasks:
            return tasks[0]

        # Priority: In Progress > most comments > most recent > first
        def score_task(task):
            score = 0
            status = task.get('status', '').lower()
            if status == 'in progress':
                score += 1000
            elif status == 'done':
                score += 100
            elif status == 'review':
                score += 50

            # Count comments
            comments = task.get('comments', [])
            score += len(comments) * 10

            # Recency bonus (more recent = higher score)
            last_modified = task.get('lastModified', task.get('created', ''))
            if last_modified:
                try:
                    from datetime import datetime
                    dt = datetime.fromisoformat(last_modified.replace('Z', '+00:00'))
                    # Days since epoch (more recent = higher number)
                    score += (dt - datetime(1970, 1, 1)).days
                except (ValueError, TypeError):
                    pass

            return score

        return max(full_tasks, key=score_task)

    def _merge_task_data(self, keep_id: str, remove_id: str, all_tasks: list[dict]):
        """Merge data from removed task into kept task."""
        keep_task = next((t for t in all_tasks if t['id'] == keep_id), None)
        remove_task = next((t for t in all_tasks if t['id'] == remove_id), None)

        if not keep_task or not remove_task:
            return

        # Merge dependencies
        keep_deps = set(keep_task.get('dependencies', []))
        remove_deps = set(remove_task.get('dependencies', []))
        keep_deps.update(remove_deps)
        keep_deps.discard(keep_id)  # Remove self-reference
        keep_deps.discard(remove_id)  # Remove reference to removed task
        keep_task['dependencies'] = list(keep_deps)

        # Merge comments (add note about merge)
        keep_comments = keep_task.get('comments', [])
        remove_comments = remove_task.get('comments', [])
        merge_comment = {
            'type': 'note',
            'content': f"Merged with duplicate task {remove_id}: {remove_task.get('name', '')}",
            'created': datetime.now().isoformat()
        }
        keep_comments.append(merge_comment)
        keep_comments.extend(remove_comments)
        keep_task['comments'] = keep_comments

        # Merge tags
        keep_tags = set(keep_task.get('tags', []))
        remove_tags = set(remove_task.get('tags', []))
        keep_tags.update(remove_tags)
        keep_task['tags'] = list(keep_tags)

        # Update last modified
        keep_task['lastModified'] = datetime.now().isoformat()

    def _generate_insights(self, analysis_results: dict) -> str:
        """Generate insights from duplicate detection results."""
        insights = []

        total_duplicates = analysis_results.get('duplicates_found', 0)
        if total_duplicates == 0:
            insights.append("✅ **No duplicates found!** Your Todo2 task list is clean.")
        else:
            insights.append(f"⚠️ **Found {total_duplicates} duplicate issues**")

            if self.duplicates['duplicate_ids']:
                insights.append(f"🚨 **CRITICAL**: {len(self.duplicates['duplicate_ids'])} duplicate task IDs (data integrity issue)")

            if self.duplicates['exact_name_matches']:
                insights.append(f"🔍 **{len(self.duplicates['exact_name_matches'])} exact name matches** (likely duplicates)")

            if self.duplicates['similar_name_matches']:
                insights.append(f"🔍 **{len(self.duplicates['similar_name_matches'])} similar name matches** (potential duplicates)")

            if self.duplicates['self_dependencies']:
                insights.append(f"⚠️ **{len(self.duplicates['self_dependencies'])} self-dependencies** (invalid)")

        return "\n\n".join(insights)

    def _generate_report(self, analysis_results: dict, insights: Optional[str] = None) -> str:
        """Generate markdown report."""
        report = f"""# Todo2 Duplicate Task Detection Report

**Generated**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
**Total Tasks Analyzed**: {analysis_results.get('total_tasks', 0)}
**Duplicates Found**: {analysis_results.get('duplicates_found', 0)}

---

## Summary

"""

        # Duplicate IDs (critical)
        if self.duplicates['duplicate_ids']:
            report += f"### ⚠️ Critical: Duplicate Task IDs ({len(self.duplicates['duplicate_ids'])})\n\n"
            report += "**This should never happen!** Multiple tasks with the same ID found.\n\n"
            for dup in self.duplicates['duplicate_ids']:
                report += f"**Task ID: {dup['id']}** (appears {dup['count']} times)\n\n"
                for task in dup['tasks']:
                    report += f"- `{task['id']}`: {task['name']} (Status: {task['status']}, Created: {task['created']})\n"
                report += "\n"

        # Exact name matches
        if self.duplicates['exact_name_matches']:
            report += f"### 🔍 Exact Name Matches ({len(self.duplicates['exact_name_matches'])})\n\n"
            report += "Tasks with identical names but different IDs:\n\n"
            for match in self.duplicates['exact_name_matches']:
                report += f"**Name**: \"{match['name']}\" ({match['count']} tasks)\n\n"
                for task in match['tasks']:
                    report += f"- `{task['id']}`: {task['name']} (Status: {task['status']}, Priority: {task['priority']})\n"
                report += "\n"

        # Similar name matches
        if self.duplicates['similar_name_matches']:
            report += f"### 🔍 Similar Name Matches ({len(self.duplicates['similar_name_matches'])})\n\n"
            report += f"Tasks with similar names (similarity ≥ {self.similarity_threshold * 100:.0f}%):\n\n"
            for match in self.duplicates['similar_name_matches'][:20]:  # Limit to top 20
                report += f"**Similarity**: {match['similarity'] * 100:.1f}%\n"
                for task in match['tasks']:
                    report += f"- `{task['id']}`: {task['name']} (Status: {task['status']})\n"
                report += "\n"

        # Similar descriptions
        if self.duplicates['similar_description_matches']:
            report += f"### 📝 Similar Description Matches ({len(self.duplicates['similar_description_matches'])})\n\n"
            report += f"Tasks with similar descriptions (similarity ≥ {self.similarity_threshold * 100:.0f}%):\n\n"
            for match in self.duplicates['similar_description_matches'][:10]:  # Limit to top 10
                report += f"**Similarity**: {match['similarity'] * 100:.1f}%\n"
                for task in match['tasks']:
                    report += f"- `{task['id']}`: {task['name']} (Status: {task['status']})\n"
                report += "\n"

        # Self-dependencies
        if self.duplicates['self_dependencies']:
            report += f"### ⚠️ Self-Dependencies ({len(self.duplicates['self_dependencies'])})\n\n"
            report += "Tasks that depend on themselves (invalid):\n\n"
            for task in self.duplicates['self_dependencies']:
                report += f"- `{task['id']}`: {task['name']}\n"
                report += f"  Dependencies: {', '.join(task['dependencies'])}\n\n"

        # Recommendations
        report += "---\n\n## Recommendations\n\n"
        if analysis_results.get('duplicates_found', 0) == 0:
            report += "✅ **No duplicates found!** Your Todo2 task list is clean.\n\n"
        else:
            report += "### Action Items\n\n"
            if self.duplicates['duplicate_ids']:
                report += "1. **CRITICAL**: Fix duplicate IDs immediately - this is a data integrity issue\n"
            if self.duplicates['exact_name_matches']:
                report += "2. Review exact name matches and consolidate duplicates\n"
            if self.duplicates['similar_name_matches']:
                report += "3. Review similar name matches for potential consolidation\n"
            if self.duplicates['self_dependencies']:
                report += "4. Remove self-dependencies (tasks cannot depend on themselves)\n"
            report += "\n"

        report += "### How to Fix\n\n"
        report += "1. Review the duplicate tasks listed above\n"
        report += "2. Determine which task to keep (usually the one with more comments/history)\n"
        report += "3. Update dependencies pointing to deleted tasks\n"
        report += "4. Delete duplicate tasks\n"
        report += "5. Re-run this script to verify fixes\n\n"

        report += "---\n\n"
        report += "*Report generated by `scripts/automate_todo2_duplicate_detection.py`*\n"

        return report

    def _create_followup_tasks(self, analysis_results: dict):
        """Create follow-up tasks if duplicates found."""
        if analysis_results.get('duplicates_found', 0) > 0:
            # Only create follow-up if not auto-fixing
            if not self.auto_fix:
                self.results['followup_tasks'].append({
                    'name': 'Review and fix duplicate Todo2 tasks',
                    'description': f"Found {analysis_results.get('duplicates_found', 0)} duplicate issues. Review report and fix.",
                    'priority': 'high' if self.duplicates['duplicate_ids'] else 'medium'
                })


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description='Detect duplicate Todo2 tasks')
    parser.add_argument('--config', type=str, default='scripts/todo2_duplicate_config.json',
                       help='Path to configuration file')
    parser.add_argument('--output', type=str, help='Override output path')
    parser.add_argument('--threshold', type=float, default=0.85,
                       help='Similarity threshold (0.0-1.0, default: 0.85)')
    parser.add_argument('--auto-fix', action='store_true',
                       help='Automatically fix duplicates (experimental)')
    parser.add_argument('--dry-run', action='store_true',
                       help='Dry run mode (no changes)')

    args = parser.parse_args()

    # Load config
    from project_management_automation.utils import find_project_root
    project_root = find_project_root()
    config_path = project_root / args.config
    if config_path.exists():
        with open(config_path) as f:
            config = json.load(f)
    else:
        config = {}

    # Override with command-line args
    if args.output:
        config['output_path'] = args.output
    if args.threshold:
        config['similarity_threshold'] = args.threshold
    if args.auto_fix:
        config['auto_fix'] = True

    # Run detector
    detector = Todo2DuplicateDetector(config)
    results = detector.run()

    if results and 'error' not in results:
        # Write report to file
        report = results.get('report', '')
        if report:
            detector.output_path.parent.mkdir(parents=True, exist_ok=True)
            with open(detector.output_path, 'w') as f:
                f.write(report)
            logger.info(f"Report written to: {detector.output_path}")

        duplicates_found = results.get('duplicates_found', 0)
        print("\n✅ Duplicate detection complete!")
        print(f"   Report: {detector.output_path}")
        print(f"   Duplicates found: {duplicates_found}")
    else:
        logger.error("Duplicate detection failed")
        sys.exit(1)


if __name__ == '__main__':
    main()
