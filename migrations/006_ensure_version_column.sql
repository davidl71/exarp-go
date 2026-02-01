-- Ensure version column exists (T-1768324162836). Ignores duplicate if present.
ALTER TABLE tasks ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
