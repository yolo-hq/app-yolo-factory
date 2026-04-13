-- Auto-generated migration
-- Review carefully before applying

SET lock_timeout = '10s';
SET statement_timeout = '60s';

-- Add column factory_steps.result_status
ALTER TABLE factory_steps ADD COLUMN IF NOT EXISTS result_status TEXT NOT NULL;

-- Drop column factory_prds.total_tasks
-- DESTRUCTIVE: This will permanently delete the column and its data
ALTER TABLE factory_prds DROP COLUMN IF EXISTS total_tasks;

-- Drop column factory_prds.completed_tasks
-- DESTRUCTIVE: This will permanently delete the column and its data
ALTER TABLE factory_prds DROP COLUMN IF EXISTS completed_tasks;

-- Drop column factory_prds.total_cost_usd
-- DESTRUCTIVE: This will permanently delete the column and its data
ALTER TABLE factory_prds DROP COLUMN IF EXISTS total_cost_usd;

