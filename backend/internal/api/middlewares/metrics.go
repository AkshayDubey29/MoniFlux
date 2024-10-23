// backend/internal/api/middlewares/metrics.go

package middlewares

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics defines the Prometheus metrics to be collected
type Metrics struct {
	requestsTotal    *prometheus.CounterVec
	responseDuration *prometheus.HistogramVec
}

// NewMetrics initializes and registers Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "moniflux_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"path", "method", "status"},
		),
		responseDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "moniflux_response_duration_seconds",
				Help:    "Histogram of response latencies for HTTP requests.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"path", "method"},
		),
	}

	// Register metrics with Prometheus
	prometheus.MustRegister(m.requestsTotal)
	prometheus.MustRegister(m.responseDuration)

	return m
}

// MetricsMiddleware collects Prometheus metrics for each HTTP request
func (m *Metrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		timer := prometheus.NewTimer(m.responseDuration.WithLabelValues(r.URL.Path, r.Method))
		defer timer.ObserveDuration()

		// Capture response status
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		// Increment request counter
		m.requestsTotal.WithLabelValues(r.URL.Path, r.Method, http.StatusText(rec.status)).Inc()
	})
}

// ExposeMetricsHandler returns the Prometheus metrics handler
func ExposeMetricsHandler() http.Handler {
	return promhttp.Handler()
}
