package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Workflow struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"userId" bson:"userId" validate:"required"`
	WebsiteID   primitive.ObjectID `json:"websiteId" bson:"websiteId" validate:"required"`
	Name        string             `json:"name" bson:"name" validate:"required"`
	Description string             `json:"description" bson:"description"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
	
	// Workflow configuration
	Triggers    []WorkflowTrigger  `json:"triggers" bson:"triggers"`
	Actions     []WorkflowAction   `json:"actions" bson:"actions"`
	Conditions  []WorkflowCondition `json:"conditions" bson:"conditions"`
	
	// Execution settings
	ExecutionMode string            `json:"executionMode" bson:"executionMode"` // immediate, scheduled, manual
	Schedule      *WorkflowSchedule `json:"schedule,omitempty" bson:"schedule,omitempty"`
	
	// Metadata
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	LastRun     *time.Time         `json:"lastRun,omitempty" bson:"lastRun,omitempty"`
	RunCount    int64              `json:"runCount" bson:"runCount"`
	
	// Analytics
	SuccessRate float64            `json:"successRate" bson:"successRate"`
	AvgRunTime  float64            `json:"avgRunTime" bson:"avgRunTime"` // milliseconds
}

type WorkflowTrigger struct {
	ID       string                 `json:"id" bson:"id"`
	Type     string                 `json:"type" bson:"type"` // event, schedule, webhook, manual
	Config   map[string]interface{} `json:"config" bson:"config"`
	IsActive bool                   `json:"isActive" bson:"isActive"`
}

type WorkflowAction struct {
	ID       string                 `json:"id" bson:"id"`
	Type     string                 `json:"type" bson:"type"` // email, webhook, analytics, notification
	Config   map[string]interface{} `json:"config" bson:"config"`
	Order    int                    `json:"order" bson:"order"`
	IsActive bool                   `json:"isActive" bson:"isActive"`
}

type WorkflowCondition struct {
	ID       string                 `json:"id" bson:"id"`
	Field    string                 `json:"field" bson:"field"`
	Operator string                 `json:"operator" bson:"operator"` // equals, contains, greater_than, etc.
	Value    interface{}            `json:"value" bson:"value"`
	LogicOp  string                 `json:"logicOp" bson:"logicOp"` // and, or
}

type WorkflowSchedule struct {
	Type       string    `json:"type" bson:"type"` // cron, interval
	Expression string    `json:"expression" bson:"expression"`
	Timezone   string    `json:"timezone" bson:"timezone"`
	StartDate  time.Time `json:"startDate" bson:"startDate"`
	EndDate    *time.Time `json:"endDate,omitempty" bson:"endDate,omitempty"`
}

type WorkflowExecution struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WorkflowID primitive.ObjectID `json:"workflowId" bson:"workflowId"`
	Status     string             `json:"status" bson:"status"` // pending, running, completed, failed
	StartTime  time.Time          `json:"startTime" bson:"startTime"`
	EndTime    *time.Time         `json:"endTime,omitempty" bson:"endTime,omitempty"`
	Duration   int64              `json:"duration" bson:"duration"` // milliseconds
	
	// Execution details
	TriggerData map[string]interface{} `json:"triggerData" bson:"triggerData"`
	Results     []ActionResult         `json:"results" bson:"results"`
	Error       string                 `json:"error,omitempty" bson:"error,omitempty"`
	
	// Metadata
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type ActionResult struct {
	ActionID  string                 `json:"actionId" bson:"actionId"`
	Status    string                 `json:"status" bson:"status"` // success, failed, skipped
	Output    map[string]interface{} `json:"output" bson:"output"`
	Error     string                 `json:"error,omitempty" bson:"error,omitempty"`
	Duration  int64                  `json:"duration" bson:"duration"` // milliseconds
}

type VisitorEvent struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WebsiteID  primitive.ObjectID `json:"websiteId" bson:"websiteId"`
	VisitorID  string             `json:"visitorId" bson:"visitorId"`
	SessionID  string             `json:"sessionId" bson:"sessionId"`
	
	// Event details
	EventType  string                 `json:"eventType" bson:"eventType"` // pageview, click, form_submit, etc.
	EventData  map[string]interface{} `json:"eventData" bson:"eventData"`
	
	// Page information
	URL        string `json:"url" bson:"url"`
	Title      string `json:"title" bson:"title"`
	Referrer   string `json:"referrer" bson:"referrer"`
	
	// User agent and device info
	UserAgent  string `json:"userAgent" bson:"userAgent"`
	IPAddress  string `json:"ipAddress" bson:"ipAddress"`
	Country    string `json:"country" bson:"country"`
	City       string `json:"city" bson:"city"`
	
	// Timestamps
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
