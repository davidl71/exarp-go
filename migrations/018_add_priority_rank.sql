-- 018_add_priority_rank.sql — Optional numeric sort key within the same named priority band.
-- Lower values sort earlier (after named-priority tiering in claim/backlog/list ordering).
ALTER TABLE tasks ADD COLUMN priority_rank INTEGER NOT NULL DEFAULT 0;
