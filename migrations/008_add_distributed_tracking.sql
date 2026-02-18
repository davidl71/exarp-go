-- SQLite Migration: Add Distributed Tracking Fields
-- Version: 8
-- Description: Add assigned_to, host, agent for distributed tracking and aggregation
-- project_id already exists in schema. This migration adds assignee/host/agent fields.

-- Persistent assignee (who owns the task, distinct from lock assignee)
ALTER TABLE tasks ADD COLUMN assigned_to TEXT;

-- Hostname where task was created or last modified (for distributed aggregation)
ALTER TABLE tasks ADD COLUMN host TEXT;

-- Agent ID that created or last modified the task (e.g. general-Davids-Mac-mini-12345)
ALTER TABLE tasks ADD COLUMN agent TEXT;

-- Indexes for aggregation queries
CREATE INDEX IF NOT EXISTS idx_tasks_assigned_to ON tasks(assigned_to) WHERE assigned_to IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_host ON tasks(host) WHERE host IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent) WHERE agent IS NOT NULL;
