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
