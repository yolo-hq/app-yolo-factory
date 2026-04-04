CREATE TABLE factory_lint_results (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES factory_runs(id),
    task_id TEXT NOT NULL REFERENCES factory_tasks(id),
    passed BOOLEAN NOT NULL DEFAULT false,
    checks_run INTEGER NOT NULL DEFAULT 0,
    checks_passed INTEGER NOT NULL DEFAULT 0,
    checks_failed INTEGER NOT NULL DEFAULT 0,
    findings TEXT NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
