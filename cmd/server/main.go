package main

import (
	"context"
	"os"
	"time"

	"code-harvest.conner.dev/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Set by linker flags
var port string
var uri string

var heartbeatTTL = time.Minute * 10
var heartbeatInterval = time.Second * 10

func main() {
	// Connect to mongodb. Cancel the context and disconnect from the client before main exits.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// We can't store the sessions without a connection to mongo.
	if err != nil {
		panic(err)
	}

	app := &CodeHarvestApp{
		logger: logger.New(os.Stdout, logger.LevelInfo),
		ctx:    ctx,
		client: client,
	}

	// The ECG runs a heartbeat check to see if a session has gone stale (no activity for x minutes).
	ecg := ECG{
		check:     app.checkHeartbeat,
		stopChan:  make(chan bool),
		heartbeat: time.NewTicker(heartbeatInterval),
	}

	// The RPC server makes our handlers accessible to the client over tcp.
	rpcServer := RPCServer{
		rcvr: app,
	}

	app.logger.PrintDebug("Starting up..", nil)

	// Run blocks until we receive a shutdown signal or an error.
	if err = run(app.handleShutdown, &ecg, &rpcServer); err != nil {
		app.logger.PrintError(err, nil)
	}

	// Check if we are shutting down because of an error or a signal.
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}

	app.logger.PrintInfo("Server shutdown successfully", nil)
}
