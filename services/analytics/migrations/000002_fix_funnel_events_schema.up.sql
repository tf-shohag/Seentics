-- Fix funnel_events table schema to match the Go model
-- Add missing columns and rename completed to converted

-- Add missing columns
ALTER TABLE funnel_events 
ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS last_activity TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS completed_steps INTEGER[],
ADD COLUMN IF NOT EXISTS converted BOOLEAN DEFAULT FALSE;

-- Copy data from completed to converted column
UPDATE funnel_events SET converted = completed WHERE completed IS NOT NULL;

-- Drop the old completed column (optional - can be kept for backward compatibility)
-- ALTER TABLE funnel_events DROP COLUMN IF EXISTS completed;

-- Update indexes for better performance
CREATE INDEX IF NOT EXISTS idx_funnel_events_converted ON funnel_events(converted) WHERE converted = true;
CREATE INDEX IF NOT EXISTS idx_funnel_events_started_at ON funnel_events(started_at) WHERE started_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_funnel_events_last_activity ON funnel_events(last_activity) WHERE last_activity IS NOT NULL;
