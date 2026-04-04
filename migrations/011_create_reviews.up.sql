CREATE TABLE factory_reviews (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES factory_runs(id),
    task_id TEXT NOT NULL REFERENCES factory_tasks(id),
    session_id TEXT,
    model TEXT NOT NULL,
    verdict TEXT NOT NULL,
    reasons TEXT DEFAULT '[]',
    anti_patterns TEXT DEFAULT '[]',
    criteria_results TEXT NOT NULL,
    suggestions TEXT DEFAULT '[]',
    cost_usd REAL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
