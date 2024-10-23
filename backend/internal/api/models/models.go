// backend/internal/api/models/models.go

package models

import (
	"errors"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	jwt "github.com/golang-jwt/jwt/v4"
)

// User represents a user entity in the system.
type User struct {
	UserID    string    `json:"userID" bson:"userID"`
	Username  string    `json:"username" bson:"username"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"password" bson:"password"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// Claims represents the JWT claims.
type Claims struct {
	UserID string `json:"userID" bson:"userID"`
	jwt.RegisteredClaims
}

// Ensure that Claims implements the jwt.Claims interface.
var _ jwt.Claims = &Claims{}

// Test represents a load test configuration and status.
type Test struct {
	TestID        string             `json:"testID" bson:"testID"`
	UserID        string             `json:"userID" bson:"userID"`
	LogType       string             `json:"logType" bson:"logType"`
	LogRate       int                `json:"logRate" bson:"logRate"`
	LogSize       int                `json:"logSize" bson:"logSize"`
	MetricsRate   int                `json:"metricsRate" bson:"metricsRate"`
	TraceRate     int                `json:"traceRate" bson:"traceRate"`
	Duration      int                `json:"duration" bson:"duration"`
	Destination   common.Destination `json:"destination" bson:"destination"`
	Status        string             `json:"status" bson:"status"`
	ScheduledTime time.Time          `json:"scheduledTime,omitempty" bson:"scheduledTime,omitempty"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// LogEntry represents a log entry.
type LogEntry struct {
	TestID    string    `json:"testID" bson:"testID"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Message   string    `json:"message" bson:"message"`
	Level     string    `json:"level" bson:"level"`
}

// Metric represents a metric data point.
type Metric struct {
	TestID    string    `json:"testID" bson:"testID"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Value     float64   `json:"value" bson:"value"`
}

// Trace represents a trace data point.
type Trace struct {
	TestID    string    `json:"testID" bson:"testID"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	TraceID   string    `json:"traceID" bson:"traceID"`
	SpanID    string    `json:"spanID" bson:"spanID"`
	Operation string    `json:"operation" bson:"operation"`
	Duration  int       `json:"duration" bson:"duration"`
}

// ScheduleRequest represents a request to schedule a load test.
type ScheduleRequest struct {
	TestID   string    `json:"testID" bson:"testID"`
	UserID   string    `json:"userID" bson:"userID"`
	Schedule time.Time `json:"schedule" bson:"schedule"`
}

// CancelRequest represents a request to cancel a load test.
type CancelRequest struct {
	TestID string `json:"testID" bson:"testID"`
}

// RestartRequest represents a request to restart a load test.
type RestartRequest struct {
	TestID   string `json:"testID" bson:"testID"`
	LogRate  int    `json:"logRate" bson:"logRate"`
	Duration int    `json:"duration" bson:"duration"`
	// Add other configuration fields as needed.
}

// TestResults represents the results of a load test.
type TestResults struct {
	TestID      string    `json:"testID" bson:"testID"`
	Results     string    `json:"results" bson:"results"`
	CompletedAt time.Time `json:"completedAt" bson:"completedAt"`
}

// ValidationError represents a structured validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrInvalidToken is returned when a JWT token is invalid.
var ErrInvalidToken = errors.New("invalid token")
