// backend/internal/loadgen/generators/generator.go

package generators

import (
	"math/rand"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/sirupsen/logrus"
)

// GeneratorService handles the generation of logs, metrics, and traces.
type GeneratorService struct {
	logger       *logrus.Logger
	logRate      int     // Logs per second
	metricsRate  int     // Metrics per second
	traceRate    int     // Traces per second
	logSize      int     // Approximate size of each log entry (in bytes)
	metricsValue float64 // Value of the generated metrics
}

// NewGeneratorService creates a new GeneratorService instance with configured rates and sizes.
func NewGeneratorService(logger *logrus.Logger, logRate, metricsRate, traceRate, logSize int, metricsValue float64) *GeneratorService {
	return &GeneratorService{
		logger:       logger,
		logRate:      logRate,
		metricsRate:  metricsRate,
		traceRate:    traceRate,
		logSize:      logSize,
		metricsValue: metricsValue,
	}
}

// GenerateLogs simulates log generation at the configured rate and size.
func (gs *GeneratorService) GenerateLogs(testID string, duration time.Duration) []models.LogEntry {
	gs.logger.Infof("Starting log generation for TestID: %s", testID)
	var logs []models.LogEntry
	ticker := time.NewTicker(time.Second / time.Duration(gs.logRate))
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ticker.C:
			if time.Since(startTime) > duration {
				gs.logger.Infof("Log generation complete for TestID: %s", testID)
				return logs
			}
			log := models.LogEntry{
				TestID:    testID,
				Timestamp: time.Now(),
				Message:   gs.generateRandomLogMessage(),
				Level:     "INFO", // Example log level
			}
			logs = append(logs, log)
			gs.logger.Debugf("Generated log: %+v", log)
		}
	}
}

// GenerateMetrics simulates metric generation at the configured rate.
func (gs *GeneratorService) GenerateMetrics(testID string, duration time.Duration) []models.Metric {
	gs.logger.Infof("Starting metric generation for TestID: %s", testID)
	var metrics []models.Metric
	ticker := time.NewTicker(time.Second / time.Duration(gs.metricsRate))
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ticker.C:
			if time.Since(startTime) > duration {
				gs.logger.Infof("Metric generation complete for TestID: %s", testID)
				return metrics
			}
			metric := models.Metric{
				TestID:    testID,
				Timestamp: time.Now(),
				Value:     gs.generateRandomMetricValue(),
			}
			metrics = append(metrics, metric)
			gs.logger.Debugf("Generated metric: %+v", metric)
		}
	}
}

// GenerateTraces simulates trace generation at the configured rate.
func (gs *GeneratorService) GenerateTraces(testID string, duration time.Duration) []models.Trace {
	gs.logger.Infof("Starting trace generation for TestID: %s", testID)
	var traces []models.Trace
	ticker := time.NewTicker(time.Second / time.Duration(gs.traceRate))
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ticker.C:
			if time.Since(startTime) > duration {
				gs.logger.Infof("Trace generation complete for TestID: %s", testID)
				return traces
			}
			trace := models.Trace{
				TestID:    testID,
				Timestamp: time.Now(),
				TraceID:   gs.generateRandomTraceID(),
				SpanID:    gs.generateRandomSpanID(),
				Operation: "operation_example", // Example operation name
				Duration:  rand.Intn(1000),     // Simulated trace duration in ms
			}
			traces = append(traces, trace)
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

// randomString generates a random alphanumeric string of the specified length.
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
