package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Debugging script to identify and fix analytics data issues
// Run this to diagnose why analytics data is empty

type DebugResult struct {
	Query       string      `json:"query"`
	Description string      `json:"description"`
	Result      interface{} `json:"result"`
	Error       string      `json:"error,omitempty"`
}

func main() {
	// Database connection string - update with your actual credentials
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://username:password@localhost:5432/analytics_db?sslmode=disable"
		fmt.Println("Warning: Using default database URL. Set DATABASE_URL environment variable.")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	websiteID := "68d8d06d2b9f57ed64fbbb6a"
	results := []DebugResult{}

	// 1. Check total events in database
	result1 := checkTotalEvents(ctx, db)
	results = append(results, result1)

	// 2. Check events for specific website
	result2 := checkWebsiteEvents(ctx, db, websiteID)
	results = append(results, result2)

	// 3. Check recent events (last 24 hours)
	result3 := checkRecentEvents(ctx, db, websiteID)
	results = append(results, result3)

	// 4. Check custom events aggregated
	result4 := checkCustomEvents(ctx, db, websiteID)
	results = append(results, result4)

	// 5. Run the exact dashboard query
	result5 := checkDashboardQuery(ctx, db, websiteID)
	results = append(results, result5)

	// 6. Check event processing logs
	result6 := checkEventTypes(ctx, db, websiteID)
	results = append(results, result6)

	// Output results as JSON
	output, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(output))

	// Write to file for analysis
	err = os.WriteFile("analytics_debug_results.json", output, 0644)
	if err != nil {
		log.Printf("Failed to write results to file: %v", err)
	} else {
		fmt.Println("\nResults written to analytics_debug_results.json")
	}
}

func checkTotalEvents(ctx context.Context, db *pgxpool.Pool) DebugResult {
	query := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(DISTINCT website_id) as unique_websites,
			COUNT(DISTINCT visitor_id) as unique_visitors,
			COUNT(DISTINCT session_id) as unique_sessions,
			MIN(timestamp) as earliest_event,
			MAX(timestamp) as latest_event
		FROM events`

	var result struct {
		TotalEvents    int       `json:"total_events"`
		UniqueWebsites int       `json:"unique_websites"`
		UniqueVisitors int       `json:"unique_visitors"`
		UniqueSessions int       `json:"unique_sessions"`
		EarliestEvent  time.Time `json:"earliest_event"`
		LatestEvent    time.Time `json:"latest_event"`
	}

	err := db.QueryRow(ctx, query).Scan(
		&result.TotalEvents, &result.UniqueWebsites, &result.UniqueVisitors,
		&result.UniqueSessions, &result.EarliestEvent, &result.LatestEvent,
	)

	debugResult := DebugResult{
		Query:       query,
		Description: "Check total events in database",
		Result:      result,
	}

	if err != nil {
		debugResult.Error = err.Error()
	}

	return debugResult
}

func checkWebsiteEvents(ctx context.Context, db *pgxpool.Pool, websiteID string) DebugResult {
	query := `
		SELECT 
			event_type,
			COUNT(*) as total_events,
			COUNT(DISTINCT visitor_id) as unique_visitors,
			COUNT(DISTINCT session_id) as unique_sessions,
			MIN(timestamp) as first_event,
			MAX(timestamp) as last_event
		FROM events 
		WHERE website_id = $1
		GROUP BY event_type
		ORDER BY total_events DESC`

	rows, err := db.Query(ctx, query, websiteID)
	if err != nil {
		return DebugResult{
			Query:       query,
			Description: "Check events for specific website",
			Error:       err.Error(),
		}
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var eventType string
		var totalEvents, uniqueVisitors, uniqueSessions int
		var firstEvent, lastEvent time.Time

		err := rows.Scan(&eventType, &totalEvents, &uniqueVisitors, &uniqueSessions, &firstEvent, &lastEvent)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"event_type":      eventType,
			"total_events":    totalEvents,
			"unique_visitors": uniqueVisitors,
			"unique_sessions": uniqueSessions,
			"first_event":     firstEvent,
			"last_event":      lastEvent,
		})
	}

	return DebugResult{
		Query:       query,
		Description: fmt.Sprintf("Check events for website %s", websiteID),
		Result:      results,
	}
}

func checkRecentEvents(ctx context.Context, db *pgxpool.Pool, websiteID string) DebugResult {
	query := `
		SELECT 
			id,
			visitor_id,
			session_id,
			event_type,
			page,
			timestamp,
			created_at
		FROM events 
		WHERE website_id = $1
		AND timestamp >= NOW() - INTERVAL '24 hours'
		ORDER BY timestamp DESC
		LIMIT 10`

	rows, err := db.Query(ctx, query, websiteID)
	if err != nil {
		return DebugResult{
			Query:       query,
			Description: "Check recent events (last 24 hours)",
			Error:       err.Error(),
		}
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, visitorID, sessionID, eventType, page string
		var timestamp, createdAt time.Time

		err := rows.Scan(&id, &visitorID, &sessionID, &eventType, &page, &timestamp, &createdAt)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"id":          id,
			"visitor_id":  visitorID,
			"session_id":  sessionID,
			"event_type":  eventType,
			"page":        page,
			"timestamp":   timestamp,
			"created_at":  createdAt,
		})
	}

	return DebugResult{
		Query:       query,
		Description: "Recent events in last 24 hours",
		Result:      results,
	}
}

func checkCustomEvents(ctx context.Context, db *pgxpool.Pool, websiteID string) DebugResult {
	query := `
		SELECT 
			event_type,
			COUNT(*) as total_records,
			SUM(count) as total_count,
			MIN(first_seen) as first_seen,
			MAX(last_seen) as last_seen
		FROM custom_events_aggregated 
		WHERE website_id = $1
		GROUP BY event_type
		ORDER BY total_count DESC`

	rows, err := db.Query(ctx, query, websiteID)
	if err != nil {
		return DebugResult{
			Query:       query,
			Description: "Check custom events aggregated",
			Error:       err.Error(),
		}
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var eventType string
		var totalRecords, totalCount int
		var firstSeen, lastSeen time.Time

		err := rows.Scan(&eventType, &totalRecords, &totalCount, &firstSeen, &lastSeen)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"event_type":     eventType,
			"total_records":  totalRecords,
			"total_count":    totalCount,
			"first_seen":     firstSeen,
			"last_seen":      lastSeen,
		})
	}

	return DebugResult{
		Query:       query,
		Description: "Custom events aggregated data",
		Result:      results,
	}
}

func checkDashboardQuery(ctx context.Context, db *pgxpool.Pool, websiteID string) DebugResult {
	query := `
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
			WHERE website_id = $1 
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
		WHERE e.website_id = $1
		AND e.timestamp >= NOW() - INTERVAL '7 days'
		AND e.event_type = 'pageview'`

	var result struct {
		PageViews       int     `json:"page_views"`
		TotalVisitors   int     `json:"total_visitors"`
		UniqueVisitors  int     `json:"unique_visitors"`
		Sessions        int     `json:"sessions"`
		BounceRate      float64 `json:"bounce_rate"`
		AvgSessionTime  float64 `json:"avg_session_time"`
		PagesPerSession float64 `json:"pages_per_session"`
	}

	err := db.QueryRow(ctx, query, websiteID).Scan(
		&result.PageViews, &result.TotalVisitors, &result.UniqueVisitors,
		&result.Sessions, &result.BounceRate, &result.AvgSessionTime, &result.PagesPerSession,
	)

	debugResult := DebugResult{
		Query:       query,
		Description: "Exact dashboard query that should return analytics data",
		Result:      result,
	}

	if err != nil {
		debugResult.Error = err.Error()
	}

	return debugResult
}

func checkEventTypes(ctx context.Context, db *pgxpool.Pool, websiteID string) DebugResult {
	query := `
		SELECT 
			event_type,
			COUNT(*) as count,
			'events_table' as source
		FROM events 
		WHERE website_id = $1
		GROUP BY event_type
		UNION ALL
		SELECT 
			event_type,
			SUM(count) as count,
			'aggregated_table' as source
		FROM custom_events_aggregated 
		WHERE website_id = $1
		GROUP BY event_type
		ORDER BY count DESC`

	rows, err := db.Query(ctx, query, websiteID)
	if err != nil {
		return DebugResult{
			Query:       query,
			Description: "Check event types distribution",
			Error:       err.Error(),
		}
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var eventType, source string
		var count int

		err := rows.Scan(&eventType, &count, &source)
		if err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"event_type": eventType,
			"count":      count,
			"source":     source,
		})
	}

	return DebugResult{
		Query:       query,
		Description: "Event types distribution across tables",
		Result:      results,
	}
}
