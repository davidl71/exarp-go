"""
Task Duration Estimation using Statistical Methods

Uses Python's standard library `statistics` module to improve task duration estimates
by learning from historical task completion data.

Features:
- Historical data analysis (learns from completed tasks)
- Statistical methods (mean, median, percentiles, confidence intervals)
- Multi-factor matching (tags, keywords, priority, complexity)
- Confidence scores and uncertainty ranges
- Multiple estimation strategies with weighted combination
"""

import json
import logging
import sys
from datetime import datetime
from pathlib import Path
from typing import Any, List, Optional

# Use Go statistics via bridge wrapper
# Add bridge directory to path (bridge is at project_root/bridge/)
from ..utils import find_project_root
project_root = find_project_root()
bridge_path = Path(project_root) / "bridge"
if str(bridge_path) not in sys.path:
    sys.path.insert(0, str(bridge_path))
from statistics_bridge import mean, median, stdev, quantile, round as round_stat, StatisticsError  # noqa: E402

logger = logging.getLogger(__name__)


class TaskDurationEstimator:
    """Statistical task duration estimator."""

    def __init__(self, project_root: Optional[Path] = None):
        """Initialize estimator with project root."""
        if project_root is None:
            project_root = find_project_root()
        self.project_root = project_root
        self.state_file = project_root / ".todo2" / "state.todo2.json"
        self._historical_data: Optional[list[dict]] = None

    def load_historical_data(self) -> list[dict]:
        """Load and process historical task data."""
        if self._historical_data is not None:
            return self._historical_data

        if not self.state_file.exists():
            logger.warning(f"State file not found: {self.state_file}")
            return []

        try:
            with open(self.state_file) as f:
                data = json.load(f)

            todos = data.get('todos', [])
            historical = []

            for task in todos:
                # Only use completed tasks with time data
                status = task.get('status', '').lower()
                if status not in ['done', 'completed']:
                    continue

                # Get actual hours if available
                actual_hours = task.get('actualHours')
                if actual_hours and actual_hours > 0:
                    historical.append({
                        'name': task.get('name', '') or task.get('content', ''),
                        'details': task.get('details', '') or task.get('long_description', ''),
                        'tags': task.get('tags', []),
                        'priority': task.get('priority', 'medium'),
                        'estimated_hours': task.get('estimatedHours', 0),
                        'actual_hours': actual_hours,
                        'created': task.get('created'),
                        'completed_at': task.get('completedAt'),
                    })
                    continue

                # Calculate active work time using EstimationLearner (excludes idle time, uses comment fallback)
                try:
                    from .estimation_learner import EstimationLearner
                    learner = EstimationLearner(self.project_root)
                    
                    # Load comments for comment-based fallback
                    comments = []
                    if 'comments' in data:
                        comments = [c for c in data.get('comments', []) if c.get('todoId') == task.get('id')]
                    
                    calculated_hours, method = learner._calculate_active_work_time(task, comments)
                    
                    if calculated_hours and calculated_hours > 0:
                        historical.append({
                            'name': task.get('name', '') or task.get('content', ''),
                            'details': task.get('details', '') or task.get('long_description', ''),
                            'tags': task.get('tags', []),
                            'priority': task.get('priority', 'medium'),
                            'estimated_hours': task.get('estimatedHours', 0),
                            'actual_hours': calculated_hours,
                            'created': task.get('created'),
                            'completed_at': task.get('completedAt'),
                        })
                        continue
                except Exception as e:
                    logger.debug(f"Could not calculate active work time for task {task.get('id')}: {e}")

                # Last resort: Calculate duration from timestamps if available (includes idle time - less accurate)
                created_str = task.get('created')
                completed_str = task.get('completedAt') or task.get('lastModified')

                if created_str and completed_str:
                    try:
                        created = self._parse_datetime(created_str)
                        completed = self._parse_datetime(completed_str)

                        if created and completed and completed > created:
                            # Calculate hours (minimum 0.5 hours)
                            # Note: This includes idle time, so it's less accurate than status/comment-based methods
                            duration_hours = max(0.5, (completed - created).total_seconds() / 3600)
                            historical.append({
                                'name': task.get('name', '') or task.get('content', ''),
                                'details': task.get('details', '') or task.get('long_description', ''),
                                'tags': task.get('tags', []),
                                'priority': task.get('priority', 'medium'),
                                'estimated_hours': task.get('estimatedHours', 0),
                                'actual_hours': duration_hours,
                                'created': created_str,
                                'completed_at': completed_str,
                            })
                    except (ValueError, TypeError) as e:
                        logger.debug(f"Failed to parse timestamps for task: {e}")
                        continue

            self._historical_data = historical
            logger.info(f"Loaded {len(historical)} historical task records")
            return historical

        except Exception as e:
            logger.error(f"Failed to load historical data: {e}", exc_info=True)
            return []

    def _parse_datetime(self, dt_str: str) -> Optional[datetime]:
        """Parse datetime string to datetime object."""
        if not dt_str:
            return None

        # Try common formats
        formats = [
            '%Y-%m-%dT%H:%M:%S.%f',
            '%Y-%m-%dT%H:%M:%S',
            '%Y-%m-%dT%H:%M:%S.%f%z',
            '%Y-%m-%dT%H:%M:%S%z',
            '%Y-%m-%d %H:%M:%S',
        ]

        for fmt in formats:
            try:
                if fmt.endswith('z') or '+00:00' in dt_str:
                    return datetime.strptime(dt_str.replace('Z', '+00:00'), fmt.replace('z', '%z'))
                return datetime.strptime(dt_str, fmt)
            except ValueError:
                continue

        return None

    def estimate(
        self,
        name: str,
        details: str = "",
        tags: Optional[List[str]] = None,
        priority: str = "medium",
        use_historical: bool = True,
    ) -> dict[str, Any]:
        """
        Estimate task duration using multiple statistical methods.

        Returns:
            Dictionary with estimate, confidence, method, and metadata
        """
        if tags is None:
            tags = []

        text = (name + " " + details).lower()

        # Strategy 1: Historical data matching (if available)
        historical_estimate = None
        historical_confidence = 0.0
        if use_historical:
            historical_data = self.load_historical_data()
            if historical_data:
                historical_estimate, historical_confidence = self._estimate_from_history(
                    text, tags, priority, historical_data
                )

        # Strategy 2: Keyword-based heuristic (fallback)
        heuristic_estimate = self._estimate_from_keywords(text)
        heuristic_confidence = 0.3  # Lower confidence for heuristics

        # Strategy 3: Priority-based adjustment
        priority_multiplier = self._get_priority_multiplier(priority)

        # Combine estimates with weighted average
        if historical_estimate and historical_confidence > 0.2:
            # Use historical if we have good match
            base_estimate = historical_estimate
            confidence = historical_confidence
            method = "historical_match"
        else:
            # Use heuristic with priority adjustment
            base_estimate = heuristic_estimate * priority_multiplier
            confidence = heuristic_confidence
            method = "keyword_heuristic"

        # Apply priority multiplier
        final_estimate = base_estimate * priority_multiplier

        # Round to reasonable precision
        final_estimate = round(final_estimate, 1)

        # Calculate confidence interval (assuming normal distribution)
        std_dev = final_estimate * 0.3  # 30% coefficient of variation
        lower_bound = max(0.5, final_estimate - 1.96 * std_dev)  # 95% CI
        upper_bound = final_estimate + 1.96 * std_dev

        return {
            'estimate_hours': final_estimate,
            'confidence': min(0.95, confidence),  # Cap at 95%
            'method': method,
            'lower_bound': round(lower_bound, 1),
            'upper_bound': round(upper_bound, 1),
            'metadata': {
                'historical_match': historical_estimate is not None,
                'historical_confidence': historical_confidence,
                'heuristic_estimate': heuristic_estimate,
                'priority_multiplier': priority_multiplier,
            }
        }

    def _estimate_from_history(
        self,
        text: str,
        tags: list[str],
        priority: str,
        historical: list[dict]
    ) -> tuple[float, float]:
        """Estimate using historical data matching."""
        matches = []
        text_words = set(text.split())

        for record in historical:
            score = 0.0
            record_text = (record['name'] + " " + record['details']).lower()
            record_words = set(record_text.split())

            # Text similarity (word overlap)
            if text_words and record_words:
                word_overlap = len(text_words & record_words) / len(text_words | record_words)
                score += word_overlap * 0.5

            # Tag matching
            record_tags = [t.lower() for t in record.get('tags', [])]
            if tags and record_tags:
                tag_overlap = len({t.lower() for t in tags} & set(record_tags))
                if record_tags:
                    score += (tag_overlap / len(record_tags)) * 0.3

            # Priority matching
            if priority.lower() == record.get('priority', 'medium').lower():
                score += 0.2

            if score > 0.1:  # Minimum threshold
                matches.append({
                    'actual_hours': record['actual_hours'],
                    'score': score,
                })

        if not matches:
            return None, 0.0

        # Sort by score and take top matches
        matches.sort(key=lambda x: x['score'], reverse=True)
        top_matches = matches[:min(10, len(matches))]

        # Weighted average of top matches
        total_weight = sum(m['score'] for m in top_matches)
        if total_weight == 0:
            return None, 0.0

        weighted_sum = sum(m['actual_hours'] * m['score'] for m in top_matches)
        estimate = weighted_sum / total_weight

        # Confidence based on number and quality of matches
        avg_score = total_weight / len(top_matches)
        confidence = min(0.9, 0.3 + avg_score * 0.6)

        return estimate, confidence

    def _estimate_from_keywords(self, text: str) -> float:
        """Estimate using keyword heuristics (improved version)."""
        # Quick/simple tasks
        quick_keywords = ['quick', 'simple', 'minor', 'small', 'fix typo', 'update version', 'bump']
        if any(kw in text for kw in quick_keywords):
            return 0.5

        # Small tasks
        small_keywords = ['add', 'create', 'implement', 'setup', 'install', 'configure']
        if any(kw in text for kw in small_keywords):
            return 2.0

        # Medium tasks
        medium_keywords = ['refactor', 'migrate', 'integrate', 'update', 'improve', 'enhance']
        if any(kw in text for kw in medium_keywords):
            return 3.0

        # Large tasks
        large_keywords = ['complex', 'major', 'rewrite', 'redesign', 'architecture', 'system']
        if any(kw in text for kw in large_keywords):
            return 4.0

        # Default
        return 2.0

    def _get_priority_multiplier(self, priority: str) -> float:
        """Get time multiplier based on priority."""
        priority = priority.lower()
        multipliers = {
            'low': 0.8,      # Low priority tasks often take less effort
            'medium': 1.0,   # Baseline
            'high': 1.2,     # High priority tasks often more complex/urgent
            'critical': 1.5, # Critical tasks may have hidden complexity
        }
        return multipliers.get(priority, 1.0)

    def get_statistics(self) -> str:
        """Get statistics about historical task durations. Returns JSON string."""
        import json
        historical = self.load_historical_data()

        if not historical:
            return json.dumps({
                'count': 0,
                'message': 'No historical data available'
            }, indent=2)

        actual_hours = [r['actual_hours'] for r in historical]

        try:
            stats = {
                'count': len(actual_hours),
                'mean': round_stat(mean(actual_hours), 2),
                'median': round_stat(median(actual_hours), 2),
                'stdev': round_stat(stdev(actual_hours), 2),
                'min': round_stat(min(actual_hours), 2),
                'max': round_stat(max(actual_hours), 2),
            }

            # Percentiles using quantile (more accurate than manual calculation)
            if len(actual_hours) > 1:
                stats['p25'] = round_stat(quantile(actual_hours, 0.25), 2)
                stats['p75'] = round_stat(quantile(actual_hours, 0.75), 2)
                stats['p90'] = round_stat(quantile(actual_hours, 0.90), 2)

            # Accuracy metrics (if estimated hours available)
            with_estimates = [
                (r['estimated_hours'], r['actual_hours'])
                for r in historical
                if r.get('estimated_hours', 0) > 0
            ]

            if with_estimates:
                errors = [abs(est - act) / act for est, act in with_estimates]
                stats['estimation_accuracy'] = {
                    'count': len(with_estimates),
                    'mean_absolute_error': round_stat(mean(errors), 2),
                    'mean_error': round_stat(mean([est - act for est, act in with_estimates]), 2),
                }

            return json.dumps(stats, indent=2)

        except (StatisticsError, RuntimeError, ValueError) as e:
            import json
            logger.warning(f"Statistics calculation failed: {e}")
            return json.dumps({
                'count': len(actual_hours),
                'error': str(e)
            }, indent=2)


def estimate_task_duration(
    name: str,
    details: str = "",
    tags: Optional[List[str]] = None,
    priority: str = "medium",
    use_historical: bool = True,
    project_root: Optional[Path] = None
) -> float:
    """
    Convenience function for simple duration estimation.

    Returns just the estimated hours (backward compatible).
    """
    estimator = TaskDurationEstimator(project_root)
    result = estimator.estimate(name, details, tags, priority, use_historical)
    return result['estimate_hours']


def estimate_task_duration_detailed(
    name: str,
    details: str = "",
    tags: Optional[List[str]] = None,
    priority: str = "medium",
    use_historical: bool = True,
    project_root: Optional[Path] = None
) -> dict[str, Any]:
    """
    Get detailed duration estimation with confidence and metadata.

    Returns full estimation dictionary.
    """
    estimator = TaskDurationEstimator(project_root)
    return estimator.estimate(name, details, tags, priority, use_historical)
