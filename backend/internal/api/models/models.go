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
// This line makes the usage explicit, preventing the import from being removed.
var _ jwt.Claims = &Claims{}

// Optionally, add a dummy reference if needed
var _ = jwt.New

// Test represents a load test configuration and status.
type Test struct {
	TestID        string             `json:"testID" bson:"testID"`                                   // Unique identifier for the test.
	UserID        string             `json:"userID" bson:"userID"`                                   // ID of the user who initiated the test.
	LogType       string             `json:"logType" bson:"logType"`                                 // Type of log to be generated (e.g., Catalina, Nginx).
	LogRate       int                `json:"logRate" bson:"logRate"`                                 // Rate at which logs are generated.
	LogSize       int                `json:"logSize" bson:"logSize"`                                 // Size of each log entry.
	MetricsRate   int                `json:"metricsRate" bson:"metricsRate"`                         // Rate at which metrics are generated.
	TraceRate     int                `json:"traceRate" bson:"traceRate"`                             // Rate at which traces are generated.
	Duration      int                `json:"duration" bson:"duration"`                               // Duration for which the test runs (in seconds).
	Destination   common.Destination `json:"destination" bson:"destination"`                         // Destination details for payload delivery.
	Status        string             `json:"status" bson:"status"`                                   // Current status of the test (e.g., running, completed).
	ScheduledTime time.Time          `json:"scheduledTime,omitempty" bson:"scheduledTime,omitempty"` // Time when the test is scheduled.
	CreatedAt     time.Time          `json:"createdAt" bson:"createdAt"`                             // Timestamp of when the test was created.
	UpdatedAt     time.Time          `json:"updatedAt" bson:"updatedAt"`                             // Timestamp of when the test was last updated.
}

// LogEntry represents a log entry.
type LogEntry struct {
	TestID    string    `json:"testID" bson:"testID"`       // Test ID associated with this log entry.
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // Time the log was generated.
	Message   string    `json:"message" bson:"message"`     // Log message.
	Level     string    `json:"level" bson:"level"`         // Log level (e.g., INFO, ERROR).
}

// Metric represents a metric data point.
type Metric struct {
	TestID    string    `json:"testID" bson:"testID"`       // Test ID associated with this metric.
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // Time the metric was recorded.
	Value     float64   `json:"value" bson:"value"`         // Value of the metric.
}

// Trace represents a trace data point.
type Trace struct {
	TestID    string    `json:"testID" bson:"testID"`       // Test ID associated with this trace.
	Timestamp time.Time `json:"timestamp" bson:"timestamp"` // Time the trace was recorded.
	TraceID   string    `json:"traceID" bson:"traceID"`     // ID of the trace.
	SpanID    string    `json:"spanID" bson:"spanID"`       // ID of the trace span.
	Operation string    `json:"operation" bson:"operation"` // Operation name.
	Duration  int       `json:"duration" bson:"duration"`   // Duration of the trace in milliseconds.
}

// ScheduleRequest represents a request to schedule a load test.
type ScheduleRequest struct {
	TestID   string    `json:"testID" bson:"testID"`     // ID of the test to schedule.
	UserID   string    `json:"userID" bson:"userID"`     // ID of the user scheduling the test.
	Schedule time.Time `json:"schedule" bson:"schedule"` // Time to schedule the test.
}

type CancelRequest struct {
	TestID string `json:"testID" bson:"testID"` // ID of the test to cancel.
}

// RestartRequest represents a request to restart a load test.
type RestartRequest struct {
	TestID   string `json:"testID" bson:"testID"`     // ID of the test to restart.
	LogRate  int    `json:"logRate" bson:"logRate"`   // Updated log rate.
	Duration int    `json:"duration" bson:"duration"` // Updated duration.
	// Add other configuration fields as needed.
}

// TestResults represents the results of a load test.
type TestResults struct {
	TestID      string    `json:"testID" bson:"testID"`           // ID of the test.
	Results     string    `json:"results" bson:"results"`         // Results summary.
	CompletedAt time.Time `json:"completedAt" bson:"completedAt"` // Time when the test was completed.
}

// ValidationError represents a structured validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrInvalidToken is returned when a JWT token is invalid.
var ErrInvalidToken = errors.New("invalid token")

// Config represents the application's configuration settings.
type Config struct {
	APIPort           string               `mapstructure:"api_port"`
	LogLevel          string               `mapstructure:"log_level"`
	LogFormat         string               `mapstructure:"log_format"`
	LogOutput         string               `mapstructure:"log_output"`
	MongoURI          string               `mapstructure:"mongo_uri"`
	MongoDB           string               `mapstructure:"mongo_db"`
	JWTSecret         string               `mapstructure:"jwt_secret"`
	JWTExpiry         string               `mapstructure:"jwt_expiry"`
	AllowedOrigins    []string             `mapstructure:"allowed_origins"`
	RateLimit         common.RateLimit     `mapstructure:"rate_limit"`
	SecurityRateLimit common.RateLimit     `mapstructure:"security.rate_limiting"`
	Metrics           common.Metrics       `mapstructure:"metrics"`
	EnableTLS         bool                 `mapstructure:"enable_tls"`
	TLSCertPath       string               `mapstructure:"tls_cert_path"`
	TLSKeyPath        string               `mapstructure:"tls_key_path"`
	Destinations      []common.Destination `mapstructure:"destinations"`
	LogRate           int                  `mapstructure:"log_rate"`
	MetricsRate       int                  `mapstructure:"metrics_rate"`
	TraceRate         int                  `mapstructure:"trace_rate"`
	LogSize           int                  `mapstructure:"log_size"`
	MetricsValue      float64              `mapstructure:"metrics_value"`
	DefaultRoles      []string             `mapstructure:"default_roles"`
	Monitoring        common.Monitoring    `mapstructure:"monitoring"`
}
