package middlewares // backend/internal/api/middlewares/logging.go

import (
	"time"

	"github.com/sirupsen/logrus"
	"net/http"
)

// LoggingMiddleware logs each incoming HTTP request and its corresponding response.
func LoggingMiddleware(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Capture response details
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			duration := time.Since(startTime)

			logger.WithFields(logrus.Fields{
				"method":       r.Method,
				"path":         r.URL.Path,
				"status":       rec.status,
				"duration_ms":  duration.Milliseconds(),
				"remote_addr":  r.RemoteAddr,
				"user_agent":   r.UserAgent(),
				"request_time": startTime.Format(time.RFC3339),
			}).Info("Handled request")
		})
	}
}

// statusRecorder is a wrapper to capture the HTTP status code
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code
func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
