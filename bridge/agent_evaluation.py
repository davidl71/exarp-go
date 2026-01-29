#!/usr/bin/env python3
"""
AgentScope Evaluation Bridge for exarp-go

Executes AgentScope evaluations via Python bridge.
This allows exarp-go to use AgentScope for advanced agent evaluation.
"""

import json
import sys
import argparse
from pathlib import Path

try:
    from agentscope.eval import Evaluator
    AGENTSCOPE_AVAILABLE = True
except ImportError:
    AGENTSCOPE_AVAILABLE = False
    print("Warning: AgentScope not installed. Install with: pip install agentscope")


def load_evaluation_config(config_path: str) -> dict:
    """Load evaluation configuration from YAML file."""
    try:
        import yaml
    except ImportError:
        # Fallback to JSON if PyYAML not available
        import json
        config_file = Path(config_path)
        if config_path.endswith('.yaml') or config_path.endswith('.yml'):
            # Try JSON version
            json_path = config_path.replace('.yaml', '.json').replace('.yml', '.json')
            config_file = Path(json_path)
        
        if not config_file.exists():
            raise FileNotFoundError(f"Evaluation config not found: {config_path}. Install PyYAML with: pip install pyyaml")
        
        with open(config_file, 'r') as f:
            return json.load(f)
    
    config_file = Path(config_path)
    if not config_file.exists():
        raise FileNotFoundError(f"Evaluation config not found: {config_path}")
    
    with open(config_file, 'r') as f:
        config = yaml.safe_load(f)
    
    return config


def evaluate_agent_behavior(config: dict, output_dir: str) -> dict:
    """Evaluate agent behavior using AgentScope."""
    if not AGENTSCOPE_AVAILABLE:
        return {
            "error": "AgentScope not available",
            "message": "Install AgentScope with: pip install agentscope"
        }
    
    try:
        # Create output directory
        output_path = Path(output_dir)
        output_path.mkdir(parents=True, exist_ok=True)
        
        # Initialize evaluator
        evaluator = Evaluator(config=config)
        
        # Run evaluation
        results = evaluator.run()
        
        # Save results
        results_file = output_path / "evaluation_results.json"
        with open(results_file, 'w') as f:
            json.dump(results, f, indent=2)
        
        return {
            "status": "success",
            "results": results,
            "output_file": str(results_file)
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "type": type(e).__name__
        }


def main():
    """Main entry point for agent evaluation."""
    parser = argparse.ArgumentParser(description="AgentScope evaluation for exarp-go")
    parser.add_argument(
        "--config",
        type=str,
        default=".github/agentscope_eval.yaml",
        help="Path to evaluation configuration file"
    )
    parser.add_argument(
        "--output",
        type=str,
        default="agent_results",
        help="Output directory for evaluation results"
    )
    
    args = parser.parse_args()
    
    # Load configuration
    try:
        config = load_evaluation_config(args.config)
    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        print("Creating default configuration...", file=sys.stderr)
        # Create default config if not found
        default_config = {
            "evaluator": {
                "name": "exarp-go-agent-evaluator",
                "type": "basic"
            },
            "test_cases": [
                {
                    "name": "mcp_tool_validation",
                    "description": "Validate MCP tools work correctly"
                },
                {
                    "name": "task_workflow_validation",
                    "description": "Validate task workflow functionality"
                }
            ]
        }
        config = default_config
    
    # Run evaluation
    result = evaluate_agent_behavior(config, args.output)
    
    # Output results as JSON
    print(json.dumps(result, indent=2))
    
    # Exit with appropriate code
    if result.get("status") == "error":
        sys.exit(1)
    sys.exit(0)


if __name__ == "__main__":
    main()
