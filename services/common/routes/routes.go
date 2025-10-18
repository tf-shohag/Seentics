package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/middleware"
	"github.com/seentics/seentics/services/common/services"
)

func SetupUserRoutes(router *gin.RouterGroup) {
	// Authentication routes (no auth required)
	auth := router.Group("/auth")
	{
		auth.POST("/register", services.Register)
		auth.POST("/login", services.Login)
		auth.POST("/forgot-password", services.ForgotPassword)
		auth.POST("/reset-password", services.ResetPassword)
		auth.GET("/verify-email/:token", services.VerifyEmail)
		
		// OAuth routes
		auth.POST("/google", services.GoogleAuth)
		auth.POST("/github", services.GitHubAuth)
		auth.GET("/oauth/health", services.OAuthHealthCheck)
	}

	// Protected user routes
	users := router.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("/profile", services.GetProfile)
		users.PUT("/profile", services.UpdateProfile)
		users.DELETE("/account", services.DeleteAccount)
		users.POST("/change-password", services.ChangePassword)
	}

	// Website management routes
	websites := router.Group("/websites")
	websites.Use(middleware.AuthMiddleware())
	{
		websites.GET("/", services.GetWebsites)
		websites.POST("/", services.CreateWebsite)
		websites.GET("/by-site-id/:id", services.GetWebsiteBySiteID)
		websites.GET("/:id", services.GetWebsite)
		websites.PUT("/:id", services.UpdateWebsite)
		websites.DELETE("/:id", services.DeleteWebsite)
		websites.POST("/:id/regenerate-tracking", services.RegenerateTrackingID)
	}

	// Privacy routes
	privacy := router.Group("/privacy")
	privacy.Use(middleware.AuthMiddleware())
	{
		privacy.GET("/settings", services.GetPrivacySettings)
		privacy.PUT("/settings", services.UpdatePrivacySettings)
		privacy.GET("/export", services.ExportUserData)
		privacy.GET("/download", services.DownloadUserData)
		privacy.GET("/requests", services.GetPrivacyRequests)
		privacy.POST("/requests", services.CreatePrivacyRequest)
		privacy.GET("/compliance", services.GetComplianceStatus)
		privacy.POST("/delete", services.DeleteUserData)
	}

	// Billing routes (cloud features only)
	billing := router.Group("/billing")
	billing.Use(middleware.AuthMiddleware())
	billing.Use(middleware.RequireCloudFeatures())
	{
		billing.GET("/subscription", services.GetSubscription)
		billing.GET("/usage", services.GetSubscriptionUsage)
		billing.POST("/checkout", services.CreateCheckoutSession)
		billing.POST("/create-portal", services.CreatePortalSession)
		billing.GET("/invoices", services.GetInvoices)
	}

	// Support routes
	support := router.Group("/support")
	support.Use(middleware.AuthMiddleware())
	{
		support.POST("/contact", services.CreateSupportTicket)
		support.POST("/tickets", services.CreateSupportTicket)
		support.GET("/tickets", services.GetSupportTickets)
		support.GET("/tickets/:id", services.GetSupportTicket)
	}

	// Events routes
	events := router.Group("/events")
	events.Use(middleware.OptionalAuthMiddleware())
	{
		events.POST("/track", services.TrackEvent)
		events.GET("/", middleware.AuthMiddleware(), services.GetEvents)
	}

	// Webhook routes (no auth required)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/lemonsqueezy", services.HandleLemonSqueezyWebhook)
	}
}

func SetupWorkflowRoutes(router *gin.RouterGroup) {
	// All workflow routes require authentication
	router.Use(middleware.AuthMiddleware())

	// Workflow CRUD operations
	router.GET("/", services.GetWorkflows)
	router.POST("/", middleware.CheckSubscriptionLimit("workflows"), services.CreateWorkflow)
	router.GET("/:id", services.GetWorkflow)
	router.PUT("/:id", services.UpdateWorkflow)
	router.DELETE("/:id", services.DeleteWorkflow)

	// Workflow execution
	router.POST("/:id/execute", services.ExecuteWorkflow)
	router.GET("/:id/executions", services.GetWorkflowExecutions)
	router.GET("/:id/executions/:executionId", services.GetWorkflowExecution)

	// Workflow analytics
	analytics := router.Group("/analytics")
	{
		analytics.GET("/overview", services.GetWorkflowAnalytics)
		analytics.GET("/:id/stats", services.GetWorkflowStats)
		analytics.GET("/:id/performance", services.GetWorkflowPerformance)
		
		// Detailed workflow analytics endpoints
		analytics.GET("/workflow/:id/activity", services.GetWorkflowActivity)
		analytics.GET("/workflow/:id/chart", services.GetWorkflowSummary)
		analytics.GET("/workflow/:id/nodes", services.GetWorkflowNodePerformance)
		analytics.GET("/workflow/:id/triggers", services.GetWorkflowTriggerTypes)
		analytics.GET("/workflow/:id/actions", services.GetWorkflowActionTypes)
		analytics.GET("/workflow/:id/hourly", services.GetWorkflowHourlyData)
		analytics.GET("/funnel/:id", services.GetWorkflowFunnelData)
	}

	// Workflow stats summary endpoint
	router.GET("/stats/summary", services.GetWorkflowsSummary)

	// Workflow stats (frontend expects this exact path)
	router.GET("/:id/stats", services.GetWorkflowStats)
	router.PATCH("/:id/status", services.UpdateWorkflowStatus)

	// Visitor tracking
	visitor := router.Group("/visitor")
	visitor.Use(middleware.OptionalAuthMiddleware())
	{
		visitor.POST("/track", services.TrackVisitorEvent)
		visitor.GET("/events", middleware.AuthMiddleware(), services.GetVisitorEvents)
		visitor.GET("/analytics", middleware.AuthMiddleware(), services.GetVisitorAnalytics)
	}
}

func SetupAdminRoutes(router *gin.RouterGroup) {
	// All admin routes require authentication and admin role
	router.Use(middleware.AuthMiddleware())
	// TODO: Add admin role middleware

	// User management
	users := router.Group("/users")
	{
		users.GET("/", services.AdminGetUsers)
		users.GET("/:id", services.AdminGetUser)
		users.PUT("/:id", services.AdminUpdateUser)
		users.DELETE("/:id", services.AdminDeleteUser)
		users.POST("/:id/suspend", services.AdminSuspendUser)
		users.POST("/:id/activate", services.AdminActivateUser)
	}

	// Website management
	websites := router.Group("/websites")
	{
		websites.GET("/", services.AdminGetWebsites)
		websites.GET("/:id", services.AdminGetWebsite)
		websites.DELETE("/:id", services.AdminDeleteWebsite)
	}

	// System statistics
	router.GET("/stats", services.GetAdminStats)
	router.GET("/health", services.GetSystemHealth)
}

func SetupInternalRoutes(router *gin.RouterGroup) {
	// Internal routes for microservice communication
	// These routes are used by other services and should not require user authentication
	// but should validate API key

	// User validation for other services
	router.GET("/users/:id/validate", services.ValidateUser)
	router.GET("/websites/:id/validate", services.ValidateWebsite)
	
	// Workflow triggers from external services
	router.POST("/workflows/:id/trigger", services.TriggerWorkflow)
	
	// Analytics data for other services
	router.GET("/analytics/summary", services.GetAnalyticsSummary)
}
