# Agno-Go learnings and exarp-go adoption

Patterns from [Agno-Go](https://github.com/rexleimo/agno-Go) (multi-agent framework in Go) that exarp-go can adopt without depending on Agno-Go. See [.cursor/plans/context-cache-compression.plan.md](../.cursor/plans/context-cache-compression.plan.md) for the full plan.

## Summary table

| Learning | Agno-Go | exarp-go adoption |
|----------|---------|-------------------|
| **Token metrics in API** | Input/output/cache/reasoning tokens per run/session/message; exposed via OpenAPI. | **Adopted:** `token_estimate` in JSON envelope for session (prime), report (overview/scorecard/briefing), task_workflow (list). Uses `config.TokensPerChar()`; see [CONTEXT_REDUCTION_OPTIONS.md](CONTEXT_REDUCTION_OPTIONS.md). |
| **Per-session observability** | Session storage, metrics per session. | **Adopted:** Prime and list responses carry `token_estimate` so clients know the “cost” of this session’s context. Handoff and session state already provide observability. |
| **Semantic compression** | `EnableSemanticCompression`, max tokens post-compression, optional semantic model. | **Consider:** Optional rule-based abbreviation (preserve dates/IDs/versions); no separate model. See plan §4. |
| **Lightweight lifecycle** | Agent ~180ns, ~1.2KB; request-scoped semantics. | **Consider:** Request-scoped cache ([ctxcache](https://pkg.go.dev/github.com/lawlielt/ctxcache)) for per-request memoization without cross-request state; see plan §1. |
| **Multi-provider** | OpenAI, Anthropic, Gemini, Ollama, etc. | **Already:** exarp-go has fm/ollama/mlx/localai. Token accuracy (e.g. tiktoken-go) matters most for OpenAI-compatible; see plan §3. |
| **Structured logging / health** | AgentOS REST, health checks, CORS, timeouts. | **Already:** Health tool and structured logging. Token/cost hints don’t require new infra. |

## References

- Agno-Go: [github.com/rexleimo/agno-Go](https://github.com/rexleimo/agno-Go)
- Agno-Go semantic compression (blog): [Agno-Go: Semantic Compression for Efficient LLM Interactions](https://devevangelista.medium.com/agno-go-semantic-compression-for-efficient-llm-interactions-1545fc7673b1)
- Plan: [.cursor/plans/context-cache-compression.plan.md](../.cursor/plans/context-cache-compression.plan.md)
- Context reduction: [CONTEXT_REDUCTION_OPTIONS.md](CONTEXT_REDUCTION_OPTIONS.md)
- Token estimation (tiktoken): [GO_AI_ECOSYSTEM.md §5.5](GO_AI_ECOSYSTEM.md#55-token-estimation-nice-to-have)
