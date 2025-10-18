package config

import (
	"os"
)

type Config struct {
	Port        string
	Environment string
	MongoURI    string
	RedisURL    string
	APIKey      string
	JWTSecret   string
	
	// Email configuration
	EmailAPIKey string
	EmailFrom   string
	
	// OAuth configuration
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	FrontendURL        string
	
	// Cloud features
	CloudFeaturesEnabled bool
	
	// Lemon Squeezy configuration
	LemonSqueezyAPIKey            string
	LemonSqueezyWebhookSecret     string
	LemonSqueezyStandardVariantID string
	LemonSqueezyProVariantID      string
	
	// Webhook configuration
	WebhookSecret string
	
	// External service URLs
	AnalyticsServiceURL string
	
	// Database names
	UserDBName     string
	WorkflowDBName string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "3004"),
		Environment: getEnv("NODE_ENV", "development"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		APIKey:      getEnv("API_KEY", ""),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		
		EmailAPIKey: getEnv("EMAIL_API_KEY", ""),
		EmailFrom:   getEnv("EMAIL_FROM", "noreply@seentics.com"),
		
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		
		CloudFeaturesEnabled: getEnv("CLOUD_FEATURES_ENABLED", "false") == "true",
		
		LemonSqueezyAPIKey:            getEnv("LEMONSQUEEZY_API_KEY", ""),
		LemonSqueezyWebhookSecret:     getEnv("LEMONSQUEEZY_WEBHOOK_SECRET", ""),
		LemonSqueezyStandardVariantID: getEnv("LEMONSQUEEZY_STANDARD_VARIANT_ID", ""),
		LemonSqueezyProVariantID:      getEnv("LEMONSQUEEZY_PRO_VARIANT_ID", ""),
		
		WebhookSecret: getEnv("WEBHOOK_SECRET", ""),
		
		AnalyticsServiceURL: getEnv("ANALYTICS_SERVICE_URL", "http://localhost:3002"),
		
		UserDBName:     getEnv("USER_DB_NAME", "seentics_users"),
		WorkflowDBName: getEnv("WORKFLOW_DB_NAME", "seentics_workflows"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
