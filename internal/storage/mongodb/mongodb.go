package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type db struct {
	uri        string
	database   string
	collection string
	client     *mongo.Client
}

func New(uri, database, collection string) *db {
	return &db{
		uri:        uri,
		database:   database,
		collection: collection,
	}
}

func (m *db) Connect() func() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.uri))
	// Can't proceed without a database connection.
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

func (m *db) Save(item interface{}) error {
	_, err := m.client.Database(m.database).
		Collection(m.collection).
		InsertOne(context.Background(), item)
	return err
}
