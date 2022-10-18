package main

import (
	"context"
	"os"
	"time"

	"code-harvest.conner.dev/internal/server"
	"code-harvest.conner.dev/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Set by linker flags
var port string
var uri string

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)

	// Connect to mongodb. Cancel the context and disconnect from the client before main exits.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.PrintFatal(err, nil)
		}
	}()

	// We can't store the sessions without a connection to mongo.
	if err != nil {
		log.PrintFatal(err, nil)
	}

	log.PrintInfo("Starting up the server...", nil)
	storage := server.NewMongoStorage(client, "codeharvest", "sessions")
	server.New(log, storage).Start(port)
	log.PrintInfo("Shutting down...", nil)
}
