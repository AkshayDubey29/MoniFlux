// backend/internal/db/mongo/mongo.go

package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/AkshayDubey29/MoniFlux/backend/internal/common"
)

// MongoClient wraps the MongoDB client and provides access to the database and collections.
type MongoClient struct {
	Client *mongo.Client
	DB     *mongo.Database
	Logger *logrus.Logger
}

// NewMongoClient initializes a new MongoDB client based on the provided configuration.
// It sets up connection options, establishes the connection, and pings the database to ensure connectivity.
func NewMongoClient(cfg *common.Config, logger *logrus.Logger) (*MongoClient, error) {
	// Define MongoDB client options
	clientOpts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetMaxPoolSize(100).                        // Maximum number of connections in the pool
		SetMinPoolSize(10).                         // Minimum number of connections in the pool
		SetConnectTimeout(10 * time.Second).        // Timeout for establishing connections
		SetMaxConnIdleTime(5 * time.Minute).        // Maximum idle time for connections
		SetHeartbeatInterval(10 * time.Second).     // Interval between heartbeat pings to the primary
		SetServerSelectionTimeout(5 * time.Second). // Timeout for server selection
		SetRetryWrites(true).                       // Enable retryable writes
		SetRetryReads(true).                        // Enable retryable reads
		SetDirect(false).                           // Enable read preference and server selection
		SetAppName("MoniFlux")                      // Application name for MongoDB logs and monitoring

	// Create a new MongoDB client
	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		logger.Errorf("Failed to create MongoDB client: %v", err)
		return nil, err
	}

	// Create a context with a timeout for connecting to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Connect to MongoDB
	err = client.Connect(ctx)
	if err != nil {
		logger.Errorf("Failed to connect to MongoDB: %v", err)
		return nil, err
	}

	// Ping the MongoDB server to verify connectivity
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Errorf("Failed to ping MongoDB: %v", err)
		return nil, err
	}

	logger.Info("Successfully connected to MongoDB")

	// Access the specified database
	db := client.Database(cfg.MongoDB)

	return &MongoClient{
		Client: client,
		DB:     db,
		Logger: logger,
	}, nil
}

// Disconnect gracefully disconnects the MongoDB client.
// It ensures that all pending operations are completed before closing the connection.
func (m *MongoClient) Disconnect(ctx context.Context) error {
	if m.Client == nil {
		return errors.New("MongoClient is not initialized")
	}

	err := m.Client.Disconnect(ctx)
	if err != nil {
		m.Logger.Errorf("Error disconnecting MongoDB client: %v", err)
		return err
	}

	m.Logger.Info("Successfully disconnected from MongoDB")
	return nil
}

// Ping checks the connectivity to the MongoDB server.
// It can be used for health checks to ensure the database is reachable.
func (m *MongoClient) Ping(ctx context.Context) error {
	if m.Client == nil {
		return errors.New("MongoClient is not initialized")
	}

	err := m.Client.Ping(ctx, nil)
	if err != nil {
		m.Logger.Errorf("Ping to MongoDB failed: %v", err)
		return err
	}

	m.Logger.Info("Ping to MongoDB succeeded")
	return nil
}

// GetCollection returns a MongoDB collection based on the provided name.
// It abstracts the collection access, making it easier to manage database operations.
func (m *MongoClient) GetCollection(name string) *mongo.Collection {
	return m.DB.Collection(name)
}

// InsertOne inserts a single document into the specified collection.
// It returns the inserted ID or an error if the operation fails.
func (m *MongoClient) InsertOne(ctx context.Context, collectionName string, document interface{}) (interface{}, error) {
	collection := m.GetCollection(collectionName)
	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		m.Logger.Errorf("Failed to insert document into %s: %v", collectionName, err)
		return nil, err
	}
	return result.InsertedID, nil
}

// FindOne retrieves a single document from the specified collection based on the filter.
// It decodes the result into the provided result interface.
func (m *MongoClient) FindOne(ctx context.Context, collectionName string, filter interface{}, result interface{}) error {
	collection := m.GetCollection(collectionName)
	err := collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			m.Logger.Warnf("No document found in %s with filter %+v", collectionName, filter)
			return err
		}
		m.Logger.Errorf("Failed to find document in %s: %v", collectionName, err)
		return err
	}
	return nil
}

// UpdateOne updates a single document in the specified collection based on the filter.
// It returns the update result or an error if the operation fails.
func (m *MongoClient) UpdateOne(ctx context.Context, collectionName string, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	collection := m.GetCollection(collectionName)
	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		m.Logger.Errorf("Failed to update document in %s: %v", collectionName, err)
		return nil, err
	}
	return result, nil
}

// FindAll retrieves all documents from the specified collection based on the filter.
// It returns a slice of results or an error if the operation fails.
func (m *MongoClient) FindAll(ctx context.Context, collectionName string, filter interface{}, results interface{}) error {
	collection := m.GetCollection(collectionName)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		m.Logger.Errorf("Failed to find documents in %s: %v", collectionName, err)
		return err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, results); err != nil {
		m.Logger.Errorf("Failed to decode documents from %s: %v", collectionName, err)
		return err
	}

	return nil
}

// CreateIndex creates an index on the specified collection based on the index model.
// It returns the index name or an error if the operation fails.
func (m *MongoClient) CreateIndex(ctx context.Context, collectionName string, indexModel mongo.IndexModel) (string, error) {
	collection := m.GetCollection(collectionName)
	indexView := collection.Indexes()

	indexName, err := indexView.CreateOne(ctx, indexModel)
	if err != nil {
		m.Logger.Errorf("Failed to create index on %s: %v", collectionName, err)
		return "", err
	}

	m.Logger.Infof("Successfully created index %s on %s", indexName, collectionName)
	return indexName, nil
}
