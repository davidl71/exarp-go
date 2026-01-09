# SQLite Migration - Code Generation Strategy

**Created:** 2026-01-09  
**Approach:** Generate as much as possible, build less often, use Makefile CI  
**Goal:** Reduce manual coding by 60-70%

## Strategy Overview

Instead of writing all code manually, generate:
1. **SQL schema** from Go structs
2. **CRUD boilerplate** from struct definitions
3. **Test scaffolding** from function signatures
4. **Migration scripts** from schema changes
5. **Temporary shell scripts** for validation/testing (instead of building)

## What Can Be Generated

### 1. SQL Schema Generation

**From:** Go struct definitions  
**To:** SQL CREATE TABLE statements

**Tool:** Generate from `Todo2Task` struct in `internal/tools/todo2_utils.go`

```bash
# Generate schema from struct
go run scripts/generate-schema.go > migrations/001_initial_schema.sql
```

**Benefits:**
- Single source of truth (Go struct)
- Auto-generate indexes
- Auto-generate foreign keys
- No manual SQL writing

---

### 2. CRUD Function Generation

**From:** Struct definition + method signatures  
**To:** Complete CRUD implementations

**Template-Based Generation:**

```go
// Template for CreateTask
func CreateTask(task *Todo2Task) error {
    tx, err := db.Begin()
    if err != nil { return err }
    defer tx.Rollback()
    
    // Generated: Insert task
    _, err = tx.Exec(`INSERT INTO tasks (...) VALUES (...)`, ...)
    
    // Generated: Insert tags (loop)
    for _, tag := range task.Tags {
        _, err = tx.Exec(`INSERT INTO task_tags ...`, ...)
    }
    
    // Generated: Insert dependencies (loop)
    for _, dep := range task.Dependencies {
        _, err = tx.Exec(`INSERT INTO task_dependencies ...`, ...)
    }
    
    return tx.Commit()
}
```

**Generation Script:**
```bash
# Generate all CRUD functions
go run scripts/generate-crud.go \
    --struct Todo2Task \
    --output internal/database/tasks.go \
    --template templates/crud.go.tmpl
```

**Benefits:**
- Generate 80% of CRUD code
- Only customize transaction logic
- Consistent patterns
- Less manual coding

---

### 3. Test Scaffolding Generation

**From:** Function signatures  
**To:** Test function stubs with table-driven test structure

**Use Existing Tool:** `testing(action=suggest)` from exarp-go

```bash
# Generate test scaffolding
make test-suggest target=internal/database/tasks.go
# Or use exarp-go tool directly
./bin/exarp-go testing action=suggest target_file=internal/database/tasks.go
```

**Generated Test Structure:**
```go
func TestCreateTask(t *testing.T) {
    tests := []struct {
        name    string
        task    *Todo2Task
        wantErr bool
    }{
        {
            name: "valid task",
            task: &Todo2Task{...},
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateTask(tt.task)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Benefits:**
- Generate test structure
- Fill in test cases manually
- Consistent test patterns

---

### 4. Migration Script Generation

**From:** Current JSON structure  
**To:** Migration Go code

**Generation Script:**
```bash
# Analyze JSON and generate migration code
go run scripts/generate-migration.go \
    --input .todo2/state.todo2.json \
    --output internal/database/migrate_json.go
```

**Generated Code:**
- JSON parsing logic
- Data conversion functions
- Validation logic
- Error handling

**Benefits:**
- Auto-detect JSON structure
- Generate conversion logic
- Handle edge cases automatically

---

### 5. Temporary Shell Scripts (Instead of Building)

**Strategy:** Generate shell scripts for validation/testing instead of building full binary

**Use Cases:**

#### Schema Validation Script
```bash
# Generate: scripts/validate-schema.sh
#!/bin/bash
# Validate SQL schema syntax
sqlite3 .todo2/todo2.db < migrations/001_initial_schema.sql
if [ $? -eq 0 ]; then
    echo "✅ Schema valid"
else
    echo "❌ Schema invalid"
    exit 1
fi
```

#### Data Migration Test Script
```bash
# Generate: scripts/test-migration.sh
#!/bin/bash
# Test migration without building binary
go run -tags dev cmd/migrate-todo2/main.go --dry-run
```

#### Query Validation Script
```bash
# Generate: scripts/test-queries.sh
#!/bin/bash
# Test SQL queries directly
sqlite3 .todo2/todo2.db <<EOF
SELECT * FROM tasks WHERE status = 'Todo';
SELECT COUNT(*) FROM task_tags WHERE tag = 'sqlite';
EOF
```

**Benefits:**
- No build time
- Fast iteration
- Use existing Makefile targets
- Temporary (delete after validation)

---

## Generation Tools to Create

### 1. Schema Generator (`scripts/generate-schema.go`)

**Input:** Go struct  
**Output:** SQL CREATE TABLE statements

**Features:**
- Parse Go struct tags
- Generate indexes from struct fields
- Generate foreign keys from relationships
- Output to SQL file

**Usage:**
```bash
make generate-schema
# Or
go run scripts/generate-schema.go
```

---

### 2. CRUD Generator (`scripts/generate-crud.go`)

**Input:** Struct definition + template  
**Output:** CRUD function implementations

**Templates:**
- `templates/crud-create.go.tmpl`
- `templates/crud-read.go.tmpl`
- `templates/crud-update.go.tmpl`
- `templates/crud-delete.go.tmpl`

**Usage:**
```bash
make generate-crud
# Or
go run scripts/generate-crud.go --struct Todo2Task
```

---

### 3. Test Generator (`scripts/generate-tests.go`)

**Input:** Go file with functions  
**Output:** Test file with scaffolding

**Uses:** Existing `testing(action=suggest)` tool

**Usage:**
```bash
make generate-tests TARGET=internal/database/tasks.go
# Or use exarp-go
./bin/exarp-go testing action=suggest target_file=internal/database/tasks.go
```

---

### 4. Migration Generator (`scripts/generate-migration.go`)

**Input:** JSON file  
**Output:** Migration Go code

**Usage:**
```bash
make generate-migration
# Or
go run scripts/generate-migration.go --input .todo2/state.todo2.json
```

---

### 5. Shell Script Generator (`scripts/generate-scripts.sh`)

**Input:** Task requirements  
**Output:** Temporary shell scripts

**Usage:**
```bash
make generate-scripts TYPE=schema-validation
make generate-scripts TYPE=query-test
make generate-scripts TYPE=migration-dry-run
```

---

## Makefile Integration

### New Makefile Targets

```makefile
##@ Code Generation

generate-schema: ## Generate SQL schema from Go structs
	@echo "$(BLUE)Generating SQL schema...$(NC)"
	@go run scripts/generate-schema.go > migrations/001_initial_schema.sql
	@echo "$(GREEN)✅ Schema generated: migrations/001_initial_schema.sql$(NC)"

generate-crud: ## Generate CRUD functions from structs
	@echo "$(BLUE)Generating CRUD functions...$(NC)"
	@go run scripts/generate-crud.go --struct Todo2Task --output internal/database/tasks.go
	@echo "$(GREEN)✅ CRUD functions generated$(NC)"

generate-tests: ## Generate test scaffolding
	@echo "$(BLUE)Generating test scaffolding...$(NC)"
	@if [ -z "$(TARGET)" ]; then \
		echo "$(RED)Error: TARGET required. Use: make generate-tests TARGET=internal/database/tasks.go$(NC)"; \
		exit 1; \
	fi
	@./bin/exarp-go testing action=suggest target_file=$(TARGET) > $(TARGET)_test.go.tmp
	@echo "$(GREEN)✅ Test scaffolding generated$(NC)"

generate-migration: ## Generate migration code from JSON
	@echo "$(BLUE)Generating migration code...$(NC)"
	@go run scripts/generate-migration.go --input .todo2/state.todo2.json --output internal/database/migrate_json.go
	@echo "$(GREEN)✅ Migration code generated$(NC)"

generate-scripts: ## Generate temporary shell scripts
	@echo "$(BLUE)Generating shell scripts...$(NC)"
	@if [ -z "$(TYPE)" ]; then \
		echo "$(RED)Error: TYPE required. Use: make generate-scripts TYPE=schema-validation$(NC)"; \
		exit 1; \
	fi
	@go run scripts/generate-scripts.go --type $(TYPE) --output scripts/tmp/$(TYPE).sh
	@chmod +x scripts/tmp/$(TYPE).sh
	@echo "$(GREEN)✅ Script generated: scripts/tmp/$(TYPE).sh$(NC)"

generate-all: generate-schema generate-crud generate-migration ## Generate all code
	@echo "$(GREEN)✅ All code generation complete$(NC)"

##@ Validation (No Build Required)

validate-schema: ## Validate SQL schema syntax (no build)
	@echo "$(BLUE)Validating schema...$(NC)"
	@sqlite3 .todo2/todo2.db < migrations/001_initial_schema.sql 2>&1 || echo "$(YELLOW)⚠️  Database file doesn't exist yet (expected)$(NC)"
	@sqlite3 :memory: < migrations/001_initial_schema.sql && echo "$(GREEN)✅ Schema syntax valid$(NC)" || (echo "$(RED)❌ Schema syntax invalid$(NC)" && exit 1)

test-queries: ## Test SQL queries directly (no build)
	@echo "$(BLUE)Testing queries...$(NC)"
	@if [ ! -f .todo2/todo2.db ]; then \
		echo "$(YELLOW)⚠️  Database not created yet. Run migration first.$(NC)"; \
		exit 0; \
	fi
	@sqlite3 .todo2/todo2.db < scripts/tmp/test-queries.sql && echo "$(GREEN)✅ Queries work$(NC)"

migrate-dry-run: ## Test migration without committing (no build)
	@echo "$(BLUE)Running migration dry-run...$(NC)"
	@go run -tags dev cmd/migrate-todo2/main.go --dry-run

##@ Development Workflow

dev-generate: generate-all validate-schema ## Generate code and validate (fast iteration)
	@echo "$(GREEN)✅ Code generated and validated (no build required)$(NC)"

dev-test-queries: generate-scripts test-queries ## Generate query test script and run (no build)
	@echo "$(GREEN)✅ Query testing complete$(NC)"
```

---

## Workflow: Generate → Code → Test Sprint

### Phase 1: Generate Everything (No Testing)

```bash
# Generate all code at once
make generate-all

# This creates:
# - migrations/001_initial_schema.sql (from struct)
# - internal/database/tasks.go (CRUD boilerplate)
# - internal/database/migrate_json.go (migration code)
# - Test scaffolding files
```

**Time:** 5-10 minutes (one-time generation)  
**No Testing:** ✅ Just generate

---

### Phase 2: Code Sprint (No Testing)

**Work through all tasks without testing:**
- Customize generated CRUD functions
- Add transaction logic
- Handle edge cases
- Implement business logic
- Fix obvious issues

**Generated code handles:**
- ✅ SQL statements
- ✅ Struct scanning
- ✅ Basic error handling
- ✅ Transaction structure
- ✅ Loop patterns (tags, dependencies)

**Time:** 2-3 days of pure coding  
**No Testing:** ✅ Build momentum, don't interrupt flow

---

### Phase 3: Validation Sprint (No Build)

**After coding sprint, validate without building:**

```bash
# Validate schema syntax (no build)
make validate-schema

# Test queries directly (no build)
make generate-scripts TYPE=query-test
make test-queries

# Test migration logic (dry-run, no build)
make migrate-dry-run
```

**Time:** 10-15 minutes  
**No Build Required:** ✅ Use shell scripts

---

### Phase 4: Build & Test Sprint

**Build once, then run all tests:**

```bash
# Build once
make build

# Run all tests via Makefile
make test                    # All tests
make test-coverage          # With coverage
make test-integration       # Integration tests
make sanity-check           # Sanity checks
```

**Build Frequency:** Once per sprint (not per fix)  
**Test Frequency:** Once per sprint (batch testing)

---

## Code Generation Templates

### CRUD Create Template

```go
// Template: templates/crud-create.go.tmpl
func Create{{.StructName}}({{.VarName}} *{{.StructName}}) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Insert main record
    _, err = tx.Exec(`
        INSERT INTO {{.TableName}} (
            {{range .Fields}}{{.Name}},{{end}}
        ) VALUES (
            {{range .Fields}}?,{{end}}
        )
    `, {{range .Fields}}{{.VarName}}.{{.FieldName}},{{end}})
    if err != nil {
        return fmt.Errorf("insert {{.TableName}}: %w", err)
    }
    
    {{range .Relationships}}
    // Insert {{.Type}} relationships
    for _, {{.VarName}} := range {{$.VarName}}.{{.FieldName}} {
        _, err = tx.Exec(`
            INSERT INTO {{.TableName}} ({{.ForeignKey}}, {{.ValueColumn}})
            VALUES (?, ?)
        `, {{$.VarName}}.ID, {{.VarName}})
        if err != nil {
            return fmt.Errorf("insert {{.Type}}: %w", err)
        }
    }
    {{end}}
    
    return tx.Commit()
}
```

### CRUD Read Template

```go
// Template: templates/crud-read.go.tmpl
func Get{{.StructName}}(id string) (*{{.StructName}}, error) {
    var {{.VarName}} {{.StructName}}
    
    // Query main record
    err := db.QueryRow(`
        SELECT {{range .Fields}}{{.Name}},{{end}}
        FROM {{.TableName}}
        WHERE id = ?
    `, id).Scan({{range .Fields}}&{{$.VarName}}.{{.FieldName}},{{end}})
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("{{.TableName}} not found: %w", err)
        }
        return nil, fmt.Errorf("query {{.TableName}}: %w", err)
    }
    
    {{range .Relationships}}
    // Load {{.Type}} relationships
    rows, err := db.Query(`
        SELECT {{.ValueColumn}}
        FROM {{.TableName}}
        WHERE {{.ForeignKey}} = ?
    `, id)
    if err != nil {
        return nil, fmt.Errorf("query {{.Type}}: %w", err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var {{.VarName}} string
        if err := rows.Scan(&{{.VarName}}); err != nil {
            return nil, fmt.Errorf("scan {{.Type}}: %w", err)
        }
        {{$.VarName}}.{{.FieldName}} = append({{$.VarName}}.{{.FieldName}}, {{.VarName}})
    }
    {{end}}
    
    return &{{.VarName}}, nil
}
```

---

## Shell Script Templates

### Schema Validation Script

```bash
#!/bin/bash
# Generated: scripts/tmp/validate-schema.sh

set -e

DB_PATH=".todo2/todo2.db"
SCHEMA_PATH="migrations/001_initial_schema.sql"

echo "Validating SQL schema..."

# Test in memory database
sqlite3 :memory: < "$SCHEMA_PATH" && echo "✅ Schema syntax valid" || {
    echo "❌ Schema syntax invalid"
    exit 1
}

# Test with actual database (if exists)
if [ -f "$DB_PATH" ]; then
    sqlite3 "$DB_PATH" < "$SCHEMA_PATH" && echo "✅ Schema applies to existing database" || {
        echo "⚠️  Schema conflicts with existing database (may need migration)"
    }
fi
```

### Query Test Script

```bash
#!/bin/bash
# Generated: scripts/tmp/test-queries.sh

set -e

DB_PATH=".todo2/todo2.db"

if [ ! -f "$DB_PATH" ]; then
    echo "⚠️  Database not found. Run migration first."
    exit 0
fi

echo "Testing SQL queries..."

sqlite3 "$DB_PATH" <<EOF
-- Test: Get tasks by status
SELECT COUNT(*) as todo_count FROM tasks WHERE status = 'Todo';

-- Test: Get tasks by tag
SELECT COUNT(*) as tagged_count 
FROM tasks t
JOIN task_tags tt ON t.id = tt.task_id
WHERE tt.tag = 'sqlite';

-- Test: Get dependencies
SELECT COUNT(*) as deps_count FROM task_dependencies;
EOF

echo "✅ Queries executed successfully"
```

---

## Revised Timeline (With Generation + Batch Testing)

### Original Vibe Coding: 22-35 days
### With Code Generation: **12-18 days** (40-50% reduction)
### With Generation + Batch Testing: **10-14 days** (55-60% reduction)

**Breakdown:**

| Phase | Original | With Generation | With Batch Testing | Savings |
|-------|----------|-----------------|-------------------|---------|
| Task 1: Research | 2-3 days | **1-2 days** | **1-2 days** | -1 day |
| Task 2: Infrastructure | 4-6 days | **2-3 days** | **2-3 days** | -2-3 days |
| Task 3: CRUD | 10-15 days | **5-8 days** | **4-6 days** | -6-9 days |
| Task 4: Migration | 4-6 days | **2-3 days** | **2-3 days** | -2-3 days |
| Task 5: Simplification | 2-3 days | **2-3 days** | **1-2 days** | -1 day |
| **Total** | **22-35 days** | **12-18 days** | **10-14 days** | **-12-21 days** |

**Batch Testing Benefits:**
- No testing interruptions during coding
- Build once per sprint (not per fix)
- Test once per sprint (batch all tests)
- Faster flow state
- Less context switching

---

## Generation Workflow Example (Batch Testing)

### Sprint 1: Generate & Code (2-3 days, No Testing)

**Day 1: Generate Everything**
```bash
# Generate all code
make generate-all

# Creates:
# ✅ migrations/001_initial_schema.sql
# ✅ internal/database/tasks.go (80% complete)
# ✅ internal/database/migrate_json.go (70% complete)
# ✅ Test scaffolding files
```

**Day 2-3: Code Sprint (No Testing)**
```bash
# Pure coding - no testing interruptions
# Customize generated code:
# - Complex transaction logic
# - Edge case handling
# - Business logic
# - Error messages
# - All CRUD functions
# - All query functions
# - All relationship functions

# Build momentum, don't test yet!
```

**Time:** 2-3 days of pure coding  
**No Testing:** ✅ Don't interrupt flow

---

### Sprint 2: Validation & Testing (1 day)

**Morning: Quick Validation (No Build)**
```bash
# Validate without building
make validate-schema
make generate-scripts TYPE=query-test
make test-queries
make migrate-dry-run
```

**Afternoon: Build Once**
```bash
# Build once for the entire sprint
make build
```

**Evening: Test Sprint (All Tests)**
```bash
# Run all tests via Makefile
make test                    # All Go tests
make test-coverage          # Coverage report
make test-integration       # Integration tests
make sanity-check           # Sanity checks

# Fix any critical issues found
# Don't test each fix - batch fixes, then test again
```

**Time:** 1 day  
**Build Frequency:** Once  
**Test Frequency:** Once (batch)

---

### Sprint 3: Fix & Finalize (1-2 days)

**If tests found issues:**
```bash
# Fix all issues found in test sprint
# Don't test each fix - fix everything, then test

# After all fixes:
make build
make test
```

**Time:** 1-2 days  
**Approach:** Batch fixes, batch testing

---

## Benefits Summary

1. **60-70% less manual coding** - Generate boilerplate
2. **5-10x fewer builds** - Build once per sprint
3. **Batch testing** - Test once per sprint (not per fix)
4. **Faster flow state** - No testing interruptions
5. **Consistent patterns** - Generated code is uniform
6. **Single source of truth** - Go structs → SQL → CRUD
7. **Leverage existing CI** - Use Makefile targets
8. **Temporary scripts** - Delete after validation
9. **Less context switching** - Code in sprints, test in sprints

## Implementation Priority

1. **High Priority:**
   - Schema generator (biggest win)
   - CRUD generator (saves most time)
   - Makefile targets (leverage existing CI)
   - Batch test targets (test sprint workflow)

2. **Medium Priority:**
   - Test scaffolding generator
   - Migration code generator

3. **Low Priority:**
   - Shell script generator (nice to have)

## Batch Testing Strategy

### Sprint-Based Workflow

**Sprint 1: Generate & Code (2-3 days)**
- Generate all code
- Code all functions
- **No testing** - maintain flow state
- **No building** - use validation scripts

**Sprint 2: Validate & Test (1 day)**
- Quick validation (shell scripts)
- Build once
- Run all tests (batch)
- Document issues

**Sprint 3: Fix & Finalize (1-2 days)**
- Fix all issues (batch)
- Build once
- Test once (batch)
- Done

### Makefile Test Sprint Target

```makefile
test-sprint: build ## Run all tests in batch (sprint testing)
	@echo "$(BLUE)Running test sprint...$(NC)"
	@echo "$(BLUE)1. Go tests...$(NC)"
	@$(MAKE) test-go
	@echo "$(BLUE)2. Test coverage...$(NC)"
	@$(MAKE) test-coverage-go
	@echo "$(BLUE)3. Integration tests...$(NC)"
	@$(MAKE) test-integration
	@echo "$(BLUE)4. Sanity checks...$(NC)"
	@$(MAKE) sanity-check
	@echo "$(GREEN)✅ Test sprint complete$(NC)"

validate-sprint: ## Validate without building (pre-sprint)
	@echo "$(BLUE)Running validation sprint...$(NC)"
	@$(MAKE) validate-schema
	@$(MAKE) generate-scripts TYPE=query-test
	@$(MAKE) test-queries
	@$(MAKE) migrate-dry-run
	@echo "$(GREEN)✅ Validation sprint complete$(NC)"
```

## Next Steps

1. Create `scripts/generate-schema.go` - Parse Go struct → SQL
2. Create `scripts/generate-crud.go` - Generate CRUD from templates
3. Add Makefile targets for generation
4. Create template files
5. Test generation workflow

**Estimated Time to Create Generators:** 2-3 days  
**Time Saved:** 10-17 days  
**ROI:** 3-5x (worth it!)
