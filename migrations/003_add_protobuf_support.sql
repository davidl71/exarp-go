-- SQLite Migration: Add Protobuf Support for Tasks
-- Version: 3
-- Created: 2026-01-12
-- Description: Add protobuf column for task metadata while maintaining JSON compatibility

-- Note: These columns may already exist if migration was previously applied as version 2.
-- SQLite will return error if columns exist, which we handle in migration code.

-- Add protobuf metadata column (BLOB for binary protobuf data)
ALTER TABLE tasks ADD COLUMN metadata_protobuf BLOB;

-- Add format indicator to track which format is used
-- 'json' = JSON format (legacy), 'protobuf' = protobuf format (new)
ALTER TABLE tasks ADD COLUMN metadata_format TEXT DEFAULT 'json';

-- Create index on metadata_format for efficient queries
CREATE INDEX IF NOT EXISTS idx_tasks_metadata_format ON tasks(metadata_format);

-- Update existing tasks to have 'json' format
UPDATE tasks SET metadata_format = 'json' WHERE metadata_format IS NULL;
