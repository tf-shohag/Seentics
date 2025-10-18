package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/config"
	"github.com/seentics/seentics/services/common/database"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subscription plans mapping
var SUBSCRIPTION_PLANS = map[string]string{
	"free":     "free",
	"standard": "standard",
	"pro":      "pro",
}

// Lemon Squeezy webhook data structure
type LemonSqueezyWebhookData struct {
	ID         string `json:"id"`
	Attributes struct {
		CustomData struct {
			UserID string `json:"user_id"`
		} `json:"custom_data"`
		VariantID   string    `json:"variant_id"`
		CustomerID  string    `json:"customer_id"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		RenewsAt    time.Time `json:"renews_at"`
		CancelledAt time.Time `json:"cancelled_at"`
		InvoiceID   string    `json:"invoice_id"`
		Total       int       `json:"total"`
		Currency    string    `json:"currency"`
	} `json:"attributes"`
}

// Subscription model for database operations
type Subscription struct {
	ID                         primitive.ObjectID `bson:"_id,omitempty"`
	UserID                     primitive.ObjectID `bson:"userId"`
	Plan                       string             `bson:"plan"`
	Status                     string             `bson:"status"`
	LemonSqueezyCustomerID     string             `bson:"lemonSqueezyCustomerId"`
	LemonSqueezySubscriptionID string             `bson:"lemonSqueezySubscriptionId"`
	LemonSqueezyVariantID      string             `bson:"lemonSqueezyVariantId"`
	CurrentPeriodStart         time.Time          `bson:"currentPeriodStart"`
	CurrentPeriodEnd           *time.Time         `bson:"currentPeriodEnd"`
	CancelAtPeriodEnd          bool               `bson:"cancelAtPeriodEnd"`
	CancelledAt                *time.Time         `bson:"cancelledAt"`
	CreatedAt                  time.Time          `bson:"createdAt"`
	UpdatedAt                  time.Time          `bson:"updatedAt"`
}

// Webhook Services - Lemon Squeezy Integration
func HandleLemonSqueezyWebhook(c *gin.Context) {
	// Get webhook signature
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing webhook signature"})
		return
	}

	// Read raw body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Verify webhook signature
	cfg := config.Load()
	if !verifyLemonSqueezySignature(string(body), signature, cfg.LemonSqueezyWebhookSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Parse webhook payload
	var webhook struct {
		Meta struct {
			EventName string `json:"event_name"`
		} `json:"meta"`
		Data LemonSqueezyWebhookData `json:"data"`
	}

	if err := json.Unmarshal(body, &webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	eventName := webhook.Meta.EventName
	logrus.Info("Received Lemon Squeezy webhook:", eventName)

	// Handle different webhook events
	switch eventName {
	case "subscription_created":
		err = handleSubscriptionCreated(webhook.Data)
	case "subscription_updated":
		err = handleSubscriptionUpdated(webhook.Data)
	case "subscription_cancelled":
		err = handleSubscriptionCancelled(webhook.Data)
	case "subscription_resumed":
		err = handleSubscriptionResumed(webhook.Data)
	case "subscription_expired":
		err = handleSubscriptionExpired(webhook.Data)
	case "subscription_paused":
		err = handleSubscriptionPaused(webhook.Data)
	case "subscription_unpaused":
		err = handleSubscriptionUnpaused(webhook.Data)
	case "subscription_payment_failed":
		err = handleSubscriptionPaymentFailed(webhook.Data)
	case "subscription_payment_success":
		err = handleSubscriptionPaymentSuccess(webhook.Data)
	case "order_created":
		err = handleOrderCreated(webhook.Data)
	default:
		logrus.Warn("Unhandled webhook event:", eventName)
	}

	if err != nil {
		logrus.Error("Webhook processing error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook processing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Verify Lemon Squeezy webhook signature
func verifyLemonSqueezySignature(payload, signature, secret string) bool {
	if secret == "" {
		logrus.Warn("Lemon Squeezy webhook secret not configured")
		return false
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// Handle subscription created
func handleSubscriptionCreated(data LemonSqueezyWebhookData) error {
	userID := data.Attributes.CustomData.UserID
	if userID == "" {
		logrus.Error("No user_id in subscription created webhook")
		return nil
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		logrus.Error("Invalid user ID:", userID)
		return err
	}

	// Get variant to plan mapping from config
	cfg := config.Load()
	var plan string
	switch data.Attributes.VariantID {
	case cfg.LemonSqueezyStandardVariantID:
		plan = SUBSCRIPTION_PLANS["standard"]
	case cfg.LemonSqueezyProVariantID:
		plan = SUBSCRIPTION_PLANS["pro"]
	default:
		logrus.Error("Unknown variant ID:", data.Attributes.VariantID)
		return nil
	}

	// Find or create subscription
	collection := database.GetUserDB().Collection("subscriptions")

	subscription := Subscription{
		UserID:                     userObjectID,
		Plan:                       plan,
		Status:                     "active",
		LemonSqueezySubscriptionID: data.ID,
		LemonSqueezyCustomerID:     data.Attributes.CustomerID,
		LemonSqueezyVariantID:      data.Attributes.VariantID,
		CurrentPeriodStart:         data.Attributes.CreatedAt,
		CurrentPeriodEnd:           &data.Attributes.RenewsAt,
		CancelAtPeriodEnd:          false,
		CancelledAt:                nil,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	// Upsert subscription
	_, err = collection.ReplaceOne(
		context.Background(),
		bson.M{"userId": userObjectID},
		subscription,
	)
	if err != nil {
		logrus.Error("Failed to create/update subscription:", err)
		return err
	}

	// Update user subscription info
	userCollection := database.GetUserDB().Collection("users")
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": userObjectID},
		bson.M{"$set": bson.M{
			"subscriptionPlan":   plan,
			"subscriptionStatus": "active",
			"updatedAt":          time.Now(),
		}},
	)
	if err != nil {
		logrus.Error("Failed to update user subscription:", err)
		return err
	}

	logrus.Info("Subscription created for user:", userID, "plan:", plan)
	return nil
}

// Handle subscription updated
func handleSubscriptionUpdated(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	update := bson.M{
		"$set": bson.M{
			"status":             data.Attributes.Status,
			"currentPeriodStart": data.Attributes.CreatedAt,
			"currentPeriodEnd":   data.Attributes.RenewsAt,
			"updatedAt":          time.Now(),
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to update subscription:", err)
		return err
	}

	logrus.Info("Subscription updated:", data.ID)
	return nil
}

// Handle subscription cancelled
func handleSubscriptionCancelled(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	update := bson.M{
		"$set": bson.M{
			"status":            "cancelled",
			"cancelAtPeriodEnd": true,
			"cancelledAt":       data.Attributes.CancelledAt,
			"updatedAt":         time.Now(),
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to cancel subscription:", err)
		return err
	}

	logrus.Info("Subscription cancelled:", data.ID)
	return nil
}

// Handle subscription resumed
func handleSubscriptionResumed(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	update := bson.M{
		"$set": bson.M{
			"status":            "active",
			"cancelAtPeriodEnd": false,
			"cancelledAt":       nil,
			"updatedAt":         time.Now(),
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to resume subscription:", err)
		return err
	}

	logrus.Info("Subscription resumed:", data.ID)
	return nil
}

// Handle subscription expired
func handleSubscriptionExpired(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	// Find subscription to get user ID
	var subscription Subscription
	err := collection.FindOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
	).Decode(&subscription)
	if err != nil {
		logrus.Error("Subscription not found:", data.ID)
		return err
	}

	// Downgrade to free plan
	update := bson.M{
		"$set": bson.M{
			"plan":                       SUBSCRIPTION_PLANS["free"],
			"status":                     "expired",
			"lemonSqueezySubscriptionId": nil,
			"lemonSqueezyVariantId":      nil,
			"currentPeriodEnd":           nil,
			"updatedAt":                  time.Now(),
		},
	}

	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to expire subscription:", err)
		return err
	}

	// Update user subscription info
	userCollection := database.GetUserDB().Collection("users")
	_, err = userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": subscription.UserID},
		bson.M{"$set": bson.M{
			"subscriptionPlan":   SUBSCRIPTION_PLANS["free"],
			"subscriptionStatus": "expired",
			"updatedAt":          time.Now(),
		}},
	)
	if err != nil {
		logrus.Error("Failed to update user subscription:", err)
		return err
	}

	logrus.Info("Subscription expired, downgraded to free:", data.ID)
	return nil
}

// Handle subscription paused
func handleSubscriptionPaused(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	update := bson.M{
		"$set": bson.M{
			"status":    "paused",
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to pause subscription:", err)
		return err
	}

	logrus.Info("Subscription paused:", data.ID)
	return nil
}

// Handle subscription unpaused
func handleSubscriptionUnpaused(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	update := bson.M{
		"$set": bson.M{
			"status":    "active",
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to unpause subscription:", err)
		return err
	}

	logrus.Info("Subscription unpaused:", data.ID)
	return nil
}

// Handle subscription payment failed
func handleSubscriptionPaymentFailed(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	update := bson.M{
		"$set": bson.M{
			"status":    "past_due",
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to update subscription payment failed:", err)
		return err
	}

	logrus.Info("Subscription payment failed:", data.ID)
	return nil
}

// Handle subscription payment success
func handleSubscriptionPaymentSuccess(data LemonSqueezyWebhookData) error {
	collection := database.GetUserDB().Collection("subscriptions")

	// Create invoice record
	invoice := bson.M{
		"lemonSqueezyInvoiceId": data.Attributes.InvoiceID,
		"amount":                data.Attributes.Total,
		"currency":              data.Attributes.Currency,
		"status":                "paid",
		"paidAt":                time.Now(),
		"createdAt":             time.Now(),
	}

	update := bson.M{
		"$set": bson.M{
			"status":    "active",
			"updatedAt": time.Now(),
		},
		"$push": bson.M{
			"invoices": invoice,
		},
	}

	_, err := collection.UpdateOne(
		context.Background(),
		bson.M{"lemonSqueezySubscriptionId": data.ID},
		update,
	)
	if err != nil {
		logrus.Error("Failed to update subscription payment success:", err)
		return err
	}

	logrus.Info("Subscription payment successful:", data.ID)
	return nil
}

// Handle order created (one-time purchases)
func handleOrderCreated(data LemonSqueezyWebhookData) error {
	logrus.Info("Order created:", data.ID)
	// Handle one-time purchases if needed
	return nil
}
