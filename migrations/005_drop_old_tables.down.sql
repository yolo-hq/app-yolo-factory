CREATE TABLE repos (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    local_path TEXT DEFAULT '',
    target_branch TEXT NOT NULL DEFAULT 'main',
    default_model TEXT NOT NULL DEFAULT 'sonnet',
    feedback_loops TEXT DEFAULT '[]',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL REFERENCES repos(id),
    title TEXT NOT NULL,
    body TEXT DEFAULT '',
    type TEXT NOT NULL DEFAULT 'auto',
    status TEXT NOT NULL DEFAULT 'queued',
    priority INTEGER NOT NULL DEFAULT 3,
    model TEXT DEFAULT '',
    labels TEXT DEFAULT '[]',
    parent_id TEXT REFERENCES tasks(id),
    depends_on TEXT DEFAULT '[]',
    cost REAL NOT NULL DEFAULT 0,
    run_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    timeout_secs INTEGER NOT NULL DEFAULT 600,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_tasks_repo_id ON tasks(repo_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_priority ON tasks(priority);

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

CREATE TABLE questions (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id),
    run_id TEXT NOT NULL REFERENCES runs(id),
    repo_id TEXT NOT NULL REFERENCES repos(id),
    status TEXT NOT NULL DEFAULT 'open',
    context TEXT DEFAULT '',
    tried TEXT DEFAULT '',
    body TEXT NOT NULL,
    resolution TEXT DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
CREATE INDEX idx_questions_task_id ON questions(task_id);
CREATE INDEX idx_questions_status ON questions(status);
