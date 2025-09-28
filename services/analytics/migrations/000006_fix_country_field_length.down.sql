-- Rollback country field length fix
-- WARNING: This may cause data loss if country names are longer than 2 characters

-- Remove the index we added
DROP INDEX IF EXISTS idx_events_country;

-- Revert country field back to original size (this may fail if data exists)
ALTER TABLE events ALTER COLUMN country TYPE VARCHAR(2);

-- Revert other geo fields back to original sizes
ALTER TABLE events ALTER COLUMN city TYPE VARCHAR(100);  -- This was already 100, no change needed
ALTER TABLE events ALTER COLUMN region TYPE VARCHAR(100);  -- This field didn't exist in original schema
ALTER TABLE events ALTER COLUMN continent TYPE VARCHAR(100);  -- This field didn't exist in original schema
