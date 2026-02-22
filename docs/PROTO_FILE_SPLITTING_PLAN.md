# Proto File Splitting Plan

**Goal:** Split `proto/tools.proto` into multiple smaller `.proto` files for better context fit, maintainability, and clearer ownership—without changing Go imports or handler code.

**Current state:** One large `tools.proto` (~815 lines, 80+ messages) generates a single `tools.pb.go` (~6k lines). All code imports `proto` and uses types like `proto.MemoryRequest`, `proto.TaskWorkflowRequest`, etc.

---

## Strategy: Same Package, Multiple Files

- **Keep a single Go package:** `package exarp.tools` and `option go_package = "github.com/davidl71/exarp-go/proto"` in every new file.
- **Result:** protoc produces multiple `.pb.go` files (e.g. `tools_common.pb.go`, `tools_memory.pb.go`) all in package `proto`. No changes needed in `internal/tools` or `protobuf_helpers.go`.
- **Shared types:** Put types used by more than one domain in a common file and have other files `import` it.

---

## Proposed File Layout

| File | Contents | Approx. lines |
|------|----------|---------------|
| `tools_common.proto` | Shared types: `AIInsights`, `Metrics`, `BriefingQuote`, `BriefingData`, `PlanningSnippet`, `ProjectInfo`, `HealthData`, `CodebaseMetrics`, `TaskMetrics`, `ProjectPhase`, `RiskOrBlocker`, `NextAction`, `ProjectOverviewData`, `ScorecardData` (used by Report and others) | ~90 |
| `tools_memory.proto` | `Memory`, `MemoryRequest`, `MemoryResponse` | ~45 |
| `tools_context.proto` | `ContextItem`, `ContextRequest`, `ContextResponse`, `ContextSummary` | ~35 |
| `tools_report.proto` | `ReportRequest`, `ReportResponse` (imports tools_common for AIInsights, etc.) | ~25 |
| `tools_task.proto` | `TaskAnalysisRequest/Response`, `TaskWorkflowRequest/Response`, `TaskSummary`, `SyncResults`, `TaskDiscoveryRequest`, `AnalyzeAlignmentRequest`, `AutomationRequest`, `AutomationResponse`, `EstimationResult`, `HealthReport`, `InferTaskProgressResponse` | ~160 |
| `tools_testing.proto` | `TestingRequest`, `TestingResponse` | ~25 |
| `tools_config.proto` | `GenerateConfigRequest` | ~20 |
| `tools_health.proto` | `HealthRequest` | ~25 |
| `tools_security_lint.proto` | `SecurityRequest`, `LintRequest` | ~40 |
| `tools_estimation_git.proto` | `EstimationRequest`, `GitToolsRequest`, `GitToolsResponse` | ~55 |
| `tools_session.proto` | `SessionRequest`, `SessionDetection`, `SessionAgentContext`, `SessionWorkflow`, `LockCleanupReport`, `SessionPrimeResult`, `SessionHandoffResult` | ~95 |
| `tools_misc.proto` | `WorkflowModeRequest`, `SetupHooksRequest`, `CheckAttributionRequest`, `AddExternalToolHintsRequest`, `MemoryMaintRequest`, `ToolCatalogRequest` | ~75 |
| `tools_llm.proto` | `OllamaRequest`, `MLXRequest`, `PromptTrackingRequest`, `RecommendRequest`, `InferSessionModeRequest`, `ContextBudgetRequest` | ~95 |

Exact message-to-file mapping can be tuned (e.g. merge small files into `tools_misc.proto` or split task further).

---

## Import Rules

- **tools_common.proto:** No import of other tools_*.proto. Contains only shared message definitions.
- **tools_report.proto:** `import "proto/tools_common.proto";` and use e.g. `exarp.tools.AIInsights` (same package, so just `AIInsights` in that file).
- **Other tools_*.proto:** No cross-imports unless a type is shared (then move to tools_common and import it).

Protobuf same-package imports: use `import "proto/tools_common.proto";` and keep `package exarp.tools` in both. Generated Go stays in one package.

---

## Build and Tooling

1. **Makefile**  
   - In the `proto` target, replace `proto/tools.proto` with the list of new files, e.g.  
     `proto/tools_common.proto proto/tools_memory.proto proto/tools_context.proto ...`  
   - Order: list `tools_common.proto` first so it’s generated before files that depend on it (protoc accepts any order for same-package files; source_relative puts each in its own `.pb.go`).

2. **proto-check**  
   - Already runs over `proto/*.proto`; no change if new files live under `proto/`.

3. **proto-clean**  
   - Remove generated `proto/tools_*.pb.go` (or add an explicit list). Do **not** remove `proto/todo2.pb.go`, `proto/config.pb.go`, `proto/bridge.pb.go`.

4. **buf**  
   - If using `buf generate`, ensure the proto module includes all new `proto/*.proto` files (default inclusion by path is usually sufficient).

---

## Migration Steps

1. Create `tools_common.proto` with shared types only; generate and run tests.
2. Create one domain file (e.g. `tools_memory.proto`), move messages from `tools.proto`, delete those from `tools.proto`, run `make proto` and tests.
3. Repeat for each domain until `tools.proto` is empty; then remove `tools.proto` and add any remaining messages to the appropriate split file.
4. Update `proto/README.md` and `docs/PROTOBUF_USAGE.md` to describe the split layout.

---

## What Stays the Same

- **Go imports:** All handlers keep using `proto.MemoryRequest`, `proto.ReportRequest`, etc.
- **Handler and helper code:** No changes in `internal/tools/protobuf_helpers.go` or tool handlers.
- **Package and wire format:** Same package name and wire format; no API/version change.

---

## References

- Current schemas: `proto/tools.proto`, `proto/README.md`
- Build: `make proto`, `make proto-check`, `Makefile` (proto targets)
- Usage: `docs/PROTOBUF_USAGE.md`, `docs/PROTOBUF_IMPLEMENTATION_STATUS.md`
