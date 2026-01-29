-- SQLite Migration: Tag Cache Tables
-- Version: 4
-- Created: 2026-01-29
-- Description: Add tables for caching discovered tags, canonical mappings, and file-based tag discoveries

-- ============================================================================
-- Tag Cache Tables
-- ============================================================================

-- Canonical tag mappings: Store canonical tag rules
CREATE TABLE IF NOT EXISTS canonical_tags (
    old_tag TEXT PRIMARY KEY,
    new_tag TEXT NOT NULL,
    category TEXT,  -- e.g., 'scorecard', 'type', 'domain', 'batch'
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Discovered tags from files: Cache tag discoveries from markdown files
CREATE TABLE IF NOT EXISTS discovered_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    file_hash TEXT,  -- MD5 hash of file content for cache invalidation
    tag TEXT NOT NULL,
    source TEXT NOT NULL,  -- 'hashtag', 'bracket', 'explicit', 'llm'
    llm_suggested INTEGER DEFAULT 0,  -- 1 if suggested by LLM
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(file_path, tag)
);

-- Tag frequency cache: Store tag usage statistics
CREATE TABLE IF NOT EXISTS tag_frequency (
    tag TEXT PRIMARY KEY,
    count INTEGER NOT NULL DEFAULT 0,
    last_seen_at INTEGER,
    is_canonical INTEGER DEFAULT 0,  -- 1 if this is a canonical tag
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- File-to-task tag matches: Cache which files provide tags for which tasks
CREATE TABLE IF NOT EXISTS file_task_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    task_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    applied INTEGER DEFAULT 0,  -- 1 if tag was applied to task
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(file_path, task_id, tag),
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- ============================================================================
-- Indexes for Performance
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_canonical_tags_new ON canonical_tags(new_tag);
CREATE INDEX IF NOT EXISTS idx_canonical_tags_category ON canonical_tags(category);

CREATE INDEX IF NOT EXISTS idx_discovered_tags_file ON discovered_tags(file_path);
CREATE INDEX IF NOT EXISTS idx_discovered_tags_tag ON discovered_tags(tag);
CREATE INDEX IF NOT EXISTS idx_discovered_tags_source ON discovered_tags(source);

CREATE INDEX IF NOT EXISTS idx_tag_frequency_count ON tag_frequency(count DESC);
CREATE INDEX IF NOT EXISTS idx_tag_frequency_canonical ON tag_frequency(is_canonical);

CREATE INDEX IF NOT EXISTS idx_file_task_tags_file ON file_task_tags(file_path);
CREATE INDEX IF NOT EXISTS idx_file_task_tags_task ON file_task_tags(task_id);
CREATE INDEX IF NOT EXISTS idx_file_task_tags_tag ON file_task_tags(tag);

-- ============================================================================
-- Initial Data: Populate canonical tags
-- ============================================================================

-- Scorecard: testing
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('testing-validation', 'testing', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('test', 'testing', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('tests', 'testing', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('validation', 'testing', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('integration', 'testing', 'scorecard');

-- Scorecard: documentation
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('documentation', 'docs', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('doc', 'docs', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('reporting', 'docs', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('inventory', 'docs', 'scorecard');

-- Scorecard: security
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('security-scan', 'security', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('vulnerability', 'security', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('vulnerabilities', 'security', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('audit', 'security', 'scorecard');

-- Scorecard: build/CI
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('cicd', 'build', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('ci', 'build', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('cd', 'build', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('ci-cd', 'build', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('github-actions', 'build', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('automation', 'build', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('uv', 'build', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('setup', 'build', 'scorecard');

-- Scorecard: performance
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('optimization', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('profiling', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('pgo', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('gc', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('memory', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('caching', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('file-io', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('hot-reload', 'performance', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('metrics', 'performance', 'scorecard');

-- Scorecard: linting
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('spelling', 'linting', 'scorecard');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('cspell', 'linting', 'scorecard');

-- Type: bug
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('bug-fix', 'bug', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('fix', 'bug', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('bugfix', 'bug', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('json-parsing', 'bug', 'type');

-- Type: feature
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('enhancement', 'feature', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('enhancements', 'feature', 'type');

-- Type: refactor
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('refactoring', 'refactor', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('code-quality', 'refactor', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('cleanup', 'refactor', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('consolidation', 'refactor', 'type');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('naming', 'refactor', 'type');

-- Domain: migration
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('bridge', 'migration', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('legacy-code', 'migration', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('python-cleanup', 'migration', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('native', 'migration', 'domain');

-- Domain: config
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('configuration', 'config', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('configuration-system', 'config', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('pyproject', 'config', 'domain');

-- Domain: CLI
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('tui', 'cli', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('bubble-tea', 'cli', 'domain');

-- Domain: MCP
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('mcp-config', 'mcp', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('mcp-go-core', 'mcp', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('mcp-go-core-extraction', 'mcp', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('go-sdk', 'mcp', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('framework-agnostic', 'mcp', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('framework-improvements', 'mcp', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('todo-sync', 'mcp', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('protobuf-integration', 'mcp', 'domain');

-- Domain: LLM/AI
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('mlx', 'llm', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('apple-foundation-models', 'llm', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('apple-silicon', 'llm', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('npu', 'llm', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('swift', 'llm', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('devwisdom-go', 'llm', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('multi-agent', 'llm', 'domain');

-- Domain: database
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('sqlite', 'database', 'domain');

-- Domain: concurrency
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('goroutines', 'concurrency', 'domain');

-- Domain: git
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('git-hooks', 'git', 'domain');

-- Domain: planning
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('strategy', 'planning', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('coordination', 'planning', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('task-breakdown', 'planning', 'domain');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('crew-roles', 'planning', 'domain');

-- Batch/phase normalization
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('batch-1', 'batch', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('batch-2', 'batch', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('batch1', 'batch', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('batch2', 'batch', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('batch3', 'batch', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('phase-1', 'phase', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('phase-2', 'phase', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('phase-3', 'phase', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('phase-4', 'phase', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('phase-5', 'phase', 'batch');
INSERT OR IGNORE INTO canonical_tags (old_tag, new_tag, category) VALUES ('phase-6', 'phase', 'batch');

-- Populate initial tag frequency for canonical tags
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('testing', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('docs', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('security', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('build', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('performance', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('linting', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('bug', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('feature', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('refactor', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('migration', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('config', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('cli', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('mcp', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('llm', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('database', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('concurrency', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('git', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('planning', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('batch', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('phase', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('epic', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('workflow', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('research', 0, 1);
INSERT OR IGNORE INTO tag_frequency (tag, count, is_canonical) VALUES ('analysis', 0, 1);
