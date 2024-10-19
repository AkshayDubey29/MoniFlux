package mongo

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "github.com/AkshayDubey29/MoniFlux/internal/config/v1"
)

// MongoClient wraps the MongoDB client
type MongoClient struct {
    Client *mongo.Client
    DB     *mongo.Database
}

// NewMongoClient initializes a new MongoDB client
func NewMongoClient(cfg *v1.Config) (*MongoClient, error) {
    clientOptions := options.Client().ApplyURI(cfg.MongoURI)
    client, err := mongo.NewClient(clientOptions)
    if err != nil {
        return nil, err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err = client.Connect(ctx)
    if err != nil {
        return nil, err
    }

    db := client.Database(cfg.MongoDB)

    return &MongoClient{
        Client: client,
        DB:     db,
    }, nil
}
