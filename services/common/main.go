package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/seentics/seentics/services/common/config"
	"github.com/seentics/seentics/services/common/database"
	"github.com/seentics/seentics/services/common/middleware"
	"github.com/seentics/seentics/services/common/routes"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Set log level
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Initialize database connections
	if err := database.InitMongoDB(cfg.MongoURI); err != nil {
		logrus.Fatal("Failed to connect to MongoDB: ", err)
	}

	if err := database.InitRedis(cfg.RedisURL); err != nil {
		logrus.Fatal("Failed to connect to Redis: ", err)
	}

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	router.Use(cors.New(corsConfig))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "OK",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "common-service",
		})
	})

	// API Key middleware for all routes except health
	router.Use(middleware.APIKeyMiddleware(cfg.APIKey))

	// Setup routes
	api := router.Group("/api/v1")
	{
		// User management routes
		userRoutes := api.Group("/user")
		routes.SetupUserRoutes(userRoutes)

		// Workflow management routes
		workflowRoutes := api.Group("/workflows")
		routes.SetupWorkflowRoutes(workflowRoutes)

		// Admin routes
		adminRoutes := api.Group("/admin")
		routes.SetupAdminRoutes(adminRoutes)

		// Internal routes for microservice communication
		internalRoutes := api.Group("/internal")
		routes.SetupInternalRoutes(internalRoutes)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Common Service starting on port %s", cfg.Port)
		logrus.Infof("Environment: %s", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatal("Failed to start server: ", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatal("Server forced to shutdown: ", err)
	}

	// Close database connections
	database.CloseMongoDB()
	database.CloseRedis()

	logrus.Info("Server exited")
}
