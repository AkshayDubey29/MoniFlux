// backend/internal/controllers/controller.go

package controllers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TestTask represents an ongoing load test with its cancellation function.
type TestTask struct {
	CancelFunc context.CancelFunc
}

// LoadGenController manages load generation operations.
type LoadGenController struct {
	Config      *common.Config
	Logger      *logrus.Logger
	MongoClient *mongo.Client

	mu    sync.Mutex
	tests map[string]*TestTask
}

// NewLoadGenController creates a new LoadGenController.
func NewLoadGenController(cfg *common.Config, log *logrus.Logger, mongoClient *mongo.Client) *LoadGenController {
	return &LoadGenController{
		Config:      cfg,
		Logger:      log,
		MongoClient: mongoClient,
		tests:       make(map[string]*TestTask),
	}
}

// StartTest initiates a new load test or updates an existing one if allowed.
func (c *LoadGenController) StartTest(ctx context.Context, test *models.Test) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{"testID": test.TestID}

	var existingTest models.Test
	err := collection.FindOne(ctx, filter).Decode(&existingTest)
	isNewTest := errors.Is(err, mongo.ErrNoDocuments)

	if isNewTest {
		// Assign a unique TestID if not provided.
		if test.TestID == "" {
			test.TestID = uuid.New().String()
		}

		// Initialize status and timestamps.
		test.Status = "Running"
		test.CreatedAt = time.Now()
		test.UpdatedAt = time.Now()

		// Insert as a new test.
		_, err = collection.InsertOne(ctx, test)
		if err != nil {
			c.Logger.Errorf("Failed to insert new test into database: %v", err)
			return err
		}
		c.Logger.Infof("Test %s started and inserted as new", test.TestID)
	} else {
		// Check if the existing test is in a state that allows starting.
		if existingTest.Status == "Running" {
			return fmt.Errorf("test with ID %s is already running", test.TestID)
		}
		if existingTest.Status != "Cancelled" && existingTest.Status != "Completed" && existingTest.Status != "Error" {
			return fmt.Errorf("test with ID %s cannot be started in its current state: %s", test.TestID, existingTest.Status)
		}

		// Update the existing test's configuration and status.
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
			c.Logger.Errorf("Failed to update existing test in database: %v", err)
			return err
		}
		c.Logger.Infof("Test %s configuration updated and started", test.TestID)
	}

	return c.startOrRestartLoadGeneration(ctx, test)
}

// startOrRestartLoadGeneration starts or restarts the load generation for a test.
func (c *LoadGenController) startOrRestartLoadGeneration(ctx context.Context, test *models.Test) error {
	if task, exists := c.tests[test.TestID]; exists {
		task.CancelFunc()
		delete(c.tests, test.TestID)
		c.Logger.Infof("Existing load generation for test %s stopped", test.TestID)
	}

	// Create a new cancellable context for the load generation process.
	loadCtx, cancel := context.WithCancel(context.Background())
	c.tests[test.TestID] = &TestTask{CancelFunc: cancel}

	// Start load generation in a separate goroutine.
	go c.generateLoad(loadCtx, test)

	c.Logger.Infof("Load generation task started for test %s", test.TestID)
	return nil
}

// generateLoad simulates load generation based on the test configuration.
func (c *LoadGenController) generateLoad(ctx context.Context, test *models.Test) {
	defer func() {
		// Upon completion or cancellation, remove the test from the running tests map.
		c.mu.Lock()
		delete(c.tests, test.TestID)
		c.mu.Unlock()
	}()

	if test.LogRate <= 0 && test.MetricsRate <= 0 && test.TraceRate <= 0 {
		c.Logger.Errorf("No valid rate specified for test %s. At least one of LogRate, MetricsRate, or TraceRate must be > 0.", test.TestID)
		c.updateTestStatus(context.Background(), test.TestID, "Error")
		return
	}

	var logTicker, metricTicker, traceTicker *time.Ticker

	if test.LogRate > 0 {
		logInterval := time.Second / time.Duration(test.LogRate)
		logTicker = time.NewTicker(logInterval)
		defer logTicker.Stop()
	}

	if test.MetricsRate > 0 {
		metricInterval := time.Second / time.Duration(test.MetricsRate)
		metricTicker = time.NewTicker(metricInterval)
		defer metricTicker.Stop()
	}

	if test.TraceRate > 0 {
		traceInterval := time.Second / time.Duration(test.TraceRate)
		traceTicker = time.NewTicker(traceInterval)
		defer traceTicker.Stop()
	}

	done := time.After(time.Duration(test.Duration) * time.Second)

	for {
		select {
		case <-ctx.Done():
			c.Logger.Infof("Load test cancelled: %s", test.TestID)
			c.updateTestStatus(context.Background(), test.TestID, "Cancelled")
			return
		case <-done:
			c.Logger.Infof("Load test completed: %s", test.TestID)
			c.updateTestStatus(context.Background(), test.TestID, "Completed")
			return
		case <-logTicker.C:
			// Generate log asynchronously to prevent blocking
			go func() {
				if err := c.generateLog(context.Background(), test); err != nil {
					c.Logger.Errorf("Error generating log for test %s: %v", test.TestID, err)
					c.updateTestStatus(context.Background(), test.TestID, "Error")
				}
			}()
		case <-metricTicker.C:
			// Generate metric asynchronously to prevent blocking
			go func() {
				if err := c.generateMetric(context.Background(), test); err != nil {
					c.Logger.Errorf("Error generating metric for test %s: %v", test.TestID, err)
					c.updateTestStatus(context.Background(), test.TestID, "Error")
				}
			}()
		case <-traceTicker.C:
			// Generate trace asynchronously to prevent blocking
			go func() {
				if err := c.generateTrace(context.Background(), test); err != nil {
					c.Logger.Errorf("Error generating trace for test %s: %v", test.TestID, err)
					c.updateTestStatus(context.Background(), test.TestID, "Error")
				}
			}()
		}
	}
}

// generateLog simulates log generation.
func (c *LoadGenController) generateLog(ctx context.Context, test *models.Test) error {
	// Simulate processing time based on LogSize (milliseconds)
	if test.LogSize > 0 {
		time.Sleep(time.Duration(test.LogSize) * time.Millisecond)
	}

	logEntry := models.LogEntry{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		Message:   "Simulated log entry",
		Level:     test.LogType,
	}

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("logs")
	if _, err := collection.InsertOne(ctx, logEntry); err != nil {
		return err
	}

	c.Logger.Debugf("Log generated for test %s", test.TestID)
	return nil
}

// generateMetric simulates metric generation.
func (c *LoadGenController) generateMetric(ctx context.Context, test *models.Test) error {
	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	metric := models.Metric{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		Value:     42.0, // Example metric value
	}

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("metrics")
	if _, err := collection.InsertOne(ctx, metric); err != nil {
		return err
	}

	c.Logger.Debugf("Metric generated for test %s", test.TestID)
	return nil
}

// generateTrace simulates trace generation.
func (c *LoadGenController) generateTrace(ctx context.Context, test *models.Test) error {
	// Simulate processing time
	time.Sleep(30 * time.Millisecond)

	trace := models.Trace{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		TraceID:   uuid.New().String(),
		SpanID:    uuid.New().String(),
		Operation: "SimulatedOperation",
		Duration:  100, // Duration in milliseconds
	}

	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("traces")
	if _, err := collection.InsertOne(ctx, trace); err != nil {
		return err
	}

	c.Logger.Debugf("Trace generated for test %s", test.TestID)
	return nil
}

// updateTestStatus updates the status of a test in the database.
func (c *LoadGenController) updateTestStatus(ctx context.Context, testID, status string) error {
	collection := c.MongoClient.Database(c.Config.MongoDB).Collection("tests")
	filter := bson.M{
		"testID": testID,
	}
	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"updatedAt":   time.Now(),
			"completedAt": time.Now(),
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
		return err
	}

	// Only allow scheduling if the test is in "Pending" or "Scheduled" state.
	if test.Status != "Pending" && test.Status != "Scheduled" {
		return fmt.Errorf("test with ID %s cannot be scheduled in its current state: %s", scheduleReq.TestID, test.Status)
	}

	// Update the test's scheduledTime and status.
	update := bson.M{
		"$set": bson.M{
			"scheduledTime": scheduleReq.Schedule,
			"status":        "Scheduled",
			"updatedAt":     time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	c.Logger.Infof("Test %s scheduled to start at %v", scheduleReq.TestID, scheduleReq.Schedule)

	// Start a goroutine to execute the test at the scheduled time.
	go c.scheduleTestExecution(scheduleReq.TestID, scheduleReq.Schedule)

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
		return err
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
		return fmt.Errorf("failed to update test status in DB for testID %s", testID)
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
		return err
	}

	// Check if the test status allows restarting.
	if test.Status != "Completed" && test.Status != "Cancelled" && test.Status != "Error" {
		return fmt.Errorf("test with ID %s cannot be restarted in its current state: %s", restartReq.TestID, test.Status)
	}

	// Update the test's configuration if provided.
	updatedFields := bson.M{}
	if restartReq.LogRate > 0 {
		c.Logger.Infof("Updating LogRate to %d logs/sec for test %s", restartReq.LogRate, restartReq.TestID)
		test.LogRate = restartReq.LogRate
		updatedFields["logRate"] = restartReq.LogRate
	}
	if restartReq.MetricsRate > 0 {
		c.Logger.Infof("Updating MetricsRate to %d metrics/sec for test %s", restartReq.MetricsRate, restartReq.TestID)
		test.MetricsRate = restartReq.MetricsRate
		updatedFields["metricsRate"] = restartReq.MetricsRate
	}
	if restartReq.TraceRate > 0 {
		c.Logger.Infof("Updating TraceRate to %d traces/sec for test %s", restartReq.TraceRate, restartReq.TestID)
		test.TraceRate = restartReq.TraceRate
		updatedFields["traceRate"] = restartReq.TraceRate
	}
	if restartReq.Duration > 0 {
		c.Logger.Infof("Updating Duration to %d seconds for test %s", restartReq.Duration, restartReq.TestID)
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
		return fmt.Errorf("failed to update test %s in database: %v", restartReq.TestID, err)
	}

	c.Logger.Infof("Test %s configuration updated for restart", restartReq.TestID)

	// If the test was previously running, cancel the existing load generation.
	if task, exists := c.tests[restartReq.TestID]; exists {
		task.CancelFunc()
		delete(c.tests, restartReq.TestID)
		c.Logger.Infof("Existing load generation for test %s stopped for restart", restartReq.TestID)
	}

	// Start load generation with updated configuration.
	err = c.startOrRestartLoadGeneration(ctx, &test)
	if err != nil {
		c.Logger.Errorf("Failed to restart load generation for test %s: %v", restartReq.TestID, err)
		c.updateTestStatus(context.Background(), restartReq.TestID, "Error")
		return err
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
		return err
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
		return nil, err
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
		return nil, err
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
		return nil, err
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

	// Assign a unique TestID if not provided.
	if test.TestID == "" {
		test.TestID = uuid.New().String()
	}

	// Initialize status and timestamps.
	test.Status = "Pending"
	test.CreatedAt = time.Now()
	test.UpdatedAt = time.Now()

	// Insert as a new test.
	_, err = collection.InsertOne(ctx, test)
	if err != nil {
		c.Logger.Errorf("Failed to insert new test into database: %v", err)
		return err
	}

	c.Logger.Infof("Test %s created successfully", test.TestID)
	return nil
}
