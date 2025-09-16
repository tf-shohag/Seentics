package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AdminStats represents the admin dashboard statistics
type AdminStats struct {
	TotalUsers      int                    `json:"total_users"`
	TotalWebsites   int                    `json:"total_websites"`
	TotalWorkflows  int                    `json:"total_workflows"`
	TotalFunnels    int                    `json:"total_funnels"`
	TotalEvents     int64                  `json:"total_events"`
	RecentUsers     []interface{}          `json:"recent_users"`
	RecentWebsites  []interface{}          `json:"recent_websites"`
	RecentWorkflows []interface{}          `json:"recent_workflows"`
	RecentFunnels   []interface{}          `json:"recent_funnels"`
	SystemHealth    map[string]interface{} `json:"system_health"`
	Timestamp       time.Time              `json:"timestamp"`
}

// GetAdminStats fetches comprehensive admin statistics
func GetAdminStats(w http.ResponseWriter, r *http.Request) {
	stats := &AdminStats{
		Timestamp: time.Now(),
		SystemHealth: map[string]interface{}{
			"status": "healthy",
		},
	}

	// Fetch user statistics
	userStats, err := fetchUserStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch user stats: %v", err), http.StatusInternalServerError)
		return
	}
	stats.TotalUsers = int(userStats["total"].(float64))
	stats.RecentUsers = userStats["recent"].([]interface{})

	// Fetch website statistics
	websiteStats, err := fetchWebsiteStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch website stats: %v", err), http.StatusInternalServerError)
		return
	}
	stats.TotalWebsites = int(websiteStats["total"].(float64))
	stats.RecentWebsites = websiteStats["recent"].([]interface{})

	// Fetch workflow statistics
	workflowStats, err := fetchWorkflowStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch workflow stats: %v", err), http.StatusInternalServerError)
		return
	}
	stats.TotalWorkflows = int(workflowStats["total"].(float64))
	stats.RecentWorkflows = workflowStats["recent"].([]interface{})

	// Fetch funnel statistics
	funnelStats, err := fetchFunnelStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch funnel stats: %v", err), http.StatusInternalServerError)
		return
	}
	stats.TotalFunnels = int(funnelStats["total"].(float64))
	stats.RecentFunnels = funnelStats["recent"].([]interface{})

	// Fetch analytics statistics
	analyticsStats, err := fetchAnalyticsStats()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch analytics stats: %v", err), http.StatusInternalServerError)
		return
	}
	stats.TotalEvents = int64(analyticsStats["total_events"].(float64))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetUsersList fetches paginated list of users
func GetUsersList(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")
	
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "20"
	}

	users, err := fetchUsersList(page, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch users list: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetWebsitesList fetches paginated list of websites
func GetWebsitesList(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")
	
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "20"
	}

	websites, err := fetchWebsitesList(page, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch websites list: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(websites)
}

// Helper functions to fetch data from services
func fetchUserStats() (map[string]interface{}, error) {
	return makeServiceRequest(os.Getenv("USER_SERVICE_URL") + "/api/v1/admin/stats")
}

func fetchWebsiteStats() (map[string]interface{}, error) {
	return makeServiceRequest(os.Getenv("USER_SERVICE_URL") + "/api/v1/admin/websites/stats")
}

func fetchWorkflowStats() (map[string]interface{}, error) {
	return makeServiceRequest(os.Getenv("WORKFLOW_SERVICE_URL") + "/api/v1/admin/stats")
}

func fetchFunnelStats() (map[string]interface{}, error) {
	return makeServiceRequest(os.Getenv("ANALYTICS_SERVICE_URL") + "/api/v1/admin/funnels/stats")
}

func fetchAnalyticsStats() (map[string]interface{}, error) {
	return makeServiceRequest(os.Getenv("ANALYTICS_SERVICE_URL") + "/api/v1/admin/analytics/stats")
}

func fetchUsersList(page, limit string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/admin/users?page=%s&limit=%s", os.Getenv("USER_SERVICE_URL"), page, limit)
	return makeServiceRequest(url)
}

func fetchWebsitesList(page, limit string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/admin/websites?page=%s&limit=%s", os.Getenv("USER_SERVICE_URL"), page, limit)
	return makeServiceRequest(url)
}

// makeServiceRequest makes HTTP request to internal services
func makeServiceRequest(url string) (map[string]interface{}, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add API key for inter-service communication
	globalAPIKey := os.Getenv("GLOBAL_API_KEY")
	if globalAPIKey != "" {
		req.Header.Set("X-API-Key", globalAPIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
