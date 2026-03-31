> **Archived** 2026-03-31 — snapshot of `docs/CONTEXT_TOOLS_COMPARISON.md` before trim. The live page is a short comparison plus a pointer to [CONTEXT_AND_CTX_AUDIT_2026-03.md](../../CONTEXT_AND_CTX_AUDIT_2026-03.md). Exarp task **T-1774976982390803000**.

# Context Tools Comparison: `context` vs `context_budget`

**Original date:** 2026-01-08  
**Purpose:** Compare the `context` and `context_budget` tools to understand their differences, similarities, and use cases

---

## Status 2026-03 (read this first)

**Canonical audit:** [CONTEXT_AND_CTX_AUDIT_2026-03.md](CONTEXT_AND_CTX_AUDIT_2026-03.md) — `context.Context` propagation, FM chain vs Apple-only copy, and follow-up tasks.

| Topic | Current truth |
|-------|----------------|
| **`context` tool** | Unified MCP tool: actions `summarize`, `budget`, `batch`, `count` — **`handlers_ai.go`** → **`context.go`** / **`context_shared.go`**. |
| **AI summarization** | **`DefaultFMProvider()`** (FM **chain**, typically **Ollama → stub** in `fm_chain.go`); not Python, not Apple-only. |
| **`context` action `budget`** | Native **`handleContextBudget`** in **`context.go`** (same implementation shape as dedicated `context_budget` tool). |
| **`context_budget` tool** | Still native Go in **`context.go`**; separate MCP registration for budget-only callers. |
| **Request cancellation** | Summarize/budget/batch/count honor **`ctx`** (canceled/deadline) on hot paths; see audit doc. |

**Below:** Older sections still mention Python bridge and Apple-only paths. Treat them as **historical** unless they match the table above. A full doc rewrite is tracked as an exarp task.

---

## Executive Summary

| Tool | Type | Implementation | Primary Use Case |
|------|------|----------------|------------------|
| **`context`** | Unified multi-action tool | Native Go (`handlers_ai` + `context` / `context_shared`; FM chain for summarize/batch) | Summarize, batch budget analysis, token count |
| **`context_budget`** | Single-purpose tool | Native Go only | Token budget analysis and planning |

**Key Insight:** `context` offers `summarize`, `budget`, `batch`, and `count` in one tool; `context_budget` is the dedicated budget tool. Both budget paths use native Go token estimation in `context.go`.

---

## Detailed Comparison

### 1. **Tool Architecture**

#### `context` Tool
- **Type:** Unified multi-action tool
- **Actions:** `summarize`, `budget`, `batch`, `count`
- **Implementation (2026-03):** Native Go
  - `summarize` / `batch`: **`HandleContextSummarizeShared`** / FM chain via **`DefaultFMProvider().Generate`**
  - `budget`: **`handleContextBudget`** in **`context.go`**
  - `count`: **`handleContextCount`**
- **Location:** `internal/tools/handlers_ai.go`, `context.go`, `context_shared.go`

#### `context_budget` Tool
- **Type:** Single-purpose tool
- **Actions:** None (single function)
- **Implementation:** Native Go only
- **Location:** `internal/tools/context.go`

---

### 2. **Functionality Comparison**

#### `context` Tool - `summarize` Action

**Purpose:** Compress verbose outputs using AI-powered summarization

**Features:**
- ✅ **AI-Powered:** Uses Apple Foundation Models (native Go) when available
- ✅ **Multiple Levels:** `brief`, `detailed`, `key_metrics`, `actionable`
- ✅ **Smart Summarization:** Semantic understanding, not just pattern matching
- ✅ **Privacy:** On-device processing with Apple FM
- ✅ **Fallback:** Python bridge when Apple FM unavailable

**Input:**
```json
{
  "action": "summarize",
  "data": "{\"status\":\"success\",\"score\":85,\"issues\":3}",
  "level": "brief"
}
```

**Output:**
```json
{
  "summary": "System achieved success with score 85, 3 issues identified",
  "method": "apple_foundation_models",
  "duration_ms": 1506,
  "token_estimate": {
    "original": 47,
    "summarized": 52,
    "reduction_percent": "-10.6"
  }
}
```

**Use Cases:**
- Compress large tool outputs before adding to context
- Reduce token usage for verbose JSON responses
- Extract key information from detailed reports
- Generate actionable summaries

---

#### `context_budget` Tool

**Purpose:** Analyze token usage and suggest context reduction strategy

**Features:**
- ✅ **Native Go:** Fast, no Python bridge overhead
- ✅ **Token Estimation:** Estimates tokens for multiple items
- ✅ **Budget Analysis:** Compares total tokens vs budget
- ✅ **Recommendations:** Suggests which items to summarize
- ✅ **Strategy Suggestions:** Provides reduction strategy

**Input:**
```json
{
  "items": "[{\"data\":\"...\"}, {\"data\":\"...\"}]",
  "budget_tokens": 4000
}
```

**Output:**
```json
{
  "total_tokens": 5234,
  "budget_tokens": 4000,
  "over_budget": true,
  "reduction_needed": 1234,
  "items": [
    {
      "index": 0,
      "tokens": 2341,
      "percent_of_budget": 58.5,
      "recommendation": "summarize_brief"
    },
    {
      "index": 1,
      "tokens": 1567,
      "percent_of_budget": 39.2,
      "recommendation": "summarize_key_metrics"
    }
  ],
  "strategy": "Summarize 2 items using 'brief' level. Estimated savings: 2736 tokens."
}
```

**Use Cases:**
- Plan context reduction before adding multiple items
- Identify which items consume most tokens
- Get recommendations for summarization levels
- Estimate token savings from summarization

---

### 3. **Performance Comparison**

| Metric | `context` (summarize) | `context_budget` |
|--------|----------------------|-----------------|
| **Implementation** | Native Go (Apple FM) / Python Bridge | Native Go only |
| **Typical Runtime** | 1-2 seconds (AI processing) | <10ms (token estimation) |
| **Dependencies** | Apple FM (optional) / Python | None (pure Go) |
| **Platform Support** | All (Python fallback) | All |
| **Token Estimation** | After summarization | Before summarization |

---

### 4. **When to Use Each Tool**

#### Use `context` Tool (`action="summarize"`) When:
- ✅ You have **one item** to summarize
- ✅ You want **AI-powered semantic summarization**
- ✅ You need **intelligent compression** (not just token counting)
- ✅ You want **multiple summarization levels** (brief, detailed, etc.)
- ✅ You're on **Apple Silicon** (for native Go + Apple FM performance)
- ✅ You need **privacy-preserving** on-device processing

**Example:**
```python
# After running a health check tool
health_result = health_check()
summary = context(action="summarize", data=health_result, level="brief")
# → "Health: 85/100, 3 issues, 2 actions"
```

---

#### Use `context_budget` Tool When:
- ✅ You have **multiple items** to analyze
- ✅ You want to **plan** context reduction before summarizing
- ✅ You need **fast token estimation** (no AI processing)
- ✅ You want **recommendations** on which items to summarize
- ✅ You need **budget analysis** (total vs budget comparison)
- ✅ You want **native Go performance** (no Python bridge)

**Example:**
```python
# Before adding multiple tool results to context
items = [tool1_result, tool2_result, tool3_result]
budget_analysis = context_budget(items=items, budget_tokens=4000)
# → Recommendations: "Summarize items 0 and 1 using 'brief' level"
```

---

### 5. **Workflow Integration**

#### Recommended Workflow:

1. **Planning Phase:** Use `context_budget` to analyze multiple items
   ```python
   budget_analysis = context_budget(items=all_tool_results, budget_tokens=4000)
   # Identify which items need summarization
   ```

2. **Summarization Phase:** Use `context` to summarize identified items
   ```python
   for item in items_to_summarize:
       summary = context(action="summarize", data=item, level="brief")
       # Add summary to context instead of full item
   ```

3. **Result:** Reduced context with key information preserved

---

### 6. **Implementation Details**

#### `context` Tool Implementation

**Native Go Path (Apple FM):**
- File: `internal/tools/context_shared.go` (`handleContextSummarizeNative` / `HandleContextSummarizeShared`)
- Build tags: `darwin && arm64 && cgo`
- Uses: Apple Foundation Models `summarize` action
- Fallback: Python bridge if Apple FM unavailable

**Python Bridge Path:**
- File: `bridge/context/tool.py`
- Uses: Pattern-based summarization
- Always available (fallback)

#### `context_budget` Tool Implementation

**Native Go Only:**
- File: `internal/tools/context.go`
- Function: `handleContextBudget()`
- Algorithm: Character-based token estimation (0.25 tokens/char)
- No dependencies: Pure Go, no external calls

---

### 7. **Token Estimation Methods**

#### `context` Tool
- **After Summarization:** Estimates tokens in original vs summarized text
- **Method:** Character-based estimation (0.25 tokens/char)
- **Purpose:** Show reduction achieved

#### `context_budget` Tool
- **Before Summarization:** Estimates tokens for each item
- **Method:** Character-based estimation (0.25 tokens/char)
- **Purpose:** Plan reduction strategy

**Note:** Both use the same estimation method (`TOKENS_PER_CHAR = 0.25`), but `context_budget` provides more detailed per-item analysis.

---

### 8. **Recommendation Logic**

#### `context_budget` Recommendations

Based on item size vs budget ratio:

| Ratio | Recommendation | Meaning |
|-------|----------------|---------|
| > 50% | `summarize_brief` | Item is >50% of budget, summarize aggressively |
| > 25% | `summarize_key_metrics` | Item is 25-50% of budget, extract metrics only |
| > 10% | `keep_detailed` | Item is 10-25% of budget, keep with details |
| ≤ 10% | `keep_full` | Item is <10% of budget, keep as-is |

---

### 9. **Migration Status**

| Tool | Status | Notes |
|------|--------|-------|
| `context` (summarize) | ✅ **Migrated** | Native Go with Apple FM (best candidate) |
| `context` (budget) | ⚠️ **Python Bridge** | Could migrate to use `context_budget` internally |
| `context` (batch) | ⚠️ **Python Bridge** | Could migrate to native Go |
| `context_budget` | ✅ **Native Go** | Already fully migrated |

**Potential Optimization:**
- `context` tool's `budget` action could route to `context_budget` instead of Python bridge
- This would provide native Go performance for budget analysis

---

### 10. **Summary Table**

| Feature | `context` (summarize) | `context_budget` |
|---------|----------------------|------------------|
| **Purpose** | Summarize single item | Analyze multiple items |
| **AI-Powered** | ✅ Yes (Apple FM) | ❌ No (token estimation) |
| **Implementation** | Native Go + Python | Native Go only |
| **Speed** | 1-2 seconds | <10ms |
| **Input** | Single JSON string | Array of items |
| **Output** | Summarized text | Budget analysis + recommendations |
| **Use Case** | Compress after tool execution | Plan before adding to context |
| **Privacy** | ✅ On-device (Apple FM) | ✅ On-device (native Go) |
| **Platform** | All (Python fallback) | All |

---

## Recommendations

### For Users:

1. **Use `context_budget` first** when you have multiple items to add to context
   - Get recommendations on which items to summarize
   - Plan your context reduction strategy

2. **Use `context` (summarize) second** to actually summarize identified items
   - Leverage AI-powered summarization
   - Get intelligent compression

3. **Workflow:**
   ```
   Plan → context_budget → Identify items → context (summarize) → Add to context
   ```

### For Developers:

1. **Consider routing `context` (budget) to `context_budget`** instead of Python bridge
   - Provides native Go performance
   - Reduces Python bridge overhead
   - Maintains unified `context` tool interface

2. **Keep both tools** - they serve different purposes:
   - `context_budget`: Fast planning and analysis
   - `context` (summarize): AI-powered summarization

---

## Related (2026-03)

- [CONTEXT_AND_CTX_AUDIT_2026-03.md](CONTEXT_AND_CTX_AUDIT_2026-03.md) — `ctx` behavior, FM chain, exarp task IDs for doc rewrite and design decisions.

---

## Conclusion

Both tools are valuable and serve complementary purposes:

- **`context_budget`**: Fast, native Go tool for planning and analysis
- **`context` (summarize)**: AI-powered tool for intelligent summarization

The recommended workflow is to use `context_budget` for planning, then `context` (summarize) for execution. This provides both speed (planning) and intelligence (summarization).

