"""
Estimation Learning Module

Learns from actual vs estimated task completion times to improve future estimates.
Analyzes patterns in estimation errors and adjusts estimation strategies accordingly.
"""

import json
import logging
import statistics
from collections import defaultdict
from pathlib import Path
from typing import Any, Dict, Optional

from ..utils import find_project_root

logger = logging.getLogger(__name__)


class EstimationLearner:
    """
    Learns from estimation accuracy to improve future estimates.

    Analyzes:
    - Patterns in over/under-estimation
    - Which task types are consistently mis-estimated
    - MLX vs statistical accuracy
    - Adjustments needed for different priorities/tags
    """

    def __init__(self, project_root: Path | None = None):
        """Initialize learner with project root."""
        if project_root is None:
            project_root = find_project_root()
        self.project_root = project_root
        self.state_file = project_root / ".todo2" / "state.todo2.json"
        self.learning_cache_file = project_root / ".todo2" / "estimation_learning.json"

    def analyze_estimation_accuracy(self) -> dict[str, Any]:
        """
        Analyze estimation accuracy from historical data.

        Returns:
            Dictionary with accuracy analysis including:
            - Overall accuracy metrics
            - Error patterns by tag, priority, method
            - Recommendations for improvement
        """
        if not self.state_file.exists():
            return {
                'success': False,
                'message': 'No task data available',
                'accuracy_metrics': {}
            }

        try:
            with open(self.state_file) as f:
                data = json.load(f)

            tasks = data.get('todos', [])

            # Get completed tasks with both estimates and actuals
            completed_tasks = []
            for task in tasks:
                status = task.get('status', '').lower()
                if status not in ['done', 'completed']:
                    continue

                estimated = task.get('estimatedHours')
                actual = task.get('actualHours')

                # Calculate active work time from status changes (more accurate than elapsed time)
                if not actual:
                    actual = self._calculate_active_work_time(task)
                    if not actual or actual <= 0:
                        continue

                if estimated and estimated > 0 and actual and actual > 0:
                    completed_tasks.append({
                        'name': task.get('name', ''),
                        'details': task.get('details', '') or task.get('long_description', ''),
                        'tags': task.get('tags', []),
                        'priority': task.get('priority', 'medium'),
                        'estimated_hours': estimated,
                        'actual_hours': actual,
                        'error': actual - estimated,
                        'error_pct': ((actual - estimated) / estimated) * 100 if estimated > 0 else 0,
                        'abs_error_pct': abs((actual - estimated) / estimated) * 100 if estimated > 0 else 0,
                    })

            if not completed_tasks:
                return {
                    'success': False,
                    'message': 'No completed tasks with both estimates and actuals',
                    'completed_tasks_count': 0
                }

            # Calculate overall accuracy metrics
            errors = [t['error'] for t in completed_tasks]
            error_pcts = [t['error_pct'] for t in completed_tasks]
            abs_error_pcts = [t['abs_error_pct'] for t in completed_tasks]

            accuracy_metrics = {
                'total_tasks': len(completed_tasks),
                'mean_error': round(statistics.mean(errors), 2),
                'median_error': round(statistics.median(errors), 2),
                'mean_absolute_error': round(statistics.mean([abs(e) for e in errors]), 2),
                'mean_error_percentage': round(statistics.mean(error_pcts), 2),
                'mean_absolute_error_percentage': round(statistics.mean(abs_error_pcts), 2),
                'over_estimated_count': len([e for e in errors if e < 0]),  # Negative error = over-estimated
                'under_estimated_count': len([e for e in errors if e > 0]),  # Positive error = under-estimated
                'accurate_count': len([e for e in abs_error_pcts if e < 20]),  # Within 20%
            }

            # Analyze patterns by tag
            tag_accuracy = self._analyze_by_tag(completed_tasks)

            # Analyze patterns by priority
            priority_accuracy = self._analyze_by_priority(completed_tasks)

            # Analyze over/under estimation patterns
            patterns = self._identify_patterns(completed_tasks)

            # Generate recommendations
            recommendations = self._generate_recommendations(
                accuracy_metrics,
                tag_accuracy,
                priority_accuracy,
                patterns
            )

            return {
                'success': True,
                'accuracy_metrics': accuracy_metrics,
                'tag_accuracy': tag_accuracy,
                'priority_accuracy': priority_accuracy,
                'patterns': patterns,
                'recommendations': recommendations,
            }

        except Exception as e:
            logger.error(f"Failed to analyze estimation accuracy: {e}", exc_info=True)
            return {
                'success': False,
                'error': str(e)
            }

    def _analyze_by_tag(self, tasks: list[dict]) -> dict[str, dict[str, float]]:
        """Analyze estimation accuracy by tag."""
        tag_stats = defaultdict(lambda: {'errors': [], 'count': 0})

        for task in tasks:
            tags = task.get('tags', [])
            if not tags:
                tags = ['untagged']

            for tag in tags:
                tag_stats[tag]['errors'].append(task['error_pct'])
                tag_stats[tag]['count'] += 1

        # Calculate statistics per tag
        result = {}
        for tag, stats in tag_stats.items():
            if stats['count'] >= 2:  # Need at least 2 tasks for meaningful stats
                errors = stats['errors']
                result[tag] = {
                    'count': stats['count'],
                    'mean_error_pct': round(statistics.mean(errors), 2),
                    'mean_abs_error_pct': round(statistics.mean([abs(e) for e in errors]), 2),
                    'bias': 'over-estimate' if statistics.mean(errors) < -10 else
                           'under-estimate' if statistics.mean(errors) > 10 else 'balanced',
                }

        return result

    def _analyze_by_priority(self, tasks: list[dict]) -> dict[str, dict[str, float]]:
        """Analyze estimation accuracy by priority."""
        priority_stats = defaultdict(lambda: {'errors': [], 'count': 0})

        for task in tasks:
            priority = task.get('priority', 'medium').lower()
            priority_stats[priority]['errors'].append(task['error_pct'])
            priority_stats[priority]['count'] += 1

        # Calculate statistics per priority
        result = {}
        for priority, stats in priority_stats.items():
            if stats['count'] >= 2:
                errors = stats['errors']
                result[priority] = {
                    'count': stats['count'],
                    'mean_error_pct': round(statistics.mean(errors), 2),
                    'mean_abs_error_pct': round(statistics.mean([abs(e) for e in errors]), 2),
                    'bias': 'over-estimate' if statistics.mean(errors) < -10 else
                           'under-estimate' if statistics.mean(errors) > 10 else 'balanced',
                }

        return result

    def _identify_patterns(self, tasks: list[dict]) -> dict[str, Any]:
        """Identify patterns in estimation errors."""
        patterns = {
            'consistently_over_estimated': [],
            'consistently_under_estimated': [],
            'high_variance_tags': [],
        }

        # Find tags with consistent bias
        tag_errors = defaultdict(list)
        for task in tasks:
            for tag in task.get('tags', []):
                tag_errors[tag].append(task['error_pct'])

        for tag, errors in tag_errors.items():
            if len(errors) >= 3:  # Need multiple samples
                mean_error = statistics.mean(errors)
                stdev = statistics.stdev(errors) if len(errors) > 1 else 0

                if mean_error < -15 and stdev < 20:  # Consistently over-estimated
                    patterns['consistently_over_estimated'].append({
                        'tag': tag,
                        'mean_error_pct': round(mean_error, 2),
                        'count': len(errors),
                        'recommendation': f'Reduce estimates for {tag} tasks by {abs(round(mean_error, 0))}%'
                    })
                elif mean_error > 15 and stdev < 20:  # Consistently under-estimated
                    patterns['consistently_under_estimated'].append({
                        'tag': tag,
                        'mean_error_pct': round(mean_error, 2),
                        'count': len(errors),
                        'recommendation': f'Increase estimates for {tag} tasks by {round(mean_error, 0)}%'
                    })
                elif stdev > 30:  # High variance (unpredictable)
                    patterns['high_variance_tags'].append({
                        'tag': tag,
                        'stdev': round(stdev, 2),
                        'count': len(errors),
                        'recommendation': f'High variance in {tag} tasks - consider breaking down into smaller tasks'
                    })

        return patterns

    def _generate_recommendations(
        self,
        metrics: dict[str, Any],
        tag_accuracy: dict[str, dict],
        priority_accuracy: dict[str, dict],
        patterns: dict[str, list]
    ) -> list[str]:
        """Generate actionable recommendations based on analysis."""
        recommendations = []

        # Overall accuracy
        mae_pct = metrics.get('mean_absolute_error_percentage', 0)
        if mae_pct > 40:
            recommendations.append(
                f"⚠️ High estimation error ({mae_pct:.1f}% MAE). Consider providing more detailed task descriptions."
            )
        elif mae_pct > 25:
            recommendations.append(
                f"Estimation accuracy is moderate ({mae_pct:.1f}% MAE). Room for improvement."
            )
        else:
            recommendations.append(
                f"✅ Good estimation accuracy ({mae_pct:.1f}% MAE)."
            )

        # Bias detection
        mean_error_pct = metrics.get('mean_error_percentage', 0)
        if mean_error_pct < -15:
            recommendations.append(
                f"📉 Systematically over-estimating by {abs(mean_error_pct):.1f}%. Consider reducing estimates by ~{abs(mean_error_pct) * 0.8:.0f}%."
            )
        elif mean_error_pct > 15:
            recommendations.append(
                f"📈 Systematically under-estimating by {mean_error_pct:.1f}%. Consider increasing estimates by ~{mean_error_pct * 0.8:.0f}%."
            )

        # Tag-specific recommendations
        for tag, stats in tag_accuracy.items():
            if stats['bias'] == 'over-estimate' and stats['mean_abs_error_pct'] > 25:
                recommendations.append(
                    f"Tag '{tag}': Consistently over-estimated (avg {stats['mean_error_pct']:.1f}%). "
                    f"Reduce estimates by ~{abs(round(stats['mean_error_pct'] * 0.8, 0))}%."
                )
            elif stats['bias'] == 'under-estimate' and stats['mean_abs_error_pct'] > 25:
                recommendations.append(
                    f"Tag '{tag}': Consistently under-estimated (avg {stats['mean_error_pct']:.1f}%). "
                    f"Increase estimates by ~{round(stats['mean_error_pct'] * 0.8, 0)}%."
                )

        # Pattern-based recommendations
        for pattern in patterns.get('consistently_over_estimated', []):
            recommendations.append(pattern['recommendation'])

        for pattern in patterns.get('consistently_under_estimated', []):
            recommendations.append(pattern['recommendation'])

        return recommendations

    def get_adjustment_factors(self) -> dict[str, float]:
        """
        Get adjustment factors based on learned patterns.

        Returns dictionary mapping tags/priorities to adjustment multipliers
        that can be applied to future estimates.
        """
        analysis = self.analyze_estimation_accuracy()

        if not analysis.get('success'):
            return {}

        adjustments = {}

        # Tag adjustments
        for tag, stats in analysis.get('tag_accuracy', {}).items():
            if stats['count'] >= 3:  # Need sufficient data
                mean_error = stats.get('mean_error_pct', 0)
                if abs(mean_error) > 10:  # Significant bias
                    # Adjustment factor: if over-estimated by 20%, multiply by 0.8
                    # if under-estimated by 20%, multiply by 1.2
                    adjustment = 1.0 + (mean_error / 100.0) * 0.8  # 80% correction to avoid over-adjusting
                    adjustments[f'tag:{tag}'] = max(0.5, min(2.0, adjustment))  # Clamp to reasonable range

        # Priority adjustments
        for priority, stats in analysis.get('priority_accuracy', {}).items():
            if stats['count'] >= 3:
                mean_error = stats.get('mean_error_pct', 0)
                if abs(mean_error) > 10:
                    adjustment = 1.0 + (mean_error / 100.0) * 0.8
                    adjustments[f'priority:{priority}'] = max(0.5, min(2.0, adjustment))

        return adjustments

    def save_learning_data(self, data: dict[str, Any]) -> bool:
        """Save learned patterns to cache file."""
        try:
            self.learning_cache_file.parent.mkdir(parents=True, exist_ok=True)
            with open(self.learning_cache_file, 'w') as f:
                json.dump(data, f, indent=2)
            return True
        except Exception as e:
            logger.error(f"Failed to save learning data: {e}")
            return False

    def _calculate_active_work_time(self, task: dict) -> float | None:
        """
        Calculate active work time from status change history.

        Only counts time when task was "In Progress", not idle time.
        Tracks transitions: Todo -> In Progress -> Todo/Done

        Returns:
            Total active work hours, or None if cannot calculate
        """
        from datetime import datetime

        changes = task.get('changes', [])
        if not changes:
            return None

        def parse_datetime(dt_str):
            if not dt_str:
                return None
            try:
                return datetime.fromisoformat(dt_str.replace('Z', '+00:00'))
            except Exception:
                return None

        # Track In Progress periods
        task.get('status', 'Todo')
        created_time = parse_datetime(task.get('created'))

        # Build timeline of status changes
        status_changes = []
        for change in changes:
            if change.get('field') == 'status':
                timestamp = parse_datetime(change.get('timestamp'))
                old_value = change.get('oldValue', '').strip()
                new_value = change.get('newValue', '').strip()

                if timestamp and old_value and new_value:
                    status_changes.append({
                        'time': timestamp,
                        'from': old_value,
                        'to': new_value
                    })

        # Sort by time
        status_changes.sort(key=lambda x: x['time'])

        # Calculate total time in "In Progress" state
        total_seconds = 0
        in_progress_start = None

        # Start from creation time if available
        current_status_state = 'Todo'

        if created_time and status_changes:
            # Initialize from first status change
            first_change = status_changes[0]
            if first_change['from'] == 'Todo' and first_change['to'] == 'In Progress':
                first_change['time']
                current_status_state = 'In Progress'
                in_progress_start = first_change['time']

        # Process all status changes
        for change in status_changes:
            change_time = change['time']
            change['from']
            to_status = change['to']

            # If we were In Progress, accumulate time until this change
            if current_status_state == 'In Progress' and in_progress_start:
                elapsed = (change_time - in_progress_start).total_seconds()
                if elapsed > 0:
                    total_seconds += elapsed

            # Update current state
            current_status_state = to_status

            # Track when entering In Progress
            if to_status == 'In Progress':
                in_progress_start = change_time
            else:
                in_progress_start = None

        # Handle final state if still In Progress (shouldn't happen for completed tasks)
        # or if task completed while In Progress
        if current_status_state == 'In Progress' and in_progress_start:
            completed_time = parse_datetime(task.get('completedAt')) or parse_datetime(task.get('lastModified'))
            if completed_time and completed_time > in_progress_start:
                elapsed = (completed_time - in_progress_start).total_seconds()
                if elapsed > 0:
                    total_seconds += elapsed

        # Convert to hours
        if total_seconds > 0:
            hours = total_seconds / 3600.0
            # Cap at reasonable maximum (e.g., 100 hours per task)
            return min(hours, 100.0)

        return None

    def load_learning_data(self) -> Optional[Dict[str, Any]]:
        """Load learned patterns from cache file."""
        try:
            if self.learning_cache_file.exists():
                with open(self.learning_cache_file) as f:
                    return json.load(f)
        except Exception as e:
            logger.debug(f"Failed to load learning data: {e}")
        return None

