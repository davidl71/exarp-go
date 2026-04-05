ALTER TABLE task_activities ADD COLUMN activity_type_enum INTEGER NOT NULL DEFAULT 0;
ALTER TABLE task_execution_runs ADD COLUMN status_enum INTEGER NOT NULL DEFAULT 0;
ALTER TABLE task_verifications ADD COLUMN kind_enum INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_task_activities_type_enum ON task_activities(activity_type_enum);
CREATE INDEX IF NOT EXISTS idx_task_execution_runs_status_enum ON task_execution_runs(status_enum);
CREATE INDEX IF NOT EXISTS idx_task_verifications_kind_enum ON task_verifications(kind_enum);
