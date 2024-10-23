// backend/internal/api/handlers/handler.go

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/controllers"
	validator "github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

// Handler encapsulates the controller, validator, and logger.
type Handler struct {
	Controller *controllers.LoadGenController
	Validator  *validator.Validate
	Logger     *logrus.Logger
}

// NewHandler creates a new Handler instance.
func NewHandler(controller *controllers.LoadGenController, logger *logrus.Logger) *Handler {
	return &Handler{
		Controller: controller,
		Validator:  validator.New(),
		Logger:     logger,
	}
}

// StartTestHandler handles the initiation of a new load test.
func (h *Handler) StartTestHandler(w http.ResponseWriter, r *http.Request) {
	var test models.Test
	if err := json.NewDecoder(r.Body).Decode(&test); err != nil {
		h.Logger.Errorf("Failed to decode test: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the test struct.
	if err := h.Validator.Struct(test); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Start the test using the controller.
	if err := h.Controller.StartTest(r.Context(), &test); err != nil {
		h.Logger.Errorf("Failed to start test: %v", err)
		http.Error(w, "Failed to start test", http.StatusInternalServerError)
		return
	}

	// Respond with the created test.
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(test)
}

// ScheduleTestHandler handles scheduling a load test.
func (h *Handler) ScheduleTestHandler(w http.ResponseWriter, r *http.Request) {
	var schedule models.ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		h.Logger.Errorf("Failed to decode schedule request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the schedule request.
	if err := h.Validator.Struct(schedule); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Schedule the test using the controller.
	if err := h.Controller.ScheduleTest(r.Context(), &schedule); err != nil {
		h.Logger.Errorf("Failed to schedule test: %v", err)
		http.Error(w, "Failed to schedule test", http.StatusInternalServerError)
		return
	}

	// Respond with the scheduled request.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

// CancelTestHandler handles cancelling a load test.
func (h *Handler) CancelTestHandler(w http.ResponseWriter, r *http.Request) {
	var cancelReq models.CancelRequest
	if err := json.NewDecoder(r.Body).Decode(&cancelReq); err != nil {
		h.Logger.Errorf("Failed to decode cancel request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the cancel request.
	if err := h.Validator.Struct(cancelReq); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Cancel the test using the controller.
	if err := h.Controller.CancelTest(r.Context(), cancelReq.TestID); err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			http.Error(w, "Test not found", http.StatusNotFound)
			return
		}
		h.Logger.Errorf("Failed to cancel test: %v", err)
		http.Error(w, "Failed to cancel test", http.StatusInternalServerError)
		return
	}

	// Respond with a success message.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

// RestartTestHandler handles restarting a load test.
func (h *Handler) RestartTestHandler(w http.ResponseWriter, r *http.Request) {
	var restartReq models.RestartRequest
	if err := json.NewDecoder(r.Body).Decode(&restartReq); err != nil {
		h.Logger.Errorf("Failed to decode restart request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the restart request.
	if err := h.Validator.Struct(restartReq); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Restart the test using the controller.
	if err := h.Controller.RestartTest(r.Context(), &restartReq); err != nil {
		h.Logger.Errorf("Failed to restart test: %v", err)
		http.Error(w, "Failed to restart test", http.StatusInternalServerError)
		return
	}

	// Respond with the restarted request.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(restartReq)
}

// SaveResultsHandler handles saving the results of a load test.
func (h *Handler) SaveResultsHandler(w http.ResponseWriter, r *http.Request) {
	var results models.TestResults
	if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
		h.Logger.Errorf("Failed to decode test results: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the test results.
	if err := h.Validator.Struct(results); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Save the results using the controller.
	if err := h.Controller.SaveResults(r.Context(), &results); err != nil {
		h.Logger.Errorf("Failed to save test results: %v", err)
		http.Error(w, "Failed to save test results", http.StatusInternalServerError)
		return
	}

	// Respond with the saved results.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

// GetAllTestsHandler retrieves all active and scheduled tests.
func (h *Handler) GetAllTestsHandler(w http.ResponseWriter, r *http.Request) {
	tests, err := h.Controller.GetAllTests(r.Context())
	if err != nil {
		h.Logger.Errorf("Failed to get all tests: %v", err)
		http.Error(w, "Failed to retrieve tests", http.StatusInternalServerError)
		return
	}

	// Respond with the list of tests.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tests)
}

// GetTestByIDHandler retrieves a specific test by its TestID.
func (h *Handler) GetTestByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID := vars["testID"]

	test, err := h.Controller.GetTestByID(r.Context(), testID)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			http.Error(w, "Test not found", http.StatusNotFound)
			return
		}
		h.Logger.Errorf("Failed to get test by ID: %v", err)
		http.Error(w, "Failed to retrieve test", http.StatusInternalServerError)
		return
	}

	// Respond with the test details.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(test)
}
