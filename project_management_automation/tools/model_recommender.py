"""
AI Model Selection Recommender Tool

Recommends optimal AI model based on task type and complexity.
Based on Cursor IDE Best Practice #4.
"""

from typing import Optional
import json
import logging
import time

logger = logging.getLogger(__name__)

# Import error handler
try:
    from ..error_handler import (
        ErrorCode,
        format_error_response,
        format_success_response,
        log_automation_execution,
    )
except ImportError:

    def format_success_response(data, message=None):
        return {"success": True, "data": data, "timestamp": time.time()}

    def format_error_response(error, error_code, include_traceback=False):
        return {"success": False, "error": {"code": str(error_code), "message": str(error)}}

    def log_automation_execution(name, duration, success, error=None):
        logger.info(f"{name}: {duration:.2f}s, success={success}")

    class ErrorCode:
        AUTOMATION_ERROR = "AUTOMATION_ERROR"


# Model recommendations based on task type
MODEL_RECOMMENDATIONS = {
    "claude-sonnet": {
        "name": "Claude Sonnet 4",
        "best_for": [
            "Complex multi-file implementations",
            "Architecture decisions",
            "Code review with nuanced feedback",
            "Long context comprehension",
            "Reasoning-heavy tasks",
        ],
        "keywords": [
            "architecture", "design", "complex", "multi-file", "refactor",
            "analyze", "explain", "reason", "understand", "comprehensive",
        ],
        "task_types": ["architecture", "review", "analysis", "complex_implementation"],
        "cost": "higher",
        "speed": "moderate",
    },
    "claude-haiku": {
        "name": "Claude Haiku",
        "best_for": [
            "Quick code completions",
            "Simple bug fixes",
            "Syntax corrections",
            "Fast iterations",
            "Cost-sensitive workflows",
        ],
        "keywords": [
            "quick", "simple", "fix", "typo", "syntax", "format",
            "small", "minor", "fast", "autocomplete",
        ],
        "task_types": ["quick_fix", "completion", "formatting", "simple_edit"],
        "cost": "low",
        "speed": "fast",
    },
    "gpt-4o": {
        "name": "GPT-4o",
        "best_for": [
            "General coding tasks",
            "API integrations",
            "Multi-modal tasks with images",
            "Quick prototyping",
        ],
        "keywords": [
            "api", "integration", "prototype", "general", "standard",
            "image", "vision", "screenshot",
        ],
        "task_types": ["general", "integration", "prototyping", "multimodal"],
        "cost": "moderate",
        "speed": "fast",
    },
    "o1-preview": {
        "name": "o1-preview",
        "best_for": [
            "Mathematical reasoning",
            "Algorithm design",
            "Complex problem solving",
            "Scientific computing",
        ],
        "keywords": [
            "math", "algorithm", "proof", "reasoning", "scientific",
            "calculation", "logic", "theorem",
        ],
        "task_types": ["math", "algorithm", "reasoning", "scientific"],
        "cost": "highest",
        "speed": "slow",
    },
    "gemini-pro": {
        "name": "Gemini Pro",
        "best_for": [
            "Large codebase analysis",
            "Very long context windows",
            "Cross-file understanding",
        ],
        "keywords": [
            "large", "codebase", "context", "many files", "whole project",
        ],
        "task_types": ["large_context", "codebase_analysis"],
        "cost": "moderate",
        "speed": "moderate",
    },
    "ollama-llama3.2": {
        "name": "Ollama Llama 3.2",
        "best_for": [
            "Local development without API costs",
            "Privacy-sensitive tasks",
            "Offline development",
            "General coding tasks",
            "Quick prototyping",
        ],
        "keywords": [
            "local", "offline", "private", "free", "general", "prototype",
            "development", "no-cost", "self-hosted",
        ],
        "task_types": ["general", "local", "privacy", "offline"],
        "cost": "free",
        "speed": "moderate",
    },
    "ollama-mistral": {
        "name": "Ollama Mistral",
        "best_for": [
            "Fast local inference",
            "Code generation",
            "Quick iterations",
            "Cost-free development",
        ],
        "keywords": [
            "fast", "local", "code", "generation", "quick", "free",
            "efficient", "lightweight",
        ],
        "task_types": ["code_generation", "quick_fix", "local"],
        "cost": "free",
        "speed": "fast",
    },
    "ollama-codellama": {
        "name": "Ollama CodeLlama",
        "best_for": [
            "Code-specific tasks",
            "Code completion",
            "Code explanation",
            "Local code analysis",
        ],
        "keywords": [
            "code", "programming", "completion", "explain", "analyze",
            "local", "codebase", "syntax",
        ],
        "task_types": ["code_analysis", "code_generation", "local"],
        "cost": "free",
        "speed": "moderate",
    },
    "ollama-phi3": {
        "name": "Ollama Phi-3",
        "best_for": [
            "Small model for quick tasks",
            "Resource-constrained environments",
            "Fast local inference",
            "Simple code edits",
        ],
        "keywords": [
            "small", "lightweight", "fast", "quick", "simple",
            "resource-efficient", "mobile",
        ],
        "task_types": ["quick_fix", "simple_edit", "local"],
        "cost": "free",
        "speed": "fast",
    },
}


def recommend_model(
    task_description: Optional[str] = None,
    task_type: Optional[str] = None,
    optimize_for: str = "quality",
    include_alternatives: bool = True,
) -> str:
    """
    [HINT: Model selection. Recommends optimal AI model based on task type.]

    📊 Output: Recommended model, confidence, alternatives
    🔧 Side Effects: None (advisory only)
    📁 Analyzes: Task description, keywords, task type
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "What model should I use for implementing a complex authentication system?"

    Model Guidelines:
    - Claude Sonnet: Complex tasks, architecture, code review
    - Claude Haiku: Quick fixes, completions, simple edits
    - GPT-4o: General tasks, API work, multimodal
    - o1-preview: Math, algorithms, complex reasoning
    - Gemini Pro: Large codebase analysis
    - Ollama models: Local/offline tasks, privacy-sensitive, cost-free (llama3.2, mistral, codellama, phi3)

    Args:
        task_description: Description of the task
        task_type: Optional explicit task type
        optimize_for: "quality", "speed", or "cost"
        include_alternatives: Include alternative recommendations

    Returns:
        JSON with model recommendation
    """
    start_time = time.time()

    try:
        content = (task_description or "").lower()

        # Score each model
        model_scores = {}
        for model_id, model_data in MODEL_RECOMMENDATIONS.items():
            score = 0

            # Keyword matching
            for kw in model_data["keywords"]:
                if kw in content:
                    score += 2

            # Task type matching
            if task_type and task_type in model_data["task_types"]:
                score += 5

            # Optimization preference
            if optimize_for == "speed" and model_data["speed"] == "fast":
                score += 3
            elif optimize_for == "cost" and model_data["cost"] == "low":
                score += 3
            elif optimize_for == "quality" and model_data["cost"] in ["higher", "highest"]:
                score += 2

            model_scores[model_id] = score

        # Get best model
        best_model = max(model_scores.keys(), key=lambda x: model_scores[x])
        best_score = model_scores[best_model]
        best_data = MODEL_RECOMMENDATIONS[best_model]

        # Calculate confidence
        total_score = sum(model_scores.values()) or 1
        confidence = min(best_score / total_score * 100 + 30, 95)  # Base 30% confidence

        result = {
            "recommended_model": best_data["name"],
            "model_id": best_model,
            "confidence": round(confidence, 1),
            "best_for": best_data["best_for"][:3],
            "cost": best_data["cost"],
            "speed": best_data["speed"],
            "optimization": optimize_for,
        }

        if include_alternatives:
            # Get alternatives sorted by score
            sorted_models = sorted(
                model_scores.items(), key=lambda x: x[1], reverse=True
            )
            result["alternatives"] = [
                {
                    "model": MODEL_RECOMMENDATIONS[m]["name"],
                    "model_id": m,
                    "score": s,
                    "cost": MODEL_RECOMMENDATIONS[m]["cost"],
                }
                for m, s in sorted_models[1:3]  # Top 2 alternatives
            ]

        # Add tip based on task
        if "local" in content or "offline" in content or "private" in content or "free" in content:
            result["tip"] = "For local/offline tasks, consider Ollama models (llama3.2, mistral, codellama)"
        elif "complex" in content or "architecture" in content:
            result["tip"] = "For complex tasks, Claude Sonnet provides the best reasoning"
        elif "quick" in content or "simple" in content:
            result["tip"] = "Claude Haiku is 5x faster for simple tasks"
        elif "image" in content or "screenshot" in content:
            result["tip"] = "GPT-4o supports multimodal input for images"

        duration = time.time() - start_time
        log_automation_execution("recommend_model", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("recommend_model", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)


def list_available_models() -> str:
    """
    [HINT: Model catalog. Lists all available models with capabilities.]

    📊 Output: Model catalog with best-for, cost, speed info
    🔧 Side Effects: None
    📁 Static data
    ⏱️ Typical Runtime: <1 second

    Example Prompt:
    "What AI models are available and what are they good for?"

    Returns:
        JSON with model catalog
    """
    start_time = time.time()

    try:
        catalog = []
        for model_id, model_data in MODEL_RECOMMENDATIONS.items():
            catalog.append({
                "model_id": model_id,
                "name": model_data["name"],
                "best_for": model_data["best_for"],
                "task_types": model_data["task_types"],
                "cost": model_data["cost"],
                "speed": model_data["speed"],
            })

        result = {
            "models": catalog,
            "count": len(catalog),
            "tip": "Use recommend_model for task-specific recommendations",
        }

        duration = time.time() - start_time
        log_automation_execution("list_available_models", duration, True)

        return json.dumps(format_success_response(result), indent=2)

    except Exception as e:
        duration = time.time() - start_time
        log_automation_execution("list_available_models", duration, False, e)
        error_response = format_error_response(e, ErrorCode.AUTOMATION_ERROR)
        return json.dumps(error_response, indent=2)

