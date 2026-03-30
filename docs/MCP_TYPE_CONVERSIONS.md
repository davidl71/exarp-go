# MCP arguments, maps, and type conversions

**Last updated:** 2026-03-30  

This doc is the **starting map** for “type casting” work in exarp-go: where values cross boundaries (JSON, protobuf, `map[string]interface{}`, DB metadata), what Go types you actually get, and **where to start** hardening or refactoring.

It complements [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md) (what is implemented) and [PROTOBUF_INTEGRATION.md](PROTOBUF_INTEGRATION.md) (handler pattern).

---

## Summary table

| Layer | Typical input | Typical output | Mechanism | Primary code |
|-------|----------------|----------------|-----------|--------------|
| **MCP / CLI args** | `json.RawMessage` | `map[string]interface{}` **or** typed `proto.Message` | JSON object detected first → `json.Unmarshal` into map; else protobuf binary → `proto.Unmarshal` | `internal/framework/request.go` — `ParseRequest` |
| **Protobuf → legacy params** | `proto.Message` | `map[string]interface{}` | `protojson.Marshal` → `json.Unmarshal` into map; optional empty-string filter, array stringify, **float64→int** for named fields | `internal/framework/request.go` — `ProtobufToParams` |
| **Tool wrapper** | raw args | filled `params` + defaults | `WrapHandler`: parse → convert → `ApplyDefaults` → `NativeHandler` | `internal/tools/handlers_wrap.go` |
| **Safe param helpers** | `map[string]interface{}` | scalars, slices, enums | `spf13/cast` + trim: `ParamString`, `ParamBool`, `ParamInt`, `ParamIntOK`, `ParamFloat64`, `ParamFloat64OK`, `ParamStringSlice`, `RequireParam`, `ParamEnum` | `internal/tools/params_helpers.go` |
| **String normalization** | status/priority strings | canonical lowercase / title case | Lookup tables (not numeric casting) | `internal/tools/normalization.go` |
| **Per-tool proto glue** | args | `*Request` + `*RequestToParams` | `Parse*Request` + `ProtobufToOptions` (e.g. `ConvertFloat64ToInt`, `Float64ToIntFields`) | `internal/tools/protobuf_helpers.go` (+ split files) |
| **Task metadata (DB)** | `map` / proto | blob + format flag | `SerializeTaskMetadata`; protobuf vs JSON | `internal/database/tasks.go`, `internal/models/todo2_protobuf.go` |
| **Nested / dynamic JSON** | LLM or analysis `map` | Go fields | Widespread `v.(string)`, `v.(float64)`, `[]interface{}` iteration | `task_analysis_*.go`, `session_*.go`, `report_data.go`, etc. |

---

## Implemented (wave 1)

**As of 2026-03-30**, numeric and slice coercion helpers live in `internal/tools/params_helpers.go` and are covered by `params_helpers_test.go`:

- **`ParamInt` / `ParamIntOK`** — JSON `float64`, ints, numeric strings.
- **`ParamFloat64` / `ParamFloat64OK`** — same boundary, for fractional params (e.g. temperature).
- **`ParamStringSlice`** — JSON arrays, `[]interface{}`, comma-separated strings (via `cast.ToStringSlice`).
- **`ParamStringSliceTrimmed`** — same as `ParamStringSlice`, plus per-element trim and dropping empty strings.
- **`ParamStringSliceTrimmedCommaSeparated`** — string values split on commas (trim each token); arrays still use one element per array item (tags, deps, `recommended_tools`, ownership lists, `modified_task_ids`).

**Call sites migrated in this wave** (non-exhaustive): `task_analysis_deps.go`, `task_analysis_deps_analysis.go`, `task_analysis_shared.go`, `task_workflow_maintenance.go`, `task_workflow_agent.go` (claim/batch_claim `lease_minutes`, `count`, **`task_ids`**), `report.go`, `automation_scheduled.go`, `recommend.go`, `task_execute.go`, `context.go`, `fm_plan_execute.go`, `infer_task_progress.go`, `sampling_tool.go`, `estimation_shared_v2.go`, `ollama_native.go`, `prompt_tracking.go`, `memory_maint_utils.go`, `apple_foundation_helpers.go`.

**Wave 2 (params consistency):** `task_workflow_crud.go` / `parse*FromParams` and `parseStringSliceFromParams` → `ParamStringSliceTrimmed`; `task_workflow_create_ai.go` (tags/deps); `session_handoff.go` (`modified_task_ids`); `automation_native.go` (`action`, `cursor_agent_prompt`); `tool_catalog.go` (`action`, `tool_name`); `recommend.go` (model/workflow/advisor string fields).

**Still good next targets** for raw `.(string)` / nested maps: `prompt_tracking.go`, `fm_plan_execute.go`, `sampling_tool.go`, `estimation_shared_v2.go`, `task_execute.go`, `context.go`, `task_analysis_deps*.go`, `session_helpers_handoff.go`, `task_analysis_graph.go`, `report_data.go`, `ollama_native_handlers.go`, `todo2_json.go`.

---

## Where to start (recommended order)

1. **Understand the two inbound paths**  
   - **JSON path:** `encoding/json` decodes numbers into `interface{}` as **`float64`**, arrays as **`[]interface{}`**, objects as **`map[string]interface{}`**.  
   - **Protobuf path:** After `ProtobufToParams`, the map still comes from JSON unmarshaling, so numeric fields are still **`float64`** unless you pass **`ConvertFloat64ToInt`** and **`Float64ToIntFields`** for specific keys.

2. **Use `params_helpers.go` before scattered asserts**  
   Wave 1 added `ParamInt`, `ParamIntOK`, `ParamFloat64`, `ParamFloat64OK`, `ParamStringSlice` alongside `ParamString`, `ParamBool`, `RequireParam`, `ParamEnum`.  
   **Next incremental wins:** replace remaining `params["…"].(string)` / nested `map` chains in high-churn files (see **Implemented** and the list in step 3).

3. **Target files with many remaining assertions**  
   Re-grep `internal/tools` for `.(string)` / `.(float64)` / `.(int)` when planning the next wave; prioritize files listed under **Still good next targets** above.  
   Treat **tests** separately: they assert on constructed maps where `int` may already be correct.

4. **Watch for `.(int)` on JSON-backed maps**  
   Maps built in Go code may store `int`; **JSON-unmarshaled** maps almost never have `int` for numeric literals — use `float64` or `cast.ToInt64E` / a shared helper.

5. **Structured outputs from LLMs / aggregators**  
   Long term, prefer **`json.Unmarshal` into small structs** or **`mapstructure`/similar** for stable subtrees instead of deep `map[string]interface{}` chains.

---

## Related references

| Topic | Location |
|--------|----------|
| Parse + protobuf preference vs JSON object | `internal/framework/request.go` |
| Defaults merge (empty string / zero) | `internal/framework/request.go` — `ApplyDefaults`, `isZeroValue` |
| Example `*RequestToParams` options | `internal/tools/protobuf_helpers.go` — e.g. `MemoryRequestToParams` |
| Protobuf status / roadmap | [PROTOBUF_IMPLEMENTATION_STATUS.md](PROTOBUF_IMPLEMENTATION_STATUS.md) |
| Generic parser (sibling library, same idea) | `mcp-go-core/pkg/mcp/request/parser.go` — `ParseRequest` |

---

## Out of scope here

- **Aether / Rust** NATS and `prost` conversions — see that repo’s boundary docs.  
- **Cursor / MCP client** serialization — exarp-go receives JSON objects on the wire for typical tool calls; binary protobuf is supported for selected paths.
