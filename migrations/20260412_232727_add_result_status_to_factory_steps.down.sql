-- Auto-generated rollback migration

-- Rollback: Drop column factory_prds.total_cost_usd
ALTER TABLE factory_prds ADD COLUMN total_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0;

-- Rollback: Drop column factory_prds.completed_tasks
ALTER TABLE factory_prds ADD COLUMN completed_tasks INTEGER NOT NULL DEFAULT 0;

-- Rollback: Drop column factory_prds.total_tasks
ALTER TABLE factory_prds ADD COLUMN total_tasks INTEGER NOT NULL DEFAULT 0;

-- Rollback: Add column factory_steps.result_status
ALTER TABLE factory_steps DROP COLUMN result_status;

