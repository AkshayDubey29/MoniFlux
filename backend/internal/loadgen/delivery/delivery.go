// backend/internal/loadgen/delivery/delivery.go

package delivery

import (
	"context"
	"fmt"
	"sync"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/sirupsen/logrus"
)

// DeliveryService handles the delivery of logs, metrics, and traces to configured destinations.
type DeliveryService struct {
	logger       *logrus.Logger
	handlers     []DestinationHandler
	handlerMutex sync.Mutex
}

// NewDeliveryService initializes a new DeliveryService with appropriate handlers based on destinations.
func NewDeliveryService(logger *logrus.Logger, destinations []common.Destination) (*DeliveryService, error) {
	handlers := make([]DestinationHandler, 0, len(destinations))

	for _, dest := range destinations {
		switch dest.Type {
		case "http":
			handler := NewHTTPDestinationHandler(dest, logger)
			handlers = append(handlers, handler)
		case "file":
			handler, err := NewFileDestinationHandler(dest, logger)
			if err != nil {
				logger.Errorf("Failed to initialize FileDestinationHandler for destination %s: %v", dest.Name, err)
				continue // Skip this destination and proceed with others
			}
			handlers = append(handlers, handler)
		default:
			logger.Errorf("Unsupported destination type: %s", dest.Type)
			continue // Skip unsupported destination types
		}
	}

	if len(handlers) == 0 {
		logger.Warn("No valid destinations configured for delivery")
	}

	return &DeliveryService{
		logger:   logger,
		handlers: handlers,
	}, nil
}

// SendLogs sends a batch of log entries to all configured destinations.
func (ds *DeliveryService) SendLogs(ctx context.Context, logs []models.LogEntry) error {
	ds.handlerMutex.Lock()
	defer ds.handlerMutex.Unlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(ds.handlers))

	for _, handler := range ds.handlers {
		wg.Add(1)
		go func(h DestinationHandler) {
			defer wg.Done()
			if err := h.SendLog(logs); err != nil {
				errChan <- fmt.Errorf("failed to send logs: %w", err)
			}
		}(handler)
	}

	wg.Wait()
	close(errChan)

	// Aggregate errors
	var finalErr error
	for err := range errChan {
		ds.logger.Error(err)
		finalErr = err // Overwrite with the last error
	}

	return finalErr
}

// SendMetrics sends a batch of metric entries to all configured destinations.
func (ds *DeliveryService) SendMetrics(ctx context.Context, metrics []models.Metric) error {
	ds.handlerMutex.Lock()
	defer ds.handlerMutex.Unlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(ds.handlers))

	for _, handler := range ds.handlers {
		wg.Add(1)
		go func(h DestinationHandler) {
			defer wg.Done()
			if err := h.SendMetric(metrics); err != nil {
				errChan <- fmt.Errorf("failed to send metrics: %w", err)
			}
		}(handler)
	}

	wg.Wait()
	close(errChan)

	// Aggregate errors
	var finalErr error
	for err := range errChan {
		ds.logger.Error(err)
		finalErr = err // Overwrite with the last error
	}

	return finalErr
}

// SendTraces sends a batch of trace entries to all configured destinations.
func (ds *DeliveryService) SendTraces(ctx context.Context, traces []models.Trace) error {
	ds.handlerMutex.Lock()
	defer ds.handlerMutex.Unlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(ds.handlers))

	for _, handler := range ds.handlers {
		wg.Add(1)
		go func(h DestinationHandler) {
			defer wg.Done()
			if err := h.SendTrace(traces); err != nil {
				errChan <- fmt.Errorf("failed to send traces: %w", err)
			}
		}(handler)
	}

	wg.Wait()
	close(errChan)

	// Aggregate errors
	var finalErr error
	for err := range errChan {
		ds.logger.Error(err)
		finalErr = err // Overwrite with the last error
	}

	return finalErr
}

// Close gracefully shuts down all destination handlers.
func (ds *DeliveryService) Close() error {
	ds.handlerMutex.Lock()
	defer ds.handlerMutex.Unlock()

	var finalErr error

	for _, handler := range ds.handlers {
		if err := handler.Close(); err != nil {
			ds.logger.Errorf("Failed to close handler: %v", err)
			finalErr = err
		}
	}

	return finalErr
}
