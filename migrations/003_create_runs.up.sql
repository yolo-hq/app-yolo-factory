CREATE TABLE runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id),
    repo_id TEXT NOT NULL REFERENCES repos(id),
    agent TEXT NOT NULL DEFAULT 'claude-cli',
    model TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    cost REAL NOT NULL DEFAULT 0,
    duration INTEGER NOT NULL DEFAULT 0,
    log_url TEXT DEFAULT '',
    error TEXT DEFAULT '',
    commit_hash TEXT DEFAULT '',
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_runs_task_id ON runs(task_id);
CREATE INDEX idx_runs_status ON runs(status);
