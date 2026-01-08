"""
MLX-Enhanced Task Duration Estimator

Combines statistical methods with MLX semantic understanding for improved
task duration estimation accuracy.

Features:
- Semantic task similarity matching (vs simple word overlap)
- MLX-based complexity analysis
- Hybrid approach: statistical + MLX (weighted combination)
- Graceful fallback to statistical-only if MLX unavailable
"""

import json
import logging
import re
from typing import Any, Dict, List, Optional

from .task_duration_estimator import TaskDurationEstimator

logger = logging.getLogger(__name__)

# Try to import MLX integration
try:
    from .mlx_integration import MLX_AVAILABLE, generate_with_mlx
    MLX_ENABLED = MLX_AVAILABLE
except ImportError:
    MLX_ENABLED = False
    logger.debug("MLX integration not available")


class MLXEnhancedTaskEstimator(TaskDurationEstimator):
    """
    Enhanced task duration estimator using MLX for semantic understanding.

    Combines statistical methods (from parent class) with MLX semantic analysis
    for improved estimation accuracy, especially for novel tasks or tasks with
    semantic similarity but low word overlap.

    Now includes adaptive learning from actual vs estimated completion times.
    """

    def __init__(
        self,
        project_root=None,
        use_mlx: bool = True,
        mlx_weight: float = 0.3,
        mlx_model: str = "mlx-community/Phi-3.5-mini-instruct-4bit",
        use_learning: bool = True
    ):
        """
        Initialize MLX-enhanced estimator.

        Args:
            project_root: Project root path (defaults to find_project_root)
            use_mlx: Enable MLX enhancement (default: True)
            mlx_weight: Weight for MLX estimate in hybrid (0.0-1.0, default: 0.3)
                       1.0 - mlx_weight = weight for statistical estimate
            mlx_model: MLX model to use for analysis
            use_learning: Enable adaptive learning from past estimates (default: True)
        """
        super().__init__(project_root)
        self.use_mlx = use_mlx and MLX_ENABLED
        self.mlx_weight = max(0.0, min(1.0, mlx_weight))  # Clamp to [0, 1]
        self.statistical_weight = 1.0 - self.mlx_weight
        self.mlx_model = mlx_model
        self.use_learning = use_learning

        # Initialize learner if enabled
        self.learner = None
        self.adjustment_factors = {}
        if self.use_learning:
            try:
                from .estimation_learner import EstimationLearner
                self.learner = EstimationLearner(project_root)
                self.adjustment_factors = self.learner.get_adjustment_factors()
                if self.adjustment_factors:
                    logger.debug(f"Loaded {len(self.adjustment_factors)} adjustment factors from learning")
            except Exception as e:
                logger.debug(f"Learning not available: {e}")
                self.use_learning = False

        if self.use_mlx:
            logger.debug(f"MLX enhancement enabled (weight: {self.mlx_weight})")
        else:
            logger.debug("MLX enhancement disabled or unavailable")

    def estimate(
        self,
        name: str,
        details: str = "",
        tags: Optional[List[str]] = None,
        priority: str = "medium",
        use_historical: bool = True,
    ) -> dict[str, Any]:
        """
        Estimate task duration with MLX semantic enhancement.

        Uses hybrid approach:
        - Statistical estimate (from parent class): Historical data + keyword matching
        - MLX estimate: Semantic understanding + complexity analysis
        - Combined: Weighted average of both

        Falls back gracefully to statistical-only if MLX unavailable or fails.

        Returns:
            Dictionary with estimate, confidence, method, and metadata
        """
        # Get base statistical estimate
        base_estimate = super().estimate(name, details, tags, priority, use_historical)

        # If MLX disabled or unavailable, return statistical estimate
        if not self.use_mlx:
            return base_estimate

        # Get MLX-enhanced estimate
        mlx_estimate = None
        try:
            mlx_estimate = self._mlx_semantic_estimate(name, details, tags or [], priority)
        except Exception as e:
            logger.debug(f"MLX estimation failed, using statistical only: {e}")

        # Combine estimates if MLX estimate available
        if mlx_estimate and mlx_estimate.get('estimate_hours'):
            combined = self._combine_estimates(base_estimate, mlx_estimate)
        else:
            combined = base_estimate

        # Apply learned adjustments if learning enabled
        if self.use_learning and self.adjustment_factors:
            combined = self._apply_learned_adjustments(combined, tags or [], priority)

        return combined

    def _mlx_semantic_estimate(
        self,
        name: str,
        details: str,
        tags: list[str],
        priority: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get MLX-based semantic estimate.

        Uses MLX to:
        1. Analyze task complexity and scope
        2. Understand semantic meaning (vs keyword matching)
        3. Estimate duration based on semantic understanding
        """
        # Build analysis prompt
        tags_str = ", ".join(tags) if tags else "none"
        analysis_prompt = f"""Analyze this task and estimate how long it will take to complete.

Task: {name}
Details: {details}
Tags: {tags_str}
Priority: {priority}

Consider:
- Technical complexity (simple, moderate, complex)
- Scope of work (small fix, feature, major refactor)
- Research/testing needs (none, some, extensive)
- Integration complexity (standalone, single system, multi-system)

Respond ONLY with a JSON object in this exact format:
{{
    "estimate_hours": <number between 0.5 and 20>,
    "confidence": <number between 0.0 and 1.0>,
    "complexity": <number between 1 and 10>,
    "reasoning": "<brief 1-2 sentence explanation>"
}}

Example response:
{{
    "estimate_hours": 3.5,
    "confidence": 0.75,
    "complexity": 6,
    "reasoning": "Moderate complexity feature requiring API integration and testing"
}}
"""

        try:
            # Call MLX generation
            result = generate_with_mlx(
                prompt=analysis_prompt,
                model=self.mlx_model,
                max_tokens=200,
                verbose=False
            )

            # Parse MLX response
            data = json.loads(result)
            if data.get('success') and 'data' in data:
                generated_text = data['data'].get('generated_text', '')
                parsed = self._parse_mlx_response(generated_text)
                if parsed:
                    logger.debug(f"MLX estimate: {parsed.get('estimate_hours')}h (confidence: {parsed.get('confidence')})")
                    return parsed

            logger.debug("MLX returned success but no parseable estimate")

        except json.JSONDecodeError as e:
            logger.debug(f"Failed to parse MLX JSON response: {e}")
        except Exception as e:
            logger.debug(f"MLX estimation error: {e}")

        return None

    def _parse_mlx_response(self, text: str) -> Optional[Dict[str, Any]]:
        """
        Parse MLX response to extract estimation JSON.

        MLX may return:
        - Raw JSON
        - JSON embedded in text
        - JSON with markdown formatting
        """
        # Try to extract JSON object from response
        # Look for JSON object pattern
        json_patterns = [
            r'\{[^{}]*"estimate_hours"[^{}]*\}',  # Single-line JSON
            r'\{.*?"estimate_hours".*?\}',  # Multi-line JSON
        ]

        for pattern in json_patterns:
            match = re.search(pattern, text, re.DOTALL)
            if match:
                try:
                    json_str = match.group(0)
                    # Clean up common formatting issues
                    json_str = json_str.replace('\n', ' ').strip()
                    parsed = json.loads(json_str)

                    # Validate required fields
                    if 'estimate_hours' in parsed:
                        # Ensure valid ranges
                        estimate_hours = float(parsed['estimate_hours'])
                        estimate_hours = max(0.5, min(20.0, estimate_hours))  # Clamp to reasonable range

                        confidence = float(parsed.get('confidence', 0.6))
                        confidence = max(0.0, min(1.0, confidence))

                        complexity = int(parsed.get('complexity', 5))
                        complexity = max(1, min(10, complexity))

                        return {
                            'estimate_hours': round(estimate_hours, 1),
                            'confidence': confidence,
                            'complexity': complexity,
                            'reasoning': parsed.get('reasoning', 'MLX semantic analysis'),
                            'method': 'mlx_semantic'
                        }
                except (json.JSONDecodeError, ValueError, KeyError) as e:
                    logger.debug(f"Failed to parse JSON from pattern match: {e}")
                    continue

        # Fallback: Try to extract just the estimate_hours number
        hours_match = re.search(r'"estimate_hours"\s*:\s*(\d+\.?\d*)', text)
        if hours_match:
            try:
                estimate_hours = float(hours_match.group(1))
                estimate_hours = max(0.5, min(20.0, estimate_hours))
                return {
                    'estimate_hours': round(estimate_hours, 1),
                    'confidence': 0.5,  # Lower confidence if we only got hours
                    'reasoning': 'Extracted estimate from MLX response',
                    'method': 'mlx_extracted'
                }
            except ValueError:
                pass

        return None

    def _combine_estimates(
        self,
        statistical_estimate: dict[str, Any],
        mlx_estimate: dict[str, Any]
    ) -> dict[str, Any]:
        """
        Combine statistical and MLX estimates using weighted average.

        Args:
            statistical_estimate: Estimate from parent class
            mlx_estimate: Estimate from MLX semantic analysis

        Returns:
            Combined estimate dictionary
        """
        stat_hours = statistical_estimate['estimate_hours']
        mlx_hours = mlx_estimate['estimate_hours']

        # Weighted average
        combined_hours = (
            stat_hours * self.statistical_weight +
            mlx_hours * self.mlx_weight
        )
        combined_hours = round(combined_hours, 1)

        # Combine confidence (take higher, with slight boost for hybrid)
        stat_confidence = statistical_estimate.get('confidence', 0.5)
        mlx_confidence = mlx_estimate.get('confidence', 0.5)
        combined_confidence = max(stat_confidence, mlx_confidence)

        # Boost confidence slightly for hybrid approach (we have more information)
        combined_confidence = min(0.95, combined_confidence + 0.05)

        # Calculate bounds (use statistical bounds as base, adjusted by MLX)
        lower_bound = min(
            statistical_estimate.get('lower_bound', combined_hours * 0.7),
            mlx_hours * 0.8
        )
        upper_bound = max(
            statistical_estimate.get('upper_bound', combined_hours * 1.3),
            mlx_hours * 1.2
        )

        return {
            'estimate_hours': combined_hours,
            'confidence': round(combined_confidence, 2),
            'method': 'hybrid_statistical_mlx',
            'lower_bound': round(lower_bound, 1),
            'upper_bound': round(upper_bound, 1),
            'metadata': {
                **statistical_estimate.get('metadata', {}),
                'statistical_estimate': stat_hours,
                'mlx_estimate': mlx_hours,
                'mlx_confidence': mlx_confidence,
                'mlx_complexity': mlx_estimate.get('complexity'),
                'mlx_reasoning': mlx_estimate.get('reasoning'),
                'statistical_weight': self.statistical_weight,
                'mlx_weight': self.mlx_weight,
                'enhancement': 'mlx_semantic_matching'
            }
        }

    def _apply_learned_adjustments(
        self,
        estimate: dict[str, Any],
        tags: list[str],
        priority: str
    ) -> dict[str, Any]:
        """
        Apply learned adjustment factors based on past estimation accuracy.

        Adjusts estimate based on patterns learned from actual vs estimated times.
        """
        if not self.adjustment_factors:
            return estimate

        adjustments = []

        # Apply tag-based adjustments
        for tag in tags:
            key = f'tag:{tag}'
            if key in self.adjustment_factors:
                adjustments.append(self.adjustment_factors[key])

        # Apply priority-based adjustments
        priority_key = f'priority:{priority.lower()}'
        if priority_key in self.adjustment_factors:
            adjustments.append(self.adjustment_factors[priority_key])

        # Apply adjustments (average if multiple)
        if adjustments:
            import statistics
            avg_adjustment = statistics.mean(adjustments)
            adjusted_hours = estimate['estimate_hours'] * avg_adjustment

            estimate['estimate_hours'] = round(adjusted_hours, 1)
            estimate['method'] = f"{estimate.get('method', 'unknown')}_with_learning"

            if 'metadata' not in estimate:
                estimate['metadata'] = {}
            estimate['metadata']['learning_adjustment'] = round(avg_adjustment, 3)
            estimate['metadata']['adjustment_count'] = len(adjustments)

            # Recalculate bounds with adjustment
            estimate['lower_bound'] = round(estimate.get('lower_bound', adjusted_hours * 0.7) * avg_adjustment, 1)
            estimate['upper_bound'] = round(estimate.get('upper_bound', adjusted_hours * 1.3) * avg_adjustment, 1)

            logger.debug(f"Applied learning adjustment: {avg_adjustment:.3f}x to {estimate['estimate_hours']}h")

        return estimate


# Convenience functions for backward compatibility
def estimate_task_duration_mlx_enhanced(
    name: str,
    details: str = "",
    tags: Optional[List[str]] = None,
    priority: str = "medium",
    use_historical: bool = True,
    use_mlx: bool = True,
    mlx_weight: float = 0.3,
    project_root=None
) -> float:
    """
    Convenience function for MLX-enhanced duration estimation.

    Returns just the estimated hours (backward compatible).
    """
    estimator = MLXEnhancedTaskEstimator(
        project_root=project_root,
        use_mlx=use_mlx,
        mlx_weight=mlx_weight
    )
    result = estimator.estimate(name, details, tags, priority, use_historical)
    return result['estimate_hours']


def estimate_task_duration_mlx_enhanced_detailed(
    name: str,
    details: str = "",
    tags: Optional[List[str]] = None,
    priority: str = "medium",
    use_historical: bool = True,
    use_mlx: bool = True,
    mlx_weight: float = 0.3,
    project_root=None
) -> dict[str, Any]:
    """
    Get detailed MLX-enhanced duration estimation with confidence and metadata.

    Returns full estimation dictionary.
    """
    estimator = MLXEnhancedTaskEstimator(
        project_root=project_root,
        use_mlx=use_mlx,
        mlx_weight=mlx_weight
    )
    return estimator.estimate(name, details, tags, priority, use_historical)

