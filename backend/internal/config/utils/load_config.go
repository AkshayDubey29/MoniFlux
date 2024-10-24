// backend/internal/config/utils/load_config.go

package utils

import (
	"fmt"
	"strings"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/spf13/viper"
)

// LoadConfig loads the application configuration from the specified path.
// It supports reading from configuration files, environment variables, and setting default values.
// The function returns a pointer to the Config struct and an error if the loading fails.
func LoadConfig(path string) (*common.Config, error) {
	// Initialize Viper
	v := viper.New()

	// Set the file name and path if provided
	if path != "" {
		v.SetConfigFile(path)
	}

	// Enable environment variables reading
	v.AutomaticEnv()

	// Set environment variable prefix to avoid conflicts
	v.SetEnvPrefix("MONIFLUX")

	// Replace dots in env variables with underscores
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults(v)

	// Read the configuration file
	if err := v.ReadInConfig(); err != nil {
		// If the config file is not found, proceed with environment variables and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal the configuration into the Config struct
	var config common.Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// Validate the configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation error: %w", err)
	}

	return &config, nil
}

// setDefaults sets the default configuration values.
// These defaults are used if the configuration file or environment variables do not provide specific values.
func setDefaults(v *viper.Viper) {
	// Set default configuration values
	v.SetDefault("server.api_port", "8080")
	v.SetDefault("server.loadgen_port", "9080")

	v.SetDefault("server.read_timeout", 15)
	v.SetDefault("server.write_timeout", 15)
	v.SetDefault("server.idle_timeout", 60)

	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "json")
	v.SetDefault("log_output", "stdout")
	v.SetDefault("log_file_path", "/var/log/moniflux/app.log")

	v.SetDefault("mongo_uri", "mongodb://mongodb:27017")
	v.SetDefault("mongo_db", "moniflux")

	v.SetDefault("jwt_secret", "default-jwt-secret")
	v.SetDefault("jwt_expiry", "24h")

	v.SetDefault("allowed_origins", []string{"https://frontend.example.com"})

	v.SetDefault("rate_limit.requests_per_minute", 100)
	v.SetDefault("rate_limit.burst", 20)

	v.SetDefault("security.rate_limiting.requests_per_minute", 1000)
	v.SetDefault("security.rate_limiting.burst", 200)
	v.SetDefault("security.rate_limiting.cooldown_seconds", 60)

	v.SetDefault("metrics.prometheus_enabled", true)
	v.SetDefault("metrics.prometheus_endpoint", "/metrics")
	v.SetDefault("metrics.prometheus_port", 2112)
	v.SetDefault("metrics.namespace", "moniflux")
	v.SetDefault("metrics.subsystem", "api_server")

	v.SetDefault("enable_tls", false)
	v.SetDefault("tls_cert_path", "/path/to/cert.pem")
	v.SetDefault("tls_key_path", "/path/to/key.pem")

	v.SetDefault("destinations", []common.Destination{
		{Name: "destination1", Endpoint: "https://destination1.example.com/api", Port: 443, APIKey: "your_destination1_api_key_here"},
		{Name: "destination2", Endpoint: "https://destination2.example.com/api", Port: 443, APIKey: "your_destination2_api_key_here"},
	})

	v.SetDefault("log_rate", 100)
	v.SetDefault("metrics_rate", 50)
	v.SetDefault("trace_rate", 20)
	v.SetDefault("log_size", 1)
	v.SetDefault("metrics_value", 100.0)

	v.SetDefault("default_roles", []string{"admin", "editor", "viewer"})

	v.SetDefault("monitoring.health_check_interval", "5m")

	v.SetDefault("environment", "production")

	v.SetDefault("features.enable_debug_mode", false)
	v.SetDefault("features.enable_cache", true)
	v.SetDefault("features.beta_features.new_dashboard", false)
	v.SetDefault("features.beta_features.advanced_reports", false)

	v.SetDefault("cache.type", "redis")
	v.SetDefault("cache.redis.uri", "redis://localhost:6379")
	v.SetDefault("cache.redis.password", "")
	v.SetDefault("cache.redis.db", 0)
	v.SetDefault("cache.redis.pool_size", 20)
	v.SetDefault("cache.redis.idle_timeout", "300s")

	v.SetDefault("paths.log_file", "/var/log/moniflux/app.log")
	v.SetDefault("paths.data_dir", "/var/lib/moniflux/data")
	v.SetDefault("paths.temp_dir", "/tmp/moniflux")
}

// validateConfig performs validation on the loaded configuration.
// It ensures that essential configurations are set correctly.
func validateConfig(config *common.Config) error {
	// Example validation; extend as necessary
	if config.JWTSecret == "" {
		return fmt.Errorf("jwt_secret must be set")
	}

	if config.JWTExpiry == "" {
		return fmt.Errorf("jwt_expiry must be set")
	}

	if config.Server.APIPort == "" {
		return fmt.Errorf("server.api_port must be set")
	}

	if config.Server.LoadgenPort == "" {
		return fmt.Errorf("server.loadgen_port must be set")
	}

	// Add more validation rules as needed

	return nil
}
