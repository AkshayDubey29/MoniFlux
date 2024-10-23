package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/api/routers"
	"github.com/AkshayDubey29/MoniFlux/backend/pkg/config"
	"github.com/AkshayDubey29/MoniFlux/backend/pkg/db"
	"github.com/AkshayDubey29/MoniFlux/backend/pkg/logger"
)

func main() {
	// Load configuration from config.yaml
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		// If configuration fails to load, log the error and exit
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize custom logger based on the configuration
	customLogger := logger.NewLogger(cfg.LogLevel)
	customLogger.Info("Custom logger initialized")

	// Initialize MongoDB connection
	if err := db.InitializeMongo(cfg); err != nil {
		customLogger.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	customLogger.Info("MongoDB initialized")

	// Set up the API router with all routes and middleware
	router := routers.SetupRouter()

	// Define the HTTP server with timeouts and the router as the handler
	srv := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a separate goroutine to allow graceful shutdown
	go func() {
		customLogger.Infof("Starting API server on port %s", cfg.APIPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			customLogger.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Channel to listen for interrupt or terminate signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	customLogger.Info("Shutdown signal received, initiating graceful shutdown...")

	// Create a context with timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt to gracefully shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		customLogger.Fatalf("Server forced to shutdown: %v", err)
	}

	// Disconnect MongoDB client
	if db.MongoDBClient != nil {
		if err := db.MongoDBClient.Client.Disconnect(ctx); err != nil {
			customLogger.Errorf("Error disconnecting MongoDB client: %v", err)
		} else {
			customLogger.Info("MongoDB connection closed")
		}
	}

	customLogger.Info("Server gracefully stopped")
}
