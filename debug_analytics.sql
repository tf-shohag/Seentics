-- Debug script to check analytics data
-- Run these queries to diagnose the issue

-- 1. Check if any events exist in the events table
SELECT 
    COUNT(*) as total_events,
    COUNT(DISTINCT website_id) as unique_websites,
    COUNT(DISTINCT visitor_id) as unique_visitors,
    COUNT(DISTINCT session_id) as unique_sessions
FROM events;

-- 2. Check events for your specific website
SELECT 
    COUNT(*) as total_events,
    COUNT(DISTINCT visitor_id) as unique_visitors,
    COUNT(DISTINCT session_id) as unique_sessions,
    event_type,
    MIN(timestamp) as first_event,
    MAX(timestamp) as last_event
FROM events 
WHERE website_id = '68d8d06d2b9f57ed64fbbb6a'
GROUP BY event_type
ORDER BY total_events DESC;

-- 3. Check recent events for your website (last 24 hours)
SELECT 
    id,
    visitor_id,
    session_id,
    event_type,
    page,
    timestamp,
    created_at
FROM events 
WHERE website_id = '68d8d06d2b9f57ed64fbbb6a'
AND timestamp >= NOW() - INTERVAL '24 hours'
ORDER BY timestamp DESC
LIMIT 10;

-- 4. Check custom events aggregated table
SELECT 
    COUNT(*) as total_custom_events,
    event_type,
    SUM(count) as total_count,
    MIN(first_seen) as first_seen,
    MAX(last_seen) as last_seen
FROM custom_events_aggregated 
WHERE website_id = '68d8d06d2b9f57ed64fbbb6a'
GROUP BY event_type
ORDER BY total_count DESC;

-- 5. Check if there are any events at all in the database
SELECT 
    'events' as table_name,
    COUNT(*) as row_count,
    MIN(timestamp) as earliest,
    MAX(timestamp) as latest
FROM events
UNION ALL
SELECT 
    'custom_events_aggregated' as table_name,
    COUNT(*) as row_count,
    MIN(first_seen) as earliest,
    MAX(last_seen) as latest
FROM custom_events_aggregated;

-- 6. Check the exact dashboard query that's failing
WITH session_stats AS (
    SELECT 
        session_id,
        COUNT(*) as page_count,
        CASE 
            WHEN COUNT(*) > 1 THEN 
                LEAST(EXTRACT(EPOCH FROM (MAX(timestamp) - MIN(timestamp))), 1800)
            ELSE 
                COALESCE(MAX(time_on_page), 30)
        END as session_duration
    FROM events
    WHERE website_id = '68d8d06d2b9f57ed64fbbb6a'
    AND timestamp >= NOW() - INTERVAL '7 days'
    AND event_type = 'pageview'
    GROUP BY session_id
)
SELECT 
    COUNT(*) as page_views,
    COUNT(DISTINCT e.session_id) as total_visitors,
    COUNT(DISTINCT e.visitor_id) as unique_visitors,
    COUNT(DISTINCT e.session_id) as sessions,
    COALESCE(
        (COUNT(DISTINCT CASE WHEN s.page_count = 1 THEN e.session_id END) * 100.0) / 
        NULLIF(COUNT(DISTINCT e.session_id), 0), 0
    ) as bounce_rate,
    COALESCE(AVG(s.session_duration), 0) as avg_session_time,
    COALESCE(COUNT(*) * 1.0 / NULLIF(COUNT(DISTINCT e.session_id), 0), 0) as pages_per_session
FROM events e
INNER JOIN session_stats s ON e.session_id = s.session_id
WHERE e.website_id = '68d8d06d2b9f57ed64fbbb6a'
AND e.timestamp >= NOW() - INTERVAL '7 days'
AND e.event_type = 'pageview';
