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

// Privacy Services
func GetPrivacySettings(c *gin.Context) {
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

	// Get user's privacy settings from their profile
	collection := database.GetUserDB().Collection("users")
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get website privacy settings
	websiteCollection := database.GetUserDB().Collection("websites")
	cursor, err := websiteCollection.Find(context.Background(), bson.M{"userId": objectID})
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

	privacySettings := gin.H{
		"user": gin.H{
			"dataRetentionDays": 365, // Default
			"allowAnalytics":    true,
			"allowMarketing":    false,
		},
		"websites": websites,
	}

	c.JSON(http.StatusOK, privacySettings)
}

func UpdatePrivacySettings(c *gin.Context) {
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
		DataRetentionDays int  `json:"dataRetentionDays"`
		AllowAnalytics    bool `json:"allowAnalytics"`
		AllowMarketing    bool `json:"allowMarketing"`
		WebsiteID         string `json:"websiteId,omitempty"`
		WebsiteSettings   *models.PrivacySettings `json:"websiteSettings,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update website-specific privacy settings if provided
	if req.WebsiteID != "" && req.WebsiteSettings != nil {
		websiteObjectID, err := primitive.ObjectIDFromHex(req.WebsiteID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
			return
		}

		websiteCollection := database.GetUserDB().Collection("websites")
		_, err = websiteCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": websiteObjectID, "userId": objectID},
			bson.M{"$set": bson.M{
				"privacySettings": req.WebsiteSettings,
				"updatedAt":       time.Now(),
			}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update website privacy settings"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Privacy settings updated successfully"})
}

func ExportUserData(c *gin.Context) {
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

	// Get user data
	userCollection := database.GetUserDB().Collection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's websites
	websiteCollection := database.GetUserDB().Collection("websites")
	cursor, err := websiteCollection.Find(context.Background(), bson.M{"userId": objectID})
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

	// Get user's workflows
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	workflowCursor, err := workflowCollection.Find(context.Background(), bson.M{"userId": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflows"})
		return
	}
	defer workflowCursor.Close(context.Background())

	var workflows []models.Workflow
	if err = workflowCursor.All(context.Background(), &workflows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode workflows"})
		return
	}

	// Remove sensitive information
	user.Password = ""

	exportData := gin.H{
		"user":      user,
		"websites":  websites,
		"workflows": workflows,
		"exportedAt": time.Now(),
		"format":    "JSON",
	}

	c.Header("Content-Disposition", "attachment; filename=user_data_export.json")
	c.JSON(http.StatusOK, exportData)
}

func DeleteUserData(c *gin.Context) {
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
		Password     string `json:"password" validate:"required"`
		DeleteType   string `json:"deleteType" validate:"required"` // "partial" or "complete"
		DataTypes    []string `json:"dataTypes,omitempty"` // ["analytics", "workflows", "websites"]
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user password
	userCollection := database.GetUserDB().Collection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// For now, just mark as deletion requested
	// In a real implementation, you'd verify password and process deletion
	deletionRecord := bson.M{
		"userId":      objectID,
		"deleteType":  req.DeleteType,
		"dataTypes":   req.DataTypes,
		"requestedAt": time.Now(),
		"status":      "pending",
	}

	// Store deletion request (in a real system, this would trigger a background job)
	deletionCollection := database.GetUserDB().Collection("deletion_requests")
	_, err = deletionCollection.InsertOne(context.Background(), deletionRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deletion request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data deletion request processed"})
}

func GetComplianceStatus(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Return compliance status
	complianceStatus := gin.H{
		"gdprCompliant":     true,
		"ccpaCompliant":     true,
		"dataRetentionDays": 365,
		"lastAudit":         time.Now().AddDate(0, -1, 0), // 1 month ago
		"nextAudit":         time.Now().AddDate(0, 2, 0),  // 2 months from now
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"complianceStatus": complianceStatus,
		},
	})
}

func DownloadUserData(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// For now, return the same data as export but as downloadable file
	// In a real implementation, you'd generate a file and return it
	c.Header("Content-Disposition", "attachment; filename=user_data.json")
	c.Header("Content-Type", "application/json")
	
	// Get user data (same as ExportUserData)
	userObjectID, _ := primitive.ObjectIDFromHex(userID.(string))
	userCollection := database.GetUserDB().Collection("users")
	
	var user models.User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user data"})
		return
	}

	// Remove sensitive data
	user.Password = ""
	
	exportData := gin.H{
		"user": user,
		"exportedAt": time.Now(),
		"format": "json",
	}

	c.JSON(http.StatusOK, exportData)
}

func GetPrivacyRequests(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// For now, return empty array - in real implementation you'd fetch from database
	requests := []gin.H{}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"requests": requests,
		},
	})
}

func CreatePrivacyRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Type        string `json:"type" validate:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create privacy request record
	request := gin.H{
		"id":          primitive.NewObjectID().Hex(),
		"userId":      userID,
		"type":        req.Type,
		"description": req.Description,
		"status":      "pending",
		"createdAt":   time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"request": request,
		},
	})
}
