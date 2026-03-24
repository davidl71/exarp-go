-- 013_add_automation_runs.sql
-- Track automation run executions for overlap detection and schedule metadata.

CREATE TABLE IF NOT EXISTS automation_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_label TEXT NOT NULL,
    action TEXT NOT NULL,
    pid INTEGER NOT NULL,
    host TEXT,
    status TEXT NOT NULL,
    started_at INTEGER NOT NULL,
    ended_at INTEGER,
    error_text TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_automation_runs_pid ON automation_runs(pid);
CREATE INDEX IF NOT EXISTS idx_automation_runs_action_status ON automation_runs(action, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_automation_runs_active_label ON automation_runs(schedule_label) WHERE status = 'running';
