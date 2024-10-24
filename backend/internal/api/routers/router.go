package routers

import (
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/handlers"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/middlewares"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/controllers"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/services/authentication"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// SetupRouter initializes the router with all necessary routes and middlewares.
// Parameters:
// - logger: Instance of logrus.Logger for logging purposes.
// - controller: Instance of LoadGenController to handle business logic.
// - authService: Instance of AuthenticationService to handle authentication.
// - config: Application configuration containing settings for middlewares.
func SetupRouter(logger *logrus.Logger, controller *controllers.LoadGenController, authService *authentication.AuthenticationService, config *common.Config) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// Initialize middlewares
	requestIDMiddleware := middlewares.RequestIDMiddleware
	recoveryMiddleware := middlewares.RecoveryMiddleware(logger)
	loggingMiddleware := middlewares.LoggingMiddleware(logger)
	// Initialize AuthMiddleware with AuthenticationService and logger
	authMiddleware := middlewares.NewAuthMiddleware(authService, logger).MiddlewareFunc
	// Initialize CORSMiddleware with AllowedOrigins and logger
	corsMiddleware := middlewares.CORSMiddleware(config.AllowedOrigins, logger)

	// Setup Rate Limiter
	rateLimitInterval := rate.Every(time.Minute / time.Duration(config.RateLimit.RequestsPerMinute))
	rateLimiter := middlewares.NewRateLimiter(rateLimitInterval, config.RateLimit.Burst, logger)
	rateLimitMiddleware := middlewares.RateLimitMiddleware(rateLimiter)

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
	// 7. Metrics
	router.Use(recoveryMiddleware)
	router.Use(loggingMiddleware)
	router.Use(requestIDMiddleware)
	router.Use(securityHeadersMiddleware)
	router.Use(corsMiddleware)
	router.Use(rateLimitMiddleware)
	router.Use(metricsMiddleware)

	// Apply authentication middleware to all routes except /health
	apiRouter := router.PathPrefix("/").Subrouter()
	apiRouter.Use(authMiddleware)

	// Initialize handlers with dependencies
	h := handlers.NewHandler(controller, authService, logger)

	// Define API routes with their respective handlers
	apiRouter.HandleFunc("/start-test", h.StartTest).Methods("POST")
	logger.Infof("Registered POST /start-test endpoint")

	apiRouter.HandleFunc("/schedule-test", h.ScheduleTest).Methods("POST")
	logger.Infof("Registered POST /schedule-test endpoint")

	apiRouter.HandleFunc("/create-test", h.CreateTest).Methods("POST")
	logger.Infof("Registered POST /create-test endpoint")

	apiRouter.HandleFunc("/cancel-test", h.CancelTest).Methods("POST")
	logger.Infof("Registered POST /cancel-test endpoint")

	apiRouter.HandleFunc("/restart-test", h.RestartTest).Methods("POST")
	logger.Infof("Registered POST /restart-test endpoint")

	apiRouter.HandleFunc("/save-results", h.SaveResults).Methods("POST")
	logger.Infof("Registered POST /save-results endpoint")

	apiRouter.HandleFunc("/get-all-tests", h.GetAllTests).Methods("GET")
	logger.Infof("Registered GET /get-all-tests endpoint")

	// User registration endpoint
	router.HandleFunc("/register", h.RegisterUser).Methods("POST")
	logger.Infof("Registered POST /register endpoint")

	// Authentication endpoint
	router.HandleFunc("/authenticate", h.AuthenticateUser).Methods("POST")
	logger.Infof("Registered POST /authenticate endpoint")

	// Health Check Endpoint (Unprotected)
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	logger.Infof("Registered GET /health endpoint")

	// Metrics Endpoint (Protected or Unprotected based on your needs)
	router.Handle("/metrics", metrics.ExposeMetricsHandler()).Methods("GET")
	logger.Infof("Registered GET /metrics endpoint")

	return router
}
