-- Add comprehensive geolocation fields to events table
ALTER TABLE events ADD COLUMN IF NOT EXISTS country_code VARCHAR(2);
ALTER TABLE events ADD COLUMN IF NOT EXISTS continent VARCHAR(50);
ALTER TABLE events ADD COLUMN IF NOT EXISTS region VARCHAR(100);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_events_country_code ON events(country_code);
CREATE INDEX IF NOT EXISTS idx_events_continent ON events(continent);
CREATE INDEX IF NOT EXISTS idx_events_region ON events(region);

-- Update existing NULL values to 'Unknown' for consistency
UPDATE events SET country_code = 'XX' WHERE country_code IS NULL AND country IS NOT NULL;
UPDATE events SET continent = 'Unknown' WHERE continent IS NULL;
UPDATE events SET region = 'Unknown' WHERE region IS NULL;
