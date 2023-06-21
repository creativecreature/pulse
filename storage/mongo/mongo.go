package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Constants for the collection names
const (
	daily   = "daily"
	weekly  = "weekly"
	monthly = "monthly"
	yearly  = "yearly"
)

type db struct {
	uri      string
	database string
	client   *mongo.Client
}

func NewDB(uri, database string) *db {
	return &db{
		uri:      uri,
		database: database,
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
