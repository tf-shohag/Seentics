-- Runs on first initialization of the database volume
-- Ensure TimescaleDB extension exists for the primary database
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Example: you can add schema migrations here if desired
-- CREATE SCHEMA IF NOT EXISTS analytics;
