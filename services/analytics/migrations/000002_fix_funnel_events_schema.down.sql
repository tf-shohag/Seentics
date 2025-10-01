-- Rollback funnel_events schema changes

-- Drop the new indexes
DROP INDEX IF EXISTS idx_funnel_events_converted;
DROP INDEX IF EXISTS idx_funnel_events_started_at;
DROP INDEX IF EXISTS idx_funnel_events_last_activity;

-- Remove the new columns
ALTER TABLE funnel_events 
DROP COLUMN IF EXISTS started_at,
DROP COLUMN IF EXISTS last_activity,
DROP COLUMN IF EXISTS completed_steps,
DROP COLUMN IF EXISTS converted;
