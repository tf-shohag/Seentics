package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seentics/seentics/services/common/database"
	"github.com/seentics/seentics/services/common/models"
	"github.com/seentics/seentics/services/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetWebsites(c *gin.Context) {
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

	collection := database.GetUserDB().Collection("websites")
	cursor, err := collection.Find(context.Background(), bson.M{"userId": objectID})
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

	c.JSON(http.StatusOK, websites)
}

func CreateWebsite(c *gin.Context) {
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
		Name   string `json:"name" validate:"required"`
		URL    string `json:"url" validate:"required,url"`
		Domain string `json:"domain" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate tracking ID
	trackingID := uuid.New().String()

	website := models.Website{
		ID:         primitive.NewObjectID(),
		UserID:     objectID,
		Name:       req.Name,
		URL:        req.URL,
		Domain:     req.Domain,
		TrackingID: trackingID,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		AllowedOrigins: []string{req.Domain},
		DataRetention:  365, // Default 1 year
		PrivacySettings: models.PrivacySettings{
			CookieConsent:     true,
			AnonymizeIPs:      true,
			DataRetentionDays: 365,
			GDPRCompliant:     true,
		},
	}

	collection := database.GetUserDB().Collection("websites")
	_, err = collection.InsertOne(context.Background(), website)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create website"})
		return
	}

	c.JSON(http.StatusCreated, website)
}

func GetWebsite(c *gin.Context) {
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

	websiteID := c.Param("id")
	websiteObjectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	collection := database.GetUserDB().Collection("websites")
	var website models.Website
	err = collection.FindOne(context.Background(), bson.M{
		"_id":    websiteObjectID,
		"userId": userObjectID,
	}).Decode(&website)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}

	c.JSON(http.StatusOK, website)
}

func UpdateWebsite(c *gin.Context) {
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

	websiteID := c.Param("id")
	websiteObjectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	var updateData struct {
		Name            string   `json:"name"`
		URL             string   `json:"url"`
		Domain          string   `json:"domain"`
		IsActive        *bool    `json:"isActive"`
		AllowedOrigins  []string `json:"allowedOrigins"`
		DataRetention   int      `json:"dataRetention"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update document
	updateDoc := bson.M{
		"updatedAt": time.Now(),
	}

	if updateData.Name != "" {
		updateDoc["name"] = updateData.Name
	}
	if updateData.URL != "" {
		updateDoc["url"] = updateData.URL
	}
	if updateData.Domain != "" {
		updateDoc["domain"] = updateData.Domain
	}
	if updateData.IsActive != nil {
		updateDoc["isActive"] = *updateData.IsActive
	}
	if updateData.AllowedOrigins != nil {
		updateDoc["allowedOrigins"] = updateData.AllowedOrigins
	}
	if updateData.DataRetention > 0 {
		updateDoc["dataRetention"] = updateData.DataRetention
	}

	collection := database.GetUserDB().Collection("websites")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": websiteObjectID, "userId": userObjectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update website"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Website updated successfully"})
}

func DeleteWebsite(c *gin.Context) {
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

	websiteID := c.Param("id")
	websiteObjectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	collection := database.GetUserDB().Collection("websites")
	result, err := collection.DeleteOne(context.Background(), bson.M{
		"_id":    websiteObjectID,
		"userId": userObjectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete website"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}

	// TODO: Delete related data (analytics, workflows, etc.)

	c.JSON(http.StatusOK, gin.H{"message": "Website deleted successfully"})
}

func RegenerateTrackingID(c *gin.Context) {
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

	websiteID := c.Param("id")
	websiteObjectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}

	// Generate new tracking ID
	newTrackingID := uuid.New().String()

	collection := database.GetUserDB().Collection("websites")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": websiteObjectID, "userId": userObjectID},
		bson.M{"$set": bson.M{
			"trackingId": newTrackingID,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate tracking ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Tracking ID regenerated successfully",
		"trackingId": newTrackingID,
	})
}
