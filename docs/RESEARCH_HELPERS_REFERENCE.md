# Research Helpers Reference

**Date:** 2026-01-07  
**Status:** ✅ **Implemented**  
**Location:** `mcp_stdio_tools/research_helpers.py`

---

## Overview

Research helper functions provide a unified interface for parallel research execution using:
- **CodeLlama** (MLX/Ollama) - Code review and architecture analysis
- **Context7** - Library documentation retrieval
- **Tractatus Thinking** - Structured reasoning and logical decomposition

---

## CodeLlama Helpers

### `codellama_analyze_code(code, task_description, model, use_mlx)`

Analyze code using CodeLlama for code review.

**Parameters:**
- `code` (str): Code to analyze
- `task_description` (str): Context/task description
- `model` (str): Model name (default: "mlx-community/CodeLlama-7b-mlx")
- `use_mlx` (bool): Use MLX if True, Ollama if False

**Returns:**
```python
{
    "tool": "codellama",
    "model": "mlx-community/CodeLlama-7b-mlx",
    "analysis": "Code quality assessment",
    "findings": ["finding1", "finding2"],
    "recommendations": ["rec1", "rec2"]
}
```

**Example:**
```python
result = await codellama_analyze_code(
    code="func RegisterTool(...) { ... }",
    task_description="Review Go tool registration pattern",
    model="mlx-community/CodeLlama-7b-mlx"
)
```

### `codellama_analyze_architecture(architecture_doc, model)`

Analyze architecture documents using CodeLlama.

**Parameters:**
- `architecture_doc` (str): Architecture document or design
- `model` (str): Model name

**Returns:**
```python
{
    "tool": "codellama",
    "model": "mlx-community/CodeLlama-7b-mlx",
    "analysis": "Architecture assessment",
    "strengths": ["strength1", "strength2"],
    "weaknesses": ["weakness1", "weakness2"],
    "recommendations": ["rec1", "rec2"]
}
```

### `codellama_review_design(design, requirements, model)`

Review design against requirements.

**Parameters:**
- `design` (str): Design document or code
- `requirements` (str): Requirements to check
- `model` (str): Model name

**Returns:**
```python
{
    "tool": "codellama",
    "model": "mlx-community/CodeLlama-7b-mlx",
    "review": "Design review",
    "compliance": {"req1": True, "req2": False},
    "recommendations": ["rec1", "rec2"]
}
```

---

## Context7 Helpers

### `context7_resolve_library(query, library_name)`

Resolve library name to Context7 library ID.

**Parameters:**
- `query` (str): Search query
- `library_name` (str): Library name (e.g., "go-sdk")

**Returns:**
- Library ID (str) like "/modelcontextprotocol/go-sdk" or None

**Example:**
```python
library_id = await context7_resolve_library(
    query="Go MCP SDK",
    library_name="go-sdk"
)
# Returns: "/modelcontextprotocol/go-sdk"
```

### `context7_query_documentation(library_id, query)`

Query library documentation using Context7.

**Parameters:**
- `library_id` (str): Context7 library ID
- `query` (str): Documentation query

**Returns:**
```python
{
    "tool": "context7",
    "library_id": "/modelcontextprotocol/go-sdk",
    "query": "How to register tools?",
    "documentation": "Documentation content",
    "examples": ["example1", "example2"],
    "api_info": {"method": "RegisterTool", "params": [...]}
}
```

### `context7_get_library_info(library_name, query)`

Get comprehensive library information.

**Parameters:**
- `library_name` (str): Library name
- `query` (str): Information query (default: "library overview and usage")

**Returns:**
Complete library information including documentation

---

## Tractatus Thinking Helpers

### `tractatus_analyze_concept(concept, depth_limit)`

Analyze a concept using Tractatus Thinking.

**Parameters:**
- `concept` (str): Concept to analyze (e.g., "What is framework-agnostic design?")
- `depth_limit` (int): Maximum analysis depth (default: 5)

**Returns:**
```python
{
    "tool": "tractatus",
    "concept": "What is framework-agnostic design?",
    "session_id": "session_123",
    "analysis": "Concept analysis",
    "propositions": ["prop1", "prop2"],
    "logical_structure": {"structure": "..."}
}
```

**Example:**
```python
result = await tractatus_analyze_concept(
    concept="What is framework-agnostic MCP server design?",
    depth_limit=5
)
```

### `tractatus_decompose_problem(problem)`

Decompose a problem using Tractatus.

**Parameters:**
- `problem` (str): Problem statement

**Returns:**
Problem decomposition with components and relationships

### `tractatus_reason_about(statement)`

Perform logical reasoning about a statement.

**Parameters:**
- `statement` (str): Statement to reason about

**Returns:**
Reasoning results with logical structure

---

## Parallel Research Execution

### `execute_parallel_research(...)`

Execute research using all available tools in parallel.

**Parameters:**
- `task_description` (str): Task description
- `code` (str, optional): Code to analyze
- `architecture_doc` (str, optional): Architecture document
- `library_names` (List[str], optional): Libraries to research
- `concepts` (List[str], optional): Concepts to analyze
- `web_search_queries` (List[str], optional): Web search queries

**Returns:**
```python
{
    "task_description": "Task description",
    "codellama_results": [...],
    "context7_results": [...],
    "tractatus_results": [...],
    "errors": [...]
}
```

**Example:**
```python
results = await execute_parallel_research(
    task_description="Implement framework-agnostic design",
    code=existing_code,
    architecture_doc=architecture_doc,
    library_names=["go-sdk", "mcp-go"],
    concepts=["What is framework-agnostic design?"],
    web_search_queries=["Go framework abstraction patterns 2026"]
)
```

### `format_research_comment(research_results)`

Format research results as Todo2 comment.

**Parameters:**
- `research_results` (Dict): Results from `execute_parallel_research`

**Returns:**
Formatted markdown string for Todo2 `research_with_links` comment

**Example:**
```python
comment = format_research_comment(results)
# Use in Todo2: add_comments(todo_id, research_with_links, comment)
```

---

## Usage Examples

### Example 1: Code Review

```python
from mcp_stdio_tools.research_helpers import codellama_analyze_code

result = await codellama_analyze_code(
    code="""
    func RegisterTool(name string, handler ToolHandler) error {
        // implementation
    }
    """,
    task_description="Review Go tool registration pattern",
    model="mlx-community/CodeLlama-7b-mlx"
)
```

### Example 2: Library Documentation

```python
from mcp_stdio_tools.research_helpers import context7_get_library_info

result = await context7_get_library_info(
    library_name="go-sdk",
    query="How to register tools in Go MCP SDK?"
)
```

### Example 3: Concept Analysis

```python
from mcp_stdio_tools.research_helpers import tractatus_analyze_concept

result = await tractatus_analyze_concept(
    concept="What is framework-agnostic MCP server design?",
    depth_limit=5
)
```

### Example 4: Parallel Research

```python
from mcp_stdio_tools.research_helpers import (
    execute_parallel_research,
    format_research_comment
)

# Execute parallel research
results = await execute_parallel_research(
    task_description="Implement framework-agnostic design",
    code=existing_code,
    library_names=["go-sdk"],
    concepts=["What is framework-agnostic design?"]
)

# Format for Todo2
comment = format_research_comment(results)
```

---

## Integration with MCP Tools

These helpers are designed to integrate with MCP tools:

- **CodeLlama:** Uses `mlx` or `ollama` MCP tools
- **Context7:** Uses `context7_resolve_library_id` and `context7_query_docs` MCP tools
- **Tractatus:** Uses `tractatus_thinking` MCP tool

**Note:** Current implementation provides structure. Full integration requires calling actual MCP tools.

---

## See Also

- [PARALLEL_RESEARCH_WORKFLOW.md](./PARALLEL_RESEARCH_WORKFLOW.md) - Complete workflow guide
- [MODEL_ASSISTED_WORKFLOW.md](./MODEL_ASSISTED_WORKFLOW.md) - Model integration details

---

**Status:** ✅ **Helper functions created.** Ready for integration with MCP tools.

