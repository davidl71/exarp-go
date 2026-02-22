# Token Estimation Plan

**Tag hints:** `#feature` `#planning`

> Plan for adding token counting and cost estimation capabilities to exarp-go.
> These are future enhancements — no code changes required yet.

## Overview

Token estimation would allow exarp-go to:
- Count tokens accurately (matching OpenAI's tokenizer)
- Truncate content to exact token budgets
- Estimate API costs before sending requests
- Report token usage in the scorecard

## Task Breakdown

### 1. tiktoken-go as Optional Dependency (T-1771536390174034000)

**Approach:** Add `github.com/pkoukk/tiktoken-go` as an optional dependency.

- Build-tag gated (`//go:build tiktoken`) so it doesn't increase binary size for users who don't need it
- Fallback to word-based estimation (1 token ≈ 0.75 words) when tiktoken unavailable
- Provide `TokenCounter` interface in `internal/tools/` with two implementations:
  - `tiktokenCounter` (accurate, optional)
  - `heuristicCounter` (fast, always available)

### 2. Use tiktoken in Context Tool (T-1771535440218686000)

**Integration point:** `internal/tools/context.go` — `handleContextBudget()`

- Replace character-based budget calculation with token-based
- Use `TokenCounter` interface so it works with or without tiktoken
- Add `token_count` field to budget response alongside existing `char_count`

### 3. Scorecard OpenAI Token Count (T-1771535451393947000)

**Integration point:** `internal/tools/scorecard_go.go`

- Add "Estimated Context Tokens" metric to scorecard
- Count tokens across: task descriptions, code files, docs
- Show in scorecard output alongside existing metrics
- Use heuristic counter by default (fast enough for scorecard)

### 4. Exact Truncation to N Tokens (T-1771535454285198000)

**Integration point:** `internal/tools/context.go` — summarization pipeline

- Add `TruncateToTokens(text string, maxTokens int) string` utility
- Binary search on the token boundary for efficiency
- Preserve word boundaries (don't cut mid-word)
- Used by context summarize/batch when budget is specified in tokens

### 5. API Cost Estimation (T-1771535456833220000)

**Integration point:** New file `internal/tools/cost_estimation.go`

- Pricing table for common models (GPT-4o, Claude 3.5, etc.)
- `EstimateCost(inputTokens, outputTokens int, model string) float64`
- Expose via `text_generate` tool response as optional `estimated_cost` field
- Update pricing table periodically (store in config or constants)

## Implementation Order

```
1. tiktoken-go dependency  →  2. context tool integration
                           →  3. scorecard metric
                           →  4. exact truncation
                           →  5. cost estimation (depends on 1)
```

## Decision: Build-Tag Gating

tiktoken-go pulls in a ~4MB encoding data file. Using a build tag keeps the default binary lean:

```go
//go:build tiktoken

package tools

import "github.com/pkoukk/tiktoken-go"

type tiktokenCounter struct { enc *tiktoken.Tiktoken }
func (c *tiktokenCounter) Count(text string) int { ... }
```

```go
//go:build !tiktoken

package tools

type heuristicCounter struct{}
func (c *heuristicCounter) Count(text string) int {
    return len(strings.Fields(text)) * 4 / 3 // ~0.75 words per token
}
```

## Status

All five tasks are documented here as a cohesive plan. Implementation deferred to a future phase when token-aware context management becomes a priority.
