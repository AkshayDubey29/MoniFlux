// backend/internal/controllers/controller.go

package controllers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/v1"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/db/mongo"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

// LoadGenController manages load generation operations.
type LoadGenController struct {
	Config      *v1.Config
	Logger      *logrus.Logger
	MongoClient *mongo.MongoClient

	mu    sync.Mutex
	tests map[string]*TestTask
}

// TestTask represents an ongoing load test with its cancellation function.
type TestTask struct {
	CancelFunc context.CancelFunc
}

// NewLoadGenController creates a new LoadGenController.
func NewLoadGenController(cfg *v1.Config, log *logrus.Logger, mongoClient *mongo.MongoClient) *LoadGenController {
	return &LoadGenController{
		Config:      cfg,
		Logger:      log,
		MongoClient: mongoClient,
		tests:       make(map[string]*TestTask),
	}
}

// StartTest initiates a new load test.
func (c *LoadGenController) StartTest(ctx context.Context, test *models.Test) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Assign a unique TestID if not provided.
	if test.TestID == "" {
		test.TestID = uuid.New().String()
	}

	// Initialize status and timestamps.
	test.Status = "Running"
	test.CreatedAt = time.Now()
	test.UpdatedAt = time.Now()

	// Insert the test into the database.
	if _, err := c.MongoClient.InsertOne(ctx, "tests", test); err != nil {
		c.Logger.Errorf("Failed to insert test into database: %v", err)
		return err
	}

	// Create a cancellable context for the load generation process.
	loadCtx, cancel := context.WithCancel(context.Background())

	// Store the TestTask for future cancellation.
	c.tests[test.TestID] = &TestTask{
		CancelFunc: cancel,
	}

	// Start load generation in a separate goroutine.
	go c.generateLoad(loadCtx, test)

	c.Logger.Infof("Load test started: %s", test.TestID)
	return nil
}

// generateLoad simulates load generation.
func (c *LoadGenController) generateLoad(ctx context.Context, test *models.Test) {
	ticker := time.NewTicker(time.Duration(test.LogRate) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.Logger.Infof("Load test cancelled: %s", test.TestID)
			c.updateTestStatus(context.Background(), test.TestID, "Cancelled")
			return
		case <-ticker.C:
			// Simulate log, metric, and trace generation.
			if err := c.generateLog(ctx, test); err != nil {
				c.Logger.Errorf("Error generating log for test %s: %v", test.TestID, err)
				c.updateTestStatus(context.Background(), test.TestID, "Error")
				return
			}

			if err := c.generateMetric(ctx, test); err != nil {
				c.Logger.Errorf("Error generating metric for test %s: %v", test.TestID, err)
				c.updateTestStatus(context.Background(), test.TestID, "Error")
				return
			}

			if err := c.generateTrace(ctx, test); err != nil {
				c.Logger.Errorf("Error generating trace for test %s: %v", test.TestID, err)
				c.updateTestStatus(context.Background(), test.TestID, "Error")
				return
			}
		}
	}
}

// generateLog simulates log generation.
func (c *LoadGenController) generateLog(ctx context.Context, test *models.Test) error {
	time.Sleep(time.Duration(test.LogSize) * time.Millisecond)

	logEntry := models.LogEntry{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		Message:   "Simulated log entry",
		Level:     "INFO",
	}

	if _, err := c.MongoClient.InsertOne(ctx, "logs", logEntry); err != nil {
		return err
	}

	c.Logger.Debugf("Log generated for test %s", test.TestID)
	return nil
}

// generateMetric simulates metric generation.
func (c *LoadGenController) generateMetric(ctx context.Context, test *models.Test) error {
	time.Sleep(50 * time.Millisecond)

	metric := models.Metric{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		Value:     42.0, // Example metric value
	}

	if _, err := c.MongoClient.InsertOne(ctx, "metrics", metric); err != nil {
		return err
	}

	c.Logger.Debugf("Metric generated for test %s", test.TestID)
	return nil
}

// generateTrace simulates trace generation.
func (c *LoadGenController) generateTrace(ctx context.Context, test *models.Test) error {
	time.Sleep(30 * time.Millisecond)

	trace := models.Trace{
		TestID:    test.TestID,
		Timestamp: time.Now(),
		TraceID:   uuid.New().String(),
		SpanID:    uuid.New().String(),
	}

	if _, err := c.MongoClient.InsertOne(ctx, "traces", trace); err != nil {
		return err
	}

	c.Logger.Debugf("Trace generated for test %s", test.TestID)
	return nil
}

// updateTestStatus updates the status of a test in the database.
func (c *LoadGenController) updateTestStatus(ctx context.Context, testID, status string) error {
	filter := map[string]interface{}{
		"testID": testID,
	}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":      status,
			"updatedAt":   time.Now(),
			"completedAt": time.Now(),
		},
	}

	if _, err := c.MongoClient.UpdateOne(ctx, "tests", filter, update); err != nil {
		c.Logger.Errorf("Failed to update status for test %s: %v", testID, err)
		return err
	}

	c.Logger.Infof("Test %s status updated to %s", testID, status)
	return nil
}

// ScheduleTest schedules a test to start at a specified time.
func (c *LoadGenController) ScheduleTest(ctx context.Context, schedule *models.ScheduleRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var test models.Test
	err := c.MongoClient.FindOne(ctx, "tests", map[string]interface{}{
		"testID": schedule.TestID,
	}, &test)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return fmt.Errorf("test with ID %s not found", schedule.TestID)
		}
		return err
	}

	if test.Status != "Pending" && test.Status != "Scheduled" {
		return fmt.Errorf("test with ID %s cannot be scheduled in its current state: %s", schedule.TestID, test.Status)
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"scheduledTime": schedule.Schedule,
			"status":        "Scheduled",
			"updatedAt":     time.Now(),
		},
	}
	if _, err = c.MongoClient.UpdateOne(ctx, "tests", map[string]interface{}{
		"testID": schedule.TestID,
	}, update); err != nil {
		return err
	}

	go c.scheduleTestExecution(ctx, schedule.TestID, schedule.Schedule)
	return nil
}

// scheduleTestExecution starts the test when the scheduled time arrives.
func (c *LoadGenController) scheduleTestExecution(ctx context.Context, testID string, startTime time.Time) {
	timer := time.NewTimer(time.Until(startTime))
	defer timer.Stop()

	select {
	case <-timer.C:
		c.mu.Lock()
		defer c.mu.Unlock()

		var test models.Test
		err := c.MongoClient.FindOne(ctx, "tests", map[string]interface{}{
			"testID": testID,
		}, &test)
		if err != nil {
			c.Logger.Errorf("Failed to retrieve test %s for scheduled start: %v", testID, err)
			return
		}

		if test.Status != "Scheduled" {
			return
		}

		if err = c.StartTest(ctx, &test); err != nil {
			c.Logger.Errorf("Failed to start scheduled test %s: %v", testID, err)
		}
	case <-ctx.Done():
		c.Logger.Infof("Scheduling for test %s cancelled", testID)
	}
}

// CancelTest cancels a running or scheduled test.
func (c *LoadGenController) CancelTest(ctx context.Context, testID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel running tests.
	if task, exists := c.tests[testID]; exists {
		task.CancelFunc()
		delete(c.tests, testID)
		c.Logger.Infof("Cancelled running test: %s", testID)
		return nil
	}

	// Cancel scheduled tests.
	filter := map[string]interface{}{
		"testID": testID,
		"status": "Scheduled",
	}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":    "Cancelled",
			"updatedAt": time.Now(),
		},
	}

	result, err := c.MongoClient.UpdateOne(ctx, "tests", filter, update)
	if err != nil {
		c.Logger.Errorf("Failed to cancel scheduled test %s: %v", testID, err)
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no running or scheduled test found with ID %s", testID)
	}

	c.Logger.Infof("Cancelled scheduled test: %s", testID)
	return nil
}

// RestartTest restarts an existing test with updated configurations.
func (c *LoadGenController) RestartTest(ctx context.Context, restartReq *models.RestartRequest) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var test models.Test
	err := c.MongoClient.FindOne(ctx, "tests", map[string]interface{}{
		"testID": restartReq.TestID,
	}, &test)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			return fmt.Errorf("test with ID %s not found", restartReq.TestID)
		}
		return err
	}

	if test.Status != "Completed" && test.Status != "Cancelled" && test.Status != "Error" {
		return fmt.Errorf("test with ID %s cannot be restarted in its current state: %s", restartReq.TestID, test.Status)
	}

	// Update the test configurations.
	updateFields := make(map[string]interface{})
	if restartReq.LogRate > 0 {
		updateFields["logRate"] = restartReq.LogRate
	}
	if restartReq.Duration > 0 {
		updateFields["duration"] = restartReq.Duration
	}

	if len(updateFields) > 0 {
		update := map[string]interface{}{
			"$set": updateFields,
		}
		if _, err := c.MongoClient.UpdateOne(ctx, "tests", map[string]interface{}{
			"testID": restartReq.TestID,
		}, update); err != nil {
			return fmt.Errorf("failed to update test configurations: %w", err)
		}
	}

	// Reset the test status.
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":      "Running",
			"updatedAt":   time.Now(),
			"createdAt":   time.Now(),
			"completedAt": nil,
		},
	}
	if _, err := c.MongoClient.UpdateOne(ctx, "tests", map[string]interface{}{
		"testID": restartReq.TestID,
	}, update); err != nil {
		return fmt.Errorf("failed to reset test status: %w", err)
	}

	if err := c.StartTest(ctx, &test); err != nil {
		return fmt.Errorf("failed to restart test: %w", err)
	}

	c.Logger.Infof("Test %s restarted successfully", restartReq.TestID)
	return nil
}

// SaveResults saves the results of a completed test.
func (c *LoadGenController) SaveResults(ctx context.Context, results *models.TestResults) error {
	var test models.Test
	if err := c.MongoClient.FindOne(ctx, "tests", map[string]interface{}{
		"testID": results.TestID,
	}, &test); err != nil {
		return err
	}

	if test.Status != "Completed" && test.Status != "Error" {
		return fmt.Errorf("test with ID %s cannot have results saved in its current state: %s", results.TestID, test.Status)
	}

	if _, err := c.MongoClient.InsertOne(ctx, "test_results", results); err != nil {
		return fmt.Errorf("failed to save test results: %w", err)
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":      "Results Saved",
			"updatedAt":   time.Now(),
			"completedAt": results.CompletedAt,
		},
	}
	if _, err := c.MongoClient.UpdateOne(ctx, "tests", map[string]interface{}{
		"testID": results.TestID,
	}, update); err != nil {
		return fmt.Errorf("failed to update test status after saving results: %w", err)
	}

	c.Logger.Infof("Results saved for test %s", results.TestID)
	return nil
}

// GetAllTests retrieves all active and scheduled tests.
func (c *LoadGenController) GetAllTests(ctx context.Context) ([]models.Test, error) {
	var tests []models.Test
	if err := c.MongoClient.FindAll(ctx, "tests", map[string]interface{}{}, &tests); err != nil {
		c.Logger.Errorf("Failed to retrieve all tests: %v", err)
		return nil, err
	}

	c.Logger.Infof("Retrieved %d tests from the database", len(tests))
	return tests, nil
}

// GetTestByID retrieves a specific test by its TestID.
func (c *LoadGenController) GetTestByID(ctx context.Context, testID string) (*models.Test, error) {
	var test models.Test
	if err := c.MongoClient.FindOne(ctx, "tests", map[string]interface{}{
		"testID": testID,
	}, &test); err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
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

		if err := c.updateTestStatus(ctx, testID, "Stopped"); err != nil {
			c.Logger.Errorf("Failed to update status for stopped test %s: %v", testID, err)
		}
	}

	return nil
}
