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

// Events Services
func TrackEvent(c *gin.Context) {
	var req struct {
		WebsiteID   string                 `json:"websiteId" validate:"required"`
		VisitorID   string                 `json:"visitorId" validate:"required"`
		SessionID   string                 `json:"sessionId" validate:"required"`
		EventType   string                 `json:"eventType" validate:"required"`
		EventData   map[string]interface{} `json:"eventData"`
		URL         string                 `json:"url" validate:"required"`
		Title       string                 `json:"title"`
		Referrer    string                 `json:"referrer"`
		UserAgent   string                 `json:"userAgent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	websiteObjectID, err := primitive.ObjectIDFromHex(req.WebsiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	// Verify website exists and is active
	websiteCollection := database.GetUserDB().Collection("websites")
	var website models.Website
	err = websiteCollection.FindOne(context.Background(), bson.M{
		"_id":      websiteObjectID,
		"isActive": true,
	}).Decode(&website)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found or inactive"})
		return
	}

	// Get client IP and basic geo info
	clientIP := c.ClientIP()
	
	// Create visitor event
	event := models.VisitorEvent{
		ID:        primitive.NewObjectID(),
		WebsiteID: websiteObjectID,
		VisitorID: req.VisitorID,
		SessionID: req.SessionID,
		EventType: req.EventType,
		EventData: req.EventData,
		URL:       req.URL,
		Title:     req.Title,
		Referrer:  req.Referrer,
		UserAgent: req.UserAgent,
		IPAddress: clientIP,
		Country:   "Unknown", // In real implementation, use GeoIP service
		City:      "Unknown",
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}

	// Store in workflow database
	collection := database.GetWorkflowDB().Collection("visitor_events")
	_, err = collection.InsertOne(context.Background(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track event"})
		return
	}

	// TODO: Trigger any workflows that match this event

	c.JSON(http.StatusOK, gin.H{
		"message": "Event tracked successfully",
		"eventId": event.ID.Hex(),
	})
}

func GetEvents(c *gin.Context) {
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

	// Query parameters
	websiteID := c.Query("websiteId")
	eventType := c.Query("eventType")
	limit := 100 // Default limit

	// Build filter - first get user's websites
	websiteCollection := database.GetUserDB().Collection("websites")
	cursor, err := websiteCollection.Find(context.Background(), bson.M{"userId": userObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch websites"})
		return
	}
	defer cursor.Close(context.Background())

	var websites []models.Website
	if err = cursor.All(context.Background(), &websites); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode websites"})
		return
	}

	// Extract website IDs
	var websiteIDs []primitive.ObjectID
	for _, website := range websites {
		websiteIDs = append(websiteIDs, website.ID)
	}

	// Build event filter
	eventFilter := bson.M{"websiteId": bson.M{"$in": websiteIDs}}
	
	if websiteID != "" {
		websiteObjectID, err := primitive.ObjectIDFromHex(websiteID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
			return
		}
		eventFilter["websiteId"] = websiteObjectID
	}
	
	if eventType != "" {
		eventFilter["eventType"] = eventType
	}

	// Get events
	eventCollection := database.GetWorkflowDB().Collection("visitor_events")
	eventCursor, err := eventCollection.Find(context.Background(), eventFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	defer eventCursor.Close(context.Background())

	var events []models.VisitorEvent
	count := 0
	for eventCursor.Next(context.Background()) && count < limit {
		var event models.VisitorEvent
		if err := eventCursor.Decode(&event); err != nil {
			continue
		}
		events = append(events, event)
		count++
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// Visitor Services
func TrackVisitorEvent(c *gin.Context) {
	// This is the same as TrackEvent for now
	TrackEvent(c)
}

func GetVisitorEvents(c *gin.Context) {
	// This is the same as GetEvents for now
	GetEvents(c)
}

func GetVisitorAnalytics(c *gin.Context) {
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

	websiteID := c.Query("websiteId")
	if websiteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Website ID is required"})
		return
	}

	websiteObjectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	// Verify website belongs to user
	websiteCollection := database.GetUserDB().Collection("websites")
	var website models.Website
	err = websiteCollection.FindOne(context.Background(), bson.M{
		"_id":    websiteObjectID,
		"userId": userObjectID,
	}).Decode(&website)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}

	// Get analytics data
	eventCollection := database.GetWorkflowDB().Collection("visitor_events")
	
	// Total events
	totalEvents, _ := eventCollection.CountDocuments(context.Background(), bson.M{"websiteId": websiteObjectID})
	
	// Unique visitors (approximate)
	uniqueVisitors, _ := eventCollection.Distinct(context.Background(), "visitorId", bson.M{"websiteId": websiteObjectID})
	
	// Page views
	pageViews, _ := eventCollection.CountDocuments(context.Background(), bson.M{
		"websiteId": websiteObjectID,
		"eventType": "pageview",
	})

	// Top pages
	pipeline := []bson.M{
		{"$match": bson.M{"websiteId": websiteObjectID, "eventType": "pageview"}},
		{"$group": bson.M{
			"_id":   "$url",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 10},
	}
	
	topPagesCursor, _ := eventCollection.Aggregate(context.Background(), pipeline)
	var topPages []bson.M
	topPagesCursor.All(context.Background(), &topPages)

	analytics := gin.H{
		"totalEvents":    totalEvents,
		"uniqueVisitors": len(uniqueVisitors),
		"pageViews":      pageViews,
		"topPages":       topPages,
		"period":         "last_30_days", // In real implementation, make this configurable
	}

	c.JSON(http.StatusOK, analytics)
}
