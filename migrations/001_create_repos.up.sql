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
