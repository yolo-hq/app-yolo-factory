-- 001_initial.up.sql
-- Factory schema: all 10 tables in FK-safe order

-- 1. factory_projects (no deps)
CREATE TABLE factory_projects (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    name TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    repo_url TEXT NOT NULL,
    local_path TEXT NOT NULL DEFAULT '',
    use_worktrees BOOLEAN DEFAULT false,
    default_branch TEXT NOT NULL DEFAULT 'main',
    maintenance_branches TEXT DEFAULT '[]',
    default_model TEXT NOT NULL DEFAULT 'sonnet',
    escalation_model TEXT NOT NULL DEFAULT 'opus',
    escalation_after_retries INTEGER NOT NULL DEFAULT 2,
    budget_per_task_usd REAL NOT NULL DEFAULT 2.0,
    budget_per_prd_usd REAL NOT NULL DEFAULT 20.0,
    budget_monthly_usd REAL NOT NULL DEFAULT 200.0,
    budget_warning_at REAL NOT NULL DEFAULT 0.8,
    spent_this_month_usd REAL NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    timeout_secs INTEGER NOT NULL DEFAULT 600,
    auto_merge BOOLEAN NOT NULL DEFAULT true,
    auto_start BOOLEAN NOT NULL DEFAULT false,
    push_failed_branches BOOLEAN NOT NULL DEFAULT false,
    setup_commands TEXT DEFAULT '[]',
    test_commands TEXT DEFAULT '["go build ./...","go test ./..."]'
);

-- 2. factory_prds (FK -> projects)
CREATE TABLE factory_prds (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    project_id TEXT NOT NULL REFERENCES factory_projects(id),
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    source TEXT NOT NULL DEFAULT 'manual',
    created_by TEXT NOT NULL DEFAULT 'human',
    body TEXT NOT NULL,
    acceptance_criteria TEXT NOT NULL,
    design_decisions TEXT DEFAULT '[]',
    total_tasks INTEGER DEFAULT 0,
    completed_tasks INTEGER DEFAULT 0,
    failed_tasks INTEGER DEFAULT 0,
    total_cost_usd REAL DEFAULT 0,
    approved_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- 3. factory_tasks (FK -> prds, projects)
CREATE TABLE factory_tasks (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    prd_id TEXT REFERENCES factory_prds(id),
    project_id TEXT REFERENCES factory_projects(id),
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    spec TEXT NOT NULL,
    acceptance_criteria TEXT NOT NULL,
    branch TEXT NOT NULL,
    model TEXT DEFAULT '',
    sequence INTEGER NOT NULL,
    depends_on TEXT DEFAULT '[]',
    run_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    cost_usd REAL DEFAULT 0,
    summary TEXT,
    commit_hash TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP
);

-- 4. factory_runs (FK -> tasks)
CREATE TABLE factory_runs (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    task_id TEXT NOT NULL REFERENCES factory_tasks(id),
    agent_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    model TEXT NOT NULL,
    session_id TEXT,
    session_name TEXT,
    escalated_model TEXT,
    cost_usd REAL DEFAULT 0,
    tokens_in INTEGER DEFAULT 0,
    tokens_out INTEGER DEFAULT 0,
    duration_ms INTEGER DEFAULT 0,
    num_turns INTEGER DEFAULT 0,
    commit_hash TEXT,
    branch_name TEXT,
    files_changed TEXT DEFAULT '[]',
    result TEXT,
    error TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- 5. factory_steps (FK -> runs)
CREATE TABLE factory_steps (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    run_id TEXT NOT NULL REFERENCES factory_runs(id),
    phase TEXT NOT NULL,
    skill TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    model TEXT NOT NULL,
    session_id TEXT,
    cost_usd REAL DEFAULT 0,
    tokens_in INTEGER DEFAULT 0,
    tokens_out INTEGER DEFAULT 0,
    duration_ms INTEGER DEFAULT 0,
    input_summary TEXT,
    output_summary TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- 6. factory_reviews (FK -> runs, tasks)
CREATE TABLE factory_reviews (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    run_id TEXT NOT NULL REFERENCES factory_runs(id),
    task_id TEXT NOT NULL REFERENCES factory_tasks(id),
    session_id TEXT,
    model TEXT NOT NULL,
    verdict TEXT NOT NULL,
    reasons TEXT DEFAULT '[]',
    anti_patterns TEXT DEFAULT '[]',
    criteria_results TEXT NOT NULL,
    suggestions TEXT DEFAULT '[]',
    cost_usd REAL DEFAULT 0
);

-- 7. factory_questions (FK -> tasks, runs)
CREATE TABLE factory_questions (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    task_id TEXT NOT NULL REFERENCES factory_tasks(id),
    run_id TEXT NOT NULL REFERENCES factory_runs(id),
    body TEXT NOT NULL,
    context TEXT,
    confidence TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    answer TEXT,
    answered_by TEXT,
    answer_session_id TEXT,
    answered_at TIMESTAMP
);

-- 8. factory_suggestions (FK -> projects, nullable FK -> tasks)
CREATE TABLE factory_suggestions (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    project_id TEXT NOT NULL REFERENCES factory_projects(id),
    source TEXT NOT NULL,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    priority TEXT NOT NULL DEFAULT 'medium',
    status TEXT NOT NULL DEFAULT 'pending',
    converted_task_id TEXT REFERENCES factory_tasks(id)
);

-- 9. factory_insights (FK -> projects nullable)
CREATE TABLE factory_insights (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    project_id TEXT REFERENCES factory_projects(id),
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    metric_data TEXT NOT NULL DEFAULT '{}',
    recommendation TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'medium',
    status TEXT NOT NULL DEFAULT 'pending'
);

-- 10. factory_lint_results (FK -> runs, tasks)
CREATE TABLE factory_lint_results (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    run_id TEXT NOT NULL REFERENCES factory_runs(id),
    task_id TEXT NOT NULL REFERENCES factory_tasks(id),
    passed BOOLEAN NOT NULL DEFAULT false,
    checks_run INTEGER NOT NULL DEFAULT 0,
    checks_passed INTEGER NOT NULL DEFAULT 0,
    checks_failed INTEGER NOT NULL DEFAULT 0,
    findings TEXT NOT NULL DEFAULT '[]'
);

-- Indexes
CREATE INDEX idx_factory_tasks_project_status ON factory_tasks(project_id, status);
CREATE INDEX idx_factory_tasks_prd_sequence ON factory_tasks(prd_id, sequence);
CREATE INDEX idx_factory_tasks_status ON factory_tasks(status);

CREATE INDEX idx_factory_runs_task_id ON factory_runs(task_id);
CREATE INDEX idx_factory_runs_status ON factory_runs(status);

CREATE INDEX idx_factory_steps_run_id ON factory_steps(run_id);

CREATE INDEX idx_factory_reviews_run_id ON factory_reviews(run_id);
CREATE INDEX idx_factory_reviews_task_id ON factory_reviews(task_id);

CREATE INDEX idx_factory_questions_status ON factory_questions(status);
CREATE INDEX idx_factory_questions_task_id ON factory_questions(task_id);

CREATE INDEX idx_factory_suggestions_project_status ON factory_suggestions(project_id, status);

CREATE INDEX idx_factory_insights_project_status ON factory_insights(project_id, status);
CREATE INDEX idx_factory_insights_category ON factory_insights(category);

CREATE INDEX idx_factory_lint_results_run_id ON factory_lint_results(run_id);
CREATE INDEX idx_factory_lint_results_task_id ON factory_lint_results(task_id);
