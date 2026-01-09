# SQLite Migration - Research Notes

**Task:** sqlite-migration-1  
**Status:** In Progress  
**Started:** 2026-01-09

## Research: SQLite Go Best Practices

### Key Findings

1. **Driver Choice:**
   - `github.com/mattn/go-sqlite3` - CGO required, most popular
   - `modernc.org/sqlite` - Pure Go, no CGO, good alternative
   - Recommendation: Start with `go-sqlite3` for compatibility

2. **Connection Management:**
   - Use `database/sql` package (standard library)
   - Single connection per database file (SQLite limitation)
   - Use connection pooling for concurrent access
   - Set `_busy_timeout` pragma for concurrent writes

3. **Schema Design:**
   - Normalize data (separate tables for tags, dependencies)
   - Use foreign keys with CASCADE
   - Create indexes on frequently queried columns
   - Use INTEGER PRIMARY KEY for auto-increment

4. **Transactions:**
   - Always use transactions for multi-step operations
   - Use `BEGIN IMMEDIATE` for write transactions
   - Rollback on error, commit on success

5. **Performance:**
   - Use prepared statements for repeated queries
   - Enable WAL mode for better concurrency
   - Use `PRAGMA synchronous = NORMAL` for better performance
   - Index foreign keys and frequently filtered columns

## Schema Design from Todo2Task Struct

### Current Struct Analysis

```go
type Todo2Task struct {
    ID             string                 `json:"id"`
    Content        string                 `json:"content"`
    LongDescription string                `json:"long_description,omitempty"`
    Status         string                 `json:"status"`
    Priority       string                 `json:"priority,omitempty"`
    Tags           []string               `json:"tags,omitempty"`
    Dependencies   []string               `json:"dependencies,omitempty"`
    Completed      bool                   `json:"completed,omitempty"`
    Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
```

### Additional Fields from JSON Analysis

- `name` (string, optional)
- `created` (ISO 8601 string)
- `lastModified` (ISO 8601 string, optional)
- `completedAt` (ISO 8601 string, optional)
- `taskNumber` (integer, optional)
- `estimatedHours` (float, optional)
- `actualHours` (float, optional)
- `project_id` (string, optional)
- `changes` (array of change objects)
- `comments` (array of comment objects)
- `activities` (array of activity objects)

## Schema Design Decisions

### Tables Needed

1. **tasks** - Main task table
2. **task_tags** - Many-to-many relationship
3. **task_dependencies** - Many-to-many relationship
4. **task_changes** - Change history
5. **task_comments** - Task comments
6. **task_activities** - Activity log
7. **schema_migrations** - Migration tracking

### Key Design Decisions

1. **Primary Keys:** Use TEXT for task IDs (preserve existing IDs)
2. **Timestamps:** Store as TEXT (ISO 8601) for compatibility
3. **Metadata:** Store as JSON TEXT (SQLite JSON support)
4. **Relationships:** Separate tables with foreign keys
5. **Indexes:** Status, priority, project_id, tags, dependencies

## Schema Created

✅ **Initial schema SQL file created:** `migrations/001_initial_schema.sql`

**Key Features:**
- 7 tables: tasks, task_tags, task_dependencies, task_changes, task_comments, task_activities, schema_migrations
- Foreign keys with CASCADE deletes
- Comprehensive indexes for performance
- WAL mode enabled for concurrency
- Busy timeout for concurrent access

## Next Steps

1. ✅ Create initial schema SQL file - DONE
2. Design migration system
3. Document schema decisions
4. Create schema validation approach
5. Test schema with sample data

