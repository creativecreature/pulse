package storage

import (
	"context"

	"code-harvest.conner.dev/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBStorage struct {
	database   string
	collection string
	client     *mongo.Client
}

func New(client *mongo.Client, database, collection string) *MongoDBStorage {
	return &MongoDBStorage{client: client, database: database, collection: collection}
}

func (m *MongoDBStorage) Save(s *models.Session) error {
	sessionCollection := m.client.Database("codeharvest").Collection("sessions")
	_, err := sessionCollection.InsertOne(context.Background(), s)
	return err
}
