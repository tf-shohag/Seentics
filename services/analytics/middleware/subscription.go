package middleware

import (
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
		} `json:"usage"`
	} `json:"data"`
}

func NewSubscriptionMiddleware(logger zerolog.Logger) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		usersServiceURL: os.Getenv("USER_SERVICE_URL"),
		globalAPIKey:    os.Getenv("GLOBAL_API_KEY"),
		logger:          logger,
	}
}

// CheckFunnelLimit middleware to check if user can create more funnels
func (s *SubscriptionMiddleware) CheckFunnelLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
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
	default:
		return false, fmt.Errorf("unknown resource type: %s", resourceType)
	}
}
