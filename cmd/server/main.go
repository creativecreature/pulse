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
	port       string
)

func main() {
	log := logger.New(os.Stdout, logger.LevelInfo)
	server, err := server.New(
		serverName,
		server.WithLog(log),
		server.WithStorage(storage.DiskStorage()),
	)

	properties := map[string]string{}
	properties["serverName"] = serverName
	properties["port"] = port

	if err != nil {
		log.PrintFatal(err, properties)
	}

	err = server.Start(port)
	if err != nil {
		log.PrintFatal(err, properties)
	}
}
