-- 001_initial.down.sql
-- Drop all Factory tables in reverse FK order

DROP TABLE IF EXISTS factory_lint_results;
DROP TABLE IF EXISTS factory_insights;
DROP TABLE IF EXISTS factory_suggestions;
DROP TABLE IF EXISTS factory_questions;
DROP TABLE IF EXISTS factory_reviews;
DROP TABLE IF EXISTS factory_steps;
DROP TABLE IF EXISTS factory_runs;
DROP TABLE IF EXISTS factory_tasks;
DROP TABLE IF EXISTS factory_prds;
DROP TABLE IF EXISTS factory_projects;
