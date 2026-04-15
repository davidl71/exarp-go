# Task Tools Comparison: `task_analysis`, `task_discovery`, `task_workflow`

**Date:** 2026-01-08  
**Purpose:** Compare the three task-related tools to understand their differences, similarities, and use cases

---

## Executive Summary

| Tool | Primary Purpose | Implementation | Actions |
|------|----------------|-----------------|---------|
| **`task_analysis`** | Analyze existing tasks | Python Bridge | `duplicates`, `tags`, `hierarchy`, `dependencies`, `parallelization` |
| **`task_discovery`** | Find tasks from code/docs | Python Bridge | `comments`, `markdown`, `orphans`, `all` |
| **`task_workflow`** | Manage task lifecycle | Python Bridge | `sync`, `approve`, `clarify`, `clarity`, `cleanup` |

**Key Insight:** These three tools work together in a task management workflow:
1. **Discover** tasks (`task_discovery`) → Find tasks from various sources
2. **Analyze** tasks (`task_analysis`) → Understand task structure and quality
3. **Manage** tasks (`task_workflow`) → Handle task lifecycle and workflow

---

## Detailed Comparison

### 1. **Tool Architecture**

#### `task_analysis` Tool
- **Type:** Analysis tool (read-only analysis)
- **Implementation:** Python Bridge
- **Location:** `project_management_automation/tools/consolidated_analysis.py`
- **Actions:** 5 analysis actions
- **Output:** Analysis reports, recommendations

#### `task_discovery` Tool
- **Type:** Discovery tool (find/create tasks)
- **Implementation:** Python Bridge
- **Location:** `project_management_automation/tools/consolidated_analysis.py`
- **Actions:** 4 discovery actions
- **Output:** Found tasks, optionally creates tasks

#### `task_workflow` Tool
- **Type:** Workflow management tool (modify tasks)
- **Implementation:** Python Bridge
- **Location:** `project_management_automation/tools/consolidated_workflow.py`
- **Actions:** 5 workflow actions
- **Output:** Task updates, workflow status

---

### 2. **Functionality Comparison**

#### `task_analysis` Tool

**Purpose:** Analyze existing tasks for quality, structure, and optimization opportunities

**Actions:**

1. **`action="duplicates"`** - Find duplicate tasks
   - Detects similar tasks using similarity threshold (default: 0.85)
   - Can auto-fix duplicates (merge/delete)
   - Output: List of duplicate pairs with similarity scores

2. **`action="tags"`** - Analyze and consolidate tags
   - Identifies tag patterns and inconsistencies
   - Suggests tag consolidation
   - Can remove specified tags
   - Output: Tag analysis and recommendations

3. **`action="hierarchy"`** - Analyze task hierarchy
   - Examines parent-child relationships
   - Identifies hierarchy issues (orphans, cycles)
   - Suggests hierarchy improvements
   - Output: Hierarchy structure and recommendations

4. **`action="dependencies"`** - Analyze task dependencies
   - Maps dependency chains
   - Identifies circular dependencies
   - Finds blocked tasks
   - Output: Dependency graph and issues

5. **`action="parallelization"`** - Find parallelization opportunities
   - Identifies tasks that can run in parallel
   - Analyzes dependency constraints
   - Suggests parallel execution strategies
   - Output: Parallelization recommendations

**Key Features:**
- ✅ Read-only analysis (doesn't modify tasks by default)
- ✅ Dry-run mode (preview changes before applying)
- ✅ Auto-fix option (for duplicates)
- ✅ Custom rules support
- ✅ Multiple output formats (text, JSON)

**Use Cases:**
- Clean up duplicate tasks
- Optimize task structure
- Identify blocked tasks
- Plan parallel execution
- Improve task organization

---

#### `task_discovery` Tool

**Purpose:** Find tasks from various sources (code comments, markdown, orphaned tasks)

**Actions:**

1. **`action="comments"`** - Find tasks in code comments
   - Scans code files for TODO/FIXME/HACK comments
   - Extracts task information from comments
   - Supports custom file patterns
   - Output: List of discovered tasks

2. **`action="markdown"`** - Find tasks in markdown files
   - Scans markdown files for task lists
   - Extracts tasks from documentation
   - Supports custom doc paths
   - Output: List of discovered tasks

3. **`action="orphans"`** - Find orphaned tasks
   - Identifies tasks without proper structure
   - Finds tasks missing dependencies
   - Detects incomplete task definitions
   - Output: List of orphaned tasks

4. **`action="all"`** - Comprehensive discovery
   - Combines all discovery methods
   - Scans comments, markdown, and orphans
   - Most comprehensive option
   - Output: Combined list of all discovered tasks

**Key Features:**
- ✅ Multiple discovery sources
- ✅ Optional task creation (`create_tasks` parameter)
- ✅ Custom file patterns
- ✅ FIXME inclusion toggle
- ✅ Output to file support

**Use Cases:**
- Find tasks in code comments
- Extract tasks from documentation
- Identify missing tasks
- Discover orphaned tasks
- Create tasks from discovered items

---

#### `task_workflow` Tool

**Purpose:** Manage task lifecycle and workflow operations

**Actions:**

1. **`action="sync"`** - Sync tasks across systems
   - Synchronizes tasks between Todo2 and other systems
   - Updates task status
   - Resolves conflicts
   - Output: Sync results and changes

2. **`action="approve"`** - Batch approve tasks
   - Approves tasks in "Review" status
   - Moves tasks to "Todo" or specified status
   - Supports tag filtering
   - Output: Approval results

3. **`action="clarify"`** - Handle task clarification
   - Lists tasks needing clarification
   - Adds clarification questions
   - Resolves clarification requests
   - Sub-actions: `list`, `add`, `resolve`
   - Output: Clarification status

4. **`action="clarity"`** - Improve task clarity
   - Analyzes task descriptions
   - Suggests clarity improvements
   - Uses MLX for task hour estimation
   - Output: Clarity recommendations

5. **`action="cleanup"`** - Clean up stale tasks
   - Identifies stale tasks (default: 2 hours)
   - Removes or updates stale tasks
   - Supports custom stale threshold
   - Output: Cleanup results

**Key Features:**
- ✅ Task modification (changes task state)
- ✅ Batch operations
- ✅ Status transitions
- ✅ Tag filtering
- ✅ Auto-apply option
- ✅ Dry-run mode

**Use Cases:**
- Sync tasks between systems
- Approve reviewed tasks
- Handle task clarifications
- Improve task descriptions
- Clean up stale tasks

---

### 3. **Workflow Integration**

#### Typical Task Management Workflow

```
1. DISCOVER → task_discovery(action="all")
   ↓
   Find tasks from code, docs, orphans
   
2. ANALYZE → task_analysis(action="duplicates")
   ↓
   Clean up duplicates, analyze structure
   
3. ANALYZE → task_analysis(action="dependencies")
   ↓
   Understand dependencies, find blockers
   
4. WORKFLOW → task_workflow(action="sync")
   ↓
   Sync tasks across systems
   
5. WORKFLOW → task_workflow(action="approve")
   ↓
   Approve ready tasks
   
6. WORKFLOW → task_workflow(action="clarify")
   ↓
   Handle clarification requests
```

---

### 4. **Action Matrix**

| Action | `task_analysis` | `task_discovery` | `task_workflow` |
|--------|----------------|------------------|-----------------|
| **Read Tasks** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Modify Tasks** | ⚠️ Optional (auto_fix) | ⚠️ Optional (create_tasks) | ✅ Yes |
| **Create Tasks** | ❌ No | ✅ Yes | ❌ No |
| **Delete Tasks** | ⚠️ Optional (duplicates) | ❌ No | ⚠️ Optional (cleanup) |
| **Update Status** | ❌ No | ❌ No | ✅ Yes |
| **Analyze Structure** | ✅ Yes | ❌ No | ❌ No |
| **Find Tasks** | ❌ No | ✅ Yes | ❌ No |
| **Sync Systems** | ❌ No | ❌ No | ✅ Yes |

---

### 5. **Implementation Status**

| Tool | Implementation | Migration Status | AI Usage |
|------|---------------|------------------|----------|
| `task_analysis` | Python Bridge | ⚠️ Not migrated | ❌ None (candidate for Apple FM) |
| `task_discovery` | Python Bridge | ⚠️ Not migrated | ❌ None (candidate for Apple FM) |
| `task_workflow` | Python Bridge | ⚠️ Not migrated | ✅ MLX (clarity action) |

**Migration Opportunities:**
- All three tools are Python Bridge (could migrate to native Go)
- `task_analysis` (hierarchy) is a candidate for Apple Foundation Models
- `task_workflow` (clarity) already uses MLX

---

### 6. **Use Case Examples**

#### Example 1: Clean Up Task Backlog

```python
# Step 1: Discover all tasks
discovered = task_discovery(action="all")

# Step 2: Find duplicates
duplicates = task_analysis(action="duplicates", similarity_threshold=0.85)

# Step 3: Analyze hierarchy
hierarchy = task_analysis(action="hierarchy")

# Step 4: Clean up stale tasks
cleanup = task_workflow(action="cleanup", stale_threshold_hours=24)
```

#### Example 2: Prepare for Sprint

```python
# Step 1: Find tasks in code
code_tasks = task_discovery(action="comments", include_fixme=True)

# Step 2: Analyze dependencies
deps = task_analysis(action="dependencies")

# Step 3: Find parallelization opportunities
parallel = task_analysis(action="parallelization")

# Step 4: Sync and approve ready tasks
sync = task_workflow(action="sync")
approve = task_workflow(action="approve", status="Review")
```

#### Example 3: Improve Task Quality

```python
# Step 1: Analyze tags
tags = task_analysis(action="tags", include_recommendations=True)

# Step 2: Improve clarity
clarity = task_workflow(action="clarity", task_id="T-123")

# Step 3: Handle clarifications
clarify = task_workflow(action="clarify", sub_action="list")
```

---

### 7. **Parameter Comparison**

#### Common Parameters

| Parameter | `task_analysis` | `task_discovery` | `task_workflow` |
|-----------|----------------|------------------|-----------------|
| `action` | ✅ Required | ✅ Required | ✅ Required |
| `dry_run` | ✅ Yes (default: true) | ❌ No | ✅ Yes (default: false) |
| `output_path` | ✅ Yes | ✅ Yes | ✅ Yes |
| `output_format` | ✅ Yes (text/json) | ❌ No | ✅ Yes (text/json) |

#### Unique Parameters

**`task_analysis`:**
- `similarity_threshold` (0.85) - For duplicate detection
- `auto_fix` (false) - Auto-fix duplicates
- `custom_rules` - Custom analysis rules
- `remove_tags` - Tags to remove
- `include_recommendations` (true) - Include recommendations

**`task_discovery`:**
- `file_patterns` - Custom file patterns to scan
- `include_fixme` (true) - Include FIXME comments
- `doc_path` - Custom documentation path
- `create_tasks` (false) - Create tasks from discoveries

**`task_workflow`:**
- `status` ("Review") - Filter by status
- `new_status` ("Todo") - Target status
- `task_id` - Single task ID
- `task_ids` - Multiple task IDs
- `filter_tag` - Filter by tag
- `sub_action` ("list") - For clarify action
- `clarification_text` - Clarification question
- `decision` - Decision for clarification
- `stale_threshold_hours` (2.0) - Stale task threshold
- `auto_apply` (false) - Auto-apply changes
- `move_to_todo` (true) - Move to Todo after approval

---

### 8. **Output Comparison**

#### `task_analysis` Output

**Format:** Text or JSON

**Content:**
- Analysis results (duplicates, tags, hierarchy, etc.)
- Recommendations
- Statistics (counts, percentages)
- Actionable insights

**Example:**
```
Found 5 duplicate task pairs:
- T-10 and T-15 (similarity: 0.92)
- T-20 and T-25 (similarity: 0.88)
...

Recommendations:
- Merge T-10 and T-15
- Consolidate tags: "bug" and "bugfix" → "bug"
```

#### `task_discovery` Output

**Format:** Text

**Content:**
- List of discovered tasks
- Source information (file, line number)
- Task details (description, priority)
- Creation status (if create_tasks=true)

**Example:**
```
Discovered 12 tasks:

From comments:
- TODO: Fix authentication bug (src/auth.py:45)
- FIXME: Update API documentation (docs/api.md:12)

From markdown:
- Implement user dashboard (docs/roadmap.md)
- Add unit tests (docs/testing.md)
```

#### `task_workflow` Output

**Format:** Text or JSON

**Content:**
- Operation results (sync, approve, clarify, etc.)
- Task updates
- Status changes
- Workflow status

**Example:**
```
Synced 15 tasks:
- T-10: Status updated (Review → Todo)
- T-15: Dependencies resolved
...

Approved 8 tasks:
- T-20, T-21, T-22, ...
```

---

### 9. **AI Usage Comparison**

| Tool | AI Usage | Model | Purpose |
|------|----------|-------|---------|
| `task_analysis` | ❌ None | N/A | **Candidate for Apple FM** (hierarchy classification) |
| `task_discovery` | ❌ None | N/A | **Candidate for Apple FM** (semantic extraction) |
| `task_workflow` | ✅ MLX | MLX | Task hour estimation (clarity action) |

**Potential AI Enhancements:**

1. **`task_analysis` (hierarchy):**
   - Use Apple FM to classify tasks into hierarchy levels
   - Generate hierarchy recommendations

2. **`task_discovery`:**
   - Use Apple FM for semantic task extraction from comments
   - Classify discovered tasks by type/priority

3. **`task_workflow` (clarify):**
   - Use Apple FM to generate clarification questions
   - Auto-answer clarification requests

---

### 10. **Summary Table**

| Feature | `task_analysis` | `task_discovery` | `task_workflow` |
|---------|----------------|------------------|-----------------|
| **Primary Purpose** | Analyze tasks | Find tasks | Manage tasks |
| **Actions** | 5 (duplicates, tags, hierarchy, deps, parallel) | 4 (comments, markdown, orphans, all) | 5 (sync, approve, clarify, clarity, cleanup) |
| **Modifies Tasks** | ⚠️ Optional | ⚠️ Optional | ✅ Yes |
| **Creates Tasks** | ❌ No | ✅ Yes | ❌ No |
| **Implementation** | Python Bridge | Python Bridge | Python Bridge |
| **AI Usage** | ❌ None | ❌ None | ✅ MLX (clarity) |
| **Dry Run** | ✅ Yes | ❌ No | ✅ Yes |
| **Output Format** | Text/JSON | Text | Text/JSON |
| **Use Case** | Quality analysis | Task discovery | Lifecycle management |

---

## Recommendations

### For Users:

1. **Use `task_discovery` first** to find all tasks
2. **Use `task_analysis` second** to understand and improve task structure
3. **Use `task_workflow` third** to manage task lifecycle

### For Developers:

1. **Consider migrating to native Go** for better performance
2. **Add Apple FM support** to `task_analysis` (hierarchy) and `task_discovery` (semantic extraction)
3. **Enhance `task_workflow` (clarify)** with Apple FM for auto-question generation

---

## Conclusion

The three task tools work together to provide a complete task management solution:

- **`task_discovery`**: Find tasks from various sources
- **`task_analysis`**: Understand and improve task quality
- **`task_workflow`**: Manage task lifecycle and workflow

All three tools are currently Python Bridge implementations with opportunities for native Go migration and AI enhancement.

