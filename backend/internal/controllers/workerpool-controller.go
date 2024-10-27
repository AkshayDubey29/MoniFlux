// workerpool-controller.go

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/sirupsen/logrus"
)

// DestinationType defines the type of destination for logs.
type DestinationType string

const (
	FileDestination DestinationType = "file"
	HTTPDestination DestinationType = "http"
)

// WorkerPool manages a pool of workers to process log, metric, and trace entries concurrently.
type WorkerPool struct {
	numWorkers      int
	jobs            chan interface{} // Can accept any type of job entry (logs, metrics, traces)
	wg              sync.WaitGroup
	file            *os.File
	logger          *logrus.Logger
	batchSize       int           // Number of entries per batch
	batchDelay      time.Duration // Maximum delay before flushing a batch
	destinationType DestinationType
	httpEndpoint    string // Used if destinationType is HTTP
	successCount    int64
	failureCount    int64
	mu              sync.Mutex // Protects successCount and failureCount
	shutdownOnce    sync.Once  // Ensures Shutdown is called only once
}

// NewWorkerPool initializes a new WorkerPool with a specified number of workers, destination, batch size, and batch delay.
func NewWorkerPool(numWorkers int, destinationType DestinationType, destinationEndpoint string, logger *logrus.Logger, batchSize int, batchDelay time.Duration) (*WorkerPool, error) {
	var file *os.File
	var err error

	if destinationType == FileDestination {
		// Open the file in append mode with buffering
		filePath := destinationEndpoint
		if filePath == "" {
			return nil, fmt.Errorf("filePath cannot be empty for file destination")
		}
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
	}

	wp := &WorkerPool{
		numWorkers:      numWorkers,
		jobs:            make(chan interface{}, numWorkers*10000), // Increased buffer size
		file:            file,
		logger:          logger,
		batchSize:       batchSize,
		batchDelay:      batchDelay,
		destinationType: destinationType,
		httpEndpoint:    destinationEndpoint,
	}

	wp.start()
	return wp, nil
}

// start initializes the worker goroutines.
func (wp *WorkerPool) start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker processes each job (log, metric, trace) based on the destination type.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	wp.logger.Debugf("Worker %d started", id)

	for job := range wp.jobs {
		switch entry := job.(type) {
		case models.LogEntry:
			if wp.destinationType == FileDestination {
				wp.processLog(entry)
			} else if wp.destinationType == HTTPDestination {
				wp.processLogHTTP(entry)
			}
		case models.Metric:
			if wp.destinationType == FileDestination {
				wp.processMetric(entry)
			} else if wp.destinationType == HTTPDestination {
				wp.processMetricHTTP(entry)
			}
		case models.Trace:
			if wp.destinationType == FileDestination {
				wp.processTrace(entry)
			} else if wp.destinationType == HTTPDestination {
				wp.processTraceHTTP(entry)
			}
		default:
			wp.logger.Errorf("Worker %d: Unknown job type: %T", id, job)
		}
	}
	wp.logger.Debugf("Worker %d stopped", id)
}

// processLog handles LogEntry by writing to a file.
func (wp *WorkerPool) processLog(logEntry models.LogEntry) {
	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		wp.logger.Errorf("Failed to marshal log entry: %v", err)
		return
	}
	if _, err := wp.file.Write(append(jsonData, '\n')); err != nil {
		wp.logger.Errorf("Failed to write log entry to file: %v", err)
	}
}

// processLogHTTP handles LogEntry by sending it to an HTTP endpoint.
func (wp *WorkerPool) processLogHTTP(logEntry models.LogEntry) {
	wp.sendHTTPEntry(logEntry)
}

// processMetric handles Metric entries by writing to a file.
func (wp *WorkerPool) processMetric(metric models.Metric) {
	jsonData, err := json.Marshal(metric)
	if err != nil {
		wp.logger.Errorf("Failed to marshal metric entry: %v", err)
		return
	}
	if _, err := wp.file.Write(append(jsonData, '\n')); err != nil {
		wp.logger.Errorf("Failed to write metric entry to file: %v", err)
	}
}

// processMetricHTTP handles Metric entries by sending them to an HTTP endpoint.
func (wp *WorkerPool) processMetricHTTP(metric models.Metric) {
	wp.sendHTTPEntry(metric)
}

// processTrace handles Trace entries by writing to a file.
func (wp *WorkerPool) processTrace(trace models.Trace) {
	jsonData, err := json.Marshal(trace)
	if err != nil {
		wp.logger.Errorf("Failed to marshal trace entry: %v", err)
		return
	}
	if _, err := wp.file.Write(append(jsonData, '\n')); err != nil {
		wp.logger.Errorf("Failed to write trace entry to file: %v", err)
	}
}

// processTraceHTTP handles Trace entries by sending them to an HTTP endpoint.
func (wp *WorkerPool) processTraceHTTP(trace models.Trace) {
	wp.sendHTTPEntry(trace)
}

// sendHTTPEntry sends any entry (log, metric, trace) to the HTTP endpoint with retry logic.
func (wp *WorkerPool) sendHTTPEntry(entry interface{}) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		wp.logger.Errorf("Failed to marshal entry: %v", err)
		wp.incrementFailure()
		return
	}

	var attempt int
	maxAttempts := 3
	backoff := time.Second

	for attempt = 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest("POST", wp.httpEndpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			wp.logger.Errorf("Attempt %d: Failed to create HTTP request: %v", attempt, err)
			wp.incrementFailure()
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{
			Timeout: 5 * time.Second, // Set a timeout for the HTTP request
		}

		resp, err := client.Do(req)
		if err != nil {
			wp.logger.Errorf("Attempt %d: Failed to send entry to HTTP endpoint: %v", attempt, err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				wp.incrementSuccess()
				return
			} else {
				wp.logger.Errorf("Attempt %d: Received non-success status code %d from HTTP endpoint", attempt, resp.StatusCode)
			}
		}

		// Wait before retrying
		time.Sleep(backoff)
		backoff *= 2 // Exponential backoff
	}

	// After max attempts, log failure
	wp.logger.Errorf("All %d attempts failed to send entry to HTTP endpoint", maxAttempts)
	wp.incrementFailure()
}

// Submit enqueues a log, metric, or trace entry for processing.
func (wp *WorkerPool) Submit(entry interface{}) {
	select {
	case wp.jobs <- entry:
	default:
		wp.logger.Warn("Job channel is full, dropping entry")
	}
}

// Shutdown gracefully shuts down the worker pool and closes the file.
func (wp *WorkerPool) Shutdown() error {
	var err error
	wp.shutdownOnce.Do(func() {
		close(wp.jobs)
		wp.wg.Wait()
		if wp.file != nil {
			err = wp.file.Close()
			if err != nil {
				wp.logger.Errorf("Failed to close log file: %v", err)
			}
		}
	})
	return err
}

// incrementSuccess safely increments the successCount.
func (wp *WorkerPool) incrementSuccess() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.successCount++
}

// incrementFailure safely increments the failureCount.
func (wp *WorkerPool) incrementFailure() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.failureCount++
}

// GetCounts returns the number of successful and failed HTTP requests.
func (wp *WorkerPool) GetCounts() (successes int64, failures int64) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.successCount, wp.failureCount
}
