// backend/internal/services/monitoring/service.go

package monitoring

import (
	"context"
	"errors"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/v1"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/db/mongo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/mongo"
)

// MonitoringService handles metrics collection and health checks.
type MonitoringService struct {
	config         *v1.Config
	logger         *logrus.Logger
	mongoClient    *mongo.Client
	healthCheckCol *mongo.Collection

	// Prometheus metrics
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	errorCounter    *prometheus.CounterVec
}

// NewMonitoringService creates a new instance of MonitoringService.
func NewMonitoringService(cfg *v1.Config, logger *logrus.Logger, mongoClient *mongo.MongoClient) *MonitoringService {
	healthCol := mongoClient.GetCollection("health_checks")

	// Initialize Prometheus metrics
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moniflux_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "moniflux_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	errorCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moniflux_http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "endpoint", "error"},
	)

	// Register Prometheus metrics
	prometheus.MustRegister(requestCounter, requestDuration, errorCounter)

	return &MonitoringService{
		config:          cfg,
		logger:          logger,
		mongoClient:     mongoClient,
		healthCheckCol:  healthCol,
		requestCounter:  requestCounter,
		requestDuration: requestDuration,
		errorCounter:    errorCounter,
	}
}

// RecordRequest records an HTTP request metric.
func (ms *MonitoringService) RecordRequest(method, endpoint, status string, duration time.Duration) {
	ms.requestCounter.WithLabelValues(method, endpoint, status).Inc()
	ms.requestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordError records an HTTP error metric.
func (ms *MonitoringService) RecordError(method, endpoint, errMsg string) {
	ms.errorCounter.WithLabelValues(method, endpoint, errMsg).Inc()
}

// PerformHealthCheck performs a health check for a given service.
func (ms *MonitoringService) PerformHealthCheck(ctx context.Context, serviceName string, checkFunc func() error) error {
	// Remove unused variable or use it appropriately
	// startTime := time.Now()

	status := "healthy"
	details := ""

	err := checkFunc()
	if err != nil {
		status = "unhealthy"
		details = err.Error()
		ms.logger.Errorf("Health check failed for %s: %v", serviceName, err)
	} else {
		ms.logger.Infof("Health check passed for %s", serviceName)
	}

	healthCheck := &HealthCheck{
		ServiceName: serviceName,
		Status:      status,
		CheckedAt:   time.Now(),
		Details:     details,
	}

	_, err = ms.healthCheckCol.InsertOne(ctx, healthCheck)
	if err != nil {
		ms.logger.Errorf("Failed to record health check for %s: %v", serviceName, err)
		return errors.New("internal server error")
	}

	return err
}

// SetupPrometheusHandler sets up the Prometheus HTTP handler.
func (ms *MonitoringService) SetupPrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// Example of a periodic health check runner
func (ms *MonitoringService) StartHealthCheckScheduler(ctx context.Context, interval time.Duration, services map[string]func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				ms.logger.Info("Stopping health check scheduler")
				return
			case <-ticker.C:
				for serviceName, checkFunc := range services {
					if err := ms.PerformHealthCheck(ctx, serviceName, checkFunc); err != nil {
						ms.logger.Errorf("Health check error for %s: %v", serviceName, err)
					}
				}
			}
		}
	}()
}
