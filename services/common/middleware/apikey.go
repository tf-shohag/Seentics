package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware(expectedAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip API key validation for health check
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Skip if no API key is configured
		if expectedAPIKey == "" {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		if apiKey != expectedAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
