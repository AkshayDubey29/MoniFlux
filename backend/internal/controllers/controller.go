// backend/internal/controllers/controller.go

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestTask represents a running load test with its cancel function and worker pool.
type TestTask struct {
	CancelFunc context.CancelFunc
	WorkerPool *WorkerPool
}

// LoadGenController manages the main load generation operations.
type LoadGenController struct {
	MongoClient *mongo.Client
	Config      *common.Config
	Logger      *logrus.Logger
	Validator   *validator.Validate
	mu          sync.Mutex
	tests       map[string]*TestTask
}

// NewLoadGenController initializes a new LoadGenController.
func NewLoadGenController(cfg *common.Config, log *logrus.Logger, mongoClient *mongo.Client) *LoadGenController {
	return &LoadGenController{
		Config:      cfg,
		Logger:      log,
		MongoClient: mongoClient,
		Validator:   validator.New(),
		tests:       make(map[string]*TestTask),
	}
}

func generateRandomMessage(size int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	message := make([]rune, size)
	for i := range message {
		message[i] = letters[rand.Intn(len(letters))]
	}
	return string(message)
}

// Generates a random metric value, adjust as needed.
func generateRandomMetricValue() float64 {
	return rand.Float64() * 100 // Example: Random value between 0 and 100
}

// Generates a random duration for traces in milliseconds.
func generateRandomDuration() int {
	return rand.Intn(500) + 50 // Example: Random duration between 50ms and 550ms
}

// determineNumberOfWorkers calculates the number of workers based on log rate and size.
func determineNumberOfWorkers(logRate int, logSize int) int {
	// Assume each worker can handle a certain number of logs per second.
	// Adjust this multiplier based on system benchmarking.
	logsPerWorker := 10000 // Each worker handles 10,000 logs/sec.

	numWorkers := logRate / logsPerWorker
	if logRate%logsPerWorker != 0 {
		numWorkers++
	}

	// Set a minimum and maximum limit.
	if numWorkers < 1 {
		numWorkers = 1
	}
	if numWorkers > 10000 { // Arbitrary upper limit to prevent excessive workers.
		numWorkers = 10000
	}

	return numWorkers
}

// assignDefaults sets default values based on destination type and other properties.
func (c *LoadGenController) assignDefaults(test *models.Test) {
	switch test.Destination.Type {
	case "file":
		if test.Destination.FileCount == 0 {
			test.Destination.FileCount = 10
			c.Logger.Infof("Defaulting FileCount to %d for test %s", test.Destination.FileCount, test.TestID)
		}
		if test.Destination.FileFreq == 0 {
			test.Destination.FileFreq = 5
			c.Logger.Infof("Defaulting FileFreq to %d minutes for test %s", test.Destination.FileFreq, test.TestID)
		}
		if test.Destination.FilePath == "" {
			test.Destination.FilePath = "/tmp/default-output.log"
			c.Logger.Infof("Defaulting FilePath to %s for test %s", test.Destination.FilePath, test.TestID)
		}
	case "http":
		if test.Destination.Port == 0 {
			test.Destination.Port = 80
			c.Logger.Infof("Defaulting Port to %d for test %s", test.Destination.Port, test.TestID)
		}
		if test.Destination.Endpoint == "" {
			test.Destination.Endpoint = "http://localhost/api"
			c.Logger.Infof("Defaulting Endpoint to %s for test %s", test.Destination.Endpoint, test.TestID)
		}
		if test.Destination.APIKey == "" {
			test.Destination.APIKey = "default-api-key"
			c.Logger.Infof("Defaulting APIKey for test %s", test.TestID)
		}
	default:
		c.Logger.Warnf("Unknown destination type '%s' for test %s", test.Destination.Type, test.TestID)
	}
	if test.LogRate == 0 {
		test.LogRate = 50
		c.Logger.Infof("Defaulting LogRate to %d for test %s", test.LogRate, test.TestID)
	}
	if test.MetricsRate == 0 {
		test.MetricsRate = 20
		c.Logger.Infof("Defaulting MetricsRate to %d for test %s", test.MetricsRate, test.TestID)
	}
	if test.TraceRate == 0 {
		test.TraceRate = 10
		c.Logger.Infof("Defaulting TraceRate to %d for test %s", test.TraceRate, test.TestID)
	}
	if test.Duration == 0 {
		test.Duration = 300
		c.Logger.Infof("Defaulting Duration to %d seconds for test %s", test.Duration, test.TestID)
	}
}

// StartTest initiates or updates a load test.
func (c *LoadGenController) StartTest(ctx context.Context, test *models.Test) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Assign default values based on destination type before validation.
	c.assignDefaults(test)

	// Validate the test configuration.
	if err := c.Validator.Struct(test); err != nil {
		c.Logger.Errorf("Validation failed for test %s: %v", test.TestID, err)
		return fmt.Errorf("validation failed: %w", err)
	}

	// Access MongoDB collection and check for an existing test.
	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": test.TestID}

	var existingTest models.Test
	err := collection.FindOne(ctx, filter).Decode(&existingTest)
	isNewTest := errors.Is(err, mongo.ErrNoDocuments)

	if isNewTest {
		// Set a unique TestID and initialize test status and timestamps.
		if test.TestID == "" {
			test.TestID = uuid.New().String()
		}
		test.Status = "Running"
		test.CreatedAt, test.UpdatedAt = time.Now(), time.Now()

		// Insert the new test into the database.
		_, err = collection.InsertOne(ctx, test)
		if err != nil {
			c.Logger.Errorf("Failed to insert test %s: %v", test.TestID, err)
			return fmt.Errorf("failed to insert test: %w", err)
		}
		c.Logger.Infof("Test %s started and inserted as new", test.TestID)
	} else {
		// Ensure the test is in a startable state.
		if existingTest.Status == "Running" {
			return fmt.Errorf("test with ID %s is already running", test.TestID)
		}
		if existingTest.Status != "Cancelled" && existingTest.Status != "Completed" && existingTest.Status != "Error" {
			return fmt.Errorf("test with ID %s cannot be started in its current state: %s", test.TestID, existingTest.Status)
		}

		// Update the existing test's configuration and set it to "Running".
		update := bson.M{
			"$set": bson.M{
				"logRate":       test.LogRate,
				"metricsRate":   test.MetricsRate,
				"traceRate":     test.TraceRate,
				"logSize":       test.LogSize,
				"duration":      test.Duration,
				"status":        "Running",
				"updatedAt":     time.Now(),
				"completedAt":   time.Time{},
				"scheduledTime": time.Time{},
			},
		}

		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.Logger.Errorf("Failed to update test %s: %v", test.TestID, err)
			return fmt.Errorf("failed to update test: %w", err)
		}
		c.Logger.Infof("Test %s configuration updated and started", test.TestID)
	}

	// Determine Destination Type and Endpoint
	destinationType := FileDestination
	destinationValue := ""

	if test.Destination.Type == "http" {
		destinationType = HTTPDestination
		destinationValue = test.Destination.Endpoint
		if destinationValue == "" {
			c.Logger.Errorf("HTTP endpoint must be specified for HTTP destination")
			return fmt.Errorf("HTTP endpoint must be specified for HTTP destination")
		}
	} else if test.Destination.Type == "file" {
		destinationType = FileDestination
		destinationValue = test.Destination.FilePath
		if destinationValue == "" {
			c.Logger.Errorf("filePath must be specified for file destination")
			return fmt.Errorf("filePath must be specified for file destination")
		}
	} else {
		c.Logger.Errorf("Unsupported destination type: %s", test.Destination.Type)
		return fmt.Errorf("unsupported destination type: %s", test.Destination.Type)
	}

	// Create a self-contained cancellable context based on the test's duration
	loadCtx, cancel := context.WithTimeout(context.Background(), time.Duration(test.Duration)*time.Second)

	// Initialize WorkerPool based on the destination type
	numWorkers := determineNumberOfWorkers(test.LogRate, test.LogSize)
	batchSize := 1000                    // Customize as needed
	batchDelay := 100 * time.Millisecond // Adjust as necessary

	// Log initialization details
	if destinationType == FileDestination {
		c.Logger.Infof("Initializing WorkerPool with filePath: %s for test %s", destinationValue, test.TestID)
	} else {
		c.Logger.Infof("Initializing WorkerPool with httpEndpoint: %s for test %s", destinationValue, test.TestID)
	}

	wp, err := NewWorkerPool(numWorkers, destinationType, destinationValue, c.Logger, batchSize, batchDelay)
	if err != nil {
		c.Logger.Errorf("Failed to initialize WorkerPool for test %s: %v", test.TestID, err)
		cancel()
		return fmt.Errorf("failed to initialize WorkerPool: %w", err)
	}

	// Register the test with its CancelFunc and WorkerPool
	c.tests[test.TestID] = &TestTask{
		CancelFunc: cancel,
		WorkerPool: wp,
	}

	// Start the load generation in a new goroutine
	go func() {
		defer func() {
			// Shutdown resources and log upon task completion or error
			err := wp.Shutdown()
			if err != nil {
				c.Logger.Errorf("Failed to shutdown WorkerPool for test %s: %v", test.TestID, err)
			}
			cancel()
		}()

		// Generate load; handle any errors encountered during the process
		if err := c.generateLoad(loadCtx, test, wp); err != nil {
			c.Logger.Errorf("Load generation for test %s failed: %v", test.TestID, err)
			c.updateTestStatus(context.Background(), test.TestID, "Error")
		} else {
			c.updateTestStatus(context.Background(), test.TestID, "Completed")
		}
	}()

	c.Logger.Infof("Load generation task started for test %s with %d workers", test.TestID, numWorkers)
	return nil
}

// generateLoad simulates load generation based on test configuration.
// It generates logs, metrics, and traces as per the configured rates.
// controller.go

// controller.go

func (c *LoadGenController) generateLoad(ctx context.Context, test *models.Test, wp *WorkerPool) error {
	c.Logger.Infof("Starting load generation for test %s with duration %d seconds", test.TestID, test.Duration)

	// Calculate total logs, metrics, and traces to generate based on rates and duration.
	totalLogs := test.LogRate * test.Duration
	totalMetrics := test.MetricsRate * test.Duration
	totalTraces := test.TraceRate * test.Duration

	// Initialize tickers for precise rate control.
	logInterval := time.Second / time.Duration(test.LogRate)
	metricInterval := time.Second / time.Duration(test.MetricsRate)
	traceInterval := time.Second / time.Duration(test.TraceRate)

	// Ensure that intervals are not zero to prevent ticker misconfiguration.
	if logInterval <= 0 {
		logInterval = time.Millisecond // Minimum interval.
	}
	if metricInterval <= 0 {
		metricInterval = time.Millisecond
	}
	if traceInterval <= 0 {
		traceInterval = time.Millisecond
	}

	logTicker := time.NewTicker(logInterval)
	defer logTicker.Stop()

	metricTicker := time.NewTicker(metricInterval)
	defer metricTicker.Stop()

	traceTicker := time.NewTicker(traceInterval)
	defer traceTicker.Stop()

	// Channel to signal completion.
	done := time.After(time.Duration(test.Duration) * time.Second)

	// Counters for generated logs, metrics, and traces.
	var generatedLogs, generatedMetrics, generatedTraces int

	startTime := time.Now()

	for {
		select {
		case <-done:
			c.Logger.Infof("Load test duration completed: %s", test.TestID)
			// Optionally, log final counts if HTTP destination
			if test.Destination.Type == "http" {
				successes, failures := wp.GetCounts()
				c.Logger.Infof("Load test %s completed. Successes: %d, Failures: %d", test.TestID, successes, failures)
			}
			return nil

		case <-ctx.Done():
			c.Logger.Infof("Load test context cancelled: %s, Reason: %v", test.TestID, ctx.Err())
			// Optionally, log final counts if HTTP destination
			if test.Destination.Type == "http" {
				successes, failures := wp.GetCounts()
				c.Logger.Infof("Load test %s cancelled. Successes: %d, Failures: %d", test.TestID, successes, failures)
			}
			return ctx.Err()

		case <-logTicker.C:
			if generatedLogs >= totalLogs {
				logTicker.Stop()
				continue
			}
			logEntry := models.LogEntry{
				TestID:    test.TestID,
				Timestamp: time.Now().UTC(), // Ensure correct type
				Message:   generateRandomMessage(test.LogSize),
				Level:     test.LogType,
			}
			wp.Submit(logEntry)
			generatedLogs++

			// Optional: Log progress at intervals.
			if generatedLogs%100000 == 0 {
				elapsed := time.Since(startTime).Seconds()
				c.Logger.Infof("Generated %d logs for test %s in %.2f seconds", generatedLogs, test.TestID, elapsed)
				if test.Destination.Type == "http" {
					successes, failures := wp.GetCounts()
					c.Logger.Infof("HTTP Logs - Successes: %d, Failures: %d", successes, failures)
				}
			}

		case <-metricTicker.C:
			if generatedMetrics >= totalMetrics {
				metricTicker.Stop()
				continue
			}
			metric := models.Metric{
				TestID:    test.TestID,
				Timestamp: time.Now().UTC(), // Ensure correct type
				Value:     generateRandomMetricValue(),
			}
			wp.Submit(metric)
			generatedMetrics++

			// Optional: Log progress at intervals.
			if generatedMetrics%50000 == 0 {
				elapsed := time.Since(startTime).Seconds()
				c.Logger.Infof("Generated %d metrics for test %s in %.2f seconds", generatedMetrics, test.TestID, elapsed)
			}

		case <-traceTicker.C:
			if generatedTraces >= totalTraces {
				traceTicker.Stop()
				continue
			}
			trace := models.Trace{
				TestID:    test.TestID,
				Timestamp: time.Now().UTC(), // Ensure correct type
				TraceID:   uuid.New().String(),
				SpanID:    uuid.New().String(),
				Operation: "SimulatedOperation",
				Duration:  generateRandomDuration(),
			}
			wp.Submit(trace)
			generatedTraces++

			// Optional: Log progress at intervals.
			if generatedTraces%50000 == 0 {
				elapsed := time.Since(startTime).Seconds()
				c.Logger.Infof("Generated %d traces for test %s in %.2f seconds", generatedTraces, test.TestID, elapsed)
			}
		}
	}
}

// generateLog simulates log generation and sends it to the configured destination.
// Deprecated: Using WorkerPool instead.
func (c *LoadGenController) generateLog(test *models.Test) error {
	logEntry := models.LogEntry{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		Message:   "Simulated log entry",
		Level:     test.LogType,
	}
	return c.sendToDestination(test.Destination, logEntry)
}

// generateMetric simulates metric generation and sends it to the configured destination.
// Deprecated: Using WorkerPool instead.
func (c *LoadGenController) generateMetric(test *models.Test) error {
	metric := models.Metric{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		Value:     42.0, // Example metric value.
	}
	return c.sendToDestination(test.Destination, metric)
}

// generateTrace simulates trace generation and sends it to the configured destination.
// Deprecated: Using WorkerPool instead.
func (c *LoadGenController) generateTrace(test *models.Test) error {
	trace := models.Trace{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		TraceID:   uuid.New().String(),
		SpanID:    uuid.New().String(),
		Operation: "SimulatedOperation",
		Duration:  100, // Duration in milliseconds.
	}
	return c.sendToDestination(test.Destination, trace)
}

// monitorConfigUpdates monitors for configuration changes in MongoDB and applies them.
func (c *LoadGenController) monitorConfigUpdates(ctx context.Context, testID string) {
	ticker := time.NewTicker(10 * time.Second) // Poll interval.
	defer ticker.Stop()

	c.Logger.Infof("Started monitoring config updates for test %s", testID)

	for {
		select {
		case <-ticker.C:
			updatedConfig, err := c.fetchUpdatedConfig(ctx, testID)
			if err != nil {
				c.Logger.Errorf("Error fetching updated configuration for test %s: %v", testID, err)
				continue
			}

			// Compare updatedConfig with current config.
			if c.hasConfigChanged(testID, updatedConfig) {
				c.Logger.Infof("Configuration change detected for test %s", testID)
				c.applyConfigUpdates(updatedConfig)
			} else {
				c.Logger.Debugf("No configuration change detected for test %s", testID)
			}

		case <-ctx.Done():
			c.Logger.Infof("Stopped monitoring for config updates on test %s", testID)
			return
		}
	}
}

// hasConfigChanged checks if there are any changes in the configuration.
func (c *LoadGenController) hasConfigChanged(testID string, updatedConfig *models.Test) bool {
	// Fetch the current test configuration.
	currentTest, err := c.GetTestByID(context.Background(), testID)
	if err != nil {
		c.Logger.Errorf("Error fetching current configuration for test %s: %v", testID, err)
		return false
	}

	// Compare relevant fields.
	if currentTest.LogRate != updatedConfig.LogRate ||
		currentTest.MetricsRate != updatedConfig.MetricsRate ||
		currentTest.TraceRate != updatedConfig.TraceRate ||
		currentTest.Duration != updatedConfig.Duration ||
		currentTest.Destination.Type != updatedConfig.Destination.Type ||
		currentTest.Destination.FilePath != updatedConfig.Destination.FilePath {
		return true
	}

	return false
}

// fetchUpdatedConfig retrieves the latest test configuration from MongoDB.
func (c *LoadGenController) fetchUpdatedConfig(ctx context.Context, testID string) (*models.Test, error) {
	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	var updatedConfig models.Test
	err := collection.FindOne(ctx, bson.M{"testID": testID}).Decode(&updatedConfig)
	if err != nil {
		return nil, err
	}
	return &updatedConfig, nil
}

// sendToDestination sends data to the configured destination based on type.
func (c *LoadGenController) sendToDestination(destination common.Destination, data interface{}) error {
	switch destination.Type {
	case "file":
		return c.writeLogToFile(destination.FilePath, data)
	case "http":
		return c.sendLogToHTTP(destination.Endpoint, data, destination.APIKey)
	default:
		return fmt.Errorf("unknown destination type: %s", destination.Type)
	}
}

// writeLogToFile writes data to a specified file in JSON format.
func (c *LoadGenController) writeLogToFile(filePath string, data interface{}) error {
	c.Logger.Infof("Attempting to write data to file: %s", filePath)

	// Ensure the directory exists.
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.Logger.Errorf("Failed to create directories for file path %s: %v", filePath, err)
		return err
	}

	// Serialize data to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.Logger.Errorf("Failed to marshal data for file %s: %v", filePath, err)
		return err
	}

	// Write JSON data to file with a newline.
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		c.Logger.Errorf("Failed to open log file %s: %v", filePath, err)
		return err
	}
	defer file.Close()

	if _, err := file.Write(append(jsonData, '\n')); err != nil {
		c.Logger.Errorf("Failed to write to log file %s: %v", filePath, err)
		return err
	}

	c.Logger.Infof("Data successfully written to file: %s", filePath)
	return nil
}

// sendLogToHTTP sends data to a specified HTTP endpoint as JSON.
func (c *LoadGenController) sendLogToHTTP(endpoint string, data interface{}, apiKey string) error {
	// Serialize data to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		c.Logger.Errorf("Failed to marshal data for HTTP endpoint %s: %v", endpoint, err)
		return err
	}

	// Create HTTP request.
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		c.Logger.Errorf("Failed to create HTTP request for endpoint %s: %v", endpoint, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	// Send HTTP request.
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.Logger.Errorf("Failed to send data to HTTP endpoint %s: %v", endpoint, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.Logger.Errorf("Received non-success status code %d from HTTP endpoint %s", resp.StatusCode, endpoint)
		return fmt.Errorf("received status code %d from endpoint", resp.StatusCode)
	}

	c.Logger.Debugf("Data sent to HTTP endpoint %s successfully", endpoint)
	return nil
}

// applyConfigUpdates applies configuration changes to the running test.
func (c *LoadGenController) applyConfigUpdates(updatedConfig *models.Test) {
	testID := updatedConfig.TestID

	// Cancel the existing load generation.
	if task, exists := c.tests[testID]; exists {
		task.CancelFunc()
		delete(c.tests, testID)
		c.Logger.Infof("Existing load generation for test %s stopped for configuration update", testID)
	}

	// Start load generation with updated configuration.
	go func() {
		if err := c.StartTest(context.Background(), updatedConfig); err != nil {
			c.Logger.Errorf("Failed to apply updated configuration for test %s: %v", testID, err)
			c.updateTestStatus(context.Background(), testID, "Error")
		}
	}()
}

// updateTestStatus updates the status of a test in the database.
func (c *LoadGenController) updateTestStatus(ctx context.Context, testID, status string) error {
	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": testID}
	update := bson.M{
		"$set": bson.M{
			"status":        status,
			"updatedAt":     time.Now(),
			"completedAt":   time.Now(),
			"scheduledTime": time.Time{},
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.Logger.Errorf("Failed to update status for test %s: %v", testID, err)
		return err
	}

	c.Logger.Infof("Test %s status updated to %s", testID, status)
	return nil
}

// ScheduleTest schedules a test to start at a specified time.
func (c *LoadGenController) ScheduleTest(ctx context.Context, scheduleReq *models.ScheduleRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": scheduleReq.TestID}

	var test models.Test
	err := collection.FindOne(ctx, filter).Decode(&test)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("test with ID %s not found", scheduleReq.TestID)
		}
		return fmt.Errorf("error retrieving test: %w", err)
	}

	// Only allow scheduling if the test is in "Pending" or "Scheduled" state.
	if test.Status != "Pending" && test.Status != "Scheduled" {
		return fmt.Errorf("test with ID %s cannot be scheduled in its current state: %s", scheduleReq.TestID, test.Status)
	}

	// Update the test's scheduledTime and status.
	update := bson.M{
		"$set": bson.M{
			"scheduledTime": scheduleReq.ScheduleAt,
			"status":        "Scheduled",
			"updatedAt":     time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to schedule test: %w", err)
	}

	c.Logger.Infof("Test %s scheduled to start at %v", scheduleReq.TestID, scheduleReq.ScheduleAt)

	// Start a goroutine to execute the test at the scheduled time.
	go c.scheduleTestExecution(scheduleReq.TestID, scheduleReq.ScheduleAt)

	return nil
}

// scheduleTestExecution starts the test when the scheduled time arrives.
func (c *LoadGenController) scheduleTestExecution(testID string, startTime time.Time) {
	timerDuration := time.Until(startTime)
	if timerDuration < 0 {
		c.Logger.Errorf("Scheduled start time %v is in the past for test %s", startTime, testID)
		return
	}

	timer := time.NewTimer(timerDuration)
	defer timer.Stop()

	select {
	case <-timer.C:
		c.mu.Lock()
		defer c.mu.Unlock()

		collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
		filter := bson.M{"testID": testID}

		var test models.Test
		err := collection.FindOne(context.Background(), filter).Decode(&test)
		if err != nil {
			c.Logger.Errorf("Failed to retrieve test %s for scheduled start: %v", testID, err)
			return
		}

		// Only start if the test is still in "Scheduled" status.
		if test.Status != "Scheduled" {
			c.Logger.Infof("Test %s is no longer in 'Scheduled' status. Current status: %s", testID, test.Status)
			return
		}

		// Start the test.
		err = c.StartTest(context.Background(), &test)
		if err != nil {
			c.Logger.Errorf("Failed to start scheduled test %s: %v", testID, err)
			c.updateTestStatus(context.Background(), testID, "Error")
			return
		}

		c.Logger.Infof("Scheduled test %s started successfully", testID)
	}
}

// CancelTest cancels a running or scheduled test.
func (c *LoadGenController) CancelTest(ctx context.Context, testID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Logger.Infof("Attempting to cancel test with ID: %s", testID)

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": testID}

	var test models.Test
	err := collection.FindOne(ctx, filter).Decode(&test)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("test with ID %s not found", testID)
		}
		c.Logger.Errorf("Error fetching test %s: %v", testID, err)
		return fmt.Errorf("error fetching test: %w", err)
	}

	// Check if the test is already completed or cancelled.
	if test.Status == "Completed" || test.Status == "Cancelled" {
		c.Logger.Infof("Test with ID %s is already %s", testID, test.Status)
		return fmt.Errorf("test with ID %s is already %s", testID, test.Status)
	}

	// If the test is running, cancel the load generation.
	if test.Status == "Running" {
		if task, exists := c.tests[testID]; exists {
			task.CancelFunc()
			delete(c.tests, testID)
			c.Logger.Infof("Cancellation signal sent for running test %s", testID)
		} else {
			c.Logger.Warnf("Test %s is marked as running but no task found in memory", testID)
		}
	}

	// Update the test's status to "Cancelled" in the database.
	update := bson.M{
		"$set": bson.M{
			"status":        "Cancelled",
			"completedAt":   time.Now(),
			"updatedAt":     time.Now(),
			"scheduledTime": time.Time{},
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.Logger.Errorf("Failed to update test status in DB for testID %s: %v", testID, err)
		return fmt.Errorf("failed to update test status in DB for testID %s: %w", testID, err)
	}

	c.Logger.Infof("Test %s successfully cancelled", testID)
	return nil
}

// RestartTest restarts an existing test with updated configurations.
func (c *LoadGenController) RestartTest(ctx context.Context, restartReq *models.RestartRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Logger.Infof("Received request to restart test with ID: %s", restartReq.TestID)

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": restartReq.TestID}

	var test models.Test
	err := collection.FindOne(ctx, filter).Decode(&test)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("test with ID %s not found", restartReq.TestID)
		}
		c.Logger.Errorf("Error retrieving test with ID %s: %v", restartReq.TestID, err)
		return fmt.Errorf("error retrieving test: %w", err)
	}

	// Check if the test status allows restarting.
	if test.Status != "Completed" && test.Status != "Cancelled" && test.Status != "Error" {
		return fmt.Errorf("test with ID %s cannot be restarted in its current state: %s", restartReq.TestID, test.Status)
	}

	// Update the test's configuration if provided.
	updatedFields := bson.M{}
	if restartReq.LogRate > 0 {
		test.LogRate = restartReq.LogRate
		updatedFields["logRate"] = restartReq.LogRate
	}
	if restartReq.MetricsRate > 0 {
		test.MetricsRate = restartReq.MetricsRate
		updatedFields["metricsRate"] = restartReq.MetricsRate
	}
	if restartReq.TraceRate > 0 {
		test.TraceRate = restartReq.TraceRate
		updatedFields["traceRate"] = restartReq.TraceRate
	}
	if restartReq.Duration > 0 {
		test.Duration = restartReq.Duration
		updatedFields["duration"] = restartReq.Duration
	}

	if len(updatedFields) == 0 {
		c.Logger.Warnf("No valid configuration fields provided to update for test %s", restartReq.TestID)
		return fmt.Errorf("no valid configuration fields provided to update")
	}

	// Update the test's status and reset relevant fields.
	updatedFields["status"] = "Running"
	updatedFields["updatedAt"] = time.Now()
	updatedFields["completedAt"] = time.Time{}
	updatedFields["scheduledTime"] = time.Time{}

	update := bson.M{
		"$set": updatedFields,
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		c.Logger.Errorf("Failed to update test %s in database: %v", restartReq.TestID, err)
		return fmt.Errorf("failed to update test %s in database: %w", restartReq.TestID, err)
	}

	c.Logger.Infof("Test %s configuration updated for restart", restartReq.TestID)

	// If the test was previously running, cancel the existing load generation.
	if task, exists := c.tests[restartReq.TestID]; exists {
		task.CancelFunc()
		delete(c.tests, restartReq.TestID)
		c.Logger.Infof("Existing load generation for test %s stopped for restart", restartReq.TestID)
	}

	// Start load generation with updated configuration.
	err = c.StartTest(ctx, &test)
	if err != nil {
		c.Logger.Errorf("Failed to restart load generation for test %s: %v", restartReq.TestID, err)
		c.updateTestStatus(context.Background(), restartReq.TestID, "Error")
		return fmt.Errorf("failed to restart load generation for test %s: %w", restartReq.TestID, err)
	}

	c.Logger.Infof("Test %s restarted successfully", restartReq.TestID)
	return nil
}

// SaveResults saves the results of a completed test.
func (c *LoadGenController) SaveResults(ctx context.Context, results *models.TestResults) error {
	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": results.TestID}

	var test models.Test
	err := collection.FindOne(ctx, filter).Decode(&test)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("test with ID %s not found", results.TestID)
		}
		return fmt.Errorf("error retrieving test: %w", err)
	}

	// Check if the test is in a state that allows saving results.
	if test.Status != "Completed" && test.Status != "Error" {
		return fmt.Errorf("test with ID %s cannot have results saved in its current state: %s", results.TestID, test.Status)
	}

	// Insert the test results.
	resultsCollection := c.MongoClient.Database(c.Config.MongoDB).Collection("test_results")
	_, err = resultsCollection.InsertOne(ctx, results)
	if err != nil {
		return fmt.Errorf("failed to save test results: %w", err)
	}

	// Update the test's status to "Results Saved".
	update := bson.M{
		"$set": bson.M{
			"status":        "Results Saved",
			"updatedAt":     time.Now(),
			"completedAt":   results.CompletedAt,
			"scheduledTime": time.Time{},
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update test status after saving results: %w", err)
	}

	c.Logger.Infof("Results saved for test %s", results.TestID)
	return nil
}

// GetAllTests retrieves all active and scheduled tests.
func (c *LoadGenController) GetAllTests(ctx context.Context) ([]models.Test, error) {
	var tests []models.Test
	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.Logger.Errorf("Failed to retrieve all tests: %v", err)
		return nil, fmt.Errorf("failed to retrieve tests: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var test models.Test
		if err := cursor.Decode(&test); err != nil {
			c.Logger.Errorf("Failed to decode test: %v", err)
			continue
		}
		tests = append(tests, test)
	}

	if err := cursor.Err(); err != nil {
		c.Logger.Errorf("Cursor error: %v", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	c.Logger.Infof("Retrieved %d tests from the database", len(tests))
	return tests, nil
}

// GetTestByID retrieves a specific test by its TestID.
func (c *LoadGenController) GetTestByID(ctx context.Context, testID string) (*models.Test, error) {
	var test models.Test
	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": testID}

	err := collection.FindOne(ctx, filter).Decode(&test)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("test with ID %s not found", testID)
		}
		return nil, fmt.Errorf("error retrieving test: %w", err)
	}

	return &test, nil
}

// StopAllTests gracefully stops all running tests.
func (c *LoadGenController) StopAllTests(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for testID, task := range c.tests {
		task.CancelFunc()
		delete(c.tests, testID)
		c.Logger.Infof("Stopped test: %s", testID)

		err := c.updateTestStatus(ctx, testID, "Stopped")
		if err != nil {
			c.Logger.Errorf("Failed to update status for stopped test %s: %v", testID, err)
		}
	}

	c.Logger.Infof("All running tests have been stopped")
	return nil
}

// CreateTest inserts a new test into the database.
func (c *LoadGenController) CreateTest(ctx context.Context, test *models.Test) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": test.TestID}

	var existingTest models.Test
	err := collection.FindOne(ctx, filter).Decode(&existingTest)
	isNewTest := errors.Is(err, mongo.ErrNoDocuments)

	if !isNewTest {
		return fmt.Errorf("test with ID %s already exists", test.TestID)
	}

	if test.TestID == "" {
		test.TestID = uuid.New().String()
	}
	test.Status = "Pending"
	test.CreatedAt, test.UpdatedAt = time.Now(), time.Now()

	_, err = collection.InsertOne(ctx, test)
	if err != nil {
		c.Logger.Errorf("Failed to insert test: %v", err)
		return fmt.Errorf("failed to insert test: %w", err)
	}

	c.Logger.Infof("Test %s created successfully", test.TestID)
	return nil
}
