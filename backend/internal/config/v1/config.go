// backend/internal/config/v1/config.go

package v1

// Config represents the application's configuration settings.

// RateLimit represents the rate limiting configuration.
type RateLimit struct {
	RequestsPerMin int `mapstructure:"requests_per_minute"`
	Burst          int `mapstructure:"burst"`
}

// Metrics represents the metrics configuration.
type Metrics struct {
	PrometheusEnabled  bool   `mapstructure:"prometheus_enabled"`
	PrometheusEndpoint string `mapstructure:"prometheus_endpoint"`
	PrometheusPort     int    `mapstructure:"prometheus_port"`
}

// Monitoring represents the monitoring configuration.
type Monitoring struct {
	HealthCheckInterval string `mapstructure:"health_check_interval"`
}
