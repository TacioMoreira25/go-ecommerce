package database

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStore implements the Storage interface for MongoDB
type MongoStore struct {
	dbName string
	client *mongo.Client
	DB *mongo.Database
}

// NewMongoStore creates a new MongoStore
func NewMongoStore(dbName string) *MongoStore {
	return &MongoStore{
		dbName: dbName,
	}
}

// Connect connects to a MongoDB instance
func (s *MongoStore) Connect(ctx context.Context, uri string) error {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	// Ping the primary
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}
	s.client = client
	s.DB = s.client.Database(s.dbName)

	return nil
}

// Disconnect disconnects from the MongoDB instance
func (s *MongoStore) Disconnect(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}
