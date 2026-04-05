# exarp-go Implementation Plan

**Generated**: 2026-03-08  
**Updated**: 2026-04-05 (llamacpp / direct GGUF path removed from product)  
**Total Tasks**: 473 (102 Todo, 371 Done)  
**Scope**: Remaining work after cleanup and architectural unification (local LLMs via Ollama + `text_generate`, Apple FM chain)

---

## Executive Summary

After recent MCP/CLI drift fixes and cleanup of wrong-project tasks, exarp-go has **~100 remaining tasks** organized into clear implementation tracks:

1. **Compatibility & Polish** - Language-neutral tooling
2. **TUI Enhancements** - Better task management UX
3. **Test Coverage** - Fill remaining gaps
4. **Documentation** - User-facing guides

**Critical Path**: TUI enhancements and compatibility improvements  
**Quick Wins**: ✅ Phase 1 complete! (OpenCode flags, iTerm fix, test fixes, TUI shortcuts)

---

## Track 1: Compatibility & Polish ✨

**Priority**: MEDIUM  
**Timeline**: 1 week  
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
- [x] T-1771362279364 - ~~Document MLX + OpenCode~~ — exarp-go no longer ships an `mlx` tool; OpenCode MLX is independent (see `OPENCODE_INTEGRATION.md` §5).

**Acceptance**: Tools documented as language-specific or enhanced for multi-language

---

## Track 2: TUI Enhancements 🖥️

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

## Track 3: Test Coverage 🧪

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

## Track 4: Documentation & Infrastructure 📚

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

## Track 5: Advanced Features (Future) 🚀

**Priority**: LOW  
**Timeline**: Ongoing / backlog  
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

### Phase 2: Compatibility & docs (Week 2-3)
**Focus**: Language-neutral tooling and MCP polish

1. Testing tool strategy decision
2. Scorecard / report language-neutral audit
3. OpenCode-friendly hints and CLI flags

**Outcome**: Clear multi-language story; fewer Go-centric defaults

---

### Phase 3: TUI Enhancements (Week 4)
**Focus**: Developer UX improvements

1. Handoff view in TUI
2. Task creation from TUI
3. Bulk operations
4. Error handling improvements

**Outcome**: TUI is production-ready

---

### Phase 4: Documentation & Infrastructure (Week 5+)
**Focus**: Sustainability, onboarding

1. Consolidated documentation
2. CI/CD improvements
3. Ansible stabilization
4. Test coverage expansion

**Outcome**: Project is maintainable long-term

---

## Metrics & Success Criteria

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

### Risk 1: Model compatibility issues (Ollama / FM)
**Likelihood**: MEDIUM  
**Impact**: MEDIUM  
**Mitigation**:
- Create compatibility matrix
- Test with common models (llama3.2, phi, etc.)
- Graceful degradation when unsupported

### Risk 2: Scope creep on TUI
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
- **Focus**: Ollama + `text_generate` for portable local LLMs; Apple FM via `provider=fm` where CGO/darwin allow

---

**Next Steps**:
1. Review and approve this plan
2. Start Phase 1 (Quick Wins)
3. Execute Phase 2 (compatibility + docs)
4. Update plan weekly based on progress

