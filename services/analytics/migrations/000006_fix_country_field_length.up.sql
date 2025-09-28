-- Fix country field length issue
-- The country field is too short (VARCHAR(2)) and causing insert failures

-- Increase country field length to accommodate full country names
ALTER TABLE events ALTER COLUMN country TYPE VARCHAR(100);

-- Also ensure other geo fields have adequate length
ALTER TABLE events ALTER COLUMN city TYPE VARCHAR(100);
ALTER TABLE events ALTER COLUMN region TYPE VARCHAR(100);
ALTER TABLE events ALTER COLUMN continent TYPE VARCHAR(100);

-- Add index for better performance on the now-larger country field
CREATE INDEX IF NOT EXISTS idx_events_country ON events(country) WHERE country IS NOT NULL;
