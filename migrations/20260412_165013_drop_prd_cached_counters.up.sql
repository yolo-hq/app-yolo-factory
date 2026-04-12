-- Drop cached counter columns from factory_prds.
-- completed_tasks, total_tasks, total_cost_usd are now virtual computed fields
-- derived from factory_tasks via GROUP BY aggregation.

SET lock_timeout = '10s';
SET statement_timeout = '60s';

ALTER TABLE factory_prds DROP COLUMN IF EXISTS completed_tasks;
ALTER TABLE factory_prds DROP COLUMN IF EXISTS total_tasks;
ALTER TABLE factory_prds DROP COLUMN IF EXISTS total_cost_usd;
