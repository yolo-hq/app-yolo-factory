-- Auto-generated migration
-- Review carefully before applying

SET lock_timeout = '10s';
SET statement_timeout = '60s';

-- Create table factory_insights
CREATE TABLE IF NOT EXISTS factory_insights (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    project_id TEXT NOT NULL,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    metric_data TEXT NOT NULL DEFAULT '{}',
    recommendation TEXT NOT NULL DEFAULT '',
    priority TEXT NOT NULL DEFAULT 'medium',
    status TEXT NOT NULL DEFAULT 'pending'
);

-- Create table factory_lint_results
CREATE TABLE IF NOT EXISTS factory_lint_results (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    run_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    passed BOOLEAN NOT NULL DEFAULT false,
    checks_run INTEGER NOT NULL DEFAULT 0,
    checks_passed INTEGER NOT NULL DEFAULT 0,
    checks_failed INTEGER NOT NULL DEFAULT 0,
    findings TEXT NOT NULL DEFAULT '[]'
);

-- Create table factory_prds
CREATE TABLE IF NOT EXISTS factory_prds (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    source TEXT NOT NULL DEFAULT 'manual',
    created_by TEXT NOT NULL DEFAULT 'human',
    body TEXT NOT NULL,
    acceptance_criteria TEXT NOT NULL,
    design_decisions TEXT NOT NULL DEFAULT '[]',
    total_tasks INTEGER NOT NULL DEFAULT 0,
    completed_tasks INTEGER NOT NULL DEFAULT 0,
    failed_tasks INTEGER NOT NULL DEFAULT 0,
    total_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    approved_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

-- Create table factory_projects
CREATE TABLE IF NOT EXISTS factory_projects (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    name TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    repo_url TEXT NOT NULL,
    local_path TEXT NOT NULL DEFAULT '',
    use_worktrees BOOLEAN NOT NULL DEFAULT false,
    default_branch TEXT NOT NULL DEFAULT 'main',
    maintenance_branches TEXT NOT NULL DEFAULT '[]',
    default_model TEXT NOT NULL DEFAULT 'sonnet',
    escalation_model TEXT NOT NULL DEFAULT 'opus',
    escalation_after_retries INTEGER NOT NULL DEFAULT 2,
    budget_per_task_usd DOUBLE PRECISION NOT NULL DEFAULT 2.0,
    budget_per_prd_usd DOUBLE PRECISION NOT NULL DEFAULT 20.0,
    budget_monthly_usd DOUBLE PRECISION NOT NULL DEFAULT 200.0,
    budget_warning_at DOUBLE PRECISION NOT NULL DEFAULT 0.8,
    spent_this_month_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    timeout_secs INTEGER NOT NULL DEFAULT 600,
    auto_merge BOOLEAN NOT NULL DEFAULT true,
    auto_start BOOLEAN NOT NULL DEFAULT false,
    push_failed_branches BOOLEAN NOT NULL DEFAULT false,
    setup_commands TEXT NOT NULL DEFAULT '[]',
    test_commands TEXT NOT NULL DEFAULT '[]'
);

-- Create table factory_questions
CREATE TABLE IF NOT EXISTS factory_questions (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    task_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    body TEXT NOT NULL,
    context TEXT NOT NULL,
    confidence TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    answer TEXT NOT NULL,
    answered_by TEXT NOT NULL,
    answer_session_id TEXT NOT NULL,
    answered_at TIMESTAMPTZ
);

-- Create table factory_reviews
CREATE TABLE IF NOT EXISTS factory_reviews (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    run_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    model TEXT NOT NULL,
    verdict TEXT NOT NULL,
    reasons TEXT NOT NULL DEFAULT '[]',
    anti_patterns TEXT NOT NULL DEFAULT '[]',
    criteria_results TEXT NOT NULL,
    suggestions TEXT NOT NULL DEFAULT '[]',
    cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0
);

-- Create table factory_runs
CREATE TABLE IF NOT EXISTS factory_runs (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    task_id TEXT NOT NULL,
    agent_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    model TEXT NOT NULL,
    session_id TEXT NOT NULL,
    session_name TEXT NOT NULL,
    escalated_model TEXT NOT NULL,
    cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    tokens_in INTEGER NOT NULL DEFAULT 0,
    tokens_out INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    num_turns INTEGER NOT NULL DEFAULT 0,
    commit_hash TEXT NOT NULL,
    branch_name TEXT NOT NULL,
    files_changed TEXT NOT NULL DEFAULT '[]',
    result TEXT NOT NULL,
    error TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    completed_at TIMESTAMPTZ
);

-- Create table factory_steps
CREATE TABLE IF NOT EXISTS factory_steps (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    run_id TEXT NOT NULL,
    phase TEXT NOT NULL,
    skill TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    model TEXT NOT NULL,
    session_id TEXT NOT NULL,
    cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    tokens_in INTEGER NOT NULL DEFAULT 0,
    tokens_out INTEGER NOT NULL DEFAULT 0,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    input_summary TEXT NOT NULL,
    output_summary TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    completed_at TIMESTAMPTZ
);

-- Create table factory_suggestions
CREATE TABLE IF NOT EXISTS factory_suggestions (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    project_id TEXT NOT NULL,
    source TEXT NOT NULL,
    category TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    priority TEXT NOT NULL DEFAULT 'medium',
    status TEXT NOT NULL DEFAULT 'pending',
    converted_task_id TEXT NOT NULL
);

-- Create table factory_tasks
CREATE TABLE IF NOT EXISTS factory_tasks (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    prd_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    spec TEXT NOT NULL,
    acceptance_criteria TEXT NOT NULL,
    branch TEXT NOT NULL,
    model TEXT NOT NULL DEFAULT '',
    sequence INTEGER NOT NULL,
    depends_on TEXT NOT NULL DEFAULT '[]',
    run_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    summary TEXT NOT NULL,
    commit_hash TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

