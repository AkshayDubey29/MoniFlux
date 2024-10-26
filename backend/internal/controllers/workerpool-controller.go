// workerpool-controller.go

package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/sirupsen/logrus"
)

// WorkerPool manages a pool of workers to process log, metric, and trace entries concurrently.
type WorkerPool struct {
	numWorkers   int
	jobs         chan interface{} // Can accept any type of job entry (logs, metrics, traces)
	wg           sync.WaitGroup
	file         *os.File
	logger       *logrus.Logger
	batchSize    int           // Number of entries per batch
	batchDelay   time.Duration // Maximum delay before flushing a batch
	shutdownOnce sync.Once     // Ensures Shutdown is called only once
}

// NewWorkerPool initializes a new WorkerPool with a specified number of workers, log file, batch size, and batch delay.
func NewWorkerPool(numWorkers int, filePath string, logger *logrus.Logger, batchSize int, batchDelay time.Duration) (*WorkerPool, error) {
	// Open the file in append mode with buffering
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	wp := &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan interface{}, numWorkers*1000), // Buffered channel
		file:       file,
		logger:     logger,
		batchSize:  batchSize,
		batchDelay: batchDelay,
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

// worker processes each job (log, metric, trace) and writes them to the file.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	wp.logger.Debugf("Worker %d started", id)

	for job := range wp.jobs {
		switch entry := job.(type) {
		case models.LogEntry:
			wp.processLog(entry)
		case models.Metric:
			wp.processMetric(entry)
		case models.Trace:
			wp.processTrace(entry)
		default:
			wp.logger.Errorf("Worker %d: Unknown job type: %T", id, job)
		}
	}
	wp.logger.Debugf("Worker %d stopped", id)
}

// processLog handles LogEntry
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

// processMetric handles Metric
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

// processTrace handles Trace
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
		err = wp.file.Close()
	})
	return err
}
