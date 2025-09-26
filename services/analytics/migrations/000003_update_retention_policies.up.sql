-- Update retention policies to 1 year for all analytics data
-- This migration standardizes all retention policies to 1 year

-- Remove existing retention policies
SELECT remove_retention_policy('events', if_exists => TRUE);
SELECT remove_retention_policy('funnel_events', if_exists => TRUE);
SELECT remove_retention_policy('custom_events_aggregated', if_exists => TRUE);

-- Add new retention policies with 1 year retention for all tables
SELECT add_retention_policy('events', INTERVAL '1 year', if_not_exists => TRUE);
SELECT add_retention_policy('funnel_events', INTERVAL '1 year', if_not_exists => TRUE);
SELECT add_retention_policy('custom_events_aggregated', INTERVAL '1 year', if_not_exists => TRUE);
