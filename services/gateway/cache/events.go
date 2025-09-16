package cache

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// EventsTracker handles Redis-based monthly events tracking
type EventsTracker struct {
	client *redis.Client
}

// EventsCounter represents user's monthly events data
type EventsCounter struct {
	UserID       string    `json:"user_id"`
	Count        int64     `json:"count"`
	LastUpdated  time.Time `json:"last_updated"`
	MonthYear    string    `json:"month_year"` // Format: "2024-01"
}

// UserEventsBatch represents batch data for sync
type UserEventsBatch struct {
	UserID    string `json:"user_id"`
	Count     int64  `json:"count"`
	MonthYear string `json:"month_year"`
}

// NewEventsTracker creates a new events tracker
func NewEventsTracker() *EventsTracker {
	return &EventsTracker{
		client: GetRedisClient(),
	}
}

// IncrementUserEvents increments the monthly events counter for a user
func (et *EventsTracker) IncrementUserEvents(ctx context.Context, userID string, count int64) error {
	if count <= 0 {
		count = 1
	}

	monthYear := time.Now().Format("2006-01")
	key := fmt.Sprintf("events:user:%s:%s", userID, monthYear)
	
	// Increment counter with expiration (35 days to handle month transitions)
	pipe := et.client.Pipeline()
	pipe.IncrBy(ctx, key, count)
	pipe.Expire(ctx, key, 35*24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to increment user events: %w", err)
	}

	// Update last activity timestamp
	timestampKey := fmt.Sprintf("events:user:%s:last_activity", userID)
	et.client.Set(ctx, timestampKey, time.Now().Unix(), 35*24*time.Hour)

	return nil
}

// GetUserEventsCount gets the current monthly events count for a user
func (et *EventsTracker) GetUserEventsCount(ctx context.Context, userID string) (int64, error) {
	monthYear := time.Now().Format("2006-01")
	key := fmt.Sprintf("events:user:%s:%s", userID, monthYear)
	
	result, err := et.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // No events tracked yet
		}
		return 0, fmt.Errorf("failed to get user events count: %w", err)
	}
	
	count, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse count: %w", err)
	}
	
	return count, nil
}

// CheckUserEventsLimit checks if user can track more events based on their plan
func (et *EventsTracker) CheckUserEventsLimit(ctx context.Context, userID, userPlan string, eventCount int64) (bool, int64, int64, error) {
	// Plan limits
	planLimits := map[string]int64{
		"free":     1000,
		"standard": 100000,
		"pro":      500000,
	}
	
	limit, exists := planLimits[userPlan]
	if !exists {
		limit = planLimits["free"] // Default to free plan
	}
	
	currentCount, err := et.GetUserEventsCount(ctx, userID)
	if err != nil {
		return false, 0, 0, err
	}
	
	canTrack := (currentCount + eventCount) <= limit
	return canTrack, currentCount, limit, nil
}

// GetAllActiveUsers gets all users who have tracked events this month
func (et *EventsTracker) GetAllActiveUsers(ctx context.Context) ([]string, error) {
	monthYear := time.Now().Format("2006-01")
	pattern := fmt.Sprintf("events:user:*:%s", monthYear)
	
	keys, err := et.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	
	userIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		// Extract user ID from key: events:user:USER_ID:MONTH_YEAR
		parts := splitKey(key)
		if len(parts) >= 4 {
			userIDs = append(userIDs, parts[2])
		}
	}
	
	return userIDs, nil
}

// GetUserEventsBatch gets events data for multiple users for batch sync
func (et *EventsTracker) GetUserEventsBatch(ctx context.Context, userIDs []string) ([]UserEventsBatch, error) {
	monthYear := time.Now().Format("2006-01")
	batch := make([]UserEventsBatch, 0, len(userIDs))
	
	pipe := et.client.Pipeline()
	
	// Prepare pipeline commands
	cmds := make(map[string]*redis.StringCmd)
	for _, userID := range userIDs {
		key := fmt.Sprintf("events:user:%s:%s", userID, monthYear)
		cmds[userID] = pipe.Get(ctx, key)
	}
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err.Error() != "redis: nil" {
		return nil, fmt.Errorf("failed to get user events batch: %w", err)
	}
	
	// Process results
	for userID, cmd := range cmds {
		result, err := cmd.Result()
		if err != nil {
			if err == redis.Nil {
				continue // No events for this user
			}
			log.Printf("Error getting count for user %s: %v", userID, err)
			continue
		}
		
		count, err := strconv.ParseInt(result, 10, 64)
		if err != nil {
			log.Printf("Error parsing count for user %s: %v", userID, err)
			continue
		}
		
		if count > 0 {
			batch = append(batch, UserEventsBatch{
				UserID:    userID,
				Count:     count,
				MonthYear: monthYear,
			})
		}
	}
	
	return batch, nil
}

// ResetUserEventsCounter resets the monthly counter for a user (called after successful sync)
func (et *EventsTracker) ResetUserEventsCounter(ctx context.Context, userID string) error {
	monthYear := time.Now().Format("2006-01")
	key := fmt.Sprintf("events:user:%s:%s", userID, monthYear)
	
	err := et.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to reset user events counter: %w", err)
	}
	
	return nil
}

// GetEventsStats gets overall events statistics
func (et *EventsTracker) GetEventsStats(ctx context.Context) (map[string]interface{}, error) {
	monthYear := time.Now().Format("2006-01")
	pattern := fmt.Sprintf("events:user:*:%s", monthYear)
	
	keys, err := et.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get events stats: %w", err)
	}
	
	totalUsers := len(keys)
	totalEvents := int64(0)
	
	if totalUsers > 0 {
		pipe := et.client.Pipeline()
		cmds := make([]*redis.StringCmd, len(keys))
		
		for i, key := range keys {
			cmds[i] = pipe.Get(ctx, key)
		}
		
		_, err := pipe.Exec(ctx)
		if err != nil && err.Error() != "redis: nil" {
			return nil, fmt.Errorf("failed to execute stats pipeline: %w", err)
		}
		
		for _, cmd := range cmds {
			result, err := cmd.Result()
			if err == nil {
				if count, parseErr := strconv.ParseInt(result, 10, 64); parseErr == nil {
					totalEvents += count
				}
			}
		}
	}
	
	return map[string]interface{}{
		"month_year":    monthYear,
		"total_users":   totalUsers,
		"total_events":  totalEvents,
		"average_events": func() float64 {
			if totalUsers > 0 {
				return float64(totalEvents) / float64(totalUsers)
			}
			return 0
		}(),
	}, nil
}

// Helper function to split Redis key
func splitKey(key string) []string {
	return strings.Split(key, ":")
}
