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
	Username  string               `bson:"username" json:"username" validate:"required,min=3,max=30"`
	Email     string               `bson:"email" json:"email" validate:"required,email"`
	Password  string               `bson:"password" json:"password" validate:"required,min=8"`
	Roles     []primitive.ObjectID `bson:"roles" json:"roles" validate:"required,dive,required"`
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
	TestID        string             `json:"testID" bson:"testID" validate:"required"`
	UserID        string             `json:"userID" bson:"userID" validate:"required"`
	LogType       string             `json:"logType" bson:"logType" validate:"required,oneof=INFO WARN ERROR DEBUG"`
	LogRate       int                `json:"logRate,omitempty" bson:"logRate" validate:"omitempty,min=1"`         // Logs per second
	LogSize       int                `json:"logSize,omitempty" bson:"logSize" validate:"omitempty,min=1"`         // Size of each log entry in bytes
	MetricsRate   int                `json:"metricsRate,omitempty" bson:"metricsRate" validate:"omitempty,min=1"` // Metrics per second
	TraceRate     int                `json:"traceRate,omitempty" bson:"traceRate" validate:"omitempty,min=1"`     // Traces per second
	Duration      int                `json:"duration" bson:"duration" validate:"required,min=1"`                  // Duration in seconds
	Destination   common.Destination `json:"destination" bson:"destination" validate:"required"`
	Status        string             `json:"status" bson:"status" validate:"required,oneof=Pending Running Completed Cancelled"`
	ScheduledTime time.Time          `json:"scheduledTime,omitempty" bson:"scheduledTime,omitempty"`
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`
	CompletedAt   time.Time          `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
}

// LogEntry represents a log entry.
type LogEntry struct {
	TestID    string    `json:"testID" bson:"testID" validate:"required"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp" validate:"required"`
	Message   string    `json:"message" bson:"message" validate:"required"`
	Level     string    `json:"level" bson:"level" validate:"required,oneof=INFO WARN ERROR"`
}

// Metric represents a metric data point.
type Metric struct {
	TestID    string    `json:"testID" bson:"testID" validate:"required"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp" validate:"required"`
	Value     float64   `json:"value" bson:"value" validate:"required"`
}

// Trace represents a trace data point.
type Trace struct {
	TestID    string    `json:"testID" bson:"testID" validate:"required"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp" validate:"required"`
	TraceID   string    `json:"traceID" bson:"traceID" validate:"required,len=16"`
	SpanID    string    `json:"spanID" bson:"spanID" validate:"required,len=8"`
	Operation string    `json:"operation" bson:"operation" validate:"required"`
	Duration  int       `json:"duration" bson:"duration" validate:"required,min=1"` // Duration in ms
}

// ScheduleRequest represents a request to schedule a load test.
type ScheduleRequest struct {
	TestID     string    `json:"testID" bson:"testID" validate:"required"`
	ScheduleAt time.Time `json:"scheduleAt" bson:"scheduleAt" validate:"required"`
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
	TestID      string     `json:"testID" bson:"testID" validate:"required"`
	CompletedAt time.Time  `json:"completedAt" bson:"completedAt" validate:"required"`
	Logs        []LogEntry `json:"logs" bson:"logs" validate:"dive"`
	Metrics     []Metric   `json:"metrics" bson:"metrics" validate:"dive"`
	Traces      []Trace    `json:"traces" bson:"traces" validate:"dive"`
}

// ValidationError represents a structured validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Custom errors
var (
	ErrInvalidToken           = errors.New("invalid token")
	ErrTestNotFound           = errors.New("test not found")
	ErrTestAlreadyExists      = errors.New("test already exists")
	ErrTestAlreadyCompleted   = errors.New("test already completed")
	ErrTestAlreadyCancelled   = errors.New("test already cancelled")
	ErrDestinationUnsupported = errors.New("unsupported destination type")
)
