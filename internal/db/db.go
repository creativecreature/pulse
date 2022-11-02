package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	uri        string
	database   string
	collection string
	client     *mongo.Client
}

func New(uri, database, collection string) *mongoDB {
	return &mongoDB{
		uri:        uri,
		database:   database,
		collection: collection,
	}
}

func (m *mongoDB) Connect() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.uri))

	// I can't proceed without a database connection.
	if err != nil {
		panic(err)
	}

	m.client = client

	return func() {
		err := client.Disconnect(ctx)
		if err != nil {
			panic(err)
		}
	}
}

func (m *mongoDB) Save(item interface{}) error {
	_, err := m.client.Database(m.database).Collection(m.collection).InsertOne(context.Background(), item)
	return err
}
