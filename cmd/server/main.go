package main

import (
	"os"

	"code-harvest.conner.dev/internal/server"
	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/pkg/logger"
)

// Set by linker flags
var (
	serverName string
	uri        string
	port       string
)

func main() {
	server, err := server.New(
		serverName,
		server.WithLog(logger.New(os.Stdout, logger.LevelInfo)),
		server.WithStorage(storage.MongoDB(uri, "codeharvest", "sessions")),
	)
	if err != nil {
		panic(err)
	}

	server.Start(port)
}
