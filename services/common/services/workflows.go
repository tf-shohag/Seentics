package services

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/database"
	"github.com/seentics/seentics/services/common/models"
	"github.com/seentics/seentics/services/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetWorkflows(c *gin.Context) {
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

	collection := database.GetWorkflowDB().Collection("workflows")
	cursor, err := collection.Find(context.Background(), bson.M{"userId": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workflows"})
		return
	}
	defer cursor.Close(context.Background())

	var workflows []models.Workflow
	if err = cursor.All(context.Background(), &workflows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode workflows"})
		return
	}

	c.JSON(http.StatusOK, workflows)
}

func CreateWorkflow(c *gin.Context) {
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

	var req struct {
		Name          string                     `json:"name" validate:"required"`
		Description   string                     `json:"description"`
		WebsiteID     string                     `json:"websiteId" validate:"required"`
		Triggers      []models.WorkflowTrigger   `json:"triggers"`
		Actions       []models.WorkflowAction    `json:"actions"`
		Conditions    []models.WorkflowCondition `json:"conditions"`
		ExecutionMode string                     `json:"executionMode" validate:"required"`
		Schedule      *models.WorkflowSchedule   `json:"schedule,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	websiteObjectID, err := primitive.ObjectIDFromHex(req.WebsiteID)
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

	workflow := models.Workflow{
		ID:            primitive.NewObjectID(),
		UserID:        userObjectID,
		WebsiteID:     websiteObjectID,
		Name:          req.Name,
		Description:   req.Description,
		IsActive:      true,
		Triggers:      req.Triggers,
		Actions:       req.Actions,
		Conditions:    req.Conditions,
		ExecutionMode: req.ExecutionMode,
		Schedule:      req.Schedule,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		RunCount:      0,
		SuccessRate:   0.0,
		AvgRunTime:    0.0,
	}

	collection := database.GetWorkflowDB().Collection("workflows")
	_, err = collection.InsertOne(context.Background(), workflow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

func GetWorkflow(c *gin.Context) {
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

	workflowID := c.Param("id")
	workflowObjectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	collection := database.GetWorkflowDB().Collection("workflows")
	var workflow models.Workflow
	err = collection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

func UpdateWorkflow(c *gin.Context) {
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

	workflowID := c.Param("id")
	workflowObjectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	var updateData struct {
		Name          string                     `json:"name"`
		Description   string                     `json:"description"`
		IsActive      *bool                      `json:"isActive"`
		Triggers      []models.WorkflowTrigger   `json:"triggers"`
		Actions       []models.WorkflowAction    `json:"actions"`
		Conditions    []models.WorkflowCondition `json:"conditions"`
		ExecutionMode string                     `json:"executionMode"`
		Schedule      *models.WorkflowSchedule   `json:"schedule"`
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
	if updateData.Description != "" {
		updateDoc["description"] = updateData.Description
	}
	if updateData.IsActive != nil {
		updateDoc["isActive"] = *updateData.IsActive
	}
	if updateData.Triggers != nil {
		updateDoc["triggers"] = updateData.Triggers
	}
	if updateData.Actions != nil {
		updateDoc["actions"] = updateData.Actions
	}
	if updateData.Conditions != nil {
		updateDoc["conditions"] = updateData.Conditions
	}
	if updateData.ExecutionMode != "" {
		updateDoc["executionMode"] = updateData.ExecutionMode
	}
	if updateData.Schedule != nil {
		updateDoc["schedule"] = updateData.Schedule
	}

	collection := database.GetWorkflowDB().Collection("workflows")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": workflowObjectID, "userId": userObjectID},
		bson.M{"$set": updateDoc},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow updated successfully"})
}

func DeleteWorkflow(c *gin.Context) {
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

	workflowID := c.Param("id")
	workflowObjectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	collection := database.GetWorkflowDB().Collection("workflows")
	result, err := collection.DeleteOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// TODO: Delete related executions

	c.JSON(http.StatusOK, gin.H{"message": "Workflow deleted successfully"})
}

func UpdateWorkflowStatus(c *gin.Context) {
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

	workflowID := c.Param("id")
	workflowObjectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	var req struct {
		Status string `json:"status" validate:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := []string{"Active", "Paused", "Draft"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be Active, Paused, or Draft"})
		return
	}

	collection := database.GetWorkflowDB().Collection("workflows")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{
			"_id":    workflowObjectID,
			"userId": userObjectID,
		},
		bson.M{"$set": bson.M{
			"status":    req.Status,
			"isActive":  req.Status == "Active",
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update workflow status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow status updated successfully"})
}

func ExecuteWorkflow(c *gin.Context) {
	// Use the real workflow execution implementation
	ExecuteWorkflowReal(c)
}

func GetWorkflowExecutions(c *gin.Context) {
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

	workflowID := c.Param("id")
	workflowObjectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	// Verify workflow belongs to user
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	var workflow models.Workflow
	err = workflowCollection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	collection := database.GetWorkflowDB().Collection("workflow_executions")
	cursor, err := collection.Find(context.Background(), bson.M{"workflowId": workflowObjectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch executions"})
		return
	}
	defer cursor.Close(context.Background())

	var executions []models.WorkflowExecution
	if err = cursor.All(context.Background(), &executions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode executions"})
		return
	}

	c.JSON(http.StatusOK, executions)
}

func GetWorkflowExecution(c *gin.Context) {
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

	workflowID := c.Param("id")
	workflowObjectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	executionID := c.Param("executionId")
	executionObjectID, err := primitive.ObjectIDFromHex(executionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid execution ID"})
		return
	}

	// Verify workflow belongs to user
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	var workflow models.Workflow
	err = workflowCollection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	collection := database.GetWorkflowDB().Collection("workflow_executions")
	var execution models.WorkflowExecution
	err = collection.FindOne(context.Background(), bson.M{
		"_id":        executionObjectID,
		"workflowId": workflowObjectID,
	}).Decode(&execution)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Execution not found"})
		return
	}

	c.JSON(http.StatusOK, execution)
}
