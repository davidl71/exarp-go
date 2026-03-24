-- 010_add_task_indexes.sql — Add indexes on host, assignee, and agent for query performance.

-- Index on host for host-specific queries
CREATE INDEX IF NOT EXISTS idx_tasks_host ON tasks(host);

-- Index on assignee for agent lock queries
CREATE INDEX IF NOT EXISTS idx_tasks_assignee ON tasks(assigned_to);

-- Index on agent for agent-specific queries
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent);