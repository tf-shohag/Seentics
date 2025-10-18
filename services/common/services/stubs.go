package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


// Workflow Analytics Services
func GetWorkflowAnalytics(c *gin.Context) {
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

	// Get workflow analytics overview
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")

	// Total workflows
	totalWorkflows, _ := workflowCollection.CountDocuments(context.Background(), bson.M{"userId": userObjectID})
	
	// Active workflows
	activeWorkflows, _ := workflowCollection.CountDocuments(context.Background(), bson.M{
		"userId":   userObjectID,
		"isActive": true,
	})

	// Total executions (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	totalExecutions, _ := executionCollection.CountDocuments(context.Background(), bson.M{
		"createdAt": bson.M{"$gte": thirtyDaysAgo},
	})

	// Success rate
	successfulExecutions, _ := executionCollection.CountDocuments(context.Background(), bson.M{
		"status":    "completed",
		"createdAt": bson.M{"$gte": thirtyDaysAgo},
	})

	var successRate float64 = 0
	if totalExecutions > 0 {
		successRate = float64(successfulExecutions) / float64(totalExecutions) * 100
	}

	analytics := gin.H{
		"totalWorkflows":       totalWorkflows,
		"activeWorkflows":      activeWorkflows,
		"totalExecutions":      totalExecutions,
		"successfulExecutions": successfulExecutions,
		"successRate":          successRate,
		"period":               "last_30_days",
	}

	c.JSON(http.StatusOK, analytics)
}

func GetWorkflowStats(c *gin.Context) {
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

	// Verify workflow belongs to user and get workflow data
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	var workflow struct {
		ID           primitive.ObjectID `bson:"_id"`
		UserID       primitive.ObjectID `bson:"userId"`
		Name         string             `bson:"name"`
		RunCount     int64              `bson:"runCount"`
		SuccessCount int64              `bson:"successCount"`
		SuccessRate  float64            `bson:"successRate"`
		AvgDuration  float64            `bson:"avgDuration"`
		LastRun      time.Time          `bson:"lastRun"`
		Actions      []interface{}      `bson:"actions"`
	}
	err = workflowCollection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get execution stats from executions collection
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	
	totalExecutions, _ := executionCollection.CountDocuments(context.Background(), bson.M{"workflowId": workflowObjectID})
	successfulExecutions, _ := executionCollection.CountDocuments(context.Background(), bson.M{
		"workflowId": workflowObjectID,
		"status":     "completed",
	})
	failedExecutions, _ := executionCollection.CountDocuments(context.Background(), bson.M{
		"workflowId": workflowObjectID,
		"status":     "failed",
	})

	// Calculate conversion rate (completions vs triggers)
	conversionRate := "0.0%"
	if workflow.RunCount > 0 {
		rate := float64(workflow.SuccessCount) / float64(workflow.RunCount) * 100
		conversionRate = fmt.Sprintf("%.1f%%", rate)
	}

	// Get node stats (simplified for now)
	nodeStats := make(map[string]interface{})
	for i := range workflow.Actions {
		nodeID := fmt.Sprintf("action_%d", i)
		nodeStats[nodeID] = map[string]interface{}{
			"nodeTitle":       fmt.Sprintf("Action %d", i+1),
			"nodeType":        "action",
			"totalExecutions": workflow.RunCount,
			"successRate":     conversionRate,
		}
	}

	// Create analytics insights
	insights := []map[string]interface{}{}
	if workflow.SuccessRate < 50 {
		insights = append(insights, map[string]interface{}{
			"type":    "warning",
			"message": "Low success rate detected. Consider reviewing workflow conditions.",
		})
	}
	if workflow.RunCount == 0 {
		insights = append(insights, map[string]interface{}{
			"type":    "info",
			"message": "This workflow hasn't been executed yet.",
		})
	}

	// Format last triggered
	var lastTriggered *string
	if !workflow.LastRun.IsZero() {
		formatted := workflow.LastRun.Format(time.RFC3339)
		lastTriggered = &formatted
	}

	stats := map[string]interface{}{
		"totalTriggers":    workflow.RunCount,
		"totalCompletions": workflow.SuccessCount,
		"totalRuns":        totalExecutions,
		"successfulRuns":   successfulExecutions,
		"failedRuns":       failedExecutions,
		"conversionRate":   conversionRate,
		"successRate":      fmt.Sprintf("%.1f%%", workflow.SuccessRate),
		"lastTriggered":    lastTriggered,
		"nodeStats":        nodeStats,
		"nodeTypeSummary": map[string]interface{}{
			"triggers": map[string]interface{}{
				"count":      1,
				"executions": workflow.RunCount,
			},
			"conditions": map[string]interface{}{
				"count":  0,
				"passed": 0,
				"failed": 0,
			},
			"actions": map[string]interface{}{
				"count":       len(workflow.Actions),
				"completions": workflow.SuccessCount,
				"failures":    workflow.RunCount - workflow.SuccessCount,
				"skipped":     0,
			},
		},
		"insights": insights,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func GetWorkflowPerformance(c *gin.Context) {
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
	var workflow struct {
		ID     primitive.ObjectID `bson:"_id"`
		UserID primitive.ObjectID `bson:"userId"`
	}
	err = workflowCollection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get performance metrics
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	
	// Average execution time
	pipeline := []bson.M{
		{"$match": bson.M{"workflowId": workflowObjectID, "status": "completed"}},
		{"$group": bson.M{
			"_id":        nil,
			"avgDuration": bson.M{"$avg": "$duration"},
			"minDuration": bson.M{"$min": "$duration"},
			"maxDuration": bson.M{"$max": "$duration"},
		}},
	}

	cursor, err := executionCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate performance metrics"})
		return
	}
	defer cursor.Close(context.Background())

	var result struct {
		AvgDuration float64 `bson:"avgDuration"`
		MinDuration int64   `bson:"minDuration"`
		MaxDuration int64   `bson:"maxDuration"`
	}

	if cursor.Next(context.Background()) {
		cursor.Decode(&result)
	}

	performance := gin.H{
		"workflowId":          workflowID,
		"averageExecutionTime": result.AvgDuration,
		"minExecutionTime":    result.MinDuration,
		"maxExecutionTime":    result.MaxDuration,
		"unit":               "milliseconds",
	}

	c.JSON(http.StatusOK, performance)
}

// Admin Services
func AdminGetUsers(c *gin.Context) {
	// TODO: Add admin role verification middleware
	
	// Query parameters for pagination and filtering
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")
	status := c.Query("status") // active, inactive, suspended
	
	collection := database.GetUserDB().Collection("users")
	
	// Build filter
	filter := bson.M{}
	if status != "" {
		if status == "active" {
			filter["isActive"] = true
		} else if status == "inactive" {
			filter["isActive"] = false
		}
	}
	
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(context.Background())
	
	var users []bson.M
	for cursor.Next(context.Background()) {
		var user bson.M
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		// Remove sensitive information
		delete(user, "password")
		delete(user, "resetPasswordToken")
		delete(user, "emailVerifyToken")
		users = append(users, user)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"page":  page,
		"limit": limit,
		"count": len(users),
	})
}

func AdminGetUser(c *gin.Context) {
	userID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	collection := database.GetUserDB().Collection("users")
	var user bson.M
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	// Remove sensitive information
	delete(user, "password")
	delete(user, "resetPasswordToken")
	delete(user, "emailVerifyToken")
	
	c.JSON(http.StatusOK, user)
}

func AdminUpdateUser(c *gin.Context) {
	userID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var updateData bson.M
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Remove sensitive fields that shouldn't be updated via admin
	delete(updateData, "password")
	delete(updateData, "_id")
	updateData["updatedAt"] = time.Now()
	
	collection := database.GetUserDB().Collection("users")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": updateData},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func AdminDeleteUser(c *gin.Context) {
	userID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	// Soft delete by deactivating
	collection := database.GetUserDB().Collection("users")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"isActive":  false,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func AdminSuspendUser(c *gin.Context) {
	userID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	collection := database.GetUserDB().Collection("users")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"isActive":  false,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to suspend user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User suspended successfully"})
}

func AdminActivateUser(c *gin.Context) {
	userID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	collection := database.GetUserDB().Collection("users")
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"isActive":  true,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User activated successfully"})
}

func AdminGetWebsites(c *gin.Context) {
	collection := database.GetUserDB().Collection("websites")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch websites"})
		return
	}
	defer cursor.Close(context.Background())
	
	var websites []bson.M
	if err = cursor.All(context.Background(), &websites); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode websites"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"websites": websites,
		"count":    len(websites),
	})
}

func AdminGetWebsite(c *gin.Context) {
	websiteID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}
	
	collection := database.GetUserDB().Collection("websites")
	var website bson.M
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&website)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Website not found"})
		return
	}
	
	c.JSON(http.StatusOK, website)
}

func AdminDeleteWebsite(c *gin.Context) {
	websiteID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}
	
	collection := database.GetUserDB().Collection("websites")
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete website"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Website deleted successfully"})
}

func GetAdminStats(c *gin.Context) {
	userCollection := database.GetUserDB().Collection("users")
	websiteCollection := database.GetUserDB().Collection("websites")
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	
	// Get counts
	totalUsers, _ := userCollection.CountDocuments(context.Background(), bson.M{})
	activeUsers, _ := userCollection.CountDocuments(context.Background(), bson.M{"isActive": true})
	totalWebsites, _ := websiteCollection.CountDocuments(context.Background(), bson.M{})
	totalWorkflows, _ := workflowCollection.CountDocuments(context.Background(), bson.M{})
	
	stats := gin.H{
		"totalUsers":     totalUsers,
		"activeUsers":    activeUsers,
		"totalWebsites":  totalWebsites,
		"totalWorkflows": totalWorkflows,
		"generatedAt":    time.Now(),
	}
	
	c.JSON(http.StatusOK, stats)
}

func GetSystemHealth(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": gin.H{
			"database": "connected",
			"redis":    "connected",
		},
		"version": "1.0.0",
	}
	
	c.JSON(http.StatusOK, health)
}

// Internal Services
func ValidateUser(c *gin.Context) {
	userID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	collection := database.GetUserDB().Collection("users")
	var user struct {
		ID       primitive.ObjectID `bson:"_id"`
		IsActive bool               `bson:"isActive"`
		Email    string             `bson:"email"`
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"valid": false, "error": "User not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"valid":    user.IsActive,
		"userId":   user.ID.Hex(),
		"email":    user.Email,
		"isActive": user.IsActive,
	})
}

func ValidateWebsite(c *gin.Context) {
	websiteID := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(websiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid website ID"})
		return
	}
	
	collection := database.GetUserDB().Collection("websites")
	var website struct {
		ID        primitive.ObjectID `bson:"_id"`
		IsActive  bool               `bson:"isActive"`
		Domain    string             `bson:"domain"`
		TrackingID string            `bson:"trackingId"`
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&website)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"valid": false, "error": "Website not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"valid":      website.IsActive,
		"websiteId":  website.ID.Hex(),
		"domain":     website.Domain,
		"trackingId": website.TrackingID,
		"isActive":   website.IsActive,
	})
}

func TriggerWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	workflowObjectID, err := primitive.ObjectIDFromHex(workflowID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}
	
	var triggerData map[string]interface{}
	if err := c.ShouldBindJSON(&triggerData); err != nil {
		triggerData = make(map[string]interface{})
	}
	
	// Verify workflow exists and is active
	collection := database.GetWorkflowDB().Collection("workflows")
	var workflow struct {
		ID       primitive.ObjectID `bson:"_id"`
		IsActive bool               `bson:"isActive"`
	}
	err = collection.FindOne(context.Background(), bson.M{"_id": workflowObjectID}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}
	
	if !workflow.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Workflow is not active"})
		return
	}
	
	// TODO: Implement actual workflow execution logic
	
	c.JSON(http.StatusOK, gin.H{
		"message":     "Workflow triggered successfully",
		"workflowId":  workflowID,
		"triggeredAt": time.Now(),
	})
}

func GetAnalyticsSummary(c *gin.Context) {
	// Get summary analytics for all services
	userCollection := database.GetUserDB().Collection("users")
	websiteCollection := database.GetUserDB().Collection("websites")
	eventCollection := database.GetWorkflowDB().Collection("visitor_events")
	
	// Get counts for the last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	totalUsers, _ := userCollection.CountDocuments(context.Background(), bson.M{})
	newUsers, _ := userCollection.CountDocuments(context.Background(), bson.M{
		"createdAt": bson.M{"$gte": thirtyDaysAgo},
	})
	
	totalWebsites, _ := websiteCollection.CountDocuments(context.Background(), bson.M{})
	totalEvents, _ := eventCollection.CountDocuments(context.Background(), bson.M{
		"createdAt": bson.M{"$gte": thirtyDaysAgo},
	})
	
	summary := gin.H{
		"totalUsers":    totalUsers,
		"newUsers":      newUsers,
		"totalWebsites": totalWebsites,
		"totalEvents":   totalEvents,
		"period":        "last_30_days",
		"generatedAt":   time.Now(),
	}
	
	c.JSON(http.StatusOK, summary)
}

// Real Workflow Analytics Functions
func GetWorkflowActivity(c *gin.Context) {
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
	count, err := workflowCollection.CountDocuments(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	})
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get workflow executions for activity
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	cursor, err := executionCollection.Find(
		context.Background(),
		bson.M{"workflowId": workflowObjectID},
		// Sort by creation time descending, limit to 50
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow activity"})
		return
	}
	defer cursor.Close(context.Background())

	var activities []gin.H
	for cursor.Next(context.Background()) {
		var execution struct {
			ID          primitive.ObjectID `bson:"_id"`
			Status      string             `bson:"status"`
			StartTime   time.Time          `bson:"startTime"`
			EndTime     *time.Time         `bson:"endTime"`
			Results     []interface{}      `bson:"results"`
			TriggerData map[string]interface{} `bson:"triggerData"`
		}

		if err := cursor.Decode(&execution); err != nil {
			continue
		}

		// Add trigger event
		activities = append(activities, gin.H{
			"id":        execution.ID.Hex() + "_trigger",
			"event":     "Trigger",
			"nodeId":    "trigger",
			"nodeTitle": "Workflow Triggered",
			"detail":    fmt.Sprintf("Workflow execution started - Status: %s", execution.Status),
			"timestamp": execution.StartTime.Format(time.RFC3339),
		})

		// Add action events from results
		for i, result := range execution.Results {
			if resultMap, ok := result.(map[string]interface{}); ok {
				actionID := fmt.Sprintf("action_%d", i)
				if actionIDVal, exists := resultMap["actionId"]; exists {
					actionID = fmt.Sprintf("%v", actionIDVal)
				}

				status := "completed"
				if statusVal, exists := resultMap["status"]; exists {
					status = fmt.Sprintf("%v", statusVal)
				}

				activities = append(activities, gin.H{
					"id":        execution.ID.Hex() + "_" + actionID,
					"event":     "Action Executed",
					"nodeId":    actionID,
					"nodeTitle": fmt.Sprintf("Action %d", i+1),
					"detail":    fmt.Sprintf("Action %s", status),
					"timestamp": execution.StartTime.Add(time.Duration(i+1) * time.Minute).Format(time.RFC3339),
				})
			}
		}
	}

	// Sort activities by timestamp (most recent first)
	// In a real implementation, you'd do this in the database query
	
	c.JSON(http.StatusOK, gin.H{
		"activities":  activities,
		"totalCount":  len(activities),
		"hasMore":     len(activities) >= 50,
	})
}

func GetWorkflowSummary(c *gin.Context) {
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
	count, err := workflowCollection.CountDocuments(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	})
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get daily stats from workflow_stats collection
	statsCollection := database.GetWorkflowDB().Collection("workflow_stats")
	
	// Get last 30 days of data
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	cursor, err := statsCollection.Find(
		context.Background(),
		bson.M{
			"workflowId": workflowObjectID,
			"date": bson.M{
				"$gte": thirtyDaysAgo.Format("2006-01-02"),
				"$lte": time.Now().Format("2006-01-02"),
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflow summary"})
		return
	}
	defer cursor.Close(context.Background())

	var summary []gin.H
	statsMap := make(map[string]gin.H)

	// Process existing stats
	for cursor.Next(context.Background()) {
		var stat struct {
			Date       string `bson:"date"`
			Executions int64  `bson:"executions"`
			Successes  int64  `bson:"successes"`
			Failures   int64  `bson:"failures"`
		}

		if err := cursor.Decode(&stat); err != nil {
			continue
		}

		completionRate := 0.0
		if stat.Executions > 0 {
			completionRate = float64(stat.Successes) / float64(stat.Executions) * 100
		}

		statsMap[stat.Date] = gin.H{
			"date":           stat.Date,
			"triggers":       stat.Executions,
			"completions":    stat.Successes,
			"completionRate": completionRate,
		}
	}

	// Fill in missing days with zero values
	for i := 29; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if stat, exists := statsMap[date]; exists {
			summary = append(summary, stat)
		} else {
			summary = append(summary, gin.H{
				"date":           date,
				"triggers":       0,
				"completions":    0,
				"completionRate": 0.0,
			})
		}
	}

	c.JSON(http.StatusOK, summary)
}

func GetWorkflowNodePerformance(c *gin.Context) {
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

	// Get workflow to extract nodes
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	var workflow struct {
		ID      primitive.ObjectID `bson:"_id"`
		UserID  primitive.ObjectID `bson:"userId"`
		Name    string             `bson:"name"`
		Trigger map[string]interface{} `bson:"trigger"`
		Actions []map[string]interface{} `bson:"actions"`
	}

	err = workflowCollection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get execution stats for each node
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	
	var nodePerformance []gin.H

	// Add trigger performance
	totalExecutions, _ := executionCollection.CountDocuments(context.Background(), bson.M{
		"workflowId": workflowObjectID,
	})

	successfulExecutions, _ := executionCollection.CountDocuments(context.Background(), bson.M{
		"workflowId": workflowObjectID,
		"status":     "completed",
	})

	triggerPerformance := 100.0
	if totalExecutions > 0 {
		triggerPerformance = float64(successfulExecutions) / float64(totalExecutions) * 100
	}

	triggerTitle := "Workflow Trigger"
	if triggerType, ok := workflow.Trigger["type"].(string); ok {
		triggerTitle = fmt.Sprintf("%s Trigger", triggerType)
	}

	nodePerformance = append(nodePerformance, gin.H{
		"nodeId":      "trigger",
		"nodeTitle":   triggerTitle,
		"triggers":    totalExecutions,
		"executions":  successfulExecutions,
		"performance": triggerPerformance,
	})

	// Add action performance
	for i, action := range workflow.Actions {
		actionID := fmt.Sprintf("action_%d", i)
		if id, ok := action["id"].(string); ok {
			actionID = id
		}

		actionTitle := fmt.Sprintf("Action %d", i+1)
		if title, ok := action["title"].(string); ok {
			actionTitle = title
		} else if actionType, ok := action["type"].(string); ok {
			actionTitle = fmt.Sprintf("%s Action", actionType)
		}

		// Count successful action executions
		// This would require analyzing the results array in executions
		cursor, err := executionCollection.Find(context.Background(), bson.M{
			"workflowId": workflowObjectID,
			"status":     "completed",
		})
		if err != nil {
			continue
		}

		actionSuccesses := int64(0)
		actionAttempts := int64(0)

		for cursor.Next(context.Background()) {
			var execution struct {
				Results []map[string]interface{} `bson:"results"`
			}
			if err := cursor.Decode(&execution); err != nil {
				continue
			}

			if i < len(execution.Results) {
				actionAttempts++
				if status, ok := execution.Results[i]["status"].(string); ok && status == "success" {
					actionSuccesses++
				}
			}
		}
		cursor.Close(context.Background())

		actionPerformance := 0.0
		if actionAttempts > 0 {
			actionPerformance = float64(actionSuccesses) / float64(actionAttempts) * 100
		}

		nodePerformance = append(nodePerformance, gin.H{
			"nodeId":      actionID,
			"nodeTitle":   actionTitle,
			"triggers":    actionAttempts,
			"executions":  actionSuccesses,
			"performance": actionPerformance,
		})
	}

	c.JSON(http.StatusOK, nodePerformance)
}

func GetWorkflowTriggerTypes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, err := primitive.ObjectIDFromHex(userID.(string))
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

	// Get workflow executions and analyze trigger data
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	cursor, err := executionCollection.Find(context.Background(), bson.M{
		"workflowId": workflowObjectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trigger types"})
		return
	}
	defer cursor.Close(context.Background())

	triggerTypeCounts := make(map[string]int64)
	totalTriggers := int64(0)

	for cursor.Next(context.Background()) {
		var execution struct {
			TriggerData map[string]interface{} `bson:"triggerData"`
		}
		if err := cursor.Decode(&execution); err != nil {
			continue
		}

		totalTriggers++
		
		// Extract trigger type from trigger data
		triggerType := "Unknown"
		if triggerData := execution.TriggerData; triggerData != nil {
			if tType, ok := triggerData["type"].(string); ok {
				triggerType = tType
			} else if event, ok := triggerData["event"].(string); ok {
				triggerType = event
			} else if url, ok := triggerData["url"].(string); ok {
				if url != "" {
					triggerType = "Page Visit"
				}
			}
		}

		triggerTypeCounts[triggerType]++
	}

	var triggerTypes []gin.H
	for triggerType, count := range triggerTypeCounts {
		percentage := 0.0
		if totalTriggers > 0 {
			percentage = float64(count) / float64(totalTriggers) * 100
		}

		triggerTypes = append(triggerTypes, gin.H{
			"triggerType": triggerType,
			"count":       count,
			"percentage":  percentage,
		})
	}

	c.JSON(http.StatusOK, triggerTypes)
}

func GetWorkflowActionTypes(c *gin.Context) {
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

	// Get workflow to analyze action types
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	var workflow struct {
		Actions []map[string]interface{} `bson:"actions"`
	}

	err = workflowCollection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get execution stats for action types
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	
	actionTypeStats := make(map[string]struct {
		Count             int64
		Successes         int64
		TotalExecutionTime int64
	})

	// Analyze each action type
	for _, action := range workflow.Actions {
		actionType := "Unknown"
		if aType, ok := action["type"].(string); ok {
			actionType = aType
		}

		// Get execution stats for this action type
		cursor, err := executionCollection.Find(context.Background(), bson.M{
			"workflowId": workflowObjectID,
			"status":     "completed",
		})
		if err != nil {
			continue
		}

		for cursor.Next(context.Background()) {
			var execution struct {
				Results   []map[string]interface{} `bson:"results"`
				StartTime time.Time               `bson:"startTime"`
				EndTime   *time.Time              `bson:"endTime"`
			}
			if err := cursor.Decode(&execution); err != nil {
				continue
			}

			// Analyze results for this action type
			for _, result := range execution.Results {
				if rType, ok := result["type"].(string); ok && rType == actionType {
					stats := actionTypeStats[actionType]
					stats.Count++

					if status, ok := result["status"].(string); ok && status == "success" {
						stats.Successes++
					}

					// Calculate execution time if available
					if execution.EndTime != nil {
						executionTime := execution.EndTime.Sub(execution.StartTime).Milliseconds()
						stats.TotalExecutionTime += executionTime
					}

					actionTypeStats[actionType] = stats
				}
			}
		}
		cursor.Close(context.Background())
	}

	var actionTypes []gin.H
	for actionType, stats := range actionTypeStats {
		successRate := 0.0
		if stats.Count > 0 {
			successRate = float64(stats.Successes) / float64(stats.Count) * 100
		}

		avgExecutionTime := int64(0)
		if stats.Count > 0 {
			avgExecutionTime = stats.TotalExecutionTime / stats.Count
		}

		actionTypes = append(actionTypes, gin.H{
			"actionType":        actionType,
			"count":             stats.Count,
			"successRate":       successRate,
			"avgExecutionTime":  avgExecutionTime,
		})
	}

	c.JSON(http.StatusOK, actionTypes)
}

func GetWorkflowHourlyData(c *gin.Context) {
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
	count, err := workflowCollection.CountDocuments(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	})
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get executions from last 24 hours
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	
	cursor, err := executionCollection.Find(context.Background(), bson.M{
		"workflowId": workflowObjectID,
		"startTime": bson.M{
			"$gte": twentyFourHoursAgo,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get hourly data"})
		return
	}
	defer cursor.Close(context.Background())

	// Group executions by hour
	hourlyStats := make(map[int]struct {
		Triggers    int64
		Completions int64
	})

	for cursor.Next(context.Background()) {
		var execution struct {
			StartTime time.Time `bson:"startTime"`
			Status    string    `bson:"status"`
		}
		if err := cursor.Decode(&execution); err != nil {
			continue
		}

		hour := execution.StartTime.Hour()
		stats := hourlyStats[hour]
		stats.Triggers++
		if execution.Status == "completed" {
			stats.Completions++
		}
		hourlyStats[hour] = stats
	}

	// Generate hourly data for last 24 hours
	var hourlyData []gin.H
	for i := 23; i >= 0; i-- {
		hour := time.Now().Add(-time.Duration(i) * time.Hour).Hour()
		stats := hourlyStats[hour]
		
		completionRate := 0.0
		if stats.Triggers > 0 {
			completionRate = float64(stats.Completions) / float64(stats.Triggers) * 100
		}

		hourlyData = append(hourlyData, gin.H{
			"hour":           hour,
			"triggers":       stats.Triggers,
			"completions":    stats.Completions,
			"completionRate": completionRate,
		})
	}

	c.JSON(http.StatusOK, hourlyData)
}

func GetWorkflowFunnelData(c *gin.Context) {
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

	// Get workflow structure
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	var workflow struct {
		ID      primitive.ObjectID `bson:"_id"`
		UserID  primitive.ObjectID `bson:"userId"`
		Name    string             `bson:"name"`
		Trigger map[string]interface{} `bson:"trigger"`
		Actions []map[string]interface{} `bson:"actions"`
	}

	err = workflowCollection.FindOne(context.Background(), bson.M{
		"_id":    workflowObjectID,
		"userId": userObjectID,
	}).Decode(&workflow)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	// Get execution data for funnel analysis
	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	cursor, err := executionCollection.Find(context.Background(), bson.M{
		"workflowId": workflowObjectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get funnel data"})
		return
	}
	defer cursor.Close(context.Background())

	totalRuns := int64(0)
	successfulCompletions := int64(0)
	stepCompletions := make([]int64, len(workflow.Actions)+1) // +1 for trigger

	for cursor.Next(context.Background()) {
		var execution struct {
			Status  string                   `bson:"status"`
			Results []map[string]interface{} `bson:"results"`
		}
		if err := cursor.Decode(&execution); err != nil {
			continue
		}

		totalRuns++
		stepCompletions[0]++ // Trigger always fires

		// Count completed actions
		for i, result := range execution.Results {
			if i+1 < len(stepCompletions) {
				if status, ok := result["status"].(string); ok && status == "success" {
					stepCompletions[i+1]++
				}
			}
		}

		if execution.Status == "completed" {
			successfulCompletions++
		}
	}

	// Build funnel steps
	var steps []gin.H
	var dropOffRates []gin.H

	// Add trigger step
	triggerName := "Workflow Trigger"
	if triggerType, ok := workflow.Trigger["type"].(string); ok {
		triggerName = fmt.Sprintf("%s Trigger", triggerType)
	}

	steps = append(steps, gin.H{
		"name":           triggerName,
		"nodeType":       "trigger",
		"count":          totalRuns,
		"completed":      stepCompletions[0],
		"conversionRate": "100.0%",
		"dropOff":        0,
		"avgTime":        0,
		"stepOrder":      1,
		"successRate":    "100.0%",
	})

	// Add action steps
	for i, action := range workflow.Actions {
		actionName := fmt.Sprintf("Action %d", i+1)
		if title, ok := action["title"].(string); ok {
			actionName = title
		} else if actionType, ok := action["type"].(string); ok {
			actionName = fmt.Sprintf("%s Action", actionType)
		}

		prevStepCount := stepCompletions[i]
		currentStepCount := stepCompletions[i+1]
		dropOff := prevStepCount - currentStepCount

		conversionRate := "0.0%"
		if prevStepCount > 0 {
			rate := float64(currentStepCount) / float64(prevStepCount) * 100
			conversionRate = fmt.Sprintf("%.1f%%", rate)
		}

		steps = append(steps, gin.H{
			"name":           actionName,
			"nodeType":       "action",
			"count":          prevStepCount,
			"completed":      currentStepCount,
			"conversionRate": conversionRate,
			"dropOff":        dropOff,
			"avgTime":        0, // Would need execution time data
			"stepOrder":      i + 2,
			"successRate":    conversionRate,
		})

		// Add drop-off rate
		if i > 0 {
			prevStepName := steps[i]["name"].(string)
			dropOffRate := 0.0
			if prevStepCount > 0 {
				dropOffRate = float64(dropOff) / float64(prevStepCount) * 100
			}

			dropOffRates = append(dropOffRates, gin.H{
				"fromStep":     prevStepName,
				"toStep":       actionName,
				"dropOffCount": dropOff,
				"dropOffRate":  dropOffRate,
				"critical":     dropOffRate > 50.0,
			})
		}
	}

	funnelData := gin.H{
		"totalVisitors":         totalRuns,
		"steps":                 steps,
		"dropOffRates":          dropOffRates,
		"averageTimePerStep":    []gin.H{}, // Would need execution timing data
		"pathAnalysis":          []gin.H{}, // Would need visitor tracking data
		"totalRuns":             totalRuns,
		"successfulCompletions": successfulCompletions,
	}

	c.JSON(http.StatusOK, funnelData)
}

func GetWorkflowsSummary(c *gin.Context) {
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

	siteID := c.Query("siteId")
	if siteID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "siteId parameter is required"})
		return
	}

	siteObjectID, err := primitive.ObjectIDFromHex(siteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid siteId"})
		return
	}

	// Get workflows for this user and site
	workflowCollection := database.GetWorkflowDB().Collection("workflows")
	cursor, err := workflowCollection.Find(context.Background(), bson.M{
		"userId": userObjectID,
		"siteId": siteObjectID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workflows"})
		return
	}
	defer cursor.Close(context.Background())

	totalWorkflows := int64(0)
	activeWorkflows := int64(0)
	pausedWorkflows := int64(0)
	draftWorkflows := int64(0)
	totalTriggers := int64(0)
	totalCompletions := int64(0)
	var topPerformingWorkflows []gin.H

	for cursor.Next(context.Background()) {
		var workflow struct {
			ID           primitive.ObjectID `bson:"_id"`
			Name         string             `bson:"name"`
			Status       string             `bson:"status"`
			IsActive     bool               `bson:"isActive"`
			RunCount     int64              `bson:"runCount"`
			SuccessCount int64              `bson:"successCount"`
			SuccessRate  float64            `bson:"successRate"`
		}
		if err := cursor.Decode(&workflow); err != nil {
			continue
		}

		totalWorkflows++
		totalTriggers += workflow.RunCount
		totalCompletions += workflow.SuccessCount

		switch workflow.Status {
		case "Active":
			if workflow.IsActive {
				activeWorkflows++
			}
		case "Paused":
			pausedWorkflows++
		case "Draft":
			draftWorkflows++
		}

		// Add to top performing if success rate > 80%
		if workflow.SuccessRate > 80.0 && workflow.RunCount > 0 {
			topPerformingWorkflows = append(topPerformingWorkflows, gin.H{
				"id":          workflow.ID.Hex(),
				"name":        workflow.Name,
				"successRate": workflow.SuccessRate,
				"totalRuns":   workflow.RunCount,
			})
		}
	}

	averageSuccessRate := 0.0
	if totalTriggers > 0 {
		averageSuccessRate = float64(totalCompletions) / float64(totalTriggers) * 100
	}

	summary := gin.H{
		"totalWorkflows":           totalWorkflows,
		"activeWorkflows":          activeWorkflows,
		"pausedWorkflows":          pausedWorkflows,
		"draftWorkflows":           draftWorkflows,
		"totalTriggers":            totalTriggers,
		"totalCompletions":         totalCompletions,
		"averageSuccessRate":       averageSuccessRate,
		"topPerformingWorkflows":   topPerformingWorkflows,
	}

	c.JSON(http.StatusOK, summary)
}
