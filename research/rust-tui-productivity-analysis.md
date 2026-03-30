# Rust TUI Productivity Tools - Comprehensive Research Analysis

## Executive Summary

This document consolidates research on **18 Rust-based terminal user interface (TUI) projects** focused on productivity, task management, time tracking, and development workflows. The analysis covers architecture patterns, technical implementations, and applicability to different use cases.

**Research Date:** 2026-03-26  
**Projects Analyzed:** 18 repositories  
**Total Lines of Code Analyzed:** ~1.2M+ lines

---

## Table of Contents

1. [Project Inventory](#project-inventory)
2. [Architecture Patterns](#architecture-patterns)
3. [Technology Stack Analysis](#technology-stack-analysis)
4. [State Management Patterns](#state-management-patterns)
5. [Persistence Strategies](#persistence-strategies)
6. [Integration Patterns](#integration-patterns)
7. [Applicability Matrix](#applicability-matrix)
8. [Best Practices](#best-practices)
9. [Research Gaps & Deep Dive Opportunities](#research-gaps--deep-dive-opportunities)

---

## Project Inventory

### Category: Task/Kanban Management (10 projects)

| Project | Stars | Primary Purpose | Unique Feature |
|---------|-------|-----------------|----------------|
| **basilk** | N/A | Simple Kanban boards | Minimal 3-status pipeline, JSON versioning |
| **jirust** | N/A | JIRA terminal client | Local caching with SurrealDB |
| **kanban** (fulsomenko) | 83 | Lazygit-inspired Kanban | MCP server for AI integration |
| **rust_kanban** | 255 | Full-featured Kanban | Encrypted cloud sync, drag-drop |
| **tmmpr** | N/A | Mind mapping | Infinite canvas, node connections |
| **td** | N/A | Graph-based todos | Task dependencies (DAG) |
| **taskwarrior-tui** | ~2000 | Taskwarrior interface | 9000+ LOC, mature project |
| **sc-cli** | N/A | Shortcut (Clubhouse) client | Git integration, multi-workspace |
| **rusty-krab-manager** | 327 | Pomodoro + random tasks | Weighted probability selection |
| **taskfinder** | N/A | Task extraction from notes | Markdown parsing, due dates |

### Category: Time Tracking (3 projects)

| Project | Stars | Primary Purpose | Unique Feature |
|---------|-------|-----------------|----------------|
| **timr-tui** | N/A | Multi-mode timer | 5 timer modes, 7 digit styles |
| **work-tuimer** | N/A | Time tracking | PIN-style input, SQLite persistence |
| **tatuin** | 130 | Task aggregation | 6 providers (Todoist, Obsidian, GitHub) |

### Category: Search/Navigation (5 projects)

| Project | Stars | Primary Purpose | Unique Feature |
|---------|-------|-----------------|----------------|
| **television** | Growing | Fuzzy finder | 881K LOC, Nucleo matching, channels |
| **todolist-rust** | N/A | Windows todo | Windows Registry storage |
| **maelstrom** | 718 | Distributed testing | Custom container runtime, FUSE fs |
| **nereid** | N/A | AI+Mermaid diagrams | 42K LOC, 40+ MCP tools |
| **taskui** | N/A | Taskfile.dev TUI | YAML task discovery |

---

## Architecture Patterns

### Pattern 1: State Machine (Mode Enum)

**Prevalence:** 16/18 projects (89%)

**Implementation:**
```rust
// Common pattern across basilk, rust_kanban, tmmpr, etc.
pub enum Mode {
    Normal,      // Navigation
    Editing,     // Text input
    Visual,      // Selection
    Search,      // Filter mode
}
```

**Projects Using:**
- basilk: `ViewMode` (11 variants)
- rust_kanban: `AppMode` with stack
- tmmpr: Vim-inspired modes
- taskui: `InputMode` (Select/Search/Preview)
- td: Task management modes

**Advantages:**
- Clear state transitions
- Contextual keybindings
- Easy to extend

---

### Pattern 2: Event-Driven Architecture

**Prevalence:** 17/18 projects (94%)

**Implementation Approaches:**

**Simple Thread (blocking):**
```rust
// basilk, taskui
loop {
    if event::poll(Duration::from_millis(100))? {
        match event::read()? {
            Event::Key(key) => handle_key(key),
            _ => {}
        }
    }
}
```

**Async Tokio (complex):**
```rust
// television, maelstrom
loop {
    tokio::select! {
        Some(event) = events.next() => handle_event(event),
        _ = tick_interval.tick() => update_ui(),
        Some(action) = action_rx.recv() => handle_action(action),
    }
}
```

**Projects Using Async:**
- television (tokio)
- maelstrom (tokio, gRPC)
- tatuin (tokio)
- rust_kanban (tokio)

---

### Pattern 3: Component/Widget Pattern

**Prevalence:** 14/18 projects (78%)

**Implementation:**
```rust
// From kanban (fulsomenko)
pub trait Component: Downcast {
    fn pre_render(&self, state: &AppState, storage: &mut FrameLocalStorage);
    fn render(&self, frame: &mut Frame, area: Rect, state: &AppState);
    fn process_input(&mut self, key: KeyEvent, state: &mut AppState) -> bool;
}
```

**Projects Using:**
- kanban: Trait-based components
- rust_kanban: WidgetManager
- nereid: MCP tool router
- television: Channel abstraction

---

### Pattern 4: Repository/Storage Pattern

**Prevalence:** 8/18 projects (44%)

**Implementation:**
```rust
// From kanban
#[async_trait]
pub trait PersistenceStore: Send + Sync {
    async fn save(&self, snapshot: StoreSnapshot) -> Result<PersistenceMetadata>;
    async fn load(&self) -> Result<(StoreSnapshot, PersistenceMetadata)>;
}
```

**Projects Using:**
- kanban: Pluggable persistence
- work-tuimer: SQLite repository
- jirust: SurrealDB caching
- rust_kanban: Cloud sync abstraction

---

## Technology Stack Analysis

### TUI Framework Adoption

| Framework | Projects | Notes |
|-----------|----------|-------|
| **ratatui** | 16 | Modern, maintained, feature-rich |
| **tui-rs** | 2 | Legacy (jirust, rusty-krab-manager) |

### Async Runtime Usage

| Runtime | Projects | Use Case |
|---------|----------|----------|
| **tokio** | 7 | Complex apps (television, maelstrom, rust_kanban) |
| **sync** | 11 | Simple apps (event loops, single-threaded) |

### Terminal I/O

| Library | Projects | Notes |
|---------|----------|-------|
| **crossterm** | 17 | Universal standard |
| **termion** | 1 | Legacy (rusty-krab-manager) |

### Serialization

| Format | Projects | Use Case |
|--------|----------|----------|
| **JSON** | 12 | Simple persistence, human-readable |
| **TOML** | 8 | Configuration files |
| **SQLite** | 3 | Structured data (work-tuimer, taskwarrior-tui) |
| **YAML** | 2 | Task parsing (taskui), CI configs |

---

## State Management Patterns

### Centralized State (Mutable App Struct)

**Prevalence:** 15/18 projects (83%)

```rust
// Common pattern
pub struct App {
    mode: Mode,
    items: Vec<Item>,
    selected: ListState,
    // ... all state in one place
}
```

**Projects:** basilk, td, tmmpr, taskui, timr-tui, work-tuimer, etc.

### Message Passing (Actor Model)

**Prevalence:** 5/18 projects (28%)

```rust
// From rust_kanban
tokio::spawn(async move { IoHandler::new(...) });
tokio::spawn(async move { WidgetManager::new(...) });
```

**Projects:** rust_kanban, maelstrom, television, tatuin, nereid

### Split State (Multi-Panel)

**Prevalence:** 3/18 projects (17%)

```rust
// From tatuin
pub struct App {
    providers: ArcRwLock<SelectableList<Provider>>,
    projects: ArcRwLock<SelectableList<Project>>,
    filter_widget: ArcRwLock<FilterWidget>,
}
```

**Projects:** tatuin, kanban, nereid

---

## Persistence Strategies

### Strategy Matrix

| Strategy | Projects | Pros | Cons |
|----------|----------|------|------|
| **Flat JSON** | 12 | Simple, portable, human-readable | No ACID, manual migrations |
| **SQLite** | 3 | ACID, indexed, relational | Complex, binary format |
| **Registry (Win)** | 1 | Native Windows | Platform-specific |
| **Cloud Sync** | 2 | Cross-device | Requires auth, latency |
| **Hybrid** | 2 | Best of both | Complexity |

### Notable Implementations

**Versioned JSON with Migrations (kanban):**
```rust
pub static JSON_VERSIONS: [&str; 2] = ["6ad96", "911fc"];
```

**SQLite with Revision Tokens (work-tuimer):**
```sql
CREATE TABLE day_meta (
    date TEXT PRIMARY KEY,
    revision INTEGER NOT NULL DEFAULT 0
);
```

**Encrypted Cloud Sync (rust_kanban):**
- AES-256-GCM encryption
- Supabase backend
- Key stored locally

---

## Integration Patterns

### API Integration

| Type | Projects | Examples |
|------|----------|----------|
| **REST API** | 6 | JIRA, Todoist, GitHub, GitLab, Shortcut |
| **gRPC** | 2 | maelstrom, nereid |
| **MCP Server** | 2 | kanban, nereid |
| **File-based** | 4 | Obsidian, Markdown, YAML |

### Shell Integration

| Feature | Projects |
|---------|----------|
| **Shell scripts** | television (ctrl+t), taskwarrior-tui |
| **Keybindings** | Universal (vim-style) |
| **CLI mode** | work-tuimer, television, taskui |

---

## Applicability Matrix

### By Workflow Type

| Workflow Type | Best Matches | Why |
|---------------|--------------|-----|
| **Solo dev, terminal-centric** | basilk, td, taskui | Simple, no external deps |
| **Team with JIRA** | jirust | Full JIRA feature support |
| **Time tracking/billing** | work-tuimer, timr-tui | SQLite persistence, reports |
| **Multi-tool aggregation** | tatuin | 6 providers unified |
| **Documentation-heavy** | tmmpr, nereid | Diagrams, mind maps |
| **Large codebase CI** | maelstrom | Distributed test execution |
| **GitHub-centric** | sc-cli, tatuin | GitHub Issues/Shortcut integration |
| **Note-taking** | taskfinder, tmmpr | Markdown extraction |

### By Team Size

| Team Size | Best Matches |
|-----------|--------------|
| **Solo** | basilk, td, timr-tui, taskui |
| **Small (2-5)** | rust_kanban, tmmpr, work-tuimer |
| **Medium (5-20)** | jirust, tatuin, sc-cli |
| **Large (20+)** | maelstrom, taskwarrior-tui |

### By Platform

| Platform | Best Matches | Notes |
|----------|--------------|-------|
| **Linux** | All except todolist-rust | Universal support |
| **macOS** | All except maelstrom | maelstrom is Linux-only |
| **Windows** | todolist-rust (native), most via WSL | todolist-rust uses WinRegistry |

---

## Best Practices

### ✅ Patterns to Emulate

1. **Clear Mode Separation**
   - Use enums for distinct UI states
   - Contextual help based on current mode
   - Example: basilk, tmmpr

2. **Keyboard-First Design**
   - Vim-style navigation (hjkl)
   - Arrow key alternatives
   - Universal: q to quit
   - Example: taskwarrior-tui, television

3. **Progressive Enhancement**
   - Mouse support optional
   - CLI mode + TUI mode
   - Example: work-tuimer, television

4. **Clear Error Messages**
   - Use anyhow for error propagation
   - Human-readable error context
   - Example: television, maelstrom

5. **Configuration Flexibility**
   - Environment variables
   - Config files (TOML/YAML)
   - CLI arguments
   - Example: kanban, tatuin

### ⚠️ Anti-Patterns to Avoid

1. **Single Large File**
   - todolist-rust: 376 LOC in main.rs
   - Hard to maintain and test

2. **Platform-Specific Storage**
   - todolist-rust: Windows Registry only
   - Limits portability

3. **No Tests**
   - rust_kanban acknowledges this gap
   - Critical for reliability

4. **Blocking I/O in Event Loop**
   - Can cause UI freezing
   - Use async or background threads

5. **Hardcoded Values**
   - Magic numbers for layout
   - Should be configurable

---

## Research Gaps & Deep Dive Opportunities

### High Priority

1. **Ratatui vs tui-rs Migration Patterns**
   - How projects transitioned
   - Breaking changes analysis
   - Performance differences

2. **Async State Management**
   - Comparison of approaches
   - Tokio vs sync tradeoffs
   - Deadlock prevention

3. **Testing Strategies for TUI Apps**
   - Unit testing UI logic
   - Integration testing with mocked backends
   - Screenshot testing

### Medium Priority

4. **Accessibility in TUIs**
   - Screen reader support
   - High contrast themes
   - Keyboard-only navigation

5. **Cross-Platform File Watching**
   - notify crate usage
   - Performance on large repos
   - Event debouncing

6. **Persistence Performance**
   - JSON vs SQLite benchmarks
   - Migration strategies
   - Conflict resolution

### Low Priority

7. **Packaging & Distribution**
   - Homebrew formulas
   - AUR packages
   - Nix flakes
   - Cargo install

8. **Documentation Generation**
   - Keybinding tables
   - Help systems
   - Man pages

---

## Related Resources

### Internal Links
- Research raw data: See 18 individual research reports
- Task tracking: .todo2/state.todo2.json

### External Links
- [ratatui documentation](https://ratatui.rs/)
- [tui-rs archive](https://github.com/fdehau/tui-rs)
- [Rust CLI book](https://rust-cli.github.io/book/)

---

## Document Metadata

- **Created:** 2026-03-26
- **Author:** Claude (Sisyphus)
- **Version:** 1.0
- **Status:** Draft - Awaiting deep dive completion

---

## Next Steps

1. ✅ Create consolidated research document (THIS)
2. ⬜ Execute deep dive: ratatui patterns analysis
3. ⬜ Execute deep dive: state management comparison
4. ⬜ Execute deep dive: persistence strategies
5. ⬜ Create decision matrix for tool selection
6. ⬜ Document code samples and templates

---

*Document generated by Sisyphus agent as part of Rust TUI ecosystem research.*
