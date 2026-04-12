-- Restore cached counter columns to factory_prds.

SET lock_timeout = '10s';
SET statement_timeout = '60s';

ALTER TABLE factory_prds ADD COLUMN IF NOT EXISTS completed_tasks INTEGER NOT NULL DEFAULT 0;
ALTER TABLE factory_prds ADD COLUMN IF NOT EXISTS total_tasks INTEGER NOT NULL DEFAULT 0;
ALTER TABLE factory_prds ADD COLUMN IF NOT EXISTS total_cost_usd DOUBLE PRECISION NOT NULL DEFAULT 0;
