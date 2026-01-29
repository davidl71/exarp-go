"""
MCP Prompts – legacy stub (retired).

All prompts migrated to exarp-go (stdio://prompts). See internal/prompts/templates.go
and docs/PROMPTS_MIGRATION_AND_OBSOLESCENCE_PLAN.md.

Consumers (project_scorecard, project_overview) use LEGACY_PROMPTS_COUNT for metrics
only; canonical prompt count is from exarp-go.
"""
LEGACY_PROMPTS_COUNT = 0  # All prompts in Go; Python manifest retired
