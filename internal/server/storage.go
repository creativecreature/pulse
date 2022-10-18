package server

import (
	"context"

	"code-harvest.conner.dev/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type Storage interface {
	Save(s *models.Session) error
}

type MongoDBStorage struct {
	database   string
	collection string
	client     *mongo.Client
}

func NewMongoStorage(client *mongo.Client, database, collection string) *MongoDBStorage {
	return &MongoDBStorage{client: client, database: database, collection: collection}
}

func (m *MongoDBStorage) Save(s *models.Session) error {
	sessionCollection := m.client.Database("codeharvest").Collection("sessions")
	_, err := sessionCollection.InsertOne(context.Background(), s)
	return err
}

type MemoryStorage struct {
	sessions []*models.Session
}

func (m *MemoryStorage) Save(s *models.Session) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *MemoryStorage) Get() []*models.Session {
	return m.sessions
}
