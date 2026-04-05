# Model-Assisted Workflow: Local LLMs (FM, Ollama)

**Date:** 2026-01-07 (updated 2026-04-05)  
**Status:** 📋 Design Phase  
**Purpose:** Leverage local models (Apple FM chain, Ollama) for task breakdown, execution, and prompt optimization

---

## Executive Summary


## External Tool Hints

For documentation on external libraries used in this document, use Context7:

- **cloud**: Use `resolve-library-id` then `query-docs` for cloud documentation



This document outlines how to use local LLM models (Ollama, FM via `text_generate`) to enhance the Todo2 workflow by:
1. **Task Breakdown** - Decomposing complex tasks into manageable subtasks
2. **Easy Task Execution** - Automating routine/simple tasks using local models
3. **Prompt Optimization** - Iteratively refining prompts for better AI responses

---

## End-user guide

**When to use local models:** Use local models for routine work (code review, documentation, task breakdown, simple code generation, privacy-sensitive tasks). Use cloud AI for complex reasoning, high-stakes decisions, or when local models fail. See [Benefits & Trade-offs](#benefits--trade-offs) and [When to Use Local Models](#when-to-use-local-models) below.

**task_execute vs prompt optimization:**

- **task_execute** — Runs the full execution flow for a Todo2 task: load task, render prompt, call local model, parse response, apply file changes (with confidence threshold), add result comment. Use when you want the AI to attempt implementing a task end-to-end.
- **Prompt optimization** — Iteratively refines a prompt via `prompt_analyzer` and `RefinePromptLoop` for clarity, specificity, and structure. Use when improving prompts for reuse or when task_execute produces poor outputs and you want to tune the prompt first.

**Running tests with real models:** Use `make test-real-models` to run integration tests that call real backends (FM/Ollama). Regular `make test` skips these tests. See [Testing with real models](#testing-with-real-models-make-test-real-models) for backend requirements and skip behavior.

**Preferred backend and local_ai_backend:**

- **Task metadata `preferred_backend`** — Optional key: `fm` or `ollama` (legacy `mlx` is ignored). When set on a task, tools that use LLMs (estimation, text_generate, report insights) respect it.
- **Tool param `local_ai_backend`** — Pass to `task_workflow` create/update or `estimation` to override the backend for that call. Stored as `preferred_backend` when creating tasks.
- **summarize** — When `local_ai_backend` is not passed, the task's `preferred_backend` is used; if unset, the default is `fm`. The param overrides task metadata for that call.
- **run_with_ai** — Uses the task's `preferred_backend` when `local_ai_backend` is not passed; optional `instruction` adds extra guidance for the model. The param overrides task metadata for that call.

**Summarize and run_with_ai backend selection:**  
`task_workflow` actions `summarize` and `run_with_ai` both use the task's `preferred_backend` (from task metadata) when the caller does not supply `local_ai_backend`. If neither is set, the default is `fm` (FM chain: Apple → Ollama → stub). CLI: `exarp-go task summarize T-xxx [--local-ai-backend fm]` and `exarp-go task run-with-ai T-xxx [--backend ollama]`.

**CLI commands with backend options:** `task create`, `task update`, `task estimate`, `task summarize`, and `task run-with-ai` accept `--local-ai-backend` or (for run-with-ai) `--backend` with values `fm` or `ollama`.

**Examples:**

```bash
# Create task with preferred backend (CLI, when --recommended-tools / local_ai_backend supported)
exarp-go -tool task_workflow -args '{"action":"create","name":"Add tests","long_description":"...","local_ai_backend":"ollama"}'

# Estimation with specific backend
exarp-go -tool estimation -args '{"action":"estimate","name":"Refactor module","local_ai_backend":"ollama"}'

# Set preferred_backend on existing task (CLI or MCP)
exarp-go task update T-123 --local-ai-backend ollama
# Or: exarp-go -tool task_workflow -args '{"action":"update","task_id":"T-123","local_ai_backend":"ollama"}'

# CLI subcommands (task estimate, summarize, run-with-ai)
exarp-go task estimate "Refactor module" --local-ai-backend ollama
exarp-go task summarize T-xxx [--local-ai-backend fm]
exarp-go task run-with-ai T-xxx [--backend ollama] [--instruction "..."]
```

---

**Key Benefits:**
- ✅ **Privacy** - All processing happens locally (no data sent to external APIs)
- ✅ **Cost Efficiency** - No API costs for routine operations
- ✅ **Speed** - Fast inference on Apple Silicon (FM / Ollama) or local GPU (Ollama)
- ✅ **Offline Capability** - Works without internet connection
- ✅ **Iterative Refinement** - Can optimize prompts through multiple iterations

---

## Research Findings

> **2026-04:** exarp-go does **not** register an `mlx` MCP tool. Use **Ollama** (e.g. codellama tags) or **`text_generate`** with `provider=fm` / `ollama` / `auto`. Historical MLX-only model IDs below are optional on-disk formats; operationally pull CodeLlama via Ollama.

### CodeLlama (recommended via Ollama)

**What it is:** Code-focused LLMs (7B–34B) for synthesis, review, and task breakdown.

**In exarp-go:** Call through **`ollama`** or **`text_generate`** with `provider=ollama` and an Ollama model tag (e.g. `codellama` variants). See `docs/MLX_ARCHITECTURE_ANALYSIS.md` only for historical MLX-community context.

### Ollama

**What It Is:**
- Local LLM runtime that manages model downloads and execution
- Supports many models (Llama, Mistral, CodeLlama, etc.)
- HTTP API for easy integration
- Cross-platform (works on Mac, Linux, Windows)

**Current Integration:**
- ✅ `ollama` tool available in MCP server
- ✅ Default model: `llama3.2`
- ✅ Supports custom models and configurations

**Models Suitable for Task Work:**
- **llama3.2** - General purpose, good for task breakdown
- **codellama** - Code-focused (pull via Ollama)
- **mistral** - Fast and capable
- **qwen2.5** - Strong reasoning and code understanding

---

## Use Cases

### 1. Task Breakdown

**Problem:** Complex tasks need to be decomposed into manageable subtasks.

**Solution:** Use CodeLlama (via Ollama) or the FM chain to analyze task descriptions and generate subtask breakdowns.

**Workflow:**
1. User creates a complex task (e.g., "Migrate Python MCP server to Go")
2. AI calls local model with task description
3. Model analyzes complexity and suggests subtask breakdown
4. AI creates Todo2 tasks based on model suggestions
5. Human reviews and approves/modifies breakdown

**Example Prompt:**
```
Analyze this task and break it down into 3-8 subtasks:

Task: "Migrate Python MCP server to Go SDK"

Requirements:
- Each subtask should be independently executable
- Subtasks should have clear acceptance criteria
- Consider dependencies between subtasks
- Estimate complexity (simple/medium/complex) for each

Format: JSON with fields: name, description, acceptance_criteria, complexity, dependencies
```

**Benefits:**
- More accurate task decomposition
- Better dependency identification
- Improved estimation accuracy

### 2. Easy Task Execution

**Problem:** Many routine tasks (code review, documentation, simple refactoring) can be automated.

**Solution:** Use local models to execute simple tasks directly, reducing AI interaction overhead.

**Workflow:**
1. AI identifies a task as "easy" (routine, well-defined, low-risk)
2. AI delegates task to local model (CodeLlama for code, general model for docs)
3. Model generates solution
4. AI reviews output and applies changes
5. Task marked as Done automatically

**Task Categories Suitable for Auto-Execution:**
- **Code Review** - Simple syntax/style checks
- **Documentation** - Adding comments, updating READMEs
- **Refactoring** - Renaming, simple restructuring
- **Test Generation** - Basic unit test creation
- **Code Formatting** - Applying style guidelines

**Example:**
```
Task: "Add error handling to function X"

Model receives:
- Function code
- Error handling requirements
- Code style guidelines

Model generates:
- Updated function with error handling
- Brief explanation of changes
```

**Benefits:**
- Faster task completion
- Reduced AI token usage
- Consistent code quality
- Parallel execution of multiple easy tasks

### 3. Prompt Optimization

**Problem:** Prompts need iterative refinement to get best results from AI.

**Solution:** Use local models to analyze and optimize prompts before sending to main AI.

**Workflow:**
1. AI generates initial prompt for a task
2. Local model analyzes prompt quality
3. Model suggests improvements (clarity, specificity, structure)
4. AI refines prompt based on suggestions
5. Optimized prompt used for actual task

**Optimization Criteria:**
- **Clarity** - Is the task clearly defined?
- **Specificity** - Are requirements specific enough?
- **Completeness** - Are all necessary details included?
- **Structure** - Is the prompt well-organized?
- **Actionability** - Can the AI execute this without clarification?

**Example:**
```
Original Prompt:
"Fix the bug in the server"

Optimized Prompt:
"Fix the bug in the MCP server where tool registration fails when tool name contains special characters. 
The fix should:
1. Validate tool names during registration
2. Return clear error messages for invalid names
3. Add unit tests for edge cases
4. Update documentation with naming rules"
```

**Benefits:**
- Better AI responses (fewer clarification requests)
- More accurate task execution
- Reduced iteration cycles
- Improved overall workflow efficiency

---

## Architecture Design

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Todo2 Workflow                        │
│  (Task Creation, Management, Execution)                 │
└─────────────────┬─────────────────────────────────────┘
                   │
                   ├──────────────────────────────────────┐
                   │                                      │
        ┌──────────▼──────────┐              ┌──────────▼──────────┐
        │  Task Analyzer      │              │  Prompt Optimizer    │
        │  (Complexity Check) │              │  (Iterative Refine)  │
        └──────────┬──────────┘              └──────────┬──────────┘
                   │                                      │
        ┌──────────▼──────────┐              ┌──────────▼──────────┐
        │  Model Router       │              │  Model Router        │
        │  (Select Best Model)│              │  (Select Best Model) │
        └──────────┬──────────┘              └──────────┬──────────┘
                   │                                      │
        ┌──────────┴──────────┐              ┌──────────┴──────────┐
        │                     │              │                     │
┌───────▼────────┐   ┌────────▼────────┐  ┌───▼────────────┐  ┌───────▼────────┐
│  FM chain      │   │ Ollama (code)  │  │ Gateway        │  │ Ollama (chat)  │
│  (text_generate│   │  codellama     │  │ (optional)     │  │  llama3.2 …    │
│   provider=fm) │   │                │  │                │  │                │
└────────────────┘   └────────────────┘  └────────────────┘  └────────────────┘
```

### Core Components

#### 1. Task Analyzer

**Purpose:** Determine if a task can benefit from model assistance.

**Logic:**
```go
type TaskComplexity string

const (
    ComplexitySimple  TaskComplexity = "simple"   // Can auto-execute
    ComplexityMedium  TaskComplexity = "medium"   // Needs breakdown
    ComplexityComplex TaskComplexity = "complex"  // Needs human review
)

func AnalyzeTask(task Task) (complexity TaskComplexity, canAutoExecute bool, needsBreakdown bool) {
    // Analyze task description, acceptance criteria, dependencies
    // Return complexity assessment and recommendations
}
```

**Decision Criteria:**
- **Simple:** Well-defined, routine, low-risk, < 1 hour estimated
- **Medium:** Needs breakdown, multiple steps, some uncertainty
- **Complex:** High-stakes, experimental, requires human judgment

#### 2. Model Router

**Purpose:** Select the best model for a given task and run generation.

**Implementation:** `internal/tools/model_router.go`

```go
type ModelRouter interface {
    SelectModel(taskType string, requirements ModelRequirements) ModelType
    Generate(ctx context.Context, model ModelType, prompt string, maxTokens int, temperature float32) (string, error)
}

// ModelType: fm, gateway, ollama-llama, ollama-codellama (see model_router.go)
// DefaultModelRouter is the shared instance.
```

**Selection Logic (`defaultModelRouter.SelectModel`):** planner/reviewer roles prefer FM when `FMAvailable()`; else if FM available return FM; else if `GatewayAvailable()` return gateway; else code-ish tasks → `ollama-codellama`, else → `ollama-llama`. See `internal/tools/model_router.go`.

**Discovery:** `LLMBackendStatus()` / `stdio://models` — no `mlx` tool.

**Usage:** `infer_task_progress`, `task_analysis` (tags action with `use_llm_semantic`).

#### 3. Task Breakdown Handler

**Purpose:** Use models to decompose complex tasks.

**Implementation:**
```go
type BreakdownRequest struct {
    TaskDescription   string
    AcceptanceCriteria []string
    Context           map[string]interface{}
}

type BreakdownResult struct {
    Subtasks []Subtask
    Dependencies map[string][]string
    Estimates   map[string]Duration
}

func BreakDownTask(ctx context.Context, req BreakdownRequest, model ModelType) (*BreakdownResult, error) {
    // Generate prompt for task breakdown
    prompt := buildBreakdownPrompt(req)
    
    // Call model
    response, err := callModel(ctx, model, prompt)
    if err != nil {
        return nil, err
    }
    
    // Parse response into structured breakdown
    return parseBreakdown(response)
}
```

#### 4. Auto-Execution Handler

**Purpose:** Execute simple tasks using local models.

**Implementation:**
```go
type ExecutionRequest struct {
    Task        Task
    CodeContext string  // Relevant code files
    Guidelines  string  // Style/requirements
}

type ExecutionResult struct {
    Changes     []Change
    Explanation string
    Confidence  float64
}

func AutoExecuteTask(ctx context.Context, req ExecutionRequest, model ModelType) (*ExecutionResult, error) {
    // Build execution prompt with context
    prompt := buildExecutionPrompt(req)
    
    // Call model
    response, err := callModel(ctx, model, prompt)
    if err != nil {
        return nil, err
    }
    
    // Parse and validate response
    result := parseExecution(response)
    
    // Apply changes if confidence is high
    if result.Confidence > 0.8 {
        return applyChanges(result.Changes)
    }
    
    return result, nil
}
```

#### 5. Prompt Optimizer

**Purpose:** Iteratively refine prompts for better AI responses.

**Implementation:**
```go
type PromptAnalysis struct {
    Clarity      float64
    Specificity  float64
    Completeness float64
    Suggestions  []string
}

func OptimizePrompt(ctx context.Context, originalPrompt string, model ModelType) (string, *PromptAnalysis, error) {
    // Analyze current prompt
    analysis := analyzePrompt(originalPrompt)
    
    // Generate suggestions
    suggestions := generateSuggestions(ctx, model, originalPrompt, analysis)
    
    // Refine prompt
    optimized := applySuggestions(originalPrompt, suggestions)
    
    return optimized, analysis, nil
}
```

---

## Integration Points

### 1. Todo2 Workflow Integration

**Task Creation Flow:**
```
User Request
    ↓
AI Creates Initial Task
    ↓
Task Analyzer (Complexity Check)
    ↓
If Complex → Model Breakdown → Create Subtasks
If Simple → Model Auto-Execution → Apply Changes
If Medium → Standard AI Processing
```

**Task Execution Flow:**
```
Task Ready for Execution
    ↓
Check Complexity
    ↓
If Simple → Auto-Execute with Model
If Complex → Standard AI Processing
    ↓
Review Results
    ↓
Mark as Done or Request Changes
```

### 2. MCP Tool Integration

**New Tools to Add:**
- `model_task_breakdown` - Decompose complex tasks
- `model_auto_execute` - Execute simple tasks
- `model_optimize_prompt` - Refine prompts

**Existing Tools to Enhance:**
- `estimation` — FM / Ollama backends
- `ollama` — full action surface
- `text_generate` — unified provider switch

### 3. Implementation notes

**Current state:** Native Go `ollama` client, `text_generate`, and `model_router.go`; MLX MCP path removed.

---

## Implementation Plan

### Phase 1: Research & Design (Current)
- ✅ Research CodeLlama / Ollama / FM capabilities
- ✅ Design architecture
- 📋 Create implementation tasks

### Phase 2: Model Integration
- [x] Create model router component
- [x] Implement Ollama HTTP client
- [x] Add model selection logic — T-207 Done; `ResolveModelForTask` (recommend + router), `text_generate` provider=auto with task_type/task_description

### Phase 3: Task Breakdown
- [ ] Implement task analyzer
- [ ] Create breakdown prompt templates
- [ ] Add breakdown result parser
- [ ] Integrate with Todo2 task creation

### Phase 4: Auto-Execution
- [ ] Implement task complexity classifier
- [x] Create execution prompt templates — `task_execution` in `internal/prompts/templates.go` (T-213)
- [x] Add change application logic — `ApplyChanges`, `ParseExecutionResponse` in `internal/tools/execution_apply.go` (T-214)
- [x] Integrate with Todo2 execution flow — `task_execute` tool in `internal/tools/task_execute.go` (T-215): load task, render template, model generate, parse, apply changes, add result comment

### Phase 5: Prompt Optimization
- [ ] Implement prompt analyzer
- [x] Define template spec — [PROMPT_OPTIMIZATION_TEMPLATE_SPEC.md](PROMPT_OPTIMIZATION_TEMPLATE_SPEC.md)
- [x] Create analysis prompt template — `prompt_optimization_analysis` in `internal/prompts/templates.go`
- [x] Create suggestions-generation template — `prompt_optimization_suggestions` (T-1770830686054)
- [x] Create refinement template — `prompt_optimization_refinement` (T-1770830686525)
- [x] Add iterative refinement loop — `RefinePromptLoop`, `GenerateSuggestions`, `RefinePrompt` in `internal/tools/prompt_analyzer.go` (T-218)
- [ ] Integrate with task creation workflow

### Phase 6: Testing & Documentation
- [x] Unit tests for all components — `execution_apply_test.go`, `prompt_analyzer_test.go` (mock generator), `task_execute_test.go` (mock ModelRouter); no real LLM required.
- [x] Integration tests with real models — See `internal/tools/real_models_integration_test.go`. Run with `make test-real-models` (requires a local backend; skipped in `go test -short`).
- [ ] Performance benchmarks
- [x] User documentation — End-user guide (when to use local models, task_execute vs prompt optimization, make test-real-models), backend requirements, and testing section

#### Testing with real models: make test-real-models

Run integration tests that call real LLM backends:

```bash
make test-real-models
```

This runs `go test -run RealModels ./internal/tools/... -timeout=120s -count=1` (no `-short`).

**Skip behavior:** Tests in `real_models_integration_test.go` call `t.Skip()` when `testing.Short()` is true. Regular `go test ./...` or `make test` uses `-short` by default, so these tests are skipped. Use `make test-real-models` (or `go test -run RealModels ./internal/tools/...` without `-short`) to run them.

**Backend requirements:** At least one local backend must be available:

| Backend | Platform | Setup |
|---------|----------|-------|
| **Apple Foundation Models (FM)** | macOS Apple Silicon, CGO build | Built-in; use `make build-apple-fm` |
| **Ollama** | Any (GPU/RAM recommended) | Install [Ollama](https://ollama.ai/), run `ollama serve`, pull a model |

For **faster Ollama-based tests**, pull a light model (e.g. `ollama pull qwen2.5:1.5b`); tests default to `qwen2.5:1.5b`. For larger families, use a **quantized tag** (e.g. `qwen2.5:7b-q4_0`) to keep runs fast and memory use low. Set `OLLAMA_DEFAULT_MODEL` / `OLLAMA_CODE_MODEL` (and optionally `OLLAMA_TEST_MODEL` / `OLLAMA_TEST_CODE_MODEL`) to use your preferred models.

Tests use `DefaultFMProvider()` and Ollama paths as configured; they skip with a clear message if no backend is available.

---

## Benefits & Trade-offs

### Benefits

1. **Privacy** - All processing local, no data leakage
2. **Cost** - No API costs for routine operations
3. **Speed** - Fast inference on local hardware
4. **Offline** - Works without internet
5. **Iteration** - Can refine prompts multiple times cheaply
6. **Parallelization** - Can run multiple model calls simultaneously

### Trade-offs

1. **Hardware Requirements** - FM benefits from Apple Silicon + CGO; Ollama needs GPU/RAM for larger models
2. **Model Quality** - Local models may be less capable than cloud models
3. **Setup Complexity** - Need to manage model downloads and updates
4. **Maintenance** - Models need periodic updates
5. **Resource Usage** - Models consume CPU/GPU/RAM

### When to Use Local Models

**Use Local Models For:**
- ✅ Routine tasks (code review, documentation)
- ✅ Task breakdown and planning
- ✅ Prompt optimization
- ✅ Simple code generation
- ✅ Privacy-sensitive tasks

**Use Cloud AI For:**
- ✅ Complex reasoning tasks
- ✅ High-stakes decisions
- ✅ Tasks requiring latest knowledge
- ✅ When local models fail

---

## Next Steps

1. **Review and approve this design**
2. **Create Todo2 tasks for implementation**
3. **Set up development environment** (Ollama; optional Apple FM CGO build)
4. **Implement Phase 2** (Model Integration)
5. **Test with real tasks**
6. **Iterate based on results**

---

## References

- [Ollama Documentation](https://ollama.ai/docs)
- [CodeLlama Paper](https://ai.meta.com/research/publications/code-llama-open-foundation-models-for-code/)

---

**Status:** Ready for implementation planning and task creation.

