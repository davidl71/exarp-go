# Tool Parameter Parsing Strategy

## Overview

MCP tool calls pass arguments as a JSON object. The Go SDK deserializes them into
`map[string]interface{}` before handing them to the tool handler. Every handler
signature looks like:

```go
func handleFoo(ctx context.Context, params map[string]interface{}) ([]framework.TextContent, error)
```

Handlers that accept raw bytes instead receive `json.RawMessage` and must
unmarshal the map themselves (see Pattern 3 below).

Because all values arrive as `interface{}`, the handler must safely extract each
parameter — handling missing keys, wrong types, and appropriate defaults — before
doing any real work.

Three patterns cover every case in the codebase.

---

## Pattern 1: Helper Functions (`ParamString`, `ParamBool`, `ParamInt`, `ParamEnum`)

**File:** `internal/tools/params_helpers.go`

The preferred approach for scalar params. Helpers in `params_helpers.go` wrap
the `github.com/spf13/cast` library, which converts any reasonable value to the
target type without panicking.

### Functions

| Helper | Returns | Notes |
|---|---|---|
| `ParamString(params, key)` | `string` (trimmed) | Never panics; empty string on miss |
| `ParamBool(params, key, default)` | `bool` | Returns `default` on miss or wrong type |
| `ParamInt(params, key, default)` | `int` | Returns `default` on miss or wrong type |
| `ParamEnum(params, key, valid, default)` | `(string, error)` | Case-insensitive; error on unrecognized value |
| `RequireParam(params, key)` | `(string, error)` | Error if missing or empty |
| `RequireEnum(params, key, valid)` | `(string, error)` | Combined require + enum check |
| `HasKey(params, key)` | `bool` | True even if value is nil/empty |

### Examples

**Dispatcher in `session.go` (line 38):**
```go
action := strings.TrimSpace(cast.ToString(params["action"]))
if action == "" {
    action = "prime"
}
```
`cast.ToString` is the same underlying mechanism as `ParamString`. Both trim
whitespace and return an empty string if the key is absent.

**Dispatcher in `task_workflow_native.go` (line 22):**
```go
action := ParamString(params, "action")
if action == "" {
    action = "list"
}
```

**Boolean with default in `report.go` (line 33):**
```go
includePlanning := cast.ToBool(params["include_planning"])
```
`cast.ToBool` returns `false` for missing keys — the implicit default here.

**Boolean with explicit default in `task_workflow_crud.go` (lines 44–47):**
```go
dryRun := false
if _, ok := params["dry_run"]; ok {
    dryRun = cast.ToBool(params["dry_run"])
}
```
`HasKey` + `cast.ToBool` is used when `false` is a valid explicit value and
you need to distinguish "not provided" from "explicitly false".

**Required string in `task_workflow_actions.go` (lines 68–71):**
```go
taskID := cast.ToString(params["task_id"])
if taskID == "" {
    return nil, fmt.Errorf("apply_approval_result requires task_id")
}
```

**Enum validation with `RequireEnum`:**
```go
action, err := RequireEnum(params, "action", []string{"list", "create", "update", "delete"})
if err != nil {
    return nil, err
}
```

**Optional enum with default via `ParamEnum`:**
```go
format, err := ParamEnum(params, "output_format", []string{"json", "markdown", "html", "text"}, "text")
if err != nil {
    return nil, err
}
```

---

## Pattern 2: Direct Type Assertion `params["key"].(Type)`

Used for numeric and boolean types where the JSON decoder always produces a
specific Go type, or where fallback to a default is expressed inline.

### When JSON produces predictable types

The standard `encoding/json` decoder always decodes JSON numbers as `float64`
when the target is `interface{}`, and JSON booleans as `bool`. This makes direct
type assertions safe when the schema enforces the type.

**Float64 in `apple_foundation.go` (lines 150–157):**
```go
if temp, ok := params["temperature"].(float64); ok {
    temp32 := float32(temp)
    options.Temperature = &temp32
}
if maxTokens, ok := params["max_tokens"].(float64); ok {
    maxTokensInt := int(maxTokens)
    options.MaxTokens = &maxTokensInt
}
```
Always use the two-value assertion `v, ok := params["key"].(Type)` — never the
single-value form, which panics on type mismatch or nil.

**String with inline default in `automation_native.go` (lines 42–44):**
```go
action, _ := params["action"].(string)
if action == "" {
    action = "daily"
}
```

**Conditional field in `report.go` (lines 90–96):**
```go
if sc, ok := params["score"].(float64); ok {
    score = sc
} else if sc, ok := params["score"].(int); ok {
    score = float64(sc)
} else {
    score = 50.0 // Default score
}
```
Multiple type branches handle callers that may send integer or float values.

**Elicitation result in `task_workflow_crud.go` (lines 76–82):**
```go
if content != nil {
    if proceed, ok := content["proceed"].(bool); ok && !proceed {
        return framework.FormatResult(...cancelled...), nil
    }
    if dr, ok := content["dry_run"].(bool); ok && dr {
        dryRun = true
    }
}
```

### When to prefer Pattern 1 over Pattern 2

Use helpers (`cast.ToString`, `ParamString`, etc.) when:
- The value might arrive as a number, bool, or string depending on the caller.
- You want automatic whitespace trimming.
- The param is required and you want a standard error message.

Use direct assertion `.(type)` when:
- The JSON schema strictly defines the type (e.g., `"type": "number"`).
- The param is optional and the zero value is acceptable as a default.
- You are handling a nested map returned from elicitation or a sub-call.

---

## Pattern 3: JSON Re-marshal for Complex or Nested Params

Some params carry structured data: a JSON array of task definitions, a
JSON string representing a nested struct, or a `[]interface{}` that must be
decoded into a typed slice. Two sub-patterns cover this.

### 3a: JSON string param decoded to typed struct

**Batch task creation in `task_workflow_create_ai.go` (lines 47–60):**
```go
var taskDefs []map[string]interface{}
switch v := tasksParam.(type) {
case string:
    if err := json.Unmarshal([]byte(v), &taskDefs); err != nil {
        return nil, fmt.Errorf("tasks param must be a valid JSON array: %w", err)
    }
case []interface{}:
    for _, item := range v {
        if m, ok := item.(map[string]interface{}); ok {
            taskDefs = append(taskDefs, m)
        }
    }
default:
    return nil, fmt.Errorf("tasks param must be a JSON array string or array of objects")
}
```
The `tasks` schema field is `"type": "string"` (a JSON-encoded array). Some
callers send it pre-decoded as `[]interface{}`; the type switch handles both.

**task_journal in `session_handoff.go` (lines 185–196):**
```go
if journal, ok := params["task_journal"].([]interface{}); ok && len(taskJournal) == 0 {
    for _, v := range journal {
        if m, ok := v.(map[string]interface{}); ok {
            taskJournal = append(taskJournal, m)
        }
    }
}
if journalRaw, ok := params["task_journal"].(string); ok && journalRaw != "" && len(taskJournal) == 0 {
    var decoded []map[string]interface{}
    if err := json.Unmarshal([]byte(journalRaw), &decoded); err == nil {
        taskJournal = decoded
    }
}
```
The field accepts both a pre-decoded `[]interface{}` and a JSON string; the
handler tries each form in order.

### 3b: Re-marshal params map to pass to a handler expecting `json.RawMessage`

**`handlers.go` (lines 53–57) bridging map to a `json.RawMessage` handler:**
```go
argsJSON, err := json.Marshal(params)
if err != nil {
    return nil, fmt.Errorf("failed to marshal params: %w", err)
}
return handleGenerateConfigNative(ctx, argsJSON)
```

**`handlers_ai.go` (lines 271–274) map to typed struct:**
```go
paramsJSON, _ := json.Marshal(paramsMap)
if err := json.Unmarshal(paramsJSON, &params); err != nil {
    return nil, fmt.Errorf("failed to convert params: %w", err)
}
```
This is the standard way to convert a `map[string]interface{}` into a named
struct when that struct has JSON tags — marshal then unmarshal into the target
type.

**ParseTaskIDsFromParams in `task_workflow_common.go` (lines 34–44)** handles a
comma-separated string that might also be a JSON array:
```go
if ids, ok := params["task_ids"].(string); ok && ids != "" {
    var parsed []string
    if json.Unmarshal([]byte(ids), &parsed) == nil {
        for _, id := range parsed {
            add(id)
        }
    } else {
        for _, id := range strings.Split(ids, ",") {
            add(id)
        }
    }
}
```
Attempt JSON parse first; fall back to comma-split. This makes the field
tolerant of both `["T-1","T-2"]` and `"T-1,T-2"`.

---

## Schema Definitions and Their Relationship to Parsing

Tool schemas are registered in `registry_core.go`, `registry_ai.go`,
`registry_infra.go`, and `registry_misc.go` using `server.RegisterTool(name,
description, framework.ToolSchema{...}, handler)`.

Each property has a `"type"` key (`"string"`, `"boolean"`, `"number"`) and
optionally a `"default"`. The schema is informational for MCP clients — it is
not enforced by the Go runtime. The handler must perform its own validation.

Patterns to match schema types to parsing:

| Schema type | Preferred parsing |
|---|---|
| `"string"` | `ParamString` / `cast.ToString` / `RequireParam` |
| `"boolean"` | `ParamBool` / `cast.ToBool` / `.(bool)` assertion |
| `"number"` | `ParamInt` / `cast.ToInt` / `.(float64)` assertion |
| `"string"` + `"enum"` | `ParamEnum` / `RequireEnum` |
| `"string"` (JSON-encoded array/object) | Pattern 3a: `json.Unmarshal([]byte(v), &target)` |

Default values declared in the schema are applied via `framework.ApplyDefaults`
in `WrapHandler` (see `handlers_wrap.go`). Handlers that are called directly
(not through `WrapHandler`) must apply defaults manually.

---

## Best Practices

### Required parameters

Use `RequireParam` for string params that must be present:
```go
taskID, err := RequireParam(params, "task_id")
if err != nil {
    return nil, err // message: missing required parameter "task_id"
}
```

Or the inline cast pattern with an explicit error:
```go
taskID := cast.ToString(params["task_id"])
if taskID == "" {
    return nil, fmt.Errorf("action requires task_id")
}
```

### Optional parameters with defaults

```go
// String with default
format := ParamString(params, "output_format")
if format == "" {
    format = "text"
}

// Bool with default false (zero value — just use cast directly)
dryRun := cast.ToBool(params["dry_run"])

// Int with explicit default
limit := ParamInt(params, "limit", 17)
```

### Distinguishing "not provided" from "explicit false/zero"

```go
if HasKey(params, "dry_run") {
    dryRun = cast.ToBool(params["dry_run"])
} else {
    dryRun = false // caller did not provide the param
}
```

### Enum validation

Always validate enum values against an explicit allow-list; provide the list
in the error message:
```go
action, err := ParamEnum(params, "action",
    []string{"list", "create", "update", "delete"},
    "list",
)
if err != nil {
    return nil, err // message: invalid value "foo" for "action": must be one of list, create, update, delete
}
```

### Default output paths

Use the helpers in `params_helpers.go` rather than writing custom logic:
```go
outputPath := DefaultReportOutputPath(projectRoot, "REPORT.md", params)
// Returns params["output_path"] if set; else projectRoot/out/REPORT.md (generated) or projectRoot/docs/REPORT.md (user-facing)
```

---

## Common Pitfalls

### 1. Panicking single-value type assertion

```go
// WRONG — panics if key is missing or value is not a string
action := params["action"].(string)

// CORRECT — safe two-value form
action, _ := params["action"].(string)

// BETTER — use the helper, which also trims whitespace
action := ParamString(params, "action")
```

### 2. Nil map access

The params map is created by the framework before the handler is called, so it
is never nil in production. In tests, always initialize it:
```go
params := map[string]interface{}{}           // empty but non-nil
params := map[string]interface{}{"action": "list"}
```

### 3. JSON numbers always arrive as float64

`encoding/json` decodes all numbers to `float64` when the target is
`interface{}`. Do not assert `.(int)` on a JSON number:
```go
// WRONG — always fails for JSON numbers
count := params["limit"].(int)

// CORRECT
count := ParamInt(params, "limit", 17)
// or
if f, ok := params["limit"].(float64); ok {
    count = int(f)
}
```

### 4. Forgetting to trim whitespace on string params

LLM agents sometimes inject leading/trailing spaces. Always trim:
```go
// cast.ToString does NOT trim — wrap it or use ParamString
action := strings.TrimSpace(cast.ToString(params["action"]))
// or
action := ParamString(params, "action")   // trims internally
```

### 5. Missing default after WrapHandler short-circuits

`framework.ApplyDefaults` only fills in defaults for keys that are missing,
empty strings, or zero values. If your handler is called directly (not via
`WrapHandler`), apply defaults explicitly before branching on param values.

### 6. Mutating the shared params map

Handlers that forward a subset of params to sub-handlers should copy the map
rather than mutate the original to avoid surprising side-effects:
```go
subParams := make(map[string]interface{}, len(params))
for k, v := range params {
    subParams[k] = v
}
subParams["action"] = "update"
return handleTaskWorkflowUpdate(ctx, subParams)
```
