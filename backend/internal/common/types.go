// backend/internal/common/types.go

package common

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RateLimit defines the structure for rate limiting configurations.
type RateLimit struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute" json:"requestsPerMinute" bson:"requestsPerMinute"`
	Burst             int `mapstructure:"burst" json:"burst" bson:"burst"`
}

// Metrics defines the structure for metrics configurations.
type Metrics struct {
	PrometheusEnabled  bool   `mapstructure:"prometheus_enabled" json:"prometheusEnabled" bson:"prometheusEnabled"`
	PrometheusEndpoint string `mapstructure:"prometheus_endpoint" json:"prometheusEndpoint" bson:"prometheusEndpoint"`
	PrometheusPort     int    `mapstructure:"prometheus_port" json:"prometheusPort" bson:"prometheusPort"`
}

// Monitoring defines the structure for monitoring configurations.
type Monitoring struct {
	HealthCheckInterval string `mapstructure:"health_check_interval" json:"healthCheckInterval" bson:"healthCheckInterval"`
}

// Destination represents where the payloads are delivered.
type Destination struct {
	Name     string `mapstructure:"name" json:"name" bson:"name"`             // Name of the destination.
	Endpoint string `mapstructure:"endpoint" json:"endpoint" bson:"endpoint"` // Endpoint URL for the destination.
	Port     int    `mapstructure:"port" json:"port" bson:"port"`             // Port number for the destination.
	APIKey   string `mapstructure:"api_key" json:"apiKey" bson:"apiKey"`      // API key for authentication.
}

// ServerConfig represents the server configuration section.
type ServerConfig struct {
	Port         string `mapstructure:"port" json:"port" bson:"port"`                          // Port on which the server runs.
	ReadTimeout  int    `mapstructure:"read_timeout" json:"readTimeout" bson:"readTimeout"`    // Read timeout in seconds.
	WriteTimeout int    `mapstructure:"write_timeout" json:"writeTimeout" bson:"writeTimeout"` // Write timeout in seconds.
	IdleTimeout  int    `mapstructure:"idle_timeout" json:"idleTimeout" bson:"idleTimeout"`    // Idle timeout in seconds.
}

// Config represents the application's configuration settings.
type Config struct {
	Server            ServerConfig  `mapstructure:"server" json:"server" bson:"server"`
	LogLevel          string        `mapstructure:"log_level" json:"logLevel" bson:"logLevel"`
	LogFormat         string        `mapstructure:"log_format" json:"logFormat" bson:"logFormat"`
	LogOutput         string        `mapstructure:"log_output" json:"logOutput" bson:"logOutput"`
	LogFilePath       string        `mapstructure:"log_file_path" json:"logFilePath" bson:"logFilePath"` // Path to log file.
	MongoURI          string        `mapstructure:"mongo_uri" json:"mongoURI" bson:"mongoURI"`
	MongoDB           string        `mapstructure:"mongo_db" json:"mongoDB" bson:"mongoDB"`
	JWTSecret         string        `mapstructure:"jwt_secret" json:"jwtSecret" bson:"jwtSecret"`
	JWTExpiry         string        `mapstructure:"jwt_expiry" json:"jwtExpiry" bson:"jwtExpiry"`
	AllowedOrigins    []string      `mapstructure:"allowed_origins" json:"allowedOrigins" bson:"allowedOrigins"`
	RateLimit         RateLimit     `mapstructure:"rate_limit" json:"rateLimit" bson:"rateLimit"`
	SecurityRateLimit RateLimit     `mapstructure:"security.rate_limiting" json:"securityRateLimit" bson:"securityRateLimit"`
	Metrics           Metrics       `mapstructure:"metrics" json:"metrics" bson:"metrics"`
	EnableTLS         bool          `mapstructure:"enable_tls" json:"enableTLS" bson:"enableTLS"`
	TLSCertPath       string        `mapstructure:"tls_cert_path" json:"tlsCertPath" bson:"tlsCertPath"`
	TLSKeyPath        string        `mapstructure:"tls_key_path" json:"tlsKeyPath" bson:"tlsKeyPath"`
	Destinations      []Destination `mapstructure:"destinations" json:"destinations" bson:"destinations"`
	LogRate           int           `mapstructure:"log_rate" json:"logRate" bson:"logRate"`
	MetricsRate       int           `mapstructure:"metrics_rate" json:"metricsRate" bson:"metricsRate"`
	TraceRate         int           `mapstructure:"trace_rate" json:"traceRate" bson:"traceRate"`
	LogSize           int           `mapstructure:"log_size" json:"logSize" bson:"logSize"`
	MetricsValue      float64       `mapstructure:"metrics_value" json:"metricsValue" bson:"metricsValue"`
	DefaultRoles      []string      `mapstructure:"default_roles" json:"defaultRoles" bson:"defaultRoles"`
	Monitoring        Monitoring    `mapstructure:"monitoring" json:"monitoring" bson:"monitoring"`
	APIPort           string
	ServerPort        string `mapstructure:"server_port" json:"serverPort" bson:"serverPort"`
}

// User represents a user in the system.
type User struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username  string               `bson:"username" json:"username"`
	Email     string               `bson:"email" json:"email"`
	Password  string               `bson:"password" json:"password"`
	Roles     []primitive.ObjectID `bson:"roles" json:"roles"`
	CreatedAt time.Time            `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updatedAt"`
}
