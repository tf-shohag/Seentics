-- Rollback retention policies to original values
-- This reverts the retention policies to their previous values

-- Remove the 1 year retention policies
SELECT remove_retention_policy('events', if_exists => TRUE);
SELECT remove_retention_policy('funnel_events', if_exists => TRUE);
SELECT remove_retention_policy('custom_events_aggregated', if_exists => TRUE);

-- Restore original retention policies
SELECT add_retention_policy('events', INTERVAL '2 years', if_not_exists => TRUE);
SELECT add_retention_policy('funnel_events', INTERVAL '2 years', if_not_exists => TRUE);
SELECT add_retention_policy('custom_events_aggregated', INTERVAL '1 year', if_not_exists => TRUE);
