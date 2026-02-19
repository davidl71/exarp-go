# Apple Foundation Models: Planner–Worker Flow and Comparison

**Reference:** [LLMs Calling LLMs: Building AI Agents with Apple's Foundation Models and Tool Calling](https://www.natashatherobot.com/p/ai-agents-apples-foundation-models-tool-calling) (Natasha the Robot, Jul 2025).

This doc describes how exarp-go uses Apple Foundation Models (and the FM chain), how that compares to the article’s Swift/agent pattern, what we added (plan-and-execute in Go), and what we could improve.

---

## 1. How We Use FM Today

### 1.1 Single-session, no tool calling

- **API:** We use `go-foundationmodels`: `fm.NewSession()`, `sess.RespondWithOptions(prompt, maxTokens, temperature)`. We do **not** use `RegisterTool` / `RespondWithTools` in our code.
- **Surface:** MCP tool `apple_foundation_models` with actions: `status`, `hardware`, `models`, `generate`, `respond`, `summarize`, `classify`.
- **Internal use:** `DefaultFMProvider().Generate(ctx, prompt, maxTokens, temperature)` is used by:
  - `task_analysis` (hierarchy, classification),
  - `task_workflow` (clarify),
  - `task_discovery`, `estimation`, `context`, `infer_task_progress`, report fallback.
- **Backend chain:** `DefaultFMProvider()` is a chain: **Apple FM → Ollama → stub**. So when Apple FM isn’t available (e.g. non‑darwin or no CGO), Ollama is used; otherwise we return a clear “not supported” style message.

So today: **one session per request, plain text in/out, no FM-level tool calling.**

### 1.2 Context and limits

- **Context window:** 4096 tokens (input + output) per session, same as in the article.
- **No guided generation in our FM path:** We don’t use structured/guided output types on the coordinator; the article notes that using a “generating” type on the coordinator can conflict with tool calling. Our usage is consistent with “no guided generation” for the coordinator.

---

## 2. Article’s Pattern vs Ours

| Aspect | Article (Swift / Foundation Models) | exarp-go (Go) |
|--------|-------------------------------------|---------------|
| **Agent definition** | LLM running tools in a loop, possibly with other LLMs | We don’t implement “LLM as agent” inside FM; the MCP server is the agent, tools are handlers |
| **Tool calling** | Coordinator has **tools** (Planner + Worker) that are **other LLM sessions** | We don’t expose LLMs as FM tools; we orchestrate in Go with multiple `Generate` calls |
| **Plan–execute** | Coordinator → Planner (tool) → Workers (tools), all via FM tool API | **New:** `fm_plan_and_execute` does plan → workers in **Go**: one planner call, N worker calls in parallel, then combine |
| **Context multiplication** | Each tool = new session ⇒ multiple 4096-token windows | Same idea: planner = 1 call, each worker = 1 call ⇒ multiple effective windows |
| **Backend** | Apple Foundation Models only | FM chain: Apple FM first, then Ollama, then stub |

So we get **similar “plan and execute” behavior** by doing orchestration in Go and using the existing FM (or Ollama) for each step, without using the Foundation Models **tool-calling** API.

---

## 3. What We Added: `fm_plan_and_execute`

- **Tool:** `fm_plan_and_execute`
- **Backend:** `DefaultFMProvider()` (Apple FM or Ollama).
- **Flow:**
  1. **Planner:** One `Generate` with a prompt that asks for a JSON array of subtask strings (or we parse numbered/bullet lines).
  2. **Workers:** For each subtask, one `Generate` in parallel (“Complete this subtask concisely…”).
  3. **Combine:** Results are concatenated in subtask order and returned as a single text (markdown sections).
- **Parameters:** `task` (required), `max_subtasks` (default 5, cap 20), `plan_max_tokens`, `worker_max_tokens` (defaults 512).
- **Parsing:** We try JSON array first; if that fails, we strip markdown code fences and retry JSON; then fall back to numbered/bullet line parsing.

This gives a **planner–worker style flow** without implementing “LLM as tool” inside Foundation Models.

---

## 4. How We Use FM vs the Article (Summary)

- **Same:** 4096-token limit, plan–execute idea (decompose → execute in parallel → combine), benefit of multiple “context windows” via multiple calls.
- **Different:**
  - **Orchestration:** We do it in Go (one planner call + N worker calls); the article does it in Swift with FM **tool calling** (Planner and Worker are tools that run other sessions).
  - **Tool calling:** We don’t use `RegisterTool` / `RespondWithTools`; the article’s coordinator uses tools for Planner and Workers.
  - **Guided generation:** We don’t use a “generating” type on a coordinator; the article explicitly warns against it when using tool calling.

---

## 5. What We Could Improve

### 5.1 FM-native tool calling (Planner/Worker as tools)

- **Idea:** Use go-foundationmodels’ `RegisterTool` and `RespondWithTools` so that the **coordinator** has two tools: “Planner” and “Worker”. Each tool’s `Execute` would create a new session and run the appropriate prompt, then return the result. That would mirror the article’s “LLM in your LLM” pattern.
- **Pros:** Closer to the article; the model decides when to call Planner vs Workers and in what order.
- **Cons:** go-foundationmodels and the article both note that FM tool calling can be **inconsistent**; we’d need to validate behavior and possibly add retries or fallbacks.

### 5.2 Session and context management

- **Article:** Don’t call `respond(to:)` again before the previous call returns; watch coordinator context (it sees all tool inputs/outputs).
- **Us:** Each `Generate` is a new session (or we use the chain), so we don’t have “one long-lived coordinator session.” For a future FM-tool-calling coordinator, we’d need: no concurrent `RespondWithTools` on the same session, and a strategy so the coordinator context doesn’t overflow (e.g. truncate or summarize tool results).

### 5.3 Guided generation vs tool calling

- **Article:** Avoid guided generation (e.g. `generating: FinalAnswer.self`) on the coordinator when using tool calling.
- **Us:** We don’t use guided generation on FM today. If we add FM tool calling, we should keep the coordinator as plain text (e.g. `RespondWithTools` without a structured output type).

### 5.4 Observability and debugging

- **Article:** Full trace of instructions, prompts, tool calls, and responses.
- **Us:** We could add optional logging or a “trace” mode for `fm_plan_and_execute` (planner prompt, parsed subtasks, per-worker prompt/result, combined output) to aid debugging and tuning.

### 5.5 Recursion / “Worker spawns workers”

- **Article:** Recursive delegation (worker that can call itself) hit framework limits and led to an explosion of workers; they advise avoiding it without solid testing.
- **Us:** Our current flow is one level only (planner → workers). If we ever add FM-based tools, we should avoid giving the Worker tool an instance of itself or a generic “delegate” tool that can recurse without a hard cap.

### 5.6 Mapping planner output to worker input

- **Article:** Stresses that when tools are chained, the output of one must match the input of the next (e.g. Planner outputs subtasks, Worker takes one subtask).
- **Us:** We already enforce that by parsing the planner response into `[]string` and passing each string as the worker prompt. If we moved to FM tool calling, we’d define the Planner tool’s output schema and the Worker tool’s input schema so they align.

---

## 6. References

- Article: [LLMs Calling LLMs: Building AI Agents with Apple's Foundation Models and Tool Calling](https://www.natashatherobot.com/p/ai-agents-apples-foundation-models-tool-calling)
- FM in exarp-go: `internal/tools/apple_foundation.go`, `fm_apple.go`, `fm_chain.go`, `fm_provider.go`
- Plan–execute tool: `internal/tools/fm_plan_execute.go`
- LLM abstraction: `docs/LLM_NATIVE_ABSTRACTION_PATTERNS.md`, `.cursor/rules/llm-tools.mdc`
- go-foundationmodels: `vendor/github.com/blacktop/go-foundationmodels/` (doc.go, fm.go — tool calling, session, context)
