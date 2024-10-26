// backend/internal/common/types.go

package common

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RateLimit defines the structure for rate limiting configurations.
type RateLimit struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute" json:"requestsPerMinute" bson:"requestsPerMinute" validate:"required,min=1"`
	Burst             int `mapstructure:"burst" json:"burst" bson:"burst" validate:"required,min=1"`
}

// Metrics defines the structure for metrics configurations.
type Metrics struct {
	PrometheusEnabled  bool   `mapstructure:"prometheus_enabled" json:"prometheusEnabled" bson:"prometheusEnabled"`
	PrometheusEndpoint string `mapstructure:"prometheus_endpoint" json:"prometheusEndpoint" bson:"prometheusEndpoint" validate:"required_if=PrometheusEnabled true,url"`
	PrometheusPort     int    `mapstructure:"prometheus_port" json:"prometheusPort" bson:"prometheusPort" validate:"required_if=PrometheusEnabled true,min=1,max=65535"`
}

// Monitoring defines the structure for monitoring configurations.
type Monitoring struct {
	HealthCheckInterval string `mapstructure:"health_check_interval" json:"healthCheckInterval" bson:"healthCheckInterval" validate:"required,nonzero"`
}

// Destination represents where the payloads are delivered.
type Destination struct {
	Type      string `mapstructure:"type" json:"type" bson:"type" validate:"required,oneof=http file"`
	Name      string `mapstructure:"name" json:"name" bson:"name"`
	Endpoint  string `mapstructure:"endpoint" json:"endpoint" bson:"endpoint" validate:"omitempty,required_if=Type http,url"`
	Port      int    `mapstructure:"port" json:"port" bson:"port" validate:"omitempty,required_if=Type http,min=1,max=65535"`
	APIKey    string `mapstructure:"api_key" json:"apiKey" bson:"apiKey" validate:"omitempty"`
	FilePath  string `mapstructure:"file_path" json:"filePath" bson:"filePath" validate:"omitempty,required_if=Type file"`
	FileCount int    `mapstructure:"file_count" json:"fileCount" bson:"fileCount" validate:"omitempty,required_if=Type file,min=1"`
	FileFreq  int    `mapstructure:"file_freq" json:"fileFreq" bson:"fileFreq" validate:"omitempty,required_if=Type file,min=1"` // Frequency in minutes
}

// ServerConfig represents the server configuration section.
type ServerConfig struct {
	APIPort      string `mapstructure:"api_port" json:"apiPort" bson:"apiPort" validate:"required,port"`
	LoadgenPort  string `mapstructure:"loadgen_port" json:"loadgenPort" bson:"loadgenPort" validate:"required,port"`
	LoadgenURL   string `mapstructure:"loadgen_url" json:"loadgenUrl" bson:"loadgenUrl" validate:"required,url"`
	ReadTimeout  int    `mapstructure:"read_timeout" json:"readTimeout" bson:"readTimeout" validate:"required,min=1"`
	WriteTimeout int    `mapstructure:"write_timeout" json:"writeTimeout" bson:"writeTimeout" validate:"required,min=1"`
	IdleTimeout  int    `mapstructure:"idle_timeout" json:"idleTimeout" bson:"idleTimeout" validate:"required,min=1"`
}

// Config represents the application's configuration settings.
type Config struct {
	Server            ServerConfig  `mapstructure:"server" json:"server" bson:"server"`
	LoadgenURL        string        `mapstructure:"loadgen_url" json:"loadgenUrl" bson:"loadgenUrl" validate:"required,url"`
	LogLevel          string        `mapstructure:"log_level" json:"logLevel" bson:"logLevel" validate:"required,oneof=debug info warn error fatal"`
	LogFormat         string        `mapstructure:"log_format" json:"logFormat" bson:"logFormat" validate:"required,oneof=json text"`
	LogOutput         string        `mapstructure:"log_output" json:"logOutput" bson:"logOutput" validate:"required,oneof=stdout stderr file"`
	LogFilePath       string        `mapstructure:"log_file_path" json:"logFilePath" bson:"logFilePath" validate:"required_if=LogOutput file"`
	MongoURI          string        `mapstructure:"mongo_uri" json:"mongoURI" bson:"mongoURI" validate:"required,url"`
	MongoDB           string        `mapstructure:"mongo_db" json:"mongoDB" bson:"mongoDB" validate:"required"`
	JWTSecret         string        `mapstructure:"jwt_secret" json:"jwtSecret" bson:"jwtSecret" validate:"required,min=32"`
	JWTExpiry         string        `mapstructure:"jwt_expiry" json:"jwtExpiry" bson:"jwtExpiry" validate:"required"`
	AllowedOrigins    []string      `mapstructure:"allowed_origins" json:"allowedOrigins" bson:"allowedOrigins" validate:"required,dive,url"`
	RateLimit         RateLimit     `mapstructure:"rate_limit" json:"rateLimit" bson:"rateLimit"`
	SecurityRateLimit RateLimit     `mapstructure:"security.rate_limiting" json:"securityRateLimit" bson:"securityRateLimit"`
	Metrics           Metrics       `mapstructure:"metrics" json:"metrics" bson:"metrics"`
	EnableTLS         bool          `mapstructure:"enable_tls" json:"enableTLS" bson:"enableTLS"`
	TLSCertPath       string        `mapstructure:"tls_cert_path" json:"tlsCertPath" bson:"tlsCertPath" validate:"required_if=EnableTLS true"`
	TLSKeyPath        string        `mapstructure:"tls_key_path" json:"tlsKeyPath" bson:"tlsKeyPath" validate:"required_if=EnableTLS true"`
	Destinations      []Destination `mapstructure:"destinations" json:"destinations" bson:"destinations" validate:"required,dive"`
	LogRate           int           `mapstructure:"log_rate" json:"logRate" bson:"logRate" validate:"required,min=1"`
	MetricsRate       int           `mapstructure:"metrics_rate" json:"metricsRate" bson:"metricsRate" validate:"required,min=1"`
	TraceRate         int           `mapstructure:"trace_rate" json:"traceRate" bson:"traceRate" validate:"required,min=1"`
	LogSize           int           `mapstructure:"log_size" json:"logSize" bson:"logSize" validate:"required,min=1"`
	MetricsValue      float64       `mapstructure:"metrics_value" json:"metricsValue" bson:"metricsValue" validate:"required"`
	DefaultRoles      []string      `mapstructure:"default_roles" json:"defaultRoles" bson:"defaultRoles" validate:"required,dive,required"`
	Monitoring        Monitoring    `mapstructure:"monitoring" json:"monitoring" bson:"monitoring"`
	ServerPort        string        `mapstructure:"server_port" json:"serverPort" bson:"serverPort" validate:"required,port"`
}

// User represents a user in the system.
type User struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username  string               `bson:"username" json:"username" validate:"required,min=3,max=30"`
	Email     string               `bson:"email" json:"email" validate:"required,email"`
	Password  string               `bson:"password" json:"password" validate:"required,min=8"`
	Roles     []primitive.ObjectID `bson:"roles" json:"roles" validate:"required,dive,required"`
	CreatedAt time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updatedAt"`
}

// ValidationError represents a validation error for a specific field.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

//// Test defines the structure of a load test.
//type Test struct {
//	TestID      string      `json:"testID" validate:"required"`
//	Duration    int         `json:"duration" validate:"required,gt=0"` // Duration in seconds
//	Destination Destination `json:"destination" validate:"required,dive"`
//	// Add other relevant fields as needed
//}

// CancelRequest represents a request to cancel a load test.
type CancelRequest struct {
	TestID string `json:"testID" validate:"required"`
}

// RestartRequest represents a request to restart a load test.
type RestartRequest struct {
	TestID string `json:"testID" validate:"required"`
}

// TestResults represents the results of a load test.
type TestResults struct {
	TestID      string    `json:"testID" validate:"required"`
	CompletedAt time.Time `json:"completedAt"`
	Logs        []LogEntry
	Metrics     []Metric
	Traces      []Trace
}

// LogEntry represents a single log entry.
type LogEntry struct {
	TestID    string    `json:"testID" bson:"testID"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Message   string    `json:"message" bson:"message"`
	Level     string    `json:"level" bson:"level"`
}

// Metric represents a single metric data point.
type Metric struct {
	TestID    string    `json:"testID" bson:"testID"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Value     float64   `json:"value" bson:"value"`
}

// Trace represents a single trace data point.
type Trace struct {
	TestID    string    `json:"testID" bson:"testID"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	TraceID   string    `json:"traceID" bson:"traceID"`
	SpanID    string    `json:"spanID" bson:"spanID"`
	Operation string    `json:"operation" bson:"operation"`
	Duration  int       `json:"duration" bson:"duration"` // Duration in ms
}
