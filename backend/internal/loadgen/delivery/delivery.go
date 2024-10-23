// backend/internal/loadgen/delivery/delivery.go

package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models" // Added import for models
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/sirupsen/logrus"
)

// DeliveryService handles the delivery of logs, metrics, and traces to configured destinations.
type DeliveryService struct {
	logger       *logrus.Logger
	destinations []common.Destination
	client       *http.Client
}

// NewDeliveryService initializes a new DeliveryService.
// Parameters:
// - logger: Instance of logrus.Logger for logging purposes.
// - destinations: Slice of Destination structs defining where to send the payloads.
func NewDeliveryService(logger *logrus.Logger, destinations []common.Destination) *DeliveryService {
	return &DeliveryService{
		logger:       logger,
		destinations: destinations,
		client: &http.Client{
			Timeout: 10 * time.Second, // Set a timeout for HTTP requests
		},
	}
}

// SendLogs sends a batch of log entries to all configured destinations.
// Parameters:
// - ctx: Context for managing request deadlines and cancellations.
// - logs: Slice of LogEntry structs to be sent.
func (ds *DeliveryService) SendLogs(ctx context.Context, logs []models.LogEntry) error {
	return ds.deliverPayload(ctx, logs, "logs")
}

// SendMetrics sends a batch of metric entries to all configured destinations.
// Parameters:
// - ctx: Context for managing request deadlines and cancellations.
// - metrics: Slice of Metric structs to be sent.
func (ds *DeliveryService) SendMetrics(ctx context.Context, metrics []models.Metric) error {
	return ds.deliverPayload(ctx, metrics, "metrics")
}

// SendTraces sends a batch of trace entries to all configured destinations.
// Parameters:
// - ctx: Context for managing request deadlines and cancellations.
// - traces: Slice of Trace structs to be sent.
func (ds *DeliveryService) SendTraces(ctx context.Context, traces []models.Trace) error {
	return ds.deliverPayload(ctx, traces, "traces")
}

// deliverPayload is a generic method to send any payload type to all destinations.
// Parameters:
// - ctx: Context for managing request deadlines and cancellations.
// - payload: The data to be sent, which should be serializable to JSON.
// - payloadType: A string indicating the type of payload (e.g., "logs", "metrics", "traces").
// Returns:
// - error: An error if any delivery fails.
func (ds *DeliveryService) deliverPayload(ctx context.Context, payload interface{}, payloadType string) error {
	if len(ds.destinations) == 0 {
		ds.logger.Warn("No destinations configured for delivery")
		return nil
	}

	// Serialize the payload to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		ds.logger.Errorf("Failed to marshal %s payload: %v", payloadType, err)
		return err
	}

	// Iterate over each destination and send the payload
	for _, dest := range ds.destinations {
		// Construct the full URL
		url := fmt.Sprintf("http://%s:%d/%s", dest.Endpoint, dest.Port, payloadType)

		// Create a new HTTP POST request with context
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(data))
		if err != nil {
			ds.logger.Errorf("Failed to create HTTP request for %s: %v", payloadType, err)
			continue // Proceed to the next destination
		}

		// Set appropriate headers
		req.Header.Set("Content-Type", "application/json")

		// Send the HTTP request
		resp, err := ds.client.Do(req)
		if err != nil {
			ds.logger.Errorf("Failed to send %s to %s:%d - %v", payloadType, dest.Endpoint, dest.Port, err)
			continue // Proceed to the next destination
		}

		// Ensure the response body is closed
		resp.Body.Close()

		// Check for non-success status codes
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			ds.logger.Errorf("Received non-success status code %d when sending %s to %s:%d", resp.StatusCode, payloadType, dest.Endpoint, dest.Port)
			continue // Proceed to the next destination
		}

		ds.logger.Infof("Successfully sent %s to %s:%d", payloadType, dest.Endpoint, dest.Port)
	}

	return nil
}

// Example usage of SendLogs function
/*
func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	destinations := []common.Destination{
		{
			Port:     8081,
			Endpoint: "localhost",
		},
		{
			Port:     8082,
			Endpoint: "remote-server.com",
		},
	}

	deliveryService := NewDeliveryService(logger, destinations)

	logs := []models.LogEntry{
		{
			TestID:    "test123",
			Timestamp: time.Now(),
			Message:   "Sample log message",
			Level:     "INFO",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := deliveryService.SendLogs(ctx, logs)
	if err != nil {
		logger.Errorf("Error sending logs: %v", err)
	}
}
*/

// Additional functions and methods can be implemented as needed, such as retry mechanisms, batching, etc.
