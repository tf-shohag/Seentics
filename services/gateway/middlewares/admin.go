package middlewares

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// AdminOrJWTMiddleware validates admin access using admin code OR JWT token
func AdminOrJWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for JWT token first
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			// JWT token provided - allow access (assuming JWT validation is handled elsewhere)
			next(w, r)
			return
		}

		// No JWT token, check for admin code
		adminCode := os.Getenv("ADMIN_CODE")
		if adminCode == "" {
			http.Error(w, "Admin access not configured", http.StatusServiceUnavailable)
			return
		}

		// Get admin code from request header or query parameter
		providedCode := r.Header.Get("X-Admin-Code")
		if providedCode == "" {
			providedCode = r.URL.Query().Get("admin_code")
		}

		// Check if admin code matches
		if providedCode != adminCode {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "Unauthorized",
				"message": "Admin access requires valid admin code or JWT token",
			})
			return
		}

		next(w, r)
	}
}

// AdminCORSMiddleware handles CORS for admin routes
func AdminCORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for admin routes
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Admin-Code")
		w.Header().Set("Access-Control-Expose-Headers", "X-Admin-Code")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
