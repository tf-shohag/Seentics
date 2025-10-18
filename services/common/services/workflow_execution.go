package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seentics/seentics/services/common/config"
	"github.com/seentics/seentics/services/common/database"
	"github.com/seentics/seentics/services/common/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkflowExecutionEngine struct {
	config *config.Config
}

type ExecutionContext struct {
	WorkflowID       primitive.ObjectID
	UserID           primitive.ObjectID
	WebsiteID        primitive.ObjectID
	VisitorID        string
	SessionID        string
	TriggerData      map[string]interface{}
	IdentifiedUser   map[string]interface{}
	LocalStorageData map[string]interface{}
}

type ActionResult struct {
	ActionID  string                 `json:"actionId"`
	Status    string                 `json:"status"` // success, failed, skipped
	Output    map[string]interface{} `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Duration  int64                  `json:"duration"` // milliseconds
	StartTime time.Time              `json:"startTime"`
	EndTime   time.Time              `json:"endTime"`
}

// Real workflow execution implementation
func ExecuteWorkflowReal(c *gin.Context) {
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

	var triggerData map[string]interface{}
	if err := c.ShouldBindJSON(&triggerData); err != nil {
		triggerData = make(map[string]interface{})
	}

	// Get workflow
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

	if !workflow.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Workflow is not active"})
		return
	}

	// Create execution context
	ctx := &ExecutionContext{
		WorkflowID:       workflowObjectID,
		UserID:           userObjectID,
		WebsiteID:        workflow.WebsiteID,
		TriggerData:      triggerData,
		IdentifiedUser:   make(map[string]interface{}),
		LocalStorageData: make(map[string]interface{}),
	}

	// Extract visitor and session info from trigger data
	if visitorID, ok := triggerData["visitorId"].(string); ok {
		ctx.VisitorID = visitorID
	}
	if sessionID, ok := triggerData["sessionId"].(string); ok {
		ctx.SessionID = sessionID
	}
	if identifiedUser, ok := triggerData["identifiedUser"].(map[string]interface{}); ok {
		ctx.IdentifiedUser = identifiedUser
	}
	if localStorage, ok := triggerData["localStorageData"].(map[string]interface{}); ok {
		ctx.LocalStorageData = localStorage
	}

	// Create execution record
	execution := models.WorkflowExecution{
		ID:          primitive.NewObjectID(),
		WorkflowID:  workflowObjectID,
		Status:      "running",
		StartTime:   time.Now(),
		TriggerData: triggerData,
		Results:     []models.ActionResult{},
		CreatedAt:   time.Now(),
	}

	executionCollection := database.GetWorkflowDB().Collection("workflow_executions")
	_, err = executionCollection.InsertOne(context.Background(), execution)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create execution"})
		return
	}

	// Execute workflow
	engine := &WorkflowExecutionEngine{config: config.Load()}
	results, err := engine.ExecuteWorkflow(ctx, &workflow)

	// Update execution record
	endTime := time.Now()
	duration := endTime.Sub(execution.StartTime).Milliseconds()
	status := "completed"
	errorMsg := ""

	if err != nil {
		status = "failed"
		errorMsg = err.Error()
		logrus.Error("Workflow execution failed:", err)
	}

	// Convert ActionResult to models.ActionResult
	modelResults := make([]models.ActionResult, len(results))
	for i, result := range results {
		modelResults[i] = models.ActionResult{
			ActionID: result.ActionID,
			Status:   result.Status,
			Output:   result.Output,
			Error:    result.Error,
			Duration: result.Duration,
		}
	}

	_, err = executionCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": execution.ID},
		bson.M{"$set": bson.M{
			"status":   status,
			"endTime":  endTime,
			"duration": duration,
			"results":  modelResults,
			"error":    errorMsg,
		}},
	)
	if err != nil {
		logrus.Error("Failed to update execution record:", err)
	}

	// Update workflow statistics
	go updateWorkflowStats(workflowObjectID, status == "completed", duration)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Workflow executed",
		"executionId": execution.ID.Hex(),
		"status":      status,
		"duration":    duration,
		"results":     results,
		"error":       errorMsg,
	})
}

func (e *WorkflowExecutionEngine) ExecuteWorkflow(ctx *ExecutionContext, workflow *models.Workflow) ([]ActionResult, error) {
	var results []ActionResult

	logrus.Info("Starting workflow execution", logrus.Fields{
		"workflowId": ctx.WorkflowID.Hex(),
		"userId":     ctx.UserID.Hex(),
		"visitorId":  ctx.VisitorID,
	})

	// Check conditions first
	if !e.evaluateConditions(ctx, workflow.Conditions) {
		logrus.Info("Workflow conditions not met, skipping execution")
		return results, nil
	}

	// Execute actions in order
	for _, action := range workflow.Actions {
		if !action.IsActive {
			continue
		}

		result := e.executeAction(ctx, action)
		results = append(results, result)

		// Stop execution if action failed and it's critical
		if result.Status == "failed" {
			logrus.Error("Action failed, stopping workflow execution", logrus.Fields{
				"actionId": action.ID,
				"error":    result.Error,
			})
			break
		}
	}

	return results, nil
}

func (e *WorkflowExecutionEngine) evaluateConditions(ctx *ExecutionContext, conditions []models.WorkflowCondition) bool {
	if len(conditions) == 0 {
		return true // No conditions means always execute
	}

	// Simple condition evaluation - can be enhanced
	for _, condition := range conditions {
		if !e.evaluateCondition(ctx, condition) {
			return false
		}
	}

	return true
}

func (e *WorkflowExecutionEngine) evaluateCondition(ctx *ExecutionContext, condition models.WorkflowCondition) bool {
	// Get field value from context
	var fieldValue interface{}

	switch condition.Field {
	case "visitorId":
		fieldValue = ctx.VisitorID
	case "sessionId":
		fieldValue = ctx.SessionID
	case "user.email":
		if email, ok := ctx.IdentifiedUser["email"]; ok {
			fieldValue = email
		}
	case "user.name":
		if name, ok := ctx.IdentifiedUser["name"]; ok {
			fieldValue = name
		}
	default:
		// Check in trigger data
		if value, ok := ctx.TriggerData[condition.Field]; ok {
			fieldValue = value
		}
	}

	// Evaluate condition based on operator
	switch condition.Operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", condition.Value)
	case "not_equals":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", condition.Value)
	case "contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		valueStr := fmt.Sprintf("%v", condition.Value)
		return strings.Contains(fieldStr, valueStr)
	case "not_contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		valueStr := fmt.Sprintf("%v", condition.Value)
		return !strings.Contains(fieldStr, valueStr)
	case "exists":
		return fieldValue != nil && fieldValue != ""
	case "not_exists":
		return fieldValue == nil || fieldValue == ""
	default:
		logrus.Warn("Unknown condition operator:", condition.Operator)
		return false
	}
}

func (e *WorkflowExecutionEngine) executeAction(ctx *ExecutionContext, action models.WorkflowAction) ActionResult {
	startTime := time.Now()
	result := ActionResult{
		ActionID:  action.ID,
		Status:    "success",
		Output:    make(map[string]interface{}),
		StartTime: startTime,
	}

	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime).Milliseconds()
	}()

	actionType, ok := action.Config["type"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "Action type not specified"
		return result
	}

	logrus.Info("Executing action", logrus.Fields{
		"actionId":   action.ID,
		"actionType": actionType,
	})

	switch actionType {
	case "webhook":
		return e.executeWebhookAction(ctx, action)
	case "email":
		return e.executeEmailAction(ctx, action)
	case "track_event":
		return e.executeTrackEventAction(ctx, action)
	case "delay":
		return e.executeDelayAction(ctx, action)
	case "conditional":
		return e.executeConditionalAction(ctx, action)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("Unknown action type: %s", actionType)
		return result
	}
}

func (e *WorkflowExecutionEngine) executeWebhookAction(ctx *ExecutionContext, action models.WorkflowAction) ActionResult {
	result := ActionResult{
		ActionID:  action.ID,
		Status:    "success",
		Output:    make(map[string]interface{}),
		StartTime: time.Now(),
	}

	webhookURL, ok := action.Config["url"].(string)
	if !ok {
		result.Status = "failed"
		result.Error = "Webhook URL not specified"
		return result
	}

	method := "POST"
	if m, ok := action.Config["method"].(string); ok {
		method = m
	}

	// Prepare payload
	payload := map[string]interface{}{
		"workflowId":       ctx.WorkflowID.Hex(),
		"visitorId":        ctx.VisitorID,
		"sessionId":        ctx.SessionID,
		"triggerData":      ctx.TriggerData,
		"identifiedUser":   ctx.IdentifiedUser,
		"localStorageData": ctx.LocalStorageData,
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	// Add custom payload if specified
	if customPayload, ok := action.Config["payload"].(map[string]interface{}); ok {
		for k, v := range customPayload {
			payload[k] = e.substituteVariables(fmt.Sprintf("%v", v), ctx)
		}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		result.Status = "failed"
		result.Error = "Failed to marshal payload"
		return result
	}

	// Create request
	req, err := http.NewRequest(method, webhookURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		result.Status = "failed"
		result.Error = "Failed to create request"
		return result
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Seentics-Workflow/1.0")

	// Add custom headers
	if headers, ok := action.Config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			req.Header.Set(k, e.substituteVariables(fmt.Sprintf("%v", v), ctx))
		}
	}

	// Add HMAC signature if secret is configured
	if e.config.WebhookSecret != "" {
		signature := e.generateHMACSignature(payloadJSON, e.config.WebhookSecret)
		req.Header.Set("X-Seentics-Signature", signature)
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.Output["statusCode"] = resp.StatusCode
	result.Output["url"] = webhookURL

	if resp.StatusCode >= 400 {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Webhook returned status %d", resp.StatusCode)
	}

	return result
}

func (e *WorkflowExecutionEngine) executeEmailAction(ctx *ExecutionContext, action models.WorkflowAction) ActionResult {
	result := ActionResult{
		ActionID:  action.ID,
		Status:    "success",
		Output:    make(map[string]interface{}),
		StartTime: time.Now(),
	}

	// For now, just log the email action
	// In a real implementation, you'd integrate with an email service
	to, _ := action.Config["to"].(string)
	subject, _ := action.Config["subject"].(string)
	body, _ := action.Config["body"].(string)

	// Substitute variables
	to = e.substituteVariables(to, ctx)
	subject = e.substituteVariables(subject, ctx)
	body = e.substituteVariables(body, ctx)

	logrus.Info("Email action executed", logrus.Fields{
		"to":      to,
		"subject": subject,
	})

	result.Output["to"] = to
	result.Output["subject"] = subject
	result.Output["message"] = "Email sent successfully (simulated)"

	return result
}

func (e *WorkflowExecutionEngine) executeTrackEventAction(ctx *ExecutionContext, action models.WorkflowAction) ActionResult {
	result := ActionResult{
		ActionID:  action.ID,
		Status:    "success",
		Output:    make(map[string]interface{}),
		StartTime: time.Now(),
	}

	eventName, _ := action.Config["eventName"].(string)
	eventData, _ := action.Config["eventData"].(map[string]interface{})

	// Create visitor event
	event := models.VisitorEvent{
		ID:        primitive.NewObjectID(),
		WebsiteID: ctx.WebsiteID,
		VisitorID: ctx.VisitorID,
		SessionID: ctx.SessionID,
		EventType: eventName,
		EventData: eventData,
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}

	// Store event
	collection := database.GetWorkflowDB().Collection("visitor_events")
	_, err := collection.InsertOne(context.Background(), event)
	if err != nil {
		result.Status = "failed"
		result.Error = "Failed to track event"
		return result
	}

	result.Output["eventId"] = event.ID.Hex()
	result.Output["eventName"] = eventName

	return result
}

func (e *WorkflowExecutionEngine) executeDelayAction(ctx *ExecutionContext, action models.WorkflowAction) ActionResult {
	result := ActionResult{
		ActionID:  action.ID,
		Status:    "success",
		Output:    make(map[string]interface{}),
		StartTime: time.Now(),
	}

	delaySeconds, ok := action.Config["delaySeconds"].(float64)
	if !ok {
		delaySeconds = 1.0
	}

	time.Sleep(time.Duration(delaySeconds) * time.Second)

	result.Output["delaySeconds"] = delaySeconds

	return result
}

func (e *WorkflowExecutionEngine) executeConditionalAction(ctx *ExecutionContext, action models.WorkflowAction) ActionResult {
	result := ActionResult{
		ActionID:  action.ID,
		Status:    "success",
		Output:    make(map[string]interface{}),
		StartTime: time.Now(),
	}

	// Get condition configuration
	conditions, ok := action.Config["conditions"].([]interface{})
	if !ok {
		result.Status = "failed"
		result.Error = "No conditions specified for conditional action"
		return result
	}

	// Get the logic operator (AND/OR)
	logicOperator := "AND"
	if op, ok := action.Config["logicOperator"].(string); ok {
		logicOperator = strings.ToUpper(op)
	}

	// Evaluate all conditions
	conditionResults := make([]bool, len(conditions))
	conditionDetails := make([]map[string]interface{}, len(conditions))

	for i, conditionInterface := range conditions {
		conditionMap, ok := conditionInterface.(map[string]interface{})
		if !ok {
			result.Status = "failed"
			result.Error = fmt.Sprintf("Invalid condition format at index %d", i)
			return result
		}

		field, _ := conditionMap["field"].(string)
		operator, _ := conditionMap["operator"].(string)
		value := conditionMap["value"]

		conditionResult := e.evaluateConditionValue(ctx, field, operator, value)
		conditionResults[i] = conditionResult

		conditionDetails[i] = map[string]interface{}{
			"field":    field,
			"operator": operator,
			"value":    value,
			"result":   conditionResult,
		}
	}

	// Apply logic operator
	var finalResult bool
	if logicOperator == "OR" {
		finalResult = false
		for _, condResult := range conditionResults {
			if condResult {
				finalResult = true
				break
			}
		}
	} else { // AND (default)
		finalResult = true
		for _, condResult := range conditionResults {
			if !condResult {
				finalResult = false
				break
			}
		}
	}

	// Execute actions based on condition result
	if finalResult {
		// Execute success actions
		if successActions, ok := action.Config["successActions"].([]interface{}); ok {
			result.Output["executedBranch"] = "success"
			result.Output["successActionsCount"] = len(successActions)
			// In a full implementation, you would execute these actions
		}
	} else {
		// Execute failure actions
		if failureActions, ok := action.Config["failureActions"].([]interface{}); ok {
			result.Output["executedBranch"] = "failure"
			result.Output["failureActionsCount"] = len(failureActions)
			// In a full implementation, you would execute these actions
		}
	}

	result.Output["conditionResult"] = finalResult
	result.Output["logicOperator"] = logicOperator
	result.Output["conditionDetails"] = conditionDetails
	result.Output["conditionsEvaluated"] = len(conditions)

	return result
}

func (e *WorkflowExecutionEngine) evaluateConditionValue(ctx *ExecutionContext, field, operator string, value interface{}) bool {
	// Get field value from context
	var fieldValue interface{}

	switch field {
	case "visitorId":
		fieldValue = ctx.VisitorID
	case "sessionId":
		fieldValue = ctx.SessionID
	case "user.email":
		if email, ok := ctx.IdentifiedUser["email"]; ok {
			fieldValue = email
		}
	case "user.name":
		if name, ok := ctx.IdentifiedUser["name"]; ok {
			fieldValue = name
		}
	case "user.id":
		if id, ok := ctx.IdentifiedUser["id"]; ok {
			fieldValue = id
		}
	default:
		// Check in trigger data
		if val, ok := ctx.TriggerData[field]; ok {
			fieldValue = val
		} else if strings.HasPrefix(field, "localStorage.") {
			key := strings.TrimPrefix(field, "localStorage.")
			if val, ok := ctx.LocalStorageData[key]; ok {
				fieldValue = val
			}
		}
	}

	// Convert values to strings for comparison
	fieldStr := fmt.Sprintf("%v", fieldValue)
	valueStr := fmt.Sprintf("%v", value)

	// Evaluate condition based on operator
	switch operator {
	case "equals", "==":
		return fieldStr == valueStr
	case "not_equals", "!=":
		return fieldStr != valueStr
	case "contains":
		return strings.Contains(fieldStr, valueStr)
	case "not_contains":
		return !strings.Contains(fieldStr, valueStr)
	case "starts_with":
		return strings.HasPrefix(fieldStr, valueStr)
	case "ends_with":
		return strings.HasSuffix(fieldStr, valueStr)
	case "exists":
		return fieldValue != nil && fieldStr != ""
	case "not_exists":
		return fieldValue == nil || fieldStr == ""
	case "greater_than", ">":
		if fieldNum, err := parseNumber(fieldValue); err == nil {
			if valueNum, err := parseNumber(value); err == nil {
				return fieldNum > valueNum
			}
		}
		return false
	case "less_than", "<":
		if fieldNum, err := parseNumber(fieldValue); err == nil {
			if valueNum, err := parseNumber(value); err == nil {
				return fieldNum < valueNum
			}
		}
		return false
	case "greater_than_or_equal", ">=":
		if fieldNum, err := parseNumber(fieldValue); err == nil {
			if valueNum, err := parseNumber(value); err == nil {
				return fieldNum >= valueNum
			}
		}
		return false
	case "less_than_or_equal", "<=":
		if fieldNum, err := parseNumber(fieldValue); err == nil {
			if valueNum, err := parseNumber(value); err == nil {
				return fieldNum <= valueNum
			}
		}
		return false
	case "regex_match":
		// TODO: Implement regex matching safely
		return false
	default:
		logrus.Warn("Unknown condition operator:", operator)
		return false
	}
}

func parseNumber(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot parse %v as number", value)
	}
}

func (e *WorkflowExecutionEngine) substituteVariables(template string, ctx *ExecutionContext) string {
	result := template

	// Replace visitor variables
	result = strings.ReplaceAll(result, "{{visitorId}}", ctx.VisitorID)
	result = strings.ReplaceAll(result, "{{sessionId}}", ctx.SessionID)

	// Replace user variables
	if email, ok := ctx.IdentifiedUser["email"].(string); ok {
		result = strings.ReplaceAll(result, "{{user.email}}", email)
	}
	if name, ok := ctx.IdentifiedUser["name"].(string); ok {
		result = strings.ReplaceAll(result, "{{user.name}}", name)
	}

	// Replace timestamp
	result = strings.ReplaceAll(result, "{{timestamp}}", time.Now().Format(time.RFC3339))

	return result
}

func (e *WorkflowExecutionEngine) generateHMACSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func updateWorkflowStats(workflowID primitive.ObjectID, success bool, duration int64) {
	collection := database.GetWorkflowDB().Collection("workflows")
	statsCollection := database.GetWorkflowDB().Collection("workflow_stats")
	
	// Get current workflow stats
	var currentWorkflow struct {
		RunCount     int64   `bson:"runCount"`
		SuccessCount int64   `bson:"successCount"`
		TotalDuration int64  `bson:"totalDuration"`
		AvgDuration  float64 `bson:"avgDuration"`
	}
	
	err := collection.FindOne(context.Background(), bson.M{"_id": workflowID}).Decode(&currentWorkflow)
	if err != nil {
		// Initialize if not found
		currentWorkflow.RunCount = 0
		currentWorkflow.SuccessCount = 0
		currentWorkflow.TotalDuration = 0
		currentWorkflow.AvgDuration = 0
	}

	// Calculate new stats
	newRunCount := currentWorkflow.RunCount + 1
	newSuccessCount := currentWorkflow.SuccessCount
	if success {
		newSuccessCount++
	}
	
	newTotalDuration := currentWorkflow.TotalDuration + duration
	newAvgDuration := float64(newTotalDuration) / float64(newRunCount)
	successRate := float64(newSuccessCount) / float64(newRunCount) * 100

	// Update workflow document with comprehensive stats
	update := bson.M{
		"$set": bson.M{
			"runCount":     newRunCount,
			"successCount": newSuccessCount,
			"totalDuration": newTotalDuration,
			"avgDuration":  newAvgDuration,
			"successRate":  successRate,
			"lastRun":      time.Now(),
			"updatedAt":    time.Now(),
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": workflowID}, update)
	if err != nil {
		logrus.Error("Failed to update workflow stats:", err)
	}

	// Also store daily stats for analytics
	today := time.Now().Format("2006-01-02")
	dailyStatsUpdate := bson.M{
		"$inc": bson.M{
			"executions": 1,
		},
		"$set": bson.M{
			"date":      today,
			"updatedAt": time.Now(),
		},
	}

	if success {
		dailyStatsUpdate["$inc"].(bson.M)["successes"] = 1
	} else {
		dailyStatsUpdate["$inc"].(bson.M)["failures"] = 1
	}

	// Upsert daily stats
	_, err = statsCollection.UpdateOne(
		context.Background(),
		bson.M{
			"workflowId": workflowID,
			"date":       today,
		},
		dailyStatsUpdate,
		// Add upsert option
	)
	if err != nil {
		logrus.Error("Failed to update daily workflow stats:", err)
	}

	// Update hourly stats for real-time analytics
	currentHour := time.Now().Format("2006-01-02T15")
	hourlyStatsUpdate := bson.M{
		"$inc": bson.M{
			"executions": 1,
		},
		"$set": bson.M{
			"hour":      currentHour,
			"updatedAt": time.Now(),
		},
	}

	if success {
		hourlyStatsUpdate["$inc"].(bson.M)["successes"] = 1
	} else {
		hourlyStatsUpdate["$inc"].(bson.M)["failures"] = 1
	}

	hourlyCollection := database.GetWorkflowDB().Collection("workflow_hourly_stats")
	_, err = hourlyCollection.UpdateOne(
		context.Background(),
		bson.M{
			"workflowId": workflowID,
			"hour":       currentHour,
		},
		hourlyStatsUpdate,
	)
	if err != nil {
		logrus.Error("Failed to update hourly workflow stats:", err)
	}
}
