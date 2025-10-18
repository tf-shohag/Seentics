package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SupportTicket struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"userId"`
	Subject     string             `json:"subject" bson:"subject"`
	Description string             `json:"description" bson:"description"`
	Priority    string             `json:"priority" bson:"priority"` // low, medium, high, urgent
	Status      string             `json:"status" bson:"status"`     // open, in_progress, resolved, closed
	Category    string             `json:"category" bson:"category"` // technical, billing, feature_request, bug_report
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	ResolvedAt  *time.Time         `json:"resolvedAt,omitempty" bson:"resolvedAt,omitempty"`
	
	// Support agent info
	AssignedTo  string `json:"assignedTo,omitempty" bson:"assignedTo,omitempty"`
	
	// Communication
	Messages    []SupportMessage `json:"messages,omitempty" bson:"messages,omitempty"`
}

type SupportMessage struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	From      string             `json:"from" bson:"from"` // user or agent email
	Message   string             `json:"message" bson:"message"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	IsInternal bool              `json:"isInternal" bson:"isInternal"` // internal agent notes
}

// Support Services
func CreateSupportTicket(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userEmail, _ := c.Get("email")

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Subject     string `json:"subject" validate:"required"`
		Description string `json:"description" validate:"required"`
		Priority    string `json:"priority" validate:"required"`
		Category    string `json:"category" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate priority and category
	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "urgent": true}
	validCategories := map[string]bool{"technical": true, "billing": true, "feature_request": true, "bug_report": true}

	if !validPriorities[req.Priority] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority"})
		return
	}

	if !validCategories[req.Category] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category"})
		return
	}

	ticket := SupportTicket{
		ID:          primitive.NewObjectID(),
		UserID:      objectID,
		Subject:     req.Subject,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      "open",
		Category:    req.Category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Messages: []SupportMessage{
			{
				ID:        primitive.NewObjectID(),
				From:      userEmail.(string),
				Message:   req.Description,
				CreatedAt: time.Now(),
				IsInternal: false,
			},
		},
	}

	collection := database.GetUserDB().Collection("support_tickets")
	_, err = collection.InsertOne(context.Background(), ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create support ticket"})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func GetSupportTickets(c *gin.Context) {
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

	// Query parameters for filtering
	status := c.Query("status")
	priority := c.Query("priority")
	category := c.Query("category")

	// Build filter
	filter := bson.M{"userId": objectID}
	if status != "" {
		filter["status"] = status
	}
	if priority != "" {
		filter["priority"] = priority
	}
	if category != "" {
		filter["category"] = category
	}

	collection := database.GetUserDB().Collection("support_tickets")
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch support tickets"})
		return
	}
	defer cursor.Close(context.Background())

	var tickets []SupportTicket
	if err = cursor.All(context.Background(), &tickets); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode support tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tickets": tickets,
		"count":   len(tickets),
	})
}

func GetSupportTicket(c *gin.Context) {
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

	ticketID := c.Param("id")
	ticketObjectID, err := primitive.ObjectIDFromHex(ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	collection := database.GetUserDB().Collection("support_tickets")
	var ticket SupportTicket
	err = collection.FindOne(context.Background(), bson.M{
		"_id":    ticketObjectID,
		"userId": userObjectID,
	}).Decode(&ticket)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Support ticket not found"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func AddSupportMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userEmail, _ := c.Get("email")
	userObjectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ticketID := c.Param("id")
	ticketObjectID, err := primitive.ObjectIDFromHex(ticketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	var req struct {
		Message string `json:"message" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := SupportMessage{
		ID:        primitive.NewObjectID(),
		From:      userEmail.(string),
		Message:   req.Message,
		CreatedAt: time.Now(),
		IsInternal: false,
	}

	collection := database.GetUserDB().Collection("support_tickets")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": ticketObjectID, "userId": userObjectID},
		bson.M{
			"$push": bson.M{"messages": message},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message added successfully",
		"messageId": message.ID.Hex(),
	})
}
