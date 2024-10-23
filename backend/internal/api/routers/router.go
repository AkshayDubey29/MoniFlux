// backend/internal/api/routers/router.go

package routers

import (
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/handlers"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/middlewares"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/controllers"
	"github.com/gorilla/mux"
)

// SetupRouter initializes the router with all necessary routes and middlewares.
// Parameters:
// - logger: Instance of logrus.Logger for logging purposes.
// - controller: Instance of LoadGenController to handle business logic.
// - config: Application configuration containing settings for middlewares.
func SetupRouter(logger *logrus.Logger, controller *controllers.LoadGenController, config *common.Config) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Initialize middlewares
	requestIDMiddleware := middlewares.RequestIDMiddleware
	recoveryMiddleware := middlewares.RecoveryMiddleware(logger)
	loggingMiddleware := middlewares.LoggingMiddleware(logger)
	authMiddleware := middlewares.AuthMiddleware(config.JWTSecret, logger)
	corsMiddleware := middlewares.CORSMiddleware(config.AllowedOrigins, logger)

	// Setup Rate Limiter
	// rate.Every defines the interval between events, so we calculate it based on RequestsPerMinute
	// For example, 60 requests per minute => 1 request per second
	rateLimit := rate.Every(time.Minute / time.Duration(config.RateLimit.RequestsPerMinute))
	limiter := rate.NewLimiter(rateLimit, config.RateLimit.Burst)
	rateLimitMiddleware := middlewares.RateLimitMiddleware(limiter, logger)

	// Initialize Metrics Middleware
	metrics := middlewares.NewMetrics()
	metricsMiddleware := metrics.MetricsMiddleware

	// Initialize Security Headers Middleware
	securityHeadersMiddleware := middlewares.SecurityHeadersMiddleware

	// Apply global middlewares in the order of:
	// 1. Recovery (to catch panics)
	// 2. Logging
	// 3. Request ID
	// 4. Security Headers
	// 5. CORS
	// 6. Rate Limiting
	// 7. Authentication
	// 8. Metrics
	router.Use(recoveryMiddleware)
	router.Use(loggingMiddleware)
	router.Use(requestIDMiddleware)
	router.Use(securityHeadersMiddleware)
	router.Use(corsMiddleware)
	router.Use(rateLimitMiddleware)
	router.Use(authMiddleware)
	router.Use(metricsMiddleware)

	// Initialize handlers with dependencies
	h := handlers.NewHandler(controller, logger)

	// Define API routes with their respective handlers
	router.HandleFunc("/start-test", h.StartTest).Methods("POST")
	router.HandleFunc("/schedule-test", h.ScheduleTest).Methods("POST")
	router.HandleFunc("/cancel-test", h.CancelTest).Methods("POST")
	router.HandleFunc("/restart-test", h.RestartTest).Methods("POST")
	router.HandleFunc("/save-results", h.SaveResults).Methods("POST")
	router.HandleFunc("/get-all-tests", h.GetAllTests).Methods("GET")

	// Health Check Endpoint (Unprotected)
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Metrics Endpoint (Protected or Unprotected based on your needs)
	router.Handle("/metrics", metrics.ExposeMetricsHandler()).Methods("GET")

	return router
}