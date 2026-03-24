CREATE TABLE IF NOT EXISTS task_execution_runs (
    run_id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    agent_id TEXT NOT NULL DEFAULT '',
    host TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'running',
    summary TEXT NOT NULL DEFAULT '',
    files_touched TEXT NOT NULL DEFAULT '[]',
    commands_run TEXT NOT NULL DEFAULT '[]',
    notes TEXT NOT NULL DEFAULT '',
    started_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    ended_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_execution_runs_task_started
    ON task_execution_runs(task_id, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_execution_runs_status_started
    ON task_execution_runs(status, started_at DESC);

CREATE TABLE IF NOT EXISTS task_verifications (
    verification_id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    run_id TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL,
    command TEXT NOT NULL DEFAULT '',
    result TEXT NOT NULL,
    details TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (run_id) REFERENCES task_execution_runs(run_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_task_verifications_task_created
    ON task_verifications(task_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_verifications_run_created
    ON task_verifications(run_id, created_at DESC);

CREATE TABLE IF NOT EXISTS task_progress_entries (
    progress_id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    run_id TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL,
    files TEXT NOT NULL DEFAULT '[]',
    remaining_work TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (run_id) REFERENCES task_execution_runs(run_id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_task_progress_entries_task_created
    ON task_progress_entries(task_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_progress_entries_run_created
    ON task_progress_entries(run_id, created_at DESC);
