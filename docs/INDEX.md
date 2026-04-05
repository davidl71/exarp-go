# exarp-go Documentation Index

**Last Updated**: 2026-03-08  
**Quick Start**: See `GETTING_STARTED.md` (if it exists) or start with `CURSOR_MCP_SETUP.md`

---

## 🚀 Getting Started

| Document | Description |
|----------|-------------|
| `CURSOR_MCP_SETUP.md` | Set up exarp-go with Cursor IDE |
| `OPENCODE_INTEGRATION.md` | Use exarp-go with OpenCode |
| `PORTABLE_MCP_RUNNER.md` | Per-project MCP setup |
| `CLI_MAKE_CI_USAGE.md` | Command-line usage guide |

---

## 📚 User Guides

### Preferred Tool Surface
Use these first unless you specifically need a backend-specific or compatibility tool:

- `task_workflow`, `task_analysis`, `task_discovery`
- `report`, `health`, `session`, `automation`
- `testing`, `lint`, `security`, `git_tools`
- `memory`, `memory_maint`, `recommend`, `text_generate`
- `workflow_mode`, `tool_catalog`, `generate_config`, `setup_hooks`

Advanced/specialist tools such as `ollama`, `cursor_cloud_agent`, and `fm_plan_and_execute` still exist, but they are not the recommended starting surface. Unified generation is via `text_generate` (`fm`, `ollama`, `insight`, `localai`, `gateway`, `auto`).

Compatibility aliases still exist for migration:
- `task_execute` -> prefer `task_workflow`
- `infer_session_mode` -> prefer `session`
- `scan_dependency_security` -> prefer `security`
- `context_budget` -> prefer `context`

### Task Management
| Document | Description |
|----------|-------------|
| `TASK_TOOLS_GUIDE.md` | Guide to task_workflow, task_analysis, and task-discovery flows |
| `TASK_TOOLS_COMPARISON.md` | Detailed comparison of task tools |
| `EXARP_CLI_SHORTCUTS.md` | CLI shortcuts reference |

### Development Workflow
| Document | Description |
|----------|-------------|
| `WORKFLOW_USAGE.md` | Workflow usage guide |
| `MODEL_ASSISTED_WORKFLOW.md` | AI-assisted workflow with local LLMs |
| `HANDOFF_VIA_GIT.md` | Session handoff via git |

### Code Quality
| Document | Description |
|----------|-------------|
| `LINT_TARGETS.md` | All linting make targets in one place |
| `LINTERS_CONFIGURATION.md` | Linter configuration guide |
| `SCORECARD_GO_IMPLEMENTATION.md` | Project scorecard feature |

---

## 🔧 Technical Reference

### Architecture
| Document | Description |
|----------|-------------|
| `FRAMEWORK_AGNOSTIC_DESIGN.md` | Framework-agnostic architecture |
| `DEVWISDOM_GO_LESSONS.md` | Go development best practices |
| `TOOL_LANGUAGE_COMPATIBILITY_MATRIX.md` | Tool language compatibility |

### AI/LLM Stack
| Document | Description |
|----------|-------------|
| `GO_AI_ECOSYSTEM.md` | AI backend stack (Ollama, Apple FM, LocalAI, gateway) |
| `archive/llamacpp-removed/` | Historical llama.cpp/GGUF docs (product path removed; use Ollama) |
| `LLM_NATIVE_ABSTRACTION_PATTERNS.md` | LLM abstraction patterns |

### MCP Integration
| Document | Description |
|----------|-------------|
| `MCP_FRAMEWORKS_COMPARISON.md` | MCP framework comparison |
| `CURSOR_API_AND_CLI_INTEGRATION.md` | Cursor integration details |

### Database & Storage
| Document | Description |
|----------|-------------|
| `PROTOBUF_USAGE.md` | Protobuf usage and tooling |
| `TASK_CONTENT_HASH_DESIGN.md` | Task content hashing |

---

## 📋 Project Management

### Planning & Roadmap
| Document | Description |
|----------|-------------|
| `IMPLEMENTATION_PLAN.md` | **UPDATED** 6-week implementation roadmap |
| `out/PROJECT_OVERVIEW.md` | Auto-generated project status |
| `PRD.md` | Product requirements document |
| `PROJECT_GOALS.md` | Project goals and objectives |

### Analysis & Reports
| Document | Description |
|----------|-------------|
| `out/TASK_ANALYSIS_DUPLICATES.md` | Auto-generated duplicate analysis |
| `out/TAG_ANALYSIS_RESULT.json` | Tag analysis (auto-generated) |
| `DEADCODE_REPORT.md` | Dead code analysis |

---

## 🛠️ Development

### Build & Deploy
| Document | Description |
|----------|-------------|
| `PORTABLE_MCP_RUNNER.md` | Portable runner for MCP |
| `CLI_MAKE_CI_USAGE.md` | Make targets and CI usage |

### Testing
| Document | Description |
|----------|-------------|
| `DEV_TEST_AUTOMATION.md` | Development and test automation |
| `TESTING_SUMMARY.md` | Testing strategy summary |

---

## 🗂️ Archive

Historical and outdated documentation has been moved to `docs/archive/`.

**Total Active Docs**: ~286 files (many are auto-generated reports)  
**Key User-Facing Docs**: ~30 files

---

## 📖 Quick Reference

**Most Important Docs for New Users**:
1. `CURSOR_MCP_SETUP.md` or `OPENCODE_INTEGRATION.md` - Get started with your IDE
2. `TASK_TOOLS_GUIDE.md` - Learn the primary task-management surface
3. `CLI_MAKE_CI_USAGE.md` - CLI reference
4. `LINT_TARGETS.md` - Code quality tools
5. `GO_AI_ECOSYSTEM.md` - Local LLM setup

**Most Important for Contributors**:
1. `IMPLEMENTATION_PLAN.md` - Current roadmap
2. `FRAMEWORK_AGNOSTIC_DESIGN.md` - Architecture
3. `DEVWISDOM_GO_LESSONS.md` - Best practices
4. `TOOL_LANGUAGE_COMPATIBILITY_MATRIX.md` - Language support

---

## 🔍 Finding Documentation

**By Topic**:
```bash
# Find all docs about tasks
ls docs/*TASK*.md

# Find all docs about MCP
ls docs/*MCP*.md

# Search doc content
rg "keyword" docs/*.md
```

**Auto-Generated** (updated by tools):
- `out/PROJECT_OVERVIEW.md`
- `out/TASK_ANALYSIS_DUPLICATES.md`
- `out/TAG_ANALYSIS_RESULT.json`
- Various `*_REPORT.md` files

**User-Maintained** (manually updated):
- `IMPLEMENTATION_PLAN.md`
- `TASK_TOOLS_GUIDE.md`
- `LINT_TARGETS.md`
- Integration guides

---

## 📝 Documentation Standards

- **Format**: Markdown (.md)
- **Linting**: gomarklint (run `make lint-all`)
- **Style**: Clear, concise, code examples
- **Updates**: Update this index when adding major new docs

---

**Need help?** Start with the relevant guide above or search with `rg "your-topic" docs/`
