---
status: Done
title: Gateway Provider and Optional Model Param
todos:
    - content: Create docs/plans/2026-02-27-gateway-provider.plan.md with YAML todos and milestone checkboxes
      id: T-1
      status: done
    - content: Run task_workflow sync_from_plan to create Todo2 tasks from the plan
      id: T-2
      status: done
    - content: Add gateway_provider.go (TextGenerator, OPENAI_GATEWAY_* env, HTTP /v1/chat/completions)
      id: T-3
      status: done
    - content: Register provider=gateway in text_generate and LLMBackendStatus in llm_backends.go
      id: T-4
      status: done
    - content: Add optional model param to text_generate for localai and gateway; extend LocalAI/Gateway with model override
      id: T-5
      status: done
    - content: Add gateway_provider_test.go and update GO_AI_ECOSYSTEM or research doc
      id: T-6
      status: done
    - content: Document compose pattern exarp-go → llm-router → RouteLLM → providers
      id: T-7
      status: done
    - content: Document Cursor + kcolemangt/llm-router (Base URL, ngrok, provider=gateway)
      id: T-8
      status: done
    - content: Optional ModelRouter + gateway for provider=auto
      id: T-9
      status: done
updated: "2026-03-08"
---

# Gateway Provider and Optional Model Param

Implemented and verified `provider=gateway` plus optional `model` override per [LLM_ROUTER_AND_ROUTELLM_RESEARCH.md](../research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md) §6.5.

**Verification summary:**
- `internal/tools/gateway_provider.go` implements the OpenAI-compatible gateway provider with `OPENAI_GATEWAY_BASE_URL`, optional `OPENAI_GATEWAY_MODEL`, and optional `OPENAI_GATEWAY_API_KEY`.
- `internal/tools/text_generate.go` supports `provider=gateway` and optional `model` overrides for `localai` and `gateway`.
- `internal/tools/llm_backends.go` advertises gateway availability.
- `internal/tools/model_router.go` includes `ModelGateway` for `provider=auto` selection when configured.
- `internal/tools/gateway_provider_test.go` exists and the focused verification suite passed on 2026-03-08.
- `docs/GO_AI_ECOSYSTEM.md` and `docs/research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md` document the composed llm-router/RouteLLM and Cursor integration patterns.

**Task IDs:** This plan uses short sequential IDs (T-1..T-9) for a small, self-contained set of tasks. `T-2` is represented in the plan only; the synced Todo2 set contains the implementation tasks (`T-1`, `T-3`..`T-9`).

## Milestones

- [x] **Create plan doc** (T-1)
- [x] **Run sync_from_plan** (T-2)
- [x] **Add gateway_provider.go** (T-3)
- [x] **Register gateway in text_generate and llm_backends** (T-4)
- [x] **Optional model param and GenerateWithModel** (T-5)
- [x] **Tests and docs** (T-6)
- [x] **Document compose pattern** (T-7)
- [x] **Document Cursor + kcolemangt/llm-router** (T-8)
- [x] **Optional ModelRouter + gateway** (T-9)

## Reference

- [internal/tools/gateway_provider.go](internal/tools/gateway_provider.go) — OpenAI-compatible gateway implementation
- [internal/tools/text_generate.go](internal/tools/text_generate.go) — provider switch and model override dispatch
- [internal/tools/llm_backends.go](internal/tools/llm_backends.go) — discovery
- [internal/tools/model_router.go](internal/tools/model_router.go) — provider=auto gateway selection
- [docs/GO_AI_ECOSYSTEM.md](docs/GO_AI_ECOSYSTEM.md) — gateway usage and Cursor note
- [docs/research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md](docs/research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md) — routing composition details
