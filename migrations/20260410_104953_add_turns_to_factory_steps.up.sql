-- Auto-generated migration
-- Review carefully before applying

SET lock_timeout = '10s';
SET statement_timeout = '60s';

-- Add column factory_steps.turns
ALTER TABLE factory_steps ADD COLUMN IF NOT EXISTS turns INTEGER NOT NULL DEFAULT 0;

