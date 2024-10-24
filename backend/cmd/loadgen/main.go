// backend/cmd/loadgen/main.go

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/handlers"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/utils"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/controllers"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/db/mongo"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration (use default config file path or environment variable)
	configFile := "/app/configs/config.yaml" // Default path
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// If the file doesn't exist, don't use it
		configFile = ""
	}

	cfg, err := utils.LoadConfig(configFile)
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB client
	mongoClient, err := mongo.NewMongoClient(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	// Initialize controller with MongoClient's internal client
	controller := controllers.NewLoadGenController(cfg, logger, mongoClient.Client)

	// Initialize handlers
	handler := handlers.NewHandler(controller, logger)

	// Set up router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/tests", handler.StartTest).Methods("POST")
	router.HandleFunc("/tests/schedule", handler.ScheduleTest).Methods("POST")
	router.HandleFunc("/tests/cancel", handler.CancelTest).Methods("POST")
	router.HandleFunc("/tests/restart", handler.RestartTest).Methods("POST")
	router.HandleFunc("/tests/results", handler.SaveResults).Methods("POST")
	router.HandleFunc("/tests", handler.GetAllTests).Methods("GET")
	router.HandleFunc("/tests/{testID}", handler.GetTestByID).Methods("GET")
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET") // Health Check Endpoint

	// Start HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	go func() {
		logger.Infof("Starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Graceful shutdown on interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("Shutting down server...")

	// Shutdown the server with a timeout
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Fatalf("Server Shutdown Failed:%+v", err)
	}

	logger.Info("Server exited gracefully")
}
