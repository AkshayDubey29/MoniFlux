// backend/internal/api/models/models.go

package models

import (
	"errors"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	jwt "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user entity in the system.
type User struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username  string               `bson:"username" json:"username"`
	Email     string               `bson:"email" json:"email"`
	Password  string               `bson:"password" json:"password"`
	Roles     []primitive.ObjectID `bson:"roles" json:"roles"`
	CreatedAt time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updatedAt"`
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
	LogRate       int                `json:"logRate,omitempty" validate:"omitempty,min=1"`       // Logs per second
	LogSize       int                `json:"logSize,omitempty" validate:"omitempty,min=1"`       // Milliseconds per log
	MetricsRate   int                `json:"metricsRate,omitempty" validate:"omitempty,min=1"`   // Metrics per second
	TraceRate     int                `json:"traceRate,omitempty" validate:"omitempty,min=1"`     // Traces per second
	Duration      int                `json:"duration" bson:"duration" validate:"required,min=1"` // Duration in seconds
	Destination   common.Destination `json:"destination" bson:"destination"`
	Status        string             `json:"status" bson:"status"`
	ScheduledTime time.Time          `json:"scheduledTime,omitempty" bson:"scheduledTime,omitempty"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	CompletedAt   time.Time          `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
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
	TestID   string    `json:"testID" bson:"testID" validate:"required"`
	UserID   string    `json:"userID" bson:"userID" validate:"required"`
	Schedule time.Time `json:"schedule" bson:"schedule" validate:"required"`
}

// CancelRequest represents a request to cancel a load test.
type CancelRequest struct {
	TestID string `json:"testID" bson:"testID" validate:"required"`
}

// RestartRequest represents a request to restart a load test.
type RestartRequest struct {
	TestID      string `json:"testID" bson:"testID" validate:"required"`
	LogRate     int    `json:"logRate,omitempty" bson:"logRate" validate:"omitempty,min=1"`
	MetricsRate int    `json:"metricsRate,omitempty" bson:"metricsRate" validate:"omitempty,min=1"`
	TraceRate   int    `json:"traceRate,omitempty" bson:"traceRate" validate:"omitempty,min=1"`
	Duration    int    `json:"duration" bson:"duration" validate:"required,min=1"`
	// Add other configuration fields as needed.
}

// TestResults represents the results of a load test.
type TestResults struct {
	TestID      string    `json:"testID" bson:"testID" validate:"required"`
	Results     string    `json:"results" bson:"results" validate:"required"`
	CompletedAt time.Time `json:"completedAt" bson:"completedAt" validate:"required"`
}

// ValidationError represents a structured validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Custom errors
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTestNotFound = errors.New("test not found")
)
