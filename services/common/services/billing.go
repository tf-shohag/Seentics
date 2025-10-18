package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/database"
	"github.com/seentics/seentics/services/common/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Billing Services
func GetSubscription(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collection := database.GetUserDB().Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	subscription := gin.H{
		"plan":             user.SubscriptionPlan,
		"status":           user.SubscriptionStatus,
		"stripeCustomerId": user.StripeCustomerID,
		"features": gin.H{
			"maxWebsites":        getMaxWebsites(user.SubscriptionPlan),
			"maxWorkflows":       getMaxWorkflows(user.SubscriptionPlan),
			"analyticsRetention": getAnalyticsRetention(user.SubscriptionPlan),
			"supportLevel":       getSupportLevel(user.SubscriptionPlan),
		},
	}

	c.JSON(http.StatusOK, subscription)
}

func CreateCheckoutSession(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		PlanID    string `json:"planId" validate:"required"`
		ReturnURL string `json:"returnUrl" validate:"required"`
		CancelURL string `json:"cancelUrl" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, you would:
	// 1. Create a Stripe checkout session
	// 2. Store session info in database
	// 3. Return the checkout URL

	// Mock response for now
	checkoutSession := gin.H{
		"sessionId":   "cs_test_" + primitive.NewObjectID().Hex(),
		"checkoutUrl": "https://checkout.stripe.com/pay/cs_test_" + primitive.NewObjectID().Hex(),
		"planId":      req.PlanID,
		"createdAt":   time.Now(),
	}

	c.JSON(http.StatusOK, checkoutSession)
}

func CreatePortalSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		ReturnURL string `json:"returnUrl" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user to check if they have a Stripe customer ID
	collection := database.GetUserDB().Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.StripeCustomerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active subscription found"})
		return
	}

	// In a real implementation, you would create a Stripe portal session
	portalSession := gin.H{
		"portalUrl": "https://billing.stripe.com/p/session/" + primitive.NewObjectID().Hex(),
		"createdAt": time.Now(),
	}

	c.JSON(http.StatusOK, portalSession)
}

func GetInvoices(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user to check Stripe customer ID
	collection := database.GetUserDB().Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Mock invoices for now
	// In a real implementation, you would fetch from Stripe
	invoices := []gin.H{
		{
			"id":          "in_" + primitive.NewObjectID().Hex(),
			"amount":      2900, // $29.00 in cents
			"currency":    "usd",
			"status":      "paid",
			"date":        time.Now().AddDate(0, -1, 0),
			"description": "Seentics Pro Plan",
			"downloadUrl": "/api/v1/user/billing/invoices/in_" + primitive.NewObjectID().Hex() + "/download",
		},
		{
			"id":          "in_" + primitive.NewObjectID().Hex(),
			"amount":      2900,
			"currency":    "usd",
			"status":      "paid",
			"date":        time.Now().AddDate(0, -2, 0),
			"description": "Seentics Pro Plan",
			"downloadUrl": "/api/v1/user/billing/invoices/in_" + primitive.NewObjectID().Hex() + "/download",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"hasMore":  false,
	})
}

// Helper functions for subscription features
func getMaxWebsites(plan string) int {
	switch plan {
	case "free":
		return 1
	case "pro":
		return 10
	case "enterprise":
		return -1 // unlimited
	default:
		return 1
	}
}

func getMaxWorkflows(plan string) int {
	switch plan {
	case "free":
		return 3
	case "pro":
		return 50
	case "enterprise":
		return -1 // unlimited
	default:
		return 3
	}
}

func getAnalyticsRetention(plan string) int {
	switch plan {
	case "free":
		return 30 // 30 days
	case "pro":
		return 365 // 1 year
	case "enterprise":
		return 1095 // 3 years
	default:
		return 30
	}
}

func getSupportLevel(plan string) string {
	switch plan {
	case "free":
		return "email"
	case "standard":
		return "priority"
	case "pro":
		return "24/7"
	default:
		return "email"
	}
}

func GetSubscriptionUsage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user subscription info
	userCollection := database.GetUserDB().Collection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get current usage counts
	websiteCollection := database.GetUserDB().Collection("websites")
	websiteCount, _ := websiteCollection.CountDocuments(context.Background(), bson.M{"userId": userObjectID})

	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	workflowCount, _ := workflowCollection.CountDocuments(context.Background(), bson.M{"userId": userObjectID})

	// Get plan limits
	maxWebsites := getMaxWebsites(user.SubscriptionPlan)
	maxWorkflows := getMaxWorkflows(user.SubscriptionPlan)

	usage := gin.H{
		"plan": user.SubscriptionPlan,
		"status": user.SubscriptionStatus,
		"usage": gin.H{
			"websites": gin.H{
				"current": websiteCount,
				"limit":   maxWebsites,
			},
			"workflows": gin.H{
				"current": workflowCount,
				"limit":   maxWorkflows,
			},
			"monthlyEvents": gin.H{
				"current": 0, // TODO: Implement actual event counting
				"limit":   1000, // TODO: Get from plan limits
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    usage,
	})
}
