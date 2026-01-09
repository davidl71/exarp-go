# SQLite Schema Design Document

**Created:** 2026-01-09  
**Status:** Complete  
**Task:** sqlite-migration-1

## Overview

This document describes the database schema design for migrating Todo2 tasks from JSON file storage to SQLite database.

## Design Principles

1. **Preserve Existing Data:** All fields from JSON structure are preserved
2. **Normalize Relationships:** Tags and dependencies in separate tables
3. **Maintain History:** Change tracking, comments, and activities preserved
4. **Performance:** Indexes on frequently queried columns
5. **Concurrency:** WAL mode and busy timeout for concurrent access

## Schema Structure

### 7 Tables

1. **tasks** - Main task storage (15 fields)
2. **task_tags** - Many-to-many tags relationship
3. **task_dependencies** - Many-to-many dependencies relationship
4. **task_changes** - Change history tracking
5. **task_comments** - Task comments storage
6. **task_activities** - Activity log
7. **schema_migrations** - Migration version tracking

### Key Design Decisions

1. **Primary Keys:** TEXT (preserve existing task IDs like "T-0", "T-NaN")
2. **Timestamps:** TEXT (ISO 8601) for compatibility with JSON format
3. **Booleans:** INTEGER (0/1) - SQLite standard
4. **Metadata:** TEXT (JSON blob) for flexible storage
5. **Foreign Keys:** Enabled with CASCADE deletes
6. **Indexes:** 15 indexes for common query patterns

## Field Mappings

### Todo2Task Struct → tasks Table

| Struct Field | SQL Column | Type | Notes |
|--------------|------------|------|-------|
| ID | id | TEXT PRIMARY KEY | Preserve existing IDs |
| Content | content | TEXT | |
| LongDescription | long_description | TEXT | |
| Status | status | TEXT NOT NULL | Default 'Todo' |
| Priority | priority | TEXT | |
| Tags | (separate table) | task_tags | Many-to-many |
| Dependencies | (separate table) | task_dependencies | Many-to-many |
| Completed | completed | INTEGER | 0/1 boolean |
| Metadata | metadata | TEXT | JSON blob |

### Additional Fields from JSON

| JSON Field | SQL Column | Type | Notes |
|------------|------------|------|-------|
| name | name | TEXT | |
| created | created | TEXT | ISO 8601 |
| lastModified | last_modified | TEXT | ISO 8601 |
| completedAt | completed_at | TEXT | ISO 8601 |
| taskNumber | task_number | INTEGER | |
| estimatedHours | estimated_hours | REAL | |
| actualHours | actual_hours | REAL | |
| project_id | project_id | TEXT | |

## Relationships

### One-to-Many

- tasks → task_changes (one task, many changes)
- tasks → task_comments (one task, many comments)
- tasks → task_activities (one task, many activities)

### Many-to-Many

- tasks ↔ task_tags (via task_tags table)
- tasks ↔ task_dependencies (via task_dependencies table)

## Indexes

### Performance Indexes

**tasks table:**
- `idx_tasks_status` - Filter by status
- `idx_tasks_priority` - Filter by priority
- `idx_tasks_created` - Sort by creation time
- `idx_tasks_completed` - Filter completed tasks
- `idx_tasks_last_modified` - Sort by modification time
- `idx_tasks_project` - Filter by project

**task_tags table:**
- `idx_task_tags_task` - Get tags for task
- `idx_task_tags_tag` - Get tasks by tag

**task_dependencies table:**
- `idx_task_deps_task` - Get dependencies for task
- `idx_task_deps_depends` - Get dependents (reverse lookup)

**Other tables:**
- Change, comment, activity indexes for efficient queries

## SQLite Configuration

### PRAGMA Settings

```sql
PRAGMA foreign_keys = ON;        -- Enable foreign key constraints
PRAGMA journal_mode = WAL;       -- Write-Ahead Logging for concurrency
PRAGMA synchronous = NORMAL;     -- Safe with WAL, better performance
PRAGMA busy_timeout = 30000;     -- 30 second timeout for concurrent access
```

## Migration Strategy

1. **Backup:** Create backup of `.todo2/state.todo2.json`
2. **Create Schema:** Run `migrations/001_initial_schema.sql`
3. **Migrate Data:** Convert JSON to SQL INSERT statements
4. **Validate:** Compare counts and verify data integrity
5. **Rollback:** Support rollback if migration fails

## Next Steps

1. ✅ Schema design complete
2. ✅ Initial schema SQL file created
3. ⏭️ Create database connection package
4. ⏭️ Implement migration system
5. ⏭️ Create data migration tool

