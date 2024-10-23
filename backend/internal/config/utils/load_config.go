// backend/internal/config/utils/load_config.go

package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/common" // Updated import
)

// LoadConfig loads the application configuration from the specified path.
// It supports reading from configuration files, environment variables, and setting default values.
// The function returns a pointer to the Config struct and an error if the loading fails.
func LoadConfig(path string) (*common.Config, error) { // Changed to common.Config
	// Initialize Viper
	v := viper.New()

	// Set the file name and path
	v.SetConfigFile(path)

	// Set the file type (auto-detected by Viper)
	// If the file extension is not provided in 'path', specify it here
	// v.SetConfigType("yaml")

	// Set default values
	setDefaults(v)

	// Enable environment variables reading
	v.AutomaticEnv()

	// Set environment variable prefix to avoid conflicts
	v.SetEnvPrefix("MONIFLUX")

	// Replace dots in env variables with underscores
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read the configuration file
	if err := v.ReadInConfig(); err != nil {
		// If the config file is not found, proceed with environment variables and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal the configuration into the Config struct
	var config common.Config // Changed to common.Config
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
	// Example default settings; adjust as necessary
	v.SetDefault("api_port", "8080")
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "json")
	v.SetDefault("log_output", "stdout")
	v.SetDefault("mongo_uri", "mongodb://mongodb:27017")
	v.SetDefault("mongo_db", "moniflux")
	v.SetDefault("jwt_secret", "default-jwt-secret")
	v.SetDefault("jwt_expiry", "24h")
	v.SetDefault("allowed_origins", []string{"http://localhost:3000"})
	v.SetDefault("rate_limit.requests_per_minute", 100) // Changed field name
	v.SetDefault("rate_limit.burst", 20)
	v.SetDefault("security.rate_limiting.requests_per_minute", 100) // Changed field name
	v.SetDefault("security.rate_limiting.burst", 20)
	v.SetDefault("metrics.prometheus_enabled", true)
	v.SetDefault("metrics.prometheus_endpoint", "/metrics")
	v.SetDefault("metrics.prometheus_port", 9090)
	v.SetDefault("enable_tls", false)
	v.SetDefault("tls_cert_path", "")
	v.SetDefault("tls_key_path", "")
	v.SetDefault("destinations", []common.Destination{ // Changed to common.Destination
		{Endpoint: "localhost", Port: 8081},
		{Endpoint: "remote-server.com", Port: 8082},
	})
	v.SetDefault("log_rate", 10)
	v.SetDefault("metrics_rate", 5)
	v.SetDefault("trace_rate", 2)
	v.SetDefault("log_size", 512)
	v.SetDefault("metrics_value", 100.0)
	v.SetDefault("default_roles", []string{"admin", "editor", "viewer"})
	v.SetDefault("monitoring.health_check_interval", "5m")
}

// validateConfig performs validation on the loaded configuration.
// It ensures that essential configurations are set correctly.
func validateConfig(config *common.Config) error { // Changed to common.Config
	// Example validation; extend as necessary
	if config.JWTSecret == "" {
		return fmt.Errorf("jwt_secret must be set")
	}

	if config.JWTExpiry == "" {
		return fmt.Errorf("jwt_expiry must be set")
	}

	if config.APIPort == "" {
		return fmt.Errorf("api_port must be set")
	}

	// Add more validation rules as needed

	return nil
}
