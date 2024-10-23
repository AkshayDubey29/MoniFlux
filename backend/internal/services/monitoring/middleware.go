// backend/internal/services/monitoring/middleware.go

package monitoring

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// MonitoringMiddleware provides HTTP middleware for logging and metrics.
type MonitoringMiddleware struct {
	monitoringService *MonitoringService
	logger            *logrus.Logger
}

// NewMonitoringMiddleware creates a new instance of MonitoringMiddleware.
func NewMonitoringMiddleware(ms *MonitoringService, logger *logrus.Logger) *MonitoringMiddleware {
	return &MonitoringMiddleware{
		monitoringService: ms,
		logger:            logger,
	}
}

// MiddlewareFunc is the HTTP middleware function that logs requests and records metrics.
func (mm *MonitoringMiddleware) MiddlewareFunc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Use a ResponseWriter wrapper to capture the status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(startTime)
		method := r.Method
		endpoint := r.URL.Path
		status := http.StatusText(rw.statusCode)

		// Record the request metrics
		mm.monitoringService.RecordRequest(method, endpoint, status, duration)

		// Log the request details
		mm.logger.WithFields(logrus.Fields{
			"method":   method,
			"endpoint": endpoint,
			"status":   rw.statusCode,
			"duration": duration.Seconds(),
		}).Info("Handled HTTP request")
	})
}

// responseWriter is a wrapper around http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code for logging and metrics.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
