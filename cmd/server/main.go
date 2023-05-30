package main

import (
	"os"
	"path"

	"code-harvest.conner.dev/internal/server"
	"code-harvest.conner.dev/internal/storage"
	"code-harvest.conner.dev/pkg/logger"
)

// Set by linker flags
var (
	serverName string
	port       string
)

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	server, err := server.New(
		serverName,
		server.WithLog(logger.New(os.Stdout, logger.LevelInfo)),
		server.WithStorage(storage.DiskStorage(path.Join(homeDir, ".code-harvest"))),
	)
	if err != nil {
		panic(err)
	}

	server.Start(port)
}
