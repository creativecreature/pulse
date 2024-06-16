package main

import (
	"context"
	"os"
	"path"
	"time"

	"github.com/creativecreature/pulse/mongo"
	"github.com/creativecreature/pulse/server"
)

// ldflags.
var (
	uri        string
	db         string
	serverName string
	port       string
)

func main() {
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	client := mongo.New(uri, db)
	defer func() {
		disconnectErr := client.Disconnect(timeoutCtx)
		if disconnectErr != nil {
			panic(disconnectErr)
		}
	}()

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	// Create the path for the log storages segment files.
	segmentPath := path.Join(userHomeDir, ".pulse", "segments")

	server, err := server.New(
		serverName,
		segmentPath,
		client,
	)
	if err != nil {
		panic(err)
	}

	err = server.Start(port)
	if err != nil {
		panic(err)
	}
}
