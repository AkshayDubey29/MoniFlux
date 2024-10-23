// pkg/db/db.go

package db

import (
	"context"
	"log"
	"time"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/config/v1"
	dbmongo "github.com/AkshayDubey29/MoniFlux/backend/internal/db/mongo" // Alias to avoid conflict with external mongo package
	"github.com/sirupsen/logrus"
	mongoPkg "go.mongodb.org/mongo-driver/mongo"             // Alias for external mongo package
	mongoDriver "go.mongodb.org/mongo-driver/mongo/readpref" // Alias for external mongo readpref
)

// MongoDBClient is a global MongoDB client that can be accessed by other parts of the application.
var MongoDBClient *dbmongo.MongoClient

// InitializeMongo initializes the MongoDB client and connects to the MongoDB server.
// It verifies the connection by pinging the MongoDB server and stores the client globally.
func InitializeMongo(cfg *v1.Config, logger *logrus.Logger) error {
	// Initialize the MongoDB client using the internal mongo package
	client, err := dbmongo.NewMongoClient(cfg, logger)
	if err != nil {
		logger.Errorf("Error creating MongoDB client: %v", err)
		return err
	}

	// Context for pinging the MongoDB server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ping the MongoDB server to verify the connection
	if err := client.Client.Ping(ctx, mongoDriver.Primary()); err != nil {
		logger.Errorf("Error pinging MongoDB server: %v", err)
		return err
	}

	logger.Info("Successfully connected to MongoDB")

	// Set the global MongoDB client
	MongoDBClient = client
	return nil
}

// GetMongoDB returns the MongoDB database instance that can be used for database operations.
func GetMongoDB() *mongoPkg.Database {
	if MongoDBClient == nil {
		log.Fatalf("MongoDB client has not been initialized")
	}
	return MongoDBClient.DB
}

// CloseMongoConnection closes the connection to the MongoDB server.
func CloseMongoConnection(logger *logrus.Logger) error {
	if MongoDBClient != nil {
		logger.Info("Closing MongoDB connection")
		return MongoDBClient.Client.Disconnect(context.Background())
	}
	return nil
}
