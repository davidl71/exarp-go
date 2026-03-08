# exarp-go Implementation Plan

**Generated**: 2026-03-08  
**Total Tasks**: 509 (153 Todo, 356 Done)  
**Scope**: Remaining work after cleanup and architectural unification

---

## Executive Summary

After recent MCP/CLI drift fixes and cleanup of wrong-project tasks, exarp-go has **153 remaining tasks** organized into clear implementation tracks:

1. **llamacpp Integration** - 34 high-priority tasks (Mac/Metal GPU support)
2. **Compatibility & Polish** - Language-neutral tooling
3. **TUI Enhancements** - Better task management UX
4. **Test Coverage** - Fill remaining gaps
5. **Documentation** - User-facing guides

**Critical Path**: llamacpp integration (Waves 1-7)  
**Quick Wins**: OpenCode flags (--quiet, --json), TUI shortcuts

---

## Track 1: llamacpp Integration (Mac/Metal) 🔥

**Priority**: HIGH  
**Timeline**: 2-3 weeks  
**Value**: Local LLM support with GPU acceleration for Mac users

### Wave 1: Build Foundation (Week 1, Days 1-2)
**Goal**: Get llamacpp building on macOS/arm64

- [ ] T-1771367222946 - Test go-skynet/go-llama.cpp build on macOS/arm64 (HIGH)
- [ ] T-1771367262618 - Add go-skynet/go-llama.cpp dependency to go.mod (HIGH)
- [ ] T-1771367292653 - Update Makefile with libbinding.a build target (HIGH)
- [ ] T-1771367294624 - Add build-llamacpp make target (MEDIUM)
- [ ] T-1771367296219 - Document build requirements for llama.cpp (LOW)
- [ ] T-1771367297447 - Create build script for llama.cpp submodule (MEDIUM)

**Acceptance**: `make build-llamacpp` succeeds on macOS

---

### Wave 2: Core Implementation (Week 1, Days 3-5)
**Goal**: ModelManager and GPU detection working

- [ ] T-1771367264021 - Create llamacpp.go with CGO build tags (HIGH)
- [ ] T-1771367269400 - Add GPU detection for Metal and CUDA (HIGH)
- [ ] T-1771367249199 - Create ModelManager singleton with reference counting (HIGH)
- [ ] T-1771367250500 - Implement LRU cache for loaded models (HIGH)
- [ ] T-1771367254117 - Implement thread-safe concurrent model access (HIGH)

**Acceptance**: Can load/unload models safely with GPU detection

---

### Wave 3: Ollama Integration (Week 2, Days 1-3)
**Goal**: Auto-discover and load GGUF models from Ollama

- [ ] T-1771367227570 - Design Ollama manifest parsing strategy (HIGH)
- [ ] T-1771367235348 - Implement Ollama manifest parser (HIGH)
- [ ] T-1771367236869 - Create model discovery from Ollama storage (HIGH)
- [ ] T-1771367238424 - Map Ollama model names to GGUF blob paths (HIGH)
- [ ] T-1771367266058 - Implement GGUF model loader with dual path resolution (HIGH)
- [ ] T-1771367242253 - Test GGUF loading from Ollama blobs (HIGH)

**Acceptance**: `llamacpp models` lists available Ollama GGUF models

---

### Wave 4: Text Generation API (Week 2, Days 4-5)
**Goal**: Generate text via llamacpp

- [ ] T-1771367225957 - Design llamacpp tool schema and API (HIGH)
- [ ] T-1771367267805 - Implement text generation endpoint (HIGH)
- [ ] T-1771367270980 - Implement context management and tokenization (MEDIUM)
- [ ] T-1771367272521 - Create llamacpp provider for TextGenerator interface (HIGH)
- [ ] T-1771367279593 - Register llamacpp tool in registry.go (HIGH)

**Acceptance**: `llamacpp generate prompt="test"` returns text

---

### Wave 5: Integration (Week 3, Days 1-2)
**Goal**: Wire into provider chain and config

- [ ] T-1771367281667 - Update provider chain with llamacpp fallback (HIGH)
- [ ] T-1771367286109 - Update text_generate tool for llamacpp provider (HIGH)
- [ ] T-1771367283347 - Add llamacpp configuration to config schema (MEDIUM)
- [ ] T-1771367284807 - Add llamacpp to stdio://models resource (MEDIUM)
- [ ] T-1771367252253 - Add model warmup and preload configuration (MEDIUM)
- [ ] T-1771367255624 - Add memory usage monitoring and limits (MEDIUM)
- [ ] T-1771367240441 - Add model alias mapping system (MEDIUM)

**Acceptance**: `text_generate provider=llamacpp` works end-to-end

---

### Wave 6: Testing (Week 3, Days 3-4)
**Goal**: Verify Metal GPU, integration tests

- [ ] T-1771367308667 - Test GPU offloading on Metal and CUDA (MEDIUM)
- [ ] T-1771367303630 - Unit tests for llamacpp tool handlers (MEDIUM)
- [ ] T-1771367305621 - Integration tests with Ollama models (MEDIUM)
- [ ] T-1771367307285 - Benchmark vs Ollama HTTP performance (LOW)

**Acceptance**: Tests pass, Metal GPU utilized

---

### Wave 7: Documentation (Week 3, Day 5)
**Goal**: User-facing docs and examples

- [ ] T-1771367224337 - Document GGUF model compatibility requirements (MEDIUM)
- [ ] T-1771367312044 - Update text-generate skill with llamacpp option (LOW)
- [ ] T-1771367310542 - Update LLM_EXPOSURE_OPPORTUNITIES.md documentation (LOW)

**Acceptance**: Docs merged, examples tested

---

## Track 2: Compatibility & Polish ✨

**Priority**: MEDIUM  
**Timeline**: 1 week (can run parallel to llamacpp)  
**Value**: Language-neutral, cross-platform improvements

### Quick Wins (Days 1-2)

#### OpenCode Integration
- [ ] T-1771360895822 - Add --quiet flag to suppress verbose CLI output (HIGH)
- [ ] T-1771360896308 - Add --json flag for structured machine-readable output (HIGH)
- [ ] T-1771360897057 - Add --concise flag to strip emojis/decorative lines (MEDIUM)
- [ ] T-1771360897632 - Optimize tool descriptions and hints for MCP clients (MEDIUM)
- [ ] T-1771360822084 - Add --quiet and --json flags to exarp CLI (MEDIUM)

**Acceptance**: `exarp-go task list --quiet --json` works in scripts

#### Mac-Specific Fixes
- [ ] T-1771360298493 - Fix runIniTermTab: detect ITERM_SESSION_ID (HIGH)
- [ ] T-1771369913069 - Add timeout guard for Apple FM in task discovery (MEDIUM)

---

### Language Neutrality (Days 3-4)

- [ ] T-1772958892671074000 - Decide testing tool strategy: Go-only vs multi-language (MEDIUM)
- [ ] T-1772958900371299000 - Audit and neutralize Go-centric narrative in scorecard (LOW)
- [ ] T-1771361056782 - Add opencode-friendly hints to tool descriptions (MEDIUM)
- [ ] T-1771362279364 - Document MLX + OpenCode integration setup (MEDIUM)

**Acceptance**: Tools documented as language-specific or enhanced for multi-language

---

## Track 3: TUI Enhancements 🖥️

**Priority**: MEDIUM  
**Timeline**: 3-5 days  
**Value**: Better developer UX for task management

### High Priority
- [ ] T-1771456779351 - TUI: Inline status change shortcuts (d=Done, i=In Progress) (HIGH)
- [ ] T-1771355107420 - Handoffs list display and detail view in TUI (HIGH)

### Medium Priority
- [ ] T-1771456781612 - TUI: Task creation from TUI (n=new task) (MEDIUM)
- [ ] T-1771456780061 - TUI: Bulk status update for selected tasks (MEDIUM)
- [ ] T-1771543674831 - TUI: Improve error handling and user feedback (MEDIUM)
- [ ] T-1771543666094 - TUI: Route through MCP tools instead of direct calls (MEDIUM)
- [ ] T-1771543659626 - TUI: Extract action handlers into tui_update_actions.go (MEDIUM)

### Low Priority (Refactoring)
- [ ] T-1771543663481 - TUI: Extract sort/filter logic into tui_update_sorts.go (MEDIUM)
- [ ] T-1771543664811 - TUI: Decompose model struct into per-view state (LOW)
- [ ] T-1771543670781 - TUI: Deduplicate code between TUI and TUI3270 (LOW)
- [ ] T-1771543672166 - TUI: Create state machine for mode transitions (LOW)
- [ ] T-1771543673354 - TUI: Add keyboard shortcut customization (LOW)

**Acceptance**: TUI is faster, more intuitive, well-tested

---

## Track 4: Test Coverage 🧪

**Priority**: LOW (except high-priority fixes)  
**Timeline**: Ongoing  
**Value**: Confidence, regression prevention

### Fixes (Do First)
- [ ] T-1771245912597 - Fix TestRealModels_AnalyzeTask: model output parsing (MEDIUM) ⚠️ **ONLY FAILING TEST**

### New Coverage (Optional, Low Priority)
- [ ] T-1771748461437 - Add behavior validation tests for classify (LOW)
- [ ] T-1771748460125 - Add integration test for task preferred_backend (LOW)
- [ ] T-1771748458942 - Test GenerateWithOptions public function (LOW)
- [ ] T-1771748456311 - Test classify action with custom categories (LOW)
- [ ] T-1771543669306 - TUI: Add 3270 TUI test coverage (LOW)
- [ ] T-1771543667742 - TUI: Add unit tests for keyboard handlers (MEDIUM)
- [ ] T-1771171357749 - Unit tests for task_execute (MEDIUM)
- [ ] T-1771171359038 - Unit tests for prompt_analyzer (MEDIUM)
- [ ] T-1771171360013 - Unit tests for execution_apply (MEDIUM)

**Acceptance**: Test coverage >80%, no flaky tests

---

## Track 5: Documentation & Infrastructure 📚

**Priority**: LOW  
**Timeline**: Ongoing  
**Value**: Onboarding, maintenance

### Documentation
- [ ] T-1771460266378 - Docs: Document make lint-* targets in one place (LOW)
- [ ] T-1771460279379 - Docs: Add result comments to completed ansible tasks (LOW)
- [ ] T-1771460263674 - Docs: Document task_analysis and task_execute (MEDIUM)
- [ ] T-1771460263062 - Docs: Add exarp abilities audit to use-exarp-tools (MEDIUM)
- [ ] T-1771460276135 - Docs: Document ansible-galaxy and offline flow (LOW)
- [ ] T-1771355092457 - Validate exarp-go with OpenCode MCP (MEDIUM)

### Infrastructure
- [ ] T-1771456783942 - Make: Add check-security standalone target (LOW)
- [ ] T-1771456783140 - Make: Clean up dev.sh Python/pytest references (LOW)
- [ ] T-1771456782344 - Make: Add queue-status target (LOW)
- [ ] T-1771456781612 - Make: Add task-show target (LOW)
- [ ] T-1771460265620 - CI: Add lint-shellcheck, lint-yaml, lint-ansible (MEDIUM)
- [ ] T-1771460272225 - Optional: Add actionlint to default linters list (LOW)

### Ansible
- [ ] T-1771459746925 - Ansible-lint: Ensure community.general collection (HIGH)
- [ ] T-1771459748745 - Ansible-lint: Add .ansible-lint config (MEDIUM)
- [ ] T-1771457680959 - Ansible: Add integration test (vagrant/docker) (LOW)
- [ ] T-1771457731564 - Ansible: Verify playbook on Ubuntu/Docker (LOW)

---

## Track 6: Advanced Features (Future) 🚀

**Priority**: LOW  
**Timeline**: Post-llamacpp  
**Value**: Nice-to-have enhancements

### Task Workflow
- [ ] T-1771524914059 - Task content hash (optional): memory exact-dedup (LOW)
- [ ] T-1771524906950 - Task content hash: shared NormalizeForComparison (MEDIUM)
- [ ] T-1771524908368 - Task content hash: replace normalizeTaskContent (MEDIUM)
- [ ] T-1771524909926 - Task content hash: duplicate detection by content (MEDIUM)
- [ ] T-1771524911395 - Task content hash: sync conflict reporting (MEDIUM)
- [ ] T-1771524912621 - Task content hash: sanity check use shared norm (LOW)

### Handoff & Sessions
- [ ] T-1771527444192 - Handoff: optional restore from point-in-time snapshot (LOW)

### Database
- [ ] T-1771353163496 - Add rqlite driver (driver_rqlite.go, gorqlite) (MEDIUM)
- [ ] T-1771353168386 - Wire config and env for rqlite (DB_DRIVER, DB_DSN) (MEDIUM)
- [ ] T-1771353169844 - Document self-hosted rqlite setup (docs) (MEDIUM)
- [ ] T-1771353170943 - Tests for rqlite driver and config (MEDIUM)
- [ ] T-1771354500144 - Asynq producer: enqueue task_id from exarp-go (MEDIUM)

### AI/LLM
- [ ] T-1771252286533 - General agent abstraction for run-agent-in-task (HIGH)
- [ ] T-1771252280227 - Optional stdio://llm/status resource (LOW)
- [ ] T-1771252276374 - Document AI/LLM stack in main docs (MEDIUM)
- [ ] T-1771252272139 - Optional LocalAI backend (OpenAI-compatible) (LOW)
- [ ] T-1771252268378 - Evaluate langchaingo or Go AI SDK (LOW)

---

## Recommended Implementation Order

### Phase 1: Quick Wins (Week 1)
**Focus**: Immediate value, low effort

1. OpenCode flags (--quiet, --json, --concise) ⚡
2. iTerm fix (runIniTermTab detection) ⚡
3. Fix TestRealModels_AnalyzeTask (only failing test) ⚡
4. TUI inline shortcuts (d=Done, i=In Progress) ⚡

**Outcome**: Better CLI/MCP integration, tests green

---

### Phase 2: llamacpp Foundation (Week 2-3)
**Focus**: Build infrastructure for local LLM

1. Waves 1-2: Build system + ModelManager
2. Wave 3: Ollama integration
3. Wave 4: Text generation API

**Outcome**: Can generate text with llamacpp on Mac

---

### Phase 3: Integration & Polish (Week 4)
**Focus**: Complete llamacpp, compatibility fixes

1. Wave 5: Provider chain integration
2. Wave 6-7: Testing + docs
3. Testing tool strategy decision
4. Scorecard language-neutral audit

**Outcome**: llamacpp fully integrated, language-neutral

---

### Phase 4: TUI Enhancements (Week 5)
**Focus**: Developer UX improvements

1. Handoff view in TUI
2. Task creation from TUI
3. Bulk operations
4. Error handling improvements

**Outcome**: TUI is production-ready

---

### Phase 5: Documentation & Infrastructure (Week 6+)
**Focus**: Sustainability, onboarding

1. Consolidated documentation
2. CI/CD improvements
3. Ansible stabilization
4. Test coverage expansion

**Outcome**: Project is maintainable long-term

---

## Metrics & Success Criteria

### llamacpp Track
- ✅ Builds successfully on macOS/arm64
- ✅ Detects and uses Metal GPU
- ✅ Loads GGUF models from Ollama
- ✅ Generates text via MCP tool
- ✅ Performance benchmarks show >2x speedup vs Ollama HTTP
- ✅ Integration tests pass
- ✅ Documentation complete

### Compatibility Track
- ✅ All tools have explicit language compatibility status
- ✅ Testing tool strategy documented
- ✅ Scorecard output is language-neutral
- ✅ OpenCode integration validated
- ✅ CLI has --quiet and --json modes

### Quality Metrics
- ✅ Test suite: 100% passing (currently 99%+)
- ✅ Test coverage: >80%
- ✅ No critical bugs
- ✅ Documentation current
- ✅ Build time: <2min on Mac

---

## Risk Mitigation

### Risk 1: llamacpp CGO complexity
**Likelihood**: MEDIUM  
**Impact**: HIGH  
**Mitigation**: 
- Test early on macOS/arm64
- Have fallback to Ollama HTTP if CGO fails
- Document known issues clearly

### Risk 2: Model compatibility issues
**Likelihood**: MEDIUM  
**Impact**: MEDIUM  
**Mitigation**:
- Create compatibility matrix
- Test with common models (llama3.2, phi, etc.)
- Graceful degradation when unsupported

### Risk 3: Scope creep on TUI
**Likelihood**: HIGH  
**Impact**: LOW  
**Mitigation**:
- Focus on high-priority items only
- Defer "nice-to-have" refactorings
- Ship incrementally

---

## Notes

- **Total cleaned**: Removed 11 wrong-project tasks (mcp-go-core, IBKR, Jenkins)
- **Test status**: 7 "fix test" tasks marked Done (tests now pass)
- **MCP/CLI drift**: Recently resolved, compatibility matrix updated
- **Focus**: llamacpp is the critical path for Mac users

---

**Next Steps**:
1. Review and approve this plan
2. Start Phase 1 (Quick Wins)
3. Begin Wave 1 (llamacpp build foundation)
4. Update plan weekly based on progress

