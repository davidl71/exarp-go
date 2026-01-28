# LLM / Text-Generation Abstractions → Reducing Python Dependency

**Question:** Can the recent LLM and text-generation abstractions reduce Python dependency?

**Short answer:** Yes. Adding an **Ollama-backed `TextGenerator`** and using it in the **FM chain** (Apple FM → Ollama → stub) would let several tools stop using the Python bridge for LLM/text. **ReportInsightProvider** can also be extended to prefer native backends (FM, then Ollama) before MLX, reducing how often we hit the bridge.

---

## Current Abstractions

| Abstraction | Interface | Implementations | Bridge? |
|-------------|-----------|-----------------|---------|
| **TextGenerator** | `Supported() bool`; `Generate(ctx, prompt, maxTokens, temp) (string, error)` | FMProvider, ReportInsightProvider | No (FM); Yes (ReportInsight via MLX) |
| **FMProvider** | Embeds TextGenerator | Apple FM (darwin/arm64/cgo), **stub** (others) | No |
| **ReportInsightProvider** | Embeds TextGenerator | **Composite:** MLX (bridge) → FM | Yes (MLX) |
| **OllamaProvider** | `Invoke(ctx, params) ([]TextContent, error)` | Native HTTP → bridge fallback | Yes (fallback only) |

- **DefaultFMProvider()** → Apple FM or stub. No bridge.
- **DefaultReportInsight()** → MLX then FM. Uses bridge for MLX.
- **DefaultOllama()** → Native then bridge. Uses bridge only on native failure.

---

## Where Python Is Used for LLM / Text Today

| Use | What uses it | Python's role |
|-----|----------------|---------------|
| **FMProvider** | task_analysis hierarchy, task_workflow clarify, context summarize, estimation, task_discovery | None. Apple or stub only. |
| **ReportInsightProvider** | Report/scorecard insights | **MLX via bridge** (then FM). |
| **context `action=summarize`** | When !FMAvailable or native fails | **Bridge** (Python summarizes, often via Ollama). |
| **task_workflow `action=clarify`** | When Apple FM unavailable | **Handler fallback to bridge** (Python does clarify, e.g. Ollama). |
| **ollama tool** | generate, docs, quality, summary, etc. | **Bridge fallback** when native fails. |

So Python is used for:

1. **MLX** – report insights (ReportInsightProvider).
2. **context summarize** – when no Apple FM (e.g. Linux).
3. **task_workflow clarify** – when no Apple FM (handler fallback).
4. **ollama** – only when native Ollama fails.

---

## How Abstractions Can Reduce Python Use

### 1. Ollama as TextGenerator (main lever)

**Idea:** Implement a `TextGenerator` that calls **native** Ollama generate (HTTP). Use it in a **chain** for foundation-model-style use.

**Concrete design:**

- Add **`ollamaTextGenerator`** (e.g. in `ollama_native.go` or `fm_ollama.go`):
  - Implements `TextGenerator`: `Supported()` checks Ollama status (or we try and fail); `Generate(ctx, prompt, maxTokens, temp)` calls `handleOllamaGenerate`-style logic with a default model (e.g. config or `llama3.2`).
- Introduce an **FM chain** used as **DefaultFMProvider**:
  - **Apple FM** (when darwin/arm64/cgo) → **Ollama** (when running) → **stub**.

All current `DefaultFMProvider()` call sites then get **Ollama on non–Apple-FM platforms** without Python:

| Tool | Today | With Ollama in FM chain |
|------|--------|---------------------------|
| **context summarize** | Bridge when !FMAvailable | Native (Ollama) when no Apple FM → no bridge for summarize |
| **task_workflow clarify** | Handler fallback to bridge | Native (Ollama) → **remove** clarify fallback to bridge |
| **task_analysis hierarchy** | ErrFMNotSupported off–darwin | Native (Ollama) when Ollama running |
| **task_discovery** | Same | Same |
| **estimation** | Same | Same |

**Result:** We can remove the **task_workflow** Python fallback for clarify and avoid **context** bridge for summarize on Linux when Ollama is available. No new Python; we reuse existing native Ollama HTTP.

### 2. ReportInsightProvider: FM → Ollama → MLX

**Today:** MLX (bridge) first, then FM.

**Change:** Use a **native-first** order:

1. **FM** (Apple)  
2. **Ollama** (new `TextGenerator` as above)  
3. **MLX** (bridge) only if both FM and Ollama fail or are unavailable  

**Result:** When Apple FM or Ollama works, we never call the bridge for report insights. Python (MLX) is used only as last resort.

### 3. Optional: ReportInsight uses “FM chain” instead of raw FM

**Today:** ReportInsight fallback is `DefaultFMProvider()` (Apple or stub).

**Change:** ReportInsight fallback uses the **same FM chain** as above (Apple → Ollama → stub). Then “FM” in the ReportInsight flow already includes Ollama, and we only hit MLX when both Apple FM and Ollama are unavailable.

---

## What Stays on Python (for now)

- **MLX** – No native Go implementation. Remains bridge-only. With (1) and (2), we use it less often (e.g. only when no Apple FM and no Ollama).
- **Ollama bridge fallback** – When native Ollama HTTP fails (e.g. wrong host, timeout), we still fall back to the bridge. We could eventually try to remove this if native coverage is good enough.
- **context/report/other** – Non-LLM Python use (e.g. batch logic, security, testing, etc.) is unchanged by these abstractions.

---

## Implementation Outline

1. **`ollamaTextGenerator`**
   - Implement `TextGenerator` using native Ollama generate (reuse `handleOllamaGenerate` or a small wrapper).
   - `Supported()`: e.g. Ollama status check or “always try.”
   - Config or default for model (e.g. `llama3.2`).

2. **FM chain**
   - Add a composite `FMProvider` that tries, in order: Apple FM → Ollama `TextGenerator` → stub.
   - Wire it as `DefaultFMProvider()` (or equivalent) so all current FM call sites use the chain.

3. **ReportInsightProvider**
   - Change order to: FM chain (Apple → Ollama) → MLX.
   - Optionally, use the same FM chain type as `DefaultFMProvider` for the “FM” step.

4. **Remove task_workflow clarify fallback**
   - Once the FM chain includes Ollama, drop the handler fallback to the Python bridge for clarify.

5. **Context**
   - No handler change needed if context summarize already uses `DefaultFMProvider()` (or the FM chain). It will automatically use Ollama when Apple FM is unavailable.

6. **Tests & docs**
   - Update `LLM_NATIVE_ABSTRACTION_PATTERNS.md`, `TASK_WORKFLOW_NATIVE_ONLY_ANALYSIS.md`, and any “when we use the bridge” docs.

---

## Summary

| Change | Python reduction |
|--------|-------------------|
| Ollama as `TextGenerator` + FM chain (Apple → Ollama → stub) | context summarize, task_workflow clarify, task_analysis, task_discovery, estimation use native Ollama off–darwin; **clarify fallback removed**. |
| ReportInsight: FM → Ollama → MLX | Report/scorecard insights use bridge (MLX) only when both FM and Ollama fail. |

The existing **TextGenerator / FMProvider / ReportInsightProvider** abstractions are already the right place to plug this in; the main new piece is an **Ollama-backed TextGenerator** and a **chain FMProvider** that uses it. That directly reduces Python use for LLM/text across several tools.

---

## Optional Tools We Didn’t Implement — Do They Help?

**Question:** There were tools we didn’t implement in the LLM abstraction (optional). Will implementing them help reduce Python?

**Short answer:** It depends **which** “optional” you mean. **Exposure** (unified `text_generate` tool, LLM prompt, `stdio://llm/status`) does **not** reduce Python. **Abstraction** work (Ollama as `TextGenerator`, FM chain, ReportInsight reorder) **does**.

---

### Optional exposure (docs: “Optional (not implemented)”)

From `LLM_EXPOSURE_OPPORTUNITIES.md`:

| Optional | What it is | Reduces Python? |
|----------|------------|------------------|
| **Unified `text_generate` tool** | Single tool calling `DefaultFMProvider()` or `DefaultReportInsight()` | **No.** Same backends, same bridge use. Just another entry point. |
| **LLM prompt** (“generate” / “llm”) | Prompt directing AI to use fm/ollama/mlx | **No.** Discovery/UX only. |
| **`stdio://llm/status`** | Dedicated resource for LLM backend status | **No.** Discovery only (status already in `stdio://models` backends). |

Implementing these improves **discovery and UX** but does **not** change which backends we call or when we use the bridge. **No** Python reduction.

---

### Optional abstraction work (what actually helps)

These are “optional” in the sense we **haven’t implemented them yet**:

| Optional | What it is | Reduces Python? |
|----------|------------|------------------|
| **Ollama as `TextGenerator`** | `TextGenerator` impl using native Ollama generate | **Yes.** Enables FM chain and ReportInsight reorder. |
| **FM chain** (Apple → Ollama → stub) | `DefaultFMProvider` tries Apple FM, then Ollama, then stub | **Yes.** context, task_workflow clarify, task_analysis, etc. use native Ollama off–darwin; clarify fallback can be removed. |
| **ReportInsight: FM → Ollama → MLX** | Native-first order before MLX (bridge) | **Yes.** Report insights use bridge only when FM and Ollama both fail. |

Implementing **these** reduces Python use. They extend the **abstraction** (new backend, new chain), not the **exposure** (new tool/prompt/resource).

---

### Summary

- **Implementing optional *tools* (e.g. `text_generate`) / prompts / resources:** **No** Python reduction.
- **Implementing optional *abstraction* work (Ollama `TextGenerator`, FM chain, ReportInsight reorder):** **Yes** — that’s what reduces Python dependency.
