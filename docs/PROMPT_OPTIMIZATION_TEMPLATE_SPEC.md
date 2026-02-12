# Prompt Optimization Template Spec

**Purpose:** Define variable placeholders and output format for prompt optimization templates used with local models (Ollama, MLX, Apple Foundation Models).

**Reference:** [MODEL_ASSISTED_WORKFLOW.md](MODEL_ASSISTED_WORKFLOW.md) — Prompt Optimizer section

---

## Template Variables

All optimization templates use `{variable_name}` placeholders (per `internal/prompts/substituteTemplate`).

| Variable    | Required | Description |
|-------------|----------|-------------|
| `{prompt}`  | Yes      | The prompt text to analyze or optimize |
| `{context}` | No       | Optional context (e.g. task type, project domain) |
| `{task_type}` | No    | Hint for task-type variants (e.g. "code", "docs", "planning") |

---

## Analysis Output Schema

The **analysis** template produces structured output. Parse as JSON when possible.

### Dimensions (0.0–1.0 scale)

| Dimension    | Description |
|--------------|-------------|
| **clarity** | Is the task clearly defined? Unambiguous wording? |
| **specificity** | Are requirements specific enough? Concrete vs vague? |
| **completeness** | Are all necessary details included? No critical gaps? |
| **structure** | Is the prompt well-organized? Logical flow? |
| **actionability** | Can the AI execute this without clarification? |

### Output Format

```json
{
  "clarity": 0.85,
  "specificity": 0.6,
  "completeness": 0.7,
  "structure": 0.8,
  "actionability": 0.5,
  "summary": "Brief overall assessment.",
  "notes": ["Optional per-dimension notes."]
}
```

### Fallback (non-JSON)

If the model returns plain text, parse line-by-line for scores:
- `Clarity: 0.85` or `clarity: 0.85`
- `Specificity: 0.6` or `specificity: 0.6`
- etc.

---

## Suggestions Format

The **suggestions** template produces improvement suggestions.

### Output Format

```json
{
  "suggestions": [
    {
      "dimension": "specificity",
      "issue": "What is missing or vague.",
      "recommendation": "Concrete improvement."
    }
  ]
}
```

### Fallback (plain text)

- Numbered list: `1. ... 2. ...`
- Or bullet list with `dimension:` prefix

---

## Task Type Variants

When `{task_type}` is set, apply type-specific guidance:

| Task Type | Analysis/Suggestions focus | Refinement focus |
|-----------|----------------------------|------------------|
| **code**  | File paths, APIs, test requirements, concrete I/O | Include file paths, APIs, test requirements |
| **docs**  | Structure, audience, scope | Clear structure, audience, scope |
| **general** (or empty) | Balanced across all dimensions | Balanced refinement |

---

## Template Types

| Template              | Variables       | Output          | Task |
|-----------------------|-----------------|-----------------|------|
| Analysis              | prompt, context | Analysis schema | T-1770830685549 |
| Suggestions generation| prompt, analysis, context | Suggestions format | T-1770830686054 |
| Refinement            | prompt, suggestions, context | Refined prompt text | T-1770830686525 |

---

## Model Compatibility

- **Ollama** — JSON mode preferred when supported; otherwise instruct to return structured JSON.
- **MLX** — Same as Ollama; Phi-3.5, CodeLlama, etc.
- **Apple FM** — Same pattern; instruct clearly for structured output.

---

## References

- `internal/prompts/templates.go` — `substituteTemplate` uses `{key}` from `args`
- `docs/MODEL_ASSISTED_WORKFLOW.md` — Optimization criteria and `PromptAnalysis` struct
