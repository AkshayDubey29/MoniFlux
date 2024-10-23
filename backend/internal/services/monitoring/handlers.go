// backend/internal/services/monitoring/handlers.go

package monitoring

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MonitoringHandlers encapsulates handlers related to monitoring.
type MonitoringHandlers struct {
	monitoringService *MonitoringService
	logger            *logrus.Logger
}

// NewMonitoringHandlers creates a new instance of MonitoringHandlers.
func NewMonitoringHandlers(ms *MonitoringService, logger *logrus.Logger) *MonitoringHandlers {
	return &MonitoringHandlers{
		monitoringService: ms,
		logger:            logger,
	}
}

// GetHealthCheckHistoryHandler retrieves the health check history for a specific service.
func (mh *MonitoringHandlers) GetHealthCheckHistoryHandler(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" {
		http.Error(w, "Missing 'service' query parameter", http.StatusBadRequest)
		return
	}

	// Define the time range for the history (e.g., last 24 hours)
	since := time.Now().Add(-24 * time.Hour)

	// Query the database for health checks of the specified service since the defined time
	filter := bson.M{
		"service_name": serviceName,
		"checked_at": bson.M{
			"$gte": since,
		},
	}

	opts := options.Find().SetSort(bson.D{{"checked_at", -1}}).SetLimit(100) // Limit to last 100 records

	cursor, err := mh.monitoringService.healthCheckCol.Find(r.Context(), filter, opts)
	if err != nil {
		mh.logger.Errorf("Error fetching health check history: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var healthChecks []HealthCheck
	for cursor.Next(r.Context()) {
		var hc HealthCheck
		if err := cursor.Decode(&hc); err != nil {
			mh.logger.Errorf("Error decoding health check record: %v", err)
			continue
		}
		healthChecks = append(healthChecks, hc)
	}

	if err := cursor.Err(); err != nil {
		mh.logger.Errorf("Cursor error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Respond with the health check history
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(healthChecks)
}

// HealthCheckStatusHandler provides a simple health check endpoint for the monitoring service itself.
func (mh *MonitoringHandlers) HealthCheckStatusHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
