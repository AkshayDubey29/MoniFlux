// backend/internal/api/handlers/handler.go

package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/models"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/controllers"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/services/authentication"
	validator "github.com/go-playground/validator/v10" // Aliased import for clarity
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

// Handler encapsulates the controller, validator, and logger.
type Handler struct {
	Controller  *controllers.LoadGenController
	AuthService *authentication.AuthenticationService
	Validator   *validator.Validate
	Logger      *logrus.Logger
}

// NewHandler creates a new Handler instance.
func NewHandler(controller *controllers.LoadGenController, authService *authentication.AuthenticationService, logger *logrus.Logger) *Handler {
	return &Handler{
		Controller:  controller,
		AuthService: authService,
		Validator:   validator.New(),
		Logger:      logger,
	}
}

// StartTest handles the initiation of a new load test.
func (h *Handler) StartTest(w http.ResponseWriter, r *http.Request) {
	var test models.Test
	if err := json.NewDecoder(r.Body).Decode(&test); err != nil {
		h.Logger.Errorf("Failed to decode test: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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
		w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(test)
}

// ScheduleTest handles scheduling a load test.
func (h *Handler) ScheduleTest(w http.ResponseWriter, r *http.Request) {
	var schedule models.ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		h.Logger.Errorf("Failed to decode schedule request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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
		w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedule)
}

// CancelTest handles cancelling a load test.
func (h *Handler) CancelTest(w http.ResponseWriter, r *http.Request) {
	var cancelReq models.CancelRequest
	if err := json.NewDecoder(r.Body).Decode(&cancelReq); err != nil {
		h.Logger.Errorf("Failed to decode cancel request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Cancel the test using the controller.
	err := h.Controller.CancelTest(r.Context(), cancelReq.TestID)
	if err != nil {
		if errors.Is(err, mongoDriver.ErrNoDocuments) {
			http.Error(w, "Test not found", http.StatusNotFound)
			return
		}
		h.Logger.Errorf("Failed to cancel test: %v", err)
		http.Error(w, "Failed to cancel test", http.StatusInternalServerError)
		return
	}

	// Respond with a success message.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
}

// RestartTest handles restarting a load test.
func (h *Handler) RestartTest(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("Entered RestartTest handler")

	var restartReq models.RestartRequest

	// Log the request body for debugging
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.Logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	h.Logger.Debugf("Request body: %s", string(bodyBytes))

	// Decode the request payload
	if err := json.Unmarshal(bodyBytes, &restartReq); err != nil {
		h.Logger.Errorf("Failed to decode restart request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	h.Logger.Debugf("Decoded RestartRequest: %+v", restartReq)

	// Validate the restart request
	if err := h.Validator.Struct(restartReq); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		h.Logger.Debug("Sent validation error response")
		return
	}
	h.Logger.Debug("RestartRequest validation successful")

	// Construct the internal call to the loadgen service's /tests/restart endpoint
	loadgenURL := fmt.Sprintf("%s/tests/restart", h.Controller.Config.Server.LoadgenURL)
	h.Logger.Debugf("Internal call to %s", loadgenURL)

	// Perform the HTTP request to loadgen service
	resp, err := http.Post(loadgenURL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		h.Logger.Errorf("Internal call to /tests/restart failed: %v", err)
		http.Error(w, "Failed to restart test", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read response from loadgen
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.Logger.Errorf("Failed to read response from loadgen service: %v", err)
		http.Error(w, "Failed to read response from loadgen service", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.Logger.Errorf("Loadgen service returned error: %s", string(body))
		http.Error(w, "Loadgen service error", http.StatusInternalServerError)
		return
	}

	h.Logger.Debug("RestartTest internal call successful")

	// Respond with the success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body) // Return the response from loadgen to the client
}

// SaveResults handles saving the results of a load test.
func (h *Handler) SaveResults(w http.ResponseWriter, r *http.Request) {
	var results models.TestResults
	if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
		h.Logger.Errorf("Failed to decode test results: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
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
		w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

// GetAllTests handles retrieving all active and scheduled tests.
func (h *Handler) GetAllTests(w http.ResponseWriter, r *http.Request) {
	tests, err := h.Controller.GetAllTests(r.Context())
	if err != nil {
		h.Logger.Errorf("Failed to get all tests: %v", err)
		http.Error(w, "Failed to retrieve tests", http.StatusInternalServerError)
		return
	}

	// Respond with the list of tests.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tests)
}

// GetTestByID handles retrieving a specific test by its TestID.
func (h *Handler) GetTestByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	testID, exists := vars["testID"]
	if !exists {
		h.Logger.Errorf("TestID not provided in URL")
		http.Error(w, "TestID is required", http.StatusBadRequest)
		return
	}

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(test)
}

// RegisterUser handles user registration.
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Errorf("Failed to decode registration request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the registration request.
	if err := h.Validator.Struct(req); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Register the user using the authentication service.
	if err := h.AuthService.RegisterUser(req.Username, req.Email, req.Password); err != nil {
		h.Logger.Errorf("Failed to register user: %v", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// AuthenticateUser handles user authentication.
func (h *Handler) AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Errorf("Failed to decode authentication request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate the authentication request.
	if err := h.Validator.Struct(req); err != nil {
		h.Logger.Errorf("Validation error: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}

	// Authenticate the user using the authentication service.
	token, err := h.AuthService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		h.Logger.Errorf("Failed to authenticate user: %v", err)
		http.Error(w, "Failed to authenticate user", http.StatusUnauthorized)
		return
	}

	// Respond with the JWT token.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// CreateTest handles the creation of a new load test.
func (h *Handler) CreateTest(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debugf("Received request to create test at %v", time.Now())

	// Log incoming request body
	var requestBodyBytes []byte
	requestBodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Errorf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	h.Logger.Debugf("CreateTest request body: %s", string(requestBodyBytes))

	// Decode the request body into the Test struct
	var test models.Test
	if err := json.Unmarshal(requestBodyBytes, &test); err != nil {
		h.Logger.Errorf("Failed to decode create-test request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	h.Logger.Debugf("Decoded Test object: %+v", test)

	// Validate the test struct
	if err := h.Validator.Struct(test); err != nil {
		h.Logger.Errorf("Validation error in create-test: %v", err)
		var validationErrors []models.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   err.Field(),
				Message: err.Tag(),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(validationErrors)
		return
	}
	h.Logger.Debug("Test object passed validation")

	// Call the controller to create the test
	h.Logger.Debug("Calling Controller.CreateTest")
	if err := h.Controller.CreateTest(r.Context(), &test); err != nil {
		h.Logger.Errorf("Failed to create test: %v", err)
		http.Error(w, "Failed to create test", http.StatusInternalServerError)
		return
	}
	h.Logger.Debugf("Test created successfully: %+v", test)

	// Respond with the created test
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(test); err != nil {
		h.Logger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
	h.Logger.Debug("CreateTest response sent successfully")
}

// HealthCheck handles the /health endpoint.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
