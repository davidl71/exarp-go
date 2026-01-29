-- SQLite Migration: Task tag suggestions cache
-- Version: 5
-- Description: Cache task-level tag suggestions (action=tags) for reuse as LLM hints everywhere

-- Task-level tag suggestions: applied/suggested tags from action=tags (keyword + LLM semantic)
-- Used as quick hints for LLM semantic analysis of task name and description
CREATE TABLE IF NOT EXISTS task_tag_suggestions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    source TEXT NOT NULL,  -- 'keyword', 'llm_semantic', 'tags'
    applied INTEGER DEFAULT 1,  -- 1 if tag was applied to task
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(task_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_task_tag_suggestions_task ON task_tag_suggestions(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tag_suggestions_tag ON task_tag_suggestions(tag);
