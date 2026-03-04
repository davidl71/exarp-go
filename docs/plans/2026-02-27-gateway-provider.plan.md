---
status: Todo
title: Gateway Provider and Optional Model Param
todos:
    - content: Create docs/plans/2026-02-27-gateway-provider.plan.md with YAML todos and milestone checkboxes
      id: T-1
      status: pending
    - content: Run task_workflow sync_from_plan to create Todo2 tasks from the plan
      id: T-2
      status: pending
    - content: Add gateway_provider.go (TextGenerator, OPENAI_GATEWAY_* env, HTTP /v1/chat/completions)
      id: T-3
      status: pending
    - content: Register provider=gateway in text_generate and LLMBackendStatus in llm_backends.go
      id: T-4
      status: pending
    - content: Add optional model param to text_generate for localai and gateway; extend LocalAI/Gateway with model override
      id: T-5
      status: pending
    - content: Add gateway_provider_test.go and update GO_AI_ECOSYSTEM or research doc
      id: T-6
      status: pending
    - content: Document compose pattern exarp-go → llm-router → RouteLLM → providers
      id: T-7
      status: pending
    - content: Document Cursor + kcolemangt/llm-router (Base URL, ngrok, provider=gateway)
      id: T-8
      status: pending
    - content: Optional ModelRouter + gateway for provider=auto
      id: T-9
      status: pending
updated: "2026-02-27"
---

# Gateway Provider and Optional Model Param

Implement `provider=gateway` and optional `model` param per [LLM_ROUTER_AND_ROUTELLM_RESEARCH.md](../research/LLM_ROUTER_AND_ROUTELLM_RESEARCH.md) §6.5.

**Task IDs:** This plan uses short sequential IDs (T-1..T-9) for a small, self-contained set of tasks. Other exarp plans (e.g. mcp-go-core-extraction, exarp-go-generated) use **epoch-based IDs** (e.g. T-1772056740723802000) so tasks sort and align with the rest of the project. For new plans that should match the main task DB, prefer epoch-style IDs; sync_from_plan accepts any valid `T-<digits>` ID.

## Milestones

- [ ] **Create plan doc** (T-1)
- [ ] **Run sync_from_plan** (T-2)
- [ ] **Add gateway_provider.go** (T-3)
- [ ] **Register gateway in text_generate and llm_backends** (T-4)
- [ ] **Optional model param and GenerateWithModel** (T-5)
- [ ] **Tests and docs** (T-6)
- [ ] **Document compose pattern** (T-7)
- [ ] **Document Cursor + kcolemangt/llm-router** (T-8)
- [ ] **Optional ModelRouter + gateway** (T-9)

## Reference

- [internal/tools/localai_provider.go](internal/tools/localai_provider.go) — pattern to mirror
- [internal/tools/text_generate.go](internal/tools/text_generate.go) — provider switch
- [internal/tools/llm_backends.go](internal/tools/llm_backends.go) — discovery
