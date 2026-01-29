# Backlog Execution Plan

**Generated:** 2026-01-29 01:10:57

**Backlog:** 425 tasks (Todo + In Progress)

**Waves:** 7 dependency levels

## Wave 0

| ID | Content | Priority |
|----|--------|----------|
| T-1768319001463 | Configuration System Epic | high |
| T-1768319224557 | Review 6 remaining standalone tasks for epic assignment o... | high |
| T-1768319355360 | Review dependencies between subtasks within epics | high |
| T-1768319664765 | Set priorities for all 7 epics and key subtasks | high |
| T-1768325400284 | Analyze current mcp-go-core integration status in devwisd... | high |
| T-1768325405127 | Complete logging migration by removing internal/logging d... | high |
| T-1768325715211 | Performance audit: Python bridge subprocess overhead | high |
| T-1768325728818 | Performance audit: File I/O caching | high |
| T-1768325734895 | Performance audit: Database query optimization | high |
| T-1768326545827 | Consolidate logging systems (mcp-go-core vs internal/logg... | high |
| T-1768326690159 | Migrate mcp-go-core logger to use slog (standard library) | high |
| T-1768326823408 | Investigate analyze_alignment JSON parsing error in git h... | high |
| T-105 | Document gotoHuman API/tools | medium |
| T-106 | Test basic approval request flow | medium |
| T-107 | Create approval request helper function | medium |
| T-108 | Enhance `task_workflow` tool with approval action | medium |
| T-109 | Add approval request when task moves to Review | medium |
| T-110 | Test approval workflow end-to-end | medium |
| T-111 | Auto-sync Review tasks with gotoHuman | medium |
| T-112 | Handle approval/rejection responses | medium |
| T-113 | Update Todo2 task status based on approval | medium |
| T-114 | Add approval history tracking | medium |
| T-115 | Add approval timeout handling | medium |
| T-116 | Add approval reminders | medium |
| T-117 | Add approval delegation | medium |
| T-118 | Add approval analytics | medium |
| T-119 | Test gotoHuman server loads in Cursor | medium |
| T-120 | List available gotoHuman tools | medium |
| T-121 | Test sending approval request | medium |
| T-122 | Test checking approval status | medium |
| T-123 | Test task moves to Review → approval request sent | medium |
| T-124 | Test human approves → task moves to Done | medium |
| T-125 | Test human rejects → task moves back to In Progress | medium |
| T-126 | Test approval timeout handling | medium |
| T-127 | Complete workflow: Task → Review → Approval → Done | medium |
| T-128 | Multiple concurrent approval requests | medium |
| T-129 | Approval cancellation | medium |
| T-130 | Approval delegation | medium |
| T-131 | List all available forms | medium |
| T-132 | Get schema for each form | medium |
| T-133 | Test basic approval request | medium |
| T-134 | Document form fields and usage | medium |
| T-135 | Create approval form mapping | medium |
| T-136 | Enhance `task_workflow` tool | medium |
| T-137 | Auto-send approval on Review state | medium |
| T-138 | Handle approval responses | medium |
| T-139 | Approval status checking | medium |
| T-140 | Approval timeout handling | medium |
| T-141 | Approval history tracking | medium |
| T-142 | Approval analytics | medium |
| T-143 | Set up HTTP server with SSE endpoint | medium |
| T-144 | Configure OAuth if needed | medium |
| T-145 | Use `url` and `transport` in `mcp.json` | medium |
| T-146 | Handle authentication tokens | medium |
| T-147 | Support multiple concurrent connections | medium |
| T-148 | Go handler created in `handlers.go` | medium |
| T-149 | Tool registered in `registry.go` (appropriate batch) | medium |
| T-150 | Tool added to `bridge/execute_tool.py` | medium |
| T-151 | All parameters mapped correctly | medium |
| T-152 | Default values handled | medium |
| T-153 | Error handling implemented | medium |
| T-154 | Integration test written | medium |
| T-155 | Documentation updated | medium |
| T-156 | Tool tested via MCP interface | medium |
| T-157 | Analyze `project-management-automation/project_management... | medium |
| T-158 | Compare with `exarp-go/internal/tools/registry.go` | medium |
| T-159 | Identify FastMCP dependencies vs standalone functions | medium |
| T-160 | Map tool dependencies | medium |
| T-161 | Create priority matrix (high/medium/low) | medium |
| T-162 | Full integration test coverage | medium |
| T-163 | MCP interface validation | medium |
| T-164 | Tool/prompt/resource functionality | medium |
| T-165 | Error handling validation | medium |
| T-166 | Performance benchmarks (if applicable) | medium |
| T-167 | Update README.md | medium |
| T-168 | Create MIGRATION_COMPLETE.md | medium |
| T-169 | Create MIGRATION_VALIDATION.md | medium |
| T-170 | Update project-management-automation README (deprecation) | medium |
| T-171 | Remove unused code | medium |
| T-172 | Update comments | medium |
| T-173 | Code organization | medium |
| T-174 | Deprecation notices | medium |
| T-175 | Review and approve migration plan | medium |
| T-176 | Set up test environment | medium |
| T-1768170886572 | Implement automation tool sprint action in native Go | medium |
| T-1768253987147 | Update devwisdom-go to use mcp-go-core | medium |
| T-1768312778714 | Review local commits to identify relevant changes | medium |
| T-1768318349855 | Create protobuf schemas for tool handler arguments | medium |
| T-1768318359061 | Update tool handlers to use protobuf request messages | medium |
| T-1768318368598 | Migrate Python bridge to protobuf communication | medium |
| T-1768318377076 | Migrate memory system to protobuf file format | medium |
| T-1768318391643 | Migrate report and scorecard to protobuf | medium |
| T-1768318471621 | Testing & Validation Epic | medium |
| T-1768318471623 | Python Code Cleanup Epic | medium |
| T-1768319001461 | Protobuf Integration Epic | medium |
| T-1768320725711 | Test Planning Link | medium |
| T-1768320773738 | Test Validation - Valid Path | medium |
| T-1768321828916 | Analyze task-related code for simplification opportunities | medium |
| T-1768324162836 | Fix missing version field in SQLite schema | medium |
| T-1768324344223 | Fix schema versioning and duplicate migration versions | medium |
| T-1768325408714 | Integrate mcp-go-core response formatting utilities | medium |
| T-1768325421830 | Evaluate migrating ConsultationLogger to mcp-go-core or k... | medium |
| T-1768325426665 | Identify code that can be migrated from devwisdom-go to m... | medium |
| T-1768325741814 | Performance audit: Profile-Guided Optimization (PGO) | medium |
| T-1768325747897 | Performance audit: Memory allocation optimization | medium |
| T-1768325753706 | Performance audit: Concurrent tool execution | medium |
| T-1768327631413 | Migrate CLI to use mcp-go-core CLI utilities | medium |
| T-1769531781411 | Remove check_attribution Python fallback | medium | ✅ Done (2026-01-27) |
| T-1769531785535 | Remove session Python fallback | medium | ✅ Done (2026-01-27) |
| T-1769531787256 | Remove memory_maint Python fallback | medium | ✅ Done (2026-01-27) |
| T-1769531792537 | Update migration docs for 4 removed Python fallbacks | medium | ✅ Done (2026-01-29; PYTHON_REMAINING_ANALYSIS, BACKLOG, etc.) |
| T-177 | Create baseline performance metrics | medium |
| T-178 | Backup current code | medium |
| T-179 | Create `internal/tools/statistics.go` | medium |
| T-180 | Implement helper functions | medium |
| T-181 | Write unit tests | medium |
| T-182 | Verify tests pass | medium |
| T-183 | Create `bridge/statistics_bridge.py` | medium |
| T-184 | Implement bridge functions | medium |
| T-185 | Test bridge integration | medium |
| T-186 | Document bridge usage | medium |
| T-187 | Migrate `get_statistics()` function | medium |
| T-188 | Replace error analysis calculations | medium |
| T-189 | Update exception handling | medium |
| T-190 | Run integration tests | medium |
| T-191 | Compare Python vs Go results | medium |
| T-192 | Migrate `analyze_estimation_accuracy()` | medium |
| T-193 | Migrate `_analyze_by_tag()` | medium |
| T-194 | Migrate `_analyze_by_priority()` | medium |
| T-195 | Migrate `_identify_patterns()` | medium |
| T-196 | Verify grouped analysis results | medium |
| T-197 | Migrate `_apply_learned_adjustments()` | medium |
| T-198 | Verify adjustment calculations | medium |
| T-199 | All tests pass | medium |
| T-200 | Performance benchmarks meet targets | medium |
| T-202 | Production deployment | medium |
| T-203 | Monitor for issues | medium |
| T-204 | Create model router component | medium |
| T-205 | Implement MLX handler (via Python bridge) | medium |
| T-206 | Implement Ollama HTTP client | medium |
| T-207 | Add model selection logic | medium |
| T-208 | Implement task analyzer | medium |
| T-209 | Create breakdown prompt templates | medium |
| T-210 | Add breakdown result parser | medium |
| T-211 | Integrate with Todo2 task creation | medium |
| T-212 | Implement task complexity classifier | medium |
| T-213 | Create execution prompt templates | medium |
| T-214 | Add change application logic | medium |
| T-215 | Integrate with Todo2 execution flow | medium |
| T-216 | Implement prompt analyzer | medium |
| T-217 | Create optimization prompt templates | medium |
| T-218 | Add iterative refinement loop | medium |
| T-219 | Integrate with task creation workflow | medium |
| T-220 | Unit tests for all components | medium |
| T-221 | Integration tests with real models | medium |
| T-222 | Performance benchmarks | medium |
| T-223 | User documentation | medium |
| T-224 | Create research result aggregator | medium |
| T-225 | Implement conflict detection | medium |
| T-226 | Add agent performance tracking | medium |
| T-227 | Analyze task dependencies | medium |
| T-228 | Create parallel execution framework | medium |
| T-229 | Implement shared state management | medium |
| T-230 | Design agent pool system | medium |
| T-231 | Implement load balancing | medium |
| T-232 | Add monitoring dashboard | medium |
| T-233 | Create native Go implementation file | medium |
| T-234 | Implement tool logic (or hybrid pattern) | medium |
| T-235 | Update handler in `internal/tools/handlers.go` | medium |
| T-236 | Register tool in `internal/tools/registry.go` (if new reg... | medium |
| T-237 | Write integration tests | medium |
| T-238 | Update documentation | medium |
| T-239 | Keep Python bridge as fallback (if hybrid) | medium |
| T-240 | Verify tool works via MCP interface | medium |
| T-241 | Update migration status docs | medium |
| T-242 | All medium complexity tools (Phase 2) migrated (hybrid wh... | medium |
| T-243 | Complex tools evaluated and migrated where practical (Pha... | medium |
| T-244 | All resources migrated to native Go (Phase 4) - 0/6 compl... | medium |
| T-245 | Prompts evaluated (Phase 5) - migrate if beneficial, othe... | medium |
| T-246 | All migrated tools have unit and integration tests | medium |
| T-247 | Performance improvements measured and documented | medium |
| T-248 | All 4 remaining actions implemented in native Go | medium |
| T-249 | Handler routes to native implementation for all actions | medium |
| T-250 | Python bridge fallback maintained for error cases | medium |
| T-251 | Unit tests for each action | medium |
| T-252 | Integration tests via MCP interface | medium |
| T-253 | `orphans` action implemented in native Go | medium |
| T-254 | Uses task_analysis for dependency checking | medium |
| T-255 | Integration tests pass | medium |
| T-256 | All 3 remaining actions implemented | medium |
| T-258 | Tests pass | medium |
| T-260 | All medium complexity tools migrated (3B) | medium |
| T-261 | High complexity tools evaluated and migrated where practi... | medium |
| T-262 | Automation tool documented as intentional Python bridge r... | medium |
| T-263 | All migrated tools have tests | medium |
| T-264 | Hybrid patterns documented | medium |
| T-265 | No regressions in functionality | medium |
| T-266 | Create `internal/security/path.go` | medium |
| T-267 | Implement `ValidatePath()` function | medium |
| T-268 | Update `linting.go` to use path validation | medium |
| T-269 | Update `scorecard_go.go` to use path validation | medium |
| T-270 | Update `bridge/python.go` to use path validation | medium |
| T-271 | Add tests for path boundary enforcement | medium |
| T-272 | Create `internal/security/ratelimit.go` | medium |
| T-273 | Implement sliding window rate limiter | medium |
| T-274 | Add middleware to framework | medium |
| T-275 | Configure default limits | medium |
| T-276 | Add per-tool limits | medium |
| T-277 | Add tests for rate limiting | medium |
| T-278 | Design access control model | medium |
| T-279 | Implement permission system | medium |
| T-280 | Add configuration support | medium |
| T-281 | Update tool/resource handlers | medium |
| T-282 | Add audit logging | medium |
| T-283 | Test directory traversal prevention (`../../../etc/passwd`) | medium |
| T-284 | Test symlink handling | medium |
| T-285 | Test absolute paths outside project root | medium |
| T-286 | Test relative paths that escape root | medium |
| T-287 | Test edge cases (empty paths, special characters) | medium |
| T-288 | Test request limit enforcement | medium |
| T-289 | Test sliding window accuracy | medium |
| T-290 | Test per-tool limits | medium |
| T-291 | Test concurrent requests | medium |
| T-292 | Test limit reset behavior | medium |
| T-293 | Test tool access permissions | medium |
| T-294 | Test resource access permissions | medium |
| T-295 | Test deny list functionality | medium |
| T-296 | Test allow list functionality | medium |
| T-297 | Backup `.todo2/state.todo2.json` | medium |
| T-298 | Verify JSON structure is valid | medium |
| T-299 | Count tasks in JSON file | medium |
| T-300 | Test database schema creation | medium |
| T-301 | Review migration plan | medium |
| T-302 | Run migration tool (dry-run first) | medium |
| T-303 | Verify all tasks migrated | medium |
| T-304 | Validate data integrity | medium |
| T-305 | Test query operations | medium |
| T-306 | Test concurrent access | medium |
| T-307 | Update all code to use database | medium |
| T-308 | Remove file locking code | medium |
| T-309 | Remove JSON cache for tasks | medium |
| T-310 | Update tests | medium |
| T-311 | Restore JSON backup | medium |
| T-312 | Remove database file | medium |
| T-313 | Revert code changes | medium |
| T-314 | Verify system works with JSON | medium |
| T-315 | Background goroutine for periodic cleanup | medium |
| T-316 | CLI command for lock status monitoring | medium |
| T-317 | Scheduled cleanup via cron | medium |
| T-318 | Heartbeat mechanism (agents report alive) | medium |
| T-319 | Process monitoring (verify agent process exists) | medium |
| T-320 | Lock statistics and metrics | medium |
| T-321 | Alerting system for stale lock patterns | medium |
| T-322 | Review status updates category | medium |
| T-323 | Check for broken references | medium |
| T-324 | Update this document if needed | medium |
| T-325 | Review implementation summaries | medium |
| T-326 | Check retention dates | medium |
| T-327 | Delete expired files | medium |
| T-328 | Review all categories | medium |
| T-329 | Update retention policy if needed | medium |
| T-330 | Archive this policy document if superseded | medium |
| T-331 | Create `go.mod` with Go SDK dependencies | medium |
| T-332 | Create `cmd/server/main.go` - Basic server skeleton | medium |
| T-333 | Create `internal/framework/server.go` - Framework interfaces | medium |
| T-334 | Create `internal/framework/factory.go` - Framework factory | medium |
| T-335 | Create `internal/framework/adapters/gosdk/adapter.go` - G... | medium |
| T-336 | Create `internal/bridge/python.go` - Python bridge mechanism | medium |
| T-337 | Create `bridge/execute_tool.py` - Python tool executor | medium |
| T-338 | Create `internal/config/config.go` - Configuration manage... | medium |
| T-339 | Test basic server startup | medium |
| T-340 | Set up Go project structure | medium |
| T-341 | Install and configure Go SDK | medium |
| T-342 | Create basic STDIO server skeleton | medium |
| T-343 | Implement Python bridge mechanism | medium |
| T-344 | Test basic tool execution via bridge | medium |
| T-345 | Migrate simple tools (no MLX dependency) | medium |
| T-346 | Implement tool registration system | medium |
| T-347 | Test tool execution | medium |
| T-348 | Verify Cursor integration | medium |
| T-349 | Migrate medium-complexity tools | medium |
| T-350 | Implement prompt system | medium |
| T-351 | Test prompts in Cursor | medium |
| T-352 | Migrate advanced tools | medium |
| T-353 | Implement resource handlers | medium |
| T-354 | Test resources in Cursor | medium |
| T-355 | Handle MLX tools (`ollama`, `mlx`) | medium |
| T-356 | Implement MLX bridge or alternative | medium |
| T-357 | Test MLX tools | medium |
| T-358 | Comprehensive testing | medium |
| T-359 | Performance optimization | medium |
| T-360 | Documentation | medium |
| T-361 | Deployment preparation | medium |
| T-362 | Create Go project structure | medium |
| T-363 | Set up Go module and dependencies | medium |
| T-364 | Implement basic STDIO server | medium |
| T-365 | Create Python bridge mechanism | medium |
| T-366 | Test bridge with one simple tool | medium |
| T-367 | Migrate `analyze_alignment` | medium |
| T-368 | Migrate `generate_config` | medium |
| T-369 | Migrate `health` | medium |
| T-373 | Test all Batch 1 tools in Cursor | medium |
| T-374 | Migrate `memory` | medium |
| T-375 | Migrate `memory_maint` | medium |
| T-376 | Migrate `report` | medium |
| T-377 | Migrate `security` | medium |
| T-378 | Migrate `task_analysis` | medium |
| T-379 | Migrate `task_discovery` | medium |
| T-380 | Migrate `task_workflow` | medium |
| T-381 | Migrate `testing` | medium |
| T-382 | Migrate `automation` | medium |
| T-383 | Migrate `tool_catalog` | medium |
| T-384 | Migrate `workflow_mode` | medium |
| T-385 | Migrate `lint` | medium |
| T-386 | Migrate `estimation` | medium |
| T-387 | Migrate `git_tools` | medium |
| T-388 | Migrate `session` | medium |
| T-389 | Migrate `infer_session_mode` | medium |
| T-390 | Handle `ollama` (via HTTP or bridge) | medium |
| T-391 | Handle `mlx` (via bridge) | medium |
| T-392 | Update README | medium |
| T-393 | Create migration guide | medium |
| T-394 | Deploy to production | medium |
| T-395 | Code review of existing patterns (CodeLlama) | medium |
| T-396 | Library documentation (Context7) | medium |
| T-397 | Logical decomposition (Tractatus) | medium |
| T-398 | Latest patterns (Web Search) | medium |
| T-399 | Verify `bin/exarp-go` exists and is executable | medium |
| T-400 | Test Go server: `./bin/exarp-go` responds to MCP requests | medium |
| T-401 | Verify all 24 tools work via Go server | medium |
| T-402 | Verify all 8 prompts work via Go server | medium |
| T-403 | Verify all 6 resources work via Go server | medium |
| T-404 | Check MCP config uses Go binary (not Python server) | medium |
| T-405 | Backup old files (optional): `mkdir -p docs/archive && mv... | medium |
| T-53 | Ollama uses "model" not "name" | medium |
| T-64 | Tool may not be available on non-Apple platforms (conditi... | medium |
| T-65 | Requires Swift bridge to be built and Apple Intelligence ... | medium |
| T-66 | Requires Swift bridge to be built | medium |
| T-67 | Requires Swift bridge | medium |
| T-68 | apple_foundation_models is conditionally compiled | medium |
| T-69 | Add `version`, `assignee`, `assigned_at`, `lock_until` fi... | medium |
| T-70 | Create migration script | medium |
| T-71 | Implement `ClaimTaskForAgent()` with SELECT FOR UPDATE | medium |
| T-72 | Implement `ReleaseTask()` | medium |
| T-73 | Update `UpdateTask()` to use version checks | medium |
| T-74 | Add Python wrapper for claim operations | medium |
| T-75 | Implement lease renewal mechanism | medium |
| T-76 | Add dead agent cleanup job | medium |
| T-77 | Update task workflow tools to use atomic claims | medium |
| T-78 | Fine-grained file locking (for Git operations) | medium |
| T-79 | Agent heartbeat system | medium |
| T-80 | Lock timeout monitoring/alerts | medium |
| T-81 | Performance metrics (lock contention, wait times) | medium |
| T-82 | Memory system migrated (blocks Phase 4) | medium |
| T-83 | Session tool migrated (high value) | medium |
| T-84 | Phase 4 resources migrated (after memory) | medium |
| T-85 | All critical path items complete | medium |
| T-86 | Open Settings (Cmd+, / Ctrl+,) | medium |
| T-87 | Navigate to Beta Features | medium |
| T-88 | Enable "Cursor: Background Agents" | medium |
| T-89 | Verify Pro plan subscription | medium |
| T-90 | Restart Cursor | medium |
| T-91 | Press `Ctrl+E` / `Cmd+E` to open sidebar | medium |
| T-92 | Verify Background Agents sidebar appears | medium |
| T-93 | Install Remote - SSH extension in Cursor | medium |
| T-94 | Configure SSH config (`~/.ssh/config`) | medium |
| T-95 | Test SSH connection: `ssh davids-mac-mini.tailf62197.ts.net` | medium |
| T-96 | Connect via Remote - SSH (F1 → "Remote-SSH: Connect to ... | medium |
| T-1768318471624 | Enhancements Epic | low |
| T-1768325006421 | Verify tasks 3-5 completion status | low |
| T-1768327860912 | Refactor CLI subcommand parsing to use mcp-go-core ParseArgs | low |
| T-1768327866933 | Use mcp-go-core DetectMode for execution mode detection | low |
| T-1768327872868 | Use mcp-go-core IsTTYFile for colored output detection | low |

## Wave 1

| ID | Content | Priority |
|----|--------|----------|
| T-1768326701962 | Migrate exarp-go CLI to use mcp-go-core logger | high |
| T-1768170876574 | Verify estimation tool native implementation completeness | medium |
| T-1768223765685 | Analyze Python code for safe removal | medium |
| T-1768223926189 | Create Python code removal plan document | medium |
| T-1768251813711 | Add Tests for Factory and Config Packages | medium |
| T-1768251816882 | Create Error Types for Better Error Handling | medium |
| T-1768251818361 | Add Builder Pattern for Server Configuration | medium |
| T-1768251821268 | Remove Empty Directories or Add Placeholders | medium |
| T-1768251822699 | Add Package-Level Documentation | medium |
| T-1768251824236 | Consider Options Pattern for Adapter Construction | medium |
| T-1768253988999 | Implement Logger Integration in Adapter | medium |
| T-1768253990892 | Implement Middleware Support in Adapter | medium |
| T-1768253992311 | Complete SSE Transport Implementation | medium |
| T-1768253994622 | Implement CLI Utilities | medium |
| T-1768312586170 | Pull latest changes from git repository | medium |
| T-1768313222082 | Fix Todo2 sync logic to properly handle database save errors | medium |
| T-1768316808114 | Create protobuf schemas for high-priority items | medium |
| T-1768316817909 | Set up protobuf build tooling and update Ansible | medium |
| T-1768316828486 | Implement Phase 1: Task Management protobuf migration | medium |
| T-1768317335877 | Run migration to add protobuf columns to database | medium |
| T-1768317405631 | Test protobuf serialization with real tasks | medium |
| T-1768317407961 | Measure performance improvements for protobuf serialization | medium |
| T-1768317410079 | Update ListTasks and other database operations to use pro... | medium |
| T-1768317520843 | T1.5.1: Create Protobuf Conversion Layer | medium |
| T-1768317889365 | ate the `analyze_alignment` tool from Python to Go server... | medium |
| T-1768317926143 | Generate a comprehensive project scorecard showing health... | medium |
| T-1768317926144 | Investigate cspell configuration in exarp-go project and ... | medium |
| T-1768317926147 | ate 6 simple tools to Go server using Python bridge, impl... | medium |
| T-1768317926149 | le MLX tools (ollama, mlx) with appropriate integration s... | medium |
| T-1768317926150 | rehensive testing, performance optimization, and complete... | medium |
| T-1768317926158 | MLX-enhanced time estimates and improve task clarity usin... | medium |
| T-1768317926161 | te comprehensive documentation for the parallel migration... | medium |
| T-1768317926167 | rmine if Go linting functionality can be implemented dire... | medium |
| T-1768317926168 | ace Python bridge dependency for lint tool with native Go... | medium |
| T-1768317926169 | tify and document any leftover Python code that should ha... | medium |
| T-1768317926171 | ve leftover Python code files identified in audit.

📋 ... | medium |
| T-1768317926177 | gn comprehensive migration strategy for remaining Python ... | medium |
| T-1768317926181 | lete testing, validation, and cleanup of migration from p... | medium |
| T-1768317926183 | Test Task with Database Estimation | medium |
| T-1768317926184 | Migrate session tool to native Go implementation | medium |
| T-1768317926185 | Complete context tool native Go migration (batch action) | medium |
| T-1768317926186 | Migrate ollama tool to native Go HTTP client | medium |
| T-1768317926187 | Complete memory_maint tool native Go migration | medium |
| T-1768317926188 | Migrate prompt_tracking tool to native Go | medium |

## Wave 2

| ID | Content | Priority |
|----|--------|----------|
| T-1768251811938 | Extract Type Conversion Helpers | medium |
| T-1768251815345 | Extract Context Validation Helper | medium |
| T-1768251819753 | Extract Request Validation Helpers | medium |
| T-1768317530753 | T1.5.2: Add Protobuf Format Support to Loader | medium |

## Wave 3

| ID | Content | Priority |
|----|--------|----------|
| T-1768317539591 | T1.5.3: Add Protobuf CLI Commands | medium |

## Wave 4

| ID | Content | Priority |
|----|--------|----------|
| T-1768317546936 | T1.5.4: Add Schema Synchronization Validation | medium |

## Wave 5

| ID | Content | Priority |
|----|--------|----------|
| T-1768317554758 | T1.5.5: Update Documentation for Protobuf Integration | medium |

## Wave 7

| ID | Content | Priority |
|----|--------|----------|
| T-1768268676474 | Phase 1.7: Update Documentation | medium |

## Full order

T-1768319001463, T-1768319224557, T-1768319355360, T-1768319664765, T-1768325400284, T-1768325405127, T-1768325715211, T-1768325728818, T-1768325734895, T-1768326545827, T-1768326690159, T-1768326823408, T-105, T-106, T-107, T-108, T-109, T-110, T-111, T-112, T-113, T-114, T-115, T-116, T-117, T-118, T-119, T-120, T-121, T-122, T-123, T-124, T-125, T-126, T-127, T-128, T-129, T-130, T-131, T-132, T-133, T-134, T-135, T-136, T-137, T-138, T-139, T-140, T-141, T-142, T-143, T-144, T-145, T-146, T-147, T-148, T-149, T-150, T-151, T-152, T-153, T-154, T-155, T-156, T-157, T-158, T-159, T-160, T-161, T-162, T-163, T-164, T-165, T-166, T-167, T-168, T-169, T-170, T-171, T-172, T-173, T-174, T-175, T-176, T-1768170886572, T-1768253987147, T-1768312778714, T-1768318349855, T-1768318359061, T-1768318368598, T-1768318377076, T-1768318391643, T-1768318471621, T-1768318471623, T-1768319001461, T-1768320725711, T-1768320773738, T-1768321828916, T-1768324162836, T-1768324344223, T-1768325408714, T-1768325421830, T-1768325426665, T-1768325741814, T-1768325747897, T-1768325753706, T-1768327631413, T-1769531781411, T-1769531785535, T-1769531787256, T-1769531792537, T-177, T-178, T-179, T-180, T-181, T-182, T-183, T-184, T-185, T-186, T-187, T-188, T-189, T-190, T-191, T-192, T-193, T-194, T-195, T-196, T-197, T-198, T-199, T-200, T-202, T-203, T-204, T-205, T-206, T-207, T-208, T-209, T-210, T-211, T-212, T-213, T-214, T-215, T-216, T-217, T-218, T-219, T-220, T-221, T-222, T-223, T-224, T-225, T-226, T-227, T-228, T-229, T-230, T-231, T-232, T-233, T-234, T-235, T-236, T-237, T-238, T-239, T-240, T-241, T-242, T-243, T-244, T-245, T-246, T-247, T-248, T-249, T-250, T-251, T-252, T-253, T-254, T-255, T-256, T-258, T-260, T-261, T-262, T-263, T-264, T-265, T-266, T-267, T-268, T-269, T-270, T-271, T-272, T-273, T-274, T-275, T-276, T-277, T-278, T-279, T-280, T-281, T-282, T-283, T-284, T-285, T-286, T-287, T-288, T-289, T-290, T-291, T-292, T-293, T-294, T-295, T-296, T-297, T-298, T-299, T-300, T-301, T-302, T-303, T-304, T-305, T-306, T-307, T-308, T-309, T-310, T-311, T-312, T-313, T-314, T-315, T-316, T-317, T-318, T-319, T-320, T-321, T-322, T-323, T-324, T-325, T-326, T-327, T-328, T-329, T-330, T-331, T-332, T-333, T-334, T-335, T-336, T-337, T-338, T-339, T-340, T-341, T-342, T-343, T-344, T-345, T-346, T-347, T-348, T-349, T-350, T-351, T-352, T-353, T-354, T-355, T-356, T-357, T-358, T-359, T-360, T-361, T-362, T-363, T-364, T-365, T-366, T-367, T-368, T-369, T-373, T-374, T-375, T-376, T-377, T-378, T-379, T-380, T-381, T-382, T-383, T-384, T-385, T-386, T-387, T-388, T-389, T-390, T-391, T-392, T-393, T-394, T-395, T-396, T-397, T-398, T-399, T-400, T-401, T-402, T-403, T-404, T-405, T-53, T-64, T-65, T-66, T-67, T-68, T-69, T-70, T-71, T-72, T-73, T-74, T-75, T-76, T-77, T-78, T-79, T-80, T-81, T-82, T-83, T-84, T-85, T-86, T-87, T-88, T-89, T-90, T-91, T-92, T-93, T-94, T-95, T-96, T-1768318471624, T-1768325006421, T-1768327860912, T-1768327866933, T-1768327872868, T-1768326701962, T-1768170876574, T-1768223765685, T-1768223926189, T-1768251813711, T-1768251816882, T-1768251818361, T-1768251821268, T-1768251822699, T-1768251824236, T-1768253988999, T-1768253990892, T-1768253992311, T-1768253994622, T-1768312586170, T-1768313222082, T-1768316808114, T-1768316817909, T-1768316828486, T-1768317335877, T-1768317405631, T-1768317407961, T-1768317410079, T-1768317520843, T-1768317889365, T-1768317926143, T-1768317926144, T-1768317926147, T-1768317926149, T-1768317926150, T-1768317926158, T-1768317926161, T-1768317926167, T-1768317926168, T-1768317926169, T-1768317926171, T-1768317926177, T-1768317926181, T-1768317926183, T-1768317926184, T-1768317926185, T-1768317926186, T-1768317926187, T-1768317926188, T-1768251811938, T-1768251815345, T-1768251819753, T-1768317530753, T-1768317539591, T-1768317546936, T-1768317554758, T-1768268676474
