package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/config"
)

// Plan limits configuration
var PLAN_LIMITS = map[string]map[string]interface{}{
	"free": {
		"websites":      1,
		"monthlyEvents": 1000,
		"workflows":     1,
		"funnels":       1,
		"features":      []string{"basic_analytics", "email_support", "privacy_compliant"},
	},
	"standard": {
		"websites":      5,
		"monthlyEvents": 100000,
		"workflows":     10,
		"funnels":       10,
		"features":      []string{"advanced_analytics", "priority_support", "custom_domains", "api_access", "advanced_integrations", "ab_testing"},
	},
	"pro": {
		"websites":      10,
		"monthlyEvents": 500000,
		"workflows":     30,
		"funnels":       30,
		"features":      []string{"enterprise_analytics", "24_7_support", "white_label", "advanced_api", "custom_integrations", "advanced_ab_testing", "team_collaboration", "custom_reports"},
	},
}

// CloudFeaturesMiddleware checks if cloud features are enabled
func CloudFeaturesMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()
		c.Set("cloudFeaturesEnabled", cfg.CloudFeaturesEnabled)
		c.Next()
	}
}

// CheckSubscriptionLimit checks if user can perform action based on their plan
func CheckSubscriptionLimit(limitType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()
		
		// Skip usage limits in open source deployment
		if !cfg.CloudFeaturesEnabled {
			c.Next()
			return
		}

		userID := c.GetHeader("X-User-ID")
		userPlan := c.GetHeader("X-User-Plan")
		userStatus := c.GetHeader("X-User-Status")

		if userPlan == "" {
			userPlan = "free"
		}

		if userStatus == "" {
			userStatus = "active"
		}

		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User authentication required",
			})
			c.Abort()
			return
		}

		// Check if user is active
		if userStatus != "active" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User account is not active",
			})
			c.Abort()
			return
		}

		// Get plan limits
		planLimits, exists := PLAN_LIMITS[userPlan]
		if !exists {
			planLimits = PLAN_LIMITS["free"]
		}

		// For now, we'll implement basic limit checking
		// In a full implementation, you'd check current usage from database
		limit, hasLimit := planLimits[limitType]
		if hasLimit {
			// Store limit info for later use
			c.Set("userPlan", userPlan)
			c.Set("planLimit", limit)
		}

		c.Next()
	}
}

// RequireCloudFeatures middleware that blocks access if cloud features are disabled
func RequireCloudFeatures() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()
		
		if !cfg.CloudFeaturesEnabled {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "This feature is only available in the cloud version",
				"code":  "CLOUD_FEATURE_REQUIRED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckFeatureAccess checks if user's plan has access to a specific feature
func CheckFeatureAccess(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Load()
		
		// Skip feature checks in open source deployment
		if !cfg.CloudFeaturesEnabled {
			c.Next()
			return
		}

		userPlan := c.GetHeader("X-User-Plan")
		if userPlan == "" {
			userPlan = "free"
		}

		planLimits, exists := PLAN_LIMITS[userPlan]
		if !exists {
			planLimits = PLAN_LIMITS["free"]
		}

		features, hasFeatures := planLimits["features"].([]string)
		if !hasFeatures {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Feature access information not available",
			})
			c.Abort()
			return
		}

		// Check if feature is available in plan
		hasAccess := false
		for _, f := range features {
			if f == feature {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Feature not available in your current plan",
				"feature": feature,
				"plan":    userPlan,
				"code":    "FEATURE_NOT_AVAILABLE",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
