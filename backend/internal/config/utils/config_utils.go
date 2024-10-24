package utils

import (
	"fmt"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// SetupRateLimiter initializes a rate limiter based on the provided configuration.
// It logs warnings and sets default values if the configuration is invalid or missing.
func SetupRateLimiter(cfg *common.Config, logger *logrus.Logger) *rate.Limiter {
	// Validate RateLimit configuration
	if cfg.RateLimit.RequestsPerMinute <= 0 {
		logger.Warn("RateLimit.RequestsPerMinute is not set or invalid, defaulting to 100")
		cfg.RateLimit.RequestsPerMinute = 100
	}

	if cfg.RateLimit.Burst <= 0 {
		logger.Warn("RateLimit.Burst is not set or invalid, defaulting to 20")
		cfg.RateLimit.Burst = 20
	}

	// Initialize the rate limiter
	limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit.RequestsPerMinute), cfg.RateLimit.Burst)
	logger.Infof("Rate limiter set to %d requests per minute with burst %d", cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.Burst)

	return limiter
}

// SetupSecurityRateLimiter initializes a security-specific rate limiter based on the provided configuration.
// It logs warnings and sets default values if the configuration is invalid or missing.
func SetupSecurityRateLimiter(cfg *common.Config, logger *logrus.Logger) *rate.Limiter {
	// Validate SecurityRateLimit configuration
	if cfg.SecurityRateLimit.RequestsPerMinute <= 0 {
		logger.Warn("SecurityRateLimit.RequestsPerMinute is not set or invalid, defaulting to 100")
		cfg.SecurityRateLimit.RequestsPerMinute = 100
	}

	if cfg.SecurityRateLimit.Burst <= 0 {
		logger.Warn("SecurityRateLimit.Burst is not set or invalid, defaulting to 20")
		cfg.SecurityRateLimit.Burst = 20
	}

	// Initialize the security rate limiter
	limiter := rate.NewLimiter(rate.Limit(cfg.SecurityRateLimit.RequestsPerMinute), cfg.SecurityRateLimit.Burst)
	logger.Infof("Security rate limiter set to %d requests per minute with burst %d", cfg.SecurityRateLimit.RequestsPerMinute, cfg.SecurityRateLimit.Burst)

	return limiter
}

// ValidateConfig performs additional validation on the loaded configuration.
// It ensures that essential configurations are set correctly.
func ValidateConfig(cfg *common.Config) error {
	// Validate JWT Secret and Expiry
	if cfg.JWTSecret == "" {
		return fmt.Errorf("jwt_secret must be set")
	}

	if cfg.JWTExpiry == "" {
		return fmt.Errorf("jwt_expiry must be set")
	}

	// Validate API Port
	if cfg.Server.APIPort == "" {
		return fmt.Errorf("api_port must be set")
	}

	// Validate Loadgen URL
	if cfg.Server.LoadgenURL == "" {
		return fmt.Errorf("loadgen_url must be set")
	}

	// Additional validation rules can be added here as needed
	return nil
}

// ParseDuration parses the JWTExpiry string into a time.Duration.
// It returns an error if the parsing fails.
func ParseDuration(expiryStr string) (time.Duration, error) {
	duration, err := time.ParseDuration(expiryStr)
	if err != nil {
		return 0, fmt.Errorf("invalid JWTExpiry format: %w", err)
	}
	return duration, nil
}
