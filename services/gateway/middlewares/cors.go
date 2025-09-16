package middlewares

import (
	"net/http"
	"os"
	"strings"
)

// CORS middleware - handles all CORS logic
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Get CORS configuration from environment
		corsOrigins := getCORSOrigins()

		// Check if origin is allowed
		if isOriginAllowed(origin, corsOrigins) {
			// Set CORS headers (these will be the ONLY CORS headers set)
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization,X-API-Key,X-Site-ID,X-Domain")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length,Content-Range")
			w.Header().Set("Access-Control-Max-Age", "1728000")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getCORSOrigins returns allowed CORS origins from environment
func getCORSOrigins() []string {
	corsConfig := os.Getenv("CORS_ORIGIN")
	if corsConfig == "" {
		// Clean up your defaults - remove duplicates and trailing slashes
		return []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"https://seentics.com",
			"https://www.seentics.com",
			"https://api.seentics.com",
		}
	}

	// Split and clean up origins
	origins := strings.Split(corsConfig, ",")
	var cleanOrigins []string
	for _, origin := range origins {
		cleaned := strings.TrimSpace(strings.TrimSuffix(origin, "/"))
		if cleaned != "" {
			cleanOrigins = append(cleanOrigins, cleaned)
		}
	}
	return cleanOrigins
}

// isOriginAllowed checks if the request origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	// Normalize origin by removing trailing slash
	normalizedOrigin := strings.TrimSuffix(origin, "/")

	for _, allowed := range allowedOrigins {
		if allowed == normalizedOrigin {
			return true
		}
	}
	return false
}
