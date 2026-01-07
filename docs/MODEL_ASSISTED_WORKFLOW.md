# Model-Assisted Workflow: CodeLlama, MLX, and Ollama Integration

**Date:** 2026-01-07  
**Status:** 📋 Design Phase  
**Purpose:** Leverage local models (CodeLlama/MLX/Ollama) for task breakdown, execution, and prompt optimization

---

## Executive Summary

This document outlines how to use local LLM models (CodeLlama via MLX, Ollama models) to enhance the Todo2 workflow by:
1. **Task Breakdown** - Decomposing complex tasks into manageable subtasks
2. **Easy Task Execution** - Automating routine/simple tasks using local models
3. **Prompt Optimization** - Iteratively refining prompts for better AI responses

**Key Benefits:**
- ✅ **Privacy** - All processing happens locally (no data sent to external APIs)
- ✅ **Cost Efficiency** - No API costs for routine operations
- ✅ **Speed** - Fast inference on Apple Silicon (MLX) or local GPU (Ollama)
- ✅ **Offline Capability** - Works without internet connection
- ✅ **Iterative Refinement** - Can optimize prompts through multiple iterations

---

## Research Findings

### CodeLlama (via MLX)

**What It Is:**
- CodeLlama is a collection of code-focused LLMs (7B-34B parameters)
- Available in MLX format for Apple Silicon optimization
- Designed for code synthesis, understanding, and analysis

**Capabilities:**
- Code generation and completion
- Code review and analysis
- Architecture assessment
- Documentation generation
- Task decomposition (breaking down complex problems)

**Current Usage in Project:**
- ✅ Already used for architecture analysis (`docs/MLX_ARCHITECTURE_ANALYSIS.md`)
- ✅ Integrated in `estimation` tool (MLX-enhanced task duration estimation)
- ✅ Available via `mlx` MCP tool

**Models Available:**
- `mlx-community/CodeLlama-7b-mlx` - 7B parameter model
- `mlx-community/CodeLlama-13b-mlx` - 13B parameter model
- `mlx-community/CodeLlama-34b-mlx` - 34B parameter model (larger, more capable)

### MLX Framework

**What It Is:**
- Apple's machine learning framework optimized for Apple Silicon
- Provides efficient inference on Neural Engine + Metal GPU
- Python API similar to NumPy

**Current Integration:**
- ✅ `mlx` tool available in MCP server
- ✅ Supports multiple models (Phi-3.5, CodeLlama, Mistral, Qwen)
- ✅ Used for code analysis and estimation

**Models Suitable for Task Work:**
- **CodeLlama** - Best for code-related tasks, architecture analysis
- **Phi-3.5** - Good for general reasoning, task breakdown
- **Mistral** - Balanced performance for various tasks
- **Qwen** - Strong for code and reasoning

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
- **codellama** - Code-focused, similar to MLX CodeLlama
- **mistral** - Fast and capable
- **qwen2.5** - Strong reasoning and code understanding

---

## Use Cases

### 1. Task Breakdown

**Problem:** Complex tasks need to be decomposed into manageable subtasks.

**Solution:** Use CodeLlama/MLX to analyze task descriptions and generate subtask breakdowns.

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
┌───────▼────────┐   ┌────────▼────────┐  ┌───▼────────┐  ┌───────▼────────┐
│  MLX Handler   │   │ Ollama Handler │  │ MLX Handler│  │ Ollama Handler │
│  (CodeLlama)   │   │  (Various)     │  │ (Phi-3.5)  │  │  (llama3.2)    │
└────────────────┘   └────────────────┘  └────────────┘  └────────────────┘
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

**Purpose:** Select the best model for a given task.

**Selection Logic:**
```go
type ModelType string

const (
    ModelMLXCodeLlama ModelType = "mlx-codellama"  // Code tasks
    ModelMLXPhi35     ModelType = "mlx-phi35"       // General reasoning
    ModelOllamaLlama  ModelType = "ollama-llama"    // General tasks
    ModelOllamaCode   ModelType = "ollama-codellama" // Code tasks
)

func SelectModel(taskType TaskType, requirements ModelRequirements) ModelType {
    // Select based on:
    // - Task type (code vs. general)
    // - Hardware availability (Apple Silicon for MLX)
    // - Model capabilities
    // - Performance requirements
}
```

**Selection Rules:**
- **Code tasks** → CodeLlama (MLX or Ollama)
- **General reasoning** → Phi-3.5 (MLX) or llama3.2 (Ollama)
- **Apple Silicon** → Prefer MLX (faster, more efficient)
- **Other platforms** → Use Ollama

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
- `estimation` - Already uses MLX, can be enhanced
- `mlx` - Add task-specific actions
- `ollama` - Add task-specific actions

### 3. Go Migration Considerations

**Current State:**
- MLX tools use Python bridge (Apple Silicon specific)
- Ollama can use HTTP API (cross-platform)

**Migration Strategy:**
- **Phase 1:** Keep MLX via Python bridge (already planned)
- **Phase 2:** Add Ollama HTTP client in Go (no bridge needed)
- **Phase 3:** Create Go model router and handlers
- **Phase 4:** Integrate with Todo2 workflow

---

## Implementation Plan

### Phase 1: Research & Design (Current)
- ✅ Research CodeLlama, MLX, Ollama capabilities
- ✅ Design architecture
- 📋 Create implementation tasks

### Phase 2: Model Integration
- [ ] Create model router component
- [ ] Implement MLX handler (via Python bridge)
- [ ] Implement Ollama HTTP client
- [ ] Add model selection logic

### Phase 3: Task Breakdown
- [ ] Implement task analyzer
- [ ] Create breakdown prompt templates
- [ ] Add breakdown result parser
- [ ] Integrate with Todo2 task creation

### Phase 4: Auto-Execution
- [ ] Implement task complexity classifier
- [ ] Create execution prompt templates
- [ ] Add change application logic
- [ ] Integrate with Todo2 execution flow

### Phase 5: Prompt Optimization
- [ ] Implement prompt analyzer
- [ ] Create optimization prompt templates
- [ ] Add iterative refinement loop
- [ ] Integrate with task creation workflow

### Phase 6: Testing & Documentation
- [ ] Unit tests for all components
- [ ] Integration tests with real models
- [ ] Performance benchmarks
- [ ] User documentation

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

1. **Hardware Requirements** - MLX needs Apple Silicon, Ollama needs GPU/RAM
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
3. **Set up development environment** (MLX, Ollama)
4. **Implement Phase 2** (Model Integration)
5. **Test with real tasks**
6. **Iterate based on results**

---

## References

- [MLX Framework](https://mlx-framework.org/)
- [MLX CodeLlama Models](https://huggingface.co/mlx-community/CodeLlama-7b-mlx)
- [Ollama Documentation](https://ollama.ai/docs)
- [CodeLlama Paper](https://ai.meta.com/research/publications/code-llama-open-foundation-models-for-code/)

---

**Status:** Ready for implementation planning and task creation.

