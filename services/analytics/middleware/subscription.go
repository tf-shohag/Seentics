package middleware

import (
	"analytics-app/config"
	"analytics-app/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type SubscriptionMiddleware struct {
	usersServiceURL string
	globalAPIKey    string
	logger          zerolog.Logger
	usageTracker    *utils.UsageTracker
}

type UsageResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Plan  string `json:"plan"`
		Usage struct {
			Funnels struct {
				Current   int  `json:"current"`
				Limit     int  `json:"limit"`
				CanCreate bool `json:"canCreate"`
			} `json:"funnels"`
			MonthlyEvents struct {
				Current  int  `json:"current"`
				Limit    int  `json:"limit"`
				CanTrack bool `json:"canTrack"`
			} `json:"monthlyEvents"`
		} `json:"usage"`
	} `json:"data"`
}

func NewSubscriptionMiddleware(logger zerolog.Logger) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		usersServiceURL: os.Getenv("USER_SERVICE_URL"),
		globalAPIKey:    os.Getenv("GLOBAL_API_KEY"),
		logger:          logger,
		usageTracker:    utils.NewUsageTracker(logger),
	}
}

// CheckFunnelLimit middleware to check if user can create more funnels
func (s *SubscriptionMiddleware) CheckFunnelLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip usage limits in open source deployment
		if config.IsOpenSource() {
			c.Next()
			return
		}

		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			s.logger.Error().Msg("Missing user ID in request headers")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		if s.usersServiceURL == "" || s.globalAPIKey == "" {
			s.logger.Warn().Msg("Subscription checking disabled: missing configuration")
			c.Set("user_id", userID)
			c.Next()
			return
		}

		// Check current usage limits
		canCreate, err := s.checkUsageLimits(userID, "funnels")
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to check usage limits")
			// Fail open - allow creation if API is unavailable
			c.Set("user_id", userID)
			c.Next()
			return
		}

		if !canCreate {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Funnel limit reached",
				"message": "You have reached your plan's funnel limit. Please upgrade to create more funnels.",
				"code":    "LIMIT_REACHED",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func (s *SubscriptionMiddleware) checkUsageLimits(userID, resourceType string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/user/billing/usage", s.usersServiceURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create usage request: %w", err)
	}

	req.Header.Set("X-API-Key", s.globalAPIKey)
	req.Header.Set("X-User-ID", userID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send usage request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("usage check failed with status: %d", resp.StatusCode)
	}

	var usageResp UsageResponse
	if err := json.NewDecoder(resp.Body).Decode(&usageResp); err != nil {
		return false, fmt.Errorf("failed to decode usage response: %w", err)
	}

	if !usageResp.Success {
		return false, fmt.Errorf("usage check API returned error")
	}

	switch resourceType {
	case "funnels":
		return usageResp.Data.Usage.Funnels.CanCreate, nil
	case "monthlyEvents":
		return usageResp.Data.Usage.MonthlyEvents.CanTrack, nil
	default:
		return false, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

// CheckEventLimit middleware to check if user can track events (in-memory check)
func (s *SubscriptionMiddleware) CheckEventLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip usage limits in open source deployment
		if config.IsOpenSource() {
			c.Next()
			return
		}

		userID := c.GetHeader("x-user-id")
		if userID == "" {
			s.logger.Warn().Msg("No user ID found in headers for event limit check")
			c.Next() // Allow anonymous tracking
			return
		}

		// Fast in-memory check (no database call)
		canTrack, err := s.usageTracker.CheckEventLimit(userID, 1)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to check event limits")
			// Fail open - allow tracking if cache check fails
			c.Next()
			return
		}

		if !canTrack {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Monthly event limit exceeded",
				"message": "You have reached your plan's monthly event tracking limit. Please upgrade your plan to continue tracking events.",
				"code":    "EVENT_LIMIT_REACHED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckBatchEventLimit middleware for batch event endpoints (in-memory check)
func (s *SubscriptionMiddleware) CheckBatchEventLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip usage limits in open source deployment
		if config.IsOpenSource() {
			c.Next()
			return
		}

		userID := c.GetHeader("x-user-id")
		if userID == "" {
			s.logger.Warn().Msg("No user ID found in headers for batch event limit check")
			c.Next()
			return
		}

		// For batch events, we need to check if the user can track the number of events in the batch
		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			c.Abort()
			return
		}

		// Get events count from batch
		events, ok := requestBody["events"].([]interface{})
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid events array"})
			c.Abort()
			return
		}

		eventCount := len(events)
		
		// Fast in-memory check (no database call)
		canTrack, err := s.usageTracker.CheckEventLimit(userID, eventCount)
		if err != nil {
			s.logger.Error().Err(err).Str("user_id", userID).Int("event_count", eventCount).Msg("Failed to check batch event limits")
			// Re-bind the JSON for the next handler
			c.Set("request_body", requestBody)
			c.Next()
			return
		}

		if !canTrack {
			s.logger.Warn().Str("user_id", userID).Int("event_count", eventCount).Msg("User would exceed monthly event limits with batch")
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Monthly event limit would be exceeded",
				"message": fmt.Sprintf("Processing %d events would exceed your plan's monthly limit. Please upgrade your plan.", eventCount),
				"code":    "BATCH_EVENT_LIMIT_REACHED",
			})
			c.Abort()
			return
		}

		// Re-bind the JSON for the next handler
		c.Set("request_body", requestBody)
		c.Next()
	}
}


// IncrementEventUsage increments the event usage counter in memory (for events only)
func (s *SubscriptionMiddleware) IncrementEventUsage(userID string, eventCount int) error {
	if userID == "" {
		return nil // Skip for anonymous events
	}

	// Use in-memory tracking for events (will be flushed every 10 minutes)
	return s.usageTracker.TrackEvents(userID, eventCount)
}

// IncrementResourceUsage increments usage for non-event resources (real-time to database)
func (s *SubscriptionMiddleware) IncrementResourceUsage(userID, resourceType string, count int) error {
	if userID == "" {
		return nil
	}

	if s.usersServiceURL == "" || s.globalAPIKey == "" {
		return nil // Skip if not configured
	}

	// Prepare increment request
	incrementData := map[string]interface{}{
		"type":  resourceType,
		"count": count,
	}

	jsonData, err := json.Marshal(incrementData)
	if err != nil {
		return fmt.Errorf("failed to marshal increment data: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/user/billing/usage/increment", s.usersServiceURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create increment request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.globalAPIKey)
	req.Header.Set("x-user-id", userID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to increment usage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("increment usage returned status %d", resp.StatusCode)
	}

	s.logger.Debug().Str("user_id", userID).Str("resource_type", resourceType).Int("count", count).Msg("Successfully incremented resource usage")
	return nil
}

// GetUsageCacheStats returns statistics about the usage cache for monitoring
func (s *SubscriptionMiddleware) GetUsageCacheStats() map[string]interface{} {
	return s.usageTracker.GetCacheStats()
}

// Shutdown gracefully shuts down the usage tracker
func (s *SubscriptionMiddleware) Shutdown() {
	s.usageTracker.Shutdown()
}
