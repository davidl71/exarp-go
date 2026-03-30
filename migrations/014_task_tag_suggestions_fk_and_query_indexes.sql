-- 014_task_tag_suggestions_fk_and_query_indexes.sql
-- Version: 14
-- Enforce task_tag_suggestions.task_id -> tasks(id) with ON DELETE CASCADE; add list/join indexes.

-- Drop orphaned suggestion rows (no matching task) so the rebuilt table satisfies FK.
DELETE FROM task_tag_suggestions
WHERE task_id NOT IN (SELECT id FROM tasks);

CREATE TABLE task_tag_suggestions_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    source TEXT NOT NULL,
    applied INTEGER DEFAULT 1,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(task_id, tag),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

INSERT INTO task_tag_suggestions_new (id, task_id, tag, source, applied, created_at)
SELECT id, task_id, tag, source, applied, created_at FROM task_tag_suggestions;

DROP TABLE task_tag_suggestions;

ALTER TABLE task_tag_suggestions_new RENAME TO task_tag_suggestions;

CREATE INDEX IF NOT EXISTS idx_task_tag_suggestions_task ON task_tag_suggestions(task_id);
CREATE INDEX IF NOT EXISTS idx_task_tag_suggestions_tag ON task_tag_suggestions(tag);

-- ListTasks: status filter + ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_tasks_status_created_at ON tasks(status, created_at DESC);

-- ListTasks with tag filter: lookup by tag then task_id
CREATE INDEX IF NOT EXISTS idx_task_tags_tag_task ON task_tags(tag, task_id);

-- Comments loaded per task, often scoped by type
CREATE INDEX IF NOT EXISTS idx_task_comments_task_type ON task_comments(task_id, comment_type);
