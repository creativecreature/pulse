package main

import (
	"os"

	"code-harvest.conner.dev/logger"
	"code-harvest.conner.dev/server"
	"code-harvest.conner.dev/storage"
)

// ldflags.
var (
	serverName string
	port       string
)

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)

	server, err := server.New(
		serverName,
		server.WithLog(log),
		server.WithStorage(storage.DiskStorage()),
	)
	if err != nil {
		log.PrintFatal(err, nil)
	}

	err = server.Start(port)
	if err != nil {
		log.PrintFatal(err, nil)
	}
}
