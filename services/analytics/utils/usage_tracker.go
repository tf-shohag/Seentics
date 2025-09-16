package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type UsageTracker struct {
	usersServiceURL string
	globalAPIKey    string
	logger          zerolog.Logger
}

type UsageRequest struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

func NewUsageTracker(logger zerolog.Logger) *UsageTracker {
	return &UsageTracker{
		usersServiceURL: os.Getenv("USER_SERVICE_URL"),
		globalAPIKey:    os.Getenv("GLOBAL_API_KEY"),
		logger:          logger,
	}
}

// IncrementUsage increments the usage counter for a specific resource type
func (u *UsageTracker) IncrementUsage(userID, resourceType string, count int) error {
	if u.usersServiceURL == "" || u.globalAPIKey == "" {
		u.logger.Warn().Msg("Usage tracking disabled: missing configuration")
		return nil
	}

	reqBody := UsageRequest{
		Type:  resourceType,
		Count: count,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal usage request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/user/billing/usage/increment", u.usersServiceURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create usage request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", u.globalAPIKey)
	req.Header.Set("X-User-ID", userID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send usage request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("usage tracking failed with status: %d", resp.StatusCode)
	}

	u.logger.Info().
		Str("user_id", userID).
		Str("type", resourceType).
		Int("count", count).
		Msg("Usage tracking successful")

	return nil
}

// DecrementUsage decrements the usage counter for a specific resource type
func (u *UsageTracker) DecrementUsage(userID, resourceType string, count int) error {
	return u.IncrementUsage(userID, resourceType, -count)
}
