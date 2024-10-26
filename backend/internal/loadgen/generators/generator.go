// backend/internal/loadgen/generators/generator.go

package generators

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/loadgen/delivery"
	"github.com/sirupsen/logrus"
)

// GeneratorService handles the generation of logs, metrics, and traces.
type GeneratorService struct {
	logger          *logrus.Logger
	deliveryService *delivery.DeliveryService
	logRate         int     // Logs per second
	metricsRate     int     // Metrics per second
	traceRate       int     // Traces per second
	logSize         int     // Approximate size of each log entry (in bytes)
	metricsValue    float64 // Base value of the generated metrics

	wg         sync.WaitGroup
	cancelFunc context.CancelFunc
}

// NewGeneratorService creates a new GeneratorService instance with configured rates and sizes.
// It initializes the DeliveryService based on the provided destinations.
func NewGeneratorService(logger *logrus.Logger, destinations []models.Destination, logRate, metricsRate, traceRate, logSize int, metricsValue float64) (*GeneratorService, error) {
	// Convert common.Destination to models.Destination if necessary
	modelDestinations := make([]models.Destination, len(destinations))
	copy(modelDestinations, destinations)

	// Initialize the DeliveryService
	deliveryService, err := delivery.NewDeliveryService(logger, destinations)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DeliveryService: %w", err)
	}

	return &GeneratorService{
		logger:          logger,
		deliveryService: deliveryService,
		logRate:         logRate,
		metricsRate:     metricsRate,
		traceRate:       traceRate,
		logSize:         logSize,
		metricsValue:    metricsValue,
	}, nil
}

// StartGenerating starts the log, metric, and trace generation concurrently.
// It runs for the specified duration or until the context is canceled.
func (gs *GeneratorService) StartGenerating(ctx context.Context, test models.Test) {
	// Create a cancellable context
	ctx, gs.cancelFunc = context.WithCancel(ctx)
	gs.wg.Add(3)

	// Start generating logs, metrics, and traces
	go gs.generateLogs(ctx, test.TestID, time.Duration(test.Duration)*time.Second)
	go gs.generateMetrics(ctx, test.TestID, time.Duration(test.Duration)*time.Second)
	go gs.generateTraces(ctx, test.TestID, time.Duration(test.Duration)*time.Second)
}

// StopGenerating stops the generation processes gracefully.
func (gs *GeneratorService) StopGenerating() {
	if gs.cancelFunc != nil {
		gs.cancelFunc()
	}
	gs.wg.Wait()
	gs.deliveryService.Close()
}

// generateLogs simulates log generation at the configured rate and size.
// Logs are sent directly to the DeliveryService.
func (gs *GeneratorService) generateLogs(ctx context.Context, testID string, duration time.Duration) {
	defer gs.wg.Done()
	ticker := time.NewTicker(time.Second / time.Duration(gs.logRate))
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			gs.logger.Infof("Log generation cancelled for TestID: %s", testID)
			return
		case <-ticker.C:
			if time.Since(startTime) > duration {
				gs.logger.Infof("Log generation complete for TestID: %s", testID)
				return
			}
			log := models.LogEntry{
				TestID:    testID,
				Timestamp: time.Now(),
				Message:   gs.generateRandomLogMessage(),
				Level:     gs.randomLogLevel(),
			}
			if err := gs.deliveryService.SendLogs(ctx, []models.LogEntry{log}); err != nil {
				gs.logger.Errorf("Failed to send log: %v", err)
			}
			gs.logger.Debugf("Generated log: %+v", log)
		}
	}
}

// generateMetrics simulates metric generation at the configured rate.
// Metrics are sent directly to the DeliveryService.
func (gs *GeneratorService) generateMetrics(ctx context.Context, testID string, duration time.Duration) {
	defer gs.wg.Done()
	ticker := time.NewTicker(time.Second / time.Duration(gs.metricsRate))
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			gs.logger.Infof("Metric generation cancelled for TestID: %s", testID)
			return
		case <-ticker.C:
			if time.Since(startTime) > duration {
				gs.logger.Infof("Metric generation complete for TestID: %s", testID)
				return
			}
			metric := models.Metric{
				TestID:    testID,
				Timestamp: time.Now(),
				Value:     gs.generateRandomMetricValue(),
			}
			if err := gs.deliveryService.SendMetrics(ctx, []models.Metric{metric}); err != nil {
				gs.logger.Errorf("Failed to send metric: %v", err)
			}
			gs.logger.Debugf("Generated metric: %+v", metric)
		}
	}
}

// generateTraces simulates trace generation at the configured rate.
// Traces are sent directly to the DeliveryService.
func (gs *GeneratorService) generateTraces(ctx context.Context, testID string, duration time.Duration) {
	defer gs.wg.Done()
	ticker := time.NewTicker(time.Second / time.Duration(gs.traceRate))
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			gs.logger.Infof("Trace generation cancelled for TestID: %s", testID)
			return
		case <-ticker.C:
			if time.Since(startTime) > duration {
				gs.logger.Infof("Trace generation complete for TestID: %s", testID)
				return
			}
			trace := models.Trace{
				TestID:    testID,
				Timestamp: time.Now(),
				TraceID:   gs.generateRandomTraceID(),
				SpanID:    gs.generateRandomSpanID(),
				Operation: gs.randomOperationName(),
				Duration:  gs.generateRandomTraceDuration(),
			}
			if err := gs.deliveryService.SendTraces(ctx, []models.Trace{trace}); err != nil {
				gs.logger.Errorf("Failed to send trace: %v", err)
			}
			gs.logger.Debugf("Generated trace: %+v", trace)
		}
	}
}

// generateRandomLogMessage creates a mock log message of approximately the configured size.
func (gs *GeneratorService) generateRandomLogMessage() string {
	// Simulate generating a log message with random content
	return randomString(gs.logSize)
}

// generateRandomMetricValue generates a random metric value around the configured value.
func (gs *GeneratorService) generateRandomMetricValue() float64 {
	// Simulate generating a random metric value with some variation
	return gs.metricsValue + rand.Float64()*10 - 5 // Adds random variation between -5 and +5
}

// generateRandomTraceID creates a mock Trace ID for traces.
func (gs *GeneratorService) generateRandomTraceID() string {
	return randomString(16) // Simulate a 16-character Trace ID
}

// generateRandomSpanID creates a mock Span ID for traces.
func (gs *GeneratorService) generateRandomSpanID() string {
	return randomString(8) // Simulate an 8-character Span ID
}

// randomOperationName generates a random operation name for traces.
func (gs *GeneratorService) randomOperationName() string {
	operations := []string{"operation_login", "operation_fetch_data", "operation_logout", "operation_update_profile"}
	return operations[rand.Intn(len(operations))]
}

// generateRandomTraceDuration generates a random trace duration in milliseconds.
func (gs *GeneratorService) generateRandomTraceDuration() int {
	return rand.Intn(1000) + 100 // Duration between 100ms and 1100ms
}

// randomString generates a random alphanumeric string of the specified length.
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
