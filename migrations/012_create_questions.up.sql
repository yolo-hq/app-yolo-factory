CREATE TABLE factory_questions (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES factory_tasks(id),
    run_id TEXT NOT NULL REFERENCES factory_runs(id),
    body TEXT NOT NULL,
    context TEXT,
    confidence TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    answer TEXT,
    answered_by TEXT,
    answer_session_id TEXT,
    answered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
