// backend/internal/common/types.go

package common

// RateLimit defines the structure for rate limiting configurations.
type RateLimit struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}

// Metrics defines the structure for metrics configurations.
type Metrics struct {
	Enabled  bool    `mapstructure:"enabled"`
	Port     string  `mapstructure:"port"`
	Interval float64 `mapstructure:"interval"`
}

// Monitoring defines the structure for monitoring configurations.
type Monitoring struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	Interval int    `mapstructure:"interval"`
}

// Destination represents where the payloads are delivered.
type Destination struct {
	Port     int    `mapstructure:"port" bson:"port"`
	Endpoint string `mapstructure:"endpoint" bson:"endpoint"`
}

// Config represents the application's configuration settings.
type Config struct {
	APIPort           string        `mapstructure:"api_port"`
	LogLevel          string        `mapstructure:"log_level"`
	LogFormat         string        `mapstructure:"log_format"`
	LogOutput         string        `mapstructure:"log_output"`
	MongoURI          string        `mapstructure:"mongo_uri"`
	MongoDB           string        `mapstructure:"mongo_db"`
	JWTSecret         string        `mapstructure:"jwt_secret"`
	JWTExpiry         string        `mapstructure:"jwt_expiry"`
	AllowedOrigins    []string      `mapstructure:"allowed_origins"`
	RateLimit         RateLimit     `mapstructure:"rate_limit"`
	SecurityRateLimit RateLimit     `mapstructure:"security.rate_limiting"`
	Metrics           Metrics       `mapstructure:"metrics"`
	EnableTLS         bool          `mapstructure:"enable_tls"`
	TLSCertPath       string        `mapstructure:"tls_cert_path"`
	TLSKeyPath        string        `mapstructure:"tls_key_path"`
	Destinations      []Destination `mapstructure:"destinations"`
	LogRate           int           `mapstructure:"log_rate"`
	MetricsRate       int           `mapstructure:"metrics_rate"`
	TraceRate         int           `mapstructure:"trace_rate"`
	LogSize           int           `mapstructure:"log_size"`
	MetricsValue      float64       `mapstructure:"metrics_value"`
	DefaultRoles      []string      `mapstructure:"default_roles"`
	Monitoring        Monitoring    `mapstructure:"monitoring"`
}
