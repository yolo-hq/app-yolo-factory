-- factory_projects
CREATE INDEX idx_factory_projects_status ON factory_projects(status);
CREATE INDEX idx_factory_projects_name ON factory_projects(name);

-- factory_prds
CREATE INDEX idx_factory_prds_project_id ON factory_prds(project_id);
CREATE INDEX idx_factory_prds_status ON factory_prds(status);

-- factory_tasks
CREATE INDEX idx_factory_tasks_prd_id ON factory_tasks(prd_id);
CREATE INDEX idx_factory_tasks_project_id ON factory_tasks(project_id);
CREATE INDEX idx_factory_tasks_status ON factory_tasks(status);
CREATE INDEX idx_factory_tasks_sequence ON factory_tasks(sequence);

-- factory_runs
CREATE INDEX idx_factory_runs_task_id ON factory_runs(task_id);
CREATE INDEX idx_factory_runs_status ON factory_runs(status);

-- factory_steps
CREATE INDEX idx_factory_steps_run_id ON factory_steps(run_id);
CREATE INDEX idx_factory_steps_status ON factory_steps(status);

-- factory_reviews
CREATE INDEX idx_factory_reviews_run_id ON factory_reviews(run_id);
CREATE INDEX idx_factory_reviews_task_id ON factory_reviews(task_id);
CREATE INDEX idx_factory_reviews_verdict ON factory_reviews(verdict);

-- factory_questions
CREATE INDEX idx_factory_questions_task_id ON factory_questions(task_id);
CREATE INDEX idx_factory_questions_run_id ON factory_questions(run_id);
CREATE INDEX idx_factory_questions_status ON factory_questions(status);

-- factory_suggestions
CREATE INDEX idx_factory_suggestions_project_id ON factory_suggestions(project_id);
CREATE INDEX idx_factory_suggestions_status ON factory_suggestions(status);
CREATE INDEX idx_factory_suggestions_category ON factory_suggestions(category);
