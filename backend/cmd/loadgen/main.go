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
	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/v1"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/controllers"
	"github.com/AkshayDubey29/MoniFlux/backend/internal/db/mongo"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	// Initialize logger with appropriate settings.
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration.
	cfg, err := v1.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB client.
	mongoClient, err := mongo.NewMongoClient(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer func() {
		if err := mongoClient.Close(); err != nil {
			logger.Errorf("Failed to close MongoDB client: %v", err)
		}
	}()

	// Initialize controller with MongoClient.
	controller := controllers.NewLoadGenController(cfg, logger, mongoClient)

	// Initialize handlers.
	handler := handlers.NewHandler(controller, logger)

	// Set up router.
	router := mux.NewRouter()

	// Define routes.
	router.HandleFunc("/tests", handler.StartTestHandler).Methods("POST")
	router.HandleFunc("/tests/schedule", handler.ScheduleTestHandler).Methods("POST")
	router.HandleFunc("/tests/cancel", handler.CancelTestHandler).Methods("POST")
	router.HandleFunc("/tests/restart", handler.RestartTestHandler).Methods("POST")
	router.HandleFunc("/tests/results", handler.SaveResultsHandler).Methods("POST")
	router.HandleFunc("/tests", handler.GetAllTestsHandler).Methods("GET")
	router.HandleFunc("/tests/{testID}", handler.GetTestByIDHandler).Methods("GET")

	// Start HTTP server.
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: router,
	}

	// Run server in a goroutine.
	go func() {
		logger.Infof("Starting server on port %d", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Graceful shutdown on interrupt signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("Shutting down server...")

	// Shutdown the server with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server Shutdown Failed:%+v", err)
	}

	logger.Info("Server exited gracefully")
}
