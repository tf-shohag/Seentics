-- Remove geolocation fields from events table
DROP INDEX IF EXISTS idx_events_region;
DROP INDEX IF EXISTS idx_events_continent;
DROP INDEX IF EXISTS idx_events_country_code;

ALTER TABLE events 
DROP COLUMN IF EXISTS region,
DROP COLUMN IF EXISTS continent,
DROP COLUMN IF EXISTS country_code;
